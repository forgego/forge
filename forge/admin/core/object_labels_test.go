package core

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type labelTestItem struct {
	schema.BaseSchema
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (labelTestItem) Meta() schema.Meta {
	return schema.Meta{
		TableName: "label_test_items",
	}
}

func (labelTestItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
	}
}

func TestAdmin_ObjectLabels(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "object_labels_test.sqlite")

	database, err := db.NewDB(dbPath)
	require.NoError(t, err)
	defer database.Close()

	_, err = database.Exec(`
		CREATE TABLE label_test_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	res1, err := database.Exec(`INSERT INTO label_test_items (name) VALUES (?)`, "Item One")
	require.NoError(t, err)
	id1, err := res1.LastInsertId()
	require.NoError(t, err)

	res2, err := database.Exec(`INSERT INTO label_test_items (name) VALUES (?)`, "Item Two")
	require.NoError(t, err)
	id2, err := res2.LastInsertId()
	require.NoError(t, err)

	res3, err := database.Exec(`INSERT INTO label_test_items (name) VALUES (?)`, "Item Three")
	require.NoError(t, err)
	id3, err := res3.LastInsertId()
	require.NoError(t, err)

	manager, err := orm.NewManagerWithDB[labelTestItem]("label_test_items", database)
	require.NoError(t, err)

	admin, err := NewAdmin[labelTestItem](labelTestItem{}, manager, &Config[labelTestItem]{})
	require.NoError(t, err)

	// Ensure *Admin[T] satisfies LabelResolver
	var resolver LabelResolver = admin

	t.Run("empty ids returns empty map", func(t *testing.T) {
		emptyNil, err := resolver.ObjectLabels(ctx, nil)
		require.NoError(t, err)
		assert.Empty(t, emptyNil)

		emptySlice, err := resolver.ObjectLabels(ctx, []interface{}{})
		require.NoError(t, err)
		assert.Empty(t, emptySlice)
	})

	t.Run("resolves exactly requested object ids", func(t *testing.T) {
		labels, err := resolver.ObjectLabels(ctx, []interface{}{id1, id3})
		require.NoError(t, err)
		require.Len(t, labels, 2)
		assert.Equal(t, "Item One", labels[fmt.Sprint(id1)])
		assert.Equal(t, "Item Three", labels[fmt.Sprint(id3)])

		_, hasID2 := labels[fmt.Sprint(id2)]
		assert.False(t, hasID2)
	})

	t.Run("resolves float64 ids from json unmarshaling", func(t *testing.T) {
		labels, err := resolver.ObjectLabels(ctx, []interface{}{float64(id1), float64(id2)})
		require.NoError(t, err)
		require.Len(t, labels, 2)
		assert.Equal(t, "Item One", labels[fmt.Sprint(id1)])
		assert.Equal(t, "Item Two", labels[fmt.Sprint(id2)])
	})

	t.Run("caps resolution at 1000 ids", func(t *testing.T) {
		ids := make([]interface{}, 1005)
		ids[0] = id1
		for i := 1; i < 1000; i++ {
			ids[i] = int64(9000 + i)
		}
		// id2 is placed beyond the 1000 cap limit
		ids[1000] = id2
		for i := 1001; i < 1005; i++ {
			ids[i] = int64(9000 + i)
		}

		labels, err := resolver.ObjectLabels(ctx, ids)
		require.NoError(t, err)
		assert.Equal(t, "Item One", labels[fmt.Sprint(id1)])
		_, hasID2 := labels[fmt.Sprint(id2)]
		assert.False(t, hasID2, "id beyond 1000 cap should not be resolved")
	})
}
