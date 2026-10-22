package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/admin/core"
	"github.com/stretchr/testify/require"
)

type observedExportWriter struct {
	*httptest.ResponseRecorder
	writes int
}

func (w *observedExportWriter) Write(data []byte) (int, error) {
	w.writes++
	return w.ResponseRecorder.Write(data)
}

func (w *observedExportWriter) WriteHeader(status int) {
	w.writes++
	w.ResponseRecorder.WriteHeader(status)
}

type observedExportAdmin struct {
	mockAdmin
	writer             *observedExportWriter
	pageTwoReached     bool
	wroteBeforePageTwo bool
}

func (a *observedExportAdmin) ListObjects(_ context.Context, params core.ListParams) (*core.PaginatedResponse, error) {
	if params.Page == 1 {
		return &core.PaginatedResponse{Count: 3, Results: []map[string]any{
			{"id": 1}, {"id": 2},
		}}, nil
	}
	a.pageTwoReached = true
	a.wroteBeforePageTwo = a.writer.writes > 0
	return &core.PaginatedResponse{Count: 3, Results: []map[string]any{{"id": 3}}}, nil
}

func TestCSVExportWritesAfterFetchingAllPages(t *testing.T) {
	w := &observedExportWriter{ResponseRecorder: httptest.NewRecorder()}
	admin := &observedExportAdmin{mockAdmin: mockAdmin{
		modelName: "items",
		metadata:  &core.Metadata{ListDisplay: []string{"id"}, Pagination: core.PaginationConfig{MaxPageSize: 2}},
	}, writer: w}
	NewRouter(core.NewRegistry()).handleExport(admin)(w,
		httptest.NewRequest(http.MethodGet, "/items/export?format=csv", nil))

	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, admin.pageTwoReached)
	require.False(t, admin.wroteBeforePageTwo)
	require.Equal(t, "id\n1\n2\n3\n", w.Body.String())
}
