package state

import (
	"os"
	"path/filepath"
	"testing"
)

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
