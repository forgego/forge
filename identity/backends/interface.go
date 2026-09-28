package backends

import (
	"context"

	"github.com/forgego/forge/identity/models"
)

// AuthenticationBackend defines the interface for authentication backends
// This follows the Strategy pattern - different backends implement different auth methods
type AuthenticationBackend interface {
	// Authenticate attempts to authenticate using this backend
	// credentials is a map of credential types to values (e.g., {"username": "user", "password": "pass"})
	// Returns the authenticated user if successful, nil if not applicable, error if failed
	Authenticate(ctx context.Context, credentials map[string]string) (*models.User, error)

	// GetUser retrieves a user by identifier (for token-based auth)
	// identifier could be a token, session key, etc.
	GetUser(ctx context.Context, identifier string) (*models.User, error)

	// Supports returns true if this backend can handle the given credential type
	// credentialType examples: "password", "token", "oauth", "saml"
	Supports(credentialType string) bool

	// Name returns the backend name (e.g., "password", "token", "oauth")
	Name() string
}

// BackendRegistry manages authentication backends
type BackendRegistry interface {
	// Register registers a backend
	Register(backend AuthenticationBackend)

	// Get retrieves a backend by name
	Get(name string) (AuthenticationBackend, error)

	// GetByCredentialType retrieves backends that support a credential type
	GetByCredentialType(credentialType string) []AuthenticationBackend

	// Authenticate attempts authentication with all registered backends
	// Returns the first successful authentication
	Authenticate(ctx context.Context, credentials map[string]string) (*models.User, error)

	// GetUser attempts to get a user using all backends
	GetUser(ctx context.Context, identifier string) (*models.User, error)
}
