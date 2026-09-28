package cli

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/tests/testhelpers"
)

// TestCLIPostgresAppJourney is the release gate for a freshly generated
// external application on PostgreSQL. It builds nothing by hand that Forge
// generates: `forge new` scaffolds the project outside the repository,
// `forge generate --api --strict` writes the models' code and REST API,
// `forge makemigrations` and `forge migrate` build the schema, and the
// scaffolded server serves real HTTP. The test only writes what a developer
// writes: models, route wiring, one line in main.go, and a small command
// that uses the transaction API.
//
// When RUN_POSTGRES_TESTS=1 or FORGE_REQUIRE_DB=1 is set, a missing
// PostgreSQL fails the test instead of skipping it.
func TestCLIPostgresAppJourney(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	database, dbEnv := requireJourneyPostgres(t, ctx)

	workdir, cleanup := testhelpers.TempWorkdir(t, "forge-e2e-postgres-app-")
	defer cleanup()
	projectDir := filepath.Join(workdir, "tracker")
	app := &journeyApp{t: t, ctx: ctx, dir: projectDir, env: dbEnv, db: database}

	// Scaffold a PostgreSQL project and an app, then add the developer's code.
	app.forgeIn(workdir, "new", "tracker", "--path", projectDir,
		"--template", "simple", "--database", "postgres", "--docker=false")
	require.NoError(t, patchGeneratedProject(projectDir))
	app.forge("add", "app", "work")
	app.writeModels("")
	testhelpers.WriteFileString(t, filepath.Join(projectDir, "app", "work", "routes.go"), trackerRoutes)
	testhelpers.WriteFileString(t, filepath.Join(projectDir, "cmd", "txcheck", "main.go"), trackerTxCommand)
	registerAppInMain(t, filepath.Join(projectDir, "cmd", "server", "main.go"))
	port := randomPort(t)
	updateServerPort(t, filepath.Join(projectDir, "config", "config.yaml"), port)

	// Generate, compile and migrate.
	app.generate()
	out := app.forge("makemigrations", "initial", "--auto", "--models", "./app/work")
	assert.Contains(t, out, "000001_initial.up.sql")
	app.forge("migrate", "up")
	assert.Contains(t, app.forge("migrate", "status"), "Current Version: 1")
	for _, fk := range []string{"fk_tasks_project_id", "fk_tasks_owner_id"} {
		assert.Equal(t, 1, app.count(`SELECT count(*) FROM pg_constraint WHERE conname = $1 AND contype = 'f'`, fk),
			"foreign key %s must exist", fk)
	}
	assert.Contains(t, app.forge("makemigrations", "unchanged", "--auto", "--models", "./app/work"), "No changes detected")
	assert.Len(t, app.migrationFiles(), 2, "unchanged models must not write a migration")

	// Members authenticate with API tokens; they have no API of their own.
	var alice, bob int64
	require.NoError(t, database.QueryRowContext(ctx,
		`INSERT INTO members (name, api_token) VALUES ('alice', 'token-alice') RETURNING id`).Scan(&alice))
	require.NoError(t, database.QueryRowContext(ctx,
		`INSERT INTO members (name, api_token) VALUES ('bob', 'token-bob') RETURNING id`).Scan(&bob))

	stop := app.startServer(port)
	api := &journeyClient{t: t, base: fmt.Sprintf("http://127.0.0.1:%s/api/v1", port)}

	// Unauthenticated and unknown-token requests are rejected and write nothing.
	status, body := api.do(http.MethodGet, "/projects/", "", nil)
	assert.Equal(t, http.StatusForbidden, status, "no credentials: %s", body)
	assert.Equal(t, "Authentication credentials were not provided", body["detail"])
	status, body = api.do(http.MethodPost, "/projects/", "", map[string]any{"name": "Anonymous"})
	assert.Equal(t, http.StatusForbidden, status, "anonymous create: %s", body)
	status, body = api.do(http.MethodPost, "/projects/", "not-a-token", map[string]any{"name": "Forged"})
	assert.Equal(t, http.StatusUnauthorized, status, "unknown token: %s", body)
	assert.Zero(t, app.count(`SELECT count(*) FROM projects`), "rejected creates must not write")

	// Create and read back a project.
	status, body = api.do(http.MethodPost, "/projects/", "token-alice", map[string]any{
		"name": "Apollo", "description": "Moon program",
		"id": 777, "created_at": "2000-01-01T00:00:00Z", "updated_at": "2000-01-01T00:00:00Z",
	})
	require.Equal(t, http.StatusCreated, status, "create project: %s", body)
	projectID := jsonID(t, body)
	assert.NotEqual(t, int64(777), projectID, "the primary key comes from the database, not the body")
	assert.Equal(t, "Apollo", body["name"])
	before := app.projectRow(projectID)
	assert.Equal(t, "Apollo", before.name)
	assert.Greater(t, before.createdAt.Year(), 2000, "created_at comes from the database, not the body")
	assert.Zero(t, app.count(`SELECT count(*) FROM projects WHERE id = 777`))

	status, body = api.do(http.MethodGet, fmt.Sprintf("/projects/%d", projectID), "token-bob", nil)
	require.Equal(t, http.StatusOK, status, "get project: %s", body)
	assert.Equal(t, "Moon program", body["description"])
	status, body = api.do(http.MethodGet, "/projects/", "token-bob", nil)
	require.Equal(t, http.StatusOK, status, "list projects: %s", body)
	assert.EqualValues(t, 1, body["count"])

	// Validation errors are 400s that name the field and change nothing.
	status, body = api.do(http.MethodPost, "/projects/", "token-alice", map[string]any{"description": "no name"})
	assert.Equal(t, http.StatusBadRequest, status, "missing name: %s", body)
	assert.Contains(t, body["errors"], "name")
	status, body = api.do(http.MethodPost, "/projects/", "token-alice", map[string]any{"name": strings.Repeat("x", 101)})
	assert.Equal(t, http.StatusBadRequest, status, "name too long: %s", body)
	assert.Contains(t, body["errors"], "name")
	status, body = api.doRaw(http.MethodPost, "/projects/", "token-alice", []byte(`{"name":`))
	assert.Equal(t, http.StatusBadRequest, status, "malformed JSON: %s", body)
	status, body = api.do(http.MethodPatch, fmt.Sprintf("/projects/%d", projectID), "token-alice", map[string]any{"name": ""})
	assert.Equal(t, http.StatusBadRequest, status, "blank name: %s", body)
	status, body = api.do(http.MethodPut, fmt.Sprintf("/projects/%d", projectID), "token-alice", map[string]any{"description": "no name"})
	assert.Equal(t, http.StatusBadRequest, status, "PUT without name: %s", body)
	assert.Equal(t, 1, app.count(`SELECT count(*) FROM projects`), "rejected creates must not write")
	assert.Equal(t, before, app.projectRow(projectID), "rejected updates must not change the row")

	// The primary key and auto fields cannot be changed through the body.
	status, body = api.do(http.MethodPatch, fmt.Sprintf("/projects/%d", projectID), "token-alice", map[string]any{
		"name": "Apollo 11", "id": projectID + 100, "created_at": "2000-01-01T00:00:00Z",
	})
	require.Equal(t, http.StatusOK, status, "rename project: %s", body)
	assert.Equal(t, projectID, jsonID(t, body))
	after := app.projectRow(projectID)
	assert.Equal(t, "Apollo 11", after.name)
	assert.True(t, before.createdAt.Equal(after.createdAt), "created_at must not change: %v -> %v", before.createdAt, after.createdAt)
	assert.Zero(t, app.count(`SELECT count(*) FROM projects WHERE id = $1`, projectID+100))

	// Tasks reference a project and an owner.
	status, body = api.do(http.MethodPost, "/tasks/", "token-alice", map[string]any{
		"project_id": projectID, "owner_id": alice, "title": "Build the lander",
	})
	require.Equal(t, http.StatusCreated, status, "create task: %s", body)
	taskID := jsonID(t, body)
	assert.Equal(t, false, body["done"])
	status, body = api.do(http.MethodPost, "/tasks/", "token-bob", map[string]any{
		"project_id": projectID, "owner_id": bob, "title": "Train the crew",
	})
	require.Equal(t, http.StatusCreated, status, "create second task: %s", body)
	bobTaskID := jsonID(t, body)

	status, body = api.do(http.MethodPost, "/tasks/", "token-alice", map[string]any{
		"project_id": projectID + 1000, "owner_id": alice, "title": "Orphan",
	})
	assert.Equal(t, http.StatusBadRequest, status, "unknown project: %s", body)
	status, body = api.do(http.MethodPost, "/tasks/", "token-alice", map[string]any{"project_id": projectID, "owner_id": alice})
	assert.Equal(t, http.StatusBadRequest, status, "missing title: %s", body)
	assert.Contains(t, body["errors"], "title")
	assert.Equal(t, 2, app.count(`SELECT count(*) FROM tasks`), "rejected task creates must not write")

	status, body = api.do(http.MethodGet, fmt.Sprintf("/tasks/?project_id=%d", projectID), "token-bob", nil)
	require.Equal(t, http.StatusOK, status, "filter tasks by project: %s", body)
	assert.EqualValues(t, 2, body["count"])
	status, body = api.do(http.MethodGet, fmt.Sprintf("/tasks/?owner_id=%d", alice), "token-bob", nil)
	require.Equal(t, http.StatusOK, status, "filter tasks by owner: %s", body)
	assert.EqualValues(t, 1, body["count"])

	// Only a task's owner may change or delete it.
	status, body = api.do(http.MethodPatch, fmt.Sprintf("/tasks/%d", taskID), "token-bob", map[string]any{"title": "Hijacked"})
	assert.Equal(t, http.StatusForbidden, status, "non-owner update: %s", body)
	status, body = api.do(http.MethodDelete, fmt.Sprintf("/tasks/%d", taskID), "token-bob", nil)
	assert.Equal(t, http.StatusForbidden, status, "non-owner delete: %s", body)
	assert.Equal(t, "Build the lander", app.taskTitle(taskID), "a denied write must not change the row")

	status, body = api.do(http.MethodPatch, fmt.Sprintf("/tasks/%d", taskID), "token-alice", map[string]any{"title": "Land", "done": true})
	require.Equal(t, http.StatusOK, status, "owner update: %s", body)
	assert.Equal(t, true, body["done"])
	assert.Equal(t, "Land", app.taskTitle(taskID))

	status, body = api.do(http.MethodDelete, fmt.Sprintf("/tasks/%d", bobTaskID), "token-bob", nil)
	require.Equal(t, http.StatusNoContent, status, "owner delete: %s", body)
	assert.Zero(t, app.count(`SELECT count(*) FROM tasks WHERE id = $1`, bobTaskID))
	status, body = api.do(http.MethodGet, fmt.Sprintf("/tasks/%d", bobTaskID), "token-alice", nil)
	assert.Equal(t, http.StatusNotFound, status, "deleted task: %s", body)

	// A transaction whose second write fails leaves nothing behind.
	out, code := app.txcheck("Rolled back", -1)
	assert.Equal(t, 3, code, "txcheck output: %s", out)
	assert.Contains(t, out, "rolled back")
	assert.Zero(t, app.count(`SELECT count(*) FROM projects WHERE name = 'Rolled back'`), "the first write must be rolled back")
	out, code = app.txcheck("Committed", alice)
	require.Equal(t, 0, code, "txcheck output: %s", out)
	assert.Equal(t, 1, app.count(`SELECT count(*) FROM projects p JOIN tasks t ON t.project_id = p.id WHERE p.name = 'Committed'`))

	stop()

	// Change the schema: add a field, regenerate, and migrate existing rows.
	app.writeModels(`schema.Int32Field("priority", schema.Default(0)),`)
	app.generate()
	out = app.forge("makemigrations", "add_priority", "--auto", "--models", "./app/work")
	assert.Contains(t, out, "000002_add_priority.up.sql")
	assert.Len(t, app.migrationFiles(), 4)
	upSQL, err := os.ReadFile(filepath.Join(projectDir, "migrations", "000002_add_priority.up.sql"))
	require.NoError(t, err)
	assert.Contains(t, string(upSQL), `ALTER TABLE tasks ADD COLUMN "priority" INTEGER DEFAULT 0;`)
	assert.NotContains(t, string(upSQL), "ALTER COLUMN", "only the new column may change")
	assert.NotContains(t, string(upSQL), "CONSTRAINT", "foreign keys are unchanged")
	app.forge("migrate", "up")
	assert.Contains(t, app.forge("migrate", "status"), "Current Version: 2")
	assert.Contains(t, app.forge("makemigrations", "unchanged", "--auto", "--models", "./app/work"), "No changes detected")

	assert.Zero(t, app.count(`SELECT count(*) FROM tasks WHERE priority IS DISTINCT FROM 0`), "existing rows take the default")
	assert.Equal(t, 1, app.count(`SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'projects' AND column_name = 'created_at' AND is_nullable = 'NO' AND column_default = 'now()'`),
		"created_at keeps NOT NULL DEFAULT now()")

	stop = app.startServer(port)
	defer stop()
	status, body = api.do(http.MethodGet, fmt.Sprintf("/tasks/%d", taskID), "token-bob", nil)
	require.Equal(t, http.StatusOK, status, "existing task after migration: %s", body)
	assert.Equal(t, "Land", body["title"])
	assert.EqualValues(t, 0, body["priority"])
	status, body = api.do(http.MethodPatch, fmt.Sprintf("/tasks/%d", taskID), "token-alice", map[string]any{"priority": 5})
	require.Equal(t, http.StatusOK, status, "set new field: %s", body)
	assert.Equal(t, 5, app.count(`SELECT priority FROM tasks WHERE id = $1`, taskID))
	status, body = api.do(http.MethodGet, "/projects/", "token-alice", nil)
	require.Equal(t, http.StatusOK, status, "existing projects after migration: %s", body)
	assert.EqualValues(t, 2, body["count"])
	status, body = api.do(http.MethodPost, "/projects/", "token-alice", map[string]any{"name": "Artemis"})
	require.Equal(t, http.StatusCreated, status, "create after migration: %s", body)
	assert.Equal(t, 1, app.count(`SELECT count(*) FROM projects WHERE name = 'Artemis' AND created_at IS NOT NULL`))
}

