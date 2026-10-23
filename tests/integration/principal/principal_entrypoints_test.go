package principal_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	adminrest "github.com/forgego/forge/admin/api/rest"
	admincore "github.com/forgego/forge/admin/core"
	forgeapi "github.com/forgego/forge/api"
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	forgehttp "github.com/forgego/forge/server"
	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type entrypointTarget struct {
	ID    int64  `json:"id"`
	Value string `json:"value"`
}

type entrypointSerializer struct{ *forgeapi.BaseSerializer }

func newEntrypointSerializer() forgeapi.Serializer {
	return &entrypointSerializer{BaseSerializer: forgeapi.NewBaseSerializer(map[string]interface{}{})}
}
func (*entrypointSerializer) Fields() []string { return []string{"value"} }

type sqliteTargetWriter struct {
	db     *sql.DB
	limit  int
	offset int
}

func (w sqliteTargetWriter) Count(context.Context) (int64, error) {
	var count int64
	err := w.db.QueryRow("SELECT COUNT(*) FROM entrypoint_targets").Scan(&count)
	return count, err
}
func (w sqliteTargetWriter) All(context.Context) ([]*entrypointTarget, error) {
	rows, err := w.db.Query("SELECT id, value FROM entrypoint_targets LIMIT ? OFFSET ?", w.limit, w.offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var targets []*entrypointTarget
	for rows.Next() {
		target := new(entrypointTarget)
		if err := rows.Scan(&target.ID, &target.Value); err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, rows.Err()
}
func (w sqliteTargetWriter) Offset(n int) *sqliteTargetWriter { w.offset = n; return &w }
func (w sqliteTargetWriter) Limit(n int) *sqliteTargetWriter  { w.limit = n; return &w }
func (w sqliteTargetWriter) Get(_ context.Context, id int64) (*entrypointTarget, error) {
	target := new(entrypointTarget)
	err := w.db.QueryRow("SELECT id, value FROM entrypoint_targets WHERE id = ?", id).Scan(&target.ID, &target.Value)
	return target, err
}

func (w sqliteTargetWriter) Create(_ context.Context, object interface{}) error {
	target := object.(*entrypointTarget)
	_, err := w.db.Exec("UPDATE entrypoint_targets SET value = ? WHERE id = 1", target.Value)
	return err
}
func (w sqliteTargetWriter) Update(context.Context, *entrypointTarget) error { return nil }
func (w sqliteTargetWriter) Delete(context.Context, *entrypointTarget) error { return nil }

type tokenStore map[string]string

func (s tokenStore) IssueToken(_ context.Context, username string, _ time.Duration) (string, error) {
	return username, nil
}
func (s tokenStore) ValidateToken(_ context.Context, token string) (string, bool, error) {
	username, ok := s[token]
	return username, ok, nil
}
func (tokenStore) RevokeToken(context.Context, string) error { return nil }

type apiAuth struct{ user *models.User }

func (a apiAuth) Authenticate(*http.Request) (*authentication.AuthResult, error) {
	if a.user == nil {
		return nil, nil
	}
	return authentication.NewAuthResult(a.user, "test-token"), nil
}
func (apiAuth) AuthenticateHeader(*http.Request) string { return "Bearer" }

type targetAdmin struct{ db *sql.DB }

func (*targetAdmin) SetDB(*db.DB)            {}
func (*targetAdmin) ModelName() string       { return "entrypoint_targets" }
func (*targetAdmin) ModelType() reflect.Type { return reflect.TypeOf(entrypointTarget{}) }
func (*targetAdmin) GetMetadata(context.Context, interface{}) (*admincore.Metadata, error) {
	return &admincore.Metadata{}, nil
}
func (*targetAdmin) HasAddPermission(context.Context, interface{}) bool                 { return true }
func (*targetAdmin) HasChangePermission(context.Context, interface{}, interface{}) bool { return true }
func (*targetAdmin) HasDeletePermission(context.Context, interface{}, interface{}) bool { return true }
func (*targetAdmin) HasViewPermission(context.Context, interface{}, interface{}) bool   { return true }
func (*targetAdmin) HasModulePermission(context.Context, interface{}) bool              { return true }
func (*targetAdmin) ManagerInterface() interface{}                                      { return nil }
func (*targetAdmin) ConfigInterface() interface{}                                       { return nil }
func (*targetAdmin) GetHistory(context.Context, string) ([]admincore.LogEntry, error) {
	return nil, nil
}
func (*targetAdmin) LogAction(context.Context, interface{}, string, string, admincore.ActionType, string) error {
	return nil
}
func (*targetAdmin) PageType() string { return "list" }
func (*targetAdmin) ListObjects(context.Context, admincore.ListParams) (*admincore.PaginatedResponse, error) {
	return nil, nil
}
func (*targetAdmin) GetObject(context.Context, interface{}) (interface{}, error) { return nil, nil }
func (a *targetAdmin) CreateObject(_ context.Context, data map[string]interface{}) (interface{}, error) {
	_, err := a.db.Exec("UPDATE entrypoint_targets SET value = ? WHERE id = 1", data["value"])
	return data, err
}
func (*targetAdmin) UpdateObject(context.Context, interface{}, map[string]interface{}) (interface{}, error) {
	return nil, nil
}
func (*targetAdmin) DeleteObject(context.Context, interface{}) error { return nil }
func (*targetAdmin) ExecuteAction(context.Context, string, []interface{}, map[string]interface{}) (interface{}, error) {
	return nil, nil
}
func (*targetAdmin) Autocomplete(context.Context, string, int) ([]admincore.AutocompleteItem, error) {
	return nil, nil
}

type sqliteBusinessCommand struct{ db *sql.DB }

func (c *sqliteBusinessCommand) Definition() *cobra.Command {
	return &cobra.Command{Use: "mutate"}
}
func (c *sqliteBusinessCommand) Execute(ctx *core.Context, _ []string) error {
	if ctx.Principal == nil || ctx.Principal.Kind != "user" {
		return errors.New("business command requires a resolved user")
	}
	_, err := c.db.Exec("UPDATE entrypoint_targets SET value = 'changed' WHERE id = 1")
	return err
}

func TestPrincipalEntryPointsDenyUnavailableUsers(t *testing.T) {
	for _, tc := range []struct {
		name     string
		username string
		apiCode  int
	}{
		{name: "inactive user", username: "inactive", apiCode: http.StatusForbidden},
		{name: "unknown user", username: "unknown", apiCode: http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			isolatedDir := filepath.Join(t.TempDir(), "work")
			require.NoError(t, os.MkdirAll(filepath.Join(isolatedDir, "config"), 0755))
			require.NoError(t, os.MkdirAll(filepath.Join(filepath.Dir(isolatedDir), "config"), 0755))
			require.NoError(t, os.Chdir(isolatedDir))
			t.Cleanup(func() { _ = os.Chdir(originalDir) })
			dsn := filepath.Join(t.TempDir(), "principals.sqlite")
			sqlDB, err := sql.Open("sqlite3", dsn)
			if err != nil {
				t.Fatal(err)
			}
			defer sqlDB.Close()
			require.NoError(t, createEntrypointSchema(sqlDB))
			_, err = sqlDB.Exec("INSERT INTO entrypoint_targets (id, value) VALUES (1, 'original')")
			require.NoError(t, err)
			if tc.username == "inactive" {
				_, err = sqlDB.Exec("INSERT INTO users (id, username, email, password, is_active, is_staff, is_superuser, is_locked) VALUES (1, 'inactive', 'inactive@example.invalid', '', 0, 1, 0, 0)")
				require.NoError(t, err)
			}

			assertTarget := func(entrypoint string) {
				t.Helper()
				var value string
				require.NoError(t, sqlDB.QueryRow("SELECT value FROM entrypoint_targets WHERE id = 1").Scan(&value))
				assert.Equal(t, "original", value, "%s must leave the target row unchanged", entrypoint)
			}
			identityDB := mustForgeDB(t, dsn)
			defer identityDB.Close()
			userStore := repository.NewUserRepository(identityDB)
			if tc.username == "inactive" {
				_, err := userStore.GetByUsername(context.Background(), tc.username)
				require.NoError(t, err)
			}

			adminRegistry := admincore.NewRegistry()
			require.NoError(t, adminRegistry.Register(&targetAdmin{db: sqlDB}))
			adminRouter := adminrest.NewRouter(adminRegistry).
				SetUserRepository(userStore).
				SetStores(adminrest.Stores{Tokens: tokenStore{"token": tc.username}})
			adminMux := chi.NewRouter()
			adminRouter.RegisterRoutes(adminMux)
			adminReq := httptest.NewRequest(http.MethodPost, "/api/entrypoint_targets/", strings.NewReader(`{"value":"changed"}`))
			adminReq.Header.Set("Authorization", "Bearer token")
			adminRec := httptest.NewRecorder()
			adminMux.ServeHTTP(adminRec, adminReq)
			require.Equal(t, http.StatusUnauthorized, adminRec.Code, adminRec.Body.String())
			assertTarget("admin REST")

			var apiUser *models.User
			if tc.username == "inactive" {
				apiUser = &models.User{ID: 1, Username: "inactive", IsActive: false, IsStaff: true}
			}
			apiView := forgeapi.NewBaseViewSet(newEntrypointSerializer, &sqliteTargetWriter{db: sqlDB, limit: 100}, &entrypointTarget{})
			apiView.Authentication = []authentication.Authentication{apiAuth{user: apiUser}}
			apiView.Permissions = []permissions.Permission{permissions.NewIsAuthenticated(), permissions.NewIsStaffUser()}
			apiRouter := forgeapi.NewRouter("/api")
			apiRouter.Register("entrypoint_targets", apiView)
			apiMux := forgehttp.NewRouter()
			apiRouter.RegisterRoutes(apiMux)
			apiReq := httptest.NewRequest(http.MethodPost, "/api/entrypoint_targets/", strings.NewReader(`{"value":"changed"}`))
			apiRec := httptest.NewRecorder()
			apiMux.ServeHTTP(apiRec, apiReq)
			require.Equal(t, tc.apiCode, apiRec.Code, apiRec.Body.String())
			assertTarget("API viewset")

			t.Setenv("FORGE_DATABASE_DRIVER", "sqlite")
			t.Setenv("FORGE_DATABASE_NAME", dsn)
			registry := core.GetRegistry()
			registry.RegisterCommand("mutate", &sqliteBusinessCommand{db: sqlDB})
			root := registry.BuildRootCommand()
			root.SetArgs([]string{"--as", tc.username, "mutate"})
			require.ErrorContains(t, root.Execute(), "resolve CLI principal")
			assertTarget("CLI --as")
		})
	}
}

func createEntrypointSchema(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE entrypoint_targets (id INTEGER PRIMARY KEY, value TEXT NOT NULL);
CREATE TABLE users (
	id INTEGER PRIMARY KEY, username TEXT NOT NULL, email TEXT NOT NULL, password TEXT NOT NULL,
		first_name TEXT, last_name TEXT, is_active BOOLEAN NOT NULL DEFAULT 1, is_staff BOOLEAN NOT NULL DEFAULT 0,
		is_superuser BOOLEAN NOT NULL DEFAULT 0, is_locked BOOLEAN NOT NULL DEFAULT 0,
		email_verified BOOLEAN NOT NULL DEFAULT 0, phone_number TEXT, phone_verified BOOLEAN NOT NULL DEFAULT 0,
		timezone TEXT, locale TEXT, language TEXT,
	bio TEXT, website TEXT, location TEXT, avatar TEXT, password_changed_at TIMESTAMP, password_expires_at TIMESTAMP,
		must_change_password BOOLEAN NOT NULL DEFAULT 0, locked_at TIMESTAMP, locked_reason TEXT,
		failed_login_attempts INTEGER NOT NULL DEFAULT 0,
		last_failed_login_at TIMESTAMP, email_verified_at TIMESTAMP,
		date_joined TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, last_login TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP)`)
	return err
}

func mustForgeDB(t *testing.T, dsn string) *db.DB {
	t.Helper()
	database, err := db.NewDBWithDriver("sqlite", dsn)
	require.NoError(t, err)
	return database
}
