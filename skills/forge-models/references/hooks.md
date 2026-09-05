---
sidebar_position: 6
---

# Hooks

Hooks let you run code at specific points in a model's lifecycle. Use them for validation, data transformation, side effects, and more.

Complete reference for model lifecycle hooks.

## Hook Types

### BeforeCreate

Called before creating a new object:

```go
BeforeCreate: func(ctx context.Context, instance interface{}) error {
    user := instance.(*User)
    return nil
}
```

### AfterCreate

Called after creating a new object:

```go
AfterCreate: func(ctx context.Context, instance interface{}) error {
    user := instance.(*User)
    // Send notification, log, etc.
    return nil
}
```

### BeforeUpdate

Called before updating an object:

```go
BeforeUpdate: func(ctx context.Context, instance interface{}) error {
    user := instance.(*User)
    // Validate changes, etc.
    return nil
}
```

### AfterUpdate

Called after updating an object:

```go
AfterUpdate: func(ctx context.Context, instance interface{}) error {
    user := instance.(*User)
    // Update cache, etc.
    return nil
}
```

### BeforeSave

Called before saving (create or update):

```go
BeforeSave: func(ctx context.Context, instance interface{}) error {
    user := instance.(*User)
    return nil
}
```

### AfterSave

Called after saving (create or update):

```go
AfterSave: func(ctx context.Context, instance interface{}) error {
    user := instance.(*User)
    return nil
}
```

### BeforeDelete

Called before deleting an object:

```go
BeforeDelete: func(ctx context.Context, instance interface{}) error {
    user := instance.(*User)
    return nil
}
```

### AfterDelete

Called after deleting an object:

```go
AfterDelete: func(ctx context.Context, instance interface{}) error {
    user := instance.(*User)
    // Cleanup, etc.
    return nil
}
```

### Clean

Custom validation:

```go
Clean: func(ctx context.Context, instance interface{}) error {
    user := instance.(*User)
    if user.Username == "" {
        return errors.New("username is required")
    }
    return nil
}
```

### Validate

Additional validation:

```go
Validate: func(ctx context.Context, instance interface{}) error {
    user := instance.(*User)
    // Additional validation
    return nil
}
```

## Hook Execution Order

### On Create

1. `BeforeSave`
2. `BeforeCreate`
3. Database insert
4. `AfterCreate`
5. `AfterSave`

### On Update

1. `BeforeSave`
2. `BeforeUpdate`
3. Database update
4. `AfterUpdate`
5. `AfterSave`

### On Delete

1. `BeforeDelete`
2. Database delete
3. `AfterDelete`

## Examples

### Password Hashing

```go
func (User) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        BeforeCreate: func(ctx context.Context, instance interface{}) error {
            user := instance.(*User)
            if user.Password != "" {
                hashed, err := auth.HashPassword(user.Password)
                if err != nil {
                    return err
                }
                user.Password = hashed
            }
            return nil
        },
        BeforeUpdate: func(ctx context.Context, instance interface{}) error {
            user := instance.(*User)
            // Only hash if password changed
            if user.Password != "" && !strings.HasPrefix(user.Password, "$2a$") {
                hashed, err := auth.HashPassword(user.Password)
                if err != nil {
                    return err
                }
                user.Password = hashed
            }
            return nil
        },
    }
}
```

### Timestamps

```go
func (Post) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        BeforeCreate: func(ctx context.Context, instance interface{}) error {
            post := instance.(*Post)
            now := time.Now()
            post.CreatedAt = now
            post.UpdatedAt = now
            return nil
        },
        BeforeUpdate: func(ctx context.Context, instance interface{}) error {
            post := instance.(*Post)
            post.UpdatedAt = time.Now()
            return nil
        },
    }
}
```

### Validation

```go
func (User) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        Clean: func(ctx context.Context, instance interface{}) error {
            user := instance.(*User)
            if user.Username == "" {
                return errors.New("username is required")
            }
            if !isValidEmail(user.Email) {
                return errors.New("invalid email address")
            }
            return nil
        },
    }
}
```

## Best Practices

1. **Keep hooks fast** - Hooks should be fast and not block
2. **Handle errors** - Return errors to prevent save/delete
3. **Use appropriate hooks** - Use BeforeSave for common logic, BeforeCreate/BeforeUpdate for specific logic
4. **Don't modify relations in hooks** - Modify relations separately
5. **Use context** - Use context for cancellation and timeouts

## See Also

- [Models Guide](/docs/guides/models) - Model usage guide
- [Schema Reference](/docs/api-reference/schema) - Model definition

