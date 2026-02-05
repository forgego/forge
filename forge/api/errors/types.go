package errors

import (
	"fmt"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation-error"
	// ErrorTypeAuthentication represents authentication errors
	ErrorTypeAuthentication ErrorType = "authentication-error"
	// ErrorTypeAuthorization represents authorization errors
	ErrorTypeAuthorization ErrorType = "authorization-error"
	// ErrorTypeNotFound represents not found errors
	ErrorTypeNotFound ErrorType = "not-found-error"
	// ErrorTypeConflict represents conflict errors
	ErrorTypeConflict ErrorType = "conflict-error"
	// ErrorTypeRateLimit represents rate limit errors
	ErrorTypeRateLimit ErrorType = "rate-limit-error"
	// ErrorTypeInternal represents internal server errors
	ErrorTypeInternal ErrorType = "internal-error"
	// ErrorTypeBadRequest represents bad request errors
	ErrorTypeBadRequest ErrorType = "bad-request-error"
	// ErrorTypeMethodNotAllowed represents method not allowed errors
	ErrorTypeMethodNotAllowed ErrorType = "method-not-allowed-error"
	// ErrorTypeNotAcceptable represents not acceptable errors
	ErrorTypeNotAcceptable ErrorType = "not-acceptable-error"
	// ErrorTypeUnsupportedMediaType represents unsupported media type errors
	ErrorTypeUnsupportedMediaType ErrorType = "unsupported-media-type-error"
)

// ProblemError is an error that implements the error interface and can be converted to a Problem
// (FieldError is defined in problem.go)
type ProblemError interface {
	error
	ToProblem() *Problem
	GetType() ErrorType
	GetCode() string
	GetStatus() int
}

// BaseError is the base implementation of ProblemError
type BaseError struct {
	Type     ErrorType
	Code     string
	Status   int
	Title    string
	Detail   string
	Instance string
	Meta     map[string]interface{}
}

// Error implements the error interface
func (e *BaseError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return e.Title
}

// GetType returns the error type
func (e *BaseError) GetType() ErrorType {
	return e.Type
}

// GetCode returns the error code
func (e *BaseError) GetCode() string {
	return e.Code
}

// GetStatus returns the HTTP status code
func (e *BaseError) GetStatus() int {
	return e.Status
}

// ToProblem converts the error to a Problem Details structure
func (e *BaseError) ToProblem() *Problem {
	return &Problem{
		Type:     e.getTypeURI(),
		Title:    e.Title,
		Status:   e.Status,
		Detail:   e.Detail,
		Instance: e.Instance,
		Code:     e.Code,
		Meta:     e.Meta,
	}
}

// getTypeURI generates the type URI for the problem
func (e *BaseError) getTypeURI() string {
	// Default base URL, can be overridden via configuration
	baseURL := "https://api.example.com/problems"
	return fmt.Sprintf("%s/%s", baseURL, string(e.Type))
}

// NewBaseError creates a new base error
func NewBaseError(errType ErrorType, code string, status int, title, detail string) *BaseError {
	return &BaseError{
		Type:   errType,
		Code:   code,
		Status: status,
		Title:  title,
		Detail: detail,
		Meta:   make(map[string]interface{}),
	}
}

// WithInstance sets the instance URI
func (e *BaseError) WithInstance(instance string) *BaseError {
	e.Instance = instance
	return e
}

// WithMeta sets metadata
func (e *BaseError) WithMeta(key string, value interface{}) *BaseError {
	if e.Meta == nil {
		e.Meta = make(map[string]interface{})
	}
	e.Meta[key] = value
	return e
}

// WithMetaMap sets multiple metadata entries
func (e *BaseError) WithMetaMap(meta map[string]interface{}) *BaseError {
	if e.Meta == nil {
		e.Meta = make(map[string]interface{})
	}
	for k, v := range meta {
		e.Meta[k] = v
	}
	return e
}

// SetTypeBaseURL sets the base URL for problem type URIs
// This is a global setting that affects all errors
var typeBaseURL = "https://api.example.com/problems"

// SetTypeBaseURL sets the base URL for problem type URIs
func SetTypeBaseURL(url string) {
	typeBaseURL = url
}

// GetTypeBaseURL returns the current base URL for problem type URIs
func GetTypeBaseURL() string {
	return typeBaseURL
}
