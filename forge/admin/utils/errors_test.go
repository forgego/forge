package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAdminError(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
	}{
		{"not found", http.StatusNotFound, "Resource not found"},
		{"forbidden", http.StatusForbidden, "Access denied"},
		{"bad request", http.StatusBadRequest, "Invalid input"},
		{"internal error", http.StatusInternalServerError, "Internal error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAdminError(tt.code, tt.message)
			if err.Code != tt.code {
				t.Errorf("AdminError.Code = %d, want %d", err.Code, tt.code)
			}
			if err.Message != tt.message {
				t.Errorf("AdminError.Message = %q, want %q", err.Message, tt.message)
			}
			if err.Details == nil {
				t.Error("AdminError.Details should be initialized")
			}
		})
	}
}

func TestAdminError_Error(t *testing.T) {
	err := NewAdminError(http.StatusNotFound, "Not found")
	if err.Error() != "Not found" {
		t.Errorf("Error() = %q, want %q", err.Error(), "Not found")
	}
}

func TestAdminError_WithDetail(t *testing.T) {
	err := NewAdminError(http.StatusBadRequest, "Validation failed")
	result := err.WithDetail("field", "email")
	
	if result != err {
		t.Error("WithDetail() should return the same error for chaining")
	}
	if err.Details["field"] != "email" {
		t.Errorf("Details['field'] = %v, want 'email'", err.Details["field"])
	}
}

func TestAdminError_WithDetail_Multiple(t *testing.T) {
	err := NewAdminError(http.StatusBadRequest, "Validation failed")
	err.WithDetail("field", "email").WithDetail("error", "invalid format")
	
	if len(err.Details) != 2 {
		t.Errorf("Details length = %d, want 2", len(err.Details))
	}
}

func TestAdminError_HTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"with code", http.StatusNotFound, http.StatusNotFound},
		{"zero code", 0, http.StatusInternalServerError},
		{"custom code", 418, 418},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAdminError(tt.code, "test")
			if status := err.HTTPStatus(); status != tt.expected {
				t.Errorf("HTTPStatus() = %d, want %d", status, tt.expected)
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *AdminError
		expected int
	}{
		{"ErrModelNotFound", ErrModelNotFound, http.StatusNotFound},
		{"ErrInstanceNotFound", ErrInstanceNotFound, http.StatusNotFound},
		{"ErrPermissionDenied", ErrPermissionDenied, http.StatusForbidden},
		{"ErrValidationFailed", ErrValidationFailed, http.StatusBadRequest},
		{"ErrInvalidID", ErrInvalidID, http.StatusBadRequest},
		{"ErrInvalidFormData", ErrInvalidFormData, http.StatusBadRequest},
		{"ErrActionNotFound", ErrActionNotFound, http.StatusNotFound},
		{"ErrExportFailed", ErrExportFailed, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.HTTPStatus() != tt.expected {
				t.Errorf("%s.HTTPStatus() = %d, want %d", tt.name, tt.err.HTTPStatus(), tt.expected)
			}
		})
	}
}

func TestHandleAdminError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "admin error",
			err:          NewAdminError(http.StatusNotFound, "Not found"),
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			
			if tt.err != nil {
				HandleAdminError(rr, tt.err)
				if rr.Code != tt.expectedCode {
					t.Errorf("HandleAdminError() status = %d, want %d", rr.Code, tt.expectedCode)
				}
			}
		})
	}
}

func TestHandleAdminError_GenericError(t *testing.T) {
	rr := httptest.NewRecorder()
	genericErr := error(&testError{msg: "something went wrong"})
	HandleAdminError(rr, genericErr)
	
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("HandleAdminError() status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestAdminError_NilDetails(t *testing.T) {
	err := &AdminError{
		Code:    http.StatusBadRequest,
		Message: "test",
	}
	// HTTPStatus should still work with nil Details
	if status := err.HTTPStatus(); status != http.StatusBadRequest {
		t.Errorf("HTTPStatus() = %d, want %d", status, http.StatusBadRequest)
	}
}
