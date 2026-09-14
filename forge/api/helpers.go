package api

import (
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/parsers"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/renderers"
	"github.com/forgego/forge/api/throttling"
)

// GetDefaultAuthentication returns default authentication classes.
func GetDefaultAuthentication() []authentication.Authentication {
	settings := GetSettings()
	if settings == nil {
		return nil
	}
	return settings.DefaultAuthentication
}

// GetDefaultPermissions returns default permission classes.
func GetDefaultPermissions() []permissions.Permission {
	settings := GetSettings()
	if settings == nil {
		return nil
	}
	return settings.DefaultPermissions
}

// GetDefaultThrottles returns default throttle classes.
func GetDefaultThrottles() []throttling.Throttle {
	settings := GetSettings()
	if settings == nil {
		return nil
	}
	return settings.DefaultThrottles
}

// GetDefaultRenderers returns default renderers.
func GetDefaultRenderers() []renderers.Renderer {
	settings := GetSettings()
	if settings == nil {
		return nil
	}
	return settings.DefaultRenderers
}

// GetDefaultParsers returns default parsers.
func GetDefaultParsers() []parsers.Parser {
	settings := GetSettings()
	if settings == nil {
		return nil
	}
	return settings.DefaultParsers
}

// SetupDefaultAPI sets up the API with sensible defaults.
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

// SetupCompleteAPI sets up a complete API with all features.
func SetupCompleteAPI() {
	// Initialize with defaults
	Initialize()

	// Set up default renderers
	SetDefaultRenderers(
		renderers.NewJSONRenderer(),
		renderers.NewXMLRenderer(),
		renderers.NewHTMLRenderer(),
	)

	// Set up default parsers
	SetDefaultParsers(
		parsers.NewJSONParser(),
		parsers.NewFormParser(),
		parsers.NewMultiPartParser(),
	)
}

// SetDefaultRenderers sets default renderers.
func SetDefaultRenderers(rendererList ...renderers.Renderer) {
	settingsWriteMu.Lock()
	defer settingsWriteMu.Unlock()

	cur := globalSettings.Load()
	var next Settings
	if cur != nil {
		next = *cur
	} else {
		next = *DefaultSettings()
	}
	next.DefaultRenderers = append([]renderers.Renderer(nil), rendererList...)
	globalSettings.Store(&next)
}

// SetDefaultParsers sets default parsers.
func SetDefaultParsers(parserList ...parsers.Parser) {
	settingsWriteMu.Lock()
	defer settingsWriteMu.Unlock()

	cur := globalSettings.Load()
	var next Settings
	if cur != nil {
		next = *cur
	} else {
		next = *DefaultSettings()
	}
	next.DefaultParsers = append([]parsers.Parser(nil), parserList...)
	globalSettings.Store(&next)
}
