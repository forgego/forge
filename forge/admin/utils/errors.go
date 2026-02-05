package utils

import (
	"fmt"
	"net/http"
)

// AdminError represents an admin-specific error
type AdminError struct {
	Code    int
	Message string
	Details map[string]interface{}
}

// Error implements the error interface
func (e *AdminError) Error() string {
	return e.Message
}

// NewAdminError creates a new admin error
func NewAdminError(code int, message string) *AdminError {
	return &AdminError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetail adds a detail to the error
func (e *AdminError) WithDetail(key string, value interface{}) *AdminError {
	e.Details[key] = value
	return e
}

// HTTPStatus returns the HTTP status code
func (e *AdminError) HTTPStatus() int {
	if e.Code != 0 {
		return e.Code
	}
	return http.StatusInternalServerError
}

// Common admin errors
var (
	ErrModelNotFound     = NewAdminError(http.StatusNotFound, "Model not found")
	ErrInstanceNotFound  = NewAdminError(http.StatusNotFound, "Instance not found")
	ErrPermissionDenied  = NewAdminError(http.StatusForbidden, "Permission denied")
	ErrValidationFailed  = NewAdminError(http.StatusBadRequest, "Validation failed")
	ErrInvalidID         = NewAdminError(http.StatusBadRequest, "Invalid ID")
	ErrInvalidFormData   = NewAdminError(http.StatusBadRequest, "Invalid form data")
	ErrActionNotFound    = NewAdminError(http.StatusNotFound, "Action not found")
	ErrExportFailed      = NewAdminError(http.StatusInternalServerError, "Export failed")
)

// HandleAdminError handles an admin error and writes HTTP response
func HandleAdminError(w http.ResponseWriter, err error) {
	adminErr, ok := err.(*AdminError)
	if !ok {
		adminErr = NewAdminError(http.StatusInternalServerError, err.Error())
	}
	
	w.WriteHeader(adminErr.HTTPStatus())
	fmt.Fprintf(w, `{"error": "%s", "code": %d}`, adminErr.Message, adminErr.Code)
}
