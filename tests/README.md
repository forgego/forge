# Forge Testing Suite

Comprehensive testing suite for the Forge framework's schema and migration system.

## Quick Start

```bash
cd tests
go test ./...
```

## Test Organization

```
tests/
├── integration/          # Integration tests with real databases
│   ├── migrate/         # Migration system tests (45+ tests)
│   ├── orm/             # ORM/Query tests
│   └── schema/          # Schema builder tests
├── e2e/                 # End-to-end CLI tests
│   └── cli/
├── helpers/             # Test assertion helpers
├── infra/               # Infrastructure setup (docker, filesystem)
├── testhelpers/         # Test utilities
├── testdata/            # Test fixtures and models
│   └── models/          # Model definitions for testing
├── TESTING.md           # Comprehensive testing guide
└── README.md            # This file
```

## Test Coverage

### Migration System (45+ tests)

**Generation & Change Detection:**
- Create/drop/rename tables
- Add/drop/rename/modify columns
- Add/drop indexes (simple, unique, composite, GIN, GiST, functional, partial, covering)
- Add/drop foreign keys and constraints
- No-change detection

**Execution:**
- Migrate up/down
- Migrate to specific version
- Rollback to specific version
- Partial rollback

**PostgreSQL Features:**
- GIN/GiST indexes
- JSONB operations
- Array types
- Custom enum types
- Partial/functional/covering indexes
- UUID type
- Numeric precision
- Timestamp with time zone

**Advanced Fields:**
- Generated columns (STORED/VIRTUAL)
- Custom DB column names
- Column comments
- Database-level defaults
- Custom constraints (CHECK)
- Min/Max value constraints

**Recovery & Status:**
- Force recovery from dirty state
- Migration status reporting
- Version tracking

**Scenarios:**
- Full e-commerce schema
- Incremental e-commerce migrations
- Complex schema evolution

### Schema System (3+ tests)

- All field types (Int64, String, Text, Bool, DateTime, Decimal, JSON, UUID, etc.)
- Field options (Required, Unique, MaxLength, MinValue, etc.)
- Database options (DBColumn, DBType, DBDefault, DBComment, etc.)
- Presentation options (VerboseName, HelpText, Editable, Serialize)
- Complex models with multiple field types

### ORM System (5+ tests)

- Field expressions
- Comparison expressions
- Query building (Q objects)
- SQL generation
- Identifier escaping

## Running Tests

### All Tests

```bash
cd tests
go test ./...
```

### Specific Package

```bash
# Migration tests only
go test ./integration/migrate -v

# Schema tests only
go test ./integration/schema -v

# ORM tests only
go test ./integration/orm -v
```

### Specific Test

```bash
go test ./integration/migrate -run TestGeneration_CreateTable -v
```

### With Coverage

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Prerequisites

- **PostgreSQL**: Running on `localhost:5432`
  - User: `postgres`
  - Password: `123`
  - Or set environment variables: `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`
- **Go**: 1.21 or higher

## Test Helpers

### PostgreSQL Setup

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

### Assertions

```go
// Table and column assertions
helpers.AssertTableExists(ctx, t, db, "postgres", "table_name")
helpers.AssertColumnExists(ctx, t, db, "postgres", "table_name", "column_name")
helpers.AssertForeignKeyExists(ctx, t, db, "postgres", "table_name", "column_name")

// Migration state
helpers.AssertMigrationState(ctx, t, database, migrationsDir, expectedVersion, expectDirty)

// Row count
helpers.AssertRowCount(ctx, t, db, "table_name", expectedCount)
```

## Test Fixtures

Pre-defined model fixtures in `testdata/models/`:

- **basic_models.go**: User, Product, Order
- **relationships_models.go**: Author, Post, Tag, UserProfile (with ForeignKey, OneToOne, ManyToMany)
- **complex_fields_models.go**: Event, Settings (with JSON, Bytes, Decimal, DateTime)
- **postgres_features_models.go**: ProductWithJSONB, UserWithUUID, DocumentWithTimestamps, OrderWithStatusEnum

## Documentation

See [TESTING.md](./TESTING.md) for comprehensive testing guide including:
- Detailed test categories
- Writing new tests
- Best practices
- Troubleshooting

## Contributing

When adding new features:

1. Write tests first (TDD)
2. Add tests to the appropriate package (`integration/migrate`, `integration/schema`, etc.)
3. Update documentation (this README and TESTING.md)
4. Ensure all tests pass: `go test ./...`
5. Check test coverage

## Test Status

✅ **All major features tested**
- Schema definition and builders
- Migration generation (all change types)
- Migration execution (up/down/to version/rollback)
- PostgreSQL-specific features
- Advanced field options
- Recovery scenarios
- Full schema evolution scenarios

## Quick Reference

| Command | Description |
|---------|-------------|
| `go test ./...` | Run all tests |
| `go test ./integration/migrate -v` | Run migration tests with verbose output |
| `go test ./integration/migrate -run TestName` | Run specific test |
| `go test ./... -short` | Run only short tests (skip slow integration tests) |
| `go test ./... -coverprofile=coverage.out` | Generate coverage report |
| `go test ./... -count=1` | Disable test caching |

## Troubleshooting

### Connection Refused

Ensure PostgreSQL is running:
```bash
psql -h localhost -U postgres -c "SELECT 1"
```

### Slow Tests

Integration tests use real databases and are slower. This is expected.

### Build Errors

```bash
go mod tidy
go mod download
```

## References

- [Forge Schema Package](../forge/schema/)
- [Forge Migration Package](../forge/migrate/)
- [Forge DB Package](../forge/db/)
- [Comprehensive Testing Guide](./TESTING.md)
