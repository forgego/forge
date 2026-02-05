# Testing Guide

This document provides comprehensive information about testing the Forge framework's schema and migration system.

## Table of Contents

- [Test Organization](#test-organization)
- [Running Tests](#running-tests)
- [Test Categories](#test-categories)
- [Migration System Tests](#migration-system-tests)
- [Schema System Tests](#schema-system-tests)
- [Writing New Tests](#writing-new-tests)
- [Test Helpers](#test-helpers)

## Test Organization

Tests are organized by type and purpose:

```
tests/
├── integration/          # Integration tests with real databases
│   ├── migrate/         # Migration system tests
│   ├── orm/             # ORM/Query tests
│   └── schema/          # Schema builder tests
├── e2e/                 # End-to-end CLI tests
│   └── cli/
├── helpers/             # Test assertion helpers
├── infra/               # Infrastructure setup (docker, filesystem)
├── testhelpers/         # Test utilities
└── testdata/            # Test fixtures and models
```

## Running Tests

### Prerequisites

- PostgreSQL running on `localhost:5432` (or set `POSTGRES_HOST`/`POSTGRES_PORT`)
- PostgreSQL user: `postgres`, password: `123` (or set `POSTGRES_USER`/`POSTGRES_PASSWORD`)
- Go 1.21+

### Run All Tests

```bash
cd tests
go test ./...
```

### Run Specific Test Package

```bash
# Migration tests only
go test ./integration/migrate -v

# Schema tests only
go test ./integration/schema -v

# ORM tests only
go test ./integration/orm -v

# E2E CLI tests
go test ./e2e/cli -v
```

### Run Specific Test

```bash
go test ./integration/migrate -run TestGeneration_CreateTable -v
```

### Run With Coverage

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Categories

### 🔴 Integration Tests (`integration/`)

Real database and filesystem tests. These test actual SQL execution and database state.

**Migration Tests** (`integration/migrate/`):
- Generation (change detection, SQL generation)
- Execution (up/down, migrate to version, rollback)
- Reversibility (ensuring migrations can be reversed)
- Edge cases (renames, type changes, constraints)
- PostgreSQL-specific features (GIN/GiST indexes, JSONB, UUIDs)
- Advanced fields (generated columns, custom constraints)
- Recovery (dirty state, force recovery)
- E-commerce scenarios (incremental, full schema)

**Schema Tests** (`integration/schema/`):
- Field builders (all field types)
- Complex models (multiple field types, relationships)
- Field options (validation, database, presentation)

**ORM Tests** (`integration/orm/`):
- Query building
- Type-safe field expressions
- SQL generation

### 🟡 E2E Tests (`e2e/`)

Full system tests that run actual CLI commands.

- `forge makemigrations`
- `forge migrate`
- `forge rollback`

### 🛠 Helper Packages

**`helpers/`**: Pure assertion helpers (no I/O)
- `assertions.go` - Complex assertions
- `sql_assertions.go` - SQL/database assertions
- `migration_assertions.go` - Migration-specific assertions

**`infra/`**: Infrastructure management
- `docker/` - PostgreSQL container lifecycle
- `filesystem/` - Temporary directories, file operations

**`testhelpers/`**: Test utilities
- PostgreSQL connection helpers
- Migration execution helpers
- Filesystem helpers

## Migration System Tests

### Change Detection Tests

Test that the migration generator correctly detects schema changes.

**Covered Changes:**
- Create table
- Drop table
- Rename table
- Add column
- Drop column
- Rename column
- Modify column type
- Add index (simple, unique, composite, GIN, GiST, functional, partial, covering)
- Drop index
- Add foreign key
- Drop foreign key
- Add constraint (unique, check)
- Drop constraint

**Example:**

```go
func TestGeneration_CreateTable(t *testing.T) {
    // Setup database and temp directories
    // Create model definition
    // Generate migration
    // Verify migration files exist and contain expected SQL
}
```

### Execution Tests

Test that migrations can be applied and rolled back correctly.

**Test Cases:**
- `MigrateUp` - Apply all pending migrations
- `MigrateDown` - Roll back migrations
- `MigrateTo` - Migrate to specific version
- `RollbackTo` - Rollback to specific version
- No changes detected
- Migration with data preservation

**Example:**

```go
func TestExecution_MigrateUp(t *testing.T) {
    // Create migration files
    // Apply migrations
    // Verify tables/columns exist
    // Verify migration version
}
```

### Reversibility Tests

Test that migrations can be reversed without data loss (when possible).

**Test Cases:**
- Simple up/down cycle
- Multiple migration rollback sequence
- Reapply after rollback
- Rollback with data
- Partial rollback

### Edge Case Tests

Test complex schema changes that may have subtle issues.

**Test Cases:**
- Column rename (vs drop+add)
- Column type change
- Table rename
- Add column with default
- Add NOT NULL column without default
- Drop column with data
- Add/drop foreign keys
- Add/drop unique constraints
- Add/drop indexes

### PostgreSQL-Specific Tests

Test PostgreSQL-specific features.

**Test Cases:**
- GIN indexes (for JSONB)
- GiST indexes (for spatial data)
- JSONB operations
- Array column types
- Custom enum types
- Partial indexes
- Functional indexes
- Covering indexes
- UUID type
- Numeric precision
- Timestamp with time zone

### Advanced Field Tests

Test advanced schema features.

**Test Cases:**
- Generated columns (STORED/VIRTUAL)
- Custom DB column names
- Column comments
- Database-level defaults (DBDefault)
- Custom constraints (CHECK)
- Min/Max value constraints

### Recovery Tests

Test recovery from migration failures.

**Test Cases:**
- Force recovery from dirty state
- Migration status reporting
- Detailed status (applied/pending)
- Version tracking

## Schema System Tests

### Field Type Tests

Test all field types and their builders.

**Field Types:**
- Int64, Int32
- String, Text, Email, URL
- Bool
- DateTime, Date, Time
- Float32, Float64, Decimal
- JSON (maps to JSONB in PostgreSQL)
- Bytes
- UUID

### Field Option Tests

Test field configuration options.

**Validation Options:**
- `Required()` / `Optional()`
- `Unique()`
- `MaxLength()` / `MinLength()`
- `MaxValue()` / `MinValue()`
- `MaxDigits()` / `DecimalPlaces()` (for Decimal)
- `Blank()`
- `Choices()` (enum-like behavior)
- `Validators()` (custom validation)

**Database Options:**
- `DBColumn()` - Custom column name
- `DBType()` - Custom SQL type
- `DBDefault()` - Database-level default
- `DBComment()` - Column comment
- `DBIndex()` - Create index
- `DBCollation()` - Character collation
- `GeneratedColumn()` - Generated/computed column

**Presentation Options:**
- `VerboseName()` - Human-readable name
- `HelpText()` - Help text for forms
- `Editable()` - Control editability
- `Serialize()` - Control serialization

### Relationship Tests

Test relationship types.

**Relationship Types:**
- `ForeignKey()` - One-to-many
- `OneToOne()` - One-to-one
- `ManyToMany()` - Many-to-many (via junction table)

**Options:**
- `OnDelete()` - CASCADE, SET NULL, RESTRICT, etc.
- `RelatedName()` - Reverse relation name

### Meta Options Tests

Test model-level metadata.

**Meta Options:**
- `TableName` - Custom table name
- `Indexes` - Table indexes
- `CustomSQL` - Custom SQL statements (e.g., CREATE TYPE)

## Writing New Tests

### Basic Test Structure

```go
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
    "github.com/forgego/forge/migrate"
    "github.com/forgego/forge/tests/helpers"
    "github.com/forgego/forge/tests/testhelpers"
)

func TestYourFeature(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    // Setup PostgreSQL
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

    // Setup temp directories
    tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "your_feature_")
    defer cleanupTemp()
    modelsDir := filepath.Join(tempDir, "models")
    migrationsDir := filepath.Join(tempDir, "migrations")
    require.NoError(t, os.MkdirAll(modelsDir, 0755))
    require.NoError(t, os.MkdirAll(migrationsDir, 0755))

    // Create model files
    modelContent := `package models
    // ... your model definition
    `
    require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "model.go"), []byte(modelContent), 0644))

    // Generate migration
    gen, err := migrate.NewGenerator(modelsDir, migrationsDir)
    require.NoError(t, err)
    err = gen.GenerateMigrations("your_migration_name")
    require.NoError(t, err)

    // Apply migration
    database, err := db.NewDB(dsn)
    require.NoError(t, err)
    defer database.Close()

    runner, err := db.NewMigrationRunner(database, migrationsDir)
    require.NoError(t, err)
    defer runner.Close()

    err = runner.Migrate(ctx)
    require.NoError(t, err)

    // Verify results
    helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "your_table")
    // ... more assertions
}
```

### Using Test Helpers

**PostgreSQL Setup:**

```go
opts := testhelpers.PostgresOpts{
    UseDirect: true,
    Host:      "localhost",
    Port:      "5432",
    User:      "postgres",
    Password:  "123",
    DBName:    fmt.Sprintf("test_%d", time.Now().UnixNano()),
}
postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
require.NoError(t, err)
defer cleanup()
```

**Temporary Directories:**

```go
tempDir, cleanup := testhelpers.TempDirInTests(t, "prefix_")
defer cleanup()
```

**Assertions:**

```go
// Table exists
helpers.AssertTableExists(ctx, t, db, "postgres", "table_name")

// Column exists
helpers.AssertColumnExists(ctx, t, db, "postgres", "table_name", "column_name")

// Foreign key exists
helpers.AssertForeignKeyExists(ctx, t, db, "postgres", "table_name", "column_name")

// Migration state
helpers.AssertMigrationState(ctx, t, database, migrationsDir, expectedVersion, expectDirty)

// Row count
helpers.AssertRowCount(ctx, t, db, "table_name", expectedCount)
```

## Test Coverage Goals

### Minimum Coverage

- ✅ All field types
- ✅ All field options (validation, database, presentation)
- ✅ All relationship types
- ✅ All change types (create, drop, modify, rename for tables/columns)
- ✅ All index types
- ✅ Foreign key constraints
- ✅ Unique constraints
- ✅ Migration execution (up/down/to version/rollback)
- ✅ Migration reversibility
- ✅ PostgreSQL-specific features
- ✅ Edge cases
- ✅ Recovery scenarios

### Extended Coverage

- ✅ Generated columns
- ✅ Custom DB options
- ✅ Database-level defaults
- ✅ Constraint validation
- ✅ Migration status tracking
- ✅ Force recovery

## Best Practices

1. **Unique Database Names**: Always use timestamps in database names to avoid conflicts
2. **Cleanup**: Always defer cleanup functions
3. **Context Timeouts**: Use reasonable timeouts (60s for most tests)
4. **Assertions**: Use helper assertions for better error messages
5. **Test Isolation**: Each test should be independent
6. **Descriptive Names**: Test names should describe what they test
7. **Documentation**: Add comments explaining complex test scenarios

## Troubleshooting

### Tests Fail with "connection refused"

Ensure PostgreSQL is running:
```bash
psql -h localhost -U postgres -c "SELECT 1"
```

Set correct environment variables:
```bash
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=123
```

### Tests Fail with "database already exists"

Tests create unique database names with timestamps. If you see this error, it might be from concurrent tests. Wait a moment and retry.

### Slow Tests

Integration tests are slower because they use real databases. To run faster unit tests only:
```bash
go test ./orm -short
```

### Build Errors

Ensure all dependencies are installed:
```bash
go mod tidy
go mod download
```

## Contributing

When adding new features:

1. Write tests first (TDD)
2. Add tests to the appropriate package
3. Update this documentation
4. Ensure all tests pass: `go test ./...`
5. Check test coverage

## References

- [Forge Schema Package](../forge/schema/)
- [Forge Migration Package](../forge/migrate/)
- [Forge DB Package](../forge/db/)
- [Test Helpers](./testhelpers/)
- [Test Fixtures](./testdata/)
