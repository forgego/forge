package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/cli/templates"
	codegen "github.com/forgego/forge/codegen"
	"github.com/stretchr/testify/require"
)

// runProjectCommand runs a project command with parsed flags, as the CLI does.
func runProjectCommand(t *testing.T, cmd interface {
	Execute(*core.Context, []string) error
}, definition func() *core.Context, flags []string, args []string) {
	t.Helper()
	ctx := definition()
	require.NoError(t, ctx.Cmd.ParseFlags(flags))
	require.NoError(t, cmd.Execute(ctx, args))
}

// TestAddAPIScaffoldCompilesAndRegisters follows `forge add app --example`,
// `forge generate` and `forge add api`: the scaffolded viewset must compile
// and pass Router.Register's configuration check. The api.go.tmpl template
// gets the same treatment in a second app.
func TestAddAPIScaffoldCompilesAndRegisters(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a scaffolded project")
	}
	forgePath, err := filepath.Abs("../../..")
	require.NoError(t, err)
	projectPath := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(projectPath, "go.mod"),
		[]byte("module example.com/addapi\n\ngo 1.23\n\nrequire github.com/forgego/forge v0.0.0\n\nreplace github.com/forgego/forge => "+forgePath+"\n"),
		0o644,
	))

	originalWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	require.NoError(t, os.Chdir(projectPath))

	addApp := NewAddAppCommand()
	addAPI := NewAddAPICommand()
	for _, app := range []string{"blog", "catalog"} {
		runProjectCommand(t, addApp, func() *core.Context {
			return &core.Context{Cmd: addApp.Definition()}
		}, []string{"--example"}, []string{app})
		appPath := filepath.Join(projectPath, "app", app)
		require.NoError(t, codegen.NewGenerator(appPath, appPath).Generate())
	}

	// forge add api appends a viewset to app/blog/api.go.
	runProjectCommand(t, addAPI, func() *core.Context {
		return &core.Context{Cmd: addAPI.Definition()}
	}, []string{"--app", "blog", "--model", "Example"}, []string{"examples"})

	// The embedded api.go.tmpl renders a viewset into app/catalog.
	rendered, err := templates.RenderTemplate("api.go.tmpl", templates.TemplateData{
		AppName: "catalog", ModelName: "Example", ResourceName: "examples",
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "app", "catalog", "api_examples.go"), rendered, 0o644))

	for app, register := range map[string]string{"blog": "RegisterExamplesAPI", "catalog": "RegisterexamplesAPI"} {
		require.NoError(t, os.WriteFile(filepath.Join(projectPath, "app", app, "api_scaffold_test.go"), []byte(`package `+app+`

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/forgego/forge/server"
)

func TestScaffoldedAPIRegisters(t *testing.T) {
	router := server.NewRouter()
	`+register+`(router)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/examples/", strings.NewReader(`+"`"+`{"name":"x","nmae":"typo"}`+"`"+`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "nmae") {
		t.Fatalf("unknown key: got %d %s", rec.Code, rec.Body.String())
	}
}
`), 0o644))
	}

	for _, args := range [][]string{{"vet", "-mod=mod", "./..."}, {"test", "-mod=mod", "./..."}} {
		cmd := exec.Command("go", args...)
		cmd.Dir = projectPath
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "go %v in the scaffolded project failed:\n%s", args, output)
	}
}
