package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/forgego/forge/admin/api/rest"
	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/schema"
	"github.com/forgego/forge/stores"
	"github.com/forgego/forge/stores/storestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type storeNote struct {
	schema.BaseSchema
	ID    int64  `db:"id" json:"id"`
	Title string `db:"title" json:"title"`
}

type siteTestUsers struct{}

func (siteTestUsers) GetByUsername(_ context.Context, username string) (*models.User, error) {
	if username != "admin" {
		return nil, repository.ErrUserNotFound
	}
	return &models.User{ID: 1, Username: username, IsActive: true, IsStaff: true, IsSuperuser: true}, nil
}

type unavailableSiteUsers struct{}

func (unavailableSiteUsers) GetByUsername(context.Context, string) (*models.User, error) {
	return nil, errors.New("identity table is unavailable")
}

type inactiveSiteUsers struct{}

func (inactiveSiteUsers) GetByUsername(_ context.Context, username string) (*models.User, error) {
	return &models.User{ID: 1, Username: username, IsStaff: true, IsSuperuser: true}, nil
}

func (storeNote) Meta() schema.Meta { return schema.Meta{TableName: "store_notes"} }

func (storeNote) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
	}
}

func createNotesTable(t *testing.T, database *db.DB) {
	t.Helper()
	ddl := "CREATE TABLE store_notes (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL)"
	if database.Driver == "postgres" {
		ddl = "CREATE TABLE store_notes (id BIGSERIAL PRIMARY KEY, title TEXT NOT NULL)"
	}
	_, err := database.Exec(ddl)
	require.NoError(t, err)
}

// storeSite is one admin instance: a site on its own connection with the
// database stores, serving over HTTP.
func storeSite(t *testing.T, database *db.DB) *httptest.Server {
	t.Helper()
	site := NewSite("stores")
	site.SetUserStore(siteTestUsers{})
	// Registered before the stores are chosen, as init() registrations are.
	_, err := RegisterWithSite(site, &core.Config[storeNote]{})
	require.NoError(t, err)
	site.SetDB(database)
	require.NoError(t, site.UseDatabaseStores(context.Background()))
	ts := httptest.NewServer(site.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func adminCall(t *testing.T, base, method, path, token string, body any) (int, map[string]any) {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, base+path, reader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	out := map[string]any{}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func TestSite_DatabaseStoresSharedBetweenInstances(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "secret")
	for _, backend := range storestest.Backends() {
		t.Run(backend.Name, func(t *testing.T) {
			first, dsn := backend.OpenMigrated(t)
			createNotesTable(t, first)
			a := storeSite(t, first)
			b := storeSite(t, storestest.Open(t, backend.Driver, dsn))

			status, body := adminCall(t, a.URL, http.MethodPost, "/api/login", "", map[string]string{"username": "admin", "password": "secret"})
			require.Equal(t, http.StatusOK, status, "%v", body)
			token := body["token"].(string)

			status, _ = adminCall(t, b.URL, http.MethodGet, "/api/config", token, nil)
			assert.Equal(t, http.StatusOK, status, "a token issued on A authenticates on B")

			status, body = adminCall(t, a.URL, http.MethodPost, "/api/store_notes/", token, map[string]string{"title": "Shared"})
			require.Equal(t, http.StatusCreated, status, "%v", body)
			id := fmt.Sprint(body["id"])
			status, body = adminCall(t, b.URL, http.MethodGet, "/api/store_notes/"+id+"/history", token, nil)
			require.Equal(t, http.StatusOK, status, "%v", body)
			assert.Contains(t, fmt.Sprint(body), "add", "history written on A is listed on B")
			assert.Contains(t, fmt.Sprint(body), "admin", "the admin user is recorded by name")

			status, _ = adminCall(t, a.URL, http.MethodPost, "/api/saved-views/store_notes/", token, map[string]any{"name": "Mine", "ordering": []string{"-id"}})
			require.Equal(t, http.StatusCreated, status)
			status, body = adminCall(t, b.URL, http.MethodGet, "/api/saved-views/store_notes/", token, nil)
			require.Equal(t, http.StatusOK, status)
			assert.Contains(t, fmt.Sprint(body["views"]), "Mine", "a view saved on A is listed on B")

			status, _ = adminCall(t, b.URL, http.MethodPost, "/api/logout", token, nil)
			require.Equal(t, http.StatusOK, status)
			status, _ = adminCall(t, a.URL, http.MethodGet, "/api/config", token, nil)
			assert.Equal(t, http.StatusUnauthorized, status, "logging out on B revokes the token on A")
		})
	}
}

func TestSite_UseDatabaseStoresRequiresMigratedDatabase(t *testing.T) {
	ctx := context.Background()
	site := NewSite("stores")
	require.ErrorContains(t, site.UseDatabaseStores(ctx), "call SetDB first")

	site.SetDB(storestest.Open(t, "sqlite3", storestest.SQLitePath(t)))
	require.ErrorIs(t, site.UseDatabaseStores(ctx), stores.ErrNotMigrated)
}

func TestSite_ConfiguredCredentialPrincipalUsesLoginIdentitySource(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "configured-admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "secret")

	for _, tc := range []struct {
		name  string
		users rest.AdminUserStore
		want  int
	}{
		{name: "identity store unavailable", users: unavailableSiteUsers{}, want: http.StatusOK},
		{name: "inactive identity row", users: inactiveSiteUsers{}, want: http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			site := NewSite("configured")
			site.SetDB(storestest.OpenMemory(t))
			site.SetUserStore(tc.users)
			server := httptest.NewServer(site.Handler())
			defer server.Close()

			status, body := adminCall(t, server.URL, http.MethodPost, "/api/login", "", map[string]string{
				"username": "configured-admin", "password": "secret",
			})
			require.Equal(t, http.StatusOK, status, "%v", body)
			status, body = adminCall(t, server.URL, http.MethodGet, "/api/config", body["token"].(string), nil)
			assert.Equal(t, tc.want, status, "%v", body)
		})
	}
}

type recordingHistory struct{ entries []core.LogEntry }

func (h *recordingHistory) LogAction(_ context.Context, entry core.LogEntry) error {
	h.entries = append(h.entries, entry)
	return nil
}

func (h *recordingHistory) GetHistory(context.Context, string, string) ([]core.LogEntry, error) {
	return h.entries, nil
}

func TestSite_SharedHistoryKeepsExplicitHistoryManagers(t *testing.T) {
	ctx := context.Background()
	shared := &recordingHistory{}
	own := &recordingHistory{}
	legacy := NewHistoryManager("title")

	defaulted, err := core.NewAdmin[storeNote](storeNote{}, nil, &core.Config[storeNote]{})
	require.NoError(t, err)
	explicit, err := core.NewAdmin[storeNote](storeNote{}, nil, &core.Config[storeNote]{HistoryManager: own})
	require.NoError(t, err)
	compat, err := core.NewAdmin[storeNote](storeNote{}, nil, &core.Config[storeNote]{HistoryManager: legacy})
	require.NoError(t, err)

	for _, a := range []*core.Admin[storeNote]{defaulted, explicit, compat} {
		a.UseSiteHistory(shared)
	}
	user := map[string]interface{}{"username": "alice", "role": "superuser"}
	require.NoError(t, defaulted.LogAction(ctx, user, "1", "one", core.ActionAdd, ""))
	require.NoError(t, explicit.LogAction(ctx, user, "2", "two", core.ActionAdd, ""))
	require.NoError(t, compat.LogAction(ctx, user, "3", "three", core.ActionAdd, ""))

	require.Len(t, shared.entries, 2, "the default and admin.HistoryManager write to the shared history")
	assert.Equal(t, "1", shared.entries[0].ObjectID)
	assert.Equal(t, "3", shared.entries[1].ObjectID)
	assert.Equal(t, "alice", shared.entries[0].UserID, "a map user is recorded by username")
	require.Len(t, own.entries, 1, "a config's own HistoryManager is kept")
}

func TestSite_SetDBAppliesDatabaseStoresAutomatically(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "secret")

	db := storestest.OpenMemory(t)
	createNotesTable(t, db)

	site := NewSite("autostores")
	site.SetUserStore(siteTestUsers{})
	_, err := RegisterWithSite(site, &core.Config[storeNote]{})
	require.NoError(t, err)

	// Apps only call SetDB; UseDatabaseStores is applied automatically.
	site.SetDB(db)

	server := httptest.NewServer(site.Handler())
	defer server.Close()

	status, body := adminCall(t, server.URL, http.MethodPost, "/api/login", "", map[string]string{
		"username": "admin", "password": "secret",
	})
	require.Equal(t, http.StatusOK, status, "%v", body)
	token := body["token"].(string)

	status, _ = adminCall(t, server.URL, http.MethodGet, "/api/config", token, nil)
	assert.Equal(t, http.StatusOK, status)
}

