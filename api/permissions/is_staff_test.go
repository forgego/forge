package permissions

import (
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/stretchr/testify/assert"
)

type staffTestUser struct {
	Staff     bool
	Superuser bool
	Admin     bool
}

func (u *staffTestUser) GetID() string                            { return "1" }
func (u *staffTestUser) IsAuthenticated() bool                    { return true }
func (u *staffTestUser) IsStaff() bool                            { return u.Staff }
func (u *staffTestUser) IsSuperuser() bool                        { return u.Superuser }
func (u *staffTestUser) IsAdmin() bool                            { return u.Admin }
func (u *staffTestUser) HasPermission(permissionCode string) bool { return false }

func TestIsStaffUser(t *testing.T) {
	tests := []struct {
		name     string
		user     interface{}
		expected bool
	}{
		{"anonymous denied", nil, false},
		{"regular user denied", &staffTestUser{}, false},
		{"staff allowed", &staffTestUser{Staff: true}, true},
		{"superuser allowed", &staffTestUser{Superuser: true}, true},
		{"admin allowed", &staffTestUser{Admin: true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm := NewIsStaffUser()
			req := httptest.NewRequest("GET", "/test", nil)
			view := &MockViewSet{Action: "list"}
			if tt.user != nil {
				authentication.SetUserOnRequest(req, tt.user)
			}
			assert.Equal(t, tt.expected, perm.HasPermission(req, view))
			assert.Equal(t, tt.expected, perm.HasObjectPermission(req, view, nil))
		})
	}
}

func TestIsStaffUser_MessageAndCode(t *testing.T) {
	perm := NewIsStaffUser()
	assert.Equal(t, "You do not have permission to perform this action", perm.GetMessage())
	assert.Equal(t, "permission_denied", perm.GetCode())
}
