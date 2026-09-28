package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	forgeerrors "github.com/forgego/forge/errors"
	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type schemaInputModel struct {
	schema.BaseSchema
	ID           int64     `json:"id" db:"id"`
	LargeNumber  int64     `json:"large_number" db:"large_number"`
	Metadata     []byte    `json:"metadata" db:"metadata"`
	Payload      []byte    `json:"payload" db:"payload"`
	StartsAt     time.Time `json:"starts_at" db:"starts_at"`
	DateOnly     time.Time `json:"date_only" db:"date_only"`
	OccurredAt   time.Time `json:"occurred_at" db:"occurred_at"`
	Status       string    `json:"status" db:"status"`
	Title        string    `json:"title" db:"title"`
	HookObserved int64     `json:"-" db:"-"`
}

type typedJSONDetails struct {
	Enabled bool   `json:"enabled"`
	Label   string `json:"label"`
}

type typedJSONModel struct {
	schema.BaseSchema
	Tags     []string               `json:"tags" db:"tags"`
	Details  typedJSONDetails       `json:"details" db:"details"`
	Settings map[string]interface{} `json:"settings" db:"settings"`
}

func (typedJSONModel) Fields() []schema.Field {
	return []schema.Field{
		schema.JSONField("tags"),
		schema.JSONField("details"),
		schema.JSONField("settings"),
	}
}

type requiredSchemaModel struct {
	schema.BaseSchema
	ID   int64  `json:"id" db:"id"`
	Name string `json:"display_name" db:"display_name_col"`
}

func (requiredSchemaModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		{Name: "name", DBColumn: "display_name_col", Type: schema.TypeString, Required: true, Editable: true, Serialize: true},
	}
}

type requiredSchemaManager struct {
	createCalled bool
}

func TestPopulateFromMap_ResolvesSchemaAndDatabaseAliases(t *testing.T) {
	for _, alias := range []string{"name", "display_name_col", "display_name", "Name"} {
		t.Run(alias, func(t *testing.T) {
			model := &requiredSchemaModel{}
			require.NoError(t, populateFromMap(model, map[string]interface{}{alias: "resolved"}))
			assert.Equal(t, "resolved", model.Name)
		})
	}
}

func (m *requiredSchemaManager) Create(_ context.Context, model interface{}) error {
	m.createCalled = true
	model.(*requiredSchemaModel).ID = 1
	return nil
}

func (*requiredSchemaManager) Get(context.Context, int64) (interface{}, error) {
	return &requiredSchemaModel{ID: 1}, nil
}

func (*requiredSchemaManager) Update(context.Context, interface{}) error { return nil }
func (*requiredSchemaManager) Delete(context.Context, interface{}) error { return nil }

func (schemaInputModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("large_number"),
		schema.JSONField("metadata"),
		schema.BytesField("payload"),
		schema.TimeField("starts_at"),
		schema.DateField("date_only"),
		schema.DateTimeField("occurred_at"),
		schema.StringField("status", schema.Required(), schema.Default("draft")),
		schema.StringField("title"),
	}
}

func (m *schemaInputModel) Validate() error {
	if m.Status == "" {
		return errors.New("status is required")
	}
	return nil
}

func (m *schemaInputModel) BeforeCreate(context.Context) error {
	m.HookObserved = m.ID
	return nil
}

type schemaInputSerializer struct{ *BaseSerializer }

func newSchemaInputSerializer() Serializer {
	return &schemaInputSerializer{BaseSerializer: NewBaseSerializer(make(map[string]interface{}))}
}

type schemaInputManager struct {
	mu     sync.Mutex
	items  map[int64]*schemaInputModel
	nextID int64
	last   *schemaInputModel
}

func (m *schemaInputManager) Create(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item := model.(*schemaInputModel)
	if err := item.BeforeCreate(ctx); err != nil {
		return err
	}
	m.nextID++
	item.ID = m.nextID
	cp := *item
	m.last = &cp
	if m.items == nil {
		m.items = make(map[int64]*schemaInputModel)
	}
	m.items[item.ID] = &cp
	return nil
}

