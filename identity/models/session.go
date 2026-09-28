package models

import (
	"time"
)

// UserSession represents a user session
type UserSession struct {
	ID           int64      `json:"id" db:"id"`
	UserID       int64      `json:"user_id" db:"user_id"`
	SessionKey   string     `json:"session_key" db:"session_key"` // Unique session identifier
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	LastActivity time.Time  `json:"last_activity" db:"last_activity"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"` // nil for non-expiring sessions
	IsRememberMe bool       `json:"is_remember_me" db:"is_remember_me"`
}

// IsExpired checks if the session is expired
func (s *UserSession) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

// IsActive checks if the session is active (not expired)
func (s *UserSession) IsActive() bool {
	return !s.IsExpired()
}
