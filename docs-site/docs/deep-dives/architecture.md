---
id: architecture
title: Architecture
sidebar_label: Architecture
---

# forge Framework - Architecture Documentation

## Overview

forge is a full-stack web framework for Go that provides Django-like functionality with type safety. This document describes the complete architecture of the framework.

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                          │
│  User Models, Views, Controllers, Routes                      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                  Code Generation Layer                        │
│  AST Parser → Schema Analysis → Code Generator                │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Framework API Layer                        │
│  Admin, API, ORM, Identity, Filter                            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Database Layer                             │
│  SQL Builder, Query Execution, Transactions, Migrations       │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                         │
│  HTTP Server, Security, Logging, Config, Server               │
└─────────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. Schema System (`forge/schema/`)

**Purpose:** Declarative model definitions

**Components:**
- `schema.go` - Core Schema interface
- `field.go` - Field definitions with builders
- `relation.go` - Relationship definitions
- `meta.go` - Model metadata
- `hooks.go` - Lifecycle hooks
- `constraint_builder.go` - Database constraints

**Key Interfaces:**
```go
type Schema interface {
    Fields() []Field
    Relations() []Relation
    Meta() Meta
    Hooks() *ModelHooks
}
```

### 2. Code Generation (`forge/codegen/`)

**Purpose:** Generate type-safe Go code from schema definitions

**Components:**
- `ast_parser.go` - Go AST parser for schema extraction
- `generator.go` - Main generation orchestrator
- `writer.go` - Code file writer
- `templates.go` - Code generation templates
- `templates/` - Template files

**Generated Files:**
- Model structs
- FieldExpr definitions
- Manager with CRUD operations
- QuerySet wrappers

### 3. ORM System (`forge/orm/`)

**Purpose:** Type-safe database queries and operations

**Components:**
- `queryset.go` - QuerySet implementation
- `manager.go` - Manager with CRUD operations
- `field_expr.go` - Type-safe field accessors
- `expression.go` - Query expressions
- `sql_builder.go` - SQL generation
- `update_builder.go` - Update operations
- `aggregates.go` - Aggregate functions
- `annotations.go` - Annotation support

**Key Types:**
- `QuerySet[T]` - Type-safe query interface
- `Manager[T]` - Model manager
- `FieldExpr[T]` - Type-safe field access
- `Expression` - Query conditions

### 4. Admin System (`forge/admin/`)

**Purpose:** Django-like admin interface

**Components:**
- `admin.go` - Type-safe Admin[T] and Config[T]
- `registry.go` - Admin model registry
- `list_view.go` - List view with pagination, search, filtering
- `detail_view.go` - Detail view
- `form_view.go` - Create/update forms
- `fields.go` - Field expressions
- `filters.go` - Filter system
- `widgets.go` - Form widgets
- `actions.go` - Bulk actions
- `export.go` - CSV/JSON export
- `inlines.go` - Related model editing
- `http/` - HTTP handlers

**Key Types:**
- `Admin[T]` - Type-safe admin interface
- `Config[T]` - Admin configuration
- `FormData` - Form data handling

### 5. API Framework (`forge/api/`)

**Purpose:** DRF-like REST API framework

**Components:**
- `viewset.go` - BaseViewSet with CRUD
- `serializer.go` - Serializer system
- `serializers/` - Field serializers
- `authentication/` - Auth backends (Token, JWT, Basic, Session, API Key)
- `permissions/` - Permission classes
- `throttling/` - Rate limiting
- `renderers/` - Response renderers (JSON, XML, YAML, HTML, CSV)
- `parsers/` - Request parsers
- `filters/` - Field filtering
- `pagination.go` - Pagination
- `exceptions/` - Exception handling
- `versioning/` - API versioning
- `caching/` - Cache backends
- `docs/` - OpenAPI documentation

**Key Types:**
- `ViewSet` - Base viewset interface
- `Serializer` - Serializer interface
- `Authentication` - Auth interface
- `Permission` - Permission interface

### 6. Filter System (`forge/filter/`)

**Purpose:** Advanced filtering system with AST support

**Components:**
- `filter.go` - Base filter interface
- `filterset.go` - FilterSet implementation
- `parser.go` - Query string parser
- `ast.go` - AST representation
- `expression_converter.go` - Expression conversion
- `filters/` - Filter implementations
- `widgets/` - Filter widgets
- `security.go` - Security validation
- `optimizer.go` - Query optimization

**Key Types:**
- `Filter[T]` - Filter interface
- `FilterSet[T]` - FilterSet implementation
- `FilterNode` - AST node

### 7. Identity System (`forge/identity/`)

**Purpose:** User management and authentication

**Components:**
- `models/` - User, Session, Permission, Group, Token models
- `repository/` - Data access layer
- `service/` - Business logic (User, Auth, Password, Permission)
- `backends/` - Authentication backends
- `serializers/` - API serializers
- `handlers/` - HTTP handlers
- `middleware/` - Authentication middleware
- `factory.go` - System setup
- `config/` - Configuration

**Key Types:**
- `IdentitySystem` - Complete identity system
- `UserService` - User management
- `AuthService` - Authentication
- `PasswordService` - Password management

### 8. Database Layer (`forge/db/`)

**Purpose:** Database connection and operations

**Components:**
- `db.go` - Database connection wrapper
- `transaction.go` - Transaction management
- `migrations.go` - Migration system integration
- `migrate/` - Migration utilities

**Features:**
- Connection pooling
- Transaction support with savepoints
- Migration management
- SQL builder integration

