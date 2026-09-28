package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// contractModel covers each request contract a field can have.
type contractModel struct {
	schema.BaseSchema
	ID        int64     `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Note      *string   `json:"note" db:"note"`
	Count     int64     `json:"count" db:"count"`
	Password  string    `json:"password" db:"password"`
	Internal  string    `json:"internal" db:"internal"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (contractModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
		schema.StringField("note"),
		schema.Int64Field("count"),
		// Write-only: accepted on input, never serialized.
		schema.StringField("password", schema.WriteOnly()),
		// Hidden: neither serialized nor writable.
		schema.StringField("internal", schema.Editable(false), schema.Serialize(false)),
		{Name: "created_at", Type: schema.TypeDateTime, Editable: true, Serialize: true, AutoNowAdd: true},
	}
}

// contractManager persists copies and counts the writes it receives. Writes
// stand in for business hooks: orm.Manager runs hooks inside Create/Update.
type contractManager struct {
	unusedDeleteOperation
	unusedListOperations
	mu      sync.Mutex
	items   map[int64]contractModel
	nextID  int64
	creates int
	updates int
}

func (m *contractManager) Create(_ context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item := model.(*contractModel)
	m.creates++
	m.nextID++
	item.ID = m.nextID
	m.items[item.ID] = *item
	return nil
}

func (m *contractManager) Get(_ context.Context, id int64) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &item, nil
}

func (m *contractManager) Update(_ context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item := model.(*contractModel)
	m.updates++
	m.items[item.ID] = *item
	return nil
}

func (m *contractManager) stored(id int64) contractModel {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.items[id]
}

type contractSerializer struct {
	*BaseSerializer
	fields []string
}

func (s *contractSerializer) Fields() []string { return s.fields }

// WriteOnlyFields declares an input the serializer consumes without a
// model field, so the strict contract accepts it.
func (s *contractSerializer) WriteOnlyFields() []string { return []string{"password_confirm"} }

type contractOptions struct {
	strict bool
	fields []string
}

// newContractHandler wires the viewset the way codegen/templates/api.tmpl
// does, with one seeded row.
func newContractHandler(opts contractOptions) (*contractManager, http.Handler) {
	note := "seeded"
	mgr := &contractManager{
		items:  map[int64]contractModel{1: {ID: 1, Title: "seeded", Note: &note, Count: 5, Password: "old", Internal: "keep"}},
		nextID: 1,
	}
	vs := NewBaseViewSet(func() Serializer {
		return &contractSerializer{BaseSerializer: NewBaseSerializer(nil), fields: opts.fields}
	}, mgr, &contractModel{})
	vs.ExcludeResponseFields = NonSerializableFields(&contractModel{})
	vs.ReadOnlyRequestFields = NonEditableFields(&contractModel{})
	vs.RejectUnknownRequestFields = opts.strict
	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	return mgr, handler
}

func decodeContractBody(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &decoded), string(body))
	return decoded
}

func TestRequestContract_StrictRejectsUnknownKeysBeforeAnyWrite(t *testing.T) {
	mgr, handler := newContractHandler(contractOptions{strict: true})
	before := mgr.stored(1)

	for _, tc := range []struct {
		method, target, body string
		unknown              []string
	}{
		{http.MethodPost, "/api/items/", `{"title":"new","titel":"typo","extra":1}`, []string{"titel", "extra"}},
		{http.MethodPut, "/api/items/1", `{"title":"changed","titel":"typo"}`, []string{"titel"}},
		{http.MethodPatch, "/api/items/1", `{"count":9,"Title ":"x"}`, []string{"Title "}},
		// A hidden, non-editable field is as unknown as a missing one, so
		// the response reveals nothing about it.
		{http.MethodPatch, "/api/items/1", `{"internal":"x"}`, []string{"internal"}},
		{http.MethodPatch, "/api/items/1", `{"Internal":"x","INTERNAL":"y"}`, []string{"Internal", "INTERNAL"}},
		// Only names the viewset reads are known; a case variant is not.
		{http.MethodPatch, "/api/items/1", `{"TITLE":"x"}`, []string{"TITLE"}},
	} {
		rec := performSchemaInputRequest(t, handler, tc.method, tc.target, tc.body)
		require.Equal(t, http.StatusBadRequest, rec.Code, "%s %s: %s", tc.method, tc.body, rec.Body.String())
		for _, key := range tc.unknown {
			assert.Contains(t, rec.Body.String(), key)
		}
		assert.Contains(t, rec.Body.String(), unknownRequestFieldMessage)
	}

	assert.Zero(t, mgr.creates, "create must not reach the manager")
	assert.Zero(t, mgr.updates, "update must not reach the manager")
	assert.Equal(t, before, mgr.stored(1), "stored data must not change")
}

