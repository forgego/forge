package exceptions

// ParseError represents a request parsing error
type ParseError struct {
	*APIException
}

// NewParseError creates a new parse error exception
func NewParseError(message string) *ParseError {
	if message == "" {
		message = "Malformed request"
	}
	return &ParseError{
		APIException: NewAPIException(
			400,
			"parse_error",
			message,
			nil,
		),
	}
}
