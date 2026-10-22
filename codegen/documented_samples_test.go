package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuickstartArticleSampleGeneratesAndQueries(t *testing.T) {
	const model = `package blog
import "github.com/forgego/forge/schema"
type Article struct { ArticleGenerated }
func (Article) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required(), schema.MaxLength(200)),
		schema.TextField("body", schema.Required()),
		schema.BoolField("published", schema.Default(false)),
	}
}
func (Article) Meta() schema.Meta { return schema.Meta{TableName: "articles", VerboseName: "Article"} }
func (Article) Relations() []schema.Relation { return nil }
func (Article) Hooks() *schema.ModelHooks { return nil }
`
	const exercise = `package blog
import (
	"context"
	"testing"
	"github.com/forgego/forge/db"
)
func TestArticleQuery(t *testing.T) {
	database, err := db.NewDBWithDriver("sqlite3", ":memory:", db.WithMaxOpenConns(1))
	if err != nil { t.Fatal(err) }
	defer database.Close()
	if _, err := database.Exec("CREATE TABLE articles (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, body TEXT NOT NULL, published INTEGER NOT NULL)"); err != nil { t.Fatal(err) }
	ArticleObjects.SetDB(database)
	ctx := context.Background()
	if err := ArticleObjects.Create(ctx, &Article{ArticleGenerated: ArticleGenerated{Title: "Go", Body: "sample", Published: true}}); err != nil { t.Fatal(err) }
	f := ArticleFieldsInstance
	qs, err := ArticleObjects.Filter(f.Published.Eq(true))
	if err != nil { t.Fatal(err) }
	articles, err := qs.OrderBy(f.Id.Desc()).Limit(10).All(ctx)
	if err != nil || len(articles) != 1 || articles[0].Title != "Go" { t.Fatalf("articles=%+v, err=%v", articles, err) }
}
`
	compileAndRunGeneratedSample(t, "blog", model, exercise)
}

func TestIntroductionArticleSampleGeneratesAndQueries(t *testing.T) {
	const model = `package blog
import "github.com/forgego/forge/schema"
type Article struct { ArticleGenerated }
func (Article) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.MaxLength(200)),
		schema.TextField("content"),
		schema.TimeField("published_at", schema.AutoNow()),
	}
}
`
	const exercise = `package blog
import (
	"context"
	"testing"
	"github.com/forgego/forge/db"
)
func TestArticleQuery(t *testing.T) {
	database, err := db.NewDBWithDriver("sqlite3", ":memory:", db.WithMaxOpenConns(1))
	if err != nil { t.Fatal(err) }
	defer database.Close()
	if _, err := database.Exec("CREATE TABLE articles (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, content TEXT, published_at TEXT)"); err != nil { t.Fatal(err) }
	ArticleObjects.SetDB(database)
	ctx := context.Background()
	if err := ArticleObjects.Create(ctx, &Article{ArticleGenerated: ArticleGenerated{Title: "Go", Content: "sample"}}); err != nil { t.Fatal(err) }
	a := ArticleFieldsInstance
	qs, err := ArticleObjects.Filter(a.Title.Contains("Go"))
	if err != nil { t.Fatal(err) }
	articles, err := qs.OrderBy(a.PublishedAt.Desc()).Limit(10).All(ctx)
	if err != nil || len(articles) != 1 || articles[0].Title != "Go" { t.Fatalf("articles=%+v, err=%v", articles, err) }
}
`
	compileAndRunGeneratedSample(t, "blog", model, exercise)
}

func compileAndRunGeneratedSample(t *testing.T, pkg, model, exercise string) {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	module := t.TempDir()
	dir := filepath.Join(module, pkg)
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "models.go"), []byte(model), 0644); err != nil {
		t.Fatal(err)
	}
	if err := NewGenerator(dir, dir).Generate(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "article_test.go"), []byte(exercise), 0644); err != nil {
		t.Fatal(err)
	}
	goMod := fmt.Sprintf("module example.com/documented\n\ngo 1.26.0\n\nrequire github.com/forgego/forge v0.0.0\n\nreplace github.com/forgego/forge => %s\n", root)
	if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}
	sum, err := os.ReadFile(filepath.Join(root, "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "go.sum"), sum, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "test", "-mod=mod", "./...")
	cmd.Dir = module
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated sample failed: %v\n%s", err, strings.TrimSpace(string(output)))
	}
}
