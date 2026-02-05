---
sidebar_position: 3
---

# Request Lifecycle

Understanding how forge processes HTTP requests helps you build better applications and debug issues.

## Overview

When a request comes in, forge processes it through several stages:

```
HTTP Request
    ↓
Router (matches route)
    ↓
Middleware Stack
    ↓
Handler/ViewSet
    ↓
Authentication
    ↓
Permissions
    ↓
Business Logic
    ↓
Response
```

## Step-by-step flow

### 1. HTTP Request arrives

A client sends an HTTP request to your forge application:

```http
GET /api/posts HTTP/1.1
Host: localhost:8000
Authorization: Bearer token123
```

### 2. Router matches route

The router (chi) matches the request to a registered route:

```go
router.Get("/api/posts", listPostsHandler)
```

### 3. Middleware stack executes

Middleware runs in registration order:

```go
router.Use(middleware.RequestID)
router.Use(middleware.Logger)
router.Use(middleware.Recoverer)
router.Use(csrf.Protect)
router.Use(auth.RequireAuth)
```

**Common middleware:**
- Request ID - Adds unique ID to each request
- Logger - Logs request details
- Recoverer - Recovers from panics
- CSRF - Protects against CSRF attacks
- Auth - Authenticates users

### 4. Handler/ViewSet executes

The matched handler or ViewSet processes the request:

```go
func listPostsHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    posts, err := Post.Objects.
        Filter(Post.Fields.Published.Equals(true)).
        All(ctx)
    
    // Return JSON response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}
```

### 5. Authentication (if required)

If the route requires authentication:

```go
user := auth.GetUser(r.Context())

// Or use middleware
router.Use(auth.RequireAuth)
```

### 6. Permissions (if required)

Check if user has required permissions:

```go
if !users.HasPermission(user, "posts.view_post") {
    http.Error(w, "Forbidden", http.StatusForbidden)
    return
}
```

### 7. Business logic

Execute your application logic:

```go
posts, err := Post.Objects.
    Filter(Post.Fields.AuthorID.Equals(user.ID)).
    All(ctx)

for _, post := range posts {
    post.ViewCount++
    Post.Objects.Update(ctx, post)
}
```

### 8. Response

Return response to client:

```go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(posts)

// Or return HTML
templates.ExecuteTemplate(w, "posts.html", posts)
```

## Example: Complete flow

Here's a complete example:

```go
router.Get("/api/posts", listPostsHandler)
router.Use(middleware.Logger)
router.Use(auth.RequireAuth)

func listPostsHandler(w http.ResponseWriter, r *http.Request) {
    user := auth.GetUser(r.Context())
    
    if !users.HasPermission(user, "posts.view_post") {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    
    ctx := r.Context()
    posts, err := Post.Objects.
        Filter(Post.Fields.Published.Equals(true)).
        All(ctx)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}
```

## ViewSet flow

For REST API ViewSets, the flow is more structured:

```
Request
    ↓
ViewSet Handler
    ↓
Authentication
    ↓
Permissions
    ↓
Throttling
    ↓
Content Negotiation
    ↓
Parse Request
    ↓
ViewSet Action (list, create, etc.)
    ↓
Serialize
    ↓
Render Response
```

## Error handling

Errors can occur at any stage:

1. **Routing errors** - 404 Not Found
2. **Authentication errors** - 401 Unauthorized
3. **Permission errors** - 403 Forbidden
4. **Validation errors** - 400 Bad Request
5. **Server errors** - 500 Internal Server Error

```go
func listPostsHandler(w http.ResponseWriter, r *http.Request) {
    posts, err := Post.Objects.All(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(posts)
}
```

## Context propagation

The request context carries important information:

```go
ctx := r.Context()
posts, err := Post.Objects.All(ctx)
```

## Performance considerations

### Middleware order matters

Place expensive middleware later in the chain:

```go
router.Use(middleware.RequestID)
router.Use(middleware.Logger)

// Expensive operations go last
router.Use(auth.RequireAuth)
router.Use(rateLimit.Middleware)
```

### Database connection pooling

forge uses connection pooling automatically:

```yaml
database:
  max_connections: 25
  max_idle_connections: 5
```

### Query optimization

Use SelectRelated and PrefetchRelated to avoid N+1 queries:

```go
// Single query with JOIN
posts, err := Post.Objects.
    SelectRelated("author").
    All(ctx)

// N+1 queries (avoid this)
posts, err := Post.Objects.All(ctx)
for _, post := range posts {
    author := post.Author
}
```

## Next steps

- **[Middleware](/docs/guides/middleware/)** - Learn about middleware
- **[Authentication](/docs/guides/authentication/)** - Understand authentication
- **[Error Handling](/docs/guides/error-handling/)** - Handle errors properly
