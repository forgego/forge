package api

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/parsers"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/renderers"
	"github.com/forgego/forge/api/throttling"
)

// Settings holds API framework configuration
type Settings struct {
	// Authentication
	DefaultAuthentication []authentication.Authentication

	// Permissions
	DefaultPermissions []permissions.Permission

	// Throttling
	DefaultThrottles []throttling.Throttle
	ThrottleRates    map[string]string

	// Renderers
	DefaultRenderers []renderers.Renderer

	// Parsers
	DefaultParsers []parsers.Parser

	// Pagination
	PageSize    int
	MaxPageSize int

	// Format
	DefaultFormat string

	// Timeouts
	RequestTimeout time.Duration

	// CORS
	CORSEnabled bool
	CORSOrigins []string
	CORSMethods []string
	CORSHeaders []string
}

// DefaultSettings returns default API settings
func DefaultSettings() *Settings {
	return &Settings{
		DefaultAuthentication: []authentication.Authentication{},
		DefaultPermissions:    []permissions.Permission{},
		DefaultThrottles:      []throttling.Throttle{},
		ThrottleRates:         make(map[string]string),
		PageSize:              20,
		MaxPageSize:           100,
		DefaultFormat:         "json",
		RequestTimeout:        30 * time.Second,
		CORSEnabled:           false,
		CORSOrigins:           []string{},
		CORSMethods:           []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		CORSHeaders:           []string{"Content-Type", "Authorization"},
	}
}

// Global settings instance
var (
	globalSettings  atomic.Pointer[Settings]
	settingsWriteMu sync.Mutex
)

func init() {
	globalSettings.Store(DefaultSettings())
}

// SetSettings sets the global API settings.
// Callers must not mutate the passed Settings pointer afterwards.
func SetSettings(settings *Settings) {
	settingsWriteMu.Lock()
	defer settingsWriteMu.Unlock()
	globalSettings.Store(settings)
}

// GetSettings returns the global API settings
func GetSettings() *Settings {
	return globalSettings.Load()
}

// Initialize initializes the API framework with default settings
func Initialize() {
	// Set default settings
	settings := DefaultSettings()

	// Set default renderers
	settings.DefaultRenderers = []renderers.Renderer{
		renderers.NewJSONRenderer(),
		renderers.NewHTMLRenderer(),
	}

	// Set default parsers
	settings.DefaultParsers = []parsers.Parser{
		parsers.NewJSONParser(),
		parsers.NewFormParser(),
	}

	SetSettings(settings)
}

// SetDefaultAuthentication sets the default authentication classes
func SetDefaultAuthentication(authClasses ...authentication.Authentication) {
	settingsWriteMu.Lock()
	defer settingsWriteMu.Unlock()

	cur := globalSettings.Load()
	var next Settings
	if cur != nil {
		next = *cur
	} else {
		next = *DefaultSettings()
	}
	next.DefaultAuthentication = append([]authentication.Authentication(nil), authClasses...)
	globalSettings.Store(&next)
}

// SetDefaultPermissions sets the default permission classes
func SetDefaultPermissions(permClasses ...permissions.Permission) {
	settingsWriteMu.Lock()
	defer settingsWriteMu.Unlock()

	cur := globalSettings.Load()
	var next Settings
	if cur != nil {
		next = *cur
	} else {
		next = *DefaultSettings()
	}
	next.DefaultPermissions = append([]permissions.Permission(nil), permClasses...)
	globalSettings.Store(&next)
}

// SetDefaultThrottles sets the default throttle classes
func SetDefaultThrottles(throttleClasses ...throttling.Throttle) {
	settingsWriteMu.Lock()
	defer settingsWriteMu.Unlock()

	cur := globalSettings.Load()
	var next Settings
	if cur != nil {
		next = *cur
	} else {
		next = *DefaultSettings()
	}
	next.DefaultThrottles = append([]throttling.Throttle(nil), throttleClasses...)
	globalSettings.Store(&next)
}
