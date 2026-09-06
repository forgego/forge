package orm

import (
	"reflect"
	"testing"
	"time"

	"github.com/forgego/forge/schema"
)

type SafeAccessorTestModel struct {
	schema.BaseSchema
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Age       int       `db:"age"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
}

func (SafeAccessorTestModel) Meta() schema.Meta {
	return schema.Meta{
		TableName: "safe_accessor_test_models",
	}
}

func (SafeAccessorTestModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int32Field("id", schema.Primary()),
		schema.StringField("name"),
		schema.Int32Field("age"),
		schema.BoolField("is_active"),
		schema.DateTimeField("created_at"),
	}
}

type SafeAccessorRelModel struct {
	schema.BaseSchema
	ID        int `db:"id"`
	ParentID  int `db:"parent_id"`
	Parent    *SafeAccessorTestModel `db:"-"`
	Title     string `db:"title"`
}

func (SafeAccessorRelModel) Meta() schema.Meta {
	return schema.Meta{
		TableName: "safe_accessor_rel_models",
	}
}

func (SafeAccessorRelModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int32Field("id", schema.Primary()),
		schema.Int32Field("parent_id"),
		schema.StringField("title"),
	}
}

func (SafeAccessorRelModel) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("parent", "SafeAccessorTestModel", schema.DBConstraint(false)),
	}
}

func init() {
	RegisterModelType("SafeAccessorTestModel", reflect.TypeOf(SafeAccessorTestModel{}))
	RegisterModelType("SafeAccessorRelModel", reflect.TypeOf(SafeAccessorRelModel{}))
}

func TestSafeAccessor_Get(t *testing.T) {
	accessor, err := NewSafeAccessor[SafeAccessorTestModel]()
	if err != nil {
		t.Fatalf("Failed to create SafeAccessor: %v", err)
	}

	instance := &SafeAccessorTestModel{
		ID:       1,
		Name:     "Test",
		Age:      25,
		IsActive: true,
	}

	tests := []struct {
		name      string
		fieldName string
		want      interface{}
		wantErr   bool
	}{
		{"Get existing string field", "name", "Test", false},
		{"Get existing int field", "age", 25, false},
		{"Get existing bool field", "is_active", true, false},
		{"Get non-existing field", "non_existing", nil, true},
		{"Get field by PascalCase", "Name", "Test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := accessor.Get(instance, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeAccessor.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("SafeAccessor.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeAccessor_Set(t *testing.T) {
	accessor, err := NewSafeAccessor[SafeAccessorTestModel]()
	if err != nil {
		t.Fatalf("Failed to create SafeAccessor: %v", err)
	}

	instance := &SafeAccessorTestModel{
		ID: 1,
	}

	tests := []struct {
		name      string
		fieldName string
		value     interface{}
		wantErr   bool
		check     func(*testing.T, *SafeAccessorTestModel)
	}{
		{
			name:      "Set existing string field",
			fieldName: "name",
			value:     "NewName",
			wantErr:   false,
			check: func(t *testing.T, m *SafeAccessorTestModel) {
				if m.Name != "NewName" {
					t.Errorf("Expected Name = 'NewName', got %v", m.Name)
				}
			},
		},
		{
			name:      "Set existing int field",
			fieldName: "age",
			value:     30,
			wantErr:   false,
			check: func(t *testing.T, m *SafeAccessorTestModel) {
				if m.Age != 30 {
					t.Errorf("Expected Age = 30, got %v", m.Age)
				}
			},
		},
		{
			name:      "Set convertible type",
			fieldName: "age",
			value:     int64(40), // Convertible to int
			wantErr:   false,
			check: func(t *testing.T, m *SafeAccessorTestModel) {
				if m.Age != 40 {
					t.Errorf("Expected Age = 40, got %v", m.Age)
				}
			},
		},
		{
			name:      "Set incompatible type",
			fieldName: "age",
			value:     "string_value",
			wantErr:   true,
			check:     func(t *testing.T, m *SafeAccessorTestModel) {},
		},
		{
			name:      "Set non-existing field",
			fieldName: "non_existing",
			value:     "value",
			wantErr:   true,
			check:     func(t *testing.T, m *SafeAccessorTestModel) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := accessor.Set(instance, tt.fieldName, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeAccessor.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.check(t, instance)
		})
	}
}

func TestSafeAccessor_GetPath(t *testing.T) {
	accessor, err := NewSafeAccessor[SafeAccessorRelModel]()
	if err != nil {
		t.Fatalf("Failed to create SafeAccessor: %v", err)
	}

	parent := &SafeAccessorTestModel{
		ID:   1,
		Name: "Parent Name",
	}

	instance := &SafeAccessorRelModel{
		ID:       2,
		ParentID: 1,
		Parent:   parent,
		Title:    "Child Title",
	}

	tests := []struct {
		name    string
		path    string
		want    interface{}
		wantErr bool
	}{
		{"Get local field", "title", "Child Title", false},
		{"Get related field", "parent__name", "Parent Name", false},
		{"Get non-existing related field", "parent__non_existing", nil, true},
		{"Get invalid path syntax", "parent___name", nil, true},
		{"Get field on nil relation", "non_existing_rel__name", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := accessor.GetPath(instance, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeAccessor.GetPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("SafeAccessor.GetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
