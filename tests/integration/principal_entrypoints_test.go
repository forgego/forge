package integration_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
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

type principalTestCase struct {
	name       string
	username   string
	wantDenied bool
}

type entrypointTarget struct {
	ID    int64  `json:"id" db:"id"`
	Value string `json:"value" db:"value"`
}

type entrypointSerializer struct{ *forgeapi.BaseSerializer }

func newEntrypointSerializer() forgeapi.Serializer {
	return &entrypointSerializer{BaseSerializer: forgeapi.NewBaseSerializer(map[string]interface{}{})}
}
func (*entrypointSerializer) Fields() []string { return []string{"id", "value"} }

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
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (w sqliteTargetWriter) Create(_ context.Context, object interface{}) error {
	target := object.(*entrypointTarget)
	_, err := w.db.Exec("UPDATE entrypoint_targets SET value = ? WHERE id = 1", target.Value)
	return err
}

func (w sqliteTargetWriter) Update(_ context.Context, target *entrypointTarget) error {
	_, err := w.db.Exec("UPDATE entrypoint_targets SET value = ? WHERE id = ?", target.Value, target.ID)
	return err
}

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
func (a *targetAdmin) GetObject(_ context.Context, id interface{}) (interface{}, error) {
	var target entrypointTarget
	err := a.db.QueryRow("SELECT id, value FROM entrypoint_targets WHERE id = ?", id).Scan(&target.ID, &target.Value)
	if err != nil {
		return nil, err
	}
	return &target, nil
}
func (a *targetAdmin) CreateObject(_ context.Context, data map[string]interface{}) (interface{}, error) {
	_, err := a.db.Exec("UPDATE entrypoint_targets SET value = ? WHERE id = 1", data["value"])
	return data, err
}
func (a *targetAdmin) UpdateObject(_ context.Context, id interface{}, data map[string]interface{}) (interface{}, error) {
	_, err := a.db.Exec("UPDATE entrypoint_targets SET value = ? WHERE id = ?", data["value"], id)
	return data, err
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
	_, err := c.db.Exec("UPDATE entrypoint_targets SET value = 'updated' WHERE id = 1")
	return err
}

type entrypointEnv struct {
	db  *sql.DB
	dsn string
}

