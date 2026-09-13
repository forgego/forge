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

type ExistsTestItem struct {
	schema.BaseSchema
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Active bool   `db:"active"`
}

func (ExistsTestItem) Meta() schema.Meta {
	return schema.Meta{TableName: "exists_test_items"}
}

func (ExistsTestItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
		schema.BoolField("active"),
	}
}

func setupExistsDB(t *testing.T) *db.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "exists_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)

	_, err = database.Exec(`
		CREATE TABLE exists_test_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			active INTEGER NOT NULL
		);
		INSERT INTO exists_test_items (name, active) VALUES ('item1', 1);
		INSERT INTO exists_test_items (name, active) VALUES ('item2', 0);
	`)
	require.NoError(t, err)

	_, err = GetModelSchema[ExistsTestItem]()
	require.NoError(t, err)

	return database
}

func TestQuerySet_Exists_Filtered(t *testing.T) {
	database := setupExistsDB(t)
	ctx := context.Background()

	tests := []struct {
		name       string
		filterExpr Expression
		wantExists bool
	}{
		{
			name:       "matching active item exists",
			filterExpr: F("active").Eq(true),
			wantExists: true,
		},
		{
			name:       "matching item by name exists",
			filterExpr: F("name").Eq("item2"),
			wantExists: true,
		},
		{
			name:       "non-matching item does not exist",
			filterExpr: F("name").Eq("nonexistent"),
			wantExists: false,
		},
		{
			name:       "non-matching active filter does not exist",
			filterExpr: And(F("name").Eq("item2"), F("active").Eq(true)),
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs, err := NewQuerySet[ExistsTestItem]("exists_test_items")
			require.NoError(t, err)
			qs = qs.SetDB(database).Filter(tt.filterExpr)

			exists, err := qs.Exists(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.wantExists, exists)
		})
	}
}

func TestQuerySet_Exists_GeneratedSQL(t *testing.T) {
	qs, err := NewQuerySet[ExistsTestItem]("exists_test_items")
	require.NoError(t, err)
	filtered := qs.Filter(F("name").Eq("item1"))

	baseQS, ok := filtered.(*BaseQuerySet[ExistsTestItem])
	require.True(t, ok)

	sql, _, err := baseQS.buildExistsSQL()
	require.NoError(t, err)

	assert.Contains(t, sql, "LIMIT 1")
	assert.NotContains(t, sql, "COUNT")
}
