package exceptions

// NotFound represents a resource not found error
type NotFound struct {
	*APIException
}

// NewNotFound creates a new not found exception
func NewNotFound(message string) *NotFound {
	if message == "" {
		message = "Not found"
	}
	return &NotFound{
		APIException: NewAPIException(
			404,
			"not_found",
			message,
			nil,
		),
	}
}
