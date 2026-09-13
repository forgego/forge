package exceptions

// PermissionDenied represents permission denial
type PermissionDenied struct {
	*APIException
}

// NewPermissionDenied creates a new permission denied exception
func NewPermissionDenied(message string) *PermissionDenied {
	if message == "" {
		message = "You do not have permission to perform this action"
	}
	return &PermissionDenied{
		APIException: NewAPIException(
			403,
			"permission_denied",
			message,
			nil,
		),
	}
}
