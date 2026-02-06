---
sidebar_position: 13
description: Generate and apply database migrations.
image: /forge-social-card.svg
---

# Migrations

Use the CLI and migration engine to generate and apply schema changes.

## What you can do

- Generate migrations from schema changes
- Apply, rollback, and inspect migration status
- Validate migrations (lint, safety, drift, checksums)

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

## Next steps

- [Models](/docs/models/)
- [Server Overview](/docs/server/overview/)
