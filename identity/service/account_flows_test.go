package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/forgego/forge/identity/config"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFailingSessionRepository simulates failure in SessionRepository operations.
type mockFailingSessionRepository struct {
	repository.SessionRepository
	deleteByUserIDError error
}

func (m *mockFailingSessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	if m.deleteByUserIDError != nil {
		return m.deleteByUserIDError
	}
	return m.SessionRepository.DeleteByUserID(ctx, userID)
}

// TestSignIn_NegativeCases tests failure modes and enumeration resistance for sign-in.
func TestSignIn_NegativeCases(t *testing.T) {
	service, testDB, ctx := setupAuthServiceTest(t)
	defer testDB.Close()

	userRepo := repository.NewUserRepository(testDB)
	hashedPassword, err := utils.HashPassword("CorrectPassword123!")
	require.NoError(t, err)

	activeUser := &models.User{
		Username: "signin_active_user",
		Email:    "signin_active@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	require.NoError(t, userRepo.Create(ctx, activeUser))

	inactiveUser := &models.User{
		Username: "signin_inactive_user",
		Email:    "signin_inactive@example.com",
		Password: hashedPassword,
		IsActive: false,
	}
	require.NoError(t, userRepo.Create(ctx, inactiveUser))

	t.Run("wrong password returns invalid credentials", func(t *testing.T) {
		_, err := service.Authenticate(ctx, &AuthenticateRequest{
			UsernameOrEmail: activeUser.Username,
			Password:        "WrongPassword123!",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("unknown user returns same invalid credentials error without enumeration", func(t *testing.T) {
		// Non-existent username
		_, errUnknownUsername := service.Authenticate(ctx, &AuthenticateRequest{
			UsernameOrEmail: "completely_unknown_user",
			Password:        "WrongPassword123!",
		})
		require.Error(t, errUnknownUsername)
		assert.ErrorIs(t, errUnknownUsername, ErrInvalidCredentials)

		// Non-existent email
		_, errUnknownEmail := service.Authenticate(ctx, &AuthenticateRequest{
			UsernameOrEmail: "nonexistent@example.com",
			Password:        "WrongPassword123!",
		})
		require.Error(t, errUnknownEmail)
		assert.ErrorIs(t, errUnknownEmail, ErrInvalidCredentials)

		// Wrong password on existing user
		_, errWrongPassword := service.Authenticate(ctx, &AuthenticateRequest{
			UsernameOrEmail: activeUser.Username,
			Password:        "WrongPassword123!",
		})

		// Both unknown user and wrong password must return identical error (no enumeration)
		assert.Equal(t, errWrongPassword, errUnknownUsername)
		assert.Equal(t, errWrongPassword, errUnknownEmail)
	})

	t.Run("inactive user returns invalid credentials without enumeration", func(t *testing.T) {
		_, err := service.Authenticate(ctx, &AuthenticateRequest{
			UsernameOrEmail: inactiveUser.Username,
			Password:        "WrongPassword123!",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials, "inactive user must not leak inactive status")
	})
}

// TestSignOut_NegativeCases tests failure modes of signing out.
func TestSignOut_NegativeCases(t *testing.T) {
	service, testDB, ctx := setupAuthServiceTest(t)
	defer testDB.Close()

	userRepo := repository.NewUserRepository(testDB)
	hashedPassword, err := utils.HashPassword("password123")
	require.NoError(t, err)

	user := &models.User{
		Username: "signout_user",
		Email:    "signout@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	require.NoError(t, userRepo.Create(ctx, user))

	t.Run("fails when signing out non-existent session", func(t *testing.T) {
		err := service.Logout(ctx, "nonexistent-session-key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("fails when signing out an already logged-out session", func(t *testing.T) {
		session, err := service.CreateSession(ctx, user.ID, &CreateSessionRequest{
			UserID:    user.ID,
			IPAddress: "127.0.0.1",
			UserAgent: "test-agent",
		})
		require.NoError(t, err)

		// First logout succeeds
		err = service.Logout(ctx, session.SessionKey)
		require.NoError(t, err)

		// Repeated logout with the same session key must fail
		err = service.Logout(ctx, session.SessionKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})
}

// TestLogoutAll_RevocationAndNegativeCases tests session revocation isolation and failure handling.
func TestLogoutAll_RevocationAndNegativeCases(t *testing.T) {
	service, testDB, ctx := setupAuthServiceTest(t)
	defer testDB.Close()

	userRepo := repository.NewUserRepository(testDB)
	sessionRepo := repository.NewSessionRepository(testDB)
	hashedPassword, err := utils.HashPassword("password123")
	require.NoError(t, err)

	user1 := &models.User{Username: "user1", Email: "u1@example.com", Password: hashedPassword, IsActive: true}
	user2 := &models.User{Username: "user2", Email: "u2@example.com", Password: hashedPassword, IsActive: true}
	require.NoError(t, userRepo.Create(ctx, user1))
	require.NoError(t, userRepo.Create(ctx, user2))

	t.Run("revokes target user sessions while preserving other users sessions", func(t *testing.T) {
		s1, err := service.CreateSession(ctx, user1.ID, &CreateSessionRequest{UserID: user1.ID})
		require.NoError(t, err)
		s2, err := service.CreateSession(ctx, user1.ID, &CreateSessionRequest{UserID: user1.ID})
		require.NoError(t, err)
		s3, err := service.CreateSession(ctx, user2.ID, &CreateSessionRequest{UserID: user2.ID})
		require.NoError(t, err)

		// Logout all for user1
		require.NoError(t, service.LogoutAll(ctx, user1.ID))

		// User1 sessions are now revoked and cannot be retrieved or refreshed
		_, err = service.GetSession(ctx, s1.SessionKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")

		_, err = service.GetSession(ctx, s2.SessionKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")

		_, err = service.RefreshSession(ctx, s1.SessionKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")

		// User2 session remains untouched and valid
		u2Session, err := service.GetSession(ctx, s3.SessionKey)
		require.NoError(t, err)
		assert.Equal(t, user2.ID, u2Session.UserID)
	})

	t.Run("fails when session repository encounters an error", func(t *testing.T) {
		expectedErr := errors.New("db delete failed")
		failingRepo := &mockFailingSessionRepository{
			SessionRepository:   sessionRepo,
			deleteByUserIDError: expectedErr,
		}
		failingService := NewAuthService(userRepo, failingRepo, nil)

		err := failingService.LogoutAll(ctx, user1.ID)
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("password change failure does not revoke other sessions", func(t *testing.T) {
		tokenRepo := repository.NewTokenRepository(testDB)
		policy := config.DefaultIdentityConfig().PasswordPolicy
		pwdService := NewPasswordService(userRepo, tokenRepo, policy)

		// Create user with current and other session
		sessCurrent, err := service.CreateSession(ctx, user1.ID, &CreateSessionRequest{UserID: user1.ID})
		require.NoError(t, err)
		sessOther, err := service.CreateSession(ctx, user1.ID, &CreateSessionRequest{UserID: user1.ID})
		require.NoError(t, err)

		// Attempt password change with wrong password
		err = pwdService.ChangePassword(ctx, user1.ID, &ChangePasswordRequest{
			CurrentPassword:   "wrong-current-password",
			NewPassword:       "ValidNewPassword123!",
			CurrentSessionKey: sessCurrent.SessionKey,
		})
		require.Error(t, err)

		// Both sessions must still be valid
		_, err = service.GetSession(ctx, sessCurrent.SessionKey)
		require.NoError(t, err)
		_, err = service.GetSession(ctx, sessOther.SessionKey)
		require.NoError(t, err)
	})
}

// TestPasswordReset_NegativeCases tests single-use, expiry, and validation failure modes.
func TestPasswordReset_NegativeCases(t *testing.T) {
	service, testDB, ctx := setupPasswordServiceTest(t)
	defer testDB.Close()

	userRepo := repository.NewUserRepository(testDB)
	tokenRepo := repository.NewTokenRepository(testDB)
	hashedPassword, err := utils.HashPassword("OldPassword123!")
	require.NoError(t, err)

	user := &models.User{
		Username: "reset_user",
		Email:    "reset@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	require.NoError(t, userRepo.Create(ctx, user))

	t.Run("single-use: reset token cannot be reused after successful reset", func(t *testing.T) {
		rawToken := "single-use-reset-token-12345"
		hashedToken, hashErr := utils.HashPassword(rawToken)
		require.NoError(t, hashErr)

		token := &models.PasswordResetToken{
			UserID:    user.ID,
			Token:     hashedToken,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		require.NoError(t, tokenRepo.CreatePasswordResetToken(ctx, token))

		// First reset must succeed
		err := service.ResetPassword(ctx, rawToken, "NewStrongPassword123!")
		require.NoError(t, err)

		// Verify password was updated
		updatedUser, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.True(t, utils.CheckPassword("NewStrongPassword123!", updatedUser.Password))

		// Second reset attempt with the same token must fail
		err = service.ResetPassword(ctx, rawToken, "AnotherPassword123!")
		require.Error(t, err)
	})

	t.Run("single-use: unhashed token cannot be reused after successful reset", func(t *testing.T) {
		rawToken := "raw-single-use-reset-token"
		token := &models.PasswordResetToken{
			UserID:    user.ID,
			Token:     rawToken,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		require.NoError(t, tokenRepo.CreatePasswordResetToken(ctx, token))

		err := service.ResetPassword(ctx, rawToken, "NewStrongPassword456!")
		require.NoError(t, err)

		// Second reset must fail
		err = service.ResetPassword(ctx, rawToken, "AnotherPassword456!")
		require.Error(t, err)
	})

	t.Run("expiry: expired reset token is rejected", func(t *testing.T) {
		rawToken := "expired-reset-token-negative"
		hashedToken, hashErr := utils.HashPassword(rawToken)
		require.NoError(t, hashErr)

		token := &models.PasswordResetToken{
			UserID:    user.ID,
			Token:     hashedToken,
			ExpiresAt: time.Now().Add(-2 * time.Hour),
		}
		require.NoError(t, tokenRepo.CreatePasswordResetToken(ctx, token))

		err := service.ResetPassword(ctx, rawToken, "NewStrongPassword123!")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("weak password does not consume reset token", func(t *testing.T) {
		rawToken := "unconsumed-reset-token"
		hashedToken, hashErr := utils.HashPassword(rawToken)
		require.NoError(t, hashErr)

		token := &models.PasswordResetToken{
			UserID:    user.ID,
			Token:     hashedToken,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		require.NoError(t, tokenRepo.CreatePasswordResetToken(ctx, token))

		// Weak password fails validation
		err := service.ResetPassword(ctx, rawToken, "short")
		require.Error(t, err)

		// Token is still valid and can be used with a strong password
		err = service.ResetPassword(ctx, rawToken, "StrongPassword789!")
		require.NoError(t, err)
	})
}

// TestEmailVerification_NegativeCases tests failure modes for email verification.
func TestEmailVerification_NegativeCases(t *testing.T) {
	service, testDB, ctx := setupUserServiceTest(t)
	defer testDB.Close()

	tokenRepo := repository.NewTokenRepository(testDB)

	user, err := service.Register(ctx, &RegisterRequest{
		Username: "email_verify_user",
		Email:    "unverified@example.com",
		Password: "Password123!",
	})
	require.NoError(t, err)
	assert.False(t, user.EmailVerified)

	t.Run("unverified login behavior: code does not enforce unverified login rejection", func(t *testing.T) {
		authSvc, authDB, authCtx := setupAuthServiceTest(t)
		defer authDB.Close()

		uRepo := repository.NewUserRepository(authDB)
		hashedPassword, hashErr := utils.HashPassword("Password123!")
		require.NoError(t, hashErr)

		unverifiedUser := &models.User{
			Username:      "unverified_login_user",
			Email:         "unverified_login@example.com",
			Password:      hashedPassword,
			IsActive:      true,
			EmailVerified: false,
		}
		require.NoError(t, uRepo.Create(authCtx, unverifiedUser))

		// User has EmailVerified = false
		u, err := authSvc.Authenticate(authCtx, &AuthenticateRequest{
			UsernameOrEmail: "unverified_login@example.com",
			Password:        "Password123!",
		})
		// Verify that in the current implementation, unverified login is not rejected
		// because AuthService does not gate authentication on email verification.
		require.NoError(t, err)
		assert.Equal(t, unverifiedUser.ID, u.ID)
		assert.False(t, u.EmailVerified)
	})

	t.Run("unverified sign-in rejected when email verification is enabled", func(t *testing.T) {
		_, authDB, authCtx := setupAuthServiceTest(t)
		defer authDB.Close()

		uRepo := repository.NewUserRepository(authDB)
		sRepo := repository.NewSessionRepository(authDB)
		backendReg := &mockBackendRegistry{userRepo: uRepo}
		enabledAuthSvc := NewAuthService(uRepo, sRepo, backendReg, WithEmailVerificationRequired(true))

		hashedPassword, hashErr := utils.HashPassword("Password123!")
		require.NoError(t, hashErr)

		unverifiedUser := &models.User{
			Username:      "unverified_signin_user",
			Email:         "unverified_signin@example.com",
			Password:      hashedPassword,
			IsActive:      true,
			EmailVerified: false,
		}
		require.NoError(t, uRepo.Create(authCtx, unverifiedUser))

		// Attempt sign-in with unverified user: rejected
		_, err := enabledAuthSvc.Authenticate(authCtx, &AuthenticateRequest{
			UsernameOrEmail: "unverified_signin@example.com",
			Password:        "Password123!",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrEmailNotVerified)

		// Wrong password on unverified user returns ErrInvalidCredentials (no enumeration)
		_, err = enabledAuthSvc.Authenticate(authCtx, &AuthenticateRequest{
			UsernameOrEmail: "unverified_signin@example.com",
			Password:        "WrongPassword123!",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)

		// Once verified, sign-in succeeds
		unverifiedUser.EmailVerified = true
		require.NoError(t, uRepo.Update(authCtx, unverifiedUser))

		u, err := enabledAuthSvc.Authenticate(authCtx, &AuthenticateRequest{
			UsernameOrEmail: "unverified_signin@example.com",
			Password:        "Password123!",
		})
		require.NoError(t, err)
		assert.Equal(t, unverifiedUser.ID, u.ID)
	})

	t.Run("expiry: expired email verification token is rejected", func(t *testing.T) {
		rawToken := "expired-email-token"
		token := &models.EmailVerificationToken{
			UserID:    user.ID,
			Token:     rawToken,
			Email:     user.Email,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		require.NoError(t, tokenRepo.CreateEmailVerificationToken(ctx, token))

		err := service.VerifyEmail(ctx, rawToken)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTokenExpired)
	})

	t.Run("single-use: email verification token cannot be reused", func(t *testing.T) {
		// Create a separate user to verify
		u2, err := service.Register(ctx, &RegisterRequest{
			Username: "single_use_verify_user",
			Email:    "singleuse@example.com",
			Password: "Password123!",
		})
		require.NoError(t, err)

		token, err := service.CreateEmailVerificationToken(ctx, u2.ID, u2.Email)
		require.NoError(t, err)

		// First verification succeeds
		err = service.VerifyEmail(ctx, token.Token)
		require.NoError(t, err)

		// Reusing the same verification token must fail
		err = service.VerifyEmail(ctx, token.Token)
		require.Error(t, err)
	})

	t.Run("token for mismatched email is rejected", func(t *testing.T) {
		u3, err := service.Register(ctx, &RegisterRequest{
			Username: "mismatch_user",
			Email:    "mismatch@example.com",
			Password: "Password123!",
		})
		require.NoError(t, err)

		rawToken := "mismatch-email-token"
		token := &models.EmailVerificationToken{
			UserID:    u3.ID,
			Token:     rawToken,
			Email:     "different@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		require.NoError(t, tokenRepo.CreateEmailVerificationToken(ctx, token))

		err = service.VerifyEmail(ctx, rawToken)
		require.ErrorIs(t, err, ErrInvalidToken)
	})
}

// TestSessionExpiry_NegativeCases tests failure modes and cleanup for expired sessions.
func TestSessionExpiry_NegativeCases(t *testing.T) {
	service, testDB, ctx := setupAuthServiceTest(t)
	defer testDB.Close()

	userRepo := repository.NewUserRepository(testDB)
	sessionRepo := repository.NewSessionRepository(testDB)

	user := &models.User{
		Username: "session_user",
		Email:    "session_user@example.com",
		Password: "hash",
		IsActive: true,
	}
	require.NoError(t, userRepo.Create(ctx, user))

	t.Run("expired session returns error and is deleted from repository", func(t *testing.T) {
		past := time.Now().Add(-15 * time.Minute)
		expiredSession := &models.UserSession{
			UserID:     user.ID,
			SessionKey: "expired-key-1",
			ExpiresAt:  &past,
		}
		require.NoError(t, sessionRepo.Create(ctx, expiredSession))

		// Accessing expired session must fail with session expired
		_, err := service.GetSession(ctx, "expired-key-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session expired")

		// Expired session must have been deleted, so subsequent lookup returns session not found
		_, err = service.GetSession(ctx, "expired-key-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("refreshing expired session returns session expired error", func(t *testing.T) {
		past := time.Now().Add(-5 * time.Minute)
		expiredSession := &models.UserSession{
			UserID:     user.ID,
			SessionKey: "expired-key-2",
			ExpiresAt:  &past,
		}
		require.NoError(t, sessionRepo.Create(ctx, expiredSession))

		_, err := service.RefreshSession(ctx, "expired-key-2")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session expired")
	})
}
