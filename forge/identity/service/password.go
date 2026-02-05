package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/forgego/forge/identity/config"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/utils"
)

// passwordService implements PasswordService interface
type passwordService struct {
	userRepo       repository.UserRepository
	tokenRepo      repository.TokenRepository
	passwordPolicy config.PasswordPolicy
}

// NewPasswordService creates a new password service
func NewPasswordService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	policy config.PasswordPolicy,
) PasswordService {
	return &passwordService{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		passwordPolicy: policy,
	}
}

// ChangePassword changes a user's password
func (s *passwordService) ChangePassword(ctx context.Context, userID int64, req *ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Verify current password
	if !utils.CheckPassword(req.CurrentPassword, user.Password) {
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password
	if err := s.ValidatePassword(req.NewPassword, user); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	now := time.Now()
	user.Password = hashedPassword
	user.PasswordChangedAt = &now
	user.MustChangePassword = false

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// RequestPasswordReset requests a password reset
func (s *passwordService) RequestPasswordReset(ctx context.Context, email string) error {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not (security best practice)
		return nil
	}

	// Generate reset token
	token, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash token before storing
	hashedToken, err := utils.HashPassword(token)
	if err != nil {
		return fmt.Errorf("failed to hash token: %w", err)
	}

	// Create reset token
	resetToken := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     hashedToken,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour expiry
	}

	if err := s.tokenRepo.CreatePasswordResetToken(ctx, resetToken); err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	// TODO: Send email with token
	// For now, just return success
	// In production, you would send an email with the token

	return nil
}

// ResetPassword resets a password using a token
func (s *passwordService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	// Get reset token
	resetToken, err := s.tokenRepo.GetPasswordResetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}

	// Check if token is used
	if resetToken.IsUsed() {
		return fmt.Errorf("token has already been used")
	}

	// Check if token is expired
	if resetToken.IsExpired() {
		return fmt.Errorf("token has expired")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, resetToken.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	// Validate new password
	if err := s.ValidatePassword(newPassword, user); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	now := time.Now()
	user.Password = hashedPassword
	user.PasswordChangedAt = &now
	user.MustChangePassword = false

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	resetToken.UsedAt = &now
	// Note: TokenRepository would need an Update method for this
	// For now, we'll delete it
	_ = s.tokenRepo.DeletePasswordResetToken(ctx, token)

	return nil
}

// ValidatePassword validates a password against policy
func (s *passwordService) ValidatePassword(password string, user *models.User) error {
	policy := s.passwordPolicy

	// Check minimum length
	if len(password) < policy.MinLength {
		return fmt.Errorf("password must be at least %d characters", policy.MinLength)
	}

	// Check complexity requirements
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSymbol := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	if policy.RequireUppercase && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if policy.RequireLowercase && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if policy.RequireNumbers && !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if policy.RequireSymbols && !hasSymbol {
		return fmt.Errorf("password must contain at least one symbol")
	}

	// Check against common passwords (simplified)
	commonPasswords := []string{"password", "123456", "password123", "admin", "letmein"}
	passwordLower := strings.ToLower(password)
	for _, common := range commonPasswords {
		if passwordLower == common {
			return fmt.Errorf("password is too common")
		}
	}

	return nil
}

// CheckPassword checks if a password matches
func (s *passwordService) CheckPassword(user *models.User, password string) bool {
	return utils.CheckPassword(password, user.Password)
}

// generateSecureToken generates a secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
