package permissions

import (
	"net/http"
)

// AllowAny allows unrestricted access
type AllowAny struct{}

// NewAllowAny creates a new AllowAny permission
func NewAllowAny() *AllowAny {
	return &AllowAny{}
}

// HasPermission always returns true
func (p *AllowAny) HasPermission(r *http.Request, view ViewSet) bool {
	return true
}

// HasObjectPermission always returns true
func (p *AllowAny) HasObjectPermission(r *http.Request, view ViewSet, obj interface{}) bool {
	return true
}

// GetMessage returns empty string (never denied)
func (p *AllowAny) GetMessage() string {
	return ""
}

// GetCode returns empty string
func (p *AllowAny) GetCode() string {
	return ""
}
