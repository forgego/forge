package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	codegen "github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

func TestApplyRejectsUnhandledChanges(t *testing.T) {
	for _, change := range []core.Change{&core.UnknownChange{SQL: "CREATE MATERIALIZED VIEW report AS SELECT 1"}, &core.RunSQL{SQL: "ALTER TABLE users ADD COLUMN x INT"}} {
		err := NewInMemoryState().Apply([]core.Change{change})
		if err == nil {
			t.Fatalf("Apply(%T) unexpectedly succeeded", change)
		}
	}
}

func TestApplyModifiesIndexState(t *testing.T) {
	manager := NewInMemoryState()
	manager.GetState().Tables["users"] = &TableState{Name: "users", Indexes: map[string]*IndexState{"old": {Name: "old", Fields: []string{"name"}}}}
	err := manager.Apply([]core.Change{&core.ModifyIndex{Table: "users", OldIndex: codegen.IndexDefinition{Name: "old"}, NewIndex: codegen.IndexDefinition{Name: "new", Fields: []string{"email"}, Unique: true}}})
	if err != nil {
		t.Fatal(err)
	}
	indexes := manager.GetState().Tables["users"].Indexes
	if indexes["old"] != nil || indexes["new"] == nil || !indexes["new"].Unique {
		t.Fatalf("indexes = %+v", indexes)
	}
}

func TestLoadRejectsUnreadableMigration(t *testing.T) {
	dir := t.TempDir()
	name := "000001_missing.up.sql"
	if err := os.Symlink(filepath.Join(dir, "gone"), filepath.Join(dir, name)); err != nil {
		t.Fatal(err)
	}
	_, err := LoadStateFromFiles(dir)
	if err == nil || !strings.Contains(err.Error(), name) {
		t.Fatalf("load error = %v, want file name", err)
	}
}

func TestGetStateDoesNotHideReplayFailure(t *testing.T) {
	dir := t.TempDir()
	if err := os.Symlink(filepath.Join(dir, "gone"), filepath.Join(dir, "000001_missing.up.sql")); err != nil {
		t.Fatal(err)
	}
	loader := NewFileStateLoader(dir)
	if got := loader.GetState(); got != nil {
		t.Fatalf("GetState returned an empty success state after replay failure: %+v", got)
	}
}

func TestLoadRejectsUnknownSchemaChange(t *testing.T) {
	dir := writeMigrations(t, map[string]string{"000001_bad.up.sql": `CREATE MATERIALIZED VIEW report AS SELECT 1;`})
	_, err := LoadStateFromFiles(dir)
	if err == nil || !strings.Contains(err.Error(), "000001_bad.up.sql") {
		t.Fatalf("load error = %v, want file name", err)
	}
}

func TestLoadRejectsUnresolvedDeferredChange(t *testing.T) {
	dir := writeMigrations(t, map[string]string{"000001_bad.up.sql": `CREATE INDEX missing_idx ON missing_table (id);`})
	_, err := LoadStateFromFiles(dir)
	if err == nil || !strings.Contains(err.Error(), "000001_bad.up.sql") {
		t.Fatalf("load error = %v, want file name", err)
	}
}

func writeMigrations(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestLoadAppliesFilesInOrder covers a table dropped in one migration and
// created again in a later one, which the loader used to create first and
// drop last.
func TestLoadAppliesFilesInOrder(t *testing.T) {
	dir := writeMigrations(t, map[string]string{
		"000001_a.up.sql": `CREATE TABLE books ("id" BIGINT PRIMARY KEY, "old" TEXT);`,
		"000002_b.up.sql": `DROP TABLE IF EXISTS books CASCADE;`,
		"000003_c.up.sql": "CREATE TABLE books (\"id\" BIGINT PRIMARY KEY, \"new\" TEXT);\nDROP TABLE IF EXISTS notes;\nCREATE TABLE notes (\"id\" BIGINT PRIMARY KEY);",
	})
	state, err := LoadStateFromFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	books := state.Tables["books"]
	if books == nil {
		t.Fatalf("re-created table missing from state: %v", state.Tables)
	}
	if _, ok := books.Columns["new"]; !ok || len(books.Columns) != 2 {
		t.Errorf("books columns %v, want id and new", books.Columns)
	}
	if state.Tables["notes"] == nil {
		t.Errorf("table created after a DROP in the same file is missing: %v", state.Tables)
	}
}

// TestLoadRetriesChangesOnLaterTables keeps the retry for migrations written
// out of order: an index on a table that a later file creates still loads.
func TestLoadRetriesChangesOnLaterTables(t *testing.T) {
	dir := writeMigrations(t, map[string]string{
		"000001_a.up.sql": `CREATE INDEX IF NOT EXISTS notes_title_idx ON notes ("title");`,
		"000002_b.up.sql": `CREATE TABLE notes ("id" BIGINT PRIMARY KEY, "title" TEXT);`,
	})
	state, err := LoadStateFromFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	notes := state.Tables["notes"]
	if notes == nil || notes.Indexes["notes_title_idx"] == nil {
		t.Fatalf("index on a later table was not retried: %+v", notes)
	}
}
