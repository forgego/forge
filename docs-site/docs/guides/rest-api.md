---
sidebar_position: 4
description: Build REST APIs with forge's ViewSet framework. Complete guide to serializers, viewsets, authentication, permissions, and API best practices.
keywords:
  - forge rest api
  - forge api
  - go rest api
  - forge viewset
  - forge serializer
  - api framework
image: /forge-social-card.svg
---

# REST API

forge's REST API framework is like Django REST Framework for Go. Instead of writing HTTP handlers for every endpoint, you use ViewSets and get CRUD operations, pagination, and filtering for free.

## Why use it?

Because writing API endpoints is repetitive:

- **Automatic CRUD** - Get all the endpoints without writing them
- **Pagination** - Built in, works out of the box
- **Filtering** - Filter by any field via query params
- **Serialization** - Control what fields get exposed
- **Authentication** - Token, JWT, session—pick what you need
- **Permissions** - Flexible permission system

## Overview

The REST API is built on the ViewSet pattern, providing:

- Full CRUD operations (Create, Read, Update, Delete)
- Automatic pagination
- Filtering and ordering
- Serialization
- Error handling

## Basic Usage

### 1. Define a Serializer

Create a serializer for your model:

```go
package models

import (
    "github.com/forgego/forge/pkg/api"
)

// UserSerializer serializes User model
type UserSerializer struct {
    *api.BaseSerializer
}

func NewUserSerializer() api.Serializer {
    return &UserSerializer{
        BaseSerializer: api.NewBaseSerializer(),
    }
}

func (s *UserSerializer) Fields() []string {
    return []string{"id", "username", "email", "is_active"}
}

func (s *UserSerializer) ReadOnlyFields() []string {
    return []string{"id", "date_joined"}
}
```

### 2. Create a ViewSet

```go
package api

import (
    "myapp/models"
    "github.com/forgego/forge/pkg/api"
)

func RegisterUserViewSet(router *api.Router) {
    viewset := api.NewBaseViewSet(
        func() api.Serializer {
            return models.NewUserSerializer()
        },
        models.User.Objects.Filter(), // QuerySet
        &models.User{},                // Model instance
    )

    router.Register("users", viewset)
}
```

### 3. Register Routes

In your `main.go`:

```go
import (
    "github.com/forgego/forge/pkg/api"
    httplib "github.com/forgego/forge/pkg/http"
)

func main() {
    // ... setup code ...

    // Create API router
    apiRouter := api.NewRouter("/api/v1")

    // Register viewsets
    api.RegisterUserViewSet(apiRouter)
    api.RegisterPostViewSet(apiRouter)

    // Register on HTTP router
    server.RegisterRoutes(func(router *httplib.Router) {
        apiRouter.RegisterRoutes(router)
    })
}
```

## API Endpoints

Once registered, your ViewSet automatically provides these endpoints:

### List (GET /api/v1/users/)

Returns paginated list of users.

**Query Parameters:**
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 20)
- `search` - Search query
- `ordering` - Order by field (use `-` for descending)
- `is_active` - Filter by field
- `username__contains` - Field lookup filters

**Response:**
```json
{
  "count": 100,
  "next": "http://example.com/api/v1/users/?page=2",
  "previous": null,
  "results": [
    {
      "id": 1,
      "username": "john",
      "email": "john@example.com",
      "is_active": true
    }
  ]
}
```

### Detail (GET /api/v1/users/\{id\}/)

Returns a single user.

**Response:**
```json
{
  "id": 1,
  "username": "john",
  "email": "john@example.com",
  "is_active": true
}
```

### Create (POST /api/v1/users/)

Creates a new user.

**Request:**
```json
{
  "username": "jane",
  "email": "jane@example.com",
  "is_active": true
}
```

**Response:**
```json
{
  "id": 2,
  "username": "jane",
  "email": "jane@example.com",
  "is_active": true
}
```

### Update (PUT /api/v1/users/\{id\}/)

Updates a user (full update).

**Request:**
```json
{
  "username": "jane_updated",
  "email": "jane@example.com",
  "is_active": true
}
```

### Partial Update (PATCH /api/v1/users/\{id\}/)

Partially updates a user.

**Request:**
```json
{
  "is_active": false
}
```

### Delete (DELETE /api/v1/users/\{id\}/)

Deletes a user.

**Response:** 204 No Content

## Pagination

Pagination is automatic. Use query parameters:

