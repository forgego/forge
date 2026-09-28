package sql

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

func TestDropDownSQL_DownSQLGeneration(t *testing.T) {
	builders := []struct {
		name    string
		builder SQLBuilder
	}{
		{name: "PostgreSQL", builder: NewPostgreSQLBuilder()},
		{name: "SQLite", builder: NewSQLiteBuilder()},
	}

	for _, tc := range builders {
		t.Run(tc.name, func(t *testing.T) {
			testDropColumnDownSQL(t, tc.builder)
			testDropTableDownSQL(t, tc.builder)
		})
	}
}

func testDropColumnDownSQL(t *testing.T, builder SQLBuilder) {
	t.Run("DropColumn with Column generates ADD COLUMN", func(t *testing.T) {
		colDef := generator.FieldDefinition{Name: "email", Type: "String"}
		change := &core.DropColumn{
			Table:      "users",
			ColumnName: "email",
			Column:     &colDef,
		}

		downSQL, err := builder.BuildDownSQL([]core.Change{change})
		require.NoError(t, err)
		assert.Contains(t, downSQL, "ADD COLUMN")
		assert.Contains(t, downSQL, "email")
	})

	t.Run("DropColumn with nil Column returns error", func(t *testing.T) {
		change := &core.DropColumn{
			Table:      "users",
			ColumnName: "email",
			Column:     nil,
		}

		_, err := builder.BuildDownSQL([]core.Change{change})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "without column definition")
	})
}

func testDropTableDownSQL(t *testing.T, builder SQLBuilder) {
	t.Run("DropTable with Definition generates CREATE TABLE", func(t *testing.T) {
		catModel := makeCategoryModel()
		change := &core.DropTable{
			Table:      "categories",
			Definition: catModel,
		}

		downSQL, err := builder.BuildDownSQL([]core.Change{change})
		require.NoError(t, err)
		assert.Contains(t, downSQL, "CREATE TABLE")
		assert.Contains(t, downSQL, "categories")
		assert.Contains(t, downSQL, "id")
		assert.Contains(t, downSQL, "name")
	})

	t.Run("DropTable with nil Definition returns error", func(t *testing.T) {
		change := &core.DropTable{
			Table:      "categories",
			Definition: nil,
		}

		_, err := builder.BuildDownSQL([]core.Change{change})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "without table definition")
	})
}

func TestSQLite_DropColumn_RoundTripExecution(t *testing.T) {
	db := openMemoryDB(t)
	_, err := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT, name TEXT);")
	require.NoError(t, err)

	builder := NewSQLiteBuilder()
	colDef := generator.FieldDefinition{Name: "email", Type: "String"}
	change := &core.DropColumn{
		Table:      "users",
		ColumnName: "email",
		Column:     &colDef,
	}

	upSQL, err := builder.BuildUpSQL([]core.Change{change})
	require.NoError(t, err)
	_, err = db.Exec(upSQL)
	require.NoError(t, err)
	assert.NotContains(t, getTableColumns(t, db, "users"), "email")

	downSQL, err := builder.BuildDownSQL([]core.Change{change})
	require.NoError(t, err)
	_, err = db.Exec(downSQL)
	require.NoError(t, err)
	assert.Contains(t, getTableColumns(t, db, "users"), "email")
}

func TestSQLite_DropTable_RoundTripExecution(t *testing.T) {
	db := openMemoryDB(t)
	catModel := makeCategoryModel()
	builder := NewSQLiteBuilder()

	createChange := &core.CreateTable{Table: catModel}
	createSQL, err := builder.BuildUpSQL([]core.Change{createChange})
	require.NoError(t, err)
	_, err = db.Exec(createSQL)
	require.NoError(t, err)
	require.True(t, tableExistsInSQLite(t, db, "categories"))

	dropChange := &core.DropTable{Table: "categories", Definition: catModel}
	upSQL, err := builder.BuildUpSQL([]core.Change{dropChange})
	require.NoError(t, err)
	_, err = db.Exec(upSQL)
	require.NoError(t, err)
	assert.False(t, tableExistsInSQLite(t, db, "categories"))

	downSQL, err := builder.BuildDownSQL([]core.Change{dropChange})
	require.NoError(t, err)
	_, err = db.Exec(downSQL)
	require.NoError(t, err)
	assert.True(t, tableExistsInSQLite(t, db, "categories"))
}

func tableExistsInSQLite(t *testing.T, db *sql.DB, tableName string) bool {
	t.Helper()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
	require.NoError(t, err)
	return count > 0
}
