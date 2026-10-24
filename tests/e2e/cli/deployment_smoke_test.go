package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeploymentSmoke(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	// 1. Setup consumer application (SQLite)
	app, port := newTrackerJourney(t, ctx, "sqlite", nil)
	app.env = map[string]string{
		"FORGE_APP_ENV":                  "production",
		"FORGE_APP_DEBUG":                "false",
		"FORGE_SECURITY_SECRET_KEY":      "production-secret-key-at-least-32-chars-long",
		"FORGE_SECURITY_CSRF_SECRET_KEY": "production-csrf-key-at-least-32-chars-long",
		"FORGE_SECURITY_SESSION_SECRET":  "production-session-key-at-least-32-chars-long",
	}

	app.forge("generate", "--api", "--strict", "--models", "./app/work", "--output", "./app/work")
	app.goCmd("mod", "tidy")
	app.forge("makemigrations", "initial", "--auto", "--models", "./app/work")

	// 2. Build server binary
	app.goCmd("build", "-o", filepath.Join("bin", serverBinaryName()), "./cmd/server")

	// 3. Schema changes: integrity check, dry-run preview, and migrate up
	app.forge("migrate", "recover", "--verify")
	dryRunOut := app.forge("migrate", "up", "--dry-run")
	assert.Contains(t, dryRunOut, "Dry-run mode: Preview of migrations")

	upOut := app.forge("migrate", "up")
	assert.Contains(t, upOut, "Migrations applied successfully")

	statusOut := app.forge("migrate", "status")
	assert.Contains(t, statusOut, "Current Version: 1")
	assert.Contains(t, statusOut, "Status: OK")

	// 4. Start server
	stop := app.startServer(port)
	defer stop()

	base := fmt.Sprintf("http://127.0.0.1:%s", port)
	httpClient := &http.Client{Timeout: 5 * time.Second}

	// Health checks: readiness 200, health 200, liveness 200
	readyResp, err := httpClient.Get(base + "/health/ready")
	require.NoError(t, err, "GET /health/ready")
	defer readyResp.Body.Close()
	require.Equal(t, http.StatusOK, readyResp.StatusCode, "readiness endpoint must return 200")
	readyBody, err := io.ReadAll(readyResp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(readyBody), `"status":"ready"`)

	healthResp, err := httpClient.Get(base + "/health")
	require.NoError(t, err, "GET /health")
	defer healthResp.Body.Close()
	require.Equal(t, http.StatusOK, healthResp.StatusCode, "health endpoint must return 200")
	healthBody, err := io.ReadAll(healthResp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(healthBody), `"status":"healthy"`)

	liveResp, err := httpClient.Get(base + "/health/live")
	require.NoError(t, err, "GET /health/live")
	defer liveResp.Body.Close()
	require.Equal(t, http.StatusOK, liveResp.StatusCode, "liveness endpoint must return 200")

	// 5. Rollback step
	// Stop server before running migration commands so the SQLite database is not locked.
	stop()

	rollbackOut := app.forge("migrate", "rollback")
	assert.Contains(t, rollbackOut, "Migration rolled back successfully")

	statusAfterRollback := app.forge("migrate", "status")
	assert.Contains(t, statusAfterRollback, "Current Version: 0")
}
