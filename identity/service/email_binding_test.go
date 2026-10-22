package service

import (
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestChangedEmailLosesPreviousVerification(t *testing.T) {
	service, database, ctx := setupUserServiceTest(t)
	defer database.Close()
	repo := repository.NewUserRepository(database)
	now := time.Now()
	user := &models.User{Username: "proof-user", Email: "old@example.com", Password: "synthetic-hash", IsActive: true, EmailVerified: true, EmailVerifiedAt: &now}
	require.NoError(t, repo.Create(ctx, user))
	same := "  OLD@EXAMPLE.COM  "
	unchanged, err := service.UpdateUser(ctx, user.ID, &UpdateUserRequest{Email: &same})
	require.NoError(t, err)
	assert.True(t, unchanged.EmailVerified)
	changed := "new@example.com"
	_, err = service.UpdateUser(ctx, user.ID, &UpdateUserRequest{Email: &changed})
	require.NoError(t, err)
	stored, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.False(t, stored.EmailVerified)
	assert.Nil(t, stored.EmailVerifiedAt)
}
func TestVerificationTokenCannotVerifyAnotherEmail(t *testing.T) {
	service, database, ctx := setupUserServiceTest(t)
	defer database.Close()
	repo := repository.NewUserRepository(database)
	user := &models.User{Username: "address-user", Email: "old@example.com", Password: "synthetic-hash", IsActive: true}
	require.NoError(t, repo.Create(ctx, user))
	tokens := repository.NewTokenRepository(database)
	token := &models.EmailVerificationToken{UserID: user.ID, Token: "synthetic-old-address-proof", Email: user.Email, ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, tokens.CreateEmailVerificationToken(ctx, token))
	// Simulate an existing address change from another writer, so proof validation
	// itself must enforce binding even without the service's flag-clearing path.
	user.Email = "new@example.com"
	require.NoError(t, repo.Update(ctx, user))
	assert.ErrorIs(t, service.VerifyEmail(ctx, token.Token), ErrInvalidToken)
	stored, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.False(t, stored.EmailVerified)
	storedToken, err := tokens.GetEmailVerificationToken(ctx, token.Token)
	require.NoError(t, err)
	assert.False(t, storedToken.IsUsed())
}

func TestVerificationAcceptsCanonicalEquivalentAddress(t *testing.T) {
	service, database, ctx := setupUserServiceTest(t)
	defer database.Close()
	repo := repository.NewUserRepository(database)
	user := &models.User{Username: "canonical-user", Email: "user@example.com", Password: "synthetic-hash", IsActive: true}
	require.NoError(t, repo.Create(ctx, user))
	token := &models.EmailVerificationToken{UserID: user.ID, Token: "synthetic-canonical-proof", Email: " USER@EXAMPLE.COM ", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, repository.NewTokenRepository(database).CreateEmailVerificationToken(ctx, token))
	require.NoError(t, service.VerifyEmail(ctx, token.Token))
}
