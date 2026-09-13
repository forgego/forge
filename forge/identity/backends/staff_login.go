package backends

import (
	"context"
	"errors"
	"fmt"
)

// ErrAdminLoginDenied indicates that the user credentials or permissions do not allow admin access.
var ErrAdminLoginDenied = errors.New("invalid admin credentials")

// AdminLoginDeniedError represents a rejected admin login attempt.
type AdminLoginDeniedError struct {
	cause error
}

func (e *AdminLoginDeniedError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%v: %v", ErrAdminLoginDenied, e.cause)
	}
	return ErrAdminLoginDenied.Error()
}

func (e *AdminLoginDeniedError) Is(target error) bool {
	return target == ErrAdminLoginDenied
}

func (e *AdminLoginDeniedError) Unwrap() error {
	return e.cause
}

// InvalidLogin signals to the admin router that this is a rejected credential/authorization error.
func (e *AdminLoginDeniedError) InvalidLogin() bool {
	return true
}

// StaffLoginAuthenticator adapts an AuthenticationBackend to authenticate admin users.
type StaffLoginAuthenticator struct {
	backend AuthenticationBackend
}

// NewStaffLoginAuthenticator constructs a new StaffLoginAuthenticator.
func NewStaffLoginAuthenticator(backend AuthenticationBackend) *StaffLoginAuthenticator {
	return &StaffLoginAuthenticator{backend: backend}
}

// AuthenticateAdmin authenticates the given credentials and ensures the user has staff or superuser privileges.
func (a *StaffLoginAuthenticator) AuthenticateAdmin(ctx context.Context, username, password string) (string, error) {
	if a.backend == nil {
		return "", &AdminLoginDeniedError{}
	}
	user, err := a.backend.Authenticate(ctx, map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrUserInactive) || errors.Is(err, ErrUserLocked) {
			return "", &AdminLoginDeniedError{cause: err}
		}
		return "", fmt.Errorf("staff login authenticate: %w", err)
	}
	if user == nil || !user.IsActive || user.IsLocked || (!user.IsStaff && !user.IsSuperuser) {
		return "", &AdminLoginDeniedError{}
	}
	return user.Username, nil
}
