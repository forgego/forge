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
