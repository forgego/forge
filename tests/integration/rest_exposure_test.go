package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	forgeapi "github.com/forgego/forge/api"
	"github.com/forgego/forge/api/docs"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/registry"
	"github.com/forgego/forge/schema"
	"github.com/forgego/forge/server"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Task 1 Models ---

type ExposureRegisteredModel struct {
	schema.BaseSchema
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func (ExposureRegisteredModel) Meta() schema.Meta {
	return schema.Meta{TableName: "registered_models"}
}

func (ExposureRegisteredModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

type ExposureUnregisteredModel struct {
	schema.BaseSchema
	ID          int64  `db:"id" json:"id"`
	SecretToken string `db:"secret_token" json:"secret_token"`
}

func (ExposureUnregisteredModel) Meta() schema.Meta {
	return schema.Meta{TableName: "unexposed_models"}
}

func (ExposureUnregisteredModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("secret_token"),
	}
}

// TestRESTExposure_UnregisteredModel_NoRouteOrDocs tests that a model registered with the
// ORM/schema but NOT registered on the api.Router has no HTTP route (404 on its would-be path)
// and no entry in router.Routes() nor in the OpenAPI paths from docs.NewOpenAPIGenerator.
func TestRESTExposure_UnregisteredModel_NoRouteOrDocs(t *testing.T) {
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "exposure_test.sqlite"), db.WithMaxOpenConns(1))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	for _, query := range []string{
		`CREATE TABLE registered_models (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL)`,
		`CREATE TABLE unexposed_models (id INTEGER PRIMARY KEY AUTOINCREMENT, secret_token TEXT NOT NULL)`,
		`INSERT INTO registered_models (id, name) VALUES (1, 'item-one')`,
		`INSERT INTO unexposed_models (id, secret_token) VALUES (1, 'secret-xyz')`,
	} {
		_, err = database.Exec(query)
		require.NoError(t, err)
	}

	// Register BOTH models with the framework registry and ORM managers
	_ = registry.RegisterModel(&ExposureRegisteredModel{})
	_ = registry.RegisterModel(&ExposureUnregisteredModel{})

	regMgr, err := orm.NewManagerWithDB[ExposureRegisteredModel]("registered_models", database)
	require.NoError(t, err)

	unregMgr, err := orm.NewManagerWithDB[ExposureUnregisteredModel]("unexposed_models", database)
	require.NoError(t, err)

	// Verify the unregistered model is fully operable via ORM
	unregItem, err := unregMgr.Get(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "secret-xyz", unregItem.SecretToken)

	// Register ONLY the registered model on the API router
	apiRouter := forgeapi.NewRouter("/api/v1")
	vs := forgeapi.NewBaseViewSet(
		func() forgeapi.Serializer { return forgeapi.NewBaseSerializer(nil) },
		regMgr,
		&ExposureRegisteredModel{},
	)
	vs.Authentication = nil
	vs.Permissions = nil
	apiRouter.Register("items", vs)

	mux := server.NewRouter()
	apiRouter.RegisterRoutes(mux)

	// 1. HTTP routes: Registered model answers 200; unregistered model returns 404 on would-be paths
	t.Run("registered route answers 200", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items/1", nil))
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "item-one")
	})

	t.Run("unregistered model has no HTTP route", func(t *testing.T) {
		for _, target := range []string{
			"/api/v1/unexposed/1",
			"/api/v1/unexposed/",
			"/api/v1/unexposed_models/1",
			"/api/v1/unexposed_models/",
			"/api/v1/secrets/1",
		} {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
			require.Equal(t, http.StatusNotFound, rec.Code, "target %s must return 404", target)
		}

		// Also verify POST to would-be path returns 404
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/unexposed_models/", bytes.NewBufferString(`{"secret_token":"leak"}`)))
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	// 2. router.Routes(): Contains no entry for unexposed resource
	t.Run("router routes exclude unregistered model", func(t *testing.T) {
		routes := apiRouter.Routes()
		require.NotEmpty(t, routes)
		for _, r := range routes {
			assert.NotEqual(t, "unexposed", r.Resource)
			assert.NotEqual(t, "unexposed_models", r.Resource)
			assert.NotContains(t, r.Path, "unexposed")
			assert.NotContains(t, r.Path, "secret")
		}
	})

	// 3. OpenAPI paths: Contains no paths for unregistered model
	t.Run("OpenAPI paths exclude unregistered model", func(t *testing.T) {
		gen := docs.NewOpenAPIGenerator("Exposure Test API", "1.0.0").WithRoutes(forgeapi.DocsRoutes(apiRouter.Routes()))
		spec := gen.Generate()
		require.NotEmpty(t, spec.Paths)

		for path := range spec.Paths {
			assert.NotContains(t, path, "unexposed")
			assert.NotContains(t, path, "secret")
		}
		assert.Contains(t, spec.Paths, "/api/v1/items/")
		assert.Contains(t, spec.Paths, "/api/v1/items/{id}")
	})
}

// --- Task 2 Models & Serializers ---

type DistinctExposureModel struct {
	schema.BaseSchema
	ID           int64  `db:"id" json:"id"`
	Title        string `db:"title" json:"title"`
	CreateOnly   string `db:"create_only" json:"create_only"`
	OpOnly       string `db:"op_only" json:"op_only"`
	SecretHidden string `db:"secret_hidden" json:"secret_hidden"`
}

func (DistinctExposureModel) Meta() schema.Meta {
	return schema.Meta{
		TableName:           "distinct_exposure_models",
		OperationOnlyFields: []string{"op_only"},
	}
}

func (DistinctExposureModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title"),
		schema.StringField("create_only"),
		schema.StringField("op_only", schema.OperationOnly()),
		schema.StringField("secret_hidden"),
	}
}

