package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile_ExtractsHooksFromStructLiteral(t *testing.T) {
	src := `package models
import (
	"context"
	"github.com/forgego/forge/schema"
)

type User struct { schema.BaseSchema }

func (User) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeCreate: beforeCreate,
		AfterSave: helpers.afterSave,
		Clean: func(instance interface{}) error { return nil },
	}
}

func beforeCreate(ctx context.Context, instance interface{}) error { return nil }
var helpers hookHelpers
type hookHelpers struct {}
func (hookHelpers) afterSave(ctx context.Context, instance interface{}) error { return nil }
`

	def := parseSingleModelDefinition(t, src)

	if def.Hooks.BeforeCreate != "beforeCreate" {
		t.Fatalf("expected BeforeCreate hook to be beforeCreate, got %q", def.Hooks.BeforeCreate)
	}
	if def.Hooks.AfterSave != "helpers.afterSave" {
		t.Fatalf("expected AfterSave hook to be helpers.afterSave, got %q", def.Hooks.AfterSave)
	}
	if def.Hooks.Clean != "<inline>" {
		t.Fatalf("expected Clean hook to be <inline>, got %q", def.Hooks.Clean)
	}
}

func TestParseFile_ExtractsHooksFromBuilderChain(t *testing.T) {
	src := `package models
import (
	"context"
	"github.com/forgego/forge/schema"
)

type User struct { schema.BaseSchema }

func (User) Hooks() *schema.ModelHooks {
	return schema.NewModelHooks().
		WithBeforeCreate(beforeCreate).
		WithAfterSave(afterSave).
		WithClean(cleanHook)
}

func beforeCreate(ctx context.Context, instance interface{}) error { return nil }
func afterSave(ctx context.Context, instance interface{}) error { return nil }
func cleanHook(instance interface{}) error { return nil }
`

	def := parseSingleModelDefinition(t, src)

	if def.Hooks.BeforeCreate != "beforeCreate" {
		t.Fatalf("expected BeforeCreate hook to be beforeCreate, got %q", def.Hooks.BeforeCreate)
	}
	if def.Hooks.AfterSave != "afterSave" {
		t.Fatalf("expected AfterSave hook to be afterSave, got %q", def.Hooks.AfterSave)
	}
	if def.Hooks.Clean != "cleanHook" {
		t.Fatalf("expected Clean hook to be cleanHook, got %q", def.Hooks.Clean)
	}
}

func TestParseFile_ExtractsHooksFromAssignedVariable(t *testing.T) {
	src := `package models
import (
	"context"
	"github.com/forgego/forge/schema"
)

type User struct { schema.BaseSchema }

func (User) Hooks() *schema.ModelHooks {
	hooks := &schema.ModelHooks{
		BeforeSave: beforeSave,
	}
	return hooks
}

func beforeSave(ctx context.Context, instance interface{}) error { return nil }
`

	def := parseSingleModelDefinition(t, src)

	if def.Hooks.BeforeSave != "beforeSave" {
		t.Fatalf("expected BeforeSave hook to be beforeSave, got %q", def.Hooks.BeforeSave)
	}
}

func TestParseFile_HandlesNilHooks(t *testing.T) {
	src := `package models
import "github.com/forgego/forge/schema"

type User struct { schema.BaseSchema }

func (User) Hooks() *schema.ModelHooks {
	return nil
}
`

	def := parseSingleModelDefinition(t, src)

	if def.Hooks.BeforeCreate != "" || def.Hooks.AfterSave != "" || def.Hooks.Clean != "" {
		t.Fatalf("expected empty hooks definition, got %#v", def.Hooks)
	}
}

