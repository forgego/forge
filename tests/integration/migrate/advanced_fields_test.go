package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/migrate"
	"github.com/forgego/forge/tests/helpers"
	"github.com/forgego/forge/tests/testhelpers"
)

// TestGeneratedColumns tests that generated columns are properly created in migrations
func TestGeneratedColumns(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "127.0.0.1",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_generated_cols_%d", time.Now().UnixNano()),
	}
	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "generated_cols_")
	defer cleanupTemp()
	modelsDir := filepath.Join(tempDir, "models")
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create model with generated column
	modelContent := `package models

import "github.com/forgego/forge/schema"

type Product struct {
	schema.BaseSchema
}

func (Product) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(200).Build(),
		schema.Decimal("price").Required().MaxDigits(10).DecimalPlaces(2).Build(),
		schema.Decimal("tax_rate").Required().MaxDigits(5).DecimalPlaces(4).Build(),
		schema.Decimal("price_with_tax").
			GeneratedColumn("price * (1 + tax_rate)", true).
			MaxDigits(10).
			DecimalPlaces(2).
			Build(),
	}
}

func (Product) Meta() schema.Meta {
	return schema.Meta{
		TableName: "products",
	}
}

func (Product) Relations() []schema.Relation {
	return []schema.Relation{}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "product.go"), []byte(modelContent), 0644))

	// Generate and apply migration
	gen, err := migrate.NewGenerator(modelsDir, migrationsDir)
	require.NoError(t, err)
	err = gen.GenerateMigrations("add_generated_column")
	require.NoError(t, err)

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Verify generated column exists and works
	helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "products")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "products", "price_with_tax")

	// Insert test data and verify generated column calculation
	_, err = postgresDB.ExecContext(ctx, `
		INSERT INTO products (name, price, tax_rate) 
		VALUES ('Test Product', 100.00, 0.0825)
	`)
	require.NoError(t, err)

	var priceWithTax float64
	err = postgresDB.QueryRowContext(ctx, `
		SELECT price_with_tax FROM products WHERE name = 'Test Product'
	`).Scan(&priceWithTax)
	require.NoError(t, err)
	require.InDelta(t, 108.25, priceWithTax, 0.01, "generated column should calculate correctly")
}

// TestFieldWithCustomDBOptions tests fields with custom DB options
func TestFieldWithCustomDBOptions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "127.0.0.1",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_db_options_%d", time.Now().UnixNano()),
	}
	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "db_options_")
	defer cleanupTemp()
	modelsDir := filepath.Join(tempDir, "models")
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create model with various DB options
	modelContent := `package models

import "github.com/forgego/forge/schema"

type Article struct {
	schema.BaseSchema
}

func (Article) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("title").
			Required().
			MaxLength(200).
			DBColumn("article_title").
			DBComment("The title of the article").
			Build(),
		schema.String("slug").
			Required().
			MaxLength(200).
			Unique().
			DBIndex().
			DBComment("URL-friendly slug").
			Build(),
		schema.Text("content").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
	}
}

func (Article) Meta() schema.Meta {
	return schema.Meta{
		TableName: "articles",
	}
}

func (Article) Relations() []schema.Relation {
	return []schema.Relation{}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "article.go"), []byte(modelContent), 0644))

	// Generate and apply migration
	gen, err := migrate.NewGenerator(modelsDir, migrationsDir)
	require.NoError(t, err)
	err = gen.GenerateMigrations("create_articles")
	require.NoError(t, err)

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Verify custom column name
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "articles", "article_title")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "articles", "slug")

	// Verify index on slug exists
	var indexExists bool
	err = postgresDB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE tablename = 'articles' 
			AND indexname LIKE '%slug%'
		)
	`).Scan(&indexExists)
	require.NoError(t, err)
	require.True(t, indexExists, "index on slug should exist")

	// Verify column comment (PostgreSQL specific)
	var comment string
	err = postgresDB.QueryRowContext(ctx, `
		SELECT col_description(
			(SELECT oid FROM pg_class WHERE relname = 'articles'), 
			(SELECT ordinal_position FROM information_schema.columns 
			 WHERE table_name = 'articles' AND column_name = 'article_title')
		)
	`).Scan(&comment)
	if err == nil {
		require.Contains(t, comment, "title", "column comment should be set")
	}
}

// TestConstraintOptions tests unique constraints and check constraints
func TestConstraintOptions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "127.0.0.1",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_constraints_%d", time.Now().UnixNano()),
	}
	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "constraints_")
	defer cleanupTemp()
	modelsDir := filepath.Join(tempDir, "models")
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create model with various constraints
	modelContent := `package models