### 9. Migration System (`forge/migrate/`)

**Purpose:** Database schema migrations

**Components:**
- `migrate.go` - Migration execution
- `core.go` - Core migration logic
- `generate/` - Migration generation
- `verify/` - Migration verification

**Features:**
- Schema detection
- Diff generation
- Migration execution
- Rollback support

### 10. HTTP & Server (`forge/server/`)

**Purpose:** HTTP server and routing

**Components:**
- `server.go` - Server wrapper
- `router.go` - Router setup
- `middleware.go` - Middleware stack
- `context.go` - Request context
- `response.go` - Response helpers
- `security.go` - Security middleware
- `static.go` - Static file serving
- `health.go` - Health checks
- `ratelimit.go` - Rate limiting

**Middleware Stack:**
1. Request ID
2. Real IP
3. Recoverer
4. Logger
5. Session
6. CSRF
7. Authentication

### 11. Logging System (`forge/log/`)

**Purpose:** Structured logging

**Components:**
- `logger.go` - Logger implementation
- `config.go` - Logging configuration
- `encoder.go` - Log encoders
- `exporters/` - Log exporters (console, file, remote)
- `hooks/` - Log hooks
- `middleware.go` - Request logging

**Features:**
- Multiple output formats (JSON, console)
- Multiple outputs (console, file, remote)
- Log levels
- Sampling in production
- Contextual logging

### 12. Configuration (`forge/config/`)

**Purpose:** Application configuration

**Components:**
- `config.go` - Viper wrapper
- `settings.go` - Framework settings
- `logging_settings.go` - Logging settings

**Features:**
- YAML, JSON, env var support
- Hierarchical configuration
- Default values

### 13. Validation (`forge/validate/`)

**Purpose:** Data validation

**Components:**
- Validator wrapper (go-playground/validator)
- Validation tag helpers
- Schema integration

### 14. Security (`forge/security/`)

**Purpose:** Security features

**Components:**
- CSRF protection
- XSS protection
- SQL injection prevention
- Input sanitization

### 15. CLI (`forge/cli/`)

**Purpose:** Command-line interface

**Components:**
- `root.go` - CLI root
- `commands/` - CLI commands
- `templates/` - Project templates
- `core/` - Core CLI functionality

**Commands:**
- `new` - Create new project
- `generate` - Generate code
- `migrate` - Run migrations
- `runserver` - Start server
- And more...

## Data Flow

### Request Flow

```
HTTP Request
  ↓
Chi Router
  ↓
Middleware Stack
  ↓
Handler (Admin/API/View)
  ↓
QuerySet / Manager
  ↓
SQL Builder
  ↓
Database
  ↓
Response (JSON/HTML)
```

### Code Generation Flow

```
Schema Definition (Go)
  ↓
AST Parser
  ↓
Schema Analysis
  ↓
Code Generator
  ↓
Generated Files:
  - Model structs
  - FieldExpr
  - Manager/QuerySet
```

### Query Execution Flow

```
QuerySet Method Call
  ↓
Expression Building
  ↓
SQL Builder
  ↓
Parameter Binding
  ↓
Database Execution
  ↓
Row Scanning
  ↓
Model Instances
```

## Extension Points

### 1. Model Extensions
- Add fields to existing models
- Add relations
- Add hooks
- Override methods

### 2. Admin Extensions
- Customize list display
- Add filters
- Add actions
- Custom widgets
- Custom forms

### 3. API Extensions
- Custom serializers
- Custom viewsets
- Custom authentication
- Custom permissions
- Custom throttling

### 4. Query Extensions
- Custom QuerySet methods
- Custom aggregates
- Custom annotations
- Custom expressions

### 5. Filter Extensions
- Custom filters
- Custom widgets
- Custom parsers

### 6. Middleware
- Custom middleware injection
- Request/response modification

### 7. Plugin System
- Register plugins
- Plugin lifecycle hooks
- Plugin dependencies

## Technology Stack

### Core
- **Go 1.25+** - Programming language
- **database/sql** - Database interface
- **go/ast** - Code generation

### HTTP & Routing
- **chi/v5** - HTTP router
- **chi/middleware** - Middleware

### Database
- **database/sql** - Standard library SQL interface
- **lib/pq** - PostgreSQL driver
- **golang-migrate/v4** - Migrations

### Security
- **gorilla/csrf** - CSRF protection
- **alexedwards/scs/v2** - Sessions
- **golang.org/x/crypto/bcrypt** - Password hashing

### Validation & Configuration
- **go-playground/validator/v10** - Validation
- **spf13/viper** - Configuration

### Logging
- **zap** - Structured logging

### Utilities
- **stretchr/testify** - Testing
- **iancoleman/strcase** - String utilities
- **google/uuid** - UUID generation

## Concurrency & Thread Safety

### Thread Safety
- Database connections are thread-safe
- QuerySets are immutable (return new instances)
- Managers are thread-safe
- Admin registry is thread-safe

### Concurrency Patterns
- Context-based cancellation
- Connection pooling
- Transaction isolation

## Performance Considerations

### Optimization Strategies
- Query optimization
- Connection pooling
- Lazy loading
- Caching (where implemented)
- Efficient data structures

### Bottlenecks
- Database queries (primary bottleneck)
- Code generation (one-time cost)
- Reflection usage (minimized)

## Memory Management

### Allocation Patterns
- Minimal allocations in hot paths
- Reusable buffers
- Efficient data structures
- Connection pooling

### Garbage Collection
- Designed to minimize GC pressure
- Efficient data structures
- Proper cleanup of resources
