---
id: migration-system
title: Migration System
sidebar_label: Migration System
---

# Migration System Feature

## Overview

The Migration System provides automatic database schema migration generation and execution. It detects schema changes from model definitions, generates SQL migrations, and executes them safely with rollback support.

## Location

`forge/db/migrate/` and `forge/migrate/`

## Status

✅ **Complete** - Production ready

## Core Components

### 1. Migration Generation (`forge/db/migrate/generate/`)

**Purpose:** Generate migration files from schema changes

**Components:**
- `generator.go` - Main migration generator
- `detector.go` - Change detection between schema states
- `dependencies.go` - Migration dependency resolution
- `squash.go` - Migration squashing (combining multiple migrations)

**Key Features:**
- AST-based schema parsing
- State comparison and diff generation
- SQL generation for multiple database dialects
- Migration file creation with up/down SQL
- Dependency tracking

### 2. State Management (`forge/db/migrate/state/`)

**Purpose:** Track and manage database schema state

**Components:**
- `manager.go` - State manager interface
- `loader.go` - Load state from migration files
- `inmemory.go` - In-memory state implementation
- `converter.go` - Convert between state formats
- `sorting.go` - Migration file sorting

**State Structure:**
```go
type SchemaState struct {
    Tables map[string]*TableState
}

type TableState struct {
    Name        string
    Columns     map[string]*ColumnState
    Indexes     map[string]*IndexState
    ForeignKeys map[string]*ForeignKeyState
    Constraints map[string]*ConstraintState
}
```

### 3. Change Detection (`forge/db/migrate/generate/detector.go`)

**Purpose:** Detect differences between current and previous schema

**Change Types:**
- `CreateTable` - Create new table
- `DropTable` - Drop table
- `RenameTable` - Rename table
- `AddColumn` - Add column to table
- `DropColumn` - Drop column from table
- `ModifyColumn` - Modify column definition
- `RenameColumn` - Rename column
- `AddIndex` - Add index
- `DropIndex` - Drop index
- `ModifyIndex` - Modify index
- `AddForeignKey` - Add foreign key constraint
- `DropForeignKey` - Drop foreign key constraint
- `AddConstraint` - Add custom constraint
- `DropConstraint` - Drop constraint
- `RunSQL` - Execute raw SQL
- `RunGo` - Execute Go code migration

### 4. SQL Generation (`forge/db/migrate/sql/`)

**Purpose:** Generate database-agnostic SQL from changes

**Components:**
- `builder.go` - Base SQL builder
- `postgres.go` - PostgreSQL-specific SQL
- `sqlite.go` - SQLite-specific SQL
- `common.go` - Common SQL utilities

**Features:**
- Dialect-aware SQL generation
- Proper identifier escaping
- Transaction support
- Rollback SQL generation

### 5. Migration Execution (`forge/db/migrate/execute/`)

**Purpose:** Execute migrations safely with rollback support

**Components:**
- `executor.go` - Migration executor (wraps golang-migrate)
- `apply.go` - Apply migrations
- `rollback.go` - Rollback migrations
- `status.go` - Check migration status
- `dryrun.go` - Dry-run migrations
- `recover.go` - Recover from failed migrations
- `validator.go` - Validate migrations
- `checksum.go` - Migration checksum verification

**Features:**
- Transaction-wrapped migrations
- Rollback support
- Migration status tracking
- Dirty state detection
- Checksum verification
- Dry-run mode

### 6. Migration Verification (`forge/db/migrate/verify/`)

**Purpose:** Verify migrations and detect schema drift

**Components:**
- `drift.go` - Detect schema drift
- `safety.go` - Safety checks for migrations
- `syntax.go` - SQL syntax validation
- `lint.go` - Migration linting
- `checksum.go` - Checksum verification

### 7. Migration Parsing (`forge/db/migrate/parse/`)

**Purpose:** Parse existing migration files

**Components:**
- `parse.go` - Main parser
- `lexer.go` - SQL lexer
- `ddl.go` - DDL statement parsing
- `tables.go` - Table structure extraction
- `classifier.go` - Statement classification

## Features

### ✅ Complete Features

