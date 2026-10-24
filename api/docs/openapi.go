package docs

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
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
	Get        *Operation   `json:"get,omitempty"`
	Post       *Operation   `json:"post,omitempty"`
	Put        *Operation   `json:"put,omitempty"`
	Patch      *Operation   `json:"patch,omitempty"`
	Delete     *Operation   `json:"delete,omitempty"`
	Options    *Operation   `json:"options,omitempty"`
	Head       *Operation   `json:"head,omitempty"`
	Trace      *Operation   `json:"trace,omitempty"`
	Parameters []*Parameter `json:"parameters,omitempty"`
}

// Operation represents an API operation
type Operation struct {
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Parameters  []*Parameter          `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]*Response  `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
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
	Ref         string             `json:"$ref,omitempty"`
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Description string             `json:"description,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Required    []string           `json:"required,omitempty"`
}

// SecurityScheme represents a security scheme in OpenAPI
type SecurityScheme struct {
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`
	In           string `json:"in,omitempty"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

// Components represents OpenAPI components
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
}

// Field represents a schema field definition. Type is the JSON schema type:
// string/integer/number/boolean/array/object.
type Field struct {
	Name      string
	Type      string
	Required  bool
	WriteOnly bool
	Hidden    bool
}

// Route represents an endpoint route definition for OpenAPI generation.
type Route struct {
	Method   string
	Path     string
	Resource string
	Action   string
	Detail   bool
	Fields   []Field
	Security []string
}

// OpenAPIGenerator generates OpenAPI specifications
type OpenAPIGenerator struct {
	Title       string
	Version     string
	Description string
	BaseURL     string
	routes      []Route
}

// NewOpenAPIGenerator creates a new OpenAPI generator
func NewOpenAPIGenerator(title, version string) *OpenAPIGenerator {
	return &OpenAPIGenerator{
		Title:   title,
		Version: version,
	}
}

// WithRoutes stores the routes on an unexported field and returns g.
func (g *OpenAPIGenerator) WithRoutes(routes []Route) *OpenAPIGenerator {
	g.routes = routes
	return g
}

var pathParamRegex = regexp.MustCompile(`\{([a-zA-Z0-9_]+)(?::[^}]+)?\}`)

func parsePath(p string) (string, []string) {
	var paramNames []string
	cleanPath := pathParamRegex.ReplaceAllStringFunc(p, func(match string) string {
		sub := pathParamRegex.FindStringSubmatch(match)
		if len(sub) > 1 {
			paramNames = append(paramNames, sub[1])
			return "{" + sub[1] + "}"
		}
		return match
	})
	return cleanPath, paramNames
}

func defaultSecurityScheme(name string) *SecurityScheme {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "bearer"):
		return &SecurityScheme{
			Type:   "http",
			Scheme: "bearer",
		}
	case strings.Contains(lower, "basic"):
		return &SecurityScheme{
			Type:   "http",
			Scheme: "basic",
		}
	case strings.Contains(lower, "api") || strings.Contains(lower, "key"):
		return &SecurityScheme{
			Type: "apiKey",
			Name: "Authorization",
			In:   "header",
		}
	default:
		return &SecurityScheme{
			Type:   "http",
			Scheme: name,
		}
	}
}

func statusCodeInt(code string) int {
	switch code {
	case "200":
		return http.StatusOK
	case "201":
		return http.StatusCreated
	case "204":
		return http.StatusNoContent
	case "400":
		return http.StatusBadRequest
	case "401":
		return http.StatusUnauthorized
	case "403":
		return http.StatusForbidden
	case "404":
		return http.StatusNotFound
	case "500":
		return http.StatusInternalServerError
	default:
		return http.StatusOK
	}
}

// Generate generates an OpenAPI specification
func (g *OpenAPIGenerator) Generate() *OpenAPISpec {
	spec := &OpenAPISpec{
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

	if len(g.routes) == 0 {
		return spec
	}

	// Unified errors: 400, 401, 403, 404, 500 reference Problem schema
	spec.Components.Schemas["Problem"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"type":       {Type: "string"},
			"title":      {Type: "string"},
			"status":     {Type: "integer"},
			"detail":     {Type: "string"},
			"instance":   {Type: "string"},
			"request_id": {Type: "string"},
		},
	}

	// Deterministic resource and field tracking in order of appearance
	var resourceOrder []string
	resourceFields := make(map[string][]Field)
	seenField := make(map[string]map[string]bool)

	for _, route := range g.routes {
		if route.Resource == "" {
			continue
		}
		if _, exists := resourceFields[route.Resource]; !exists {
			resourceFields[route.Resource] = nil
			resourceOrder = append(resourceOrder, route.Resource)
			seenField[route.Resource] = make(map[string]bool)
		}
		for _, f := range route.Fields {
			if !seenField[route.Resource][f.Name] {
				seenField[route.Resource][f.Name] = true
				resourceFields[route.Resource] = append(resourceFields[route.Resource], f)
			}
		}
	}

	// Build component schemas for each resource
	for _, res := range resourceOrder {
		fields := resourceFields[res]

		// Response schema: neither WriteOnly nor Hidden
		respProps := make(map[string]*Schema)
		for _, f := range fields {
			if !f.WriteOnly && !f.Hidden {
				respProps[f.Name] = &Schema{Type: f.Type}
			}
		}
		spec.Components.Schemas[res] = &Schema{
			Type:       "object",
			Properties: respProps,
		}

		// Input schema: not Hidden (WriteOnly included), Required listed
		inputProps := make(map[string]*Schema)
		var required []string
		for _, f := range fields {
			if !f.Hidden {
				inputProps[f.Name] = &Schema{Type: f.Type}
				if f.Required {
					required = append(required, f.Name)
				}
			}
		}
		inputSchema := &Schema{
			Type:       "object",
			Properties: inputProps,
		}
		if len(required) > 0 {
			inputSchema.Required = required
		}
		spec.Components.Schemas[res+"Input"] = inputSchema

		// List response schema: paginated results
		spec.Components.Schemas[res+"List"] = &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"count":    {Type: "integer"},
				"next":     {Type: "string"},
				"previous": {Type: "string"},
				"results": {
					Type: "array",
					Items: &Schema{
						Ref: "#/components/schemas/" + res,
					},
				},
			},
		}
	}

	// Build PathItems and operations in given route order
	for _, route := range g.routes {
		cleanPath, pathParamNames := parsePath(route.Path)

		pathItem, exists := spec.Paths[cleanPath]
		if !exists {
			pathItem = &PathItem{}
			spec.Paths[cleanPath] = pathItem
		}

		op := &Operation{
			Responses: make(map[string]*Response),
		}

		// Path parameters: convert chi {id} as-is and add parameter for each {name}
		if len(pathParamNames) > 0 {
			for _, pName := range pathParamNames {
				pType := "string"
				for _, f := range route.Fields {
					if f.Name == pName && f.Type != "" {
						pType = f.Type
						break
					}
				}
				op.Parameters = append(op.Parameters, &Parameter{
					Name:     pName,
					In:       "path",
					Required: true,
					Schema: &Schema{
						Type: pType,
					},
				})
			}
		}

		methodUpper := strings.ToUpper(route.Method)

		// Request body for POST/PUT/PATCH
		if methodUpper == http.MethodPost || methodUpper == http.MethodPut || methodUpper == http.MethodPatch {
			if route.Resource != "" {
				op.RequestBody = &RequestBody{
					Required: true,
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{
								Ref: "#/components/schemas/" + route.Resource + "Input",
							},
						},
					},
				}
			}
		}

		// Success response: 200 (201 for POST create, 204 for DELETE) referencing response schema
		successCode := "200"
		if methodUpper == http.MethodPost && (route.Action == "create" || route.Action == "") {
			successCode = "201"
		} else if methodUpper == http.MethodDelete {
			successCode = "204"
		}

		successResp := &Response{
			Description: http.StatusText(statusCodeInt(successCode)),
		}
		if route.Resource != "" {
			ref := "#/components/schemas/" + route.Resource
			if route.Action == "list" {
				ref = "#/components/schemas/" + route.Resource + "List"
			}
			successResp.Content = map[string]*MediaType{
				"application/json": {
					Schema: &Schema{
						Ref: ref,
					},
				},
			}
		}
		op.Responses[successCode] = successResp

		// Unified errors: 400, 401, 403, 404, 500 referencing Problem component schema
		for _, errCode := range []string{"400", "401", "403", "404", "500"} {
			op.Responses[errCode] = &Response{
				Description: http.StatusText(statusCodeInt(errCode)),
				Content: map[string]*MediaType{
					"application/json": {
						Schema: &Schema{
							Ref: "#/components/schemas/Problem",
						},
					},
				},
			}
		}

		// Security per operation from Route.Security, and each scheme added to components.securitySchemes
		if len(route.Security) > 0 {
			op.Security = make([]map[string][]string, 0, len(route.Security))
			for _, scheme := range route.Security {
				if scheme != "" {
					op.Security = append(op.Security, map[string][]string{
						scheme: {},
					})
					if spec.Components.SecuritySchemes == nil {
						spec.Components.SecuritySchemes = make(map[string]*SecurityScheme)
					}
					if _, ok := spec.Components.SecuritySchemes[scheme]; !ok {
						spec.Components.SecuritySchemes[scheme] = defaultSecurityScheme(scheme)
					}
				}
			}
		}

		// Slot operation into pathItem with method lower-cased
		switch methodUpper {
		case http.MethodGet:
			pathItem.Get = op
		case http.MethodPost:
			pathItem.Post = op
		case http.MethodPut:
			pathItem.Put = op
		case http.MethodPatch:
			pathItem.Patch = op
		case http.MethodDelete:
			pathItem.Delete = op
		case http.MethodOptions:
			pathItem.Options = op
		case http.MethodHead:
			pathItem.Head = op
		case "TRACE":
			pathItem.Trace = op
		}
	}

	return spec
}

// Handler returns an HTTP handler for serving OpenAPI spec
func (g *OpenAPIGenerator) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		spec := g.Generate()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(spec)
	}
}
