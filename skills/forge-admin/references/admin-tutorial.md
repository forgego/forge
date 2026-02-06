---
sidebar_position: 3
---

# Tutorial 3: Building an Admin Interface

In this tutorial, you'll learn how to create a fully-featured admin interface for your forge application. The admin system provides a Django-like interface with zero code required for basic functionality.

## What You'll Build

You'll create an admin interface for a blog application with:
- User management with custom fields
- Post management with search and filters
- Comment management with inlines
- Custom actions and permissions
- Export functionality

## Prerequisites

- Completed [Tutorial 1: Getting Started](/docs/tutorials/getting-started/)
- A working forge application
- Basic understanding of Go and the forge ORM

## Step 1: Set Up Your Models

First, let's define our models with proper schema definitions:

```go
// app/blog/models.go
package blog

import (
    "time"
    "github.com/forgego/forge/schema"
    query "github.com/forgego/forge/orm"
)

// User model
type User struct {
    ID        int64     `db:"id"`
    Username  string    `db:"username"`
    Email     string    `db:"email"`
    IsActive  bool      `db:"is_active"`
    IsStaff   bool      `db:"is_staff"`
    CreatedAt time.Time `db:"created_at"`
}

// UserSchema implements schema.Schema
type UserSchema struct{}

func (s *UserSchema) Fields() []schema.Field {
    return []schema.Field{
        {Name: "ID", Type: schema.TypeInt64, PrimaryKey: true, AutoIncrement: true},
        {Name: "Username", Type: schema.TypeString, Required: true, MaxLength: intPtr(150)},
        {Name: "Email", Type: schema.TypeEmail, Required: true},
        {Name: "IsActive", Type: schema.TypeBool, Default: true},
        {Name: "IsStaff", Type: schema.TypeBool, Default: false},
        {Name: "CreatedAt", Type: schema.TypeDateTime, AutoNowAdd: true, Editable: false},
    }
}

func (s *UserSchema) Meta() schema.Meta {
    return schema.Meta{
        TableName:         "users",
        VerboseName:      "User",
        VerboseNamePlural: "Users",
    }
}

// Post model
type Post struct {
    ID        int64     `db:"id"`
    Title     string    `db:"title"`
    Content   string    `db:"content"`
    AuthorID  int64     `db:"author_id"`
    Published bool      `db:"published"`
    CreatedAt  time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}

// PostSchema implements schema.Schema
type PostSchema struct{}

func (s *PostSchema) Fields() []schema.Field {
    return []schema.Field{
        {Name: "ID", Type: schema.TypeInt64, PrimaryKey: true, AutoIncrement: true},
        {Name: "Title", Type: schema.TypeString, Required: true, MaxLength: intPtr(200)},
        {Name: "Content", Type: schema.TypeText, Required: true},
        {Name: "AuthorID", Type: schema.TypeForeignKey, RelatedModel: "User"},
        {Name: "Published", Type: schema.TypeBool, Default: false},
        {Name: "CreatedAt", Type: schema.TypeDateTime, AutoNowAdd: true},
        {Name: "UpdatedAt", Type: schema.TypeDateTime, AutoNow: true},
    }
}

func (s *PostSchema) Relations() []schema.Relation {
    return []schema.Relation{
        {Type: schema.RelationForeignKey, Field: "AuthorID", RelatedModel: "User"},
    }
}

func (s *PostSchema) Meta() schema.Meta {
    return schema.Meta{
        TableName:         "posts",
        VerboseName:      "Post",
        VerboseNamePlural: "Posts",
        OrderBy:          []string{"-created_at"},
    }
}

// Comment model
type Comment struct {
    ID        int64     `db:"id"`
    PostID    int64     `db:"post_id"`
    Author    string    `db:"author"`
    Content   string    `db:"content"`
    CreatedAt time.Time `db:"created_at"`
}

// CommentSchema implements schema.Schema
type CommentSchema struct{}

func (s *CommentSchema) Fields() []schema.Field {
    return []schema.Field{
        {Name: "ID", Type: schema.TypeInt64, PrimaryKey: true, AutoIncrement: true},
        {Name: "PostID", Type: schema.TypeForeignKey, RelatedModel: "Post"},
        {Name: "Author", Type: schema.TypeString, Required: true},
        {Name: "Content", Type: schema.TypeText, Required: true},
        {Name: "CreatedAt", Type: schema.TypeDateTime, AutoNowAdd: true},
    }
}

func (s *CommentSchema) Relations() []schema.Relation {
    return []schema.Relation{
        {Type: schema.RelationForeignKey, Field: "PostID", RelatedModel: "Post"},
    }
}

func intPtr(i int) *int { return &i }
```

