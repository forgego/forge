package exceptions

// MethodNotAllowed represents an HTTP method not allowed error
type MethodNotAllowed struct {
	*APIException
	AllowedMethods []string
}

// NewMethodNotAllowed creates a new method not allowed exception
func NewMethodNotAllowed(allowedMethods []string) *MethodNotAllowed {
	return &MethodNotAllowed{
		APIException: NewAPIException(
			405,
			"method_not_allowed",
			"Method not allowed",
			nil,
		),
		AllowedMethods: allowedMethods,
	}
}
