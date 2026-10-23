package repository

import (
	"context"
	"time"

	"github.com/forgego/forge/identity/models"
)

// UserMutations exposes narrow and transactional identity writes without changing
// the existing UserRepository contract used by custom repositories.
// Custom repositories that do not implement this interface use full-row Update
// fallbacks for login, password changes, resets, profile changes, and email
// verification. Those fallbacks cannot prevent a concurrent lock, password, or
// email change from being overwritten. Even with UserMutations, repositories
// without UserEmailMutations cannot condition email updates or email_verified
// writes on the address read earlier. Implement both mutation interfaces to
// preserve those concurrent-change checks.
type UserMutations interface {
	SetLastLogin(ctx context.Context, userID int64, at time.Time) error
	UpdateFields(ctx context.Context, userID int64, fields map[string]any) error
	ChangePasswordIfCurrent(ctx context.Context, userID int64, oldHash, newHash string, at time.Time) error
	CommitPasswordReset(ctx context.Context, tokenID, userID int64, newHash string, at time.Time) error
	CommitEmailVerification(ctx context.Context, tokenID, userID int64, email string, at time.Time) error
}

// UserEmailMutations conditionally applies email-bound fields in one write.
// Implementations must reject an address changed since read, including when
// the submitted email is unchanged. Without this interface, email-bound writes
// cannot be protected against a concurrent address change.
type UserEmailMutations interface {
	UpdateEmailIfCurrent(ctx context.Context, userID int64, oldEmail string, fields map[string]any) error
}

// AccountStateMutations atomically updates account status and revokes sessions.
type AccountStateMutations interface {
	SetAccountStateAndRevokeSessions(ctx context.Context, userID int64, fields map[string]any) error
}

// PasswordSessionMutations atomically changes a password and revokes every
// session except the session identified by currentSessionKey.
type PasswordSessionMutations interface {
	ChangePasswordAndRevokeOtherSessions(ctx context.Context, userID int64, oldHash, newHash, currentSessionKey string, at time.Time) error
}

// TokenTransactionRepository reports whether token writes and user mutations
// share one transaction. The service only uses this claim for the built-in SQL
// repositories; custom token stores use the fail-closed fallback unless they
// implement TokenCommitRepository.
type TokenTransactionRepository interface {
	SharesUserTransaction(users UserRepository) bool
}

// TokenCommitRepository lets a custom token store explicitly own atomic token
// consumption and user mutation. Implementations must commit both writes in one
// transaction or return an error without changing either record.
type TokenCommitRepository interface {
	CommitPasswordReset(ctx context.Context, users UserMutations, tokenID, userID int64, newHash string, at time.Time) error
	CommitEmailVerification(ctx context.Context, users UserMutations, tokenID, userID int64, email string, at time.Time) error
}

// SessionActivityRepository updates activity without rewriting a stale expiry.
type SessionActivityRepository interface {
	TouchLastActivity(ctx context.Context, key string, at time.Time) error
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *models.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id int64) (*models.User, error)

	// GetByEmail retrieves a user by email (normalized)
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// GetByUsername retrieves a user by username (normalized)
	GetByUsername(ctx context.Context, username string) (*models.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *models.User) error

	// Delete deletes a user (soft delete if supported)
	Delete(ctx context.Context, id int64) error

	// List retrieves users with filters
	List(ctx context.Context, filters *UserFilters) ([]*models.User, error)

	// Count counts users matching filters
	Count(ctx context.Context, filters *UserFilters) (int64, error)

	// Exists checks if a user exists with the given email
	Exists(ctx context.Context, email string) (bool, error)

	// ExistsUsername checks if a user exists with the given username
	ExistsUsername(ctx context.Context, username string) (bool, error)
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
	Search        string // Full-text search
	Limit         int
	Offset        int
	OrderBy       []string
}

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *models.UserSession) error

	// GetByKey retrieves a session by session key
	GetByKey(ctx context.Context, key string) (*models.UserSession, error)

	// GetByUserID retrieves all sessions for a user
	GetByUserID(ctx context.Context, userID int64) ([]*models.UserSession, error)

	// Update updates a session
	Update(ctx context.Context, session *models.UserSession) error

	// Delete deletes a session
	Delete(ctx context.Context, key string) error

	// DeleteByUserID deletes all sessions for a user
	DeleteByUserID(ctx context.Context, userID int64) error

	// DeleteExpired deletes all expired sessions
	DeleteExpired(ctx context.Context) error
}

// PermissionRepository defines the interface for permission data access
type PermissionRepository interface {
	// GetByID retrieves a permission by ID
	GetByID(ctx context.Context, id int64) (*models.Permission, error)

	// GetByCodename retrieves a permission by codename
	GetByCodename(ctx context.Context, codename string) (*models.Permission, error)

	// GetByUserID retrieves all permissions for a user (direct + via groups)
	GetByUserID(ctx context.Context, userID int64) ([]*models.Permission, error)

	// AssignToUser assigns a permission to a user
	AssignToUser(ctx context.Context, userID, permissionID int64) error

	// RemoveFromUser removes a permission from a user
	RemoveFromUser(ctx context.Context, userID, permissionID int64) error

	// UserHasPermission checks if a user has a specific permission
	UserHasPermission(ctx context.Context, userID int64, codename string) (bool, error)
}

// GroupRepository defines the interface for group data access
type GroupRepository interface {
	// Create creates a new group
	Create(ctx context.Context, group *models.Group) error

	// GetByID retrieves a group by ID
	GetByID(ctx context.Context, id int64) (*models.Group, error)

	// GetByName retrieves a group by name
	GetByName(ctx context.Context, name string) (*models.Group, error)

	// Update updates a group
	Update(ctx context.Context, group *models.Group) error

	// Delete deletes a group
	Delete(ctx context.Context, id int64) error

	// List retrieves all groups
	List(ctx context.Context) ([]*models.Group, error)

	// GetByUserID retrieves all groups for a user
	GetByUserID(ctx context.Context, userID int64) ([]*models.Group, error)

	// AddUser adds a user to a group
	AddUser(ctx context.Context, groupID, userID int64) error

	// RemoveUser removes a user from a group
	RemoveUser(ctx context.Context, groupID, userID int64) error

	// AddPermission adds a permission to a group
	AddPermission(ctx context.Context, groupID, permissionID int64) error

	// RemovePermission removes a permission from a group
	RemovePermission(ctx context.Context, groupID, permissionID int64) error
}

// TokenRepository defines the interface for token data access
type TokenRepository interface {
	// CreateEmailVerificationToken creates an email verification token
	CreateEmailVerificationToken(ctx context.Context, token *models.EmailVerificationToken) error

	// GetEmailVerificationToken retrieves an email verification token
	GetEmailVerificationToken(ctx context.Context, token string) (*models.EmailVerificationToken, error)

	// MarkEmailVerificationTokenUsed marks an email verification token as used by ID
	MarkEmailVerificationTokenUsed(ctx context.Context, tokenID int64) error

	// DeleteEmailVerificationToken deletes an email verification token
	DeleteEmailVerificationToken(ctx context.Context, token string) error

	// CreatePasswordResetToken creates a password reset token
	CreatePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) error

	// GetPasswordResetToken retrieves a password reset token
	GetPasswordResetToken(ctx context.Context, token string) (*models.PasswordResetToken, error)

	// DeletePasswordResetToken deletes a password reset token
	DeletePasswordResetToken(ctx context.Context, token string) error

	// DeleteExpiredTokens deletes all expired tokens
	DeleteExpiredTokens(ctx context.Context) error
}
