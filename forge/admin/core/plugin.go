package core

import (
	"context"

	"github.com/forgego/forge/admin/components"
)

// Plugin defines the interface for admin plugins
type Plugin interface {
	// ID returns the unique identifier for the plugin
	ID() string
	
	// Name returns the human-readable name of the plugin
	Name() string
	
	// Init initializes the plugin with the admin site
	Init(ctx context.Context, site interface{}) error

	// GetPages returns a map of route paths to SDUI components
	// e.g. "settings" -> Page("Settings")
	GetPages() map[string]components.Component
	
	// GetMenuItems returns menu items to add to the sidebar
	GetMenuItems() []MenuItem
}

// MenuItem represents a sidebar menu item
type MenuItem struct {
	Label    string `json:"label"`
	Path     string `json:"path"` // e.g. "/plugins/my-plugin/settings"
	Icon     string `json:"icon,omitempty"`
	Children []MenuItem `json:"children,omitempty"`
}

// BasePlugin provides a default implementation for Plugin
type BasePlugin struct{}

func (p *BasePlugin) ID() string { return "base" }
func (p *BasePlugin) Name() string { return "Base Plugin" }
func (p *BasePlugin) Init(ctx context.Context, site interface{}) error { return nil }
func (p *BasePlugin) GetPages() map[string]components.Component { return nil }
func (p *BasePlugin) GetMenuItems() []MenuItem { return nil }

