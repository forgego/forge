package api

import (
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/parsers"
	"github.com/forgego/forge/api/renderers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentNegotiator_SelectRenderer_JSON(t *testing.T) {
	negotiator := NewContentNegotiator(
		[]renderers.Renderer{
			renderers.NewJSONRenderer(),
			renderers.NewXMLRenderer(),
		},
		[]parsers.Parser{
			parsers.NewJSONParser(),
		},
	)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/json")

	renderer := negotiator.SelectRenderer(req)
	require.NotNil(t, renderer)
	assert.Equal(t, "application/json", renderer.MediaType())
}

func TestContentNegotiator_SelectRenderer_XML(t *testing.T) {
	negotiator := NewContentNegotiator(
		[]renderers.Renderer{
			renderers.NewJSONRenderer(),
			renderers.NewXMLRenderer(),
		},
		[]parsers.Parser{
			parsers.NewJSONParser(),
		},
	)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/xml")

	renderer := negotiator.SelectRenderer(req)
	require.NotNil(t, renderer)
	assert.Equal(t, "application/xml", renderer.MediaType())
}

func TestContentNegotiator_SelectRenderer_Default(t *testing.T) {
	negotiator := NewContentNegotiator(
		[]renderers.Renderer{
			renderers.NewJSONRenderer(),
			renderers.NewXMLRenderer(),
		},
		[]parsers.Parser{
			parsers.NewJSONParser(),
		},
	)

	req := httptest.NewRequest("GET", "/test", nil)
	// No Accept header

	renderer := negotiator.SelectRenderer(req)
	require.NotNil(t, renderer)
	// Should default to first renderer (JSON)
	assert.Equal(t, "application/json", renderer.MediaType())
}

func TestContentNegotiator_SelectParser_JSON(t *testing.T) {
	negotiator := NewContentNegotiator(
		[]renderers.Renderer{
			renderers.NewJSONRenderer(),
		},
		[]parsers.Parser{
			parsers.NewJSONParser(),
			parsers.NewFormParser(),
		},
	)

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	parser := negotiator.SelectParser(req)
	require.NotNil(t, parser)
	assert.Equal(t, "application/json", parser.MediaType())
}

func TestContentNegotiator_SelectParser_Form(t *testing.T) {
	negotiator := NewContentNegotiator(
		[]renderers.Renderer{
			renderers.NewJSONRenderer(),
		},
		[]parsers.Parser{
			parsers.NewJSONParser(),
			parsers.NewFormParser(),
		},
	)

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	parser := negotiator.SelectParser(req)
	require.NotNil(t, parser)
	assert.Equal(t, "application/x-www-form-urlencoded", parser.MediaType())
}

func TestContentNegotiator_SelectParser_Default(t *testing.T) {
	negotiator := NewContentNegotiator(
		[]renderers.Renderer{
			renderers.NewJSONRenderer(),
		},
		[]parsers.Parser{
			parsers.NewJSONParser(),
			parsers.NewFormParser(),
		},
	)

	req := httptest.NewRequest("POST", "/test", nil)
	// No Content-Type header

	parser := negotiator.SelectParser(req)
	require.NotNil(t, parser)
	// Should default to first parser (JSON)
	assert.Equal(t, "application/json", parser.MediaType())
}

func TestContentNegotiator_SelectParser_WithCharset(t *testing.T) {
	negotiator := NewContentNegotiator(
		[]renderers.Renderer{
			renderers.NewJSONRenderer(),
		},
		[]parsers.Parser{
			parsers.NewJSONParser(),
		},
	)

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	parser := negotiator.SelectParser(req)
	require.NotNil(t, parser)
	assert.Equal(t, "application/json", parser.MediaType())
}

// Note: parseAcceptHeader is an internal function, testing through SelectRenderer
func TestContentNegotiator_ParseAcceptHeader(t *testing.T) {
	negotiator := NewContentNegotiator(
		[]renderers.Renderer{
			renderers.NewJSONRenderer(),
			renderers.NewXMLRenderer(),
		},
		[]parsers.Parser{
			parsers.NewJSONParser(),
		},
	)

	// Test single media type
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/json")
	renderer := negotiator.SelectRenderer(req)
	assert.NotNil(t, renderer)
	assert.Equal(t, "application/json", renderer.MediaType())

	// Test multiple media types
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Accept", "application/json, application/xml")
	renderer2 := negotiator.SelectRenderer(req2)
	assert.NotNil(t, renderer2)
	assert.Equal(t, "application/json", renderer2.MediaType()) // First match
}
