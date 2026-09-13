package docs

import (
	"encoding/json"
	"net/http"
)

// Metadata is the interface for API metadata
type Metadata interface {
	// DetermineMetadata determines metadata for a view
	DetermineMetadata(r *http.Request, view interface{}) map[string]interface{}
}

// SimpleMetadata provides simple metadata
type SimpleMetadata struct{}

// NewSimpleMetadata creates a new simple metadata provider
func NewSimpleMetadata() *SimpleMetadata {
	return &SimpleMetadata{}
}

// DetermineMetadata determines metadata for a view
func (m *SimpleMetadata) DetermineMetadata(r *http.Request, view interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":        "API",
		"description": "REST API",
	}
}

// OptionsHandler handles OPTIONS requests for API metadata
func OptionsHandler(w http.ResponseWriter, r *http.Request, view interface{}, metadata Metadata) {
	if metadata == nil {
		metadata = NewSimpleMetadata()
	}

	meta := metadata.DetermineMetadata(r, view)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Return metadata as JSON
	// In a full implementation, would use JSON encoder
	_ = meta
}

// OpenAPISpec represents an OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string               `json:"openapi"`
	Info       *Info                `json:"info"`
	Servers    []*Server            `json:"servers,omitempty"`
	Paths      map[string]*PathItem `json:"paths"`
	Components *Components          `json:"components,omitempty"`
}

// Info represents API information
type Info struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// Server represents a server URL
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// PathItem represents a path item in OpenAPI
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

// Operation represents an API operation
type Operation struct {
	Summary     string               `json:"summary,omitempty"`
	Description string               `json:"description,omitempty"`
	Tags        []string             `json:"tags,omitempty"`
	Parameters  []*Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody         `json:"requestBody,omitempty"`
	Responses   map[string]*Response `json:"responses"`
}

// Parameter represents a path/query parameter
type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"` // path, query, header, cookie
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

// RequestBody represents a request body
type RequestBody struct {
	Description string                `json:"description,omitempty"`
	Required    bool                  `json:"required,omitempty"`
	Content     map[string]*MediaType `json:"content"`
}

// Response represents a response
type Response struct {
	Description string                `json:"description"`
	Content     map[string]*MediaType `json:"content,omitempty"`
}

// MediaType represents a media type
type MediaType struct {
	Schema *Schema `json:"schema,omitempty"`
}

// Schema represents a JSON schema
type Schema struct {
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Description string             `json:"description,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Required    []string           `json:"required,omitempty"`
}

// Components represents OpenAPI components
type Components struct {
	Schemas map[string]*Schema `json:"schemas,omitempty"`
}

// OpenAPIGenerator generates OpenAPI specifications
type OpenAPIGenerator struct {
	Title       string
	Version     string
	Description string
	BaseURL     string
}

// NewOpenAPIGenerator creates a new OpenAPI generator
func NewOpenAPIGenerator(title, version string) *OpenAPIGenerator {
	return &OpenAPIGenerator{
		Title:   title,
		Version: version,
	}
}

// Generate generates an OpenAPI specification
func (g *OpenAPIGenerator) Generate() *OpenAPISpec {
	return &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:       g.Title,
			Version:     g.Version,
			Description: g.Description,
		},
		Paths: make(map[string]*PathItem),
		Components: &Components{
			Schemas: make(map[string]*Schema),
		},
	}
}

// Handler returns an HTTP handler for serving OpenAPI spec
func (g *OpenAPIGenerator) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		spec := g.Generate()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(spec)
	}
}
