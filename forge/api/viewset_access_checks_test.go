package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	forgeerrors "github.com/forgego/forge/errors"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
)

type accessCheckQueryset struct {
	items []*accessCheckModel
}

func (q *accessCheckQueryset) All(context.Context) ([]*accessCheckModel, error) {
	return q.items, nil
}

func (q *accessCheckQueryset) Count(context.Context) (int64, error) {
	return int64(len(q.items)), nil
}

func (q *accessCheckQueryset) Offset(int) *accessCheckQueryset { return q }
func (q *accessCheckQueryset) Limit(int) *accessCheckQueryset  { return q }

func (q *accessCheckQueryset) Create(context.Context, interface{}) error { return nil }

func (q *accessCheckQueryset) Get(_ context.Context, id int64) (*accessCheckModel, error) {
	for _, item := range q.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, forgeerrors.NewNotFoundErrorWithMessage("not found")
}

type accessCheckModel struct {
	ID int64 `json:"id"`
}

type accessCheckSerializer struct {
	*BaseSerializer
}

func newAccessCheckSerializer() Serializer {
	return &accessCheckSerializer{BaseSerializer: NewBaseSerializer(make(map[string]interface{}))}
}

func (s *accessCheckSerializer) Fields() []string         { return []string{"id"} }
func (s *accessCheckSerializer) ReadOnlyFields() []string { return []string{"id"} }

type denyingPermission struct {
	denyObject bool
}

func (p denyingPermission) HasPermission(_ *http.Request, view permissions.ViewSet) bool {
	return p.denyObject && view.GetAction() == "retrieve"
}

func (p denyingPermission) HasObjectPermission(
	_ *http.Request,
	_ permissions.ViewSet,
	_ interface{},
) bool {
	return !p.denyObject
}

func (denyingPermission) GetMessage() string { return "access denied" }
func (denyingPermission) GetCode() string    { return "permission_denied" }

type MockUser struct {
	ID            string
	Authenticated bool
}

func (u *MockUser) GetID() string             { return u.ID }
func (u *MockUser) IsAuthenticated() bool     { return u.Authenticated }
func (u *MockUser) IsStaff() bool             { return false }
func (u *MockUser) IsSuperuser() bool         { return false }
func (u *MockUser) HasPermission(string) bool { return false }

type MockAuth struct {
	ShouldAuth bool
	User       interface{}
	Error      error
}

func (a *MockAuth) Authenticate(*http.Request) (*authentication.AuthResult, error) {
	if a.Error != nil {
		return nil, a.Error
	}
	if !a.ShouldAuth {
		return nil, nil
	}
	return authentication.NewAuthResult(a.User, "token"), nil
}

func (a *MockAuth) AuthenticateHeader(*http.Request) string { return "Mock" }

type failingAuthentication struct{}

func (failingAuthentication) Authenticate(*http.Request) (*authentication.AuthResult, error) {
	return nil, errors.New("invalid credentials")
}

func (failingAuthentication) AuthenticateHeader(*http.Request) string { return "Test" }

type denyingThrottle struct{}

func (denyingThrottle) AllowRequest(*http.Request, interface{}) (bool, time.Duration, error) {
	return false, 7 * time.Second, nil
}

func (denyingThrottle) GetScope(*http.Request, interface{}) string { return "test" }

func newAccessCheckViewSet() *BaseViewSet {
	return NewBaseViewSet(
		newAccessCheckSerializer,
		&accessCheckQueryset{items: []*accessCheckModel{{ID: 1}}},
		&accessCheckModel{},
	)
}

func TestBaseViewSetPermissionDenial(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		viewSet := newAccessCheckViewSet()
		viewSet.Permissions = []permissions.Permission{denyingPermission{}}

		response := serveAccessCheckRequest(viewSet, http.MethodGet, "/api/items/")

		assert.Equal(t, http.StatusForbidden, response.Code)
		assert.JSONEq(t, `{"type":"https://api.example.com/problems/authorization-error","title":"Permission Denied","status":403,"detail":"access denied","instance":"/api/items/","code":"PERMISSION_DENIED"}`, response.Body.String())
	})

	t.Run("retrieve object", func(t *testing.T) {
		viewSet := newAccessCheckViewSet()
		viewSet.Permissions = []permissions.Permission{denyingPermission{denyObject: true}}

		response := serveAccessCheckRequest(viewSet, http.MethodGet, "/api/items/1")

		assert.Equal(t, http.StatusForbidden, response.Code)
		assert.JSONEq(t, `{"type":"https://api.example.com/problems/authorization-error","title":"Permission Denied","status":403,"detail":"access denied","instance":"/api/items/1","code":"PERMISSION_DENIED"}`, response.Body.String())
	})
}

func TestBaseViewSetAuthenticationFailure(t *testing.T) {
	viewSet := newAccessCheckViewSet()
	viewSet.Authentication = []authentication.Authentication{failingAuthentication{}}

	response := serveAccessCheckRequest(viewSet, http.MethodGet, "/api/items/")

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.JSONEq(t, `{"type":"https://api.example.com/problems/authentication-error","title":"Authentication Failed","status":401,"detail":"invalid credentials","instance":"/api/items/","code":"AUTHENTICATION_FAILED"}`, response.Body.String())
}

func TestBaseViewSetThrottleDenial(t *testing.T) {
	viewSet := newAccessCheckViewSet()
	viewSet.Throttles = []throttling.Throttle{denyingThrottle{}}

	response := serveAccessCheckRequest(viewSet, http.MethodGet, "/api/items/")

	assert.Equal(t, http.StatusTooManyRequests, response.Code)
	assert.Equal(t, "7", response.Header().Get("Retry-After"))
	assert.JSONEq(t, `{"type":"https://api.example.com/problems/rate-limit-error","title":"Rate Limit Exceeded","status":429,"detail":"Request was throttled","instance":"/api/items/","code":"RATE_LIMIT_EXCEEDED","meta":{"retry_after_seconds":7}}`, response.Body.String())
}

func TestBaseViewSetWithoutAccessClassesKeepsListResponse(t *testing.T) {
	response := serveAccessCheckRequest(newAccessCheckViewSet(), http.MethodGet, "/api/items/")

	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"count":1,"results":[{"id":1}]}`, response.Body.String())
}

func TestViewSetConfigPassesAccessClassesToBaseViewSet(t *testing.T) {
	auth := failingAuthentication{}
	permission := denyingPermission{}
	throttle := denyingThrottle{}
	config := &ViewSetConfig{
		Model:          &accessCheckModel{},
		Queryset:       &accessCheckQueryset{},
		Serializer:     newAccessCheckSerializer(),
		Authentication: []authentication.Authentication{auth},
		Permissions:    []permissions.Permission{permission},
		Throttles:      []throttling.Throttle{throttle},
	}

	viewSet := NewConfigurableViewSet(config)

	assert.Equal(t, config.Authentication, viewSet.Authentication)
	assert.Equal(t, config.Permissions, viewSet.Permissions)
	assert.Equal(t, config.Throttles, viewSet.Throttles)
}

func serveAccessCheckRequest(viewSet ViewSet, method, path string) *httptest.ResponseRecorder {
	router := NewRouter("/api")
	router.Register("items", viewSet)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(method, path, nil))
	return response
}
