package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockUser is a test user implementation
type MockUser struct {
	ID            string
	Authenticated bool
	Staff         bool
	Superuser     bool
}

func (u *MockUser) GetID() string                            { return u.ID }
func (u *MockUser) IsAuthenticated() bool                    { return u.Authenticated }
func (u *MockUser) IsStaff() bool                            { return u.Staff }
func (u *MockUser) IsSuperuser() bool                        { return u.Superuser }
func (u *MockUser) HasPermission(permissionCode string) bool { return false }

// MockAuthentication is a test authentication implementation
type MockAuthentication struct {
	ShouldAuthenticate bool
	User               interface{}
	Auth               interface{}
	Error              error
}

func (m *MockAuthentication) Authenticate(r *http.Request) (*AuthResult, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	if m.ShouldAuthenticate {
		return NewAuthResult(m.User, m.Auth), nil
	}
	return nil, nil
}

func (m *MockAuthentication) AuthenticateHeader(r *http.Request) string {
	return "Mock"
}

func TestAuthenticateRequest_Success(t *testing.T) {
	mockUser := &MockUser{ID: "123", Authenticated: true}
	mockAuth := &MockAuthentication{
		ShouldAuthenticate: true,
		User:               mockUser,
		Auth:               "token123",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	authClasses := []Authentication{mockAuth}

	result, err := AuthenticateRequest(req, authClasses)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, mockUser, result.User)
	assert.Equal(t, "token123", result.Auth)
}

func TestAuthenticateRequest_NoAuth(t *testing.T) {
	mockAuth := &MockAuthentication{
		ShouldAuthenticate: false,
	}

	req := httptest.NewRequest("GET", "/test", nil)
	authClasses := []Authentication{mockAuth}

	result, err := AuthenticateRequest(req, authClasses)

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestAuthenticateRequest_MultipleAuthClasses(t *testing.T) {
	// First auth class doesn't authenticate
	auth1 := &MockAuthentication{ShouldAuthenticate: false}

	// Second auth class authenticates
	mockUser := &MockUser{ID: "456", Authenticated: true}
	auth2 := &MockAuthentication{
		ShouldAuthenticate: true,
		User:               mockUser,
		Auth:               "token456",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	authClasses := []Authentication{auth1, auth2}

	result, err := AuthenticateRequest(req, authClasses)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, mockUser, result.User)
}

func TestSetUserOnRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	mockUser := &MockUser{ID: "789", Authenticated: true}

	SetUserOnRequest(req, mockUser)

	user, ok := GetUserFromRequest(req)
	require.True(t, ok)
	assert.Equal(t, mockUser, user)
}

func TestSetAuthOnRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	authData := "token789"

	SetAuthOnRequest(req, authData)

	auth, ok := GetAuthFromRequest(req)
	require.True(t, ok)
	assert.Equal(t, authData, auth)
}

func TestNewAuthResult(t *testing.T) {
	user := &MockUser{ID: "123"}
	auth := "token123"

	result := NewAuthResult(user, auth)

	assert.Equal(t, user, result.User)
	assert.Equal(t, auth, result.Auth)
}

func TestContextIntegration(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := req.Context()

	// Test WithUser
	mockUser := &MockUser{ID: "123"}
	ctx = core.WithUser(ctx, mockUser)
	req = req.WithContext(ctx)

	user, ok := core.UserFromContext(req.Context())
	require.True(t, ok)
	assert.Equal(t, mockUser, user)

	// Test WithAuth
	authData := "token123"
	ctx = core.WithAuth(ctx, authData)
	req = req.WithContext(ctx)

	auth, ok := core.AuthFromContext(req.Context())
	require.True(t, ok)
	assert.Equal(t, authData, auth)
}
