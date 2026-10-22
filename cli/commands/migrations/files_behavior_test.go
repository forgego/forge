package migrations

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
)

func captureMigrationOutput(t *testing.T, run func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	runErr := run()
	w.Close()
	os.Stdout = old
	var out bytes.Buffer
	_, err = out.ReadFrom(r)
	r.Close()
	if err != nil {
		t.Fatal(err)
	}
	return out.String(), runErr
}
func createMigrationCommand(t *testing.T, path string, flags []string, name string) (string, error) {
	t.Helper()
	c := NewMakeMigrationsCommand()
	def := c.Definition()
	if err := def.ParseFlags(append([]string{"--path", path}, flags...)); err != nil {
		t.Fatal(err)
	}
	var direct bytes.Buffer
	def.SetOut(&direct)
	stdout, err := captureMigrationOutput(t, func() error {
		return c.Execute(&core.Context{Cmd: def, Config: migrationConfig(filepath.Join(t.TempDir(), "app.db"))}, []string{name})
	})
	return direct.String() + stdout, err
}

// makemigrations writes an empty pair to the chosen directory.
func TestMakeMigrationsWritesEmptyPairToChosenDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "chosen")
	_, err := createMigrationCommand(t, dir, []string{"--empty"}, "blank")
	if err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"up", "down"} {
		data, err := os.ReadFile(filepath.Join(dir, "000001_blank."+suffix+".sql"))
		if err != nil {
			t.Fatal(err)
		}
		if len(data) != 0 {
			t.Fatalf("%s has content: %q", suffix, data)
		}
	}
}

// --auto reads the selected models and generates their table SQL.
func TestMakeMigrationsAutoGeneratesSQLFromChosenModels(t *testing.T) {
	root := t.TempDir()
	models := filepath.Join(root, "custom_models")
	migrations := filepath.Join(root, "chosen")
	if err := os.Mkdir(models, 0700); err != nil {
		t.Fatal(err)
	}
	source := `package models
import "github.com/forgego/forge/schema"
type Widget struct { schema.BaseSchema; ID int64 }
func (Widget) Fields() []schema.Field { return []schema.Field{schema.Int64Field("id",schema.Primary(),schema.AutoIncrement()),schema.StringField("name")} }
func (Widget) Meta() schema.Meta { return schema.Meta{TableName:"widgets"} }
`
	if err := os.WriteFile(filepath.Join(models, "models.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := createMigrationCommand(t, migrations, []string{"--models", models, "--auto"}, "initial")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(migrations, "000001_initial.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(string(data)), "create table if not exists widgets") {
		t.Fatalf("generated SQL: %q", data)
	}
}

// --merge writes a template for a detected version gap.
func TestMakeMigrationsMergeCreatesTemplateForVersionGap(t *testing.T) {
	dir := t.TempDir()
	writeMigrationPair(t, dir, "000001", "first")
	writeMigrationPair(t, dir, "000003", "third")
	_, err := createMigrationCommand(t, dir, []string{"--merge"}, "resolve")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "000004_merge_000001_000003.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "000001") || !strings.Contains(string(data), "000003") {
		t.Fatalf("merge file: %q", data)
	}
}

// lint reports issues from the selected migration directory.
func TestLintChecksFilesInChosenMigrationDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "000001_bad.up.sql"), []byte("CREATE TABLE widgets (id INT);"), 0600); err != nil {
		t.Fatal(err)
	}
	c := NewLintCommand()
	def := c.Definition()
	if err := def.ParseFlags([]string{"--path", dir}); err != nil {
		t.Fatal(err)
	}
	output, err := captureMigrationOutput(t, func() error { return c.Execute(&core.Context{Cmd: def}, nil) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "000001_bad.up.sql") || !strings.Contains(output, "Summary:") {
		t.Fatalf("lint output: %s", output)
	}
}

// verbose lint includes SQL parse errors from the selected files.
func TestLintVerboseIncludesParseErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "000001_bad.up.sql"), []byte("CREATE TABLE widgets ("), 0600); err != nil {
		t.Fatal(err)
	}
	c := NewLintCommand()
	def := c.Definition()
	if err := def.ParseFlags([]string{"--path", dir, "--verbose"}); err != nil {
		t.Fatal(err)
	}
	output, _ := captureMigrationOutput(t, func() error { return c.Execute(&core.Context{Cmd: def}, nil) })
	if !strings.Contains(output, "Parse errors") {
		t.Fatalf("verbose parse output: %s", output)
	}
}
