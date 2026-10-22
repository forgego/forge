package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/forgego/forge/tests/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type useCaseModel struct {
	Name, Route                string
	Owned, ReadOnly, StaffOnly bool
	Primary                    bool
}

type useCaseSpec struct {
	Name, Primary, UniqueAssign string
	Models                      []useCaseModel
	PrimaryInput                map[string]any
	Workflow                    func(*testing.T, *journeyApp, *journeyClient, int64)
}

func consumerCases() []useCaseSpec {
	return []useCaseSpec{
		{Name: "commerce", Primary: "Product", UniqueAssign: `item.Sku = "rollback-sku"; item.PriceCents = 1999`,
			Models:       []useCaseModel{{Name: "Product", Route: "products"}, {Name: "Customer", Route: "customers"}, {Name: "Order", Route: "orders", Owned: true}, {Name: "OrderLine", Route: "order-lines"}, {Name: "StockMovement", Route: "stock-movements"}},
			PrimaryInput: map[string]any{"name": "Precision keyboard", "sku": "KB-001", "price_cents": 1999, "stock": 4, "active": false}, Workflow: commerceJourney},
		{Name: "crm", Primary: "Organization", UniqueAssign: `item.Domain = "rollback.example"`,
			Models:       []useCaseModel{{Name: "Organization", Route: "organizations"}, {Name: "Contact", Route: "contacts"}, {Name: "Ticket", Route: "tickets", Owned: true}, {Name: "Comment", Route: "comments"}},
			PrimaryInput: map[string]any{"name": "Acme support", "domain": "acme.example"}, Workflow: crmJourney},
		{Name: "booking", Primary: "Resource", UniqueAssign: `item.Code = "rollback-resource"`,
			Models:       []useCaseModel{{Name: "Resource", Route: "resources"}, {Name: "Slot", Route: "slots"}, {Name: "Booking", Route: "bookings", Owned: true}},
			PrimaryInput: map[string]any{"name": "Meeting room", "code": "ROOM-A", "active": false}, Workflow: bookingJourney},
		{Name: "cms", Primary: "Author", UniqueAssign: `item.Email = "rollback@example.com"`,
			Models:       []useCaseModel{{Name: "Author", Route: "authors"}, {Name: "Post", Route: "posts", Owned: true}, {Name: "Tag", Route: "tags"}, {Name: "PostTag", Route: "post-tags"}, {Name: "Revision", Route: "revisions"}},
			PrimaryInput: map[string]any{"name": "Editor", "email": "editor@example.com"}, Workflow: cmsJourney},
		{Name: "expenses", Primary: "Employee", UniqueAssign: `item.Email = "rollback@example.com"`,
			Models:       []useCaseModel{{Name: "Employee", Route: "employees"}, {Name: "ExpenseReport", Route: "expense-reports", Owned: true}, {Name: "Expense", Route: "expenses"}, {Name: "Approval", Route: "approvals", StaffOnly: true}},
			PrimaryInput: map[string]any{"name": "Engineer", "email": "engineer@example.com"}, Workflow: expensesJourney},
	}
}

