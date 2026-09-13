package fields

import "fmt"

// BooleanField represents a boolean field
type BooleanField struct {
	*BaseField
}

// NewBooleanField creates a new boolean field
func NewBooleanField(fieldName string) *BooleanField {
	return &BooleanField{
		BaseField: NewBaseField(fieldName),
	}
}

// ToInternalValue converts representation to internal value
func (f *BooleanField) ToInternalValue(data interface{}) (interface{}, error) {
	if data == nil {
		if f.Required {
			return nil, fmt.Errorf("%s: This field is required", f.FieldName)
		}
		return f.Default, nil
	}

	switch v := data.(type) {
	case bool:
		return v, nil
	case string:
		if v == "true" || v == "1" {
			return true, nil
		}
		if v == "false" || v == "0" {
			return false, nil
		}
		return nil, fmt.Errorf("%s: Invalid boolean value", f.FieldName)
	default:
		return nil, fmt.Errorf("%s: Expected boolean, got %T", f.FieldName, data)
	}
}
