package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/forgego/forge/identity/config"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/utils"
	"github.com/stretchr/testify/require"
)

type regressionUsers struct {
	repository.UserRepository
	user             *models.User
	fullUpdates      int
	lastLoginUpdates int
	passwordUpdates  int
	fieldUpdates     int
	atomicErr        error
}

func (r *regressionUsers) GetByID(context.Context, int64) (*models.User, error) { return r.user, nil }
func (r *regressionUsers) GetByUsername(context.Context, string) (*models.User, error) {
	return r.user, nil
}
func (r *regressionUsers) GetByEmail(context.Context, string) (*models.User, error) {
	return r.user, nil
}
func (r *regressionUsers) Update(_ context.Context, _ *models.User) error {
	r.fullUpdates++
	return nil
}
func (r *regressionUsers) Exists(context.Context, string) (bool, error) { return false, nil }
func (r *regressionUsers) SetLastLogin(_ context.Context, _ int64, _ time.Time) error {
	r.lastLoginUpdates++
	return nil
}
func (r *regressionUsers) UpdateFields(context.Context, int64, map[string]any) error {
	r.fieldUpdates++
	return nil
}
func (r *regressionUsers) ChangePasswordIfCurrent(context.Context, int64, string, string, time.Time) error {
	r.passwordUpdates++
	return nil
}
func (r *regressionUsers) CommitPasswordReset(context.Context, int64, int64, string, time.Time) error {
	return r.atomicErr
}
func (r *regressionUsers) CommitEmailVerification(context.Context, int64, int64, string, time.Time) error {
	return r.atomicErr
}

type regressionTokens struct {
	repository.TokenRepository
	reset        *models.PasswordResetToken
	verification *models.EmailVerificationToken
	deletes      int
	marks        int
}

type sharedRegressionTokens struct{ *regressionTokens }

func (*sharedRegressionTokens) SharesUserTransaction(repository.UserRepository) bool { return true }

func (*sharedRegressionTokens) CommitPasswordReset(ctx context.Context, users repository.UserMutations, tokenID, userID int64, hash string, at time.Time) error {
	return users.CommitPasswordReset(ctx, tokenID, userID, hash, at)
}

func (*sharedRegressionTokens) CommitEmailVerification(ctx context.Context, users repository.UserMutations, tokenID, userID int64, email string, at time.Time) error {
	return users.CommitEmailVerification(ctx, tokenID, userID, email, at)
}

// This wrapper inherits the transaction claim, but owns token consumption.
type wrappedRegressionTokens struct {
	repository.TokenRepository
	repository.TokenTransactionRepository
	tokens *regressionTokens
}

func (r *wrappedRegressionTokens) DeletePasswordResetToken(ctx context.Context, token string) error {
	return r.tokens.DeletePasswordResetToken(ctx, token)
}

func (r *wrappedRegressionTokens) MarkEmailVerificationTokenUsed(ctx context.Context, id int64) error {
	return r.tokens.MarkEmailVerificationTokenUsed(ctx, id)
}

type successfulRegressionTokens struct{ *regressionTokens }

func (*successfulRegressionTokens) DeletePasswordResetToken(context.Context, string) error {
	return nil
}
func (*successfulRegressionTokens) MarkEmailVerificationTokenUsed(context.Context, int64) error {
	return nil
}

type emailCASUsers struct {
	*regressionUsers
	casCalls int
}

type scopedEmailUsers struct {
	*regressionUsers
	verifiedWrites int
}

func (r *scopedEmailUsers) UpdateEmailIfCurrent(_ context.Context, _ int64, _ string, _ map[string]any) error {
	r.verifiedWrites++
	return nil
}

func (r *emailCASUsers) UpdateEmailIfCurrent(context.Context, int64, string, map[string]any) error {
	r.casCalls++
	return repository.ErrConcurrentChange
}

type legacyRegressionUsers struct {
	repository.UserRepository
	user    *models.User
	updates int
}

func (r *legacyRegressionUsers) GetByID(context.Context, int64) (*models.User, error) {
	return r.user, nil
}
func (r *legacyRegressionUsers) GetByUsername(context.Context, string) (*models.User, error) {
	return r.user, nil
}
func (r *legacyRegressionUsers) GetByEmail(context.Context, string) (*models.User, error) {
	return r.user, nil
}
func (r *legacyRegressionUsers) Update(context.Context, *models.User) error { r.updates++; return nil }

func (r *regressionTokens) GetPasswordResetToken(context.Context, string) (*models.PasswordResetToken, error) {
	return r.reset, nil
}
func (r *regressionTokens) DeletePasswordResetToken(context.Context, string) error {
	r.deletes++
	return errors.New("delete failed")
}
func (r *regressionTokens) GetEmailVerificationToken(context.Context, string) (*models.EmailVerificationToken, error) {
	return r.verification, nil
}
func (r *regressionTokens) MarkEmailVerificationTokenUsed(context.Context, int64) error {
	r.marks++
	return errors.New("mark failed")
}

