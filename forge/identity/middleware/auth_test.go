package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/identity/models"
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
