package stores

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/alexedwards/scs/v2"
)

var _ scs.CtxStore = (*SessionStore)(nil)

// SessionStore keeps scs sessions in the forge_sessions table. It
// implements scs.Store and scs.CtxStore; assign it to
// scs.SessionManager.Store. server.NewServer does that when server.stores
// is database.
type SessionStore struct {
	stores *Stores
}

// Sessions returns the session store.
func (s *Stores) Sessions() *SessionStore { return &SessionStore{stores: s} }

// Find implements scs.Store.
func (st *SessionStore) Find(token string) ([]byte, bool, error) {
	return st.FindCtx(context.Background(), token)
}

// Commit implements scs.Store.
func (st *SessionStore) Commit(token string, b []byte, expiry time.Time) error {
	return st.CommitCtx(context.Background(), token, b, expiry)
}

// Delete implements scs.Store.
func (st *SessionStore) Delete(token string) error {
	return st.DeleteCtx(context.Background(), token)
}

// FindCtx implements scs.CtxStore. An expired session is not found.
func (st *SessionStore) FindCtx(ctx context.Context, token string) ([]byte, bool, error) {
	var data []byte
	err := st.stores.db.QueryRowContext(ctx,
		"SELECT data FROM forge_sessions WHERE token = $1 AND expires_at > $2",
		token, unixMillis(st.stores.Now()),
	).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// CommitCtx implements scs.CtxStore: it inserts or replaces the session.
func (st *SessionStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	if b == nil {
		b = []byte{}
	}
	_, err := st.stores.db.ExecContext(ctx,
		`INSERT INTO forge_sessions (token, data, expires_at) VALUES ($1, $2, $3)
		ON CONFLICT (token) DO UPDATE SET data = excluded.data, expires_at = excluded.expires_at`,
		token, b, unixMillis(expiry),
	)
	if err != nil {
		return err
	}
	st.stores.cleanupExpired(ctx, cleanupSessions)
	return nil
}

// DeleteCtx implements scs.CtxStore.
func (st *SessionStore) DeleteCtx(ctx context.Context, token string) error {
	_, err := st.stores.db.ExecContext(ctx, "DELETE FROM forge_sessions WHERE token = $1", token)
	return err
}
