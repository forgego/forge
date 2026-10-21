package validation

import (
	"fmt"

	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
)

// FromSchema creates a typed validator using the schema's constraints and the
// concrete struct fields behind its names. No interface-typed accessor is
// required: schema fields normally have concrete Go types.
func FromSchema[T any](schemaInstance schema.Schema) *TypedValidator[T] {
	builder := For[T]()
	if schemaInstance == nil {
		builder.err = fmt.Errorf("schema is nil")
		return builder.Build()
	}
	for _, field := range schemaInstance.Fields() {
		if field.Name == "" {
			continue
		}
		resolved, ok := schema.ResolveField(new(T), field)
		if !ok {
			builder.err = fmt.Errorf("schema field %s has no exported model field", field.Name)
			return builder.Build()
		}
		fieldBuilder := FieldFor[T, any](builder, orm.NewField[any](field.Name, ""))
		definition := field
		fieldBuilder.schemaField = &definition
		fieldBuilder.fieldIndex = resolved.StructField.Index
		// FieldValidator applies presence, range, choices and custom validators.
		// Formats are tags in this API, so retain the existing type validators.
		switch field.Type {
		case schema.TypeEmail:
			fieldBuilder.Email()
		case schema.TypeURL:
			fieldBuilder.URL()
		case schema.TypeUUID:
			fieldBuilder.UUID()
		}
	}
	return builder.Build()
}
