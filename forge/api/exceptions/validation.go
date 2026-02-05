package exceptions

// ValidationError represents validation errors
type ValidationError struct {
	*APIException
	Errors map[string][]string
}

// NewValidationError creates a new validation error
func NewValidationError(errors map[string][]string) *ValidationError {
	return &ValidationError{
		APIException: NewAPIException(
			400,
			"validation_error",
			"Validation failed",
			errors,
		),
		Errors: errors,
	}
}

// AddError adds an error for a field
func (e *ValidationError) AddError(field, message string) {
	if e.Errors == nil {
		e.Errors = make(map[string][]string)
	}
	e.Errors[field] = append(e.Errors[field], message)
}

// HasErrors returns whether there are any errors
func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}
