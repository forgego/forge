package main

import (
	"github.com/forgego/forge/admin/components"
	"github.com/forgego/forge/admin/core"
)

// ReportsPlugin is a custom plugin for the ecommerce admin
type ReportsPlugin struct {
	core.BasePlugin
}

func (p *ReportsPlugin) ID() string {
	return "reports"
}

func (p *ReportsPlugin) Name() string {
	return "Reports"
}

func (p *ReportsPlugin) GetMenuItems() []core.MenuItem {
	return []core.MenuItem{
		{
			Label: "Reports",
			Icon:  "BarChart",
			Path:  "/plugins/reports/pages/overview",
			Children: []core.MenuItem{
				{
					Label: "Overview",
					Path:  "/plugins/reports/pages/overview",
				},
				{
					Label: "Sales",
					Path:  "/plugins/reports/pages/sales",
				},
			},
		},
	}
}

func (p *ReportsPlugin) GetPages() map[string]components.Component {
	return map[string]components.Component{
		"overview": components.Page("Reports Overview").WithChildren(
			components.Grid(2).WithChildren(
				components.Card("Sales Report").WithChildren(
					components.Text("Monthly sales performance and revenue analytics"),
					components.Button("View Details", "view_sales"),
				),
				components.Card("Inventory Report").WithChildren(
					components.Text("Stock turnover, reorder thresholds, and warehouse levels"),
					components.Button("View Details", "view_inventory"),
				),
			),
		),
		"sales": components.Page("Sales Report").WithChildren(
			components.Grid(2).WithChildren(
				components.Card("Monthly Sales Trend").WithChildren(
					components.Chart("Sales Trend", "line", []map[string]interface{}{
						{"month": "Jan", "sales": 4000},
						{"month": "Feb", "sales": 3000},
						{"month": "Mar", "sales": 5000},
						{"month": "Apr", "sales": 4500},
						{"month": "May", "sales": 6200},
						{"month": "Jun", "sales": 7800},
					}).WithProp("xAxisKey", "month").WithProp("colors", []string{"#38bdf8"}),
				),
				components.Card("Revenue by Category").WithChildren(
					components.Chart("Category Breakdown", "bar", []map[string]interface{}{
						{"category": "Electronics", "share": 48},
						{"category": "Fashion", "share": 24},
						{"category": "Home & Living", "share": 16},
						{"category": "Sports", "share": 12},
					}).WithProp("xAxisKey", "category").WithProp("colors", []string{"#818cf8"}),
				),
			),
		),
	}
}
