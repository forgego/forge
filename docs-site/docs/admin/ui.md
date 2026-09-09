---
sidebar_position: 5
description: UI architecture, dashboard widgets, Recharts visualizations, custom plugin pages, and theming.
image: /forge-social-card.svg
---

# Admin UI, Widgets & Custom Plugins

The Forge Admin frontend is a standalone React 19 Single Page Application that communicates with your Go backend over a type-safe JSON API.

---

## Dashboard Widgets

Embed analytics metrics, charts, and summary cards on the admin home page or model-specific headers using `admin.Widget`:

```go
package main

import (
    "context"
    "github.com/forgego/forge/admin"
)

func SetupDashboard(site *admin.Site) {
    site.AddWidget(admin.Widget{
        ID:    "total_revenue",
        Title: "Total Revenue (30d)",
        Type:  admin.WidgetMetric,
        Width: admin.WidthHalf,
        DataHandler: func(ctx context.Context) (any, error) {
            revenue, _ := models.OrderManager.
                Filter(models.OrderExpr.Status.Eq("delivered")).
                Aggregate(ctx, orm.Sum("total_amount"))
            return admin.MetricData{
                Value:  fmt.Sprintf("$%.2f", revenue["sum"]),
                Change: "+14.2%",
                Trend:  "up",
            }, nil
        },
    })

    site.AddWidget(admin.Widget{
        ID:    "order_volume_chart",
        Title: "Orders by Status",
        Type:  admin.WidgetChartDonut,
        Width: admin.WidthHalf,
        DataHandler: func(ctx context.Context) (any, error) {
            return []admin.ChartPoint{
                {Label: "Delivered", Value: 340},
                {Label: "Processing", Value: 85},
                {Label: "Pending", Value: 24},
            }, nil
        },
    })
}
```

---

## Supported Widget Types

| Widget Type | Visualization | Common Use Case |
| :--- | :--- | :--- |
| `WidgetMetric` | KPI summary card with trend percentage and badge. | Revenue, active users, total orders. |
| `WidgetChartLine` | Recharts responsive line chart with time series tooltips. | Sales trends over 30/90 days. |
| `WidgetChartBar` | Recharts categorical bar chart. | Sales by product category. |
| `WidgetChartDonut` | Recharts donut chart with interactive hover segments. | Order status distribution, customer tiers. |
| `WidgetTable` | Compact recent activity list table. | Latest 5 orders, low-stock inventory alerts. |

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

