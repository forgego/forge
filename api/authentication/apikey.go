package authentication

import (
	"net/http"
)

// APIKeyAuthentication authenticates requests using API keys
// Can be in header (X-API-Key) or query parameter (api_key)
type APIKeyAuthentication struct {
	// HeaderName is the header name for API key (default: "X-API-Key")
	HeaderName string
	// QueryParamName is the query parameter name (default: "api_key")
	QueryParamName string
	// KeyLookup is a function that looks up a user by API key
	// Should return (user, nil) if key is valid, (nil, nil) if not found, (nil, error) if error
	KeyLookup func(key string) (interface{}, error)
}

// NewAPIKeyAuthentication creates a new API key authentication instance
func NewAPIKeyAuthentication(keyLookup func(key string) (interface{}, error)) *APIKeyAuthentication {
	return &APIKeyAuthentication{
		HeaderName:     "X-API-Key",
		QueryParamName: "api_key",
		KeyLookup:      keyLookup,
	}
}

// Authenticate attempts to authenticate using API key
func (a *APIKeyAuthentication) Authenticate(r *http.Request) (*AuthResult, error) {
	var apiKey string

	// Try header first
	if a.HeaderName != "" {
		apiKey = r.Header.Get(a.HeaderName)
	}

	// Fall back to query parameter
	if apiKey == "" && a.QueryParamName != "" {
		apiKey = r.URL.Query().Get(a.QueryParamName)
	}

	if apiKey == "" {
		return nil, nil
	}

	// Lookup user by API key
	if a.KeyLookup == nil {
		return nil, nil
	}

	user, err := a.KeyLookup(apiKey)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return NewAuthResult(user, apiKey), nil
}

// AuthenticateHeader returns empty string (API keys don't use WWW-Authenticate)
func (a *APIKeyAuthentication) AuthenticateHeader(r *http.Request) string {
	return ""
}
