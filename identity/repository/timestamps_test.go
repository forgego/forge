package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func timestampTestUser(t *testing.T) (*db.DB, *models.User) {
	t.Helper()
	database := setupTestDB(t)
	t.Cleanup(func() { require.NoError(t, database.Close()) })
	// Keep a non-UTC database session too: timestamp comparisons must not
	// depend on either the caller's offset or PostgreSQL's session timezone.
	database.SetMaxOpenConns(1)
	_, err := database.Exec("SET TIME ZONE 'Pacific/Honolulu'")
	require.NoError(t, err)
	user := &models.User{Username: "timestamp-user", Email: "timestamp@example.com", Password: "test-hash", IsActive: true}
	require.NoError(t, NewUserRepository(database).Create(context.Background(), user))
	return database, user
}

func TestTokenExpiryPreservesInstantAcrossTimeZones(t *testing.T) {
	database, user := timestampTestUser(t)
	repo := NewTokenRepository(database)
	ctx := context.Background()
	for _, offset := range []int{12600, -25200} {
		for _, expired := range []bool{false, true} {
			for _, hashed := range []bool{false, true} {
				t.Run(fmt.Sprintf("offset=%d/expired=%t/hashed=%t", offset, expired, hashed), func(t *testing.T) {
					delta := time.Hour
					if expired {
						delta = -delta
					}
					expires := time.Now().Add(delta).In(time.FixedZone("caller", offset)).Truncate(time.Microsecond)
					for _, kind := range []string{"email", "password"} {
						raw := fmt.Sprintf("%d-%t-%t-%s", offset, expired, hashed, kind)
						stored := raw
						if hashed {
							var err error
							stored, err = utils.HashPasswordWithCost(raw, 4)
							require.NoError(t, err)
						}
						var gotExpiry time.Time
						var gotExpired bool
						var err error
						if kind == "email" {
							token := &models.EmailVerificationToken{UserID: user.ID, Email: user.Email, Token: stored, ExpiresAt: expires}
							require.NoError(t, repo.CreateEmailVerificationToken(ctx, token))
							got, getErr := repo.GetEmailVerificationToken(ctx, raw)
							err = getErr
							if got != nil {
								gotExpiry, gotExpired = got.ExpiresAt, got.IsExpired()
							}
						} else {
							token := &models.PasswordResetToken{UserID: user.ID, Token: stored, ExpiresAt: expires}
							require.NoError(t, repo.CreatePasswordResetToken(ctx, token))
							got, getErr := repo.GetPasswordResetToken(ctx, raw)
							err = getErr
							if got != nil {
								gotExpiry, gotExpired = got.ExpiresAt, got.IsExpired()
							}
						}
						if hashed && expired {
							assert.Error(t, err, "%s expired hash must not be selected", kind)
							continue
						}
						require.NoError(t, err, kind)
						assert.True(t, expires.Equal(gotExpiry), "%s: want %s, got %s", kind, expires, gotExpiry)
						assert.Equal(t, expired, gotExpired, kind)
					}
				})
			}
		}
	}
	// Cleanup uses the same UTC convention as hashed-token selection.
	require.NoError(t, repo.DeleteExpiredTokens(ctx))
	for _, table := range []string{"email_verification_tokens", "password_reset_tokens"} {
		var count int
		require.NoError(t, database.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count))
		assert.Equal(t, 4, count, table)
	}
}

func TestSessionExpiryPreservesInstantAcrossTimeZones(t *testing.T) {
	database, user := timestampTestUser(t)
	repo := NewSessionRepository(database)
	ctx := context.Background()
	for _, offset := range []int{12600, -25200} {
		t.Run(fmt.Sprint(offset), func(t *testing.T) {
			expires := time.Now().Add(-time.Hour).In(time.FixedZone("caller", offset)).Truncate(time.Microsecond)
			session := &models.UserSession{UserID: user.ID, SessionKey: t.Name(), ExpiresAt: &expires}
			require.NoError(t, repo.Create(ctx, session))
			got, err := repo.GetByKey(ctx, session.SessionKey)
			require.NoError(t, err)
			require.NotNil(t, got.ExpiresAt)
			assert.True(t, expires.Equal(*got.ExpiresAt))
			assert.True(t, got.IsExpired())
			updated := expires.Add(2 * time.Hour)
			session.ExpiresAt = &updated
			require.NoError(t, repo.Update(ctx, session))
			got, err = repo.GetByKey(ctx, session.SessionKey)
			require.NoError(t, err)
			assert.True(t, updated.Equal(*got.ExpiresAt))
			assert.False(t, got.IsExpired())
		})
	}
	old := time.Now().UTC().Add(-time.Hour)
	require.NoError(t, repo.Create(ctx, &models.UserSession{UserID: user.ID, SessionKey: "expired", ExpiresAt: &old}))
	require.NoError(t, repo.Create(ctx, &models.UserSession{UserID: user.ID, SessionKey: "no-expiry"}))
	require.NoError(t, repo.DeleteExpired(ctx))
	_, err := repo.GetByKey(ctx, "expired")
	assert.Error(t, err)
	got, err := repo.GetByKey(ctx, "no-expiry")
	require.NoError(t, err)
	assert.Nil(t, got.ExpiresAt)
}

func TestUserSecurityTimestampsPreserveInstantAcrossTimeZones(t *testing.T) {
	database, user := timestampTestUser(t)
	repo := NewUserRepository(database)
	instant := time.Now().In(time.FixedZone("caller", 12600)).Truncate(time.Microsecond)
	user.PasswordChangedAt, user.PasswordExpiresAt = &instant, &instant
	user.LockedAt, user.LastFailedLoginAt = &instant, &instant
	user.EmailVerifiedAt, user.LastLogin = &instant, &instant
	require.NoError(t, repo.Update(context.Background(), user))
	got, err := repo.GetByID(context.Background(), user.ID)
	require.NoError(t, err)
	for _, actual := range []*time.Time{got.PasswordChangedAt, got.PasswordExpiresAt, got.LockedAt, got.LastFailedLoginAt, got.EmailVerifiedAt, got.LastLogin} {
		require.NotNil(t, actual)
		assert.True(t, instant.Equal(*actual), "want %s, got %s", instant, actual)
	}
}