func setupEntrypointEnv(t *testing.T, tc principalTestCase) *entrypointEnv {
	t.Helper()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	isolatedDir := filepath.Join(t.TempDir(), "work")
	require.NoError(t, os.MkdirAll(filepath.Join(isolatedDir, "config"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(filepath.Dir(isolatedDir), "config"), 0755))
	require.NoError(t, os.Chdir(isolatedDir))
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	dsn := filepath.Join(t.TempDir(), "principals.sqlite")
	sqlDB, err := sql.Open("sqlite3", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, createEntrypointSchema(sqlDB))
	_, err = sqlDB.Exec("INSERT INTO entrypoint_targets (id, value) VALUES (1, 'original')")
	require.NoError(t, err)

	switch tc.username {
	case "inactive":
		_, err = sqlDB.Exec("INSERT INTO users (id, username, email, password, is_active, is_staff, is_superuser, is_locked) VALUES (1, 'inactive', 'inactive@example.invalid', '', 0, 1, 0, 0)")
		require.NoError(t, err)
	case "active":
		_, err = sqlDB.Exec("INSERT INTO users (id, username, email, password, is_active, is_staff, is_superuser, is_locked) VALUES (1, 'active', 'active@example.invalid', '', 1, 1, 0, 0)")
		require.NoError(t, err)
	}

	return &entrypointEnv{db: sqlDB, dsn: dsn}
}

func assertTargetRow(t *testing.T, db *sql.DB, want string, entrypoint string) {
	t.Helper()
	var got string
	require.NoError(t, db.QueryRow("SELECT value FROM entrypoint_targets WHERE id = 1").Scan(&got))
	assert.Equal(t, want, got, "%s target row must be %q", entrypoint, want)
}

func testAdminRESTEntrypoint(t *testing.T, tc principalTestCase) {
	env := setupEntrypointEnv(t, tc)

	identityDB := mustForgeDB(t, env.dsn)
	defer identityDB.Close()
	userStore := repository.NewUserRepository(identityDB)

	adminRegistry := admincore.NewRegistry()
	require.NoError(t, adminRegistry.Register(&targetAdmin{db: env.db}))

	token := "token-" + tc.username
	tokens := tokenStore{token: tc.username}

	adminRouter := adminrest.NewRouter(adminRegistry).
		SetUserRepository(userStore).
		SetStores(adminrest.Stores{Tokens: tokens})

	adminMux := chi.NewRouter()
	adminRouter.RegisterRoutes(adminMux)

	req := httptest.NewRequest(http.MethodPatch, "/api/entrypoint_targets/1", strings.NewReader(`{"value":"updated"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	adminMux.ServeHTTP(rec, req)

	if tc.wantDenied {
		require.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
		assertTargetRow(t, env.db, "original", "admin REST")
	} else {
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assertTargetRow(t, env.db, "updated", "admin REST")
	}
}

func testRESTAPIEntrypoint(t *testing.T, tc principalTestCase) {
	env := setupEntrypointEnv(t, tc)

	var apiUser *models.User
	if tc.username == "inactive" {
		apiUser = &models.User{ID: 1, Username: "inactive", IsActive: false, IsStaff: true}
	} else if tc.username == "active" {
		apiUser = &models.User{ID: 1, Username: "active", IsActive: true, IsStaff: true}
	}

	apiView := forgeapi.NewBaseViewSet(newEntrypointSerializer, &sqliteTargetWriter{db: env.db, limit: 100}, &entrypointTarget{})
	apiView.Authentication = []authentication.Authentication{apiAuth{user: apiUser}}
	apiView.Permissions = []permissions.Permission{permissions.NewIsAuthenticated(), permissions.NewIsStaffUser()}

	apiRouter := forgeapi.NewRouter("/api")
	apiRouter.Register("entrypoint_targets", apiView)
	apiMux := forgehttp.NewRouter()
	apiRouter.RegisterRoutes(apiMux)

	req := httptest.NewRequest(http.MethodPatch, "/api/entrypoint_targets/1", strings.NewReader(`{"value":"updated"}`))
	rec := httptest.NewRecorder()
	apiMux.ServeHTTP(rec, req)

	if tc.wantDenied {
		if tc.username == "inactive" {
			require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
		} else {
			require.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
		}
		assertTargetRow(t, env.db, "original", "the REST api")
	} else {
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assertTargetRow(t, env.db, "updated", "the REST api")
	}
}

func testCLIBusinessCommandEntrypoint(t *testing.T, tc principalTestCase) {
	env := setupEntrypointEnv(t, tc)

	t.Setenv("FORGE_DATABASE_DRIVER", "sqlite")
	t.Setenv("FORGE_DATABASE_NAME", env.dsn)

	businessCmd := &sqliteBusinessCommand{db: env.db}
	registry := core.GetRegistry()
	registry.RegisterCommand("mutate", businessCmd)

	root := registry.BuildRootCommand()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	var err error
	if root.PersistentFlags().Lookup("as") != nil {
		root.SetArgs([]string{"--as", tc.username, "mutate"})
		err = root.Execute()
	} else {
		cliCtx := core.NewContext()
		if tc.username != "unknown" {
			identityDB := mustForgeDB(t, env.dsn)
			defer identityDB.Close()
			userStore := repository.NewUserRepository(identityDB)
			_ = cliCtx.AssumeUser(context.Background(), userStore, tc.username)
		}
		err = businessCmd.Execute(cliCtx, nil)
	}

	if tc.wantDenied {
		require.Error(t, err)
		assertTargetRow(t, env.db, "original", "CLI business command")
	} else {
		require.NoError(t, err)
		assertTargetRow(t, env.db, "updated", "CLI business command")
	}
}

func TestPrincipalEntrypoints(t *testing.T) {
	entrypoints := []struct {
		name string
		run  func(t *testing.T, tc principalTestCase)
	}{
		{
			name: "admin REST API",
			run:  testAdminRESTEntrypoint,
		},
		{
			name: "the REST api",
			run:  testRESTAPIEntrypoint,
		},
		{
			name: "CLI business command",
			run:  testCLIBusinessCommandEntrypoint,
		},
	}

	for _, ep := range entrypoints {
		t.Run(ep.name, func(t *testing.T) {
			cases := []principalTestCase{
				{
					name:       "an inactive user principal",
					username:   "inactive",
					wantDenied: true,
				},
				{
					name:       "an unknown principal",
					username:   "unknown",
					wantDenied: true,
				},
				{
					name:       "an active authorized principal",
					username:   "active",
					wantDenied: false,
				},
			}

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					ep.run(t, tc)
				})
			}
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
