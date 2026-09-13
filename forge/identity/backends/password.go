package backends

import (
	"context"
	"fmt"
	"strings"

	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/utils"
)

// Predefined errors
var (
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrUserInactive       = fmt.Errorf("user account is inactive")
	ErrUserLocked         = fmt.Errorf("user account is locked")
)

// passwordBackend implements password-based authentication
type passwordBackend struct {
	userRepo repository.UserRepository
}

// NewPasswordBackend creates a new password authentication backend
func NewPasswordBackend(userRepo repository.UserRepository) AuthenticationBackend {
	return &passwordBackend{userRepo: userRepo}
}

// Name returns the backend name
func (b *passwordBackend) Name() string {
	return "password"
}

// Supports returns true if this backend can handle the credential type
func (b *passwordBackend) Supports(credentialType string) bool {
	return credentialType == "password"
}

// Authenticate attempts to authenticate using username/email and password
func (b *passwordBackend) Authenticate(ctx context.Context, credentials map[string]string) (*models.User, error) {
	// Get username or email
	usernameOrEmail, ok := credentials["username"]
	if !ok {
		usernameOrEmail, ok = credentials["email"]
		if !ok {
			return nil, nil // Not applicable - no username/email provided
		}
	}

	// Get password
	password, ok := credentials["password"]
	if !ok {
		return nil, nil // Not applicable - no password provided
	}

	// Get user by username or email
	var user *models.User
	var err error

	if strings.Contains(usernameOrEmail, "@") {
		// Looks like an email
		user, err = b.userRepo.GetByEmail(ctx, usernameOrEmail)
	} else {
		// Looks like a username
		user, err = b.userRepo.GetByUsername(ctx, usernameOrEmail)
	}

	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Check if user is locked
	if user.IsLocked {
		return nil, ErrUserLocked
	}

	// Check password
	if !utils.CheckPassword(password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetUser retrieves a user by identifier (not supported for password backend)
func (b *passwordBackend) GetUser(ctx context.Context, identifier string) (*models.User, error) {
	// Password backend doesn't support GetUser by identifier
	// This is used for token-based authentication
	return nil, nil
}
