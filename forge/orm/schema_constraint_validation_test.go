package orm_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	_ "github.com/forgego/forge/validate"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type requiredBeforeSaveHookModel struct {
	schema.BaseSchema
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func (requiredBeforeSaveHookModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
	}
}

func (requiredBeforeSaveHookModel) Meta() schema.Meta {
	return schema.Meta{TableName: "required_before_save_hook_models"}
}

func (m *requiredBeforeSaveHookModel) BeforeSave(context.Context) error {
	m.Name = ""
	return nil
}

func TestManager_MarkedValidationRejectsSchemaViolationAfterBeforeSave(t *testing.T) {
	database, err := db.NewDB(filepath.Join(t.TempDir(), "schema_constraint_validation.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })

	_, err = database.Exec(`
		CREATE TABLE required_before_save_hook_models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT
		)
	`)
	require.NoError(t, err)

	manager, err := orm.NewManagerWithDB[requiredBeforeSaveHookModel]("required_before_save_hook_models", database)
	require.NoError(t, err)
	err = manager.Create(
		orm.WithModelValidationCompleted(context.Background()),
		&requiredBeforeSaveHookModel{Name: "valid before hook"},
	)
	require.ErrorContains(t, err, "name: is required")

	var count int
	require.NoError(t, database.QueryRow("SELECT COUNT(*) FROM required_before_save_hook_models").Scan(&count))
	assert.Zero(t, count)
}
