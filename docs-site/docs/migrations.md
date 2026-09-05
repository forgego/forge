---
sidebar_position: 12
description: Generate and apply database migrations.
image: /forge-social-card.svg
---

# Migrations

Forge provides generation, validation, and execution of migrations.

## Commands

- `forge makemigrations` (auto detect schema changes)
- `forge migrate` (apply)
- `forge migrate status`
- `forge migrate rollback`
- `forge migrate squash`
- `forge migrate force`
- `forge migrate fake`
- `forge migrate lint`

## Create a migration

```bash
forge makemigrations add_posts --auto
```

## Apply migrations

```bash
forge migrate
```

## Status and rollback

```bash
forge migrate status
forge migrate rollback
```

## Validation

The migration system includes:

- checksum validation
- safety checks
- drift detection
- SQL linting

## Next steps

- [Models](/docs/models/)
- [Server Overview](/docs/server/overview/)
