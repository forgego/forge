package serializers

import (
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
)

// FromSchema creates a typed serializer from a schema definition
// This automatically maps schema fields to serializer fields
func FromSchema[T any](schemaInstance schema.Schema) *TypedSerializer[T] {
	serializer := NewTypedSerializer[T]()

	// Get model schema from ORM (for future use)
	_, err := orm.GetModelSchema[T]()
	if err != nil {
		// Return empty serializer if schema not found
		return serializer
	}

	// Get field accessor
	accessor, err := orm.NewFieldAccessor[T]()
	if err != nil {
		return serializer
	}

	// Map schema fields to serializer fields
	fields := schemaInstance.Fields()
	for _, field := range fields {
		// Skip unexported or internal fields
		if field.Name == "" {
			continue
		}

		// Create field expression based on field type
		typedField, err := createTypedFieldFromSchema[T](field, accessor)
		if err != nil {
			// Skip fields that can't be created
			continue
		}

		// Apply schema constraints
		if field.Required {
			typedField = typedField.Required()
		}

		// Add to serializer
		serializer.AddField(field.Name, typedField)
	}

	return serializer
}

// createTypedFieldFromSchema creates a typed field from a schema field
// Returns an error if the field cannot be created
func createTypedFieldFromSchema[T any](schemaField schema.Field, accessor *orm.FieldAccessor[T]) (TypedField[T], error) {
	// Create field expression based on field type
	// This is a simplified version - in practice, you'd map each schema field type
	fieldExpr, err := orm.FieldFor[T, interface{}](accessor, schemaField.Name)
	if err != nil {
		return TypedField[T]{}, err
	}

	typedField := Field[T, interface{}](fieldExpr)

	// Apply schema options
	if schemaField.Default != nil {
		typedField = typedField.Default(schemaField.Default)
	}

	return typedField, nil
}

// Override allows customizing a specific field in a serializer
func (ts *TypedSerializer[T]) Override(fieldName string, field TypedField[T]) *TypedSerializer[T] {
	ts.fields[fieldName] = field
	return ts
}
