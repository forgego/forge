package api

import (
	"net/http"

	"github.com/forgego/forge/api/filters"
	"github.com/forgego/forge/api/parsers"
	"github.com/forgego/forge/api/renderers"
)

// EnhancedBaseViewSetIntegrated is the fully integrated viewset with all features
type EnhancedBaseViewSetIntegrated struct {
	*EnhancedBaseViewSet

	// Additional configuration
	RendererClasses   []renderers.Renderer
	ParserClasses     []parsers.Parser
	FilterBackends    []filters.FilterBackend
	ContentNegotiator *ContentNegotiator
}

// NewEnhancedBaseViewSetIntegrated creates a fully integrated viewset
func NewEnhancedBaseViewSetIntegrated(serializer func() Serializer, queryset, model interface{}) *EnhancedBaseViewSetIntegrated {
	base := NewEnhancedBaseViewSet(serializer, queryset, model)

	// Default renderers
	rendererList := []renderers.Renderer{
		renderers.NewJSONRenderer(),
		renderers.NewHTMLRenderer(),
	}

	// Default parsers
	parserList := []parsers.Parser{
		parsers.NewJSONParser(),
		parsers.NewFormParser(),
	}

	return &EnhancedBaseViewSetIntegrated{
		EnhancedBaseViewSet: base,
		RendererClasses:     rendererList,
		ParserClasses:       parserList,
		ContentNegotiator:   NewContentNegotiator(rendererList, parserList),
	}
}

// GetRenderers returns the renderer classes
func (vs *EnhancedBaseViewSetIntegrated) GetRenderers() []renderers.Renderer {
	if len(vs.RendererClasses) > 0 {
		return vs.RendererClasses
	}
	// Use global defaults
	settings := GetSettings()
	if len(settings.DefaultRenderers) > 0 {
		return settings.DefaultRenderers
	}
	return []renderers.Renderer{renderers.NewJSONRenderer()}
}

// GetParsers returns the parser classes
func (vs *EnhancedBaseViewSetIntegrated) GetParsers() []parsers.Parser {
	if len(vs.ParserClasses) > 0 {
		return vs.ParserClasses
	}
	// Use global defaults
	settings := GetSettings()
	if len(settings.DefaultParsers) > 0 {
		return settings.DefaultParsers
	}
	return []parsers.Parser{parsers.NewJSONParser()}
}

// GetFilterBackends returns the filter backends
func (vs *EnhancedBaseViewSetIntegrated) GetFilterBackends() []filters.FilterBackend {
	return vs.FilterBackends
}

// List handles GET /resource/ with full integration
func (vs *EnhancedBaseViewSetIntegrated) List(w http.ResponseWriter, r *http.Request) {
	vs.EnhancedBaseViewSet.List(w, r) // Delegate to base implementation
}

// Create handles POST /resource/ with full integration
func (vs *EnhancedBaseViewSetIntegrated) Create(w http.ResponseWriter, r *http.Request) {
	vs.EnhancedBaseViewSet.Create(w, r) // Delegate to base implementation
}

// Retrieve handles GET /resource/{id}/ with full integration
func (vs *EnhancedBaseViewSetIntegrated) Retrieve(w http.ResponseWriter, r *http.Request) {
	vs.EnhancedBaseViewSet.Retrieve(w, r) // Delegate to base implementation
}

// Update handles PUT /resource/{id}/ with full integration
func (vs *EnhancedBaseViewSetIntegrated) Update(w http.ResponseWriter, r *http.Request) {
	vs.EnhancedBaseViewSet.Update(w, r) // Delegate to base implementation
}

// PartialUpdate handles PATCH /resource/{id}/ with full integration
func (vs *EnhancedBaseViewSetIntegrated) PartialUpdate(w http.ResponseWriter, r *http.Request) {
	vs.EnhancedBaseViewSet.PartialUpdate(w, r) // Delegate to base implementation
}

// Destroy handles DELETE /resource/{id}/ with full integration
func (vs *EnhancedBaseViewSetIntegrated) Destroy(w http.ResponseWriter, r *http.Request) {
	vs.EnhancedBaseViewSet.Destroy(w, r) // Delegate to base implementation
}

// Ensure EnhancedBaseViewSetIntegrated implements ViewSet
var _ ViewSet = (*EnhancedBaseViewSetIntegrated)(nil)
