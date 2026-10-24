package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/schema"
	"github.com/forgego/forge/stores/storestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type extTestModel struct {
	schema.BaseSchema
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func (extTestModel) Meta() schema.Meta { return schema.Meta{TableName: "ext_test_models"} }
func (extTestModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
	}
}

type extTestUsers struct{}

func (extTestUsers) GetByUsername(_ context.Context, username string) (*models.User, error) {
	if username == "staff-admin" {
		return &models.User{ID: 1, Username: "staff-admin", IsActive: true, IsStaff: true, IsSuperuser: true}, nil
	}
	if username == "regular-user" {
		return &models.User{ID: 2, Username: "regular-user", IsActive: true, IsStaff: false, IsSuperuser: false}, nil
	}
	return nil, repository.ErrUserNotFound
}

func TestExtension_RegisterValidation(t *testing.T) {
	validFS := fstest.MapFS{
		"index.js": &fstest.MapFile{Data: []byte("console.log('valid');")},
	}

	t.Run("bad name rejected", func(t *testing.T) {
		badNames := []string{
			"Orders",
			"orders_v1",
			"orders v1",
			"orders/extra",
			"orders!",
			"orders.js",
			"",
			"-orders", // matches regex ^[a-z0-9-]+$, but let's test invalid characters
			"orders@v1",
		}
		for _, name := range badNames {
			if name == "-orders" {
				// regex ^[a-z0-9-]+$ accepts hyphen anywhere; test characters outside [a-z0-9-]
				continue
			}
			site := NewSite("test")
			err := site.RegisterExtension(Extension{
				Name:    name,
				Assets:  validFS,
				Version: ExtensionAPIVersion,
			})
			require.Error(t, err, "expected error for bad name %q", name)
			assert.Contains(t, err.Error(), "invalid name")
		}
	})

	t.Run("duplicate rejected", func(t *testing.T) {
		site := NewSite("test")
		err := site.RegisterExtension(Extension{
			Name:    "orders",
			Assets:  validFS,
			Version: ExtensionAPIVersion,
		})
		require.NoError(t, err)

		err = site.RegisterExtension(Extension{
			Name:    "orders",
			Assets:  validFS,
			Version: ExtensionAPIVersion,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("missing index.js rejected", func(t *testing.T) {
		// Missing index.js in FS
		site := NewSite("test")
		noIndexFS := fstest.MapFS{
			"main.js": &fstest.MapFile{Data: []byte("console.log('main');")},
		}
		err := site.RegisterExtension(Extension{
			Name:    "orders",
			Assets:  noIndexFS,
			Version: ExtensionAPIVersion,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing index.js")

		// Nil Assets
		siteNil := NewSite("test")
		errNil := siteNil.RegisterExtension(Extension{
			Name:    "orders",
			Assets:  nil,
			Version: ExtensionAPIVersion,
		})
		require.Error(t, errNil)
		assert.Contains(t, errNil.Error(), "assets filesystem cannot be nil")

		// index.js is a directory
		siteDir := NewSite("test")
		dirFS := fstest.MapFS{
			"index.js/inner.txt": &fstest.MapFile{Data: []byte("not a file")},
		}
		errDir := siteDir.RegisterExtension(Extension{
			Name:    "orders",
			Assets:  dirFS,
			Version: ExtensionAPIVersion,
		})
		require.Error(t, errDir)
		assert.Contains(t, errDir.Error(), "index.js must be a file, not a directory")
	})

	t.Run("wrong version rejected", func(t *testing.T) {
		wrongVersions := []int{0, 2, -1, 42}
		for _, v := range wrongVersions {
			site := NewSite("test")
			err := site.RegisterExtension(Extension{
				Name:    "orders",
				Assets:  validFS,
				Version: v,
			})
			require.Error(t, err, "expected error for version %d", v)
			assert.Contains(t, err.Error(), "unsupported version")
		}
	})

	t.Run("valid extension registered successfully", func(t *testing.T) {
		site := NewSite("test")
		err := site.RegisterExtension(Extension{
			Name:    "orders",
			Assets:  validFS,
			Version: ExtensionAPIVersion,
		})
		require.NoError(t, err)
		assert.Len(t, site.GetExtensions(), 1)
		assert.Equal(t, "orders", site.GetExtensions()[0].Name)
	})
}

func TestExtension_MountValidation(t *testing.T) {
	validFS := fstest.MapFS{
		"index.js": &fstest.MapFile{Data: []byte("console.log('valid');")},
	}

	t.Run("mounting with valid extensions succeeds", func(t *testing.T) {
		site := NewSite("test")
		err := site.RegisterExtension(Extension{
			Name:    "orders",
			Assets:  validFS,
			Version: ExtensionAPIVersion,
		})
		require.NoError(t, err)
		assert.NotNil(t, site.Handler())
	})

	t.Run("mounting with duplicate name panics", func(t *testing.T) {
		site := NewSite("test")
		site.extensions = []Extension{
			{Name: "orders", Assets: validFS, Version: ExtensionAPIVersion},
			{Name: "orders", Assets: validFS, Version: ExtensionAPIVersion},
		}
		assert.Panics(t, func() {
			site.Handler()
		})
	})

	t.Run("mounting with invalid name panics", func(t *testing.T) {
		site := NewSite("test")
		site.extensions = []Extension{
			{Name: "BAD_NAME", Assets: validFS, Version: ExtensionAPIVersion},
		}
		assert.Panics(t, func() {
			site.Handler()
		})
	})

	t.Run("mounting with missing index.js panics", func(t *testing.T) {
		site := NewSite("test")
		site.extensions = []Extension{
			{Name: "orders", Assets: fstest.MapFS{}, Version: ExtensionAPIVersion},
		}
		assert.Panics(t, func() {
			site.Handler()
		})
	})

	t.Run("mounting with wrong version panics", func(t *testing.T) {
		site := NewSite("test")
		site.extensions = []Extension{
			{Name: "orders", Assets: validFS, Version: 99},
		}
		assert.Panics(t, func() {
			site.Handler()
		})
	})
}

func TestExtension_AssetServing(t *testing.T) {
	ordersFS := fstest.MapFS{
		"index.js": &fstest.MapFile{
			Data: []byte("console.log('orders entry');"),
		},
		"styles/theme.css": &fstest.MapFile{
			Data: []byte(".order-box { color: red; }"),
		},
		"data.json": &fstest.MapFile{
			Data: []byte(`{"enabled":true}`),
		},
	}

	site := NewSite("test")
	err := site.RegisterExtension(Extension{
		Name:    "orders",
		Assets:  ordersFS,
		Version: ExtensionAPIVersion,
	})
	require.NoError(t, err)

	handler := site.Handler()

	t.Run("index.js served with JS content type and nosniff", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ext/orders/index.js", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "javascript")
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "console.log('orders entry');", rec.Body.String())
	})

	t.Run("nested asset served with correct content type and nosniff", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ext/orders/styles/theme.css", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/css")
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, ".order-box { color: red; }", rec.Body.String())
	})

	t.Run("directory listing returns 404", func(t *testing.T) {
		dirPaths := []string{
			"/ext/orders/",
			"/ext/orders",
			"/ext/orders/styles/",
			"/ext/orders/styles",
			"/ext/",
			"/ext",
		}

		for _, p := range dirPaths {
			req := httptest.NewRequest(http.MethodGet, p, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code, "expected 404 for directory path %q", p)
			assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		}
	})

	t.Run("traversal returns 404", func(t *testing.T) {
		traversalPaths := []string{
			"/ext/orders/../orders/index.js",
			"/ext/orders/../../secret",
			"/ext/orders/..",
			"/ext/orders/styles/../../index.js",
			"/ext/orders/%2e%2e/orders/index.js",
			"/ext/orders/%2e%2e/%2e%2e/secret",
			"/ext/orders/%2e%2e",
			"/ext/orders/styles/%2e%2e/%2e%2e/index.js",
		}

		for _, p := range traversalPaths {
			req := httptest.NewRequest(http.MethodGet, p, nil)
			req.URL.Path = p
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code, "expected 404 for traversal path %q", p)
			assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		}
	})

	t.Run("non-existent file returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ext/orders/missing.js", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	})

	t.Run("disallowed HTTP methods return 405", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
		for _, m := range methods {
			req := httptest.NewRequest(m, "/ext/orders/index.js", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "expected 405 for %s", m)
			assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		}
	})
}

