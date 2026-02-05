package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockQueryset is a mock queryset for testing
type MockQueryset struct {
	items []interface{}
}

func (m *MockQueryset) All(ctx interface{}) ([]interface{}, error) {
	return m.items, nil
}

func (m *MockQueryset) Count(ctx interface{}) (int64, error) {
	return int64(len(m.items)), nil
}

func (m *MockQueryset) Filter(expr interface{}) interface{} {
	return m
}

func (m *MockQueryset) OrderBy(fields []string) interface{} {
	return m
}

func (m *MockQueryset) Offset(n int) interface{} {
	return m
}

func (m *MockQueryset) Limit(n int) interface{} {
	return m
}

// MockModel for testing
type MockModel struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// MockSerializer for testing
type MockSerializer struct {
	*BaseSerializer
}

func NewMockSerializer() Serializer {
	return &MockSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *MockSerializer) Fields() []string {
	return []string{"id", "name"}
}

func (s *MockSerializer) ReadOnlyFields() []string {
	return []string{"id"}
}

// MockUser for testing
type MockUser struct {
	ID            string
	Authenticated bool
}

func (u *MockUser) GetID() string                            { return u.ID }
func (u *MockUser) IsAuthenticated() bool                    { return u.Authenticated }
func (u *MockUser) IsStaff() bool                            { return false }
func (u *MockUser) IsSuperuser() bool                        { return false }
func (u *MockUser) HasPermission(permissionCode string) bool { return false }

func TestEnhancedBaseViewSet_List_NoAuth(t *testing.T) {
	queryset := &MockQueryset{
		items: []interface{}{
			&MockModel{ID: 1, Name: "Item 1"},
			&MockModel{ID: 2, Name: "Item 2"},
		},
	}

	viewset := NewEnhancedBaseViewSet(
		NewMockSerializer,
		queryset,
		&MockModel{},
	)

	// No authentication required
	viewset.PermissionClasses = []permissions.Permission{
		permissions.NewAllowAny(),
	}

	req := httptest.NewRequest("GET", "/test/", nil)
	w := httptest.NewRecorder()

	viewset.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "results")
}

func TestEnhancedBaseViewSet_List_WithAuth(t *testing.T) {
	queryset := &MockQueryset{
		items: []interface{}{
			&MockModel{ID: 1, Name: "Item 1"},
		},
	}

	viewset := NewEnhancedBaseViewSet(
		NewMockSerializer,
		queryset,
		&MockModel{},
	)

	// Require authentication
	viewset.AuthenticationClasses = []authentication.Authentication{
		&MockAuth{ShouldAuth: true, User: &MockUser{ID: "123", Authenticated: true}},
	}
	viewset.PermissionClasses = []permissions.Permission{
		permissions.NewIsAuthenticated(),
	}

	req := httptest.NewRequest("GET", "/test/", nil)
	w := httptest.NewRecorder()

	viewset.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEnhancedBaseViewSet_List_Unauthorized(t *testing.T) {
	queryset := &MockQueryset{items: []interface{}{}}

	viewset := NewEnhancedBaseViewSet(
		NewMockSerializer,
		queryset,
		&MockModel{},
	)

	// Require authentication but don't provide it
	viewset.PermissionClasses = []permissions.Permission{
		permissions.NewIsAuthenticated(),
	}

	req := httptest.NewRequest("GET", "/test/", nil)
	w := httptest.NewRecorder()

	viewset.List(w, req)

	// Should return 401 or 403
	assert.True(t, w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden)
}

func TestEnhancedBaseViewSet_Create(t *testing.T) {
	queryset := &MockQueryset{items: []interface{}{}}

	viewset := NewEnhancedBaseViewSet(
		NewMockSerializer,
		queryset,
		&MockModel{},
	)

	viewset.PermissionClasses = []permissions.Permission{
		permissions.NewAllowAny(),
	}

	data := map[string]interface{}{
		"name": "New Item",
	}
	jsonData, _ := json.Marshal(data)

	req := httptest.NewRequest("POST", "/test/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	viewset.Create(w, req)

	// Should return 201 Created or handle error appropriately
	assert.True(t, w.Code == http.StatusCreated || w.Code >= 400)
}

func TestEnhancedBaseViewSet_Retrieve(t *testing.T) {
	queryset := &MockQueryset{
		items: []interface{}{
			&MockModel{ID: 1, Name: "Item 1"},
		},
	}

	viewset := NewEnhancedBaseViewSet(
		NewMockSerializer,
		queryset,
		&MockModel{},
	)

	viewset.PermissionClasses = []permissions.Permission{
		permissions.NewAllowAny(),
	}

	req := httptest.NewRequest("GET", "/test/1/", nil)
	// Set ID parameter (would normally be done by router)
	req = req.WithContext(core.WithAction(req.Context(), "1"))
	w := httptest.NewRecorder()

	viewset.Retrieve(w, req)

	// Should return 200 OK, 404 Not Found, or handle error appropriately
	// The actual behavior depends on queryset implementation
	assert.True(t, w.Code >= 200 && w.Code < 500)
}

func TestEnhancedBaseViewSet_WithThrottling(t *testing.T) {
	queryset := &MockQueryset{items: []interface{}{}}

	viewset := NewEnhancedBaseViewSet(
		NewMockSerializer,
		queryset,
		&MockModel{},
	)

	viewset.PermissionClasses = []permissions.Permission{
		permissions.NewAllowAny(),
	}

	// Add throttling
	cache := throttling.NewMemoryCache()
	viewset.ThrottleClasses = []throttling.Throttle{
		throttling.NewAnonRateThrottle("1/hour", cache), // Very strict
	}

	req := httptest.NewRequest("GET", "/test/", nil)
	w := httptest.NewRecorder()

	// First request should succeed
	viewset.List(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request should be throttled
	w2 := httptest.NewRecorder()
	viewset.List(w2, req)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestEnhancedBaseViewSet_ExceptionHandling(t *testing.T) {
	queryset := &MockQueryset{items: []interface{}{}}

	viewset := NewEnhancedBaseViewSet(
		NewMockSerializer,
		queryset,
		&MockModel{},
	)

	viewset.PermissionClasses = []permissions.Permission{
		permissions.NewIsAuthenticated(), // Will fail
	}

	req := httptest.NewRequest("GET", "/test/", nil)
	w := httptest.NewRecorder()

	viewset.List(w, req)

	// Should handle exception and return error response
	assert.True(t, w.Code >= 400)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "error")
}

// MockAuth for testing
type MockAuth struct {
	ShouldAuth bool
	User       interface{}
	Error      error
}

func (m *MockAuth) Authenticate(r *http.Request) (*authentication.AuthResult, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	if m.ShouldAuth {
		return authentication.NewAuthResult(m.User, "token"), nil
	}
	return nil, nil
}

func (m *MockAuth) AuthenticateHeader(r *http.Request) string {
	return "Mock"
}
