package core

import "fmt"

// ErrorCode represents a migration error code
type ErrorCode string

const (
	ErrInvalidChange    ErrorCode = "INVALID_CHANGE"
	ErrStateMismatch    ErrorCode = "STATE_MISMATCH"
	ErrChecksumMismatch ErrorCode = "CHECKSUM_MISMATCH"
	ErrMigrationFailed  ErrorCode = "MIGRATION_FAILED"
	ErrParseFailed      ErrorCode = "PARSE_FAILED"
	ErrValidationFailed ErrorCode = "VALIDATION_FAILED"
)

// MigrationError represents a domain-specific migration error
type MigrationError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *MigrationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *MigrationError) Unwrap() error {
	return e.Cause
}

// NewMigrationError creates a new migration error
func NewMigrationError(code ErrorCode, message string, cause error) *MigrationError {
	return &MigrationError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
