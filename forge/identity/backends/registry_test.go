package backends

import (
	"context"
	"errors"
	"testing"

	"github.com/forgego/forge/identity/repository"
)

func TestBackendRegistryGetUserReturnsRepositoryNotFound(t *testing.T) {
	_, err := NewBackendRegistry().GetUser(context.Background(), "missing")
	if !errors.Is(err, repository.ErrUserNotFound) {
		t.Fatalf("GetUser() error = %v, want repository.ErrUserNotFound", err)
	}
}
