package schema_test

import (
	"database/sql"
	"fmt"
	"testing"

	migratesql "github.com/forgego/forge/db/migrate/sql"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
)

func referentialActionDB(t *testing.T, action schema.CascadeType, nullable bool) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	column := "parent_id INTEGER NOT NULL DEFAULT 0"
	if nullable {
		column = "parent_id INTEGER"
	}
	relation := schema.ForeignKeyField("parent_id", "Parent", schema.OnDelete(action))
	ddl := fmt.Sprintf(`PRAGMA foreign_keys = ON; CREATE TABLE parents (id INTEGER PRIMARY KEY); CREATE TABLE children (id INTEGER PRIMARY KEY, %s, FOREIGN KEY (parent_id) REFERENCES parents(id) ON DELETE %s);`, column, migratesql.CascadeAction(string(relation.OnDelete)))
	if _, err := db.Exec(ddl); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO parents (id) VALUES (0), (1); INSERT INTO children (id, parent_id) VALUES (1, 1)`); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCascadeDeletesDependentRows(t *testing.T) {
	db := referentialActionDB(t, schema.CascadeCASCADE, false)
	if _, err := db.Exec(`DELETE FROM parents WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM children`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

func TestProtectBlocksReferencedDeletion(t *testing.T) {
	db := referentialActionDB(t, schema.CascadePROTECT, false)
	if _, err := db.Exec(`DELETE FROM parents WHERE id = 1`); err == nil {
		t.Fatal("referenced parent was deleted")
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM parents WHERE id = 1`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

func TestSetNullClearsDependentForeignKeys(t *testing.T) {
	db := referentialActionDB(t, schema.CascadeSET_NULL, true)
	if _, err := db.Exec(`DELETE FROM parents WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	var parent sql.NullInt64
	if err := db.QueryRow(`SELECT parent_id FROM children WHERE id = 1`).Scan(&parent); err != nil || parent.Valid {
		t.Fatalf("parent=%v err=%v", parent, err)
	}
}

func TestSetDefaultAppliesSQLDefault(t *testing.T) {
	db := referentialActionDB(t, schema.CascadeSET_DEFAULT, false)
	if _, err := db.Exec(`DELETE FROM parents WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	var parent int64
	if err := db.QueryRow(`SELECT parent_id FROM children WHERE id = 1`).Scan(&parent); err != nil || parent != 0 {
		t.Fatalf("parent=%d err=%v", parent, err)
	}
}
