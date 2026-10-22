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

type bulkUpdateTestItem struct {
	schema.BaseSchema
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Status string `db:"status"`
}

func (bulkUpdateTestItem) Meta() schema.Meta {
	return schema.Meta{TableName: "bulk_update_test_items"}
}

func (bulkUpdateTestItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
		schema.StringField("status"),
	}
}

func (bulkUpdateTestItem) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (bulkUpdateTestItem) Hooks() *schema.ModelHooks {
	return nil
}

// BulkUpdate issues one update per entry
func TestBulkUpdateIssuesOneUpdatePerEntry(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "bulk_update_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)
	defer database.Close()

	_, err = database.Exec(`
		CREATE TABLE bulk_update_test_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			status TEXT NOT NULL
		);
		INSERT INTO bulk_update_test_items (name, status) VALUES ('item1', 'pending');
		INSERT INTO bulk_update_test_items (name, status) VALUES ('item2', 'pending');
	`)
	require.NoError(t, err)

	qs, err := NewQuerySet[bulkUpdateTestItem]("bulk_update_test_items")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	updates := []UpdateMap{
		{"status": "in_progress"},
		{"status": "completed"},
	}

	err = qs.BulkUpdate(ctx, updates)
	require.NoError(t, err)

	// Verify the sequential updates were applied
	rows, err := database.Query("SELECT status FROM bulk_update_test_items ORDER BY id")
	require.NoError(t, err)
	defer rows.Close()

	var statuses []string
	for rows.Next() {
		var s string
		require.NoError(t, rows.Scan(&s))
		statuses = append(statuses, s)
	}
	require.Len(t, statuses, 2)
	assert.Equal(t, "completed", statuses[0])
	assert.Equal(t, "completed", statuses[1])
}
