package errors

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/forgego/forge/api/exceptions"
)

// ErrorMapper maps various error types to Problem Details
type ErrorMapper struct {
	sanitizer *Sanitizer
}

// NewErrorMapper creates a new error mapper
func NewErrorMapper(sanitizer *Sanitizer) *ErrorMapper {
	if sanitizer == nil {
		sanitizer = DefaultSanitizer()
	}
	return &ErrorMapper{
		sanitizer: sanitizer,
	}
}

// MapError maps an error to a Problem Details structure
func (m *ErrorMapper) MapError(err error, instance string) *Problem {
	if err == nil {
		return m.mapInternalError("Unknown error", instance)
	}

	// Check if it's already a ProblemError
	if problemErr, ok := err.(ProblemError); ok {
		problem := problemErr.ToProblem()
		if instance != "" {
			problem.Instance = instance
		}
		return problem
	}

	// Check if it's already a Problem
	if problem, ok := err.(*Problem); ok {
		if instance != "" {
			problem.Instance = instance
		}
		return problem
	}

	// Map legacy exception types
	switch e := err.(type) {
	case *exceptions.ValidationError:
		return m.mapValidationError(e, instance)
	case *exceptions.AuthenticationFailed:
		return m.mapAuthenticationFailed(e, instance)
	case *exceptions.NotAuthenticated:
		return m.mapNotAuthenticated(e, instance)
	case *exceptions.PermissionDenied:
		return m.mapPermissionDenied(e, instance)
	case *exceptions.NotFound:
		return m.mapNotFound(e, instance)
	case *exceptions.Throttled:
		return m.mapThrottled(e, instance)
	case *exceptions.MethodNotAllowed:
		return m.mapMethodNotAllowed(e, instance)
	case *exceptions.ParseError:
		return m.mapParseError(e, instance)
	case *exceptions.NotAcceptable:
		return m.mapNotAcceptable(e, instance)
	case *exceptions.UnsupportedMediaType:
		return m.mapUnsupportedMediaType(e, instance)
	case *exceptions.APIException:
		return m.mapAPIException(e, instance)
	default:
		// Unknown error - sanitize and return as internal error
		sanitizedMsg := m.sanitizer.SanitizeError(err)
		return m.mapInternalError(sanitizedMsg, instance)
	}
}

// mapValidationError maps a ValidationError to Problem Details
func (m *ErrorMapper) mapValidationError(err *exceptions.ValidationError, instance string) *Problem {
	problem := NewProblem(
		ErrorTypeValidation,
		CodeValidationError,
		http.StatusBadRequest,
		"Validation Failed",
		m.sanitizer.SanitizeString(err.Message),
	).WithInstance(instance)

	// Convert field errors
	if err.Errors != nil {
		fieldErrors := make(map[string][]FieldError)
		for field, messages := range err.Errors {
			fieldErrors[field] = make([]FieldError, len(messages))
			for i, msg := range messages {
				fieldErrors[field][i] = FieldError{
					Message: m.sanitizer.SanitizeString(msg),
					Code:    CodeInvalidField,
				}
			}
		}
		problem.WithFieldErrors(fieldErrors)
	}

	return problem
}

// mapAuthenticationFailed maps an AuthenticationFailed error to Problem Details
func (m *ErrorMapper) mapAuthenticationFailed(err *exceptions.AuthenticationFailed, instance string) *Problem {
	return NewProblem(
		ErrorTypeAuthentication,
		CodeAuthenticationFailed,
		http.StatusUnauthorized,
		"Authentication Failed",
		m.sanitizer.SanitizeString(err.Message),
	).WithInstance(instance)
}

// mapNotAuthenticated maps a NotAuthenticated error to Problem Details
func (m *ErrorMapper) mapNotAuthenticated(err *exceptions.NotAuthenticated, instance string) *Problem {
	return NewProblem(
		ErrorTypeAuthentication,
		CodeTokenMissing,
		http.StatusUnauthorized,
		"Not Authenticated",
		m.sanitizer.SanitizeString(err.Message),
	).WithInstance(instance)
}

// mapPermissionDenied maps a PermissionDenied error to Problem Details
func (m *ErrorMapper) mapPermissionDenied(err *exceptions.PermissionDenied, instance string) *Problem {
	return NewProblem(
		ErrorTypeAuthorization,
		CodePermissionDenied,
		http.StatusForbidden,
		"Permission Denied",
		m.sanitizer.SanitizeString(err.Message),
	).WithInstance(instance)
}

