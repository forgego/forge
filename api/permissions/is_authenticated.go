package permissions

import (
	"net/http"

	"github.com/forgego/forge/api/authentication"
)

// IsAuthenticated requires authentication
type IsAuthenticated struct{}

// NewIsAuthenticated creates a new IsAuthenticated permission
func NewIsAuthenticated() *IsAuthenticated {
	return &IsAuthenticated{}
}

// HasPermission checks if the request is authenticated
func (p *IsAuthenticated) HasPermission(r *http.Request, view ViewSet) bool {
	_, ok := authentication.GetUserFromRequest(r)
	return ok
}

// HasObjectPermission checks if the request is authenticated
func (p *IsAuthenticated) HasObjectPermission(r *http.Request, view ViewSet, obj interface{}) bool {
	return p.HasPermission(r, view)
}

// GetMessage returns the error message
func (p *IsAuthenticated) GetMessage() string {
	return "Authentication credentials were not provided"
}

// GetCode returns the error code
func (p *IsAuthenticated) GetCode() string {
	return "not_authenticated"
}
