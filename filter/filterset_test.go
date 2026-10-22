package filter

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
)

// MockModel for testing
type MockModel struct {
	schema.BaseSchema
	ID       int64
	Username string
	Email    string
	IsActive bool
}

func (MockModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary()),
		schema.StringField("username"),
		schema.StringField("email"),
		schema.BoolField("is_active"),
	}
}

func (MockModel) Meta() schema.Meta {
	return schema.Meta{TableName: "mock_models"}
}

func TestFilterSet_NewFilterSet(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatalf("Failed to create FilterSet: %v", err)
	}

	if fs == nil {
		t.Fatal("FilterSet is nil")
	}

	if fs.schema == nil {
		t.Error("FilterSet schema is nil")
	}
}

func TestDocumentedFilterSetRequestAPI(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatal(err)
	}
	base, err := orm.NewQuerySet[MockModel]("mock_models")
	if err != nil {
		t.Fatal(err)
	}
	security := DefaultSecurityConfig()
	security.AllowedFields["mock_models"] = []string{"username", "id"}
	security.AllowedLookups["username"] = []string{"icontains"}
	security.AllowedLookups["id"] = []string{"in"}
	fs.WithQueryset(base).WithSecurity(security)
	if _, err := fs.ApplyQueryParams(context.Background(), httptest.NewRequest("GET", "/?username__icontains=forge", nil)); err != nil {
		t.Fatalf("allowed lookup: %v", err)
	}
	if _, err := fs.ApplyQueryParams(context.Background(), httptest.NewRequest("GET", "/?id__in=1,2", nil)); err != nil {
		t.Fatalf("in lookup: %v", err)
	}
	if _, err := fs.ApplyQueryParams(context.Background(), httptest.NewRequest("GET", "/?email=hidden", nil)); err == nil {
		t.Fatal("unlisted field must be rejected")
	}
	if _, err := fs.ApplyQueryParams(context.Background(), httptest.NewRequest("GET", "/?username__gt=5", nil)); err == nil {
		t.Fatal("unlisted lookup must be rejected")
	}
}

func TestFilterSet_Where(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatalf("Failed to create FilterSet: %v", err)
	}

	builder := fs.Where("username")
	if builder == nil {
		t.Fatal("FilterBuilder is nil")
	}

	if builder.err != nil {
		t.Logf("Expected error for invalid path: %v", builder.err)
	}
}

func TestFilterNode_Validate(t *testing.T) {
	// Valid field node
	node := NewFieldNode("username", "contains", "john")
	if err := node.Validate(); err != nil {
		t.Errorf("Valid field node should not error: %v", err)
	}

	// Invalid field node (no lookup)
	invalidNode := &FilterNode{
		Op:    OpField,
		Field: "username",
	}
	if err := invalidNode.Validate(); err == nil {
		t.Error("Invalid field node should error")
	}

	// Valid AND node
	andNode := NewAndNode(
		NewFieldNode("username", "contains", "john"),
		NewFieldNode("is_active", "exact", true),
	)
	if err := andNode.Validate(); err != nil {
		t.Errorf("Valid AND node should not error: %v", err)
	}

	// Invalid AND node (only one child)
	invalidAndNode := NewAndNode(NewFieldNode("username", "exact", "john"))
	if err := invalidAndNode.Validate(); err == nil {
		t.Error("AND node with less than 2 children should error")
	}
}

func TestFilterNode_Clone(t *testing.T) {
	original := NewAndNode(
		NewFieldNode("username", "contains", "john"),
		NewFieldNode("email", "exact", "test@example.com"),
	)

	clone := original.Clone()

	if clone == nil {
		t.Fatal("Clone is nil")
	}

	if clone.Op != original.Op {
		t.Error("Clone Op does not match")
	}

	if len(clone.Children) != len(original.Children) {
		t.Error("Clone children count does not match")
	}
}

func TestFilterNode_ToJSON(t *testing.T) {
	node := NewFieldNode("username", "contains", "john")

	json, err := node.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize to JSON: %v", err)
	}

	if len(json) == 0 {
		t.Error("JSON is empty")
	}

	// Test deserialization
	deserialized, err := FromJSON(json)
	if err != nil {
		t.Fatalf("Failed to deserialize from JSON: %v", err)
	}

	if deserialized.Field != node.Field {
		t.Error("Deserialized field does not match")
	}
}
