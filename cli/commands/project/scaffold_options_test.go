package project

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

func scaffoldProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/sample\n\ngo 1.26\n"), 0600); err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func scaffoldRun(t *testing.T, c interface {
	Definition() *cobra.Command
	Execute(*core.Context, []string) error
}, flags []string, args ...string) string {
	t.Helper()
	cmd := c.Definition()
	if err := cmd.ParseFlags(flags); err != nil {
		t.Fatal(err)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	runErr := c.Execute(&core.Context{Cmd: cmd}, args)
	w.Close()
	os.Stdout = old
	out, readErr := io.ReadAll(r)
	r.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	if runErr != nil {
		t.Fatalf("command failed: %v; output: %s", runErr, out)
	}
	return string(out)
}

func scaffoldApp(t *testing.T) string {
	t.Helper()
	root := scaffoldProject(t)
	scaffoldRun(t, NewAddAppCommand(), nil, "blog")
	return root
}

func readScaffold(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestAddAppCreatesSourceFiles(t *testing.T) {
	root := scaffoldApp(t)
	for _, name := range []string{"models.go", "admin.go", "api.go"} {
		if _, err := os.Stat(filepath.Join(root, "app", "blog", name)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAddAppExampleIncludesModel(t *testing.T) {
	root := scaffoldProject(t)
	scaffoldRun(t, NewAddAppCommand(), []string{"--example"}, "blog")
	if !strings.Contains(readScaffold(t, filepath.Join(root, "app", "blog", "models.go")), "type Example struct") {
		t.Fatal("example model missing")
	}
}

func TestAddGroupsScaffoldingCommands(t *testing.T) {
	group := NewAddGroup()
	if group.Definition().Use != "add" {
		t.Fatal("wrong group")
	}
	want := map[string]bool{"app": false, "model": false, "handler": false, "api": false, "service": false}
	for _, command := range group.Commands() {
		want[strings.Fields(command.Definition().Use)[0]] = true
	}
	for name, found := range want {
		if !found {
			t.Fatalf("missing %s", name)
		}
	}
}

func TestAddAPISelectsApp(t *testing.T) {
	root := scaffoldApp(t)
	scaffoldRun(t, NewAddAPICommand(), []string{"--app", "blog", "--model", "Post"}, "posts")
	if !strings.Contains(readScaffold(t, filepath.Join(root, "app", "blog", "api.go")), "RegisterPostsAPI") {
		t.Fatal("API not written to chosen app")
	}
}

func TestAddAPISelectsModel(t *testing.T) {
	root := scaffoldApp(t)
	scaffoldRun(t, NewAddAPICommand(), []string{"--app", "blog", "--model", "Article"}, "articles")
	if !strings.Contains(readScaffold(t, filepath.Join(root, "app", "blog", "api.go")), "ArticleObjects") {
		t.Fatal("chosen model not used")
	}
}

func TestAddAPISelectsResourcePath(t *testing.T) {
	root := scaffoldApp(t)
	scaffoldRun(t, NewAddAPICommand(), []string{"--app", "blog", "--model", "Post", "--resource", "published-posts"}, "posts")
	if !strings.Contains(readScaffold(t, filepath.Join(root, "app", "blog", "api.go")), `Register("published-posts", viewset)`) {
		t.Fatal("chosen resource missing")
	}
}

func TestAddAPIDryRunPreviewsWithoutWriting(t *testing.T) {
	root := scaffoldApp(t)
	path := filepath.Join(root, "app", "blog", "api.go")
	before := readScaffold(t, path)
	out := scaffoldRun(t, NewAddAPICommand(), []string{"--app", "blog", "--model", "Post", "--dry-run"}, "posts")
	if !strings.Contains(out, "would create API posts") || readScaffold(t, path) != before {
		t.Fatal("dry-run changed API or omitted preview")
	}
}

func TestAddHandlerCreatesHTTPHandler(t *testing.T) {
	root := scaffoldApp(t)
	scaffoldRun(t, NewAddHandlerCommand(), []string{"--app", "blog", "--path", "/posts"}, "ListPosts")
	if !strings.Contains(readScaffold(t, filepath.Join(root, "app", "blog", "handlers.go")), "ListPosts") {
		t.Fatal("handler missing")
	}
}

func TestAddHandlerSelectsApp(t *testing.T) {
	root := scaffoldApp(t)
	scaffoldRun(t, NewAddHandlerCommand(), []string{"--app", "blog", "--path", "/posts"}, "ListPosts")
	if _, err := os.Stat(filepath.Join(root, "app", "blog", "handlers.go")); err != nil {
		t.Fatal(err)
	}
}

func TestAddHandlerSelectsMethod(t *testing.T) {
	root := scaffoldApp(t)
	scaffoldRun(t, NewAddHandlerCommand(), []string{"--app", "blog", "--method", "POST", "--path", "/posts"}, "CreatePost")
	if !strings.Contains(readScaffold(t, filepath.Join(root, "app", "blog", "handlers.go")), "POST") {
		t.Fatal("method missing")
	}
}

func TestAddHandlerSelectsPath(t *testing.T) {
	root := scaffoldApp(t)
	scaffoldRun(t, NewAddHandlerCommand(), []string{"--app", "blog", "--path", "/custom-posts"}, "ListPosts")
	if !strings.Contains(readScaffold(t, filepath.Join(root, "app", "blog", "handlers.go")), "/custom-posts") {
		t.Fatal("path missing")
	}
}

func TestAddHandlerDryRunPreviewsWithoutWriting(t *testing.T) {
	root := scaffoldApp(t)
	out := scaffoldRun(t, NewAddHandlerCommand(), []string{"--app", "blog", "--path", "/posts", "--dry-run"}, "ListPosts")
	if !strings.Contains(out, "would create handler") {
		t.Fatal(out)
	}
	if _, err := os.Stat(filepath.Join(root, "app", "blog", "handlers.go")); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote handler: %v", err)
	}
}

func TestAddServiceCreatesService(t *testing.T) {
	root := scaffoldApp(t)
	scaffoldRun(t, NewAddServiceCommand(), []string{"--app", "blog"}, "PostService")
	if !strings.Contains(readScaffold(t, filepath.Join(root, "app", "blog", "services.go")), "type PostService interface") {
		t.Fatal("service missing")
	}
}

func TestAddServiceSelectsApp(t *testing.T) {
	root := scaffoldApp(t)
	scaffoldRun(t, NewAddServiceCommand(), []string{"--app", "blog"}, "PostService")
	if _, err := os.Stat(filepath.Join(root, "app", "blog", "services.go")); err != nil {
		t.Fatal(err)
	}
}

func TestAddServiceSelectsDomain(t *testing.T) {
	root := scaffoldApp(t)
	scaffoldRun(t, NewAddServiceCommand(), []string{"--domain", "billing"}, "BillingService")
	if !strings.Contains(readScaffold(t, filepath.Join(root, "domain", "billing", "service.go")), "package billing") {
		t.Fatal("domain service missing")
	}
}

func TestAddServiceInjectAddsDependencyPlaceholder(t *testing.T) { // wiring is not implemented
	root := scaffoldApp(t)
	if !strings.Contains(NewAddServiceCommand().Definition().Flags().Lookup("inject").Usage, "placeholder") {
		t.Fatal("inject flag promises automatic wiring")
	}
	scaffoldRun(t, NewAddServiceCommand(), []string{"--app", "blog", "--inject"}, "PostService")
	if !strings.Contains(readScaffold(t, filepath.Join(root, "app", "blog", "services.go")), "Add dependencies and wire them manually here") {
		t.Fatal("dependency placeholder missing")
	}
}

func TestAddServiceDryRunPreviewsWithoutWriting(t *testing.T) {
	root := scaffoldApp(t)
	out := scaffoldRun(t, NewAddServiceCommand(), []string{"--app", "blog", "--dry-run"}, "PostService")
	if !strings.Contains(out, "would create service") {
		t.Fatal(out)
	}
	if _, err := os.Stat(filepath.Join(root, "app", "blog", "services.go")); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote service: %v", err)
	}
}

func TestNewUsesSelectedProjectLocation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chosen")
	scaffoldRun(t, NewNewCommand(), []string{"--path", path, "--template", "simple", "--database", "sqlite", "--docker=false"}, "sample")
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err != nil {
		t.Fatal(err)
	}
}

func TestAddModelSelectsAppInPreview(t *testing.T) {
	scaffoldApp(t)
	out := scaffoldRun(t, NewAddModelCommand(), []string{"--app", "blog", "--dry-run"}, "Post")
	if !strings.Contains(out, "model Post in app blog") {
		t.Fatal(out)
	}
}

func TestAddModelSelectsTableNameInPreview(t *testing.T) {
	scaffoldApp(t)
	out := scaffoldRun(t, NewAddModelCommand(), []string{"--app", "blog", "--table", "articles", "--dry-run"}, "Post")
	if !strings.Contains(out, "Table name: articles") {
		t.Fatal(out)
	}
}

func TestAddModelDryRunPreviewsWithoutWriting(t *testing.T) {
	root := scaffoldApp(t)
	path := filepath.Join(root, "app", "blog", "models.go")
	before := readScaffold(t, path)
	out := scaffoldRun(t, NewAddModelCommand(), []string{"--app", "blog", "--dry-run"}, "Post")
	if !strings.Contains(out, "would create model") || readScaffold(t, path) != before {
		t.Fatal("preview missing or model written")
	}
}
