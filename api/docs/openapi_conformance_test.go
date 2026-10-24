package docs_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/api/docs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type conformanceTestModel struct {
	ID int64 `json:"id"`
}

type conformanceTestQueryset struct {
	items []*conformanceTestModel
}

func (q *conformanceTestQueryset) All(context.Context) ([]*conformanceTestModel, error) {
	return q.items, nil
}

func (q *conformanceTestQueryset) Count(context.Context) (int64, error) {
	return int64(len(q.items)), nil
}

func (q *conformanceTestQueryset) Offset(int) *conformanceTestQueryset       { return q }
func (q *conformanceTestQueryset) Limit(int) *conformanceTestQueryset        { return q }
func (q *conformanceTestQueryset) Create(context.Context, interface{}) error { return nil }
func (q *conformanceTestQueryset) Update(context.Context, interface{}) error { return nil }
func (q *conformanceTestQueryset) Delete(context.Context, interface{}) error { return nil }

func (q *conformanceTestQueryset) Get(_ context.Context, id int64) (*conformanceTestModel, error) {
	for _, item := range q.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

type conformanceTestSerializer struct {
	*api.BaseSerializer
}

func newConformanceTestSerializer() api.Serializer {
	return &conformanceTestSerializer{BaseSerializer: api.NewBaseSerializer(make(map[string]interface{}))}
}

func (s *conformanceTestSerializer) Fields() []string         { return []string{"id"} }
func (s *conformanceTestSerializer) ReadOnlyFields() []string { return []string{"id"} }

func createConformanceTestViewSet() *api.BaseViewSet {
	return api.NewBaseViewSet(
		newConformanceTestSerializer,
		&conformanceTestQueryset{items: []*conformanceTestModel{{ID: 42}}},
		&conformanceTestModel{},
	)
}

type routeKey struct {
	Method string
	Path   string
}

func setupConformanceSpec(t *testing.T) (*api.Router, *docs.OpenAPISpec) {
	router := api.NewRouter("/api/v1")
	router.Register("items", createConformanceTestViewSet())
	router.Register("categories", createConformanceTestViewSet())

	router.Action("items", "publish", api.ActionConfig{
		Methods: []string{http.MethodPost},
		Detail:  true,
	}, func(w http.ResponseWriter, r *http.Request) {})

	routes := api.DocsRoutes(router.Routes())
	fields := []docs.Field{
		{Name: "id", Type: "integer", Required: false, WriteOnly: false, Hidden: false},
		{Name: "name", Type: "string", Required: true, WriteOnly: false, Hidden: false},
		{Name: "password", Type: "string", Required: true, WriteOnly: true, Hidden: false},
		{Name: "secret", Type: "string", Required: false, WriteOnly: false, Hidden: true},
	}
	for i := range routes {
		routes[i].Fields = fields
		routes[i].Security = []string{"bearerAuth"}
	}

	generator := docs.NewOpenAPIGenerator("Shop API", "1.0.0").WithRoutes(routes)
	generator.Description = "Catalog and orders"
	spec := generator.Generate()
	return router, spec
}

func TestOpenAPIConformance_PathsMatchRoutes(t *testing.T) {
	router, spec := setupConformanceSpec(t)

	specRoutes := make(map[routeKey]bool)
	for path, item := range spec.Paths {
		if item.Get != nil {
			specRoutes[routeKey{Method: "GET", Path: path}] = true
		}
		if item.Post != nil {
			specRoutes[routeKey{Method: "POST", Path: path}] = true
		}
		if item.Put != nil {
			specRoutes[routeKey{Method: "PUT", Path: path}] = true
		}
		if item.Patch != nil {
			specRoutes[routeKey{Method: "PATCH", Path: path}] = true
		}
		if item.Delete != nil {
			specRoutes[routeKey{Method: "DELETE", Path: path}] = true
		}
		if item.Options != nil {
			specRoutes[routeKey{Method: "OPTIONS", Path: path}] = true
		}
		if item.Head != nil {
			specRoutes[routeKey{Method: "HEAD", Path: path}] = true
		}
	}

	routerRoutes := make(map[routeKey]bool)
	for _, d := range router.Routes() {
		routerRoutes[routeKey{Method: d.Method, Path: d.Path}] = true
	}

	assert.Equal(t, routerRoutes, specRoutes, "spec paths must match router routes 100% in both directions")
}

func TestOpenAPIConformance_HiddenAndWriteOnlyAbsentFromResponse(t *testing.T) {
	_, spec := setupConformanceSpec(t)

	require.NotNil(t, spec.Components)
	require.NotNil(t, spec.Components.Schemas)

	// Check response schema: lacks password and secret
	respSchema, ok := spec.Components.Schemas["items"]
	require.True(t, ok, "items response schema must exist in components.schemas")
	assert.NotContains(t, respSchema.Properties, "password", "response schema must lack WriteOnly field 'password'")
	assert.NotContains(t, respSchema.Properties, "secret", "response schema must lack Hidden field 'secret'")
	assert.Contains(t, respSchema.Properties, "name", "response schema must include normal field 'name'")
	assert.Contains(t, respSchema.Properties, "id", "response schema must include normal field 'id'")

	// Check input schema: has password but not secret
	inputSchema, ok := spec.Components.Schemas["itemsInput"]
	require.True(t, ok, "itemsInput schema must exist in components.schemas")
	assert.Contains(t, inputSchema.Properties, "password", "input schema must include WriteOnly field 'password'")
	assert.NotContains(t, inputSchema.Properties, "secret", "input schema must lack Hidden field 'secret'")
	assert.Contains(t, inputSchema.Properties, "name", "input schema must include normal field 'name'")
	assert.Contains(t, inputSchema.Properties, "id", "input schema must include normal field 'id'")
	assert.Contains(t, inputSchema.Required, "password", "required WriteOnly field must be listed in Required")
	assert.Contains(t, inputSchema.Required, "name", "required field must be listed in Required")
}

func TestOpenAPIConformance_ErrorsAndSecurity(t *testing.T) {
	_, spec := setupConformanceSpec(t)

	require.NotNil(t, spec.Components)
	require.NotNil(t, spec.Components.Schemas)
	require.NotNil(t, spec.Components.SecuritySchemes)

	// Problem schema must exist in components
	problemSchema, ok := spec.Components.Schemas["Problem"]
	require.True(t, ok, "Problem schema must exist in components.schemas")
	for _, prop := range []string{"type", "title", "status", "detail", "instance", "request_id"} {
		assert.Contains(t, problemSchema.Properties, prop, "Problem schema must have %s property", prop)
	}

	// bearerAuth must exist in securitySchemes
	assert.Contains(t, spec.Components.SecuritySchemes, "bearerAuth")

	// Verify all operations
	var operations []*docs.Operation
	for _, item := range spec.Paths {
		for _, op := range []*docs.Operation{
			item.Get, item.Post, item.Put, item.Patch, item.Delete, item.Options, item.Head,
		} {
			if op != nil {
				operations = append(operations, op)
			}
		}
	}
	require.NotEmpty(t, operations)

	expectedErrors := []string{"400", "401", "403", "404", "500"}
	for _, op := range operations {
		// Verify unified errors referencing Problem
		for _, code := range expectedErrors {
			resp, ok := op.Responses[code]
			require.True(t, ok, "operation must have %s response", code)
			require.NotNil(t, resp.Content, "error response %s must have content", code)
			mediaType, ok := resp.Content["application/json"]
			require.True(t, ok, "error response %s must have application/json", code)
			require.NotNil(t, mediaType.Schema, "error response %s schema must not be nil", code)
			assert.Equal(t, "#/components/schemas/Problem", mediaType.Schema.Ref)
		}

		// Verify security is bearerAuth
		require.NotEmpty(t, op.Security, "operation must have security")
		foundBearer := false
		for _, secReq := range op.Security {
			if _, ok := secReq["bearerAuth"]; ok {
				foundBearer = true
				break
			}
		}
		assert.True(t, foundBearer, "operation security must include bearerAuth")
	}
}

func TestOpenAPIConformance_NoRoutesUnchanged(t *testing.T) {
	freshGen := docs.NewOpenAPIGenerator("Shop API", "1.0.0")
	freshGen.Description = "Catalog and orders"

	noRoutesGen := docs.NewOpenAPIGenerator("Shop API", "1.0.0")
	noRoutesGen.Description = "Catalog and orders"

	specWithoutWithRoutes := noRoutesGen.Generate()
	assert.Empty(t, specWithoutWithRoutes.Paths)

	specFresh := freshGen.Generate()

	freshJSON, err := json.Marshal(specFresh)
	require.NoError(t, err)

	noRoutesJSON, err := json.Marshal(specWithoutWithRoutes)
	require.NoError(t, err)

	assert.JSONEq(t, string(freshJSON), string(noRoutesJSON))

	// Calling WithRoutes with empty routes also yields empty Paths and same JSON
	emptyRoutesGen := docs.NewOpenAPIGenerator("Shop API", "1.0.0").WithRoutes([]docs.Route{})
	emptyRoutesGen.Description = "Catalog and orders"
	emptyRoutesJSON, err := json.Marshal(emptyRoutesGen.Generate())
	require.NoError(t, err)
	assert.JSONEq(t, string(freshJSON), string(emptyRoutesJSON))
}
