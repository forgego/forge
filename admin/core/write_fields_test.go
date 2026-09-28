package core

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	validation "github.com/forgego/forge/validate"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type writeFieldsTestItem struct {
	schema.BaseSchema
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Notes     string    `db:"notes" json:"notes"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func (writeFieldsTestItem) Meta() schema.Meta {
	return schema.Meta{
		TableName: "write_fields_test_items",
	}
}

func (writeFieldsTestItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.StringField("email"),
		schema.StringField("notes"),
		schema.DateTimeField("created_at", schema.AutoNowAdd()),
	}
}

func setupWriteFieldsDB(t *testing.T, cfg *Config[writeFieldsTestItem]) (*Admin[writeFieldsTestItem], *orm.Manager[writeFieldsTestItem], int64, time.Time) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "write_fields_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = database.Close()
	})

	_, err = database.Exec(`
		CREATE TABLE write_fields_test_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			notes TEXT,
			created_at DATETIME
		);
	`)
	require.NoError(t, err)

	initialTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	res, err := database.Exec(`INSERT INTO write_fields_test_items (name, email, notes, created_at) VALUES (?, ?, ?, ?)`,
		"Original Name", "orig@example.com", "orig note", initialTime)
	require.NoError(t, err)

	id, err := res.LastInsertId()
	require.NoError(t, err)

	manager, err := orm.NewManagerWithDB[writeFieldsTestItem]("write_fields_test_items", database)
	require.NoError(t, err)

	if cfg == nil {
		cfg = &Config[writeFieldsTestItem]{}
	}

	admin, err := NewAdmin[writeFieldsTestItem](writeFieldsTestItem{}, manager, cfg)
	require.NoError(t, err)

	return admin, manager, id, initialTime
}

func TestWriteFields_UpdateValidation_UnknownField(t *testing.T) {
	admin, _, id, _ := setupWriteFieldsDB(t, nil)
	ctx := context.Background()

	payload := map[string]interface{}{
		"name":       "new",
		"created_at": "2026-06-01T00:00:00Z",
		"unknown":    1,
	}

	_, err := admin.UpdateObject(ctx, id, payload)
	require.Error(t, err)

	var valErrs *validation.ValidationErrors
	require.True(t, errors.As(err, &valErrs), "expected *validation.ValidationErrors, got %T: %v", err, err)
	assert.Contains(t, err.Error(), "unknown")
}

func TestWriteFields_Update_AutoManagedUnchanged(t *testing.T) {
	admin, manager, id, initialTime := setupWriteFieldsDB(t, nil)
	ctx := context.Background()

	payload := map[string]interface{}{
		"name":       "new",
		"created_at": "2026-06-01T00:00:00Z",
	}

	_, err := admin.UpdateObject(ctx, id, payload)
	require.NoError(t, err)

	updated, err := manager.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "new", updated.Name)
	assert.Equal(t, initialTime.Unix(), updated.CreatedAt.Unix(), "auto-managed created_at must remain unchanged in DB")
}

func TestWriteFields_Update_ReadOnlyIgnored(t *testing.T) {
	cfg := &Config[writeFieldsTestItem]{
		ReadOnlyFields: []string{"email"},
	}
	admin, manager, id, _ := setupWriteFieldsDB(t, cfg)
	ctx := context.Background()

	payload := map[string]interface{}{
		"name":  "updated name",
		"email": "new@example.com",
	}

	_, err := admin.UpdateObject(ctx, id, payload)
	require.NoError(t, err)

	updated, err := manager.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "updated name", updated.Name)
	assert.Equal(t, "orig@example.com", updated.Email, "read-only email must be ignored and not updated")
}

func TestWriteFields_Create_ExcludeIgnored(t *testing.T) {
	cfg := &Config[writeFieldsTestItem]{
		Exclude: []string{"notes"},
	}
	admin, manager, _, _ := setupWriteFieldsDB(t, cfg)
	ctx := context.Background()

	payload := map[string]interface{}{
		"name":  "Item with excluded notes",
		"notes": "secret note",
	}

	createdObj, err := admin.CreateObject(ctx, payload)
	require.NoError(t, err)

	created, ok := createdObj.(*writeFieldsTestItem)
	require.True(t, ok)
	assert.Equal(t, "Item with excluded notes", created.Name)
	assert.Empty(t, created.Notes, "excluded notes must not be set on created instance")

	fetched, err := manager.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Item with excluded notes", fetched.Name)
	assert.Empty(t, fetched.Notes, "excluded notes must not be saved to DB")
}

func TestWriteFields_GetReadOnlyFields_Precedence(t *testing.T) {
	cfg := &Config[writeFieldsTestItem]{
		ReadOnlyFields: []string{"email"},
		GetReadOnlyFields: func(ctx context.Context, instance *writeFieldsTestItem, isNew bool) []string {
			return []string{"notes"}
		},
	}
	admin, manager, id, _ := setupWriteFieldsDB(t, cfg)
	ctx := context.Background()

	payload := map[string]interface{}{
		"name":  "precedence test",
		"email": "updated_email@example.com",
		"notes": "hacked note",
	}

	_, err := admin.UpdateObject(ctx, id, payload)
	require.NoError(t, err)

	updated, err := manager.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "precedence test", updated.Name)
	assert.Equal(t, "updated_email@example.com", updated.Email, "email should be writable because GetReadOnlyFields overrides ReadOnlyFields")
	assert.Equal(t, "orig note", updated.Notes, "notes should be ignored because GetReadOnlyFields declared it read-only")
}

func TestWriteFields_WritableFields_Unit(t *testing.T) {
	ctx := context.Background()

	s := testSchema{
		fields: []schema.Field{
			{Name: "id", Type: schema.TypeInt64, PrimaryKey: true, AutoIncrement: true, Editable: true},
			{Name: "title", Type: schema.TypeString, Editable: true},
			{Name: "content", Type: schema.TypeString, Editable: true},
			{Name: "author", Type: schema.TypeString, Editable: true},
			{Name: "secret_notes", Type: schema.TypeString, Editable: false},
			{Name: "slug", Type: schema.TypeString, Generated: true, Editable: true},
			{Name: "created_at", Type: schema.TypeDateTime, AutoNowAdd: true, Editable: true},
			{Name: "updated_at", Type: schema.TypeDateTime, AutoNow: true, Editable: true},
		},
	}

	tests := []struct {
		name     string
		admin    *Admin[struct{}]
		isNew    bool
		expected map[string]bool
	}{
		{
			name:     "nil admin returns nil",
			admin:    nil,
			isNew:    false,
			expected: nil,
		},
		{
			name:     "nil schema returns nil",
			admin:    &Admin[struct{}]{schema: nil, config: &Config[struct{}]{}},
			isNew:    false,
			expected: nil,
		},
		{
			name:     "nil config returns all editable non-auto schema fields",
			admin:    &Admin[struct{}]{schema: s, config: nil},
			isNew:    false,
			expected: map[string]bool{"title": true, "content": true, "author": true},
		},
		{
			name:     "empty config returns all editable non-auto schema fields",
			admin:    &Admin[struct{}]{schema: s, config: &Config[struct{}]{}},
			isNew:    false,
			expected: map[string]bool{"title": true, "content": true, "author": true},
		},
		{
			name: "explicit Fields restricts start set",
			admin: &Admin[struct{}]{
				schema: s,
				config: &Config[struct{}]{
					Fields: []string{"title", "content", "created_at"},
				},
			},
			isNew:    false,
			expected: map[string]bool{"title": true, "content": true},
		},
		{
			name: "GetFields callback overrides Fields",
			admin: &Admin[struct{}]{
				schema: s,
				config: &Config[struct{}]{
					Fields: []string{"title"},
					GetFields: func(ctx context.Context, instance *struct{}, isNew bool) []string {
						return []string{"title", "author"}
					},
				},
			},
			isNew:    false,
			expected: map[string]bool{"title": true, "author": true},
		},
		{
			name: "Exclude removes fields",
			admin: &Admin[struct{}]{
				schema: s,
				config: &Config[struct{}]{
					Exclude: []string{"author"},
				},
			},
			isNew:    false,
			expected: map[string]bool{"title": true, "content": true},
		},
		{
			name: "ReadOnlyFields removes fields",
			admin: &Admin[struct{}]{
				schema: s,
				config: &Config[struct{}]{
					ReadOnlyFields: []string{"content"},
				},
			},
			isNew:    false,
			expected: map[string]bool{"title": true, "author": true},
		},
		{
			name: "GetReadOnlyFields callback overrides ReadOnlyFields",
			admin: &Admin[struct{}]{
				schema: s,
				config: &Config[struct{}]{
					ReadOnlyFields: []string{"title"},
					GetReadOnlyFields: func(ctx context.Context, instance *struct{}, isNew bool) []string {
						return []string{"author"}
					},
				},
			},
			isNew:    false,
			expected: map[string]bool{"title": true, "content": true},
		},
		{
			name: "Exclude and ReadOnly both remove fields",
			admin: &Admin[struct{}]{
				schema: s,
				config: &Config[struct{}]{
					Exclude:        []string{"author"},
					ReadOnlyFields: []string{"content"},
				},
			},
			isNew:    false,
			expected: map[string]bool{"title": true},
		},
		{
			name: "auto-managed and non-editable fields removed even if in Fields",
			admin: &Admin[struct{}]{
				schema: s,
				config: &Config[struct{}]{
					Fields: []string{"id", "secret_notes", "slug", "created_at", "updated_at"},
				},
			},
			isNew:    false,
			expected: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.admin.writableFields(ctx, nil, tt.isNew)
			assert.Equal(t, tt.expected, got)
		})
	}
}
