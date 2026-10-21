package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMetaConstraintConstantsPreservePublicSchemaValues(t *testing.T) {
	for _, qualifier := range []string{"schema", "fields"} {
		t.Run(qualifier, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "models.go")
			source := `package models
import ` + qualifier + ` "github.com/forgego/forge/schema"
type Reservation struct { ` + qualifier + `.BaseSchema }
func (Reservation) Fields() []` + qualifier + `.Field { return []` + qualifier + `.Field{` + qualifier + `.Int64Field("id", ` + qualifier + `.Primary())} }
func (Reservation) Meta() ` + qualifier + `.Meta {
 return ` + qualifier + `.Meta{TableName: "reservations", Constraints: []` + qualifier + `.Constraint{
  {Name: "one_slot", Type: ` + qualifier + `.ConstraintTypeUnique, Fields: []string{"slot_id"}},
  {Name: "positive_duration", Type: ` + qualifier + `.ConstraintTypeCheck, Condition: "ends_at > starts_at"},
 }}
}`
			if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			parser := NewASTParser()
			definitions, err := parser.ParseFile(filename)
			if err != nil {
				t.Fatal(err)
			}
			if got := parser.Diagnostics(); len(got) != 0 {
				t.Fatalf("public constants diagnosed as runtime expressions: %v", got)
			}
			if len(definitions) != 1 || len(definitions[0].Meta.Constraints) != 2 {
				t.Fatalf("missing constraints: %#v", definitions)
			}
			constraints := definitions[0].Meta.Constraints
			if constraints[0].Type != "UNIQUE" || constraints[1].Type != "CHECK" {
				t.Fatalf("constraint values lost: %#v", constraints)
			}
		})
	}
}

func TestUnrelatedPackageConstantIsNotTreatedAsForgeSchema(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "models.go")
	source := `package models
import "github.com/forgego/forge/schema"
import other "example.com/unrelated"
type Reservation struct { schema.BaseSchema }
func (Reservation) Fields() []schema.Field { return nil }
func (Reservation) Meta() schema.Meta { return schema.Meta{Constraints: []schema.Constraint{{Name:"unknown", Type:other.ConstraintTypeUnique}}} }
`
	if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	parser := NewASTParser()
	if _, err := parser.ParseFile(filename); err != nil {
		t.Fatal(err)
	}
	if len(parser.Diagnostics()) != 1 {
		t.Fatalf("unrelated constant must remain unsupported: %v", parser.Diagnostics())
	}
}
