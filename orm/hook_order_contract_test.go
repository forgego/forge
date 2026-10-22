package orm

import (
	"context"
	"reflect"
	"testing"

	"github.com/forgego/forge/schema"
)

type hookOrderRecord struct {
	schema.BaseSchema
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	events *[]string
}

func (hookOrderRecord) Meta() schema.Meta { return schema.Meta{TableName: "hook_order_records"} }
func (hookOrderRecord) Fields() []schema.Field {
	return []schema.Field{schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()), schema.StringField("name")}
}
func (r *hookOrderRecord) BeforeCreate(context.Context) error {
	*r.events = append(*r.events, "before create")
	return nil
}
func (r *hookOrderRecord) AfterCreate(context.Context) error {
	*r.events = append(*r.events, "after create")
	return nil
}
func (r *hookOrderRecord) BeforeUpdate(context.Context) error {
	*r.events = append(*r.events, "before update")
	return nil
}
func (r *hookOrderRecord) AfterUpdate(context.Context) error {
	*r.events = append(*r.events, "after update")
	return nil
}
func (r *hookOrderRecord) BeforeSave(context.Context) error {
	*r.events = append(*r.events, "before save")
	return nil
}
func (r *hookOrderRecord) AfterSave(context.Context) error {
	*r.events = append(*r.events, "after save")
	return nil
}
func (r *hookOrderRecord) Clean() error { *r.events = append(*r.events, "clean"); return nil }

func TestCreationRunsHooksAroundInsertInManagerOrder(t *testing.T) {
	database := setupRelationTestDB(t)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`CREATE TABLE hook_order_records (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`); err != nil {
		t.Fatal(err)
	}
	m, err := NewManagerWithDB[hookOrderRecord]("hook_order_records", database)
	if err != nil {
		t.Fatal(err)
	}
	var events []string
	r := &hookOrderRecord{Name: "first", events: &events}
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

func TestUpdateRunsHooksAroundSQLInManagerOrder(t *testing.T) {
	database := setupRelationTestDB(t)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`CREATE TABLE hook_order_records (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`); err != nil {
		t.Fatal(err)
	}
	m, err := NewManagerWithDB[hookOrderRecord]("hook_order_records", database)
	if err != nil {
		t.Fatal(err)
	}
	var events []string
	r := &hookOrderRecord{Name: "first", events: &events}
	if err := m.Create(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	events = nil
	r.Name = "second"
	if err := m.Update(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	want := []string{"before update", "before save", "clean", "after update", "after save"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events=%v want=%v", events, want)
	}
	var stored string
	if err := database.QueryRow(`SELECT name FROM hook_order_records WHERE id = ?`, r.ID).Scan(&stored); err != nil || stored != "second" {
		t.Fatalf("stored=%q err=%v", stored, err)
	}
}
