package backends

import (
	"context"
	"database/sql"
	"errors"
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

var dummyPasswordHash = mustHash("forge-dummy-password")

var comparePassword = utils.CheckPassword

func mustHash(password string) string {
	hash, err := utils.HashPassword(password)
	if err != nil {
		panic(fmt.Sprintf("failed to hash dummy password: %v", err))
	}
	return hash
}

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

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, sql.ErrNoRows)
}

func (b *passwordBackend) lookupUser(ctx context.Context, identifier string) (*models.User, error) {
	if strings.Contains(identifier, "@") {
		return b.userRepo.GetByEmail(ctx, identifier)
	}
	return b.userRepo.GetByUsername(ctx, identifier)
}

func checkUserStatus(user *models.User) error {
	if !user.IsActive {
		return ErrUserInactive
	}
	if user.IsLocked {
		return ErrUserLocked
	}
	return nil
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
	user, err := b.lookupUser(ctx, usernameOrEmail)
	if err != nil {
		if isNotFoundError(err) {
			comparePassword(password, dummyPasswordHash)
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("authenticate: lookup user: %w", err)
	}

	// Check password
	if !comparePassword(password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Check user status
	if err := checkUserStatus(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieves a user by identifier (not supported for password backend)
func (b *passwordBackend) GetUser(ctx context.Context, identifier string) (*models.User, error) {
	// Password backend doesn't support GetUser by identifier
	// This is used for token-based authentication
	return nil, nil
}
