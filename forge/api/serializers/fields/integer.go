package fields

import (
	"fmt"
	"reflect"
)

// IntegerField represents an integer field
type IntegerField struct {
	*BaseField
	MinValue *int64
	MaxValue *int64
}

// NewIntegerField creates a new integer field
func NewIntegerField(fieldName string) *IntegerField {
	return &IntegerField{
		BaseField: NewBaseField(fieldName),
	}
}

// ToInternalValue converts representation to internal value
func (f *IntegerField) ToInternalValue(data interface{}) (interface{}, error) {
	if data == nil {
		if f.Required {
			return nil, fmt.Errorf("%s: This field is required", f.FieldName)
		}
		return f.Default, nil
	}

	var intVal int64

	switch v := data.(type) {
	case int:
		intVal = int64(v)
	case int8:
		intVal = int64(v)
	case int16:
		intVal = int64(v)
	case int32:
		intVal = int64(v)
	case int64:
		intVal = v
	case float64:
		intVal = int64(v)
	case float32:
		intVal = int64(v)
	default:
		return nil, fmt.Errorf("%s: Expected integer, got %T", f.FieldName, data)
	}

	return intVal, nil
}

// Validate validates the integer value
func (f *IntegerField) Validate(value interface{}) error {
	if err := f.BaseField.Validate(value); err != nil {
		return err
	}

	if value == nil {
		return nil
	}

	var intVal int64
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal = rv.Int()
	default:
		return fmt.Errorf("%s: Expected integer", f.FieldName)
	}

	if f.MinValue != nil && intVal < *f.MinValue {
		return fmt.Errorf("%s: Ensure this value is greater than or equal to %d", f.FieldName, *f.MinValue)
	}

	if f.MaxValue != nil && intVal > *f.MaxValue {
		return fmt.Errorf("%s: Ensure this value is less than or equal to %d", f.FieldName, *f.MaxValue)
	}

	return nil
}
