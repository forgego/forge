---
sidebar_position: 40
description: Complete feature overview for Forge framework.
image: /social-card.png
---

# Features

Forge provides a toolkit for building web applications in Go. This page outlines the major features by category. It describes what exists, not how well it is tested: the [support contract](/docs/status/) gives each capability a tier (supported, partially tested, experimental or not implemented) with its test evidence, and wins where the two differ.

## Core Framework

### Schema System
- Define models with fields, relations, metadata, and lifecycle hooks
- **Dual Syntax**: Fluent builder constructors (`schema.String("name")`) and functional options (`schema.StringField("name", ...)`)
- **Field types**: integers (Int64, Int32), strings, text, booleans, dates, times, timestamps, floats (Float64, Float32), fixed-point decimals, email, URL, UUID, JSON/JSONB, and binary bytes
- **Generated Columns**: First-class support for SQL computed columns (`schema.GeneratedColumn("expr", stored)`)
- **Database options**: custom column names (`DBColumn`), SQL default expressions (`DBDefault`), collation (`DBCollation`), column comments (`DBComment`), tablespace (`DBTablespace`), and index flags (`DBIndex`)
- **Choices & Enums**: Permitted value-label sets with `schema.Choices` and `schema.NewChoice`
- **Referential Integrity & Cascades**: `CascadeCASCADE`, `CascadePROTECT`, `CascadeSET_NULL`, `CascadeSET_DEFAULT`, `CascadeDO_NOTHING`
- **Lifecycle hooks**: `BeforeCreate`, `AfterCreate`, `BeforeUpdate`, `AfterUpdate`, `BeforeSave`, `AfterSave`, `BeforeDelete`, `AfterDelete`, and `Clean` validation

### Code Generation
- Auto-generate type-safe managers for each model
- Generate field expressions for compile-time query validation (`orm.Q`)
- Build REST serializers, viewsets and routes (`forge generate --api`)
- CLI commands: `forge generate`, `forge new`, `forge add`, `forge check`

### Type-Safe ORM
- QuerySet API for filtering, ordering, slicing, and aggregating data
- Manager for CRUD operations with full type safety
- Generated field expressions for compile-time validation
- IDE autocomplete for all queries and field access
- Prevent SQL injection and runtime errors
- Boolean query trees (`orm.And`, `orm.Or`, `orm.Not`)
- Eager loading via `SelectRelated` (SQL JOINs) and `PrefetchRelated` (batched queries)

### Database Layer & Migrations
- PostgreSQL (tested) and SQLite (experimental) with connection pooling
- **AST-Driven Migrations**: Introspects Go AST code to detect model diffs automatically
- Cryptographic SHA-256 checksums on all migration files to detect drift
- Recovery with `forge migrate recover --verify`, `forge migrate recover --clean` and `forge migrate force <version>`
- Reversible migrations as paired `.up.sql` and `.down.sql` files
- Lazy QuerySets: nothing runs until you execute the query
- Raw SQL through the `*sql.DB` embedded in `db.DB`, and transactions with savepoints (`db.WithTx`, `Manager.WithTx`)

## ORM & Queries

### Filtering & Querying
- `Filter()` - Add WHERE conditions
- `Exclude()` - Add NOT conditions
- `OrderBy()` - Sort results
- `Distinct()` - Remove duplicates
- `Limit()` / `Offset()` - Pagination
- `Count()` - Count without fetching

### Field Operations
- `Select()` - Choose specific fields
- `Only()` - Defer expensive fields
- `Defer()` - Exclude specific fields
- `Values()` - Get map of values
- `ValuesList()` - Get slice of values

### Relations
- `SelectRelated()` - Eager load foreign keys (JOIN)
- `PrefetchRelated()` - Prefetch many-to-many and reverse relations
- Foreign key relationships with cascade options
- Many-to-many relationships with through tables
- Reverse relation queries

### Aggregations
- `AggregateValues()` - Compute Count, Sum, Avg, Min and Max over a QuerySet (ungrouped)
- `Annotate()` - Add computed fields
- Not implemented yet: grouped aggregates (`GROUP BY`) and custom aggregate functions

### Bulk Operations
- `BulkCreate()` - Insert multiple records
- `BulkUpdate()` - Apply several updates (one `UPDATE` statement per entry)
- `UpdateBuilder()` - Build complex updates

## Advanced Filtering

### FilterSet System
- Declarative filter definitions
- Query parameter parsing to AST
- AST to ORM expression conversion
- Security controls and field validation
- Automatic query optimization

### Filter Features
- Text search with operators (contains, starts with, ends with)
- Numeric comparisons (equals, greater than, less than, between)
- Date and time filters
- Boolean filters
- Null/not null checks
- In/not in for sets

## Admin Interface

### Admin Registry
- Register models with admin site
- Configure list views and detail views
- Define which fields to display
- Set up search and filter options

### List Views
- Customizable column display
- Sortable columns
- Bulk actions (delete, export, custom)
- Pagination controls

### Form Views
- Auto-generated forms from models
- Field widgets (text, select, date picker, etc.)
- Form validation
- Fieldsets
- Related object selection and inline relations

### Admin Features
- Full-text search across fields
- Advanced filtering sidebar
- Action menu for bulk operations
- Export to CSV/JSON
- Change history tracking (kept in process memory; lost on restart)
- Permission-based access control

### Customization
- Register custom actions
- Per-object permission hooks
- Light and dark themes
- Admin and API plugins are not implemented (`registry.RegisterPlugin` returns `NotImplemented`)

## REST API Framework

### Serializers
- Typed serializers for models
- Enhanced serializers with relations
- Nested serializers
- Read-only and write-only fields
- Custom field serialization
- Validation rules

