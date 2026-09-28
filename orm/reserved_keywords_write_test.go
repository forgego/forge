package orm

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type OrderWithReservedKeywords struct {
	schema.BaseSchema
	ID    int64  `db:"id"`
	User  string `db:"user"`
	Group string `db:"group"`
}

func (OrderWithReservedKeywords) Meta() schema.Meta {
	return schema.Meta{TableName: "order"}
}

func (OrderWithReservedKeywords) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("user"),
		schema.StringField("group"),
	}
}

func setupReservedKeywordsDB(t *testing.T) *db.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "reserved_keywords_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)

	_, err = database.Exec(`
		CREATE TABLE "order" (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"user" TEXT NOT NULL,
			"group" TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	_, err = GetModelSchema[OrderWithReservedKeywords]()
	require.NoError(t, err)

	return database
}

func TestReservedKeywords_CreateUpdateDelete(t *testing.T) {
	database := setupReservedKeywordsDB(t)
	ctx := context.Background()

	mgr, err := NewManager[OrderWithReservedKeywords]("order")
	require.NoError(t, err)
	mgr.SetDB(database)

	order := &OrderWithReservedKeywords{
		User:  "alice",
		Group: "admins",
	}

	err = mgr.Create(ctx, order)
	require.NoError(t, err)
	assert.Equal(t, int64(1), order.ID)

	order.Group = "superusers"
	err = mgr.Update(ctx, order)
	require.NoError(t, err)

	var gotGroup string
	err = database.QueryRow(`SELECT "group" FROM "order" WHERE "id" = ?`, order.ID).Scan(&gotGroup)
	require.NoError(t, err)
	assert.Equal(t, "superusers", gotGroup)

	err = mgr.Delete(ctx, order)
	require.NoError(t, err)

	var count int
	err = database.QueryRow(`SELECT COUNT(*) FROM "order" WHERE "id" = ?`, order.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
