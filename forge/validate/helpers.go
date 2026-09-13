package validation

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/schema"
)

// ValidationError represents a validation error with field name and message
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	messages := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

// Add adds a validation error
func (e *ValidationErrors) Add(field, message string) {
	e.Errors = append(e.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are any errors
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// ValidateFields validates multiple field values against their schema definitions
func ValidateFields(validator *Validator, fields []schema.Field, values map[string]interface{}) error {
	fv := NewFieldValidator(validator)
	errors := &ValidationErrors{}

	for _, field := range fields {
		value, exists := values[field.Name]
		if !exists {
			value = nil
		}

		if err := fv.ValidateField(field, value); err != nil {
			errors.Add(field.Name, err.Error())
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateRequiredFields validates that all required fields are present
func ValidateRequiredFields(fields []schema.Field, values map[string]interface{}) error {
	errors := &ValidationErrors{}

	for _, field := range fields {
		if field.Required {
			value, exists := values[field.Name]
			if !exists || isEmptyValue(value) {
				errors.Add(field.Name, "is required")
			}
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// isEmptyValue checks if a value is empty
func isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case []byte:
		return len(v) == 0
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// BuildValidationTag builds a validation tag string from field constraints
// This is a convenience function that calls GenerateValidationTag
func BuildValidationTag(field schema.Field) string {
	return GenerateValidationTag(field)
}

// ValidateStructWithTags validates a struct using validation tags
func ValidateStructWithTags(validator *Validator, structValue interface{}) error {
	return validator.ValidateStruct(structValue)
}

// ValidateFieldValue validates a single field value against a validation tag
func ValidateFieldValue(validator *Validator, value interface{}, tag string) error {
	return validator.ValidateField(value, tag)
}
