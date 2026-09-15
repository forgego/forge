package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/forgego/forge/admin/core"
	apicore "github.com/forgego/forge/api/core"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Router handles all admin API routes
type Router struct {
	registry *core.Registry
	prefix   string
	// adminPrefix is the public mount path of the admin UI, used for
	// absolute URLs in search results. Defaults to "/admin".
	adminPrefix   string
	views         *savedViewStore
	sessions      *adminSessionStore
	loginLimiter  *loginLimiter
	authenticator LoginAuthenticator
}

// NewRouter creates a new admin API router
func NewRouter(registry *core.Registry) *Router {
	return &Router{
		registry: registry,
		prefix:   "/api",
		// adminPrefix is the public path the admin UI is served from. It is
		// used to build absolute URLs (e.g. in global search results).
		adminPrefix:  "/admin",
		views:        newSavedViewStore(),
		sessions:     newAdminSessionStore(),
		loginLimiter: newLoginLimiter(),
	}
}

// WithAdminPrefix sets the public path prefix of the admin UI (default
// "/admin") so generated URLs stay correct on custom mount points.
func (r *Router) WithAdminPrefix(prefix string) *Router {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "/admin"
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	r.adminPrefix = strings.TrimRight(prefix, "/")
	if r.adminPrefix == "" {
		r.adminPrefix = "/"
	}
	return r
}

// RegisterRoutes registers all admin routes
func (r *Router) RegisterRoutes(router chi.Router) {
	router.Route(r.prefix, func(sub chi.Router) {
		// Middleware
		sub.Use(middleware.Logger)
		sub.Use(middleware.Recoverer)
		sub.Use(cors.Handler(cors.Options{
			// Echo back the request origin instead of "*": the Fetch spec
			// forbids wildcard origins on credentialed requests, so browsers
			// would otherwise drop cookies/Authorization on cross-origin
			// admin deployments. Restrict via reverse proxy in production.
			AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		}))

		// Auth endpoints
		sub.Post("/login", r.handleLogin)

		sub.Group(func(protected chi.Router) {
			protected.Use(r.authMiddleware)

			// Auth endpoints
			protected.Post("/logout", r.handleLogout)

			// Configuration endpoint
			protected.Get("/config", r.handleConfig)

			// Metadata endpoints
			protected.Get("/meta", r.handleMetaList)
			protected.Get("/meta/{model}", r.handleMetaDetail)

			// Global search
			protected.Get("/search", r.handleGlobalSearch)

			// Plugin page endpoint
			protected.Get("/plugins/{plugin}/pages/{page}", r.handlePluginPage)

			// Saved views
			protected.Route("/saved-views/{model}", func(viewRouter chi.Router) {
				viewRouter.Get("/", r.handleSavedViewsList)
				viewRouter.Post("/", r.handleSavedViewSave)
				viewRouter.Delete("/{id}", r.handleSavedViewDelete)
			})

			// Model routes (registered dynamically)
			r.registerModelRoutes(protected)
		})
	})
}

func (r *Router) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token, err := bearerToken(req.Header.Get("Authorization"))
		if err != nil {
			respondError(w, http.StatusUnauthorized, "authentication_required", "Authentication required", nil)
			return
		}

		session, ok := r.sessions.Validate(token)
		if !ok {
			respondError(w, http.StatusUnauthorized, "authentication_required", "Invalid or expired token", nil)
			return
		}

		user := map[string]interface{}{
			"username": session.Username,
			"role":     "superuser",
		}
		ctx := apicore.WithUser(req.Context(), user)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// registerModelRoutes registers routes for each model
func (r *Router) registerModelRoutes(router chi.Router) {
	for name, admin := range r.registry.GetAll() {
		basePaths := []string{fmt.Sprintf("/%s", name)}
		if admin.ModelType() != nil {
			structName := admin.ModelType().Name()
			if structName != "" && structName != name {
				basePaths = append(basePaths, fmt.Sprintf("/%s", structName))
			}
		}

		for _, basePath := range basePaths {
			router.Route(basePath, func(sub chi.Router) {
				// List and create
				sub.Get("/", r.handleList(admin))
				sub.Get("/export", r.handleExport(admin))
				sub.Post("/", r.handleCreate(admin))

				// Detail, update, delete
				sub.Route("/{id}", func(subDetail chi.Router) {
					subDetail.Get("/", r.handleDetail(admin))
					subDetail.Get("/history", r.handleHistory(admin))
					subDetail.Patch("/", r.handleUpdate(admin))
					subDetail.Put("/", r.handleReplace(admin))
					subDetail.Delete("/", r.handleDelete(admin))
				})

				// Bulk operations
				sub.Post("/bulk-create", r.handleBulkCreate(admin))
				sub.Post("/bulk-update", r.handleBulkUpdate(admin))
				sub.Delete("/bulk-delete", r.handleBulkDelete(admin))

				// Actions
				sub.Post("/action/{action}", r.handleAction(admin))

				// Autocomplete
				sub.Get("/autocomplete", r.handleAutocomplete(admin))
			})
		}
	}
}

