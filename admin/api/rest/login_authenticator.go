package rest

import (
	"context"
	"errors"
)

// ErrInvalidLogin indicates that admin credentials are invalid or the account is not permitted.
var ErrInvalidLogin = errors.New("invalid admin credentials")

var errAdminLoginDisabled = errors.New("admin login is disabled")

// LoginAuthenticator verifies admin credentials. It returns the canonical username on success
// and ErrInvalidLogin when the credentials are wrong or the account may not use the admin.
// Implementations can also return an error satisfying interface{ InvalidLogin() bool } returning
// true to signal denied authentication without importing package rest. Any other returned error
// is treated as an infrastructure/internal failure resulting in HTTP 500.
type LoginAuthenticator interface {
	AuthenticateAdmin(ctx context.Context, username, password string) (string, error)
}

// SetLoginAuthenticator configures the custom authenticator used for admin logins.
func (r *Router) SetLoginAuthenticator(a LoginAuthenticator) {
	r.authenticator = a
}

func isInvalidLogin(err error) bool {
	if errors.Is(err, ErrInvalidLogin) {
		return true
	}
	var inv interface{ InvalidLogin() bool }
	if errors.As(err, &inv) && inv.InvalidLogin() {
		return true
	}
	return false
}
