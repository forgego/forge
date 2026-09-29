// Package stores_test runs two application instances against one
// PostgreSQL database with server.stores: database and checks that the
// state Forge used to keep in process memory is shared (#293).
package stores_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/api/throttling"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/log"
	"github.com/forgego/forge/schema"
	"github.com/forgego/forge/server"
	"github.com/forgego/forge/stores"
	"github.com/forgego/forge/stores/storestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ticket struct {
	schema.BaseSchema
	ID    int64  `db:"id" json:"id"`
	Title string `db:"title" json:"title"`
}

func (ticket) Meta() schema.Meta { return schema.Meta{TableName: "tickets"} }

func (ticket) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
	}
}

// instance is one application process: its own database connection, server,
// admin site and API throttle, as a load-balanced replica would have.
type instance struct {
	url string
}

func startInstance(t *testing.T, dsn string) *instance {
	t.Helper()
	ctx := context.Background()
	database, err := db.NewDBWithDriver("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	settings := &config.Settings{
		App:      config.AppSettings{Name: "replica", Env: "test"},
		Server:   config.ServerSettings{Host: "127.0.0.1", Port: "0", Stores: "database"},
		Security: config.SecuritySettings{SessionSecret: "multi-instance-session-secret-0123456789"},
		Admin:    config.AdminSettings{Enabled: true, Path: "/admin"},
	}

	site := admin.NewSite("replica")
	_, err = admin.RegisterWithSite(site, &core.Config[ticket]{})
	require.NoError(t, err)
	uiConfig := site.GetUIConfig()
	uiConfig.Prefix = settings.Admin.Path
	site.WithUIConfig(uiConfig)
	site.SetDB(database)
	require.NoError(t, site.UseStores(ctx, settings.Server.Stores))

	srv, err := server.NewServer(config.NewConfig(), settings, log.NewNopLogger(), server.WithDatabase(database))
	require.NoError(t, err)
	sessions := srv.SessionManager()
	require.NotNil(t, sessions)

	// The throttle counts in this instance's own connection, under the
	// name the server's default store factory gives an anonymous 5/min
	// throttle.
	shared, err := stores.New(database)
	require.NoError(t, err)
	throttle := throttling.NewAnonRateThrottle("5/min").WithStore(shared.RateLimiter("api:anon/5/min", 5, time.Minute))

	srv.RegisterRoutes(func(r *server.Router) {
		r.Get("/session/put", func(w http.ResponseWriter, req *http.Request) {
			sessions.Put(req, "user", req.URL.Query().Get("user"))
			w.WriteHeader(http.StatusNoContent)
		})
		r.Get("/session/get", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprint(w, sessions.SessionManager.GetString(req.Context(), "user"))
		})
		r.Get("/api/ping", func(w http.ResponseWriter, req *http.Request) {
			if err := throttling.CheckThrottles(req, nil, []throttling.Throttle{throttle}); err != nil {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			fmt.Fprint(w, "pong")
		})
		r.Mount(settings.Admin.Path, site.Handler())
	})
	ts := httptest.NewServer(srv.Handler)
	t.Cleanup(ts.Close)
	return &instance{url: ts.URL}
}

func (in *instance) do(t *testing.T, method, path, token string, cookie *http.Cookie, body any) (*http.Response, []byte) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, in.url+path, reader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, data
}

func (in *instance) login(t *testing.T, password string) (int, string) {
	t.Helper()
	resp, data := in.do(t, http.MethodPost, "/admin/api/login", "", nil, map[string]string{"username": "ops", "password": password})
	var body struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(data, &body)
	return resp.StatusCode, body.Token
}

func TestMultipleInstancesShareStores(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "ops")
	t.Setenv("FORGE_ADMIN_PASSWORD", "correct-horse")
	t.Cleanup(func() { throttling.SetDefaultStoreFactory(nil) })

	dsn := storestest.PostgresDSN(t)
	setup := storestest.Open(t, "postgres", dsn)
	require.NoError(t, stores.Migrate(context.Background(), setup))
	_, err := setup.Exec("CREATE TABLE tickets (id BIGSERIAL PRIMARY KEY, title TEXT NOT NULL)")
	require.NoError(t, err)

	a := startInstance(t, dsn)
	b := startInstance(t, dsn)

	t.Run("session created on A is valid on B", func(t *testing.T) {
		resp, _ := a.do(t, http.MethodGet, "/session/put?user=alice", "", nil, nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
		var cookie *http.Cookie
		for _, c := range resp.Cookies() {
			if c.Name == "forge_session" {
				cookie = c
			}
		}
		require.NotNil(t, cookie)
		_, body := b.do(t, http.MethodGet, "/session/get", "", cookie, nil)
		assert.Equal(t, "alice", string(body))
	})

	var token string
	t.Run("token issued on A authenticates on B", func(t *testing.T) {
		var status int
		status, token = a.login(t, "correct-horse")
		require.Equal(t, http.StatusOK, status)
		require.NotEmpty(t, token)
		resp, body := b.do(t, http.MethodGet, "/admin/api/config", token, nil, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	})

	t.Run("history written on A is listed on B", func(t *testing.T) {
		resp, body := a.do(t, http.MethodPost, "/admin/api/tickets/", token, nil, map[string]string{"title": "Disk full"})
		require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))
		var created struct {
			ID int64 `json:"id"`
		}
		require.NoError(t, json.Unmarshal(body, &created))

		resp, body = b.do(t, http.MethodGet, fmt.Sprintf("/admin/api/tickets/%d/history", created.ID), token, nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
		assert.Contains(t, string(body), `"action":"add"`)
		assert.Contains(t, string(body), `"user_id":"ops"`)
		assert.Contains(t, string(body), "Disk full")
	})

	t.Run("saved view on A is visible on B", func(t *testing.T) {
		resp, body := a.do(t, http.MethodPost, "/admin/api/saved-views/tickets/", token, nil,
			map[string]any{"name": "Open", "filters": map[string]any{"title": "Disk"}, "ordering": []string{"-id"}})
		require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))
		resp, body = b.do(t, http.MethodGet, "/admin/api/saved-views/tickets/", token, nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(body), `"name":"Open"`)
	})

	t.Run("throttle count is shared", func(t *testing.T) {
		codes := make([]int, 0, 6)
		for i := range 6 {
			in := a
			if i%2 == 1 {
				in = b
			}
			resp, _ := in.do(t, http.MethodGet, "/api/ping", "", nil, nil)
			codes = append(codes, resp.StatusCode)
		}
		assert.Equal(t, []int{200, 200, 200, 200, 200, 429}, codes, "5/min split across A and B")
	})

	t.Run("login lockout is shared", func(t *testing.T) {
		for i := range 5 {
			in := a
			if i%2 == 1 {
				in = b
			}
			status, _ := in.login(t, "wrong")
			require.Equal(t, http.StatusUnauthorized, status, "failure %d", i+1)
		}
		for _, in := range []*instance{a, b} {
			status, _ := in.login(t, "correct-horse")
			assert.Equal(t, http.StatusTooManyRequests, status, "five failures across A and B lock out both")
		}
	})

	t.Run("state survives a restart", func(t *testing.T) {
		c := startInstance(t, dsn)
		resp, body := c.do(t, http.MethodGet, "/admin/api/config", token, nil, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "a new instance accepts the existing token: %s", string(body))
		resp, body = c.do(t, http.MethodGet, "/admin/api/saved-views/tickets/", token, nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, strings.Contains(string(body), `"name":"Open"`))
	})
}
