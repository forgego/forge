package schema

import (
	"context"
	"reflect"
	"testing"
)

type minimalContractModel struct{ BaseSchema }

var _ Schema = minimalContractModel{}

func TestEmbeddedSchemaImplementsModelContract(t *testing.T) {
	var model Schema = minimalContractModel{}
	if model.Fields() == nil || model.Relations() == nil {
		t.Fatal("default collections are nil")
	}
}

func TestBaseSchemaSuppliesEmptyDefaults(t *testing.T) {
	b := BaseSchema{}
	if len(b.Fields()) != 0 || len(b.Relations()) != 0 || !reflect.DeepEqual(b.Meta(), Meta{}) || b.Hooks() != nil {
		t.Fatalf("defaults: fields=%v relations=%v meta=%+v hooks=%v", b.Fields(), b.Relations(), b.Meta(), b.Hooks())
	}
}

func TestMetaCarriesTableOptions(t *testing.T) {
	m := Meta{TableName: "products", OrderBy: []string{"-created_at"}, TableComment: "catalog", TableOptions: map[string]string{"ENGINE": "InnoDB"}}
	if m.TableName != "products" || m.OrderBy[0] != "-created_at" || m.TableComment != "catalog" || m.TableOptions["ENGINE"] != "InnoDB" {
		t.Fatalf("meta=%+v", m)
	}
}

func TestDatabaseLayoutOptionsPopulateField(t *testing.T) {
	f := StringField("title", DBType("VARCHAR(80)"), DBIndex(), DBCollation("NOCASE"), DBComment("display title"))
	if f.DBType != "VARCHAR(80)" || !f.DBIndex || f.DBCollation != "NOCASE" || f.DBComment != "display title" {
		t.Fatalf("field=%+v", f)
	}
}

func TestGeneratedColumnCarriesExpressionAndStorage(t *testing.T) {
	f := Int64Field("double_value", GeneratedColumn("value * 2", true))
	if !f.Generated || f.GeneratedExpr != "value * 2" || !f.IsStored {
		t.Fatalf("field=%+v", f)
	}
}

func TestOneToOneDeclaresRelationCardinality(t *testing.T) {
	r := OneToOneField("profile_id", "Profile")
	if r.Type != RelationOneToOne || r.To != "Profile" || !r.DBConstraint {
		t.Fatalf("relation=%+v", r)
	}
}

func TestDBConstraintOptionDisablesForeignKeyDeclaration(t *testing.T) {
	r := ForeignKeyField("owner_id", "User", DBConstraint(false))
	if r.DBConstraint {
		t.Fatalf("constraint still enabled: %+v", r)
	}
}

func TestBeforeCreateHookRegistersCallback(t *testing.T) {
	called := false
	h := NewModelHooks().WithBeforeCreate(func(context.Context, interface{}) error { called = true; return nil })
	if err := h.BeforeCreate(context.Background(), nil); err != nil || !called {
		t.Fatalf("called=%v err=%v", called, err)
	}
}

func TestUpdateHooksRegisterBothCallbacks(t *testing.T) {
	var order []string
	h := NewModelHooks().WithBeforeUpdate(func(context.Context, interface{}) error { order = append(order, "before"); return nil }).WithAfterUpdate(func(context.Context, interface{}) error { order = append(order, "after"); return nil })
	_ = h.BeforeUpdate(context.Background(), nil)
	_ = h.AfterUpdate(context.Background(), nil)
	if !reflect.DeepEqual(order, []string{"before", "after"}) {
		t.Fatal(order)
	}
}

func TestSaveHooksRegisterBothCallbacks(t *testing.T) {
	var order []string
	h := NewModelHooks().WithBeforeSave(func(context.Context, interface{}) error { order = append(order, "before"); return nil }).WithAfterSave(func(context.Context, interface{}) error { order = append(order, "after"); return nil })
	_ = h.BeforeSave(context.Background(), nil)
	_ = h.AfterSave(context.Background(), nil)
	if !reflect.DeepEqual(order, []string{"before", "after"}) {
		t.Fatal(order)
	}
}

func TestDeleteHooksRegisterBothCallbacks(t *testing.T) {
	var order []string
	h := NewModelHooks().WithBeforeDelete(func(context.Context, interface{}) error { order = append(order, "before"); return nil }).WithAfterDelete(func(context.Context, interface{}) error { order = append(order, "after"); return nil })
	_ = h.BeforeDelete(context.Background(), nil)
	_ = h.AfterDelete(context.Background(), nil)
	if !reflect.DeepEqual(order, []string{"before", "after"}) {
		t.Fatal(order)
	}
}
