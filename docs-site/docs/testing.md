---
title: Testing Forge Apps
description: Unit testing, API test client, shared stores test helpers, and PostgreSQL integration testing.
image: /social-card.png
---

# Testing Forge Applications

Forge provides helpers for testing models, ORM queries, REST APIs, and shared stores in isolation or against PostgreSQL 15.

## Running Tests

```bash
# Run project tests via CLI or standard Go tooling
forge test
go test -race ./...
```

## ORM & Model Unit Tests (SQLite In-Memory)

For fast unit tests of business logic and basic queries, `db.NewDBWithDriver` supports in-memory SQLite (`github.com/mattn/go-sqlite3`, requires cgo):

```go
database, err := db.NewDBWithDriver("sqlite3", ":memory:", db.WithMaxOpenConns(1))
if err != nil {
	t.Fatal(err)
}
defer database.Close()

catalog.ProductObjects.SetDB(database)
```

Note: Applying generated SQL migrations on SQLite is experimental; test schema migrations against PostgreSQL 15.

## PostgreSQL Integration Tests

When testing migrations, constraints, and production query behavior, run against PostgreSQL 15. In the Forge framework suite, setting `FORGE_REQUIRE_DB=1` causes missing database infrastructure to fail rather than skip:

```bash
FORGE_REQUIRE_DB=1 go test -race ./...
```

## API & Store Test Helpers

- **`api/testing`**: `api/testing/client.go` provides an HTTP test client for exercising registered `api.Router` / `ViewSet` endpoints with JSON payloads, headers, and status assertions.
- **`stores/storestest`**: Provides test helpers for verifying database-backed sessions, rate limiters, admin tokens, and change history.
- **`forge check`**: Validates project configuration before running or deploying.
