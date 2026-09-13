package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/service"
)

type permissionServiceStub struct {
	checkPermissionFn func(ctx context.Context, userID int64, permission string) (bool, error)
}

var _ service.PermissionService = (*permissionServiceStub)(nil)

func (s *permissionServiceStub) CheckPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	if s.checkPermissionFn != nil {
		return s.checkPermissionFn(ctx, userID, permission)
	}
	return false, nil
}

func (s *permissionServiceStub) CheckPermissions(ctx context.Context, userID int64, permissions []string) (bool, error) {
	return false, nil
}

func (s *permissionServiceStub) CheckAnyPermission(ctx context.Context, userID int64, permissions []string) (bool, error) {
	return false, nil
}

func (s *permissionServiceStub) GetUserPermissions(ctx context.Context, userID int64) ([]*models.Permission, error) {
	return nil, nil
}

func (s *permissionServiceStub) AssignPermission(ctx context.Context, userID int64, permission string) error {
	return nil
}

func (s *permissionServiceStub) RemovePermission(ctx context.Context, userID int64, permission string) error {
	return nil
}

func TestRequirePermission_UsesPermissionService(t *testing.T) {
	called := false
	svc := &permissionServiceStub{
		checkPermissionFn: func(ctx context.Context, userID int64, permission string) (bool, error) {
			called = true
			if userID != 7 || permission != "users.change_user" {
				t.Fatalf("unexpected permission check args: userID=%d permission=%q", userID, permission)
			}
			return true, nil
		},
	}

	mw := NewAuthenticationMiddlewareWithPermissionService(nil, nil, nil, svc)
	h := mw.RequirePermission("users.change_user")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", &models.User{
		ID:       7,
		IsActive: true,
	}))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
	if !called {
		t.Fatalf("expected permission service to be called")
	}
}

func TestRequirePermission_DeniesWhenPermissionServiceReturnsFalse(t *testing.T) {
	svc := &permissionServiceStub{
		checkPermissionFn: func(ctx context.Context, userID int64, permission string) (bool, error) {
			return false, nil
		},
	}

	mw := NewAuthenticationMiddlewareWithPermissionService(nil, nil, nil, svc)
	h := mw.RequirePermission("users.change_user")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", &models.User{
		ID:       7,
		IsActive: true,
	}))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestRequirePermission_ReturnsInternalErrorWhenPermissionCheckFails(t *testing.T) {
	svc := &permissionServiceStub{
		checkPermissionFn: func(ctx context.Context, userID int64, permission string) (bool, error) {
			return false, errors.New("db unavailable")
		},
	}

	mw := NewAuthenticationMiddlewareWithPermissionService(nil, nil, nil, svc)
	h := mw.RequirePermission("users.change_user")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", &models.User{
		ID:       7,
		IsActive: true,
	}))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestRequirePermission_SuperuserBypassesPermissionService(t *testing.T) {
	called := false
	svc := &permissionServiceStub{
		checkPermissionFn: func(ctx context.Context, userID int64, permission string) (bool, error) {
			called = true
			return false, nil
		},
	}

	mw := NewAuthenticationMiddlewareWithPermissionService(nil, nil, nil, svc)
	h := mw.RequirePermission("users.change_user")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", &models.User{
		ID:          1,
		IsActive:    true,
		IsSuperuser: true,
	}))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
	if called {
		t.Fatalf("did not expect permission service call for superuser")
	}
}

type fakeSessionRepo struct {
	repository.SessionRepository
	getByKeyFn func(ctx context.Context, key string) (*models.UserSession, error)
}

func (f *fakeSessionRepo) GetByKey(ctx context.Context, key string) (*models.UserSession, error) {
	if f.getByKeyFn != nil {
		return f.getByKeyFn(ctx, key)
	}
	return nil, errors.New("session not found")
}

type fakeUserRepo struct {
	repository.UserRepository
	getByIDFn func(ctx context.Context, id int64) (*models.User, error)
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return nil, errors.New("user not found")
}

