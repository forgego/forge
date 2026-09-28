package permissions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/stretchr/testify/assert"
)

type w0OwnerObject struct {
	UserID int64 `json:"user_id" db:"user_id"`
}

type w0OwnerUser struct {
	ID int64 `json:"id"`
}

type w0View struct{}

func (v *w0View) GetAction() string { return "test" }

func TestIsOwnerOrReadOnly_TaggedOwnerFields(t *testing.T) {
	perm := NewIsOwnerOrReadOnly("")
	obj := &w0OwnerObject{UserID: 7}
	view := &w0View{}

	tests := []struct {
		name          string
		method        string
		user          interface{}
		expectAllowed bool
	}{
		{
			name:          "POST with matching owner (ID 7) allowed",
			method:        http.MethodPost,
			user:          &w0OwnerUser{ID: 7},
			expectAllowed: true,
		},
		{
			name:          "POST with non-matching owner (ID 8) denied",
			method:        http.MethodPost,
			user:          &w0OwnerUser{ID: 8},
			expectAllowed: false,
		},
		{
			name:          "GET allowed without user",
			method:        http.MethodGet,
			user:          nil,
			expectAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/items/1", nil)
			if tt.user != nil {
				authentication.SetUserOnRequest(req, tt.user)
			}
			allowed := perm.HasPermission(req, view) && perm.HasObjectPermission(req, view, obj)
			assert.Equal(t, tt.expectAllowed, allowed)
		})
	}
}

type w0UserWithStaff struct {
	IsStaff bool
}

type w0UserWithSuperuser struct {
	IsSuperuser bool
}

type w0UserWithAdmin struct {
	IsAdmin bool
}

type w0UserWithAllFalse struct {
	IsAdmin     bool
	IsStaff     bool
	IsSuperuser bool
}

func TestIsAdminUser_StaffAndSuperuser(t *testing.T) {
	perm := NewIsAdminUser()
	view := &w0View{}
	dummyObj := map[string]interface{}{"id": 1}

	tests := []struct {
		name          string
		user          interface{}
		expectAllowed bool
	}{
		{
			name:          "user with IsStaff: true passes",
			user:          &w0UserWithStaff{IsStaff: true},
			expectAllowed: true,
		},
		{
			name:          "user with IsSuperuser: true passes",
			user:          &w0UserWithSuperuser{IsSuperuser: true},
			expectAllowed: true,
		},
		{
			name:          "user with IsAdmin: true passes",
			user:          &w0UserWithAdmin{IsAdmin: true},
			expectAllowed: true,
		},
		{
			name:          "user with all false fails",
			user:          &w0UserWithAllFalse{IsAdmin: false, IsStaff: false, IsSuperuser: false},
			expectAllowed: false,
		},
		{
			name:          "unauthenticated request fails",
			user:          nil,
			expectAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tt.user != nil {
				authentication.SetUserOnRequest(req, tt.user)
			}
			hasPerm := perm.HasPermission(req, view)
			hasObjPerm := perm.HasObjectPermission(req, view, dummyObj)
			assert.Equal(t, tt.expectAllowed, hasPerm)
			assert.Equal(t, tt.expectAllowed, hasObjPerm)
		})
	}
}
