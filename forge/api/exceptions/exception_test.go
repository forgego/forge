package exceptions

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAPIException_Error(t *testing.T) {
	err := NewAPIException(
		http.StatusBadRequest,
		"test_error",
		"Test error message",
		nil,
	)

	assert.Equal(t, "Test error message", err.Error())
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, "test_error", err.Code)
}

func TestAPIException_ToResponse(t *testing.T) {
	err := NewAPIException(
		http.StatusBadRequest,
		"test_error",
		"Test error message",
		map[string][]string{
			"field1": {"Error 1", "Error 2"},
		},
	)

	response := err.ToResponse()

	assert.True(t, response.Error)
	assert.Equal(t, "test_error", response.Code)
	assert.Equal(t, "Test error message", response.Message)
	assert.NotNil(t, response.Details)
	assert.Contains(t, response.Details, "field1")
}

func TestValidationError(t *testing.T) {
	details := map[string][]string{
		"email": {"Invalid email format"},
		"name":  {"Name is required"},
	}

	err := NewValidationError(details)

	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, "validation_error", err.Code)
	assert.Equal(t, details, err.Details)
}

func TestAuthenticationFailed(t *testing.T) {
	err := NewAuthenticationFailed("Invalid credentials")

	assert.Equal(t, http.StatusUnauthorized, err.Status)
	assert.Equal(t, "authentication_failed", err.Code)
	assert.Equal(t, "Invalid credentials", err.Message)
}

func TestNotAuthenticated(t *testing.T) {
	err := NewNotAuthenticated("Authentication required")

	assert.Equal(t, http.StatusUnauthorized, err.Status)
	assert.Equal(t, "not_authenticated", err.Code)
	assert.Equal(t, "Authentication required", err.Message)
}

func TestPermissionDenied(t *testing.T) {
	err := NewPermissionDenied("You don't have permission")

	assert.Equal(t, http.StatusForbidden, err.Status)
	assert.Equal(t, "permission_denied", err.Code)
	assert.Equal(t, "You don't have permission", err.Message)
}

func TestNotFound(t *testing.T) {
	err := NewNotFound("Resource not found")

	assert.Equal(t, http.StatusNotFound, err.Status)
	assert.Equal(t, "not_found", err.Code)
	assert.Equal(t, "Resource not found", err.Message)
}

func TestThrottled(t *testing.T) {
	err := NewThrottled("Too many requests", 5*time.Minute)

	assert.Equal(t, http.StatusTooManyRequests, err.Status)
	assert.Equal(t, "throttled", err.Code)
	assert.Equal(t, 5*time.Minute, err.RetryAfter)
}

func TestMethodNotAllowed(t *testing.T) {
	err := NewMethodNotAllowed([]string{"GET", "POST"})

	assert.Equal(t, http.StatusMethodNotAllowed, err.Status)
	assert.Equal(t, "method_not_allowed", err.Code)
	assert.Equal(t, []string{"GET", "POST"}, err.AllowedMethods)
}

func TestParseError(t *testing.T) {
	err := NewParseError("Invalid JSON")

	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, "parse_error", err.Code)
	assert.Equal(t, "Invalid JSON", err.Message)
}

func TestNotAcceptable(t *testing.T) {
	err := NewNotAcceptable("Unsupported media type")

	assert.Equal(t, http.StatusNotAcceptable, err.Status)
	assert.Equal(t, "not_acceptable", err.Code)
}

func TestUnsupportedMediaType(t *testing.T) {
	err := NewUnsupportedMediaType("Content type not supported")

	assert.Equal(t, http.StatusUnsupportedMediaType, err.Status)
	assert.Equal(t, "unsupported_media_type", err.Code)
}

func TestDefaultExceptionHandler(t *testing.T) {
	handler := NewDefaultExceptionHandler()

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ValidationError", NewValidationError(nil), "validation_error"},
		{"AuthenticationFailed", NewAuthenticationFailed(""), "authentication_failed"},
		{"NotAuthenticated", NewNotAuthenticated(""), "not_authenticated"},
		{"PermissionDenied", NewPermissionDenied(""), "permission_denied"},
		{"NotFound", NewNotFound(""), "not_found"},
		{"Throttled", NewThrottled("", 0), "throttled"},
		{"APIException", NewAPIException(500, "test", "test", nil), "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			response := handler.HandleException(tt.err, req)

			assert.NotNil(t, response)
			assert.True(t, response.Error)
			assert.Equal(t, tt.expected, response.Code)
		})
	}
}

func TestHandleExceptionHTTP(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{"ValidationError", NewValidationError(nil), http.StatusBadRequest},
		{"AuthenticationFailed", NewAuthenticationFailed(""), http.StatusUnauthorized},
		{"PermissionDenied", NewPermissionDenied(""), http.StatusForbidden},
		{"NotFound", NewNotFound(""), http.StatusNotFound},
		{"Throttled", NewThrottled("", 5*time.Minute), http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			HandleExceptionHTTP(w, req, tt.err, nil)

			assert.Equal(t, tt.statusCode, w.Code)

			// Check for Retry-After header for throttled requests
			if _, ok := tt.err.(*Throttled); ok {
				assert.Contains(t, w.Header().Get("Retry-After"), "300")
			}
		})
	}
}

func TestHandleExceptionHTTP_Throttled_RetryAfter(t *testing.T) {
	err := NewThrottled("Too many requests", 5*time.Minute)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	HandleExceptionHTTP(w, req, err, nil)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "300", w.Header().Get("Retry-After"))
}

func TestExceptionHandler_UnknownError(t *testing.T) {
	handler := NewDefaultExceptionHandler()
	req := httptest.NewRequest("GET", "/test", nil)

	// Unknown error type
	unknownErr := &CustomError{Message: "Unknown error"}
	response := handler.HandleException(unknownErr, req)

	assert.NotNil(t, response)
	assert.True(t, response.Error)
	assert.Equal(t, "internal_error", response.Code)
}

// CustomError for testing unknown error types
type CustomError struct {
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}
