---
id: api-framework
title: API Framework
sidebar_label: API Framework
---

# API Framework Feature

## Overview

The API Framework provides a DRF-like REST API framework with serializers, viewsets, authentication, permissions, throttling, and content negotiation.

## Location

`forge/api/`

## Status

✅ **Complete** - Production ready

## Core Components

### 1. ViewSet (`viewset.go`)

Base ViewSet interface with CRUD operations:

```go
type ViewSet interface {
    List(w http.ResponseWriter, r *http.Request)
    Create(w http.ResponseWriter, r *http.Request)
    Retrieve(w http.ResponseWriter, r *http.Request)
    Update(w http.ResponseWriter, r *http.Request)
    PartialUpdate(w http.ResponseWriter, r *http.Request)
    Destroy(w http.ResponseWriter, r *http.Request)
}
```

### 2. Serializer (`serializer.go`)

Complete serializer system with field types:

**Field Types:**
- String, Integer, Float, Boolean
- DateTime, Date, Time
- Email, URL, UUID
- JSON, Choice
- Nested, List, Dict
- File, Image

### 3. Authentication (`authentication/`)

Multiple authentication backends:
- Token authentication
- JWT authentication
- Basic authentication
- Session authentication
- API Key authentication

### 4. Permissions (`permissions/`)

Permission classes:
- AllowAny
- IsAuthenticated
- IsAdminUser
- IsOwnerOrReadOnly
- Custom permissions

### 5. Throttling (`throttling/`)

Rate limiting:
- AnonRateThrottle
- UserRateThrottle
- ScopedRateThrottle

### 6. Renderers (`renderers/`)

Response renderers:
- JSON
- XML
- YAML
- HTML
- CSV

### 7. Parsers (`parsers/`)

Request parsers:
- JSON
- XML
- Form
- MultiPart

### 8. Filters (`filters/`)

Field filtering and search

### 9. Pagination (`pagination.go`)

Pagination classes:
- PageNumber
- LimitOffset

### 10. Exceptions (`exceptions/`)

Exception handling hierarchy

### 11. Versioning (`versioning/`)

API versioning support

### 12. Caching (`caching/`)

Cache backends:
- Memory cache
- Redis cache (planned)

### 13. OpenAPI (`docs/`)

OpenAPI documentation generation

## Features

### ✅ Complete Features

1. **ViewSets** - Complete CRUD operations
2. **Serializers** - Full serializer system
3. **Authentication** - Multiple auth backends
4. **Permissions** - Complete permission system
5. **Throttling** - Rate limiting
6. **Renderers** - Multiple response formats
7. **Parsers** - Multiple request formats
8. **Filters** - Field filtering and search
9. **Pagination** - PageNumber and LimitOffset
10. **Exceptions** - Exception handling
11. **Versioning** - API versioning
12. **Caching** - Cache backends
13. **OpenAPI** - Documentation generation

## Usage Examples

### Basic ViewSet

```go
type UserViewSet struct {
    *api.BaseViewSet
}

func NewUserViewSet() *UserViewSet {
    return &UserViewSet{
        BaseViewSet: api.NewBaseViewSet(
            func() api.Serializer { return NewUserSerializer() },
            User.Objects,
            &User{},
        ),
    }
}
```

### Serializer

```go
type UserSerializer struct {
    *api.Serializer
}

func NewUserSerializer() *UserSerializer {
    s := &UserSerializer{
        Serializer: api.NewSerializer(),
    }
    s.AddField("id", api.IntegerField())
    s.AddField("username", api.StringField())
    s.AddField("email", api.EmailField())
    return s
}
```

### Authentication

```go
router.Use(api.TokenAuthenticationMiddleware)
```

### Permissions

```go
viewset.SetPermissionClasses([]api.Permission{
    api.IsAuthenticated(),
})
```

### Throttling

```go
viewset.SetThrottleClasses([]api.Throttle{
    api.UserRateThrottle(100, time.Minute),
})
```

## Integration Points

### ORM System
- ViewSets use ORM QuerySet
- Serializers work with ORM models
- Filter integration

### Identity System
- Authentication integration
- Permission integration
- User serializers

### Filter System
- Filter integration
- Query parsing
- Security validation

## Extension Points

### Custom ViewSets
- Extend BaseViewSet
- Custom actions
- Custom logic

### Custom Serializers
- Custom field types
- Custom validation
- Custom serialization

### Custom Authentication
- Custom auth backends
- Custom token validation

### Custom Permissions
- Custom permission classes
- Custom permission logic

## Best Practices

1. **Type Safety** - Use type-safe serializers
2. **Permissions** - Always check permissions
3. **Validation** - Validate all input
4. **Error Handling** - Proper error responses
5. **Versioning** - Use API versioning
6. **Documentation** - Document all endpoints

## Future Enhancements

- [ ] Auto-generate ViewSets from models
- [ ] Auto-generate Serializers from models
- [ ] Auto-register routes
- [ ] Interactive API documentation
- [ ] GraphQL support
