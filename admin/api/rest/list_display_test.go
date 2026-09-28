package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResolverAdmin struct {
	mockAdmin
	mu         sync.Mutex
	labelCalls int
	calledIDs  []interface{}
	labels     map[string]string
	labelErr   error
}

func (m *mockResolverAdmin) ObjectLabels(ctx context.Context, ids []interface{}) (map[string]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.labelCalls++
	m.calledIDs = append(m.calledIDs, ids...)
	if m.labelErr != nil {
		return nil, m.labelErr
	}
	result := make(map[string]string)
	for _, id := range ids {
		idStr := fmt.Sprint(id)
		if f, ok := id.(float64); ok {
			idStr = fmt.Sprintf("%.0f", f)
		}
		if label, ok := m.labels[idStr]; ok {
			result[idStr] = label
		}
	}
	return result, nil
}

func TestHandleList_FKDisplayLabels(t *testing.T) {
	registry := core.NewRegistry()

	parentAdmin := &mockResolverAdmin{
		mockAdmin: mockAdmin{
			modelName:     "authors",
			moduleAllowed: true,
			metadata: &core.Metadata{
				Name:        "authors",
				VerboseName: "Author",
			},
		},
		labels: map[string]string{
			"10": "Alice",
			"20": "Bob",
		},
	}
	require.NoError(t, registry.Register(parentAdmin))

	childAdmin := &mockAdmin{
		modelName:     "books",
		moduleAllowed: true,
		metadata: &core.Metadata{
			Name:        "books",
			VerboseName: "Book",
			Relations: []core.RelationMetadata{
				{
					Name:         "author",
					Type:         "ForeignKey",
					RelatedModel: "authors",
				},
			},
		},
		listResponse: &core.PaginatedResponse{
			Count:      4,
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
			Results: []map[string]interface{}{
				{"id": 1, "title": "Book 1", "author_id": 10},
				{"id": 2, "title": "Book 2", "author_id": 20},
				{"id": 3, "title": "Book 3", "author_id": 10},  // Duplicate parent id (10)
				{"id": 4, "title": "Book 4", "author_id": nil}, // NULL FK
			},
		},
	}
	require.NoError(t, registry.Register(childAdmin))

	router := NewRouter(registry)
	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	rec := httptest.NewRecorder()

	router.handleList(childAdmin)(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Count   int64                        `json:"count"`
		Results []map[string]interface{}     `json:"results"`
		Display map[string]map[string]string `json:"display"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.NotNil(t, resp.Display)
	require.Contains(t, resp.Display, "author")
	assert.Equal(t, "Alice", resp.Display["author"]["10"])
	assert.Equal(t, "Bob", resp.Display["author"]["20"])

	// Child with NULL FK must contribute nothing
	assert.Len(t, resp.Display["author"], 2)

	// Verify labels are resolved with exactly ONE query/call per relation
	assert.Equal(t, 1, parentAdmin.labelCalls)
	assert.Len(t, parentAdmin.calledIDs, 2, "Duplicate FKs should be deduplicated")

	// Results must remain untouched
	assert.Len(t, resp.Results, 4)
}

func TestHandleList_OneToOneDisplayLabels(t *testing.T) {
	registry := core.NewRegistry()

	profileAdmin := &mockResolverAdmin{
		mockAdmin: mockAdmin{
			modelName:     "profiles",
			moduleAllowed: true,
			metadata: &core.Metadata{
				Name:        "profiles",
				VerboseName: "Profile",
			},
		},
		labels: map[string]string{
			"101": "Profile for User 101",
		},
	}
	require.NoError(t, registry.Register(profileAdmin))

	userAdmin := &mockAdmin{
		modelName:     "users",
		moduleAllowed: true,
		metadata: &core.Metadata{
			Name:        "users",
			VerboseName: "User",
			Relations: []core.RelationMetadata{
				{
					Name:         "profile",
					Type:         "OneToOne",
					RelatedModel: "profiles",
				},
			},
		},
		listResponse: &core.PaginatedResponse{
			Count:      1,
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
			Results: []map[string]interface{}{
				{"id": 1, "name": "User 1", "profile": 101},
			},
		},
	}
	require.NoError(t, registry.Register(userAdmin))

	router := NewRouter(registry)
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()

	router.handleList(userAdmin)(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Display map[string]map[string]string `json:"display"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.NotNil(t, resp.Display)
	assert.Equal(t, "Profile for User 101", resp.Display["profile"]["101"])
	assert.Equal(t, 1, profileAdmin.labelCalls)
}

func TestHandleList_DisplayLabelsFallbackGracefully(t *testing.T) {
	t.Run("related model not implementing LabelResolver", func(t *testing.T) {
		registry := core.NewRegistry()

		nonResolverAdmin := &mockAdmin{
			modelName:     "categories",
			moduleAllowed: true,
			metadata: &core.Metadata{
				Name: "categories",
			},
		}
		require.NoError(t, registry.Register(nonResolverAdmin))

		childAdmin := &mockAdmin{
			modelName:     "items",
			moduleAllowed: true,
			metadata: &core.Metadata{
				Name: "items",
				Relations: []core.RelationMetadata{
					{
						Name:         "category",
						Type:         "ForeignKey",
						RelatedModel: "categories",
					},
				},
			},
			listResponse: &core.PaginatedResponse{
				Count: 1,
				Results: []map[string]interface{}{
					{"id": 1, "category_id": 5},
				},
			},
		}
		require.NoError(t, registry.Register(childAdmin))

		router := NewRouter(registry)
		req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
		rec := httptest.NewRecorder()

		router.handleList(childAdmin)(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Display map[string]map[string]string `json:"display"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Nil(t, resp.Display)
	})

	t.Run("resolver error does not fail list endpoint", func(t *testing.T) {
		registry := core.NewRegistry()

		errResolverAdmin := &mockResolverAdmin{
			mockAdmin: mockAdmin{
				modelName:     "tags",
				moduleAllowed: true,
				metadata: &core.Metadata{
					Name: "tags",
				},
			},
			labelErr: assert.AnError,
		}
		require.NoError(t, registry.Register(errResolverAdmin))

		childAdmin := &mockAdmin{
			modelName:     "articles",
			moduleAllowed: true,
			metadata: &core.Metadata{
				Name: "articles",
				Relations: []core.RelationMetadata{
					{
						Name:         "tag",
						Type:         "ForeignKey",
						RelatedModel: "tags",
					},
				},
			},
			listResponse: &core.PaginatedResponse{
				Count: 1,
				Results: []map[string]interface{}{
					{"id": 1, "tag_id": 9},
				},
			},
		}
		require.NoError(t, registry.Register(childAdmin))

		router := NewRouter(registry)
		req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
		rec := httptest.NewRecorder()

		router.handleList(childAdmin)(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Display map[string]map[string]string `json:"display"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Nil(t, resp.Display)
	})

	t.Run("metadata error does not fail list endpoint", func(t *testing.T) {
		registry := core.NewRegistry()

		childAdmin := &mockAdmin{
			modelName:     "tasks",
			moduleAllowed: true,
			metadataErr:   assert.AnError,
			listResponse: &core.PaginatedResponse{
				Count: 1,
				Results: []map[string]interface{}{
					{"id": 1, "title": "Task 1"},
				},
			},
		}
		require.NoError(t, registry.Register(childAdmin))

		router := NewRouter(registry)
		req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		rec := httptest.NewRecorder()

		router.handleList(childAdmin)(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})
}

type realParentModel struct {
	schema.BaseSchema
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func (realParentModel) Meta() schema.Meta {
	return schema.Meta{TableName: "real_parents"}
}

func (realParentModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
	}
}

type realChildModel struct {
	schema.BaseSchema
	ID       int64  `db:"id" json:"id"`
	Title    string `db:"title" json:"title"`
	ParentID *int64 `db:"parent_id" json:"parent_id"`
}

func (realChildModel) Meta() schema.Meta {
	return schema.Meta{TableName: "real_children"}
}

func (realChildModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
		schema.Int64Field("parent_id", schema.Optional()),
	}
}

func (realChildModel) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("parent", "real_parents", schema.OnDelete(schema.CascadeSET_NULL)),
	}
}

