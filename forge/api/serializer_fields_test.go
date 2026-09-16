package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fieldFilterModel struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	IsStaff     bool   `json:"is_staff"`
	IsSuperuser bool   `json:"is_superuser"`
}

type fieldFilterSerializer struct {
	*BaseSerializer
}

func newFieldFilterSerializer() Serializer {
	return &fieldFilterSerializer{BaseSerializer: NewBaseSerializer(make(map[string]interface{}))}
}

func (s *fieldFilterSerializer) New() Serializer { return newFieldFilterSerializer() }

func (s *fieldFilterSerializer) Fields() []string {
	return []string{"id", "username"}
}

func (s *fieldFilterSerializer) ReadOnlyFields() []string {
	return []string{"id", "is_staff", "is_superuser"}
}

func (s *fieldFilterSerializer) WriteOnlyFields() []string {
	return []string{"password"}
}

type fieldFilterQueryset struct {
	items []*fieldFilterModel
}

func (q *fieldFilterQueryset) All(context.Context) ([]*fieldFilterModel, error) {
	return q.items, nil
}

func (q *fieldFilterQueryset) Count(context.Context) (int64, error) {
	return int64(len(q.items)), nil
}

func (q *fieldFilterQueryset) Offset(int) *fieldFilterQueryset { return q }
func (q *fieldFilterQueryset) Limit(int) *fieldFilterQueryset  { return q }

func (q *fieldFilterQueryset) Create(_ context.Context, model interface{}) error {
	m, ok := model.(*fieldFilterModel)
	if !ok {
		return errors.New("unexpected model type")
	}
	m.ID = int64(len(q.items) + 1)
	q.items = append(q.items, m)
	return nil
}

func (q *fieldFilterQueryset) Get(_ context.Context, id int64) (*fieldFilterModel, error) {
	for _, item := range q.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, errors.New("not found")
}

func (q *fieldFilterQueryset) Update(_ context.Context, model interface{}) error { return nil }

func newFieldFilterViewSet(qs *fieldFilterQueryset) *BaseViewSet {
	return NewBaseViewSet(
		newFieldFilterSerializer,
		qs,
		&fieldFilterModel{},
	)
}

func serveFieldFilterRequest(vs ViewSet, method, path, body string) *httptest.ResponseRecorder {
	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestViewSetOutputExcludesFieldsOutsideWhitelist(t *testing.T) {
	qs := &fieldFilterQueryset{items: []*fieldFilterModel{
		{ID: 1, Username: "alice", Password: "hashed-secret", IsStaff: true, IsSuperuser: true},
	}}
	resp := serveFieldFilterRequest(newFieldFilterViewSet(qs), http.MethodGet, "/api/items/", "")
	require.Equal(t, http.StatusOK, resp.Code)
	body := resp.Body.String()
	assert.Contains(t, body, `"username":"alice"`)
	assert.NotContains(t, body, "hashed-secret")
	assert.NotContains(t, body, "password")
	assert.NotContains(t, body, "is_staff")
	assert.NotContains(t, body, "is_superuser")
}

func TestViewSetCreateIgnoresReadOnlyFields(t *testing.T) {
	qs := &fieldFilterQueryset{}
	resp := serveFieldFilterRequest(
		newFieldFilterViewSet(qs),
		http.MethodPost, "/api/items/",
		`{"username":"bob","password":"pw","is_staff":true,"is_superuser":true}`,
	)
	require.Equal(t, http.StatusCreated, resp.Code)
	require.Len(t, qs.items, 1)
	assert.Equal(t, "bob", qs.items[0].Username)
	assert.False(t, qs.items[0].IsStaff, "is_staff must not be settable through the API")
	assert.False(t, qs.items[0].IsSuperuser, "is_superuser must not be settable through the API")
	assert.NotContains(t, resp.Body.String(), "password")
}
