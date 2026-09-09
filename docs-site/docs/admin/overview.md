---
sidebar_position: 1
description: Modern React 19 Admin SPA, declarative configuration, RBAC, audit logging, and custom plugins.
image: /forge-social-card.svg
---

# Admin Console Overview

Forge includes a production-grade, auto-generated **React 19 Single Page Application (SPA)** that mounts directly to your Go backend. Inspired by Django Admin, it provides complete data management capabilities out of the box with zero frontend engineering required.

---

## Architectural Highlights

- **React 19 & TanStack Router**: Ultra-fast client-side navigation with deep linking, stateful search parameters, and optimistic UI updates.
- **Declarative Go Registration**: Register any model with search fields, list filters, custom actions, and dashboard widgets in a single Go statement.
- **Enterprise Security & Auth**: Session token cookies with SHA-256 hashing, protection against session fixation, and granular Role-Based Access Control (RBAC).
- **Modern Design**: Built with Tailwind CSS, Radix UI primitives, Lucide icons, full dark/light theme switching, and keyboard command palette (`Cmd+K` / `Ctrl+K`).
- **Extensible Plugin System**: Embed custom React dashboards, analytics reports, and external workflows directly inside the admin navigation.

---

## Quick Example

Mounting the Admin Console into your Chi router:

```go
package main

import (
    "github.com/go-chi/chi/v5"
    "github.com/forgego/forge/admin"
    "myapp/models"
)

func SetupAdmin(r chi.Router) {
    // 1. Create the Admin Site
    adminSite := admin.NewSite(admin.SiteConfig{
        Title:       "Forge Commerce Admin",
        BasePath:    "/admin",
        RequireAuth: true,
    })

    // 2. Register models with declarative configuration
    admin.Register[models.Product](adminSite, admin.ModelConfig[models.Product]{
        ListDisplay:  []string{"SKU", "Name", "Category", "Price", "InStock"},
        SearchFields: []string{"SKU", "Name"},
        ListFilter:   []string{"InStock", "Category"},
        Ordering:     []string{"-created_at"},
        Actions: []admin.Action{
            {
                ID:             "restock",
                Label:          "Restock Inventory (+50)",
                ConfirmMessage: "Are you sure you want to add 50 units to selected products?",
                Handler: func(ctx context.Context, ids []any) error {
                    return InventoryService.Restock(ctx, ids, 50)
                },
            },
        },
    })

    // 3. Mount HTTP and static assets onto Chi router
    adminSite.Mount(r)
}
```

Now navigate your browser to `http://localhost:8000/admin/`.

---

## Authentication & Security

The Admin Console includes complete security defaults:

### Admin Credentials
Configure administrative credentials via environment variables or configuration files:

```bash
export FORGE_ADMIN_USERNAME="admin"
export FORGE_ADMIN_PASSWORD="super-secret-password"
```

### Session Security
- **SHA-256 Token Hashing**: Plaintext session tokens are sent to the client as an `HttpOnly`, `SameSite=Strict`, `Secure` cookie, while only the SHA-256 hash is persisted in the session repository.
- **Timing-Safe Comparison**: Authentication uses constant-time string comparison to defend against timing attacks.
- **Brute-Force Rate Limiting**: Consecutive failed login attempts trigger progressive delays.

---

## Role-Based Access Control (RBAC)

Forge provides granular permission hooks on each registered model:

```go
admin.Register[models.Order](adminSite, admin.ModelConfig[models.Order]{
    HasViewPermission: func(ctx context.Context, u *admin.User) bool {
        return u.HasRole("admin", "support", "finance")
    },
    HasAddPermission: func(ctx context.Context, u *admin.User) bool {
        return u.HasRole("admin")
    },
    HasChangePermission: func(ctx context.Context, u *admin.User) bool {
        return u.HasRole("admin", "support")
    },
    HasDeletePermission: func(ctx context.Context, u *admin.User) bool {
        return u.HasRole("admin") // Only superadmins can delete orders
    },
})
```

If a user lacks permission, actions and mutation buttons are automatically hidden from the UI, and corresponding API endpoints reject requests with HTTP 403 Forbidden.

---

## Core Capabilities

1. **List & Table Views**: Sortable columns, live search with debouncing, multi-select checkboxes for batch operations.
2. **Faceted Filtering**: Filter by category, dates, boolean flags, or custom dropdowns with instant client-side URL sync.
3. **Saved Views**: Bookmark frequently used filter and search configurations as named views for quick access.
4. **Change History & Audit Logs**: Inspect detailed timelines of who modified each record and what fields were changed.
5. **Data Export**: One-click export of filtered datasets to CSV or JSON formats.
6. **Dashboard Widgets**: Embed Recharts analytics charts, KPI metric cards, and low-stock alerts.
7. **Custom Plugins**: Extend the sidebar navigation with custom React pages and reports.

---

## Next Steps

- **[Admin Configuration Reference](/docs/admin/config/)**: Detailed breakdown of `ModelConfig` options.
- **[Custom Actions](/docs/admin/actions/)**: Build batch processing actions with confirmation dialogs.
- **[Filters & Search](/docs/admin/filters/)**: Advanced list filtering configurations.
- **[UI Customization & Plugins](/docs/admin/ui/)**: Customize theme and register custom plugin pages.

