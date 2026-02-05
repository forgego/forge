package orm

import (
	"fmt"
	"reflect"
)

// SetValue sets a field value with type checking
func SetValue[T any, V any](ub *UpdateBuilder[T], fieldName string, value V) *UpdateBuilder[T] {
	// Validate field exists and type matches
	fieldInfo := ub.schema.GetField(fieldName)
	if fieldInfo == nil {
		panic(fmt.Sprintf("field %s not found on model", fieldName))
	}

	expectedType := fieldInfo.Type
	actualType := reflect.TypeOf(value)

	// Check if types are assignable
	if !actualType.AssignableTo(expectedType) {
		panic(fmt.Sprintf("field %s expects %v, got %v", fieldName, expectedType, actualType))
	}

	ub.updates[fieldName] = value
	return ub
}

// SetExprValue sets a field to an expression value with type checking
func SetExprValue[T any, V any](ub *UpdateBuilder[T], fieldName string, expr Expression) *UpdateBuilder[T] {
	// Validate field exists
	fieldInfo := ub.schema.GetField(fieldName)
	if fieldInfo == nil {
		panic(fmt.Sprintf("field %s not found on model", fieldName))
	}

	// Validate expression against schema
	if err := expr.Resolve(ub.schema); err != nil {
		panic(fmt.Sprintf("invalid expression for field %s: %v", fieldName, err))
	}

	ub.updates[fieldName] = expr
	return ub
}