// requireJourneyPostgres creates a database for the journey on the
// PostgreSQL configured through FORGE_TEST_DATABASE_URL or POSTGRES_*, and
// returns a connection to it with the FORGE_DATABASE_* environment the
// application and the CLI connect with.
func requireJourneyPostgres(t *testing.T, ctx context.Context) (*sql.DB, map[string]string) {
	t.Helper()
	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		DBName:    fmt.Sprintf("e2e_consumer_%d", time.Now().UnixNano()),
	}
	database, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	if err != nil {
		if os.Getenv("RUN_POSTGRES_TESTS") == "1" || os.Getenv("FORGE_REQUIRE_DB") == "1" {
			t.Fatalf("PostgreSQL is required for the release gate but unavailable: %v", err)
		}
		t.Skipf("PostgreSQL unavailable (set RUN_POSTGRES_TESTS=1 to fail instead): %v", err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Logf("drop %s: %v", opts.DBName, err)
		}
	})

	u, err := url.Parse(dsn)
	require.NoError(t, err)
	password, _ := u.User.Password()
	sslmode := u.Query().Get("sslmode")
	if sslmode == "" {
		sslmode = "disable"
	}
	return database, map[string]string{
		"FORGE_DATABASE_DRIVER":   "postgres",
		"FORGE_DATABASE_HOST":     u.Hostname(),
		"FORGE_DATABASE_PORT":     u.Port(),
		"FORGE_DATABASE_USER":     u.User.Username(),
		"FORGE_DATABASE_PASSWORD": password,
		"FORGE_DATABASE_NAME":     strings.TrimPrefix(u.Path, "/"),
		"FORGE_DATABASE_SSLMODE":  sslmode,
	}
}

