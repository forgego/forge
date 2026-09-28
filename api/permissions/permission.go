package permissions

import (
	"net/http"
)

// ViewSet is a minimal interface for permission checking
// This avoids circular dependencies - full ViewSet interface is in pkg/api
type ViewSet interface {
	// GetAction returns the current action name
	GetAction() string
}

// Permission is the interface for permission classes
type Permission interface {
	// HasPermission checks if the request has permission for the view
	// Called before the view action is executed
	HasPermission(r *http.Request, view ViewSet) bool

	// HasObjectPermission checks if the request has permission for a specific object
	// Called after the object is retrieved, before modification
	HasObjectPermission(r *http.Request, view ViewSet, obj interface{}) bool

	// GetMessage returns the error message if permission is denied
	GetMessage() string

	// GetCode returns the error code if permission is denied
	GetCode() string
}

// CheckPermissions checks a list of permissions
// Returns true if all permissions pass, false otherwise
func CheckPermissions(r *http.Request, view ViewSet, perms []Permission) bool {
	for _, perm := range perms {
		if !perm.HasPermission(r, view) {
			return false
		}
	}
	return true
}

// CheckObjectPermissions checks object-level permissions
// Returns true if all permissions pass, false otherwise
func CheckObjectPermissions(r *http.Request, view ViewSet, obj interface{}, perms []Permission) bool {
	for _, perm := range perms {
		if !perm.HasObjectPermission(r, view, obj) {
			return false
		}
	}
	return true
}

// SAFE_METHODS are HTTP methods that are considered "safe" (read-only)
var SAFE_METHODS = []string{"GET", "HEAD", "OPTIONS"}

// IsSafeMethod checks if an HTTP method is safe
func IsSafeMethod(method string) bool {
	for _, m := range SAFE_METHODS {
		if m == method {
			return true
		}
	}
	return false
}
