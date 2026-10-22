package api

import (
	"context"
	"encoding/json"
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
	"github.com/stretchr/testify/require"
)

type accessCheckQueryset struct {
	unusedUpdateOperation
	unusedDeleteOperation
	items        []*accessCheckModel
	offsetCalled bool
	limitCalled  bool
	offset       int
	limit        int
}

func (q *accessCheckQueryset) All(context.Context) ([]*accessCheckModel, error) {
	start := q.offset
	if start > len(q.items) {
		start = len(q.items)
	}
	end := len(q.items)
	if q.limitCalled && start+q.limit < end {
		end = start + q.limit
	}
	return q.items[start:end], nil
}

func (q *accessCheckQueryset) Count(context.Context) (int64, error) {
	return int64(len(q.items)), nil
}

func (q *accessCheckQueryset) Offset(n int) *accessCheckQueryset {
	q.offsetCalled, q.offset = true, n
	return q
}
func (q *accessCheckQueryset) Limit(n int) *accessCheckQueryset {
	q.limitCalled, q.limit = true, n
	return q
}

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

func TestListWithNonObjectPermissionsUsesQuerysetPagination(t *testing.T) {
	for _, tc := range []struct {
		name           string
		setPermissions func(*BaseViewSet, *testing.T)
	}{
		{"explicit", func(vs *BaseViewSet, _ *testing.T) {
			vs.Permissions = []permissions.Permission{permissions.NewAllowAny(), permissions.NewIsAuthenticatedOrReadOnly(), permissions.NewIsOwnerOrReadOnly("")}
		}},
		{"default", func(_ *BaseViewSet, t *testing.T) {
			saved := GetSettings()
			t.Cleanup(func() { SetSettings(saved) })
			SetDefaultPermissions(permissions.NewAllowAny())
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			qs := &accessCheckQueryset{items: []*accessCheckModel{{ID: 1}, {ID: 2}, {ID: 3}}}
			vs := NewBaseViewSet(newAccessCheckSerializer, qs, &accessCheckModel{})
			tc.setPermissions(vs, t)
			response := serveAccessCheckRequest(vs, http.MethodGet, "/api/items/?page=2&page_size=1")
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			assert.JSONEq(t, `{"count":3,"next":"http://example.com/api/items/?page=3&page_size=1","previous":"http://example.com/api/items/?page=1&page_size=1","results":[{"id":2}]}`, response.Body.String())
			assert.True(t, qs.offsetCalled)
			assert.True(t, qs.limitCalled)
			assert.Equal(t, 1, qs.offset)
			assert.Equal(t, 1, qs.limit)
		})
	}
}

type hideSecondObjectPermission struct{}

func (hideSecondObjectPermission) HasPermission(*http.Request, permissions.ViewSet) bool { return true }
func (hideSecondObjectPermission) HasObjectPermission(_ *http.Request, _ permissions.ViewSet, object interface{}) bool {
	return object.(*accessCheckModel).ID != 2
}
func (hideSecondObjectPermission) GetMessage() string { return "hidden" }
func (hideSecondObjectPermission) GetCode() string    { return "permission_denied" }

func TestListOmitsDeniedObjectsAndPaginatesAuthorizedRows(t *testing.T) {
	vs := NewBaseViewSet(newAccessCheckSerializer,
		&accessCheckQueryset{items: []*accessCheckModel{{ID: 1}, {ID: 2}, {ID: 3}}}, &accessCheckModel{})
	vs.Permissions = []permissions.Permission{permissions.NewAllowAny(), hideSecondObjectPermission{}}
	qs := vs.Queryset.(*accessCheckQueryset)
	for _, tc := range []struct {
		path           string
		id             float64
		next, previous bool
	}{
		{"/api/items/?page=1&page_size=1", 1, true, false},
		{"/api/items/?page=2&page_size=1", 3, false, true},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := serveAccessCheckRequest(vs, http.MethodGet, tc.path)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			var body struct {
				Count    int                      `json:"count"`
				Results  []map[string]interface{} `json:"results"`
				Next     *string                  `json:"next"`
				Previous *string                  `json:"previous"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
			assert.Equal(t, 2, body.Count)
			require.Len(t, body.Results, 1)
			assert.Equal(t, tc.id, body.Results[0]["id"])
			assert.Equal(t, tc.next, body.Next != nil)
			assert.Equal(t, tc.previous, body.Previous != nil)
			assert.False(t, qs.offsetCalled)
			assert.False(t, qs.limitCalled)
		})
	}
	denied := serveAccessCheckRequest(vs, http.MethodGet, "/api/items/2")
	assert.Equal(t, http.StatusForbidden, denied.Code)
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
