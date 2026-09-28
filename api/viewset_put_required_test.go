package api

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"sync"
	"testing"

	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// putRequiredModel has required fields that a PUT body cannot or need not
// carry: one the serializer marks read-only with the ReadonlyFields
// spelling, one hidden by json:"-" but serialized (so never request
// writable), and one schema field without a struct field.
type putRequiredModel struct {
	schema.BaseSchema
	ID     int64  `json:"id" db:"id"`
	Title  string `json:"title" db:"title"`
	Slug   string `json:"slug" db:"slug"`
	Secret string `json:"-" db:"secret"`
	Label  string `json:",omitempty" db:"label"`
}

func (putRequiredModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
		schema.StringField("slug", schema.Required()),
		schema.StringField("secret", schema.Required()),
		schema.StringField("ghost", schema.Required()),
		schema.StringField("label"),
	}
}

// putRequiredManager stores copies of one model type in memory.
type putRequiredManager struct {
	unusedDeleteOperation
	unusedListOperations
	mu    sync.Mutex
	items map[int64]interface{}
}

func (m *putRequiredManager) Create(_ context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[2] = model
	return nil
}

func (m *putRequiredManager) Get(_ context.Context, id int64) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok {
		return nil, errors.New("not found")
	}
	clone := reflect.New(reflect.TypeOf(item).Elem())
	clone.Elem().Set(reflect.ValueOf(item).Elem())
	return clone.Interface(), nil
}

func (m *putRequiredManager) Update(_ context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := m.items[1]
	reflect.ValueOf(stored).Elem().Set(reflect.ValueOf(model).Elem())
	return nil
}

// readonlySpellingSerializer declares read-only fields with the documented
// ReadonlyFields spelling, not ReadOnlyFields.
type readonlySpellingSerializer struct {
	*BaseSerializer
}

func (*readonlySpellingSerializer) ReadonlyFields() []string { return []string{"slug"} }

func newPutRequiredHandler(t *testing.T, seed interface{}, strict bool) (*putRequiredManager, http.Handler) {
	t.Helper()
	mgr := &putRequiredManager{items: map[int64]interface{}{1: seed}}
	vs := NewBaseViewSet(func() Serializer {
		return &readonlySpellingSerializer{BaseSerializer: NewBaseSerializer(nil)}
	}, mgr, reflect.New(reflect.TypeOf(seed).Elem()).Interface())
	vs.RejectUnknownRequestFields = strict
	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	return mgr, handler
}

func seededPutRequiredModel() *putRequiredModel {
	return &putRequiredModel{ID: 1, Title: "old", Slug: "fixed", Secret: "kept", Label: "tag"}
}

// A required field the serializer lists in ReadonlyFields is dropped from
// input, so PUT must not require it.
func TestUpdatePut_ReadonlyFieldsSpellingIsNotRequired(t *testing.T) {
	for _, strict := range []bool{false, true} {
		mgr, handler := newPutRequiredHandler(t, &putSlugModel{ID: 1, Title: "old", Slug: "fixed"}, strict)
		rec := performSchemaInputRequest(t, handler, http.MethodPut, "/api/items/1", `{"title":"new"}`)
		require.Equal(t, http.StatusOK, rec.Code, "strict=%v: %s", strict, rec.Body.String())
		rec = performSchemaInputRequest(t, handler, http.MethodPut, "/api/items/1", `{"title":"new","slug":"ignored"}`)
		require.Equal(t, http.StatusOK, rec.Code, "strict=%v: %s", strict, rec.Body.String())
		stored := mgr.items[1].(*putSlugModel)
		assert.Equal(t, "new", stored.Title)
		assert.Equal(t, "fixed", stored.Slug, "a read-only field keeps its stored value")
	}
}

