package permissions

import (
	"net/http"

	"github.com/forgego/forge/identity/principal"
)

// IsStaffUser requires the user to be staff or a superuser.
// It accepts users exposing IsStaff, IsSuperuser or IsAdmin (bool method or
// field), matching the existing IsAdminUser detection via reflection.
type IsStaffUser struct{}

func (*IsStaffUser) HasObjectRules(*http.Request) bool { return false }

// NewIsStaffUser creates a new IsStaffUser permission.
func NewIsStaffUser() *IsStaffUser {
	return &IsStaffUser{}
}

// HasPermission checks if the request user is staff or a superuser.
func (p *IsStaffUser) HasPermission(r *http.Request, view ViewSet) bool {
	if actor, ok := principal.PrincipalFromContext(r.Context()); ok {
		if actor.Kind != principal.KindUser || actor.ID == "" {
			return false
		}
		for _, role := range actor.Roles() {
			if role == "staff" || role == "superuser" || role == "admin" {
				return true
			}
		}
		return false
	}
	user, ok := activeLegacyUser(r)
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
