package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fencedGoSample(t *testing.T, path, after string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", path))
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	start := strings.Index(body, after)
	if start < 0 {
		t.Fatalf("anchor %q missing from %s", after, path)
	}
	body = body[start+len(after):]
	open := strings.Index(body, "```go")
	if open < 0 {
		t.Fatalf("Go fence missing after %q in %s", after, path)
	}
	body = body[open:]
	lineEnd := strings.IndexByte(body, '\n')
	body = body[lineEnd+1:]
	close := strings.Index(body, "```")
	if close < 0 {
		t.Fatalf("unclosed Go fence in %s", path)
	}
	return strings.TrimSpace(body[:close]) + "\n"
}

func TestPublishedModelSamplesGenerate(t *testing.T) {
	cases := []struct{ name, path, anchor, pkg string }{
		{"quickstart article", "docs-site/docs/quickstart.md", "Replace the example with your own model", "blog"},
		{"index article", "docs-site/docs/index.md", "Define a model and start querying", "models"},
		{"introduction article", "docs-site/docs/introduction.md", "### Familiar Patterns, Go Performance", "models"},
		{"readme product", "README.md", "A model:", "catalog"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model := fencedGoSample(t, tc.path, tc.anchor)
			if !strings.HasPrefix(model, "package ") {
				model = "package " + tc.pkg + "\nimport \"github.com/forgego/forge/schema\"\n" + model
			}
			exercise := fmt.Sprintf(`package %s
import "testing"
func TestModelFieldsCompileAndRun(t *testing.T) {
    fields := (%s{}).Fields()
    if len(fields) == 0 { t.Fatal("model has no fields") }
}
`, tc.pkg, map[string]string{"blog": "Article", "models": "Article", "catalog": "Product"}[tc.pkg])
			compileAndRunGeneratedSample(t, tc.pkg, model, exercise)
		})
	}
}

func TestPublishedTypedQuerySamplesRun(t *testing.T) {
	cases := []struct{ name, modelPath, modelAnchor, queryPath, queryAnchor, pkg, table, fields, insert string }{
		{"quickstart article", "docs-site/docs/quickstart.md", "Replace the example with your own model", "docs-site/docs/quickstart.md", "With the generated code, queries are typed", "blog", "articles", "id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, body TEXT, published BOOLEAN", "INSERT INTO articles (title, body, published) VALUES ('Go', 'body', 1)"},
		{"index article", "docs-site/docs/index.md", "Define a model and start querying", "docs-site/docs/index.md", "Type-safe queries with autocomplete", "models", "articles", "id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, content TEXT, published_at TIMESTAMP", "INSERT INTO articles (title, content) VALUES ('Go', 'body')"},
		{"readme product", "README.md", "A model:", "README.md", "A typed query, using the code", "catalog", "products", "id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, price REAL, is_active BOOLEAN", "INSERT INTO products (name, price, is_active) VALUES ('Desk', 100, 1)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model := fencedGoSample(t, tc.modelPath, tc.modelAnchor)
			if !strings.HasPrefix(model, "package ") {
				model = "package " + tc.pkg + "\nimport \"github.com/forgego/forge/schema\"\n" + model
			}
			sample := fencedGoSample(t, tc.queryPath, tc.queryAnchor)
			for _, prefix := range []string{"blog.", "models.", "catalog."} {
				sample = strings.ReplaceAll(sample, prefix, "")
			}
			sample = strings.ReplaceAll(sample, "return err", "t.Fatal(err)")
			sample += "\nif err != nil { t.Fatal(err) }; if len(" + map[string]string{"catalog": "products", "blog": "articles", "models": "articles"}[tc.pkg] + ") != 1 { t.Fatal(\"query returned wrong number of rows\") }\n"
			exercise := fmt.Sprintf(`package %s
import (
    "context"
    "testing"
    "github.com/forgego/forge/db"
    "github.com/forgego/forge/orm"
)
var _ = orm.And
func TestTypedQueryRuns(t *testing.T) {
    database, err := db.NewDBWithDriver("sqlite3", ":memory:", db.WithMaxOpenConns(1))
    if err != nil { t.Fatal(err) }
    defer database.Close()
    if _, err := database.Exec(%q); err != nil { t.Fatal(err) }
    if _, err := database.Exec(%q); err != nil { t.Fatal(err) }
    %sObjects.SetDB(database)
    ctx := context.Background()
    %s
}
`, tc.pkg, "CREATE TABLE "+tc.table+" ("+tc.fields+")", tc.insert, map[string]string{"catalog": "Product", "blog": "Article", "models": "Article"}[tc.pkg], sample)
			compileAndRunGeneratedSample(t, tc.pkg, model, exercise)
		})
	}
}
