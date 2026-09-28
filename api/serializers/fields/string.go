package fields

import (
	"fmt"
	"strings"
)

// StringField represents a string field
type StringField struct {
	*BaseField
	MaxLength      *int
	MinLength      *int
	TrimWhitespace bool
}

// NewStringField creates a new string field
func NewStringField(fieldName string) *StringField {
	return &StringField{
		BaseField: NewBaseField(fieldName),
	}
}

// ToInternalValue converts representation to internal value
func (f *StringField) ToInternalValue(data interface{}) (interface{}, error) {
	if data == nil {
		if f.Required {
			return nil, fmt.Errorf("%s: This field is required", f.FieldName)
		}
		return f.Default, nil
	}

	str, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("%s: Expected string, got %T", f.FieldName, data)
	}

	// Trim whitespace if configured
	if f.TrimWhitespace {
		str = strings.TrimSpace(str)
	}

	// Check blank
	if str == "" && !f.AllowBlank {
		if f.Required {
			return nil, fmt.Errorf("%s: This field may not be blank", f.FieldName)
		}
		return f.Default, nil
	}

	return str, nil
}

// Validate validates the string value
func (f *StringField) Validate(value interface{}) error {
	if err := f.BaseField.Validate(value); err != nil {
		return err
	}

	if value == nil {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s: Expected string", f.FieldName)
	}

	// Check length constraints
	if f.MinLength != nil && len(str) < *f.MinLength {
		return fmt.Errorf("%s: Ensure this field has at least %d characters", f.FieldName, *f.MinLength)
	}

	if f.MaxLength != nil && len(str) > *f.MaxLength {
		return fmt.Errorf("%s: Ensure this field has no more than %d characters", f.FieldName, *f.MaxLength)
	}

	return nil
}
