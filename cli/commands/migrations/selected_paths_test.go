package migrations

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
	forgeerrors "github.com/forgego/forge/errors"
)

func TestRollbackUsesSelectedMigrationDirectory(t *testing.T) {
	root := t.TempDir()
	chosen := filepath.Join(root, "chosen")
	if err := os.Mkdir(chosen, 0700); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(root, "app.db")
	writeMigrationPair(t, chosen, "000001", "widgets")
	if err := executeMigration(t, NewUpCommand(), dbPath, []string{"--path", chosen}); err != nil {
		t.Fatal(err)
	}
	if !sqliteTableExists(t, dbPath, "widgets") {
		t.Fatal("up did not apply selected migration")
	}
	if err := executeMigration(t, NewRollbackCommand(), dbPath, []string{"--path", chosen}); err != nil {
		t.Fatal(err)
	}
	if sqliteTableExists(t, dbPath, "widgets") {
		t.Fatal("rollback did not reverse selected migration")
	}
}

func TestShowReadsSelectedMigrationAndRawSQL(t *testing.T) {
	root := t.TempDir()
	chosen := filepath.Join(root, "chosen")
	if err := os.Mkdir(chosen, 0700); err != nil {
		t.Fatal(err)
	}
	writeMigrationPair(t, chosen, "000001", "widgets")
	command := NewShowCommand()
	def := command.Definition()
	if err := def.ParseFlags([]string{"--path", chosen, "--sql"}); err != nil {
		t.Fatal(err)
	}
	out, err := captureMigrationOutput(t, func() error { return command.Execute(&core.Context{Cmd: def}, []string{"000001"}) })
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Up Migration", "CREATE TABLE widgets", "Down Migration", "DROP TABLE widgets"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %s", want, out)
		}
	}
}

func TestShowPlansSelectedModels(t *testing.T) {
	root := t.TempDir()
	models := filepath.Join(root, "chosen_models")
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
	command := NewShowCommand()
	def := command.Definition()
	if err := def.ParseFlags([]string{"--models", models, "--path", filepath.Join(root, "migrations")}); err != nil {
		t.Fatal(err)
	}
	out, err := captureMigrationOutput(t, func() error {
		return command.Execute(&core.Context{Cmd: def, Config: migrationConfig(filepath.Join(root, "app.db"))}, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Migration Plan Preview") || !strings.Contains(strings.ToLower(out), "widgets") {
		t.Fatalf("plan: %s", out)
	}
}

func TestShowVerboseIncludesMigrationParseErrors(t *testing.T) {
	root := t.TempDir()
	models := filepath.Join(root, "models")
	migrations := filepath.Join(root, "migrations")
	if err := os.Mkdir(models, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(migrations, 0700); err != nil {
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
	if err := os.WriteFile(filepath.Join(migrations, "000001_broken.up.sql"), []byte("CREATE TABLE widgets ("), 0600); err != nil {
		t.Fatal(err)
	}
	command := NewShowCommand()
	def := command.Definition()
	if err := def.ParseFlags([]string{"--models", models, "--path", migrations, "--verbose"}); err != nil {
		t.Fatal(err)
	}
	out, runErr := captureMigrationOutput(t, func() error {
		return command.Execute(&core.Context{Cmd: def, Config: migrationConfig(filepath.Join(root, "app.db"))}, nil)
	})
	if runErr == nil || !strings.Contains(out, "Parse warnings:") || !strings.Contains(out, "000001_broken.up.sql") {
		t.Fatalf("parse diagnostic missing: output=%s err=%v", out, runErr)
	}
}

func TestSquashReportsUnavailableForSelectedDirectory(t *testing.T) {
	root := t.TempDir()
	chosen := filepath.Join(root, "chosen")
	if err := os.Mkdir(chosen, 0700); err != nil {
		t.Fatal(err)
	}
	writeMigrationPair(t, chosen, "000001", "widgets")
	before, err := os.ReadDir(chosen)
	if err != nil {
		t.Fatal(err)
	}
	command := NewSquashCommand()
	def := command.Definition()
	if err := def.Flags().Set("path", chosen); err != nil {
		t.Fatal(err)
	}
	err = command.Execute(&core.Context{Cmd: def}, []string{"000001", "000002", "combined"})
	if !forgeerrors.IsNotImplemented(err) {
		t.Fatalf("expected unavailable squash, got %v", err)
	}
	if !strings.Contains(def.Flags().Lookup("path").Usage, "not implemented") {
		t.Fatal("path flag promises an unavailable selection")
	}
	after, err := os.ReadDir(chosen)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("squash changed selected migration directory: before=%d after=%d", len(before), len(after))
	}
}

func TestStatusListsSelectedMigrationDirectory(t *testing.T) {
	root := t.TempDir()
	chosen := filepath.Join(root, "chosen")
	if err := os.Mkdir(chosen, 0700); err != nil {
		t.Fatal(err)
	}
	writeMigrationPair(t, chosen, "000001", "widgets")
	cmd := NewStatusCommand().Definition()
	cmd.SetContext(context.Background())
	if err := cmd.Flags().Set("path", chosen); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	cmd.SetOut(&out)
	if err := NewStatusCommand().Execute(&core.Context{Cmd: cmd, Config: migrationConfig(filepath.Join(root, "app.db"))}, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "000001_widgets.up.sql") {
		t.Fatalf("status: %s", out.String())
	}
}

func TestUpAppliesSelectedMigrationDirectory(t *testing.T) {
	root := t.TempDir()
	chosen := filepath.Join(root, "chosen")
	if err := os.Mkdir(chosen, 0700); err != nil {
		t.Fatal(err)
	}
	writeMigrationPair(t, chosen, "000001", "widgets")
	dbPath := filepath.Join(root, "app.db")
	if err := executeMigration(t, NewUpCommand(), dbPath, []string{"--path", chosen}); err != nil {
		t.Fatal(err)
	}
	if !sqliteTableExists(t, dbPath, "widgets") {
		t.Fatal("selected migration not applied")
	}
}

func TestUpDryRunPreviewsSelectedSQLWithoutApplying(t *testing.T) {
	root := t.TempDir()
	chosen := filepath.Join(root, "chosen")
	if err := os.Mkdir(chosen, 0700); err != nil {
		t.Fatal(err)
	}
	writeMigrationPair(t, chosen, "000001", "widgets")
	dbPath := filepath.Join(root, "app.db")
	cmd := NewUpCommand().Definition()
	if err := cmd.ParseFlags([]string{"--path", chosen, "--dry-run"}); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	cmd.SetOut(&out)
	if err := NewUpCommand().Execute(&core.Context{Cmd: cmd, Config: migrationConfig(dbPath)}, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "CREATE TABLE widgets") || sqliteTableExists(t, dbPath, "widgets") {
		t.Fatalf("dry-run: %s", out.String())
	}
}
