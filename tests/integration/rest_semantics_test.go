package integration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	forgeapi "github.com/forgego/forge/api"
	apierrors "github.com/forgego/forge/api/errors"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/registry"
	"github.com/forgego/forge/schema"
	"github.com/forgego/forge/server"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RESTSemanticsModel is the single exposed model used to verify standard REST semantics.
type RESTSemanticsModel struct {
	schema.BaseSchema
	ID          int64  `db:"id" json:"id"`
	Title       string `db:"title" json:"title"`
	Description string `db:"description" json:"description"`
	Status      string `db:"status" json:"status"`
	Priority    int64  `db:"priority" json:"priority"`
}

func (RESTSemanticsModel) Meta() schema.Meta {
	return schema.Meta{TableName: "rest_semantics_models"}
}

func (RESTSemanticsModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
		schema.StringField("description"),
		schema.StringField("status", schema.Default("draft")),
		schema.Int64Field("priority", schema.MinValue(1), schema.MaxValue(10)),
	}
}

// Clean provides model-level validation: title cannot be "invalid_title".
func (m *RESTSemanticsModel) Clean() error {
	if m.Title == "invalid_title" {
		return errors.New("title cannot be invalid_title")
	}
	return nil
}

// restSemanticsPermission enforces permission checks using the framework's permission mechanism.
type restSemanticsPermission struct{}

func (restSemanticsPermission) HasPermission(r *http.Request, view permissions.ViewSet) bool {
	return r.Header.Get("X-Deny-Permission") != "true"
}

func (restSemanticsPermission) HasObjectPermission(r *http.Request, view permissions.ViewSet, obj interface{}) bool {
	return true
}

func (restSemanticsPermission) GetMessage() string {
	return "Permission denied"
}

func (restSemanticsPermission) GetCode() string {
	return "permission_denied"
}

type semanticsFixture struct {
	db      *db.DB
	handler http.Handler
}

func newSemanticsFixture(t *testing.T) *semanticsFixture {
	t.Helper()

	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "semantics_test.sqlite"), db.WithMaxOpenConns(1))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	createTableQuery := `CREATE TABLE rest_semantics_models (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT DEFAULT '' NOT NULL,
		status TEXT DEFAULT 'draft' NOT NULL,
		priority INTEGER DEFAULT 1 NOT NULL
	)`
	_, err = database.Exec(createTableQuery)
	require.NoError(t, err)

	seedQuery := `INSERT INTO rest_semantics_models (id, title, description, status, priority)
		VALUES (1, 'Initial Title', 'Initial Description', 'draft', 5)`
	_, err = database.Exec(seedQuery)
	require.NoError(t, err)

	_ = registry.RegisterModel(&RESTSemanticsModel{})
	manager, err := orm.NewManagerWithDB[RESTSemanticsModel]("rest_semantics_models", database)
	require.NoError(t, err)

	vs := forgeapi.NewBaseViewSet(
		func() forgeapi.Serializer { return forgeapi.NewBaseSerializer(nil) },
		manager,
		&RESTSemanticsModel{},
	)
	vs.Authentication = nil
	vs.Permissions = []permissions.Permission{restSemanticsPermission{}}

	apiRouter := forgeapi.NewRouter("/api/v1")
	apiRouter.Register("items", vs)

	mux := server.NewRouter()
	apiRouter.RegisterRoutes(mux)

	return &semanticsFixture{
		db:      database,
		handler: mux,
	}
}

