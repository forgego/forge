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

## Release test gate

The release test gate validates that all database integration tests run and pass without being skipped. This gate is executed in CI by the `release-gate` job in `.github/workflows/test.yml`. Package binaries must run serially (`-p 1`) because all packages share `FORGE_TEST_DATABASE_URL` and `CleanupDB` truncates tables.

### Environment Variables

- `FORGE_REQUIRE_DB=1`: Ensures that an unavailable database causes tests to fail (`t.Fatalf`) instead of skipping (`t.Skipf`).
- `FORGE_TEST_DATABASE_URL`: Overrides the target PostgreSQL connection URL (e.g. `postgres://user:password@localhost:5432/dbname?sslmode=disable`). Any query parameters (such as `sslmode` or connection timeouts) are preserved across test databases derived by test helpers.

### Running with the Reporter

To summarize test output and enforce that required packages have no skipped tests, build the reporter once from `forge/` and pipe each module's `go test -json` into that binary:

```bash
# Build the reporter once from forge/:
cd forge && go build -o /tmp/testreport ./internal/tools/testreport

# From the forge module:
go test -count=1 -p 1 -json ./... | /tmp/testreport --require-no-skip '^github.com/.*/(identity|internal/testutils)'

# From the tests module:
cd ../tests
go test -count=1 -p 1 -json ./... | /tmp/testreport --require-no-skip '^github.com/forgego/forge/tests/(integration|pkg_migrations|e2e)' --allow-skip 'tests/pkg_migrations$ ^TestMigrationApplySQLite$'
```

The `--require-no-skip` flag takes a regular expression matching package paths whose tests must not be skipped. If any matching test is skipped, the reporter prints an error summary and exits with code 1. The repeatable `--allow-skip '<package-regexp> <test-regexp>'` flag exempts documented skips.

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

See `docs/TECH-DEBT.md` for the test helper contract (timestamped DB
names, `t.Cleanup()`, 60s contexts, `helpers.Assert*`).

## Writing Tests

Use the shared helpers — do not hand-roll setup:

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

```go
tempDir, cleanup := testhelpers.TempDirInTests(t, "prefix_")
defer cleanup()
```

```go
helpers.AssertTableExists(ctx, t, db, "postgres", "table_name")
helpers.AssertColumnExists(ctx, t, db, "postgres", "table_name", "column_name")
helpers.AssertMigrationState(ctx, t, database, migrationsDir, expectedVersion, expectDirty)
helpers.AssertRowCount(ctx, t, db, "table_name", expectedCount)
```

Best practices: unique timestamped database names, always defer
cleanup, 60s context timeouts, helper assertions over raw checks,
independent tests with descriptive names.

Troubleshooting: "connection refused" means PostgreSQL is not up
(`psql -h localhost -U postgres -c "SELECT 1"`, default password
`123`); "database already exists" means concurrent runs colliding —
wait and retry; run unit-only suites for speed (`go test ./orm`).

## Contributing

When adding new features:

1. Write tests first (TDD)
2. Add tests to the appropriate package (`integration/migrate`, `integration/schema`, etc.)
3. Update this README if you add suites or helpers
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
- [Forge Migration Package](../forge/db/migrate/)
- [Forge DB Package](../forge/db/)

