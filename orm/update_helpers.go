package orm

import (
	"fmt"
	"reflect"
)

// SetValue sets a field value with type checking.
// Validation errors are deferred and returned when Execute() is called.
func SetValue[T any, V any](ub *UpdateBuilder[T], fieldName string, value V) *UpdateBuilder[T] {
	// Validate field exists and type matches
	fieldInfo := ub.schema.GetField(fieldName)
	if fieldInfo == nil {
		if ub.err == nil {
			ub.err = fmt.Errorf("field %s not found on model", fieldName)
		}
		return ub
	}

	expectedType := fieldInfo.Type
	actualType := reflect.TypeOf(value)

	// Check if types are assignable
	if !actualType.AssignableTo(expectedType) {
		if ub.err == nil {
			ub.err = fmt.Errorf("field %s expects %v, got %v", fieldName, expectedType, actualType)
		}
		return ub
	}

	ub.updates[fieldName] = value
	return ub
}

// SetExprValue sets a field to an expression value with type checking.
// Validation errors are deferred and returned when Execute() is called.
func SetExprValue[T any, V any](ub *UpdateBuilder[T], fieldName string, expr Expression) *UpdateBuilder[T] {
	// Validate field exists
	fieldInfo := ub.schema.GetField(fieldName)
	if fieldInfo == nil {
		if ub.err == nil {
			ub.err = fmt.Errorf("field %s not found on model", fieldName)
		}
		return ub
	}

	// Validate expression against schema
	if err := expr.Resolve(ub.schema); err != nil {
		if ub.err == nil {
			ub.err = fmt.Errorf("invalid expression for field %s: %w", fieldName, err)
		}
		return ub
	}

	ub.updates[fieldName] = expr
	return ub
}
