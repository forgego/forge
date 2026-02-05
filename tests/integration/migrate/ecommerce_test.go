package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/tests/helpers"
	"github.com/forgego/forge/tests/testhelpers"
)

// TestFullEcommerceSchemaMigration tests creating all 15 ecommerce models in a single migration
func TestFullEcommerceSchemaMigration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	t.Logf("Connected to Postgres: %s", dsn)

	// Create a temporary directory for migrations under tests/tmp (returns relative path)
	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "ecommerce_full_")
	defer cleanupTemp()
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Path to ecommerce models
	modelsDir := filepath.Join("..", "..", "examples", "ecommerce", "models")

	// Generate migration from all ecommerce models
	err = helpers.CreateMigrationFromModels(t, modelsDir, migrationsDir, "initial_ecommerce_schema")
	require.NoError(t, err, "failed to generate migration from ecommerce models")

	// Create database connection using db package
	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	// Apply migration
	err = helpers.ApplyMigrationSequence(ctx, t, database, migrationsDir)
	require.NoError(t, err, "failed to apply ecommerce migration")

	// Expected tables from all 15 ecommerce models
	expectedTables := []string{
		"addresses",
		"brands",
		"categories",
		"customer_profiles",
		"customers",
		"inventory",
		"order_items",
		"orders",
		"payments",
		"product_variants",
		"products",
		"reviews",
		"shipping",
		"suppliers",
		"warehouses",
	}

	// Verify all tables exist
	for _, tableName := range expectedTables {
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", tableName)
	}

	// Verify key relationships exist
	// Customer -> Order (FK with PROTECT)
	helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "orders", "customer_id")
	helpers.AssertForeignKeyCascade(ctx, t, postgresDB, "orders", "customer_id", helpers.CascadeRESTRICT)

	// Product -> Brand (FK with SET_NULL)
	helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "products", "brand_id")
	helpers.AssertForeignKeyCascade(ctx, t, postgresDB, "products", "brand_id", helpers.CascadeSET_NULL)

	// Product -> Supplier (FK with SET_NULL)
	helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "products", "supplier_id")
	helpers.AssertForeignKeyCascade(ctx, t, postgresDB, "products", "supplier_id", helpers.CascadeSET_NULL)

	// Order -> OrderItem (FK with CASCADE)
	helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "order_items", "order_id")
	helpers.AssertForeignKeyCascade(ctx, t, postgresDB, "order_items", "order_id", helpers.CascadeCASCADE)

	// Product -> ProductVariant (FK with CASCADE)
	helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "product_variants", "product_id")
	helpers.AssertForeignKeyCascade(ctx, t, postgresDB, "product_variants", "product_id", helpers.CascadeCASCADE)

	// Customer -> CustomerProfile (OneToOne with CASCADE)
	helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "customer_profiles", "customer_id")

	// Verify PostgreSQL-specific data types
	// UUID columns
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "customers", "uuid")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "products", "sku")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "orders", "order_number")

	// JSONB columns
	helpers.AssertJSONBColumn(ctx, t, postgresDB, "products", "attributes")
	helpers.AssertJSONBColumn(ctx, t, postgresDB, "products", "images")
	helpers.AssertJSONBColumn(ctx, t, postgresDB, "products", "seo_data")
	helpers.AssertJSONBColumn(ctx, t, postgresDB, "customers", "metadata")
	helpers.AssertJSONBColumn(ctx, t, postgresDB, "orders", "metadata")

	// DECIMAL/NUMERIC columns
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "products", "price")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "orders", "total_amount")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "customers", "lifetime_value")

	// TIMESTAMP WITH TIME ZONE columns (DateTime fields)
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "products", "created_at")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "orders", "placed_at")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "customers", "last_login")

	// Verify indexes
	// Unique indexes
	helpers.AssertIndexExists(ctx, t, postgresDB, "postgres", "customers", "idx_customer_email")
	helpers.AssertIndexExists(ctx, t, postgresDB, "postgres", "customers", "idx_customer_uuid")
	helpers.AssertIndexExists(ctx, t, postgresDB, "postgres", "products", "idx_product_sku")
	helpers.AssertIndexExists(ctx, t, postgresDB, "postgres", "products", "idx_product_slug")
	helpers.AssertIndexExists(ctx, t, postgresDB, "postgres", "orders", "idx_order_order_number")

	// Non-unique indexes
	helpers.AssertIndexExists(ctx, t, postgresDB, "postgres", "products", "idx_product_status")
	helpers.AssertIndexExists(ctx, t, postgresDB, "postgres", "products", "idx_product_is_active")
	helpers.AssertIndexExists(ctx, t, postgresDB, "postgres", "orders", "idx_order_customer_id")
	helpers.AssertIndexExists(ctx, t, postgresDB, "postgres", "orders", "idx_order_status")

	// Verify GIN indexes for JSONB columns (if created)
	// Note: GIN indexes may need to be created manually or via migration
	// This test verifies the columns exist and are JSONB type

	// Verify composite unique constraints
	// Products have unique constraints on sku and slug (already verified above)

	// Test cascade behaviors across the model graph
	// Insert test data to verify cascade behaviors
	insertCustomerSQL := `
		INSERT INTO customers (uuid, email, first_name, last_name, password_hash, is_active)
		VALUES (gen_random_uuid(), 'test@example.com', 'Test', 'User', 'hash123', true)
		RETURNING id
	`
	var customerID int64
	err = postgresDB.QueryRowContext(ctx, insertCustomerSQL).Scan(&customerID)
	require.NoError(t, err)

	insertProductSQL := `
		INSERT INTO products (sku, name, slug, price, currency, is_active, status)
		VALUES (gen_random_uuid(), 'Test Product', 'test-product', 99.99, 'USD', true, 'active')
		RETURNING id
	`
	var productID int64
	err = postgresDB.QueryRowContext(ctx, insertProductSQL).Scan(&productID)
	require.NoError(t, err)

	// Get an address ID (create one if needed)
	insertAddressSQL := `
		INSERT INTO addresses (address_line1, city, state, postal_code, country, customer_id, type, first_name, last_name)
		VALUES ('123 Main St', 'City', 'State', '12345', 'US', $1, 'billing', 'John', 'Doe')
		RETURNING id
	`
	var addressID int64
	err = postgresDB.QueryRowContext(ctx, insertAddressSQL, customerID).Scan(&addressID)
	require.NoError(t, err)

	insertOrderSQL := `
		INSERT INTO orders (order_number, customer_id, subtotal, total_amount, currency, 
			billing_address_id, shipping_address_id, status, payment_status, shipping_status)
		VALUES (gen_random_uuid(), $1, 99.99, 99.99, 'USD', $2, $2, 'pending', 'pending', 'pending')
		RETURNING id
	`
	var orderID int64
	err = postgresDB.QueryRowContext(ctx, insertOrderSQL, customerID, addressID).Scan(&orderID)
	require.NoError(t, err)

	insertOrderItemSQL := `
		INSERT INTO order_items (order_id, product_id, product_name, product_sku, quantity, unit_price, total_price)
		VALUES ($1, $2, 'Test Product', 'TEST-SKU-001', 1, 99.99, 99.99)
		RETURNING id
	`
	var orderItemID int64
	err = postgresDB.QueryRowContext(ctx, insertOrderItemSQL, orderID, productID).Scan(&orderItemID)
	require.NoError(t, err)

	// Verify cascade: deleting order should cascade to order_items
	deleteOrderSQL := `DELETE FROM orders WHERE id = $1`
	_, err = postgresDB.ExecContext(ctx, deleteOrderSQL, orderID)
	require.NoError(t, err)

	// Verify order_item was deleted due to CASCADE
	var count int64
	err = postgresDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM order_items WHERE id = $1", orderItemID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, int64(0), count, "order_item should be deleted when order is deleted (CASCADE)")

	// Create a new order to test PROTECT constraint
	insertOrderSQL2 := `
		INSERT INTO orders (order_number, customer_id, subtotal, total_amount, currency, 
			billing_address_id, shipping_address_id, status, payment_status, shipping_status)
		VALUES (gen_random_uuid(), $1, 199.99, 199.99, 'USD', $2, $2, 'pending', 'pending', 'pending')
		RETURNING id
	`
	var orderID2 int64
	err = postgresDB.QueryRowContext(ctx, insertOrderSQL2, customerID, addressID).Scan(&orderID2)
	require.NoError(t, err)

	// Verify PROTECT: cannot delete customer with orders
	deleteCustomerSQL := `DELETE FROM customers WHERE id = $1`
	_, err = postgresDB.ExecContext(ctx, deleteCustomerSQL, customerID)
	require.Error(t, err, "should not be able to delete customer with orders (PROTECT)")

	// Verify SET_NULL: deleting brand should set brand_id to NULL
	insertBrandSQL := `
		INSERT INTO brands (name, slug)
		VALUES ('Test Brand', 'test-brand')
		RETURNING id
	`
	var brandID int64
	err = postgresDB.QueryRowContext(ctx, insertBrandSQL).Scan(&brandID)
	require.NoError(t, err)

	updateProductSQL := `UPDATE products SET brand_id = $1 WHERE id = $2`
	_, err = postgresDB.ExecContext(ctx, updateProductSQL, brandID, productID)
	require.NoError(t, err)

	deleteBrandSQL := `DELETE FROM brands WHERE id = $1`
	_, err = postgresDB.ExecContext(ctx, deleteBrandSQL, brandID)
	require.NoError(t, err)

	// Verify brand_id was set to NULL
	var brandIDAfterDelete sql.NullInt64
	err = postgresDB.QueryRowContext(ctx, "SELECT brand_id FROM products WHERE id = $1", productID).Scan(&brandIDAfterDelete)
	require.NoError(t, err)
	require.False(t, brandIDAfterDelete.Valid, "brand_id should be NULL after brand deletion (SET_NULL)")

	// Clean up remaining test data
	_, _ = postgresDB.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
	_, _ = postgresDB.ExecContext(ctx, "DELETE FROM addresses WHERE id = $1", addressID)
	// Customer deletion will fail due to PROTECT, so we'll leave it for now
}