type distinctExposureSerializer struct {
	*forgeapi.BaseSerializer
	getAction func() string
}

func (s *distinctExposureSerializer) Fields() []string {
	// Response allowlist: secret_hidden is absent!
	return []string{"id", "title", "create_only", "op_only"}
}

func (s *distinctExposureSerializer) ReadOnlyFields() []string {
	// On update / partial_update, create_only is read-only (ignored on write)
	if s.getAction != nil && (s.getAction() == "update" || s.getAction() == "partial_update") {
		return []string{"create_only"}
	}
	return nil
}

// TestRESTExposure_DistinctCreateUpdateResponseAllowlists tests that create allowlist,
// update allowlist and response allowlist are enforced separately:
//   - a field allowed on create but not update is accepted on POST and rejected or ignored on PATCH
//     (asserting stored value unchanged);
//   - a field absent from the response allowlist never appears in responses.
func TestRESTExposure_DistinctCreateUpdateResponseAllowlists(t *testing.T) {
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "distinct_test.sqlite"), db.WithMaxOpenConns(1))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	createTableQuery := `CREATE TABLE distinct_exposure_models (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		create_only TEXT DEFAULT '' NOT NULL,
		op_only TEXT DEFAULT '' NOT NULL,
		secret_hidden TEXT DEFAULT '' NOT NULL
	)`
	_, err = database.Exec(createTableQuery)
	require.NoError(t, err)

	_ = registry.RegisterModel(&DistinctExposureModel{})
	manager, err := orm.NewManagerWithDB[DistinctExposureModel]("distinct_exposure_models", database)
	require.NoError(t, err)

	var vs *forgeapi.BaseViewSet
	vs = forgeapi.NewBaseViewSet(
		func() forgeapi.Serializer {
			return &distinctExposureSerializer{
				BaseSerializer: forgeapi.NewBaseSerializer(nil),
				getAction: func() string {
					if vs == nil {
						return ""
					}
					return vs.GetAction()
				},
			}
		},
		manager,
		&DistinctExposureModel{},
	)
	vs.Authentication = nil
	vs.Permissions = nil

	apiRouter := forgeapi.NewRouter("/api/v1")
	apiRouter.Register("contracts", vs)
	mux := server.NewRouter()
	apiRouter.RegisterRoutes(mux)

	// Step 1: POST accepts create_only and op_only (allowed on create), and response allowlist hides secret_hidden
	t.Run("POST accepts fields and enforces response allowlist", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":         "Initial Contract",
			"create_only":   "created_v1",
			"op_only":       "op_v1",
			"secret_hidden": "unexposed_secret_value",
		}
		bodyBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/contracts/", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code, "POST must succeed: %s", rec.Body.String())

		var respBody map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
		assert.Equal(t, "Initial Contract", respBody["title"])
		assert.Equal(t, "created_v1", respBody["create_only"])
		assert.Equal(t, "op_v1", respBody["op_only"])
		// Field absent from response allowlist must not appear in response
		assert.NotContains(t, respBody, "secret_hidden")

		// Verify stored directly in DB: all values including secret_hidden were stored
		var title, createOnly, opOnly, secretHidden string
		err = database.QueryRow(`SELECT title, create_only, op_only, secret_hidden FROM distinct_exposure_models WHERE id = 1`).
			Scan(&title, &createOnly, &opOnly, &secretHidden)
		require.NoError(t, err)
		assert.Equal(t, "Initial Contract", title)
		assert.Equal(t, "created_v1", createOnly)
		assert.Equal(t, "op_v1", opOnly)
		assert.Equal(t, "unexposed_secret_value", secretHidden)
	})

	// Step 2: PATCH of create_only field is ignored (not allowed on update), stored value remains unchanged
	t.Run("PATCH ignores create_only field and leaves stored value unchanged", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Updated Contract Title",
			"create_only": "tampered_create_v2",
		}
		bodyBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/contracts/1", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, "PATCH must succeed: %s", rec.Body.String())

		var respBody map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
		assert.Equal(t, "Updated Contract Title", respBody["title"])
		assert.Equal(t, "created_v1", respBody["create_only"])
		assert.NotContains(t, respBody, "secret_hidden")

		// Verify straight from DB that create_only is UNCHANGED
		var currentTitle, currentCreateOnly string
		err = database.QueryRow(`SELECT title, create_only FROM distinct_exposure_models WHERE id = 1`).
			Scan(&currentTitle, &currentCreateOnly)
		require.NoError(t, err)
		assert.Equal(t, "Updated Contract Title", currentTitle)
		assert.Equal(t, "created_v1", currentCreateOnly, "create_only must remain unchanged in DB")
	})

	// Step 3: PATCH of op_only field is rejected with 403 Forbidden, stored value remains unchanged
	t.Run("PATCH rejects op_only field and leaves stored value unchanged", func(t *testing.T) {
		payload := map[string]interface{}{
			"op_only": "tampered_op_v2",
		}
		bodyBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/contracts/1", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code, "PATCH of op_only must be rejected with 403")
		assert.Contains(t, rec.Body.String(), "operation_only_field")

		// Verify straight from DB that op_only is UNCHANGED
		var currentOpOnly string
		err = database.QueryRow(`SELECT op_only FROM distinct_exposure_models WHERE id = 1`).Scan(&currentOpOnly)
		require.NoError(t, err)
		assert.Equal(t, "op_v1", currentOpOnly, "op_only must remain unchanged in DB")
	})

	// Step 4: GET responses (detail and list) never contain fields absent from the response allowlist
	t.Run("GET responses never contain fields absent from response allowlist", func(t *testing.T) {
		// Detail GET
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/contracts/1", nil))
		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "secret_hidden")
		assert.NotContains(t, rec.Body.String(), "unexposed_secret_value")

		// List GET
		recList := httptest.NewRecorder()
		mux.ServeHTTP(recList, httptest.NewRequest(http.MethodGet, "/api/v1/contracts/", nil))
		require.Equal(t, http.StatusOK, recList.Code)
		assert.NotContains(t, recList.Body.String(), "secret_hidden")
		assert.NotContains(t, recList.Body.String(), "unexposed_secret_value")
	})
}

