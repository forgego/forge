---
sidebar_position: 40
description: Complete feature list for forge.
image: /forge-social-card.svg
---

# Features

This is the full feature list, grouped by area. Each item is shipped.

## Core

- Schema system: fields, relations, metadata, hooks
- Schema meta options: indexes, constraints, permissions, ordering
- Field types: int/string/bool/time/date/datetime/float/decimal/text/email/url/uuid/json/bytes
- Validation: validators, tags, error messages
- Code generation: models, managers, field expressions
- Type-safe ORM: QuerySet, Manager CRUD, expressions
- Database layer: connections, transactions, migrations

## ORM

- Filtering, ordering, distinct, limits/offsets
- Select/only/defer
- SelectRelated and PrefetchRelated
- Aggregates and annotations
- Values and ValuesList
- Bulk updates and update builder

## Filters

- FilterSet and filter builder
- Query param parsing to AST
- AST to ORM expression conversion
- Security config and optimizer
- Filter widgets (autosuggest, SQL preview)

## Admin

- Admin registry and model config
- List views, form views, detail views
- Search, filters, ordering
- Actions, exports, history
- Widgets and templates
- UI overrides and plugin pages

## API

- Serializers (typed, enhanced)
- ViewSets and routers
- Authentication: token, session, JWT, basic, API key
- Permissions: allow-any, auth, admin, owner
- Throttling: anon/user rate limits
- Pagination and ordering/search filters
- Versioning: header, query param, URL path
- Content negotiation
- Parsers: JSON, form, multipart, XML
- Renderers: JSON, HTML, XML, CSV, YAML
- Caching backends
- Exceptions and problem details
- OpenAPI docs

## Identity

- User, session, token, group, permission models
- Auth backends (password, token)
- Services: user, auth, password, permission
- Password policy, lockout, rate limits
- Identity middleware

## Server

- Router and middleware stack
- Request ID, logging, recoverer, timeout
- CORS and compression
- Security headers
- CSRF protection and sessions
- Static files
- Health, readiness, liveness, metrics
- Profiling hook

## Logging & Errors

- Logging config: levels, formats, outputs
- Console/file/remote exporters
- Sampling and stacktrace controls
- Error codes and problem details
- Idempotency and sanitization
- Request ID support

## CLI

- new, generate, runserver, version
- makemigrations, migrate (up/down/status/rollback/squash)
- createsuperuser, auth
- add: app, api, handler, group, model, service
- dev: shell, test, check
