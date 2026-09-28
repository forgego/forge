package orm

import "github.com/forgego/forge/schema"

var schemaConstraintValidator func(instance interface{}, fields []schema.Field) error

// RegisterSchemaConstraintValidator installs schema-aware persistence validation.
// The validate package registers its implementation during initialization.
func RegisterSchemaConstraintValidator(validator func(instance interface{}, fields []schema.Field) error) {
	schemaConstraintValidator = validator
}
