package orm

import (
	"context"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/internal/testutils"
	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModel is a test model for integration tests
type TestModel struct {
	schema.BaseSchema
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Price     float64   `db:"price"`
	Available bool      `db:"available"`
	CreatedAt time.Time `db:"created_at"`
}

func (TestModel) Meta() schema.Meta {
	return schema.Meta{
		TableName: "test_models",
	}
}

func (TestModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.StringField("email", schema.Optional()),
		schema.Float64Field("price", schema.Default(0.0)),
		schema.BoolField("available", schema.Default(true)),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func TestQuerySet_Integration_Filter(t *testing.T) {
	sqlDB := testutils.SetupTestDB(t)
	testutils.CreateTestTable(t, sqlDB)
	defer sqlDB.Close()
	
	database := &db.DB{DB: sqlDB, Driver: "postgres"}

	// Insert test data
	_, err := database.Exec(`
		INSERT INTO test_models (name, email, price, available) VALUES
		('Product 1', 'test1@example.com', 10.0, true),
		('Product 2', 'test2@example.com', 20.0, true),
		('Product 3', 'test3@example.com', 30.0, false)
	`)
	require.NoError(t, err)

	qs, err := NewQuerySet[TestModel]("test_models")
	if err != nil {
		t.Fatalf("Failed to create QuerySet: %v", err)
	}

	// Set DB connection
	qs = qs.SetDB(database)

	priceField := NewField[float64]("price", "test_models")
	expr := priceField.Gt(15.0)

	filtered := qs.Filter(expr)
	assert.NotNil(t, filtered)

	// Verify filtering worked by fetching data
	results, err := filtered.All(context.Background())
	require.NoError(t, err)
	assert.Len(t, results, 2) // Product 2 (20.0) and Product 3 (30.0)
	
	for _, p := range results {
		assert.Greater(t, p.Price, 15.0)
	}
}

func TestQuerySet_Integration_OrderBy(t *testing.T) {
	sqlDB := testutils.SetupTestDB(t)
	testutils.CreateTestTable(t, sqlDB)
	defer sqlDB.Close()
	
	database := &db.DB{DB: sqlDB, Driver: "postgres"}
	
	_, err := database.Exec(`
		INSERT INTO test_models (name, price, email) VALUES
		('A', 10.0, ''), ('B', 30.0, ''), ('C', 20.0, '')
	`)
	require.NoError(t, err)

	qs, err := NewQuerySet[TestModel]("test_models")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	ordered := qs.OrderBy(Desc("price"))
	results, err := ordered.All(context.Background())
	require.NoError(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, "B", results[0].Name)
	assert.Equal(t, "C", results[1].Name)
	assert.Equal(t, "A", results[2].Name)
}

func TestQuerySet_Integration_LimitOffset(t *testing.T) {
	sqlDB := testutils.SetupTestDB(t)
	testutils.CreateTestTable(t, sqlDB)
	defer sqlDB.Close()
	
	database := &db.DB{DB: sqlDB, Driver: "postgres"}

	// Insert 15 records
	for i := 1; i <= 15; i++ {
		_, err := database.Exec(`INSERT INTO test_models (name, price, email) VALUES ($1, $2, $3)`, 
			"Product", float64(i), "")
		require.NoError(t, err)
	}

	qs, err := NewQuerySet[TestModel]("test_models")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	limited := qs.OrderBy(Asc("price")).Limit(5).Offset(5)
	results, err := limited.All(context.Background())
	require.NoError(t, err)
	
	assert.Len(t, results, 5)
	// Should be 6, 7, 8, 9, 10
	assert.Equal(t, 6.0, results[0].Price)
	assert.Equal(t, 10.0, results[4].Price)
}

func TestQuerySet_Integration_ComplexQuery(t *testing.T) {
	sqlDB := testutils.SetupTestDB(t)
	testutils.CreateTestTable(t, sqlDB)
	defer sqlDB.Close()
	
	database := &db.DB{DB: sqlDB, Driver: "postgres"}

	_, err := database.Exec(`
		INSERT INTO test_models (name, price, available, email) VALUES
		('P1', 100.0, true, ''),
		('P2', 50.0, true, ''),
		('P3', 200.0, false, ''),
		('P4', 150.0, true, '')
	`)
	require.NoError(t, err)

	qs, err := NewQuerySet[TestModel]("test_models")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	priceField := NewField[float64]("price", "test_models")
	availableField := NewField[bool]("available", "test_models")

	// Complex query: price > 60 AND available = true, ordered by price DESC
	complex := qs.
		Filter(priceField.Gt(60.0)).
		Filter(availableField.Eq(true)).
		OrderBy(Desc("price"))

	results, err := complex.All(context.Background())
	require.NoError(t, err)
	
	assert.Len(t, results, 2) // P4 (150), P1 (100)
	assert.Equal(t, "P4", results[0].Name)
	assert.Equal(t, "P1", results[1].Name)
}

func TestUpdateBuilder_Integration(t *testing.T) {
	sqlDB := testutils.SetupTestDB(t)
	testutils.CreateTestTable(t, sqlDB)
	defer sqlDB.Close()
	
	database := &db.DB{DB: sqlDB, Driver: "postgres"}

	_, err := database.Exec(`INSERT INTO test_models (id, name, price, email) VALUES (1, 'Old Name', 10.0, '')`)
	require.NoError(t, err)

	qs, err := NewQuerySet[TestModel]("test_models")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	ub, err := NewUpdateBuilder[TestModel](qs)
	require.NoError(t, err)

	ub = ub.
		Set("name", "Updated Name").
		Set("price", 99.99).
		Increment("id", int64(1)) // id becomes 2

	// Need to handle WHERE clause to update specific row, defaulting to all if not filtered
	// But UpdateBuilder usually applies to the queryset's filter. 
	// Here qs is unfiltered, so it updates all.
	
	rows, err := ub.Execute(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	// Verify update
	var name string
	var price float64
	var id int64
	err = database.QueryRow("SELECT id, name, price FROM test_models").Scan(&id, &name, &price)
	require.NoError(t, err)
	
	assert.Equal(t, "Updated Name", name)
	assert.Equal(t, 99.99, price)
	assert.Equal(t, int64(2), id)
}

// Keep existing unit tests that don't need DB or adapt them if needed
func TestExpression_Integration_StringOperations(t *testing.T) {
	// ... (Same as before, no DB needed for SQL generation tests)
	nameField := NewField[string]("name", "test_models")

	// Test Contains
	containsExpr := nameField.Contains("Product")
	builder := NewSQLBuilder()
	sql, _, err := containsExpr.ToSQL(builder)
	require.NoError(t, err)
	assert.Contains(t, sql, "LIKE")
	assert.Len(t, builder.Args(), 1)

	// Test StartsWith
	startsWithExpr := nameField.StartsWith("Product")
	builder = NewSQLBuilder()
	sql, _, err = startsWithExpr.ToSQL(builder)
	require.NoError(t, err)
	assert.Contains(t, sql, "LIKE")
	assert.Len(t, builder.Args(), 1)

	// Test EndsWith
	endsWithExpr := nameField.EndsWith("1")
	builder = NewSQLBuilder()
	sql, _, err = endsWithExpr.ToSQL(builder)
	require.NoError(t, err)
	assert.Contains(t, sql, "LIKE")
	assert.Len(t, builder.Args(), 1)
}

func TestQ_Integration_ComplexNesting(t *testing.T) {
    // ... (Same as before)
	priceField := NewField[float64]("price", "test_models")
	availableField := NewField[bool]("available", "test_models")
	nameField := NewField[string]("name", "test_models")

	// Complex: (price > 20 OR price < 10) AND (available = true OR name LIKE '%Product%')
	q1 := NewQ(priceField.Gt(20.0)).Or(NewQ(priceField.Lt(10.0)))
	q2 := NewQ(availableField.Eq(true)).Or(NewQ(nameField.Contains("Product")))
	combined := q1.And(q2)

	builder := NewSQLBuilder()
	sql, _, err := combined.ToSQL(builder)
	require.NoError(t, err)

	assert.Contains(t, sql, "OR")
	assert.Contains(t, sql, "AND")
	assert.GreaterOrEqual(t, len(builder.Args()), 4)
}

func TestSQLBuilder_Integration_ComplexQuery(t *testing.T) {
    // ... (Same as before)
	builder := NewSQLBuilder()

	// Build a complex SELECT query
	selectClause := builder.BuildSelect("test_models", []string{"id", "name", "price"}, false)
	whereClause, whereArgs := builder.BuildWhere(
		[]QueryExpr{NewFieldQueryExpr("price", OpGreater, 10.0)},
		[]QueryExpr{},
	)
	orderByClause := builder.BuildOrderBy([]string{"-price", "name"})
	limitClause := builder.BuildLimit(intPtr(10))
	offsetClause := builder.BuildOffset(intPtr(0))

	sql := selectClause + " " + whereClause + " " + orderByClause + " " + limitClause + " " + offsetClause

	assert.Contains(t, sql, "SELECT")
	assert.Contains(t, sql, "FROM")
	assert.Contains(t, sql, "WHERE")
	assert.Contains(t, sql, "ORDER BY")
	assert.Contains(t, sql, "LIMIT")
	assert.Contains(t, sql, "OFFSET")
	assert.Len(t, whereArgs, 1)
}

func intPtr(i int) *int {
	return &i
}
