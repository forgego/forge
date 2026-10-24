package rest

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type exportTestAdmin struct {
	mockAdmin
	hasViewPermission func(obj any) bool
	listCalls         []core.ListParams
	listFn            func(ctx context.Context, params core.ListParams) (*core.PaginatedResponse, error)
}

func (a *exportTestAdmin) HasViewPermission(ctx context.Context, user any, obj any) bool {
	if a.hasViewPermission != nil {
		return a.hasViewPermission(obj)
	}
	return !a.denyView
}

func (a *exportTestAdmin) ListObjects(ctx context.Context, params core.ListParams) (*core.PaginatedResponse, error) {
	a.listCalls = append(a.listCalls, params)
	if a.listFn != nil {
		return a.listFn(ctx, params)
	}
	return a.mockAdmin.ListObjects(ctx, params)
}

type exportSecretModel struct {
	schema.BaseSchema
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Secret   string `json:"secret" db:"secret"`
}

func (exportSecretModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary()),
		schema.StringField("username"),
		schema.StringField("secret", schema.WriteOnly()),
	}
}

type exportSecretAdmin struct {
	mockAdmin
}

func (*exportSecretAdmin) ModelType() reflect.Type {
	return reflect.TypeOf(exportSecretModel{})
}

func TestExportExcludesRowsUserCannotView(t *testing.T) {
	admin := &exportTestAdmin{
		mockAdmin: mockAdmin{
			modelName: "items",
			metadata:  &core.Metadata{ListDisplay: []string{"id", "name"}},
			listResponse: &core.PaginatedResponse{
				Count: 2,
				Results: []map[string]any{
					{"id": 1, "name": "allowed_row"},
					{"id": 2, "name": "forbidden_row"},
				},
			},
		},
		hasViewPermission: func(obj any) bool {
			if obj == nil {
				return true
			}
			m, ok := obj.(map[string]any)
			if !ok {
				return true
			}
			return m["name"] == "allowed_row"
		},
	}

	for _, format := range []string{"json", "csv"} {
		t.Run(format, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/items/export?format="+format, nil)
			NewRouter(core.NewRegistry()).handleExport(admin)(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			body := rec.Body.String()
			assert.Contains(t, body, "allowed_row")
			assert.NotContains(t, body, "forbidden_row")

			if format == "json" {
				var rows []map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &rows))
				require.Len(t, rows, 1)
				assert.Equal(t, "allowed_row", rows[0]["name"])
			} else {
				records, err := csv.NewReader(strings.NewReader(body)).ReadAll()
				require.NoError(t, err)
				require.Len(t, records, 2)
				assert.Equal(t, []string{"id", "name"}, records[0])
				assert.Equal(t, "allowed_row", records[1][1])
			}
		})
	}
}

func TestExportHiddenFieldsAbsentFromJSONAndCSV(t *testing.T) {
	admin := &exportSecretAdmin{
		mockAdmin: mockAdmin{
			modelName: "users",
			metadata: &core.Metadata{
				ListDisplay: []string{"id", "username", "secret"},
			},
			listResponse: &core.PaginatedResponse{
				Count: 1,
				Results: []map[string]any{
					{"id": 1, "username": "alice", "secret": "top-secret-password"},
				},
			},
		},
	}

	for _, format := range []string{"json", "csv"} {
		t.Run(format, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/users/export?format="+format, nil)
			NewRouter(core.NewRegistry()).handleExport(admin)(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			body := rec.Body.String()
			assert.Contains(t, body, "alice")
			assert.NotContains(t, body, "top-secret-password")
			assert.NotContains(t, body, "secret")

			if format == "json" {
				var rows []map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &rows))
				require.Len(t, rows, 1)
				_, hasSecret := rows[0]["secret"]
				assert.False(t, hasSecret, "hidden field 'secret' must not be in JSON")
				assert.Equal(t, "alice", rows[0]["username"])
			} else {
				records, err := csv.NewReader(strings.NewReader(body)).ReadAll()
				require.NoError(t, err)
				require.Len(t, records, 2)
				assert.Equal(t, []string{"id", "username"}, records[0])
				assert.Equal(t, []string{"1", "alice"}, records[1])
			}
		})
	}
}

