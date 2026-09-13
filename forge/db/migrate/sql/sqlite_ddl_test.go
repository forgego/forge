package sql

import (
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

type sqliteForeignKeyInfo struct {
	id       int
	seq      int
	table    string
	from     string
	to       string
	onUpdate string
	onDelete string
	match    string
}

func openMemoryDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func getTableColumns(t *testing.T, db *sql.DB, tableName string) []string {
	t.Helper()
	rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
	require.NoError(t, err)
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString
		err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk)
		require.NoError(t, err)
		columns = append(columns, name)
	}
	require.NoError(t, rows.Err())
	return columns
}

func getTableForeignKeys(t *testing.T, db *sql.DB, tableName string) []sqliteForeignKeyInfo {
	t.Helper()
	rows, err := db.Query("PRAGMA foreign_key_list(" + tableName + ")")
	require.NoError(t, err)
	defer rows.Close()

	var fks []sqliteForeignKeyInfo
	for rows.Next() {
		var fk sqliteForeignKeyInfo
		err := rows.Scan(&fk.id, &fk.seq, &fk.table, &fk.from, &fk.to, &fk.onUpdate, &fk.onDelete, &fk.match)
		require.NoError(t, err)
		fks = append(fks, fk)
	}
	require.NoError(t, rows.Err())
	return fks
}

func makeCategoryModel() *generator.ModelDefinition {
	return &generator.ModelDefinition{
		Name: "Category",
		Meta: generator.MetaDefinition{TableName: "categories"},
		Fields: []generator.FieldDefinition{
			{Name: "id", Type: "Int64", PrimaryKey: true, AutoIncrement: true},
			{Name: "name", Type: "String"},
		},
	}
}

func makeProductModel() *generator.ModelDefinition {
	return &generator.ModelDefinition{
		Name: "Product",
		Meta: generator.MetaDefinition{TableName: "products"},
		Fields: []generator.FieldDefinition{
			{Name: "id", Type: "Int64", PrimaryKey: true, AutoIncrement: true},
			{Name: "category_id", Type: "Int64", Required: true},
			{Name: "name", Type: "String"},
		},
		Relations: []generator.RelationDefinition{
			{
				Name:    "category_id",
				Type:    "ForeignKey",
				To:      "Category",
				Options: map[string]interface{}{"on_delete": "CASCADE", "on_update": "CASCADE"},
			},
		},
	}
}

func TestSQLiteBuilder_CreateTable_WithForeignKeys(t *testing.T) {
	categories := makeCategoryModel()
	products := makeProductModel()
	changes := []core.Change{
		&core.CreateTable{Table: categories},
		&core.CreateTable{Table: products},
		&core.AddForeignKey{Table: "products", Relation: products.Relations[0], TargetTable: "categories"},
	}

	builder := NewSQLiteBuilder()
	upSQL, err := builder.BuildUpSQL(changes)
	require.NoError(t, err)
	assert.NotContains(t, upSQL, "DO $$")

	db := openMemoryDB(t)
	_, err = db.Exec(upSQL)
	require.NoError(t, err)

	fks := getTableForeignKeys(t, db, "products")
	require.Len(t, fks, 1)
	assert.Equal(t, "categories", fks[0].table)
	assert.Equal(t, "category_id", fks[0].from)

	downSQL, err := builder.BuildDownSQL(changes)
	require.NoError(t, err)
	_, err = db.Exec(downSQL)
	require.NoError(t, err)
	assert.Empty(t, getTableColumns(t, db, "products"))
}

func TestSQLiteBuilder_DropColumn_Executes(t *testing.T) {
	db := openMemoryDB(t)
	_, err := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT, name TEXT);")
	require.NoError(t, err)

	builder := NewSQLiteBuilder()
	change := &core.DropColumn{Table: "users", ColumnName: "email"}
	upSQL, err := builder.BuildUpSQL([]core.Change{change})
	require.NoError(t, err)

	_, err = db.Exec(upSQL)
	require.NoError(t, err)

	cols := getTableColumns(t, db, "users")
	assert.NotContains(t, cols, "email")
	assert.Contains(t, cols, "id")
	assert.Contains(t, cols, "name")
}

func TestSQLiteBuilder_AddColumn_Rollback_Executes(t *testing.T) {
	db := openMemoryDB(t)
	_, err := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);")
	require.NoError(t, err)

	builder := NewSQLiteBuilder()
	change := &core.AddColumn{
		Table:  "users",
		Column: generator.FieldDefinition{Name: "email", Type: "String"},
	}

	upSQL, err := builder.BuildUpSQL([]core.Change{change})
	require.NoError(t, err)
	_, err = db.Exec(upSQL)
	require.NoError(t, err)
	assert.Contains(t, getTableColumns(t, db, "users"), "email")

	downSQL, err := builder.BuildDownSQL([]core.Change{change})
	require.NoError(t, err)
	_, err = db.Exec(downSQL)
	require.NoError(t, err)
	assert.NotContains(t, getTableColumns(t, db, "users"), "email")
}

func TestSQLiteBuilder_UnsupportedOperations_Error(t *testing.T) {
	rel := generator.RelationDefinition{Name: "category_id"}

	tests := []struct {
		name       string
		isDown     bool
		change     core.Change
		errSnippet string
	}{
		{
			name:       "AddForeignKey on existing table up",
			isDown:     false,
			change:     &core.AddForeignKey{Table: "products", Relation: rel, TargetTable: "categories"},
			errSnippet: "add foreign key is not supported on SQLite without rebuilding the table",
		},
		{
			name:       "AddForeignKey on existing table down",
			isDown:     true,
			change:     &core.AddForeignKey{Table: "products", Relation: rel, TargetTable: "categories"},
			errSnippet: "add foreign key is not supported on SQLite without rebuilding the table",
		},
		{
			name:       "DropForeignKey up",
			isDown:     false,
			change:     &core.DropForeignKey{Table: "products", FKName: "fk_products_category_id"},
			errSnippet: "drop foreign key is not supported on SQLite without rebuilding the table",
		},
		{
			name:       "DropForeignKey down",
			isDown:     true,
			change:     &core.DropForeignKey{Table: "products", FKName: "fk_products_category_id"},
			errSnippet: "drop foreign key is not supported on SQLite without rebuilding the table",
		},
		{
			name:       "DropConstraint up",
			isDown:     false,
			change:     &core.DropConstraint{Table: "products", ConstraintName: "chk_positive"},
			errSnippet: "drop constraint is not supported on SQLite without rebuilding the table",
		},
		{
			name:       "DropConstraint down",
			isDown:     true,
			change:     &core.DropConstraint{Table: "products", ConstraintName: "chk_positive"},
			errSnippet: "drop constraint is not supported on SQLite without rebuilding the table",
		},
	}

	builder := NewSQLiteBuilder()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.isDown {
				_, err = builder.BuildDownSQL([]core.Change{tt.change})
			} else {
				_, err = builder.BuildUpSQL([]core.Change{tt.change})
			}
			require.Error(t, err)
			var migErr *core.MigrationError
			require.True(t, errors.As(err, &migErr), "expected *core.MigrationError in chain")
			assert.Equal(t, core.ErrInvalidChange, migErr.Code)
			assert.Contains(t, err.Error(), tt.errSnippet)
		})
	}
}