func TestRestSemantics_StandardOperations(t *testing.T) {
	type testCase struct {
		name          string
		method        string
		target        string
		body          string
		headers       map[string]string
		wantStatus    int
		wantErrorCode string
		wantErrorType apierrors.ErrorType
		verify        func(t *testing.T, fix *semanticsFixture, rec *httptest.ResponseRecorder)
	}

	tests := []testCase{
		{
			name:       "PUT with full body replaces every field",
			method:     http.MethodPut,
			target:     "/api/v1/items/1",
			body:       `{"title":"Replaced Title","description":"Replaced Description","status":"active","priority":8}`,
			wantStatus: http.StatusOK,
			verify: func(t *testing.T, fix *semanticsFixture, rec *httptest.ResponseRecorder) {
				var putResp map[string]interface{}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &putResp))
				assert.Equal(t, "Replaced Title", putResp["title"])
				assert.Equal(t, "Replaced Description", putResp["description"])
				assert.Equal(t, "active", putResp["status"])
				assert.Equal(t, float64(8), putResp["priority"])

				// Re-GET to verify all fields were replaced in storage
				getRec := httptest.NewRecorder()
				getReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/1", nil)
				fix.handler.ServeHTTP(getRec, getReq)
				require.Equal(t, http.StatusOK, getRec.Code)

				var getResp map[string]interface{}
				require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &getResp))
				assert.Equal(t, "Replaced Title", getResp["title"])
				assert.Equal(t, "Replaced Description", getResp["description"])
				assert.Equal(t, "active", getResp["status"])
				assert.Equal(t, float64(8), getResp["priority"])
			},
		},
		{
			name:          "PUT missing required field returns 400",
			method:        http.MethodPut,
			target:        "/api/v1/items/1",
			body:          `{"description":"Missing Title","status":"active","priority":8}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: apierrors.CodeValidationError,
			wantErrorType: apierrors.ErrorTypeValidation,
			verify: func(t *testing.T, fix *semanticsFixture, _ *httptest.ResponseRecorder) {
				// Re-GET to verify stored row was not modified
				getRec := httptest.NewRecorder()
				getReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/1", nil)
				fix.handler.ServeHTTP(getRec, getReq)
				require.Equal(t, http.StatusOK, getRec.Code)

				var getResp map[string]interface{}
				require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &getResp))
				assert.Equal(t, "Initial Title", getResp["title"])
			},
		},
		{
			name:       "PATCH with partial body changes only given fields and keeps other fields",
			method:     http.MethodPatch,
			target:     "/api/v1/items/1",
			body:       `{"title":"Patched Title"}`,
			wantStatus: http.StatusOK,
			verify: func(t *testing.T, fix *semanticsFixture, rec *httptest.ResponseRecorder) {
				var patchResp map[string]interface{}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &patchResp))
				assert.Equal(t, "Patched Title", patchResp["title"])

				// Re-GET to verify given field changed and other fields keep their values
				getRec := httptest.NewRecorder()
				getReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/1", nil)
				fix.handler.ServeHTTP(getRec, getReq)
				require.Equal(t, http.StatusOK, getRec.Code)

				var getResp map[string]interface{}
				require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &getResp))
				assert.Equal(t, "Patched Title", getResp["title"])
				assert.Equal(t, "Initial Description", getResp["description"])
				assert.Equal(t, "draft", getResp["status"])
				assert.Equal(t, float64(5), getResp["priority"])
			},
		},
		{
			name:       "DELETE returns 204 and subsequent GET returns 404",
			method:     http.MethodDelete,
			target:     "/api/v1/items/1",
			wantStatus: http.StatusNoContent,
			verify: func(t *testing.T, fix *semanticsFixture, _ *httptest.ResponseRecorder) {
				// Re-GET to verify resource is gone and returns 404 with stable error envelope
				getRec := httptest.NewRecorder()
				getReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/1", nil)
				fix.handler.ServeHTTP(getRec, getReq)
				require.Equal(t, http.StatusNotFound, getRec.Code)

				var problem apierrors.Problem
				require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &problem))
				assert.Equal(t, http.StatusNotFound, problem.Status)
				assert.Equal(t, apierrors.CodeNotFound, problem.Code)
				assert.Contains(t, problem.Type, string(apierrors.ErrorTypeNotFound))
			},
		},
		{
			name:          "GET of missing id returns 404",
			method:        http.MethodGet,
			target:        "/api/v1/items/9999",
			wantStatus:    http.StatusNotFound,
			wantErrorCode: apierrors.CodeNotFound,
			wantErrorType: apierrors.ErrorTypeNotFound,
		},
		{
			name:          "PUT of missing id returns 404",
			method:        http.MethodPut,
			target:        "/api/v1/items/9999",
			body:          `{"title":"Replaced Title","description":"Description","status":"active","priority":5}`,
			wantStatus:    http.StatusNotFound,
			wantErrorCode: apierrors.CodeNotFound,
			wantErrorType: apierrors.ErrorTypeNotFound,
		},
		{
			name:          "PATCH of missing id returns 404",
			method:        http.MethodPatch,
			target:        "/api/v1/items/9999",
			body:          `{"title":"Patched Title"}`,
			wantStatus:    http.StatusNotFound,
			wantErrorCode: apierrors.CodeNotFound,
			wantErrorType: apierrors.ErrorTypeNotFound,
		},
		{
			name:          "DELETE of missing id returns 404",
			method:        http.MethodDelete,
			target:        "/api/v1/items/9999",
			wantStatus:    http.StatusNotFound,
			wantErrorCode: apierrors.CodeNotFound,
			wantErrorType: apierrors.ErrorTypeNotFound,
		},
		{
			name:          "Invalid value failing schema numeric constraint returns 400",
			method:        http.MethodPatch,
			target:        "/api/v1/items/1",
			body:          `{"priority":99}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: apierrors.CodeValidationError,
			wantErrorType: apierrors.ErrorTypeValidation,
		},
		{
			name:          "Invalid value failing model Clean validation returns 400",
			method:        http.MethodPatch,
			target:        "/api/v1/items/1",
			body:          `{"title":"invalid_title"}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: apierrors.CodeValidationError,
			wantErrorType: apierrors.ErrorTypeValidation,
		},
		{
			name:          "Permission-denied request returns 403",
			method:        http.MethodGet,
			target:        "/api/v1/items/1",
			headers:       map[string]string{"X-Deny-Permission": "true"},
			wantStatus:    http.StatusForbidden,
			wantErrorCode: apierrors.CodePermissionDenied,
			wantErrorType: apierrors.ErrorTypeAuthorization,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fix := newSemanticsFixture(t)

			var bodyReader *bytes.Reader
			if tc.body != "" {
				bodyReader = bytes.NewReader([]byte(tc.body))
			} else {
				bodyReader = bytes.NewReader(nil)
			}

			req := httptest.NewRequest(tc.method, tc.target, bodyReader)
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			fix.handler.ServeHTTP(rec, req)

			require.Equal(t, tc.wantStatus, rec.Code, "Response body: %s", rec.Body.String())

			// Assert stable error envelope when error code / type is expected
			if tc.wantErrorCode != "" || tc.wantErrorType != "" {
				var problem apierrors.Problem
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &problem), "Must decode RFC 7807 Problem: %s", rec.Body.String())
				assert.Equal(t, tc.wantStatus, problem.Status)
				if tc.wantErrorCode != "" {
					assert.Equal(t, tc.wantErrorCode, problem.Code, "Problem code must match")
				}
				if tc.wantErrorType != "" {
					assert.Contains(t, problem.Type, string(tc.wantErrorType), "Problem type must contain expected type URI fragment")
				}
			}

			if tc.verify != nil {
				tc.verify(t, fix, rec)
			}
		})
	}
}
