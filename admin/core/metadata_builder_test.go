package core

import (
	"encoding/json"
	"testing"

	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSchema struct {
	schema.BaseSchema
	fields    []schema.Field
	relations []schema.Relation
}

func (s testSchema) Fields() []schema.Field {
	return s.fields
}

func (s testSchema) Relations() []schema.Relation {
	return s.relations
}

func TestBuildFiltersMetadata_RelatedModelForRelationFilters(t *testing.T) {
	s := testSchema{
		fields: []schema.Field{
			{Name: "category_id", Type: schema.TypeForeignKey, Editable: true},
			{Name: "tags", Type: schema.TypeManyToMany, Editable: true},
			{Name: "is_active", Type: schema.TypeBool, Editable: true},
		},
		relations: []schema.Relation{
			{Name: "category", To: "Category", Type: schema.RelationForeignKey},
			{Name: "tags", To: "Tag", Type: schema.RelationManyToMany},
		},
	}

	config := &Config[struct{}]{
		ListFilter: []Field{
			Computed("category_id"),
			Computed("tags"),
			Computed("is_active"),
		},
	}

	filters := buildFiltersMetadata(s, config)
	require.Len(t, filters, 3)

	assert.Equal(t, "Category", filters[0].RelatedModel)
	assert.False(t, filters[0].Multiple)

	assert.Equal(t, "Tag", filters[1].RelatedModel)
	assert.True(t, filters[1].Multiple)

	assert.Empty(t, filters[2].RelatedModel)
	assert.False(t, filters[2].Multiple)
}

type testCustomField struct {
	name string
}

func (f testCustomField) GetName() string                  { return f.name }
func (f testCustomField) GetType() schema.FieldType        { return schema.TypeBytes }
func (f testCustomField) Validate(value interface{}) error { return nil }

func TestBuildFieldsMetadata_UploadSupported(t *testing.T) {
	s := testSchema{
		fields: []schema.Field{
			{Name: "document", CustomField: testCustomField{name: "file"}},
			{Name: "avatar", DBType: "file"},
			{Name: "photo", DBType: "image"},
			{Name: "title", Type: schema.TypeString},
		},
	}

	fieldsMeta, err := buildFieldsMetadata(s)
	require.NoError(t, err)
	require.Len(t, fieldsMeta, 4)

	// File and image fields have upload_supported = false
	for _, idx := range []int{0, 1, 2} {
		require.NotNil(t, fieldsMeta[idx].UploadSupported, "field %s should have UploadSupported", fieldsMeta[idx].Name)
		assert.False(t, *fieldsMeta[idx].UploadSupported, "field %s should have upload_supported = false", fieldsMeta[idx].Name)

		data, err := json.Marshal(fieldsMeta[idx])
		require.NoError(t, err)
		assert.Contains(t, string(data), `"upload_supported":false`)
	}

	// Char field has none (nil in struct, omitted in JSON)
	charField := fieldsMeta[3]
	assert.Nil(t, charField.UploadSupported, "char field should not have UploadSupported")

	charData, err := json.Marshal(charField)
	require.NoError(t, err)
	assert.NotContains(t, string(charData), "upload_supported")
}
