package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/utils"
)

// tokenRepository implements TokenRepository interface
type tokenRepository struct {
	db *db.DB
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(database *db.DB) TokenRepository {
	return &tokenRepository{db: database}
}

// CreateEmailVerificationToken creates an email verification token
func (r *tokenRepository) CreateEmailVerificationToken(ctx context.Context, token *models.EmailVerificationToken) error {
	query := `
		INSERT INTO email_verification_tokens (
			user_id, token, email, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}

	err := r.db.QueryRowContext(ctx, query,
		token.UserID,
		token.Token,
		token.Email,
		token.CreatedAt,
		token.ExpiresAt,
	).Scan(&token.ID)

	if err != nil {
		return fmt.Errorf("failed to create email verification token: %w", err)
	}

	return nil
}

// GetEmailVerificationToken retrieves an email verification token
func (r *tokenRepository) GetEmailVerificationToken(ctx context.Context, token string) (*models.EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token, email, created_at, expires_at, verified_at
		FROM email_verification_tokens
		WHERE token = $1
	`

	verificationToken := &models.EmailVerificationToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&verificationToken.ID,
		&verificationToken.UserID,
		&verificationToken.Token,
		&verificationToken.Email,
		&verificationToken.CreatedAt,
		&verificationToken.ExpiresAt,
		&verificationToken.VerifiedAt,
	)

	if err == sql.ErrNoRows {
		// Backward-compatible fallback:
		// email verification tokens are stored hashed, so plaintext lookup requires
		// scanning valid rows and verifying with bcrypt comparison.
		return r.getEmailVerificationTokenByHash(ctx, token)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get email verification token: %w", err)
	}

	return verificationToken, nil
}

// getEmailVerificationTokenByHash finds a token by comparing the plaintext with stored hashes
func (r *tokenRepository) getEmailVerificationTokenByHash(ctx context.Context, token string) (*models.EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token, email, created_at, expires_at, verified_at
		FROM email_verification_tokens
		WHERE verified_at IS NULL
		  AND expires_at > NOW()
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query email verification tokens: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		verificationToken := &models.EmailVerificationToken{}
		if scanErr := rows.Scan(
			&verificationToken.ID,
			&verificationToken.UserID,
			&verificationToken.Token,
			&verificationToken.Email,
			&verificationToken.CreatedAt,
			&verificationToken.ExpiresAt,
			&verificationToken.VerifiedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan email verification token: %w", scanErr)
		}

		if utils.CheckPassword(token, verificationToken.Token) {
			return verificationToken, nil
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate email verification tokens: %w", err)
	}

	return nil, fmt.Errorf("token not found")
}

// DeleteEmailVerificationToken deletes an email verification token
func (r *tokenRepository) DeleteEmailVerificationToken(ctx context.Context, token string) error {
	query := `DELETE FROM email_verification_tokens WHERE token = $1`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete email verification token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

// MarkEmailVerificationTokenUsed marks an email verification token as used by ID
func (r *tokenRepository) MarkEmailVerificationTokenUsed(ctx context.Context, tokenID int64) error {
	query := `UPDATE email_verification_tokens SET verified_at = $1 WHERE id = $2`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark email verification token as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

// CreatePasswordResetToken creates a password reset token
func (r *tokenRepository) CreatePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (
			user_id, token, created_at, expires_at
		) VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	now := time.Now()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}

	err := r.db.QueryRowContext(ctx, query,
		token.UserID,
		token.Token,
		token.CreatedAt,
		token.ExpiresAt,
	).Scan(&token.ID)

	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	return nil
}

// GetPasswordResetToken retrieves a password reset token
func (r *tokenRepository) GetPasswordResetToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, created_at, expires_at, used_at
		FROM password_reset_tokens
		WHERE token = $1
	`

	resetToken := &models.PasswordResetToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&resetToken.ID,
		&resetToken.UserID,
		&resetToken.Token,
		&resetToken.CreatedAt,
		&resetToken.ExpiresAt,
		&resetToken.UsedAt,
	)

	if err == sql.ErrNoRows {
		// Backward-compatible fallback:
		// password reset tokens are stored hashed, so plaintext lookup requires
		// scanning valid rows and verifying with bcrypt comparison.
		return r.getPasswordResetTokenByHash(ctx, token)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}

	return resetToken, nil
}

func (r *tokenRepository) getPasswordResetTokenByHash(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, created_at, expires_at, used_at
		FROM password_reset_tokens
		WHERE used_at IS NULL
		  AND expires_at > NOW()
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query password reset tokens: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		resetToken := &models.PasswordResetToken{}
		if scanErr := rows.Scan(
			&resetToken.ID,
			&resetToken.UserID,
			&resetToken.Token,
			&resetToken.CreatedAt,
			&resetToken.ExpiresAt,
			&resetToken.UsedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan password reset token: %w", scanErr)
		}

		if utils.CheckPassword(token, resetToken.Token) {
			return resetToken, nil
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate password reset tokens: %w", err)
	}

	return nil, fmt.Errorf("token not found")
}

// DeletePasswordResetToken deletes a password reset token
func (r *tokenRepository) DeletePasswordResetToken(ctx context.Context, token string) error {
	query := `DELETE FROM password_reset_tokens WHERE token = $1`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete password reset token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

// DeleteExpiredTokens deletes all expired tokens
func (r *tokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	// Delete expired email verification tokens
	query1 := `DELETE FROM email_verification_tokens WHERE expires_at < NOW() AND verified_at IS NULL`
	_, err := r.db.ExecContext(ctx, query1)
	if err != nil {
		return fmt.Errorf("failed to delete expired email verification tokens: %w", err)
	}

	// Delete expired password reset tokens
	query2 := `DELETE FROM password_reset_tokens WHERE expires_at < NOW() AND used_at IS NULL`
	_, err = r.db.ExecContext(ctx, query2)
	if err != nil {
		return fmt.Errorf("failed to delete expired password reset tokens: %w", err)
	}

	return nil
}
