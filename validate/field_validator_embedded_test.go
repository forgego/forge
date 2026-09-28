package validation

import (
	"errors"
	"testing"

	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/require"
)

type EmbeddedValidationGenerated struct {
	Code string `json:"code" db:"code_value"`
}

type embeddedValidationModel struct {
	EmbeddedValidationGenerated
}

type rejectingSchemaValidator struct{}

func (rejectingSchemaValidator) Validate(value interface{}) error {
	if value == "invalid" {
		return errors.New("rejected")
	}
	return nil
}

func TestFieldValidator_ValidateModelValidatesEmbeddedConcreteField(t *testing.T) {
	fields := []schema.Field{{
		Name:       "code_value",
		DBColumn:   "code_value",
		Type:       schema.TypeString,
		Validators: []schema.Validator{rejectingSchemaValidator{}},
	}}
	err := NewFieldValidator(NewValidator()).ValidateModel(
		&embeddedValidationModel{EmbeddedValidationGenerated{Code: "invalid"}},
		fields,
	)
	require.ErrorContains(t, err, "code_value: rejected")
}
