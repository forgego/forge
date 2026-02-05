package errors

import (
	"testing"
)

func TestDefaultSanitizer(t *testing.T) {
	sanitizer := DefaultSanitizer()
	if sanitizer == nil {
		t.Fatal("DefaultSanitizer should not return nil")
	}
}

func TestSanitizeError(t *testing.T) {
	sanitizer := DefaultSanitizer()

	// Test that stack traces are removed
	errMsg := "error: goroutine 1 [running]:\nmain.main()\n\t/path/to/file.go:123"
	sanitized := sanitizer.SanitizeError(&testError{msg: errMsg})

	if sanitized == errMsg {
		t.Error("Stack trace should be removed from error message")
	}
}

func TestSanitizeDatabaseErrors(t *testing.T) {
	sanitizer := DefaultSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "PostgreSQL error",
			input:    "pq: duplicate key value violates unique constraint",
			expected: "Duplicate entry",
		},
		{
			name:     "MySQL error",
			input:    "mysql error: connection refused",
			expected: "Database connection error",
		},
		{
			name:     "Duplicate key",
			input:    "duplicate key value",
			expected: "Duplicate entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSanitizePII(t *testing.T) {
	config := SanitizerConfig{
		HideStackTraces:  true,
		HideDatabaseErrs: true,
		RedactPII:        true,
		PIIPatterns:      []string{},
	}
	sanitizer := NewSanitizer(config)

	tests := []struct {
		name         string
		input        string
		shouldRedact bool
	}{
		{
			name:         "Password field",
			input:        "password=secret123",
			shouldRedact: true,
		},
		{
			name:         "Token field",
			input:        "token=abc123",
			shouldRedact: true,
		},
		{
			name:         "Normal message",
			input:        "User created successfully",
			shouldRedact: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if tt.shouldRedact {
				if result == tt.input {
					t.Error("PII should be redacted")
				}
			} else {
				if result != tt.input {
					t.Errorf("Non-PII message should not be changed, got '%s'", result)
				}
			}
		})
	}
}

func TestSanitizeMap(t *testing.T) {
	config := SanitizerConfig{
		RedactPII:   true,
		PIIPatterns: []string{},
	}
	sanitizer := NewSanitizer(config)

	data := map[string]interface{}{
		"username": "john",
		"password": "secret123",
		"email":    "john@example.com",
		"age":      30,
	}

	sanitized := sanitizer.SanitizeMap(data)

	if sanitized["password"] != "[REDACTED]" {
		t.Error("Password should be redacted")
	}
	if sanitized["username"] != "john" {
		t.Error("Username should not be redacted (not in default patterns)")
	}
}

// testError is a simple error for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
