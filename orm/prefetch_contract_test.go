package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/dialect"
	"github.com/mattn/go-sqlite3"
)

var prefetchQueryCount atomic.Int64

func init() { sql.Register("prefetch-counting-sqlite", prefetchCountingDriver{}) }

type prefetchCountingDriver struct{}

func (prefetchCountingDriver) Open(name string) (driver.Conn, error) {
	conn, err := (&sqlite3.SQLiteDriver{}).Open(name)
	if err != nil {
		return nil, err
	}
	return &prefetchCountingConn{Conn: conn}, nil
}

type prefetchCountingConn struct{ driver.Conn }

func (c *prefetchCountingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return c.Conn.(driver.ExecerContext).ExecContext(ctx, query, args)
}

func (c *prefetchCountingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	prefetchQueryCount.Add(1)
	return c.Conn.(driver.QueryerContext).QueryContext(ctx, query, args)
}

func TestPrefetchRelatedBatchesMultivaluedRelations(t *testing.T) {
	sqlDB, err := sql.Open("prefetch-counting-sqlite", filepath.Join(t.TempDir(), "prefetch.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	database := &db.DB{DB: sqlDB, Driver: "sqlite3"}
	database.SetDialect(dialect.NewSQLiteDialect())
	if _, err := database.Exec(`
		CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT);
		CREATE TABLE categories (id INTEGER PRIMARY KEY, name TEXT);
		CREATE TABLE product_categories (product_id INTEGER, category_id INTEGER);
		INSERT INTO products (id, name) VALUES (1, 'Laptop'), (2, 'Phone');
		INSERT INTO categories (id, name) VALUES (1, 'Electronics'), (2, 'Computers');
		INSERT INTO product_categories (product_id, category_id) VALUES (1, 1), (1, 2), (2, 1);
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := GetModelSchema[Category](); err != nil {
		t.Fatal(err)
	}
	if _, err := GetModelSchema[Product](); err != nil {
		t.Fatal(err)
	}
	qs, err := NewQuerySet[Product]("products")
	if err != nil {
		t.Fatal(err)
	}
	prefetchQueryCount.Store(0)
	products, err := qs.SetDB(database).OrderBy("id").PrefetchRelated("Categories").All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 2 || len(products[0].Categories) != 2 || len(products[1].Categories) != 1 || products[1].Categories[0].Name != "Electronics" {
		t.Fatalf("products=%+v", products)
	}
	if got := prefetchQueryCount.Load(); got != 3 {
		t.Fatalf("SELECT queries=%d, want base, join, target queries", got)
	}
}
