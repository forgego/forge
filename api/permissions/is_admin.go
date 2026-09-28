package permissions

import (
	"net/http"

	"github.com/forgego/forge/api/authentication"
)

// IsAdminUser requires the user to be an admin
// Assumes user has an IsAdmin() bool method or IsAdmin field
type IsAdminUser struct{}

// NewIsAdminUser creates a new IsAdminUser permission
func NewIsAdminUser() *IsAdminUser {
	return &IsAdminUser{}
}

// HasPermission checks if the user is authenticated and is an admin
func (p *IsAdminUser) HasPermission(r *http.Request, view ViewSet) bool {
	user, ok := authentication.GetUserFromRequest(r)
	if !ok {
		return false
	}

	return isAdmin(user)
}

// HasObjectPermission checks if the user is authenticated and is an admin
func (p *IsAdminUser) HasObjectPermission(r *http.Request, view ViewSet, obj interface{}) bool {
	return p.HasPermission(r, view)
}

// GetMessage returns the error message
func (p *IsAdminUser) GetMessage() string {
	return "You do not have permission to perform this action"
}

// GetCode returns the error code
func (p *IsAdminUser) GetCode() string {
	return "permission_denied"
}

// boolMember checks if obj has a bool method or field with the given name
func boolMember(obj interface{}, name string) (value, found bool) {
	if m := getMethod(obj, name); m.IsValid() && m.Type().NumIn() == 0 {
		results := m.Call(nil)
		if len(results) > 0 {
			if v, ok := results[0].Interface().(bool); ok {
				return v, true
			}
		}
	}

	if f := getField(obj, name); f.IsValid() {
		if v, ok := f.Interface().(bool); ok {
			return v, true
		}
	}

	return false, false
}

// isAdmin checks if a user is an admin using reflection
func isAdmin(user interface{}) bool {
	for _, name := range []string{"IsAdmin", "IsStaff", "IsSuperuser"} {
		if val, found := boolMember(user, name); found && val {
			return true
		}
	}

	return false
}
