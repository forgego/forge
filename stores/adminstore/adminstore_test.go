package adminstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/forgego/forge/admin/api/rest"
	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/stores"
	"github.com/forgego/forge/stores/adminstore"
	"github.com/forgego/forge/stores/storestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func twoInstances(t *testing.T, backend storestest.Backend) (*stores.Stores, *stores.Stores) {
	t.Helper()
	first, dsn := backend.OpenMigrated(t)
	a, err := stores.New(first)
	require.NoError(t, err)
	b, err := stores.New(storestest.Open(t, backend.Driver, dsn))
	require.NoError(t, err)
	return a, b
}

func TestHistory_SharedBetweenInstancesNewestFirst(t *testing.T) {
	for _, backend := range storestest.Backends() {
		t.Run(backend.Name, func(t *testing.T) {
			ctx := context.Background()
			a, b := twoInstances(t, backend)
			ha, hb := adminstore.NewHistory(a), adminstore.NewHistory(b)
			at := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

			require.NoError(t, ha.LogAction(ctx, core.LogEntry{
				Timestamp: at, UserID: "alice", UserName: "Alice", ModelName: "Products",
				ObjectID: "7", ObjectRepr: "Lamp", Action: core.ActionAdd, ChangeStats: `{"name":"Lamp"}`,
			}))
			require.NoError(t, hb.LogAction(ctx, core.LogEntry{
				UserID: "bob", ModelName: "products", ObjectID: "7", ObjectRepr: "Lamp", Action: core.ActionChange,
			}))
			require.NoError(t, ha.LogAction(ctx, core.LogEntry{
				UserID: "bob", ModelName: "products", ObjectID: "8", ObjectRepr: "Desk", Action: core.ActionDelete,
			}))

			entries, err := hb.GetHistory(ctx, "products", "7")
			require.NoError(t, err)
			require.Len(t, entries, 2, "history written on A and B is listed on B")
			assert.Equal(t, core.ActionChange, entries[0].Action, "newest first")
			assert.Equal(t, "bob", entries[0].UserID)
			assert.False(t, entries[0].Timestamp.IsZero(), "a missing timestamp is filled in")
			first := entries[1]
			assert.Equal(t, core.ActionAdd, first.Action)
			assert.Equal(t, "Alice", first.UserName)
			assert.Equal(t, "Products", first.ModelName)
			assert.Equal(t, "Lamp", first.ObjectRepr)
			assert.Equal(t, `{"name":"Lamp"}`, first.ChangeStats)
			assert.True(t, at.Equal(first.Timestamp), "timestamp %v", first.Timestamp)
			assert.Greater(t, entries[0].ID, first.ID)

			all, err := ha.GetHistory(ctx, "", "")
			require.NoError(t, err)
			assert.Len(t, all, 3)
			none, err := ha.GetHistory(ctx, "orders", "")
			require.NoError(t, err)
			assert.NotNil(t, none)
			assert.Empty(t, none)
		})
	}
}

func TestSavedViews_SharedPerUserAndModel(t *testing.T) {
	for _, backend := range storestest.Backends() {
		t.Run(backend.Name, func(t *testing.T) {
			ctx := context.Background()
			a, b := twoInstances(t, backend)
			va, vb := adminstore.NewSavedViews(a), adminstore.NewSavedViews(b)

			view, created, err := va.SaveView(ctx, "alice", "products", rest.SavedViewRequest{
				Name: "Cheap", Filters: map[string]interface{}{"price__lt": float64(10)},
				Ordering: []string{"-price"}, Display: []string{"name", "price"},
			})
			require.NoError(t, err)
			require.True(t, created)
			require.NotEmpty(t, view.ID)

			listed, err := vb.ListSavedViews(ctx, "alice", "products")
			require.NoError(t, err)
			require.Len(t, listed, 1, "a view saved on A is listed on B")
			assert.Equal(t, view.ID, listed[0].ID)
			assert.Equal(t, "Cheap", listed[0].Name)
			assert.Equal(t, map[string]interface{}{"price__lt": float64(10)}, listed[0].Filters)
			assert.Equal(t, []string{"-price"}, listed[0].Ordering)
			assert.Equal(t, []string{"name", "price"}, listed[0].Display)

			updated, created, err := vb.SaveView(ctx, "alice", "products", rest.SavedViewRequest{Name: "CHEAP", Ordering: []string{"price"}})
			require.NoError(t, err)
			assert.False(t, created, "saving under the same name, in any case, updates the view")
			assert.Equal(t, view.ID, updated.ID)
			assert.Equal(t, "Cheap", updated.Name, "the original name is kept")
			assert.Equal(t, []string{"price"}, updated.Ordering)

			for _, other := range []struct{ user, model string }{{"bob", "products"}, {"alice", "orders"}} {
				views, err := va.ListSavedViews(ctx, other.user, other.model)
				require.NoError(t, err)
				assert.Empty(t, views, "views are per user and model")
			}

			deleted, err := va.DeleteSavedView(ctx, "bob", "products", view.ID)
			require.NoError(t, err)
			assert.False(t, deleted, "another user cannot delete the view")
			deleted, err = va.DeleteSavedView(ctx, "alice", "products", view.ID)
			require.NoError(t, err)
			assert.True(t, deleted)
			listed, err = vb.ListSavedViews(ctx, "alice", "products")
			require.NoError(t, err)
			assert.Empty(t, listed)
		})
	}
}