// registerAppInMain adds the app's routes to the scaffolded server.
func registerAppInMain(t *testing.T, mainPath string) {
	t.Helper()
	content, err := os.ReadFile(mainPath)
	require.NoError(t, err)
	src := string(content)
	for old, replacement := range map[string]string{
		"\t\"github.com/forgego/forge/server\"\n)": "\t\"github.com/forgego/forge/server\"\n\n\t\"tracker/app/work\"\n)",
		"\t\t// Register admin routes\n":           "\t\twork.Register(router, database)\n\n\t\t// Register admin routes\n",
	} {
		require.Contains(t, src, old, "scaffolded main.go changed; update the journey")
		src = strings.Replace(src, old, replacement, 1)
	}
	require.NoError(t, os.WriteFile(mainPath, []byte(src), 0o644))
}

type journeyApp struct {
	t   *testing.T
	ctx context.Context
	dir string
	env map[string]string
	db  *sql.DB
}

func (a *journeyApp) forge(args ...string) string {
	a.t.Helper()
	return a.forgeIn(a.dir, args...)
}

func (a *journeyApp) forgeIn(dir string, args ...string) string {
	a.t.Helper()
	out, _, err := testhelpers.RunCLI(a.ctx, dir, a.env, args, 90*time.Second)
	require.NoError(a.t, err, "forge %s:\n%s", strings.Join(args, " "), out)
	return out
}

