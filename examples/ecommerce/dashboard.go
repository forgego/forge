package main

import (
	"github.com/forgego/forge/admin/components"
	"github.com/forgego/forge/admin/core"
)

func SetupDashboard() {
	core.SetDashboard(core.DashboardConfig{
		Title: "Ecommerce Dashboard",
		Layout: components.Grid(4).WithChildren(
			// Stats Row
			components.Stats("Total Revenue", "$128,430", "+12.5%").
				WithColSpan(1),
			components.Stats("Active Customers", "2,450", "+8.2%").
				WithColSpan(1),
			components.Stats("Pending Orders", "43", "-2.4%").
				WithColSpan(1),
			components.Stats("Inventory Value", "$840k", "+4.1%").
				WithColSpan(1),

			// Main Content
			components.Card("Revenue Trends (Last 7 Days)").
				WithColSpan(3).
				WithChildren(
					components.Chart("Revenue", "area", []map[string]interface{}{
						{"name": "Mon", "revenue": 4000, "orders": 240},
						{"name": "Tue", "revenue": 3000, "orders": 139},
						{"name": "Wed", "revenue": 2000, "orders": 980},
						{"name": "Thu", "revenue": 2780, "orders": 390},
						{"name": "Fri", "revenue": 1890, "orders": 480},
						{"name": "Sat", "revenue": 2390, "orders": 380},
						{"name": "Sun", "revenue": 3490, "orders": 430},
					}).WithProp("dataKeys", []string{"revenue", "orders"}).
					   WithProp("colors", []string{"#0ea5e9", "#22c55e"}).
					   WithProp("height", 350),
				),
			
			// Sidebar
			components.Card("Recent Activity").
				WithColSpan(1).
				WithChildren(
					components.Text("Order #1234"),
					components.Text("New User: John Doe"),
				),
		),
	})
}
