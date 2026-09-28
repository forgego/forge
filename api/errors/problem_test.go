package errors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProblem(t *testing.T) {
	problem := NewProblem(
		ErrorTypeValidation,
		CodeValidationError,
		http.StatusBadRequest,
		"Validation Failed",
		"The request contains invalid data",
	)

	if problem.Type == "" {
		t.Error("Problem type should not be empty")
	}
	if problem.Title != "Validation Failed" {
		t.Errorf("Expected title 'Validation Failed', got '%s'", problem.Title)
	}
	if problem.Status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, problem.Status)
	}
	if problem.Code != CodeValidationError {
		t.Errorf("Expected code '%s', got '%s'", CodeValidationError, problem.Code)
	}
}

func TestProblemWithInstance(t *testing.T) {
	problem := NewProblem(
		ErrorTypeNotFound,
		CodeNotFound,
		http.StatusNotFound,
		"Not Found",
		"Resource not found",
	).WithInstance("/api/v1/users/123")

	if problem.Instance != "/api/v1/users/123" {
		t.Errorf("Expected instance '/api/v1/users/123', got '%s'", problem.Instance)
	}
}

func TestProblemWithMeta(t *testing.T) {
	problem := NewProblem(
		ErrorTypeInternal,
		CodeInternalError,
		http.StatusInternalServerError,
		"Internal Error",
		"An error occurred",
	).WithMeta("request_id", "abc-123")

	if problem.Meta == nil {
		t.Error("Meta should not be nil")
	}
	if problem.Meta["request_id"] != "abc-123" {
		t.Errorf("Expected request_id 'abc-123', got '%v'", problem.Meta["request_id"])
	}
}

func TestProblemWithFieldErrors(t *testing.T) {
	problem := NewProblem(
		ErrorTypeValidation,
		CodeValidationError,
		http.StatusBadRequest,
		"Validation Failed",
		"Validation errors",
	).WithFieldError("email", "Invalid email format", CodeInvalidEmailFormatField)

	if problem.Errors == nil {
		t.Error("Errors should not be nil")
	}
	if len(problem.Errors["email"]) != 1 {
		t.Errorf("Expected 1 error for email, got %d", len(problem.Errors["email"]))
	}
	if problem.Errors["email"][0].Message != "Invalid email format" {
		t.Errorf("Expected message 'Invalid email format', got '%s'", problem.Errors["email"][0].Message)
	}
}

func TestProblemWriteJSON(t *testing.T) {
	problem := NewProblem(
		ErrorTypeValidation,
		CodeValidationError,
		http.StatusBadRequest,
		"Validation Failed",
		"Test error",
	)

	w := httptest.NewRecorder()

	err := problem.WriteJSON(w)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/problem+json" {
		t.Errorf("Expected Content-Type 'application/problem+json', got '%s'", w.Header().Get("Content-Type"))
	}

	var result Problem
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Title != "Validation Failed" {
		t.Errorf("Expected title 'Validation Failed', got '%s'", result.Title)
	}
}

func TestProblemError(t *testing.T) {
	problem := NewProblem(
		ErrorTypeNotFound,
		CodeNotFound,
		http.StatusNotFound,
		"Not Found",
		"Resource not found",
	)

	errMsg := problem.Error()
	if errMsg == "" {
		t.Error("Error() should return a non-empty string")
	}
}
