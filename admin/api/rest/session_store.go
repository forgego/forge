package rest

import (
	"context"
	"errors"
	"strings"
	"time"
)

// TokenStore keeps the bearer tokens the admin API issues at login.
// stores.AdminTokens keeps them in the database so they survive a restart
// and work on every instance.
type TokenStore interface {
	// IssueToken returns a new random token for username, valid for ttl.
	IssueToken(ctx context.Context, username string, ttl time.Duration) (string, error)
	// ValidateToken returns the username of an unexpired, unrevoked token.
	ValidateToken(ctx context.Context, token string) (username string, ok bool, err error)
	// RevokeToken invalidates a token. Unknown tokens are not an error.
	RevokeToken(ctx context.Context, token string) error
}

func bearerToken(authorizationHeader string) (string, error) {
	header := strings.TrimSpace(authorizationHeader)
	if header == "" {
		return "", errors.New("missing authorization header")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("invalid authorization scheme")
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", errors.New("missing bearer token")
	}
	return token, nil
}
