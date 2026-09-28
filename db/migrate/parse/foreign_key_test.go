package parse

import (
	"testing"

	"github.com/forgego/forge/db/migrate/core"
)

func parsedForeignKeys(t *testing.T, sql string) ([]*core.CreateTable, []*core.AddForeignKey) {
	t.Helper()
	changes, err := NewSQLParser().ParseUpSQL(sql)
	if err != nil {
		t.Fatal(err)
	}
	var tables []*core.CreateTable
	var fks []*core.AddForeignKey
	for _, change := range changes {
		switch c := change.(type) {
		case *core.CreateTable:
			tables = append(tables, c)
		case *core.AddForeignKey:
			fks = append(fks, c)
		case *core.UnknownChange:
			t.Errorf("unparsed statement: %s", c.SQL)
		}
	}
	return tables, fks
}

func assertForeignKey(t *testing.T, fk *core.AddForeignKey, column, target, onDelete, onUpdate string) {
	t.Helper()
	if fk.Relation.Name != column || fk.TargetTable != target ||
		fk.Relation.Options["on_delete"] != onDelete || fk.Relation.Options["on_update"] != onUpdate {
		t.Errorf("foreign key %s -> %s on delete %v on update %v, want %s -> %s on delete %s on update %s",
			fk.Relation.Name, fk.TargetTable, fk.Relation.Options["on_delete"], fk.Relation.Options["on_update"],
			column, target, onDelete, onUpdate)
	}
}

// TestTableForeignKeyActionsInEitherOrder covers table-level foreign keys whose
// ON UPDATE precedes ON DELETE, and schema-qualified targets.
func TestTableForeignKeyActionsInEitherOrder(t *testing.T) {
	_, fks := parsedForeignKeys(t, `CREATE TABLE books (
    "id" INTEGER PRIMARY KEY,
    "author_id" INTEGER,
    "editor_id" INTEGER,
    FOREIGN KEY (author_id) REFERENCES main.authors (id) ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_editor FOREIGN KEY ("editor_id") REFERENCES "main"."authors" ("id") ON DELETE RESTRICT
);`)
	if len(fks) != 2 {
		t.Fatalf("parsed %d foreign keys, want 2", len(fks))
	}
	assertForeignKey(t, fks[0], "author_id", "authors", "SET NULL", "CASCADE")
	assertForeignKey(t, fks[1], "editor_id", "authors", "PROTECT", "NO ACTION")
}

func TestAlterTableForeignKeyActionsInEitherOrder(t *testing.T) {
	_, fks := parsedForeignKeys(t, `CREATE TABLE books ("id" INTEGER, "author_id" INTEGER);
ALTER TABLE books ADD CONSTRAINT fk_books_author_id FOREIGN KEY ("author_id") REFERENCES public.authors (id) ON UPDATE SET NULL ON DELETE CASCADE;`)
	if len(fks) != 1 {
		t.Fatalf("parsed %d foreign keys, want 1", len(fks))
	}
	assertForeignKey(t, fks[0], "author_id", "authors", "CASCADE", "SET NULL")
}

// TestCreateTableIgnoresComments covers -- and /* */ comments inside the
// column list, whose commas and parentheses used to split or end it.
func TestCreateTableIgnoresComments(t *testing.T) {
	tables, fks := parsedForeignKeys(t, `CREATE TABLE books (
    "id" INTEGER PRIMARY KEY, -- the key, (really
    /* the author, see ) */ "author_id" INTEGER NOT NULL,
    "title" TEXT DEFAULT '-- not a comment',
    FOREIGN KEY (author_id) REFERENCES authors (id) -- trailing, comment
        ON DELETE CASCADE
);`)
	if len(tables) != 1 {
		t.Fatalf("parsed %d tables, want 1", len(tables))
	}
	fields := tables[0].Table.Fields
	if len(fields) != 3 || fields[0].Name != "id" || fields[1].Name != "author_id" || fields[2].Name != "title" {
		t.Fatalf("parsed fields %+v", fields)
	}
	if !fields[1].Required {
		t.Error("author_id lost NOT NULL")
	}
	if fields[2].Default != "-- not a comment" {
		t.Errorf("title default %#v", fields[2].Default)
	}
	if len(fks) != 1 {
		t.Fatalf("parsed %d foreign keys, want 1", len(fks))
	}
	assertForeignKey(t, fks[0], "author_id", "authors", "CASCADE", "NO ACTION")
}