func TestRequestContract_StrictAcceptsKnownEchoedAndDeclaredKeys(t *testing.T) {
	mgr, handler := newContractHandler(contractOptions{strict: true})

	// id and created_at are echoed back from a previous response; Count is
	// the Go name of a schema field; password_confirm is declared by the
	// serializer.
	rec := performSchemaInputRequest(t, handler, http.MethodPost, "/api/items/",
		`{"id":99,"created_at":"2020-01-01T00:00:00Z","title":"new","Count":3,"password":"pw","password_confirm":"pw"}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	created := mgr.stored(2)
	assert.Equal(t, int64(2), created.ID, "the manager, not the body, assigns the primary key")
	assert.True(t, created.CreatedAt.IsZero(), "created_at is database-owned")
	assert.Equal(t, "new", created.Title)
	assert.Equal(t, int64(3), created.Count)
	assert.Equal(t, "pw", created.Password, "write-only fields are writable")

	body := decodeContractBody(t, rec.Body.Bytes())
	assert.NotContains(t, body, "password", "write-only fields are never serialized")
	assert.NotContains(t, body, "internal")
	assert.Equal(t, "new", body["title"])

	// PUT and PATCH echo a full representation back without errors.
	rec = performSchemaInputRequest(t, handler, http.MethodPut, "/api/items/1",
		`{"id":1,"title":"echoed","note":"n","count":5,"created_at":"2020-01-01T00:00:00Z"}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "echoed", mgr.stored(1).Title)
}

func TestRequestContract_DefaultIgnoresUnknownKeys(t *testing.T) {
	mgr, handler := newContractHandler(contractOptions{})

	rec := performSchemaInputRequest(t, handler, http.MethodPost, "/api/items/", `{"title":"new","titel":"typo"}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, "new", mgr.stored(2).Title)

	rec = performSchemaInputRequest(t, handler, http.MethodPatch, "/api/items/1", `{"internal":"x","count":7}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "keep", mgr.stored(1).Internal, "a non-editable field is ignored, not written")
	assert.Equal(t, int64(7), mgr.stored(1).Count)
}

func TestRequestContract_NullOmittedAndZero(t *testing.T) {
	for _, strict := range []bool{false, true} {
		mgr, handler := newContractHandler(contractOptions{strict: strict})
		patch := func(body string) contractModel {
			t.Helper()
			rec := performSchemaInputRequest(t, handler, http.MethodPatch, "/api/items/1", body)
			require.Equal(t, http.StatusOK, rec.Code, "%s: %s", body, rec.Body.String())
			return mgr.stored(1)
		}

		// Omitted keys keep their stored values.
		got := patch(`{}`)
		require.NotNil(t, got.Note)
		assert.Equal(t, "seeded", *got.Note)
		assert.Equal(t, int64(5), got.Count)

		// null on a non-nullable scalar is ignored, not coerced to zero.
		got = patch(`{"count":null,"title":null}`)
		assert.Equal(t, int64(5), got.Count)
		assert.Equal(t, "seeded", got.Title)

		// An explicit zero value is written.
		got = patch(`{"count":0}`)
		assert.Equal(t, int64(0), got.Count)

		// null clears a nullable (pointer) field.
		got = patch(`{"note":null}`)
		assert.Nil(t, got.Note)
	}
}

