package principal

import (
	"context"
	"errors"
	"testing"

	"github.com/forgego/forge/identity/models"
)

type backendWithAccountState struct {
	stateErr error
}

func (backendWithAccountState) GetID() string              { return "backend-7" }
func (backendWithAccountState) GetDisplayName() string     { return "custom" }
func (u backendWithAccountState) AccountStateError() error { return u.stateErr }

func TestFromBackendRejectsUnavailableAccounts(t *testing.T) {
	for _, user := range []interface{}{
		backendWithAccountState{stateErr: models.ErrUserLocked},
		&models.User{ID: 7, IsActive: false},
	} {
		if p, ok := FromBackend(user); ok {
			t.Fatalf("FromBackend(%T) = %+v, true; want false", user, p)
		}
	}
	if p, ok := FromBackend(backendWithAccountState{}); !ok || p.ID != "backend-7" {
		t.Fatalf("active backend = %+v, %v", p, ok)
	}
}

func TestPrincipalFromUserRequiresUsableAccount(t *testing.T) {
	for _, tc := range []struct {
		name string
		user *models.User
		want error
	}{
		{name: "active", user: &models.User{ID: 7, Username: "ada", IsActive: true}},
		{name: "inactive", user: &models.User{ID: 7}, want: models.ErrUserInactive},
		{name: "locked", user: &models.User{ID: 7, IsActive: true, IsLocked: true}, want: models.ErrUserLocked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := FromUser(tc.user)
			if !errors.Is(err, tc.want) {
				t.Fatalf("FromUser() error = %v, want %v", err, tc.want)
			}
			if err == nil && (p.ID != "7" || p.Kind != KindUser || p.DisplayName != "ada") {
				t.Fatalf("unexpected principal: %+v", p)
			}
		})
	}
}

func TestPrincipalFromUserUsesUsernameWhenIDIsZero(t *testing.T) {
	p, err := FromUser(&models.User{Username: "ops", IsActive: true})
	if err != nil {
		t.Fatalf("FromUser() error = %v", err)
	}
	if p.ID != "ops" {
		t.Fatalf("principal ID = %q, want username ops", p.ID)
	}
}

func TestPrincipalContextAndBuiltins(t *testing.T) {
	p := Anonymous()
	ctx := WithPrincipal(context.Background(), p)
	got, ok := PrincipalFromContext(ctx)
	if !ok || got.Kind != KindAnonymous {
		t.Fatalf("PrincipalFromContext() = %+v, %v", got, ok)
	}
	if System("cli").ID != "cli" {
		t.Fatal("System principal should retain its name as ID")
	}
}
