package orm

import (
	"fmt"
	"reflect"
)

// SafeAccessor provides safe dynamic field access with early validation
type SafeAccessor[T any] struct {
	schema *ModelSchema
}

// NewSafeAccessor creates a new safe accessor for a model type
func NewSafeAccessor[T any]() (*SafeAccessor[T], error) {
	schema, err := GetModelSchema[T]()
	if err != nil {
		return nil, fmt.Errorf("failed to get model schema: %w", err)
	}

	return &SafeAccessor[T]{
		schema: schema,
	}, nil
}

// Get gets a field value with validation
// Validates field exists and type matches at call time
func (sa *SafeAccessor[T]) Get(instance *T, fieldName string) (interface{}, error) {
	// Validate field exists
	fieldInfo := sa.schema.GetField(fieldName)
	if fieldInfo == nil {
		// Try with PascalCase since that's how it's stored in schema (struct field name)
		pascalName := toPascalCaseSafe(fieldName)
		fieldInfo = sa.schema.GetField(pascalName)
		if fieldInfo == nil {
			return nil, fmt.Errorf("field %s not found on model", fieldName)
		}
	}

	// Get value using reflection
	v := reflect.ValueOf(instance)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	fieldValue := v.FieldByName(fieldInfo.Name)
	if !fieldValue.IsValid() {
		// Try lowercase
		fieldValue = v.FieldByName(toPascalCaseSafe(fieldName))
		if !fieldValue.IsValid() {
			return nil, fmt.Errorf("field %s not found in struct", fieldName)
		}
	}

	if fieldValue.CanInterface() {
		return fieldValue.Interface(), nil
	}

	return nil, fmt.Errorf("cannot access field %s", fieldName)
}

// Set sets a field value with validation
func (sa *SafeAccessor[T]) Set(instance *T, fieldName string, value interface{}) error {
	// Validate field exists
	fieldInfo := sa.schema.GetField(fieldName)
	if fieldInfo == nil {
		// Try with PascalCase
		pascalName := toPascalCaseSafe(fieldName)
		fieldInfo = sa.schema.GetField(pascalName)
		if fieldInfo == nil {
			return fmt.Errorf("field %s not found on model", fieldName)
		}
	}

	// Validate type
	valueType := reflect.TypeOf(value)
	if !valueType.AssignableTo(fieldInfo.Type) && !valueType.ConvertibleTo(fieldInfo.Type) {
		return fmt.Errorf("field %s expects type %v, got %v", fieldName, fieldInfo.Type, valueType)
	}

	// Set value using reflection
	v := reflect.ValueOf(instance)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	fieldValue := v.FieldByName(fieldInfo.Name)
	if !fieldValue.IsValid() {
		fieldValue = v.FieldByName(toPascalCaseSafe(fieldName))
	}

	if !fieldValue.IsValid() {
		return fmt.Errorf("field %s not found in struct", fieldName)
	}

	if !fieldValue.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}

	valueReflect := reflect.ValueOf(value)
	if valueReflect.Type().AssignableTo(fieldValue.Type()) {
		fieldValue.Set(valueReflect)
	} else if valueReflect.Type().ConvertibleTo(fieldValue.Type()) {
		fieldValue.Set(valueReflect.Convert(fieldValue.Type()))
	} else {
		return fmt.Errorf("cannot assign value of type %v to field %s of type %v", valueType, fieldName, fieldValue.Type())
	}

	return nil
}

// GetPath gets a value using a field path (supports relations)
func (sa *SafeAccessor[T]) GetPath(instance *T, path string) (interface{}, error) {
	// Use path token for fast access
	token, err := GetCachedPathToken(path, sa.schema)
	if err != nil {
		return nil, fmt.Errorf("failed to compile path %s: %w", path, err)
	}

	return token.Traverse(instance)
}

// toPascalCaseSafe converts snake_case to PascalCase (safe_accessor version)
func toPascalCaseSafe(s string) string {
	if len(s) == 0 {
		return s
	}
	result := ""
	nextUpper := true
	for _, r := range s {
		if r == '_' {
			nextUpper = true
			continue
		}
		if nextUpper {
			if r >= 'a' && r <= 'z' {
				result += string(r - 32)
			} else {
				result += string(r)
			}
			nextUpper = false
		} else {
			result += string(r)
		}
	}
	return result
}
