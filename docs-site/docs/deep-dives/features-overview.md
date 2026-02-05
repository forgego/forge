---
sidebar_position: 2
description: Complete overview of all forge framework features. Learn what's available, what's complete, and what's planned.
keywords:
  - forge features
  - feature list
  - framework features
  - complete features
image: /forge-social-card.svg
---

# Features Overview

This document provides an overview of all features in the forge framework. For detailed documentation on each feature, see the individual feature documents.

## Core Features

### 1. Schema Definition System
**Status:** ✅ Complete  
**Location:** `forge/schema/`  
**Documentation:** [Schema System Deep Dive](/docs/deep-dives/schema-system/)

Declarative model definitions with full Django field options, relationships, metadata, and lifecycle hooks.

### 2. Code Generation
**Status:** ✅ Complete  
**Location:** `forge/codegen/`  
**Documentation:** [Code Generation Guide](/docs/features/code-generation/)

AST-based code generation for models, managers, querysets, and field expressions. Generates type-safe Go code from schema definitions.

### 3. Type-Safe ORM
**Status:** ✅ Complete  
**Location:** `forge/orm/`  
**Documentation:** [ORM System Guide](/docs/features/orm-system/)

Complete Django-like ORM with type-safe queries, managers, field expressions, and SQL builder.

### 4. Admin System
**Status:** ✅ Complete  
**Location:** `forge/admin/`  
**Documentation:** [Admin System Guide](/docs/features/admin-system/)

Django-style admin interface with type-safe configuration, CRUD operations, filters, widgets, and export.

### 5. REST API Framework
**Status:** ✅ Complete  
**Location:** `forge/api/`  
**Documentation:** [API Framework Guide](/docs/features/api-framework/)

DRF-like REST API framework with serializers, viewsets, authentication, permissions, throttling, and more.

### 6. Filter System
**Status:** ✅ Complete  
**Location:** `forge/filter/`  
**Documentation:** [Filter System Guide](/docs/features/filter-system/)

Advanced filtering system with AST support, query parsing, expression conversion, security validation, and query optimization.

### 7. Identity System
**Status:** ✅ Complete  
**Location:** `forge/identity/`  
**Documentation:** [Identity System Guide](/docs/features/identity-system/)

Complete user management and authentication system with repositories, services, authentication backends, permissions, sessions, password management, and HTTP handlers.

### 8. Database Layer
**Status:** ✅ Complete  
**Location:** `forge/db/`  
**Documentation:** [Database Layer Guide](/docs/features/database-layer/)

Database connection management with pooling, transaction support with savepoints, and migration integration.

### 9. Migration System
**Status:** ✅ Complete  
**Location:** `forge/db/migrate/` and `forge/migrate/`  
**Documentation:** [Migration System Guide](/docs/features/migration-system/)

Complete migration system with schema detection, diff generation, SQL generation, execution, rollback, state management, and drift detection.

### 10. HTTP & Server
**Status:** ✅ Complete  
**Location:** `forge/server/`

HTTP server with routing, middleware stack, security, static files, and health checks.

### 11. Logging System
**Status:** ✅ Complete  
**Location:** `forge/log/`

Structured logging with multiple outputs, formats, levels, and exporters.

### 12. Configuration System
**Status:** ✅ Complete  
**Location:** `forge/config/`

Application configuration with YAML, JSON, and environment variable support.

### 13. Validation System
**Status:** ✅ Complete  
**Location:** `forge/validate/`

Data validation with go-playground/validator integration and schema support.

### 14. Security System
**Status:** ✅ Complete  
**Location:** `forge/security/`

Security features including CSRF protection, XSS prevention, and SQL injection prevention.

### 15. CLI Tools
**Status:** ✅ Complete  
**Location:** `forge/cli/`

Command-line interface for project creation, code generation, migrations, and server management.

## Feature Status Summary

### ✅ Complete Features (15)

