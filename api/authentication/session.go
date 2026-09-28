package authentication

import (
	"net/http"
)

// SessionAuthentication authenticates requests using session cookies
type SessionAuthentication struct {
	// SessionManager is the session manager interface
	// Should have GetUser(r *http.Request) (interface{}, error) method
	SessionManager interface {
		GetUser(r *http.Request) (interface{}, error)
	}
}

// NewSessionAuthentication creates a new session authentication instance
func NewSessionAuthentication(sessionManager interface {
	GetUser(r *http.Request) (interface{}, error)
}) *SessionAuthentication {
	return &SessionAuthentication{
		SessionManager: sessionManager,
	}
}

// Authenticate attempts to authenticate using session
func (a *SessionAuthentication) Authenticate(r *http.Request) (*AuthResult, error) {
	if a.SessionManager == nil {
		return nil, nil
	}

	user, err := a.SessionManager.GetUser(r)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return NewAuthResult(user, nil), nil
}

// AuthenticateHeader returns empty string (sessions don't use WWW-Authenticate)
func (a *SessionAuthentication) AuthenticateHeader(r *http.Request) string {
	return ""
}
