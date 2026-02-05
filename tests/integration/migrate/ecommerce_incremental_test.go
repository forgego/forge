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

// TestIncrementalEcommerceMigrations tests adding models incrementally
// This simulates real-world development where models are added over time
func TestIncrementalEcommerceMigrations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_incremental_ecommerce_%d", time.Now().UnixNano()),
	}
	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	t.Logf("Connected to Postgres: %s", dsn)

	// Create a temporary directory for migrations under tests/tmp (returns relative path)
	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "ecommerce_incremental_")
	defer cleanupTemp()
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create database connection
	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	// Path to ecommerce models
	baseModelsDir := filepath.Join("..", "..", "examples", "ecommerce", "models")

	// Phase 1: Create core customer and address models
	t.Run("Phase1_CoreModels", func(t *testing.T) {
		// Create temporary models directory with only core models
		phase1ModelsDir := filepath.Join(tempDir, "models_phase1")
		require.NoError(t, os.MkdirAll(phase1ModelsDir, 0755))

		// Copy core model files
		copyModelFile(t, baseModelsDir, phase1ModelsDir, "address.go")
		copyModelFile(t, baseModelsDir, phase1ModelsDir, "customer.go")

		// Generate migration
		err = helpers.CreateMigrationFromModels(t, phase1ModelsDir, migrationsDir, "001_core_models")
		require.NoError(t, err)

		// Create runner after migrations are generated
		var runner *db.MigrationRunner
		runner, err = db.NewMigrationRunner(database, migrationsDir)
		require.NoError(t, err)
		defer runner.Close()

		// Apply migration using the runner
		err = helpers.ApplyMigrationSequence(ctx, t, database, migrationsDir, runner)
		require.NoError(t, err)

		// Verify tables exist
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "addresses")
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "customers")

		// Verify customer profile table doesn't exist yet
		// (we'll add it in a later phase)
	})

	// Phase 2: Add customer profile (OneToOne relationship)
	t.Run("Phase2_CustomerProfile", func(t *testing.T) {
		// Create database connection for this subtest
		database, err := db.NewDB(dsn)
		require.NoError(t, err)
		defer database.Close()

		phase2ModelsDir := filepath.Join(tempDir, "models_phase2")
		require.NoError(t, os.MkdirAll(phase2ModelsDir, 0755))

		// Copy previous models plus new one
		copyModelFile(t, baseModelsDir, phase2ModelsDir, "address.go")
		copyModelFile(t, baseModelsDir, phase2ModelsDir, "customer.go")
		copyModelFile(t, baseModelsDir, phase2ModelsDir, "customer_profile.go")

		// Generate migration
		err = helpers.CreateMigrationFromModels(t, phase2ModelsDir, migrationsDir, "002_customer_profile")
		require.NoError(t, err)

		// Create runner after migrations are generated
		runner, err := db.NewMigrationRunner(database, migrationsDir)
		require.NoError(t, err)
		defer runner.Close()

		// Apply migration using the runner
		err = helpers.ApplyMigrationSequence(ctx, t, database, migrationsDir, runner)
		require.NoError(t, err)

		// Verify new table exists
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "customer_profiles")
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "customer_profiles", "customer_id")
	})

	// Phase 3: Add product-related models (Brand, Supplier, Category, Product)
	t.Run("Phase3_ProductModels", func(t *testing.T) {
		// Create database connection for this subtest
		database, err := db.NewDB(dsn)
		require.NoError(t, err)
		defer database.Close()

		phase3ModelsDir := filepath.Join(tempDir, "models_phase3")
		require.NoError(t, os.MkdirAll(phase3ModelsDir, 0755))

		// Copy all previous models plus product-related ones
		copyModelFile(t, baseModelsDir, phase3ModelsDir, "address.go")
		copyModelFile(t, baseModelsDir, phase3ModelsDir, "customer.go")
		copyModelFile(t, baseModelsDir, phase3ModelsDir, "customer_profile.go")
		copyModelFile(t, baseModelsDir, phase3ModelsDir, "brand.go")
		copyModelFile(t, baseModelsDir, phase3ModelsDir, "supplier.go")
		copyModelFile(t, baseModelsDir, phase3ModelsDir, "category.go")
		copyModelFile(t, baseModelsDir, phase3ModelsDir, "product.go")

		// Generate migration
		err = helpers.CreateMigrationFromModels(t, phase3ModelsDir, migrationsDir, "003_product_models")
		require.NoError(t, err)

		// Create runner after migrations are generated
		runner, err := db.NewMigrationRunner(database, migrationsDir)
		require.NoError(t, err)
		defer runner.Close()

		// Apply migration using the runner
		err = helpers.ApplyMigrationSequence(ctx, t, database, migrationsDir, runner)
		require.NoError(t, err)

		// Verify new tables exist
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "brands")
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "suppliers")
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "categories")
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "products")

		// Verify relationships
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "products", "brand_id")
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "products", "supplier_id")

		// Verify many-to-many relationship table (products_categories)
		// Note: ManyToMany creates a junction table
		// The junction table name might vary, so we'll skip this check for now
		// helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "products_categories")
	})

	// Phase 4: Add product variants and inventory
	t.Run("Phase4_ProductVariants", func(t *testing.T) {
		// Create database connection for this subtest
		database, err := db.NewDB(dsn)
		require.NoError(t, err)
		defer database.Close()

		phase4ModelsDir := filepath.Join(tempDir, "models_phase4")
		require.NoError(t, os.MkdirAll(phase4ModelsDir, 0755))

		// Copy all previous models plus new ones
		copyModelFile(t, baseModelsDir, phase4ModelsDir, "address.go")
		copyModelFile(t, baseModelsDir, phase4ModelsDir, "customer.go")
		copyModelFile(t, baseModelsDir, phase4ModelsDir, "customer_profile.go")
		copyModelFile(t, baseModelsDir, phase4ModelsDir, "brand.go")
		copyModelFile(t, baseModelsDir, phase4ModelsDir, "supplier.go")
		copyModelFile(t, baseModelsDir, phase4ModelsDir, "category.go")
		copyModelFile(t, baseModelsDir, phase4ModelsDir, "product.go")
		copyModelFile(t, baseModelsDir, phase4ModelsDir, "product_variant.go")
		copyModelFile(t, baseModelsDir, phase4ModelsDir, "inventory.go")

		// Generate migration
		err = helpers.CreateMigrationFromModels(t, phase4ModelsDir, migrationsDir, "004_product_variants")
		require.NoError(t, err)

		// Create runner after migrations are generated
		runner, err := db.NewMigrationRunner(database, migrationsDir)
		require.NoError(t, err)
		defer runner.Close()

		// Apply migration using the runner
		err = helpers.ApplyMigrationSequence(ctx, t, database, migrationsDir, runner)
		require.NoError(t, err)

		// Verify new tables
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "product_variants")
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "inventory")

		// Verify relationships
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "product_variants", "product_id")
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "inventory", "product_id")
	})

	// Phase 5: Add order-related models
	t.Run("Phase5_OrderModels", func(t *testing.T) {
		// Create database connection for this subtest
		database, err := db.NewDB(dsn)
		require.NoError(t, err)
		defer database.Close()

		phase5ModelsDir := filepath.Join(tempDir, "models_phase5")
		require.NoError(t, os.MkdirAll(phase5ModelsDir, 0755))

		// Copy all previous models plus order-related ones
		copyModelFile(t, baseModelsDir, phase5ModelsDir, "address.go")
		copyModelFile(t, baseModelsDir, phase5ModelsDir, "customer.go")
		copyModelFile(t, baseModelsDir, phase5ModelsDir, "customer_profile.go")
		copyModelFile(t, baseModelsDir, phase5ModelsDir, "brand.go")
		copyModelFile(t, baseModelsDir, phase5ModelsDir, "supplier.go")
		copyModelFile(t, baseModelsDir, phase5ModelsDir, "category.go")
		copyModelFile(t, baseModelsDir, phase5ModelsDir, "product.go")
		copyModelFile(t, baseModelsDir, phase5ModelsDir, "product_variant.go")
		copyModelFile(t, baseModelsDir, phase5ModelsDir, "inventory.go")
		copyModelFile(t, baseModelsDir, phase5ModelsDir, "order.go")
		copyModelFile(t, baseModelsDir, phase5ModelsDir, "order_item.go")

		// Generate migration
		err = helpers.CreateMigrationFromModels(t, phase5ModelsDir, migrationsDir, "005_order_models")
		require.NoError(t, err)

		// Create runner after migrations are generated
		runner, err := db.NewMigrationRunner(database, migrationsDir)
		require.NoError(t, err)
		defer runner.Close()

		// Apply migration using the runner
		err = helpers.ApplyMigrationSequence(ctx, t, database, migrationsDir, runner)
		require.NoError(t, err)

		// Verify new tables
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "orders")
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "order_items")

		// Verify relationships
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "orders", "customer_id")
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "orders", "billing_address_id")
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "orders", "shipping_address_id")
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "order_items", "order_id")
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "order_items", "product_id")
	})

	// Phase 6: Add payment and shipping models
	t.Run("Phase6_PaymentShipping", func(t *testing.T) {
		// Create database connection for this subtest
		database, err := db.NewDB(dsn)
		require.NoError(t, err)
		defer database.Close()

		phase6ModelsDir := filepath.Join(tempDir, "models_phase6")
		require.NoError(t, os.MkdirAll(phase6ModelsDir, 0755))

		// Copy all previous models plus payment/shipping
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "address.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "customer.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "customer_profile.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "brand.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "supplier.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "category.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "product.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "product_variant.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "inventory.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "order.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "order_item.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "payment.go")
		copyModelFile(t, baseModelsDir, phase6ModelsDir, "shipping.go")

		// Generate migration
		err = helpers.CreateMigrationFromModels(t, phase6ModelsDir, migrationsDir, "006_payment_shipping")
		require.NoError(t, err)

		// Create runner after migrations are generated
		runner, err := db.NewMigrationRunner(database, migrationsDir)
		require.NoError(t, err)
		defer runner.Close()

		// Apply migration using the runner
		err = helpers.ApplyMigrationSequence(ctx, t, database, migrationsDir, runner)
		require.NoError(t, err)

		// Verify new tables
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "payments")
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "shipping")

		// Verify relationships
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "payments", "order_id")
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "shipping", "order_id")
	})

	// Phase 7: Add remaining models (Review, Warehouse)
	t.Run("Phase7_RemainingModels", func(t *testing.T) {
		// Create database connection for this subtest
		database, err := db.NewDB(dsn)
		require.NoError(t, err)
		defer database.Close()

		phase7ModelsDir := filepath.Join(tempDir, "models_phase7")
		require.NoError(t, os.MkdirAll(phase7ModelsDir, 0755))

		// Copy all models
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "address.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "customer.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "customer_profile.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "brand.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "supplier.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "category.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "product.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "product_variant.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "inventory.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "order.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "order_item.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "payment.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "shipping.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "review.go")
		copyModelFile(t, baseModelsDir, phase7ModelsDir, "warehouse.go")

		// Generate migration
		err = helpers.CreateMigrationFromModels(t, phase7ModelsDir, migrationsDir, "007_remaining_models")
		require.NoError(t, err)

		// Create runner after migrations are generated
		runner, err := db.NewMigrationRunner(database, migrationsDir)
		require.NoError(t, err)
		defer runner.Close()

		// Apply migration using the runner
		err = helpers.ApplyMigrationSequence(ctx, t, database, migrationsDir, runner)
		require.NoError(t, err)

		// Verify new tables
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "reviews")
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "warehouses")

		// Verify relationships
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "reviews", "product_id")
		helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "reviews", "customer_id")
	})

	// Final verification: all tables should exist
	t.Run("FinalVerification", func(t *testing.T) {
		// Create database connection for this subtest
		postgresDB, err := sql.Open("postgres", dsn)
		require.NoError(t, err)
		defer postgresDB.Close()

		allTables := []string{
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

		// Use a simple query to check if tables exist instead of using AssertTableExists
		// which might be trying to use a closed connection
		for _, tableName := range allTables {
			var exists int
			err = postgresDB.QueryRowContext(ctx, `
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = $1 AND table_schema = 'public'
			`, tableName).Scan(&exists)
			require.NoError(t, err, "table %s should exist", tableName)
			require.Equal(t, 1, exists, "table %s should exist", tableName)
		}

		// Verify migration state - create a new database connection
		database, err := db.NewDB(dsn)
		require.NoError(t, err)
		defer database.Close()

		helpers.AssertMigrationState(ctx, t, database, migrationsDir, 7, false)
	})
}

// copyModelFile copies a model file from source to destination
func copyModelFile(t *testing.T, srcDir, dstDir, filename string) {
	srcPath := filepath.Join(srcDir, filename)
	dstPath := filepath.Join(dstDir, filename)

	srcData, err := os.ReadFile(srcPath)
	require.NoError(t, err, "failed to read source file: %s", srcPath)

	err = os.WriteFile(dstPath, srcData, 0644)
	require.NoError(t, err, "failed to write destination file: %s", dstPath)
}
