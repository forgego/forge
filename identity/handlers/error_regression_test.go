package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

var errPrivateDB = errors.New("private database connection string")

type failingUserService struct{ service.UserService }

func (failingUserService) Register(context.Context, *service.RegisterRequest) (*models.User, error) {
	return nil, errPrivateDB
}
func (failingUserService) CreateUser(context.Context, *service.CreateUserRequest) (*models.User, error) {
	return nil, errPrivateDB
}
func (failingUserService) GetUser(context.Context, int64) (*models.User, error) {
	return nil, errPrivateDB
}
func (failingUserService) ListUsers(context.Context, *service.UserFilters) ([]*models.User, int64, error) {
	return nil, 0, errPrivateDB
}
func (failingUserService) UpdateUser(context.Context, int64, *service.UpdateUserRequest) (*models.User, error) {
	return nil, errPrivateDB
}
func (failingUserService) DeleteUser(context.Context, int64) error { return errPrivateDB }

type failingAuthService struct{ service.AuthService }

func (failingAuthService) Authenticate(context.Context, *service.AuthenticateRequest) (*models.User, error) {
	return nil, errPrivateDB
}
func (failingAuthService) Logout(context.Context, string) error { return errPrivateDB }
func (failingAuthService) GetSession(context.Context, string) (*models.UserSession, error) {
	return nil, errPrivateDB
}

func TestIdentityHandlersHideUnexpectedServiceErrors(t *testing.T) {
	users := NewUserHandler(failingUserService{})
	auth := NewAuthHandler(failingAuthService{}, failingUserService{}, nil)
	tests := []struct {
		name, method, path, body string
		handler                  http.HandlerFunc
	}{
		{"register", http.MethodPost, "/auth/register", `{"username":"alice","email":"a@example.com","password":"strongpass"}`, auth.Register},
		{"login", http.MethodPost, "/auth/login", `{"username":"alice","password":"strongpass"}`, auth.Login},
		{"logout", http.MethodPost, "/auth/logout", "", auth.Logout},
		{"list", http.MethodGet, "/users", "", users.List},
		{"create", http.MethodPost, "/users", `{"username":"alice","email":"a@example.com","password":"strongpass"}`, users.Create},
		{"retrieve", http.MethodGet, "/users/1", "", users.Retrieve},
		{"update", http.MethodPut, "/users/1", `{"first_name":"Alice"}`, users.Update},
		{"delete", http.MethodDelete, "/users/1", "", users.Delete},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			req.Header.Set("X-Session-Key", "key")
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req = req.WithContext(core.WithUser(req.Context(), &models.User{ID: 1, IsSuperuser: true}))
			resp := httptest.NewRecorder()
			tc.handler(resp, req)
			require.Equal(t, http.StatusInternalServerError, resp.Code)
			require.NotContains(t, resp.Body.String(), errPrivateDB.Error())
		})
	}
}
