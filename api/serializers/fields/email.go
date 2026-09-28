package fields

import (
	"fmt"
	"strings"
)

// EmailField represents an email field
type EmailField struct {
	*StringField
}

// NewEmailField creates a new email field
func NewEmailField(fieldName string) *EmailField {
	return &EmailField{
		StringField: NewStringField(fieldName),
	}
}

// Validate validates the email value
func (f *EmailField) Validate(value interface{}) error {
	if err := f.StringField.Validate(value); err != nil {
		return err
	}

	if value == nil {
		return nil
	}

	email, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s: Expected string", f.FieldName)
	}

	if !isValidEmail(email) {
		return fmt.Errorf("%s: Enter a valid email address", f.FieldName)
	}

	return nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	if parts[0] == "" || parts[1] == "" {
		return false
	}

	// Check for domain
	if !strings.Contains(parts[1], ".") {
		return false
	}

	return true
}
