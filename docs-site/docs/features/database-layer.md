---
id: database-layer
title: Database Layer
sidebar_label: Database Layer
---

# Database Layer Feature

## Overview

The Database Layer provides database connection management, transaction handling, and SQL execution. It wraps `database/sql` with additional features like connection pooling, savepoints, and migration integration.

## Location

`forge/db/`

## Status

✅ **Complete** - Production ready

## Core Components

### 1. Database Connection (`db.go`)

**Purpose:** Database connection wrapper with connection pooling

**Structure:**
```go
type DB struct {
    *sql.DB
    Driver string
}
```

**Features:**
- Connection pooling (via database/sql)
- Multiple database driver support (PostgreSQL, SQLite)
- Connection lifecycle management
- Health checks via Ping
- DSN-based connection
- Config-based connection

**Usage:**
```go
// From DSN
db, err := db.NewDB("postgres://user:pass@localhost/dbname?sslmode=disable")
if err != nil {
    log.Fatal(err)
}

// From config
db, err := db.NewDBFromConfig(cfg)
if err != nil {
    log.Fatal(err)
}
```

### 2. Transaction Management (`transaction.go`)

**Purpose:** Transaction handling with savepoint support

**Features:**
- Transaction wrapping
- Savepoint support
- Nested transactions
- Automatic rollback on error
- Context support

**Usage:**
```go
// Simple transaction
err := db.WithTx(ctx, func(tx *db.Tx) error {
    // Multiple operations
    _, err := tx.Exec("INSERT INTO users (name) VALUES ($1)", "John")
    if err != nil {
        return err
    }
    _, err = tx.Exec("INSERT INTO profiles (user_id) VALUES ($1)", userID)
    return err
})

// With savepoint
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    log.Fatal(err)
}

sp, err := tx.CreateSavepoint("checkpoint1")
if err != nil {
    log.Fatal(err)
}

// ... operations ...

if shouldRollback {
    err = sp.RollbackToSavepoint()
} else {
    err = sp.ReleaseSavepoint()
}
```

### 3. Migration Integration (`migrations.go`)

**Purpose:** Integration with migration system

**Features:**
- Migration table management
- Migration execution
- Migration status tracking
- Migration rollback

**Usage:**
```go
import "github.com/forgego/forge/db/migrate/execute"

executor, err := execute.NewExecutor(db, "./migrations")
if err != nil {
    log.Fatal(err)
}

err = executor.Migrate(ctx)
if err != nil {
    log.Fatal(err)
}
```

### 4. SQL Building Support

**Purpose:** Support for SQL builder with proper escaping

**Features:**
- Identifier escaping
- Parameter binding
- Dialect-aware SQL
- Query rebinding (PostgreSQL to SQLite)

**Usage:**
```go
import "github.com/forgego/forge/orm"

// SQL builder handles escaping
builder := orm.NewSQLBuilder()
query := builder.BuildSelect("users", []string{"id", "name"}, nil)
// Returns: SELECT "id", "name" FROM "users"

// Rebind for different dialects
query = db.Rebind(query) // Converts $1 to ? for SQLite
```

## Features

### ✅ Complete Features

1. **Connection Pooling** - Automatic connection pooling via database/sql
2. **Multiple Drivers** - PostgreSQL and SQLite support
3. **Transaction Support** - Full transaction support with savepoints
4. **Savepoints** - Nested transaction support via savepoints
5. **Context Support** - Context-aware operations
6. **Error Handling** - Proper error wrapping and handling
7. **Health Checks** - Ping for connection health
8. **Migration Integration** - Seamless migration system integration
9. **SQL Rebinding** - Dialect-aware query rebinding
10. **Config Integration** - Configuration-based connection setup

## Usage Examples

### Basic Connection

```go
import (
    "github.com/forgego/forge/db"
    "github.com/forgego/forge/config"
)

// Create config
cfg := config.New()
cfg.Set("database.driver", "postgres")
cfg.Set("database.host", "localhost")
cfg.Set("database.port", 5432)
cfg.Set("database.user", "user")
cfg.Set("database.password", "pass")
cfg.Set("database.name", "mydb")

// Create database connection
db, err := db.NewDBFromConfig(cfg)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Test connection
if err := db.Ping(); err != nil {
    log.Fatal(err)
}
```