1. **Schema Detection** - Automatic schema detection from model definitions
2. **Change Detection** - Compare current and previous schema states
3. **SQL Generation** - Generate database-agnostic SQL
4. **Migration Execution** - Execute migrations with transactions
5. **Rollback Support** - Rollback migrations safely
6. **State Management** - Track schema state across migrations
7. **Dependency Resolution** - Handle migration dependencies
8. **Checksum Verification** - Verify migration integrity
9. **Drift Detection** - Detect schema drift
10. **Safety Checks** - Validate migrations before execution
11. **Dry-Run Mode** - Test migrations without executing
12. **Recovery** - Recover from failed migrations

## Usage Examples

### Generate Migrations

```go
import (
    "github.com/forgego/forge/db/migrate"
    "github.com/forgego/forge/db/migrate/generate"
)

// Create generator
generator := generate.NewGenerator(
    "./models",      // Models directory
    "./migrations",  // Migrations directory
    migrate.DriverPostgreSQL,
)

// Generate migrations
err := generator.GenerateMigrations("add_user_table")
if err != nil {
    log.Fatal(err)
}
```

### Execute Migrations

```go
import (
    "github.com/forgego/forge/db"
    "github.com/forgego/forge/db/migrate/execute"
)

// Create executor
executor, err := execute.NewExecutor(db, "./migrations")
if err != nil {
    log.Fatal(err)
}

// Apply all pending migrations
err = executor.Migrate(ctx)
if err != nil {
    log.Fatal(err)
}

// Rollback last migration
err = executor.Rollback(ctx)
if err != nil {
    log.Fatal(err)
}
```

### Check Migration Status

```go
status, err := executor.Status(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Current version: %d\n", status.Version)
fmt.Printf("Dirty: %v\n", status.Dirty)
fmt.Printf("Pending: %d\n", len(status.Pending))
```

### Verify Migrations

```go
import "github.com/forgego/forge/db/migrate/verify"

// Check for schema drift
drift, err := verify.DetectDrift(db, "./migrations")
if err != nil {
    log.Fatal(err)
}

if len(drift) > 0 {
    fmt.Println("Schema drift detected:")
    for _, d := range drift {
        fmt.Printf("  - %s\n", d)
    }
}
```

## Integration Points

### Schema System
- Migrations generated from schema definitions
- Schema changes detected automatically
- Field types mapped to database types

### Code Generation
- AST parser extracts model definitions
- Model definitions converted to migration state
- Generated code synchronized with migrations

### Database Layer
- Migrations executed through database connection
- Transaction support for safe migrations
- Connection pooling for migration execution

### CLI Tools
- `forge migrate` - Generate and execute migrations
- `forge migrate status` - Check migration status
- `forge migrate rollback` - Rollback migrations

## Migration File Format

Migrations are stored as SQL files:

```
migrations/
  000001_initial.up.sql
  000001_initial.down.sql
  000002_add_users.up.sql
  000002_add_users.down.sql
```

**Up Migration** (`*.up.sql`):
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(150) NOT NULL UNIQUE,
    email VARCHAR(254) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

**Down Migration** (`*.down.sql`):
```sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

## Best Practices

1. **Always Generate Migrations** - Never write migrations manually
2. **Review Generated SQL** - Check generated SQL before committing
3. **Test Migrations** - Test migrations in development first
4. **Use Transactions** - All migrations are transactional by default
5. **Version Control** - Commit all migration files
6. **Rollback Support** - Ensure down migrations are reversible
7. **Avoid Data Loss** - Be careful with DROP operations
8. **Migration Ordering** - Dependencies are handled automatically
9. **Checksums** - Verify checksums before applying migrations
10. **Schema Drift** - Regularly check for schema drift

## Safety Features

### Transaction Wrapping
All migrations are wrapped in transactions for safety.

### Dirty State Detection
Detects and prevents execution when database is in dirty state.

### Checksum Verification
Verifies migration file integrity before execution.

### Rollback Safety
Validates rollback operations before executing.

### Drift Detection
Detects differences between database and migration state.

## Extension Points

### Custom Change Types
- Define custom change types
- Custom SQL generation
- Custom validation

### Custom Dialects
- Add support for new databases
- Custom SQL generation
- Database-specific optimizations

### Custom Validators
- Add custom migration validators
- Safety checks
- Business rule validation

## Future Enhancements

- [ ] Data migrations (beyond schema)
- [ ] Migration squashing UI
- [ ] Migration testing framework
- [ ] Migration performance analysis
- [ ] Multi-database support
- [ ] Migration templates
- [ ] Migration rollback strategies
