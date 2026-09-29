package stores

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

// AdminTokens keeps admin API bearer tokens in forge_admin_tokens. Only the
// SHA-256 hash of a token is stored, so a database dump does not contain
// usable tokens. It satisfies rest.TokenStore.
type AdminTokens struct {
	stores *Stores
}

// AdminTokens returns the admin token store.
func (s *Stores) AdminTokens() *AdminTokens { return &AdminTokens{stores: s} }

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// IssueToken creates a random token for username that expires after ttl.
func (t *AdminTokens) IssueToken(ctx context.Context, username string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		return "", errors.New("invalid token ttl")
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	now := t.stores.Now()
	_, err := t.stores.db.ExecContext(ctx,
		"INSERT INTO forge_admin_tokens (token_hash, username, created_at, expires_at) VALUES ($1, $2, $3, $4)",
		hashToken(token), username, now.UTC(), unixMillis(now.Add(ttl)),
	)
	if err != nil {
		return "", err
	}
	t.stores.cleanupExpired(ctx, cleanupTokens)
	return token, nil
}

// ValidateToken returns the username of an unexpired, unrevoked token.
func (t *AdminTokens) ValidateToken(ctx context.Context, token string) (string, bool, error) {
	var username string
	err := t.stores.db.QueryRowContext(ctx,
		"SELECT username FROM forge_admin_tokens WHERE token_hash = $1 AND expires_at > $2",
		hashToken(token), unixMillis(t.stores.Now()),
	).Scan(&username)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return username, true, nil
}

// RevokeToken deletes a token. Revoking an unknown token is not an error.
func (t *AdminTokens) RevokeToken(ctx context.Context, token string) error {
	_, err := t.stores.db.ExecContext(ctx, "DELETE FROM forge_admin_tokens WHERE token_hash = $1", hashToken(token))
	return err
}
