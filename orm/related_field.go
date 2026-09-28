package orm

import (
	"fmt"
	"reflect"
)

// RelatedFieldFor creates a type-safe related field expression
// TModel is the related model type, V is the field type
// Returns an error if the relation or field is not found, or if the field type is incorrect
func RelatedFieldFor[T any, TModel any, V any](fa *FieldAccessor[T], relationName, fieldName string) (FieldExpression[V], error) {
	// Validate relation exists
	rel := fa.schema.GetRelation(relationName)
	if rel == nil {
		return FieldExpression[V]{}, fmt.Errorf("relation %s not found", relationName)
	}

	// Get target model schema and validate field
	targetSchema, err := GetModelSchema[TModel]()
	if err != nil {
		return FieldExpression[V]{}, fmt.Errorf("failed to get schema for related model: %w", err)
	}

	fieldInfo := targetSchema.GetField(fieldName)
	if fieldInfo == nil {
		return FieldExpression[V]{}, fmt.Errorf("field %s not found on related model", fieldName)
	}

	expectedType := reflect.TypeOf((*V)(nil)).Elem()
	if fieldInfo.Type != expectedType {
		return FieldExpression[V]{}, fmt.Errorf("field %s has type %v, not %v", fieldName, fieldInfo.Type, expectedType)
	}

	// Build field path
	fieldPath := relationName + "__" + fieldName
	return NewField[V](fieldPath, fa.table), nil
}
