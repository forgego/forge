package service

import (
	"context"

	"github.com/forgego/forge/identity/models"
)

// UserService defines the interface for user business logic
type UserService interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, req *CreateUserRequest) (*models.User, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, id int64, req *UpdateUserRequest) (*models.User, error)

	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, id int64) error

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id int64) (*models.User, error)

	// ListUsers retrieves users with filters
	ListUsers(ctx context.Context, filters *UserFilters) ([]*models.User, int64, error)

	// Register registers a new user (public endpoint)
	Register(ctx context.Context, req *RegisterRequest) (*models.User, error)

	// VerifyEmail verifies a user's email address
	VerifyEmail(ctx context.Context, token string) error

	// ResendVerificationEmail resends verification email
	ResendVerificationEmail(ctx context.Context, email string) error

	// CreateEmailVerificationToken creates a new email verification token
	CreateEmailVerificationToken(ctx context.Context, userID int64, email string) (*models.EmailVerificationToken, error)
}

// AuthService defines the interface for authentication business logic
type AuthService interface {
	// Authenticate authenticates a user with credentials
	Authenticate(ctx context.Context, req *AuthenticateRequest) (*models.User, error)

	// Logout logs out a user
	Logout(ctx context.Context, sessionKey string) error

	// LogoutAll logs out a user from all devices
	LogoutAll(ctx context.Context, userID int64) error

	// CreateSession creates a new session for a user
	CreateSession(ctx context.Context, userID int64, req *CreateSessionRequest) (*models.UserSession, error)

	// GetSession retrieves a session by key
	GetSession(ctx context.Context, key string) (*models.UserSession, error)

	// ListSessions lists all sessions for a user
	ListSessions(ctx context.Context, userID int64) ([]*models.UserSession, error)

	// RefreshSession refreshes a session
	RefreshSession(ctx context.Context, key string) (*models.UserSession, error)
}

// PasswordService defines the interface for password management
type PasswordService interface {
	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, userID int64, req *ChangePasswordRequest) error

	// RequestPasswordReset requests a password reset
	RequestPasswordReset(ctx context.Context, email string) error

	// ResetPassword resets a password using a token
	ResetPassword(ctx context.Context, token string, newPassword string) error

	// ValidatePassword validates a password against policy
	ValidatePassword(password string, user *models.User) error

	// CheckPassword checks if a password matches
	CheckPassword(user *models.User, password string) bool
}

// PermissionService defines the interface for permission management
type PermissionService interface {
	// CheckPermission checks if a user has a permission
	CheckPermission(ctx context.Context, userID int64, permission string) (bool, error)

	// CheckPermissions checks if a user has all specified permissions
	CheckPermissions(ctx context.Context, userID int64, permissions []string) (bool, error)

	// CheckAnyPermission checks if a user has any of the specified permissions
	CheckAnyPermission(ctx context.Context, userID int64, permissions []string) (bool, error)

	// GetUserPermissions retrieves all permissions for a user
	GetUserPermissions(ctx context.Context, userID int64) ([]*models.Permission, error)

	// AssignPermission assigns a permission to a user
	AssignPermission(ctx context.Context, userID int64, permission string) error

	// RemovePermission removes a permission from a user
	RemovePermission(ctx context.Context, userID int64, permission string) error
}

// Request/Response DTOs

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
	IsStaff   bool
	IsActive  bool
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Email       *string
	FirstName   *string
	LastName    *string
	IsActive    *bool
	IsStaff     *bool
	IsLocked    *bool
	IsSuperuser *bool
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

// AuthenticateRequest represents an authentication request
type AuthenticateRequest struct {
	UsernameOrEmail string
	Password        string
	RememberMe      bool
	IPAddress       string
	UserAgent       string
}

// CreateSessionRequest represents a request to create a session
type CreateSessionRequest struct {
	UserID     int64
	IPAddress  string
	UserAgent  string
	RememberMe bool
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string
	NewPassword     string
}

// UserFilters contains filter criteria for user queries
type UserFilters struct {
	Email         string
	Username      string
	IsActive      *bool
	IsStaff       *bool
	IsSuperuser   *bool
	IsLocked      *bool
	EmailVerified *bool
	Search        string
	Limit         int
	Offset        int
	OrderBy       []string
}

