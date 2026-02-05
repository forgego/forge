---
sidebar_position: 9
---

# Best Practices

This guide covers best practices for building maintainable applications with forge.

## Code Organization

### Project Structure

Follow the recommended project structure:

```
myproject/
├── cmd/server/main.go      # Application entry point
├── app/                    # Django-style apps
│   ├── users/
│   │   ├── models.go
│   │   ├── admin.go
│   │   └── api.go
│   └── blog/
│       ├── models.go
│       ├── admin.go
│       └── api.go
├── domain/                 # Business logic (optional)
├── infra/                  # Infrastructure (optional)
├── pkg/                    # Shared utilities
├── migrations/
└── config/
```

### App Organization

Keep apps focused and cohesive:

- **One app = One domain concept** (users, blog, shop)
- **Keep files small** - Split large files
- **Use init() for registration** - Auto-discover components
- **Group related code** - Models, admin, API in same app

## Model Design

### Field Naming

Use clear, descriptive field names:

```go
// ✅ Good
schema.String("email_address").Build()
schema.String("phone_number").Build()
schema.Time("created_at").Build()

// ❌ Bad
schema.String("email").Build()  // Too generic
schema.String("ph").Build()      // Abbreviation
schema.Time("c").Build()         // Too short
```

### Relationships

Design relationships carefully:

```go
// ✅ Good: Clear relationship
relations.ForeignKey("author", "User").
    Required().
    OnDelete(schema.Cascade).
    RelatedName("posts")

// ❌ Bad: Unclear or missing options
relations.ForeignKey("user", "User")  // Missing cascade, related name
```

### Indexes

Add indexes for frequently queried fields:

```go
func (Post) Meta() schema.Meta {
    return schema.Meta{
        Indexes: []schema.Index{
            {
                Name:   "idx_posts_author",
                Fields: []string{"author_id"},
            },
            {
                Name:   "idx_posts_created",
                Fields: []string{"created_at"},
            },
        },
    }
}
```

## Query Optimization

### Use SelectRelated

Avoid N+1 queries:

```go
// ✅ Good: Uses JOIN
posts, err := Post.Objects.
    SelectRelated("author").
    All(ctx)

// ❌ Bad: N+1 queries
posts, err := Post.Objects.All(ctx)
for _, post := range posts {
    author := post.Author  // Separate query for each post
}
```

### Use PrefetchRelated

For many-to-many and reverse foreign keys:

```go
// ✅ Good: Two queries total
users, err := User.Objects.
    PrefetchRelated("posts", "comments").
    All(ctx)

// ❌ Bad: N+1 queries
users, err := User.Objects.All(ctx)
for _, user := range users {
    posts := user.Posts  // Separate query for each user
}
```

### Limit Fields

Only fetch needed fields:

```go
// ✅ Good: Only fetch needed fields
users, err := User.Objects.
    Only("username", "email").
    All(ctx)

// ❌ Bad: Fetch all fields including large text
users, err := User.Objects.All(ctx)
```

## Security

### Input Validation

Always validate user input:

```go
// ✅ Good: Validate in multiple layers
func (Post) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        Clean: func(ctx context.Context, instance interface{}) error {
            post := instance.(*Post)
            if post.Title == "" {
                return errors.New("title is required")
            }
            return nil
        },
    }
}

// Also validate in API layer
func ValidatePostRequest(req *CreatePostRequest) error {
    if req.Title == "" {
        return errors.New("title is required")
    }
    return nil
}
```

### SQL Injection Prevention

Never concatenate SQL:

```go
// ✅ Good: Use QuerySet API
users, err := User.Objects.
    Filter(User.Fields.Username.Equals(username)).
    All(ctx)

// ❌ Bad: String concatenation
query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)
```

### Password Security

Always hash passwords:

```go
// ✅ Good: Hash in BeforeCreate hook
func (User) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        BeforeCreate: func(ctx context.Context, instance interface{}) error {
            user := instance.(*User)
            hashed, err := backends.HashPassword(user.Password)
            if err != nil {
                return err
            }
            user.Password = hashed
            return nil
        },
    }
}
```

## Error Handling

### Consistent Error Responses

Return consistent error formats:

