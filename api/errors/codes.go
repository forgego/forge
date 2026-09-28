package errors

import (
	"fmt"
	"sync"
)

// ErrorCode represents an application-level error code
type ErrorCode struct {
	Code        string
	Description string
	HTTPStatus  int
	Type        ErrorType
	Version     string // API version this code belongs to
}

// ErrorCodeRegistry maintains a registry of all error codes
type ErrorCodeRegistry struct {
	codes map[string]*ErrorCode
	mu    sync.RWMutex
}

var (
	globalRegistry = &ErrorCodeRegistry{
		codes: make(map[string]*ErrorCode),
	}
)

// RegisterErrorCode registers a new error code
func RegisterErrorCode(code *ErrorCode) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if _, exists := globalRegistry.codes[code.Code]; exists {
		return fmt.Errorf("error code %s already registered", code.Code)
	}

	globalRegistry.codes[code.Code] = code
	return nil
}

// GetErrorCode retrieves an error code by its code string
func GetErrorCode(code string) (*ErrorCode, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	ec, exists := globalRegistry.codes[code]
	return ec, exists
}

// GetAllErrorCodes returns all registered error codes
func GetAllErrorCodes() []*ErrorCode {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	codes := make([]*ErrorCode, 0, len(globalRegistry.codes))
	for _, code := range globalRegistry.codes {
		codes = append(codes, code)
	}
	return codes
}

// GetErrorCodesByVersion returns all error codes for a specific API version
func GetErrorCodesByVersion(version string) []*ErrorCode {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	codes := make([]*ErrorCode, 0)
	for _, code := range globalRegistry.codes {
		if code.Version == version {
			codes = append(codes, code)
		}
	}
	return codes
}

// Common error codes - these are registered automatically
const (
	// Validation errors
	CodeValidationError = "VALIDATION_ERROR"
	CodeInvalidField    = "INVALID_FIELD"
	CodeMissingField    = "MISSING_FIELD"
	CodeInvalidFormat   = "INVALID_FORMAT"
	CodeValueTooShort   = "VALUE_TOO_SHORT"
	CodeValueTooLong    = "VALUE_TOO_LONG"
	CodeInvalidEmail    = "INVALID_EMAIL"
	CodeInvalidURL      = "INVALID_URL"
	CodeInvalidUUID     = "INVALID_UUID"
	CodeInvalidDate     = "INVALID_DATE"
	CodeInvalidNumber   = "INVALID_NUMBER"
	CodeValueOutOfRange = "VALUE_OUT_OF_RANGE"
	CodeDuplicateValue  = "DUPLICATE_VALUE"

	// Authentication errors
	CodeAuthenticationFailed = "AUTHENTICATION_FAILED"
	CodeInvalidCredentials   = "INVALID_CREDENTIALS"
	CodeTokenExpired         = "TOKEN_EXPIRED"
	CodeTokenInvalid         = "TOKEN_INVALID"
	CodeTokenMissing         = "TOKEN_MISSING"
	CodeSessionExpired       = "SESSION_EXPIRED"

	// Authorization errors
	CodePermissionDenied        = "PERMISSION_DENIED"
	CodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	CodeForbidden               = "FORBIDDEN"
	CodeNotOwner                = "NOT_OWNER"

	// Not found errors
	CodeNotFound         = "NOT_FOUND"
	CodeResourceNotFound = "RESOURCE_NOT_FOUND"
	CodeUserNotFound     = "USER_NOT_FOUND"
	CodeRecordNotFound   = "RECORD_NOT_FOUND"

	// Conflict errors
	CodeConflict          = "CONFLICT"
	CodeDuplicateResource = "DUPLICATE_RESOURCE"
	CodeResourceExists    = "RESOURCE_EXISTS"

	// Rate limit errors
	CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	CodeTooManyRequests   = "TOO_MANY_REQUESTS"

	// Internal errors
	CodeInternalError      = "INTERNAL_ERROR"
	CodeDatabaseError      = "DATABASE_ERROR"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeTimeout            = "TIMEOUT"

	// Bad request errors
	CodeBadRequest           = "BAD_REQUEST"
	CodeInvalidJSON          = "INVALID_JSON"
	CodeInvalidXML           = "INVALID_XML"
	CodeInvalidContentType   = "INVALID_CONTENT_TYPE"
	CodeMissingRequiredField = "MISSING_REQUIRED_FIELD"

	// Method errors
	CodeMethodNotAllowed     = "METHOD_NOT_ALLOWED"
	CodeNotAcceptable        = "NOT_ACCEPTABLE"
	CodeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
)

