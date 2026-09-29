package orm

import (
	"reflect"

	"github.com/forgego/forge/schema"
)

var schemaConstraintValidator func(instance interface{}, fields []schema.Field) error

// RegisterSchemaConstraintValidator installs schema-aware persistence validation.
// The validate package registers its implementation during initialization.
func RegisterSchemaConstraintValidator(validator func(instance interface{}, fields []schema.Field) error) {
	schemaConstraintValidator = validator
}

// validationFields returns the fields to validate before a write. On an insert
// it leaves out a field that is empty and whose Default names a database
// function (gen_random_uuid(), now(), ...): the database fills it, so it is
// not missing. An update writes the value as it is, so it checks every field.
func validationFields(instance interface{}, fields []schema.Field, creating bool) []schema.Field {
	if !creating {
		return fields
	}
	kept := make([]schema.Field, 0, len(fields))
	for _, field := range fields {
		if schema.IsDatabaseFunctionDefault(field.Default) {
			if value, err := getSchemaFieldValue(instance, field); err == nil && (value == nil || reflect.ValueOf(value).IsZero()) {
				continue
			}
		}
		kept = append(kept, field)
	}
	return kept
}
