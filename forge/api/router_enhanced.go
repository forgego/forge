package api

import (
	"net/http"

	"github.com/forgego/forge/api/docs"
	forgehttp "github.com/forgego/forge/server"
)

// ActionConfig represents configuration for a custom action
type ActionConfig struct {
	// Methods are the HTTP methods allowed for this action
	Methods []string
	// Detail indicates if this is a detail action (requires ID) or list action
	Detail bool
	// URLPath is the custom URL path (default: action name)
	URLPath string
	// URLName is the URL name for reversing
	URLName string
}

// Action represents a custom viewset action
type Action struct {
	Name    string
	Config  *ActionConfig
	Handler http.HandlerFunc
}

// ActionRegistry is a registry for custom actions
type ActionRegistry struct {
	actions map[string]*Action
}

// NewActionRegistry creates a new action registry
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[string]*Action),
	}
}

// Register registers a custom action
func (ar *ActionRegistry) Register(name string, config *ActionConfig, handler http.HandlerFunc) {
	ar.actions[name] = &Action{
		Name:    name,
		Config:  config,
		Handler: handler,
	}
}

// GetAction gets an action by name
func (ar *ActionRegistry) GetAction(name string) (*Action, bool) {
	action, ok := ar.actions[name]
	return action, ok
}

// GetAllActions returns all registered actions
func (ar *ActionRegistry) GetAllActions() map[string]*Action {
	return ar.actions
}

// EnhancedRouter provides enhanced routing with action support
type EnhancedRouter struct {
	*Router
	actionRegistry *ActionRegistry
	metadata       docs.Metadata
}

// NewEnhancedRouter creates a new enhanced router
func NewEnhancedRouter(prefix string) *EnhancedRouter {
	return &EnhancedRouter{
		Router:         NewRouter(prefix),
		actionRegistry: NewActionRegistry(),
		metadata:       docs.NewSimpleMetadata(),
	}
}

// RegisterAction registers a custom action
func (r *EnhancedRouter) RegisterAction(name string, config *ActionConfig, handler http.HandlerFunc) {
	r.actionRegistry.Register(name, config, handler)
}

// RegisterEnhanced registers an enhanced viewset with full feature support
func (r *EnhancedRouter) RegisterEnhanced(resource string, vs *EnhancedBaseViewSetIntegrated) {
	// Register standard CRUD routes
	r.Register(resource, vs)

	// Register custom actions if any
	// This would be enhanced to automatically discover actions
}

// RegisterRoutesEnhanced registers all routes including custom actions
func (r *EnhancedRouter) RegisterRoutesEnhanced(router *forgehttp.Router) {
	// Register standard routes
	r.RegisterRoutes(router)

	// Register custom actions
	allActions := r.actionRegistry.GetAllActions()
	for name, action := range allActions {
		path := r.prefix + "/" + name
		if action.Config.Detail {
			path = r.prefix + "/{id}/" + name
		}

		router.Route(path, func(sub *forgehttp.Router) {
			for _, method := range action.Config.Methods {
				switch method {
				case "GET":
					sub.Get("/", action.Handler)
				case "POST":
					sub.Post("/", action.Handler)
				case "PUT":
					sub.Put("/", action.Handler)
				case "PATCH":
					sub.Patch("/", action.Handler)
				case "DELETE":
					sub.Delete("/", action.Handler)
				}
			}
		})
	}
}

// OptionsHandler handles OPTIONS requests for metadata
func (r *EnhancedRouter) OptionsHandler(w http.ResponseWriter, req *http.Request, view interface{}) {
	docs.OptionsHandler(w, req, view, r.metadata)
}
