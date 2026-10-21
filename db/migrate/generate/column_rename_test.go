package generate

import (
	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
	"testing"
)

func TestRenamedColumnEmitsDropAndAdd(t *testing.T) {
	previous := []generator.FieldDefinition{{Name: "old_name", Type: "String"}}
	current := []generator.FieldDefinition{{Name: "new_name", Type: "String"}}
	changes, err := (&Detector{}).detectColumnChanges("products", current, previous)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 {
		t.Fatalf("changes = %#v", changes)
	}
	var added, dropped bool
	for _, c := range changes {
		switch v := c.(type) {
		case *core.AddColumn:
			added = v.Column.Name == "new_name"
		case *core.DropColumn:
			dropped = v.ColumnName == "old_name"
		}
	}
	if !added || !dropped {
		t.Fatalf("rename must emit add and drop: %#v", changes)
	}
}
