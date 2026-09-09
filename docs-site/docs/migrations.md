---
sidebar_position: 12
description: Automatic AST model change detection, migration workflows, checksum validation, and disaster recovery.
image: /forge-social-card.svg
---

# Migrations & Disaster Recovery

Forge includes a zero-configuration database migration engine inspired by Django and Rails. It uses Go's Abstract Syntax Tree (`go/ast`) parser to compare your declared models against existing migration states, automatically generating reversible, deterministic migration scripts.

---

## Core Commands

```bash
# Auto-detect schema changes and generate a new timestamped migration
forge makemigrations [name] --auto

# Apply pending migrations to the database
forge migrate

# Inspect migration status across all apps
forge migrate status

# Rollback the last applied migration
forge migrate rollback

# Verify checksum integrity across all migration files
forge migrate recover --verify

# Force sync migration table state in case of disaster or manual DBA intervention
forge migrate recover --force
```

---

## The Migration Workflow

1. **Modify Your Model**: Update fields, indexes, or relations in `models/`.
2. **Generate Migration**: Run `forge makemigrations add_field_name --auto`. Forge's AST analyzer compares current model definitions against previous migration state and produces a new file under `migrations/`.
3. **Review Generated Code**: Migrations are pure Go code with explicit `Up` and `Down` functions.
4. **Apply to Database**: Run `forge migrate`. Forge wraps execution in a database transaction, records applied timestamps and cryptographic SHA-256 checksums in the `forge_migrations` table.

---

## Anatomy of a Migration File

```go
package migrations

import (
    "context"
    "github.com/forgego/forge/migration"
)

func init() {
    migration.Register(migration.Migration{
        ID:   "20260908120000_add_product_price_with_tax",
        App:  "catalog",
        Dependencies: []string{
            "20260907100000_initial_catalog",
        },
        Up: func(ctx context.Context, schema *migration.SchemaBuilder) error {
            return schema.AlterTable("products", func(table *migration.TableBuilder) {
                table.AddDecimal("price_with_tax", 10, 2).
                    Generated("price * (1 + tax_rate)", true)
                table.AddIndex("idx_products_price_with_tax", "price_with_tax")
            })
        },
        Down: func(ctx context.Context, schema *migration.SchemaBuilder) error {
            return schema.AlterTable("products", func(table *migration.TableBuilder) {
                table.DropIndex("idx_products_price_with_tax")
                table.DropColumn("price_with_tax")
            })
        },
    })
}
```

---

## Disaster Recovery & Checksum Verification

In high-velocity teams, merge conflicts or manual hotfixes by DBAs can cause schema drift or mismatched migration states. Forge provides robust recovery primitives:

### 1. Checksum Drift Verification
```bash
forge migrate recover --verify
```
Computes the SHA-256 checksum of every migration file on disk and compares it against the recorded checksum stored in the `forge_migrations` table. If any applied migration file has been edited after application, Forge reports the exact drift and refuses to apply subsequent migrations without confirmation.

### 2. Force Reconciliation
```bash
forge migrate recover --force
```
Re-aligns the migration history table with current disk state, recording missing entries or updating hashes when an emergency schema patch was applied directly in production.

---

## Best Practices for Production

- **Always run migrations in CI/CD**: Include `forge migrate status` and `forge migrate recover --verify` in your deployment pipeline before launching application pods.
- **Atomic Execution**: On PostgreSQL and SQLite, DDL statements run inside transactions. If an index creation fails, the entire migration rolls back cleanly without leaving half-migrated state.
- **Concurrent Index Creation**: On production PostgreSQL databases with large tables, use `schema.AddIndexConcurrently` to avoid table locking.

---

## Next Steps

- **[Models & Schema DSL](/docs/models/)**: Define fields and constraints that power migrations.
- **[Quickstart Guide](/docs/quickstart/)**: Run your first migration in 60 seconds.

