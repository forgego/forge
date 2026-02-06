---
name: forge-migrations-schema
description: Work with Forge schema primitives and database migrations. Use when defining schema behavior, generating migrations, applying or rolling back migrations, or validating migration files.
---

# Forge Migrations & Schema

## Overview
Use schema primitives to define models and manage database changes with the migration system.

## When to Use
- The task mentions schema primitives or migration files.
- You need to generate, apply, roll back, or validate migrations.
- The user asks about migration status, linting, or fake migrations.

## Quick Start
1. Update models and schema definitions.
2. Generate migrations with `forge makemigrations <name> --auto`.
3. Apply with `forge migrate` and check status with `forge migrate status`.

## Common Tasks
- Preview migration plans with `forge migrate show`.
- Lint migrations with `forge migrate lint` and `--verbose` when debugging.
- Roll back with `forge migrate down 1` or apply to a version with `forge migrate up <n>`.
- Use `forge migrate fake` or `--fake-initial` for existing databases.

## Gotchas
- Migration files must live under `migrations/` and follow `*_*.up.sql` and `*_*.down.sql` naming.
- Dependency auto-detection relies on relations and references in model definitions.
- Always verify pending migrations before applying in production.

## References
- [Migration guide](references/migrations.md)
- [Schema system](references/schema-system.md)
- [Migration system internals](references/migration-system.md)