### ViewSets & Routers
- Generic viewsets for CRUD operations
- Automatic URL routing
- Custom action methods
- Bulk operations
- Filtering and search
- Pagination support

### Authentication
- Token authentication
- Session authentication
- JWT authentication
- Basic authentication
- API key authentication
- Custom auth backends
- Multi-auth support

### Permissions
- Allow any access
- Authenticated users only
- Admin users only
- Owner-based permissions
- Object-level permissions
- Custom permission classes

### Throttling
- Per-user rate limits
- Anonymous user limits
- Custom throttle rules
- Scope-based throttling
- Burst protection

### API Features
- Page-number pagination (`page`, `page_size`)
- Ordering by fields
- Search across fields
- Filtering with query params
- Content negotiation
- API versioning (header, query, path)

### Data Formats
- **Parsers**: JSON, form data, multipart, XML
- **Renderers**: JSON, HTML, XML, CSV, YAML
- Custom parsers and renderers
- Accept header handling

### OpenAPI
- OpenAPI 3.0 document with the `info` block (paths and schemas are not generated yet)

## Identity & Auth

### User Management
- User model with authentication
- Session management
- Token generation and validation
- Groups and permissions
- Custom user models

### Auth Services
- User service for CRUD
- Authentication service
- Password hashing and verification
- Permission checking
- Token management

### Security Features
- Password policy enforcement
- Account lockout after failed attempts
- Rate limiting for auth endpoints
- Session timeout
- CSRF protection

## Server & Middleware

### HTTP Server
- Built on standard library
- Customizable router
- Middleware stack
- Static file serving
- Graceful shutdown on SIGINT/SIGTERM (`StartWithGracefulShutdown`)

### Middleware
Request IDs, error handling, logging, sessions and CSRF are installed by `server.NewServer`; the rest are available to add with `Router.Use`.

- Request ID tracking
- Structured logging
- Panic recovery
- Request timeout
- CORS support
- Compression (gzip)
- Security headers
- CSRF protection
- Session management

### Monitoring
- Health check endpoints
- Readiness probes
- Liveness probes
- Metrics endpoint (reports uptime only; no Prometheus integration)
- Request logging
- Performance profiling

## Logging & Error Handling

### Logging
- Structured logging with levels (debug, info, warn, error)
- Multiple output formats (console, JSON)
- Destinations: console and file (remote output is not implemented)
- Sampling for high-volume logs
- Stacktrace control
- Request ID correlation

### Error Handling
- Standardized error codes
- Problem details format (RFC 7807)
- Error sanitization for security
- Idempotency keys
- User-friendly error messages
- Developer debug info

## CLI Tools

### Project Management
- `forge new` - Create new project
- `forge generate` - Generate code from models
- `forge runserver` - Start development server
- `forge version` - Show version info

### Database Migrations
- `forge makemigrations` - Generate migrations
- `forge migrate up` - Apply all pending migrations
- `forge migrate status` - Check migration state
- `forge migrate rollback` - Roll back the last applied migration
- `forge migrate recover`, `forge migrate force` - Recover from a dirty or altered history
- `forge migrate squash` - Not implemented yet

### User Management
- `forge createsuperuser` - Create admin user
- `forge auth` - Auth utilities

### Code Scaffolding
- `forge add app` - Create new app
- `forge add api` - Generate API viewset
- `forge add handler` - Create request handler
- `forge add model` - Add new model
- `forge add service` - Create service layer

### Development Tools
- `forge shell` - Interactive Go shell
- `forge test` - Run tests
- `forge check` - Validate configuration

## Configuration

### Database Config
- Connection settings (host, port, database, user, password)
- Connection pooling (max connections, idle connections, lifetime)
- SSL mode (`database.sslmode`)

### Server Config
- Host and port binding
- Read and write timeouts, graceful shutdown timeout, request size limit
- Health check path, metrics endpoint, static files

### Security Config
- Secret key, session secret and CSRF secret (required in production)
- CSRF-exempt paths

### Logging Config
- Log level and output format
- Console and file destinations, file rotation
- Sampling rules

## What Makes Forge Different

### Developer Experience
- Familiar patterns from Django/Rails
- Less boilerplate than typical Go frameworks
- Fast prototyping to production
- Comprehensive documentation
- CLI for common tasks

### Type Safety
- Compile-time query validation
- IDE autocomplete everywhere
- Refactoring support
- No string-based queries
- Generated code for consistency

### Batteries Included
- Don't choose between 10 ORMs
- Authentication works out of the box
- Admin interface ready to use
- Security built-in, not bolted on
- One framework, one workflow

### Path to Production
- Forge is pre-1.0; check the [support contract](/docs/status/) for what is tested
- Production secrets are validated before the server listens
- Health, readiness and liveness endpoints
- Graceful shutdown and structured request logging
- A single-instance [deployment guide](/docs/deployment/)

## Extensibility

### Plugin System
Not implemented: `registry.RegisterPlugin` rejects admin and API plugins with `NotImplemented`.

### Extension Points
- Custom field types (experimental: generation and migrations do not understand them)
- Custom validators
- Custom middleware
- Custom serializers
- Custom admin actions
- Custom filters

### Integration
- Works with standard library
- Compatible with popular Go packages
- Doesn't fight the language
- Use existing tools when needed

## Learn More

- [Quick Start](/docs/quickstart) - Get started in 5 minutes
- [Models Guide](/docs/models) - Define your data
- [ORM Guide](/docs/orm) - Query your database
- [Admin Guide](/docs/admin/overview) - Customize admin interface
- [API Guide](/docs/api/overview) - Build REST APIs
- [API Reference](/docs/api-reference) - Complete API documentation
