package errors

import "fmt"

// Common framework errors
var (
	ErrNotImplemented = &NotImplementedError{}
	ErrNotFound       = &NotFoundError{}
	ErrInvalidInput   = &InvalidInputError{}
	ErrDatabaseError  = &DatabaseError{}
)

// NotImplementedError indicates a feature is not yet implemented
type NotImplementedError struct {
	Feature string
}

func (e *NotImplementedError) Error() string {
	if e.Feature != "" {
		return fmt.Sprintf("feature not implemented: %s", e.Feature)
	}
	return "feature not implemented"
}

func NewNotImplementedError(feature string) *NotImplementedError {
	return &NotImplementedError{Feature: feature}
}

// NotFoundError indicates a resource was not found
type NotFoundError struct {
	ID       interface{}
	Resource string
}

func (e *NotFoundError) Error() string {
	if e.Resource != "" && e.ID != nil {
		return fmt.Sprintf("%s with id %v not found", e.Resource, e.ID)
	}
	if e.Resource != "" {
		return fmt.Sprintf("%s not found", e.Resource)
	}
	return "resource not found"
}

func NewNotFoundError(resource string, id interface{}) *NotFoundError {
	return &NotFoundError{Resource: resource, ID: id}
}

// NewNotFoundErrorWithMessage creates a NotFoundError with a custom message
func NewNotFoundErrorWithMessage(message string) *NotFoundError {
	return &NotFoundError{Resource: message}
}

// MultipleObjectsReturnedError indicates that a query expected one result but got multiple
type MultipleObjectsReturnedError struct {
	Resource string
	Count    int
}

func (e *MultipleObjectsReturnedError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("get() returned more than one %s -- it returned %d", e.Resource, e.Count)
	}
	return fmt.Sprintf("get() returned more than one result -- it returned %d", e.Count)
}

func NewMultipleObjectsReturnedError(resource string, count int) *MultipleObjectsReturnedError {
	return &MultipleObjectsReturnedError{Resource: resource, Count: count}
}

// InvalidInputError indicates invalid input was provided
type InvalidInputError struct {
	Field   string
	Message string
}

func (e *InvalidInputError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("invalid input for field %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("invalid input: %s", e.Message)
}

func NewInvalidInputError(field, message string) *InvalidInputError {
	return &InvalidInputError{Field: field, Message: message}
}

// DatabaseError indicates a database operation failed
type DatabaseError struct {
	Err       error
	Operation string
}

func (e *DatabaseError) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("database error during %s: %v", e.Operation, e.Err)
	}
	return fmt.Sprintf("database error: %v", e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

func NewDatabaseError(operation string, err error) *DatabaseError {
	return &DatabaseError{Operation: operation, Err: err}
}

// IsNotImplemented checks if an error is a NotImplementedError
func IsNotImplemented(err error) bool {
	_, ok := err.(*NotImplementedError)
	return ok
}

// IsNotFound checks if an error is a NotFoundError
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// IsInvalidInput checks if an error is an InvalidInputError
func IsInvalidInput(err error) bool {
	_, ok := err.(*InvalidInputError)
	return ok
}

// IsDatabaseError checks if an error is a DatabaseError
func IsDatabaseError(err error) bool {
	_, ok := err.(*DatabaseError)
	return ok
}
