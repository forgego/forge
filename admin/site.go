package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/forgego/forge/admin/api/rest"
	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/admin/ui"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/server"
	"github.com/forgego/forge/stores"
	"github.com/forgego/forge/stores/adminstore"
)

// Site represents an admin site instance
type Site struct {
	Name               string
	Title              string
	Header             string
	IndexTitle         string
	SiteURL            string
	db                 *db.DB
	registry           *core.Registry
	uiConfig           UIConfig
	loginAuthenticator rest.LoginAuthenticator
	// history and apiStores are set by UseDatabaseStores.
	history   core.HistoryManager
	apiStores rest.Stores
}

// SetLoginAuthenticator sets the custom login authenticator for the admin site.
func (s *Site) SetLoginAuthenticator(a rest.LoginAuthenticator) {
	s.loginAuthenticator = a
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

// UseDatabaseStores keeps the admin's bearer tokens, login lockout counts,
// saved views and change history in the database set with SetDB, instead of
// in process memory, so they survive a restart and are shared by every
// instance using that database. It is what server.stores: database selects
// in a project created by forge new. The framework store tables must exist
// (forge migrate up with server.stores: database, or stores.Migrate);
// otherwise it returns an error wrapping stores.ErrNotMigrated.
//
// Registered models whose config leaves HistoryManager unset, or uses
// admin.HistoryManager, write their history to the database; a model with
// its own HistoryManager keeps it. Call it before Handler.
func (s *Site) UseDatabaseStores(ctx context.Context) error {
	if s.db == nil {
		return errors.New("admin: UseDatabaseStores needs a database; call SetDB first")
	}
	if err := stores.CheckMigrated(ctx, s.db); err != nil {
		return err
	}
	shared, err := stores.New(s.db)
	if err != nil {
		return err
	}
	s.history = adminstore.NewHistory(shared)
	s.apiStores = rest.Stores{
		Tokens:        shared.AdminTokens(),
		SavedViews:    adminstore.NewSavedViews(shared),
		LoginAttempts: shared.LoginAttempts(adminLoginMaxFailures, adminLoginWindow),
	}
	for _, registered := range s.registry.GetAll() {
		useSiteHistory(registered, s.history)
	}
	return nil
}

// UseStores applies a server.stores value: "database" calls
// UseDatabaseStores, "memory" (or empty) keeps the in-memory stores.
func (s *Site) UseStores(ctx context.Context, kind string) error {
	parsed, err := stores.ParseKind(kind)
	if err != nil {
		return err
	}
	if parsed != stores.KindDatabase {
		return nil
	}
	return s.UseDatabaseStores(ctx)
}

// The admin login lockout policy, the same as the in-memory default: five
// failed logins per client IP or username lock it out for 15 minutes.
const (
	adminLoginMaxFailures = 5
	adminLoginWindow      = 15 * time.Minute
)

func useSiteHistory(registered core.AdminInterface, hm core.HistoryManager) {
	if hm == nil {
		return
	}
	if user, ok := registered.(interface{ UseSiteHistory(core.HistoryManager) }); ok {
		user.UseSiteHistory(hm)
	}
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
	if s.loginAuthenticator != nil {
		apiRouter.SetLoginAuthenticator(s.loginAuthenticator)
	}
	apiRouter.SetStores(s.apiStores)
	apiRouter.WithAdminPrefix(s.uiConfig.Prefix)
	apiRouter.RegisterRoutes(r)

	// 2. Serve Static UI Assets
	prefix := normalizeAdminPrefix(s.uiConfig.Prefix)

	// Determine the route pattern
	// Since we might be mounted, we should use a wildcard that matches everything passed to this handler
	// If prefix is set, StaticFS will handle stripping it from the path
	routePattern := "/*"

	transformOpt := server.WithIndexTransform(adminIndexTransform(prefix))

	if s.uiConfig.Source == UISourceStatic && s.uiConfig.StaticDir != "" {
		// Serve from local directory
		handler := server.StaticFiles("", s.uiConfig.StaticDir, server.WithPrefix(prefix), server.WithIndexFiles("index.html"), server.WithFallback("index.html"), server.WithDisableCache(true), transformOpt)
		r.Handle(routePattern, handler)
		if prefix == "" {
			r.Handle("/", handler)
		}
	} else if s.uiConfig.Source == UISourceEmbedded && s.uiConfig.EmbedFS != nil {
		// Serve from embedded FS
		handler := server.StaticFS("", s.uiConfig.EmbedFS, server.WithPrefix(prefix), server.WithIndexFiles("index.html"), server.WithFallback("index.html"), server.WithDisableCache(true), transformOpt)
		r.Handle(routePattern, handler)
		if prefix == "" {
			r.Handle("/", handler)
		}
	}

	return r
}

func normalizeAdminPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		return ""
	}
	return "/" + prefix
}

var headTagRegex = regexp.MustCompile(`(?i)<head(\s[^>]*)?>`)

func adminIndexTransform(prefix string) func([]byte) []byte {
	return func(content []byte) []byte {
		out := content

		// Rewrite build's default "/admin/" asset base to runtime prefix
		if prefix != "/admin" {
			out = bytes.ReplaceAll(out, []byte("\"/admin/"), []byte("\""+prefix+"/"))
			out = bytes.ReplaceAll(out, []byte("'/admin/"), []byte("'"+prefix+"/"))
		}

		// Inject runtime prefix meta tag immediately after <head> or prepend if absent
		metaTag := []byte(fmt.Sprintf(`<meta name="forge-admin-prefix" content="%s">`, html.EscapeString(prefix)))
		loc := headTagRegex.FindIndex(out)
		if loc != nil {
			res := make([]byte, 0, len(out)+len(metaTag))
			res = append(res, out[:loc[1]]...)
			res = append(res, metaTag...)
			res = append(res, out[loc[1]:]...)
			return res
		}

		res := make([]byte, 0, len(out)+len(metaTag))
		res = append(res, metaTag...)
		res = append(res, out...)
		return res
	}
}
