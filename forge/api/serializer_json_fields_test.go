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

	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jsonRoundTripModel has a schema.JSON field (generated as []byte) and a
// schema.Bytes field (also []byte): only the JSON field may be decoded into
// structured JSON on responses.
type jsonRoundTripModel struct {
	schema.BaseSchema
	ID       int64  `json:"id" db:"id"`
	Metadata []byte `json:"metadata" db:"metadata"`
	Blob     []byte `json:"blob" db:"blob"`
}

type jsonDBTagModel struct {
	schema.BaseSchema
	Payload []byte `json:"data" db:"metadata_json"`
}

type JSONEmbeddedGenerated struct {
	Payload []byte `json:"data" db:"metadata_json"`
}

type jsonEmbeddedModel struct {
	schema.BaseSchema
	JSONEmbeddedGenerated
}

func (jsonEmbeddedModel) Fields() []schema.Field {
	return []schema.Field{{Name: "metadata", DBColumn: "metadata_json", Type: schema.TypeJSON}}
}

func (jsonDBTagModel) Fields() []schema.Field {
	return []schema.Field{{Name: "metadata", DBColumn: "metadata_json", Type: schema.TypeJSON}}
}

func (jsonRoundTripModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id"),
		schema.JSONField("metadata"),
		schema.BytesField("blob"),
	}
}

type jsonRoundTripSerializer struct {
	*BaseSerializer
}

func newJSONRoundTripSerializer() Serializer {
	return &jsonRoundTripSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

type jsonRoundTripManager struct {
	mu     sync.Mutex
	items  map[int64]*jsonRoundTripModel
	nextID int64
}

func (m *jsonRoundTripManager) Create(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := model.(*jsonRoundTripModel)
	if !ok {
		return errors.New("unexpected model type")
	}
	m.nextID++
	item.ID = m.nextID
	cp := *item
	if m.items == nil {
		m.items = make(map[int64]*jsonRoundTripModel)
	}
	m.items[item.ID] = &cp
	return nil
}

func (m *jsonRoundTripManager) Get(ctx context.Context, id int64) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *item
	return &cp, nil
}

func (m *jsonRoundTripManager) Update(ctx context.Context, model interface{}) error {
	return nil
}

func (m *jsonRoundTripManager) Delete(ctx context.Context, model interface{}) error {
	return nil
}

func newJSONRoundTripHandler(mgr *jsonRoundTripManager) *forgehttp.Router {
	vs := NewBaseViewSet(newJSONRoundTripSerializer, mgr, &jsonRoundTripModel{})
	router := NewRouter("/api")
	router.Register("docs", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	return handler
}

func TestSerializeModel_JSONFieldReturnsObjectBytesFieldStaysBase64(t *testing.T) {
	mgr := &jsonRoundTripManager{}
	handler := newJSONRoundTripHandler(mgr)

	req := httptest.NewRequest(http.MethodPost, "/api/docs/",
		bytes.NewBufferString(`{"metadata":{"enabled":true},"blob":"aGVsbG8="}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	// SerializeModel preserves valid JSON bytes as a RawMessage; marshal the
	// result to verify that the response representation remains an object.
	serialized := SerializeModel(mgr.items[1])
	serializedJSON, err := json.Marshal(serialized)
	require.NoError(t, err)
	assert.Contains(t, string(serializedJSON), `"metadata":{"enabled":true}`,
		"JSON field must marshal as structured JSON, not base64")

	var created map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	assert.Equal(t, map[string]interface{}{"enabled": true}, created["metadata"],
		"JSON field must come back as structured JSON, not base64")
	assert.Equal(t, "aGVsbG8=", created["blob"],
		"Bytes field must keep base64 representation")

	// Retrieve the same row: the object must round-trip.
	req = httptest.NewRequest(http.MethodGet, "/api/docs/1", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var fetched map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &fetched))
	assert.Equal(t, map[string]interface{}{"enabled": true}, fetched["metadata"])
	assert.Equal(t, "aGVsbG8=", fetched["blob"])
}

func TestSerializeModel_InvalidStoredJSONFallsBackWithoutPanic(t *testing.T) {
	m := &jsonRoundTripModel{ID: 1, Metadata: []byte("not-json{{{"), Blob: []byte("hello")}
	var got map[string]interface{}
	require.NotPanics(t, func() {
		got = SerializeModel(m)
	})
	raw, ok := got["metadata"].([]byte)
	require.True(t, ok, "invalid JSON bytes must fall back to the current []byte representation")
	assert.Equal(t, []byte("not-json{{{"), raw)
	assert.Equal(t, []byte("hello"), got["blob"])
}

func TestSerializeModel_JSONFieldMatchesConcreteDBTag(t *testing.T) {
	serialized, err := json.Marshal(SerializeModel(&jsonDBTagModel{Payload: []byte(`{"enabled":true}`)}))
	require.NoError(t, err)
	assert.JSONEq(t, `{"data":{"enabled":true}}`, string(serialized))
}

func TestSerializeModel_JSONFieldResolvesAnonymousEmbedding(t *testing.T) {
	serialized, err := json.Marshal(SerializeModel(&jsonEmbeddedModel{
		JSONEmbeddedGenerated: JSONEmbeddedGenerated{Payload: []byte(`{"enabled":true}`)},
	}))
	require.NoError(t, err)
	assert.JSONEq(t, `{"data":{"enabled":true}}`, string(serialized))
}