func (a *journeyApp) goCmd(args ...string) {
	a.t.Helper()
	cmd := exec.CommandContext(a.ctx, "go", args...)
	cmd.Dir = a.dir
	out, err := cmd.CombinedOutput()
	require.NoError(a.t, err, "go %s:\n%s", strings.Join(args, " "), out)
}

// writeModels writes the models, with extraTaskField added to Task.
func (a *journeyApp) writeModels(extraTaskField string) {
	a.t.Helper()
	src := strings.Replace(trackerModels, trackerExtraTaskField, extraTaskField, 1)
	testhelpers.WriteFileString(a.t, filepath.Join(a.dir, "app", "work", "models.go"), src)
}

// generate runs the code generator and proves the result compiles and vets.
func (a *journeyApp) generate() {
	a.t.Helper()
	a.forge("generate", "--api", "--strict", "--models", "./app/work", "--output", "./app/work")
	a.goCmd("mod", "tidy")
	a.goCmd("build", "./...")
	a.goCmd("vet", "./...")
	a.goCmd("build", "-o", "bin/server", "./cmd/server")
	a.goCmd("build", "-o", "bin/txcheck", "./cmd/txcheck")
}

func (a *journeyApp) migrationFiles() []string {
	a.t.Helper()
	files, err := filepath.Glob(filepath.Join(a.dir, "migrations", "*.sql"))
	require.NoError(a.t, err)
	return files
}