// handleAction executes a bulk action
func (r *Router) handleAction(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)
		actionName := chi.URLParam(req, "action")

		// Check change permission (actions typically require change permission)
		if !admin.HasChangePermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to execute actions", nil)
			return
		}

		// Parse request
		var request core.BulkActionRequest
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_body", err.Error(), nil)
			return
		}
		if len(request.IDs) == 0 {
			respondError(w, http.StatusBadRequest, "invalid_body", "Request body must include at least one ID", nil)
			return
		}

		// Call implementation
		response, err := admin.ExecuteAction(ctx, actionName, request.IDs, request.Params)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "action_failed", err.Error(), nil)
			return
		}

		status := http.StatusOK
		if bulkResponse, ok := response.(*core.BulkActionResponse); ok {
			status = actionResponseStatusCode(bulkResponse)
		}

		respondJSON(w, status, response)
	}
}

func actionResponseStatusCode(response *core.BulkActionResponse) int {
	if response == nil || len(response.Errors) == 0 {
		return http.StatusOK
	}
	if response.Affected > 0 {
		return http.StatusMultiStatus
	}
	return http.StatusBadRequest
}

// handleGlobalSearch handles global search across all models
func (r *Router) handleGlobalSearch(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	user, _ := apicore.UserFromContext(ctx)
	query := req.URL.Query().Get("q")

	if query == "" {
		respondError(w, http.StatusBadRequest, "missing_query", "Search query is required", nil)
		return
	}

	// Optional `models` param (comma-separated) restricts the search scope.
	allowedModels := make(map[string]bool)
	if modelsParam := strings.TrimSpace(req.URL.Query().Get("models")); modelsParam != "" {
		for _, name := range strings.Split(modelsParam, ",") {
			if trimmed := strings.TrimSpace(name); trimmed != "" {
				allowedModels[trimmed] = true
			}
		}
	}

	allAdmins := r.registry.GetAll()
	results := make([]core.SearchResultGroup, 0)

	adminBase := r.adminPrefix
	if adminBase == "" {
		adminBase = "/admin"
	}

	for name, admin := range allAdmins {
		if len(allowedModels) > 0 && !allowedModels[name] {
			continue
		}
		if !admin.HasViewPermission(ctx, user, nil) {
			continue
		}
		items, err := admin.Autocomplete(ctx, query, 5)
		if err != nil || len(items) == 0 {
			continue
		}

		group := core.SearchResultGroup{
			Model: name,
			Count: len(items),
			Items: make([]core.SearchResultItem, 0, len(items)),
		}

		for _, item := range items {
			group.Items = append(group.Items, core.SearchResultItem{
				ID:    item.Value,
				Title: item.Label,
				URL:   fmt.Sprintf("%s/%s/%v/view", strings.TrimRight(adminBase, "/"), name, item.Value),
			})
		}

		results = append(results, group)
	}

	respondJSON(w, http.StatusOK, core.SearchResponse{
		Results: results,
	})
}

// handlePluginPage returns the SDUI layout for a plugin page
func (r *Router) handlePluginPage(w http.ResponseWriter, req *http.Request) {
	pluginID := chi.URLParam(req, "plugin")
	pageID := chi.URLParam(req, "page")

	plugin, err := r.registry.GetPlugin(pluginID)
	if err != nil {
		respondError(w, http.StatusNotFound, "plugin_not_found", "Plugin not found", nil)
		return
	}

	pages := plugin.GetPages()
	if pages == nil {
		respondError(w, http.StatusNotFound, "page_not_found", "Page not found", nil)
		return
	}

	page, ok := pages[pageID]
	if !ok {
		respondError(w, http.StatusNotFound, "page_not_found", "Page not found", nil)
		return
	}

	respondJSON(w, http.StatusOK, page)
}
