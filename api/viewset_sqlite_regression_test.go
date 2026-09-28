package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var apiSQLiteCleanCalls atomic.Int64

type apiSQLiteModel struct {
	schema.BaseSchema
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email" validate:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (apiSQLiteModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("email", schema.Required(), schema.Unique()),
		schema.DateTimeField("created_at", schema.AutoNowAdd()),
	}
}

func (*apiSQLiteModel) Clean() error {
	apiSQLiteCleanCalls.Add(1)
	return nil
}

type apiSQLiteSerializer struct{ *BaseSerializer }

func newAPISQLiteSerializer() Serializer {
	return &apiSQLiteSerializer{BaseSerializer: NewBaseSerializer(make(map[string]interface{}))}
}

func setupAPISQLiteViewSet(t *testing.T) (*orm.Manager[apiSQLiteModel], *forgehttp.Router) {
	t.Helper()
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "api.sqlite"), db.WithMaxOpenConns(1))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })
	_, err = database.Exec(`CREATE TABLE api_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)

	manager, err := orm.NewManagerWithDB[apiSQLiteModel]("api_items", database)
	require.NoError(t, err)
	vs := NewBaseViewSet(newAPISQLiteSerializer, manager, &apiSQLiteModel{})
	vs.ReadOnlyRequestFields = NonEditableFields(&apiSQLiteModel{})
	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	return manager, handler
}

func TestBaseViewSet_Create_ValidatesModelExactlyOnce(t *testing.T) {
	manager, handler := setupAPISQLiteViewSet(t)

	apiSQLiteCleanCalls.Store(0)
	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"email":"api@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, int64(1), apiSQLiteCleanCalls.Load())

	apiSQLiteCleanCalls.Store(0)
	direct := &apiSQLiteModel{Email: "direct@example.com"}
	require.NoError(t, manager.Create(context.Background(), direct))
	assert.Equal(t, int64(1), apiSQLiteCleanCalls.Load())
}

func TestBaseViewSet_Create_RefetchesDatabaseGeneratedValues(t *testing.T) {
	manager, handler := setupAPISQLiteViewSet(t)
	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"email":"fresh@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var response struct {
		ID        int64  `json:"id"`
		CreatedAt string `json:"created_at"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.NotZero(t, response.ID)
	createdAt, err := time.Parse(time.RFC3339, response.CreatedAt)
	require.NoError(t, err)
	assert.False(t, createdAt.IsZero())

	stored, err := manager.Get(context.Background(), response.ID)
	require.NoError(t, err)
	assert.True(t, stored.CreatedAt.Equal(createdAt), "response timestamp %s differs from stored timestamp %s", createdAt, stored.CreatedAt)
}

func TestBaseViewSet_Create_MapsSQLiteUniqueViolationToConflict(t *testing.T) {
	manager, handler := setupAPISQLiteViewSet(t)
	require.NoError(t, manager.Create(context.Background(), &apiSQLiteModel{Email: "duplicate@example.com"}))

	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"email":"duplicate@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
	body := strings.ToLower(rec.Body.String())
	assert.NotContains(t, body, "unique constraint")
	assert.NotContains(t, body, "api_items")
	assert.NotContains(t, body, "sqlite")
}
