---
sidebar_position: 8
---

# Common Patterns

This guide covers common patterns and best practices for building applications with forge.

## Model Patterns

### Soft Delete Pattern

Implement soft deletes by adding a `deleted_at` field:

```go
func (Post) Fields() []schema.Field {
    return []schema.Field{
        // ... other fields ...
        schema.Time("deleted_at").Null().Build(),
    }
}

// Custom QuerySet method
func (qs *PostQuerySet) Active(ctx context.Context) (*PostQuerySet, error) {
    return qs.Filter(Post.Fields.DeletedAt.IsNull()), nil
}
```

### Timestamps Pattern

Add created/updated timestamps to all models:

```go
func (Post) Fields() []schema.Field {
    return []schema.Field{
        // ... other fields ...
        schema.Time("created_at").AutoNowAdd().Build(),
        schema.Time("updated_at").AutoNow().Build(),
    }
}
```

### UUID Primary Keys

Use UUIDs instead of auto-incrementing integers:

```go
func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.UUID("id").Primary().Build(),
        // ... other fields ...
    }
}
```

## Query Patterns

### Pagination Pattern

Implement consistent pagination:

```go
type PaginatedResult[T any] struct {
    Items      []*T
    Total      int64
    Page       int
    PageSize   int
    TotalPages int
}

func Paginate[T any](
    qs QuerySet[T],
    page, pageSize int,
    ctx context.Context,
) (*PaginatedResult[T], error) {
    total, err := qs.Count(ctx)
    if err != nil {
        return nil, err
    }

    items, err := qs.
        Limit(pageSize).
        Offset((page - 1) * pageSize).
        All(ctx)
    if err != nil {
        return nil, err
    }

    totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

    return &PaginatedResult[T]{
        Items:      items,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: totalPages,
    }, nil
}
```

### Search Pattern

Implement full-text search:

```go
func SearchPosts(query string, ctx context.Context) ([]*Post, error) {
    return Post.Objects.
        Filter(
            Post.Fields.Title.Contains(query).
                Or(Post.Fields.Content.Contains(query)),
        ).
        OrderBy("-created_at").
        All(ctx)
}
```

### Filter Chain Pattern

Build dynamic filters:

```go
func BuildPostFilters(filters map[string]interface{}) QueryExpr {
    var conditions []QueryExpr

    if published, ok := filters["published"].(bool); ok {
        conditions = append(conditions, Post.Fields.Published.Equals(published))
    }

    if authorID, ok := filters["author_id"].(int64); ok {
        conditions = append(conditions, Post.Fields.AuthorID.Equals(authorID))
    }

    if len(conditions) == 0 {
        return nil
    }

    result := conditions[0]
    for i := 1; i < len(conditions); i++ {
        result = result.And(conditions[i])
    }

    return result
}
```

## Service Layer Pattern

### Service Interface

Define service interfaces:

```go
type PostService interface {
    CreatePost(ctx context.Context, req *CreatePostRequest) (*Post, error)
    UpdatePost(ctx context.Context, id int64, req *UpdatePostRequest) (*Post, error)
    DeletePost(ctx context.Context, id int64) error
    GetPost(ctx context.Context, id int64) (*Post, error)
    ListPosts(ctx context.Context, filters *PostFilters) ([]*Post, error)
}
```

### Service Implementation

Implement services with business logic:

```go
type postService struct {
    db *db.DB
}

func NewPostService(db *db.DB) PostService {
    return &postService{db: db}
}

func (s *postService) CreatePost(ctx context.Context, req *CreatePostRequest) (*Post, error) {
    // Validate request
    if req.Title == "" {
        return nil, errors.New("title is required")
    }

    // Create post
    post := &Post{
        Title:     req.Title,
        Content:   req.Content,
        Published: req.Published,
    }

    err := Post.Objects.Create(ctx, post)
    if err != nil {
        return nil, err
    }

    return post, nil
}
```

## Handler Patterns

### HTTP Handler Pattern

Create consistent HTTP handlers:

