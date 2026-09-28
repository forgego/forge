package models

import (
	"time"
)

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID         int64      `json:"id" db:"id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	Token      string     `json:"token" db:"token"` // Hashed token
	Email      string     `json:"email" db:"email"` // Email being verified
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at" db:"verified_at"`
}

// IsExpired checks if the token is expired
func (t *EmailVerificationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used
func (t *EmailVerificationToken) IsUsed() bool {
	return t.VerifiedAt != nil
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        int64      `json:"id" db:"id"`
	UserID    int64      `json:"user_id" db:"user_id"`
	Token     string     `json:"token" db:"token"` // Hashed token
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at" db:"used_at"`
}

// IsExpired checks if the token is expired
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}
