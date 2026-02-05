---
sidebar_position: 4
---

# Project Structure

Understanding forge's project structure helps you organize your code effectively.

## Default structure

When you create a new forge project, you get this structure:

```
myproject/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── app/                          # Django-style apps
│   └── blog/
│       ├── models.go            # ORM models
│       ├── admin.go             # Admin configuration
│       └── api.go               # REST/GraphQL endpoints
├── migrations/                   # Database migrations
├── config/
│   └── config.yaml              # Configuration
├── static/                       # Static files
├── templates/                    # HTML templates
├── go.mod
└── README.md
```

## App structure

Each app follows a flat file structure:

- `models.go` - ORM model definitions
- `admin.go` - Admin configuration (auto-registers via `init()`)
- `api.go` - REST/GraphQL API endpoints (auto-registers via `init()`)
- `handlers.go` - HTTP handlers (optional)
- `services.go` - App-level services (optional)

## Key directories

### `cmd/server/`

Contains your application entry point:

```go
// cmd/server/main.go
package main

import (
    "github.com/forgego/forge/config"
    "github.com/forgego/forge/db"
    // ...
)

func main() {
    // Application setup
}
```

### `app/`

Contains your Django-style apps. Each app is a self-contained module:

```
app/
├── users/
│   ├── models.go
│   ├── admin.go
│   └── api.go
└── blog/
    ├── models.go
    ├── admin.go
    └── api.go
```

### `migrations/`

Contains database migration files:

```
migrations/
├── 000001_initial.up.sql
├── 000001_initial.down.sql
├── 000002_add_users.up.sql
└── 000002_add_users.down.sql
```

### `config/`

Contains configuration files:

```
config/
├── config.yaml          # Main configuration
├── development.yaml     # Development overrides
└── production.yaml      # Production overrides
```

## Advanced structure (optional)

For larger projects, you can use an advanced structure:

```
myproject/
├── cmd/server/main.go
├── app/                  # Django-style apps
├── domain/               # Pure business logic (optional)
│   ├── user/
│   │   ├── entity.go
│   │   └── service.go
│   └── billing/
│       └── service.go
├── infra/                # Infrastructure (optional)
│   ├── db/
│   ├── redis/
│   └── email/
└── pkg/                  # Shared utilities (optional)
    ├── validators/
    └── utils/
```

## Best practices

### 1. Keep apps focused

Each app should represent a single domain concept:

```
✅ Good:
app/
├── users/        # User management
├── blog/          # Blog functionality
└── shop/          # E-commerce

❌ Bad:
app/
└── everything/    # Too broad
```

### 2. Use flat files

Keep models, admin, and API in single files per app:

```
✅ Good:
app/blog/models.go
app/blog/admin.go
app/blog/api.go

❌ Bad:
app/blog/models/user.go
app/blog/models/post.go
app/blog/models/comment.go
```

### 3. Auto-registration

Use `init()` functions for automatic discovery:

```go
// app/blog/admin.go
package blog

import "github.com/forgego/forge/admin"

func init() {
    admin.Register(&admin.Config[Post]{})
    admin.Register(&admin.Config[Comment]{})
}
```

### 4. Shared code

Use `pkg/` for cross-app utilities:

```
pkg/
├── validators/    # Shared validators
├── utils/         # Shared utilities
└── types/         # Shared types
```

### 5. Domain logic

Use `domain/` for pure business logic:

```
domain/
├── user/
│   ├── entity.go      # User entity
│   ├── service.go     # User business logic
│   └── repository.go   # User data access interface
```

## File organization

### Models

Keep related models in the same file:

```go
// app/blog/models.go
package blog

type Post struct { /* ... */ }
type Comment struct { /* ... */ }
type Tag struct { /* ... */ }
```

### Admin

Register all models in one file:

```go
// app/blog/admin.go
package blog

func init() {
    admin.Register(&admin.Config[Post]{})
    admin.Register(&admin.Config[Comment]{})
    admin.Register(&admin.Config[Tag]{})
}
```

### API

Group related endpoints:

```go
// app/blog/api.go
package blog

func init() {
    RegisterPostRoutes(router)
    RegisterCommentRoutes(router)
}
```

## Next steps

- **[Models Guide](/docs/guides/models)** - Learn about models
- **[Admin Guide](/docs/guides/admin)** - Configure admin interface
- **[REST API Guide](/docs/guides/rest-api)** - Build APIs
