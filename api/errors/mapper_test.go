package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/forgego/forge/api/exceptions"
)

func TestMapError_UnknownError(t *testing.T) {
	mapper := NewErrorMapper(DefaultSanitizer())

	err := errors.New("unknown error")
	problem := mapper.MapError(err, "/test")

	if problem.Status != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, problem.Status)
	}
	if problem.Code != CodeInternalError {
		t.Errorf("Expected code '%s', got '%s'", CodeInternalError, problem.Code)
	}
}

func TestMapError_ValidationError(t *testing.T) {
	mapper := NewErrorMapper(DefaultSanitizer())

	valErr := exceptions.NewValidationError(map[string][]string{
		"email": {"Invalid email format"},
	})
	problem := mapper.MapError(valErr, "/test")

	if problem.Status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, problem.Status)
	}
	if problem.Code != CodeValidationError {
		t.Errorf("Expected code '%s', got '%s'", CodeValidationError, problem.Code)
	}
	if problem.Errors == nil {
		t.Error("Field errors should not be nil")
	}
}

func TestMapError_NotFound(t *testing.T) {
	mapper := NewErrorMapper(DefaultSanitizer())

	notFoundErr := exceptions.NewNotFound("User not found")
	problem := mapper.MapError(notFoundErr, "/test")

	if problem.Status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, problem.Status)
	}
	if problem.Code != CodeNotFound {
		t.Errorf("Expected code '%s', got '%s'", CodeNotFound, problem.Code)
	}
}

func TestMapError_AuthenticationFailed(t *testing.T) {
	mapper := NewErrorMapper(DefaultSanitizer())

	authErr := exceptions.NewAuthenticationFailed("Invalid credentials")
	problem := mapper.MapError(authErr, "/test")

	if problem.Status != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, problem.Status)
	}
	if problem.Code != CodeAuthenticationFailed {
		t.Errorf("Expected code '%s', got '%s'", CodeAuthenticationFailed, problem.Code)
	}
}

func TestMapError_PermissionDenied(t *testing.T) {
	mapper := NewErrorMapper(DefaultSanitizer())

	permErr := exceptions.NewPermissionDenied("Access denied")
	problem := mapper.MapError(permErr, "/test")

	if problem.Status != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, problem.Status)
	}
	if problem.Code != CodePermissionDenied {
		t.Errorf("Expected code '%s', got '%s'", CodePermissionDenied, problem.Code)
	}
}

func TestMapError_Throttled(t *testing.T) {
	mapper := NewErrorMapper(DefaultSanitizer())

	throttledErr := exceptions.NewThrottled("Rate limit exceeded", 60)
	problem := mapper.MapError(throttledErr, "/test")

	if problem.Status != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, problem.Status)
	}
	if problem.Code != CodeRateLimitExceeded {
		t.Errorf("Expected code '%s', got '%s'", CodeRateLimitExceeded, problem.Code)
	}
	if problem.Meta == nil {
		t.Error("Meta should contain retry_after_seconds")
	}
}

func TestMapPanic(t *testing.T) {
	mapper := NewErrorMapper(DefaultSanitizer())

	problem := mapper.MapPanic("panic: test panic", "/test")

	if problem.Status != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, problem.Status)
	}
	if problem.Code != CodeInternalError {
		t.Errorf("Expected code '%s', got '%s'", CodeInternalError, problem.Code)
	}
}
