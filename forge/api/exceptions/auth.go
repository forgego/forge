package exceptions

// AuthenticationFailed represents authentication failure
type AuthenticationFailed struct {
	*APIException
}

// NewAuthenticationFailed creates a new authentication failed exception
func NewAuthenticationFailed(message string) *AuthenticationFailed {
	if message == "" {
		message = "Authentication failed"
	}
	return &AuthenticationFailed{
		APIException: NewAPIException(
			401,
			"authentication_failed",
			message,
			nil,
		),
	}
}

// NotAuthenticated represents unauthenticated request
type NotAuthenticated struct {
	*APIException
}

// NewNotAuthenticated creates a new not authenticated exception
func NewNotAuthenticated(message string) *NotAuthenticated {
	if message == "" {
		message = "Authentication credentials were not provided"
	}
	return &NotAuthenticated{
		APIException: NewAPIException(
			401,
			"not_authenticated",
			message,
			nil,
		),
	}
}
