package filter_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/filter"
	"github.com/forgego/forge/stores"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func newMigratedSQLiteDB(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.NewDBWithDriver("sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	err = stores.Migrate(context.Background(), database)
	require.NoError(t, err)
	return database
}

func TestDatabaseFilterStorage_SaveLoadDeleteList(t *testing.T) {
	database := newMigratedSQLiteDB(t)
	storage := filter.NewDatabaseFilterStorage(database)

	fs, err := filter.NewFilterSet[filter.MockModel]()
	require.NoError(t, err)
	fs.Where("username").Equals("alice")
	ast := fs.GetAST()
	require.NotNil(t, ast)

	// 1. Save
	f1 := &filter.SavedFilter{
		ID:            "filter_user1_1",
		OwnerID:       "user1",
		Name:          "Alice Filter",
		Description:   "Finds user alice",
		AST:           ast,
		DialectHint:   "sqlite3",
		SchemaVersion: "1.0",
		Public:        false,
		Version:       1,
		SampleCount:   10,
		EstimatedCost: 5,
	}
	err = storage.Save(f1)
	require.NoError(t, err)

	// 2. Load
	loaded, err := storage.Load("filter_user1_1")
	require.NoError(t, err)
	require.Equal(t, f1.ID, loaded.ID)
	require.Equal(t, f1.OwnerID, loaded.OwnerID)
	require.Equal(t, f1.Name, loaded.Name)
	require.Equal(t, f1.Description, loaded.Description)
	require.Equal(t, f1.DialectHint, loaded.DialectHint)
	require.Equal(t, f1.SchemaVersion, loaded.SchemaVersion)
	require.False(t, loaded.Public)
	require.Equal(t, 1, loaded.Version)
	require.Equal(t, int64(10), loaded.SampleCount)
	require.Equal(t, 5, loaded.EstimatedCost)
	require.NotNil(t, loaded.AST)
	require.Equal(t, ast.Op, loaded.AST.Op)
	require.Equal(t, ast.Field, loaded.AST.Field)
	require.Equal(t, ast.Value, loaded.AST.Value)

	// 3. Update
	loaded.Description = "Updated description"
	loaded.Name = "Updated Name"
	loaded.SampleCount = 42
	err = storage.Update(loaded)
	require.NoError(t, err)

	reloaded, err := storage.Load("filter_user1_1")
	require.NoError(t, err)
	require.Equal(t, "Updated Name", reloaded.Name)
	require.Equal(t, "Updated description", reloaded.Description)
	require.Equal(t, int64(42), reloaded.SampleCount)

	// 4. List (with ownerID isolation and public filter visibility)
	// Add public filter for user2
	f2 := &filter.SavedFilter{
		ID:      "filter_user2_pub",
		OwnerID: "user2",
		Name:    "Public Filter",
		Public:  true,
		AST:     ast,
	}
	require.NoError(t, storage.Save(f2))

	// Add private filter for user2
	f3 := &filter.SavedFilter{
		ID:      "filter_user2_priv",
		OwnerID: "user2",
		Name:    "Private Filter",
		Public:  false,
		AST:     ast,
	}
	require.NoError(t, storage.Save(f3))

	// User1 should see: f1 (owned) and f2 (public), but NOT f3 (private to user2)
	listUser1, err := storage.List("user1")
	require.NoError(t, err)
	require.Len(t, listUser1, 2)
	ids1 := map[string]bool{listUser1[0].ID: true, listUser1[1].ID: true}
	require.True(t, ids1["filter_user1_1"])
	require.True(t, ids1["filter_user2_pub"])
	require.False(t, ids1["filter_user2_priv"])

	// User2 should see: f2 (owned/pub) and f3 (owned), but NOT f1 (private to user1)
	listUser2, err := storage.List("user2")
	require.NoError(t, err)
	require.Len(t, listUser2, 2)
	ids2 := map[string]bool{listUser2[0].ID: true, listUser2[1].ID: true}
	require.True(t, ids2["filter_user2_pub"])
	require.True(t, ids2["filter_user2_priv"])
	require.False(t, ids2["filter_user1_1"])

	// 5. Delete
	err = storage.Delete("filter_user1_1")
	require.NoError(t, err)

	_, err = storage.Load("filter_user1_1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "filter not found")

	listAfterDelete, err := storage.List("user1")
	require.NoError(t, err)
	require.Len(t, listAfterDelete, 1)
	require.Equal(t, "filter_user2_pub", listAfterDelete[0].ID)
}

func TestDatabaseFilterStorage_SaveAndLoadFilterHelpers(t *testing.T) {
	database := newMigratedSQLiteDB(t)
	storage := filter.NewDatabaseFilterStorage(database)

	fs, err := filter.NewFilterSet[filter.MockModel]()
	require.NoError(t, err)
	fs.Where("username").Equals("bob")

	saved, err := filter.SaveFilter(fs, "bob_filter", "find bob", storage)
	require.NoError(t, err)
	require.NotEmpty(t, saved.ID)

	loadedFs, err := filter.LoadFilter[filter.MockModel](saved.ID, storage)
	require.NoError(t, err)
	require.NotNil(t, loadedFs)
	require.NotNil(t, loadedFs.GetAST())
	require.Equal(t, "username", loadedFs.GetAST().Field)
	require.Equal(t, "bob", loadedFs.GetAST().Value)
}

func TestDatabaseFilterStorage_ConcurrentSaveAndList(t *testing.T) {
	database := newMigratedSQLiteDB(t)
	storage := filter.NewDatabaseFilterStorage(database)

	for i := 0; i < 10; i++ {
		filterID := fmt.Sprintf("bench_filter_%d", i)
		err := storage.Save(&filter.SavedFilter{
			ID:        filterID,
			OwnerID:   "bench_user",
			Name:      fmt.Sprintf("Bench %d", i),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		require.NoError(t, err)
	}

	list, err := storage.List("bench_user")
	require.NoError(t, err)
	require.Len(t, list, 10)
}