// init registers common error codes
func init() {
	// Validation errors
	registerCode(CodeValidationError, "Validation failed", 400, ErrorTypeValidation)
	registerCode(CodeInvalidField, "Invalid field value", 400, ErrorTypeValidation)
	registerCode(CodeMissingField, "Required field is missing", 400, ErrorTypeValidation)
	registerCode(CodeInvalidFormat, "Invalid format", 400, ErrorTypeValidation)
	registerCode(CodeValueTooShort, "Value is too short", 400, ErrorTypeValidation)
	registerCode(CodeValueTooLong, "Value is too long", 400, ErrorTypeValidation)
	registerCode(CodeInvalidEmail, "Invalid email format", 400, ErrorTypeValidation)
	registerCode(CodeInvalidURL, "Invalid URL format", 400, ErrorTypeValidation)
	registerCode(CodeInvalidUUID, "Invalid UUID format", 400, ErrorTypeValidation)
	registerCode(CodeInvalidDate, "Invalid date format", 400, ErrorTypeValidation)
	registerCode(CodeInvalidNumber, "Invalid number format", 400, ErrorTypeValidation)
	registerCode(CodeValueOutOfRange, "Value is out of range", 400, ErrorTypeValidation)
	registerCode(CodeDuplicateValue, "Duplicate value not allowed", 400, ErrorTypeValidation)

	// Authentication errors
	registerCode(CodeAuthenticationFailed, "Authentication failed", 401, ErrorTypeAuthentication)
	registerCode(CodeInvalidCredentials, "Invalid credentials", 401, ErrorTypeAuthentication)
	registerCode(CodeTokenExpired, "Token has expired", 401, ErrorTypeAuthentication)
	registerCode(CodeTokenInvalid, "Invalid token", 401, ErrorTypeAuthentication)
	registerCode(CodeTokenMissing, "Token is missing", 401, ErrorTypeAuthentication)
	registerCode(CodeSessionExpired, "Session has expired", 401, ErrorTypeAuthentication)

	// Authorization errors
	registerCode(CodePermissionDenied, "Permission denied", 403, ErrorTypeAuthorization)
	registerCode(CodeInsufficientPermissions, "Insufficient permissions", 403, ErrorTypeAuthorization)
	registerCode(CodeForbidden, "Forbidden", 403, ErrorTypeAuthorization)
	registerCode(CodeNotOwner, "Not the owner of the resource", 403, ErrorTypeAuthorization)

	// Not found errors
	registerCode(CodeNotFound, "Resource not found", 404, ErrorTypeNotFound)
	registerCode(CodeResourceNotFound, "Resource not found", 404, ErrorTypeNotFound)
	registerCode(CodeUserNotFound, "User not found", 404, ErrorTypeNotFound)
	registerCode(CodeRecordNotFound, "Record not found", 404, ErrorTypeNotFound)

	// Conflict errors
	registerCode(CodeConflict, "Conflict", 409, ErrorTypeConflict)
	registerCode(CodeDuplicateResource, "Duplicate resource", 409, ErrorTypeConflict)
	registerCode(CodeResourceExists, "Resource already exists", 409, ErrorTypeConflict)

	// Rate limit errors
	registerCode(CodeRateLimitExceeded, "Rate limit exceeded", 429, ErrorTypeRateLimit)
	registerCode(CodeTooManyRequests, "Too many requests", 429, ErrorTypeRateLimit)

	// Internal errors
	registerCode(CodeInternalError, "Internal server error", 500, ErrorTypeInternal)
	registerCode(CodeDatabaseError, "Database error", 500, ErrorTypeInternal)
	registerCode(CodeServiceUnavailable, "Service unavailable", 503, ErrorTypeInternal)
	registerCode(CodeTimeout, "Request timeout", 504, ErrorTypeInternal)

	// Bad request errors
	registerCode(CodeBadRequest, "Bad request", 400, ErrorTypeBadRequest)
	registerCode(CodeInvalidJSON, "Invalid JSON", 400, ErrorTypeBadRequest)
	registerCode(CodeInvalidXML, "Invalid XML", 400, ErrorTypeBadRequest)
	registerCode(CodeInvalidContentType, "Invalid content type", 400, ErrorTypeBadRequest)
	registerCode(CodeMissingRequiredField, "Missing required field", 400, ErrorTypeBadRequest)

	// Method errors
	registerCode(CodeMethodNotAllowed, "Method not allowed", 405, ErrorTypeMethodNotAllowed)
	registerCode(CodeNotAcceptable, "Not acceptable", 406, ErrorTypeNotAcceptable)
	registerCode(CodeUnsupportedMediaType, "Unsupported media type", 415, ErrorTypeUnsupportedMediaType)
}

// registerCode is a helper to register error codes
func registerCode(code, description string, httpStatus int, errType ErrorType) {
	_ = RegisterErrorCode(&ErrorCode{
		Code:        code,
		Description: description,
		HTTPStatus:  httpStatus,
		Type:        errType,
		Version:     "v1", // Default version
	})
}

// RegisterErrorCodeWithVersion registers an error code with a specific API version
func RegisterErrorCodeWithVersion(code, description string, httpStatus int, errType ErrorType, version string) error {
	return RegisterErrorCode(&ErrorCode{
		Code:        code,
		Description: description,
		HTTPStatus:  httpStatus,
		Type:        errType,
		Version:     version,
	})
}
