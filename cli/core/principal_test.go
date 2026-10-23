package core

import (
	"context"
	"testing"

	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/identity/repository"
	"github.com/stretchr/testify/require"
)

type principalUsers struct {
	repository.UserRepository
	user *models.User
	err  error
}

func (r principalUsers) GetByUsername(context.Context, string) (*models.User, error) {
	return r.user, r.err
}

func TestNewContextDefaultsToSystemPrincipal(t *testing.T) {
	ctx := NewContext()
	require.NotNil(t, ctx.Principal)
	require.Equal(t, principal.System("cli"), *ctx.Principal)
}

func TestAssumeUserRejectsUnusableAccounts(t *testing.T) {
	for _, tc := range []struct {
		name string
		user *models.User
		err  error
	}{
		{name: "inactive", user: &models.User{ID: 7, Username: "disabled"}},
		{name: "locked", user: &models.User{ID: 8, Username: "locked", IsActive: true, IsLocked: true}},
		{name: "unknown", err: repository.ErrUserNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewContext()
			before := *ctx.Principal
			err := ctx.AssumeUser(context.Background(), principalUsers{user: tc.user, err: tc.err}, tc.name)
			require.Error(t, err)
			require.Equal(t, before, *ctx.Principal)
		})
	}
}

func TestAssumeUserSetsResolvedPrincipal(t *testing.T) {
	ctx := NewContext()
	user := &models.User{ID: 9, Username: "staff", IsActive: true, IsStaff: true}
	require.NoError(t, ctx.AssumeUser(context.Background(), principalUsers{user: user}, user.Username))
	require.Equal(t, "9", ctx.Principal.ID)
	require.Equal(t, principal.KindUser, ctx.Principal.Kind)
	require.Equal(t, []string{"staff"}, ctx.Principal.Roles())
}
