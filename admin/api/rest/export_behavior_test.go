package rest

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/forgego/forge/admin/core"
	"github.com/stretchr/testify/require"
)

type exportPagingAdmin struct {
	mockAdmin
	calls []core.ListParams
}

func (a *exportPagingAdmin) ListObjects(_ context.Context, params core.ListParams) (*core.PaginatedResponse, error) {
	a.calls = append(a.calls, params)
	if params.Page == 1 {
		return &core.PaginatedResponse{Count: 3, Results: []map[string]any{
			{"id": 1, "name": "first", "status": "open"},
			{"id": 2, "name": "second", "status": "open"},
		}}, nil
	}
	return &core.PaginatedResponse{Count: 3, Results: []map[string]any{
		{"id": 3, "name": "third", "status": "open"},
	}}, nil
}

func TestCSVExportIncludesEveryPageWithActiveFilters(t *testing.T) {
	admin := &exportPagingAdmin{mockAdmin: mockAdmin{
		modelName: "items",
		metadata:  &core.Metadata{ListDisplay: []string{"id", "name"}, Pagination: core.PaginationConfig{MaxPageSize: 2}},
	}}
	rec := httptest.NewRecorder()
	NewRouter(core.NewRegistry()).handleExport(admin)(rec,
		httptest.NewRequest(http.MethodGet, "/items/export?format=csv&status=open&search=fi&ordering=name", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "text/csv", rec.Header().Get("Content-Type"))
	rows, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	require.NoError(t, err)
	require.Equal(t, [][]string{{"id", "name"}, {"1", "first"}, {"2", "second"}, {"3", "third"}}, rows)
	require.Len(t, admin.calls, 2)
	require.Equal(t, []int{1, 2}, []int{admin.calls[0].Page, admin.calls[1].Page})
	for _, call := range admin.calls {
		require.Equal(t, "open", call.Filters["status"])
		require.Equal(t, "fi", call.Search)
		require.Equal(t, []string{"name"}, call.Ordering)
	}
}

func TestCSVExportKeepsListedColumnsWhenClientHidesOne(t *testing.T) {
	admin := &mockAdmin{
		modelName:    "items",
		metadata:     &core.Metadata{ListDisplay: []string{"id", "name"}},
		listResponse: &core.PaginatedResponse{Count: 1, Results: []map[string]any{{"id": 1, "name": "first"}}},
	}
	rec := httptest.NewRecorder()
	NewRouter(core.NewRegistry()).handleExport(admin)(rec,
		httptest.NewRequest(http.MethodGet, "/items/export?format=csv&hidden=name", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	rows, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	require.NoError(t, err)
	require.Equal(t, []string{"id", "name"}, rows[0])
	require.Equal(t, []string{"1", "first"}, rows[1])
}

func TestCSVExportEscapesFormulaCellsButPreservesNumericValues(t *testing.T) {
	rec := httptest.NewRecorder()
	err := writeCSVExport(rec, "items", []string{"formula", "number", "numeric_text"}, nil,
		[]map[string]any{{"formula": "=1+1", "number": -5, "numeric_text": "-5"}})
	require.NoError(t, err)
	rows, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	require.NoError(t, err)
	require.Equal(t, []string{"'=1+1", "-5", "'-5"}, rows[1])
}

func TestJSONExportUsesFilteredListRatherThanSelectedIDs(t *testing.T) {
	admin := &mockAdmin{
		modelName: "items",
		metadata:  &core.Metadata{ListDisplay: []string{"id", "name"}},
		listResponse: &core.PaginatedResponse{Count: 2, Results: []map[string]any{
			{"id": 1, "name": "first"}, {"id": 2, "name": "second"},
		}},
	}
	rec := httptest.NewRecorder()
	NewRouter(core.NewRegistry()).handleExport(admin)(rec,
		httptest.NewRequest(http.MethodGet, "/items/export?format=json&ids=1", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &rows))
	require.Len(t, rows, 2)
	require.Equal(t, float64(2), rows[1]["id"])
}
