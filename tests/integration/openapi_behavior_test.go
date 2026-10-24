package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	forgeapi "github.com/forgego/forge/api"
	"github.com/forgego/forge/api/authentication"
	apicore "github.com/forgego/forge/api/core"
	"github.com/forgego/forge/api/docs"
	"github.com/forgego/forge/api/exceptions"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/operation"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/forgego/forge/tests/testdata/reference_app/operations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	openAPITestValidToken  = "valid-token"
	openAPITestDeniedToken = "denied-token"
)

type OpenAPIOrderModel struct {
	schema.BaseSchema
	ID         int64  `db:"id" json:"id"`
	OrderID    int64  `db:"-" json:"order_id,omitempty"`
	Status     string `db:"status" json:"status"`
	ApprovedBy string `db:"approved_by" json:"approved_by"`
}

func (OpenAPIOrderModel) Meta() schema.Meta {
	return schema.Meta{
		TableName: "orders",
	}
}

func (OpenAPIOrderModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("status", schema.Required()),
		schema.StringField("approved_by", schema.Optional()),
	}
}

type openAPIOrderSerializer struct {
	*forgeapi.BaseSerializer
}

func newOpenAPIOrderSerializer() forgeapi.Serializer {
	return &openAPIOrderSerializer{
		BaseSerializer: forgeapi.NewBaseSerializer(nil),
	}
}

func (s *openAPIOrderSerializer) Fields() []string {
	return []string{"id", "order_id", "status", "approved_by"}
}

var openAPIOrderDocsFields = []docs.Field{
	{Name: "id", Type: "integer", Required: false},
	{Name: "order_id", Type: "integer", Required: false},
	{Name: "status", Type: "string", Required: true},
	{Name: "approved_by", Type: "string", Required: false},
}

type openAPIBearerAuth struct{}

func (a *openAPIBearerAuth) Authenticate(r *http.Request) (*authentication.AuthResult, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, nil
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, nil
	}
	token := parts[1]
	switch token {
	case openAPITestValidToken:
		p := approvalPrincipal(true)
		*r = *r.WithContext(apicore.WithPrincipal(r.Context(), p))
		return authentication.NewAuthResult(p, token), nil
	case openAPITestDeniedToken:
		p := approvalPrincipal(false)
		*r = *r.WithContext(apicore.WithPrincipal(r.Context(), p))
		return authentication.NewAuthResult(p, token), nil
	default:
		return nil, authentication.ErrInvalidToken
	}
}

func (a *openAPIBearerAuth) AuthenticateHeader(_ *http.Request) string {
	return "Bearer"
}

func resolveOpenAPIActor(r *http.Request) (operation.Actor, error) {
	if p, ok := apicore.PrincipalFromContext(r.Context()); ok {
		return principal.Actor(p), nil
	}
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, exceptions.NewNotAuthenticated("Authentication credentials were not provided.")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, exceptions.NewNotAuthenticated("Authentication credentials were not provided.")
	}
	switch parts[1] {
	case openAPITestValidToken:
		return principal.Actor(approvalPrincipal(true)), nil
	case openAPITestDeniedToken:
		return principal.Actor(approvalPrincipal(false)), nil
	default:
		return nil, exceptions.NewAuthenticationFailed("Invalid token.")
	}
}

func setupOpenAPITestFixture(t *testing.T) (*approvalFixture, *docs.OpenAPISpec, http.Handler) {
	t.Helper()
	fixture := newApprovalFixture(t)

	manager, err := orm.NewManagerWithDB[OpenAPIOrderModel]("orders", fixture.database)
	require.NoError(t, err)

	vs := forgeapi.NewBaseViewSet(
		newOpenAPIOrderSerializer,
		manager,
		&OpenAPIOrderModel{},
	)
	vs.Authentication = []authentication.Authentication{&openAPIBearerAuth{}}
	vs.Permissions = []permissions.Permission{permissions.NewIsAuthenticated()}

	apiRouter := forgeapi.NewRouter("/api")
	apiRouter.Register("orders", vs)

	approveHandler := forgeapi.OperationAction(fixture.runner, operations.ApproveOrder, resolveOpenAPIActor)
	apiRouter.Action("orders", "approve", forgeapi.ActionConfig{
		Methods: []string{http.MethodPost},
		Detail:  false,
		URLPath: "approve",
	}, approveHandler)

	routes := forgeapi.DocsRoutes(apiRouter.Routes())
	for i := range routes {
		routes[i].Fields = openAPIOrderDocsFields
		routes[i].Security = []string{"bearerAuth"}
	}

	generator := docs.NewOpenAPIGenerator("t", "1").WithRoutes(routes)
	spec := generator.Generate()

	serverRouter := forgehttp.NewRouter()
	apiRouter.RegisterRoutes(serverRouter)

	return fixture, spec, serverRouter
}

