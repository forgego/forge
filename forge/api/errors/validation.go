package errors

import (
	"fmt"
	"net/http"
	"strings"
)

// ValidationError represents a validation error with field-level details
type ValidationError struct {
	*BaseError
	FieldErrors map[string][]FieldError
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	if message == "" {
		message = "Validation failed"
	}

	return &ValidationError{
		BaseError: NewBaseError(
			ErrorTypeValidation,
			CodeValidationError,
			http.StatusBadRequest,
			"Validation Failed",
			message,
		),
		FieldErrors: make(map[string][]FieldError),
	}
}

// AddFieldError adds a field-level error
func (e *ValidationError) AddFieldError(field, message, code string) *ValidationError {
	if e.FieldErrors == nil {
		e.FieldErrors = make(map[string][]FieldError)
	}

	// Use default code if not provided
	if code == "" {
		code = CodeInvalidField
	}

	e.FieldErrors[field] = append(e.FieldErrors[field], FieldError{
		Message: message,
		Code:    code,
	})
	return e
}

// AddFieldErrors adds multiple errors for a field
func (e *ValidationError) AddFieldErrors(field string, errors []FieldError) *ValidationError {
	if e.FieldErrors == nil {
		e.FieldErrors = make(map[string][]FieldError)
	}
	e.FieldErrors[field] = append(e.FieldErrors[field], errors...)
	return e
}

// SetFieldErrors sets all field errors at once
func (e *ValidationError) SetFieldErrors(errors map[string][]FieldError) *ValidationError {
	e.FieldErrors = errors
	return e
}

// HasErrors returns whether there are any field errors
func (e *ValidationError) HasErrors() bool {
	return len(e.FieldErrors) > 0
}

// GetFieldError returns errors for a specific field
func (e *ValidationError) GetFieldError(field string) []FieldError {
	if e.FieldErrors == nil {
		return nil
	}
	return e.FieldErrors[field]
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}

	if len(e.FieldErrors) > 0 {
		var fields []string
		for field := range e.FieldErrors {
			fields = append(fields, field)
		}
		return fmt.Sprintf("Validation failed for fields: %s", strings.Join(fields, ", "))
	}

	return e.Title
}

// ToProblem converts the validation error to a Problem Details structure
func (e *ValidationError) ToProblem() *Problem {
	problem := NewProblem(
		e.Type,
		e.Code,
		e.Status,
		e.Title,
		e.Detail,
	).WithInstance(e.Instance)

	// Add field errors
	if len(e.FieldErrors) > 0 {
		problem.WithFieldErrors(e.FieldErrors)
	}

	// Add metadata
	if len(e.Meta) > 0 {
		problem.WithMetaMap(e.Meta)
	}

	return problem
}

// GetType returns the error type
func (e *ValidationError) GetType() ErrorType {
	return ErrorTypeValidation
}

// GetCode returns the error code
func (e *ValidationError) GetCode() string {
	return e.Code
}

// GetStatus returns the HTTP status code
func (e *ValidationError) GetStatus() int {
	return http.StatusBadRequest
}

// NewFieldValidationError creates a validation error for a single field
func NewFieldValidationError(field, message, code string) *ValidationError {
	err := NewValidationError(fmt.Sprintf("Validation failed for field '%s'", field))
	err.AddFieldError(field, message, code)
	return err
}

// NewMultipleFieldValidationError creates a validation error for multiple fields
func NewMultipleFieldValidationError(fieldErrors map[string][]FieldError) *ValidationError {
	err := NewValidationError("Validation failed for multiple fields")
	err.SetFieldErrors(fieldErrors)
	return err
}

// Common validation error codes for fields (field-specific)
const (
	CodeInvalidEmailFormatField  = "INVALID_EMAIL_FORMAT"
	CodePasswordTooShortField    = "PASSWORD_TOO_SHORT"
	CodePasswordTooLongField     = "PASSWORD_TOO_LONG"
	CodePasswordTooWeakField     = "PASSWORD_TOO_WEAK"
	CodeInvalidURLFormatField    = "INVALID_URL_FORMAT"
	CodeInvalidUUIDFormatField   = "INVALID_UUID_FORMAT"
	CodeInvalidDateFormatField   = "INVALID_DATE_FORMAT"
	CodeInvalidNumberFormatField = "INVALID_NUMBER_FORMAT"
	CodeRequiredFieldField       = "REQUIRED_FIELD"
)

// RegisterValidationErrorCodes registers common validation error codes
func init() {
	RegisterErrorCodeWithVersion(CodeInvalidEmailFormatField, "Invalid email format", 400, ErrorTypeValidation, "v1")
	RegisterErrorCodeWithVersion(CodePasswordTooShortField, "Password is too short", 400, ErrorTypeValidation, "v1")
	RegisterErrorCodeWithVersion(CodePasswordTooLongField, "Password is too long", 400, ErrorTypeValidation, "v1")
	RegisterErrorCodeWithVersion(CodePasswordTooWeakField, "Password is too weak", 400, ErrorTypeValidation, "v1")
	RegisterErrorCodeWithVersion(CodeInvalidURLFormatField, "Invalid URL format", 400, ErrorTypeValidation, "v1")
	RegisterErrorCodeWithVersion(CodeInvalidUUIDFormatField, "Invalid UUID format", 400, ErrorTypeValidation, "v1")
	RegisterErrorCodeWithVersion(CodeInvalidDateFormatField, "Invalid date format", 400, ErrorTypeValidation, "v1")
	RegisterErrorCodeWithVersion(CodeInvalidNumberFormatField, "Invalid number format", 400, ErrorTypeValidation, "v1")
	RegisterErrorCodeWithVersion(CodeRequiredFieldField, "Required field is missing", 400, ErrorTypeValidation, "v1")
}
