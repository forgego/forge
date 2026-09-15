package orm

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "exec_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)

	_, err = database.Exec(`
		CREATE TABLE test_exec_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	return database
}

func TestExecuteInsert_MatchesDefault(t *testing.T) {
	ctx := context.Background()
	db1 := setupTestDB(t)
	db2 := setupTestDB(t)

	query := "INSERT INTO test_exec_items (name) VALUES (?) RETURNING id"
	args := []interface{}{"item1"}

	id1, err1 := ExecuteInsert(ctx, db1, query, args)
	require.NoError(t, err1)

	id2, err2 := ExecuteInsertTx(ctx, db2.DB, db2.Dialect(), query, args)
	require.NoError(t, err2)

	require.Equal(t, id2, id1)
}

func TestExecuteBulkInsert_MatchesDefault(t *testing.T) {
	ctx := context.Background()
	db1 := setupTestDB(t)
	db2 := setupTestDB(t)

	query := "INSERT INTO test_exec_items (name) VALUES (?), (?) RETURNING id"
	args := []interface{}{"item1", "item2"}

	ids1, err1 := ExecuteBulkInsert(ctx, db1, query, args)
	require.NoError(t, err1)

	ids2, err2 := ExecuteBulkInsertTx(ctx, db2.DB, db2.Dialect(), query, args)
	require.NoError(t, err2)

	require.Equal(t, ids2, ids1)
}

func TestExecuteUpdate_MatchesDefault(t *testing.T) {
	ctx := context.Background()
	db1 := setupTestDB(t)
	db2 := setupTestDB(t)

	_, err := db1.Exec("INSERT INTO test_exec_items (id, name) VALUES (1, 'initial')")
	require.NoError(t, err)
	_, err = db2.Exec("INSERT INTO test_exec_items (id, name) VALUES (1, 'initial')")
	require.NoError(t, err)

	query := "UPDATE test_exec_items SET name = ? WHERE id = ?"
	args := []interface{}{"updated", 1}

	rows1, err1 := ExecuteUpdate(ctx, db1, query, args)
	require.NoError(t, err1)

	rows2, err2 := ExecuteUpdateTx(ctx, db2.DB, db2.Dialect(), query, args)
	require.NoError(t, err2)

	require.Equal(t, rows2, rows1)
}

func TestExecuteDelete_MatchesDefault(t *testing.T) {
	ctx := context.Background()
	db1 := setupTestDB(t)
	db2 := setupTestDB(t)

	_, err := db1.Exec("INSERT INTO test_exec_items (id, name) VALUES (1, 'initial')")
	require.NoError(t, err)
	_, err = db2.Exec("INSERT INTO test_exec_items (id, name) VALUES (1, 'initial')")
	require.NoError(t, err)

	query := "DELETE FROM test_exec_items WHERE id = ?"
	args := []interface{}{1}

	rows1, err1 := ExecuteDelete(ctx, db1, query, args)
	require.NoError(t, err1)

	rows2, err2 := ExecuteDeleteTx(ctx, db2.DB, db2.Dialect(), query, args)
	require.NoError(t, err2)

	require.Equal(t, rows2, rows1)
}
