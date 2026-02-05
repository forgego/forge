package api

import (
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/parsers"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/renderers"
	"github.com/forgego/forge/api/throttling"
)

// GetDefaultAuthentication returns default authentication classes
func GetDefaultAuthentication() []authentication.Authentication {
	settings := GetSettings()
	return settings.DefaultAuthentication
}

// GetDefaultPermissions returns default permission classes
func GetDefaultPermissions() []permissions.Permission {
	settings := GetSettings()
	return settings.DefaultPermissions
}

// GetDefaultThrottles returns default throttle classes
func GetDefaultThrottles() []throttling.Throttle {
	settings := GetSettings()
	return settings.DefaultThrottles
}

// GetDefaultRenderers returns default renderers
func GetDefaultRenderers() []renderers.Renderer {
	settings := GetSettings()
	return settings.DefaultRenderers
}

// GetDefaultParsers returns default parsers
func GetDefaultParsers() []parsers.Parser {
	settings := GetSettings()
	return settings.DefaultParsers
}

// SetupDefaultAPI sets up the API with sensible defaults
func SetupDefaultAPI() {
	Initialize()

	// Set default renderers
	SetDefaultRenderers(
		renderers.NewJSONRenderer(),
		renderers.NewHTMLRenderer(),
	)

	// Set default parsers
	SetDefaultParsers(
		parsers.NewJSONParser(),
		parsers.NewFormParser(),
	)
}

// SetDefaultRenderers sets default renderers
func SetDefaultRenderers(rendererList ...renderers.Renderer) {
	settings := GetSettings()
	settings.DefaultRenderers = rendererList
}

// SetDefaultParsers sets default parsers
func SetDefaultParsers(parserList ...parsers.Parser) {
	settings := GetSettings()
	settings.DefaultParsers = parserList
}

// CreateDefaultViewSet creates a viewset with default settings
func CreateDefaultViewSet(serializer func() Serializer, queryset, model interface{}) *EnhancedBaseViewSetIntegrated {
	vs := NewEnhancedBaseViewSetIntegrated(serializer, queryset, model)

	// Apply defaults from settings
	settings := GetSettings()

	if len(settings.DefaultAuthentication) > 0 {
		vs.AuthenticationClasses = settings.DefaultAuthentication
	}

	if len(settings.DefaultPermissions) > 0 {
		vs.PermissionClasses = settings.DefaultPermissions
	}

	if len(settings.DefaultThrottles) > 0 {
		vs.ThrottleClasses = settings.DefaultThrottles
	}

	return vs
}
