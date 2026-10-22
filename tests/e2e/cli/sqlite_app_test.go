package cli

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteFreshUserCanManageProjectsThroughAdminAndAPI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	app, port := newTrackerJourney(t, ctx, "sqlite", nil)
	app.generate()
	assert.Contains(t, app.forge("makemigrations", "initial", "--auto", "--models", "./app/work"), "000001_initial.up.sql")
	app.forge("migrate", "up")
	assert.Contains(t, app.forge("migrate", "status"), "Current Version: 1")

	database, err := sql.Open("sqlite3", filepath.Join(app.dir, "forge.db"))
	require.NoError(t, err)
	defer database.Close()
	app.db = database
	_, err = database.ExecContext(ctx, `INSERT INTO members (name, api_token) VALUES (?, ?)`, "alice", "token-alice")
	require.NoError(t, err)

	app.env = map[string]string{"FORGE_ADMIN_USERNAME": "admin", "FORGE_ADMIN_PASSWORD": "secret"}
	stop := startSQLiteRunserver(t, app, port)
	defer stop()

	base := fmt.Sprintf("http://127.0.0.1:%s", port)
	admin := &journeyClient{t: t, base: base + "/admin/api", authScheme: "Bearer"}
	status, body := admin.do(http.MethodPost, "/login", "", map[string]any{"username": "admin", "password": "secret"})
	require.Equal(t, http.StatusOK, status, "admin login: %v", body)
	adminToken, ok := body["token"].(string)
	require.True(t, ok && adminToken != "", "admin login returned no token: %v", body)

	status, body = admin.do(http.MethodPost, "/projects/", adminToken, map[string]any{"name": "Admin project"})
	require.Equal(t, http.StatusCreated, status, "admin create: %v", body)
	adminID := jsonID(t, body)
	status, body = admin.do(http.MethodGet, "/projects/", adminToken, nil)
	require.Equal(t, http.StatusOK, status, "admin list: %v", body)
	assert.EqualValues(t, 1, body["count"])
	status, body = admin.do(http.MethodPatch, fmt.Sprintf("/projects/%d/", adminID), adminToken, map[string]any{"name": "Updated by admin"})
	require.Equal(t, http.StatusOK, status, "admin update: %v", body)
	assert.Equal(t, "Updated by admin", body["name"])
	status, body = admin.do(http.MethodDelete, fmt.Sprintf("/projects/%d/", adminID), adminToken, nil)
	require.Equal(t, http.StatusNoContent, status, "admin delete: %v", body)
	assert.Zero(t, app.count(`SELECT count(*) FROM projects WHERE id = ?`, adminID))

	api := &journeyClient{t: t, base: base + "/api/v1"}
	status, body = api.do(http.MethodPost, "/projects/", "token-alice", map[string]any{"name": "API project"})
	require.Equal(t, http.StatusCreated, status, "API create: %v", body)
	apiID := jsonID(t, body)
	status, body = api.do(http.MethodGet, "/projects/", "token-alice", nil)
	require.Equal(t, http.StatusOK, status, "API list: %v", body)
	assert.EqualValues(t, 1, body["count"])
	status, body = api.do(http.MethodPatch, fmt.Sprintf("/projects/%d", apiID), "token-alice", map[string]any{"name": "Updated by API"})
	require.Equal(t, http.StatusOK, status, "API update: %v", body)
	assert.Equal(t, "Updated by API", body["name"])
	status, body = api.do(http.MethodDelete, fmt.Sprintf("/projects/%d", apiID), "token-alice", nil)
	require.Equal(t, http.StatusNoContent, status, "API delete: %v", body)
	assert.Zero(t, app.count(`SELECT count(*) FROM projects WHERE id = ?`, apiID))

	status, body = admin.do(http.MethodGet, "/config", adminToken, nil)
	require.Equal(t, http.StatusOK, status, "admin token use: %v", body)
	status, body = admin.do(http.MethodPost, "/logout", adminToken, nil)
	require.Equal(t, http.StatusOK, status, "admin logout: %v", body)
	status, body = admin.do(http.MethodGet, "/config", adminToken, nil)
	assert.Equal(t, http.StatusUnauthorized, status, "revoked admin token: %v", body)
}

// startSQLiteRunserver builds the CLI from this checkout and exercises its
// runserver command. A process group lets cleanup stop go run and its server.
func startSQLiteRunserver(t *testing.T, app *journeyApp, port string) func() {
	t.Helper()
	root, err := repoRoot()
	require.NoError(t, err)
	cliPath := filepath.Join(app.dir, "bin", "forge")
	build := exec.CommandContext(app.ctx, "go", "build", "-o", cliPath, "./cmd/forge")
	build.Dir = root
	out, err := build.CombinedOutput()
	require.NoError(t, err, "build forge CLI: %s", out)

	cmd := exec.Command(cliPath, "runserver")
	cmd.Dir = app.dir
	cmd.Env = app.environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var logs bytes.Buffer
	cmd.Stdout = &logs
	cmd.Stderr = &logs
	require.NoError(t, cmd.Start())
	stopped := false
	stop := func() {
		if stopped {
			return
		}
		stopped = true
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(15 * time.Second):
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			<-done
		}
		if t.Failed() {
			t.Logf("runserver logs:\n%s", logs.String())
		}
	}
	t.Cleanup(stop)
	waitForHealthy(t, port)
	return stop
}
