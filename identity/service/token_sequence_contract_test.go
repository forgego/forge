package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/forgego/forge/identity/config"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
)

type sequencingUsers struct {
	*regressionUsers
	consumed *bool
}

func (u *sequencingUsers) CommitPasswordReset(context.Context, int64, int64, string, time.Time) error {
	if !*u.consumed {
		return errors.New("user written before token consumption")
	}
	u.passwordUpdates++
	return nil
}

type sequencingTokens struct {
	repository.TokenRepository
	reset    *models.PasswordResetToken
	consumed *bool
}

func (t *sequencingTokens) GetPasswordResetToken(context.Context, string) (*models.PasswordResetToken, error) {
	return t.reset, nil
}
func (t *sequencingTokens) DeletePasswordResetToken(context.Context, string) error {
	*t.consumed = true
	return nil
}

func TestCustomTokenStoreConsumesBeforeUserWrite(t *testing.T) {
	consumed := false
	u := &sequencingUsers{regressionUsers: &regressionUsers{user: &models.User{ID: 1, Password: "old"}}, consumed: &consumed}
	tokens := &sequencingTokens{reset: &models.PasswordResetToken{ID: 2, UserID: 1, Token: "stored", ExpiresAt: time.Now().Add(time.Hour)}, consumed: &consumed}
	s := NewPasswordService(u, tokens, config.DefaultIdentityConfig().PasswordPolicy)
	if err := s.ResetPassword(context.Background(), "raw", "StrongPassword123!"); err != nil {
		t.Fatal(err)
	}
	if !consumed || u.passwordUpdates != 1 {
		t.Fatalf("consumed=%v writes=%d", consumed, u.passwordUpdates)
	}
}
