package orm

import (
	"context"
	"path/filepath"
	"strings"
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
	Credit    float64      `db:"credit"`
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
		schema.Float64Field("credit"),
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
			credit REAL NOT NULL DEFAULT 0,
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
	assert.Equal(t, `SELECT "orders".* FROM "orders" LEFT JOIN "customers" AS "customer" ON "customer"."id" = "orders"."customer_id" LEFT JOIN "companies" AS "customer__company" ON "customer__company"."id" = "customer"."company_id" WHERE "customer__company"."name" = ?`, sql)

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
	const expectedSQL = `SELECT * FROM "orders" WHERE "total" = ?`
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

func setupToManyDedupeDB(t *testing.T) *db.DB {
	t.Helper()
	database := setupRelationTestDB(t)

	// Companies: 1: Acme Corp, 2: Beta LLC
	_, err := database.Exec(`INSERT INTO companies (id, name) VALUES (1, 'Acme Corp'), (2, 'Beta LLC')`)
	require.NoError(t, err)

	// Customers:
	// Customer 1: Acme, Company 1 (Acme Corp)
	// Customer 2: Bob, Company 2 (Beta LLC)
	// Customer 3: Charlie, Company 1 (Acme Corp)
	_, err = database.Exec(`INSERT INTO customers (id, name, company_id) VALUES (1, 'Acme', 1), (2, 'Bob', 2), (3, 'Charlie', 1)`)
	require.NoError(t, err)

	// Orders:
	// Customer 1 (Acme): 3 orders over 100
	// Customer 2 (Bob): 1 order over 100
	// Customer 3 (Charlie): 1 order under 100
	_, err = database.Exec(`INSERT INTO orders (id, total, customer_id) VALUES
		(1, 150.0, 1),
		(2, 200.0, 1),
		(3, 250.0, 1),
		(4, 300.0, 2),
		(5, 50.0, 3)`)
	require.NoError(t, err)

	return database
}

func TestRelationPathJoin_ToManyFilter_DeduplicatesParentRows(t *testing.T) {
	database := setupToManyDedupeDB(t)
	defer database.Close()

	ctx := context.Background()
	customerQS, err := NewQuerySet[TestCustomer]("customers")
	require.NoError(t, err)
	customerQS = customerQS.SetDB(database)

	// Customer with 3 orders over 100 (Acme) and one with a single order over 100 (Bob):
	// Filter(F("orders__total").Gt(100)).All returns 2 customers, each once; Count returns 2.
	filtered := customerQS.Filter(F("orders__total").Gt(100.0))

	customers, err := filtered.All(ctx)
	require.NoError(t, err)
	require.Len(t, customers, 2)
	assert.Equal(t, int64(1), customers[0].ID)
	assert.Equal(t, "Acme", customers[0].Name)
	assert.Equal(t, int64(2), customers[1].ID)
	assert.Equal(t, "Bob", customers[1].Name)

	count, err := filtered.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRelationPathJoin_ToManyFilter_Pagination(t *testing.T) {
	database := setupToManyDedupeDB(t)
	defer database.Close()

	ctx := context.Background()
	customerQS, err := NewQuerySet[TestCustomer]("customers")
	require.NoError(t, err)
	customerQS = customerQS.SetDB(database)

	// Same filter plus OrderBy("name"), Limit(1), Offset(1):
	// returns exactly the second customer by name (pagination is over distinct parents).
	customers, err := customerQS.
		Filter(F("orders__total").Gt(100.0)).
		OrderBy("name").
		Limit(1).
		Offset(1).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, customers, 1)
	assert.Equal(t, int64(2), customers[0].ID)
	assert.Equal(t, "Bob", customers[0].Name)
}

func TestRelationPathJoin_ToManyFilter_CombinedWithToOneOrderBy(t *testing.T) {
	database := setupToManyDedupeDB(t)
	defer database.Close()

	ctx := context.Background()
	customerQS, err := NewQuerySet[TestCustomer]("customers")
	require.NoError(t, err)
	customerQS = customerQS.SetDB(database)

	// To-many filter combined with a to-one ordering path (OrderBy("company__name"))
	// runs without SQL error and returns distinct rows in company-name order.
	customers, err := customerQS.
		Filter(F("orders__total").Gt(100.0)).
		OrderBy("company__name").
		All(ctx)
	require.NoError(t, err)
	require.Len(t, customers, 2)
	assert.Equal(t, "Acme", customers[0].Name)
	assert.Equal(t, "Bob", customers[1].Name)

	// Descending order: Beta LLC > Acme Corp
	customersDesc, err := customerQS.
		Filter(F("orders__total").Gt(100.0)).
		OrderBy("-company__name").
		All(ctx)
	require.NoError(t, err)
	require.Len(t, customersDesc, 2)
	assert.Equal(t, "Bob", customersDesc[0].Name)
	assert.Equal(t, "Acme", customersDesc[1].Name)

	// Assert SQL structure
	sql, _, err := customerQS.
		Filter(F("orders__total").Gt(100.0)).
		OrderBy("company__name").(*BaseQuerySet[TestCustomer]).buildSQL()
	require.NoError(t, err)
	assert.Equal(t, `SELECT "customers".* FROM "customers" LEFT JOIN "companies" AS "company" ON "company"."id" = "customers"."company_id" WHERE "customers"."id" IN (SELECT "customers"."id" FROM "customers" LEFT JOIN "orders" AS "orders" ON "orders"."customer_id" = "customers"."id" WHERE "orders"."total" > ?) ORDER BY "company"."name" ASC`, sql)
}

func TestRelationPathJoin_AnnotationOnRelatedField_RegistersJoin(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()
	seedRelationTestData(t, database)

	qs, err := NewQuerySet[TestOrder]("orders")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	nameField := NewField[string]("customer__name", "orders")
	ann := NewExpressionAnnotation("customer_name", nameField)
	qs = qs.Annotate(ann)

	base := qs.(*BaseQuerySet[TestOrder])
	sql, args, err := base.buildSQL()
	require.NoError(t, err)
	assert.Contains(t, sql, `LEFT JOIN "customers" AS "customer" ON "customer"."id" = "orders"."customer_id"`)

	rows, err := database.Query(sql, args...)
	require.NoError(t, err)
	defer rows.Close()

	type result struct {
		id           int64
		total        float64
		customerID   int64
		customerName string
	}
	var results []result
	for rows.Next() {
		var r result
		err := rows.Scan(&r.id, &r.total, &r.customerID, &r.customerName)
		require.NoError(t, err)
		results = append(results, r)
	}
	require.NoError(t, rows.Err())
	require.Len(t, results, 3)
	assert.Equal(t, "Acme", results[0].customerName)
	assert.Equal(t, "Acme", results[1].customerName)
	assert.Equal(t, "Bob", results[2].customerName)

	// Verify no duplicate joins when both annotation and filter reference customer__name
	qs2, err := NewQuerySet[TestOrder]("orders")
	require.NoError(t, err)
	qs2 = qs2.SetDB(database)
	qs2 = qs2.Annotate(NewExpressionAnnotation("customer_name", NewField[string]("customer__name", "orders")))
	qs2 = qs2.Filter(F("customer__name").Eq("Acme"))
	sql2, _, err := qs2.(*BaseQuerySet[TestOrder]).buildSQL()
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(sql2, `LEFT JOIN "customers"`))
}
