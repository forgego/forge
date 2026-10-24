package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	admincore "github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/api/serializers"
	codegen "github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate"
	migratecore "github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/state"
	"github.com/forgego/forge/tests/testdata/models/conformance"
)

func TestSharedFactsConformance(t *testing.T) {
	fixtureDir := filepath.Join("..", "testdata", "models", "conformance")

	// Parsed schema: the source declaration is the canonical fact set.
	definitions, err := codegen.NewASTParser().ParseDirectory(fixtureDir)
	if err != nil {
		t.Fatal(err)
	}
	item := parsedModel(t, definitions, "CatalogItem")
	assertParsedField(t, item, "optional_note", false, 40, nil, false)
	assertParsedField(t, item, "name", true, 60, nil, false)
	assertParsedField(t, item, "state", true, 20, "draft", false)
	assertParsedField(t, item, "external_code", true, 32, nil, true)
	assertParsedRelation(t, item, "owner_id", "ForeignKeyField", "CASCADE", "")
	assertParsedRelation(t, item, "tags", "ManyToManyField", "", "catalog_item_tags")

	// Generated Go: validation tags and Go types must retain the same constraints.
	generatedDir := t.TempDir()
	if err := codegen.NewGenerator(fixtureDir, generatedDir).Generate(); err != nil {
		t.Fatal(err)
	}
	generated, err := os.ReadFile(filepath.Join(generatedDir, "gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	generatedText := string(generated)
	for _, want := range []string{
		"OptionalNote string `json:\"optional_note\" db:\"optional_note\" validate:\"max=40\"`",
		"Name string `json:\"name\" db:\"name\" validate:\"required,max=60\"`",
		"State string `json:\"state\" db:\"state\" validate:\"required,max=20\"`",
		"ExternalCode string `json:\"external_code\" db:\"external_code\" validate:\"required,unique,max=32\"`",
		"schema.StringField(\"optional_note\", schema.Optional(), schema.MaxLength(40))",
		"schema.StringField(\"state\", schema.Required(), schema.MaxLength(20), schema.Default(\"draft\"))",
		"schema.StringField(\"external_code\", schema.Required(), schema.MaxLength(32), schema.Unique())",
		"schema.ForeignKeyField(\"owner_id\", \"Owner\", schema.OnDelete(schema.CascadeType(\"CASCADE\")))",
		"schema.ManyToManyField(\"tags\", \"Tag\", schema.Through(\"catalog_item_tags\"))",
	} {
		if !strings.Contains(generatedText, want) {
			t.Errorf("generated code does not contain %q", want)
		}
	}

	// Migration state: inspect the generated SQL state, not just migration text.
	migrationsDir := t.TempDir()
	if err := migrate.GenerateForDriver("shared facts", fixtureDir, migrationsDir, migratecore.DriverSQLite); err != nil {
		t.Fatal(err)
	}
	migrationState, err := state.LoadStateFromFiles(migrationsDir)
	if err != nil {
		t.Fatal(err)
	}
	table := migrationState.Tables["catalog_items"]
	if table == nil {
		t.Fatal("migration state has no catalog_items table")
	}
	assertMigrationColumn(t, table.Columns["optional_note"], false, 40, nil, false)
	assertMigrationColumn(t, table.Columns["name"], true, 60, nil, false)
	assertMigrationColumn(t, table.Columns["state"], true, 20, "draft", false)
	assertMigrationColumn(t, table.Columns["external_code"], true, 32, nil, true)
	if !hasForeignKey(table.ForeignKeys, "owner_id", "owners", "CASCADE") {
		t.Errorf("migration state foreign keys = %+v, want owner_id -> owners ON DELETE CASCADE", table.ForeignKeys)
	}
	if migrationState.Tables["catalog_item_tags"] == nil {
		t.Error("M2M through model did not produce catalog_item_tags migration state")
	}

	// Admin metadata mirrors schema nullability, length, default, uniqueness and relations.
	admin, err := admincore.NewAdmin[conformance.CatalogItem](conformance.CatalogItem{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	adminMeta, err := admin.GetMetadata(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	assertAdminField(t, adminMeta.Fields, "optional_note", false, 40, nil, false)
	assertAdminField(t, adminMeta.Fields, "name", true, 60, nil, false)
	assertAdminField(t, adminMeta.Fields, "state", true, 20, "draft", false)
	assertAdminField(t, adminMeta.Fields, "external_code", true, 32, nil, true)
	assertAdminRelation(t, adminMeta.Relations, "owner_id", "ForeignKey", "Owner", "CASCADE", "")
	assertAdminRelation(t, adminMeta.Relations, "tags", "ManyToMany", "Tag", "", "catalog_item_tags")

	// REST serializer metadata and validation both derive from FromSchema.
	serializer := serializers.FromSchema[conformance.CatalogItem](conformance.CatalogItem{})
	restMeta := serializer.SchemaMetadata()
	if _, err := json.Marshal(restMeta); err != nil {
		t.Fatalf("REST schema metadata is not JSON serializable: %v", err)
	}
	assertRESTField(t, restMeta.Fields, "optional_note", false, 40, nil, false)
	assertRESTField(t, restMeta.Fields, "name", true, 60, nil, false)
	assertRESTField(t, restMeta.Fields, "state", true, 20, "draft", false)
	assertRESTField(t, restMeta.Fields, "external_code", true, 32, nil, true)
	assertRESTRelation(t, restMeta.Relations, "owner_id", "ForeignKey", "Owner", "CASCADE", "")
	assertRESTRelation(t, restMeta.Relations, "tags", "ManyToMany", "Tag", "", "catalog_item_tags")
	valid := conformance.CatalogItem{}
	if err := serializer.Deserialize(map[string]interface{}{
		"optional_note": nil,
		"name":          "item",
		"external_code": "sku-1",
		"owner_id":      int64(1),
	}, &valid); err != nil {
		t.Fatalf("REST serializer rejected optional nullable field/default: %v", err)
	}
	if valid.State != "draft" {
		t.Errorf("REST serializer default state = %q, want draft", valid.State)
	}
	tooLong := conformance.CatalogItem{}
	if err := serializer.Deserialize(map[string]interface{}{"name": strings.Repeat("x", 61), "external_code": "sku-2", "owner_id": int64(1)}, &tooLong); err == nil {
		t.Error("REST serializer accepted name beyond max length")
	}
	missingRequired := conformance.CatalogItem{}
	if err := serializer.Deserialize(map[string]interface{}{"external_code": "sku-3", "owner_id": int64(1)}, &missingRequired); err == nil {
		t.Error("REST serializer accepted missing required name")
	}
}

func assertRESTField(t *testing.T, fields []serializers.SchemaFieldMetadata, name string, required bool, maxLength int, defaultValue interface{}, unique bool) {
	t.Helper()
	for _, field := range fields {
		if field.Name == name {
			if field.Required != required || field.MaxLength != maxLength || field.Default != defaultValue || field.Unique != unique {
				t.Errorf("REST field %s = required:%v max_length:%d default:%v unique:%v", name, field.Required, field.MaxLength, field.Default, field.Unique)
			}
			return
		}
	}
	t.Errorf("REST field %q not found", name)
}

func assertRESTRelation(t *testing.T, relations []serializers.SchemaRelationMetadata, name, kind, target, onDelete, through string) {
	t.Helper()
	for _, relation := range relations {
		if relation.Name == name {
			if relation.Type != kind || relation.To != target || relation.OnDelete != onDelete || relation.Through != through {
				t.Errorf("REST relation %s = %+v", name, relation)
			}
			return
		}
	}
	t.Errorf("REST relation %q not found", name)
}

func parsedModel(t *testing.T, defs []*codegen.ModelDefinition, name string) *codegen.ModelDefinition {
	t.Helper()
	for _, def := range defs {
		if def.Name == name {
			return def
		}
	}
	t.Fatalf("parsed model %q not found", name)
	return nil
}

func assertParsedField(t *testing.T, model *codegen.ModelDefinition, name string, required bool, maxLength int, defaultValue interface{}, unique bool) {
	t.Helper()
	for _, field := range model.Fields {
		if field.Name != name {
			continue
		}
		gotUnique, _ := field.Options["unique"].(bool)
		if field.Required != required || field.Default != defaultValue || gotUnique != unique || field.Options["max_length"] != maxLength {
			t.Errorf("parsed field %s = required:%v max_length:%v default:%v unique:%v", name, field.Required, field.Options["max_length"], field.Default, field.Options["unique"])
		}
		return
	}
	t.Errorf("parsed field %q not found", name)
}

func assertParsedRelation(t *testing.T, model *codegen.ModelDefinition, name, kind, onDelete, through string) {
	t.Helper()
	for _, relation := range model.Relations {
		if relation.Name == name {
			gotOnDelete, _ := relation.Options["on_delete"].(string)
			gotThrough, _ := relation.Options["through"].(string)
			if relation.Type != kind || gotOnDelete != onDelete || gotThrough != through {
				t.Errorf("parsed relation %s = %+v", name, relation)
			}
			return
		}
	}
	t.Errorf("parsed relation %q not found", name)
}

func assertMigrationColumn(t *testing.T, column *state.ColumnState, required bool, maxLength int, defaultValue interface{}, unique bool) {
	t.Helper()
	if column == nil {
		t.Error("migration column is missing")
		return
	}
	if column.Required != required || column.Default != defaultValue || column.Unique != unique || column.Options["max_length"] != maxLength {
		t.Errorf("migration column %s = required:%v max_length:%v default:%v unique:%v", column.Name, column.Required, column.Options["max_length"], column.Default, column.Unique)
	}
}

func hasForeignKey(foreignKeys map[string]*state.ForeignKeyState, column, target, onDelete string) bool {
	for _, fk := range foreignKeys {
		if fk.Column == column && fk.TargetTable == target && fk.OnDelete == onDelete {
			return true
		}
	}
	return false
}

func assertAdminField(t *testing.T, fields []admincore.FieldMetadata, name string, required bool, maxLength int, defaultValue interface{}, unique bool) {
	t.Helper()
	for _, field := range fields {
		if field.Name == name {
			if field.Required != required || field.MaxLength != maxLength || field.DefaultValue != defaultValue || field.Unique != unique {
				t.Errorf("admin field %s = required:%v max_length:%d default:%v unique:%v", name, field.Required, field.MaxLength, field.DefaultValue, field.Unique)
			}
			return
		}
	}
	t.Errorf("admin field %q not found", name)
}

func assertAdminRelation(t *testing.T, relations []admincore.RelationMetadata, name, kind, target, onDelete, through string) {
	t.Helper()
	for _, relation := range relations {
		if relation.Name == name {
			if relation.Type != kind || relation.RelatedModel != target || relation.OnDelete != onDelete || relation.Through != through {
				t.Errorf("admin relation %s = %+v", name, relation)
			}
			return
		}
	}
	t.Errorf("admin relation %q not found", name)
}
