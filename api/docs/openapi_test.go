package docs

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
