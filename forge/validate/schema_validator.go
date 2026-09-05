package validation

import (
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
)

// FromSchema creates a typed validator from a schema definition
// This automatically generates validators based on schema field constraints
func FromSchema[T any](schemaInstance schema.Schema) *TypedValidator[T] {
	builder := For[T]()

	// Get model schema from ORM (for future use)
	_, err := orm.GetModelSchema[T]()
	if err != nil {
		// Return empty validator if schema not found
		return builder.Build()
	}

	// Get field accessor
	accessor, err := orm.NewFieldAccessor[T]()
	if err != nil {
		return builder.Build()
	}

	// Map schema fields to validators
	fields := schemaInstance.Fields()
	for _, field := range fields {
		// Skip unexported fields
		if field.Name == "" {
			continue
		}

		// Create field expression
		fieldExpr, err := orm.FieldFor[T, interface{}](accessor, field.Name)
		if err != nil {
			// Skip fields that can't be accessed
			continue
		}

		// Build field validator using FieldFor helper
		fieldBuilder := FieldFor[T, interface{}](builder, fieldExpr)

		// Apply schema constraints
		if field.Required {
			fieldBuilder = fieldBuilder.WithRequired()
		}

		if field.MinLength != nil {
			fieldBuilder = fieldBuilder.WithMinLength(*field.MinLength)
		}

		if field.MaxLength != nil {
			fieldBuilder = fieldBuilder.WithMaxLength(*field.MaxLength)
		}

		if field.MinValue != nil {
			fieldBuilder = fieldBuilder.Min(*field.MinValue)
		}

		if field.MaxValue != nil {
			fieldBuilder = fieldBuilder.Max(*field.MaxValue)
		}

		// Type-specific validators
		switch field.Type {
		case schema.TypeEmail:
			fieldBuilder = fieldBuilder.Email()
		case schema.TypeURL:
			fieldBuilder = fieldBuilder.URL()
		case schema.TypeUUID:
			fieldBuilder = fieldBuilder.UUID()
		}

		// Unique constraint
		if field.Unique {
			fieldBuilder = fieldBuilder.WithUnique()
		}

		// Choices
		if len(field.Choices) > 0 {
			choices := make([]interface{}, len(field.Choices))
			for i, choice := range field.Choices {
				choices[i] = choice.Value
			}
			fieldBuilder = fieldBuilder.WithChoices(choices...)
		}
	}

	return builder.Build()
}


