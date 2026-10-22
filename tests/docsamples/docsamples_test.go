package docsamples

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func sample(t *testing.T, root, path, anchor string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	pos := strings.Index(content, anchor)
	if pos < 0 {
		t.Fatalf("missing anchor %q", anchor)
	}
	content = content[pos+len(anchor):]
	pos = strings.Index(content, "```go")
	if pos < 0 {
		t.Fatalf("missing Go sample in %s", path)
	}
	content = content[pos:]
	content = content[strings.IndexByte(content, '\n')+1:]
	pos = strings.Index(content, "```")
	if pos < 0 {
		t.Fatalf("unclosed Go sample in %s", path)
	}
	return strings.TrimSpace(content[:pos]) + "\n"
}

func removeHeader(src string) string {
	if strings.HasPrefix(src, "package ") {
		src = src[strings.IndexByte(src, '\n')+1:]
	}
	src = strings.TrimSpace(src)
	if strings.HasPrefix(src, "import (") {
		src = src[strings.Index(src, "\n)")+2:]
	} else if strings.HasPrefix(src, "import ") {
		src = src[strings.IndexByte(src, '\n')+1:]
	}
	return strings.TrimSpace(src)
}

const imports = `package catalog
import (
    "context"
    "errors"
    "net/http"
    "net/http/httptest"
    "strconv"
    "testing"
    "github.com/forgego/forge/api"
    "github.com/forgego/forge/api/docs"
    "github.com/forgego/forge/api/throttling"
    dbpkg "github.com/forgego/forge/db"
    "github.com/forgego/forge/identity"
    "github.com/forgego/forge/orm"
    "github.com/forgego/forge/schema"
    "github.com/forgego/forge/server"
    "github.com/go-chi/chi/v5"
)
var _ = []any{context.Background, errors.New, http.MethodGet, httptest.NewRecorder, strconv.Itoa, api.NewRouter, docs.NewOpenAPIGenerator, throttling.NewUserRateThrottle, dbpkg.NewDBWithDriver, identity.SetupIdentitySystem, orm.F, schema.Int64Field, server.NewRouter, chi.URLParam}
`

const support = `
type Product struct {
    schema.BaseSchema
    ID int64 ` + "`json:\"id\" db:\"id\"`" + `
    Sku string ` + "`json:\"sku\" db:\"sku\"`" + `
    Name string ` + "`json:\"name\" db:\"name\"`" + `
    Price float64 ` + "`json:\"price\" db:\"price\"`" + `
    InStock bool ` + "`json:\"in_stock\" db:\"in_stock\"`" + `
    IsFeatured bool ` + "`json:\"is_featured\" db:\"is_featured\"`" + `
    Archived bool ` + "`json:\"archived\" db:\"archived\"`" + `
    Stock int64 ` + "`json:\"stock\" db:\"stock\"`" + `
}
func (Product) Fields() []schema.Field { return []schema.Field{
    schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
    schema.StringField("sku"), schema.StringField("name"),
    schema.Float64Field("price"), schema.BoolField("in_stock"),
    schema.BoolField("is_featured"), schema.BoolField("archived"), schema.Int64Field("stock"),
} }
var ProductObjects *orm.Manager[Product]
var ProductManager *orm.Manager[Product]
var ProductExpr = struct{ InStock orm.Field[bool]; Price orm.Field[float64]; Stock orm.Field[int64] }{
    orm.NewField[bool]("in_stock", "products"), orm.NewField[float64]("price", "products"), orm.NewField[int64]("stock", "products"),
}
func setup(t *testing.T) {
    t.Helper()
    database, err := dbpkg.NewDBWithDriver("sqlite3", ":memory:", dbpkg.WithMaxOpenConns(1))
    if err != nil { t.Fatal(err) }
    t.Cleanup(func(){ _ = database.Close() })
    _, err = database.Exec("CREATE TABLE products (id INTEGER PRIMARY KEY AUTOINCREMENT, sku TEXT, name TEXT, price REAL, in_stock BOOLEAN, is_featured BOOLEAN NOT NULL DEFAULT 0, archived BOOLEAN NOT NULL DEFAULT 0, stock INTEGER)")
    if err != nil { t.Fatal(err) }
    ProductObjects, err = orm.NewManagerWithDB[Product]("products", database)
    if err != nil { t.Fatal(err) }
    ProductManager = ProductObjects
}
func request(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
    rec := httptest.NewRecorder()
    req := httptest.NewRequest(method, path, stringsReader(body))
    router.ServeHTTP(rec, req)
    return rec
}
func stringsReader(s string) *strings.Reader { return strings.NewReader(s) }
`

