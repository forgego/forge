package config

import "time"

// IdentityConfig holds configuration for the identity system
type IdentityConfig struct {
	// Password policy
	PasswordPolicy PasswordPolicy

	// Account lockout
	LockoutConfig LockoutConfig

	// Rate limiting
	RateLimitConfig RateLimitConfig

	// Email settings
	EmailVerificationRequired bool
	EmailFromAddress          string

	// Session settings
	SessionTimeout         time.Duration
	RememberMeDuration     time.Duration
	SessionCleanupInterval time.Duration
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumbers   bool
	RequireSymbols   bool
	MaxAge           time.Duration // Password expiration (0 = no expiration)
	HistoryCount     int           // Prevent reuse of last N passwords
}

// DefaultPasswordPolicy returns default password policy
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:        12,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSymbols:   false, // Optional by default
		MaxAge:           0,     // No expiration by default
		HistoryCount:     5,     // Remember last 5 passwords
	}
}

// LockoutConfig defines account lockout settings
type LockoutConfig struct {
	MaxFailedAttempts int           // Maximum failed login attempts
	LockoutDuration   time.Duration // Lockout duration (0 = permanent until manual unlock)
	PermanentLockout  bool          // Require admin unlock
	ResetWindow       time.Duration // Window to reset counter
}

// DefaultLockoutConfig returns default lockout configuration
func DefaultLockoutConfig() LockoutConfig {
	return LockoutConfig{
		MaxFailedAttempts: 5,
		LockoutDuration:   15 * time.Minute,
		PermanentLockout:  false,
		ResetWindow:       15 * time.Minute,
	}
}

// RateLimitConfig defines rate limiting settings
type RateLimitConfig struct {
	Login             RateLimit
	PasswordReset     RateLimit
	EmailVerification RateLimit
	Registration      RateLimit
}

// RateLimit defines a rate limit
type RateLimit struct {
	Limit  int           // Number of requests
	Window time.Duration // Time window
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Login: RateLimit{
			Limit:  5,
			Window: 15 * time.Minute,
		},
		PasswordReset: RateLimit{
			Limit:  3,
			Window: 1 * time.Hour,
		},
		EmailVerification: RateLimit{
			Limit:  3,
			Window: 1 * time.Hour,
		},
		Registration: RateLimit{
			Limit:  5,
			Window: 1 * time.Hour,
		},
	}
}

// DefaultIdentityConfig returns default identity system configuration
func DefaultIdentityConfig() *IdentityConfig {
	return &IdentityConfig{
		PasswordPolicy:            DefaultPasswordPolicy(),
		LockoutConfig:             DefaultLockoutConfig(),
		RateLimitConfig:           DefaultRateLimitConfig(),
		EmailVerificationRequired: false, // Optional by default
		EmailFromAddress:          "noreply@example.com",
		SessionTimeout:            24 * time.Hour,
		RememberMeDuration:        30 * 24 * time.Hour, // 30 days
		SessionCleanupInterval:    1 * time.Hour,
	}
}
