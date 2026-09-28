package fields

import (
	"fmt"
	"reflect"
)

// FloatField represents a float field
type FloatField struct {
	*BaseField
	MinValue *float64
	MaxValue *float64
}

// NewFloatField creates a new float field
func NewFloatField(fieldName string) *FloatField {
	return &FloatField{
		BaseField: NewBaseField(fieldName),
	}
}

// ToInternalValue converts representation to internal value
func (f *FloatField) ToInternalValue(data interface{}) (interface{}, error) {
	if data == nil {
		if f.Required {
			return nil, fmt.Errorf("%s: This field is required", f.FieldName)
		}
		return f.Default, nil
	}

	var floatVal float64

	switch v := data.(type) {
	case float32:
		floatVal = float64(v)
	case float64:
		floatVal = v
	case int:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	default:
		rv := reflect.ValueOf(data)
		if rv.Kind() == reflect.Float32 || rv.Kind() == reflect.Float64 {
			floatVal = rv.Float()
		} else {
			return nil, fmt.Errorf("%s: Expected float, got %T", f.FieldName, data)
		}
	}

	return floatVal, nil
}

// Validate validates the float value
func (f *FloatField) Validate(value interface{}) error {
	if err := f.BaseField.Validate(value); err != nil {
		return err
	}

	if value == nil {
		return nil
	}

	var floatVal float64
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		floatVal = rv.Float()
	default:
		return fmt.Errorf("%s: Expected float", f.FieldName)
	}

	if f.MinValue != nil && floatVal < *f.MinValue {
		return fmt.Errorf("%s: Ensure this value is greater than or equal to %g", f.FieldName, *f.MinValue)
	}

	if f.MaxValue != nil && floatVal > *f.MaxValue {
		return fmt.Errorf("%s: Ensure this value is less than or equal to %g", f.FieldName, *f.MaxValue)
	}

	return nil
}