## Step 2: Generate Code

Generate type-safe code from your models:

```bash
forge generate
```

This creates:
- Generated model structs
- Type-safe field expressions
- Manager with CRUD operations
- QuerySet for filtering

## Step 3: Basic Admin Registration

Create an admin configuration file:

```go
// app/blog/admin.go
package blog

import (
    "context"
    "github.com/forgego/forge/admin"
    "github.com/forgego/forge/admin/http"
    admincore "github.com/forgego/forge/admin/core"
    "github.com/forgego/forge/schema"
    query "github.com/forgego/forge/orm"
)

var (
    userAdmin   *admincore.Admin[User]
    postAdmin   *admincore.Admin[Post]
    commentAdmin *admincore.Admin[Comment]
)

func InitAdmin(
    userManager *query.Manager[User],
    postManager *query.Manager[Post],
    commentManager *query.Manager[Comment],
) error {
    // User Admin - Basic configuration
    userSchema := &UserSchema{}
    userAdmin, err := admincore.Register[User](
        schema.NewSchema(userSchema),
        userManager,
        &admincore.Config[User]{
            // List display fields
            ListDisplay: []string{"username", "email", "is_active", "is_staff", "created_at"},
            
            // Search fields
            SearchFields: []string{"username", "email"},
            
            // List filters
            ListFilter: []string{"is_active", "is_staff", "created_at"},
            
            // Ordering
            Ordering: []string{"-created_at"},
        },
    )
    if err != nil {
        return err
    }

    // Register for HTTP handlers
    http.RegisterAdminForHTTP(userAdmin)

    return nil
}
```

## Step 4: Advanced Admin Configuration

Let's enhance the Post admin with more features:

```go
func InitPostAdmin(postManager *query.Manager[Post]) error {
    postSchema := &PostSchema{}
    
    postAdmin, err := admincore.Register[Post](
        schema.NewSchema(postSchema),
        postManager,
        &admincore.Config[Post]{
            // List display
            ListDisplay: []string{"title", "author", "published", "created_at"},
            
            // List display links (clickable fields)
            ListDisplayLinks: []string{"title"},
            
            // Search
            SearchFields: []string{"title", "content"},
            
            // Filters
            ListFilter: []string{"published", "author", "created_at"},
            
            // Date hierarchy for navigation
            DateHierarchy: "created_at",
            
            // Ordering
            Ordering: []string{"-created_at"},
            
            // Items per page
            ListPerPage: 25,
            
            // Read-only fields
            ReadOnlyFields: []string{"created_at", "updated_at"},
            
            // Fieldsets for form organization
            Fieldsets: []admincore.Fieldset[Post]{
                {
                    Title: "Post Information",
                    Fields: []string{"title", "content", "author"},
                },
                {
                    Title: "Publishing",
                    Fields: []string{"published"},
                },
                {
                    Title: "Timestamps",
                    Fields: []string{"created_at", "updated_at"},
                },
            },
        },
    )
    if err != nil {
        return err
    }

    http.RegisterAdminForHTTP(postAdmin)
    return nil
}
```

## Step 5: Adding Inlines

Add comments as an inline to the Post admin:

```go
func InitPostAdminWithInlines(
    postManager *query.Manager[Post],
    commentManager *query.Manager[Comment],
) error {
    postSchema := &PostSchema{}
    
    postAdmin, err := admincore.Register[Post](
        schema.NewSchema(postSchema),
        postManager,
        &admincore.Config[Post]{
            // ... previous configuration ...
            
            // Add inlines
            Inlines: []admincore.Inline[Post, Comment]{
                admincore.TabularInline[Post, Comment](
                    commentManager,
                    "post_id", // Foreign key field
                    []string{"author", "content", "created_at"},
                ),
            },
        },
    )
    if err != nil {
        return err
    }

    http.RegisterAdminForHTTP(postAdmin)
    return nil
}
```

