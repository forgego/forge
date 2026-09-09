package core

import (
	"context"
	"testing"

	"github.com/forgego/forge/schema"
)

// permTestModel is a minimal schema for permission tests.
type permTestModel struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (permTestModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100)),
	}
}

func (permTestModel) Relations() []schema.Relation { return nil }

func (permTestModel) Meta() schema.Meta {
	return schema.Meta{TableName: "perm_test_models", VerboseName: "PermTest"}
}

func (permTestModel) Hooks() *schema.ModelHooks { return nil }

func newPermTestAdmin(t *testing.T, cfg *Config[permTestModel]) *Admin[permTestModel] {
	t.Helper()
	a, err := NewAdmin[permTestModel](permTestModel{}, nil, cfg)
	if err != nil {
		t.Fatalf("NewAdmin failed: %v", err)
	}
	return a
}

func TestPermissions_DefaultDeny(t *testing.T) {
	a := newPermTestAdmin(t, &Config[permTestModel]{})
	ctx := context.Background()

	if a.HasAddPermission(ctx, nil) {
		t.Error("expected default deny for Add")
	}
	if a.HasChangePermission(ctx, nil, nil) {
		t.Error("expected default deny for Change")
	}
	if a.HasDeletePermission(ctx, nil, nil) {
		t.Error("expected default deny for Delete")
	}
	if a.HasViewPermission(ctx, nil, nil) {
		t.Error("expected default deny for View")
	}
	if a.HasModulePermission(ctx, nil) {
		t.Error("expected default deny for Module")
	}

	meta, err := a.GetMetadata(ctx, nil)
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}
	if meta.Permissions != (PermissionMetadata{}) {
		t.Errorf("expected zero permissions by default, got %+v", meta.Permissions)
	}
}

func TestPermissions_ExplicitGrant(t *testing.T) {
	cfg := &Config[permTestModel]{
		HasViewPermission: func(ctx context.Context, a *Admin[permTestModel], user interface{}, obj *permTestModel) bool {
			u, _ := user.(string)
			return u == "alice"
		},
	}
	a := newPermTestAdmin(t, cfg)

	alice := "alice"
	bob := "bob"

	if !a.HasViewPermission(context.Background(), alice, nil) {
		t.Error("expected grant for alice")
	}
	if a.HasViewPermission(context.Background(), bob, nil) {
		t.Error("expected deny for bob")
	}
}

func TestPermissions_MetadataIsolation(t *testing.T) {
	cfg := &Config[permTestModel]{
		HasViewPermission: func(ctx context.Context, a *Admin[permTestModel], user interface{}, obj *permTestModel) bool {
			u, _ := user.(string)
			return u == "alice"
		},
	}
	a := newPermTestAdmin(t, cfg)

	alice := "alice"
	bob := "bob"

	metaBob, err := a.GetMetadata(context.Background(), bob)
	if err != nil {
		t.Fatalf("GetMetadata(bob) failed: %v", err)
	}
	metaAlice, err := a.GetMetadata(context.Background(), alice)
	if err != nil {
		t.Fatalf("GetMetadata(alice) failed: %v", err)
	}

	if metaBob.Permissions.View {
		t.Error("bob metadata leaked View permission")
	}
	if !metaAlice.Permissions.View {
		t.Error("alice metadata missing View permission")
	}

	// The shared cache must not carry per-user permissions.
	metaBob2, err := a.GetMetadata(context.Background(), bob)
	if err != nil {
		t.Fatalf("GetMetadata(bob) again failed: %v", err)
	}
	if metaBob2.Permissions.View {
		t.Error("cached metadata leaked alice's permission into bob's response")
	}
}
