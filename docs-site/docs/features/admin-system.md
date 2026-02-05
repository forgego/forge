---
id: admin-system
title: Admin System
sidebar_label: Admin System
---

# Admin System Feature

## Overview

The Admin System provides a Django-like admin interface with full type safety using Go generics. It includes complete CRUD operations, filtering, searching, widgets, actions, and export functionality.

## Location

`forge/admin/`

## Status

✅ **Complete** - Production ready

## Core Components

### 1. Admin Interface (`admin.go`)

Type-safe admin interface:

```go
type Admin[T any] struct {
    model   T
    manager *query.Manager[T]
    config  *Config[T]
    name    string
}
```

**Key Methods:**
- `Register()` - Register a model with admin
- `GetQueryset()` - Get base queryset
- `SaveModel()` - Save model instance
- `DeleteModel()` - Delete model instance
- `HasAddPermission()` - Check add permission
- `HasChangePermission()` - Check change permission
- `HasDeletePermission()` - Check delete permission

### 2. Admin Configuration (`config.go`)

Complete configuration system with all Django admin options:

**List View Configuration:**
- `ListDisplay` - Fields to display in list view
- `ListDisplayLinks` - Fields that link to detail view
- `ListEditable` - Fields editable in list view
- `ListFilter` - Filters for list view
- `SearchFields` - Fields searchable in list view
- `DateHierarchy` - Field for date-based navigation
- `Ordering` - Default ordering
- `ListPerPage` - Items per page
- `ListMaxShowAll` - Maximum items for "show all"
- `ListSelectRelated` - Fields for select_related optimization
- `ListPrefetchRelated` - Fields for prefetch_related optimization
- `ShowFullResultCount` - Show full vs filtered count
- `PreserveFilters` - Preserve filters when navigating
- `EmptyValueDisplay` - Display value for empty/null fields
- `SortableBy` - Fields that can be sorted

**Form Configuration:**
- `Fields` - Explicit field ordering
- `Exclude` - Fields to exclude
- `ReadOnlyFields` - Read-only fields
- `PrepopulatedFields` - Auto-populated fields
- `RawIDFields` - Raw ID input for foreign keys
- `AutocompleteFields` - Autocomplete for foreign keys
- `RadioFields` - Radio buttons for choices
- `Fieldsets` - Form field grouping
- `GetForm` - Custom form generator
- `GetFields` - Dynamic field list
- `GetFieldsets` - Dynamic fieldsets
- `GetReadOnlyFields` - Dynamic read-only fields
- `GetPrepopulatedFields` - Dynamic prepopulated fields

**Related Models:**
- `Inlines` - Related model editing

**Actions:**
- `Actions` - Bulk actions

**Permissions:**
- `PermissionChecker` - Permission checker
- `HasAddPermission` - Add permission check
- `HasChangePermission` - Change permission check
- `HasDeletePermission` - Delete permission check
- `HasViewPermission` - View permission check
- `HasModulePermission` - Module permission check

**View Customization:**
- `GetQueryset` - Custom queryset hook
- `SaveModel` - Custom save hook
- `DeleteModel` - Custom delete hook

**Advanced Features:**
- `SaveAs` - Show "Save as new" button
- `SaveAsContinue` - Continue editing after "save as new"
- `SaveOnTop` - Show save buttons at top
- `ViewOnSite` - Get URL for viewing object
- `ShowChangeLink` - Show change link in inlines

**Form Field Customization:**
- `FormFieldOverrides` - Override widgets for fields
- `FormFieldForForeignKey` - Custom foreign key field
- `FormFieldForManyToMany` - Custom many-to-many field

### 3. List View (`list_view.go`)

Complete list view implementation:
- Pagination
- Filtering
- Searching
- Sorting
- Bulk actions
- Export

### 4. Detail View (`detail_view.go`)

Detail view for viewing/editing individual objects:
- Form rendering
- Field display
- Inlines
- Actions

### 5. Form View (`form_view.go`)

Form handling for create/update:
- Form generation
- Form validation
- Form saving
- Form errors

### 6. Filters (`filters.go`, `filters_advanced.go`)

Filter system:
- Boolean filters
- Choice filters
- Date filters
- Number filters
- Text filters
- Related filters
- Custom filters

### 7. Widgets (`widgets.go`, `widgets_advanced.go`)

Form widgets:
- Text input
- Textarea
- Select
- Checkbox
- Radio buttons
- Date picker
- Time picker
- DateTime picker
- File upload
- Image upload
- Rich text editor
- Autocomplete
- Raw ID

### 8. Actions (`actions.go`)

Bulk actions:
- Type-safe action definition
- Action execution
- Action permissions
- Action confirmation

### 9. Export (`export.go`)

Export functionality:
- CSV export
- JSON export
- Custom export formats

### 10. Inlines (`inlines.go`)

Related model editing:
- Tabular inline
- Stacked inline
- Inline permissions
- Inline validation

