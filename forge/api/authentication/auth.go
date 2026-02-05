package authentication

import (
	"net/http"

	"github.com/forgego/forge/api/core"
)

// Authentication is the interface for authentication classes
type Authentication interface {
	// Authenticate attempts to authenticate the request
	// Returns AuthResult if successful, nil if not applicable, error if failed
	Authenticate(r *http.Request) (*AuthResult, error)

	// AuthenticateHeader returns the WWW-Authenticate header value
	// Used for 401 responses to tell client how to authenticate
	AuthenticateHeader(r *http.Request) string
}

// AuthResult represents a successful authentication
type AuthResult struct {
	User interface{} // Authenticated user
	Auth interface{} // Authentication credentials (token, etc.)
}

// NewAuthResult creates a new authentication result
func NewAuthResult(user, auth interface{}) *AuthResult {
	return &AuthResult{
		User: user,
		Auth: auth,
	}
}

// AuthenticateRequest authenticates a request using a list of authentication classes
// Returns the first successful authentication result
func AuthenticateRequest(r *http.Request, authClasses []Authentication) (*AuthResult, error) {
	for _, auth := range authClasses {
		result, err := auth.Authenticate(r)
		if err != nil {
			// Authentication failed, continue to next
			continue
		}
		if result != nil {
			// Authentication successful
			return result, nil
		}
		// Authentication not applicable, try next
	}

	// No authentication succeeded
	return nil, nil
}

// SetUserOnRequest sets the authenticated user on the request
func SetUserOnRequest(r *http.Request, user interface{}) {
	ctx := core.WithUser(r.Context(), user)
	*r = *r.WithContext(ctx)
}

// SetAuthOnRequest sets authentication credentials on the request
func SetAuthOnRequest(r *http.Request, auth interface{}) {
	ctx := core.WithAuth(r.Context(), auth)
	*r = *r.WithContext(ctx)
}

// GetUserFromRequest retrieves the authenticated user from the request
func GetUserFromRequest(r *http.Request) (interface{}, bool) {
	return core.UserFromContext(r.Context())
}

// GetAuthFromRequest retrieves authentication credentials from the request
func GetAuthFromRequest(r *http.Request) (interface{}, bool) {
	return core.AuthFromContext(r.Context())
}
