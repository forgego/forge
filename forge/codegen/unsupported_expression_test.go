package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnsupportedExpressionsProduceDiagnostics(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		line    string
		message string
	}{
		{"helper call", "return []schema.Field{\n\t\tdefaultFields(),\n\t}", "defaultFields()", "helper call defaultFields"},
		{"computed variable", "computed := defaultField()\n\treturn []schema.Field{\n\t\tcomputed,\n\t}", "computed,", "value of computed is computed at run time"},
		{"loop", "fields := []schema.Field{}\n\tfor range []int{1} {\n\t\tfields = append(fields, schema.StringField(\"name\"))\n\t}\n\treturn fields", "for range", "loops cannot be evaluated"},
		{"conditional", "fields := []schema.Field{}\n\tif enabled {\n\t\tfields = append(fields, schema.StringField(\"name\"))\n\t}\n\treturn fields", "if enabled", "conditional assembly cannot be evaluated"},
		{"computed append", "fields := []schema.Field{}\n\tfields = append(fields, more()...)\n\treturn fields", "more()...", "append of a computed slice"},
		{"direct helper return", "return defaultFields()", "return defaultFields()", "helper call defaultFields"},
		{"nested helper option", "return []schema.Field{\n\t\tschema.StringField(\"name\", requiredOption()),\n\t}", "requiredOption()", "helper call requiredOption"},
		{"chained helper default", "return []schema.Field{\n\t\tschema.String(\"name\").Default(loadDefault()).Build(),\n\t}", "loadDefault()", "helper call loadDefault"},
		{"unresolved identifier", "return []schema.Field{\n\t\tnameField,\n\t}", "nameField,", "field element nameField"},
		{"direct struct literal", "return []schema.Field{\n\t\tschema.Field{},\n\t}", "schema.Field{}", "field element"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := "package models\nimport \"github.com/forgego/forge/schema\"\n\ntype Product struct { schema.BaseSchema }\n\nfunc (Product) Fields() []schema.Field {\n\t" + test.body + "\n}\n\nvar enabled bool\nfunc defaultFields() schema.Field { return schema.StringField(\"name\") }\nfunc defaultField() schema.Field { return schema.StringField(\"name\") }\nfunc more() []schema.Field { return nil }\n"
			filename := filepath.Join(t.TempDir(), "models.go")
			if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			parser := NewASTParser()
			if _, err := parser.ParseFile(filename); err != nil {
				t.Fatal(err)
			}
			diagnostics := parser.Diagnostics()
			if len(diagnostics) != 1 {
				t.Fatalf("expected one diagnostic, got %#v", diagnostics)
			}
			if diagnostics[0].Line != sourceLine(source, test.line) {
				t.Errorf("line = %d, want %d", diagnostics[0].Line, sourceLine(source, test.line))
			}
			if !strings.Contains(diagnostics[0].Message, test.message) {
				t.Errorf("message = %q, want substring %q", diagnostics[0].Message, test.message)
			}
		})
	}
}

func TestSupportedExpressionsProduceNoDiagnostics(t *testing.T) {
	source := `package models
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema }
func (Product) Fields() []schema.Field {
	var fields = []schema.Field{schema.StringField("name")}
	fields = append(fields, schema.StringField("status"))
return fields
}`
	filename := filepath.Join(t.TempDir(), "models.go")
	if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	parser := NewASTParser()
	if _, err := parser.ParseFile(filename); err != nil {
		t.Fatal(err)
	}
	if diagnostics := parser.Diagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
}

func TestMetaComputedValueProducesDiagnostic(t *testing.T) {
	source := "package models\nimport \"github.com/forgego/forge/schema\"\n\ntype Product struct { schema.BaseSchema }\n\nfunc (Product) Meta() schema.Meta {\n\treturn schema.Meta{TableName: tableName()}\n}\n\nfunc tableName() string { return \"products\" }\n"
	filename := filepath.Join(t.TempDir(), "models.go")
	if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	parser := NewASTParser()
	if _, err := parser.ParseFile(filename); err != nil {
		t.Fatal(err)
	}
	diagnostics := parser.Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Line != sourceLine(source, "tableName()") {
		t.Errorf("line = %d, want %d", diagnostics[0].Line, sourceLine(source, "tableName()"))
	}
	if !strings.Contains(diagnostics[0].Message, "helper call tableName") {
		t.Errorf("message = %q, want substring %q", diagnostics[0].Message, "helper call tableName")
	}
}

func TestNewlySupportedExpressionsProduceNoDiagnostics(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"assigned constructor", "field := schema.StringField(\"name\")\n\treturn []schema.Field{field}"},
		{"unrelated control flow", "if invalid {\n\t\tpanic(\"bad\")\n\t}\n\treturn []schema.Field{schema.StringField(\"name\")}"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := "package models\nimport \"github.com/forgego/forge/schema\"\n\ntype Product struct { schema.BaseSchema }\n\nfunc (Product) Fields() []schema.Field {\n\t" + test.body + "\n}\n\nvar invalid bool\n"
			filename := filepath.Join(t.TempDir(), "models.go")
			if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			parser := NewASTParser()
			if _, err := parser.ParseFile(filename); err != nil {
				t.Fatal(err)
			}
			if diagnostics := parser.Diagnostics(); len(diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %#v", diagnostics)
			}
		})
	}
}

func TestDiagnosticStringAndSort(t *testing.T) {
	diagnostic := Diagnostic{File: "models.go", Line: 12, Column: 3, Model: "Product", Method: "Fields", Message: "unsupported"}
	if got, want := diagnostic.String(), "models.go:12:3: Product.Fields: unsupported"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	parser := NewASTParser()
	parser.diagnostics = []Diagnostic{
		{File: "b.go", Line: 1, Column: 1},
		{File: "a.go", Line: 2, Column: 1},
		{File: "a.go", Line: 1, Column: 2},
		{File: "a.go", Line: 1, Column: 1},
	}
	diagnostics := parser.Diagnostics()
	for i, want := range []Diagnostic{{File: "a.go", Line: 1, Column: 1}, {File: "a.go", Line: 1, Column: 2}, {File: "a.go", Line: 2, Column: 1}, {File: "b.go", Line: 1, Column: 1}} {
		if diagnostics[i].File != want.File || diagnostics[i].Line != want.Line || diagnostics[i].Column != want.Column {
			t.Errorf("diagnostic %d = %#v, want %#v", i, diagnostics[i], want)
		}
	}
}

func sourceLine(source, text string) int {
	for line, value := range strings.Split(source, "\n") {
		if strings.Contains(value, text) {
			return line + 1
		}
	}
	return 0
}
