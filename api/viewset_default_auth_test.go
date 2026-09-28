package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
)

type defaultAuthQueryset struct{ accessCheckQueryset }

func (q *defaultAuthQueryset) Delete(context.Context, *accessCheckModel) error { return nil }

func TestBaseViewSetDefaultAccess(t *testing.T) {
	for _, tc := range []struct {
		name          string
		configure     func()
		explicitEmpty bool
		status        int
	}{
		{"authentication", func() { SetDefaultAuthentication(failingAuthentication{}) }, false, http.StatusUnauthorized},
		{"permission", func() { SetDefaultPermissions(denyingPermission{}) }, false, http.StatusForbidden},
		{"explicit empty permissions", func() { SetDefaultPermissions(denyingPermission{}) }, true, 0},
		{"explicit empty authentication", func() { SetDefaultAuthentication(failingAuthentication{}) }, true, 0},
		{"no defaults", func() {}, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			saved := GetSettings()
			t.Cleanup(func() { SetSettings(saved) })
			SetSettings(DefaultSettings())
			vs := NewBaseViewSet(newAccessCheckSerializer,
				&defaultAuthQueryset{accessCheckQueryset{items: []*accessCheckModel{{ID: 1}}}},
				&accessCheckModel{})
			if tc.explicitEmpty {
				vs.Authentication = []authentication.Authentication{}
				vs.Permissions = []permissions.Permission{}
			}
			// Configure after construction to require request-time resolution.
			tc.configure()
			router := NewRouter("/api")
			router.Register("items", vs)
			handler := forgehttp.NewRouter()
			router.RegisterRoutes(handler)
			for _, request := range []struct {
				method, path string
				success      int
			}{
				{http.MethodPost, "/api/items/", http.StatusCreated},
				{http.MethodDelete, "/api/items/1", http.StatusNoContent},
			} {
				t.Run(request.method, func(t *testing.T) {
					req := httptest.NewRequest(request.method, request.path, strings.NewReader("{}"))
					req.Header.Set("Content-Type", "application/json")
					response := httptest.NewRecorder()
					handler.ServeHTTP(response, req)
					want := tc.status
					if want == 0 {
						want = request.success
					}
					assert.Equal(t, want, response.Code, response.Body.String())
				})
			}
		})
	}
}

func TestBaseViewSetDefaultObjectPermissions(t *testing.T) {
	saved := GetSettings()
	t.Cleanup(func() { SetSettings(saved) })
	SetSettings(DefaultSettings())
	vs := newAccessCheckViewSet()
	SetDefaultPermissions(denyingPermission{denyObject: true})
	response := serveAccessCheckRequest(vs, http.MethodGet, "/api/items/1")
	assert.Equal(t, http.StatusForbidden, response.Code, response.Body.String())
}