```go
type ErrorResponse struct {
    Error   string            `json:"error"`
    Details map[string]string `json:"details,omitempty"`
}

func HandleError(w http.ResponseWriter, err error) {
    var status int
    var message string

    switch e := err.(type) {
    case *NotFoundError:
        status = http.StatusNotFound
        message = e.Message
    case *ValidationError:
        status = http.StatusBadRequest
        message = e.Message
        // Include validation details
    default:
        status = http.StatusInternalServerError
        message = "Internal server error"
        // Log actual error
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
```

### Don't Expose Internal Errors

Never expose internal errors to users:

```go
// ✅ Good: Generic error message
if err != nil {
    log.Error("Database error", err)
    http.Error(w, "An error occurred", http.StatusInternalServerError)
    return
}

// ❌ Bad: Expose internal error
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

## Testing

### Write Tests

Test business logic thoroughly:

```go
func TestCreatePost(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()

    post := &Post{
        Title:     "Test Post",
        Content:   "Test Content",
        Published: true,
    }

    err := Post.Objects.Create(ctx, post)
    require.NoError(t, err)
    assert.NotZero(t, post.ID)
    assert.NotZero(t, post.CreatedAt)
}
```

### Test Edge Cases

Test error conditions:

```go
func TestCreatePostValidation(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()

    post := &Post{
        Title: "", // Missing required field
    }

    err := Post.Objects.Create(ctx, post)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "title is required")
}
```

## Performance

### Use Connection Pooling

Configure connection pool:

```yaml
database:
  max_connections: 25
  max_idle_connections: 5
  max_lifetime: 5m
```

### Cache Expensive Operations

Cache query results:

```go
func GetPopularPosts(ctx context.Context) ([]*Post, error) {
    cacheKey := "popular_posts"

    if cached, err := cache.Get(ctx, cacheKey); err == nil {
        return cached.([]*Post), nil
    }

    posts, err := Post.Objects.
        Filter(Post.Fields.Published.Equals(true)).
        OrderBy("-view_count").
        Limit(10).
        All(ctx)
    if err != nil {
        return nil, err
    }

    cache.Set(ctx, cacheKey, posts, 5*time.Minute)
    return posts, nil
}
```

### Use Transactions

For multi-step operations:

```go
func CreatePostWithTags(ctx context.Context, post *Post, tags []*Tag) error {
    return db.WithTx(ctx, func(tx *db.Tx) error {
        if err := Post.Objects.Create(ctx, post); err != nil {
            return err
        }

        for _, tag := range tags {
            if err := PostTag.Objects.Create(ctx, &PostTag{
                PostID: post.ID,
                TagID:  tag.ID,
            }); err != nil {
                return err
            }
        }

        return nil
    })
}
```

## Documentation

### Document Public APIs

Document exported functions:

```go
// CreatePost creates a new post with the given data.
// It validates the input and returns an error if validation fails.
func CreatePost(ctx context.Context, req *CreatePostRequest) (*Post, error) {
    // ...
}
```

### Add Examples

Include usage examples:

```go
// Example:
//   post, err := CreatePost(ctx, &CreatePostRequest{
//       Title: "My Post",
//       Content: "Post content",
//   })
```

## Configuration

### Use Environment Variables

For sensitive configuration:

```yaml
database:
  password: "${DB_PASSWORD}"
  host: "${DB_HOST}"
```

### Separate Configs

Different configs for different environments:

```
config/
├── development.yaml
├── staging.yaml
└── production.yaml
```

## Deployment

### Health Checks

Add health check endpoints:

```go
router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    if err := db.Ping(); err != nil {
        http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
})
```

### Logging

Use structured logging:

```go
logger.Info("Post created",
    zap.Int64("post_id", post.ID),
    zap.String("title", post.Title),
)
```

## Summary

1. **Organize code** - Follow project structure
2. **Design models** - Clear names, proper relationships
3. **Optimize queries** - Use SelectRelated, PrefetchRelated
4. **Secure** - Validate input, hash passwords
5. **Handle errors** - Consistent error responses
6. **Test** - Write comprehensive tests
7. **Performance** - Cache, pool connections
8. **Document** - Document public APIs
9. **Configure** - Use environment variables
10. **Deploy** - Health checks, logging

## Next Steps

- [Common Patterns](/docs/guides/common-patterns) - Reusable patterns
- [Performance Guide](/docs/advanced/performance) - Performance optimization
- [Security Guide](/docs/guides/security) - Security best practices
