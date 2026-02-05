---
sidebar_position: 3
description: Learn how to use and customize forge's auto-generated admin interface. Django-style admin panel with CRUD operations, search, filtering, and more.
keywords:
  - forge admin
  - admin interface
  - django admin go
  - auto-generated admin
  - forge admin panel
image: /forge-social-card.svg
---

# Admin Interface

The admin interface is forge's auto-generated admin panel. Register your models and you get a full web UI for managing your data. No code required.

## Why use it?

Because building admin interfaces is boring:

- **Zero code** - Just register your models
- **Full CRUD** - Create, read, update, delete—all there
- **Search and filters** - Find what you need fast
- **Bulk actions** - Update or delete multiple records at once
- **Export** - Download your data as CSV or JSON

## Quick start

### 1. Define Your Schema

First, ensure your models implement the schema interface:

```go
type UserSchema struct{}

func (s *UserSchema) Fields() []schema.Field {
    return []schema.Field{
        {Name: "ID", Type: schema.TypeInt64, PrimaryKey: true},
        {Name: "Username", Type: schema.TypeString, Required: true},
        {Name: "Email", Type: schema.TypeEmail, Required: true},
        {Name: "IsActive", Type: schema.TypeBool, Default: true},
    }
}

func (s *UserSchema) Meta() schema.Meta {
    return schema.Meta{
        TableName: "users",
        VerboseName: "User",
        VerboseNamePlural: "Users",
    }
}
```

### 2. Register Models with Admin

```go
import (
    admincore "github.com/forgego/forge/admin/core"
    "github.com/forgego/forge/admin/http"
    "github.com/forgego/forge/schema"
    query "github.com/forgego/forge/orm"
)

func main() {
    // Setup database and managers
    database := db.NewDBFromConfig(cfg)
    userManager := query.NewManager[User](database)
    
    // Register admin
    userSchema := &UserSchema{}
    userAdmin, err := admincore.Register[User](
        schema.NewSchema(userSchema),
        userManager,
        &admincore.Config[User]{
            ListDisplay: []string{"username", "email", "is_active"},
            SearchFields: []string{"username", "email"},
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Register for HTTP handlers
    http.RegisterAdminForHTTP(userAdmin)
    
    // Setup router
    router := httplib.NewRouter()
    adminRouter := http.NewRouter(admin.GetGlobalRegistry())
    adminRouter.RegisterRoutes(router, "/admin")
    
    // Start server
    log.Fatal(http.ListenAndServe(":8080", router))
}
```

### 3. Access Admin

Start your server and visit `http://localhost:8000/admin/`

You'll see:
- A list of all registered models
- Links to view, add, and manage each model
- Search and filtering capabilities
- Full CRUD operations

## Admin Configuration Options

### List Display

Control which fields appear in the list view:

```go
config := &admincore.Config[User]{
    ListDisplay: []string{"username", "email", "is_active", "created_at"},
}
```

### List Display Links

Make fields clickable to navigate to detail view:

```go
config := &admincore.Config[User]{
    ListDisplay: []string{"username", "email"},
    ListDisplayLinks: []string{"username"}, // username is clickable
}
```

### Search Fields

Enable search on specific fields:

```go
config := &admincore.Config[User]{
    SearchFields: []string{"username", "email"},
}
```

### List Filter

Add filters to the sidebar:

```go
config := &admincore.Config[Post]{
    ListFilter: []string{"published", "created_at", "author"},
}
```

You can also use custom filter objects:

```go
import adminfilters "github.com/forgego/forge/admin/filters"

config := &admincore.Config[Post]{
    ListFilter: []admincore.Filter[Post]{
        adminfilters.NewBooleanFilter[Post]("published"),
        adminfilters.NewDateFilter[Post]("created_at"),
        adminfilters.NewRelatedFieldListFilter[Post, User](
            authorField,
            userManager,
        ),
    },
}
```

### Date Hierarchy

Add date-based navigation:

```go
config := &admincore.Config[Post]{
    DateHierarchy: "created_at",
}
```

### Ordering

Set default ordering:

```go
config := &admincore.Config[Post]{
    Ordering: []string{"-created_at", "title"}, // Descending by created_at, then ascending by title
}
```

### Read Only Fields

Make fields read-only in forms:

```go
config := &admincore.Config[User]{
    ReadOnlyFields: []string{"created_at", "last_login"},
}
```

### Fieldsets

Organize form fields into groups:

```go
config := &admincore.Config[User]{
    Fieldsets: []admincore.Fieldset[User]{
        {
            Title: "Personal Information",
            Fields: []string{"username", "email"},
        },
        {
            Title: "Permissions",
            Fields: []string{"is_active", "is_staff", "is_superuser"},
            Collapsible: true, // Optional: make collapsible
        },
        {
            Title: "Important Dates",
            Fields: []string{"created_at", "last_login"},
        },
    },
}
```

### List Per Page

Control pagination:

```go
config := &admincore.Config[Post]{
    ListPerPage: 25, // Items per page
    ListMaxShowAll: 100, // Max items for "show all"
}
```

## Advanced Customization

### Custom Queryset

Override the base queryset:

```go
config := &admincore.Config[Post]{
    GetQueryset: func(ctx context.Context, admin *admincore.Admin[Post], qs query.QuerySet[Post]) (query.QuerySet[Post], error) {
        // Only show published posts to non-staff users
        if user := GetUserFromContext(ctx); user != nil && !user.IsStaff {
            return qs.Filter(publishedField.Eq(true)), nil
        }
        return qs, nil
    },
}
```

### Custom Save Logic

Customize how models are saved:

```go
config := &admincore.Config[Post]{
    SaveModel: func(ctx context.Context, admin *admincore.Admin[Post], instance *Post, formData admincore.FormData, isNew bool) error {
        // Custom validation
        if instance.Title == "" {
            return errors.New("title is required")
        }
        
        // Set timestamps
        if isNew {
            instance.CreatedAt = time.Now()
        }
        instance.UpdatedAt = time.Now()
        
        // Call default save
        if isNew {
            return admin.Manager().Create(ctx, instance)
        }
        return admin.Manager().Update(ctx, instance)
    },
}
```

### Custom Delete Logic

Customize delete behavior:

```go
config := &admincore.Config[Post]{
    DeleteModel: func(ctx context.Context, admin *admincore.Admin[Post], instance *Post) error {
        // Soft delete instead of hard delete
        instance.Deleted = true
        return admin.Manager().Update(ctx, instance)
    },
}
```

### Custom Actions

Add custom bulk actions:

```go
import (
    "context"
    admincore "github.com/forgego/forge/admin/core"
)

config := &admincore.Config[Post]{
    Actions: []admincore.Action[Post]{
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
    },
}
```

### Inlines

Add related model editing:

```go
config := &admincore.Config[Post]{
    Inlines: []admincore.Inline[Post, Comment]{
        admincore.TabularInline[Post, Comment](
            commentManager,
            "post_id", // Foreign key field name
            []string{"author", "content", "created_at"}, // Fields to display
        ),
        // Or use StackedInline for a different layout
        admincore.StackedInline[Post, Comment](
            commentManager,
            "post_id",
            []string{"author", "content", "created_at"},
        ),
    },
}
```

## Features

### List View

The list view provides:

- **Pagination** - Automatic pagination with configurable page size
- **Search** - Full-text search on specified fields
- **Sorting** - Click column headers to sort
- **Filtering** - Filter by field values
- **Inline Editing** - Double-click cells to edit inline (coming soon)
- **Bulk Actions** - Select multiple items for bulk operations

### Detail View

View all fields of a model instance:

- Read-only display of all fields
- Related objects display
- History/audit trail (coming soon)

### Create/Edit Forms

Auto-generated forms with:

- Proper input types for different field types
- Client and server-side validation
- Rich text editing for textarea fields
- Date pickers for date/time fields
- Enhanced dropdowns for select fields
- File upload support (coming soon)