func (m *schemaInputManager) Get(_ context.Context, id int64) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok {
		return nil, forgeerrors.NewNotFoundErrorWithMessage("not found")
	}
	cp := *item
	return &cp, nil
}

func (m *schemaInputManager) Update(_ context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item := model.(*schemaInputModel)
	cp := *item
	m.items[item.ID] = &cp
	m.last = &cp
	return nil
}

func (*schemaInputManager) Delete(context.Context, interface{}) error { return nil }

func newSchemaInputHandler(mgr *schemaInputManager) *forgehttp.Router {
	vs := NewBaseViewSet(newSchemaInputSerializer, mgr, &schemaInputModel{})
	vs.ReadOnlyRequestFields = NonEditableFields(&schemaInputModel{})
	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	return handler
}

func performSchemaInputRequest(t *testing.T, handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestPopulateFromMap_UsesSchemaTypeForJSONAndBytes(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{name: "array", value: []interface{}{json.Number("1"), json.Number("2")}, want: `[1,2]`},
		{name: "string", value: "text", want: `"text"`},
		{name: "number", value: json.Number("42"), want: `42`},
		{name: "object", value: map[string]interface{}{"enabled": true}, want: `{"enabled":true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &schemaInputModel{}
			require.NoError(t, populateFromMap(model, map[string]interface{}{"metadata": tt.value}))
			assert.JSONEq(t, tt.want, string(model.Metadata))
		})
	}

	model := &schemaInputModel{}
	require.NoError(t, populateFromMap(model, map[string]interface{}{"payload": "aGVsbG8="}))
	assert.Equal(t, []byte("hello"), model.Payload)
}

func TestPopulateFromMap_JSONSchemaSupportsConcreteGoTypes(t *testing.T) {
	model := &typedJSONModel{}
	err := populateFromMap(model, map[string]interface{}{
		"tags":     []interface{}{"read", "write"},
		"details":  map[string]interface{}{"enabled": true, "label": "primary"},
		"settings": map[string]interface{}{"retries": json.Number("3")},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"read", "write"}, model.Tags)
	assert.Equal(t, typedJSONDetails{Enabled: true, Label: "primary"}, model.Details)
	assert.Equal(t, float64(3), model.Settings["retries"])
}

func TestBaseViewSet_Create_PreservesLargeInteger(t *testing.T) {
	mgr := &schemaInputManager{}
	rec := performSchemaInputRequest(t, newSchemaInputHandler(mgr), http.MethodPost, "/api/items/", `{"large_number":9007199254740993}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	require.NotNil(t, mgr.last)
	assert.Equal(t, int64(9007199254740993), mgr.last.LargeNumber)
}

func TestBaseViewSet_RejectsJSONNullBodies(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		mgr := &schemaInputManager{}
		rec := performSchemaInputRequest(t, newSchemaInputHandler(mgr), http.MethodPost, "/api/items/", `null`)
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Empty(t, mgr.items)
		assert.Nil(t, mgr.last)
	})

	for _, method := range []string{http.MethodPut, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			original := &schemaInputModel{ID: 1, Status: "published", Title: "unchanged"}
			mgr := &schemaInputManager{items: map[int64]*schemaInputModel{1: original}, nextID: 1}
			rec := performSchemaInputRequest(t, newSchemaInputHandler(mgr), method, "/api/items/1", `null`)
			require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.Equal(t, "unchanged", mgr.items[1].Title)
			assert.Nil(t, mgr.last)
		})
	}
}

func TestBaseViewSet_Create_ValidatesRequiredSchemaFieldWithoutTags(t *testing.T) {
	mgr := &requiredSchemaManager{}
	vs := NewBaseViewSet(newSchemaInputSerializer, mgr, &requiredSchemaModel{})
	router := NewRouter("/api")
	router.Register("required-items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	rec := performSchemaInputRequest(t, handler, http.MethodPost, "/api/required-items/", `{}`)
	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), `"display_name"`)
	assert.False(t, mgr.createCalled)
}

func TestBaseViewSet_Create_RejectsInvalidByteArrays(t *testing.T) {
	for _, body := range []string{
		`{"payload":[300]}`,
		`{"payload":[1,"x"]}`,
	} {
		t.Run(body, func(t *testing.T) {
			mgr := &schemaInputManager{}
			rec := performSchemaInputRequest(t, newSchemaInputHandler(mgr), http.MethodPost, "/api/items/", body)
			require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.Nil(t, mgr.last)
		})
	}
}

