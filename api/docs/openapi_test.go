package docs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pins the document shown in docs-site/docs/api/openapi.md: only the info
// block is filled; paths and schemas are not generated.
func TestOpenAPIHandlerServesInfoOnlyDocument(t *testing.T) {
	generator := NewOpenAPIGenerator("Shop API", "1.0.0")
	generator.Description = "Catalog and orders"

	rec := httptest.NewRecorder()
	generator.Handler()(rec, httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{
		"openapi": "3.0.0",
		"info": {"title": "Shop API", "version": "1.0.0", "description": "Catalog and orders"},
		"paths": {},
		"components": {}
	}`, rec.Body.String())
}

func TestOpenAPIRouteServesGeneratedDocument(t *testing.T) {
	generator := NewOpenAPIGenerator("Shop API", "1.0.0")
	generator.Description = "Catalog and orders"
	router := forgehttp.NewRouter()
	router.Get("/api/openapi.json", generator.Handler())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	var doc map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Equal(t, "3.0.0", doc["openapi"])
	require.Equal(t, map[string]any{"title": "Shop API", "version": "1.0.0", "description": "Catalog and orders"}, doc["info"])
	require.Empty(t, doc["paths"])
	require.Empty(t, doc["components"])
	require.Equal(t, "Shop API", generator.Generate().Info.Title)
}
