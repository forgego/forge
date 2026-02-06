---
sidebar_position: 0
description: Current implementation status of forge framework features. See what's working right now.
keywords:
  - forge status
  - implementation status
  - forge features
  - django go status
  - framework progress
image: /forge-social-card.svg
---

# Implementation Status

This page is a quick, honest snapshot of what forge supports today.

## Implemented

### Core
- Schema system (fields, relations, metadata, hooks)
- Code generation (models, managers, field expressions)
- Type-safe ORM (QuerySet, Manager CRUD, field expressions)
- Database layer (connections, transactions, migrations)

### Admin
- Auto-generated admin CRUD
- List views (search, filtering, ordering)
- Form views with validation
- Actions and exports

### API
- ViewSets (CRUD)
- Serializers and validation
- Authentication and permissions
- Throttling and pagination
- OpenAPI docs

### Supporting
- Filter system
- Identity system
- CLI tooling (new, generate, migrations, runserver)

## Notes

If a feature is not listed above, assume it is not shipped yet. Check the docs
before relying on advanced behavior.

## Getting Started

1. **Install** - `go install github.com/forgego/forge/cli/cmd@latest`
2. **Create** - `forge new myapp`
3. **Define models** - write schema definitions
4. **Generate** - `forge generate`
5. **Run** - `forge runserver`
