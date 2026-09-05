package permissions

import (
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/stretchr/testify/assert"
)

// MockViewSet is a test viewset implementation
type MockViewSet struct {
	Action string
}

func (m *MockViewSet) GetAction() string {
	return m.Action
}

// MockUser is a test user
type MockUser struct {
	ID            string
	Authenticated bool
	Staff         bool
	Superuser     bool
	Admin         bool
}

func (u *MockUser) GetID() string                            { return u.ID }
func (u *MockUser) IsAuthenticated() bool                    { return u.Authenticated }
func (u *MockUser) IsStaff() bool                            { return u.Staff }
func (u *MockUser) IsSuperuser() bool                        { return u.Superuser }
func (u *MockUser) IsAdmin() bool                            { return u.Admin }
func (u *MockUser) HasPermission(permissionCode string) bool { return false }

func TestAllowAny(t *testing.T) {
	perm := NewAllowAny()
	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{Action: "list"}

	// Should always allow
	assert.True(t, perm.HasPermission(req, view))
	assert.True(t, perm.HasObjectPermission(req, view, nil))

	// Test with authenticated user
	authUser := &MockUser{Authenticated: true}
	authentication.SetUserOnRequest(req, authUser)
	assert.True(t, perm.HasPermission(req, view))
}

func TestIsAuthenticated_Authenticated(t *testing.T) {
	perm := NewIsAuthenticated()
	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{Action: "list"}

	authUser := &MockUser{Authenticated: true}
	authentication.SetUserOnRequest(req, authUser)

	assert.True(t, perm.HasPermission(req, view))
	assert.True(t, perm.HasObjectPermission(req, view, nil))
}

func TestIsAuthenticated_NotAuthenticated(t *testing.T) {
	perm := NewIsAuthenticated()
	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{Action: "list"}

	// No user set
	assert.False(t, perm.HasPermission(req, view))
	assert.False(t, perm.HasObjectPermission(req, view, nil))
}

func TestIsAuthenticatedOrReadOnly_ReadMethod(t *testing.T) {
	perm := NewIsAuthenticatedOrReadOnly()
	view := &MockViewSet{Action: "list"}

	testCases := []struct {
		method   string
		expected bool
	}{
		{"GET", true},
		{"HEAD", true},
		{"OPTIONS", true},
		{"POST", false},
		{"PUT", false},
		{"PATCH", false},
		{"DELETE", false},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/test", nil)
			// No user set (not authenticated)
			assert.Equal(t, tc.expected, perm.HasPermission(req, view))
		})
	}
}

func TestIsAuthenticatedOrReadOnly_WriteMethod_Authenticated(t *testing.T) {
	perm := NewIsAuthenticatedOrReadOnly()
	req := httptest.NewRequest("POST", "/test", nil)
	view := &MockViewSet{Action: "create"}

	authUser := &MockUser{Authenticated: true}
	authentication.SetUserOnRequest(req, authUser)

	assert.True(t, perm.HasPermission(req, view))
}

func TestIsAdminUser_Admin(t *testing.T) {
	perm := NewIsAdminUser()
	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{Action: "list"}

	adminUser := &MockUser{Authenticated: true, Admin: true}
	authentication.SetUserOnRequest(req, adminUser)

	assert.True(t, perm.HasPermission(req, view))
	assert.True(t, perm.HasObjectPermission(req, view, map[string]interface{}{"id": "1"}))
}

func TestIsAdminUser_NotAdmin(t *testing.T) {
	perm := NewIsAdminUser()
	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{Action: "list"}

	regularUser := &MockUser{Authenticated: true, Admin: false}
	authentication.SetUserOnRequest(req, regularUser)

	assert.False(t, perm.HasPermission(req, view))
}

func TestIsOwnerOrReadOnly_ReadMethod(t *testing.T) {
	perm := NewIsOwnerOrReadOnly("user_id")
	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{Action: "list"}

	// Read methods should always be allowed
	assert.True(t, perm.HasPermission(req, view))
}

func TestIsOwnerOrReadOnly_Owner(t *testing.T) {
	perm := NewIsOwnerOrReadOnly("UserID")
	req := httptest.NewRequest("PUT", "/test", nil)
	view := &MockViewSet{Action: "update"}

	user := &MockUser{ID: "123", Authenticated: true}
	authentication.SetUserOnRequest(req, user)

	// Mock object with matching user_id (using struct for reflection)
	// Field name must match exactly what's passed to NewIsOwnerOrReadOnly
	type MockObject struct {
		UserID string
	}
	obj := &MockObject{UserID: "123"}

	assert.True(t, perm.HasObjectPermission(req, view, obj))
}

func TestIsOwnerOrReadOnly_NotOwner(t *testing.T) {
	perm := NewIsOwnerOrReadOnly("user_id")
	req := httptest.NewRequest("PUT", "/test", nil)
	view := &MockViewSet{Action: "update"}

	user := &MockUser{ID: "123", Authenticated: true}
	authentication.SetUserOnRequest(req, user)

	// Mock object with different user_id
	obj := map[string]interface{}{
		"user_id": "456",
	}

	assert.False(t, perm.HasObjectPermission(req, view, obj))
}

func TestCheckPermissions_AllPass(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{Action: "list"}

	authUser := &MockUser{Authenticated: true}
	authentication.SetUserOnRequest(req, authUser)

	perms := []Permission{
		NewIsAuthenticated(),
		NewAllowAny(),
	}

	assert.True(t, CheckPermissions(req, view, perms))
}

func TestCheckPermissions_OneFails(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{Action: "list"}

	// No user set
	perms := []Permission{
		NewIsAuthenticated(),
		NewAllowAny(),
	}

	assert.False(t, CheckPermissions(req, view, perms))
}

func TestCheckObjectPermissions_AllPass(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{Action: "retrieve"}

	user := &MockUser{ID: "123", Authenticated: true}
	authentication.SetUserOnRequest(req, user)

	obj := map[string]interface{}{"user_id": "123"}

	perms := []Permission{
		NewIsAuthenticated(),
		NewIsOwnerOrReadOnly("user_id"),
	}

	assert.True(t, CheckObjectPermissions(req, view, obj, perms))
}

func TestPermissionMessages(t *testing.T) {
	tests := []struct {
		name     string
		perm     Permission
		expected string
	}{
		{"AllowAny", NewAllowAny(), ""},
		{"IsAuthenticated", NewIsAuthenticated(), "Authentication credentials were not provided"},
		{"IsAuthenticatedOrReadOnly", NewIsAuthenticatedOrReadOnly(), "Authentication credentials were not provided"},
		{"IsAdminUser", NewIsAdminUser(), "You do not have permission to perform this action"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.perm.GetMessage())
		})
	}
}