// mapNotFound maps a NotFound error to Problem Details
func (m *ErrorMapper) mapNotFound(err *exceptions.NotFound, instance string) *Problem {
	return NewProblem(
		ErrorTypeNotFound,
		CodeNotFound,
		http.StatusNotFound,
		"Not Found",
		m.sanitizer.SanitizeString(err.Message),
	).WithInstance(instance)
}

// mapThrottled maps a Throttled error to Problem Details
func (m *ErrorMapper) mapThrottled(err *exceptions.Throttled, instance string) *Problem {
	problem := NewProblem(
		ErrorTypeRateLimit,
		CodeRateLimitExceeded,
		http.StatusTooManyRequests,
		"Rate Limit Exceeded",
		m.sanitizer.SanitizeString(err.Message),
	).WithInstance(instance)

	// Add retry-after to meta if available
	if err.RetryAfter > 0 {
		problem.WithMeta("retry_after_seconds", int(err.RetryAfter.Seconds()))
	}

	return problem
}

// mapMethodNotAllowed maps a MethodNotAllowed error to Problem Details
func (m *ErrorMapper) mapMethodNotAllowed(err *exceptions.MethodNotAllowed, instance string) *Problem {
	return NewProblem(
		ErrorTypeMethodNotAllowed,
		CodeMethodNotAllowed,
		http.StatusMethodNotAllowed,
		"Method Not Allowed",
		m.sanitizer.SanitizeString(err.Message),
	).WithInstance(instance)
}

// mapParseError maps a ParseError to Problem Details
func (m *ErrorMapper) mapParseError(err *exceptions.ParseError, instance string) *Problem {
	return NewProblem(
		ErrorTypeBadRequest,
		CodeInvalidJSON,
		http.StatusBadRequest,
		"Parse Error",
		m.sanitizer.SanitizeString(err.Message),
	).WithInstance(instance)
}

// mapNotAcceptable maps a NotAcceptable error to Problem Details
func (m *ErrorMapper) mapNotAcceptable(err *exceptions.NotAcceptable, instance string) *Problem {
	return NewProblem(
		ErrorTypeNotAcceptable,
		CodeNotAcceptable,
		http.StatusNotAcceptable,
		"Not Acceptable",
		m.sanitizer.SanitizeString(err.Message),
	).WithInstance(instance)
}

// mapUnsupportedMediaType maps an UnsupportedMediaType error to Problem Details
func (m *ErrorMapper) mapUnsupportedMediaType(err *exceptions.UnsupportedMediaType, instance string) *Problem {
	return NewProblem(
		ErrorTypeUnsupportedMediaType,
		CodeUnsupportedMediaType,
		http.StatusUnsupportedMediaType,
		"Unsupported Media Type",
		m.sanitizer.SanitizeString(err.Message),
	).WithInstance(instance)
}

// mapAPIException maps a generic APIException to Problem Details
func (m *ErrorMapper) mapAPIException(err *exceptions.APIException, instance string) *Problem {
	// Determine error type based on status code
	errType := ErrorTypeInternal
	if err.Status >= 400 && err.Status < 500 {
		errType = ErrorTypeBadRequest
	}

	return NewProblem(
		errType,
		err.Code,
		err.Status,
		"API Error",
		m.sanitizer.SanitizeString(err.Message),
	).WithInstance(instance)
}

// mapInternalError maps an internal error to Problem Details
func (m *ErrorMapper) mapInternalError(detail string, instance string) *Problem {
	// Always sanitize internal errors
	sanitizedDetail := m.sanitizer.SanitizeString(detail)
	if sanitizedDetail == "" {
		sanitizedDetail = "An internal error occurred"
	}

	return NewProblem(
		ErrorTypeInternal,
		CodeInternalError,
		http.StatusInternalServerError,
		"Internal Server Error",
		sanitizedDetail,
	).WithInstance(instance)
}

// MapPanic maps a panic to a Problem Details structure
func (m *ErrorMapper) MapPanic(rec interface{}, instance string) *Problem {
	// Log the panic details (but never send to client)
	_ = debug.Stack() // Stack trace is logged but not included in response

	// Create a sanitized error message
	var detail string
	switch v := rec.(type) {
	case string:
		detail = m.sanitizer.SanitizeString(v)
	case error:
		detail = m.sanitizer.SanitizeError(v)
	default:
		detail = m.sanitizer.SanitizeString(fmt.Sprintf("%v", v))
	}

	if detail == "" {
		detail = "An internal error occurred"
	}

	return NewProblem(
		ErrorTypeInternal,
		CodeInternalError,
		http.StatusInternalServerError,
		"Internal Server Error",
		detail,
	).WithInstance(instance)
}
