package orm

import (
	"context"
	"errors"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/stretchr/testify/require"
)

func TestManagerWithTxRawQueriesUseTransaction(t *testing.T) {
	database, manager := setupTxTestDB(t)
	ctx := context.Background()
	rollback := errors.New("rollback raw transaction")
	boundValue := "raw-value' OR 1=1 --"
	var bound *Manager[TxTestItem]

	err := database.WithTx(ctx, func(tx *db.Tx) error {
		bound = manager.WithTx(tx)
		if _, err := bound.RawExec(ctx,
			"INSERT INTO tx_test_items (name, value) VALUES ($1, $2)",
			boundValue, int64(10),
		); err != nil {
			return err
		}

		rows, err := bound.RawQuery(ctx,
			"SELECT name FROM tx_test_items WHERE name = $1", boundValue)
		if err != nil {
			return err
		}
		defer rows.Close()
		require.True(t, rows.Next())
		var name string
		require.NoError(t, rows.Scan(&name))
		require.Equal(t, boundValue, name)

		outside, err := manager.RawQuery(ctx,
			"SELECT COUNT(*) FROM tx_test_items WHERE name = $1", boundValue)
		if err != nil {
			return err
		}
		defer outside.Close()
		require.True(t, outside.Next())
		var outsideCount int
		require.NoError(t, outside.Scan(&outsideCount))
		require.Zero(t, outsideCount)

		qs := bound.QuerySet().Filter(F("name").Eq(boundValue))
		affected, err := qs.Update(ctx, UpdateMap{"value": int64(25)})
		require.NoError(t, err)
		require.Equal(t, int64(1), affected)
		items, err := qs.All(ctx)
		require.NoError(t, err)
		require.Equal(t, int64(25), items[0].Value)
		return rollback
	})
	require.ErrorIs(t, err, rollback)

	rows, err := manager.RawQuery(ctx,
		"SELECT COUNT(*) FROM tx_test_items WHERE name = $1", "raw-value")
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var count int
	require.NoError(t, rows.Scan(&count))
	require.Zero(t, count)

	_, err = bound.RawExec(ctx, "DELETE FROM tx_test_items WHERE name = $1", "raw-value")
	require.Error(t, err)
	require.Contains(t, err.Error(), "transaction")
	_, err = bound.QuerySet().All(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "transaction")
	_, err = bound.QuerySet().Update(ctx, UpdateMap{"value": int64(1)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "transaction")
}

func TestManagerWithTxRawQueryRollsBackOnPanic(t *testing.T) {
	database, manager := setupTxTestDB(t)
	ctx := context.Background()
	var bound *Manager[TxTestItem]
	require.PanicsWithValue(t, "stop", func() {
		_ = database.WithTx(ctx, func(tx *db.Tx) error {
			bound = manager.WithTx(tx)
			_, err := bound.RawExec(ctx,
				"INSERT INTO tx_test_items (name, value) VALUES ($1, $2)",
				"panic-value", int64(10),
			)
			if err != nil {
				return err
			}
			panic("stop")
		})
	})

	rows, err := manager.RawQuery(ctx,
		"SELECT COUNT(*) FROM tx_test_items WHERE name = $1", "panic-value")
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var count int
	require.NoError(t, rows.Scan(&count))
	require.Zero(t, count)
}

func TestManagerWithTxRawExecCommits(t *testing.T) {
	database, manager := setupTxTestDB(t)
	ctx := context.Background()
	err := database.WithTx(ctx, func(tx *db.Tx) error {
		_, err := manager.WithTx(tx).RawExec(ctx,
			"INSERT INTO tx_test_items (name, value) VALUES ($1, $2)",
			"committed-value", int64(7),
		)
		return err
	})
	require.NoError(t, err)

	rows, err := manager.RawQuery(ctx,
		"SELECT value FROM tx_test_items WHERE name = $1", "committed-value")
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var value int64
	require.NoError(t, rows.Scan(&value))
	require.Equal(t, int64(7), value)
}