import "github.com/forgego/forge/schema"

type Employee struct {
	schema.BaseSchema
}

func (Employee) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("email").Required().MaxLength(255).Unique().Build(),
		schema.String("employee_id").Required().MaxLength(50).Unique().Build(),
		schema.Int32("age").MinValue(18).MaxValue(120).Build(),
		schema.Decimal("salary").MinValue(0).MaxDigits(12).DecimalPlaces(2).Build(),
	}
}

func (Employee) Meta() schema.Meta {
	return schema.Meta{
		TableName: "employees",
		Indexes: []schema.Index{
			schema.Index("email", "employee_id").Unique().Build(),
		},
	}
}

func (Employee) Relations() []schema.Relation {
	return []schema.Relation{}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "employee.go"), []byte(modelContent), 0644))

	// Generate and apply migration
	gen, err := migrate.NewGenerator(modelsDir, migrationsDir)
	require.NoError(t, err)
	err = gen.GenerateMigrations("create_employees")
	require.NoError(t, err)

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Verify table and constraints
	helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "employees")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "employees", "email")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "employees", "employee_id")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "employees", "age")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "employees", "salary")

	// Test unique constraint on email
	_, err = postgresDB.ExecContext(ctx, `
		INSERT INTO employees (email, employee_id, age, salary) 
		VALUES ('test@example.com', 'EMP001', 30, 50000.00)
	`)
	require.NoError(t, err)

	// Duplicate email should fail
	_, err = postgresDB.ExecContext(ctx, `
		INSERT INTO employees (email, employee_id, age, salary) 
		VALUES ('test@example.com', 'EMP002', 25, 45000.00)
	`)
	require.Error(t, err, "duplicate email should violate unique constraint")
}

// TestDBDefaultValues tests database-level defaults
func TestDBDefaultValues(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "127.0.0.1",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_db_defaults_%d", time.Now().UnixNano()),
	}
	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "db_defaults_")
	defer cleanupTemp()
	modelsDir := filepath.Join(tempDir, "models")
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create model with DB defaults
	modelContent := `package models

import "github.com/forgego/forge/schema"

type Task struct {
	schema.BaseSchema
}

func (Task) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("title").Required().MaxLength(200).Build(),
		schema.String("status").Required().MaxLength(50).DBDefault("'pending'").Build(),
		schema.Bool("is_completed").DBDefault("false").Build(),
		schema.DateTime("created_at").DBDefault("CURRENT_TIMESTAMP").Build(),
		schema.Int32("priority").DBDefault("5").Build(),
	}
}

func (Task) Meta() schema.Meta {
	return schema.Meta{
		TableName: "tasks",
	}
}

func (Task) Relations() []schema.Relation {
	return []schema.Relation{}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "task.go"), []byte(modelContent), 0644))

	// Generate and apply migration
	gen, err := migrate.NewGenerator(modelsDir, migrationsDir)
	require.NoError(t, err)
	err = gen.GenerateMigrations("create_tasks")
	require.NoError(t, err)

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Verify table exists
	helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "tasks")

	// Insert row without specifying default fields
	_, err = postgresDB.ExecContext(ctx, `INSERT INTO tasks (title) VALUES ('Test Task')`)
	require.NoError(t, err)

	// Verify defaults were applied
	var status string
	var isCompleted bool
	var priority int
	err = postgresDB.QueryRowContext(ctx, `
		SELECT status, is_completed, priority 
		FROM tasks WHERE title = 'Test Task'
	`).Scan(&status, &isCompleted, &priority)
	require.NoError(t, err)
	require.Equal(t, "pending", status, "status should default to 'pending'")
	require.False(t, isCompleted, "is_completed should default to false")
	require.Equal(t, 5, priority, "priority should default to 5")
}
