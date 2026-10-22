package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpiredTokenCleanupRollsBackBothTables(t *testing.T) {
	r := mutationDB(t, "")
	ctx := context.Background()
	expired := "2020-01-01 00:00:00"
	_, err := r.db.ExecContext(ctx, `INSERT INTO email_verification_tokens (id, user_id, expires_at) VALUES (2, 1, $1)`, expired)
	require.NoError(t, err)
	_, err = r.db.ExecContext(ctx, `INSERT INTO password_reset_tokens (id, user_id, expires_at) VALUES (2, 1, $1)`, expired)
	require.NoError(t, err)
	_, err = r.db.ExecContext(ctx, `CREATE TRIGGER fail_reset_cleanup BEFORE DELETE ON password_reset_tokens BEGIN SELECT RAISE(ABORT, 'reset cleanup failed'); END`)
	require.NoError(t, err)
	require.ErrorContains(t, NewTokenRepository(r.db).DeleteExpiredTokens(ctx), "reset cleanup failed")
	var count int
	require.NoError(t, r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM email_verification_tokens WHERE id = 2`).Scan(&count))
	require.Equal(t, 1, count)
}
