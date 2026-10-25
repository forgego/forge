package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serializerOutputModel struct {
	ID           int64  `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	InternalNote string `json:"internal_note" db:"internal_note"`
}

type serializerOutputSerializer struct {
	*BaseSerializer
}

func newSerializerOutputSerializer() Serializer {
	return &serializerOutputSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *serializerOutputSerializer) Fields() []string {
	return []string{"id", "name"}
}

func (s *serializerOutputSerializer) ReadOnlyFields() []string {
	return []string{"id"}
}

func (s *serializerOutputSerializer) New() Serializer {
	return newSerializerOutputSerializer()
}

type serializerOutputQueryset struct {
	unusedDeleteOperation
	items []*serializerOutputModel
}

func (q *serializerOutputQueryset) All(context.Context) ([]*serializerOutputModel, error) {
	result := make([]*serializerOutputModel, len(q.items))
	for i, item := range q.items {
		cp := *item
		result[i] = &cp
	}
	return result, nil
}

func (q *serializerOutputQueryset) Count(context.Context) (int64, error) {
	return int64(len(q.items)), nil
}

func (q *serializerOutputQueryset) Offset(int) *serializerOutputQueryset { return q }
func (q *serializerOutputQueryset) Limit(int) *serializerOutputQueryset  { return q }

func (q *serializerOutputQueryset) Get(_ context.Context, id int64) (*serializerOutputModel, error) {
	for _, item := range q.items {
		if item.ID == id {
			cp := *item
			return &cp, nil
		}
	}
	return nil, errors.New("not found")
}

func (q *serializerOutputQueryset) Create(_ context.Context, item *serializerOutputModel) error {
	item.ID = int64(len(q.items) + 1)
	item.InternalNote = "secret_internal_note"
	stored := *item
	q.items = append(q.items, &stored)
	return nil
}

func (q *serializerOutputQueryset) Update(_ context.Context, item *serializerOutputModel) error {
	for i, existing := range q.items {
		if existing.ID == item.ID {
			existing.Name = item.Name
			if item.InternalNote != "" {
				existing.InternalNote = item.InternalNote
			}
			q.items[i] = existing
			return nil
		}
	}
	return errors.New("not found")
}

func createSerializerOutputViewSet(qs *serializerOutputQueryset) *BaseViewSet {
	return NewBaseViewSet(
		newSerializerOutputSerializer,
		qs,
		&serializerOutputModel{},
	)
}

func TestViewSetSerializerOutput_FieldVisibility(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantName   string
	}{
		{
			name:       "create",
			method:     http.MethodPost,
			path:       "/api/v1/items/",
			body:       `{"name":"Created Item"}`,
			wantStatus: http.StatusCreated,
			wantName:   "Created Item",
		},
		{
			name:       "retrieve",
			method:     http.MethodGet,
			path:       "/api/v1/items/1",
			wantStatus: http.StatusOK,
			wantName:   "Initial Item",
		},
		{
			name:       "list",
			method:     http.MethodGet,
			path:       "/api/v1/items/",
			wantStatus: http.StatusOK,
			wantName:   "Initial Item",
		},
		{
			name:       "update",
			method:     http.MethodPut,
			path:       "/api/v1/items/1",
			body:       `{"name":"Updated Item"}`,
			wantStatus: http.StatusOK,
			wantName:   "Updated Item",
		},
		{
			name:       "custom_detail_action",
			method:     http.MethodGet,
			path:       "/api/v1/items/1/preview",
			wantStatus: http.StatusOK,
			wantName:   "Initial Item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs := &serializerOutputQueryset{
				items: []*serializerOutputModel{
					{
						ID:           1,
						Name:         "Initial Item",
						InternalNote: "classified_stored_note",
					},
				},
			}
			vs := createSerializerOutputViewSet(qs)

			apiRouter := NewRouter("/api/v1")
			apiRouter.Register("items", vs)
			apiRouter.Action("items", "preview", ActionConfig{
				Methods: []string{http.MethodGet},
				Detail:  true,
			}, func(w http.ResponseWriter, r *http.Request) {
				id, err := strconv.ParseInt(forgehttp.URLParam(r, "id"), 10, 64)
				if err != nil {
					_ = forgehttp.SendError(w, http.StatusBadRequest, "invalid id")
					return
				}
				item, err := qs.Get(r.Context(), id)
				if err != nil {
					_ = forgehttp.SendError(w, http.StatusNotFound, "not found")
					return
				}
				_ = forgehttp.SendJSON(w, http.StatusOK, vs.serializeResponse(item))
			})

			httpRouter := forgehttp.NewRouter()
			apiRouter.RegisterRoutes(httpRouter)

			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			rec := httptest.NewRecorder()
			httpRouter.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			bodyStr := rec.Body.String()
			assert.Contains(t, bodyStr, "name")
			assert.Contains(t, bodyStr, tt.wantName)
			assert.NotContains(t, bodyStr, "internal_note")
			assert.NotContains(t, bodyStr, "classified")
			assert.NotContains(t, bodyStr, "secret")

			var responseData map[string]interface{}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &responseData))
			if tt.name == "list" {
				results, ok := responseData["results"].([]interface{})
				require.True(t, ok)
				require.NotEmpty(t, results)
				for _, r := range results {
					itemMap, ok := r.(map[string]interface{})
					require.True(t, ok)
					assert.Contains(t, itemMap, "name")
					assert.NotContains(t, itemMap, "internal_note")
				}
			} else {
				assert.Contains(t, responseData, "name")
				assert.NotContains(t, responseData, "internal_note")
			}
		})
	}
}
