---
sidebar_position: 0
description: Complete overview of all forge framework features, from core ORM to advanced admin interface and REST API framework.
keywords:
  - forge features
  - django go features
  - go framework features
  - orm features
  - admin interface
  - rest api
image: /forge-social-card.svg
---

# Features Overview

forge focuses on the parts of web development that are slow or repetitive:
schema design, data access, admin tooling, and API plumbing. Everything is
type-safe and designed to work together.

## Core building blocks

### 🏗️ Schema system
- Declarative models with a broad set of field types
- Field options for defaults, indexing, uniqueness, and validation
- Relationships with typed accessors
- Hooks for create/update/delete lifecycle events

### 🔧 Code generation
- Generates managers, querysets, and field expressions
- Keeps import management and formatting consistent
- Lets you focus on schema instead of boilerplate

### 🗄️ Type-safe ORM
- QuerySet API (Filter, Exclude, OrderBy, Limit, Offset, Distinct)
- Query methods (All, Get, First, Last, Count, Exists)
- SQL builder with parameter binding

## Admin interface

- Auto-generated CRUD UI from your schemas
- List views with search, filtering, sorting, and bulk actions
- Create/edit forms with widgets, fieldsets, and inline relations
- CSV/JSON export and autocomplete for relations
- Plugin slots for custom pages and UI extensions

## REST API framework

- BaseViewSet for standard CRUD endpoints
- Serializer system with validation hooks
- Auth backends (API key, basic, token, session, JWT)
- Permissions, throttling helpers, and pagination
- Renderers/parsers for JSON, XML, CSV, and YAML
- OpenAPI generator for schema output
- API versioning support

## Migrations & CLI

- Built-in migration helpers with `forge migrate`
- Project scaffolding and code generation via `forge new` and `forge generate`
- Local dev server with `forge runserver`

## Supporting systems

- Identity and session primitives
- Configuration via YAML/JSON/environment variables
- Structured logging configuration
- Middleware-based HTTP server built on chi
- Plugin registry for model, admin, API, and CLI extensions

Want the full breakdown? Use the feature-specific docs in the sidebar or jump
into the [schema system](/docs/features/schema-system/) and
[ORM system](/docs/features/orm-system/).

### Ready to start?
- [Installation guide](/docs/getting-started/installation)
- [Quick start](/docs/getting-started/quickstart)
- [Examples](/docs/examples/blog)