func TestResetFailureDoesNotWritePassword(t *testing.T) {
	u := &regressionUsers{user: &models.User{ID: 1, Password: "old"}, atomicErr: errors.New("token delete failed")}
	tokens := &regressionTokens{reset: &models.PasswordResetToken{ID: 2, UserID: 1, Token: "stored", ExpiresAt: time.Now().Add(time.Hour)}}
	s := NewPasswordService(u, &sharedRegressionTokens{tokens}, config.DefaultIdentityConfig().PasswordPolicy)
	require.ErrorContains(t, s.ResetPassword(context.Background(), "raw", "StrongPassword123!"), "token delete failed")
	require.Equal(t, 0, u.fullUpdates)
	require.Equal(t, 0, tokens.deletes)
}

func TestVerificationFailureDoesNotWriteUser(t *testing.T) {
	u := &regressionUsers{user: &models.User{ID: 1, Email: "a@example.com"}, atomicErr: errors.New("token mark failed")}
	tokens := &regressionTokens{verification: &models.EmailVerificationToken{ID: 2, UserID: 1, Email: "a@example.com", ExpiresAt: time.Now().Add(time.Hour)}}
	s := NewUserService(u, &sharedRegressionTokens{tokens}, nil)
	require.ErrorContains(t, s.VerifyEmail(context.Background(), "raw"), "token mark failed")
	require.Equal(t, 0, u.fullUpdates)
	require.Equal(t, 0, tokens.marks)
}

func TestCustomTokenRepositoryIsUsedForPasswordReset(t *testing.T) {
	u := &regressionUsers{user: &models.User{ID: 1, Password: "old"}}
	tokens := &regressionTokens{reset: &models.PasswordResetToken{ID: 2, UserID: 1, Token: "stored", ExpiresAt: time.Now().Add(time.Hour)}}
	s := NewPasswordService(u, tokens, config.DefaultIdentityConfig().PasswordPolicy)
	require.ErrorContains(t, s.ResetPassword(context.Background(), "raw", "StrongPassword123!"), "delete failed")
	require.Equal(t, 1, tokens.deletes)
}

func TestCustomTokenRepositoryIsUsedForEmailVerification(t *testing.T) {
	u := &regressionUsers{user: &models.User{ID: 1, Email: "a@example.com"}}
	tokens := &regressionTokens{verification: &models.EmailVerificationToken{ID: 2, UserID: 1, Email: "a@example.com", ExpiresAt: time.Now().Add(time.Hour)}}
	s := NewUserService(u, tokens, nil)
	require.ErrorContains(t, s.VerifyEmail(context.Background(), "raw"), "mark failed")
	require.Equal(t, 1, tokens.marks)
}

func TestInheritedTransactionClaimDoesNotSkipCustomResetTokenConsume(t *testing.T) {
	u := &regressionUsers{user: &models.User{ID: 1, Password: "old"}}
	tokens := &regressionTokens{reset: &models.PasswordResetToken{ID: 2, UserID: 1, Token: "stored", ExpiresAt: time.Now().Add(time.Hour)}}
	shared := &sharedRegressionTokens{tokens}
	s := NewPasswordService(u, &wrappedRegressionTokens{shared, shared, tokens}, config.DefaultIdentityConfig().PasswordPolicy)
	require.ErrorContains(t, s.ResetPassword(context.Background(), "raw", "StrongPassword123!"), "delete failed")
	require.Equal(t, 1, tokens.deletes)
	require.Zero(t, u.passwordUpdates)
}

func TestInheritedTransactionClaimDoesNotSkipCustomVerificationTokenConsume(t *testing.T) {
	u := &regressionUsers{user: &models.User{ID: 1, Email: "a@example.com"}}
	tokens := &regressionTokens{verification: &models.EmailVerificationToken{ID: 2, UserID: 1, Email: "a@example.com", ExpiresAt: time.Now().Add(time.Hour)}}
	shared := &sharedRegressionTokens{tokens}
	s := NewUserService(u, &wrappedRegressionTokens{shared, shared, tokens}, nil)
	require.ErrorContains(t, s.VerifyEmail(context.Background(), "raw"), "mark failed")
	require.Equal(t, 1, tokens.marks)
	require.Zero(t, u.fieldUpdates)
}

func TestCustomResetTokenUsesScopedUserPasswordWrite(t *testing.T) {
	u := &regressionUsers{user: &models.User{ID: 1, Password: "old"}}
	tokens := &successfulRegressionTokens{&regressionTokens{reset: &models.PasswordResetToken{ID: 2, UserID: 1, Token: "stored", ExpiresAt: time.Now().Add(time.Hour)}}}
	s := NewPasswordService(u, tokens, config.DefaultIdentityConfig().PasswordPolicy)
	require.NoError(t, s.ResetPassword(context.Background(), "raw", "StrongPassword123!"))
	require.Equal(t, 1, u.passwordUpdates)
	require.Zero(t, u.fullUpdates)
}