func TestSite_SetDBUnmigratedDatabaseReturns503(t *testing.T) {
	database, err := db.NewDBWithDriver("sqlite3", ":memory:", db.WithMaxOpenConns(1), db.WithMaxIdleConns(1))
	require.NoError(t, err)
	defer database.Close()

	site := NewSite("unmigrated")
	site.SetDB(database)
	require.Error(t, site.storesErr)

	server := httptest.NewServer(site.Handler())
	defer server.Close()

	// Every request gets 503
	status, body := adminCall(t, server.URL, http.MethodGet, "/api/config", "", nil)
	assert.Equal(t, http.StatusServiceUnavailable, status)
	assert.Equal(t, "admin stores unavailable: run forge migrate up", body["error"])

	status, body = adminCall(t, server.URL, http.MethodPost, "/api/login", "", map[string]string{"username": "admin", "password": "x"})
	assert.Equal(t, http.StatusServiceUnavailable, status)
	assert.Equal(t, "admin stores unavailable: run forge migrate up", body["error"])

	// Clear storesErr when a later UseDatabaseStores succeeds
	storestest.Migrate(t, database)
	require.NoError(t, site.UseDatabaseStores(context.Background()))
	assert.NoError(t, site.storesErr)

	// Handler now succeeds and doesn't return 503
	recoveredServer := httptest.NewServer(site.Handler())
	defer recoveredServer.Close()
	status, body = adminCall(t, recoveredServer.URL, http.MethodGet, "/api/config", "", nil)
	assert.Equal(t, http.StatusUnauthorized, status)
	assert.NotEqual(t, "admin stores unavailable: run forge migrate up", body["error"])
}

func TestSite_HandlerLogsStoresErrorOnce(t *testing.T) {
	var buf bytes.Buffer
	prevWriter := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prevWriter)

	database, err := db.NewDBWithDriver("sqlite3", ":memory:", db.WithMaxOpenConns(1), db.WithMaxIdleConns(1))
	require.NoError(t, err)
	defer database.Close()

	site := NewSite("unmigrated_log")
	site.SetDB(database)

	handler := site.Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req)
	assert.Equal(t, http.StatusServiceUnavailable, rec1.Code)

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)
	assert.Equal(t, http.StatusServiceUnavailable, rec2.Code)

	logged := buf.String()
	assert.Contains(t, logged, "forge/admin: stores unavailable:")
	assert.Equal(t, 1, strings.Count(logged, "forge/admin: stores unavailable:"))
}
