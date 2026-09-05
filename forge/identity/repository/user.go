package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/models"
)

// scanNullableString scans a nullable string field
func scanNullableString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *db.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(database *db.DB) UserRepository {
	return &userRepository{db: database}
}

// normalizeEmail normalizes an email address (lowercase domain)
func normalizeEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[0] + "@" + strings.ToLower(parts[1])
	}
	return email
}

// normalizeUsername normalizes a username
func normalizeUsername(username string) string {
	return strings.TrimSpace(username)
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			username, email, password, first_name, last_name,
			is_active, is_staff, is_superuser, is_locked, email_verified,
			phone_number, phone_verified, timezone, locale, language,
			bio, website, location, avatar,
			date_joined, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		RETURNING id
	`

	now := time.Now()
	if user.DateJoined.IsZero() {
		user.DateJoined = now
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	err := r.db.QueryRowContext(ctx, query,
		normalizeUsername(user.Username),
		normalizeEmail(user.Email),
		user.Password,
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.IsStaff,
		user.IsSuperuser,
		user.IsLocked,
		user.EmailVerified,
		user.PhoneNumber,
		user.PhoneVerified,
		user.Timezone,
		user.Locale,
		user.Language,
		user.Bio,
		user.Website,
		user.Location,
		user.Avatar,
		user.DateJoined,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, username, email, password, first_name, last_name,
			is_active, is_staff, is_superuser, is_locked, email_verified,
			phone_number, phone_verified, timezone, locale, language,
			bio, website, location, avatar,
			password_changed_at, password_expires_at, must_change_password,
			locked_at, locked_reason, failed_login_attempts, last_failed_login_at,
			email_verified_at, date_joined, last_login, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := &models.User{}

	// Use sql.NullString for nullable string fields
	var firstName, lastName, phoneNumber, timezone, locale, language sql.NullString
	var bio, website, location, avatar, lockedReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&firstName, &lastName,
		&user.IsActive, &user.IsStaff, &user.IsSuperuser, &user.IsLocked, &user.EmailVerified,
		&phoneNumber, &user.PhoneVerified, &timezone, &locale, &language,
		&bio, &website, &location, &avatar,
		&user.PasswordChangedAt, &user.PasswordExpiresAt, &user.MustChangePassword,
		&user.LockedAt, &lockedReason, &user.FailedLoginAttempts, &user.LastFailedLoginAt,
		&user.EmailVerifiedAt, &user.DateJoined, &user.LastLogin,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	// Convert nullable strings to regular strings
	user.FirstName = scanNullableString(firstName)
	user.LastName = scanNullableString(lastName)
	user.PhoneNumber = scanNullableString(phoneNumber)
	user.Timezone = scanNullableString(timezone)
	user.Locale = scanNullableString(locale)
	user.Language = scanNullableString(language)
	user.Bio = scanNullableString(bio)
	user.Website = scanNullableString(website)
	user.Location = scanNullableString(location)
	user.Avatar = scanNullableString(avatar)
	user.LockedReason = scanNullableString(lockedReason)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email (normalized)
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password, first_name, last_name,
			is_active, is_staff, is_superuser, is_locked, email_verified,
			phone_number, phone_verified, timezone, locale, language,
			bio, website, location, avatar,
			password_changed_at, password_expires_at, must_change_password,
			locked_at, locked_reason, failed_login_attempts, last_failed_login_at,
			email_verified_at, date_joined, last_login, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &models.User{}
	normalizedEmail := normalizeEmail(email)

	// Use sql.NullString for nullable string fields
	var firstName, lastName, phoneNumber, timezone, locale, language sql.NullString
	var bio, website, location, avatar, lockedReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, normalizedEmail).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&firstName, &lastName,
		&user.IsActive, &user.IsStaff, &user.IsSuperuser, &user.IsLocked, &user.EmailVerified,
		&phoneNumber, &user.PhoneVerified, &timezone, &locale, &language,
		&bio, &website, &location, &avatar,
		&user.PasswordChangedAt, &user.PasswordExpiresAt, &user.MustChangePassword,
		&user.LockedAt, &lockedReason, &user.FailedLoginAttempts, &user.LastFailedLoginAt,
		&user.EmailVerifiedAt, &user.DateJoined, &user.LastLogin,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	// Convert nullable strings to regular strings
	user.FirstName = scanNullableString(firstName)
	user.LastName = scanNullableString(lastName)
	user.PhoneNumber = scanNullableString(phoneNumber)
	user.Timezone = scanNullableString(timezone)
	user.Locale = scanNullableString(locale)
	user.Language = scanNullableString(language)
	user.Bio = scanNullableString(bio)
	user.Website = scanNullableString(website)
	user.Location = scanNullableString(location)
	user.Avatar = scanNullableString(avatar)
	user.LockedReason = scanNullableString(lockedReason)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by username (normalized)
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password, first_name, last_name,
			is_active, is_staff, is_superuser, is_locked, email_verified,
			phone_number, phone_verified, timezone, locale, language,
			bio, website, location, avatar,
			password_changed_at, password_expires_at, must_change_password,
			locked_at, locked_reason, failed_login_attempts, last_failed_login_at,
			email_verified_at, date_joined, last_login, created_at, updated_at, deleted_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`

	user := &models.User{}
	normalizedUsername := normalizeUsername(username)

	// Use sql.NullString for nullable string fields
	var firstName, lastName, phoneNumber, timezone, locale, language sql.NullString
	var bio, website, location, avatar, lockedReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, normalizedUsername).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&firstName, &lastName,
		&user.IsActive, &user.IsStaff, &user.IsSuperuser, &user.IsLocked, &user.EmailVerified,
		&phoneNumber, &user.PhoneVerified, &timezone, &locale, &language,
		&bio, &website, &location, &avatar,
		&user.PasswordChangedAt, &user.PasswordExpiresAt, &user.MustChangePassword,
		&user.LockedAt, &lockedReason, &user.FailedLoginAttempts, &user.LastFailedLoginAt,
		&user.EmailVerifiedAt, &user.DateJoined, &user.LastLogin,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	// Convert nullable strings to regular strings
	user.FirstName = scanNullableString(firstName)
	user.LastName = scanNullableString(lastName)
	user.PhoneNumber = scanNullableString(phoneNumber)
	user.Timezone = scanNullableString(timezone)
	user.Locale = scanNullableString(locale)
	user.Language = scanNullableString(language)
	user.Bio = scanNullableString(bio)
	user.Website = scanNullableString(website)
	user.Location = scanNullableString(location)
	user.Avatar = scanNullableString(avatar)
	user.LockedReason = scanNullableString(lockedReason)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET
			username = $2, email = $3, password = $4, first_name = $5, last_name = $6,
			is_active = $7, is_staff = $8, is_superuser = $9, is_locked = $10, email_verified = $11,
			phone_number = $12, phone_verified = $13, timezone = $14, locale = $15, language = $16,
			bio = $17, website = $18, location = $19, avatar = $20,
			password_changed_at = $21, password_expires_at = $22, must_change_password = $23,
			locked_at = $24, locked_reason = $25, failed_login_attempts = $26, last_failed_login_at = $27,
			email_verified_at = $28, last_login = $29, updated_at = $30
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		normalizeUsername(user.Username),
		normalizeEmail(user.Email),
		user.Password,
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.IsStaff,
		user.IsSuperuser,
		user.IsLocked,
		user.EmailVerified,
		user.PhoneNumber,
		user.PhoneVerified,
		user.Timezone,
		user.Locale,
		user.Language,
		user.Bio,
		user.Website,
		user.Location,
		user.Avatar,
		user.PasswordChangedAt,
		user.PasswordExpiresAt,
		user.MustChangePassword,
		user.LockedAt,
		user.LockedReason,
		user.FailedLoginAttempts,
		user.LastFailedLoginAt,
		user.EmailVerifiedAt,
		user.LastLogin,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete deletes a user (soft delete)
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List retrieves users with filters
func (r *userRepository) List(ctx context.Context, filters *UserFilters) ([]*models.User, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filters.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, normalizeEmail(filters.Email))
		argIndex++
	}

	if filters.Username != "" {
		conditions = append(conditions, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, normalizeUsername(filters.Username))
		argIndex++
	}

	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filters.IsActive)
		argIndex++
	}

	if filters.IsStaff != nil {
		conditions = append(conditions, fmt.Sprintf("is_staff = $%d", argIndex))
		args = append(args, *filters.IsStaff)
		argIndex++
	}

	if filters.IsSuperuser != nil {
		conditions = append(conditions, fmt.Sprintf("is_superuser = $%d", argIndex))
		args = append(args, *filters.IsSuperuser)
		argIndex++
	}

	if filters.IsLocked != nil {
		conditions = append(conditions, fmt.Sprintf("is_locked = $%d", argIndex))
		args = append(args, *filters.IsLocked)
		argIndex++
	}

	if filters.EmailVerified != nil {
		conditions = append(conditions, fmt.Sprintf("email_verified = $%d", argIndex))
		args = append(args, *filters.EmailVerified)
		argIndex++
	}

	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(username ILIKE $%d OR email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)",
			argIndex, argIndex, argIndex, argIndex))
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
		argIndex += 4
	}

	query := `
		SELECT id, username, email, password, first_name, last_name,
			is_active, is_staff, is_superuser, is_locked, email_verified,
			phone_number, phone_verified, timezone, locale, language,
			bio, website, location, avatar,
			password_changed_at, password_expires_at, must_change_password,
			locked_at, locked_reason, failed_login_attempts, last_failed_login_at,
			email_verified_at, date_joined, last_login, created_at, updated_at, deleted_at
		FROM users
		WHERE ` + strings.Join(conditions, " AND ")

	// Add ordering
	if len(filters.OrderBy) > 0 {
		query += " ORDER BY " + strings.Join(filters.OrderBy, ", ")
	} else {
		query += " ORDER BY created_at DESC"
	}

	// Add limit and offset
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}

		// Use sql.NullString for nullable string fields
		var firstName, lastName, phoneNumber, timezone, locale, language sql.NullString
		var bio, website, location, avatar, lockedReason sql.NullString

		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password,
			&firstName, &lastName,
			&user.IsActive, &user.IsStaff, &user.IsSuperuser, &user.IsLocked, &user.EmailVerified,
			&phoneNumber, &user.PhoneVerified, &timezone, &locale, &language,
			&bio, &website, &location, &avatar,
			&user.PasswordChangedAt, &user.PasswordExpiresAt, &user.MustChangePassword,
			&user.LockedAt, &lockedReason, &user.FailedLoginAttempts, &user.LastFailedLoginAt,
			&user.EmailVerifiedAt, &user.DateJoined, &user.LastLogin,
			&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		// Convert nullable strings to regular strings
		user.FirstName = scanNullableString(firstName)
		user.LastName = scanNullableString(lastName)
		user.PhoneNumber = scanNullableString(phoneNumber)
		user.Timezone = scanNullableString(timezone)
		user.Locale = scanNullableString(locale)
		user.Language = scanNullableString(language)
		user.Bio = scanNullableString(bio)
		user.Website = scanNullableString(website)
		user.Location = scanNullableString(location)
		user.Avatar = scanNullableString(avatar)
		user.LockedReason = scanNullableString(lockedReason)

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// Count counts users matching filters
func (r *userRepository) Count(ctx context.Context, filters *UserFilters) (int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filters.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, normalizeEmail(filters.Email))
		argIndex++
	}

	if filters.Username != "" {
		conditions = append(conditions, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, normalizeUsername(filters.Username))
		argIndex++
	}

	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filters.IsActive)
		argIndex++
	}

	if filters.IsStaff != nil {
		conditions = append(conditions, fmt.Sprintf("is_staff = $%d", argIndex))
		args = append(args, *filters.IsStaff)
		argIndex++
	}

	if filters.IsSuperuser != nil {
		conditions = append(conditions, fmt.Sprintf("is_superuser = $%d", argIndex))
		args = append(args, *filters.IsSuperuser)
		argIndex++
	}

	if filters.IsLocked != nil {
		conditions = append(conditions, fmt.Sprintf("is_locked = $%d", argIndex))
		args = append(args, *filters.IsLocked)
		argIndex++
	}

	if filters.EmailVerified != nil {
		conditions = append(conditions, fmt.Sprintf("email_verified = $%d", argIndex))
		args = append(args, *filters.EmailVerified)
		argIndex++
	}

	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(username ILIKE $%d OR email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)",
			argIndex, argIndex, argIndex, argIndex))
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	query := "SELECT COUNT(*) FROM users WHERE " + strings.Join(conditions, " AND ")

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// Exists checks if a user exists with the given email
func (r *userRepository) Exists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, normalizeEmail(email)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}

// ExistsUsername checks if a user exists with the given username
func (r *userRepository) ExistsUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, normalizeUsername(username)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if username exists: %w", err)
	}

	return exists, nil
}

