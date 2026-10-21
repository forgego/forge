package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/admin/core"
	"github.com/stretchr/testify/require"
)

type selectiveViewAdmin struct{ mockAdmin }

func (*selectiveViewAdmin) HasViewPermission(_ context.Context, _ any, obj any) bool {
	if obj == nil {
		return true
	}
	return obj.(map[string]any)["id"] != 2
}

func TestListOmitsObjectsWithoutViewPermission(t *testing.T) {
	admin := &selectiveViewAdmin{mockAdmin: mockAdmin{listResponse: &core.PaginatedResponse{
		Count: 2, Results: []map[string]any{{"id": 1, "name": "public"}, {"id": 2, "name": "private"}},
	}}}
	rec := httptest.NewRecorder()
	NewRouter(core.NewRegistry()).handleList(admin)(rec, httptest.NewRequest(http.MethodGet, "/items", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Results []map[string]any `json:"results"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Results, 1)
	require.Equal(t, "public", body.Results[0]["name"])
}

func TestExportOmitsObjectsWithoutViewPermission(t *testing.T) {
	admin := &selectiveViewAdmin{mockAdmin: mockAdmin{
		metadata:     &core.Metadata{ListDisplay: []string{"id", "name"}},
		listResponse: &core.PaginatedResponse{Count: 2, Results: []map[string]any{{"id": 1, "name": "public"}, {"id": 2, "name": "private"}}},
	}}
	for _, format := range []string{"json", "csv"} {
		t.Run(format, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewRouter(core.NewRegistry()).handleExport(admin)(rec, httptest.NewRequest(http.MethodGet, "/items/export?format="+format, nil))
			require.Equal(t, http.StatusOK, rec.Code)
			require.Contains(t, rec.Body.String(), "public")
			require.NotContains(t, rec.Body.String(), "private")
		})
	}
}

type selectiveRelatedAdmin struct{ mockResolverAdmin }

func (*selectiveRelatedAdmin) HasViewPermission(_ context.Context, _ any, obj any) bool {
	if obj == nil {
		return true
	}
	return fmt.Sprint(obj.(map[string]any)["id"]) != "20"
}
func (*selectiveRelatedAdmin) GetObject(_ context.Context, id any) (any, error) {
	return map[string]any{"id": id}, nil
}

func TestDisplayLabelsOmitRelatedObjectsWithoutViewPermission(t *testing.T) {
	registry := core.NewRegistry()
	related := &selectiveRelatedAdmin{mockResolverAdmin: mockResolverAdmin{
		mockAdmin: mockAdmin{modelName: "authors"},
		labels:    map[string]string{"10": "public", "20": "private"},
	}}
	require.NoError(t, registry.Register(related))
	admin := &mockAdmin{metadata: &core.Metadata{Relations: []core.RelationMetadata{{Name: "author", Type: "ForeignKey", RelatedModel: "authors"}}}}
	response := &core.PaginatedResponse{Results: []map[string]any{{"author_id": 10}, {"author_id": 20}}}
	NewRouter(registry).attachDisplayLabels(context.Background(), admin, nil, response)
	require.Equal(t, map[string]string{"10": "public"}, response.Display["author"], fmt.Sprint(response.Display))
}
