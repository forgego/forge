package permissions

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/identity/principal"
)

// generatedMember mirrors the primary key spelling produced by Forge's model
// generator and the active-account field used by generated auth models.
type generatedMember struct {
	Id       int64
	IsActive bool
}

func TestGeneratedModelTokenAuthenticationSetsPrincipalForWrites(t *testing.T) {
	member := &generatedMember{Id: 42, IsActive: true}
	auth := authentication.NewTokenAuthentication(func(token string) (interface{}, error) {
		if token != "member-token" {
			return nil, nil
		}
		return member, nil
	})
	request := httptest.NewRequest("POST", "/api/v1/projects/", nil)
	request.Header.Set("Authorization", "Token member-token")
	result, err := auth.Authenticate(request)
	if err != nil || result == nil {
		t.Fatalf("Authenticate() = (%v, %v), want authenticated member", result, err)
	}
	authentication.SetUserOnRequest(request, result.User)

	actor, ok := principal.PrincipalFromContext(request.Context())
	if !ok || actor.Kind != principal.KindUser || actor.ID != "42" {
		t.Fatalf("principal = %+v, %v; want active generated member 42", actor, ok)
	}
	if !NewIsAuthenticated().HasPermission(request, nil) {
		t.Fatal("active generated member must be allowed to POST")
	}
}

func TestLegacyPermissionFallbackRequiresActiveUser(t *testing.T) {
	for _, tc := range []struct {
		name   string
		active bool
		want   bool
	}{
		{name: "active", active: true, want: true},
		{name: "inactive", active: false, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/api/v1/projects/", nil)
			request = request.WithContext(context.WithValue(request.Context(), core.UserKey, &generatedMember{Id: 42, IsActive: tc.active}))
			if got := NewIsAuthenticated().HasPermission(request, nil); got != tc.want {
				t.Fatalf("HasPermission() = %v, want %v", got, tc.want)
			}
		})
	}
}
