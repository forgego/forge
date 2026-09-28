package orm

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TxTestItem struct {
	schema.BaseSchema
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Value int64  `db:"value"`
}

func (TxTestItem) Meta() schema.Meta {
	return schema.Meta{TableName: "tx_test_items"}
}

func (TxTestItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
		schema.Int64Field("value"),
	}
}

func setupTxTestDB(t *testing.T) (*db.DB, *Manager[TxTestItem]) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "tx_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)

	_, err = database.Exec(`
		CREATE TABLE tx_test_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value INTEGER NOT NULL
		);
	`)
	require.NoError(t, err)

	_, err = GetModelSchema[TxTestItem]()
	require.NoError(t, err)

	mgr, err := NewManagerWithDB[TxTestItem]("tx_test_items", database)
	require.NoError(t, err)

	return database, mgr
}

func TestManager_WithTx_Rollback(t *testing.T) {
	database, mgr := setupTxTestDB(t)
	ctx := context.Background()

	errRollback := errors.New("force rollback")
	err := database.WithTx(ctx, func(tx *db.Tx) error {
		txMgr := mgr.WithTx(tx)
		item := &TxTestItem{Name: "item-rollback", Value: 42}
		if err := txMgr.Create(ctx, item); err != nil {
			return err
		}
		assert.NotZero(t, item.ID)

		fetched, err := txMgr.Get(ctx, item.ID)
		assert.NoError(t, err)
		assert.NotNil(t, fetched)
		assert.Equal(t, "item-rollback", fetched.Name)

		return errRollback
	})
	require.ErrorIs(t, err, errRollback)

	count, err := mgr.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestManager_WithTx_Commit(t *testing.T) {
	database, mgr := setupTxTestDB(t)
	ctx := context.Background()

	var createdID int64
	err := database.WithTx(ctx, func(tx *db.Tx) error {
		txMgr := mgr.WithTx(tx)
		item := &TxTestItem{Name: "item-commit", Value: 100}
		if err := txMgr.Create(ctx, item); err != nil {
			return err
		}
		createdID = item.ID
		return nil
	})
	require.NoError(t, err)

	count, err := mgr.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	item, err := mgr.Get(ctx, createdID)
	require.NoError(t, err)
	assert.Equal(t, "item-commit", item.Name)
	assert.Equal(t, int64(100), item.Value)
}

func TestQuerySet_WithTx_UncommittedRows(t *testing.T) {
	database, mgr := setupTxTestDB(t)
	ctx := context.Background()

	err := database.WithTx(ctx, func(tx *db.Tx) error {
		txMgr := mgr.WithTx(tx)
		item := &TxTestItem{Name: "item-filter", Value: 200}
		if err := txMgr.Create(ctx, item); err != nil {
			return err
		}

		qs, err := txMgr.Filter(F("name").Eq("item-filter"))
		if err != nil {
			return err
		}

		items, err := qs.All(ctx)
		if err != nil {
			return err
		}
		assert.Len(t, items, 1)
		assert.Equal(t, "item-filter", items[0].Name)
		assert.Equal(t, int64(200), items[0].Value)

		return nil
	})
	require.NoError(t, err)
}

func TestManager_WithTx_BulkCreate_Rollback(t *testing.T) {
	database, mgr := setupTxTestDB(t)
	ctx := context.Background()

	errRollback := errors.New("bulk rollback")
	err := database.WithTx(ctx, func(tx *db.Tx) error {
		txMgr := mgr.WithTx(tx)
		items := []*TxTestItem{
			{Name: "bulk-1", Value: 10},
			{Name: "bulk-2", Value: 20},
		}
		if err := txMgr.BulkCreate(ctx, items); err != nil {
			return err
		}

		count, err := txMgr.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)

		return errRollback
	})
	require.ErrorIs(t, err, errRollback)

	count, err := mgr.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestManager_WithTx_UpdateAndRollback(t *testing.T) {
	database, mgr := setupTxTestDB(t)
	ctx := context.Background()

	item := &TxTestItem{Name: "original", Value: 1}
	require.NoError(t, mgr.Create(ctx, item))

	errRollback := errors.New("update rollback")
	err := database.WithTx(ctx, func(tx *db.Tx) error {
		txMgr := mgr.WithTx(tx)
		item.Name = "modified-in-tx"
		if err := txMgr.Update(ctx, item); err != nil {
			return err
		}

		readBack, err := txMgr.Get(ctx, item.ID)
		assert.NoError(t, err)
		assert.Equal(t, "modified-in-tx", readBack.Name)
		return errRollback
	})
	require.ErrorIs(t, err, errRollback)

	after, err := mgr.Get(ctx, item.ID)
	require.NoError(t, err)
	assert.Equal(t, "original", after.Name)
}

func TestManager_WithTx_DeleteAndRollback(t *testing.T) {
	database, mgr := setupTxTestDB(t)
	ctx := context.Background()

	item := &TxTestItem{Name: "to-delete", Value: 5}
	require.NoError(t, mgr.Create(ctx, item))

	errRollback := errors.New("delete rollback")
	err := database.WithTx(ctx, func(tx *db.Tx) error {
		txMgr := mgr.WithTx(tx)
		if err := txMgr.Delete(ctx, item); err != nil {
			return err
		}

		count, err := txMgr.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
		return errRollback
	})
	require.ErrorIs(t, err, errRollback)

	count, err := mgr.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestQuerySet_WithTx_Update_Delete(t *testing.T) {
	database, mgr := setupTxTestDB(t)
	ctx := context.Background()

	require.NoError(t, mgr.Create(ctx, &TxTestItem{Name: "target", Value: 10}))
	require.NoError(t, mgr.Create(ctx, &TxTestItem{Name: "other", Value: 20}))

	err := database.WithTx(ctx, func(tx *db.Tx) error {
		txMgr := mgr.WithTx(tx)
		qs, err := txMgr.Filter(F("name").Eq("target"))
		if err != nil {
			return err
		}

		affected, err := qs.Update(ctx, UpdateMap{"value": int64(99)})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		delAffected, err := qs.Delete(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), delAffected)

		count, err := txMgr.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
		return nil
	})
	require.NoError(t, err)

	count, err := mgr.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestGetDBTX_And_GetDialect_Resolution(t *testing.T) {
	database, _ := setupTxTestDB(t)
	ctx := context.Background()

	dbtx, err := GetDBTX(database)
	require.NoError(t, err)
	assert.Equal(t, database.DB, dbtx)

	dbtxSQL, err := GetDBTX(database.DB)
	require.NoError(t, err)
	assert.Equal(t, database.DB, dbtxSQL)

	d, err := GetDialect(database)
	require.NoError(t, err)
	assert.Equal(t, "sqlite", d.Name())

	err = database.WithTx(ctx, func(tx *db.Tx) error {
		assert.Equal(t, database, tx.DB())

		txDBTX, err := GetDBTX(tx)
		assert.NoError(t, err)
		assert.Equal(t, tx.Tx, txDBTX)

		sqlTxDBTX, err := GetDBTX(tx.Tx)
		assert.NoError(t, err)
		assert.Equal(t, tx.Tx, sqlTxDBTX)

		rawDialect, err := GetDialect(tx.Tx)
		require.NoError(t, err)
		assert.Equal(t, NewSQLBuilder().Placeholder(1), rawDialect.Placeholder(1))

		txDialect, err := GetDialect(tx)
		assert.NoError(t, err)
		assert.Equal(t, "sqlite", txDialect.Name())

		return nil
	})
	require.NoError(t, err)

	_, err = GetDBTX("invalid")
	assert.Error(t, err)

	_, err = GetDialect("invalid")
	assert.Error(t, err)

	var nilTx *db.Tx
	_, err = GetDBTX(nilTx)
	assert.Error(t, err)
	_, err = GetDialect(nilTx)
	assert.Error(t, err)

	var nilSQLTx *sql.Tx
	_, err = GetDialect(nilSQLTx)
	assert.Error(t, err)

	var nilSQLDB *sql.DB
	_, err = GetDBTX(nilSQLDB)
	assert.Error(t, err)
}
