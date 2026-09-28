package core

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAction(t *testing.T) {
	tests := []struct {
		name   string
		action string
		label  string
	}{
		{
			name:   "basic action",
			action: "delete",
			label:  "Delete Selected",
		},
		{
			name:   "empty name",
			action: "",
			label:  "Empty Action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(ctx context.Context, instances []*struct{}) error { return nil }
			action := NewAction[struct{}](tt.action, tt.label, handler)
			if action.Name != tt.action {
				t.Errorf("Action.Name = %q, want %q", action.Name, tt.action)
			}
			if action.Label != tt.label {
				t.Errorf("Action.Label = %q, want %q", action.Label, tt.label)
			}
		})
	}
}

func TestAction_WithDescription(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithDescription("Test description")
	if result.Description != "Test description" {
		t.Errorf("Action.Description = %q, want %q", result.Description, "Test description")
	}
}

func TestAction_WithPermissions(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithPermissions("admin", "delete")
	if len(result.Permissions) != 2 {
		t.Errorf("Action.Permissions length = %d, want 2", len(result.Permissions))
	}
}

func TestAction_WithConfirmation(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithConfirmation("Are you sure?")
	if result.Confirmation != "Are you sure?" {
		t.Errorf("Action.Confirmation = %q, want %q", result.Confirmation, "Are you sure?")
	}
}

func TestAction_WithDangerous(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithDangerous(true)
	if !result.Dangerous {
		t.Error("Action.Dangerous should be true")
	}
}

func TestAction_WithIcon(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithIcon("trash")
	if result.Icon != "trash" {
		t.Errorf("Action.Icon = %q, want %q", result.Icon, "trash")
	}
}

func TestAction_WithUIComponent(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithUIComponent("ConfirmDialog")
	if result.UIComponent != "ConfirmDialog" {
		t.Errorf("Action.UIComponent = %q, want %q", result.UIComponent, "ConfirmDialog")
	}
}

func TestNewFieldset(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
	}{
		{
			name:   "empty fieldset",
			fields: []string{},
		},
		{
			name:   "single field",
			fields: []string{"name"},
		},
		{
			name:   "multiple fields",
			fields: []string{"name", "email", "created_at"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldset := NewFieldset[struct{}](tt.name, tt.fields...)
			if fieldset.Name != tt.name {
				t.Errorf("Fieldset.Name = %q, want %q", fieldset.Name, tt.name)
			}
			if len(fieldset.Fields) != len(tt.fields) {
				t.Errorf("Fieldset.Fields length = %d, want %d", len(fieldset.Fields), len(tt.fields))
			}
		})
	}
}

func TestFieldset_WithCollapsed(t *testing.T) {
	fieldset := Fieldset[struct{}]{Name: "test"}
	result := fieldset.WithCollapsed(true)
	if !result.Collapsed {
		t.Error("Fieldset.Collapsed should be true")
	}
}

func TestFieldset_WithDescription(t *testing.T) {
	fieldset := Fieldset[struct{}]{Name: "test"}
	result := fieldset.WithDescription("Test description")
	if result.Description != "Test description" {
		t.Errorf("Fieldset.Description = %q, want %q", result.Description, "Test description")
	}
}

func TestComputed(t *testing.T) {
	method := Computed("get_full_name")
	if method.Path() != "get_full_name" {
		t.Errorf("Method.Path() = %q, want %q", method.Path(), "get_full_name")
	}
}

func TestMethod_Path(t *testing.T) {
	method := Method("custom_method")
	if method.Path() != "custom_method" {
		t.Errorf("Method.Path() = %q, want %q", method.Path(), "custom_method")
	}
}

func TestRadioLayout_Constants(t *testing.T) {
	if RadioHorizontal != "horizontal" {
		t.Errorf("RadioHorizontal = %q, want %q", RadioHorizontal, "horizontal")
	}
	if RadioVertical != "vertical" {
		t.Errorf("RadioVertical = %q, want %q", RadioVertical, "vertical")
	}
}

func TestConfig_Defaults(t *testing.T) {
	config := Config[struct{}]{}
	// Test that a new config has zero values
	if config.ListPerPage != 0 {
		t.Errorf("Config.ListPerPage = %d, want 0", config.ListPerPage)
	}
	if config.PageType != "" {
		t.Errorf("Config.PageType = %q, want empty", config.PageType)
	}
}

func TestInlineRelationConfig_Fields(t *testing.T) {
	config := InlineRelationConfig{
		Type:         "one_to_many",
		Label:        "Items",
		Fields:       []string{"name", "quantity"},
		RelatedModel: "Item",
		RelatedField: "order_id",
	}
	if config.Type != "one_to_many" {
		t.Errorf("Type = %q, want %q", config.Type, "one_to_many")
	}
	if config.Label != "Items" {
		t.Errorf("Label = %q, want %q", config.Label, "Items")
	}
	if len(config.Fields) != 2 {
		t.Errorf("Fields length = %d, want 2", len(config.Fields))
	}
}

func TestInlineConfig_Fields(t *testing.T) {
	config := InlineConfig{
		ListDisplay: []string{"name", "email"},
	}
	if len(config.ListDisplay) != 2 {
		t.Errorf("ListDisplay length = %d, want 2", len(config.ListDisplay))
	}
}

func TestFilter_Type(t *testing.T) {
	filter := Filter[struct{}]{
		Name:    "status",
		Label:   "Status",
		Type:    "choice",
		Choices: []Choice{{Value: "active", Label: "Active"}},
	}
	if filter.Name != "status" {
		t.Errorf("Filter.Name = %q, want %q", filter.Name, "status")
	}
	if filter.Type != "choice" {
		t.Errorf("Filter.Type = %q, want %q", filter.Type, "choice")
	}
}

