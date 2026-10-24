package orm

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/forgego/forge/schema"
)

type hookOrderTestRecord struct {
	schema.BaseSchema
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	events *[]string
}

func (hookOrderTestRecord) Meta() schema.Meta {
	return schema.Meta{TableName: "hook_order_test_records"}
}
func (hookOrderTestRecord) Fields() []schema.Field {
	return []schema.Field{schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()), schema.StringField("name")}
}
func (r *hookOrderTestRecord) BeforeCreate(context.Context) error {
	*r.events = append(*r.events, "before create")
	return nil
}
func (r *hookOrderTestRecord) AfterCreate(context.Context) error {
	*r.events = append(*r.events, "after create")
	return nil
}
func (r *hookOrderTestRecord) BeforeUpdate(context.Context) error {
	*r.events = append(*r.events, "before update")
	return nil
}
func (r *hookOrderTestRecord) AfterUpdate(context.Context) error {
	*r.events = append(*r.events, "after update")
	return nil
}
func (r *hookOrderTestRecord) BeforeSave(context.Context) error {
	*r.events = append(*r.events, "before save")
	return nil
}
func (r *hookOrderTestRecord) AfterSave(context.Context) error {
	*r.events = append(*r.events, "after save")
	return nil
}
func (r *hookOrderTestRecord) Clean() error { *r.events = append(*r.events, "clean"); return nil }

type hookOrderRollbackRecord struct {
	schema.BaseSchema
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	events *[]string
}

func (hookOrderRollbackRecord) Meta() schema.Meta {
	return schema.Meta{TableName: "hook_order_rollback_records"}
}
func (hookOrderRollbackRecord) Fields() []schema.Field {
	return []schema.Field{schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()), schema.StringField("name")}
}
func (r *hookOrderRollbackRecord) BeforeCreate(context.Context) error {
	*r.events = append(*r.events, "before create")
	return nil
}
func (r *hookOrderRollbackRecord) AfterCreate(context.Context) error {
	*r.events = append(*r.events, "after create")
	return nil
}
func (r *hookOrderRollbackRecord) BeforeUpdate(context.Context) error {
	*r.events = append(*r.events, "before update")
	return nil
}
func (r *hookOrderRollbackRecord) AfterUpdate(context.Context) error {
	*r.events = append(*r.events, "after update")
	return nil
}
func (r *hookOrderRollbackRecord) BeforeSave(context.Context) error {
	*r.events = append(*r.events, "before save")
	return errors.New("hook error rollback")
}
func (r *hookOrderRollbackRecord) AfterSave(context.Context) error {
	*r.events = append(*r.events, "after save")
	return nil
}
func (r *hookOrderRollbackRecord) Clean() error { *r.events = append(*r.events, "clean"); return nil }

func TestHooksRunInTransactionOrder(t *testing.T) {
	database := setupRelationTestDB(t)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`CREATE TABLE hook_order_test_records (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`); err != nil {
		t.Fatal(err)
	}
	m, err := NewManagerWithDB[hookOrderTestRecord]("hook_order_test_records", database)
	if err != nil {
		t.Fatal(err)
	}
	var events []string
	r := &hookOrderTestRecord{Name: "first", events: &events}
	if err := m.Create(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	want := []string{"before create", "before save", "clean", "after create", "after save"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events=%v want=%v", events, want)
	}
	if r.ID == 0 {
		t.Fatal("insert did not assign an ID")
	}
}

func TestHookErrorRollsBackInsert(t *testing.T) {
	database := setupRelationTestDB(t)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`CREATE TABLE hook_order_rollback_records (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`); err != nil {
		t.Fatal(err)
	}
	m, err := NewManagerWithDB[hookOrderRollbackRecord]("hook_order_rollback_records", database)
	if err != nil {
		t.Fatal(err)
	}
	var events []string
	r := &hookOrderRollbackRecord{Name: "rollback-test", events: &events}
	if err := m.Create(context.Background(), r); err == nil {
		t.Fatal("expected error from hook, got nil")
	}
	var count int
	err = database.QueryRow(`SELECT COUNT(*) FROM hook_order_rollback_records`).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows inserted, got %d", count)
	}
}
