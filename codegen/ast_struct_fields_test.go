package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func parseStructFieldSource(t *testing.T, source string) []Diagnostic {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "models.go")
	if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	parser := NewASTParser()
	if _, err := parser.ParseFile(filename); err != nil {
		t.Fatal(err)
	}
	return parser.Diagnostics()
}

func TestStructFieldWithoutSchemaEntryProducesDiagnostic(t *testing.T) {
	source := `package users
import "github.com/forgego/forge/schema"
type User struct {
	schema.BaseSchema
	ID           int64  ` + "`db:\"id\"`" + `
	Email        string ` + "`json:\"email\"`" + `
	PasswordHash string ` + "`json:\"-\"`" + `
}
func (User) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("email"),
	}
}
`
	diagnostics := parseStructFieldSource(t, source)
	if len(diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %#v", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Line != sourceLine(source, "PasswordHash string") {
		t.Errorf("line = %d, want %d", diagnostic.Line, sourceLine(source, "PasswordHash string"))
	}
	if diagnostic.Model != "User" || diagnostic.Method != "Fields" {
		t.Errorf("model/method = %s/%s, want User/Fields", diagnostic.Model, diagnostic.Method)
	}
	if !strings.Contains(diagnostic.Message, `struct field PasswordHash has no Fields() or Relations() entry for "password_hash"`) {
		t.Errorf("message = %q", diagnostic.Message)
	}
}

func TestStructFieldsMatchingSchemaProduceNoDiagnostics(t *testing.T) {
	source := `package blog
import "github.com/forgego/forge/schema"
type Post struct {
	schema.BaseSchema
	ID       int64    ` + "`db:\"id\"`" + `
	Title    string   ` + "`db:\"display_title\"`" + `
	AuthorID int64    ` + "`db:\"author_id\"`" + `
	Author   *Author  ` + "`db:\"author\"`" + `
	Tags     []Tag
	Cached   string   ` + "`db:\"-\"`" + `
	internal string
}
func (Post) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.DBColumn("display_title")),
		schema.Int64Field("author_id"),
	}
}
func (Post) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("Author", "Author"),
		schema.ManyToManyField("Tags", "Tag", schema.Through("post_tags")),
	}
}
type Author struct{}
type Tag struct{}
`
	if diagnostics := parseStructFieldSource(t, source); len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
}

func TestSchemaModelWithoutFieldsMethodProducesDiagnostic(t *testing.T) {
	source := `package models
import "github.com/forgego/forge/schema"
type User struct {
	schema.BaseSchema
	Password string ` + "`db:\"password\"`" + `
}
`
	diagnostics := parseStructFieldSource(t, source)
	if len(diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Line != sourceLine(source, "type User struct") {
		t.Errorf("line = %d, want %d", diagnostics[0].Line, sourceLine(source, "type User struct"))
	}
	if !strings.Contains(diagnostics[0].Message, "has no Fields() method") {
		t.Errorf("message = %q", diagnostics[0].Message)
	}
}

func TestGeneratedEmbeddingWithoutFieldsMethodProducesNoDiagnostic(t *testing.T) {
	source := `package models
type UserGenerated struct{}
type User struct {
	UserGenerated
}
`
	if diagnostics := parseStructFieldSource(t, source); len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
}
