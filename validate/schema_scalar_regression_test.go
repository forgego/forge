package validation

import (
	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneratedValidationTagsAcceptRequiredScalarZeroValues(t *testing.T) {
	for _, test := range []struct {
		name  string
		field schema.Field
		value any
	}{
		{"int64", schema.Int64Field("value", schema.Required(), schema.MinValue(0)), int64(0)},
		{"int32", schema.Int32Field("value", schema.Required()), int32(0)},
		{"float64", schema.Float64Field("value", schema.Required()), float64(0)},
		{"float32", schema.Float32Field("value", schema.Required()), float32(0)},
		{"decimal", schema.DecimalField("value", schema.Required()), float64(0)},
		{"boolean", schema.BoolField("value", schema.Required()), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.NoError(t, NewValidator().ValidateField(test.value, GenerateValidationTag(test.field)))
		})
	}
	var missing *int64
	assert.Error(t, NewValidator().ValidateField(missing, GenerateValidationTag(schema.Int64Field("value", schema.Required()))))
	assert.Error(t, NewValidator().ValidateField("", GenerateValidationTag(schema.StringField("value", schema.Required()))))
	assert.Error(t, NewValidator().ValidateField(int64(-1), GenerateValidationTag(schema.Int64Field("value", schema.Required(), schema.MinValue(0)))))
	// Explicit application tags retain their requested semantics.
	assert.Error(t, NewValidator().ValidateField(int64(0), GenerateValidationTag(schema.Int64Field("value", schema.ValidationTag("required")))))
}
