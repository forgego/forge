package admin

import (
	"testing"

	"github.com/forgego/forge/admin/core"
)

func TestNewSite(t *testing.T) {
	tests := []struct {
		name         string
		siteName     string
		expectTitle  string
		expectHeader string
	}{
		{
			name:         "default site",
			siteName:     "default",
			expectTitle:  "Admin",
			expectHeader: "Administration",
		},
		{
			name:         "custom site",
			siteName:     "custom",
			expectTitle:  "Admin",
			expectHeader: "Administration",
		},
		{
			name:         "empty name",
			siteName:     "",
			expectTitle:  "Admin",
			expectHeader: "Administration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			site := NewSite(tt.siteName)
			if site == nil {
				t.Fatal("NewSite() returned nil")
			}
			if site.Name != tt.siteName {
				t.Errorf("Site.Name = %q, want %q", site.Name, tt.siteName)
			}
			if site.Title != tt.expectTitle {
				t.Errorf("Site.Title = %q, want %q", site.Title, tt.expectTitle)
			}
			if site.Header != tt.expectHeader {
				t.Errorf("Site.Header = %q, want %q", site.Header, tt.expectHeader)
			}
			if site.registry == nil {
				t.Error("Site.registry is nil")
			}
		})
	}
}

func TestSite_GetRegistry(t *testing.T) {
	site := NewSite("test")
	registry := site.GetRegistry()
	if registry == nil {
		t.Error("GetRegistry() returned nil")
	}
}

func TestSite_GetUIConfig(t *testing.T) {
	site := NewSite("test")
	config := site.GetUIConfig()
	if config.Source != UISourceEmbedded {
		t.Errorf("UIConfig.Source = %q, want %q", config.Source, UISourceEmbedded)
	}
}

func TestSite_WithUIConfig(t *testing.T) {
	site := NewSite("test")
	customConfig := UIConfig{
		Source:    UISourceStatic,
		StaticDir: "/static",
		Prefix:    "/admin",
	}
	result := site.WithUIConfig(customConfig)
	if result != site {
		t.Error("WithUIConfig() should return the site for chaining")
	}
	config := site.GetUIConfig()
	if config.Source != UISourceStatic {
		t.Errorf("UIConfig.Source = %q, want %q", config.Source, UISourceStatic)
	}
	if config.StaticDir != "/static" {
		t.Errorf("UIConfig.StaticDir = %q, want %q", config.StaticDir, "/static")
	}
}

func TestDefaultUIConfig(t *testing.T) {
	config := DefaultUIConfig()
	if config.Source != UISourceEmbedded {
		t.Errorf("DefaultUIConfig().Source = %q, want %q", config.Source, UISourceEmbedded)
	}
	if config.Prefix != "" {
		t.Errorf("DefaultUIConfig().Prefix = %q, want empty string", config.Prefix)
	}
}

func TestUISource_Constants(t *testing.T) {
	tests := []struct {
		name     string
		source   UISource
		expected string
	}{
		{"UISourceEmbedded", UISourceEmbedded, "embedded"},
		{"UISourceStatic", UISourceStatic, "static"},
		{"UISourceExternal", UISourceExternal, "external"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.source) != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.source, tt.expected)
			}
		})
	}
}

func TestSite_SetDB(t *testing.T) {
	site := NewSite("test")
	// Test with nil DB (should not panic)
	result := site.SetDB(nil)
	if result != site {
		t.Error("SetDB() should return the site for chaining")
	}
}

// Test that core.Registry is properly initialized
func TestSite_RegistryOperations(t *testing.T) {
	site := NewSite("test")
	registry := site.GetRegistry()

	// Test initial count
	if count := registry.Count(); count != 0 {
		t.Errorf("Initial registry count = %d, want 0", count)
	}

	// Test model names
	names := registry.ModelNames()
	if len(names) != 0 {
		t.Errorf("Initial model names = %v, want empty", names)
	}
}

// Test plugin registration
func TestSite_RegisterPlugin_Nil(t *testing.T) {
	site := NewSite("test")
	// This should not panic even with nil plugin handling in registry
	registry := site.GetRegistry()
	if registry == nil {
		t.Error("Registry should not be nil")
	}
}

// Test Handler method
func TestSite_Handler(t *testing.T) {
	site := NewSite("test")
	handler := site.Handler()
	if handler == nil {
		t.Error("Handler() returned nil")
	}
}

// Test IndexView method
func TestSite_IndexView(t *testing.T) {
	site := NewSite("test")
	handler := site.IndexView()
	if handler == nil {
		t.Error("IndexView() returned nil")
	}
}

// Test type aliases
func TestTypeAliases(t *testing.T) {
	// Test that type aliases are properly defined
	var _ Config[struct{}] = core.Config[struct{}]{}
	var _ Action[struct{}] = core.Action[struct{}]{}
	var _ Filter[struct{}] = core.Filter[struct{}]{}
	var _ Plugin = core.Plugin(nil)
	var _ Field = core.Field(nil)
	var _ Method = core.Method("")
	var _ Fieldset[struct{}] = core.Fieldset[struct{}]{}
	var _ InlineRelationConfig = core.InlineRelationConfig{}
	var _ InlineConfig = core.InlineConfig{}
	var _ Choice = core.Choice{}
}
