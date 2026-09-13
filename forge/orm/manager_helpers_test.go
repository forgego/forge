package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInsertSQL_UsesSchemaFieldsAndSkipsOptionalZeroValues(t *testing.T) {
	instance := testModel{
		Name: "Widget",
	}

	sql, values, columns, err := BuildInsertSQL(instance, "test_table", "id")
	require.NoError(t, err)

	assert.Equal(t, `INSERT INTO "test_table" ("name") VALUES ($1) RETURNING "id"`, sql)
	assert.Equal(t, []interface{}{"Widget"}, values)
	assert.Equal(t, []string{"name"}, columns)
}

func TestBuildInsertSQL_RequiredFieldIncludedEvenWhenZeroValue(t *testing.T) {
	instance := testModel{
		Name: "",
	}

	sql, values, columns, err := BuildInsertSQL(instance, "test_table", "id")
	require.NoError(t, err)

	assert.Equal(t, `INSERT INTO "test_table" ("name") VALUES ($1) RETURNING "id"`, sql)
	assert.Equal(t, []interface{}{""}, values)
	assert.Equal(t, []string{"name"}, columns)
}

func TestBuildBulkInsertSQL_ConsistentColumns(t *testing.T) {
	instances := []interface{}{
		testModel{Name: "A"},
		testModel{Name: "B"},
	}

	sql, values, columns, err := BuildBulkInsertSQL(instances, "test_table", "id")
	require.NoError(t, err)

	assert.Equal(t, `"name"`, EscapeIdentifier(columns[0]))
	assert.Equal(t, `INSERT INTO "test_table" ("name") VALUES ($1), ($2) RETURNING "id"`, sql)
	assert.Equal(t, []interface{}{"A", "B"}, values)
	assert.Equal(t, []string{"name"}, columns)
}

func TestBuildBulkInsertSQL_RejectsInconsistentColumns(t *testing.T) {
	instances := []interface{}{
		testModel{Name: "A"},
		testModel{Name: "B", Email: "b@example.com"},
	}

	_, _, _, err := BuildBulkInsertSQL(instances, "test_table", "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires consistent columns")
}

func TestBuildUpdateSQL_QuotesIdentifiers(t *testing.T) {
	instance := testModel{
		ID:   42,
		Name: "Widget",
	}

	sql, values, err := BuildUpdateSQL(instance, "order", "id")
	require.NoError(t, err)

	assert.Contains(t, sql, `UPDATE "order" SET`)
	assert.Contains(t, sql, `"name" = $1`)
	assert.Contains(t, sql, `WHERE "id" = $5`)
	assert.Equal(t, []interface{}{"Widget", "", float64(0), false, int64(42)}, values)
}

func TestBuildDeleteSQL_QuotesIdentifiers(t *testing.T) {
	sql, values := BuildDeleteSQL("order", "id", int64(42))
	assert.Equal(t, `DELETE FROM "order" WHERE "id" = $1`, sql)
	assert.Equal(t, []interface{}{int64(42)}, values)
}