func TestBaseViewSet_Create_AcceptsIntegralByteArray(t *testing.T) {
	mgr := &schemaInputManager{}
	rec := performSchemaInputRequest(t, newSchemaInputHandler(mgr), http.MethodPost, "/api/items/", `{"payload":[0,1.0,255]}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	require.NotNil(t, mgr.last)
	assert.Equal(t, []byte{0, 1, 255}, mgr.last.Payload)
}

func TestPopulateFromMap_UsesSchemaTemporalFormats(t *testing.T) {
	for _, value := range []string{"14:30:00", "14:30"} {
		t.Run(value, func(t *testing.T) {
			model := &schemaInputModel{}
			require.NoError(t, populateFromMap(model, map[string]interface{}{"starts_at": value}))
			assert.Equal(t, 14, model.StartsAt.Hour())
			assert.Equal(t, 30, model.StartsAt.Minute())
		})
	}

	model := &schemaInputModel{}
	require.NoError(t, populateFromMap(model, map[string]interface{}{"occurred_at": "2026-09-17T14:30:00Z"}))
	assert.Equal(t, 2026, model.OccurredAt.Year())
}

func TestBaseViewSet_TemporalFieldsRoundTripThroughCreateAndUpdate(t *testing.T) {
	mgr := &schemaInputManager{}
	handler := newSchemaInputHandler(mgr)
	body := `{"starts_at":"14:30:05","date_only":"2026-09-19","occurred_at":"2026-09-19T14:30:05Z"}`

	created := performSchemaInputRequest(t, handler, http.MethodPost, "/api/items/", body)
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &response))
	assert.Equal(t, "14:30:05", response["starts_at"])
	assert.Equal(t, "2026-09-19", response["date_only"])

	updateBody, err := json.Marshal(response)
	require.NoError(t, err)
	updated := performSchemaInputRequest(t, handler, http.MethodPut, "/api/items/1", string(updateBody))
	require.Equal(t, http.StatusOK, updated.Code, updated.Body.String())
	require.NoError(t, json.Unmarshal(updated.Body.Bytes(), &response))
	assert.Equal(t, "14:30:05", response["starts_at"])
	assert.Equal(t, "2026-09-19", response["date_only"])
}

func TestBaseViewSet_Create_IgnoresAutoIncrementIDBeforeHooks(t *testing.T) {
	mgr := &schemaInputManager{}
	rec := performSchemaInputRequest(t, newSchemaInputHandler(mgr), http.MethodPost, "/api/items/", `{"id":999}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	require.NotNil(t, mgr.last)
	assert.Zero(t, mgr.last.HookObserved)
	assert.Equal(t, int64(1), mgr.last.ID)
}

func TestBaseViewSet_Create_AppliesSchemaDefaultsBeforeValidation(t *testing.T) {
	t.Run("omitted", func(t *testing.T) {
		mgr := &schemaInputManager{}
		rec := performSchemaInputRequest(t, newSchemaInputHandler(mgr), http.MethodPost, "/api/items/", `{}`)
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		require.NotNil(t, mgr.last)
		assert.Equal(t, "draft", mgr.last.Status)
	})

	t.Run("supplied", func(t *testing.T) {
		mgr := &schemaInputManager{}
		rec := performSchemaInputRequest(t, newSchemaInputHandler(mgr), http.MethodPost, "/api/items/", `{"status":"published"}`)
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		require.NotNil(t, mgr.last)
		assert.Equal(t, "published", mgr.last.Status)
	})

	t.Run("update unaffected", func(t *testing.T) {
		mgr := &schemaInputManager{items: map[int64]*schemaInputModel{1: {ID: 1, Status: "published"}}, nextID: 1}
		rec := performSchemaInputRequest(t, newSchemaInputHandler(mgr), http.MethodPatch, "/api/items/1", `{"title":"updated"}`)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.NotNil(t, mgr.last)
		assert.Equal(t, "published", mgr.last.Status)
	})
}
