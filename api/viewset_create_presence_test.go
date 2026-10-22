package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
)

type requiredScalarRequestModel struct {
	schema.BaseSchema
	Count   int64 `json:"count" db:"count"`
	Enabled bool  `json:"enabled" db:"enabled"`
}

func (requiredScalarRequestModel) Fields() []schema.Field {
	return []schema.Field{schema.Int64Field("count", schema.Required()), schema.BoolField("enabled", schema.Required())}
}

type requiredScalarManager struct{ calls int }

func (m *requiredScalarManager) Create(context.Context, interface{}) error { m.calls++; return nil }

func TestCreateRequiresScalarPresenceWithoutRejectingZeroAndFalse(t *testing.T) {
	for _, test := range []struct {
		name, body string
		status     int
	}{
		{"present zero and false", `{"count":0,"enabled":false}`, http.StatusCreated},
		{"null number", `{"count":null,"enabled":false}`, http.StatusBadRequest},
		{"null boolean", `{"count":0,"enabled":null}`, http.StatusBadRequest},
		{"both omitted", `{}`, http.StatusBadRequest},
		{"number omitted", `{"enabled":false}`, http.StatusBadRequest},
		{"boolean omitted", `{"count":0}`, http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			manager := &requiredScalarManager{}
			vs := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, manager, &requiredScalarRequestModel{})
			vs.Authentication = []authentication.Authentication{}
			vs.Permissions = []permissions.Permission{}
			response := httptest.NewRecorder()
			vs.Create(response, httptest.NewRequest(http.MethodPost, "/items/", strings.NewReader(test.body)))
			assert.Equal(t, test.status, response.Code, response.Body.String())
			if test.status == http.StatusBadRequest {
				assert.Zero(t, manager.calls)
			} else {
				assert.Equal(t, 1, manager.calls)
			}
		})
	}
}

// DBDefault fields may be omitted on create: the manager leaves their zero
// values out of INSERT so PostgreSQL can fill them. PUT still requires them.
type requiredDatabaseDefaultModel struct {
	schema.BaseSchema
	Count   int64 `json:"count" db:"count"`
	Enabled bool  `json:"enabled" db:"enabled"`
}

func (requiredDatabaseDefaultModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("count", schema.Required(), schema.DBDefault("7")),
		schema.BoolField("enabled", schema.Required(), schema.DBDefault("true")),
	}
}
func TestCreateAllowsOmittedRequiredDatabaseDefaults(t *testing.T) {
	manager := &requiredScalarManager{}
	vs := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, manager, &requiredDatabaseDefaultModel{})
	vs.Authentication = []authentication.Authentication{}
	vs.Permissions = []permissions.Permission{}
	response := httptest.NewRecorder()
	vs.Create(response, httptest.NewRequest(http.MethodPost, "/items/", strings.NewReader(`{}`)))
	assert.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	assert.Equal(t, 1, manager.calls)
}