### Transaction with Error Handling

```go
err := db.WithTx(ctx, func(tx *db.Tx) error {
    // Insert user
    result, err := tx.ExecContext(ctx,
        "INSERT INTO users (username, email) VALUES ($1, $2)",
        "john", "john@example.com",
    )
    if err != nil {
        return err
    }

    userID, err := result.LastInsertId()
    if err != nil {
        return err
    }

    // Insert profile
    _, err = tx.ExecContext(ctx,
        "INSERT INTO profiles (user_id, bio) VALUES ($1, $2)",
        userID, "Software developer",
    )
    return err
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
}
```

### Savepoint Example

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

// Create savepoint
sp1, err := tx.CreateSavepoint("before_complex_operation")
if err != nil {
    log.Fatal(err)
}

// Complex operation
err = performComplexOperation(tx)
if err != nil {
    // Rollback to savepoint
    if err := sp1.RollbackToSavepoint(); err != nil {
        log.Fatal(err)
    }
    // Continue with alternative path
    err = performAlternativeOperation(tx)
    if err != nil {
        return err
    }
} else {
    // Release savepoint
    if err := sp1.ReleaseSavepoint(); err != nil {
        log.Fatal(err)
    }
}

// Commit transaction
return tx.Commit()
```

### Query Execution

```go
// Simple query
rows, err := db.QueryContext(ctx, "SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
    var id int64
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("User %d: %s\n", id, name)
}

// Exec with result
result, err := db.ExecContext(ctx,
    "UPDATE users SET last_login = $1 WHERE id = $2",
    time.Now(), userID,
)
if err != nil {
    log.Fatal(err)
}

affected, err := result.RowsAffected()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Updated %d rows\n", affected)
```

## Integration Points

### ORM System
- ORM uses database connection for queries
- Transaction support for ORM operations
- Connection pooling for performance

### Migration System
- Migrations executed through database connection
- Transaction-wrapped migrations
- Migration state stored in database

### Configuration System
- Database config from configuration
- Environment variable support
- Default values

### Server System
- Database connection in request context
- Connection lifecycle management
- Health check endpoints

## Connection Pooling

The database layer uses `database/sql` connection pooling:

**Configuration:**
```go
// Set connection pool settings
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(10 * time.Minute)
```

**Best Practices:**
- Set appropriate pool sizes
- Use connection timeouts
- Monitor connection usage
- Handle connection errors gracefully

## Error Handling

All database operations return errors that should be checked:

```go
db, err := db.NewDB(dsn)
if err != nil {
    // Handle connection error
    log.Fatal(err)
}

err = db.Ping()
if err != nil {
    // Handle ping error
    log.Fatal(err)
}
```

## Security

### SQL Injection Prevention
- Always use parameterized queries
- Never concatenate user input into SQL
- Use SQL builder for safe query construction

### Connection Security
- Use SSL for production connections
- Store credentials securely
- Use connection string validation

## Performance

### Connection Pooling
- Reuse connections efficiently
- Configure pool sizes appropriately
- Monitor connection usage

### Query Optimization
- Use prepared statements for repeated queries
- Batch operations when possible
- Use transactions for multiple operations

## Best Practices

1. **Always Check Errors** - Check all database operation errors
2. **Use Transactions** - Use transactions for multiple operations
3. **Close Resources** - Always close rows and connections
4. **Context Support** - Use context for cancellation and timeouts
5. **Connection Pooling** - Configure pool sizes appropriately
6. **Parameterized Queries** - Always use parameterized queries
7. **Health Checks** - Regular health checks for connections
8. **Error Handling** - Proper error handling and logging
9. **Resource Management** - Proper resource cleanup
10. **Security** - Secure connection strings and credentials

## Future Enhancements

- [ ] Connection retry logic
- [ ] Query result caching
- [ ] Query logging and profiling
- [ ] Connection monitoring
- [ ] Read/write splitting
- [ ] Database sharding support
- [ ] Connection health metrics
