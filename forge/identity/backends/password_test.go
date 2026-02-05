package backends

import (
	"context"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/testutils"
	"github.com/forgego/forge/identity/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPasswordBackendTest(t *testing.T) (AuthenticationBackend, *db.DB, context.Context) {
	testDB := setupTestDB(t)
	userRepo := repository.NewUserRepository(testDB)
	backend := NewPasswordBackend(userRepo)
	ctx := context.Background()
	return backend, testDB, ctx
}

func setupTestDB(t *testing.T) *db.DB {
	return testutils.SetupTestDB(t)
}

func TestPasswordBackend_Authenticate(t *testing.T) {
	backend, testDB, ctx := setupPasswordBackendTest(t)
	defer testDB.Close()

	// Create test user
	userRepo := repository.NewUserRepository(testDB)
	hashedPassword, _ := utils.HashPassword("correctpassword")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("authenticates with correct username and password", func(t *testing.T) {
		credentials := map[string]string{
			"username": "testuser",
			"password": "correctpassword",
		}

		authenticated, err := backend.Authenticate(ctx, credentials)
		require.NoError(t, err)
		assert.Equal(t, user.ID, authenticated.ID)
		assert.Equal(t, user.Username, authenticated.Username)
	})

	t.Run("authenticates with email and password", func(t *testing.T) {
		credentials := map[string]string{
			"email":    "test@example.com",
			"password": "correctpassword",
		}

		authenticated, err := backend.Authenticate(ctx, credentials)
		require.NoError(t, err)
		assert.Equal(t, user.ID, authenticated.ID)
	})

	t.Run("fails with incorrect password", func(t *testing.T) {
		credentials := map[string]string{
			"username": "testuser",
			"password": "wrongpassword",
		}

		_, err := backend.Authenticate(ctx, credentials)
		assert.Error(t, err)
	})

	t.Run("fails with non-existent user", func(t *testing.T) {
		credentials := map[string]string{
			"username": "nonexistent",
			"password": "password123",
		}

		_, err := backend.Authenticate(ctx, credentials)
		assert.Error(t, err)
	})

	t.Run("fails with inactive user", func(t *testing.T) {
		// Deactivate user
		user.IsActive = false
		userRepo.Update(ctx, user)

		credentials := map[string]string{
			"username": "testuser",
			"password": "correctpassword",
		}

		_, err := backend.Authenticate(ctx, credentials)
		assert.Error(t, err)
	})

	t.Run("fails with locked user", func(t *testing.T) {
		// Reactivate and lock user
		user.IsActive = true
		user.IsLocked = true
		userRepo.Update(ctx, user)

		credentials := map[string]string{
			"username": "testuser",
			"password": "correctpassword",
		}

		_, err := backend.Authenticate(ctx, credentials)
		assert.Error(t, err)
	})
}

func TestPasswordBackend_Supports(t *testing.T) {
	backend, testDB, _ := setupPasswordBackendTest(t)
	defer testDB.Close()

	t.Run("supports password credential type", func(t *testing.T) {
		assert.True(t, backend.Supports("password"))
	})

	t.Run("does not support other credential types", func(t *testing.T) {
		assert.False(t, backend.Supports("token"))
		assert.False(t, backend.Supports("oauth"))
	})
}

func TestPasswordBackend_Name(t *testing.T) {
	backend, testDB, _ := setupPasswordBackendTest(t)
	defer testDB.Close()

	assert.Equal(t, "password", backend.Name())
}

func TestPasswordBackend_GetUser(t *testing.T) {
	backend, testDB, ctx := setupPasswordBackendTest(t)
	defer testDB.Close()

	t.Run("returns nil for password backend", func(t *testing.T) {
		// Password backend doesn't support GetUser by identifier
		// It returns nil, nil to indicate it's not applicable
		user, err := backend.GetUser(ctx, "identifier")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}
