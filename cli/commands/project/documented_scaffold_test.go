package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
	"github.com/spf13/viper"
)

func newDocumentedProject(t *testing.T, docker bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sample")
	cmd := NewNewCommand().Definition()
	for name, value := range map[string]string{"path": path, "template": "simple", "database": "sqlite", "docker": map[bool]string{true: "true", false: "false"}[docker]} {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := NewNewCommand().Execute(&core.Context{Cmd: cmd}, []string{"sample"}); err != nil {
		t.Fatal(err)
	}
	return path
}

// Explicit flags create a SQLite simple project without Docker files.
func TestNewSQLiteSimpleProjectWithoutDocker(t *testing.T) {
	path := newDocumentedProject(t, false)
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(path, "Dockerfile")); !os.IsNotExist(err) {
		t.Fatalf("unexpected Dockerfile: %v", err)
	}
	v := viper.New()
	v.SetConfigFile(filepath.Join(path, "config", "config.yaml"))
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if got := v.GetString("database.driver"); got != "sqlite3" {
		t.Fatalf("driver = %q", got)
	}
}

// The example option writes a real Example model into the new app.
func TestAddAppExampleWritesModel(t *testing.T) {
	path := newDocumentedProject(t, false)
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	if err := os.Chdir(path); err != nil {
		t.Fatal(err)
	}
	cmd := NewAddAppCommand().Definition()
	if err := cmd.Flags().Set("example", "true"); err != nil {
		t.Fatal(err)
	}
	if err := NewAddAppCommand().Execute(&core.Context{Cmd: cmd}, []string{"blog"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(path, "app", "blog", "models.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "type Example struct") || !strings.Contains(string(data), "func (Example) Fields()") {
		t.Fatalf("missing example model in %s", data)
	}
}

// The scaffolded server uses the documented default listener.
func TestNewProjectDefaultsToLocalhostPort8000(t *testing.T) {
	path := newDocumentedProject(t, false)
	v := viper.New()
	v.SetConfigFile(filepath.Join(path, "config", "config.yaml"))
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if got := v.GetString("server.host"); got != "localhost" {
		t.Fatalf("host = %q", got)
	}
	if got := v.GetString("server.port"); got != "8000" {
		t.Fatalf("port = %q", got)
	}
}

// The Docker option writes both entry point files.
func TestNewDockerProjectWritesDockerFiles(t *testing.T) {
	path := newDocumentedProject(t, true)
	for _, name := range []string{"Dockerfile", "compose.yaml"} {
		if _, err := os.Stat(filepath.Join(path, name)); err != nil {
			t.Fatal(err)
		}
	}
}
