package api

import (
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/api/parsers"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/renderers"
	"github.com/forgego/forge/api/throttling"
	forgehttp "github.com/forgego/forge/server"
)

// SetupCompleteAPI sets up a complete API with all features
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

// CreateProductionViewSet creates a production-ready viewset with all features
func CreateProductionViewSet(
	serializer func() Serializer,
	queryset, model interface{},
) *EnhancedBaseViewSetIntegrated {
	vs := NewEnhancedBaseViewSetIntegrated(serializer, queryset, model)

	// Apply global defaults
	settings := GetSettings()

	// Authentication
	if len(settings.DefaultAuthentication) > 0 {
		vs.AuthenticationClasses = settings.DefaultAuthentication
	}

	// Permissions
	if len(settings.DefaultPermissions) > 0 {
		vs.PermissionClasses = settings.DefaultPermissions
	}

	// Throttling
	if len(settings.DefaultThrottles) > 0 {
		vs.ThrottleClasses = settings.DefaultThrottles
	}

	// Renderers
	if len(settings.DefaultRenderers) > 0 {
		vs.RendererClasses = settings.DefaultRenderers
	}

	// Parsers
	if len(settings.DefaultParsers) > 0 {
		vs.ParserClasses = settings.DefaultParsers
	}

	return vs
}

// RegisterAPIWithDefaults registers an API with default middleware
func RegisterAPIWithDefaults(
	router *forgehttp.Router,
	prefix string,
	registerFunc func(*EnhancedRouter),
) {
	// Create enhanced router
	apiRouter := NewEnhancedRouter(prefix)

	// Apply API middleware
	apiMiddleware := core.Chain(
		APIMiddleware(),
	)

	// Register routes
	router.Route(prefix, func(sub *forgehttp.Router) {
		sub.Use(apiMiddleware)
		registerFunc(apiRouter)
		apiRouter.RegisterRoutesEnhanced(sub)
	})
}

// CompleteExample shows a complete API setup
func CompleteExample() {
	// Setup API
	SetupCompleteAPI()

	// Configure authentication
	SetDefaultAuthentication(
		authentication.NewTokenAuthentication(func(token string) (interface{}, error) {
			// Your token lookup
			return nil, nil
		}),
	)

	// Configure permissions
	SetDefaultPermissions(
		permissions.NewIsAuthenticated(),
	)

	// Configure throttling
	cache := throttling.NewMemoryCache()
	SetDefaultThrottles(
		throttling.NewAnonRateThrottle("100/hour", cache),
		throttling.NewUserRateThrottle("1000/day", cache),
	)

	// Create router
	router := forgehttp.NewRouter()

	// Register API
	RegisterAPIWithDefaults(router, "/api/v1", func(apiRouter *EnhancedRouter) {
		// Register viewsets
		// apiRouter.RegisterEnhanced("users", userViewSet)
	})
}
