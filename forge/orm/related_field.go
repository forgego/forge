package orm

import "fmt"
import "reflect"

// RelatedFieldFor creates a type-safe related field expression
// TModel is the related model type, V is the field type
func RelatedFieldFor[T any, TModel any, V any](fa *FieldAccessor[T], relationName, fieldName string) FieldExpression[V] {
	// Validate relation exists
	rel := fa.schema.GetRelation(relationName)
	if rel == nil {
		panic(fmt.Sprintf("relation %s not found", relationName))
	}

	// Get target model schema and validate field
	targetSchema, err := GetModelSchema[TModel]()
	if err != nil {
		panic(fmt.Sprintf("failed to get schema for related model: %v", err))
	}

	fieldInfo := targetSchema.GetField(fieldName)
	if fieldInfo == nil {
		panic(fmt.Sprintf("field %s not found on related model", fieldName))
	}

	expectedType := reflect.TypeOf((*V)(nil)).Elem()
	if fieldInfo.Type != expectedType {
		panic(fmt.Sprintf("field %s has type %v, not %v", fieldName, fieldInfo.Type, expectedType))
	}

	// Build field path
	fieldPath := relationName + "__" + fieldName
	return NewField[V](fieldPath, fa.table)
}