func TestChoice_Fields(t *testing.T) {
	choice := Choice{
		Value: "active",
		Label: "Active",
	}
	if choice.Value != "active" {
		t.Errorf("Choice.Value = %q, want %q", choice.Value, "active")
	}
	if choice.Label != "Active" {
		t.Errorf("Choice.Label = %q, want %q", choice.Label, "Active")
	}
}

type listObjectsTestModel struct {
	schema.BaseSchema
	ID    int64  `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
	Role  string `db:"role" json:"role"`
}

func (listObjectsTestModel) Meta() schema.Meta {
	return schema.Meta{
		TableName: "list_objects_test_models",
	}
}

func (listObjectsTestModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.StringField("email"),
		schema.StringField("role"),
	}
}

func setupListObjectsDB(t *testing.T) *Admin[listObjectsTestModel] {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "list_objects_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = database.Close()
	})

	_, err = database.Exec(`
		CREATE TABLE list_objects_test_models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			role TEXT
		);
	`)
	require.NoError(t, err)

	items := []struct {
		name, email, role string
	}{
		{"Alice Smith", "alice@example.com", "admin"},
		{"Bob Jones", "bob@domain.com", "editor"},
		{"Charlie Brown", "charlie@example.com", "viewer"},
		{"David Smith", "david@domain.com", "editor"},
		{"Eve Adams", "eve@example.com", "viewer"},
	}

	for _, it := range items {
		_, err := database.Exec(`INSERT INTO list_objects_test_models (name, email, role) VALUES (?, ?, ?)`,
			it.name, it.email, it.role)
		require.NoError(t, err)
	}

	manager, err := orm.NewManagerWithDB[listObjectsTestModel]("list_objects_test_models", database)
	require.NoError(t, err)

	cfg := &Config[listObjectsTestModel]{
		SearchFields: []Field{Method("name"), Method("email")},
		ListPerPage:  10,
	}

	admin, err := NewAdmin[listObjectsTestModel](listObjectsTestModel{}, manager, cfg)
	require.NoError(t, err)

	return admin
}

func TestListObjects_FiltersOrderingPagination(t *testing.T) {
	admin := setupListObjectsDB(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		params        ListParams
		expectedCount int64
		expectedNames []string
		expectedPage  int
		expectedPages int
		expectedSize  int
	}{
		{
			name: "search across search fields - name",
			params: ListParams{
				Search:   "Smith",
				Ordering: []string{"id"},
			},
			expectedCount: 2,
			expectedNames: []string{"Alice Smith", "David Smith"},
			expectedPage:  1,
			expectedPages: 1,
			expectedSize:  10,
		},
		{
			name: "search across search fields - email",
			params: ListParams{
				Search:   "domain.com",
				Ordering: []string{"id"},
			},
			expectedCount: 2,
			expectedNames: []string{"Bob Jones", "David Smith"},
			expectedPage:  1,
			expectedPages: 1,
			expectedSize:  10,
		},
		{
			name: "an __icontains lookup",
			params: ListParams{
				Filters: map[string]interface{}{
					"name__icontains": "brown",
				},
			},
			expectedCount: 1,
			expectedNames: []string{"Charlie Brown"},
			expectedPage:  1,
			expectedPages: 1,
			expectedSize:  10,
		},
		{
			name: "an __in lookup",
			params: ListParams{
				Filters: map[string]interface{}{
					"role__in": "admin, viewer",
				},
				Ordering: []string{"id"},
			},
			expectedCount: 3,
			expectedNames: []string{"Alice Smith", "Charlie Brown", "Eve Adams"},
			expectedPage:  1,
			expectedPages: 1,
			expectedSize:  10,
		},
		{
			name: "ordering asc with an unsafe field ignored",
			params: ListParams{
				Ordering: []string{"name", "unsafe;drop table", "nonexistent_field"},
			},
			expectedCount: 5,
			expectedNames: []string{"Alice Smith", "Bob Jones", "Charlie Brown", "David Smith", "Eve Adams"},
			expectedPage:  1,
			expectedPages: 1,
			expectedSize:  10,
		},
		{
			name: "ordering desc with an unsafe field ignored",
			params: ListParams{
				Ordering: []string{"-name", "1=1", "unknown"},
			},
			expectedCount: 5,
			expectedNames: []string{"Eve Adams", "David Smith", "Charlie Brown", "Bob Jones", "Alice Smith"},
			expectedPage:  1,
			expectedPages: 1,
			expectedSize:  10,
		},
		{
			name: "page 2 of size 2 with total count",
			params: ListParams{
				Page:     2,
				PageSize: 2,
				Ordering: []string{"id"},
			},
			expectedCount: 5,
			expectedNames: []string{"Charlie Brown", "David Smith"},
			expectedPage:  2,
			expectedPages: 3,
			expectedSize:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := admin.ListObjects(ctx, tt.params)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, resp.Count)
			assert.Equal(t, tt.expectedPage, resp.Page)
			assert.Equal(t, tt.expectedPages, resp.TotalPages)
			assert.Equal(t, tt.expectedSize, resp.PageSize)

			items, ok := resp.Results.([]*listObjectsTestModel)
			require.True(t, ok, "expected []*listObjectsTestModel, got %T", resp.Results)
			names := make([]string, len(items))
			for i, item := range items {
				names[i] = item.Name
			}
			assert.Equal(t, tt.expectedNames, names)
		})
	}
}
