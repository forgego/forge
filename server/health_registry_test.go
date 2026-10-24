package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/log"
	"github.com/forgego/forge/server"
	"github.com/forgego/forge/stores/storestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHealthServer(t *testing.T, name string) *server.Server {
	t.Helper()
	cfg := config.NewConfig()
	settings := &config.Settings{
		App: config.AppSettings{
			Name: name,
			Env:  "test",
		},
		Server: config.ServerSettings{
			Host:            "127.0.0.1",
			Port:            "0",
			HealthCheckPath: "/health",
		},
	}
	db := storestest.OpenMemory(t)
	srv, err := server.NewServer(cfg, settings, log.NewNopLogger(), server.WithDatabase(db))
	require.NoError(t, err)
	require.NotNil(t, srv)
	return srv
}

func TestHealthRegistryIsolation(t *testing.T) {
	srvA := newTestHealthServer(t, "srv-a")
	srvB := newTestHealthServer(t, "srv-b")

	// Verify Uptime accessor works on registry
	assert.True(t, srvA.Health().Uptime() >= 0)
	assert.True(t, srvB.Health().Uptime() >= 0)

	// Both servers start ready
	recA := httptest.NewRecorder()
	srvA.Handler.ServeHTTP(recA, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	assert.Equal(t, http.StatusOK, recA.Code)

	recB := httptest.NewRecorder()
	srvB.Handler.ServeHTTP(recB, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	assert.Equal(t, http.StatusOK, recB.Code)

	// A failing check registered on A's Health() makes A's /health/ready unhealthy while B's stays healthy
	srvA.Health().RegisterFunc("fail_check_a", func(ctx context.Context) error {
		return errors.New("failing check on A")
	})

	recA = httptest.NewRecorder()
	srvA.Handler.ServeHTTP(recA, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	assert.Equal(t, http.StatusServiceUnavailable, recA.Code)
	var statusA server.HealthStatus
	require.NoError(t, json.Unmarshal(recA.Body.Bytes(), &statusA))
	assert.Equal(t, "not ready", statusA.Status)
	assert.Equal(t, "not ready", statusA.Checks["fail_check_a"])

	recB = httptest.NewRecorder()
	srvB.Handler.ServeHTTP(recB, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	assert.Equal(t, http.StatusOK, recB.Code)
	var statusB server.HealthStatus
	require.NoError(t, json.Unmarshal(recB.Body.Bytes(), &statusB))
	assert.Equal(t, "ready", statusB.Status)
	assert.NotContains(t, statusB.Checks, "fail_check_a")

	// Clean up with Unregister
	srvA.Health().Unregister("fail_check_a")

	recA = httptest.NewRecorder()
	srvA.Handler.ServeHTTP(recA, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	assert.Equal(t, http.StatusOK, recA.Code)
	require.NoError(t, json.Unmarshal(recA.Body.Bytes(), &statusA))
	assert.Equal(t, "ready", statusA.Status)

	// Default-registry checks still apply to both (compat)
	server.RegisterHealthCheckFunc("compat_default_fail", func(ctx context.Context) error {
		return errors.New("default registry failed")
	})
	t.Cleanup(func() {
		server.UnregisterHealthCheck("compat_default_fail")
	})

	recA = httptest.NewRecorder()
	srvA.Handler.ServeHTTP(recA, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	assert.Equal(t, http.StatusServiceUnavailable, recA.Code)

	recB = httptest.NewRecorder()
	srvB.Handler.ServeHTTP(recB, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	assert.Equal(t, http.StatusServiceUnavailable, recB.Code)

	// Clean up default-registry check with Unregister
	server.UnregisterHealthCheck("compat_default_fail")

	recA = httptest.NewRecorder()
	srvA.Handler.ServeHTTP(recA, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	assert.Equal(t, http.StatusOK, recA.Code)

	recB = httptest.NewRecorder()
	srvB.Handler.ServeHTTP(recB, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	assert.Equal(t, http.StatusOK, recB.Code)

	// A name in the server registry wins over the default registry
	server.RegisterHealthCheckFunc("override_check", func(ctx context.Context) error {
		return errors.New("default check fails")
	})
	t.Cleanup(func() {
		server.UnregisterHealthCheck("override_check")
	})

	srvA.Health().RegisterFunc("override_check", func(ctx context.Context) error {
		return nil
	})
	t.Cleanup(func() {
		srvA.Health().Unregister("override_check")
	})

	recA = httptest.NewRecorder()
	srvA.Handler.ServeHTTP(recA, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	assert.Equal(t, http.StatusOK, recA.Code)

	recB = httptest.NewRecorder()
	srvB.Handler.ServeHTTP(recB, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	assert.Equal(t, http.StatusServiceUnavailable, recB.Code)

	// Cleanup
	srvA.Health().Unregister("override_check")
	server.UnregisterHealthCheck("override_check")
}
