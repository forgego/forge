package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/forgego/forge/schema"
)

type functionDefaultModel struct {
	schema.BaseSchema
	ID     int64  `json:"id"`
	UID    string `json:"uid"`
	Status string `json:"status"`
}

func (functionDefaultModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.UUIDField("uid", schema.Default("gen_random_uuid()")),
		schema.StringField("status", schema.Default("draft")),
	}
}

// A create request that omits a field gets its literal Default, but not a
// database function default such as gen_random_uuid(), which the database
// fills when the column is left out of the INSERT.
func TestApplySchemaDefaults_LeavesDatabaseFunctionsToTheDatabase(t *testing.T) {
	data := map[string]interface{}{}
	applySchemaDefaults(&functionDefaultModel{}, data)
	assert.Equal(t, "draft", data["status"])
	_, set := data["uid"]
	assert.False(t, set, "gen_random_uuid() must not be sent as the value")
}

type requiredFunctionDefaultModel struct {
	schema.BaseSchema
	ID  int64  `json:"id"`
	UID string `json:"uid"`
}

func (requiredFunctionDefaultModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.UUIDField("uid", schema.Required(), schema.Default("gen_random_uuid()")),
	}
}

// A create may leave a Required field with a database function default
// empty; an update that writes it empty is rejected.
func TestValidateModelInstance_DatabaseFunctionDefaultIsExemptOnCreateOnly(t *testing.T) {
	assert.NoError(t, validateModelInstance(&requiredFunctionDefaultModel{}, true))
	err := validateModelInstance(&requiredFunctionDefaultModel{}, false)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "required")
	}
}