func TestPublishedAPISamplesCompileAndRun(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, path, anchor, extra, exercise string }{
		{"branching queryset", "docs-site/docs/orm.md", "Every filter or ordering method", "", `setup(t)
ctx := context.Background()
for _, p := range []Product{{Name:"budget", Price:20, InStock:true},{Name:"premium", Price:600, InStock:true}} { if err := ProductObjects.Create(ctx, &p); err != nil { t.Fatal(err) } }
`},
		{"complete product API", "docs-site/docs/api/overview.md", "## Complete Example", "", `setup(t)
router := server.NewRouter()
RegisterProductAPI(router)
rec := request(router, http.MethodGet, "/api/v1/products/", "")
if rec.Code != 200 { t.Fatalf("status=%d: %s", rec.Code, rec.Body.String()) }
`},
		{"custom pagination", "docs-site/docs/api/pagination.md", "## Paginating custom handlers", "", `r := httptest.NewRequest(http.MethodGet, "/products?page=2&page_size=1", nil)
w := httptest.NewRecorder()
loadRows := func(offset, size int) ([]string, int) { return []string{"second"}, 3 }
`},
		{"product actions", "docs-site/docs/api/viewsets.md", "## Adding Custom Action Routes", "", `setup(t)
if err := ProductObjects.Create(context.Background(), &Product{Name:"desk", Price:100, IsFeatured:true}); err != nil { t.Fatal(err) }
apiRouter := api.NewRouter("/api/v1")
`},
		{"manager scoped routes", "docs-site/docs/api/viewsets.md", "## Scoping the Records a ViewSet Serves", "", `setup(t)
viewset := api.NewBaseViewSet(NewProductSerializer, ActiveProducts{ProductObjects}, &Product{})
if viewset.Queryset == nil { t.Fatal("missing manager") }
if err := ProductObjects.Create(context.Background(), &Product{Name:"visible"}); err != nil { t.Fatal(err) }
if err := ProductObjects.Create(context.Background(), &Product{Name:"hidden", Archived:true}); err != nil { t.Fatal(err) }
apiRouter := api.NewRouter("/api/v1")
apiRouter.Register("products", viewset)
router := server.NewRouter()
apiRouter.RegisterRoutes(router)
list := request(router, http.MethodGet, "/api/v1/products/", "")
if list.Code != 200 || !strings.Contains(list.Body.String(), "visible") || strings.Contains(list.Body.String(), "hidden") { t.Fatalf("list=%d %s", list.Code, list.Body.String()) }
detail := request(router, http.MethodGet, "/api/v1/products/2", "")
if detail.Code != 404 { t.Fatalf("archived detail status=%d", detail.Code) }
`},
		{"nested serializer", "docs-site/docs/api/serializers.md", "## Nested Relational Serialization", "type Customer struct { Name string `json:\"name\"` }\n", `order := Order{Customer: &Customer{Name:"Ada"}}
if order.Customer.Name != "Ada" { t.Fatal("nested model missing") }
`},
		{"OpenAPI serving", "docs-site/docs/api/openapi.md", "## Serving the document", "", `router := server.NewRouter()
`},
		{"Meta helpers", "docs-site/docs/api-reference/schema.md", "### Common Meta Helpers", "", `meta := (Product{}).Meta()
if len(meta.Indexes) != 3 || len(meta.Constraints) != 1 { t.Fatalf("meta=%+v", meta) }
`},
		{"order hooks", "docs-site/docs/api-reference/hooks.md", "## Practical Example", "type Order struct { ID int64; ShippingAmount, TotalAmount, Subtotal, TaxAmount float64 }\nfunc NotifyOrderPlaced(int64) {}\n", `order := &Order{Subtotal:10, TaxAmount:2, ShippingAmount:3}
hooks := (Order{}).Hooks()
if err := hooks.BeforeSave(context.Background(), order); err != nil { t.Fatal(err) }
if order.TotalAmount != 15 { t.Fatalf("total=%v", order.TotalAmount) }
`},
		{"custom manager", "docs-site/docs/api-reference/manager.md", "## Custom Managers", "", `setup(t)
m := &CustomProductManager{ProductManager}
if _, err := m.InStockOnly(); err != nil { t.Fatal(err) }
`},
		{"identity setup", "docs-site/docs/identity.md", "## Setup example", "", `var db *dbpkg.DB
`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snippet := removeHeader(sample(t, root, tc.path, tc.anchor))
			extra := tc.extra
			exercise := tc.exercise
			switch tc.name {
			case "branching queryset":
				exercise += snippet + `
if err != nil || len(budgetItems) != 1 || len(premiumItems) != 1 { t.Fatalf("branches: budget=%v premium=%v err=%v", budgetItems, premiumItems, err) }
`
				snippet = ""
			case "custom pagination":
				exercise += snippet + `
if w.Code != 200 { t.Fatalf("status=%d: %s", w.Code, w.Body.String()) }
`
				snippet = ""
			case "product actions":
				exercise += `RegisterProductActions(apiRouter)
router := server.NewRouter()
apiRouter.RegisterRoutes(router)
rec := request(router, http.MethodGet, "/api/v1/products/featured", "")
if rec.Code != 200 { t.Fatalf("status=%d: %s", rec.Code, rec.Body.String()) }
`
			case "manager scoped routes":
				pos := strings.Index(snippet, "viewset :=")
				if pos < 0 {
					t.Fatal("viewset construction missing")
				}
				exercise = strings.Replace(exercise, "viewset := api.NewBaseViewSet(NewProductSerializer, ActiveProducts{ProductObjects}, &Product{})", snippet[pos:], 1)
				snippet = snippet[:pos]
			case "OpenAPI serving":
				exercise += snippet + `
rec := request(router, http.MethodGet, "/api/openapi.json", "")
if rec.Code != 200 { t.Fatalf("status=%d: %s", rec.Code, rec.Body.String()) }
`
				snippet = ""
			case "identity setup":
				exercise += snippet
				snippet = ""
			}
			if tc.name == "product actions" || tc.name == "manager scoped routes" {
				extra += `
type ProductSerializer struct { *api.BaseSerializer }
func NewProductSerializer() api.Serializer { return &ProductSerializer{api.NewBaseSerializer(nil)} }
`
			}
			source := imports + "\n" + support + "\n" + extra + "\n" + snippet + "\nfunc TestSample(t *testing.T) {\n" + exercise + "\n}\n"
			source = strings.Replace(source, "\"testing\"", "\"testing\"\n\"strings\"", 1)
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "sample_test.go"), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			mod := fmt.Sprintf("module example.com/docsample\n\ngo 1.26.10\n\nrequire github.com/forgego/forge v0.0.0\n\nreplace github.com/forgego/forge => %s\n", root)
			if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0644); err != nil {
				t.Fatal(err)
			}
			sum, err := os.ReadFile(filepath.Join(root, "go.sum"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "go.sum"), sum, 0644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "test", "-mod=mod", "./...")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=", "GOTOOLCHAIN=go1.26.10")
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("published sample did not compile and run: %v\n%s", err, output)
			}
		})
	}
}