func TestSQLHistorySurvivesReopen(t *testing.T) {
	ctx := context.Background()
	backend := storestest.Backend{Name: "sqlite", Driver: "sqlite3", NewDSN: storestest.SQLitePath}
	db1, dsn := backend.OpenMigrated(t)

	s1, err := stores.New(db1)
	require.NoError(t, err)
	h1 := adminstore.NewHistory(s1)

	order1 := core.LogEntry{
		UserID:      "alice",
		UserName:    "Alice",
		ModelName:   "Order",
		ObjectID:    "1",
		ObjectRepr:  "Order #1",
		Action:      core.ActionAdd,
		ChangeStats: `{"status":"pending"}`,
	}
	order2 := core.LogEntry{
		UserID:      "bob",
		UserName:    "Bob",
		ModelName:   "Order",
		ObjectID:    "2",
		ObjectRepr:  "Order #2",
		Action:      core.ActionChange,
		ChangeStats: `{"status":"shipped"}`,
	}
	customer := core.LogEntry{
		UserID:      "alice",
		UserName:    "Alice",
		ModelName:   "Customer",
		ObjectID:    "9",
		ObjectRepr:  "Customer Acme",
		Action:      core.ActionAdd,
		ChangeStats: `{"name":"Acme"}`,
	}

	require.NoError(t, h1.LogAction(ctx, order1))
	require.NoError(t, h1.LogAction(ctx, order2))
	require.NoError(t, h1.LogAction(ctx, customer))

	tx, err := db1.BeginTx(ctx, nil)
	require.NoError(t, err)
	require.NoError(t, h1.LogActionWithTx(ctx, tx, core.LogEntry{
		UserID:      "charlie",
		UserName:    "Charlie",
		ModelName:   "Order",
		ObjectID:    "3",
		ObjectRepr:  "Order #3",
		Action:      core.ActionDelete,
		ChangeStats: `{"status":"canceled"}`,
	}))
	require.NoError(t, tx.Rollback())

	require.NoError(t, db1.Close())

	db2 := storestest.Open(t, backend.Driver, dsn)
	s2, err := stores.New(db2)
	require.NoError(t, err)
	h2 := adminstore.NewHistory(s2)

	entries, err := h2.GetHistory(ctx, "order", "1")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, order1.UserID, entries[0].UserID)
	assert.Equal(t, order1.UserName, entries[0].UserName)
	assert.Equal(t, order1.Action, entries[0].Action)
	assert.Equal(t, order1.ObjectRepr, entries[0].ObjectRepr)
	assert.Equal(t, order1.ChangeStats, entries[0].ChangeStats)

	all, err := h2.GetHistory(ctx, "", "")
	require.NoError(t, err)
	require.Len(t, all, 3)
	assert.Equal(t, customer.ObjectRepr, all[0].ObjectRepr)
	assert.Equal(t, order2.ObjectRepr, all[1].ObjectRepr)
	assert.Equal(t, order1.ObjectRepr, all[2].ObjectRepr)
	assert.Greater(t, all[0].ID, all[1].ID)
	assert.Greater(t, all[1].ID, all[2].ID)

	rolledBack, err := h2.GetHistory(ctx, "order", "3")
	require.NoError(t, err)
	assert.Empty(t, rolledBack)

	tx2, err := db2.BeginTx(ctx, nil)
	require.NoError(t, err)
	require.NoError(t, h2.LogActionWithTx(ctx, tx2, core.LogEntry{
		UserID:      "david",
		UserName:    "David",
		ModelName:   "Order",
		ObjectID:    "4",
		ObjectRepr:  "Order #4",
		Action:      core.ActionAdd,
		ChangeStats: `{"status":"draft"}`,
	}))
	require.NoError(t, tx2.Rollback())

	rolledBack2, err := h2.GetHistory(ctx, "order", "4")
	require.NoError(t, err)
	assert.Empty(t, rolledBack2)

	require.NoError(t, db2.Close())

	db3 := storestest.Open(t, backend.Driver, dsn)
	s3, err := stores.New(db3)
	require.NoError(t, err)
	h3 := adminstore.NewHistory(s3)

	rolledBack3, err := h3.GetHistory(ctx, "order", "4")
	require.NoError(t, err)
	assert.Empty(t, rolledBack3)

	all3, err := h3.GetHistory(ctx, "", "")
	require.NoError(t, err)
	assert.Len(t, all3, 3)
}
