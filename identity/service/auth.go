package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/forgego/forge/identity/backends"
	"github.com/forgego/forge/identity/config"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
)

var ErrInvalidCredentials = fmt.Errorf("invalid credentials")

// authService implements AuthService interface
type authService struct {
	userRepo                 repository.UserRepository
	sessionRepo              repository.SessionRepository
	backendRegistry          backends.BackendRegistry
	requireEmailVerification bool
}

// AuthServiceOption configures an AuthService instance.
type AuthServiceOption func(*authService)

// WithEmailVerificationRequired configures whether email verification is required for authentication.
func WithEmailVerificationRequired(required bool) AuthServiceOption {
	return func(s *authService) {
		s.requireEmailVerification = required
	}
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	backendRegistry backends.BackendRegistry,
	opts ...AuthServiceOption,
) AuthService {
	svc := &authService{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		backendRegistry: backendRegistry,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// NewAuthServiceWithConfig creates a new authentication service with identity configuration.
func NewAuthServiceWithConfig(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	backendRegistry backends.BackendRegistry,
	cfg *config.IdentityConfig,
) AuthService {
	var opts []AuthServiceOption
	if cfg != nil && cfg.EmailVerificationRequired {
		opts = append(opts, WithEmailVerificationRequired(true))
	}
	return NewAuthService(userRepo, sessionRepo, backendRegistry, opts...)
}

// Authenticate authenticates a user with credentials using the backend registry
func (s *authService) Authenticate(ctx context.Context, req *AuthenticateRequest) (*models.User, error) {
	// Build credentials map - password backend checks for "username" or "email" key
	// It will determine if the value is an email or username internally
	credentials := map[string]string{
		"username": req.UsernameOrEmail, // Backend will check if it's email or username
		"password": req.Password,
	}

	// Try authentication with backend registry
	user, err := s.backendRegistry.Authenticate(ctx, credentials)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := user.AccountStateError(); err != nil {
		return nil, err
	}

	if s.requireEmailVerification && !user.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	// Update last login
	now := time.Now()
	if mutations, ok := s.userRepo.(repository.UserMutations); ok {
		if err := mutations.SetLastLogin(ctx, user.ID, now); err != nil {
			return nil, fmt.Errorf("failed to update last login: %w", err)
		}
	} else {
		user.LastLogin = &now
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to update last login: %w", err)
		}
	}
	user.LastLogin = &now

	return user, nil
}

// Logout logs out a user
func (s *authService) Logout(ctx context.Context, sessionKey string) error {
	return s.sessionRepo.Delete(ctx, sessionKey)
}

// LogoutAll logs out a user from all devices
func (s *authService) LogoutAll(ctx context.Context, userID int64) error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

// CreateSession creates a new session for a user
func (s *authService) CreateSession(ctx context.Context, userID int64, req *CreateSessionRequest) (*models.UserSession, error) {
	// Generate session key
	sessionKey, err := generateSessionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session key: %w", err)
	}

	// Calculate expiration
	var expiresAt *time.Time
	if req.RememberMe {
		exp := time.Now().Add(30 * 24 * time.Hour) // 30 days
		expiresAt = &exp
	}

	// Create session
	session := &models.UserSession{
		UserID:       userID,
		SessionKey:   sessionKey,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		LastActivity: time.Now(),
		CreatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
		IsRememberMe: req.RememberMe,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSession retrieves a session by key
func (s *authService) GetSession(ctx context.Context, key string) (*models.UserSession, error) {
	session, err := s.sessionRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Check if expired
	if session.IsExpired() {
		// Delete expired session
		if err := s.sessionRepo.Delete(ctx, key); err != nil {
			return nil, fmt.Errorf("failed to delete expired session: %w", err)
		}
		return nil, fmt.Errorf("session expired")
	}
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("session user not found: %w", err)
	}
	if err := user.AccountStateError(); err != nil {
		return nil, err
	}

	return session, nil
}

// ListSessions lists all sessions for a user
func (s *authService) ListSessions(ctx context.Context, userID int64) ([]*models.UserSession, error) {
	return s.sessionRepo.GetByUserID(ctx, userID)
}

// RefreshSession refreshes a session
func (s *authService) RefreshSession(ctx context.Context, key string) (*models.UserSession, error) {
	session, err := s.GetSession(ctx, key)
	if err != nil {
		return nil, err
	}

	activity, ok := s.sessionRepo.(repository.SessionActivityRepository)
	// Update last activity without writing a possibly stale expiry.
	session.LastActivity = time.Now()
	var updateErr error
	if ok {
		updateErr = activity.TouchLastActivity(ctx, key, session.LastActivity)
	} else {
		updateErr = s.sessionRepo.Update(ctx, session)
	}
	if err := updateErr; err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	return session, nil
}

// generateSessionKey generates a secure random session key
func generateSessionKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
