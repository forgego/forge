package exceptions

import (
	"net/http"
	"strconv"

	"github.com/forgego/forge/api/core"
)

// ExceptionHandler handles exceptions and converts them to HTTP responses
type ExceptionHandler interface {
	HandleException(err error, r *http.Request) *ErrorResponse
}

// DefaultExceptionHandler is the default exception handler
type DefaultExceptionHandler struct{}

// NewDefaultExceptionHandler creates a new default exception handler
func NewDefaultExceptionHandler() *DefaultExceptionHandler {
	return &DefaultExceptionHandler{}
}

// HandleException handles an exception and returns an error response
func (h *DefaultExceptionHandler) HandleException(err error, r *http.Request) *ErrorResponse {
	// Check for known exception types
	switch e := err.(type) {
	case *ValidationError:
		return e.ToResponse()
	case *AuthenticationFailed:
		return e.ToResponse()
	case *NotAuthenticated:
		return e.ToResponse()
	case *PermissionDenied:
		return e.ToResponse()
	case *NotFound:
		return e.ToResponse()
	case *Throttled:
		return e.ToResponse()
	case *MethodNotAllowed:
		return e.ToResponse()
	case *ParseError:
		return e.ToResponse()
	case *NotAcceptable:
		return e.ToResponse()
	case *UnsupportedMediaType:
		return e.ToResponse()
	case *APIException:
		return e.ToResponse()
	default:
		// Unknown error, return generic 500
		return &ErrorResponse{
			Error:   true,
			Code:    "internal_error",
			Message: "An internal error occurred",
			Details: nil,
		}
	}
}

// HandleExceptionHTTP handles an exception and writes HTTP response
func HandleExceptionHTTP(w http.ResponseWriter, r *http.Request, err error, handler ExceptionHandler) {
	if handler == nil {
		handler = NewDefaultExceptionHandler()
	}

	errorResp := handler.HandleException(err, r)

	// Determine status code from error type
	statusCode := http.StatusInternalServerError
	switch e := err.(type) {
	case *ValidationError:
		statusCode = e.Status
	case *AuthenticationFailed:
		statusCode = e.Status
	case *NotAuthenticated:
		statusCode = e.Status
	case *PermissionDenied:
		statusCode = e.Status
	case *NotFound:
		statusCode = e.Status
	case *Throttled:
		statusCode = e.Status
	case *MethodNotAllowed:
		statusCode = e.Status
	case *ParseError:
		statusCode = e.Status
	case *NotAcceptable:
		statusCode = e.Status
	case *UnsupportedMediaType:
		statusCode = e.Status
	case *APIException:
		statusCode = e.Status
	}

	// Handle throttled exception specially for Retry-After header
	if throttled, ok := err.(*Throttled); ok {
		retryAfter := int(throttled.RetryAfter.Seconds())
		if retryAfter > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		}
	}

	// Write response
	response := core.NewResponse(w)
	response.Status(statusCode).JSON(errorResp)
}

// Global exception handler
var globalHandler ExceptionHandler = NewDefaultExceptionHandler()

// SetExceptionHandler sets the global exception handler
func SetExceptionHandler(handler ExceptionHandler) {
	globalHandler = handler
}

// GetExceptionHandler returns the global exception handler
func GetExceptionHandler() ExceptionHandler {
	return globalHandler
}
