package generate_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/generate"
)

// The table definitions of the SQLite rebuild migration below, as the SQLite
// builder renders them for functionalModels.
const (
	sqliteAuthorColumns = `    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "name" TEXT NOT NULL UNIQUE,
    "score" REAL DEFAULT 1.500000,
    "ratio" REAL,
    "visits" INTEGER DEFAULT 0,
    "status" TEXT DEFAULT 'Active',
    "motto" TEXT DEFAULT 'read, then write it''s',
    "rank" INTEGER DEFAULT (1 + 2),
    "balance" REAL,
    "born_on" TIMESTAMP,
    "seen_at" TIMESTAMP,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP`
	sqliteAuthorColumnList = `"id", "name", "score", "ratio", "visits", "status", "motto", "rank", "balance", "born_on", "seen_at", "created_at", "updated_at"`
	sqliteBookColumns      = `    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "author_id" INTEGER NOT NULL,
    "editor_id" INTEGER,
    "reviewer_id" INTEGER,
    "pages" INTEGER DEFAULT 1,
    "isbn" TEXT,
    FOREIGN KEY (author_id) REFERENCES authors (id) ON DELETE CASCADE ON UPDATE NO ACTION,
    FOREIGN KEY (editor_id) REFERENCES authors (id) ON DELETE SET NULL ON UPDATE NO ACTION,`
	sqliteBookColumnList = `"id", "author_id", "editor_id", "reviewer_id", "pages", "isbn"`
)

// sqliteRebuild returns a hand-written SQLite table rebuild, as the
// migrations guide describes it: create the new table, copy the rows, drop
// the old table, rename the new one, and re-create the indexes.
func sqliteRebuild(table, definition, columns, indexes string) string {
	return "CREATE TABLE " + table + "_new (\n" + definition + "\n);\n" +
		"INSERT INTO " + table + "_new (" + columns + ")\n    SELECT " + columns + " FROM " + table + ";\n" +
		"DROP TABLE " + table + ";\n" +
		"ALTER TABLE " + table + "_new RENAME TO " + table + ";\n" + indexes
}