## Step 6: Custom Actions

Add bulk actions to your admin:

```go
import (
    adminadvanced "github.com/forgego/forge/admin/advanced"
)

func InitPostAdminWithActions(postManager *query.Manager[Post]) error {
    // ... previous configuration ...
    
    // Create action manager
    actionManager := adminadvanced.NewActionManager(postAdmin)
    
    // Add built-in actions
    builtinActions := adminadvanced.NewBuiltinActions(postAdmin)
    actionManager.RegisterAction(builtinActions.DeleteSelected())
    
    // Add custom actions
    actionManager.RegisterAction(
        admincore.NewAction[Post](
            "publish",
            "Publish selected posts",
            func(ctx context.Context, posts []*Post) error {
                for _, post := range posts {
                    post.Published = true
                    if err := postManager.Update(ctx, post); err != nil {
                        return err
                    }
                }
                return nil
            },
        ),
    )
    
    actionManager.RegisterAction(
        admincore.NewAction[Post](
            "unpublish",
            "Unpublish selected posts",
            func(ctx context.Context, posts []*Post) error {
                for _, post := range posts {
                    post.Published = false
                    if err := postManager.Update(ctx, post); err != nil {
                        return err
                    }
                }
                return nil
            },
        ),
    )
    
    // ... rest of configuration ...
}
```

## Step 7: Custom Filters

Create custom filters for advanced filtering:

```go
import (
    adminfilters "github.com/forgego/forge/admin/filters"
)

func InitPostAdminWithCustomFilters(postManager *query.Manager[Post]) error {
    // ... previous configuration ...
    
    // Custom date range filter
    dateFilter := adminfilters.NewDateRangeFilter[Post](
        admincore.FieldExpr[Post, time.Time]{
            Name: "created_at",
            Get: func(p *Post) time.Time { return p.CreatedAt },
        },
    )
    
    // Custom related filter
    authorFilter := adminfilters.NewRelatedFieldListFilter[Post, User](
        admincore.FieldExpr[Post, *User]{
            Name: "author",
            Get: func(p *Post) *User { return p.Author },
        },
        userManager,
    )
    
    // Add to config
    config.ListFilter = []admincore.Filter[Post]{
        dateFilter,
        authorFilter,
    }
    
    // ... rest of configuration ...
}
```

## Step 8: Permissions

Add permission checking to your admin:

```go
import (
    "github.com/forgego/forge/identity"
)

func InitAdminWithPermissions(
    userManager *query.Manager[User],
    permissionChecker *identity.PermissionChecker,
) error {
    userAdmin, err := admincore.Register[User](
        schema.NewSchema(&UserSchema{}),
        userManager,
        &admincore.Config[User]{
            // ... previous configuration ...
            
            // Permission checker
            PermissionChecker: permissionChecker,
            
            // Custom permission checks
            HasAddPermission: func(ctx context.Context, user interface{}) bool {
                // Check if user has permission to add users
                return permissionChecker.HasPermission(
                    ctx,
                    user,
                    "auth.add_user",
                )
            },
            
            HasChangePermission: func(ctx context.Context, user interface{}, obj *User) bool {
                // Users can only edit themselves unless they're staff
                if staff, ok := user.(*User); ok {
                    if !staff.IsStaff && staff.ID != obj.ID {
                        return false
                    }
                }
                return permissionChecker.HasPermission(
                    ctx,
                    user,
                    "auth.change_user",
                )
            },
            
            HasDeletePermission: func(ctx context.Context, user interface{}, obj *User) bool {
                // Only staff can delete users
                if staff, ok := user.(*User); ok {
                    if !staff.IsStaff {
                        return false
                    }
                }
                return permissionChecker.HasPermission(
                    ctx,
                    user,
                    "auth.delete_user",
                )
            },
        },
    )
    
    // ... rest of setup ...
}
```

## Step 9: Setting Up HTTP Routes

