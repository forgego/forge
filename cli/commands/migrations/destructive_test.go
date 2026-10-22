package migrations

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
)

func sqliteColumnExists(t *testing.T, path, table, column string) bool {
	t.Helper()
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	rows, err := database.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name, kind string
		var notNull, pk int
		var defaultValue sql.NullString
		if err := rows.Scan(&id, &name, &kind, &notNull, &defaultValue, &pk); err != nil {
			t.Fatal(err)
		}
		if name == column {
			return true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return false
}

func writeSQLPair(t *testing.T, dir, name, up, down string) {
	t.Helper()
	for suffix, sql := range map[string]string{"up": up, "down": down} {
		if err := os.WriteFile(filepath.Join(dir, name+"."+suffix+".sql"), []byte(sql), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func runUpForSafety(t *testing.T, dir, dbPath string, allow bool) (string, error) {
	t.Helper()
	cmd := NewUpCommand().Definition()
	if err := cmd.Flags().Set("path", dir); err != nil {
		t.Fatal(err)
	}
	if allow {
		if err := cmd.Flags().Set("allow-destructive", "true"); err != nil {
			t.Fatal(err)
		}
	}
	var out strings.Builder
	cmd.SetOut(&out)
	err := NewUpCommand().Execute(&core.Context{Cmd: cmd, Config: migrationConfig(dbPath)}, nil)
	return out.String(), err
}

func TestUpRequiresConfirmationBeforeAnyPendingMigration(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "app.db")
	writeSQLPair(t, dir, "000001_initial", "CREATE TABLE widgets (id INTEGER PRIMARY KEY, name TEXT);", "DROP TABLE widgets;")
	if _, err := runUpForSafety(t, dir, dbPath, false); err != nil {
		t.Fatal(err)
	}
	writeSQLPair(t, dir, "000002_additive", "CREATE TABLE extra (id INTEGER PRIMARY KEY);", "DROP TABLE extra;")
	writeSQLPair(t, dir, "000003_remove_name", "ALTER TABLE widgets DROP COLUMN name;", "ALTER TABLE widgets ADD COLUMN name TEXT;")
	out, err := runUpForSafety(t, dir, dbPath, false)
	if err == nil || !strings.Contains(out+err.Error(), "000003_remove_name.up.sql") || !strings.Contains(out+err.Error(), "DropColumn") || !strings.Contains(out+err.Error(), "--allow-destructive") {
		t.Fatalf("expected named destructive error, output=%q err=%v", out, err)
	}
	if sqliteTableExists(t, dbPath, "extra") {
		t.Fatal("earlier pending migration was applied")
	}
	if !sqliteColumnExists(t, dbPath, "widgets", "name") {
		t.Fatal("column was dropped")
	}
	if _, err := runUpForSafety(t, dir, dbPath, true); err != nil {
		t.Fatal(err)
	}
	if !sqliteTableExists(t, dbPath, "extra") || sqliteColumnExists(t, dbPath, "widgets", "name") {
		t.Fatal("allowed migrations did not apply")
	}
}

func TestUpAppliesAdditiveMigrationWithoutConfirmation(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "app.db")
	writeSQLPair(t, dir, "000001_initial", "CREATE TABLE widgets (id INTEGER PRIMARY KEY);", "DROP TABLE widgets;")
	if _, err := runUpForSafety(t, dir, dbPath, false); err != nil {
		t.Fatal(err)
	}
	if !sqliteTableExists(t, dbPath, "widgets") {
		t.Fatal("additive migration was not applied")
	}
}

func TestMakeMigrationsWarnsForGeneratedDropColumn(t *testing.T) {
	models := t.TempDir()
	dir := t.TempDir()
	modelPath := filepath.Join(models, "models.go")
	writeModel := func(field string) {
		t.Helper()
		source := `package models
import "github.com/forgego/forge/schema"
type Widget struct { schema.BaseSchema; ID int64 }
func (Widget) Fields() []schema.Field { return []schema.Field{schema.Int64Field("id",schema.Primary(),schema.AutoIncrement())` + field + `} }
func (Widget) Meta() schema.Meta { return schema.Meta{TableName:"widgets"} }
`
		if err := os.WriteFile(modelPath, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	run := func(name string) string {
		t.Helper()
		cmd := NewMakeMigrationsCommand().Definition()
		for key, value := range map[string]string{"path": dir, "models": models, "auto": "true"} {
			if err := cmd.Flags().Set(key, value); err != nil {
				t.Fatal(err)
			}
		}
		var out strings.Builder
		cmd.SetOut(&out)
		if err := NewMakeMigrationsCommand().Execute(&core.Context{Cmd: cmd, Config: migrationConfig(filepath.Join(dir, "app.db"))}, []string{name}); err != nil {
			t.Fatal(err)
		}
		return out.String()
	}
	writeModel(`,schema.StringField("name")`)
	run("initial")
	writeModel("")
	out := run("remove_name")
	if !strings.Contains(out, "Warning") || !strings.Contains(out, "DropColumn") || !strings.Contains(out, "name") {
		t.Fatalf("missing generated drop warning: %s", out)
	}
}
