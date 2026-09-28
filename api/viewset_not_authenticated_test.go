package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
)

type ownerOnlyPermission struct{}

func (ownerOnlyPermission) HasPermission(*http.Request, permissions.ViewSet) bool { return true }
func (ownerOnlyPermission) HasObjectPermission(*http.Request, permissions.ViewSet, interface{}) bool {
	return false
}
func (ownerOnlyPermission) GetMessage() string { return "Not the owner" }
func (ownerOnlyPermission) GetCode() string    { return "permission_denied" }

// TestBaseViewSetUnauthenticatedPermissionFailure follows Django REST
// framework: an unauthenticated request that fails a permission answers 401
// with a WWW-Authenticate challenge when the first authentication class can
// issue one, and 403 otherwise; an authenticated request stays 403.
func TestBaseViewSetUnauthenticatedPermissionFailure(t *testing.T) {
	tokens := authentication.NewTokenAuthentication(func(token string) (interface{}, error) {
		if token == "good" {
			return "alice", nil
		}
		return nil, nil
	})
	session := authentication.NewSessionAuthentication(nil)
	for _, tc := range []struct {
		name      string
		auth      []authentication.Authentication
		perms     []permissions.Permission
		path      string
		token     string
		status    int
		challenge string
	}{
		{name: "token auth, anonymous", auth: []authentication.Authentication{tokens},
			perms: []permissions.Permission{permissions.NewIsAuthenticated()}, path: "/api/items/",
			status: http.StatusUnauthorized, challenge: "Token"},
		{name: "token auth, anonymous object permission", auth: []authentication.Authentication{tokens},
			perms: []permissions.Permission{ownerOnlyPermission{}}, path: "/api/items/1",
			status: http.StatusUnauthorized, challenge: "Token"},
		{name: "token auth, authenticated", auth: []authentication.Authentication{tokens},
			perms: []permissions.Permission{ownerOnlyPermission{}}, path: "/api/items/1", token: "good",
			status: http.StatusForbidden},
		{name: "token auth, bad token", auth: []authentication.Authentication{tokens},
			perms: []permissions.Permission{permissions.NewIsAuthenticated()}, path: "/api/items/", token: "bad",
			status: http.StatusUnauthorized, challenge: "Token"},
		{name: "session first cannot challenge", auth: []authentication.Authentication{session, tokens},
			perms: []permissions.Permission{permissions.NewIsAuthenticated()}, path: "/api/items/",
			status: http.StatusForbidden},
		{name: "no authentication classes", auth: []authentication.Authentication{},
			perms: []permissions.Permission{permissions.NewIsAuthenticated()}, path: "/api/items/",
			status: http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vs := newAccessCheckViewSet()
			vs.Authentication = tc.auth
			vs.Permissions = tc.perms
			router := NewRouter("/api")
			router.Register("items", vs)
			handler := forgehttp.NewRouter()
			router.RegisterRoutes(handler)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Token "+tc.token)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			assert.Equal(t, tc.status, response.Code, response.Body.String())
			assert.Equal(t, tc.challenge, response.Header().Get("WWW-Authenticate"))
		})
	}
}
