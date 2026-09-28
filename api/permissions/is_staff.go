package permissions

import (
	"net/http"

	"github.com/forgego/forge/api/authentication"
)

// IsStaffUser requires the user to be staff or a superuser.
// It accepts users exposing IsStaff, IsSuperuser or IsAdmin (bool method or
// field), matching the existing IsAdminUser detection via reflection.
type IsStaffUser struct{}

// NewIsStaffUser creates a new IsStaffUser permission.
func NewIsStaffUser() *IsStaffUser {
	return &IsStaffUser{}
}

// HasPermission checks if the request user is staff or a superuser.
func (p *IsStaffUser) HasPermission(r *http.Request, view ViewSet) bool {
	user, ok := authentication.GetUserFromRequest(r)
	if !ok {
		return false
	}
	return isAdmin(user)
}

// HasObjectPermission checks if the request user is staff or a superuser.
func (p *IsStaffUser) HasObjectPermission(r *http.Request, view ViewSet, obj interface{}) bool {
	return p.HasPermission(r, view)
}

// GetMessage returns the error message.
func (p *IsStaffUser) GetMessage() string {
	return "You do not have permission to perform this action"
}

// GetCode returns the error code.
func (p *IsStaffUser) GetCode() string {
	return "permission_denied"
}
