package authentication

import (
	"errors"
	"net/http"
	"strings"
)

// ErrInvalidToken is returned when a token is not recognized.
var ErrInvalidToken = errors.New("invalid token")

// TokenAuthentication authenticates requests using a token in the Authorization header
// Format: Authorization: Token <token>
type TokenAuthentication struct {
	// TokenLookup is a function that looks up a user by token
	// Should return (user, nil) if token is valid, (nil, nil) if not found, (nil, error) if error.
	// An unknown token (nil, nil) is rejected with ErrInvalidToken.
	TokenLookup func(token string) (interface{}, error)
}

// NewTokenAuthentication creates a new token authentication instance
func NewTokenAuthentication(tokenLookup func(token string) (interface{}, error)) *TokenAuthentication {
	return &TokenAuthentication{
		TokenLookup: tokenLookup,
	}
}

// Authenticate attempts to authenticate using token from Authorization header
func (a *TokenAuthentication) Authenticate(r *http.Request) (*AuthResult, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, nil // No auth header, not an error
	}

	// Parse "Token <token>" format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "token" {
		return nil, nil // Not our format, try next auth class
	}

	token := parts[1]
	if token == "" {
		return nil, nil
	}

	// Lookup user by token
	if a.TokenLookup == nil {
		return nil, nil // No lookup function configured
	}

	user, err := a.TokenLookup(token)
	if err != nil {
		return nil, err // Error during lookup
	}

	if user == nil {
		return nil, ErrInvalidToken // token supplied but not recognized: reject instead of falling through
	}

	// Success
	return NewAuthResult(user, token), nil
}

// AuthenticateHeader returns the WWW-Authenticate header value
func (a *TokenAuthentication) AuthenticateHeader(r *http.Request) string {
	return "Token"
}
