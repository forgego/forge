package backends

import (
	"context"
	"fmt"

	"github.com/forgego/forge/identity/models"
)

// backendRegistry implements BackendRegistry interface
type backendRegistry struct {
	backends []AuthenticationBackend
	byName   map[string]AuthenticationBackend
}

// NewBackendRegistry creates a new backend registry
func NewBackendRegistry() BackendRegistry {
	return &backendRegistry{
		backends: make([]AuthenticationBackend, 0),
		byName:   make(map[string]AuthenticationBackend),
	}
}

// Register registers a backend
func (r *backendRegistry) Register(backend AuthenticationBackend) {
	name := backend.Name()
	if _, exists := r.byName[name]; !exists {
		r.backends = append(r.backends, backend)
		r.byName[name] = backend
	}
}

// Get retrieves a backend by name
func (r *backendRegistry) Get(name string) (AuthenticationBackend, error) {
	backend, ok := r.byName[name]
	if !ok {
		return nil, fmt.Errorf("backend %s not found", name)
	}
	return backend, nil
}

// GetByCredentialType retrieves backends that support a credential type
func (r *backendRegistry) GetByCredentialType(credentialType string) []AuthenticationBackend {
	var supported []AuthenticationBackend
	for _, backend := range r.backends {
		if backend.Supports(credentialType) {
			supported = append(supported, backend)
		}
	}
	return supported
}

// Authenticate attempts authentication with all registered backends
// Returns the first successful authentication
func (r *backendRegistry) Authenticate(ctx context.Context, credentials map[string]string) (*models.User, error) {
	// Determine credential type from credentials map
	credentialType := "password" // Default
	if _, ok := credentials["token"]; ok {
		credentialType = "token"
	} else if _, ok := credentials["oauth_token"]; ok {
		credentialType = "oauth"
	}

	// Try backends that support this credential type
	supportedBackends := r.GetByCredentialType(credentialType)
	if len(supportedBackends) == 0 {
		// Fallback: try all backends
		supportedBackends = r.backends
	}

	// Try each backend
	for _, backend := range supportedBackends {
		user, err := backend.Authenticate(ctx, credentials)
		if err != nil {
			// Authentication failed, continue to next backend
			continue
		}
		if user != nil {
			// Authentication successful
			return user, nil
		}
		// Backend not applicable, try next
	}

	// No backend succeeded
	return nil, fmt.Errorf("authentication failed")
}

// GetUser attempts to get a user using all backends
func (r *backendRegistry) GetUser(ctx context.Context, identifier string) (*models.User, error) {
	// Try all backends
	for _, backend := range r.backends {
		user, err := backend.GetUser(ctx, identifier)
		if err != nil {
			continue
		}
		if user != nil {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}
