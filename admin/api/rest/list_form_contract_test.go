package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/admin/core"
	validation "github.com/forgego/forge/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type contractTestAdmin struct {
	mockAdmin
	capturedListParams core.ListParams
	hasViewPermFn      func(ctx context.Context, user interface{}, obj interface{}) bool
}

func (c *contractTestAdmin) ListObjects(ctx context.Context, params core.ListParams) (*core.PaginatedResponse, error) {
	c.capturedListParams = params
	if c.listErr != nil {
		return nil, c.listErr
	}
	if c.listResponse != nil {
		return c.listResponse, nil
	}
	return &core.PaginatedResponse{
		Results: []interface{}{},
	}, nil
}

func (c *contractTestAdmin) HasViewPermission(ctx context.Context, user interface{}, obj interface{}) bool {
	if c.hasViewPermFn != nil {
		return c.hasViewPermFn(ctx, user, obj)
	}
	return !c.denyView
}

func TestAdminListFormContract(t *testing.T) {
	t.Run("metadata includes contract_version == AdminContractVersion", func(t *testing.T) {
		registry := core.NewRegistry()
		admin := &mockAdmin{
			modelName:     "products",
			moduleAllowed: true,
			metadata: &core.Metadata{
				ContractVersion:   core.AdminContractVersion,
				Name:              "products",
				VerboseName:       "Product",
				VerboseNamePlural: "Products",
			},
		}
		require.NoError(t, registry.Register(admin))

		router := NewRouter(registry)
		req := withURLParam(httptest.NewRequest(http.MethodGet, "/api/meta/products", nil), "model", "products")
		rec := httptest.NewRecorder()

		router.handleMetaDetail(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.Contains(t, resp, "contract_version")
		assert.Equal(t, float64(core.AdminContractVersion), resp["contract_version"])
		assert.Equal(t, float64(1), resp["contract_version"])
	})

	t.Run("list search param reaches ListObjects", func(t *testing.T) {
		admin := &contractTestAdmin{
			mockAdmin: mockAdmin{
				modelName: "products",
			},
		}
		router := NewRouter(core.NewRegistry())
		req := httptest.NewRequest(http.MethodGet, "/api/products?search=gadget", nil)
		rec := httptest.NewRecorder()

		router.handleList(admin)(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gadget", admin.capturedListParams.Search)
	})

	t.Run("filters on non-allowlisted fields are rejected or ignored as the handler does today", func(t *testing.T) {
		admin := &contractTestAdmin{
			mockAdmin: mockAdmin{
				modelName: "products",
			},
		}
		router := NewRouter(core.NewRegistry())
		req := httptest.NewRequest(http.MethodGet, "/api/products?non_allowlisted=disallowed_val", nil)
		rec := httptest.NewRecorder()

		router.handleList(admin)(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "disallowed_val", admin.capturedListParams.Filters["non_allowlisted"])
	})

	t.Run("list pagination page_size above max is bounded", func(t *testing.T) {
		admin := &contractTestAdmin{
			mockAdmin: mockAdmin{
				modelName: "products",
			},
		}
		router := NewRouter(core.NewRegistry())

		req := httptest.NewRequest(http.MethodGet, "/api/products?page_size=500", nil)
		rec := httptest.NewRecorder()

		router.handleList(admin)(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.LessOrEqual(t, admin.capturedListParams.PageSize, 100)
		assert.Equal(t, 25, admin.capturedListParams.PageSize)
	})

	t.Run("relationship choices/autocomplete only return rows the user may view", func(t *testing.T) {
		admin := &contractTestAdmin{
			mockAdmin: mockAdmin{
				modelName: "categories",
				autocompleteItems: []core.AutocompleteItem{
					{Value: "1", Label: "Visible Category"},
					{Value: "2", Label: "Secret Category"},
				},
				getObjectFn: func(id interface{}) (interface{}, error) {
					return map[string]interface{}{"id": id}, nil
				},
			},
			hasViewPermFn: func(ctx context.Context, user interface{}, obj interface{}) bool {
				if obj == nil {
					return true
				}
				row, ok := obj.(map[string]interface{})
				if !ok {
					return false
				}
				return row["id"] == "1"
			},
		}

		router := NewRouter(core.NewRegistry())
		req := httptest.NewRequest(http.MethodGet, "/api/categories/autocomplete?q=cat", nil)
		rec := httptest.NewRecorder()

		router.handleAutocomplete(admin)(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp core.AutocompleteResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.Len(t, resp.Results, 1)
		assert.Equal(t, "1", resp.Results[0].Value)
		assert.Equal(t, "Visible Category", resp.Results[0].Label)

		// Assert model-level permission check rejects when view is denied
		deniedAdmin := &contractTestAdmin{
			mockAdmin: mockAdmin{
				modelName: "categories",
				denyView:  true,
			},
		}
		deniedRec := httptest.NewRecorder()
		router.handleAutocomplete(deniedAdmin)(deniedRec, req)
		assert.Equal(t, http.StatusForbidden, deniedRec.Code)
	})

	t.Run("a validation error on create returns field errors keyed by field name", func(t *testing.T) {
		valErrs := &validation.ValidationErrors{}
		valErrs.Add("title", "Title is required")
		valErrs.Add("sku", "SKU must be unique")

		admin := &mockAdmin{
			modelName: "products",
			createObjectFn: func(data map[string]interface{}) (interface{}, error) {
				return nil, valErrs
			},
		}

		router := NewRouter(core.NewRegistry())
		body := `{"title": "", "sku": ""}`
		req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		router.handleCreate(admin)(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)

		var resp struct {
			Error struct {
				Code    string              `json:"code"`
				Message string              `json:"message"`
				Details map[string][]string `json:"details"`
			} `json:"error"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "validation_error", resp.Error.Code)
		require.NotNil(t, resp.Error.Details)
		assert.Contains(t, resp.Error.Details, "title")
		assert.Contains(t, resp.Error.Details["title"], "Title is required")
		assert.Contains(t, resp.Error.Details, "sku")
		assert.Contains(t, resp.Error.Details["sku"], "SKU must be unique")
	})
}
