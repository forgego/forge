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

forge provides a comprehensive set of features that make building web applications in Go as productive as Django. All features are type-safe and work together seamlessly.

## Core Features

### 🏗️ Schema System
Define your models declaratively with type safety:
- **15+ Field Types** - Int64, String, Bool, Time, Date, DateTime, Float32/64, Decimal, Text, Email, URL, UUID, JSON, Bytes
- **Complete Field Options** - Required, Blank, Default, Primary, AutoIncrement, Unique, Index, MaxLength, Choices, and more
- **Relationships** - ForeignKey, OneToOne, ManyToMany, OneToMany with full type safety
- **Model Metadata** - TableName, ordering, permissions, constraints
- **Lifecycle Hooks** - BeforeSave, AfterSave, BeforeCreate, AfterCreate, BeforeUpdate, AfterUpdate, BeforeDelete, AfterDelete

### 🔧 Code Generation
AST-based code generation that eliminates boilerplate:
- **Model Structs** - Generated with proper types and database tags
- **FieldExpr** - Type-safe field accessors for queries
- **Managers** - CRUD operations with proper error handling
- **QuerySets** - Type-safe query building and execution
- **Import Management** - Automatic import handling and formatting

### 🗄️ Type-Safe ORM
Django-like ORM with Go's type safety:
- **QuerySet API** - Filter, Exclude, OrderBy, Limit, Offset, Distinct
- **Query Methods** - All, Get, First, Last, Count, Exists
- **Type Safety** - Compile-time checking of field names and types
- **SQL Builder** - Proper parameter binding and SQL injection prevention
- **Manager Operations** - Create, Update, Delete with hooks

## Admin Interface

### 🎛️ Auto-Generated Admin
Django-style admin interface that works out of the box:
- **Type-Safe Admin[T]** - Generic-based admin interface
- **Complete Config[T]** - All Django admin options supported
- **Automatic Registration** - Register models and get full CRUD

### 📋 List Views
Rich list views with advanced features:
- **Pagination** - Efficient pagination with customizable page sizes
- **Search** - Full-text search across multiple fields
- **Filtering** - Advanced filtering with multiple filter types
- **Sorting** - Multi-column sorting with proper handling
- **Actions** - Bulk operations with custom actions

### 📝 Form Views
Powerful create and edit forms:
- **Auto-Generated Forms** - Created from model definitions
- **Validation** - Server-side and client-side validation
- **Widgets** - Rich form widgets for all field types
- **Fieldsets** - Organize fields into logical groups
- **Inlines** - Edit related models inline

### 🎨 Widget System
Rich form widgets for every need:
- **Basic Widgets** - Text, Textarea, Select, Checkbox, Radio
- **Date/Time Widgets** - Date, Time, DateTime pickers
- **File Widgets** - File and Image upload with preview
- **Rich Text** - WYSIWYG editor integration
- **Advanced Widgets** - Autocomplete, RawID, custom widgets

### 🔍 Filter System
Advanced filtering capabilities:
- **Filter Types** - Boolean, Choice, Date, Number, Text, Related
- **Custom Filters** - Create custom filter logic
- **Filter Widgets** - Appropriate UI widgets for each filter type
- **Security** - Field whitelist/blacklist, lookup restrictions

## REST API Framework

### 🚀 BaseViewSet
Complete CRUD operations out of the box:
- **Standard CRUD** - Create, Read, Update, Delete, List
- **Type Safety** - Generic-based with compile-time checking
- **Customization** - Override any method for custom behavior
- **Integration** - Works seamlessly with QuerySet and Manager

### 📦 Serializer System
Flexible data serialization:
- **Field Serializers** - Serializers for all field types
- **Validation** - Automatic validation with custom validators
- **Nested Serializers** - Handle complex object relationships
- **Custom Serialization** - Create custom serialization logic

### 🔐 Authentication & Authorization
Complete auth system:
- **Multiple Backends** - Token, JWT, Basic, Session, API Key
- **Permission Classes** - AllowAny, IsAuthenticated, IsAdminUser, IsOwnerOrReadOnly
- **Custom Auth** - Create custom authentication backends
- **Integration** - Works with identity system

### ⚡ Performance Features
Built-in performance optimizations:
- **Throttling** - AnonRateThrottle, UserRateThrottle, ScopedRateThrottle
- **Pagination** - PageNumber and LimitOffset pagination
- **Caching** - Built-in cache backends
- **Filtering** - Efficient field filtering
- **Versioning** - API versioning support