// --- Task 3 Models ---

type FieldDenialTestModel struct {
	schema.BaseSchema
	ID                 int64  `db:"id" json:"id"`
	PublicName         string `db:"public_name" json:"public_name"`
	WriteOnlyPassword  string `db:"write_only_password" json:"write_only_password"`
	ReadOnlyAuditNote  string `db:"read_only_audit_note" json:"read_only_audit_note"`
	OperationProtected string `db:"operation_protected" json:"operation_protected"`
	HiddenConfig       string `db:"hidden_config" json:"hidden_config"`
}

func (FieldDenialTestModel) Meta() schema.Meta {
	return schema.Meta{
		TableName:           "field_denial_models",
		OperationOnlyFields: []string{"operation_protected"},
	}
}

func (FieldDenialTestModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("public_name"),
		schema.StringField("write_only_password", schema.WriteOnly()),
		schema.StringField("read_only_audit_note", schema.Editable(false)),
		schema.StringField("operation_protected", schema.OperationOnly()),
		schema.StringField("hidden_config", schema.Editable(false), schema.Serialize(false)),
	}
}

// TestRESTExposure_FieldLevelReadAndWriteDenials tests:
//   - Field-level read denial: denied field missing from GET body (retrieve and list);
//   - Field-level write denial: write to a denied field fails (or is ignored per current documented
//     behaviour in docs/) and the stored value read straight from the DB is unchanged after each denied write.
func TestRESTExposure_FieldLevelReadAndWriteDenials(t *testing.T) {
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "denials_test.sqlite"), db.WithMaxOpenConns(1))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	createTableQuery := `CREATE TABLE field_denial_models (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_name TEXT NOT NULL,
		write_only_password TEXT DEFAULT '' NOT NULL,
		read_only_audit_note TEXT DEFAULT '' NOT NULL,
		operation_protected TEXT DEFAULT '' NOT NULL,
		hidden_config TEXT DEFAULT '' NOT NULL
	)`
	_, err = database.Exec(createTableQuery)
	require.NoError(t, err)

	seedQuery := `INSERT INTO field_denial_models (id, public_name, write_only_password, read_only_audit_note, operation_protected, hidden_config)
		VALUES (1, 'initial_public', 'super_secret_pw', 'audit_fixed_v1', 'protected_state', 'internal_sys_config')`
	_, err = database.Exec(seedQuery)
	require.NoError(t, err)

	_ = registry.RegisterModel(&FieldDenialTestModel{})
	manager, err := orm.NewManagerWithDB[FieldDenialTestModel]("field_denial_models", database)
	require.NoError(t, err)

	vs := forgeapi.NewBaseViewSet(
		func() forgeapi.Serializer { return forgeapi.NewBaseSerializer(nil) },
		manager,
		&FieldDenialTestModel{},
	)
	vs.ExcludeResponseFields = forgeapi.NonSerializableFields(&FieldDenialTestModel{})
	vs.ReadOnlyRequestFields = forgeapi.NonEditableFields(&FieldDenialTestModel{})
	vs.RejectUnknownRequestFields = true
	vs.Authentication = nil
	vs.Permissions = nil

	apiRouter := forgeapi.NewRouter("/api/v1")
	apiRouter.Register("denials", vs)
	mux := server.NewRouter()
	apiRouter.RegisterRoutes(mux)

	// 1. Field-level read denial: Denied fields are missing from GET body
	t.Run("field-level read denial: denied fields missing from GET detail and list", func(t *testing.T) {
		// Detail GET
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/denials/1", nil))
		require.Equal(t, http.StatusOK, rec.Code)

		var detailResp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &detailResp))

		assert.Contains(t, detailResp, "public_name")
		assert.Contains(t, detailResp, "read_only_audit_note")
		assert.Contains(t, detailResp, "operation_protected")

		// write_only_password is write-only: denied from read
		assert.NotContains(t, detailResp, "write_only_password")
		assert.NotContains(t, detailResp, "super_secret_pw")

		// hidden_config is non-serializable: denied from read
		assert.NotContains(t, detailResp, "hidden_config")
		assert.NotContains(t, detailResp, "internal_sys_config")

		// List GET
		recList := httptest.NewRecorder()
		mux.ServeHTTP(recList, httptest.NewRequest(http.MethodGet, "/api/v1/denials/", nil))
		require.Equal(t, http.StatusOK, recList.Code)

		assert.NotContains(t, recList.Body.String(), "write_only_password")
		assert.NotContains(t, recList.Body.String(), "super_secret_pw")
		assert.NotContains(t, recList.Body.String(), "hidden_config")
		assert.NotContains(t, recList.Body.String(), "internal_sys_config")
	})

	// 2. Field-level write denial: read-only field write is ignored per docs, stored value unchanged
	t.Run("field-level write denial: write to read-only field ignored, stored DB value unchanged", func(t *testing.T) {
		payload := map[string]interface{}{
			"public_name":          "modified_public",
			"read_only_audit_note": "tampered_audit_attempt",
		}
		bodyBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/denials/1", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, "request must succeed: %s", rec.Body.String())

		// Assert straight from the DB that public_name changed but read_only_audit_note is unchanged
		var publicName, auditNote string
		err = database.QueryRow(`SELECT public_name, read_only_audit_note FROM field_denial_models WHERE id = 1`).
			Scan(&publicName, &auditNote)
		require.NoError(t, err)
		assert.Equal(t, "modified_public", publicName)
		assert.Equal(t, "audit_fixed_v1", auditNote, "stored value straight from DB must remain unchanged")
	})

	// 3. Field-level write denial: operation-only field write fails with 403, stored value unchanged
	t.Run("field-level write denial: write to operation-only field fails 403, stored DB value unchanged", func(t *testing.T) {
		payload := map[string]interface{}{
			"operation_protected": "tampered_op_state",
		}
		bodyBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/denials/1", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code, "write to operation-only field must return 403")
		assert.Contains(t, rec.Body.String(), "operation_only_field")

		// Assert straight from the DB that stored value is unchanged
		var opProtected string
		err = database.QueryRow(`SELECT operation_protected FROM field_denial_models WHERE id = 1`).Scan(&opProtected)
		require.NoError(t, err)
		assert.Equal(t, "protected_state", opProtected, "stored value straight from DB must remain unchanged")
	})

	// 4. Field-level write denial: write to hidden non-editable field fails with 400, stored value unchanged
	t.Run("field-level write denial: write to hidden non-editable field fails 400, stored DB value unchanged", func(t *testing.T) {
		payload := map[string]interface{}{
			"hidden_config": "tampered_hidden_config",
		}
		bodyBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/denials/1", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code, "write to hidden non-editable field must return 400")
		assert.Contains(t, rec.Body.String(), "Unknown field.")

		// Assert straight from the DB that stored value is unchanged
		var hiddenConfig string
		err = database.QueryRow(`SELECT hidden_config FROM field_denial_models WHERE id = 1`).Scan(&hiddenConfig)
		require.NoError(t, err)
		assert.Equal(t, "internal_sys_config", hiddenConfig, "stored value straight from DB must remain unchanged")
	})

	// 5. Field-level write denial on create (POST): read-only field write ignored on create
	t.Run("field-level write denial on POST: write to read-only field ignored, stored DB value unchanged", func(t *testing.T) {
		payload := map[string]interface{}{
			"public_name":          "item_two",
			"write_only_password":  "second_secret_pw",
			"read_only_audit_note": "tampered_on_create",
		}
		bodyBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/denials/", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code, "POST must succeed: %s", rec.Body.String())

		// write_only_password is not exposed in the POST response body
		assert.NotContains(t, rec.Body.String(), "second_secret_pw")

		// Assert straight from the DB that read_only_audit_note was ignored on write
		var publicName, pw, auditNote string
		err = database.QueryRow(`SELECT public_name, write_only_password, read_only_audit_note FROM field_denial_models WHERE id = 2`).
			Scan(&publicName, &pw, &auditNote)
		require.NoError(t, err)
		assert.Equal(t, "item_two", publicName)
		assert.Equal(t, "second_secret_pw", pw, "write-only field is writable on POST")
		assert.Equal(t, "", auditNote, "read-only field ignored on write, stored value straight from DB remains default")
	})
}
