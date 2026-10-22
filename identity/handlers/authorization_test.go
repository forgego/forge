package handlers

import (
	"context"
	"fmt"
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

type authorizationUsers struct {
	service.UserService
	users map[int64]*models.User
}

func (s *authorizationUsers) ListUsers(context.Context, *service.UserFilters) ([]*models.User, int64, error) {
	var users []*models.User
	for _, user := range s.users {
		users = append(users, user)
	}
	return users, int64(len(users)), nil
}
func (s *authorizationUsers) GetUser(_ context.Context, id int64) (*models.User, error) {
	if user := s.users[id]; user != nil {
		return user, nil
	}
	return nil, service.ErrUserNotFound
}
func (s *authorizationUsers) UpdateUser(_ context.Context, id int64, req *service.UpdateUserRequest) (*models.User, error) {
	user := s.users[id]
	if user == nil {
		return nil, service.ErrUserNotFound
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.IsStaff != nil {
		user.IsStaff = *req.IsStaff
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.IsLocked != nil {
		user.IsLocked = *req.IsLocked
	}
	return user, nil
}
func (s *authorizationUsers) DeleteUser(_ context.Context, id int64) error {
	delete(s.users, id)
	return nil
}

type authorizationAuth struct {
	service.AuthService
	sessions map[string]*models.UserSession
}

func (s *authorizationAuth) GetSession(_ context.Context, key string) (*models.UserSession, error) {
	if session := s.sessions[key]; session != nil {
		return session, nil
	}
	return nil, fmt.Errorf("session not found")
}
func (s *authorizationAuth) Logout(_ context.Context, key string) error {
	delete(s.sessions, key)
	return nil
}

func authorizationRequest(method, path, body string, actor *models.User) (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r = r.WithContext(core.WithUser(r.Context(), actor))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "2")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	return r, httptest.NewRecorder()
}

func TestUserRoutesEnforceActorAndTarget(t *testing.T) {
	actor := &models.User{ID: 1, IsActive: true}
	staff := &models.User{ID: 3, IsActive: true, IsStaff: true}
	superuser := &models.User{ID: 4, IsActive: true, IsSuperuser: true}
	for _, tc := range []struct {
		name, method, body string
		actor              *models.User
		want               int
		staffAfter         bool
		firstName          string
	}{
		{"ordinary user cannot list users", "list", "", actor, 403, false, ""},
		{"staff can list users", "list", "", staff, 200, false, ""},
		{"superuser can list users", "list", "", superuser, 200, false, ""},
		{"ordinary user cannot retrieve another user", "retrieve", "", actor, 403, false, ""},
		{"ordinary user can retrieve self", "retrieve", "", &models.User{ID: 2, IsActive: true}, 200, false, ""},
		{"staff can retrieve another user", "retrieve", "", staff, 200, false, ""},
		{"superuser can retrieve another user", "retrieve", "", superuser, 200, false, ""},
		{"ordinary user cannot promote self", "update", `{"is_staff":true}`, &models.User{ID: 2, IsActive: true}, 403, false, ""},
		{"forbidden flag prevents profile changes", "update", `{"first_name":"Changed","is_staff":false}`, &models.User{ID: 2, IsActive: true}, 403, false, ""},
		{"ordinary user cannot promote another user", "update", `{"is_staff":true}`, actor, 403, false, ""},
		{"ordinary user cannot update another user", "update", `{"first_name":"Changed"}`, actor, 403, false, ""},
		{"staff cannot update another user", "update", `{"first_name":"Changed"}`, staff, 403, false, ""},
		{"ordinary user can update own name", "update", `{"first_name":"Changed"}`, &models.User{ID: 2, IsActive: true}, 200, false, "Changed"},
		{"superuser can promote another user", "update", `{"is_staff":true}`, superuser, 200, true, ""},
		{"ordinary user cannot lock self", "update", `{"is_locked":true}`, &models.User{ID: 2, IsActive: true}, 403, false, ""},
		{"ordinary user cannot deactivate self", "update", `{"is_active":false}`, &models.User{ID: 2, IsActive: true}, 403, false, ""},
		{"ordinary user cannot delete another user", "delete", "", actor, 403, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &authorizationUsers{users: map[int64]*models.User{2: {ID: 2, IsActive: true, Email: "target@example.com"}}}
			h := NewUserHandler(s)
			r, w := authorizationRequest(http.MethodGet, "/users/2", tc.body, tc.actor)
			switch tc.method {
			case "list":
				h.List(w, r)
			case "retrieve":
				h.Retrieve(w, r)
			case "update":
				h.Update(w, r)
			case "delete":
				h.Delete(w, r)
			}
			require.Equal(t, tc.want, w.Code)
			require.Equal(t, tc.staffAfter, s.users[2].IsStaff)
			require.Equal(t, tc.firstName, s.users[2].FirstName)
			require.True(t, s.users[2].IsActive)
			require.False(t, s.users[2].IsLocked)
		})
	}
}

func TestLogoutCannotRevokeAnotherUsersSession(t *testing.T) {
	s := &authorizationAuth{sessions: map[string]*models.UserSession{"other": {UserID: 2}}}
	h := NewAuthHandler(s, nil, nil)
	r, w := authorizationRequest(http.MethodPost, "/auth/logout", `{"session_key":"other"}`, &models.User{ID: 1})
	h.Logout(w, r)
	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, s.sessions, "other")
}

func TestLogoutRevokesOwnSession(t *testing.T) {
	s := &authorizationAuth{sessions: map[string]*models.UserSession{"own": {UserID: 1}}}
	h := NewAuthHandler(s, nil, nil)
	r, w := authorizationRequest(http.MethodPost, "/auth/logout", `{"session_key":"own"}`, &models.User{ID: 1})
	h.Logout(w, r)
	require.Equal(t, http.StatusNoContent, w.Code)
	require.NotContains(t, s.sessions, "own")
}
