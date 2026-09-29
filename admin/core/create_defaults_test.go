package core

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type createDefaultsItem struct {
	schema.BaseSchema
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	Priority  int32     `db:"priority" json:"priority"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Note      *string   `db:"note" json:"note"`
}

func (createDefaultsItem) Meta() schema.Meta {
	return schema.Meta{TableName: "create_defaults_items"}
}

func (createDefaultsItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.BoolField("is_active", schema.Default(true)),
		schema.Int32Field("priority", schema.Default(5)),
		schema.StringField("status", schema.Required(), schema.Default("draft")),
		schema.DateTimeField("created_at", schema.Default(time.Now)),
		schema.StringField("note", schema.Optional(), schema.Default("none")),
	}
}

func setupCreateDefaultsAdmin(t *testing.T) (*Admin[createDefaultsItem], *orm.Manager[createDefaultsItem]) {
	t.Helper()
	database, err := db.NewDB(filepath.Join(t.TempDir(), "create_defaults.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	_, err = database.Exec(`
		CREATE TABLE create_defaults_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT 0,
			priority INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT '',
			created_at DATETIME,
			note TEXT
		);
	`)
	require.NoError(t, err)

	manager, err := orm.NewManagerWithDB[createDefaultsItem]("create_defaults_items", database)
	require.NoError(t, err)
	admin, err := NewAdmin[createDefaultsItem](createDefaultsItem{}, manager, &Config[createDefaultsItem]{})
	require.NoError(t, err)
	return admin, manager
}

func TestCreateObject_AppliesSchemaDefaultsForOmittedKeys(t *testing.T) {
	admin, manager := setupCreateDefaultsAdmin(t)
	ctx := context.Background()

	// status is required but has a Default, so omitting it is not an error.
	created, err := admin.CreateObject(ctx, map[string]interface{}{"name": "defaults"})
	require.NoError(t, err)
	item := created.(*createDefaultsItem)

	stored, err := manager.Get(ctx, item.ID)
	require.NoError(t, err)
	assert.True(t, stored.IsActive, "omitted Default(true) bool must store true")
	assert.Equal(t, int32(5), stored.Priority)
	assert.Equal(t, "draft", stored.Status)
	assert.False(t, stored.CreatedAt.IsZero(), "callable default must be evaluated")
}

func TestCreateObject_ExplicitZeroOverridesDefault(t *testing.T) {
	admin, manager := setupCreateDefaultsAdmin(t)
	ctx := context.Background()

	created, err := admin.CreateObject(ctx, map[string]interface{}{
		"name":      "explicit",
		"is_active": false,
		"priority":  0,
		"status":    "published",
	})
	require.NoError(t, err)
	item := created.(*createDefaultsItem)

	stored, err := manager.Get(ctx, item.ID)
	require.NoError(t, err)
	assert.False(t, stored.IsActive, "explicit false must not be replaced by the default")
	assert.Equal(t, int32(0), stored.Priority)
	assert.Equal(t, "published", stored.Status)
}

func TestCreateObject_RequiredWithoutDefaultStillRequired(t *testing.T) {
	admin, _ := setupCreateDefaultsAdmin(t)
	_, err := admin.CreateObject(context.Background(), map[string]interface{}{"is_active": true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestBuildFieldsMetadata_DefaultValues(t *testing.T) {
	fields, err := buildFieldsMetadata(createDefaultsItem{})
	require.NoError(t, err)
	byName := map[string]FieldMetadata{}
	for _, f := range fields {
		byName[f.Name] = f
	}
	assert.Equal(t, true, byName["is_active"].DefaultValue)
	assert.Equal(t, "draft", byName["status"].DefaultValue)
	assert.Nil(t, byName["created_at"].DefaultValue, "a callable default has no static value to prefill")
	assert.True(t, byName["created_at"].HasDefault, "a callable default is still a default the create form must not require")
	assert.True(t, byName["status"].HasDefault)
	assert.False(t, byName["name"].HasDefault)

	// The metadata must stay JSON-encodable even with a callable default.
	_, err = json.Marshal(fields)
	require.NoError(t, err)
}

func TestCreateObject_ExplicitNullOverridesDefault(t *testing.T) {
	admin, manager := setupCreateDefaultsAdmin(t)
	ctx := context.Background()

	created, err := admin.CreateObject(ctx, map[string]interface{}{"name": "omitted"})
	require.NoError(t, err)
	omitted, err := manager.Get(ctx, created.(*createDefaultsItem).ID)
	require.NoError(t, err)
	require.NotNil(t, omitted.Note, "an omitted key gets the schema default")
	assert.Equal(t, "none", *omitted.Note)

	created, err = admin.CreateObject(ctx, map[string]interface{}{"name": "null", "note": nil})
	require.NoError(t, err)
	null, err := manager.Get(ctx, created.(*createDefaultsItem).ID)
	require.NoError(t, err)
	assert.Nil(t, null.Note, "an explicit null is stored as NULL, not replaced by the default")
}