### 11. Fieldsets (`fieldsets.go`)

Form field grouping:
- Fieldset definition
- Collapsible fieldsets
- Fieldset permissions

### 12. Permissions (`permissions.go`)

Permission system:
- Permission checking
- Permission decorators
- Custom permissions

### 13. HTTP Handlers (`http/`)

Complete HTTP handlers:
- List handler
- Detail handler
- Create handler
- Update handler
- Delete handler
- Action handler
- Export handler

### 14. Registry (`registry.go`)

Admin model registry:
- Model registration
- Model lookup
- Global registry

## Features

### ✅ Complete Features

1. **Type-Safe Admin** - Full type safety with generics
2. **List View** - Pagination, filtering, searching, sorting
3. **Detail View** - View and edit individual objects
4. **Form Handling** - Create and update forms
5. **Filters** - Advanced filtering system
6. **Search** - Full-text search
7. **Widgets** - Rich form widgets
8. **Actions** - Bulk actions
9. **Export** - CSV and JSON export
10. **Inlines** - Related model editing
11. **Fieldsets** - Form field grouping
12. **Permissions** - Permission system
13. **HTTP Handlers** - Complete HTTP handlers
14. **Registry** - Model registry

## Usage Examples

### Basic Admin Registration

```go
userAdmin := admin.Register(
    &models.User{},
    models.User.Objects,
    &admin.Config[*models.User]{
        ListDisplay: []admin.FieldExpr[*models.User, any]{
            admin.StringField("username", getter, setter),
            admin.StringField("email", getter, setter),
            admin.BoolField("is_active", getter, setter),
        },
        ListFilter: []admin.Filter[*models.User]{
            admin.NewBooleanFilter(
                admin.BoolField("is_active", getter, setter),
            ),
        },
        SearchFields: []admin.FieldExpr[*models.User, any]{
            admin.StringField("username", getter, setter),
            admin.StringField("email", getter, setter),
        },
    },
)
```

### Admin with Inlines

```go
postAdmin := admin.Register(
    &models.Post{},
    models.Post.Objects,
    &admin.Config[*models.Post]{
        ListDisplay: []admin.FieldExpr[*models.Post, any]{
            admin.StringField("title", getter, setter),
            admin.StringField("author", getter, setter),
        },
        Inlines: []admin.Inline[*models.Post, any]{
            admin.TabularInline(
                &models.Comment{},
                models.Comment.Objects,
                parentField,
                []admin.FieldExpr[*models.Comment, any]{
                    admin.StringField("author", getter, setter),
                    admin.StringField("content", getter, setter),
                },
            ),
        },
    },
)
```

### Admin with Actions

```go
userAdmin := admin.Register(
    &models.User{},
    models.User.Objects,
    &admin.Config[*models.User]{
        Actions: []admin.Action[*models.User]{
            admin.NewAction(
                "activate",
                "Activate selected users",
                func(ctx context.Context, users []*models.User) error {
                    for _, user := range users {
                        user.IsActive = true
                        if err := models.User.Objects.Update(ctx, user); err != nil {
                            return err
                        }
                    }
                    return nil
                },
            ),
        },
    },
)
```

### Admin with Fieldsets

```go
userAdmin := admin.Register(
    &models.User{},
    models.User.Objects,
    &admin.Config[*models.User]{
        Fieldsets: []admin.Fieldset[*models.User]{
            admin.NewFieldset(
                "Account Information",
                admin.StringField("username", getter, setter),
                admin.StringField("email", getter, setter),
            ),
            admin.NewFieldset(
                "Permissions",
                admin.BoolField("is_active", getter, setter),
                admin.BoolField("is_staff", getter, setter),
            ),
        },
    },
)
```

## Integration Points

### ORM System
- Uses ORM QuerySet for data access
- Type-safe field expressions
- Manager integration

### Filter System
- Filter integration
- Query parsing
- Security validation

### HTTP Server
- HTTP handlers
- Routing integration
- Middleware support

### Identity System
- Permission integration
- Authentication middleware
- User management

## Extension Points

### Custom Widgets
- Define custom widgets
- Custom widget rendering
- Custom widget parsing

### Custom Filters
- Define custom filters
- Custom filter logic
- Custom filter widgets

### Custom Actions
- Define custom actions
- Custom action logic
- Custom action permissions

### Custom Forms
- Custom form generation
- Custom form validation
- Custom form saving

## Best Practices

1. **Type Safety** - Use type-safe field expressions
2. **Permissions** - Always check permissions
3. **Validation** - Validate all input
4. **Performance** - Use select_related and prefetch_related
5. **User Experience** - Provide helpful error messages
6. **Security** - Sanitize all output

## Future Enhancements

- [ ] HTML template rendering
- [ ] Rich text editor
- [ ] File/image uploads
- [ ] History/audit logging
- [ ] Autocomplete UI improvements
- [ ] Date/time picker improvements
