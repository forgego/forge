package backends

import (
	"context"
	"fmt"
	"strings"

	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
)

// tokenBackend implements token-based authentication
type tokenBackend struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

// NewTokenBackend creates a new token authentication backend
func NewTokenBackend(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
) AuthenticationBackend {
	return &tokenBackend{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// Name returns the backend name
func (b *tokenBackend) Name() string {
	return "token"
}

// Supports returns true if this backend can handle the credential type
func (b *tokenBackend) Supports(credentialType string) bool {
	return credentialType == "token" || credentialType == "bearer"
}

// Authenticate attempts to authenticate using a token
func (b *tokenBackend) Authenticate(ctx context.Context, credentials map[string]string) (*models.User, error) {
	// Get token
	token, ok := credentials["token"]
	if !ok {
		token, ok = credentials["bearer"]
		if !ok {
			return nil, nil // Not applicable
		}
	}

	// Get user by token (via session)
	return b.GetUser(ctx, token)
}

// GetUser retrieves a user by token/session key
func (b *tokenBackend) GetUser(ctx context.Context, identifier string) (*models.User, error) {
	// Try to get session by key
	session, err := b.sessionRepo.GetByKey(ctx, identifier)
	if err != nil {
		return nil, nil // Not found, not an error
	}

	// Check if session is expired
	if session.IsExpired() {
		return nil, fmt.Errorf("session expired")
	}

	// Get user
	user, err := b.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Check if user is locked
	if user.IsLocked {
		return nil, fmt.Errorf("user account is locked")
	}

	return user, nil
}

// ExtractTokenFromHeader extracts token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	// Check for Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}
