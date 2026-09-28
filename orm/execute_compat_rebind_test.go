package orm

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupExecuteCompatRebindDB(t *testing.T) *db.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "execute_compat_rebind_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)

	_, err = database.Exec(`
		CREATE TABLE test_compat_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	return database
}

func TestExecuteCompatRebind_PlaceholdersAndLiteralSafety(t *testing.T) {
	ctx := context.Background()
	database := setupExecuteCompatRebindDB(t)

	// ExecuteInsert with $1/$2 placeholders and RETURNING id
	insertSQL := "INSERT INTO test_compat_records (title, description) VALUES ($1, $2) RETURNING id"
	id1, err := ExecuteInsert(ctx, database, insertSQL, []interface{}{"item1", "first record"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), id1)

	id2, err := ExecuteInsert(ctx, database, insertSQL, []interface{}{"item2", "second record"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), id2)

	// ExecuteUpdate with $N placeholders
	updateSQL := "UPDATE test_compat_records SET title = $1, description = $2 WHERE id = $3"
	rowsUpdated, err := ExecuteUpdate(ctx, database, updateSQL, []interface{}{"item1_updated", "updated description", id1})
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsUpdated)

	// Verify the update took effect
	var title, desc string
	err = database.DB.QueryRowContext(ctx, "SELECT title, description FROM test_compat_records WHERE id = ?", id1).Scan(&title, &desc)
	require.NoError(t, err)
	assert.Equal(t, "item1_updated", title)
	assert.Equal(t, "updated description", desc)

	// ExecuteDelete with $N placeholders
	deleteSQL := "DELETE FROM test_compat_records WHERE id = $1"
	rowsDeleted, err := ExecuteDelete(ctx, database, deleteSQL, []interface{}{id1})
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsDeleted)

	// Verify row was deleted
	var count int
	err = database.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_compat_records WHERE id = ?", id1).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Assert a string literal containing '$1' in the SQL is not rewritten
	literalInsertSQL := "INSERT INTO test_compat_records (title, description) VALUES ($1, 'fixed $1 literal') RETURNING id"
	id3, err := ExecuteInsert(ctx, database, literalInsertSQL, []interface{}{"item3"})
	require.NoError(t, err)
	assert.Equal(t, int64(3), id3)

	var literalDesc string
	err = database.DB.QueryRowContext(ctx, "SELECT description FROM test_compat_records WHERE id = ?", id3).Scan(&literalDesc)
	require.NoError(t, err)
	assert.Equal(t, "fixed $1 literal", literalDesc)

	// Assert RebindPlaceholders preserves string literals containing '$1' directly
	literalQuery := "SELECT * FROM test_compat_records WHERE title = $1 AND description = 'price is $1'"
	reboundQuery := database.RebindPlaceholders(literalQuery)
	assert.Equal(t, "SELECT * FROM test_compat_records WHERE title = ?1 AND description = 'price is $1'", reboundQuery)

	// ExecuteBulkInsert with $N placeholders and RETURNING id
	bulkInsertSQL := "INSERT INTO test_compat_records (title, description) VALUES ($1, $2), ($3, $4) RETURNING id"
	bulkIDs, err := ExecuteBulkInsert(ctx, database, bulkInsertSQL, []interface{}{"bulk1", "desc1", "bulk2", "desc2"})
	require.NoError(t, err)
	assert.Equal(t, []int64{4, 5}, bulkIDs)

	// Verify nil database returns error
	_, err = ExecuteInsert(ctx, nil, insertSQL, nil)
	assert.Error(t, err)

	_, err = ExecuteBulkInsert(ctx, nil, bulkInsertSQL, nil)
	assert.Error(t, err)

	_, err = ExecuteUpdate(ctx, nil, updateSQL, nil)
	assert.Error(t, err)

	_, err = ExecuteDelete(ctx, nil, deleteSQL, nil)
	assert.Error(t, err)
}