var openAPIParamRegex = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

func findOpenAPIOperation(t *testing.T, spec *docs.OpenAPISpec, method, path string) *docs.Operation {
	t.Helper()
	item, ok := spec.Paths[path]
	if !ok {
		for specPath, pItem := range spec.Paths {
			pattern := "^" + openAPIParamRegex.ReplaceAllString(specPath, "[^/]+") + "$"
			if matched, _ := regexp.MatchString(pattern, path); matched {
				item = pItem
				break
			}
		}
	}
	require.NotNil(t, item, "path %s must match an OpenAPI path in spec.Paths", path)

	var op *docs.Operation
	switch strings.ToUpper(method) {
	case http.MethodGet:
		op = item.Get
	case http.MethodPost:
		op = item.Post
	case http.MethodPut:
		op = item.Put
	case http.MethodPatch:
		op = item.Patch
	case http.MethodDelete:
		op = item.Delete
	}
	require.NotNil(t, op, "operation %s %s must exist in OpenAPI spec", method, path)
	return op
}

func assertProblemBody(t *testing.T, bodyBytes []byte, expectedStatus int) {
	t.Helper()
	var prob map[string]interface{}
	require.NoError(t, json.Unmarshal(bodyBytes, &prob), "body must be valid JSON Problem: %s", string(bodyBytes))
	assert.NotEmpty(t, prob["type"], "Problem body must have 'type': %s", string(bodyBytes))
	assert.NotEmpty(t, prob["title"], "Problem body must have 'title': %s", string(bodyBytes))
	assert.Equal(t, float64(expectedStatus), prob["status"], "Problem body 'status' mismatch: %s", string(bodyBytes))
}

func TestOpenAPIBehavior(t *testing.T) {
	t.Run("TestOpenAPIBehavior_Unauthenticated", func(t *testing.T) {
		TestOpenAPIBehavior_Unauthenticated(t)
	})
	t.Run("TestOpenAPIBehavior_Denied", func(t *testing.T) {
		TestOpenAPIBehavior_Denied(t)
	})
	t.Run("TestOpenAPIBehavior_InvalidInputOnCreate", func(t *testing.T) {
		TestOpenAPIBehavior_InvalidInputOnCreate(t)
	})
	t.Run("TestOpenAPIBehavior_Success", func(t *testing.T) {
		TestOpenAPIBehavior_Success(t)
	})
}

// 1. Unauthenticated request to a secured operation -> status 401 (or what the spec lists)
// and the status is documented in that operation's responses; the body has Problem keys (type,title,status).
func TestOpenAPIBehavior_Unauthenticated(t *testing.T) {
	_, spec, handler := setupOpenAPITestFixture(t)

	testCases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "resource retrieve",
			method: http.MethodGet,
			path:   "/api/orders/1",
		},
		{
			name:   "resource create",
			method: http.MethodPost,
			path:   "/api/orders/",
			body:   `{"status":"pending"}`,
		},
		{
			name:   "approve action",
			method: http.MethodPost,
			path:   "/api/orders/approve",
			body:   `{"order_id":1}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			op := findOpenAPIOperation(t, spec, tc.method, tc.path)
			require.Contains(t, op.Responses, "401", "401 status must be documented in operation responses")

			var bodyReader *bytes.Reader
			if tc.body != "" {
				bodyReader = bytes.NewReader([]byte(tc.body))
			} else {
				bodyReader = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(tc.method, tc.path, bodyReader)
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusUnauthorized, rec.Code, "response status must be 401: %s", rec.Body.String())
			assertProblemBody(t, rec.Body.Bytes(), http.StatusUnauthorized)
		})
	}
}

// 2. Denied (authenticated, no permission) -> 403 documented, Problem body.
func TestOpenAPIBehavior_Denied(t *testing.T) {
	_, spec, handler := setupOpenAPITestFixture(t)

	op := findOpenAPIOperation(t, spec, http.MethodPost, "/api/orders/approve")
	require.Contains(t, op.Responses, "403", "403 status must be documented in operation responses")

	req := httptest.NewRequest(http.MethodPost, "/api/orders/approve", bytes.NewReader([]byte(`{"order_id":1}`)))
	req.Header.Set("Authorization", "Bearer "+openAPITestDeniedToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code, "response status must be 403: %s", rec.Body.String())
	assertProblemBody(t, rec.Body.Bytes(), http.StatusForbidden)
}

// 3. Invalid input on create -> 400 documented, Problem body.
func TestOpenAPIBehavior_InvalidInputOnCreate(t *testing.T) {
	_, spec, handler := setupOpenAPITestFixture(t)

	op := findOpenAPIOperation(t, spec, http.MethodPost, "/api/orders/")
	require.Contains(t, op.Responses, "400", "400 status must be documented in operation responses")

	testCases := []struct {
		name string
		body string
	}{
		{
			name: "malformed JSON",
			body: `{"status": `,
		},
		{
			name: "missing required field",
			body: `{}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/orders/", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Authorization", "Bearer "+openAPITestValidToken)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Code, "response status must be 400: %s", rec.Body.String())
			assertProblemBody(t, rec.Body.Bytes(), http.StatusBadRequest)
		})
	}
}