// TestGeneratedConsumerUseCases runs five separate projects, processes and
// PostgreSQL databases. Artifacts are opt-in; ordinary CI cleans up after itself.
func TestGeneratedConsumerUseCases(t *testing.T) {
	for _, spec := range consumerCases() {
		t.Run(spec.Name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			retain := os.Getenv("FORGE_CONSUMER_ARTIFACTS") != ""
			database, env := journeyPostgres(t, ctx, retain)
			base := t.TempDir()
			if retain {
				base = os.Getenv("FORGE_CONSUMER_ARTIFACTS")
				require.NoError(t, os.MkdirAll(base, 0o700))
				var err error
				base, err = os.MkdirTemp(base, spec.Name+"-")
				require.NoError(t, err)
			}
			project := filepath.Join(base, spec.Name)
			app := &journeyApp{t: t, ctx: ctx, dir: project, env: env, db: database}
			// A broken enclosing /tmp/.git must not make temporary consumer
			// builds fail while Go tries to stamp unrelated VCS metadata.
			env["GOFLAGS"] = strings.TrimSpace(os.Getenv("GOFLAGS") + " -buildvcs=false")
			env["FORGE_ADMIN_USERNAME"], env["FORGE_ADMIN_PASSWORD"] = "operator", "local-acceptance-only"
			t.Logf("Generated project: %s", project)
			app.forgeIn(base, "new", spec.Name, "--path", project, "--template", "simple", "--database", "postgres", "--docker=false")
			require.NoError(t, patchGeneratedProject(project))
			app.forge("add", "app", "work")
			models, err := os.ReadFile(filepath.Join("testdata", "usecases", spec.Name, "models.go"))
			require.NoError(t, err)
			testhelpers.WriteFileString(t, filepath.Join(project, "app", "work", "models.go"), string(models))
			spec.Models[0].Primary = true
			spec.Models = append([]useCaseModel{{Name: "Member", Route: "members", ReadOnly: true}}, spec.Models...)
			writeConsumerTemplate(t, filepath.Join(project, "app", "work", "routes.go"), useCaseRoutes, spec)
			writeConsumerTemplate(t, filepath.Join(project, "app", "work", "consumer_test.go"), useCaseORMTest, spec)
			registerAppInMainForModule(t, filepath.Join(project, "cmd", "server", "main.go"), spec.Name)
			port := randomPort(t)
			updateServerPort(t, filepath.Join(project, "config", "config.yaml"), port)
			writeConsumerEnvironment(t, project, env)
			app.forge("generate", "--api", "--strict", "--models", "./app/work", "--output", "./app/work")
			app.goCmd("mod", "tidy")
			buildConsumerServer(app)
			app.goCmd("vet", "./...")
			app.forge("makemigrations", "initial", "--auto", "--models", "./app/work")
			app.forge("migrate", "up")
			assert.Contains(t, app.forge("migrate", "status"), "Current Version: 1")
			assert.Contains(t, app.forge("makemigrations", "unchanged", "--auto", "--models", "./app/work"), "No changes detected")
			assert.Len(t, app.migrationFiles(), 2)
			for _, member := range []struct {
				name  string
				staff bool
			}{{"alice", false}, {"bob", false}, {"staff", true}} {
				digest := sha256.Sum256([]byte("token-" + member.name))
				_, err := database.ExecContext(ctx, `INSERT INTO members (name,email,api_token,is_staff) VALUES ($1,$2,$3,$4)`, member.name, member.name+"@example.com", hex.EncodeToString(digest[:]), member.staff)
				require.NoError(t, err)
			}
			stop := app.startServer(port)
			client := &journeyClient{t: t, base: "http://127.0.0.1:" + port + "/api/v1"}
			primaryID := commonConsumerJourney(t, app, client, spec)
			spec.Workflow(t, app, client, primaryID)
			app.goCmd("test", "-race", "-count=1", "./app/work")
			checkConsumerThrottling(t, client)
			stop()
			checkConsumerEvolution(t, app, string(models), spec.Models[1].Route)
			if retain {
				manifest, err := json.MarshalIndent(map[string]any{"name": spec.Name, "project": project, "port": port, "route": spec.Models[1].Route, "input": spec.PrimaryInput}, "", "  ")
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(filepath.Join(base, "consumer.json"), manifest, 0o600))
			}
			// Leave retained applications built against the final schema.
			buildConsumerServer(app)
		})
	}
}

func buildConsumerServer(app *journeyApp) {
	args := []string{"build"}
	// Browser acceptance needs the separately built admin bundle. Backend CI
	// deliberately works from a clean checkout without generated frontend assets.
	if os.Getenv("FORGE_CONSUMER_BROWSER") == "1" {
		args = append(args, "-tags", "embed")
	}
	app.goCmd(append(args, "-o", "bin/server", "./cmd/server")...)
}

func writeConsumerTemplate(t *testing.T, path, source string, data any) {
	t.Helper()
	tmpl, err := template.New("consumer").Parse(source)
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, data))
	testhelpers.WriteFileString(t, path, buf.String())
}