func TestHandleList_RealSQLiteIntegration(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	dbPath := filepath.Join(t.TempDir(), "real_sqlite_integration.sqlite")

	database, err := db.NewDB(dbPath)
	require.NoError(t, err)
	defer database.Close()

	_, err = database.Exec(`
		CREATE TABLE real_parents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);
		CREATE TABLE real_children (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			parent_id INTEGER,
			FOREIGN KEY(parent_id) REFERENCES real_parents(id)
		);
	`)
	require.NoError(t, err)

	res1, err := database.Exec(`INSERT INTO real_parents (name) VALUES (?)`, "Parent Alice")
	require.NoError(t, err)
	p1, err := res1.LastInsertId()
	require.NoError(t, err)

	res2, err := database.Exec(`INSERT INTO real_parents (name) VALUES (?)`, "Parent Bob")
	require.NoError(t, err)
	p2, err := res2.LastInsertId()
	require.NoError(t, err)

	_, err = database.Exec(`INSERT INTO real_children (title, parent_id) VALUES (?, ?)`, "Child 1", p1)
	require.NoError(t, err)
	_, err = database.Exec(`INSERT INTO real_children (title, parent_id) VALUES (?, ?)`, "Child 2", p2)
	require.NoError(t, err)
	_, err = database.Exec(`INSERT INTO real_children (title, parent_id) VALUES (?, ?)`, "Child 3", p1)
	require.NoError(t, err)
	_, err = database.Exec(`INSERT INTO real_children (title, parent_id) VALUES (?, NULL)`, "Child 4 without parent")
	require.NoError(t, err)

	parentMgr, err := orm.NewManagerWithDB[realParentModel]("real_parents", database)
	require.NoError(t, err)
	parentAdmin, err := core.NewAdmin[realParentModel](realParentModel{}, parentMgr, &core.Config[realParentModel]{
		HasViewPermission: func(ctx context.Context, admin *core.Admin[realParentModel], user interface{}, obj *realParentModel) bool {
			return true
		},
	})
	require.NoError(t, err)

	childMgr, err := orm.NewManagerWithDB[realChildModel]("real_children", database)
	require.NoError(t, err)
	childAdmin, err := core.NewAdmin[realChildModel](realChildModel{}, childMgr, &core.Config[realChildModel]{
		HasViewPermission: func(ctx context.Context, admin *core.Admin[realChildModel], user interface{}, obj *realChildModel) bool {
			return true
		},
	})
	require.NoError(t, err)

	registry := core.NewRegistry()
	require.NoError(t, registry.Register(parentAdmin))
	require.NoError(t, registry.Register(childAdmin))

	router := NewRouter(registry)
	req := httptest.NewRequest(http.MethodGet, "/api/real_children", nil)
	rec := httptest.NewRecorder()

	router.handleList(childAdmin)(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Count   int64                        `json:"count"`
		Display map[string]map[string]string `json:"display"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.Equal(t, int64(4), resp.Count)
	require.NotNil(t, resp.Display)
	require.Contains(t, resp.Display, "parent")
	assert.Equal(t, "Parent Alice", resp.Display["parent"][fmt.Sprint(p1)])
	assert.Equal(t, "Parent Bob", resp.Display["parent"][fmt.Sprint(p2)])
	assert.Len(t, resp.Display["parent"], 2, "Null parent contributes nothing")
}

func TestAttachDisplayLabels_Characterization(t *testing.T) {
	for _, tc := range []struct {
		name       string
		rows       []map[string]interface{}
		relations  []core.RelationMetadata
		labels     map[string]string
		want       map[string]map[string]string
		labelCalls int
	}{
		{"foreign key labels", []map[string]interface{}{{"author_id": 10}, {"author": map[string]interface{}{"id": 20}}}, []core.RelationMetadata{{Name: "author", Type: "ForeignKey", RelatedModel: "authors"}}, map[string]string{"10": "Alice", "20": "Bob"}, map[string]map[string]string{"author": {"10": "Alice", "20": "Bob"}}, 1},
		{"missing related rows", []map[string]interface{}{{"author_id": 10}}, []core.RelationMetadata{{Name: "author", Type: "ForeignKey", RelatedModel: "authors"}}, map[string]string{}, nil, 1},
		{"nil and absent values", []map[string]interface{}{{"author_id": nil}, {"title": "untagged"}}, []core.RelationMetadata{{Name: "author", Type: "ForeignKey", RelatedModel: "authors"}}, map[string]string{"10": "Alice"}, nil, 0},
		{"choice and unlabelled fields", []map[string]interface{}{{"status": "draft", "title": "Post"}}, []core.RelationMetadata{{Name: "status", Type: "Choices", RelatedModel: "authors"}}, map[string]string{"draft": "Draft"}, nil, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registry := core.NewRegistry()
			related := &mockResolverAdmin{mockAdmin: mockAdmin{modelName: "authors"}, labels: tc.labels}
			require.NoError(t, registry.Register(related))
			admin := &mockAdmin{metadata: &core.Metadata{Relations: tc.relations}}
			response := &core.PaginatedResponse{Results: tc.rows}

			NewRouter(registry).attachDisplayLabels(context.Background(), admin, nil, response)

			assert.Equal(t, tc.want, response.Display)
			assert.Equal(t, tc.labelCalls, related.labelCalls)
		})
	}
}

func TestAttachDisplayLabels_Characterization_EdgeCases(t *testing.T) {
	t.Run("various ID formats and deduplication", func(t *testing.T) {
		registry := core.NewRegistry()
		related := &mockResolverAdmin{
			mockAdmin: mockAdmin{modelName: "authors"},
			labels:    map[string]string{"10": "Alice", "20": "Bob", "30": "Carol", "40": "Dave"},
		}
		require.NoError(t, registry.Register(related))
		admin := &mockAdmin{metadata: &core.Metadata{
			Relations: []core.RelationMetadata{{Name: "author", Type: "ForeignKey", RelatedModel: "authors"}},
		}}

		rows := []map[string]interface{}{
			{"author": map[string]interface{}{"ID": 20}}, // uppercase ID in map
			{"Author_id": 30},         // case-insensitive match
			{"author": float64(40.0)}, // float64
			{"author_id": 10},         // regular
			{"author_id": 10},         // duplicate
			{"author": map[string]interface{}{"other": "foo"}}, // map without id/ID
		}
		response := &core.PaginatedResponse{Results: rows}

		NewRouter(registry).attachDisplayLabels(context.Background(), admin, nil, response)

		require.NotNil(t, response.Display)
		assert.Equal(t, map[string]string{
			"10": "Alice",
			"20": "Bob",
			"30": "Carol",
			"40": "Dave",
		}, response.Display["author"])
		assert.Equal(t, 1, related.labelCalls)
		assert.Len(t, related.calledIDs, 4)
	})

	t.Run("nil and error guards", func(t *testing.T) {
		registry := core.NewRegistry()
		admin := &mockAdmin{metadata: &core.Metadata{
			Relations: []core.RelationMetadata{{Name: "author", Type: "ForeignKey", RelatedModel: "authors"}},
		}}

		// Nil registry
		resp := &core.PaginatedResponse{Results: []map[string]interface{}{{"author_id": 10}}}
		NewRouter(nil).attachDisplayLabels(context.Background(), admin, nil, resp)
		assert.Nil(t, resp.Display)

		// Nil response
		NewRouter(registry).attachDisplayLabels(context.Background(), admin, nil, nil)

		// Nil Results
		respNilRes := &core.PaginatedResponse{Results: nil}
		NewRouter(registry).attachDisplayLabels(context.Background(), admin, nil, respNilRes)
		assert.Nil(t, respNilRes.Display)

		// Metadata error
		adminErr := &mockAdmin{metadataErr: assert.AnError}
		respMetaErr := &core.PaginatedResponse{Results: []map[string]interface{}{{"author_id": 10}}}
		NewRouter(registry).attachDisplayLabels(context.Background(), adminErr, nil, respMetaErr)
		assert.Nil(t, respMetaErr.Display)

		// Empty relations
		adminEmptyRel := &mockAdmin{metadata: &core.Metadata{Relations: nil}}
		respEmptyRel := &core.PaginatedResponse{Results: []map[string]interface{}{{"author_id": 10}}}
		NewRouter(registry).attachDisplayLabels(context.Background(), adminEmptyRel, nil, respEmptyRel)
		assert.Nil(t, respEmptyRel.Display)

		// Non-slice results (cannot unmarshal to []map[string]interface{})
		respBadRows := &core.PaginatedResponse{Results: "not-a-slice"}
		NewRouter(registry).attachDisplayLabels(context.Background(), admin, nil, respBadRows)
		assert.Nil(t, respBadRows.Display)
	})
}
