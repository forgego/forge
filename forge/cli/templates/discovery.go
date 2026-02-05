package templates

// DiscoveryRule defines how to discover components in a project
type DiscoveryRule struct {
	Pattern      string
	Purpose      string
	Registration string
}

// GetDiscoveryRules returns the auto-discovery rules for the framework
func GetDiscoveryRules() []DiscoveryRule {
	return []DiscoveryRule{
		{
			Pattern:      "app/*/models.go",
			Purpose:      "ORM models",
			Registration: "Via schema registry",
		},
		{
			Pattern:      "app/*/admin.go",
			Purpose:      "Admin config (per-app)",
			Registration: "Via init() functions",
		},
		{
			Pattern:      "app/*/api.go",
			Purpose:      "API endpoints",
			Registration: "Via init() functions",
		},
		{
			Pattern:      "domain/*/service.go",
			Purpose:      "Business logic",
			Registration: "Manual wiring (optional)",
		},
		{
			Pattern:      "infra/*/client.go",
			Purpose:      "Infrastructure",
			Registration: "Manual wiring",
		},
	}
}
