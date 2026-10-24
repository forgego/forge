package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/stores"
	"github.com/forgego/forge/stores/storestest"
	"github.com/stretchr/testify/require"
)

func TestLifecycle(t *testing.T) {
	t.Run("InvalidConfigFailsBeforeListenerAndHidesSecrets", func(t *testing.T) {
		port := freeTCPPort(t)
		dbPassword := "db-secret-password-xyz123"
		secretKey := "short-secret"
		csrfKey := "short-csrf"
		sessionKey := "short-session"

		cfg := config.NewConfig()
		cfg.Set("app.env", "production")
		cfg.Set("server.host", "127.0.0.1")
		cfg.Set("server.port", port)
		cfg.Set("database.password", dbPassword)
		cfg.Set("security.secret_key", secretKey)
		cfg.Set("security.csrf_secret_key", csrfKey)
		cfg.Set("security.session_secret", sessionKey)

		settings := config.LoadSettings(cfg)

		// 1. NewServer fails before listener opens
		_, err := NewServer(cfg, settings, nil, WithDatabase(storestest.OpenMemory(t)))
		require.Error(t, err, "NewServer must fail with invalid production config")
		require.NotContains(t, err.Error(), dbPassword, "error must not leak database password")
		require.NotContains(t, err.Error(), secretKey, "error must not leak secret key")
		require.NotContains(t, err.Error(), csrfKey, "error must not leak csrf secret")
		require.NotContains(t, err.Error(), sessionKey, "error must not leak session secret")

		// Verify listener never opened
		conn, dialErr := net.DialTimeout("tcp", "127.0.0.1:"+port, 100*time.Millisecond)
		if dialErr == nil {
			_ = conn.Close()
			t.Fatal("listener should not have opened when NewServer failed")
		}

		// 2. Start() also fails before listener opens and does not leak secrets
		srv := &Server{
			Server:   &http.Server{Addr: "127.0.0.1:" + port},
			config:   cfg,
			settings: settings,
		}
		startErr := srv.Start()
		require.Error(t, startErr, "Start() must fail with invalid config")
		require.NotContains(t, startErr.Error(), dbPassword)
		require.NotContains(t, startErr.Error(), secretKey)
		require.NotContains(t, startErr.Error(), csrfKey)
		require.NotContains(t, startErr.Error(), sessionKey)

		conn, dialErr = net.DialTimeout("tcp", "127.0.0.1:"+port, 100*time.Millisecond)
		if dialErr == nil {
			_ = conn.Close()
			t.Fatal("listener should not have opened when Start() failed")
		}
	})

	t.Run("InFlightRequestCompletesWithinShutdownTimeout", func(t *testing.T) {
		port := freeTCPPort(t)
		settings := &config.Settings{
			App:    config.AppSettings{Env: "test"},
			Server: config.ServerSettings{Host: "127.0.0.1", Port: port, GracefulTimeout: 5, ReadTimeout: 5, WriteTimeout: 5},
		}
		srv, err := NewServer(config.NewConfig(), settings, nil, WithDatabase(storestest.OpenMemory(t)))
		require.NoError(t, err)

		started := make(chan struct{})
		release := make(chan struct{})
		srv.RegisterRoutes(func(r *Router) {
			r.Get("/in-flight", func(w http.ResponseWriter, _ *http.Request) {
				close(started)
				<-release
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, "in-flight completed")
			})
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		serveErr := make(chan error, 1)
		go func() { serveErr <- srv.serveUntil(ctx, nil) }()

		base := "http://127.0.0.1:" + port
		require.Eventually(t, func() bool {
			conn, err := net.Dial("tcp", "127.0.0.1:"+port)
			if err != nil {
				return false
			}
			_ = conn.Close()
			return true
		}, 5*time.Second, 10*time.Millisecond)

		type reqResult struct {
			status int
			body   string
			err    error
		}
		inFlight := make(chan reqResult, 1)
		go func() {
			resp, err := http.Get(base + "/in-flight")
			if err != nil {
				inFlight <- reqResult{err: err}
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			inFlight <- reqResult{status: resp.StatusCode, body: string(body), err: err}
		}()

		<-started

		cancel()

		// Listener should be closed during drain
		require.Eventually(t, func() bool {
			conn, err := net.Dial("tcp", "127.0.0.1:"+port)
			if err != nil {
				return true
			}
			_ = conn.Close()
			return false
		}, 5*time.Second, 10*time.Millisecond)

		select {
		case err := <-serveErr:
			t.Fatalf("serveUntil returned before in-flight request finished: %v", err)
		default:
		}

		close(release)
		res := <-inFlight
		require.NoError(t, res.err)
		require.Equal(t, http.StatusOK, res.status)
		require.Equal(t, "in-flight completed", res.body)

		select {
		case err := <-serveErr:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("serveUntil did not return after in-flight request finished")
		}
	})

	t.Run("DatabaseClosedAfterDrainNotBefore", func(t *testing.T) {
		port := freeTCPPort(t)
		database, err := db.NewDBWithDriver("sqlite3", ":memory:", db.WithMaxOpenConns(1), db.WithMaxIdleConns(1))
		require.NoError(t, err)
		require.NoError(t, stores.Migrate(context.Background(), database))

		settings := &config.Settings{
			App:    config.AppSettings{Env: "test"},
			Server: config.ServerSettings{Host: "127.0.0.1", Port: port, GracefulTimeout: 5, ReadTimeout: 5, WriteTimeout: 5},
		}
		srv, err := NewServer(config.NewConfig(), settings, nil, WithDatabase(database))
		require.NoError(t, err)

		started := make(chan struct{})
		release := make(chan struct{})
		dbAliveDuringDrain := make(chan bool, 1)

		srv.RegisterRoutes(func(r *Router) {
			r.Get("/db-drain", func(w http.ResponseWriter, _ *http.Request) {
				close(started)
				<-release
				// Ensure database is accessible during drain
				err := database.Ping(context.Background())
				dbAliveDuringDrain <- (err == nil)
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, "db-drain completed")
			})
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		serveErr := make(chan error, 1)
		go func() { serveErr <- srv.serveUntil(ctx, nil) }()

		base := "http://127.0.0.1:" + port
		require.Eventually(t, func() bool {
			conn, err := net.Dial("tcp", "127.0.0.1:"+port)
			if err != nil {
				return false
			}
			_ = conn.Close()
			return true
		}, 5*time.Second, 10*time.Millisecond)

		go func() {
			resp, err := http.Get(base + "/db-drain")
			if err == nil {
				_ = resp.Body.Close()
			}
		}()

		<-started

		// Start shutdown: drain begins
		cancel()

		// Listener closes
		require.Eventually(t, func() bool {
			conn, err := net.Dial("tcp", "127.0.0.1:"+port)
			if err != nil {
				return true
			}
			_ = conn.Close()
			return false
		}, 5*time.Second, 10*time.Millisecond)

		// Release the in-flight request so it completes its db check during drain
		close(release)
		alive := <-dbAliveDuringDrain
		require.True(t, alive, "database must NOT be closed before or during the drain")

		// Drain finishes
		select {
		case err := <-serveErr:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("serveUntil did not return after drain finished")
		}

		// Database must be closed after the drain
		pingErr := database.Ping(context.Background())
		require.Error(t, pingErr, "database must be closed after the drain")
		require.Contains(t, pingErr.Error(), "database is closed")
	})

	t.Run("ShutdownExceedingTimeoutReturnsError", func(t *testing.T) {
		port := freeTCPPort(t)
		settings := &config.Settings{
			App:    config.AppSettings{Env: "test"},
			Server: config.ServerSettings{Host: "127.0.0.1", Port: port, GracefulTimeout: 1, ReadTimeout: 5, WriteTimeout: 5},
		}
		srv, err := NewServer(config.NewConfig(), settings, nil, WithDatabase(storestest.OpenMemory(t)))
		require.NoError(t, err)

		started := make(chan struct{})
		done := make(chan struct{})
		t.Cleanup(func() { close(done) })

		srv.RegisterRoutes(func(r *Router) {
			r.Get("/hang", func(w http.ResponseWriter, _ *http.Request) {
				close(started)
				select {
				case <-done:
				case <-time.After(10 * time.Second):
				}
			})
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		serveErr := make(chan error, 1)
		go func() { serveErr <- srv.serveUntil(ctx, nil) }()

		base := "http://127.0.0.1:" + port
		require.Eventually(t, func() bool {
			conn, err := net.Dial("tcp", "127.0.0.1:"+port)
			if err != nil {
				return false
			}
			_ = conn.Close()
			return true
		}, 5*time.Second, 10*time.Millisecond)

		go func() {
			resp, err := http.Get(base + "/hang")
			if err == nil {
				_ = resp.Body.Close()
			}
		}()

		<-started

		// Trigger shutdown with an in-flight request that outlives GracefulTimeout (1s)
		cancel()

		select {
		case err := <-serveErr:
			require.Error(t, err, "shutdown exceeding timeout must return an error")
			require.Contains(t, err.Error(), "context deadline exceeded")
		case <-time.After(5 * time.Second):
			t.Fatal("serveUntil did not return within bounded timeout")
		}
	})
}
