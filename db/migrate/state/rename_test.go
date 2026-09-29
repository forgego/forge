package state

import (
	"reflect"
	"testing"
)

// TestLoadReadsBackHandWrittenRenames covers ALTER TABLE .. RENAME TO and
// RENAME COLUMN, which the loader did not parse, and the foreign keys,
// indexes and constraints that follow a renamed table or column.
func TestLoadReadsBackHandWrittenRenames(t *testing.T) {
	dir := writeMigrations(t, map[string]string{
		"000001_a.up.sql": `CREATE TABLE authors ("id" BIGINT PRIMARY KEY, "name" TEXT);
CREATE TABLE books ("id" BIGINT PRIMARY KEY, "author_id" BIGINT, "pages" INTEGER, "isbn" TEXT);
ALTER TABLE books ADD CONSTRAINT fk_books_author_id FOREIGN KEY ("author_id") REFERENCES authors (id) ON DELETE CASCADE;
ALTER TABLE books ADD CONSTRAINT books_pages_positive CHECK (pages > 0 AND "pages" < 'pages' AND pages_x IS NULL);
ALTER TABLE books ADD CONSTRAINT books_isbn_key UNIQUE ("isbn");
CREATE INDEX IF NOT EXISTS books_isbn_idx ON books ("isbn", "pages");`,
		"000002_b.up.sql": `ALTER TABLE authors RENAME TO writers;
ALTER TABLE books RENAME COLUMN pages TO page_count;
ALTER TABLE books RENAME isbn TO code;
ALTER TABLE books RENAME COLUMN author_id TO writer_id;
ALTER TABLE books RENAME TO volumes;
ALTER TABLE volumes RENAME CONSTRAINT fk_books_author_id TO fk_volumes_writer_id;`,
	})
	state, err := LoadStateFromFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if state.Tables["authors"] != nil || state.Tables["books"] != nil {
		t.Fatalf("old table names remain: %v", state.Tables)
	}
	writers, volumes := state.Tables["writers"], state.Tables["volumes"]
	if writers == nil || volumes == nil {
		t.Fatalf("renamed tables missing: %v", state.Tables)
	}
	if writers.Name != "writers" || volumes.Name != "volumes" {
		t.Errorf("table names %q, %q", writers.Name, volumes.Name)
	}
	for _, column := range []string{"id", "writer_id", "page_count", "code"} {
		if col := volumes.Columns[column]; col == nil || col.Name != column {
			t.Errorf("column %s missing or misnamed: %+v", column, col)
		}
	}
	if len(volumes.Columns) != 4 {
		t.Errorf("volumes columns %v", volumes.Columns)
	}
	fk := volumes.ForeignKeys["fk_volumes_writer_id"]
	if fk == nil || len(volumes.ForeignKeys) != 1 || fk.Column != "writer_id" || fk.TargetTable != "writers" {
		t.Errorf("foreign keys %+v", volumes.ForeignKeys)
	}
	if idx := volumes.Indexes["books_isbn_idx"]; idx == nil || !reflect.DeepEqual(idx.Fields, []string{"code", "page_count"}) {
		t.Errorf("index %+v", idx)
	}
	if unique := volumes.Constraints["books_isbn_key"]; unique == nil || !reflect.DeepEqual(unique.Fields, []string{"code"}) {
		t.Errorf("unique constraint %+v", unique)
	}
	want := `page_count > 0 AND "page_count" < 'pages' AND pages_x IS NULL`
	if check := volumes.Constraints["books_pages_positive"]; check == nil || check.Condition != want {
		t.Errorf("check constraint %+v, want condition %s", check, want)
	}
}

// A generated column's expression follows a renamed column it computes from,
// as it does in the database.
func TestRenameColumnRewritesGeneratedExpressions(t *testing.T) {
	dir := writeMigrations(t, map[string]string{
		"000001_a.up.sql": `CREATE TABLE items ("id" BIGINT PRIMARY KEY, "price" INTEGER, "double_price" INTEGER GENERATED ALWAYS AS (price * 2) STORED);`,
		"000002_b.up.sql": `ALTER TABLE items RENAME COLUMN price TO unit_price;`,
	})
	state, err := LoadStateFromFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	col := state.Tables["items"].Columns["double_price"]
	if col == nil {
		t.Fatalf("columns %v", state.Tables["items"].Columns)
	}
	if expr := col.Options["generated_expr"]; expr != "unit_price * 2" {
		t.Errorf("generated_expr %q, want %q", expr, "unit_price * 2")
	}
}