func TestParseFile_ExtractsFieldsFromAssignedVariableAndAppend(t *testing.T) {
	src := `package models
import "github.com/forgego/forge/schema"

type Product struct { schema.BaseSchema }

func (Product) Fields() []schema.Field {
	fields := []schema.Field{
		schema.Int64Field("id", schema.Primary()),
		schema.StringField("name", schema.Required(), schema.Unique()),
	}
	fields = append(fields, schema.StringField("status", schema.Choices("active", "draft", "archived")))
	return fields
}
`
	def := parseSingleModelDefinition(t, src)
	if len(def.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(def.Fields))
	}
	if def.Fields[0].Name != "id" {
		t.Errorf("expected field id, got %s", def.Fields[0].Name)
	}
	if def.Fields[1].Name != "name" {
		t.Errorf("expected field name, got %s", def.Fields[1].Name)
	}
	if def.Fields[1].ValidationTag != "required,unique" {
		t.Errorf("expected validation tag required,unique, got %q", def.Fields[1].ValidationTag)
	}
	if def.Fields[2].Name != "status" {
		t.Errorf("expected field status, got %s", def.Fields[2].Name)
	}
	if def.Fields[2].ValidationTag != "oneof=active draft archived" {
		t.Errorf("expected validation tag oneof=active draft archived, got %q", def.Fields[2].ValidationTag)
	}
}

func TestParseFile_ExtractsRelationsFromAssignedVariable(t *testing.T) {
	src := `package models
import "github.com/forgego/forge/schema"

type Order struct { schema.BaseSchema }

func (Order) Relations() []schema.Relation {
	rels := []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer", schema.OnDelete(schema.CascadeCASCADE)),
	}
	return rels
}
`
	def := parseSingleModelDefinition(t, src)
	if len(def.Relations) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(def.Relations))
	}
	if def.Relations[0].Name != "customer_id" {
		t.Errorf("expected relation name customer_id, got %s", def.Relations[0].Name)
	}
	if def.Relations[0].To != "Customer" {
		t.Errorf("expected relation to Customer, got %s", def.Relations[0].To)
	}
}

func TestParseFile_ExtractsMetaFromAssignedVariable(t *testing.T) {
	src := `package models
import "github.com/forgego/forge/schema"

type Customer struct { schema.BaseSchema }

func (Customer) Meta() schema.Meta {
	m := schema.Meta{
		TableName: "app_customers",
		VerboseName: "Customer Account",
		VerboseNamePlural: "Customer Accounts",
	}
	return m
}
`
	def := parseSingleModelDefinition(t, src)
	if def.Meta.TableName != "app_customers" {
		t.Errorf("expected table name app_customers, got %s", def.Meta.TableName)
	}
	if def.Meta.VerboseName != "Customer Account" {
		t.Errorf("expected verbose name Customer Account, got %s", def.Meta.VerboseName)
	}
	if def.Meta.VerboseNamePlural != "Customer Accounts" {
		t.Errorf("expected verbose name plural Customer Accounts, got %s", def.Meta.VerboseNamePlural)
	}
}

func TestParseFile_DecimalLargeFloatValidationTag(t *testing.T) {
	src := `package models
import "github.com/forgego/forge/schema"

type Account struct { schema.BaseSchema }

func (Account) Fields() []schema.Field {
	return []schema.Field{
		schema.FloatField("balance", schema.MaxValue(10000000.5)),
	}
}
`
	def := parseSingleModelDefinition(t, src)
	if len(def.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(def.Fields))
	}
	if def.Fields[0].ValidationTag != "lte=10000000.5" {
		t.Errorf("expected validation tag lte=10000000.5 (no scientific notation), got %q", def.Fields[0].ValidationTag)
	}
}

func parseSingleModelDefinition(t *testing.T, src string) *ModelDefinition {
	t.Helper()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "models.go")
	if err := os.WriteFile(filename, []byte(src), 0o600); err != nil {
		t.Fatalf("failed to write temp model file: %v", err)
	}

	parser := NewASTParser()
	defs, err := parser.ParseFile(filename)
	if err != nil {
		t.Fatalf("failed to parse model file: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected one model definition, got %d", len(defs))
	}

	return defs[0]
}
