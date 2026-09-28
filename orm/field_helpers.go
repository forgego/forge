package orm

import (
	"fmt"
	"reflect"
)

// FieldFor creates a type-safe field expression
// T is the model type, V is the field type
// Returns an error if the field is not found or has an incorrect type
func FieldFor[T any, V any](fa *FieldAccessor[T], name string) (FieldExpression[V], error) {
	// Validate field exists and has correct type
	fieldInfo := fa.schema.GetField(name)
	if fieldInfo == nil {
		return FieldExpression[V]{}, fmt.Errorf("field %s not found on model", name)
	}

	expectedType := reflect.TypeOf((*V)(nil)).Elem()
	if fieldInfo.Type != expectedType {
		return FieldExpression[V]{}, fmt.Errorf("field %s has type %v, not %v", name, fieldInfo.Type, expectedType)
	}

	return NewField[V](name, fa.table), nil
}
