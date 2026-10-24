package schema_test

import (
	"testing"

	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type opOnlyMetaModel struct {
	schema.BaseSchema
	ID     int64  `db:"id" json:"id"`
	Status string `db:"status" json:"status"`
}

func (opOnlyMetaModel) Meta() schema.Meta {
	return schema.Meta{
		TableName:           "orders",
		OperationOnlyFields: []string{"status"},
	}
}

func (opOnlyMetaModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary()),
		schema.StringField("status"),
	}
}

type opOnlyFieldTraitModel struct {
	schema.BaseSchema
	ID     int64  `db:"id" json:"id"`
	State  string `db:"state_col" json:"state_field"`
}

func (opOnlyFieldTraitModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary()),
		schema.StringField("state", schema.DBColumn("state_col"), schema.OperationOnly()),
	}
}

type opOnlyProviderModel struct {
	schema.BaseSchema
	ID    int64  `db:"id" json:"id"`
	Stage string `db:"stage" json:"stage"`
}

func (opOnlyProviderModel) OperationOnlyFields() []string {
	return []string{"stage"}
}

func TestOperationOnlyFieldNames_FromMeta(t *testing.T) {
	names := schema.OperationOnlyFieldNames(opOnlyMetaModel{})
	require.Contains(t, names, "status")
}

func TestOperationOnlyFieldNames_FromFieldTrait(t *testing.T) {
	names := schema.OperationOnlyFieldNames(opOnlyFieldTraitModel{})
	require.Contains(t, names, "state")
	require.Contains(t, names, "state_col")
	require.Contains(t, names, "state_field")
}

func TestOperationOnlyFieldNames_FromProvider(t *testing.T) {
	names := schema.OperationOnlyFieldNames(opOnlyProviderModel{})
	require.Contains(t, names, "stage")
}

func TestFindOperationOnlyField(t *testing.T) {
	data := map[string]interface{}{
		"title":  "Sample",
		"Status": "approved",
	}
	field, found := schema.FindOperationOnlyField(data, []string{"status"})
	assert.True(t, found)
	assert.Equal(t, "Status", field)

	field, found = schema.FindOperationOnlyField(data, []string{"other"})
	assert.False(t, found)
	assert.Empty(t, field)
}
