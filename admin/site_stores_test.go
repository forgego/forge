package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/db"
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
	// Registered before the stores are chosen, as init() registrations are.
	_, err := RegisterWithSite(site, &core.Config[storeNote]{})
	require.NoError(t, err)
	site.SetDB(database)
	require.NoError(t, site.UseStores(context.Background(), "database"))
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
	require.NoError(t, site.UseStores(ctx, "memory"), "memory needs no database")
	require.ErrorContains(t, site.UseStores(ctx, "redis"), "invalid server.stores")

	site.SetDB(storestest.Open(t, "sqlite3", storestest.SQLitePath(t)))
	require.ErrorIs(t, site.UseStores(ctx, "database"), stores.ErrNotMigrated)
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
