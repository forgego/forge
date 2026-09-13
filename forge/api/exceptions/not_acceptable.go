package exceptions

// NotAcceptable represents a content negotiation failure
type NotAcceptable struct {
	*APIException
}

// NewNotAcceptable creates a new not acceptable exception
func NewNotAcceptable(message string) *NotAcceptable {
	if message == "" {
		message = "Could not satisfy the request Accept header"
	}
	return &NotAcceptable{
		APIException: NewAPIException(
			406,
			"not_acceptable",
			message,
			nil,
		),
	}
}