// Create ignores ReadonlyFields under every name of the field, not only the
// exact key stripReadOnlyInput removes.
func TestCreate_ReadonlyFieldsSpellingIgnoresAliases(t *testing.T) {
	mgr, handler := newPutRequiredHandler(t, &putOptionalSlugModel{ID: 1, Title: "old", Slug: "fixed"}, false)
	rec := performSchemaInputRequest(t, handler, http.MethodPost, "/api/items", `{"title":"new","Slug":"forced"}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Empty(t, mgr.items[2].(*putOptionalSlugModel).Slug)
}

type putOptionalSlugModel struct {
	schema.BaseSchema
	ID    int64  `json:"id" db:"id"`
	Title string `json:"title" db:"title"`
	Slug  string `json:"slug" db:"slug"`
}

func (putOptionalSlugModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
		schema.StringField("slug"),
	}
}

// Required fields PUT can never write (json:"-" without being write-only,
// or a schema field without a struct field) are not required, so a strict
// PUT stays possible.
func TestUpdatePut_UnwritableRequiredFieldsAreNotRequired(t *testing.T) {
	for _, strict := range []bool{false, true} {
		mgr, handler := newPutRequiredHandler(t, seededPutRequiredModel(), strict)
		rec := performSchemaInputRequest(t, handler, http.MethodPut, "/api/items/1", `{"title":"new"}`)
		require.Equal(t, http.StatusOK, rec.Code, "strict=%v: %s", strict, rec.Body.String())
		assert.Equal(t, "kept", mgr.items[1].(*putRequiredModel).Secret)
	}
}

// populateFromMap looks keys up case-sensitively, so a key that differs
// only in case does not supply a required field.
func TestUpdatePut_RequiredFieldMatchIsCaseSensitive(t *testing.T) {
	mgr, handler := newPutRequiredHandler(t, &putTitleModel{ID: 1, Title: "old"}, false)
	rec := performSchemaInputRequest(t, handler, http.MethodPut, "/api/items/1", `{"TITLE":"new"}`)
	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.Contains(t, putRequestErrors(t, rec.Body.Bytes()), "title")
	assert.Equal(t, "old", mgr.items[1].(*putTitleModel).Title)

	rec = performSchemaInputRequest(t, handler, http.MethodPut, "/api/items/1", `{"title":"new"}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "new", mgr.items[1].(*putTitleModel).Title)
}

// A field tagged json:",omitempty" is serialized under its Go name, so a
// GET response echoed back by a strict PUT is accepted and written.
func TestUpdatePut_EmptyJSONNameUsesGoNameLikeResponse(t *testing.T) {
	mgr, handler := newPutRequiredHandler(t, seededPutRequiredModel(), true)
	rec := performSchemaInputRequest(t, handler, http.MethodGet, "/api/items/1", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	body := decodeContractBody(t, rec.Body.Bytes())
	require.Equal(t, "tag", body["Label"], rec.Body.String())

	rec = performSchemaInputRequest(t, handler, http.MethodPut, "/api/items/1", `{"id":1,"title":"new","slug":"fixed","Label":"changed"}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "changed", mgr.items[1].(*putRequiredModel).Label)
}

// A required write-only field, such as a password, is never serialized, so
// a client must resend it on every PUT.
func TestUpdatePut_WriteOnlyRequiredFieldMustBeResent(t *testing.T) {
	_, handler := newPutRequiredHandler(t, &hiddenRequiredPasswordModel{ID: 1, Name: "old", Password: "hash"}, true)
	rec := performSchemaInputRequest(t, handler, http.MethodPut, "/api/items/1", `{"name":"new"}`)
	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.Contains(t, putRequestErrors(t, rec.Body.Bytes()), "password")

	rec = performSchemaInputRequest(t, handler, http.MethodPut, "/api/items/1", `{"name":"new","password":"secret"}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

type hiddenRequiredPasswordModel struct {
	schema.BaseSchema
	ID       int64  `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Password string `json:"-" db:"password_hash"`
}

func (hiddenRequiredPasswordModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		{Name: "password", DBColumn: "password_hash", Type: schema.TypeString, Required: true, Editable: true, Serialize: false},
	}
}

type putTitleModel struct {
	schema.BaseSchema
	ID    int64  `json:"id" db:"id"`
	Title string `json:"title" db:"title"`
}

func (putTitleModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
	}
}

// putRequestErrors returns the field errors of a validation response.
func putRequestErrors(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	errs, ok := decodeContractBody(t, body)["errors"].(map[string]interface{})
	require.True(t, ok, string(body))
	return errs
}

type putSlugModel struct {
	schema.BaseSchema
	ID    int64  `json:"id" db:"id"`
	Title string `json:"title" db:"title"`
	Slug  string `json:"slug" db:"slug"`
}

func (putSlugModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
		schema.StringField("slug", schema.Required()),
	}
}
