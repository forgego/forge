---
sidebar_position: 4
---

# Performance Optimization

Tips and techniques for optimizing forge applications.

## Database Optimization

### Connection Pooling

Configure connection pool:

```go
database.SetMaxOpenConns(25)
database.SetMaxIdleConns(5)
database.SetConnMaxLifetime(5 * time.Minute)
```

### Use Indexes

Add indexes for frequently queried fields:

```go
func (User) Meta() schema.Meta {
    return schema.Meta{
        Indexes: []schema.Index{
            {Name: "idx_user_email", Fields: []string{"email"}},
            {Name: "idx_user_username", Fields: []string{"username"}},
        },
    }
}
```

### Query Optimization

#### Use SelectRelated

Use JOINs for foreign keys:

```go
// Good: Single query with JOIN
posts, err := PostObjects.
    SelectRelated("author").
    All(ctx)

// Bad: N+1 queries
posts, err := PostObjects.All(ctx)
for _, post := range posts {
    author := post.Author  // Separate query
}
```

#### Use PrefetchRelated

Prefetch many relations:

```go
// Good: Two queries total
users, err := UserObjects.
    PrefetchRelated("posts").
    All(ctx)

// Bad: N+1 queries
users, err := UserObjects.All(ctx)
for _, user := range users {
    posts := user.Posts  // Separate query for each user
}
```

#### Use Only/Defer

Select only needed fields:

```go
// Good: Only fetch needed fields
users, err := UserObjects.
    Only("username", "email").
    All(ctx)

// Bad: Fetch all fields including large text
users, err := UserObjects.All(ctx)
```

## Caching

### Query Result Caching

Cache frequently accessed queries:

```go
import "github.com/forgego/forge/api/caching"

cache := caching.NewMemoryCache()

key := fmt.Sprintf("user:%d", userID)
if cached, err := cache.Get(key); err == nil {
    return cached.(*User), nil
}

user, err := UserObjects.Get(ctx, userID)
if err == nil {
    cache.Set(key, user, 5*time.Minute)
}
return user, err
```

### Model Instance Caching

Cache model instances:

```go
func GetUserCached(ctx context.Context, id int64) (*User, error) {
    key := fmt.Sprintf("user:%d", id)
    
    if cached, err := cache.Get(key); err == nil {
        return cached.(*User), nil
    }
    
    user, err := UserObjects.Get(ctx, id)
    if err == nil {
        cache.Set(key, user, 10*time.Minute)
    }
    
    return user, err
}
```

## Pagination

Always paginate large result sets:

```go
func GetUsers(page, pageSize int) ([]*User, int64, error) {
    ctx := context.Background()
    
    total, err := UserObjects.Count(ctx)
    if err != nil {
        return nil, 0, err
    }
    
    users, err := UserObjects.
        Limit(pageSize).
        Offset((page - 1) * pageSize).
        OrderBy("-date_joined").
        All(ctx)
    
    return users, total, err
}
```

## Batch Operations

Use bulk operations when possible:

```go
// Good: Single query
affected, err := UserObjects.
    Filter(UserFieldsInstance.IsActive.Equals(false)).
    Update(ctx, map[string]interface{}{
        "is_active": true,
    })

// Bad: Multiple queries
users, _ := UserObjects.
    Filter(UserFieldsInstance.IsActive.Equals(false)).
    All(ctx)
for _, user := range users {
    user.IsActive = true
    UserObjects.Update(ctx, user)
}
```

## Monitoring

### Query Logging

Enable query logging in development:

```go
import forgelog "github.com/forgego/forge/log"

logger, _ := forgelog.NewLogger(true) // development mode
logger.Info("Query executed",
    zap.String("query", query),
    zap.Duration("duration", duration),
)
```

### Performance Metrics

Add performance metrics:

```go
import "github.com/prometheus/client_golang/prometheus"

var queryDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "forge_query_duration_seconds",
        Help: "Query execution time",
    },
    []string{"model", "operation"},
)
```

## Best Practices

1. **Use Indexes** - Index frequently queried fields
2. **Optimize Queries** - Use SelectRelated and PrefetchRelated
3. **Cache Results** - Cache frequently accessed data
4. **Paginate Lists** - Always paginate large result sets
5. **Use Bulk Operations** - Prefer bulk operations over loops
6. **Monitor Performance** - Track query performance
7. **Profile Your App** - Use profiling tools to find bottlenecks

## See Also

- [Queries Guide](/docs/guides/queries) - Query optimization
- [Deployment Guide](/docs/guides/deployment) - Production deployment

