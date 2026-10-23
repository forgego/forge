package core

import (
	"context"
	"testing"

	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/principal"
)

type customPrincipalUser struct{}

func (customPrincipalUser) GetID() string          { return "backend-7" }
func (customPrincipalUser) GetDisplayName() string { return "custom" }

func TestWithUserKeepsLegacyValueAndAdaptsPrincipal(t *testing.T) {
	user := &models.User{ID: 9, Username: "ada", IsActive: true}
	ctx := WithUser(context.Background(), user)
	legacy, ok := UserFromContext(ctx)
	if !ok || legacy != user {
		t.Fatal("WithUser must preserve the legacy user value")
	}
	p, ok := PrincipalFromContext(ctx)
	if !ok || p.Kind != principal.KindUser || p.ID != "9" {
		t.Fatalf("principal = %+v, %v", p, ok)
	}
}

func TestWithUserSupportsCustomBackendPrincipal(t *testing.T) {
	ctx := WithUser(context.Background(), customPrincipalUser{})
	p, ok := PrincipalFromContext(ctx)
	if !ok || p.ID != "backend-7" || p.DisplayName != "custom" {
		t.Fatalf("principal = %+v, %v", p, ok)
	}
}

func TestWithUserDoesNotAdaptInactiveUser(t *testing.T) {
	ctx := WithUser(context.Background(), &models.User{ID: 9, Username: "ada"})
	p, ok := PrincipalFromContext(ctx)
	if !ok || p.Kind == principal.KindUser {
		t.Fatal("inactive user must not become an authenticated user principal")
	}
}
