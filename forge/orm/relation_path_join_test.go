package orm

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCompany struct {
	schema.BaseSchema
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (TestCompany) Meta() schema.Meta {
	return schema.Meta{TableName: "companies"}
}

func (TestCompany) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

type TestCustomer struct {
	schema.BaseSchema
	ID        int64        `db:"id"`
	Name      string       `db:"name"`
	CompanyID int64        `db:"company_id"`
	Company   *TestCompany `db:"company"`
}

func (TestCustomer) Meta() schema.Meta {
	return schema.Meta{TableName: "customers"}
}

func (TestCustomer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
		schema.Int64Field("company_id"),
	}
}

func (TestCustomer) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("company", "TestCompany"),
		schema.ForeignKeyField("orders", "TestOrder"),
	}
}

type TestOrder struct {
	schema.BaseSchema
	ID         int64         `db:"id"`
	Total      float64       `db:"total"`
	CustomerID int64         `db:"customer_id"`
	Customer   *TestCustomer `db:"customer"`
}

func (TestOrder) Meta() schema.Meta {
	return schema.Meta{TableName: "orders"}
}

func (TestOrder) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.FloatField("total"),
		schema.Int64Field("customer_id"),
	}
}

func (TestOrder) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer", "TestCustomer"),
	}
}

func setupRelationTestDB(t *testing.T) *db.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "relation_path_join_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)

	_, err = database.Exec(`
		CREATE TABLE companies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);
		CREATE TABLE customers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			company_id INTEGER,
			FOREIGN KEY(company_id) REFERENCES companies(id)
		);
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			total REAL NOT NULL,
			customer_id INTEGER,
			FOREIGN KEY(customer_id) REFERENCES customers(id)
		);
	`)
	require.NoError(t, err)

	// Register schemas
	_, err = GetModelSchema[TestCompany]()
	require.NoError(t, err)
	_, err = GetModelSchema[TestCustomer]()
	require.NoError(t, err)
	_, err = GetModelSchema[TestOrder]()
	require.NoError(t, err)

	return database
}

func seedRelationTestData(t *testing.T, database *db.DB) {
	t.Helper()
	// Insert Companies
	_, err := database.Exec(`INSERT INTO companies (id, name) VALUES (1, 'Acme Corp'), (2, 'Beta LLC')`)
	require.NoError(t, err)

	// Insert Customers
	_, err = database.Exec(`INSERT INTO customers (id, name, company_id) VALUES (1, 'Acme', 1), (2, 'Bob', 2)`)
	require.NoError(t, err)

	// Insert Orders
	// Order 1 & 2 for customer 1 ('Acme'), Order 3 for customer 2 ('Bob')
	_, err = database.Exec(`INSERT INTO orders (id, total, customer_id) VALUES (1, 100.0, 1), (2, 200.0, 1), (3, 300.0, 2)`)
	require.NoError(t, err)
}

