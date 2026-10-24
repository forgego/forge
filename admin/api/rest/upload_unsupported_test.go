package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/forgego/forge/admin/core"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMultipartRequest(t *testing.T, method, target string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "document.pdf")
	require.NoError(t, err)
	_, err = part.Write([]byte("fake-file-content"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func assertUnsupportedCapabilityResponse(t *testing.T, rec *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)

	var payload core.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &payload)
	require.NoError(t, err)
	assert.Equal(t, "unsupported_capability", payload.Error.Code)
	assert.Equal(t, "File uploads are not supported by the admin API", payload.Error.Message)
}

func TestUploadUnsupported_PostCreateRejectsMultipart(t *testing.T) {
	admin := &mockAdmin{
		modelName: "products",
		createObjectFn: func(data map[string]interface{}) (interface{}, error) {
			t.Fatal("mock admin CreateObject should not be called")
			return nil, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := newMultipartRequest(t, http.MethodPost, "/api/products/")
	rec := httptest.NewRecorder()

	router.handleCreate(admin)(rec, req)

	assertUnsupportedCapabilityResponse(t, rec)
	assert.Empty(t, admin.createCalls)
}

func TestUploadUnsupported_PatchUpdateRejectsMultipart(t *testing.T) {
	admin := &mockAdmin{
		modelName: "products",
		updateObjectFn: func(id interface{}, data map[string]interface{}) (interface{}, error) {
			t.Fatal("mock admin UpdateObject should not be called")
			return nil, nil
		},
		getObjectFn: func(id interface{}) (interface{}, error) {
			t.Fatal("mock admin GetObject should not be called")
			return nil, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := newMultipartRequest(t, http.MethodPatch, "/api/products/1/")
	req = withURLParam(req, "id", "1")
	rec := httptest.NewRecorder()

	router.handleUpdate(admin)(rec, req)

	assertUnsupportedCapabilityResponse(t, rec)
	assert.Empty(t, admin.updateCalls)
}

func TestUploadUnsupported_PutUpdateRejectsMultipart(t *testing.T) {
	t.Run("handleUpdate", func(t *testing.T) {
		admin := &mockAdmin{
			modelName: "products",
			updateObjectFn: func(id interface{}, data map[string]interface{}) (interface{}, error) {
				t.Fatal("mock admin UpdateObject should not be called")
				return nil, nil
			},
			getObjectFn: func(id interface{}) (interface{}, error) {
				t.Fatal("mock admin GetObject should not be called")
				return nil, nil
			},
		}
		router := NewRouter(core.NewRegistry())

		req := newMultipartRequest(t, http.MethodPut, "/api/products/1/")
		req = withURLParam(req, "id", "1")
		rec := httptest.NewRecorder()

		router.handleUpdate(admin)(rec, req)

		assertUnsupportedCapabilityResponse(t, rec)
		assert.Empty(t, admin.updateCalls)
	})

	t.Run("handleReplace", func(t *testing.T) {
		admin := &mockAdmin{
			modelName: "products",
			updateObjectFn: func(id interface{}, data map[string]interface{}) (interface{}, error) {
				t.Fatal("mock admin UpdateObject should not be called")
				return nil, nil
			},
			getObjectFn: func(id interface{}) (interface{}, error) {
				t.Fatal("mock admin GetObject should not be called")
				return nil, nil
			},
		}
		router := NewRouter(core.NewRegistry())

		req := newMultipartRequest(t, http.MethodPut, "/api/products/1/")
		req = withURLParam(req, "id", "1")
		rec := httptest.NewRecorder()

		router.handleReplace(admin)(rec, req)

		assertUnsupportedCapabilityResponse(t, rec)
		assert.Empty(t, admin.updateCalls)
	})
}

func TestUploadUnsupported_ChiRouterEndpointsRejectMultipart(t *testing.T) {
	registry := core.NewRegistry()
	admin := &mockAdmin{
		modelName:     "products",
		moduleAllowed: true,
		metadata: &core.Metadata{
			VerboseName:       "Product",
			VerboseNamePlural: "Products",
			Permissions:       core.PermissionMetadata{Add: true, Change: true, View: true},
		},
		createObjectFn: func(data map[string]interface{}) (interface{}, error) {
			t.Fatal("mock admin CreateObject should not be called via router")
			return nil, nil
		},
		updateObjectFn: func(id interface{}, data map[string]interface{}) (interface{}, error) {
			t.Fatal("mock admin UpdateObject should not be called via router")
			return nil, nil
		},
	}
	require.NoError(t, registry.Register(admin))

	apiRouter := newTestRouter(registry)
	apiRouter.SetUserRepository(adminTestUsers{"admin": {ID: 1, Username: "admin", IsActive: true, IsSuperuser: true}})
	token, err := apiRouter.tokens.IssueToken(context.Background(), "admin", 1*time.Hour)
	require.NoError(t, err)

	root := chi.NewRouter()
	apiRouter.RegisterRoutes(root)

	tests := []struct {
		name   string
		method string
		target string
	}{
		{"POST create", http.MethodPost, "/api/products/"},
		{"PATCH update", http.MethodPatch, "/api/products/1/"},
		{"PUT replace", http.MethodPut, "/api/products/1/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newMultipartRequest(t, tt.method, tt.target)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			root.ServeHTTP(rec, req)

			assertUnsupportedCapabilityResponse(t, rec)
			assert.Empty(t, admin.createCalls)
			assert.Empty(t, admin.updateCalls)
		})
	}
}
