---
sidebar_position: 2
description: Learn how to query your database with forge's type-safe QuerySet API. Filter, order, aggregate, and optimize queries with compile-time type checking.
keywords:
  - forge queries
  - forge queryset
  - type-safe queries
  - go database queries
  - forge filter
  - forge orm queries
image: /img/forge-social-card.jpg
---

# Queries

The QuerySet API is how you query your database in forge. No SQL strings, no runtime errors from typos—everything is type-safe and checked at compile time.

## Why QuerySet?

QuerySet fixes the usual problems:

- **Type safety** - Your code won't compile if you typo a field name
- **No SQL injection** - Everything uses parameter binding automatically
- **IDE support** - Autocomplete works, refactoring is safe
- **Familiar API** - If you know Django's ORM, this feels the same

## Basic queries

### Get All Objects

```go
ctx := context.Background()
users, err := UserObjects.All(ctx)
```

### Get a Single Object

```go
user, err := UserObjects.Get(ctx, 1)
user, err := UserObjects.First(ctx)
user, err := UserObjects.Last(ctx)
```

### Filtering

```go
users, err := UserObjects.
    Filter(UserFieldsInstance.IsActive.Equals(true)).
    All(ctx)

// Multiple filters are ANDed together
users, err := UserObjects.
    Filter(UserFieldsInstance.IsActive.Equals(true)).
    Filter(UserFieldsInstance.IsStaff.Equals(true)).
    All(ctx)

users, err := UserObjects.
    Exclude(UserFieldsInstance.IsDeleted.Equals(true)).
    All(ctx)
```

## Field Expressions

Field expressions provide type-safe field access:

### Equality

```go
UserFieldsInstance.Username.Equals("john")
UserFieldsInstance.Age.NotEquals(18)
```

### Comparison

```go
UserFieldsInstance.Age.Greater(18)
UserFieldsInstance.Age.GreaterOrEqual(18)
UserFieldsInstance.Age.Less(65)
UserFieldsInstance.Age.LessOrEqual(65)
```

### Null Checks

```go
UserFieldsInstance.LastLogin.IsNull()
UserFieldsInstance.LastLogin.IsNotNull()
```

### Membership

```go
UserFieldsInstance.Status.In("active", "pending", "approved")
UserFieldsInstance.Status.NotIn("deleted", "banned")
```

### String Operations

```go
UserFieldsInstance.Username.Contains("john")
UserFieldsInstance.Username.StartsWith("admin")
UserFieldsInstance.Username.EndsWith(".com")
UserFieldsInstance.Username.IContains("JOHN")
```

### Range

```go
UserFieldsInstance.Age.Range(18, 65)
UserFieldsInstance.CreatedAt.Range(startDate, endDate)
```

## Complex Queries

### Combining Conditions

```go
users, err := UserObjects.
    Filter(
        UserFieldsInstance.IsActive.Equals(true).
            And(UserFieldsInstance.DateJoined.Greater(lastMonth)),
    ).
    All(ctx)

users, err := UserObjects.
    Filter(
        UserFieldsInstance.IsActive.Equals(true).
            Or(UserFieldsInstance.IsStaff.Equals(true)),
    ).
    All(ctx)

users, err := UserObjects.
    Filter(UserFieldsInstance.IsActive.Equals(true).Not()).
    All(ctx)
```

### Ordering

```go
// Single field
users, err := UserObjects.OrderBy("username").All(ctx)

// Multiple fields
users, err := UserObjects.OrderBy("username", "-date_joined").All(ctx)

// Descending (use - prefix)
users, err := UserObjects.OrderBy("-date_joined").All(ctx)
```

### Limiting and Offsetting

```go
// Limit
users, err := UserObjects.Limit(10).All(ctx)

// Offset
users, err := UserObjects.Offset(20).All(ctx)

// Pagination
users, err := UserObjects.
    Limit(10).
    Offset(20).
    All(ctx)
```

### Distinct

```go
users, err := UserObjects.Distinct().All(ctx)
```

## Aggregations

### Count

```go
count, err := UserObjects.
    Filter(UserFieldsInstance.IsActive.Equals(true)).
    Count(ctx)
```

