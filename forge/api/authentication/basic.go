package authentication

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

// BasicAuthentication authenticates requests using HTTP Basic Authentication
// Format: Authorization: Basic <base64(username:password)>
type BasicAuthentication struct {
	// UserLookup is a function that looks up and validates a user by username and password
	// Should return (user, nil) if valid, (nil, nil) if not found, (nil, error) if error
	UserLookup func(username, password string) (interface{}, error)
}

// NewBasicAuthentication creates a new basic authentication instance
func NewBasicAuthentication(userLookup func(username, password string) (interface{}, error)) *BasicAuthentication {
	return &BasicAuthentication{
		UserLookup: userLookup,
	}
}

// Authenticate attempts to authenticate using Basic Auth from Authorization header
func (a *BasicAuthentication) Authenticate(r *http.Request) (*AuthResult, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	// Parse "Basic <credentials>" format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "basic" {
		return nil, nil
	}

	credentials := parts[1]
	if credentials == "" {
		return nil, nil
	}

	// Decode base64 credentials
	decoded, err := base64.StdEncoding.DecodeString(credentials)
	if err != nil {
		return nil, err
	}

	// Split username:password
	creds := strings.SplitN(string(decoded), ":", 2)
	if len(creds) != 2 {
		return nil, errors.New("invalid basic auth format")
	}

	username := creds[0]
	password := creds[1]

	if username == "" || password == "" {
		return nil, nil
	}

	// Lookup and validate user
	if a.UserLookup == nil {
		return nil, errors.New("user lookup function not configured")
	}

	user, err := a.UserLookup(username, password)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return NewAuthResult(user, nil), nil
}

// AuthenticateHeader returns the WWW-Authenticate header value
func (a *BasicAuthentication) AuthenticateHeader(r *http.Request) string {
	return "Basic realm=\"api\""
}
