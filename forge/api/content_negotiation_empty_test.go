package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/parsers"
	"github.com/forgego/forge/api/renderers"
	"github.com/stretchr/testify/assert"
)

func TestContentNegotiator_EmptyConfigurations(t *testing.T) {
	negotiator := NewContentNegotiator([]renderers.Renderer{}, []parsers.Parser{})

	t.Run("empty renderers select renderer without accept header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		assert.NotPanics(t, func() {
			renderer := negotiator.SelectRenderer(req)
			assert.Nil(t, renderer)
		})
	})

	t.Run("empty renderers select renderer with accept header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept", "application/json")
		assert.NotPanics(t, func() {
			renderer := negotiator.SelectRenderer(req)
			assert.Nil(t, renderer)
		})
	})

	t.Run("empty parsers select parser without content-type header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		assert.NotPanics(t, func() {
			parser := negotiator.SelectParser(req)
			assert.Nil(t, parser)
		})
	})

	t.Run("empty parsers select parser with content-type header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Content-Type", "application/json")
		assert.NotPanics(t, func() {
			parser := negotiator.SelectParser(req)
			assert.Nil(t, parser)
		})
	})
}

func TestContentNegotiator_TableDrivenSelection(t *testing.T) {
	tests := []struct {
		name              string
		renderers         []renderers.Renderer
		parsers           []parsers.Parser
		headerKey         string
		headerVal         string
		expectNilRenderer bool
		expectNilParser   bool
		expectedRenderer  string
		expectedParser    string
	}{
		{
			name:              "empty renderers with empty accept",
			renderers:         []renderers.Renderer{},
			parsers:           []parsers.Parser{parsers.NewJSONParser()},
			headerKey:         "Accept",
			headerVal:         "",
			expectNilRenderer: true,
			expectedParser:    "application/json",
		},
		{
			name:             "empty parsers with empty content-type",
			renderers:        []renderers.Renderer{renderers.NewJSONRenderer()},
			parsers:          []parsers.Parser{},
			headerKey:        "Content-Type",
			headerVal:        "",
			expectedRenderer: "application/json",
			expectNilParser:  true,
		},
		{
			name:             "non-empty renderers with unmatched accept defaults to first",
			renderers:        []renderers.Renderer{renderers.NewJSONRenderer(), renderers.NewXMLRenderer()},
			parsers:          []parsers.Parser{parsers.NewJSONParser()},
			headerKey:        "Accept",
			headerVal:        "text/plain",
			expectedRenderer: "application/json",
		},
		{
			name:           "non-empty parsers with unmatched content-type defaults to first",
			renderers:      []renderers.Renderer{renderers.NewJSONRenderer()},
			parsers:        []parsers.Parser{parsers.NewJSONParser(), parsers.NewFormParser()},
			headerKey:      "Content-Type",
			headerVal:      "text/plain",
			expectedParser: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cn := NewContentNegotiator(tt.renderers, tt.parsers)
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			if tt.headerKey != "" && tt.headerVal != "" {
				req.Header.Set(tt.headerKey, tt.headerVal)
			}

			renderer := cn.SelectRenderer(req)
			if tt.expectNilRenderer {
				assert.Nil(t, renderer)
			} else if tt.expectedRenderer != "" {
				assert.NotNil(t, renderer)
				assert.Equal(t, tt.expectedRenderer, renderer.MediaType())
			}

			parser := cn.SelectParser(req)
			if tt.expectNilParser {
				assert.Nil(t, parser)
			} else if tt.expectedParser != "" {
				assert.NotNil(t, parser)
				assert.Equal(t, tt.expectedParser, parser.MediaType())
			}
		})
	}
}
