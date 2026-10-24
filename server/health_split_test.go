package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/log"
	"github.com/forgego/forge/stores/storestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthVersusReadiness_DatabaseDown(t *testing.T) {
	cfg := config.NewConfig()
	settings := &config.Settings{
		App: config.AppSettings{
			Env: "test",
		},
		Server: config.ServerSettings{
			Host:            "127.0.0.1",
			Port:            "0",
			HealthCheckPath: "/health",
		},
	}

	database := storestest.OpenMemory(t)
	srv, err := NewServer(cfg, settings, log.NewNopLogger(), WithDatabase(database))
	require.NoError(t, err)
	require.NotNil(t, srv)

	// With DB up, both liveness and readiness return 200.
	liveReq := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	liveRec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(liveRec, liveReq)
	assert.Equal(t, http.StatusOK, liveRec.Code)

	var liveStatus HealthStatus
	require.NoError(t, json.Unmarshal(liveRec.Body.Bytes(), &liveStatus))
	assert.Equal(t, "alive", liveStatus.Status)

	readyReq := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	readyRec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(readyRec, readyReq)
	assert.Equal(t, http.StatusOK, readyRec.Code)

	var readyStatus HealthStatus
	require.NoError(t, json.Unmarshal(readyRec.Body.Bytes(), &readyStatus))
	assert.Equal(t, "ready", readyStatus.Status)

	// Close the database to simulate it being down.
	require.NoError(t, database.Close())

	// With DB down, liveness still returns 200 (no dependencies).
	liveDownReq := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	liveDownRec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(liveDownRec, liveDownReq)
	assert.Equal(t, http.StatusOK, liveDownRec.Code)

	var liveDownStatus HealthStatus
	require.NoError(t, json.Unmarshal(liveDownRec.Body.Bytes(), &liveDownStatus))
	assert.Equal(t, "alive", liveDownStatus.Status)

	// With DB down, readiness returns 503 (database check fails).
	readyDownReq := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	readyDownRec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(readyDownRec, readyDownReq)
	assert.Equal(t, http.StatusServiceUnavailable, readyDownRec.Code)

	var readyDownStatus HealthStatus
	require.NoError(t, json.Unmarshal(readyDownRec.Body.Bytes(), &readyDownStatus))
	assert.Equal(t, "not ready", readyDownStatus.Status)
	assert.Equal(t, "not ready", readyDownStatus.Checks["database"])
}
