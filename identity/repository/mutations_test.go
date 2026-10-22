package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func mutationDB(t *testing.T, userCheck string) *userRepository {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	sqlDB.SetMaxOpenConns(1)
	statements := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT, password TEXT, is_locked BOOLEAN DEFAULT FALSE, email_verified BOOLEAN DEFAULT FALSE, email_verified_at DATETIME, first_name TEXT, last_name TEXT, is_active BOOLEAN, is_staff BOOLEAN, locked_at DATETIME, locked_reason TEXT, is_superuser BOOLEAN, last_login DATETIME, password_changed_at DATETIME, must_change_password BOOLEAN, updated_at DATETIME, deleted_at DATETIME` + userCheck + `)`,
		`CREATE TABLE password_reset_tokens (id INTEGER PRIMARY KEY, user_id INTEGER, used_at DATETIME, expires_at DATETIME)`,
		`CREATE TABLE email_verification_tokens (id INTEGER PRIMARY KEY, user_id INTEGER, email TEXT, verified_at DATETIME, expires_at DATETIME)`,
		`INSERT INTO users (id, email, password, is_locked, email_verified) VALUES (1, 'a@example.com', 'old', TRUE, FALSE)`,
	}
	for _, stmt := range statements {
		_, err = sqlDB.Exec(stmt)
		require.NoError(t, err)
	}
	return &userRepository{db: &db.DB{DB: sqlDB, Driver: "sqlite3"}}
}

func TestScopedIdentityMutationsPreserveConcurrentFields(t *testing.T) {
	r := mutationDB(t, "")
	ctx := context.Background()
	now := time.Now()
	require.NoError(t, r.SetLastLogin(ctx, 1, now))
	require.NoError(t, r.UpdateFields(ctx, 1, map[string]any{"first_name": "Alice"}))
	var password, first string
	var locked bool
	require.NoError(t, r.db.QueryRowContext(ctx, `SELECT password, is_locked, first_name FROM users WHERE id = 1`).Scan(&password, &locked, &first))
	require.Equal(t, "old", password)
	require.True(t, locked)
	require.Equal(t, "Alice", first)
	require.ErrorIs(t, r.ChangePasswordIfCurrent(ctx, 1, "stale", "new", now), ErrConcurrentChange)
	require.NoError(t, r.ChangePasswordIfCurrent(ctx, 1, "old", "new", now))
	require.ErrorIs(t, r.ChangePasswordIfCurrent(ctx, 1, "old", "reverted", now), ErrConcurrentChange)
}

func TestEmailChangeRejectsStaleAddressAndPreservesVerification(t *testing.T) {
	r := mutationDB(t, "")
	ctx := context.Background()
	_, err := r.db.ExecContext(ctx, `UPDATE users SET email = 'b@example.com', email_verified = TRUE WHERE id = 1`)
	require.NoError(t, err)
	require.ErrorIs(t, r.UpdateEmailIfCurrent(ctx, 1, "a@example.com", map[string]any{
		"email": "c@example.com", "email_verified": false, "email_verified_at": nil,
	}), ErrConcurrentChange)
	var email string
	var verified bool
	require.NoError(t, r.db.QueryRowContext(ctx, `SELECT email, email_verified FROM users WHERE id = 1`).Scan(&email, &verified))
	require.Equal(t, "b@example.com", email)
	require.True(t, verified)
}

func TestUnchangedEmailRejectsStaleAddressAndPreservesVerification(t *testing.T) {
	r := mutationDB(t, "")
	ctx := context.Background()
	_, err := r.db.ExecContext(ctx, `UPDATE users SET email = 'b@example.com', email_verified = TRUE WHERE id = 1`)
	require.NoError(t, err)
	require.ErrorIs(t, r.UpdateEmailIfCurrent(ctx, 1, "a@example.com", map[string]any{"email": "a@example.com"}), ErrConcurrentChange)
	var email string
	var verified bool
	require.NoError(t, r.db.QueryRowContext(ctx, `SELECT email, email_verified FROM users WHERE id = 1`).Scan(&email, &verified))
	require.Equal(t, "b@example.com", email)
	require.True(t, verified)
}

func TestResetTokenAndPasswordRollBackTogether(t *testing.T) {
	r := mutationDB(t, ", CHECK (password <> 'blocked')")
	ctx := context.Background()
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `INSERT INTO password_reset_tokens (id, user_id, expires_at) VALUES (2, 1, $1)`, now.Add(time.Hour))
	require.NoError(t, err)
	require.Error(t, r.CommitPasswordReset(ctx, 2, 1, "blocked", now))
	var count int
	require.NoError(t, r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM password_reset_tokens WHERE id = 2`).Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, r.CommitPasswordReset(ctx, 2, 1, "new", now))
	require.ErrorIs(t, r.CommitPasswordReset(ctx, 2, 1, "replayed", now), ErrConcurrentChange)
	var password string
	require.NoError(t, r.db.QueryRowContext(ctx, `SELECT password FROM users WHERE id = 1`).Scan(&password))
	require.Equal(t, "new", password)
}

func TestVerificationTokenAndUserRollBackTogether(t *testing.T) {
	r := mutationDB(t, ", CHECK (email_verified = FALSE)")
	ctx := context.Background()
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `INSERT INTO email_verification_tokens (id, user_id, email, expires_at) VALUES (2, 1, 'a@example.com', $1)`, now.Add(time.Hour))
	require.NoError(t, err)
	require.Error(t, r.CommitEmailVerification(ctx, 2, 1, "a@example.com", now))
	var used sql.NullTime
	require.NoError(t, r.db.QueryRowContext(ctx, `SELECT verified_at FROM email_verification_tokens WHERE id = 2`).Scan(&used))
	require.False(t, used.Valid)
	_, err = r.db.ExecContext(ctx, `UPDATE users SET email = 'b@example.com' WHERE id = 1`)
	require.NoError(t, err)
	require.ErrorIs(t, r.CommitEmailVerification(ctx, 2, 1, "a@example.com", now), ErrConcurrentChange)
	require.NoError(t, r.db.QueryRowContext(ctx, `SELECT verified_at FROM email_verification_tokens WHERE id = 2`).Scan(&used))
	require.False(t, used.Valid)
}

func TestEmailVerificationCannotBeReplayed(t *testing.T) {
	r := mutationDB(t, "")
	ctx := context.Background()
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `INSERT INTO email_verification_tokens (id, user_id, email, expires_at) VALUES (2, 1, 'a@example.com', $1)`, now.Add(time.Hour))
	require.NoError(t, err)
	require.NoError(t, r.CommitEmailVerification(ctx, 2, 1, "a@example.com", now))
	require.ErrorIs(t, r.CommitEmailVerification(ctx, 2, 1, "a@example.com", now), ErrConcurrentChange)
	var verified bool
	require.NoError(t, r.db.QueryRowContext(ctx, `SELECT email_verified FROM users WHERE id = 1`).Scan(&verified))
	require.True(t, verified)
}

func TestStandaloneTokenMarkRejectsReplay(t *testing.T) {
	r := mutationDB(t, "")
	ctx := context.Background()
	_, err := r.db.ExecContext(ctx, `INSERT INTO email_verification_tokens (id, user_id, email, expires_at) VALUES (2, 1, 'a@example.com', $1)`, time.Now().Add(time.Hour))
	require.NoError(t, err)
	tokens := NewTokenRepository(r.db)
	require.NoError(t, tokens.MarkEmailVerificationTokenUsed(ctx, 2))
	require.Error(t, tokens.MarkEmailVerificationTokenUsed(ctx, 2))
}
