package serializers

import (
	"fmt"

	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
)

// FromSchema creates a typed serializer from a schema definition
// This automatically maps schema fields to serializer fields
func FromSchema[T any](schemaInstance schema.Schema) *TypedSerializer[T] {
	serializer := NewTypedSerializer[T]()
	if isNilValue(schemaInstance) {
		serializer.configErr = fmt.Errorf("schema is nil")
		return serializer
	}
	serializer.schema = schemaInstance

	// Map schema fields to serializer fields
	fields := schemaInstance.Fields()
	for _, field := range fields {
		// Skip unexported or internal fields
		if field.Name == "" {
			continue
		}

		// Create field expression based on field type
		typedField, err := createTypedFieldFromSchema[T](field)
		if err != nil {
			serializer.configErr = err
			return serializer
		}

		// Apply schema constraints
		if field.Required {
			typedField = typedField.Required()
		}
		if !field.Editable || field.PrimaryKey || field.AutoIncrement || field.Generated || field.AutoNow || field.AutoNowAdd {
			typedField = typedField.ReadOnly()
		}
		if !field.Serialize {
			typedField = typedField.WriteOnly()
		}

		// Add to serializer
		serializer.AddField(field.Name, typedField)
	}

	return serializer
}

// createTypedFieldFromSchema creates a typed field from a schema field
// Returns an error if the field cannot be created
func createTypedFieldFromSchema[T any](schemaField schema.Field) (TypedField[T], error) {
	resolved, ok := schema.ResolveField(new(T), schemaField)
	if !ok {
		return TypedField[T]{}, fmt.Errorf("schema field %q has no exported model field", schemaField.Name)
	}
	typedField := Field[T](orm.NewField[interface{}](schemaField.Name, ""))
	typedField.fieldIndex = resolved.StructField.Index
	validationField := schemaField
	validationField.Required = false // Presence is checked separately; zero values are valid input.
	typedField.schemaField = &validationField

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
