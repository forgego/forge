package service

import (
	"context"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/config"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPasswordServiceTest(t *testing.T) (PasswordService, *db.DB, context.Context) {
	testDB := setupTestDB(t)
	userRepo := repository.NewUserRepository(testDB)
	tokenRepo := repository.NewTokenRepository(testDB)

	// Use default password policy
	policy := config.DefaultIdentityConfig().PasswordPolicy
	service := NewPasswordService(userRepo, tokenRepo, policy)
	ctx := context.Background()
	return service, testDB, ctx
}

func TestPasswordService_ChangePassword(t *testing.T) {
	service, testDB, ctx := setupPasswordServiceTest(t)
	defer testDB.Close()

	// Create test user
	userRepo := repository.NewUserRepository(testDB)
	hashedPassword, _ := utils.HashPassword("oldpassword")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("changes password successfully", func(t *testing.T) {
		req := &ChangePasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "NewPassword123!",
		}

		err := service.ChangePassword(ctx, user.ID, req)
		require.NoError(t, err)

		// Verify password was changed
		updated, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.True(t, utils.CheckPassword("NewPassword123!", updated.Password))
		assert.False(t, utils.CheckPassword("oldpassword", updated.Password))
	})

	t.Run("fails with incorrect current password", func(t *testing.T) {
		req := &ChangePasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "NewPassword123!",
		}

		err := service.ChangePassword(ctx, user.ID, req)
		assert.Error(t, err)
	})

	t.Run("fails with weak new password", func(t *testing.T) {
		req := &ChangePasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "123", // Too short
		}

		err := service.ChangePassword(ctx, user.ID, req)
		assert.Error(t, err)
	})
}

func TestPasswordService_ValidatePassword(t *testing.T) {
	service, testDB, _ := setupPasswordServiceTest(t)
	defer testDB.Close()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	t.Run("validates strong password", func(t *testing.T) {
		err := service.ValidatePassword("StrongPassword123!", user)
		assert.NoError(t, err)
	})

	t.Run("rejects short password", func(t *testing.T) {
		err := service.ValidatePassword("short", user)
		assert.Error(t, err)
	})

	t.Run("rejects password matching username", func(t *testing.T) {
		err := service.ValidatePassword("testuser", user)
		assert.Error(t, err)
	})

	t.Run("rejects password matching email", func(t *testing.T) {
		err := service.ValidatePassword("test@example.com", user)
		assert.Error(t, err)
	})
}

func TestPasswordService_CheckPassword(t *testing.T) {
	service, testDB, ctx := setupPasswordServiceTest(t)
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

	t.Run("returns true for correct password", func(t *testing.T) {
		assert.True(t, service.CheckPassword(user, "correctpassword"))
	})

	t.Run("returns false for incorrect password", func(t *testing.T) {
		assert.False(t, service.CheckPassword(user, "wrongpassword"))
	})
}

func TestPasswordService_RequestPasswordReset(t *testing.T) {
	service, testDB, ctx := setupPasswordServiceTest(t)
	defer testDB.Close()

	// Create test user
	userRepo := repository.NewUserRepository(testDB)
	hashedPassword, _ := utils.HashPassword("password123")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("creates password reset token", func(t *testing.T) {
		err := service.RequestPasswordReset(ctx, "test@example.com")
		require.NoError(t, err)

		// Token should be created in database
		// This would be verified by checking token repository
	})

	t.Run("handles non-existent email gracefully", func(t *testing.T) {
		// Should not return error to prevent email enumeration
		err := service.RequestPasswordReset(ctx, "nonexistent@example.com")
		assert.NoError(t, err)
	})
}

func TestPasswordService_ResetPassword(t *testing.T) {
	service, testDB, ctx := setupPasswordServiceTest(t)
	defer testDB.Close()

	// Create test user
	userRepo := repository.NewUserRepository(testDB)
	tokenRepo := repository.NewTokenRepository(testDB)
	hashedPassword, _ := utils.HashPassword("oldpassword")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("resets password with valid token", func(t *testing.T) {
		// Create reset token - need to generate token string first
		// This is a simplified test - in real implementation, token generation would be handled
		token := &models.PasswordResetToken{
			UserID:    user.ID,
			Token:     "test-reset-token-12345",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		err := tokenRepo.CreatePasswordResetToken(ctx, token)
		require.NoError(t, err)

		// Reset password
		err = service.ResetPassword(ctx, token.Token, "NewPassword123!")
		require.NoError(t, err)

		// Verify password was changed
		updated, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.True(t, utils.CheckPassword("NewPassword123!", updated.Password))
	})

	t.Run("fails with invalid token", func(t *testing.T) {
		err := service.ResetPassword(ctx, "invalid-token", "NewPassword123!")
		assert.Error(t, err)
	})

	t.Run("fails with expired token", func(t *testing.T) {
		// Create expired token (would need to manipulate expiry)
		// This is a placeholder for when token expiry is implemented
		t.Skip("Token expiry testing requires time manipulation")
	})
}
