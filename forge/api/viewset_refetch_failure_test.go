package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type refetchFailureModel struct {
	schema.BaseSchema
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func (refetchFailureModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

type refetchFailureManager struct {
	item          *refetchFailureModel
	createCalls   int
	updateCalls   int
	getCalls      int
	failEveryGet  bool
	failSecondGet bool
}

func (m *refetchFailureManager) Create(_ context.Context, model interface{}) error {
	m.createCalls++
	item := model.(*refetchFailureModel)
	item.ID = 1
	copy := *item
	m.item = &copy
	return nil
}

func (m *refetchFailureManager) Get(_ context.Context, id int64) (interface{}, error) {
	m.getCalls++
	if m.failEveryGet || (m.failSecondGet && m.getCalls == 2) {
		return nil, errors.New("refetch unavailable")
	}
	copy := *m.item
	return &copy, nil
}

func (m *refetchFailureManager) Update(_ context.Context, model interface{}) error {
	m.updateCalls++
	copy := *model.(*refetchFailureModel)
	m.item = &copy
	return nil
}

func (*refetchFailureManager) Delete(context.Context, interface{}) error { return nil }

func refetchFailureHandler(manager *refetchFailureManager) *forgehttp.Router {
	viewSet := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, manager, &refetchFailureModel{})
	router := NewRouter("/api")
	router.Register("items", viewSet)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	return handler
}

func TestBaseViewSet_CreateRefetchFailureStillReturnsCreatedOnce(t *testing.T) {
	manager := &refetchFailureManager{failEveryGet: true}
	handler := refetchFailureHandler(manager)
	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"name":"created"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, 1, manager.createCalls)
	require.NotNil(t, manager.item)
	assert.Equal(t, "created", manager.item.Name)
}

func TestBaseViewSet_UpdateRefetchFailureStillReturnsSuccess(t *testing.T) {
	for _, method := range []string{http.MethodPut, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			manager := &refetchFailureManager{
				item:          &refetchFailureModel{ID: 1, Name: "before"},
				failSecondGet: true,
			}
			handler := refetchFailureHandler(manager)
			req := httptest.NewRequest(method, "/api/items/1", bytes.NewBufferString(`{"name":"after"}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, 1, manager.updateCalls)
			assert.Equal(t, "after", manager.item.Name)
		})
	}
}