### Delete

Delete with confirmation:

- Browser confirmation dialog
- Smooth row removal with HTMX
- Cascade delete handling

## Permissions

Control access to admin with permission checking:

```go
import (
    "github.com/forgego/forge/identity"
)

// Setup permission checker
permissionChecker := identity.NewPermissionChecker(userRepo)

config := &admincore.Config[User]{
    PermissionChecker: permissionChecker,
    
    // Custom permission checks
    HasAddPermission: func(ctx context.Context, user interface{}) bool {
        return permissionChecker.HasPermission(ctx, user, "auth.add_user")
    },
    
    HasChangePermission: func(ctx context.Context, user interface{}, obj *User) bool {
        // Users can only edit themselves unless they're staff
        if staff, ok := user.(*User); ok {
            if !staff.IsStaff && staff.ID != obj.ID {
                return false
            }
        }
        return permissionChecker.HasPermission(ctx, user, "auth.change_user")
    },
    
    HasDeletePermission: func(ctx context.Context, user interface{}, obj *User) bool {
        // Only staff can delete users
        if staff, ok := user.(*User); ok {
            if !staff.IsStaff {
                return false
            }
        }
        return permissionChecker.HasPermission(ctx, user, "auth.delete_user")
    },
    
    HasViewPermission: func(ctx context.Context, user interface{}, obj *User) bool {
        return permissionChecker.HasPermission(ctx, user, "auth.view_user")
    },
}
```

## Custom Templates

You can override admin templates by placing them in your project's `templates/admin/` directory:

```
templates/
└── admin/
    ├── base.html
    ├── index.html
    ├── list.html
    ├── form.html
    └── detail.html
```

## Styling

The admin uses Bootstrap 5 and can be customized with CSS:

```css
/* Custom admin styles */
.admin-header {
    background-color: #your-brand-color;
}
```

## Best Practices

1. **Use List Display** - Show only relevant fields in list view
2. **Enable Search** - Add search fields for frequently searched columns
3. **Use Filters** - Add filters for fields users commonly filter by
4. **Organize Fieldsets** - Group related fields together
5. **Set Read-Only** - Make auto-generated fields read-only

## Examples

### Blog Admin

```go
admin.RegisterModelWithOptions(
    &models.Post{},
    admin.WithListDisplay("title", "author", "published", "created_at"),
    admin.WithSearchFields("title", "content"),
    admin.WithListFilter("published", "author", "created_at"),
    admin.WithDateHierarchy("created_at"),
    admin.WithOrdering("-created_at"),
)
```

### User Admin

```go
admin.RegisterModelWithOptions(
    &models.User{},
    admin.WithListDisplay("username", "email", "is_active", "is_staff", "date_joined"),
    admin.WithSearchFields("username", "email"),
    admin.WithListFilter("is_active", "is_staff", "date_joined"),
    admin.WithFieldsets(
        admin.Fieldset("Personal Information", "username", "email"),
        admin.Fieldset("Permissions", "is_active", "is_staff", "is_superuser"),
    ),
    admin.WithReadOnlyFields("date_joined", "last_login"),
)
```

## Export Functionality

Export your data in various formats:

```go
// Export is automatically available in the list view
// Users can export filtered results as CSV or JSON
```

The export feature:
- Exports current filtered results
- Supports CSV and JSON formats
- Respects permissions (only exports what user can view)
- Handles large datasets efficiently

## Custom Widgets

Customize form widgets for specific fields:

```go
import (
    adminwidgets "github.com/forgego/forge/admin/widgets"
)

config := &admincore.Config[Post]{
    FormFieldOverrides: map[string]admincore.Widget{
        "content": adminwidgets.NewRichTextWidget(),
        "published_at": adminwidgets.NewDateTimePicker(),
        "tags": adminwidgets.NewSelectSearchWidget(),
    },
}
```

