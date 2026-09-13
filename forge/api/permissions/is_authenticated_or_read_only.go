package permissions

import (
	"net/http"

	"github.com/forgego/forge/api/authentication"
)

// IsAuthenticatedOrReadOnly allows read access to unauthenticated users
// but requires authentication for write operations
type IsAuthenticatedOrReadOnly struct{}

// NewIsAuthenticatedOrReadOnly creates a new IsAuthenticatedOrReadOnly permission
func NewIsAuthenticatedOrReadOnly() *IsAuthenticatedOrReadOnly {
	return &IsAuthenticatedOrReadOnly{}
}

// HasPermission checks if the request is authenticated or is a safe method
func (p *IsAuthenticatedOrReadOnly) HasPermission(r *http.Request, view ViewSet) bool {
	// Allow safe methods (GET, HEAD, OPTIONS) without authentication
	if IsSafeMethod(r.Method) {
		return true
	}
	// Require authentication for write operations
	_, ok := authentication.GetUserFromRequest(r)
	return ok
}

// HasObjectPermission checks if the request is authenticated or is a safe method
func (p *IsAuthenticatedOrReadOnly) HasObjectPermission(r *http.Request, view ViewSet, obj interface{}) bool {
	return p.HasPermission(r, view)
}

// GetMessage returns the error message
func (p *IsAuthenticatedOrReadOnly) GetMessage() string {
	return "Authentication credentials were not provided"
}

// GetCode returns the error code
func (p *IsAuthenticatedOrReadOnly) GetCode() string {
	return "not_authenticated"
}