func TestRequestContract_PutAndPatchApplyOnlyProvidedKeys(t *testing.T) {
	for _, method := range []string{http.MethodPut, http.MethodPatch} {
		mgr, handler := newContractHandler(contractOptions{strict: true})

		rec := performSchemaInputRequest(t, handler, method, "/api/items/1", `{"title":"renamed"}`)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		got := mgr.stored(1)
		assert.Equal(t, "renamed", got.Title, method)
		assert.Equal(t, int64(5), got.Count, "%s keeps omitted fields", method)
		require.NotNil(t, got.Note)
		assert.Equal(t, "old", got.Password, "%s keeps omitted write-only fields", method)
	}
}

func TestRequestContract_InvalidValuesFailBeforePersistence(t *testing.T) {
	for _, tc := range []struct {
		name, method, target, body string
	}{
		{"required field emptied by PUT", http.MethodPut, "/api/items/1", `{"title":""}`},
		{"required field emptied by PATCH", http.MethodPatch, "/api/items/1", `{"title":""}`},
		{"required field missing on create", http.MethodPost, "/api/items/", `{"count":1}`},
		{"unconvertible value", http.MethodPatch, "/api/items/1", `{"count":"many"}`},
		{"fractional integer", http.MethodPost, "/api/items/", `{"title":"x","count":1.5}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mgr, handler := newContractHandler(contractOptions{strict: true})
			before := mgr.stored(1)
			rec := performSchemaInputRequest(t, handler, tc.method, tc.target, tc.body)
			require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.Zero(t, mgr.creates)
			assert.Zero(t, mgr.updates)
			assert.Equal(t, before, mgr.stored(1))
		})
	}
}

func TestRequestContract_EmptyFieldsMeansAllSerializableFields(t *testing.T) {
	_, handler := newContractHandler(contractOptions{})
	rec := performSchemaInputRequest(t, handler, http.MethodGet, "/api/items/1", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	body := decodeContractBody(t, rec.Body.Bytes())
	for _, key := range []string{"id", "title", "note", "count", "created_at"} {
		assert.Contains(t, body, key)
	}
	assert.NotContains(t, body, "password")
	assert.NotContains(t, body, "internal")

	// A non-empty Fields list trims responses but does not restrict input.
	mgr, handler := newContractHandler(contractOptions{strict: true, fields: []string{"id", "title"}})
	rec = performSchemaInputRequest(t, handler, http.MethodPatch, "/api/items/1", `{"count":8}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	body = decodeContractBody(t, rec.Body.Bytes())
	assert.Equal(t, map[string]interface{}{"id": float64(1), "title": "seeded"}, body)
	assert.Equal(t, int64(8), mgr.stored(1).Count)
}

func TestUnknownRequestFields_HiddenWriteOnlyFieldIsKnown(t *testing.T) {
	model := &hiddenPasswordModel{}
	assert.Empty(t, unknownRequestFields(model, nil, map[string]interface{}{"password": "x", "password_hash": "x"}, nil))
	assert.Equal(t, []string{"is_superuser"}, unknownRequestFields(model, nil, map[string]interface{}{"is_superuser": true}, nil))
}

// hiddenPasswordModel hides its fields from JSON. A write-only schema field
// stays a known input; a serialized one hidden by json:"-" does not.
type hiddenPasswordModel struct {
	schema.BaseSchema
	ID          int64  `json:"id" db:"id"`
	Password    string `json:"-" db:"password_hash"`
	IsSuperuser bool   `json:"-" db:"is_superuser"`
}

func (hiddenPasswordModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		{Name: "password", DBColumn: "password_hash", Type: schema.TypeString, Editable: true, Serialize: false},
		schema.BoolField("is_superuser"),
	}
}
