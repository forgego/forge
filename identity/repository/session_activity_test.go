package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSessionTouchPreservesExpiry(t *testing.T) {
	r := mutationDB(t, "")
	ctx := context.Background()
	_, err := r.db.ExecContext(ctx, `CREATE TABLE user_sessions (session_key TEXT PRIMARY KEY, last_activity DATETIME, expires_at TEXT)`)
	require.NoError(t, err)
	_, err = r.db.ExecContext(ctx, `INSERT INTO user_sessions (session_key, expires_at) VALUES ('key', '2030-01-01')`)
	require.NoError(t, err)
	activity := NewSessionRepository(r.db).(SessionActivityRepository)
	require.NoError(t, activity.TouchLastActivity(ctx, "key", time.Now()))
	var expiry string
	require.NoError(t, r.db.QueryRowContext(ctx, `SELECT expires_at FROM user_sessions WHERE session_key = 'key'`).Scan(&expiry))
	require.Equal(t, "2030-01-01", expiry)
}
