package core

import (
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