func TestCustomVerificationTokenUsesEmailBoundUserWrite(t *testing.T) {
	u := &scopedEmailUsers{regressionUsers: &regressionUsers{user: &models.User{ID: 1, Email: "a@example.com"}}}
	tokens := &successfulRegressionTokens{&regressionTokens{verification: &models.EmailVerificationToken{ID: 2, UserID: 1, Email: "a@example.com", ExpiresAt: time.Now().Add(time.Hour)}}}
	s := NewUserService(u, tokens, nil)
	require.NoError(t, s.VerifyEmail(context.Background(), "raw"))
	require.Equal(t, 1, u.verifiedWrites)
	require.Zero(t, u.fullUpdates)
}

func TestUpdateEmailRejectsConcurrentAddressChange(t *testing.T) {
	u := &emailCASUsers{regressionUsers: &regressionUsers{user: &models.User{ID: 1, Email: "a@example.com"}}}
	newEmail := "c@example.com"
	s := NewUserService(u, nil, nil)
	_, err := s.UpdateUser(context.Background(), 1, &UpdateUserRequest{Email: &newEmail})
	require.ErrorIs(t, err, repository.ErrConcurrentChange)
	require.Equal(t, 1, u.casCalls)
}

func TestUpdateUnchangedEmailRejectsConcurrentAddressChange(t *testing.T) {
	u := &emailCASUsers{regressionUsers: &regressionUsers{user: &models.User{ID: 1, Email: "a@example.com", EmailVerified: true}}}
	sameEmail := "a@example.com"
	s := NewUserService(u, nil, nil)
	_, err := s.UpdateUser(context.Background(), 1, &UpdateUserRequest{Email: &sameEmail})
	require.ErrorIs(t, err, repository.ErrConcurrentChange)
	require.Equal(t, 1, u.casCalls)
	require.Zero(t, u.fieldUpdates)
}

func TestAuthenticationTouchesOnlyLastLogin(t *testing.T) {
	hash, err := utils.HashPassword("valid-password")
	require.NoError(t, err)
	u := &regressionUsers{user: &models.User{ID: 1, Username: "alice", Password: hash, IsActive: true}}
	s := NewAuthService(u, nil, &mockBackendRegistry{userRepo: u})
	_, err = s.Authenticate(context.Background(), &AuthenticateRequest{UsernameOrEmail: "alice", Password: "valid-password"})
	require.NoError(t, err)
	require.Equal(t, 1, u.lastLoginUpdates)
	require.Equal(t, 0, u.fullUpdates)
}

func TestLegacyAuthenticationPersistsLastLogin(t *testing.T) {
	hash, err := utils.HashPassword("valid-password")
	require.NoError(t, err)
	u := &legacyRegressionUsers{user: &models.User{ID: 1, Username: "alice", Password: hash, IsActive: true}}
	s := NewAuthService(u, nil, &mockBackendRegistry{userRepo: u})
	_, err = s.Authenticate(context.Background(), &AuthenticateRequest{UsernameOrEmail: "alice", Password: "valid-password"})
	require.NoError(t, err)
	require.Equal(t, 1, u.updates)
	require.NotNil(t, u.user.LastLogin)
}

func TestResetRequiresNotifierForKnownUser(t *testing.T) {
	u := &regressionUsers{user: &models.User{ID: 1, Email: "a@example.com"}}
	s := NewPasswordService(u, &regressionTokens{}, config.DefaultIdentityConfig().PasswordPolicy)
	require.ErrorContains(t, s.RequestPasswordReset(context.Background(), "a@example.com"), "notifier")
}

func TestLegacyResetDeletionFailureLeavesPasswordUnchanged(t *testing.T) {
	u := &legacyRegressionUsers{user: &models.User{ID: 1, Password: "old"}}
	tokens := &regressionTokens{reset: &models.PasswordResetToken{ID: 2, UserID: 1, Token: "stored", ExpiresAt: time.Now().Add(time.Hour)}}
	s := NewPasswordService(u, tokens, config.DefaultIdentityConfig().PasswordPolicy)
	require.ErrorContains(t, s.ResetPassword(context.Background(), "raw", "StrongPassword123!"), "delete failed")
	require.Zero(t, u.updates)
	require.Equal(t, "old", u.user.Password)
}

func TestLegacyVerificationMarkFailureLeavesUserUnverified(t *testing.T) {
	u := &legacyRegressionUsers{user: &models.User{ID: 1, Email: "a@example.com"}}
	tokens := &regressionTokens{verification: &models.EmailVerificationToken{ID: 2, UserID: 1, Email: "a@example.com", ExpiresAt: time.Now().Add(time.Hour)}}
	s := NewUserService(u, tokens, nil)
	require.ErrorContains(t, s.VerifyEmail(context.Background(), "raw"), "mark failed")
	require.Zero(t, u.updates)
	require.False(t, u.user.EmailVerified)
}
