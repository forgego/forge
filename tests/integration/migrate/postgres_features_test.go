package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/tests/helpers"
	"github.com/forgego/forge/tests/testhelpers"
)

// TestGINIndexCreation tests creating GIN indexes for JSONB columns
func TestGINIndexCreation(t *testing.T) {
	// Using localhost PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Create table with JSONB column
	createTableSQL := `
		CREATE TABLE products (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			attributes JSONB
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Create GIN index on JSONB column
	createIndexSQL := `CREATE INDEX idx_products_attributes_gin ON products USING GIN (attributes)`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createIndexSQL)

	// Verify GIN index exists
	helpers.AssertGINIndexExists(ctx, t, postgresDB, "products", "idx_products_attributes_gin")
}

// TestGiSTIndexCreation tests creating GiST indexes
func TestGiSTIndexCreation(t *testing.T) {
	// Using localhost PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"documents"})

	// Create table with geometric or full-text search column
	// For this test, we'll use a text column with GiST (though typically used for geometric types)
	createTableSQL := `
		CREATE TABLE documents (
			id BIGSERIAL PRIMARY KEY,
			title VARCHAR(200) NOT NULL,
			content TEXT
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Create GiST index (typically used for geometric types, but can be used for full-text search)
	// Note: For full-text search, GIN is more common, but GiST can be used
	createIndexSQL := `CREATE INDEX idx_documents_content_gist ON documents USING GIST (to_tsvector('english', content))`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createIndexSQL)

	// Verify GiST index exists
	helpers.AssertGiSTIndexExists(ctx, t, postgresDB, "documents", "idx_documents_content_gist")
}

