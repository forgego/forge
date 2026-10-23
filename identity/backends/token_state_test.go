package backends

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
)

type stateSessionRepository struct{ repository.SessionRepository }

func (stateSessionRepository) GetByKey(context.Context, string) (*models.UserSession, error) {
	return &models.UserSession{UserID: 7, SessionKey: "session", ExpiresAt: ptrTime(time.Now().Add(time.Hour))}, nil
}

type stateUserRepository struct {
	repository.UserRepository
	user *models.User
}

func (r stateUserRepository) GetByID(context.Context, int64) (*models.User, error) {
	return r.user, nil
}

func ptrTime(t time.Time) *time.Time { return &t }

func TestTokenBackendRejectsUnavailableUsers(t *testing.T) {
	for _, state := range []struct {
		name string
		user *models.User
		want error
	}{
		{name: "inactive", user: &models.User{ID: 7, IsActive: false}, want: models.ErrUserInactive},
		{name: "locked", user: &models.User{ID: 7, IsActive: true, IsLocked: true}, want: models.ErrUserLocked},
	} {
		t.Run(state.name, func(t *testing.T) {
			backend := NewTokenBackend(stateUserRepository{user: state.user}, stateSessionRepository{})
			_, err := backend.GetUser(context.Background(), "session")
			if !errors.Is(err, state.want) {
				t.Fatalf("GetUser() error = %v, want %v", err, state.want)
			}
		})
	}
}
