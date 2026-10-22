package migrations

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
)

// Automatic migration generation writes matching up and down SQL.
func TestMakeMigrationsAutoWritesPairedSQLFiles(t *testing.T) {
	root := t.TempDir()
	models := filepath.Join(root, "models")
	migrations := filepath.Join(root, "migrations")
	if err := os.MkdirAll(models, 0700); err != nil {
		t.Fatal(err)
	}
	source := `package models
import "github.com/forgego/forge/schema"
type Widget struct { schema.BaseSchema; ID int64; Name string }
func (Widget) Fields() []schema.Field { return []schema.Field{schema.Int64Field("id",schema.Primary(),schema.AutoIncrement()),schema.StringField("name")} }
func (Widget) Meta() schema.Meta { return schema.Meta{TableName:"widgets"} }
`
	if err := os.WriteFile(filepath.Join(models, "models.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := NewMakeMigrationsCommand().Definition()
	for name, value := range map[string]string{"path": migrations, "models": models, "auto": "true"} {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := NewMakeMigrationsCommand().Execute(&core.Context{Cmd: cmd, Config: migrationConfig(filepath.Join(root, "app.db"))}, []string{"initial"}); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"up", "down"} {
		path := filepath.Join(migrations, "000001_initial."+suffix+".sql")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(strings.ToLower(string(data)), "widgets") {
			t.Fatalf("%s does not mention widgets: %s", path, data)
		}
	}
}

func TestMakeMigrationsVerboseUsesOrdinaryGenerationOutput(t *testing.T) {
	root := t.TempDir()
	models := filepath.Join(root, "models")
	if err := os.Mkdir(models, 0700); err != nil {
		t.Fatal(err)
	}
	source := `package models
import "github.com/forgego/forge/schema"
type Widget struct { schema.BaseSchema; ID int64 }
func (Widget) Fields() []schema.Field { return []schema.Field{schema.Int64Field("id",schema.Primary(),schema.AutoIncrement())} }
func (Widget) Meta() schema.Meta { return schema.Meta{TableName:"widgets"} }
`
	if err := os.WriteFile(filepath.Join(models, "models.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	for _, verbose := range []bool{false, true} {
		command := NewMakeMigrationsCommand()
		cmd := command.Definition()
		migrations := filepath.Join(root, map[bool]string{false: "ordinary", true: "verbose"}[verbose])
		for name, value := range map[string]string{"path": migrations, "models": models, "auto": "true"} {
			if err := cmd.Flags().Set(name, value); err != nil {
				t.Fatal(err)
			}
		}
		if verbose {
			if err := cmd.Flags().Set("verbose", "true"); err != nil {
				t.Fatal(err)
			}
		}
		var output bytes.Buffer
		cmd.SetOut(&output)
		if err := command.Execute(&core.Context{Cmd: cmd, Config: migrationConfig(filepath.Join(root, "app.db"))}, []string{"initial"}); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.ToLower(output.String()), "warning") {
			t.Fatalf("unexpected diagnostic output: %s", output.String())
		}
		if _, err := os.Stat(filepath.Join(migrations, "000001_initial.up.sql")); err != nil {
			t.Fatal(err)
		}
	}
}
