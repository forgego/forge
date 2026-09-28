package core

import (
	"context"
	"net/http"
)

// Request wraps http.Request with API-specific functionality
type Request struct {
	*http.Request
	user interface{} // Authenticated user
	auth interface{} // Authentication credentials (token, etc.)
}

// NewRequest creates a new API request from http.Request
func NewRequest(r *http.Request) *Request {
	return &Request{
		Request: r,
	}
}

// User returns the authenticated user
func (r *Request) User() interface{} {
	return r.user
}

// SetUser sets the authenticated user
func (r *Request) SetUser(user interface{}) {
	r.user = user
}

// Auth returns the authentication credentials
func (r *Request) Auth() interface{} {
	return r.auth
}

// SetAuth sets the authentication credentials
func (r *Request) SetAuth(auth interface{}) {
	r.auth = auth
}

// IsAuthenticated returns whether the request is authenticated
func (r *Request) IsAuthenticated() bool {
	return r.user != nil
}

// Context returns the request context
func (r *Request) Context() context.Context {
	return r.Request.Context()
}

// WithContext returns a new request with the given context
func (r *Request) WithContext(ctx context.Context) *Request {
	return &Request{
		Request: r.Request.WithContext(ctx),
		user:    r.user,
		auth:    r.auth,
	}
}
