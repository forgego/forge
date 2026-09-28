package project

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/identity"
	forgehttp "github.com/forgego/forge/server"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestAuthCommandExecute_ScaffoldsAuthAPIWithLoginLogoutAndJWT(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n\ngo 1.25.0\n"), 0644))
	require.NoError(t, os.Chdir(tmp))

	cmd := NewAuthCommand()
	err = cmd.Execute(&core.Context{}, nil)
	require.NoError(t, err)

	apiPath := filepath.Join(tmp, "app", "auth", "api.go")
	contentBytes, err := os.ReadFile(apiPath)
	require.NoError(t, err)
	content := string(contentBytes)

	require.Contains(t, content, `router.Post("/api/v1/auth/login", handleLogin(signingKey))`)
	require.Contains(t, content, `router.Post("/api/v1/auth/logout", handleLogout)`)
	require.Contains(t, content, "func generateJWTToken(userID, username string, signingKey []byte) (string, error)")
	require.Contains(t, content, "hmac.New(sha256.New")
	require.NotContains(t, content, "change-me")
	require.NotContains(t, content, `generateJWTToken("1"`)
	require.Contains(t, content, "identity.CheckPasswordHash(password, user.Password)")
	require.NotContains(t, strings.ToLower(content), "todo:")
}

func TestAuthCommandExecute_ReturnsErrorWhenAuthAppAlreadyExists(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n\ngo 1.25.0\n"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "app", "auth"), 0755))
	require.NoError(t, os.Chdir(tmp))

	cmd := NewAuthCommand()
	err = cmd.Execute(&core.Context{}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "auth app already exists")
}

// secureScaffoldUser mirrors the generated auth User model for tests.
type secureScaffoldUser struct {
	ID          int64  `json:"id" db:"id"`
	Username    string `json:"username" db:"username"`
	Email       string `json:"email" db:"email"`
	Password    string `json:"password" db:"password"`
	IsActive    bool   `json:"is_active" db:"is_active"`
	IsStaff     bool   `json:"is_staff" db:"is_staff"`
	IsSuperuser bool   `json:"is_superuser" db:"is_superuser"`
}

type secureScaffoldUserSerializer struct {
	*api.BaseSerializer
}

func newSecureScaffoldUserSerializer() api.Serializer {
	return &secureScaffoldUserSerializer{BaseSerializer: api.NewBaseSerializer(make(map[string]interface{}))}
}

func (s *secureScaffoldUserSerializer) New() api.Serializer {
	return newSecureScaffoldUserSerializer()
}

func (s *secureScaffoldUserSerializer) Fields() []string {
	return []string{"id", "username", "email", "is_active", "date_joined"}
}

func (s *secureScaffoldUserSerializer) ReadOnlyFields() []string {
	return []string{"id", "is_staff", "is_superuser", "date_joined", "last_login"}
}

func (s *secureScaffoldUserSerializer) WriteOnlyFields() []string {
	return []string{"password"}
}

type secureScaffoldUserStore struct {
	items []*secureScaffoldUser
}

func (s *secureScaffoldUserStore) All(context.Context) ([]*secureScaffoldUser, error) {
	return s.items, nil
}

func (s *secureScaffoldUserStore) Count(context.Context) (int64, error) {
	return int64(len(s.items)), nil
}

func (s *secureScaffoldUserStore) Offset(int) *secureScaffoldUserStore { return s }
func (s *secureScaffoldUserStore) Limit(int) *secureScaffoldUserStore  { return s }

