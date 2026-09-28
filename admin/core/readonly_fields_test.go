package core

import (
	"testing"

	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildFieldsMetadata_AutoManagedAndReadOnlyFields(t *testing.T) {
	s := testSchema{
		fields: []schema.Field{
			{Name: "name", Type: schema.TypeString, Editable: true, Required: true},
			{Name: "created_at", Type: schema.TypeDateTime, AutoNowAdd: true, Editable: true, Required: true},
			{Name: "updated_at", Type: schema.TypeDateTime, AutoNow: true, Editable: true, Required: true},
			{Name: "id", Type: schema.TypeInt64, PrimaryKey: true, AutoIncrement: true, Editable: true, Required: true},
			{Name: "slug", Type: schema.TypeString, Generated: true, Editable: true, Required: true},
			{Name: "sku", Type: schema.TypeString, PrimaryKey: true, AutoIncrement: false, Editable: true, Required: true},
			{Name: "notes", Type: schema.TypeString, Editable: false, Required: false},
		},
	}

	fieldsMeta, err := buildFieldsMetadata(s)
	require.NoError(t, err)
	require.Len(t, fieldsMeta, 7)

	fieldMap := make(map[string]FieldMetadata, len(fieldsMeta))
	for _, f := range fieldsMeta {
		fieldMap[f.Name] = f
	}

	// name: string, Editable true -> read_only false, required as declared
	nameField := fieldMap["name"]
	assert.False(t, nameField.ReadOnly)
	assert.True(t, nameField.Required)

	// created_at: AutoNowAdd, Editable true -> read_only true, required false
	createdAtField := fieldMap["created_at"]
	assert.True(t, createdAtField.ReadOnly)
	assert.False(t, createdAtField.Required)

	// updated_at: AutoNow, Editable true -> read_only true, required false
	updatedAtField := fieldMap["updated_at"]
	assert.True(t, updatedAtField.ReadOnly)
	assert.False(t, updatedAtField.Required)

	// id: PrimaryKey + AutoIncrement, Editable true -> read_only true, required false
	idField := fieldMap["id"]
	assert.True(t, idField.ReadOnly)
	assert.False(t, idField.Required)

	// slug: Generated -> read_only true, required false
	slugField := fieldMap["slug"]
	assert.True(t, slugField.ReadOnly)
	assert.False(t, slugField.Required)

	// sku: PrimaryKey without AutoIncrement, Editable -> read_only false (user-assigned key)
	skuField := fieldMap["sku"]
	assert.False(t, skuField.ReadOnly)
	assert.True(t, skuField.Required)

	// notes: Editable false -> read_only true (unchanged behavior)
	notesField := fieldMap["notes"]
	assert.True(t, notesField.ReadOnly)
	assert.False(t, notesField.Required)

	// Also verify via buildMetadata end-to-end
	meta, err := buildMetadata[struct{}](s, &Config[struct{}]{}, "Item")
	require.NoError(t, err)
	require.Len(t, meta.Fields, 7)
	metaFieldMap := make(map[string]FieldMetadata, len(meta.Fields))
	for _, f := range meta.Fields {
		metaFieldMap[f.Name] = f
	}
	assert.False(t, metaFieldMap["name"].ReadOnly)
	assert.True(t, metaFieldMap["name"].Required)
	assert.True(t, metaFieldMap["created_at"].ReadOnly)
	assert.False(t, metaFieldMap["created_at"].Required)
	assert.True(t, metaFieldMap["updated_at"].ReadOnly)
	assert.False(t, metaFieldMap["updated_at"].Required)
	assert.True(t, metaFieldMap["id"].ReadOnly)
	assert.False(t, metaFieldMap["id"].Required)
	assert.True(t, metaFieldMap["slug"].ReadOnly)
	assert.False(t, metaFieldMap["slug"].Required)
	assert.False(t, metaFieldMap["sku"].ReadOnly)
	assert.True(t, metaFieldMap["sku"].Required)
	assert.True(t, metaFieldMap["notes"].ReadOnly)
	assert.False(t, metaFieldMap["notes"].Required)
}

func TestIsAutoManaged(t *testing.T) {
	tests := []struct {
		name     string
		field    schema.Field
		expected bool
	}{
		{
			name:     "regular field",
			field:    schema.Field{Name: "title", Type: schema.TypeString},
			expected: false,
		},
		{
			name:     "auto now field",
			field:    schema.Field{Name: "updated_at", AutoNow: true},
			expected: true,
		},
		{
			name:     "auto now add field",
			field:    schema.Field{Name: "created_at", AutoNowAdd: true},
			expected: true,
		},
		{
			name:     "generated field",
			field:    schema.Field{Name: "slug", Generated: true},
			expected: true,
		},
		{
			name:     "primary key with auto increment",
			field:    schema.Field{Name: "id", PrimaryKey: true, AutoIncrement: true},
			expected: true,
		},
		{
			name:     "primary key without auto increment (user assigned)",
			field:    schema.Field{Name: "sku", PrimaryKey: true, AutoIncrement: false},
			expected: false,
		},
		{
			name:     "auto increment without primary key",
			field:    schema.Field{Name: "seq", PrimaryKey: false, AutoIncrement: true},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isAutoManaged(tt.field))
		})
	}
}
