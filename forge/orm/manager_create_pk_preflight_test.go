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

type missingPrimaryKeyFieldModel struct {
	schema.BaseSchema
	Name string `db:"name"`
}

func (missingPrimaryKeyFieldModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

func (missingPrimaryKeyFieldModel) Meta() schema.Meta {
	return schema.Meta{TableName: "missing_primary_key_field_models"}
}

func TestManager_Create_MissingConcretePrimaryKeyFieldDoesNotInsert(t *testing.T) {
	database, err := db.NewDB(filepath.Join(t.TempDir(), "missing-pk.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })
	_, err = database.Exec(`CREATE TABLE missing_primary_key_field_models (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT
	)`)
	require.NoError(t, err)

	manager, err := NewManagerWithDB[missingPrimaryKeyFieldModel]("missing_primary_key_field_models", database)
	require.NoError(t, err)
	err = manager.Create(context.Background(), &missingPrimaryKeyFieldModel{Name: "must not persist"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set primary key")

	var count int
	require.NoError(t, database.QueryRow("SELECT COUNT(*) FROM missing_primary_key_field_models").Scan(&count))
	assert.Zero(t, count)
}
