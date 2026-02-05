# Forge Admin - Complete Integration Guide

## Overview

This guide covers the complete integration of the redesigned Forge Admin system with your application.

## Architecture Summary

The new admin system is organized into these key packages:

- **core/**: Core admin types (`Admin[T]`, `Config[T]`, `Registry`, `Site`)
- **schema/**: Schema integration (auto-discovery, field/relation mapping)
- **orm/**: ORM integration (wrappers for Manager and QuerySet)
- **filter/**: Filter integration (auto-generated filters from schema)
- **fields/**: Type-safe field system
- **widgets/**: Form widget system
- **views/**: View implementations (List, Form, Detail)
- **templates/**: Template engine and rendering
- **components/**: Reusable UI components
- **http/**: HTTP handlers and routing
- **advanced/**: Advanced features (Actions, History, Permissions)
- **security/**: Security features (CSRF, XSS)
- **codegen/**: Code generation tools
- **testing/**: Testing helpers

## Complete Setup Example

### 1. Define Your Model and Schema

```go
package models

import (
    "github.com/forgego/forge/schema"
)

type User struct {
    ID        int64  `db:"id"`
    Username  string `db:"username"`
    Email     string `db:"email"`
    IsActive  bool   `db:"is_active"`
    CreatedAt time.Time `db:"created_at"`
}

type UserSchema struct{}

func (s *UserSchema) Fields() []schema.Field {
    return []schema.Field{
        {
            Name:        "ID",
            Type:        schema.TypeInt64,
            PrimaryKey:  true,
            AutoIncrement: true,
        },
        {
            Name:        "Username",
            Type:        schema.TypeString,
            Required:    true,
            MaxLength:   intPtr(150),
            VerboseName: "Username",
        },
        {
            Name:        "Email",
            Type:        schema.TypeEmail,
            Required:    true,
            VerboseName: "Email",
        },
        {
            Name:        "IsActive",
            Type:        schema.TypeBool,
            Default:     true,
            VerboseName: "Is Active",
        },
        {
            Name:        "CreatedAt",
            Type:        schema.TypeDateTime,
            AutoNowAdd:  true,
            Editable:    false,
            VerboseName: "Created At",
        },
    }
}

func (s *UserSchema) Relations() []schema.Relation {
    return []schema.Relation{}
}

func (s *UserSchema) Meta() schema.Meta {
    return schema.Meta{
        TableName:         "users",
        VerboseName:       "User",
        VerboseNamePlural: "Users",
        AppLabel:          "auth",
    }
}

func (s *UserSchema) Hooks() *schema.ModelHooks {
    return nil
}
```

### 2. Register Admin

```go
package main

import (
    "log"
    
    admincore "github.com/forgego/forge/admin/core"
    "github.com/forgego/forge/orm"
    "github.com/forgego/forge/schema"
    "your-app/models"
)

func setupAdmin() {
    // Create schema instance
    schemaInstance := &models.UserSchema{}
    
    // Get ORM manager (from your ORM setup)
    manager := orm.NewManager[models.User](db)
    
    // Create admin configuration
    config := &admincore.Config[models.User]{
        VerboseName:       "User",
        VerboseNamePlural: "Users",
        ListPerPage:       25,
        
        // Custom list display
        ListDisplay: []interface{}{
            "Username",
            "Email",
            "IsActive",
        },
        
        // Search fields
        SearchFields: []interface{}{
            "Username",
            "Email",
        },
        
        // Ordering
        Ordering: []admincore.Ordering[models.User]{
            *admincore.OrderBy[models.User]("CreatedAt").Desc(),
        },
        
        // Actions
        Actions: []admincore.Action[models.User]{
            admincore.NewAction[models.User](
                "activate",
                "Activate selected users",
                func(ctx context.Context, instances []*models.User) error {
                    for _, instance := range instances {
                        instance.IsActive = true
                        if err := manager.Update(ctx, instance); err != nil {
                            return err
                        }
                    }
                    return nil
                },
            ),
        },
    }
    
    // Register admin
    admin, err := admincore.Register(schemaInstance, manager, config)
    if err != nil {
        log.Fatal(err)
    }
    
    _ = admin // Admin is now registered in global registry
}
```

### 3. Set Up HTTP Routes

```go
package main

import (
    "log"
    "net/http"
    
    admincore "github.com/forgego/forge/admin/core"
    adminhttp "github.com/forgego/forge/admin/http"
    httplib "github.com/forgego/forge/server"
)

func setupRoutes() {
    // Get the global registry
    registry := admincore.GetGlobalRegistry()
    
    // Create router with template directory
    router := adminhttp.NewCoreRouter(registry, "./templates/admin")
    
    // Create HTTP router
    httpRouter := httplib.NewRouter()
    
    // Register admin routes
    router.RegisterRoutes(httpRouter, "/admin")
    
    // Serve static files
    httpRouter.Static("/admin/static", "./static/admin")
    
    // Start server
    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", httpRouter))
}
```

### 4. Template Setup

Create templates in `./templates/admin/`:

- `base.html` - Base layout
- `list.html` - List view
- `form.html` - Create/update form
- `detail.html` - Detail view
- `index.html` - Admin dashboard

Templates are already provided in `forge/admin/templates/templates/` - copy them to your template directory.

### 5. Static Assets

Copy static assets from `forge/admin/static/` to your static directory:

- `css/admin.css` - Admin styles
- `js/admin.js` - Admin JavaScript

## Features

### Auto-Discovery

Fields, relations, and metadata are automatically discovered from your schema. No manual field definitions needed!

### Type Safety

All operations are type-safe using Go generics:

```go
admin := &admincore.Admin[User]{}
manager := admin.Manager()  // Returns *AdminManager[User]
filterset := admin.FilterSet()  // Returns *AdminFilterSet[User]
```

### Customization

Customize any aspect:

```go
config := &admincore.Config[User]{
    // List view
    ListDisplay: []interface{}{"Name", "Email"},
    SearchFields: []interface{}{"Name", "Email"},
    
    // Form view
    Fieldsets: []admincore.Fieldset[User]{
        admincore.NewFieldset[User]("Basic Info", "Name", "Email"),
    },
    
    // Custom queryset
    GetQueryset: func(ctx context.Context, admin *admincore.Admin[User], qs orm.QuerySet[User]) (orm.QuerySet[User], error) {
        // Custom filtering
        return qs.Filter(/* custom filters */), nil
    },
}
```

### Bulk Actions

```go
config.Actions = []admincore.Action[User]{
    admincore.NewAction[User](
        "delete_selected",
        "Delete selected",
        func(ctx context.Context, instances []*User) error {
            for _, instance := range instances {
                admin.DeleteModel(ctx, instance)
            }
            return nil
        },
    ).WithPermissions("users.delete_user"),
}
```

### Export

Export data in CSV or JSON format:

```http
GET /admin/users/export/?format=csv
GET /admin/users/export/?format=json
```

### Autocomplete

Autocomplete for foreign key fields:

```http
GET /admin/users/autocomplete/?field=author&search=john&limit=10
```

## Migration from Old System

1. Replace `admin.Register` with `admincore.Register`
2. Update imports to use `admincore` package
3. Update HTTP handlers to use `CoreHandler` and `CoreRouter`
4. Move templates to new structure
5. Update field definitions to use schema-first approach

## Testing

Use testing helpers:

```go
import "github.com/forgego/forge/admin/testing"

helper, err := testing.NewTestAdminHelper(schemaInstance, manager, config)
// Use helper for testing
```

## Code Generation

Generate admin code:

```go
import "github.com/forgego/forge/admin/codegen"

generator := codegen.NewAdminGenerator("./output")
generator.GenerateAdmin("User", "models", fields)
```

## Complete Example

See `forge/admin/examples/basic_usage.go` for a complete working example.