func TestAuthenticateRequest_SessionAuth(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name           string
		sessionKey     string
		useCookie      bool
		session        *models.UserSession
		user           *models.User
		nilSessionRepo bool
		nilUserRepo    bool
		wantUser       bool
	}{
		{
			name:       "active, unlocked user with valid session",
			sessionKey: "valid-key",
			session:    &models.UserSession{UserID: 1, SessionKey: "valid-key", ExpiresAt: &future},
			user:       &models.User{ID: 1, IsActive: true, IsLocked: false},
			wantUser:   true,
		},
		{
			name:       "IsActive=false",
			sessionKey: "valid-key",
			session:    &models.UserSession{UserID: 1, SessionKey: "valid-key", ExpiresAt: &future},
			user:       &models.User{ID: 1, IsActive: false, IsLocked: false},
			wantUser:   false,
		},
		{
			name:       "IsLocked=true",
			sessionKey: "valid-key",
			session:    &models.UserSession{UserID: 1, SessionKey: "valid-key", ExpiresAt: &future},
			user:       &models.User{ID: 1, IsActive: true, IsLocked: true},
			wantUser:   false,
		},
		{
			name:       "expired session",
			sessionKey: "expired-key",
			session:    &models.UserSession{UserID: 1, SessionKey: "expired-key", ExpiresAt: &past},
			user:       &models.User{ID: 1, IsActive: true, IsLocked: false},
			wantUser:   false,
		},
		{
			name:       "no session key",
			sessionKey: "",
			session:    nil,
			user:       nil,
			wantUser:   false,
		},
		{
			name:           "nil session repo",
			sessionKey:     "valid-key",
			nilSessionRepo: true,
			user:           &models.User{ID: 1, IsActive: true, IsLocked: false},
			wantUser:       false,
		},
		{
			name:        "nil user repo",
			sessionKey:  "valid-key",
			session:     &models.UserSession{UserID: 1, SessionKey: "valid-key", ExpiresAt: &future},
			nilUserRepo: true,
			wantUser:    false,
		},
		{
			name:       "active, unlocked user with valid cookie session",
			sessionKey: "cookie-key",
			useCookie:  true,
			session:    &models.UserSession{UserID: 2, SessionKey: "cookie-key", ExpiresAt: &future},
			user:       &models.User{ID: 2, IsActive: true, IsLocked: false},
			wantUser:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sessionRepo repository.SessionRepository
			if !tt.nilSessionRepo {
				sessionRepo = &fakeSessionRepo{
					getByKeyFn: func(ctx context.Context, key string) (*models.UserSession, error) {
						if tt.session != nil && key == tt.session.SessionKey {
							return tt.session, nil
						}
						return nil, errors.New("not found")
					},
				}
			}

			var userRepo repository.UserRepository
			if !tt.nilUserRepo {
				userRepo = &fakeUserRepo{
					getByIDFn: func(ctx context.Context, id int64) (*models.User, error) {
						if tt.user != nil && id == tt.user.ID {
							return tt.user, nil
						}
						return nil, errors.New("not found")
					},
				}
			}

			mw := NewAuthenticationMiddleware(nil, sessionRepo, userRepo)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.sessionKey != "" {
				if tt.useCookie {
					req.AddCookie(&http.Cookie{Name: "session_key", Value: tt.sessionKey})
				} else {
					req.Header.Set("X-Session-Key", tt.sessionKey)
				}
			}

			user, err := mw.authenticateRequest(req.Context(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantUser {
				if user == nil {
					t.Fatalf("expected user, got nil")
				}
				if user.ID != tt.user.ID {
					t.Fatalf("expected user ID %d, got %d", tt.user.ID, user.ID)
				}
			} else {
				if user != nil {
					t.Fatalf("expected nil user, got %+v", user)
				}
			}
		})
	}
}
