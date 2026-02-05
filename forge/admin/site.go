package admin

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/forgego/forge/admin/api/rest"
	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/admin/ui"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/server"
)

// Site represents an admin site instance
type Site struct {
	Name       string
	Title      string
	Header     string
	IndexTitle string
	SiteURL    string
	db         *db.DB
	registry   *core.Registry
	uiConfig   UIConfig
}

// UISource defines where the Admin UI assets come from
type UISource string

const (
	UISourceEmbedded UISource = "embedded" // Use UI files embedded in the forge binary
	UISourceStatic   UISource = "static"   // Use UI files from a local directory
	UISourceExternal UISource = "external" // Use UI from an external URL (e.g. dev server)
)

// UIConfig configures the Admin UI
type UIConfig struct {
	Source    UISource
	StaticDir string
	EmbedFS   fs.FS
	Prefix    string // URL prefix for the admin site, e.g. "/admin"
}

func DefaultUIConfig() UIConfig {
	return UIConfig{
		Source:  UISourceEmbedded,
		Prefix:  "",
		EmbedFS: ui.GetFS(),
	}
}

// NewSite creates a new admin site
func NewSite(name string) *Site {
	return &Site{
		Name:       name,
		Title:      "Admin",
		Header:     "Administration",
		IndexTitle: "Site Administration",
		registry:   core.NewRegistry(),
		uiConfig:   DefaultUIConfig(),
	}
}

// WithUIConfig sets the UI configuration for the site
func (s *Site) WithUIConfig(config UIConfig) *Site {
	s.uiConfig = config
	return s
}

// SetDB sets the database connection for the site and all registered admins
func (s *Site) SetDB(database *db.DB) *Site {
	s.db = database
	// Propagate to existing admins
	for _, admin := range s.registry.GetAll() {
		admin.SetDB(database)
	}
	return s
}

// GetUIConfig returns the current UI configuration
func (s *Site) GetUIConfig() UIConfig {
	return s.uiConfig
}

// GetRegistry returns the registry for this site
func (s *Site) GetRegistry() *core.Registry {
	return s.registry
}

// RegisterPlugin registers a plugin with the site
func (s *Site) RegisterPlugin(ctx context.Context, p core.Plugin) error {
	if err := s.registry.RegisterPlugin(p); err != nil {
		return err
	}
	return p.Init(ctx, s)
}

// IndexView handles the admin index/dashboard
func (s *Site) IndexView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allAdmins := s.registry.GetAll()

		models := make([]map[string]interface{}, 0, len(allAdmins))
		for name, admin := range allAdmins {
			models = append(models, map[string]interface{}{
				"name":        name,
				"verboseName": name, // Should ideally come from metadata
				"modelType":   admin.ModelType().String(),
			})
		}

		allPlugins := s.registry.GetAllPlugins()
		plugins := make([]map[string]interface{}, 0, len(allPlugins))
		for _, p := range allPlugins {
			plugins = append(plugins, map[string]interface{}{
				"id":          p.ID(),
				"name":        p.Name(),
				"menuEntries": p.GetMenuItems(),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"models":  models,
			"plugins": plugins,
		})
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Handler returns the HTTP handler for the admin site
func (s *Site) Handler() http.Handler {
	r := chi.NewRouter()

	// 1. Register API Routes
	apiRouter := rest.NewRouter(s.registry)
	apiRouter.RegisterRoutes(r)

	// 2. Serve Static UI Assets
	prefix := s.uiConfig.Prefix

	// Determine the route pattern
	// Since we might be mounted, we should use a wildcard that matches everything passed to this handler
	// If prefix is set, StaticFS will handle stripping it from the path
	routePattern := "/*"

	if s.uiConfig.Source == UISourceStatic && s.uiConfig.StaticDir != "" {
		// Serve from local directory
		handler := server.StaticFiles("", s.uiConfig.StaticDir, server.WithPrefix(prefix), server.WithIndexFiles("index.html"), server.WithFallback("index.html"))
		r.Handle(routePattern, handler)
		if prefix == "" {
			r.Handle("/", handler)
		}
	} else if s.uiConfig.Source == UISourceEmbedded && s.uiConfig.EmbedFS != nil {
		// Serve from embedded FS
		handler := server.StaticFS("", s.uiConfig.EmbedFS, server.WithPrefix(prefix), server.WithIndexFiles("index.html"), server.WithFallback("index.html"), server.WithDisableCache(true))
		r.Handle(routePattern, handler)
		if prefix == "" {
			r.Handle("/", handler)
		}
	}

	return r
}
