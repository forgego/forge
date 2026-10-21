---
sidebar_position: 5
description: UI architecture, dashboard layout, charts, custom plugin pages, and theming.
image: /social-card.png
---

# Admin UI, Dashboard & Custom Plugins

The Forge Admin frontend is a standalone React 19 Single Page Application that communicates with your Go backend over a type-safe JSON API.

---

## Dashboard Layout

The main dashboard uses a server-driven component tree. Configure it with `core.SetDashboard`. The layout is global; model-specific widget pinning and `Site.AddWidget` are not available.

```go
package main

import (
    "github.com/forgego/forge/admin/components"
    "github.com/forgego/forge/admin/core"
)

func SetupDashboard() {
    points := []map[string]any{
        {"name": "Mon", "orders": 4},
        {"name": "Tue", "orders": 7},
    }
    core.SetDashboard(core.DashboardConfig{
        Title: "Sales",
        Layout: components.Grid(12).WithChildren(
            components.Stats("Orders", "11", "+3").WithColSpan(4),
            components.Chart("Orders by day", "line", points).WithColSpan(8),
        ),
    })
}
```

---

## Dashboard Components

| Component | Rendering | Common Use Case |
| :--- | :--- | :--- |
| `components.Stats(label, value, change)` | Summary card with a value and optional change text. | Revenue or order totals. |
| `components.Chart(title, "line", data)` | Recharts line chart with tooltips. | Time series. |
| `components.Chart(title, "bar", data)` | Recharts bar chart with tooltips. | Category comparisons. |
| `components.Chart(title, "area", data)` | Recharts area chart with tooltips. | Trends. |

The dashboard renderer does not currently register a table component. Use the model list pages for tabular records.

---

## Custom Plugin Pages

When your application requires dedicated interfaces beyond standard CRUD (such as batch billing engines, marketing email composers, or live logistics maps), register a **Custom Plugin**:

```go
site.RegisterPlugin(admin.Plugin{
    ID:          "sales_reports",
    Label:       "Sales Analytics",
    Icon:        "bar-chart-2",
    Route:       "/reports/sales",
    Permissions: []string{"admin", "finance"},
    Handler: func(w http.ResponseWriter, r *http.Request) {
        // Return custom JSON data or server-rendered React component bundle
        data := GenerateSalesReport(r.Context())
        json.NewEncoder(w).Encode(data)
    },
})
```

The plugin automatically appears in the Admin sidebar navigation, respecting user RBAC role permissions.

---

## Theming & Brand Customization

Customize the Admin brand and visual appearance in `admin.SiteConfig`:

```go
site := admin.NewSite(admin.SiteConfig{
    Title:     "Acme Commerce",
    BrandLogo: "/static/brand-logo.svg",
    Favicon:   "/static/favicon.ico",
    Theme: admin.ThemeConfig{
        DefaultColorMode: "dark", // "dark", "light", or "system"
        PrimaryColor:     "#0284c7",
    },
})
```

---

## Next Steps

- **[Admin Overview](/docs/admin/overview/)**: Architecture and security overview.
- **[Model Configuration](/docs/admin/config/)**: Detailed list view configuration.
- **[REST API Overview](/docs/api/overview/)**: Explore the REST API framework.