```
GET /api/v1/users/?page=2&page_size=50
```

**Response Format:**
```json
{
  "count": 100,
  "next": "http://example.com/api/v1/users/?page=3&page_size=50",
  "previous": "http://example.com/api/v1/users/?page=1&page_size=50",
  "results": [...]
}
```

## Filtering

Filter by field values:

```
GET /api/v1/users/?is_active=true
GET /api/v1/users/?username__contains=john
GET /api/v1/users/?date_joined__gte=2024-01-01
```

### Filter Lookups

- `exact` - Exact match (default)
- `iexact` - Case-insensitive exact match
- `contains` - Contains substring
- `icontains` - Case-insensitive contains
- `startswith` - Starts with
- `endswith` - Ends with
- `gt` - Greater than
- `gte` - Greater than or equal
- `lt` - Less than
- `lte` - Less than or equal
- `in` - Value in list
- `isnull` - Is null
- `range` - Range filter

## Ordering

Use the `ordering` parameter:

```
GET /api/v1/users/?ordering=-date_joined,username
```

Use `-` prefix for descending order.

## Search

Full-text search via `search` parameter:

```
GET /api/v1/users/?search=john
```

Searches across fields specified in the serializer's `SearchFields()`.

## Custom Serializers

### Custom Field Serialization

```go
type UserSerializer struct {
    *api.BaseSerializer
}

func (s *UserSerializer) SerializeField(field string, value interface{}) interface{} {
    switch field {
    case "password":
        return nil // Never serialize password
    case "date_joined":
        return value.(time.Time).Format(time.RFC3339)
    default:
        return value
    }
}
```

### Nested Serialization

```go
type PostSerializer struct {
    *api.BaseSerializer
}

func (s *PostSerializer) Fields() []string {
    return []string{"id", "title", "content", "author", "created_at"}
}

func (s *PostSerializer) NestedFields() map[string]api.Serializer {
    return map[string]api.Serializer{
        "author": NewUserSerializer(),
    }
}
```

## Custom Viewsets

Create custom viewsets for complex logic:

```go
type PostViewSet struct {
    *api.BaseViewSet
}

func NewPostViewSet() *PostViewSet {
    return &PostViewSet{
        BaseViewSet: api.NewBaseViewSet(
            NewPostSerializer,
            Post.Objects.Filter(),
            &Post{},
        ),
    }
}

func (vs *PostViewSet) Create(w http.ResponseWriter, r *http.Request) {
    // your custom logic here
}

func (vs *PostViewSet) List(w http.ResponseWriter, r *http.Request) {
    // your custom filtering logic here
}
```

## Permissions

Add permission checks:

```go
viewset := api.NewBaseViewSet(...)
viewset.SetPermissions(
    api.RequireAuthenticated(),
    api.RequirePermission("can_view_user"),
)
```

## Authentication

Add authentication to your API:

```go
apiRouter := api.NewRouter("/api/v1")
apiRouter.Use(auth.RequireAuth())
```

## Error Handling

API errors are returned in a consistent format:

```json
{
  "error": "Validation failed",
  "details": {
    "username": ["This field is required."],
    "email": ["Enter a valid email address."]
  }
}
```

## Best Practices

1. **Use Serializers** - Always use serializers to control API output
2. **Paginate Lists** - Always paginate list endpoints
3. **Filter Appropriately** - Add filters for commonly queried fields
4. **Validate Input** - Use model validation in serializers
5. **Handle Errors** - Return consistent error responses
6. **Use HTTP Status Codes** - Use appropriate status codes (200, 201, 400, 404, 500)

## Examples

### Complete API Setup

```go
package main

import (
    "github.com/forgego/forge/pkg/api"
    httplib "github.com/forgego/forge/pkg/http"
    "myapp/api"
)

func main() {
    // ... your setup code ...

    apiRouter := api.NewRouter("/api/v1")
    api.RegisterUserViewSet(apiRouter)
    api.RegisterPostViewSet(apiRouter)
    api.RegisterCategoryViewSet(apiRouter)

    server.RegisterRoutes(func(router *httplib.Router) {
        apiRouter.RegisterRoutes(router)
    })
}
```

## Next Steps

- [Queries Guide](/docs/guides/queries) - Learn about QuerySet filtering
- [Security Guide](/docs/guides/security) - Secure your API
- [Advanced Topics](/docs/advanced/plugins) - Extend the API system