// TestJSONBColumnOperations tests JSONB column operations
func TestJSONBColumnOperations(t *testing.T) {
	// Using localhost PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"products"})

	// Create table with JSONB column
	createTableSQL := `
		CREATE TABLE products (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			attributes JSONB,
			metadata JSONB
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Verify JSONB columns exist
	helpers.AssertJSONBColumn(ctx, t, postgresDB, "products", "attributes")
	helpers.AssertJSONBColumn(ctx, t, postgresDB, "products", "metadata")

	// Insert JSONB data
	testData := map[string]interface{}{
		"color":  "red",
		"size":   "large",
		"weight": 1.5,
	}
	jsonBytes, err := json.Marshal(testData)
	require.NoError(t, err)

	insertSQL := `INSERT INTO products (name, attributes) VALUES ($1, $2) RETURNING id`
	var insertedID int64
	err = postgresDB.QueryRowContext(ctx, insertSQL, "Test Product", string(jsonBytes)).Scan(&insertedID)
	require.NoError(t, err)
	require.Greater(t, insertedID, int64(0), "inserted ID should be greater than 0")

	// Query JSONB data
	var retrievedJSON string
	err = postgresDB.QueryRowContext(ctx, `SELECT attributes FROM products WHERE id = $1`, insertedID).Scan(&retrievedJSON)
	require.NoError(t, err)

	var retrievedData map[string]interface{}
	err = json.Unmarshal([]byte(retrievedJSON), &retrievedData)
	require.NoError(t, err)
	require.Equal(t, "red", retrievedData["color"])
	require.Equal(t, "large", retrievedData["size"])

	// Test JSONB query operators
	var count int
	err = postgresDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM products WHERE attributes->>'color' = 'red'
	`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

// TestArrayColumnTypes tests PostgreSQL array column types
func TestArrayColumnTypes(t *testing.T) {
	// Using localhost PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"products"})

	// Create table with array columns
	createTableSQL := `
		CREATE TABLE products (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			tags TEXT[],
			category_ids INTEGER[],
			prices NUMERIC(10, 2)[]
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Verify array columns exist
	helpers.AssertArrayColumn(ctx, t, postgresDB, "products", "tags", "text")
	helpers.AssertArrayColumn(ctx, t, postgresDB, "products", "category_ids", "integer")
	helpers.AssertArrayColumn(ctx, t, postgresDB, "products", "prices", "numeric")

	// Insert array data using PostgreSQL array syntax
	// Use array literal syntax directly in SQL for proper type handling
	insertSQL := `
		INSERT INTO products (name, tags, category_ids, prices)
		VALUES ($1, ARRAY['electronics', 'gadgets', 'new']::text[], ARRAY[1, 2, 3]::integer[], ARRAY[99.99, 149.99, 199.99]::numeric[])
	`
	_, err = postgresDB.ExecContext(ctx, insertSQL, "Test Product")
	require.NoError(t, err)

	// Query array data - use pq.Array for proper scanning
	// Since we don't have pq imported, use a workaround: scan as string and parse
	var retrievedTagsStr string
	err = postgresDB.QueryRowContext(ctx, `SELECT tags::text FROM products WHERE name = $1`, "Test Product").Scan(&retrievedTagsStr)
	require.NoError(t, err)
	// Parse the array string format: {electronics,gadgets,new}
	require.Contains(t, retrievedTagsStr, "electronics")
	require.Contains(t, retrievedTagsStr, "gadgets")
	require.Contains(t, retrievedTagsStr, "new")
}

// TestCustomPostgreSQLTypes tests custom PostgreSQL types
func TestCustomPostgreSQLTypes(t *testing.T) {
	// Using localhost PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Create custom type
	createTypeSQL := `CREATE TYPE order_status AS ENUM ('pending', 'confirmed', 'shipped', 'delivered', 'cancelled')`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTypeSQL)

	// Verify custom type exists
	helpers.AssertCustomType(ctx, t, postgresDB, "order_status")

	// Create table using custom type
	createTableSQL := `
		CREATE TABLE orders (
			id BIGSERIAL PRIMARY KEY,
			order_number VARCHAR(50) NOT NULL,
			status order_status DEFAULT 'pending'
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Insert data using custom type
	insertSQL := `INSERT INTO orders (order_number, status) VALUES ($1, $2)`
	_, err = postgresDB.ExecContext(ctx, insertSQL, "ORD-001", "confirmed")
	require.NoError(t, err)

	// Query data
	var status string
	err = postgresDB.QueryRowContext(ctx, `SELECT status FROM orders WHERE id = 1`).Scan(&status)
	require.NoError(t, err)
	require.Equal(t, "confirmed", status)
}

// TestPartialIndex tests partial indexes with WHERE clauses
func TestPartialIndex(t *testing.T) {
	// Using localhost PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"products"})

	// Create table
	createTableSQL := `
		CREATE TABLE products (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			is_active BOOLEAN DEFAULT true,
			status VARCHAR(50)
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Create partial index (only for active products)
	createIndexSQL := `CREATE INDEX idx_products_active ON products(name) WHERE is_active = true`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createIndexSQL)

	// Verify partial index exists
	helpers.AssertPartialIndex(ctx, t, postgresDB, "products", "idx_products_active", "is_active = true")
}

// TestFunctionalIndex tests functional indexes
func TestFunctionalIndex(t *testing.T) {
	// Using localhost PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"users"})

	// Create table
	createTableSQL := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			email VARCHAR(254) NOT NULL,
			username VARCHAR(150) NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Create functional index (case-insensitive email search)
	createIndexSQL := `CREATE INDEX idx_users_email_lower ON users(LOWER(email))`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createIndexSQL)

	// Verify functional index exists
	helpers.AssertFunctionalIndex(ctx, t, postgresDB, "users", "idx_users_email_lower", "LOWER(email)")
}

// TestCoveringIndex tests covering indexes with INCLUDE columns
func TestCoveringIndex(t *testing.T) {
	// Using localhost PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"orders"})

	// Create table
	createTableSQL := `
		CREATE TABLE orders (
			id BIGSERIAL PRIMARY KEY,
			customer_id BIGINT NOT NULL,
			order_date TIMESTAMP DEFAULT NOW(),
			total_amount NUMERIC(12, 2)
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Create covering index with INCLUDE columns
	createIndexSQL := `CREATE INDEX idx_orders_customer_date ON orders(customer_id, order_date) INCLUDE (total_amount)`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createIndexSQL)

	// Verify covering index exists
	helpers.AssertCoveringIndex(ctx, t, postgresDB, "orders", "idx_orders_customer_date", []string{"total_amount"})
}

// TestUUIDType tests UUID column type
func TestUUIDType(t *testing.T) {
	// Using localhost PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Enable UUID extension
	enableExtensionSQL := `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, enableExtensionSQL)

	// Create table with UUID column
	createTableSQL := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			uuid UUID DEFAULT gen_random_uuid() NOT NULL,
			email VARCHAR(254) NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Verify UUID column exists
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "users", "uuid")

	// Insert data with UUID
	insertSQL := `INSERT INTO users (email) VALUES ($1)`
	_, err = postgresDB.ExecContext(ctx, insertSQL, "test@example.com")
	require.NoError(t, err)

	// Query UUID
	var uuid string
	err = postgresDB.QueryRowContext(ctx, `SELECT uuid FROM users WHERE id = 1`).Scan(&uuid)
	require.NoError(t, err)
	require.NotEmpty(t, uuid)
	require.Len(t, uuid, 36) // UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
}

// TestNumericPrecision tests NUMERIC type with precision and scale
func TestNumericPrecision(t *testing.T) {
	// Using localhost PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"products"})

	// Create table with NUMERIC columns
	createTableSQL := `
		CREATE TABLE products (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			price NUMERIC(12, 2) NOT NULL,
			discount NUMERIC(5, 2) DEFAULT 0.00
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Verify columns exist
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "products", "price")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "products", "discount")

	// Insert data with precise decimal values
	insertSQL := `INSERT INTO products (name, price, discount) VALUES ($1, $2, $3)`
	_, err = postgresDB.ExecContext(ctx, insertSQL, "Test Product", "99.99", "10.50")
	require.NoError(t, err)

	// Query and verify precision
	var price, discount float64
	err = postgresDB.QueryRowContext(ctx, `SELECT price, discount FROM products WHERE id = 1`).Scan(&price, &discount)
	require.NoError(t, err)
	require.Equal(t, 99.99, price)
	require.Equal(t, 10.50, discount)
}

// TestTimestampWithTimeZone tests TIMESTAMP WITH TIME ZONE
func TestTimestampWithTimeZone(t *testing.T) {
	// Using localhost PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"events"})

	// Create table with TIMESTAMP WITH TIME ZONE
	createTableSQL := `
		CREATE TABLE events (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Verify columns exist
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "events", "created_at")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "events", "updated_at")

	// Insert data
	insertSQL := `INSERT INTO events (name) VALUES ($1)`
	_, err = postgresDB.ExecContext(ctx, insertSQL, "Test Event")
	require.NoError(t, err)

	// Query timestamp
	var createdAt time.Time
	err = postgresDB.QueryRowContext(ctx, `SELECT created_at FROM events WHERE id = 1`).Scan(&createdAt)
	require.NoError(t, err)
	require.False(t, createdAt.IsZero())
}
