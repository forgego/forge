package fields

import (
	"fmt"
	"net/url"
)

// URLField represents a URL field
type URLField struct {
	*StringField
}

// NewURLField creates a new URL field
func NewURLField(fieldName string) *URLField {
	return &URLField{
		StringField: NewStringField(fieldName),
	}
}

// Validate validates the URL value
func (f *URLField) Validate(value interface{}) error {
	if err := f.StringField.Validate(value); err != nil {
		return err
	}

	if value == nil {
		return nil
	}

	urlStr, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s: Expected string", f.FieldName)
	}

	if _, err := url.Parse(urlStr); err != nil {
		return fmt.Errorf("%s: Enter a valid URL", f.FieldName)
	}

	return nil
}
