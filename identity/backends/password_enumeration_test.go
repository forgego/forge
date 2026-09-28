package backends

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUserRepository struct {
	repository.UserRepository
	usersByUsername map[string]*models.User
	usersByEmail    map[string]*models.User
	err             error
}

func (f *fakeUserRepository) GetByUsername(_ context.Context, username string) (*models.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if u, ok := f.usersByUsername[username]; ok {
		return u, nil
	}
	return nil, repository.ErrUserNotFound
}

func (f *fakeUserRepository) GetByEmail(_ context.Context, email string) (*models.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if u, ok := f.usersByEmail[email]; ok {
		return u, nil
	}
	return nil, repository.ErrUserNotFound
}

func TestPasswordBackend_Enumeration(t *testing.T) {
	ctx := context.Background()
	correctPassword := "correct-password-123"
	wrongPassword := "wrong-password-456"

	hashedPassword, err := utils.HashPassword(correctPassword)
	require.NoError(t, err)

	activeUser := &models.User{
		ID:       1,
		Username: "activeuser",
		Email:    "active@example.com",
		Password: hashedPassword,
		IsActive: true,
		IsLocked: false,
	}

	inactiveUser := &models.User{
		ID:       2,
		Username: "inactiveuser",
		Email:    "inactive@example.com",
		Password: hashedPassword,
		IsActive: false,
		IsLocked: false,
	}

	lockedUser := &models.User{
		ID:       3,
		Username: "lockeduser",
		Email:    "locked@example.com",
		Password: hashedPassword,
		IsActive: true,
		IsLocked: true,
	}

	dbErr := errors.New("database connection failed")

	newRepo := func(users []*models.User, repoErr error) repository.UserRepository {
		byUsername := make(map[string]*models.User)
		byEmail := make(map[string]*models.User)
		for _, u := range users {
			byUsername[u.Username] = u
			byEmail[u.Email] = u
		}
		return &fakeUserRepository{
			usersByUsername: byUsername,
			usersByEmail:    byEmail,
			err:             repoErr,
		}
	}

	tests := []struct {
		name               string
		repo               repository.UserRepository
		credentials        map[string]string
		expectedErr        error
		expectedWrappedErr error
		notExpectedErr     error
		expectUser         bool
		expectBcryptCall   bool
	}{
		{
			name: "unknown username",
			repo: newRepo(nil, nil),
			credentials: map[string]string{
				"username": "unknown",
				"password": wrongPassword,
			},
			expectedErr:      ErrInvalidCredentials,
			expectUser:       false,
			expectBcryptCall: true,
		},
		{
			name: "unknown email",
			repo: newRepo(nil, nil),
			credentials: map[string]string{
				"email":    "unknown@example.com",
				"password": wrongPassword,
			},
			expectedErr:      ErrInvalidCredentials,
			expectUser:       false,
			expectBcryptCall: true,
		},
		{
			name: "unknown username (sql.ErrNoRows)",
			repo: newRepo(nil, sql.ErrNoRows),
			credentials: map[string]string{
				"username": "unknown_sql",
				"password": wrongPassword,
			},
			expectedErr:      ErrInvalidCredentials,
			expectUser:       false,
			expectBcryptCall: true,
		},
		{
			name: "inactive user + wrong password",
			repo: newRepo([]*models.User{inactiveUser}, nil),
			credentials: map[string]string{
				"username": "inactiveuser",
				"password": wrongPassword,
			},
			expectedErr:      ErrInvalidCredentials,
			notExpectedErr:   ErrUserInactive,
			expectUser:       false,
			expectBcryptCall: true,
		},
		{
			name: "locked user + wrong password",
			repo: newRepo([]*models.User{lockedUser}, nil),
			credentials: map[string]string{
				"username": "lockeduser",
				"password": wrongPassword,
			},
			expectedErr:      ErrInvalidCredentials,
			notExpectedErr:   ErrUserLocked,
			expectUser:       false,
			expectBcryptCall: true,
		},
		{
			name: "inactive user + correct password",
			repo: newRepo([]*models.User{inactiveUser}, nil),
			credentials: map[string]string{
				"username": "inactiveuser",
				"password": correctPassword,
			},
			expectedErr:      ErrUserInactive,
			expectUser:       false,
			expectBcryptCall: true,
		},
		{
			name: "locked user + correct password",
			repo: newRepo([]*models.User{lockedUser}, nil),
			credentials: map[string]string{
				"username": "lockeduser",
				"password": correctPassword,
			},
			expectedErr:      ErrUserLocked,
			expectUser:       false,
			expectBcryptCall: true,
		},
		{
			name: "active user + correct password",
			repo: newRepo([]*models.User{activeUser}, nil),
			credentials: map[string]string{
				"username": "activeuser",
				"password": correctPassword,
			},
			expectedErr:      nil,
			expectUser:       true,
			expectBcryptCall: true,
		},
		{
			name: "repository returns a non-not-found error",
			repo: newRepo(nil, dbErr),
			credentials: map[string]string{
				"username": "activeuser",
				"password": correctPassword,
			},
			expectedWrappedErr: dbErr,
			notExpectedErr:     ErrInvalidCredentials,
			expectUser:         false,
			expectBcryptCall:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origCompare := comparePassword
			var compareCalls int32
			comparePassword = func(password, hash string) bool {
				atomic.AddInt32(&compareCalls, 1)
				return origCompare(password, hash)
			}
			t.Cleanup(func() {
				comparePassword = origCompare
			})

			backend := NewPasswordBackend(tt.repo)
			user, err := backend.Authenticate(ctx, tt.credentials)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			}
			if tt.expectedWrappedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedWrappedErr)
			}
			if tt.notExpectedErr != nil {
				assert.False(t, errors.Is(err, tt.notExpectedErr), "error should NOT match %v, got %v", tt.notExpectedErr, err)
			}
			if tt.expectUser {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, activeUser.ID, user.ID)
			} else {
				assert.Nil(t, user)
			}
			if tt.expectBcryptCall {
				assert.Equal(t, int32(1), atomic.LoadInt32(&compareCalls), "bcrypt comparison was expected to run once")
			} else {
				assert.Equal(t, int32(0), atomic.LoadInt32(&compareCalls), "bcrypt comparison should not have run")
			}
		})
	}
}
