package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/stretchr/testify/require"
)

type staleSessionRepository struct {
	repository.SessionRepository
	fullUpdates int
	touches     int
}

func (r *staleSessionRepository) GetByKey(context.Context, string) (*models.UserSession, error) {
	expires := time.Now().Add(time.Hour)
	return &models.UserSession{SessionKey: "key", ExpiresAt: &expires}, nil
}
func (r *staleSessionRepository) Update(context.Context, *models.UserSession) error {
	r.fullUpdates++
	return nil
}
func (r *staleSessionRepository) TouchLastActivity(context.Context, string, time.Time) error {
	r.touches++
	return nil
}

type expiredSessionRepository struct{ repository.SessionRepository }

func (r *expiredSessionRepository) GetByKey(context.Context, string) (*models.UserSession, error) {
	expires := time.Now().Add(-time.Hour)
	return &models.UserSession{SessionKey: "key", ExpiresAt: &expires}, nil
}
func (r *expiredSessionRepository) Delete(context.Context, string) error {
	return errors.New("delete failed")
}

func TestRefreshSessionTouchesActivityWithoutWritingStaleExpiry(t *testing.T) {
	r := &staleSessionRepository{}
	s := NewAuthService(nil, r, nil)
	_, err := s.RefreshSession(context.Background(), "key")
	require.NoError(t, err)
	require.Equal(t, 1, r.touches)
	require.Zero(t, r.fullUpdates)
}

func TestExpiredSessionReportsDeletionFailure(t *testing.T) {
	s := NewAuthService(nil, &expiredSessionRepository{}, nil)
	_, err := s.GetSession(context.Background(), "key")
	require.ErrorContains(t, err, "delete failed")
}