### 📚 Documentation
Auto-generated API documentation:
- **OpenAPI/Swagger** - Automatic API spec generation
- **Interactive Docs** - Built-in API explorer
- **Schema Documentation** - Auto-generated from serializers

## Advanced Systems

### 🧠 Filter System
AST-based filtering with advanced features:
- **Complex Filters** - AND, OR, NOT operations with nesting
- **Query Parser** - Multiple filter formats (JSON, query string)
- **AST Representation** - Abstract syntax tree for complex filters
- **Security Validation** - Prevent injection attacks
- **Query Optimization** - Automatic query optimization
- **Persistence** - Save and share filter configurations

### 👤 Identity System
Complete user management:
- **User Models** - Django-compatible user model with extensions
- **Authentication** - Multiple authentication methods
- **Sessions** - Secure session management
- **Permissions** - Fine-grained permission system
- **Groups** - User groups with role-based access
- **Tokens** - API token management
- **Password Management** - Secure password handling with policies

### 🗃️ Database Layer
Robust database handling:
- **Connection Pooling** - Efficient connection management
- **Transactions** - Full transaction support with savepoints
- **Migration Integration** - Seamless migration system integration
- **Multiple Drivers** - PostgreSQL, SQLite support
- **Health Checks** - Database health monitoring

### 🔄 Migration System
Powerful migration system:
- **AST-Based Detection** - Automatic schema change detection
- **Change Types** - 15+ different change types supported
- **SQL Generation** - Multi-dialect SQL generation
- **Rollback Support** - Safe rollback capabilities
- **Verification** - Migration verification and drift detection
- **Safety Checks** - Pre-migration validation

## Infrastructure

### 🌐 HTTP Server
Production-ready HTTP server:
- **Chi Router** - Fast, flexible HTTP router
- **Middleware Stack** - Complete middleware for production
- **Security** - CSRF, XSS, injection protection
- **Static Files** - Efficient static file serving
- **Health Checks** - Application health monitoring
- **Rate Limiting** - Built-in rate limiting

### 📝 Logging System
Structured logging for production:
- **Zap Integration** - High-performance structured logging
- **Multiple Outputs** - Console, file, remote logging
- **Log Levels** - Configurable log levels
- **Contextual Logging** - Request tracing and context
- **Sampling** - Production log sampling

### ⚙️ Configuration
Flexible configuration management:
- **Multiple Sources** - YAML, JSON, environment variables
- **Hierarchical** - Environment-specific overrides
- **Type Safety** - Typed configuration with validation
- **Defaults** - Sensible defaults everywhere

### 🔒 Security
Security by default:
- **CSRF Protection** - Built-in CSRF protection
- **XSS Prevention** - Input sanitization and output encoding
- **SQL Injection Prevention** - Parameter binding everywhere
- **Input Validation** - Comprehensive input validation

## Developer Experience

### 🛠️ CLI Tools
Complete command-line interface:
- **Project Creation** - `forge new` creates new projects
- **Code Generation** - `forge generate` updates generated code
- **Migrations** - `forge migrate` runs database migrations
- **Development Server** - `forge runserver` for development
- **Additional Commands** - Database management, testing, and more

### 🧪 Testing
Built-in testing support:
- **Test Helpers** - Utilities for testing forge applications
- **Mock Support** - Easy mocking of database operations
- **Integration Tests** - Database testing utilities
- **Test Database** - Automatic test database setup

### 🔌 Extensibility
Everything can be extended:
- **Plugin System** - Register plugins at startup
- **Hook System** - Lifecycle hooks for customization
- **Custom Fields** - Create custom field types
- **Custom Widgets** - Build custom form widgets
- **Middleware** - Add custom middleware
- **Authentication** - Custom authentication backends

## Technology Stack

### Core Technologies
- **Go 1.25+** - Latest Go with generics
- **PostgreSQL** - Primary database with full feature support
- **Chi v5** - Fast HTTP router
- **Zap** - High-performance logging

### Key Libraries
- **database/sql** - Standard database interface
- **golang-migrate** - Migration system
- **viper** - Configuration management
- **gorilla/csrf** - CSRF protection
- **alexedwards/scs** - Session management
- **go-playground/validator** - Validation
- **golang.org/x/crypto** - Cryptographic functions

## Status

All features listed above are **fully implemented and production ready**. The MVP is complete and you can build real applications with forge today.

### What's Next?
See the [Roadmap](/docs/status/roadmap/) for planned future features and enhancements.

### Ready to Start?
- [Installation Guide](/docs/getting-started/installation)
- [Quick Start](/docs/getting-started/quickstart)
- [Examples](/docs/examples/blog)
