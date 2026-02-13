---
sidebar_position: 40
description: Complete feature overview for Forge framework.
image: /forge-social-card.svg
---

# Features

Forge provides a comprehensive toolkit for building web applications in Go. This page outlines all the major features organized by category.

## Core Framework

### Schema System
- Define models with fields, relations, metadata, and lifecycle hooks
- Field types: integers, strings, booleans, dates, times, floats, decimals, text, email, URL, UUID, JSON, and binary data
- Field options: required, unique, indexed, default values, auto-increment
- Schema metadata: table names, indexes, constraints, permissions, ordering
- Lifecycle hooks: pre-save, post-save, pre-delete, post-delete

### Code Generation
- Auto-generate type-safe managers for each model
- Generate field expressions for compile-time query validation
- Create admin interfaces automatically
- Build boilerplate for APIs and handlers
- CLI commands: `forge generate`, `forge add model`, `forge add api`

### Type-Safe ORM
- QuerySet API for filtering, ordering, and aggregating data
- Manager for CRUD operations with full type safety
- Generated field expressions for compile-time validation
- IDE autocomplete for all queries and field access
- Prevent SQL injection and runtime errors

### Database Layer
- PostgreSQL support with advanced features
- Connection pooling and transaction management
- Automatic schema migrations
- Query optimization and lazy loading
- Support for custom SQL when needed

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
- `Aggregate()` - Compute aggregate values
- `Annotate()` - Add computed fields
- Built-in aggregates: Count, Sum, Avg, Min, Max
- Custom aggregate functions
- Group by support

### Bulk Operations
- `BulkCreate()` - Insert multiple records
- `BulkUpdate()` - Update multiple records
- `UpdateBuilder()` - Build complex updates
- Batch operations for performance

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

### Filter Widgets
- Auto-suggest for text fields
- SQL preview before execution
- Filter validation and sanitization
- Custom filter implementations

## Admin Interface

### Admin Registry
- Register models with admin site
- Configure list views and detail views
- Define which fields to display
- Set up search and filter options

### List Views
- Customizable column display
- Sortable columns
- Inline editing
- Bulk actions (delete, export, custom)
- Pagination controls

### Form Views
- Auto-generated forms from models
- Field widgets (text, select, date picker, etc.)
- Form validation
- Custom form layouts
- Related object selection

### Admin Features
- Full-text search across fields
- Advanced filtering sidebar
- Action menu for bulk operations
- Export to CSV/JSON
- Change history tracking
- Permission-based access control

### Customization
- Override templates
- Add custom views and pages
- Register custom actions
- Plugin system for extensions
- Theme customization

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
- Pagination: limit/offset, cursor, page number
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
- Auto-generated OpenAPI/Swagger docs
- Interactive API explorer
- Schema definitions
- Request/response examples

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
- Graceful shutdown

### Middleware
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
- Metrics collection
- Request logging
- Performance profiling

## Logging & Error Handling

### Logging
- Structured logging with levels (debug, info, warn, error)
- Multiple output formats (console, JSON)
- Multiple destinations (console, file, remote)
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
- `forge migrate` - Apply migrations
- `forge migrate up` - Run specific migration
- `forge migrate down` - Rollback migration
- `forge migrate status` - Check migration state
- `forge migrate rollback` - Rollback last migration
- `forge migrate squash` - Combine migrations

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
- SSL/TLS settings
- Query logging
- Timezone settings

### Server Config
- Host and port binding
- Debug mode
- Secret key management
- Allowed hosts
- CORS settings
- Static files configuration

### Security Config
- CSRF protection
- Session settings
- Cookie configuration
- Security headers
- Rate limiting

### Logging Config
- Log levels per module
- Output formats
- Destinations
- Rotation settings
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

### Production Ready
- Battle-tested patterns
- Security by default
- Performance optimizations
- Monitoring and metrics
- Error handling and logging

## Extensibility

### Plugin System
- Register custom plugins
- Hook into framework lifecycle
- Extend core functionality
- Share plugins across projects

### Extension Points
- Custom field types
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