| Feature | Status | Core Functionality |
|---------|--------|-------------------|
| Schema System | ✅ Complete | Full Django field options, relations, meta, hooks |
| Code Generation | ✅ Complete | AST parsing, model/manager/queryset generation |
| ORM System | ✅ Complete | QuerySet API, Manager CRUD, SQL builder |
| Admin System | ✅ Complete | Type-safe admin, CRUD, filters, widgets, export |
| API Framework | ✅ Complete | ViewSets, Serializers, Auth, Permissions, Throttling |
| Filter System | ✅ Complete | AST-based filtering, query parsing, security |
| Identity System | ✅ Complete | User management, auth, sessions, permissions |
| Database Layer | ✅ Complete | Connections, transactions, migrations |
| Migration System | ✅ Complete | Schema detection, diff generation, execution |
| HTTP Server | ✅ Complete | Routing, middleware, security, static files |
| Logging System | ✅ Complete | Structured logging, multiple outputs |
| Configuration | ✅ Complete | YAML, JSON, env vars, hierarchical |
| Validation | ✅ Complete | Validator integration, schema support |
| Security | ✅ Complete | CSRF, XSS, SQL injection prevention |
| CLI Tools | ✅ Complete | Project creation, code generation, migrations |

### 🚧 In Progress / Partial

| Feature | Status | Notes |
|---------|--------|-------|
| Advanced ORM | 🚧 Partial | SelectRelated/PrefetchRelated structure ready |
| Aggregates | 🚧 Partial | Structure ready, implementation in progress |
| Annotations | 🚧 Partial | Structure ready, implementation in progress |

### 📋 Planned Features

See the [Roadmap](/docs/status/roadmap/) for detailed planned features.

## Feature Dependencies

```
Schema System
    ↓
Code Generation
    ↓
ORM System ←→ Database Layer
    ↓              ↓
Admin System    Migration System
    ↓
API Framework
    ↓
Filter System
    ↓
Identity System
    ↓
HTTP Server ←→ Logging System
    ↓              ↓
Configuration ←→ Security
    ↓
CLI Tools
```

## Feature Integration

### Admin + ORM
- Admin uses ORM QuerySet for data access
- Type-safe field expressions in admin
- Filter integration with ORM

### API + ORM
- ViewSets use ORM for data access
- Serializers work with ORM models
- Filter integration with API

### Filter + ORM
- Filters convert to ORM expressions
- AST-based query parsing
- Security validation

### Identity + Admin
- Permission system integration
- User management in admin
- Authentication middleware

### Identity + API
- Authentication backends
- Permission classes
- User serializers

## Feature Extensibility

All features support extension:

- **Schema System**: Custom fields, relations, hooks
- **Code Generation**: Custom templates, generators
- **ORM System**: Custom QuerySet methods, expressions
- **Admin System**: Custom widgets, filters, actions
- **API Framework**: Custom serializers, viewsets, auth
- **Filter System**: Custom filters, widgets, parsers
- **Identity System**: Custom backends, services
- **HTTP Server**: Custom middleware, handlers
- **Logging System**: Custom exporters, hooks
- **CLI Tools**: Custom commands, templates

## Feature Testing

Each feature has:
- Unit tests
- Integration tests (where applicable)
- Example usage
- Documentation

## Feature Documentation

For detailed documentation on each feature, see the Deep Dives and Guides:
- [Architecture](/docs/deep-dives/architecture/) - Complete system architecture
- [Design Principles](/docs/deep-dives/design-principles/) - Framework design principles
- [Schema System](/docs/deep-dives/schema-system/) - Schema definition system
- [Code Generation](/docs/features/code-generation/) - Code generation from schemas
- [ORM System](/docs/features/orm-system/) - Type-safe ORM system
- [Admin System](/docs/features/admin-system/) - Django-like admin interface
- [API Framework](/docs/features/api-framework/) - DRF-like REST API framework
- [Filter System](/docs/features/filter-system/) - Advanced filtering system
- [Identity System](/docs/features/identity-system/) - User management and authentication
- [Database Layer](/docs/features/database-layer/) - Database connection and transactions
- [Migration System](/docs/features/migration-system/) - Database migration system