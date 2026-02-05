---
sidebar_position: 2
---

# Architecture

Overview of forge framework architecture.

## Core Philosophy

1. **Type-Safe First**: Primary API uses generics for compile-time safety
2. **Dynamic When Needed**: Secondary API for runtime flexibility
3. **Convention over Configuration**: Sensible defaults everywhere
4. **Fully Extensible**: Everything can be extended/overridden
5. **Security by Default**: Built-in protections
6. **Code Generation**: AST-based generation for type-safe code

## Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│                    User Application                      │
│  (Schema Definitions, Models, Views, Controllers)        │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              Code Generation Layer                        │
│  (AST Parser → SQL Generation → Go Code Generation)      │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              Framework API Layer                          │
│  (QuerySet, Manager, FieldExpr, QueryExpr)               │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              Database Layer                               │
│  (SQL Builder, Parameter Binding, Transactions)          │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              Infrastructure Layer                         │
│  (HTTP Router, Middleware, Security, Config, Logging)     │
└─────────────────────────────────────────────────────────┘
```

## Component Architecture

### Schema Definition System

**Location:** `pkg/schema/`

**Purpose:** Define models declaratively in Go code

**Components:**
- `schema.go` - Core Schema interface
- `field.go` - Field definitions with builders
- `relation.go` - Relationship definitions
- `meta.go` - Model metadata
- `hooks.go` - Lifecycle hooks

### Code Generation System

**Location:** `pkg/generator/`

**Purpose:** Generate type-safe Go code from schema definitions

**Components:**
- `generator.go` - Main generation orchestrator
- `ast_parser.go` - Go AST parser for schema extraction
- `writer.go` - Code file writer
- `templates/` - Code generation templates

**Generated Files:**
- `models/*.gen.go` - Model structs
- `models/*_fields.gen.go` - FieldExpr definitions
- `models/*_manager.gen.go` - Manager with CRUD operations
- `models/*_queryset.gen.go` - Type-safe QuerySet wrapper

### Query System

**Location:** `pkg/query/`

**Purpose:** Type-safe and dynamic query building

**Components:**
- `field_expr.go` - Type-safe field accessors (`FieldExpr[T]`)
- `query_expr.go` - Query conditions (`QueryExpr`)
- `queryset.go` - QuerySet implementation with `BaseQuerySet[T]`
- `sql_builder.go` - SQL generation with proper escaping
- `aggregates.go` - Aggregate functions
- `annotations.go` - Annotation support

### Database Layer

**Location:** `pkg/db/`

**Purpose:** Database connection and transaction management

**Components:**
- `db.go` - Database connection wrapper
- `transaction.go` - Transaction management
- `migrations.go` - Migration support

### HTTP & Routing

**Location:** `pkg/http/`

**Purpose:** HTTP server and routing

**Components:**
- `router.go` - Chi router wrapper
- `server.go` - HTTP server
- `middleware.go` - Middleware stack
- `context.go` - Request context utilities
- `response.go` - Response helpers

### Admin System

**Location:** `pkg/admin/`

**Purpose:** Auto-generated admin interface

**Components:**
- `registry.go` - Model registry
- `router.go` - Admin routes
- `list.go` - List view
- `forms.go` - Form handling
- `templates/` - Admin templates

### Security

**Location:** `pkg/security/`

**Purpose:** Security features

**Components:**
- `csrf.go` - CSRF protection
- `sessions.go` - Session management
- `xss.go` - XSS protection
- `sql_injection.go` - SQL injection prevention

## Data Flow

### Request Flow

1. HTTP request arrives
2. Middleware processes (auth, CSRF, logging)
3. Router matches route
4. Handler executes
5. QuerySet/Manager queries database
6. SQL builder generates SQL
7. Database executes query
8. Results mapped to models
9. Response returned

### Code Generation Flow

1. User defines schema in Go
2. `forge generate` command runs
3. AST parser reads model files
4. Schema information extracted
5. Templates generate code
6. Generated files written to `models/*.gen.go`

## Extension Points

### Plugins

Plugins can hook into:
- Model registration
- Request/response lifecycle
- Code generation
- Admin customization

### Custom Fields

Create custom field types:
- Implement `Field` interface
- Add validation
- Define SQL type

### Custom Templates

Customize code generation templates (advanced)

## Technology Stack

- **Go 1.21+** - Programming language
- **chi/v5** - HTTP router
- **database/sql** - Database interface
- **golang-migrate** - Migrations
- **zap** - Logging
- **viper** - Configuration
- **testify** - Testing

## Design Principles

1. **Type Safety** - Compile-time checking wherever possible
2. **Performance** - Efficient queries and minimal overhead
3. **Developer Experience** - Easy to use, hard to misuse
4. **Extensibility** - Everything can be extended
5. **Security** - Secure by default
6. **Standards** - Follow Go and web standards

## See Also

- [Development Guide](/docs/contributing/development) - Contributing guide
- [Implementation status](/docs/status/implementation) - Current build status
