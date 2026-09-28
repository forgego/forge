package fields

import (
	"fmt"
	"time"
)

// DateTimeField represents a datetime field
type DateTimeField struct {
	*BaseField
	Format string
}

// NewDateTimeField creates a new datetime field
func NewDateTimeField(fieldName string) *DateTimeField {
	return &DateTimeField{
		BaseField: NewBaseField(fieldName),
		Format:    time.RFC3339,
	}
}

// ToInternalValue converts representation to internal value
func (f *DateTimeField) ToInternalValue(data interface{}) (interface{}, error) {
	if data == nil {
		if f.Required {
			return nil, fmt.Errorf("%s: This field is required", f.FieldName)
		}
		return f.Default, nil
	}

	switch v := data.(type) {
	case time.Time:
		return v, nil
	case string:
		t, err := time.Parse(f.Format, v)
		if err != nil {
			// Try RFC3339 as fallback
			t, err = time.Parse(time.RFC3339, v)
			if err != nil {
				return nil, fmt.Errorf("%s: Invalid datetime format", f.FieldName)
			}
		}
		return t, nil
	default:
		return nil, fmt.Errorf("%s: Expected datetime, got %T", f.FieldName, data)
	}
}

// ToRepresentation converts internal value to representation
func (f *DateTimeField) ToRepresentation(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	t, ok := value.(time.Time)
	if !ok {
		return value, nil
	}

	return t.Format(f.Format), nil
}
