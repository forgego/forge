package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/forgego/forge/identity/backends"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
)

// authService implements AuthService interface
type authService struct {
	userRepo        repository.UserRepository
	sessionRepo     repository.SessionRepository
	backendRegistry backends.BackendRegistry
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	backendRegistry backends.BackendRegistry,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		backendRegistry: backendRegistry,
	}
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
		return nil, fmt.Errorf("invalid credentials")
	}

	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if active (backend already checks this, but double-check for safety)
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Check if locked (backend already checks this, but double-check for safety)
	if user.IsLocked {
		return nil, ErrUserLocked
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		// Log error but don't fail authentication
		fmt.Printf("Warning: failed to update last_login: %v\n", err)
	}

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
		_ = s.sessionRepo.Delete(ctx, key)
		return nil, fmt.Errorf("session expired")
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

	// Update last activity
	session.LastActivity = time.Now()
	if err := s.sessionRepo.Update(ctx, session); err != nil {
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