func (s *secureScaffoldUserStore) Get(_ context.Context, id int64) (*secureScaffoldUser, error) {
	for _, u := range s.items {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *secureScaffoldUserStore) Create(_ context.Context, model *secureScaffoldUser) error {
	model.ID = int64(len(s.items) + 1)
	s.items = append(s.items, model)
	return nil
}

func (s *secureScaffoldUserStore) Update(_ context.Context, _ *secureScaffoldUser) error { return nil }
func (s *secureScaffoldUserStore) Delete(_ context.Context, _ *secureScaffoldUser) error { return nil }

// secureScaffoldUserViewSet mirrors the generated userViewSet wrapper: it
// strips privilege flags and bcrypt-hashes passwords before delegating.
type secureScaffoldUserViewSet struct {
	*api.BaseViewSet
}

func (vs *secureScaffoldUserViewSet) Create(w http.ResponseWriter, r *http.Request) {
	if !securePrepareUserWrite(w, r) {
		return
	}
	vs.BaseViewSet.Create(w, r)
}

func securePrepareUserWrite(w http.ResponseWriter, r *http.Request) bool {
	if r.Body == nil {
		return true
	}
	body, err := io.ReadAll(r.Body)
	_ = r.Body.Close()
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	if len(bytes.TrimSpace(body)) == 0 {
		r.Body = io.NopCloser(bytes.NewReader(body))
		return true
	}
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		return true
	}
	delete(data, "is_staff")
	delete(data, "is_superuser")
	if raw, ok := data["password"]; ok {
		password, _ := raw.(string)
		if strings.TrimSpace(password) == "" {
			delete(data, "password")
		} else {
			hashed, err := identity.HashPassword(password)
			if err != nil {
				http.Error(w, "could not process password", http.StatusInternalServerError)
				return false
			}
			data["password"] = hashed
		}
	}
	rewritten, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	r.Body = io.NopCloser(bytes.NewReader(rewritten))
	r.ContentLength = int64(len(rewritten))
	return true
}

var secureScaffoldSigningKey = []byte("test-signing-key-must-be-32-bytes!")

func newSecureScaffoldRouter(store *secureScaffoldUserStore) *forgehttp.Router {
	viewset := api.NewBaseViewSet(
		newSecureScaffoldUserSerializer,
		store,
		&secureScaffoldUser{},
	)
	staff := &secureScaffoldUser{ID: 1, Username: "admin", IsStaff: true}
	viewset.Authentication = []authentication.Authentication{
		authentication.NewJWTAuthentication(secureScaffoldSigningKey, func(authentication.JWTClaims) (interface{}, error) {
			return staff, nil
		}),
	}
	viewset.Permissions = []permissions.Permission{
		permissions.NewIsAuthenticated(),
		permissions.NewIsStaffUser(),
	}
	apiRouter := api.NewRouter("/api/v1")
	apiRouter.Register("users", &secureScaffoldUserViewSet{BaseViewSet: viewset})
	handler := forgehttp.NewRouter()
	apiRouter.RegisterRoutes(handler)
	return handler
}

func secureScaffoldAuthHeader(t *testing.T) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "admin",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString(secureScaffoldSigningKey)
	require.NoError(t, err)
	return "Bearer " + signed
}

func serveSecureScaffold(handler http.Handler, method, path, body, authHeader string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestGeneratedUsersAPIRejectsAnonymousList(t *testing.T) {
	store := &secureScaffoldUserStore{items: []*secureScaffoldUser{{ID: 1, Username: "admin", Password: "hashed", IsStaff: true}}}
	handler := newSecureScaffoldRouter(store)
	resp := serveSecureScaffold(handler, http.MethodGet, "/api/v1/users/", "", "")
	require.Equal(t, http.StatusForbidden, resp.Code)
}

func TestGeneratedUsersAPICreateUpdateDeleteRequireStaff(t *testing.T) {
	store := &secureScaffoldUserStore{}
	handler := newSecureScaffoldRouter(store)

	resp := serveSecureScaffold(handler, http.MethodPost, "/api/v1/users/", `{"username":"mallory","password":"pw","is_staff":true}`, "")
	require.Equal(t, http.StatusForbidden, resp.Code)
	require.Empty(t, store.items, "anonymous POST must not create a user")

	resp = serveSecureScaffold(handler, http.MethodPut, "/api/v1/users/1", `{"username":"x"}`, "")
	require.Equal(t, http.StatusForbidden, resp.Code)

	resp = serveSecureScaffold(handler, http.MethodDelete, "/api/v1/users/1", "", "")
	require.Equal(t, http.StatusForbidden, resp.Code)

	// Authenticated but non-staff callers are rejected as well.
	nonStaff := &secureScaffoldUser{ID: 2, Username: "user"}
	plain := api.NewBaseViewSet(newSecureScaffoldUserSerializer, store, &secureScaffoldUser{})
	plain.Authentication = []authentication.Authentication{
		authentication.NewJWTAuthentication(secureScaffoldSigningKey, func(authentication.JWTClaims) (interface{}, error) {
			return nonStaff, nil
		}),
	}
	plain.Permissions = []permissions.Permission{permissions.NewIsAuthenticated(), permissions.NewIsStaffUser()}
	apiRouter := api.NewRouter("/api/v1")
	apiRouter.Register("users", plain)
	nonStaffHandler := forgehttp.NewRouter()
	apiRouter.RegisterRoutes(nonStaffHandler)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"username": "user", "exp": time.Now().Add(time.Hour).Unix()})
	signed, err := token.SignedString(secureScaffoldSigningKey)
	require.NoError(t, err)
	resp = serveSecureScaffold(nonStaffHandler, http.MethodGet, "/api/v1/users/", "", "Bearer "+signed)
	require.Equal(t, http.StatusForbidden, resp.Code)
}

