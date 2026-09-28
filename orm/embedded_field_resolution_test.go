package orm

import (
	"testing"

	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ORMResolutionGenerated struct {
	ID    int64  `json:"id" db:"item_id"`
	Value string `json:"displayValue" db:"value_col"`
}

type ormResolutionModel struct {
	schema.BaseSchema
	ORMResolutionGenerated
}

func (ormResolutionModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement(), schema.DBColumn("item_id")),
		schema.StringField("value", schema.DBColumn("value_col")),
	}
}

func (ormResolutionModel) Meta() schema.Meta {
	return schema.Meta{TableName: "resolved_items"}
}

func TestBuildModelSchema_ResolvesFieldsInAnonymousEmbedding(t *testing.T) {
	modelSchema, err := BuildModelSchema(&ormResolutionModel{})
	require.NoError(t, err)
	valueField := modelSchema.GetField("displayValue")
	require.NotNil(t, valueField)
	assert.Equal(t, "Value", valueField.StructFieldName)
	assert.Equal(t, "value_col", valueField.DBColumn)
}

func TestBuildInsertSQL_ReadsSchemaFieldFromAnonymousEmbedding(t *testing.T) {
	model := &ormResolutionModel{ORMResolutionGenerated: ORMResolutionGenerated{Value: "embedded"}}
	query, args, columns, err := BuildInsertSQL(model, "resolved_items")
	require.NoError(t, err)
	assert.Contains(t, query, `"value_col"`)
	assert.Equal(t, []interface{}{"embedded"}, args)
	assert.Equal(t, []string{"value_col"}, columns)
}
