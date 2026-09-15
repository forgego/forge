package orm

import (
	"reflect"
	"testing"

	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TargetTagModel is the target model whose Go type name is "TargetTagModel".
type TargetTagModel struct {
	schema.BaseSchema
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (TargetTagModel) Meta() schema.Meta {
	return schema.Meta{TableName: "target_tag_models"}
}

func (TargetTagModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary()),
		schema.StringField("name"),
	}
}

// SourceArticleModel defines many-to-many relations:
// 1. "TagAlias": registered alias pointing to TargetTagModel (Go type name differs from alias).
// 2. "UnregisteredCategoryAlias": unregistered alias that should fall back to the alias name.
type SourceArticleModel struct {
	schema.BaseSchema
	ID    int64  `db:"id"`
	Title string `db:"title"`
}

func (SourceArticleModel) Meta() schema.Meta {
	return schema.Meta{TableName: "source_article_models"}
}

func (SourceArticleModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary()),
		schema.StringField("title"),
	}
}

func (SourceArticleModel) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ManyToManyField("Tags", "TagAlias", schema.Through("article_tags")),
		schema.ManyToManyField("Categories", "UnregisteredCategoryAlias", schema.Through("article_categories")),
	}
}

func TestSchema_ManyToManyThroughColumn_RegisteredAliasAndFallback(t *testing.T) {
	alias := "TagAlias"
	targetType := reflect.TypeOf(TargetTagModel{})
	sourceType := reflect.TypeOf(SourceArticleModel{})

	// Verify prerequisite: Go type name differs from the alias
	require.NotEqual(t, alias, targetType.Name(), "Go type name must differ from alias")

	// Clean up any existing entries before and after the test
	cleanup := func() {
		schemaNameMu.Lock()
		delete(schemaNameRegistry, alias)
		delete(schemaNameRegistry, targetType.Name())
		delete(schemaNameRegistry, sourceType.Name())
		schemaNameMu.Unlock()

		schemaMu.Lock()
		delete(schemaCache, targetType)
		delete(schemaCache, sourceType)
		schemaMu.Unlock()
	}
	cleanup()
	t.Cleanup(cleanup)

	// Register the target model under the alias
	RegisterModelType(alias, targetType)

	// Verify GetRegisteredTypeName resolves the alias to the target model's Go type name
	registeredName, ok := GetRegisteredTypeName(alias)
	require.True(t, ok, "expected alias to be registered")
	assert.Equal(t, targetType.Name(), registeredName)

	// Build the source schema
	sourceSchema, err := BuildModelSchema(SourceArticleModel{})
	require.NoError(t, err)
	require.NotNil(t, sourceSchema)

	// Find the relations
	var tagRel *RelationInfo
	var catRel *RelationInfo
	for i := range sourceSchema.Relations {
		if sourceSchema.Relations[i].Name == "Tags" {
			tagRel = &sourceSchema.Relations[i]
		}
		if sourceSchema.Relations[i].Name == "Categories" {
			catRel = &sourceSchema.Relations[i]
		}
	}
	require.NotNil(t, tagRel, "Tags relation must exist")
	require.NotNil(t, catRel, "Categories relation must exist")

	// 1. Assert ThroughTargetColumn is the snake_case of the registered type name + "_id"
	// TargetTagModel -> target_tag_model + "_id" = "target_tag_model_id"
	assert.Equal(t, "target_tag_model_id", tagRel.ThroughTargetColumn)
	assert.Equal(t, "source_article_model_id", tagRel.ThroughSourceColumn)

	// 2. Assert the unregistered case falls back to the alias
	// UnregisteredCategoryAlias -> unregistered_category_alias + "_id" = "unregistered_category_alias_id"
	assert.Equal(t, "unregistered_category_alias_id", catRel.ThroughTargetColumn)
	assert.Equal(t, "source_article_model_id", catRel.ThroughSourceColumn)
}
