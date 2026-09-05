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
					components.Text("Monthly sales performance"),
					components.Button("View Details", "view_sales"),
				),
				components.Card("Inventory Report").WithChildren(
					components.Text("Stock levels and turnover"),
					components.Button("View Details", "view_inventory"),
				),
			),
		),
		"sales": components.Page("Sales Report").WithChildren(
			components.Card("Monthly Sales").WithChildren(
				components.Chart("Sales Trend", "line", []map[string]interface{}{
					{"month": "Jan", "sales": 4000},
					{"month": "Feb", "sales": 3000},
					{"month": "Mar", "sales": 5000},
					{"month": "Apr", "sales": 4500},
				}).WithProp("xAxisKey", "month"),
			),
		),
	}
}
