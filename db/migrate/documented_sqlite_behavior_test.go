package migrate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/schema"
)

func applyDocumentedSQLiteMigration(t *testing.T, up string) *db.DB {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "000001_documented.up.sql"), []byte(up), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "000001_documented.down.sql"), []byte("-- no rollback in this test\n"), 0600); err != nil {
		t.Fatal(err)
	}
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "documented.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	runner, err := db.NewMigrationRunner(database, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runner.Close() })
	if err := runner.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	return database
}

func TestDocumentedMetaIndexAndCheckHelpersApply(t *testing.T) {
	models := t.TempDir()
	migrations := t.TempDir()
	// Adapt the API reference's Meta example to SQLite's supported B-tree indexes.
	const model = `package models
import "github.com/forgego/forge/schema"
type Product struct { ProductGenerated }
func (Product) Fields() []schema.Field { return []schema.Field{
	schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
	schema.StringField("sku"), schema.StringField("slug"), schema.DecimalField("price"),
} }
func (Product) Meta() schema.Meta { return schema.Meta{
	TableName: "inventory_products",
	Indexes: []schema.Index{
		schema.IndexOn("idx_product_sku", "sku"),
		schema.UniqueIndexOn("idx_product_slug", "slug"),
	},
	Constraints: []schema.Constraint{
		{Name: "chk_product_price", Type: schema.ConstraintTypeCheck, Condition: "price >= 0"},
	},
} }
`
	if err := os.WriteFile(filepath.Join(models, "products.go"), []byte(model), 0600); err != nil {
		t.Fatal(err)
	}
	if err := GenerateForDriver("products", models, migrations, core.DriverSQLite); err != nil {
		t.Fatal(err)
	}
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "products.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	runner, err := db.NewMigrationRunner(database, migrations)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runner.Close() })
	if err := runner.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO inventory_products (sku, slug, price) VALUES ('k1', 'same', 10)`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO inventory_products (sku, slug, price) VALUES ('k2', 'same', 10)`); err == nil {
		t.Fatal("unique index did not reject duplicate slug")
	}
	if _, err := database.Exec(`INSERT INTO inventory_products (sku, slug, price) VALUES ('k3', 'new', -1)`); err == nil {
		t.Fatal("check constraint did not reject negative price")
	}
	var count int
	if err := database.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='index' AND name IN ('idx_product_sku', 'idx_product_slug')`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("documented indexes count = %d, err = %v", count, err)
	}
}

func TestManyToManyThroughTableRoundTrip(t *testing.T) {
	relation := schema.ManyToManyField("tags", "Tag", schema.Through("order_tags"), schema.RelatedName("orders"))
	if relation.Type != schema.RelationManyToMany || relation.Through != "order_tags" || relation.RelatedName != "orders" {
		t.Fatalf("relation = %+v", relation)
	}
	database := applyDocumentedSQLiteMigration(t, `
CREATE TABLE orders (id INTEGER PRIMARY KEY);
CREATE TABLE tags (id INTEGER PRIMARY KEY, name TEXT NOT NULL);
CREATE TABLE order_tags (order_id INTEGER NOT NULL REFERENCES orders(id), tag_id INTEGER NOT NULL REFERENCES tags(id), PRIMARY KEY (order_id, tag_id));`)
	if _, err := database.Exec(`INSERT INTO orders VALUES (1); INSERT INTO tags VALUES (10, 'urgent'), (11, 'gift'); INSERT INTO order_tags VALUES (1, 10), (1, 11)`); err != nil {
		t.Fatal(err)
	}
	rows, err := database.Query(`SELECT tags.name FROM tags JOIN order_tags ON tags.id = order_tags.tag_id WHERE order_tags.order_id = ? ORDER BY tags.id`, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		got = append(got, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "urgent" || got[1] != "gift" {
		t.Fatalf("through table returned %v", got)
	}
}

func TestGeneratedColumnMigrationEvaluatesStoredExpression(t *testing.T) {
	field := schema.DecimalField("price_with_tax", schema.GeneratedColumn("price * (1 + tax_rate)", true))
	if !field.Generated || !field.IsStored || field.GeneratedExpr != "price * (1 + tax_rate)" {
		t.Fatalf("generated field = %+v", field)
	}
	database := applyDocumentedSQLiteMigration(t, `CREATE TABLE products (
id INTEGER PRIMARY KEY, price REAL NOT NULL, tax_rate REAL NOT NULL,
price_with_tax REAL GENERATED ALWAYS AS (price * (1 + tax_rate)) STORED);`)
	if _, err := database.Exec(`INSERT INTO products (id, price, tax_rate) VALUES (1, 100, 0.1)`); err != nil {
		t.Fatal(err)
	}
	var total float64
	if err := database.QueryRow(`SELECT price_with_tax FROM products WHERE id = 1`).Scan(&total); err != nil || total < 109.999 || total > 110.001 {
		t.Fatalf("insert computed total = %v, err = %v", total, err)
	}
	if _, err := database.Exec(`UPDATE products SET price = 200 WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT price_with_tax FROM products WHERE id = 1`).Scan(&total); err != nil || total < 219.999 || total > 220.001 {
		t.Fatalf("updated computed total = %v, err = %v", total, err)
	}
}