// 4. Success list/retrieve/create and the approve action -> documented success code;
// every JSON key in the response body exists in the documented response schema properties
// (resolve $ref in components.schemas).
func TestOpenAPIBehavior_Success(t *testing.T) {
	_, spec, handler := setupOpenAPITestFixture(t)

	testCases := []struct {
		name       string
		method     string
		path       string
		body       string
		expectCode int
	}{
		{
			name:       "list",
			method:     http.MethodGet,
			path:       "/api/orders/",
			expectCode: http.StatusOK,
		},
		{
			name:       "retrieve",
			method:     http.MethodGet,
			path:       "/api/orders/1",
			expectCode: http.StatusOK,
		},
		{
			name:       "create",
			method:     http.MethodPost,
			path:       "/api/orders/",
			body:       `{"status":"pending"}`,
			expectCode: http.StatusCreated,
		},
		{
			name:       "approve action",
			method:     http.MethodPost,
			path:       "/api/orders/approve",
			body:       `{"order_id":1}`,
			expectCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			op := findOpenAPIOperation(t, spec, tc.method, tc.path)

			codeStr := strconv.Itoa(tc.expectCode)
			respDef, ok := op.Responses[codeStr]
			require.True(t, ok, "documented success code %s must exist in operation responses", codeStr)
			require.NotNil(t, respDef.Content, "success response must have content")
			mediaType, ok := respDef.Content["application/json"]
			require.True(t, ok, "success response must define application/json content")
			require.NotNil(t, mediaType.Schema, "media type must have schema")

			// Resolve $ref in components.schemas
			ref := mediaType.Schema.Ref
			require.NotEmpty(t, ref, "schema must reference a component schema via $ref")
			const schemaPrefix = "#/components/schemas/"
			require.True(t, strings.HasPrefix(ref, schemaPrefix), "$ref %s must start with %s", ref, schemaPrefix)
			schemaName := strings.TrimPrefix(ref, schemaPrefix)

			resolvedSchema, ok := spec.Components.Schemas[schemaName]
			require.True(t, ok, "resolved schema %s must exist in components.schemas", schemaName)
			require.NotNil(t, resolvedSchema.Properties, "resolved schema %s must have properties", schemaName)

			// Perform real HTTP request
			var bodyReader *bytes.Reader
			if tc.body != "" {
				bodyReader = bytes.NewReader([]byte(tc.body))
			} else {
				bodyReader = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(tc.method, tc.path, bodyReader)
			req.Header.Set("Authorization", "Bearer "+openAPITestValidToken)
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			require.Equal(t, tc.expectCode, rec.Code, "HTTP status must match documented success code: %s", rec.Body.String())

			var bodyMap map[string]interface{}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &bodyMap), "response body must be valid JSON object: %s", rec.Body.String())
			require.NotEmpty(t, bodyMap, "response body must not be empty")

			// Every JSON key in the response body exists in the documented response schema properties
			for key := range bodyMap {
				assert.Contains(t, resolvedSchema.Properties, key, "key %q in response body must exist in documented schema %s properties (%v)", key, schemaName, resolvedSchema.Properties)
			}
		})
	}
}
