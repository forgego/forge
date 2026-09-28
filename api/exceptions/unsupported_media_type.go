package exceptions

// UnsupportedMediaType represents an unsupported content type error
type UnsupportedMediaType struct {
	*APIException
}

// NewUnsupportedMediaType creates a new unsupported media type exception
func NewUnsupportedMediaType(message string) *UnsupportedMediaType {
	if message == "" {
		message = "Unsupported media type in request"
	}
	return &UnsupportedMediaType{
		APIException: NewAPIException(
			415,
			"unsupported_media_type",
			message,
			nil,
		),
	}
}