func TestRelationPathJoin_FilterOneHop(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()
	seedRelationTestData(t, database)

	ctx := context.Background()
	qs, err := NewQuerySet[TestOrder]("orders")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	// Filter Order by customer__name == "Acme" returns only Acme's orders (All) and Count matches
	orders, err := qs.Filter(F("customer__name").Eq("Acme")).All(ctx)
	require.NoError(t, err)
	require.Len(t, orders, 2)
	assert.Equal(t, int64(1), orders[0].ID)
	assert.Equal(t, int64(2), orders[1].ID)

	count, err := qs.Filter(F("customer__name").Eq("Acme")).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRelationPathJoin_FilterTwoHops(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()
	seedRelationTestData(t, database)

	ctx := context.Background()
	qs, err := NewQuerySet[TestOrder]("orders")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	// Filter Order by customer__company__name (two hops) works
	twoHopQS := qs.Filter(F("customer__company__name").Eq("Acme Corp"))
	sql, _, err := twoHopQS.(*BaseQuerySet[TestOrder]).buildSQL()
	require.NoError(t, err)
	t.Logf("Exact SQL for two-hop filter: %s", sql)
	assert.Equal(t, `SELECT "orders".* FROM "orders" LEFT JOIN "customers" AS "customer" ON "customer"."id" = "orders"."customer_id" LEFT JOIN "companies" AS "customer__company" ON "customer__company"."id" = "customer"."company_id" WHERE "customer__company"."name" = ?1`, sql)

	orders, err := twoHopQS.All(ctx)
	require.NoError(t, err)
	require.Len(t, orders, 2)
	assert.Equal(t, int64(1), orders[0].ID)
	assert.Equal(t, int64(2), orders[1].ID)

	count, err := twoHopQS.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRelationPathJoin_OrderBy(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()
	seedRelationTestData(t, database)

	ctx := context.Background()
	qs, err := NewQuerySet[TestOrder]("orders")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	// OrderBy("customer__name") sorts correctly
	// Descending by customer__name: "Bob" orders come before "Acme" orders
	orders, err := qs.OrderBy("-customer__name", "id").All(ctx)
	require.NoError(t, err)
	require.Len(t, orders, 3)
	assert.Equal(t, int64(3), orders[0].ID) // Bob's order
	assert.Equal(t, int64(1), orders[1].ID) // Acme's order 1
	assert.Equal(t, int64(2), orders[2].ID) // Acme's order 2
}

func TestRelationPathJoin_CountDistinctJoin(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()
	seedRelationTestData(t, database)

	ctx := context.Background()
	// Customer Acme has two orders. Count on Customer filtered by a one-to-many path.
	customerQS, err := NewQuerySet[TestCustomer]("customers")
	require.NoError(t, err)
	customerQS = customerQS.SetDB(database)

	// Customer filtered by orders__total > 50
	count, err := customerQS.Filter(F("orders__total").Gt(50.0)).Count(ctx)
	require.NoError(t, err)
	// Acme has 2 orders, Bob has 1 order. Both customers have orders > 50. Total distinct customers = 2.
	// If COUNT(*) was used instead of COUNT(DISTINCT ...), count would be 3!
	assert.Equal(t, int64(2), count)

	// Also verify that SQL for Count on orders with customer join uses COUNT(DISTINCT ...)
	orderQS, err := NewQuerySet[TestOrder]("orders")
	require.NoError(t, err)
	orderQS = orderQS.SetDB(database)
	countSQL, _, err := orderQS.Filter(F("customer__name").Eq("Acme")).(*BaseQuerySet[TestOrder]).buildSQL()
	require.NoError(t, err)
	assert.Contains(t, countSQL, "LEFT JOIN")
}

func TestRelationPathJoin_NoRelationPathProducesIdenticalSQL(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()

	orderQS, err := NewQuerySet[TestOrder]("orders")
	require.NoError(t, err)
	orderQS = orderQS.SetDB(database)

	// Simple filter without relation path: assert SQL byte-identical
	sql, _, err := orderQS.Filter(F("total").Eq(100.0)).(*BaseQuerySet[TestOrder]).buildSQL()
	require.NoError(t, err)
	const expectedSQL = `SELECT * FROM "orders" WHERE "total" = ?1`
	assert.Equal(t, expectedSQL, sql)
}

func TestRelationPathJoin_UpdateDeleteError(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()
	seedRelationTestData(t, database)

	ctx := context.Background()
	orderQS, err := NewQuerySet[TestOrder]("orders")
	require.NoError(t, err)
	orderQS = orderQS.SetDB(database)

	// Update with a relation-path filter returns the "not supported" error
	_, err = orderQS.Filter(F("customer__name").Eq("Acme")).Update(ctx, UpdateMap{"total": 999.0})
	require.Error(t, err)
	assert.Equal(t, "filtering by related fields is not supported in update", err.Error())

	// Delete with a relation-path filter returns the "not supported" error
	_, err = orderQS.Filter(F("customer__name").Eq("Acme")).Delete(ctx)
	require.Error(t, err)
	assert.Equal(t, "filtering by related fields is not supported in delete", err.Error())
}