func TestExportFilterQueryParamsReachListObjects(t *testing.T) {
	admin := &exportTestAdmin{
		mockAdmin: mockAdmin{
			modelName: "items",
			metadata:  &core.Metadata{ListDisplay: []string{"id", "status"}},
			listResponse: &core.PaginatedResponse{
				Count:   0,
				Results: []map[string]any{},
			},
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/items/export?format=json&status=active&search=laptop&ordering=-created_at&category=tech", nil)
	NewRouter(core.NewRegistry()).handleExport(admin)(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, admin.listCalls, 1)
	call := admin.listCalls[0]
	assert.Equal(t, "active", call.Filters["status"])
	assert.Equal(t, "tech", call.Filters["category"])
	assert.Equal(t, "laptop", call.Search)
	assert.Equal(t, []string{"-created_at"}, call.Ordering)
}

func TestExportCountExceedsMaxExportRows(t *testing.T) {
	for _, format := range []string{"json", "csv"} {
		t.Run("upfront_count_"+format, func(t *testing.T) {
			admin := &exportTestAdmin{
				mockAdmin: mockAdmin{
					modelName: "items",
					metadata:  &core.Metadata{ListDisplay: []string{"id", "name"}},
					listResponse: &core.PaginatedResponse{
						Count: 15000,
						Results: []map[string]any{
							{"id": 1, "name": "should_not_leak"},
						},
					},
				},
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/items/export?format="+format, nil)
			NewRouter(core.NewRegistry()).handleExport(admin)(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Code)
			require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var errResp core.ErrorResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
			assert.Equal(t, "export_too_large", errResp.Error.Code)
			assert.Equal(t, fmt.Sprintf("export exceeds %d rows; narrow the filters", core.MaxExportRows), errResp.Error.Message)
			assert.Equal(t, float64(15000), errResp.Error.Details["count"])
			assert.Equal(t, float64(core.MaxExportRows), errResp.Error.Details["limit"])

			// Verify no export body rows are written
			assert.NotContains(t, rec.Body.String(), "should_not_leak")
		})
	}

	t.Run("accumulated_rows_exceed_limit", func(t *testing.T) {
		items := make([]map[string]any, core.MaxExportRows+1)
		for i := range items {
			items[i] = map[string]any{"id": i + 1, "name": fmt.Sprintf("item_%d", i+1)}
		}

		admin := &exportTestAdmin{
			mockAdmin: mockAdmin{
				modelName: "items",
				metadata:  &core.Metadata{ListDisplay: []string{"id", "name"}},
				listResponse: &core.PaginatedResponse{
					Count:   0,
					Results: items,
				},
			},
		}

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items/export?format=json", nil)
		NewRouter(core.NewRegistry()).handleExport(admin)(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		var errResp core.ErrorResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
		assert.Equal(t, "export_too_large", errResp.Error.Code)
		assert.Equal(t, fmt.Sprintf("export exceeds %d rows; narrow the filters", core.MaxExportRows), errResp.Error.Message)
		assert.Equal(t, float64(core.MaxExportRows+1), errResp.Error.Details["count"])
		assert.Equal(t, float64(core.MaxExportRows), errResp.Error.Details["limit"])
		assert.NotContains(t, rec.Body.String(), "item_1")
	})
}

func TestExportBadFormatDoesNotCallListObjects(t *testing.T) {
	for _, badFormat := range []string{"xml", "yaml", "pdf", "tsv", "invalid"} {
		t.Run(badFormat, func(t *testing.T) {
			admin := &exportTestAdmin{
				mockAdmin: mockAdmin{
					modelName: "items",
					metadata:  &core.Metadata{ListDisplay: []string{"id", "name"}},
					listResponse: &core.PaginatedResponse{
						Count: 1,
						Results: []map[string]any{
							{"id": 1, "name": "item"},
						},
					},
				},
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/items/export?format="+badFormat, nil)
			NewRouter(core.NewRegistry()).handleExport(admin)(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Code)
			var errResp core.ErrorResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
			assert.Equal(t, "invalid_format", errResp.Error.Code)
			assert.Equal(t, "Unsupported export format", errResp.Error.Message)
			assert.Empty(t, admin.listCalls, "ListObjects must not be called when format is invalid")
		})
	}
}