```go
func ListPostsHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Parse query parameters
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }

    // Get posts
    result, err := Paginate(Post.Objects, page, 20, ctx)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Serialize response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### Error Handling Pattern

Consistent error responses:

```go
type ErrorResponse struct {
    Error   string            `json:"error"`
    Details map[string]string `json:"details,omitempty"`
}

func SendError(w http.ResponseWriter, status int, message string, details ...map[string]string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)

    resp := ErrorResponse{Error: message}
    if len(details) > 0 {
        resp.Details = details[0]
    }

    json.NewEncoder(w).Encode(resp)
}
```

## Validation Patterns

### Model Validation

Validate in hooks:

```go
func (Post) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        Clean: func(ctx context.Context, instance interface{}) error {
            post := instance.(*Post)

            if post.Title == "" {
                return errors.New("title is required")
            }

            if len(post.Title) > 200 {
                return errors.New("title must be 200 characters or less")
            }

            return nil
        },
    }
}
```

### Request Validation

Validate API requests:

```go
type CreatePostRequest struct {
    Title     string `json:"title" validate:"required,max=200"`
    Content   string `json:"content" validate:"required"`
    Published bool   `json:"published"`
}

func ValidateCreatePostRequest(req *CreatePostRequest) error {
    if req.Title == "" {
        return errors.New("title is required")
    }

    if len(req.Title) > 200 {
        return errors.New("title must be 200 characters or less")
    }

    return nil
}
```

## Transaction Patterns

### Transaction Wrapper

Wrap operations in transactions:

```go
func CreatePostWithTags(ctx context.Context, post *Post, tagIDs []int64) error {
    return db.WithTx(ctx, func(tx *db.Tx) error {
        // Create post
        if err := Post.Objects.Create(ctx, post); err != nil {
            return err
        }

        // Create tag associations
        for _, tagID := range tagIDs {
            association := &PostTag{
                PostID: post.ID,
                TagID:  tagID,
            }
            if err := PostTag.Objects.Create(ctx, association); err != nil {
                return err
            }
        }

        return nil
    })
}
```

## Caching Patterns

### Query Result Caching

Cache expensive queries:

```go
func GetPopularPosts(ctx context.Context) ([]*Post, error) {
    cacheKey := "popular_posts"

    // Try cache
    if cached, err := cache.Get(ctx, cacheKey); err == nil {
        return cached.([]*Post), nil
    }

    // Query database
    posts, err := Post.Objects.
        Filter(Post.Fields.Published.Equals(true)).
        OrderBy("-view_count").
        Limit(10).
        All(ctx)
    if err != nil {
        return nil, err
    }

    // Cache result
    cache.Set(ctx, cacheKey, posts, 5*time.Minute)

    return posts, nil
}
```

## Testing Patterns

### Test Setup Pattern

Consistent test setup:

```go
func setupTestDB(t *testing.T) *db.DB {
    db, err := db.NewDB("postgres://test:test@localhost/testdb?sslmode=disable")
    if err != nil {
        t.Fatal(err)
    }

    // Run migrations
    if err := migrate.Up(db); err != nil {
        t.Fatal(err)
    }

    t.Cleanup(func() {
        migrate.Down(db)
        db.Close()
    })

    return db
}
```

### Test Data Factory

Create test data:

```go
func CreateTestPost(t *testing.T, db *db.DB) *Post {
    post := &Post{
        Title:     "Test Post",
        Content:   "Test Content",
        Published: true,
    }

    err := Post.Objects.Create(context.Background(), post)
    require.NoError(t, err)

    return post
}
```

## Best Practices

1. **Use Services for Business Logic** - Keep handlers thin
2. **Validate Early** - Validate in multiple layers
3. **Use Transactions** - For multi-step operations
4. **Cache Expensive Queries** - But invalidate properly
5. **Handle Errors Gracefully** - Return user-friendly messages
6. **Use Type Safety** - Prefer type-safe API over dynamic
7. **Write Tests** - Test business logic thoroughly
8. **Document Patterns** - Share patterns with your team

## Next Steps

- [Best Practices](/docs/guides/best-practices) - More best practices
- [Performance Guide](/docs/advanced/performance) - Performance optimization
- [Security Guide](/docs/guides/security) - Security patterns
