package migrations

import (
	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/config"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func migrationConfig(dbPath string) *config.Config {
	cfg := &config.Config{Viper: viper.New()}
	cfg.Set("database.driver", "sqlite")
	cfg.Set("database.name", dbPath)
	cfg.Set("server.stores", "memory")
	return cfg
}
func writeMigrationPair(t *testing.T, dir, version, table string) {
	t.Helper()
	for suffix, body := range map[string]string{"up": "CREATE TABLE " + table + " (id INTEGER PRIMARY KEY);", "down": "DROP TABLE " + table + ";"} {
		if err := os.WriteFile(filepath.Join(dir, version+"_"+table+"."+suffix+".sql"), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMigrateUpDryRunPreviewsWithoutApplying(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "app.db")
	writeMigrationPair(t, dir, "000001", "widgets")
	cmd := NewUpCommand().Definition()
	if err := cmd.Flags().Set("path", dir); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatal(err)
	}
	ctx := &core.Context{Cmd: cmd, Config: migrationConfig(dbPath)}
	var output strings.Builder
	cmd.SetOut(&output)
	if err := NewUpCommand().Execute(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "CREATE TABLE widgets") || !strings.Contains(output.String(), "No changes were made") {
		t.Fatalf("preview: %s", output.String())
	}
	if sqliteTableExists(t, dbPath, "widgets") {
		t.Fatal("dry run created the table")
	}
}

func TestMigrateRollbackRevertsLatestMigration(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "app.db")
	writeMigrationPair(t, dir, "000001", "first")
	writeMigrationPair(t, dir, "000002", "second")
	t.Setenv("FORGE_DATABASE_DRIVER", "sqlite")
	t.Setenv("FORGE_DATABASE_NAME", dbPath)
	t.Setenv("FORGE_SERVER_STORES", "memory")
	out, err := runMigrateCommand(t, NewUpCommand().Definition(), "--path", dir)
	if err != nil {
		t.Fatalf("up: %v: %s", err, out)
	}
	cmd := NewRollbackCommand().Definition()
	if err := cmd.Flags().Set("path", dir); err != nil {
		t.Fatal(err)
	}
	if err := NewRollbackCommand().Execute(&core.Context{Cmd: cmd, Config: migrationConfig(dbPath)}, nil); err != nil {
		t.Fatal(err)
	}
	if !sqliteTableExists(t, dbPath, "first") || sqliteTableExists(t, dbPath, "second") {
		t.Fatal("rollback should remove latest table only")
	}
}

func TestMakeMigrationsEmptyCreatesEmptyPair(t *testing.T) {
	dir := t.TempDir()
	cmd := NewMakeMigrationsCommand().Definition()
	if err := cmd.Flags().Set("path", dir); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("empty", "true"); err != nil {
		t.Fatal(err)
	}
	if err := NewMakeMigrationsCommand().Execute(&core.Context{Cmd: cmd}, []string{"blank"}); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"up", "down"} {
		data, err := os.ReadFile(filepath.Join(dir, "000001_blank."+suffix+".sql"))
		if err != nil {
			t.Fatal(err)
		}
		if len(data) != 0 {
			t.Fatalf("%s migration is not empty: %q", suffix, data)
		}
	}
}

func TestMakeMigrationsAutoReportsUnchangedSchema(t *testing.T) {
	models := t.TempDir()
	dir := t.TempDir()
	source := `package models
import "github.com/forgego/forge/schema"
type Widget struct { schema.BaseSchema; ID int64 }
func (Widget) Fields() []schema.Field { return []schema.Field{schema.Int64Field("id",schema.Primary(),schema.AutoIncrement()),schema.StringField("name")} }
func (Widget) Meta() schema.Meta { return schema.Meta{TableName:"widgets"} }
`
	if err := os.WriteFile(filepath.Join(models, "models.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(name string) string {
		cmd := NewMakeMigrationsCommand().Definition()
		for key, value := range map[string]string{"path": dir, "models": models, "auto": "true"} {
			if err := cmd.Flags().Set(key, value); err != nil {
				t.Fatal(err)
			}
		}
		var output strings.Builder
		cmd.SetOut(&output)
		if err := NewMakeMigrationsCommand().Execute(&core.Context{Cmd: cmd, Config: migrationConfig(filepath.Join(dir, "app.db"))}, []string{name}); err != nil {
			t.Fatal(err)
		}
		return output.String()
	}
	run("initial")
	if output := run("check"); !strings.Contains(output, "No changes detected") {
		t.Fatalf("unchanged schema: %s", output)
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("unchanged schema wrote files: %v", files)
	}
}