Available widgets:
- `TextInput` - Standard text input
- `Textarea` - Multi-line text input
- `Select` - Dropdown select
- `SelectSearch` - Searchable select (for foreign keys)
- `Checkbox` - Boolean checkbox
- `RadioButtons` - Radio button group
- `DatePicker` - Date picker
- `TimePicker` - Time picker
- `DateTimePicker` - Combined date/time picker
- `RichTextEditor` - WYSIWYG editor
- `FileUpload` - File upload widget
- `ImageUpload` - Image upload with preview

## Type Safety

The admin system is fully type-safe using Go generics:

```go
// Type-safe field expressions
usernameField := admincore.StringField[User](
    "username",
    func(u *User) string { return u.Username },
    func(u *User, v string) { u.Username = v },
)

// Type-safe filters
activeFilter := adminfilters.NewBooleanFilter[User](usernameField)

// Type-safe actions
action := admincore.NewAction[User](
    "activate",
    "Activate users",
    func(ctx context.Context, users []*User) error {
        // users is []*User, fully typed
        for _, user := range users {
            user.IsActive = true
        }
        return nil
    },
)
```

## Performance Optimization

### Select Related

Optimize queries with select_related:

```go
config := &admincore.Config[Post]{
    ListSelectRelated: []string{"author"}, // Join author in list view
}
```

### Prefetch Related

Optimize many-to-many and reverse foreign keys:

```go
config := &admincore.Config[Post]{
    ListPrefetchRelated: []string{"comments", "tags"}, // Prefetch related objects
}
```

## Complete Example

Here's a complete example with all features:

```go
package blog

import (
    "context"
    admincore "github.com/forgego/forge/admin/core"
    "github.com/forgego/forge/admin/http"
    "github.com/forgego/forge/schema"
    query "github.com/forgego/forge/orm"
)

func InitPostAdmin(postManager *query.Manager[Post], commentManager *query.Manager[Comment]) (*admincore.Admin[Post], error) {
    postSchema := &PostSchema{}
    
    admin, err := admincore.Register[Post](
        schema.NewSchema(postSchema),
        postManager,
        &admincore.Config[Post]{
            // List configuration
            ListDisplay: []string{"title", "author", "published", "created_at"},
            ListDisplayLinks: []string{"title"},
            ListPerPage: 25,
            Ordering: []string{"-created_at"},
            
            // Search and filters
            SearchFields: []string{"title", "content"},
            ListFilter: []string{"published", "author", "created_at"},
            DateHierarchy: "created_at",
            
            // Form configuration
            Fieldsets: []admincore.Fieldset[Post]{
                {Title: "Post", Fields: []string{"title", "content", "author"}},
                {Title: "Publishing", Fields: []string{"published"}},
            },
            ReadOnlyFields: []string{"created_at", "updated_at"},
            
            // Inlines
            Inlines: []admincore.Inline[Post, Comment]{
                admincore.TabularInline[Post, Comment](
                    commentManager,
                    "post_id",
                    []string{"author", "content", "created_at"},
                ),
            },
            
            // Actions
            Actions: []admincore.Action[Post]{
                admincore.NewAction[Post](
                    "publish",
                    "Publish selected",
                    func(ctx context.Context, posts []*Post) error {
                        for _, post := range posts {
                            post.Published = true
                            postManager.Update(ctx, post)
                        }
                        return nil
                    },
                ),
            },
            
            // Performance
            ListSelectRelated: []string{"author"},
            ListPrefetchRelated: []string{"comments"},
        },
    )
    
    if err != nil {
        return nil, err
    }
    
    http.RegisterAdminForHTTP(admin)
    return admin, nil
}
```

## Next Steps

- [Admin Tutorial](/docs/tutorials/03-admin-interface) - Step-by-step tutorial
- [REST API Guide](/docs/guides/rest-api) - Build APIs for your frontend
- [Security Guide](/docs/guides/security) - Secure your admin interface
- [Advanced Topics](/docs/advanced/plugins) - Extend the admin with plugins

