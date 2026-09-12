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
			components.Stats("Total Revenue", "$142,850", "+14.2%").
				WithColSpan(1),
			components.Stats("Total Orders", "854", "+9.8%").
				WithColSpan(1),
			components.Stats("Avg Order Value", "$167.27", "+3.4%").
				WithColSpan(1),
			components.Stats("Low Stock Alerts", "3 items", "Action required").
				WithColSpan(1),

			// Main Revenue Chart
			components.Card("Sales & Revenue Trends (Last 7 Days)").
				WithColSpan(3).
				WithChildren(
					components.Chart("Revenue", "area", []map[string]interface{}{
						{"name": "Mon", "revenue": 4200, "orders": 240},
						{"name": "Tue", "revenue": 3800, "orders": 195},
						{"name": "Wed", "revenue": 5100, "orders": 310},
						{"name": "Thu", "revenue": 4780, "orders": 290},
						{"name": "Fri", "revenue": 6290, "orders": 420},
						{"name": "Sat", "revenue": 7890, "orders": 510},
						{"name": "Sun", "revenue": 6940, "orders": 460},
					}).WithProp("dataKeys", []string{"revenue", "orders"}).
						WithProp("colors", []string{"#0ea5e9", "#22c55e"}).
						WithProp("height", 350),
				),

			// Orders by Status Breakdown
			components.Card("Order Status Breakdown").
				WithColSpan(1).
				WithChildren(
					components.Chart("Status", "bar", []map[string]interface{}{
						{"status": "Delivered", "count": 520},
						{"status": "Processing", "count": 180},
						{"status": "Shipped", "count": 110},
						{"status": "Pending", "count": 44},
					}).WithProp("dataKeys", []string{"count"}).
						WithProp("colors", []string{"#6366f1"}).
						WithProp("height", 350),
				),

			// Top Selling Products
			components.Card("Top Selling Products").
				WithColSpan(2).
				WithChildren(
					components.Text("1. UltraBook Pro 16 - $1,899.00 (142 sold)"),
					components.Text("2. Studio Wireless Headphones - $349.00 (215 sold)"),
					components.Text("3. Pro Mechanical Keyboard - $179.00 (310 sold)"),
					components.Text("4. Merino Wool Tech Hoodie - $185.00 (188 sold)"),
					components.Text("5. Matter Smart Sensor Hub - $89.00 (420 sold)"),
				),

			// Real-time Platform Health & Inventory Alerts
			components.Card("System Health & Inventory Watch").
				WithColSpan(2).
				WithChildren(
					components.Text("✅ All Warehouses Online (San Francisco, New Jersey, Frankfurt)"),
					components.Text("⚠️ Low Stock Alert: Apex Pro Running Shoes (4 units remaining)"),
					components.Text("ℹ️ Security Active: SHA-256 Cookie Sessions & Bcrypt Double-Hash"),
					components.Text("⚡ API Response Time: P95 < 25ms across all ViewSets"),
				),
		),
	})
}