// TestSQLiteHandWrittenRebuildIsReadBack covers existing SQLite tables whose
// models gain a foreign key or a Meta constraint, which makemigrations cannot
// add without a table rebuild. The hand-written rebuild the migrations guide
// describes must keep the rows, enforce the new constraints, and read back so
// the next run writes nothing.
func TestSQLiteHandWrittenRebuildIsReadBack(t *testing.T) {
	before := mustReplace(t, functionalModels, "\t\t\t{Name: \"books_pages_positive\", Type: \"CHECK\", Condition: \"pages > 0\"},\n", "")
	before = mustReplace(t, before, "\t\tschema.OneToOneField(\"reviewer_id\", \"Author\", schema.OnDelete(schema.CascadePROTECT)),\n", "")
	after := mustReplace(t, functionalModels, `return schema.Meta{TableName: "authors"}`,
		`return schema.Meta{TableName: "authors", Constraints: []schema.Constraint{{Name: "authors_visits_nonneg", Type: "CHECK", Condition: "visits >= 0"}}}`)

	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()
	file := filepath.Join(modelsDir, "models.go")
	if err := os.WriteFile(file, []byte(before), 0o600); err != nil {
		t.Fatal(err)
	}
	generateMigration(t, modelsDir, migrationsDir, core.DriverSQLite, "initial")
	if err := os.WriteFile(file, []byte(after), 0o600); err != nil {
		t.Fatal(err)
	}
	gen, err := generate.NewMigrationGeneratorForDriver(modelsDir, migrationsDir, core.DriverSQLite)
	if err != nil {
		t.Fatal(err)
	}
	if err := gen.GenerateMigrations("add_constraints"); err == nil || !strings.Contains(err.Error(), "not supported on SQLite without rebuilding the table") {
		t.Fatalf("adding constraints to existing SQLite tables: got error %v", err)
	}

	index := "CREATE INDEX IF NOT EXISTS books_isbn_idx ON books (\"isbn\");\n"
	up := sqliteRebuild("authors", sqliteAuthorColumns+",\n    CONSTRAINT authors_visits_nonneg CHECK (visits >= 0)", sqliteAuthorColumnList, "") +
		sqliteRebuild("books", sqliteBookColumns+`
    FOREIGN KEY (reviewer_id) REFERENCES authors (id) ON DELETE RESTRICT ON UPDATE NO ACTION,
    CONSTRAINT books_pages_positive CHECK (pages > 0),
    CONSTRAINT books_isbn_key UNIQUE ("isbn")`, sqliteBookColumnList, index)
	down := sqliteRebuild("books", sqliteBookColumns+`
    CONSTRAINT books_isbn_key UNIQUE ("isbn")`, sqliteBookColumnList, index) +
		sqliteRebuild("authors", sqliteAuthorColumns, sqliteAuthorColumnList, "")
	for name, content := range map[string]string{"000002_rebuild.up.sql": up, "000002_rebuild.down.sql": down} {
		if err := os.WriteFile(filepath.Join(migrationsDir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	generateMigration(t, modelsDir, migrationsDir, core.DriverSQLite, "again")
	if got := migrationFiles(t, migrationsDir); len(got) != 4 {
		extra, _ := os.ReadFile(filepath.Join(migrationsDir, got[len(got)-1]))
		t.Fatalf("regenerating after the rebuild wrote %v:\n%s", got, extra)
	}
	applyMigrations(t, core.DriverSQLite, migrationsDir)

	// Rebuild tables that hold rows, each migration in a transaction as
	// forge migrate runs it, on a connection that leaves foreign keys off.
	ctx := context.Background()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "rebuild.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	run := func(name string) {
		t.Helper()
		content, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			t.Fatal(err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			t.Fatalf("%s: %v", name, err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	}
	query := func(q string) string {
		t.Helper()
		rows, err := db.QueryContext(ctx, q)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			var value sql.NullString
			if err := rows.Scan(&value); err != nil {
				t.Fatal(err)
			}
			out = append(out, value.String)
		}
		return strings.Join(out, ",")
	}
	checkRows := func(stage string) {
		t.Helper()
		if got := query(`SELECT name || ':' || visits FROM authors ORDER BY id`); got != "ann:1,bob:2" {
			t.Errorf("%s: authors %s", stage, got)
		}
		if got := query(`SELECT id || ':' || author_id || ':' || IFNULL(editor_id, '-') || ':' || IFNULL(reviewer_id, '-') || ':' || pages || ':' || isbn FROM books ORDER BY id`); got != "1:1:2:2:10:a,2:2:-:1:20:b,3:2:1:-:30:c" {
			t.Errorf("%s: books %s", stage, got)
		}
		if got := query(`SELECT 1 FROM pragma_foreign_key_check`); got != "" {
			t.Errorf("%s: foreign key violations %s", stage, got)
		}
	}

	run("000001_initial.up.sql")
	if _, err := db.ExecContext(ctx, `INSERT INTO authors (id, name, visits) VALUES (1, 'ann', 1), (2, 'bob', 2);
INSERT INTO books (id, author_id, editor_id, reviewer_id, pages, isbn) VALUES (1, 1, 2, 2, 10, 'a'), (2, 2, NULL, 1, 20, 'b'), (3, 2, 1, NULL, 30, 'c');`); err != nil {
		t.Fatal(err)
	}
	run("000002_rebuild.up.sql")
	checkRows("after the rebuild")
	for _, violation := range []string{
		`INSERT INTO books (author_id, pages, isbn) VALUES (1, 0, 'd')`,
		`INSERT INTO authors (name, visits) VALUES ('cy', -1)`,
	} {
		if _, err := db.ExecContext(ctx, violation); err == nil || !strings.Contains(err.Error(), "CHECK constraint failed") {
			t.Errorf("%s: got %v, want a CHECK failure", violation, err)
		}
	}
	if got := query(`SELECT sql FROM sqlite_master WHERE name = 'books'`); !strings.Contains(got, "FOREIGN KEY (reviewer_id) REFERENCES authors (id) ON DELETE RESTRICT") {
		t.Errorf("rebuilt books lacks the new foreign key:\n%s", got)
	}
	if got := query(`SELECT name FROM sqlite_master WHERE type = 'index' AND name = 'books_isbn_idx'`); got != "books_isbn_idx" {
		t.Errorf("rebuilt books lacks its index")
	}
	if got := query(`SELECT seq FROM sqlite_sequence WHERE name = 'books'`); got != "3" {
		t.Errorf("rebuilt books AUTOINCREMENT sequence %q, want 3", got)
	}
	run("000002_rebuild.down.sql")
	checkRows("after rolling back the rebuild")
	run("000002_rebuild.up.sql")
	checkRows("after the rebuild again")
}
