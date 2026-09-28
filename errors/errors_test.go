package errors

import (
	"strings"
	"testing"
)

func TestConfigurationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		field    string
		expected string
	}{
		{
			name:     "with field",
			message:  "database connection is required",
			field:    "db",
			expected: "configuration error: database connection is required (field: db)",
		},
		{
			name:     "without field",
			message:  "missing configuration",
			field:    "",
			expected: "configuration error: missing configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ConfigurationError{Message: tt.message, Field: tt.field}
			if err.Error() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, err.Error())
			}
		})
	}
}

func TestNewConfigurationError(t *testing.T) {
	err := NewConfigurationError("test message", "testField")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Message != "test message" {
		t.Errorf("expected message %q, got %q", "test message", err.Message)
	}
	if err.Field != "testField" {
		t.Errorf("expected field %q, got %q", "testField", err.Field)
	}
}

func TestIsConfiguration(t *testing.T) {
	// Test with ConfigurationError
	err := NewConfigurationError("test", "field")
	if !IsConfiguration(err) {
		t.Error("expected IsConfiguration to return true for ConfigurationError")
	}

	// Test with other error type
	otherErr := NewNotFoundError("resource", 1)
	if IsConfiguration(otherErr) {
		t.Error("expected IsConfiguration to return false for other error types")
	}

	// Test with nil
	if IsConfiguration(nil) {
		t.Error("expected IsConfiguration to return false for nil")
	}
}

func TestConfigurationError_ContainsExpectedStrings(t *testing.T) {
	err := NewConfigurationError("database connection not set", "db")
	errStr := err.Error()

	if !strings.Contains(errStr, "configuration error") {
		t.Error("error message should contain 'configuration error'")
	}
	if !strings.Contains(errStr, "database connection not set") {
		t.Error("error message should contain the original message")
	}
	if !strings.Contains(errStr, "db") {
		t.Error("error message should contain the field name")
	}
}
