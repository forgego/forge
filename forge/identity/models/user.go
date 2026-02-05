package models

import (
	"time"

	"github.com/forgego/forge/schema"
)

// User represents a user in the system
// This is the enhanced user model with all features from the roadmap
type User struct {
	schema.BaseSchema

	// Core identification
	ID       int64  `json:"id" db:"id"`
	Username string `json:"username" db:"username" validate:"required,max=150"`
	Email    string `json:"email" db:"email" validate:"required,email,max=254"`
	Password string `json:"-" db:"password" validate:"required,max=128"` // Never serialize password

	// Personal information
	FirstName string `json:"first_name" db:"first_name" validate:"max=150"`
	LastName  string `json:"last_name" db:"last_name" validate:"max=150"`
	Bio       string `json:"bio" db:"bio"`
	Website   string `json:"website" db:"website"`
	Location  string `json:"location" db:"location"`
	Avatar    string `json:"avatar" db:"avatar"`

	// Contact information
	PhoneNumber   string `json:"phone_number" db:"phone_number"`
	PhoneVerified bool   `json:"phone_verified" db:"phone_verified"`

	// Preferences
	Timezone string `json:"timezone" db:"timezone"`
	Locale   string `json:"locale" db:"locale"`
	Language string `json:"language" db:"language"`

	// Status flags
	IsActive      bool `json:"is_active" db:"is_active"`
	IsStaff       bool `json:"is_staff" db:"is_staff"`
	IsSuperuser   bool `json:"is_superuser" db:"is_superuser"`
	IsLocked      bool `json:"is_locked" db:"is_locked"`
	EmailVerified bool `json:"email_verified" db:"email_verified"`

	// Account security
	PasswordChangedAt  *time.Time `json:"password_changed_at" db:"password_changed_at"`
	PasswordExpiresAt  *time.Time `json:"password_expires_at" db:"password_expires_at"`
	MustChangePassword bool       `json:"must_change_password" db:"must_change_password"`

	// Account lockout
	LockedAt            *time.Time `json:"locked_at" db:"locked_at"`
	LockedReason        string     `json:"locked_reason" db:"locked_reason"`
	FailedLoginAttempts int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LastFailedLoginAt   *time.Time `json:"last_failed_login_at" db:"last_failed_login_at"`

	// Email verification
	EmailVerifiedAt *time.Time `json:"email_verified_at" db:"email_verified_at"`

	// Timestamps
	DateJoined time.Time  `json:"date_joined" db:"date_joined"`
	LastLogin  *time.Time `json:"last_login" db:"last_login"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at" db:"deleted_at"` // Soft delete
}

// GetID returns the user's ID (implements ModelWithID)
func (u *User) GetID() int64 {
	return u.ID
}

// SetID sets the user's ID (implements ModelWithID)
func (u *User) SetID(id int64) {
	u.ID = id
}

// IsAuthenticated returns true if user is authenticated (for template use)
func (u *User) IsAuthenticated() bool {
	return u.ID > 0
}

// IsAnonymous returns false (for template use)
func (u *User) IsAnonymous() bool {
	return false
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	if u.FirstName != "" {
		return u.FirstName
	}
	if u.LastName != "" {
		return u.LastName
	}
	return u.Username
}

// GetDisplayName returns the display name (full name or username)
func (u *User) GetDisplayName() string {
	fullName := u.GetFullName()
	if fullName != "" && fullName != u.Username {
		return fullName
	}
	return u.Username
}
