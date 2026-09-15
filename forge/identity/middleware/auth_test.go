package middleware

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/service"
	"github.com/gorilla/csrf"
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
	req = req.WithContext(core.WithUser(req.Context(), &models.User{
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
	req = req.WithContext(core.WithUser(req.Context(), &models.User{
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
	req = req.WithContext(core.WithUser(req.Context(), &models.User{
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
	req = req.WithContext(core.WithUser(req.Context(), &models.User{
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

			req := httptest.NewRequest(http.MethodGet, "https://example.com/test", nil)
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

func TestAuthenticateRequest_CookiePostRequiresCSRF(t *testing.T) {
	mw := newSessionAuthMiddleware()
	req := httptest.NewRequest(http.MethodPost, "https://example.com/test", nil)
	req.AddCookie(&http.Cookie{Name: "session_key", Value: "valid-key"})

	user, err := mw.authenticateRequest(req.Context(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Fatalf("expected unauthenticated request, got user %+v", user)
	}
}

func TestAuthenticateRequest_CookiePostWithoutCSRFLogsWarningOnce(t *testing.T) {
	var logs bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	mw := newSessionAuthMiddleware()
	for range 2 {
		req := httptest.NewRequest(http.MethodPost, "https://example.com/protected", nil)
		req.AddCookie(&http.Cookie{Name: "session_key", Value: "valid-key"})

		user, err := mw.authenticateRequest(req.Context(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user != nil {
			t.Fatalf("expected unauthenticated request, got user %+v", user)
		}
	}

	if count := strings.Count(logs.String(), "session cookie ignored on unsafe request: CSRF middleware is not mounted on this route"); count != 1 {
		t.Fatalf("expected one CSRF middleware warning, got %d: %s", count, logs.String())
	}
	if !strings.Contains(logs.String(), "method=POST") || !strings.Contains(logs.String(), "path=/protected") {
		t.Fatalf("expected warning to include request method and path, got: %s", logs.String())
	}
}

func TestAuthenticateRequest_CookiePostWithCSRFIsAuthenticated(t *testing.T) {
	mw := newSessionAuthMiddleware()
	csrfMiddleware := csrf.Protect([]byte("01234567890123456789012345678901"), csrf.Secure(false), csrf.TrustedOrigins([]string{"example.com"}))
	getHandler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(csrf.Token(r)))
	}))
	getRec := httptest.NewRecorder()
	getHandler.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "https://example.com/test", nil))

	csrfToken := getRec.Body.String()
	if csrfToken == "" {
		t.Fatal("expected CSRF token from GET request")
	}
	postReq := httptest.NewRequest(http.MethodPost, "https://example.com/test", nil)
	postReq.AddCookie(&http.Cookie{Name: "session_key", Value: "valid-key"})
	for _, cookie := range getRec.Result().Cookies() {
		postReq.AddCookie(cookie)
	}
	postReq.Header.Set("X-CSRF-Token", csrfToken)
	postReq.Header.Set("Referer", "https://example.com/")
	postRec := httptest.NewRecorder()
	csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := mw.authenticateRequest(r.Context(), r)
		if err != nil || user == nil {
			t.Errorf("expected authenticated request, user=%+v err=%v", user, err)
		}
	})).ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, postRec.Code, postRec.Body.String())
	}
}

func TestAuthenticateRequest_CookieGetIsAuthenticated(t *testing.T) {
	mw := newSessionAuthMiddleware()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session_key", Value: "valid-key"})

	user, err := mw.authenticateRequest(req.Context(), req)
	if err != nil || user == nil {
		t.Fatalf("expected authenticated request, user=%+v err=%v", user, err)
	}
}

func TestAuthenticateRequest_HeaderPostDoesNotRequireCSRF(t *testing.T) {
	mw := newSessionAuthMiddleware()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("X-Session-Key", "valid-key")

	user, err := mw.authenticateRequest(req.Context(), req)
	if err != nil || user == nil {
		t.Fatalf("expected authenticated request, user=%+v err=%v", user, err)
	}
}

func newSessionAuthMiddleware() *AuthenticationMiddleware {
	future := time.Now().Add(time.Hour)
	return NewAuthenticationMiddleware(nil,
		&fakeSessionRepo{getByKeyFn: func(context.Context, string) (*models.UserSession, error) {
			return &models.UserSession{UserID: 1, ExpiresAt: &future}, nil
		}},
		&fakeUserRepo{getByIDFn: func(context.Context, int64) (*models.User, error) {
			return &models.User{ID: 1, IsActive: true}, nil
		}},
	)
}

func TestSanitizeLogValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain string",
			input: "/api/v1/resource",
			want:  "/api/v1/resource",
		},
		{
			name:  "string with \r\n",
			input: "/api/v1/resource\r\nmalicious-log-entry",
			want:  "/api/v1/resourcemalicious-log-entry",
		},
		{
			name:  "string with only \n",
			input: "/api/v1/resource\nmalicious-log-entry",
			want:  "/api/v1/resourcemalicious-log-entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeLogValue(tt.input)
			if got != tt.want {
				t.Fatalf("sanitizeLogValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
