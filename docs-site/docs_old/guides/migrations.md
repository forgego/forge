---
sidebar_position: 9
description: Create and run database migrations.
keywords:
  - forge migrations
  - database migrations
image: /forge-social-card.svg
---

# Migrations Guide

Generate and apply schema changes.

## Create a migration

```bash
forge makemigrations add_posts --auto
```

## Apply migrations

```bash
forge migrate
```

## Check status

```bash
forge migrate status
```

## Next steps

- [Models guide](/docs/guides/models)
- [Queries guide](/docs/guides/queries)
