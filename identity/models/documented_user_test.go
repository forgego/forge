package models

import "testing"

// Identity's user model carries login identity and admin flags.
func TestUserModelSupportsAuthenticationAndAdminRole(t *testing.T) {
	u := &User{ID: 42, Username: "owner", Email: "owner@example.invalid", IsActive: true, IsStaff: true, IsSuperuser: true}
	if !u.IsAuthenticated() || u.GetID() != 42 || u.GetDisplayName() == "" || !u.IsStaff || !u.IsSuperuser {
		t.Fatalf("user=%+v", u)
	}
}