func TestExtension_CustomPrefixAssetServing(t *testing.T) {
	ordersFS := fstest.MapFS{
		"index.js": &fstest.MapFile{
			Data: []byte("console.log('custom prefix');"),
		},
	}

	site := NewSite("test")
	uiCfg := site.GetUIConfig()
	uiCfg.Prefix = "/backoffice"
	site.WithUIConfig(uiCfg)

	err := site.RegisterExtension(Extension{
		Name:    "orders",
		Assets:  ordersFS,
		Version: ExtensionAPIVersion,
	})
	require.NoError(t, err)

	handler := site.Handler()

	req := httptest.NewRequest(http.MethodGet, "/backoffice/ext/orders/index.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "javascript")
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "console.log('custom prefix');", rec.Body.String())
}

func TestExtension_MetadataDiscovery(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "staff-admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "secret")

	database := storestest.OpenMemory(t)

	ordersFS := fstest.MapFS{
		"index.js": &fstest.MapFile{Data: []byte("console.log('orders');")},
	}

	site := NewSite("test")
	site.SetUserStore(extTestUsers{})
	site.SetDB(database)
	require.NoError(t, site.UseDatabaseStores(context.Background()))

	// Register a model so models list is populated
	_, err := RegisterWithSite(site, &core.Config[extTestModel]{})
	require.NoError(t, err)

	err = site.RegisterExtension(Extension{
		Name:    "orders",
		Assets:  ordersFS,
		Version: ExtensionAPIVersion,
	})
	require.NoError(t, err)

	ts := httptest.NewServer(site.Handler())
	t.Cleanup(ts.Close)

	t.Run("unauthenticated metadata does not expose extensions", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/meta", nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Unauthenticated request is rejected by authMiddleware with 401
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var payload map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&payload)
		_, hasExtensions := payload["extensions"]
		assert.False(t, hasExtensions, "unauthenticated response must not contain extensions")
	})

	t.Run("authenticated staff metadata lists registered extension", func(t *testing.T) {
		// Log in as staff
		loginStatus, loginBody := adminCall(t, ts.URL, http.MethodPost, "/api/login", "", map[string]string{
			"username": "staff-admin",
			"password": "secret",
		})
		require.Equal(t, http.StatusOK, loginStatus)
		token, ok := loginBody["token"].(string)
		require.True(t, ok)
		require.NotEmpty(t, token)

		// Request metadata with staff token
		metaStatus, metaBody := adminCall(t, ts.URL, http.MethodGet, "/api/meta", token, nil)
		require.Equal(t, http.StatusOK, metaStatus)

		// Models list is preserved
		modelsList, ok := metaBody["models"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, modelsList)

		// Extensions list is present
		extsList, ok := metaBody["extensions"].([]interface{})
		require.True(t, ok, "authenticated staff metadata must contain extensions")
		require.Len(t, extsList, 1)

		ext0, ok := extsList[0].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "orders", ext0["name"])
		assert.Equal(t, "/admin/ext/orders/index.js", ext0["entry"])
		assert.Equal(t, float64(1), ext0["version"])
	})

	t.Run("empty extensions list when none registered", func(t *testing.T) {
		cleanSite := NewSite("empty-ext")
		cleanSite.SetUserStore(extTestUsers{})
		cleanSite.SetDB(database)
		require.NoError(t, cleanSite.UseDatabaseStores(context.Background()))

		_, regErr := RegisterWithSite(cleanSite, &core.Config[extTestModel]{})
		require.NoError(t, regErr)

		cleanTs := httptest.NewServer(cleanSite.Handler())
		t.Cleanup(cleanTs.Close)

		// Log in
		loginStatus, loginBody := adminCall(t, cleanTs.URL, http.MethodPost, "/api/login", "", map[string]string{
			"username": "staff-admin",
			"password": "secret",
		})
		require.Equal(t, http.StatusOK, loginStatus)
		token, _ := loginBody["token"].(string)

		metaStatus, metaBody := adminCall(t, cleanTs.URL, http.MethodGet, "/api/meta", token, nil)
		require.Equal(t, http.StatusOK, metaStatus)

		extsList, ok := metaBody["extensions"].([]interface{})
		require.True(t, ok, "extensions should be present as a list")
		assert.Empty(t, extsList, "extensions must be empty list when none registered")
	})
}
