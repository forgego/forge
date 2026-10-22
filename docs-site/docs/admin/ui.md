---
sidebar_position: 5
description: UI architecture, dashboard layout, charts, custom plugin pages, and theming.
image: /social-card.png
---

# Admin UI, Dashboard & Custom Plugins

The Forge Admin frontend is a standalone React 19 Single Page Application that communicates with your Go backend over a type-safe JSON API.

---

## Dashboard Layout

By default, the dashboard shows model and record counts, plugins, and a model catalog; it does not include an analytics chart. To show Recharts analytics charts, configure a server-driven dashboard with `core.SetDashboard` and `components.Chart`. That layout is global; model-specific widget pinning and `Site.AddWidget` are not available.

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

Register a plugin implementing `admin.Plugin` to add server-driven pages and sidebar items. Plugin pages are built from `admin/components` values, not custom React bundles:

```go
type ReportsPlugin struct { core.BasePlugin }

func (*ReportsPlugin) ID() string { return "reports" }
func (*ReportsPlugin) Name() string { return "Reports" }
func (*ReportsPlugin) GetPages() map[string]components.Component {
    return map[string]components.Component{
        "overview": components.Page("Reports").WithChildren(components.Text("Sales overview")),
    }
}
func (*ReportsPlugin) GetMenuItems() []core.MenuItem {
    return []core.MenuItem{{Label: "Reports", Path: "/plugins/reports/pages/overview"}}
}

err := site.RegisterPlugin(ctx, &ReportsPlugin{})
```

The plugin provides menu items through `GetMenuItems`. Check `RegisterPlugin`'s error; duplicate IDs are rejected. Implement permission checks in the plugin's data operations.

---

## Theming & Brand Customization

Set the site's title directly. `admin.UIConfig` controls the asset source and URL prefix. There is no `SiteConfig` with logo, favicon, or theme fields:

```go
site := admin.NewSite("commerce")
site.Title = "Acme Commerce"
ui := admin.DefaultUIConfig()
ui.Prefix = "/admin"
site.WithUIConfig(ui)
```

---

## Next Steps

- **[Admin Overview](/docs/admin/overview/)**: Architecture and security overview.
- **[Model Configuration](/docs/admin/config/)**: Detailed list view configuration.
- **[REST API Overview](/docs/api/overview/)**: Explore the REST API framework.

