package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/forgego/forge/admin"
	admincore "github.com/forgego/forge/admin/core"
	forgeapi "github.com/forgego/forge/api"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity"
	"github.com/forgego/forge/identity/backends"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/log"
	"github.com/forgego/forge/schema"
	"github.com/forgego/forge/server"
	"github.com/forgego/forge/stores"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type product struct {
	schema.BaseSchema
	ID    int64  `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Price int64  `db:"price" json:"price"`
}

func (product) Meta() schema.Meta { return schema.Meta{TableName: "products"} }

func (product) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.Int64Field("price"),
	}
}

type exclusiveModelA struct {
	schema.BaseSchema
	ID    int64  `db:"id" json:"id"`
	Title string `db:"title" json:"title"`
}

func (exclusiveModelA) Meta() schema.Meta { return schema.Meta{TableName: "exclusive_as"} }

func (exclusiveModelA) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
	}
}

type stubProductViewSet struct {
	serverName string
}

func (s *stubProductViewSet) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"server": s.serverName, "resource": "products"})
}

func (s *stubProductViewSet) Create(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (s *stubProductViewSet) Retrieve(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *stubProductViewSet) Update(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *stubProductViewSet) PartialUpdate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *stubProductViewSet) Destroy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

type stubExclusiveViewSet struct {
	serverName string
}

func (s *stubExclusiveViewSet) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"server": s.serverName, "resource": "exclusive"})
}

func (s *stubExclusiveViewSet) Create(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (s *stubExclusiveViewSet) Retrieve(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *stubExclusiveViewSet) Update(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *stubExclusiveViewSet) PartialUpdate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *stubExclusiveViewSet) Destroy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

type testServerInstance struct {
	name      string
	baseURL   string
	database  *db.DB
	server    *server.Server
	settings  *config.Settings
	adminSite *admin.Site
	apiRouter *forgeapi.Router
	client    *http.Client
	listener  net.Listener
	serveErr  chan error
}

func newTestServerInstance(t *testing.T, name string, registerExclusive bool) *testServerInstance {
	t.Helper()
	ctx := context.Background()

	dbFile := filepath.Join(t.TempDir(), fmt.Sprintf("%s.sqlite", name))
	database, err := db.NewDBWithDriver("sqlite3", dbFile, db.WithMaxOpenConns(1))
	require.NoError(t, err)

	require.NoError(t, stores.Migrate(ctx, database))

	schemaDDL := []string{
		`CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price INTEGER
		);`,
		`CREATE TABLE IF NOT EXISTS exclusive_as (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			email TEXT NOT NULL,
			password TEXT NOT NULL,
			first_name TEXT,
			last_name TEXT,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			is_staff BOOLEAN NOT NULL DEFAULT 0,
			is_superuser BOOLEAN NOT NULL DEFAULT 0,
			is_locked BOOLEAN NOT NULL DEFAULT 0,
			email_verified BOOLEAN NOT NULL DEFAULT 0,
			phone_number TEXT,
			phone_verified BOOLEAN NOT NULL DEFAULT 0,
			timezone TEXT,
			locale TEXT,
			language TEXT,
			bio TEXT,
			website TEXT,
			location TEXT,
			avatar TEXT,
			password_changed_at TIMESTAMP,
			password_expires_at TIMESTAMP,
			must_change_password BOOLEAN NOT NULL DEFAULT 0,
			locked_at TIMESTAMP,
			locked_reason TEXT,
			failed_login_attempts INTEGER NOT NULL DEFAULT 0,
			last_failed_login_at TIMESTAMP,
			email_verified_at TIMESTAMP,
			date_joined TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);`,
	}
	for _, ddl := range schemaDDL {
		_, err = database.ExecContext(ctx, ddl)
		require.NoError(t, err)
	}

	settings := &config.Settings{
		App: config.AppSettings{
			Name:    name,
			Env:     "development",
			Version: "1.0.0",
			Debug:   false,
		},
		Server: config.ServerSettings{
			Host:            "127.0.0.1",
			Port:            "0",
			HealthCheckPath: "/health",
			InfoEndpoint:    true,
		},
		Security: config.SecuritySettings{
			SessionSecret: fmt.Sprintf("secret-session-key-%s-0123456789012345", name),
		},
		Admin: config.AdminSettings{
			Enabled: true,
			Path:    "/admin",
		},
	}

	site := admin.NewSite(name)
	uiConfig := site.GetUIConfig()
	uiConfig.Prefix = settings.Admin.Path
	site.WithUIConfig(uiConfig)
	site.SetDB(database)

	userRepo := repository.NewUserRepository(database)
	site.SetLoginAuthenticator(backends.NewStaffLoginAuthenticator(backends.NewPasswordBackend(userRepo)))

	_, err = admin.RegisterWithSite(site, &admincore.Config[product]{})
	require.NoError(t, err)

	apiRouter := forgeapi.NewRouter("/api/v1")
	apiRouter.Register("products", &stubProductViewSet{serverName: name})

	if registerExclusive {
		_, err = admin.RegisterWithSite(site, &admincore.Config[exclusiveModelA]{})
		require.NoError(t, err)
		apiRouter.Register("exclusive", &stubExclusiveViewSet{serverName: name})
	}

	srv, err := server.NewServer(config.NewConfig(), settings, log.NewNopLogger(), server.WithDatabase(database))
	require.NoError(t, err)

	srv.RegisterRoutes(func(r *server.Router) {
		apiRouter.RegisterRoutes(r)
		r.Mount(settings.Admin.Path, site.Handler())
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(listener)
	}()

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
		Timeout: 5 * time.Second,
	}

	baseURL := fmt.Sprintf("http://%s", listener.Addr().String())

	// Wait for server to accept connections
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", listener.Addr().String(), 50*time.Millisecond)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, 2*time.Second, 10*time.Millisecond)

	return &testServerInstance{
		name:      name,
		baseURL:   baseURL,
		database:  database,
		server:    srv,
		settings:  settings,
		adminSite: site,
		apiRouter: apiRouter,
		client:    client,
		listener:  listener,
		serveErr:  serveErr,
	}
}

func (inst *testServerInstance) get(t *testing.T, path string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, inst.baseURL+path, nil)
	require.NoError(t, err)
	resp, err := inst.client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}

func (inst *testServerInstance) postJSON(t *testing.T, path string, payload any) (int, string) {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, inst.baseURL+path, bytes.NewReader(data))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := inst.client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}

func (inst *testServerInstance) shutdown(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = inst.server.Shutdown(ctx)
	select {
	case err := <-inst.serveErr:
		if err != nil && err != http.ErrServerClosed {
			t.Logf("server %s exited with: %v", inst.name, err)
		}
	case <-time.After(1 * time.Second):
	}
}

func TestAppIsolation(t *testing.T) {
	baselineGoroutines := runtime.NumGoroutine()

	serverA := newTestServerInstance(t, "instance-a", true)
	serverB := newTestServerInstance(t, "instance-b", false)

	t.Run("1_RegistrationsIndependent", func(t *testing.T) {
		statusA, bodyA := serverA.get(t, "/api/v1/exclusive")
		assert.Equal(t, http.StatusOK, statusA)
		assert.Contains(t, bodyA, "exclusive")

		statusB, bodyB := serverB.get(t, "/api/v1/exclusive")
		assert.Equal(t, http.StatusNotFound, statusB, "exclusive route registered on A must return 404 on B")
		assert.NotContains(t, bodyB, "exclusive")

		// Both servers serve their own "products" endpoint
		statusProdA, bodyProdA := serverA.get(t, "/api/v1/products")
		assert.Equal(t, http.StatusOK, statusProdA)
		assert.Contains(t, bodyProdA, "instance-a")

		statusProdB, bodyProdB := serverB.get(t, "/api/v1/products")
		assert.Equal(t, http.StatusOK, statusProdB)
		assert.Contains(t, bodyProdB, "instance-b")
	})

	t.Run("2_SettingsIndependent", func(t *testing.T) {
		serverA.settings.App.Debug = true
		serverA.settings.App.Name = "renamed-a"

		assert.False(t, serverB.settings.App.Debug, "server B debug setting should remain unchanged")
		assert.Equal(t, "instance-b", serverB.settings.App.Name, "server B name setting should remain unchanged")

		statusA, bodyA := serverA.get(t, "/info")
		require.Equal(t, http.StatusOK, statusA)
		var infoA map[string]any
		require.NoError(t, json.Unmarshal([]byte(bodyA), &infoA))
		assert.Equal(t, true, infoA["debug"])
		assert.Equal(t, "renamed-a", infoA["name"])

		statusB, bodyB := serverB.get(t, "/info")
		require.Equal(t, http.StatusOK, statusB)
		var infoB map[string]any
		require.NoError(t, json.Unmarshal([]byte(bodyB), &infoB))
		assert.Equal(t, false, infoB["debug"])
		assert.Equal(t, "instance-b", infoB["name"])
	})

	t.Run("3_PrincipalsIndependent", func(t *testing.T) {
		hashA, err := identity.HashPasswordWithCost("AliceSecretPass123!", 4)
		require.NoError(t, err)

		_, err = serverA.database.Exec(`
			INSERT INTO users (username, email, password, is_active, is_staff, is_superuser)
			VALUES ('alice', 'alice@example.invalid', ?, 1, 1, 1);
		`, hashA)
		require.NoError(t, err)

		// Alice logs in on server A -> 200 OK
		statusA, bodyA := serverA.postJSON(t, "/admin/api/login", map[string]string{
			"username": "alice",
			"password": "AliceSecretPass123!",
		})
		assert.Equal(t, http.StatusOK, statusA, "login on A should succeed: %s", bodyA)
		assert.Contains(t, bodyA, "token")

		// Alice attempts to log in on server B -> 401 Unauthorized
		statusB, bodyB := serverB.postJSON(t, "/admin/api/login", map[string]string{
			"username": "alice",
			"password": "AliceSecretPass123!",
		})
		assert.Equal(t, http.StatusUnauthorized, statusB, "user created on A must fail authentication on B: %s", bodyB)
	})

	t.Run("4_HealthIndependent", func(t *testing.T) {
		serverA.server.Health().RegisterFunc("component_a", func(ctx context.Context) error {
			return errors.New("degraded on A")
		})
		t.Cleanup(func() {
			serverA.server.Health().Unregister("component_a")
		})

		statusA, bodyA := serverA.get(t, "/health/ready")
		assert.Equal(t, http.StatusServiceUnavailable, statusA)
		assert.Contains(t, bodyA, "not ready")
		assert.Contains(t, bodyA, "component_a")

		statusB, bodyB := serverB.get(t, "/health/ready")
		assert.Equal(t, http.StatusOK, statusB)
		assert.Contains(t, bodyB, "ready")
		assert.NotContains(t, bodyB, "component_a")

		// Also verify /health
		statusHealthA, bodyHealthA := serverA.get(t, "/health")
		assert.Equal(t, http.StatusServiceUnavailable, statusHealthA)
		assert.Contains(t, bodyHealthA, "unhealthy")
		assert.Contains(t, bodyHealthA, "component_a")

		statusHealthB, bodyHealthB := serverB.get(t, "/health")
		assert.Equal(t, http.StatusOK, statusHealthB)
		assert.Contains(t, bodyHealthB, "healthy")
		assert.NotContains(t, bodyHealthB, "component_a")

		serverA.server.Health().Unregister("component_a")

		statusAAfter, bodyAAfter := serverA.get(t, "/health/ready")
		assert.Equal(t, http.StatusOK, statusAAfter)
		assert.Contains(t, bodyAAfter, "ready")
	})

	t.Run("5_StoppingALeavesBServing", func(t *testing.T) {
		serverA.shutdown(t)

		// Server A should now reject requests
		req, err := http.NewRequest(http.MethodGet, serverA.baseURL+"/api/v1/products", nil)
		require.NoError(t, err)
		_, err = serverA.client.Do(req)
		assert.Error(t, err, "requests to stopped server A should fail")

		// Server B continues serving requests successfully
		statusB, bodyB := serverB.get(t, "/api/v1/products")
		assert.Equal(t, http.StatusOK, statusB, "server B must continue serving after A shut down")
		assert.Contains(t, bodyB, "instance-b")
	})

	t.Run("6_GoroutinesReturnToBaseline", func(t *testing.T) {
		serverB.shutdown(t)

		deadline := time.Now().Add(2 * time.Second)
		var finalGoroutines int
		returned := false
		for time.Now().Before(deadline) {
			finalGoroutines = runtime.NumGoroutine()
			// Small margin allows for standard Go runtime/testing subtest goroutines and background tasks
			if finalGoroutines <= baselineGoroutines+4 {
				returned = true
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		if !returned {
			buf := make([]byte, 16384)
			n := runtime.Stack(buf, true)
			t.Logf("Stack dump of %d goroutines:\n%s", finalGoroutines, string(buf[:n]))
		}
		assert.True(t, returned, "goroutine count (%d) should return to within margin of baseline (%d)", finalGoroutines, baselineGoroutines)
	})
}