func TestGeneratedUsersAPIListNeverExposesPassword(t *testing.T) {
	store := &secureScaffoldUserStore{items: []*secureScaffoldUser{
		{ID: 1, Username: "admin", Password: "$2a$10$supersecrethashvalue00000000000000000000001", IsStaff: true},
	}}
	handler := newSecureScaffoldRouter(store)
	resp := serveSecureScaffold(handler, http.MethodGet, "/api/v1/users/", "", secureScaffoldAuthHeader(t))
	require.Equal(t, http.StatusOK, resp.Code)
	body := resp.Body.String()
	require.Contains(t, body, "admin")
	require.NotContains(t, body, "supersecrethashvalue")
	require.NotContains(t, strings.ToLower(body), "password")
	require.NotContains(t, body, "is_staff")
	require.NotContains(t, body, "is_superuser")
}

func TestGeneratedUsersAPICreateHashesPasswordAndIgnoresPrivilegeFlags(t *testing.T) {
	store := &secureScaffoldUserStore{}
	handler := newSecureScaffoldRouter(store)
	resp := serveSecureScaffold(handler, http.MethodPost, "/api/v1/users/",
		`{"username":"newuser","email":"new@example.com","password":"plaintext-pw","is_staff":true,"is_superuser":true}`,
		secureScaffoldAuthHeader(t))
	require.Equal(t, http.StatusCreated, resp.Code)
	require.Len(t, store.items, 1)
	stored := store.items[0].Password
	require.NotEmpty(t, stored)
	require.NotEqual(t, "plaintext-pw", stored, "password must be stored as a hash, not plaintext")
	require.True(t, identity.CheckPasswordHash("plaintext-pw", stored))
	require.False(t, store.items[0].IsStaff, "is_staff must not be settable through the public endpoint")
	require.False(t, store.items[0].IsSuperuser, "is_superuser must not be settable through the public endpoint")
	require.NotContains(t, strings.ToLower(resp.Body.String()), "password")
}

func TestAuthScaffoldEmitsSecureUserAPI(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n\ngo 1.25.0\n"), 0644))
	require.NoError(t, os.Chdir(tmp))
	require.NoError(t, NewAuthCommand().Execute(&core.Context{}, nil))

	contentBytes, err := os.ReadFile(filepath.Join(tmp, "app", "auth", "api.go"))
	require.NoError(t, err)
	content := string(contentBytes)

	require.Contains(t, content, "authentication.NewJWTAuthentication")
	require.Contains(t, content, "permissions.NewIsAuthenticated()")
	require.Contains(t, content, "permissions.NewIsStaffUser()")
	require.Contains(t, content, "identity.HashPassword")
	require.Contains(t, content, "WriteOnlyFields")
	require.Contains(t, content, "ReadOnlyFields")
	require.Contains(t, content, "is_staff")
	require.Contains(t, content, "is_superuser")
	require.NotContains(t, content, `apiRouter.Register("users", viewset)`)
}
