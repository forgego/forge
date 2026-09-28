package exceptions

import "fmt"

// APIException is the base exception for all API errors
type APIException struct {
	Status  int
	Code    string
	Message string
	Details interface{}
}

// Error implements the error interface
func (e *APIException) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("API error: %s (status: %d)", e.Code, e.Status)
}

// NewAPIException creates a new API exception
func NewAPIException(status int, code, message string, details interface{}) *APIException {
	return &APIException{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
	}
}
