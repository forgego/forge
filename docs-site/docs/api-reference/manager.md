---
sidebar_position: 3
description: Complete Manager API reference. CRUD operations, bulk operations, and model lifecycle methods. Learn how to use forge's Manager API.
keywords:
  - forge manager api
  - manager reference
  - crud operations
  - model manager
image: /img/forge-social-card.jpg
---

# Manager

The `Manager` provides CRUD operations for your models. Each model has a default manager accessible via `Model.Objects`.

Complete API reference for Manager operations.

## Manager Methods

### Create

Create a new object:

```go
user := &User{
    Username: "john",
    Email:    "john@example.com",
}
err := User.Objects.Create(ctx, user)
```

### Get

Get object by ID:

```go
user, err := User.Objects.Get(ctx, 1)
```

### Update

Update an object:

```go
user.Username = "jane"
err := User.Objects.Update(ctx, user)
```

### Delete

Delete an object:

```go
err := User.Objects.Delete(ctx, user)
```

### All

Get all objects:

```go
users, err := User.Objects.All(ctx)
```

### Filter

Get QuerySet for filtering:

```go
qs := User.Objects.Filter(User.Fields.IsActive.Equals(true))
```

## Instance Methods

### Save

Save (create or update):

```go
err := user.Save(ctx)
```

### Delete

Delete instance:

```go
err := user.Delete(ctx)
```

## See Also

- [QuerySet Reference](/docs/api-reference/queryset) - QuerySet methods
- [Queries Guide](/docs/guides/queries) - Query usage guide