func (a *journeyApp) environ() []string {
	env := os.Environ()
	for key, value := range a.env {
		env = append(env, key+"="+value)
	}
	return env
}

// startServer starts the built server and returns a function that stops it.
func (a *journeyApp) startServer(port string) func() {
	a.t.Helper()
	cmd := exec.Command(filepath.Join(a.dir, "bin", "server"))
	cmd.Dir = a.dir
	cmd.Env = a.environ()
	var logs bytes.Buffer
	cmd.Stdout = &logs
	cmd.Stderr = &logs
	require.NoError(a.t, cmd.Start())

	stopped := false
	stop := func() {
		if stopped {
			return
		}
		stopped = true
		shutdownServer(a.t, cmd, nil)
		if a.t.Failed() {
			a.t.Logf("server logs:\n%s", logs.String())
		}
	}
	a.t.Cleanup(stop)
	waitForHealthy(a.t, port)
	return stop
}

// txcheck runs the transaction command and returns its output and exit code.
func (a *journeyApp) txcheck(projectName string, ownerID int64) (string, int) {
	a.t.Helper()
	cmd := exec.CommandContext(a.ctx, filepath.Join(a.dir, "bin", "txcheck"), projectName, fmt.Sprint(ownerID))
	cmd.Dir = a.dir
	cmd.Env = a.environ()
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(out), exitErr.ExitCode()
	}
	require.NoError(a.t, err, "txcheck: %s", out)
	return string(out), 0
}

func (a *journeyApp) count(query string, args ...any) int {
	a.t.Helper()
	var n int
	require.NoError(a.t, a.db.QueryRowContext(a.ctx, query, args...).Scan(&n), query)
	return n
}

type projectRow struct {
	name, description string
	createdAt         time.Time
}

func (a *journeyApp) projectRow(id int64) projectRow {
	a.t.Helper()
	var row projectRow
	var description sql.NullString
	require.NoError(a.t, a.db.QueryRowContext(a.ctx,
		`SELECT name, description, created_at FROM projects WHERE id = $1`, id).
		Scan(&row.name, &description, &row.createdAt))
	row.description = description.String
	return row
}

func (a *journeyApp) taskTitle(id int64) string {
	a.t.Helper()
	var title string
	require.NoError(a.t, a.db.QueryRowContext(a.ctx, `SELECT title FROM tasks WHERE id = $1`, id).Scan(&title))
	return title
}

type journeyClient struct {
	t    *testing.T
	base string
}

func (c *journeyClient) do(method, path, token string, body map[string]any) (int, map[string]any) {
	c.t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		require.NoError(c.t, err)
	}
	return c.doRaw(method, path, token, payload)
}

// doRaw sends payload as the JSON body and decodes a JSON object response.
func (c *journeyClient) doRaw(method, path, token string, payload []byte) (int, map[string]any) {
	c.t.Helper()
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequest(method, c.base+path, reader)
	require.NoError(c.t, err)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	require.NoError(c.t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(c.t, err)
	decoded := map[string]any{}
	if len(bytes.TrimSpace(raw)) > 0 {
		require.NoError(c.t, json.Unmarshal(raw, &decoded), "%s %s returned %d: %s", method, path, resp.StatusCode, raw)
	}
	return resp.StatusCode, decoded
}

func jsonID(t *testing.T, body map[string]any) int64 {
	t.Helper()
	id, ok := body["id"].(float64)
	require.True(t, ok, "response has no numeric id: %v", body)
	return int64(id)
}
