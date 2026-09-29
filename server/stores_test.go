package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/throttling"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/log"
	"github.com/forgego/forge/stores"
	"github.com/forgego/forge/stores/storestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func storesSettings(kind string) *config.Settings {
	return &config.Settings{
		App:      config.AppSettings{Name: "stores", Env: "test"},
		Server:   config.ServerSettings{Host: "127.0.0.1", Port: "0", Stores: kind},
		Security: config.SecuritySettings{SessionSecret: "test-session-secret-key-that-is-long-enough"},
	}
}

func TestNewServer_DatabaseStoresNeedAMigratedDatabase(t *testing.T) {
	cfg := config.NewConfig()

	_, err := NewServer(cfg, storesSettings("database"), log.NewNopLogger())
	require.ErrorContains(t, err, "pass server.WithDatabase(database)")

	unmigrated := storestest.Open(t, "sqlite3", storestest.SQLitePath(t))
	_, err = NewServer(cfg, storesSettings("database"), log.NewNopLogger(), WithDatabase(unmigrated))
	require.ErrorIs(t, err, stores.ErrNotMigrated)

	_, err = NewServer(cfg, storesSettings("redis"), log.NewNopLogger(), WithDatabase(unmigrated))
	require.ErrorContains(t, err, `invalid server.stores "redis"`)

	for _, kind := range []string{"", "memory"} {
		srv, err := NewServer(cfg, storesSettings(kind), log.NewNopLogger(), WithDatabase(unmigrated))
		require.NoError(t, err, "memory stores never touch the database")
		require.NotNil(t, srv.SessionManager())
	}
}

// sessionServer builds a server whose /put stores a value in the session
// and whose /get reads it back.
func sessionServer(t *testing.T, settings *config.Settings, opts ...Option) *httptest.Server {
	t.Helper()
	srv, err := NewServer(config.NewConfig(), settings, log.NewNopLogger(), opts...)
	require.NoError(t, err)
	sessions := srv.SessionManager()
	require.NotNil(t, sessions)
	srv.RegisterRoutes(func(r *Router) {
		r.Get("/put", func(w http.ResponseWriter, req *http.Request) {
			sessions.Put(req, "user", req.URL.Query().Get("user"))
			w.WriteHeader(http.StatusNoContent)
		})
		r.Get("/get", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprint(w, sessions.SessionManager.GetString(req.Context(), "user"))
		})
	})
	ts := httptest.NewServer(srv.Handler)
	t.Cleanup(ts.Close)
	return ts
}

func sessionRoundTrip(t *testing.T, writer, reader *httptest.Server) string {
	t.Helper()
	resp, err := http.Get(writer.URL + "/put?user=alice")
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	var cookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "forge_session" {
			cookie = c
		}
	}
	require.NotNil(t, cookie, "the session cookie is set")

	req, err := http.NewRequest(http.MethodGet, reader.URL+"/get", nil)
	require.NoError(t, err)
	req.AddCookie(cookie)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body)
}

func TestNewServer_DatabaseStoresShareSessionsBetweenInstances(t *testing.T) {
	t.Cleanup(func() { throttling.SetDefaultStoreFactory(nil) })
	for _, backend := range storestest.Backends() {
		t.Run(backend.Name, func(t *testing.T) {
			first, dsn := backend.OpenMigrated(t)
			second := storestest.Open(t, backend.Driver, dsn)
			a := sessionServer(t, storesSettings("database"), WithDatabase(first))
			b := sessionServer(t, storesSettings("database"), WithDatabase(second))

			assert.Equal(t, "alice", sessionRoundTrip(t, a, b), "a session created on A is read on B")

			var rows int
			require.NoError(t, first.QueryRowContext(context.Background(), "SELECT count(*) FROM forge_sessions").Scan(&rows))
			assert.Equal(t, 1, rows)
		})
	}
}

// Before #293 every instance kept sessions in its own memory; that stays
// the behavior with server.stores: memory.
func TestNewServer_MemoryStoresKeepSessionsPerInstance(t *testing.T) {
	a := sessionServer(t, storesSettings("memory"))
	b := sessionServer(t, storesSettings("memory"))
	assert.Equal(t, "alice", sessionRoundTrip(t, a, a))
	assert.Equal(t, "", sessionRoundTrip(t, a, b))
}

func TestNewServer_DatabaseStoresShareAPIThrottling(t *testing.T) {
	t.Cleanup(func() { throttling.SetDefaultStoreFactory(nil) })
	database, _ := storestest.Backends()[0].OpenMigrated(t)
	throttle := throttling.NewAnonRateThrottle("2/min")

	_, err := NewServer(config.NewConfig(), storesSettings("database"), log.NewNopLogger(), WithDatabase(database))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:1234"
	require.NoError(t, throttling.CheckThrottles(req, nil, []throttling.Throttle{throttle}))
	var hits int
	require.NoError(t, database.QueryRowContext(context.Background(),
		"SELECT hits FROM forge_rate_limits WHERE bucket = $1", "rate:api:anon/2/min:throttle_anon_10.0.0.9").Scan(&hits))
	assert.Equal(t, 1, hits, "throttles without their own store count in the database")
}
