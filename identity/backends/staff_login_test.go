package backends

import (
	"context"
	"errors"
	"testing"

	"github.com/forgego/forge/identity/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAuthBackend struct {
	authenticateFunc func(ctx context.Context, credentials map[string]string) (*models.User, error)
}

func (f *fakeAuthBackend) Authenticate(ctx context.Context, credentials map[string]string) (*models.User, error) {
	if f.authenticateFunc != nil {
		return f.authenticateFunc(ctx, credentials)
	}
	return nil, nil
}

func (f *fakeAuthBackend) GetUser(ctx context.Context, identifier string) (*models.User, error) {
	return nil, nil
}

func (f *fakeAuthBackend) Supports(credentialType string) bool {
	return credentialType == "password"
}

func (f *fakeAuthBackend) Name() string {
	return "fake"
}

func TestStaffLoginAuthenticator_StaffUser(t *testing.T) {
	fake := &fakeAuthBackend{
		authenticateFunc: func(ctx context.Context, credentials map[string]string) (*models.User, error) {
			assert.Equal(t, "alice", credentials["username"])
			assert.Equal(t, "password123", credentials["password"])
			return &models.User{
				Username: "alice",
				IsActive: true,
				IsStaff:  true,
			}, nil
		},
	}

	auth := NewStaffLoginAuthenticator(fake)
	username, err := auth.AuthenticateAdmin(context.Background(), "alice", "password123")
	require.NoError(t, err)
	assert.Equal(t, "alice", username)
}

func TestStaffLoginAuthenticator_Superuser(t *testing.T) {
	fake := &fakeAuthBackend{
		authenticateFunc: func(ctx context.Context, credentials map[string]string) (*models.User, error) {
			return &models.User{
				Username:    "boss",
				IsActive:    true,
				IsSuperuser: true,
			}, nil
		},
	}

	auth := NewStaffLoginAuthenticator(fake)
	username, err := auth.AuthenticateAdmin(context.Background(), "boss", "password123")
	require.NoError(t, err)
	assert.Equal(t, "boss", username)
}

func TestStaffLoginAuthenticator_NonStaffActiveUser(t *testing.T) {
	fake := &fakeAuthBackend{
		authenticateFunc: func(ctx context.Context, credentials map[string]string) (*models.User, error) {
			return &models.User{
				Username:    "regular",
				IsActive:    true,
				IsStaff:     false,
				IsSuperuser: false,
			}, nil
		},
	}

	auth := NewStaffLoginAuthenticator(fake)
	username, err := auth.AuthenticateAdmin(context.Background(), "regular", "password123")
	require.Error(t, err)
	assert.Empty(t, username)
	assert.True(t, errors.Is(err, ErrAdminLoginDenied))

	var inv interface{ InvalidLogin() bool }
	require.True(t, errors.As(err, &inv))
	assert.True(t, inv.InvalidLogin())
}

func TestStaffLoginAuthenticator_ErrInvalidCredentials(t *testing.T) {
	fake := &fakeAuthBackend{
		authenticateFunc: func(ctx context.Context, credentials map[string]string) (*models.User, error) {
			return nil, ErrInvalidCredentials
		},
	}

	auth := NewStaffLoginAuthenticator(fake)
	username, err := auth.AuthenticateAdmin(context.Background(), "someone", "wrongpass")
	require.Error(t, err)
	assert.Empty(t, username)
	assert.True(t, errors.Is(err, ErrAdminLoginDenied))
	assert.True(t, errors.Is(err, ErrInvalidCredentials))

	var inv interface{ InvalidLogin() bool }
	require.True(t, errors.As(err, &inv))
	assert.True(t, inv.InvalidLogin())
}

func TestStaffLoginAuthenticator_ErrUserInactive(t *testing.T) {
	fake := &fakeAuthBackend{
		authenticateFunc: func(ctx context.Context, credentials map[string]string) (*models.User, error) {
			return nil, ErrUserInactive
		},
	}

	auth := NewStaffLoginAuthenticator(fake)
	username, err := auth.AuthenticateAdmin(context.Background(), "inactive", "pass")
	require.Error(t, err)
	assert.Empty(t, username)
	assert.True(t, errors.Is(err, ErrAdminLoginDenied))
	assert.True(t, errors.Is(err, ErrUserInactive))

	var inv interface{ InvalidLogin() bool }
	require.True(t, errors.As(err, &inv))
	assert.True(t, inv.InvalidLogin())
}

func TestStaffLoginAuthenticator_ErrUserLocked(t *testing.T) {
	fake := &fakeAuthBackend{
		authenticateFunc: func(ctx context.Context, credentials map[string]string) (*models.User, error) {
			return nil, ErrUserLocked
		},
	}

	auth := NewStaffLoginAuthenticator(fake)
	username, err := auth.AuthenticateAdmin(context.Background(), "locked", "pass")
	require.Error(t, err)
	assert.Empty(t, username)
	assert.True(t, errors.Is(err, ErrAdminLoginDenied))
	assert.True(t, errors.Is(err, ErrUserLocked))

	var inv interface{ InvalidLogin() bool }
	require.True(t, errors.As(err, &inv))
	assert.True(t, inv.InvalidLogin())
}

func TestStaffLoginAuthenticator_NilUser(t *testing.T) {
	fake := &fakeAuthBackend{
		authenticateFunc: func(ctx context.Context, credentials map[string]string) (*models.User, error) {
			return nil, nil
		},
	}

	auth := NewStaffLoginAuthenticator(fake)
	username, err := auth.AuthenticateAdmin(context.Background(), "nobody", "pass")
	require.Error(t, err)
	assert.Empty(t, username)
	assert.True(t, errors.Is(err, ErrAdminLoginDenied))

	var inv interface{ InvalidLogin() bool }
	require.True(t, errors.As(err, &inv))
	assert.True(t, inv.InvalidLogin())
}

func TestStaffLoginAuthenticator_InfraError(t *testing.T) {
	dbErr := errors.New("database timeout")
	fake := &fakeAuthBackend{
		authenticateFunc: func(ctx context.Context, credentials map[string]string) (*models.User, error) {
			return nil, dbErr
		},
	}

	auth := NewStaffLoginAuthenticator(fake)
	username, err := auth.AuthenticateAdmin(context.Background(), "user", "pass")
	require.Error(t, err)
	assert.Empty(t, username)
	assert.True(t, errors.Is(err, dbErr))
	assert.False(t, errors.Is(err, ErrAdminLoginDenied))

	var inv interface{ InvalidLogin() bool }
	assert.False(t, errors.As(err, &inv))
}
