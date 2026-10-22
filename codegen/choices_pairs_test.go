package generator

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLiteralChoicePairsAreParsed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.go")
	source := `package models
import "github.com/forgego/forge/schema"
type Order struct { schema.BaseSchema }
func (Order) Fields() []schema.Field {
	return []schema.Field{
		schema.StringField("status", schema.ChoicesFromPairsOpts("pending", "Pending", "paid", "Paid")),
	}
}`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	parser := NewASTParser()
	models, err := parser.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics := parser.Diagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if len(models) != 1 || len(models[0].Fields) != 1 {
		t.Fatalf("unexpected models: %#v", models)
	}
	want := []string{"pending", "paid"}
	if got := models[0].Fields[0].Options["choices"]; !reflect.DeepEqual(got, want) {
		t.Fatalf("choices = %#v, want %#v", got, want)
	}
}

func TestComputedChoicePairsAreRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.go")
	source := `package models
import "github.com/forgego/forge/schema"
type Order struct { schema.BaseSchema }
func (Order) Fields() []schema.Field {
	return []schema.Field{
		schema.StringField("status", schema.ChoicesFromPairsOpts("pending", "Pend"+"ing")),
	}
}`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	parser := NewASTParser()
	if _, err := parser.ParseFile(path); err != nil {
		t.Fatal(err)
	}
	diagnostics := parser.Diagnostics()
	if len(diagnostics) == 0 || !strings.Contains(diagnostics[0].Message, "ChoicesFromPairsOpts") {
		t.Fatalf("expected computed choice diagnostic, got %#v", diagnostics)
	}
}