### Exists

```go
exists, err := UserObjects.
    Filter(UserFieldsInstance.Username.Equals("john")).
    Exists(ctx)
```

### Aggregates

```go
import "github.com/forgego/forge/orm"

result, err := UserObjects.
    Aggregate(
        orm.Count("id"),
        orm.Avg("age"),
        orm.Max("date_joined"),
        orm.Min("date_joined"),
        orm.Sum("score"),
    ).
    Get(ctx)
```

## Field Selection

### Select Specific Fields

```go
// Only these fields
users, err := UserObjects.
    Only("username", "email").
    All(ctx)

// Exclude these fields
users, err := UserObjects.
    Defer("password", "secret").
    All(ctx)
```

## Relations

### Accessing Relations

```go
// Get user with posts
user, err := UserObjects.Get(ctx, 1)
posts := user.Posts  // []*Post

// Prefetch related objects (efficient)
users, err := UserObjects.
    PrefetchRelated("posts", "profile").
    All(ctx)

// Select related (JOIN)
users, err := UserObjects.
    SelectRelated("profile").
    All(ctx)
```

### Filtering by Relations

```go
// Get posts by author
posts, err := PostObjects.
    Filter(PostFieldsInstance.Author.Equals(userID)).
    All(ctx)

// Get users with posts
users, err := UserObjects.
    Filter(UserFieldsInstance.Posts.Count().Greater(0)).
    All(ctx)
```

## Updates

### Update Single Object

```go
user.Username = "newusername"
err := UserObjects.Update(ctx, user)
```

### Bulk Update

```go
affected, err := UserObjects.
    Filter(UserFieldsInstance.IsActive.Equals(false)).
    Update(ctx, map[string]interface{}{
        "is_active": true,
    })
```

## Deletion

### Delete Single Object

```go
err := UserObjects.Delete(ctx, user)
```

### Bulk Delete

```go
affected, err := UserObjects.
    Filter(UserFieldsInstance.IsDeleted.Equals(true)).
    Delete(ctx)
```

## Transactions

### Using Transactions

```go
import "github.com/forgego/forge/db"

err := db.WithTx(ctx, func(tx *db.Tx) error {
    user := &User{Username: "john"}
    if err := UserObjects.Create(ctx, user); err != nil {
        return err
    }
    
    post := &Post{Author: user, Title: "Hello"}
    return PostObjects.Create(ctx, post)
})
```

## Query Optimization

### Use SelectRelated for Foreign Keys

```go
// Good: Uses JOIN
posts, err := PostObjects.
    SelectRelated("author").
    All(ctx)

// Bad: N+1 queries
posts, err := PostObjects.All(ctx)
for _, post := range posts {
    author := post.Author  // Separate query for each post
}
```

### Use PrefetchRelated for Many Relations

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

### Use Only/Defer for Large Models

```go
// Good: Only fetch needed fields
users, err := UserObjects.
    Only("username", "email").
    All(ctx)

// Bad: Fetch all fields including large text fields
users, err := UserObjects.All(ctx)
```

## Common Patterns

### Pagination

```go
func GetUsers(page, pageSize int) ([]*User, int64, error) {
    ctx := context.Background()
    
    // Get total count
    total, err := UserObjects.Count(ctx)
    if err != nil {
        return nil, 0, err
    }
    
    // Get page
    users, err := UserObjects.
        Limit(pageSize).
        Offset((page - 1) * pageSize).
        OrderBy("-date_joined").
        All(ctx)
    
    return users, total, err
}
```

### Search

```go
func SearchUsers(query string) ([]*User, error) {
    ctx := context.Background()
    
    return UserObjects.
        Filter(
            UserFieldsInstance.Username.Contains(query).
                Or(UserFieldsInstance.Email.Contains(query)),
        ).
        All(ctx)
}
```

## Next Steps

- [API Reference](/docs/api-reference/queryset) - Complete QuerySet API
- [Manager Reference](/docs/api-reference/manager) - Manager methods
- [Field Reference](/docs/api-reference/fields) - Field expression methods

