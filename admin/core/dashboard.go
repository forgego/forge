package core

import (
	"context"
	"sync"

	"github.com/forgego/forge/admin/components"
)

// DashboardConfig defines the configuration for the admin dashboard
type DashboardConfig struct {
	Title  string
	Layout components.Component
}

// DashboardRegistry manages the dashboard configuration
type DashboardRegistry struct {
	config *DashboardConfig
	mu     sync.RWMutex
}

var globalDashboardRegistry = &DashboardRegistry{}

// SetDashboard sets the global dashboard configuration
func SetDashboard(config DashboardConfig) {
	globalDashboardRegistry.mu.Lock()
	defer globalDashboardRegistry.mu.Unlock()
	globalDashboardRegistry.config = &config
}

// GetDashboard returns the current dashboard configuration
func GetDashboard(ctx context.Context) *DashboardConfig {
	globalDashboardRegistry.mu.RLock()
	defer globalDashboardRegistry.mu.RUnlock()
	return globalDashboardRegistry.config
}

func init() {
	SetDashboard(DefaultDashboard())
}

// DefaultDashboard creates a default dashboard layout
func DefaultDashboard() DashboardConfig {
	return DashboardConfig{
		Title: "Dashboard",
		Layout: components.Grid(12).WithChildren(
			components.Card("Welcome").
				WithColSpan(12).
				WithChildren(
					components.Text("Welcome to Forge Admin"),
				),
		),
	}
}