func writeConsumerEnvironment(t *testing.T, project string, env map[string]string) {
	t.Helper()
	file, err := os.OpenFile(filepath.Join(project, ".env"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	defer file.Close()
	require.NoError(t, file.Chmod(0o600))
	for key, value := range env {
		_, err = fmt.Fprintf(file, "\n%s=%s\n", key, value)
		require.NoError(t, err)
	}
}

func createConsumer(t *testing.T, client *journeyClient, route string, data map[string]any) int64 {
	t.Helper()
	status, body := client.do(http.MethodPost, "/"+route+"/", "token-alice", data)
	require.Equal(t, http.StatusCreated, status, "%s: %v", route, body)
	return jsonID(t, body)
}

func commonConsumerJourney(t *testing.T, app *journeyApp, client *journeyClient, spec useCaseSpec) int64 {
	t.Helper()
	route := spec.Models[1].Route
	status, _ := client.do(http.MethodGet, "/"+route+"/", "", nil)
	assert.Equal(t, http.StatusUnauthorized, status)
	input := make(map[string]any, len(spec.PrimaryInput)+3)
	for key, value := range spec.PrimaryInput {
		input[key] = value
	}
	input["id"], input["created_at"] = 987654321, "2000-01-01T00:00:00Z"
	id := createConsumer(t, client, route, input)
	assert.NotEqualValues(t, 987654321, id)
	status, body := client.do(http.MethodGet, fmt.Sprintf("/%s/%d", route, id), "token-bob", nil)
	require.Equal(t, http.StatusOK, status, "%v", body)
	assert.NotContains(t, body, "internal_token")
	assert.NotEqual(t, "2000-01-01T00:00:00Z", body["created_at"])
	if active, exists := spec.PrimaryInput["active"]; exists {
		assert.Equal(t, active, body["active"])
	}
	for _, bad := range []map[string]any{{"name": ""}, {"name": strings.Repeat("x", 81)}, {"typo_field": true}, {"internal_token": "injected"}} {
		status, body = client.do(http.MethodPatch, fmt.Sprintf("/%s/%d", route, id), "token-alice", bad)
		assert.Equal(t, http.StatusBadRequest, status, "%v", body)
	}
	assert.Equal(t, 1, app.count("SELECT count(*) FROM "+route))
	status, body = client.do(http.MethodGet, "/"+route+"/?name="+strings.ReplaceAll(spec.PrimaryInput["name"].(string), " ", "%20")+"&ordering=-name&page_size=1", "token-alice", nil)
	require.Equal(t, http.StatusOK, status)
	assert.EqualValues(t, 1, body["count"])
	assert.Len(t, body["results"], 1)
	status, body = client.do(http.MethodGet, "/members/", "token-alice", nil)
	require.Equal(t, http.StatusOK, status)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "api_token")
	assert.NotContains(t, string(raw), "is_staff")
	for _, model := range spec.Models {
		assert.Equal(t, 1, app.count(`SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=$1`, strings.ReplaceAll(model.Route, "-", "_")), model.Name)
	}
	assert.Equal(t, 1, app.count(`SELECT count(*) FROM pg_indexes WHERE indexname=$1`, "idx_"+route+"_name"), "declared Meta index must exist")
	// Successful PATCH and PUT preserve explicit zero values and allow repeated
	// pagination without leaking hidden fields. Use a disposable independent row.
	extra := make(map[string]any, len(spec.PrimaryInput))
	for key, value := range spec.PrimaryInput {
		extra[key] = value
	}
	extra["name"] = "ZZ pagination fixture"
	for _, field := range []string{"sku", "domain", "code", "email"} {
		if value, ok := extra[field].(string); ok {
			extra[field] = "extra-" + value
		}
	}
	extraID := createConsumer(t, client, route, extra)
	detail := fmt.Sprintf("/%s/%d", route, extraID)
	status, body = client.do(http.MethodPatch, detail, "token-alice", map[string]any{"name": "ZZ patched"})
	require.Equal(t, http.StatusOK, status, "%v", body)
	extra["name"] = "ZZ replaced"
	status, body = client.do(http.MethodPut, detail, "token-alice", extra)
	require.Equal(t, http.StatusOK, status, "%v", body)
	assert.Equal(t, "ZZ replaced", body["name"])
	var pages []int64
	for _, page := range []string{"1", "2"} {
		status, body = client.do(http.MethodGet, "/"+route+"/?ordering=name&page_size=1&page="+page, "token-alice", nil)
		require.Equal(t, http.StatusOK, status, "%v", body)
		assert.EqualValues(t, 2, body["count"])
		rows, ok := body["results"].([]any)
		require.True(t, ok)
		require.Len(t, rows, 1)
		pages = append(pages, jsonID(t, rows[0].(map[string]any)))
	}
	assert.ElementsMatch(t, []int64{id, extraID}, pages)
	status, _ = client.do(http.MethodDelete, detail, "token-alice", nil)
	require.Equal(t, http.StatusNoContent, status)
	status, _ = client.do(http.MethodGet, detail, "token-alice", nil)
	assert.Equal(t, http.StatusNotFound, status)
	return id
}

func checkConsumerThrottling(t *testing.T, client *journeyClient) {
	t.Helper()
	for i := 0; i < 2; i++ {
		status, body := client.do(http.MethodGet, "/limited/", "token-alice", nil)
		require.Equal(t, http.StatusOK, status, "%v", body)
	}
	status, _ := client.do(http.MethodGet, "/limited/", "token-alice", nil)
	assert.Equal(t, http.StatusTooManyRequests, status)
	assert.NotEmpty(t, client.lastHeader.Get("Retry-After"))
	status, body := client.do(http.MethodGet, "/limited/", "token-bob", nil)
	assert.Equal(t, http.StatusOK, status, "another user's allowance must remain independent: %v", body)
}

func checkConsumerEvolution(t *testing.T, app *journeyApp, original, table string) {
	t.Helper()
	before := app.count("SELECT count(*) FROM " + table)
	expanded := strings.Replace(original, "// ACCEPTANCE EXTRA FIELD", `schema.StringField("acceptance_note", schema.Default("migrated")),`, 1)
	testhelpers.WriteFileString(t, filepath.Join(app.dir, "app", "work", "models.go"), expanded)
	app.forge("generate", "--api", "--strict", "--models", "./app/work", "--output", "./app/work")
	app.forge("makemigrations", "evolve", "--auto", "--models", "./app/work")
	app.forge("migrate", "up")
	assert.Equal(t, before, app.count("SELECT count(*) FROM "+table+" WHERE acceptance_note = 'migrated'"))
	app.forge("migrate", "recover", "--verify")
	app.forge("migrate", "rollback")
	assert.Equal(t, before, app.count("SELECT count(*) FROM "+table))
	assert.Zero(t, app.count(`SELECT count(*) FROM information_schema.columns WHERE table_name=$1 AND column_name='acceptance_note'`, table))
	app.forge("migrate", "up")
	assert.Equal(t, before, app.count("SELECT count(*) FROM "+table+" WHERE acceptance_note = 'migrated'"))
}
