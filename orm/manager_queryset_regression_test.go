package orm

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestManager_QuerySet_FilterCountAll_NoUnsupportedConnError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "manager_qs_regression.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)

	_, err = database.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			price REAL,
			available INTEGER
		);
	`)
	require.NoError(t, err)

	mgr, err := NewManagerWithDB[testModel]("test_table", database)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, mgr.Create(ctx, &testModel{Name: "qs-reg", Email: "a@b.c", Price: 1.5, Available: true}))

	qs := mgr.QuerySet().Filter(F("name").Eq("qs-reg"))

	count, err := qs.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	items, err := qs.All(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1)
}
