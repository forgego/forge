package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/models"
)

// sessionRepository implements SessionRepository interface
type sessionRepository struct {
	db *db.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(database *db.DB) SessionRepository {
	return &sessionRepository{db: database}
}

// Create creates a new session
func (r *sessionRepository) Create(ctx context.Context, session *models.UserSession) error {
	query := `
		INSERT INTO user_sessions (
			user_id, session_key, ip_address, user_agent,
			last_activity, created_at, expires_at, is_remember_me
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now()
	if session.LastActivity.IsZero() {
		session.LastActivity = now
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}

	err := r.db.QueryRowContext(ctx, query,
		session.UserID,
		session.SessionKey,
		session.IPAddress,
		session.UserAgent,
		session.LastActivity,
		session.CreatedAt,
		session.ExpiresAt,
		session.IsRememberMe,
	).Scan(&session.ID)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByKey retrieves a session by session key
func (r *sessionRepository) GetByKey(ctx context.Context, key string) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, session_key, ip_address, user_agent,
			last_activity, created_at, expires_at, is_remember_me
		FROM user_sessions
		WHERE session_key = $1
	`

	session := &models.UserSession{}
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&session.ID,
		&session.UserID,
		&session.SessionKey,
		&session.IPAddress,
		&session.UserAgent,
		&session.LastActivity,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.IsRememberMe,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *sessionRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.UserSession, error) {
	query := `
		SELECT id, user_id, session_key, ip_address, user_agent,
			last_activity, created_at, expires_at, is_remember_me
		FROM user_sessions
		WHERE user_id = $1
		ORDER BY last_activity DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.UserSession
	for rows.Next() {
		session := &models.UserSession{}
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.SessionKey,
			&session.IPAddress,
			&session.UserAgent,
			&session.LastActivity,
			&session.CreatedAt,
			&session.ExpiresAt,
			&session.IsRememberMe,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate sessions: %w", err)
	}

	return sessions, nil
}

// Update updates a session
func (r *sessionRepository) Update(ctx context.Context, session *models.UserSession) error {
	query := `
		UPDATE user_sessions SET
			last_activity = $2, expires_at = $3
		WHERE session_key = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		session.SessionKey,
		session.LastActivity,
		session.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// Delete deletes a session by session key
func (r *sessionRepository) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM user_sessions WHERE session_key = $1`

	result, err := r.db.ExecContext(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// DeleteByUserID deletes all sessions for a user
func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM user_sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired sessions
func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM user_sessions WHERE expires_at IS NOT NULL AND expires_at < NOW()`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}