Register admin routes in your main application:

```go
// cmd/server/main.go
package main

import (
    "log"
    "net/http"
    
    "github.com/forgego/forge/admin/http"
    httplib "github.com/forgego/forge/server"
    "your-app/blog"
    "your-app/db"
)

func main() {
    // Setup database
    database := db.NewDBFromConfig(cfg)
    
    // Setup managers
    userManager := query.NewManager[blog.User](database)
    postManager := query.NewManager[blog.Post](database)
    commentManager := query.NewManager[blog.Comment](database)
    
    // Initialize admin
    if err := blog.InitAdmin(userManager, postManager, commentManager); err != nil {
        log.Fatal(err)
    }
    
    // Setup router
    router := httplib.NewRouter()
    
    // Register admin routes
    adminRouter := http.NewRouter(admin.GetGlobalRegistry())
    adminRouter.RegisterRoutes(router, "/admin")
    
    // Add authentication middleware
    router.Use(authMiddleware)
    
    // Start server
    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}
```

## Step 10: Accessing the Admin

1. Start your server:
```bash
go run cmd/server/main.go
```

2. Visit `http://localhost:8080/admin/` in your browser

3. You should see:
   - A list of all registered models
   - Links to manage Users, Posts, and Comments
   - Search and filter capabilities
   - Bulk actions

## Step 11: Customizing the Admin Site

Create a custom admin site:

```go
import (
    admincore "github.com/forgego/forge/admin/core"
)

func InitCustomAdminSite() *admincore.Site {
    site := admincore.NewSite("myblog")
    site.Title = "My Blog Administration"
    site.Header = "Blog Admin"
    site.IndexTitle = "Welcome to Blog Administration"
    
    return site
}
```

## Best Practices

1. **Use List Display** - Show only relevant fields in list view
2. **Enable Search** - Add search fields for frequently searched columns
3. **Use Filters** - Add filters for fields users commonly filter by
4. **Organize Fieldsets** - Group related fields together
5. **Set Read-Only** - Make auto-generated fields read-only
6. **Check Permissions** - Always implement permission checks
7. **Use Inlines** - Use inlines for related models when appropriate
8. **Add Actions** - Provide bulk actions for common operations

## Common Patterns

### Filtering by Related Fields

```go
ListFilter: []string{"author__username", "author__email"},
```

### Custom List Display with Methods

```go
// Add a method to your model
func (p *Post) AuthorName() string {
    if p.Author != nil {
        return p.Author.Username
    }
    return ""
}

// Use in admin
ListDisplay: []string{"title", "author_name", "published"},
```

### Custom Form Validation

```go
config.SaveModel = func(ctx context.Context, admin *admincore.Admin[Post], instance *Post, formData admincore.FormData, isNew bool) error {
    // Custom validation
    if instance.Title == "" {
        return errors.New("title is required")
    }
    
    // Call default save
    return admin.Manager().Create(ctx, instance)
}
```

## Troubleshooting

### Admin Not Showing Up

- Check that models are registered: `admin.GetGlobalRegistry().GetAll()`
- Verify HTTP routes are registered: `adminRouter.RegisterRoutes(router, "/admin")`
- Check authentication middleware is not blocking access

### Filters Not Working

- Ensure filter fields exist in the model
- Check that filter types match field types
- Verify QuerySet supports the filter operations

### Inlines Not Displaying

- Verify foreign key relationship is correct
- Check that related manager is provided
- Ensure inline fields exist in the related model

## Next Steps

- [Admin Guide](/docs/guides/admin) - Complete admin reference
- [REST API Guide](/docs/guides/rest-api) - Build APIs for your frontend
- [Security Guide](/docs/guides/security) - Secure your admin interface
- [Advanced Topics](/docs/advanced/plugins) - Extend the admin with plugins

## Summary

You've learned how to:
- ✅ Register models with the admin system
- ✅ Configure list display, search, and filters
- ✅ Add inlines for related models
- ✅ Create custom actions
- ✅ Implement permissions
- ✅ Set up HTTP routes
- ✅ Customize the admin interface

The admin system provides a powerful, type-safe interface for managing your data with minimal code required.
