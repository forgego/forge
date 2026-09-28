package exceptions

import (
	"net/http"
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
