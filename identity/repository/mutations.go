package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/forgego/forge/db"
)

var ErrConcurrentChange = errors.New("identity record changed concurrently")

// execOne is shared by all conditional identity writes. Zero affected rows are
// treated as a conflict, never as successful token consumption.
func execOne(ctx context.Context, exec func(context.Context, string, ...any) (sql.Result, error), query string, args ...any) error {
	result, err := exec(ctx, query, args...)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrConcurrentChange
	}
	return nil
}

func (r *userRepository) SetLastLogin(ctx context.Context, userID int64, at time.Time) error {
	return execOne(ctx, r.db.ExecContext, `UPDATE users SET last_login = $1 WHERE id = $2 AND deleted_at IS NULL`, at.UTC(), userID)
}

var editableUserFields = map[string]bool{
	"email": true, "email_verified": true, "email_verified_at": true,
	"first_name": true, "last_name": true, "is_active": true,
	"is_staff": true, "is_locked": true, "locked_at": true,
	"locked_reason": true, "is_superuser": true,
}

func (r *userRepository) UpdateFields(ctx context.Context, userID int64, fields map[string]any) error {
	return r.updateFields(ctx, userID, "", fields)
}

func (r *userRepository) UpdateEmailIfCurrent(ctx context.Context, userID int64, oldEmail string, fields map[string]any) error {
	if _, ok := fields["email"]; !ok {
		return fmt.Errorf("email update requires email field")
	}
	return r.updateFields(ctx, userID, normalizeEmail(oldEmail), fields)
}

func (r *userRepository) updateFields(ctx context.Context, userID int64, oldEmail string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		if !editableUserFields[key] {
			return fmt.Errorf("invalid user field %q", key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	sets := make([]string, 0, len(keys)+1)
	args := make([]any, 0, len(keys)+2)
	for _, key := range keys {
		sets = append(sets, fmt.Sprintf("%s = $%d", key, len(args)+1))
		args = append(args, fields[key])
	}
	sets = append(sets, fmt.Sprintf("updated_at = $%d", len(args)+1))
	args = append(args, time.Now().UTC(), userID)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d AND deleted_at IS NULL", strings.Join(sets, ", "), len(args))
	if oldEmail != "" {
		args = append(args, oldEmail)
		query += fmt.Sprintf(" AND lower(trim(email)) = $%d", len(args))
		if verified, ok := fields["email_verified"].(bool); ok && verified {
			query += " AND email_verified = FALSE"
		}
	}
	return execOne(ctx, r.db.ExecContext, query, args...)
}

func (r *userRepository) ChangePasswordIfCurrent(ctx context.Context, userID int64, oldHash, newHash string, at time.Time) error {
	return execOne(ctx, r.db.ExecContext, `UPDATE users SET password = $1, password_changed_at = $2, must_change_password = FALSE, updated_at = $2 WHERE id = $3 AND password = $4 AND deleted_at IS NULL`, newHash, at.UTC(), userID, oldHash)
}

func (r *userRepository) withTokenMutation(ctx context.Context, consume func(*db.Tx) error, update func(*db.Tx) error) error {
	return r.db.WithTx(ctx, func(tx *db.Tx) error {
		if err := consume(tx); err != nil {
			return err
		}
		return update(tx)
	})
}

func (r *userRepository) CommitPasswordReset(ctx context.Context, tokenID, userID int64, newHash string, at time.Time) error {
	return r.withTokenMutation(ctx, func(tx *db.Tx) error {
		return execOne(ctx, tx.ExecContext, `DELETE FROM password_reset_tokens WHERE id = $1 AND user_id = $2 AND used_at IS NULL AND expires_at > $3`, tokenID, userID, at.UTC())
	}, func(tx *db.Tx) error {
		return execOne(ctx, tx.ExecContext, `UPDATE users SET password = $1, password_changed_at = $2, must_change_password = FALSE, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`, newHash, at.UTC(), userID)
	})
}

func (r *userRepository) CommitEmailVerification(ctx context.Context, tokenID, userID int64, email string, at time.Time) error {
	return r.withTokenMutation(ctx, func(tx *db.Tx) error {
		return execOne(ctx, tx.ExecContext, `UPDATE email_verification_tokens SET verified_at = $1 WHERE id = $2 AND user_id = $3 AND lower(trim(email)) = $4 AND verified_at IS NULL AND expires_at > $1`, at.UTC(), tokenID, userID, normalizeEmail(email))
	}, func(tx *db.Tx) error {
		return execOne(ctx, tx.ExecContext, `UPDATE users SET email_verified = TRUE, email_verified_at = $1, updated_at = $1 WHERE id = $2 AND lower(trim(email)) = $3 AND email_verified = FALSE AND deleted_at IS NULL`, at.UTC(), userID, normalizeEmail(email))
	})
}
