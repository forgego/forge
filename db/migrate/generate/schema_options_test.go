package generate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/generate"
	"github.com/forgego/forge/db/migrate/parse"
	"github.com/forgego/forge/db/migrate/sql"
	"github.com/forgego/forge/db/migrate/state"
)

const schemaOptionsModels = `package models

import "github.com/forgego/forge/schema"

type Tag struct {
	schema.BaseSchema
}

func (Tag) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("slug", schema.Required(), schema.DBIndex()),
	}
}

func (Tag) Meta() schema.Meta {
	return schema.Meta{TableName: "tags"}
}

type Article struct {
	schema.BaseSchema
}

func (Article) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("tenant_id", schema.Required()),
		schema.StringField("code", schema.Required()),
		schema.StringField("title", schema.Required(), schema.DBIndex()),
	}
}

func (Article) Meta() schema.Meta {
	return schema.Meta{
		TableName:      "articles",
		UniqueTogether: [][]string{{"tenant_id", "code"}},
	}
}

func (Article) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ManyToManyField("tags", "Tag"),
		schema.ManyToManyField("featured_tags", "Tag", schema.Through("article_featured_tags")),
	}
}
`

func TestSchemaOptionsDBIndexUniqueTogetherAndManyToManyRoundTrip(t *testing.T) {
	modelsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(schemaOptionsModels), 0o600); err != nil {
		t.Fatal(err)
	}
	migDir := t.TempDir()

	gen, err := generate.NewMigrationGeneratorForDriver(modelsDir, migDir, core.DriverPostgreSQL)
	if err != nil {
		t.Fatal(err)
	}
	if err := gen.GenerateMigrations("init"); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(migDir)
	if err != nil {
		t.Fatal(err)
	}
	var upSQL string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			data, err := os.ReadFile(filepath.Join(migDir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			upSQL = string(data)
			break
		}
	}
	if upSQL == "" {
		t.Fatal("expected initial .up.sql migration to be generated")
	}

	expected := []string{
		`CREATE INDEX IF NOT EXISTS idx_tags_slug ON tags ("slug");`,
		`CREATE INDEX IF NOT EXISTS idx_articles_title ON articles ("title");`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uniq_articles_tenant_id_code ON articles ("tenant_id", "code");`,
		`CREATE TABLE IF NOT EXISTS articles_tags`,
		`CREATE TABLE IF NOT EXISTS article_featured_tags`,
		`ALTER TABLE articles_tags ADD CONSTRAINT fk_articles_tags_article_id FOREIGN KEY ("article_id") REFERENCES articles (id) ON DELETE CASCADE ON UPDATE NO ACTION;`,
		`ALTER TABLE articles_tags ADD CONSTRAINT fk_articles_tags_tag_id FOREIGN KEY ("tag_id") REFERENCES tags (id) ON DELETE CASCADE ON UPDATE NO ACTION;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uniq_articles_tags_article_id_tag_id ON articles_tags ("article_id", "tag_id");`,
		`CREATE INDEX IF NOT EXISTS idx_articles_tags_tag_id ON articles_tags ("tag_id");`,
	}
	for _, stmt := range expected {
		if !strings.Contains(upSQL, stmt) {
			t.Errorf("generated migration missing %q\nfull SQL:\n%s", stmt, upSQL)
		}
	}

	// Re-running makemigrations against the same models must be a no-op.
	if err := gen.GenerateMigrations("noop"); err != nil {
		t.Fatal(err)
	}
	entriesAfter, err := os.ReadDir(migDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entriesAfter) != len(entries) {
		t.Fatalf("expected no-op second migration, got %d files (was %d)", len(entriesAfter), len(entries))
	}
}

func TestManyToManySkipsExistingThroughModel(t *testing.T) {
	article := &generator.ModelDefinition{
		Name: "Article",
		Meta: generator.MetaDefinition{TableName: "articles"},
		Fields: []generator.FieldDefinition{
			{Name: "id", Type: "Int64Field", GoType: "int64", PrimaryKey: true, AutoIncrement: true},
		},
		Relations: []generator.RelationDefinition{
			{Name: "tags", Type: "ManyToManyField", To: "Tag", Options: map[string]interface{}{"through": "article_tags"}},
		},
	}
	tag := &generator.ModelDefinition{
		Name: "Tag",
		Meta: generator.MetaDefinition{TableName: "tags"},
		Fields: []generator.FieldDefinition{
			{Name: "id", Type: "Int64Field", GoType: "int64", PrimaryKey: true, AutoIncrement: true},
		},
	}
	explicitThrough := &generator.ModelDefinition{
		Name: "ArticleTag",
		Meta: generator.MetaDefinition{TableName: "article_tags"},
		Fields: []generator.FieldDefinition{
			{Name: "id", Type: "Int64Field", GoType: "int64", PrimaryKey: true, AutoIncrement: true},
			{Name: "article_id", Type: "Int64Field", GoType: "int64", Required: true},
			{Name: "tag_id", Type: "Int64Field", GoType: "int64", Required: true},
		},
	}

	detector := generate.NewDetector()
	changes, err := detector.DetectChanges([]*generator.ModelDefinition{article, tag, explicitThrough}, nil)
	if err != nil {
		t.Fatal(err)
	}

	createCount := 0
	for _, ch := range changes {
		if ct, ok := ch.(*core.CreateTable); ok && ct.Table.Meta.TableName == "article_tags" {
			createCount++
		}
	}
	if createCount != 1 {
		t.Fatalf("expected exactly 1 CreateTable for explicit through model article_tags, got %d", createCount)
	}
}

func TestDropIndexReversibleDownSQL(t *testing.T) {
	prev := []*generator.ModelDefinition{
		{
			Name: "Item",
			Meta: generator.MetaDefinition{
				TableName: "items",
				Indexes: []generator.IndexDefinition{
					{Name: "idx_items_code", Fields: []string{"code"}, Unique: true},
				},
			},
			Fields: []generator.FieldDefinition{
				{Name: "id", Type: "Int64Field", PrimaryKey: true},
				{Name: "code", Type: "StringField"},
			},
		},
	}
	curr := []*generator.ModelDefinition{
		{
			Name: "Item",
			Meta: generator.MetaDefinition{TableName: "items"},
			Fields: []generator.FieldDefinition{
				{Name: "id", Type: "Int64Field", PrimaryKey: true},
				{Name: "code", Type: "StringField"},
			},
		},
	}

	changes, err := generate.NewDetector().DetectChanges(curr, prev)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	dropIdx, ok := changes[0].(*core.DropIndex)
	if !ok {
		t.Fatalf("expected *core.DropIndex, got %T", changes[0])
	}
	if dropIdx.Index == nil || dropIdx.Index.Name != "idx_items_code" || !dropIdx.Index.Unique {
		t.Fatalf("expected DropIndex.Index to carry original index definition, got %+v", dropIdx.Index)
	}

	for _, driver := range []core.Driver{core.DriverPostgreSQL, core.DriverSQLite} {
		builder, err := sql.NewSQLBuilder(driver)
		if err != nil {
			t.Fatalf("%s NewSQLBuilder failed: %v", driver, err)
		}
		downSQL, err := builder.BuildDownSQL(changes)
		if err != nil {
			t.Fatalf("%s BuildDownSQL failed: %v", driver, err)
		}
		if !strings.Contains(downSQL, `CREATE UNIQUE INDEX IF NOT EXISTS idx_items_code ON items ("code");`) {
			t.Fatalf("%s down SQL did not recreate dropped index:\n%s", driver, downSQL)
		}
	}
}

func TestAlterColumnReadBack(t *testing.T) {
	parser := parse.NewSQLParser()
	mem := state.NewInMemoryState()

	createSQL := `CREATE TABLE items (id BIGINT PRIMARY KEY, title VARCHAR(50) NOT NULL DEFAULT 'untitled');`
	createChanges, err := parser.ParseUpSQL(createSQL)
	if err != nil {
		t.Fatal(err)
	}
	if err := mem.Apply(createChanges); err != nil {
		t.Fatal(err)
	}

	alterStmts := []string{
		`ALTER TABLE items ALTER COLUMN title TYPE TEXT USING title::TEXT;`,
		`ALTER TABLE items ALTER COLUMN title DROP NOT NULL;`,
		`ALTER TABLE items ALTER COLUMN title SET DEFAULT 'draft';`,
	}
	for _, stmt := range alterStmts {
		ch, err := parser.ParseUpSQL(stmt)
		if err != nil {
			t.Fatalf("ParseUpSQL(%q): %v", stmt, err)
		}
		if err := mem.Apply(ch); err != nil {
			t.Fatalf("Apply(%q): %v", stmt, err)
		}
	}

	defs := mem.GetState().ToModelDefinitions()
	if len(defs) != 1 {
		t.Fatalf("expected 1 model definition, got %d", len(defs))
	}
	var titleField *generator.FieldDefinition
	for i := range defs[0].Fields {
		if defs[0].Fields[i].Name == "title" {
			titleField = &defs[0].Fields[i]
			break
		}
	}
	if titleField == nil {
		t.Fatal("title field missing from reconstructed state")
	}
	if titleField.Type != "String" {
		t.Errorf("expected String after ALTER COLUMN TYPE TEXT, got %s", titleField.Type)
	}
	if _, hasMaxLen := titleField.Options["max_length"]; hasMaxLen {
		t.Errorf("expected max_length to be cleared after ALTER COLUMN TYPE TEXT, got %#v", titleField.Options["max_length"])
	}
	if titleField.Required {
		t.Errorf("expected Required=false after DROP NOT NULL")
	}
	if titleField.Default != "draft" {
		t.Errorf("expected Default='draft' after SET DEFAULT, got %#v", titleField.Default)
	}
}
