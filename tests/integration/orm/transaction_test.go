package orm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	"github.com/forgego/forge/tests/testhelpers"
	"github.com/stretchr/testify/require"
)

type transactionRecord struct {
	schema.BaseSchema
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Value int64  `db:"value"`
}

func (transactionRecord) Meta() schema.Meta { return schema.Meta{TableName: "forge_tx_raw_records"} }

func (transactionRecord) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
		schema.Int64Field("value"),
	}
}

func TestManagerWithTxRawQueriesOnPostgres(t *testing.T) {
	dsn := testhelpers.RequirePostgresURL(t)
	database, err := db.NewDBWithDriver("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })
	_, err = database.Exec(`CREATE TABLE forge_tx_raw_records (
		id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, value BIGINT NOT NULL
	)`)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = database.Exec(`DROP TABLE IF EXISTS forge_tx_raw_records`) })

	manager, err := orm.NewManagerWithDB[transactionRecord]("forge_tx_raw_records", database)
	require.NoError(t, err)
	ctx := context.Background()
	rollback := errors.New("rollback postgres raw transaction")
	err = database.WithTx(ctx, func(tx *db.Tx) error {
		txManager := manager.WithTx(tx)
		_, err := txManager.RawExec(ctx,
			"INSERT INTO forge_tx_raw_records (name, value) VALUES ($1, $2)",
			"postgres-value", int64(10),
		)
		if err != nil {
			return err
		}
		rows, err := txManager.RawQuery(ctx,
			"SELECT name FROM forge_tx_raw_records WHERE name = $1", "postgres-value")
		if err != nil {
			return err
		}
		defer rows.Close()
		require.True(t, rows.Next())
		var name string
		require.NoError(t, rows.Scan(&name))
		require.Equal(t, "postgres-value", name)
		if err := rows.Close(); err != nil {
			return err
		}

		outside, err := manager.RawQuery(ctx,
			"SELECT COUNT(*) FROM forge_tx_raw_records WHERE name = $1", "postgres-value")
		if err != nil {
			return err
		}
		defer outside.Close()
		require.True(t, outside.Next())
		var outsideCount int
		require.NoError(t, outside.Scan(&outsideCount))
		require.Zero(t, outsideCount)

		qs := txManager.QuerySet().Filter(orm.F("name").Eq("postgres-value"))
		affected, err := qs.Update(ctx, orm.UpdateMap{"value": int64(20)})
		require.NoError(t, err)
		require.Equal(t, int64(1), affected)
		items, err := qs.All(ctx)
		require.NoError(t, err)
		require.Equal(t, int64(20), items[0].Value)
		return rollback
	})
	require.ErrorIs(t, err, rollback)
	count, err := manager.Count(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
}
