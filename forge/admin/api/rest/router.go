package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

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
	views    *savedViewStore
}

// NewRouter creates a new admin API router
func NewRouter(registry *core.Registry) *Router {
	return &Router{
		registry: registry,
		prefix:   "/api",
		views:    newSavedViewStore(),
	}
}

// RegisterRoutes registers all admin routes
func (r *Router) RegisterRoutes(router chi.Router) {
	router.Route(r.prefix, func(sub chi.Router) {
		// Middleware
		sub.Use(middleware.Logger)
		sub.Use(middleware.Recoverer)
		sub.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		}))

		// Configuration endpoint
		sub.Get("/config", r.handleConfig)

		// Metadata endpoints
		sub.Get("/meta", r.handleMetaList)
		sub.Get("/meta/{model}", r.handleMetaDetail)

		// Auth endpoints
		sub.Post("/login", r.handleLogin)

		// Global search
		sub.Get("/search", r.handleGlobalSearch)

		// Plugin page endpoint
		sub.Get("/plugins/{plugin}/pages/{page}", r.handlePluginPage)

		// Saved views
		sub.Route("/saved-views/{model}", func(viewRouter chi.Router) {
			viewRouter.Get("/", r.handleSavedViewsList)
			viewRouter.Post("/", r.handleSavedViewSave)
		})

		// Model routes (registered dynamically)
		r.registerModelRoutes(sub)
	})
}

// registerModelRoutes registers routes for each model
func (r *Router) registerModelRoutes(router chi.Router) {
	for name, admin := range r.registry.GetAll() {
		basePath := fmt.Sprintf("/%s", name)

		router.Route(basePath, func(sub chi.Router) {
			// List and create
			sub.Get("/", r.handleList(admin))
			sub.Post("/", r.handleCreate(admin))

			// Detail, update, delete
			sub.Route("/{id}", func(subDetail chi.Router) {
				subDetail.Get("/", r.handleDetail(admin))
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

// handleConfig returns the admin configuration
func (r *Router) handleConfig(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	user, _ := apicore.UserFromContext(ctx)

	// Gather plugin metadata
	plugins := make([]map[string]interface{}, 0)
	for _, p := range r.registry.GetAllPlugins() {
		plugins = append(plugins, map[string]interface{}{
			"id":          p.ID(),
			"name":        p.Name(),
			"menuEntries": p.GetMenuItems(),
		})
	}

	config := map[string]interface{}{
		"title":       "Forge Admin",
		"version":     "1.0.0",
		"user":        user,
		"plugins":     plugins,
		"environment": "development", // TODO: Get from global config
		"dashboard":   core.GetDashboard(ctx),
	}

	respondJSON(w, http.StatusOK, config)
}

// handleMetaList returns list of all models
func (r *Router) handleMetaList(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	user, _ := apicore.UserFromContext(ctx)

	allAdmins := r.registry.GetAll()
	fmt.Printf("DEBUG: handleMetaList. Registered models: %d\n", len(allAdmins))
	for k := range allAdmins {
		fmt.Printf("DEBUG: Registered model: %s\n", k)
	}
	models := make([]core.ModelListMetadata, 0, len(allAdmins))

	for name, admin := range allAdmins {
		// Check module permission
		if !admin.HasModulePermission(ctx, user) {
			continue
		}

		meta, err := admin.GetMetadata(ctx, user)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		// TODO: Get count from manager
		count := int64(0)

		models = append(models, core.ModelListMetadata{
			Name:              name,
			VerboseName:       meta.VerboseName,
			VerboseNamePlural: meta.VerboseNamePlural,
			Icon:              meta.Icon,
			Count:             count,
			Permissions:       meta.Permissions,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"models": models,
	})
}

// handleLogin handles admin login
func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	// Simple stub for now - just return success
	// Production auth should verify credentials
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token": "dummy-token",
		"user":  map[string]string{"name": "Admin", "role": "superuser"},
	})
}

// handleMetaDetail returns detailed metadata for a model
func (r *Router) handleMetaDetail(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	modelName := chi.URLParam(req, "model")
	user, _ := apicore.UserFromContext(ctx)

	fmt.Printf("DEBUG: handleMetaDetail for %s. User: %v\n", modelName, user)

	admin, err := r.registry.Get(modelName)
	if err != nil {
		fmt.Printf("DEBUG: Model not found: %s. Error: %v\n", modelName, err)
		respondError(w, http.StatusNotFound, "model_not_found", err.Error(), nil)
		return
	}

	// Check module permission
	if !admin.HasModulePermission(ctx, user) {
		respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to access this model", nil)
		return
	}

	meta, err := admin.GetMetadata(ctx, user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}

	respondJSON(w, http.StatusOK, meta)
}

// handleList returns a paginated list of objects
func (r *Router) handleList(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)

		fmt.Printf("DEBUG: handleList for %s. User: %v\n", admin.ModelName(), user)
		// Check view permission
		hasPerm := admin.HasViewPermission(ctx, user, nil)
		fmt.Printf("DEBUG: HasViewPermission: %v\n", hasPerm)
		if !hasPerm {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to view this model", nil)
			return
		}

		// Parse query parameters
		params := parseListParams(req)

		// Call implementation
		response, err := admin.ListObjects(ctx, params)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "list_failed", err.Error(), nil)
			return
		}

		respondJSON(w, http.StatusOK, response)
	}
}

// handleDetail returns a single object
func (r *Router) handleDetail(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)
		idStr := chi.URLParam(req, "id")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_id", "Invalid ID format", nil)
			return
		}

		// Check view permission
		if !admin.HasViewPermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to view this object", nil)
			return
		}

		// Call implementation
		obj, err := admin.GetObject(ctx, id)
		if err != nil {
			respondError(w, http.StatusNotFound, "not_found", "Object not found", nil)
			return
		}

		respondJSON(w, http.StatusOK, obj)
	}
}

// handleCreate creates a new object
func (r *Router) handleCreate(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)

		// Check add permission
		if !admin.HasAddPermission(ctx, user) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to add this model", nil)
			return
		}

		// Parse body
		var data map[string]interface{}
		if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_body", err.Error(), nil)
			return
		}

		// Call implementation
		obj, err := admin.CreateObject(ctx, data)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "create_failed", err.Error(), nil)
			return
		}

		respondJSON(w, http.StatusCreated, obj)
	}
}

// handleUpdate partially updates an object
func (r *Router) handleUpdate(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)
		idStr := chi.URLParam(req, "id")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_id", "Invalid ID format", nil)
			return
		}

		// Check change permission
		if !admin.HasChangePermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to change this object", nil)
			return
		}

		// Parse body
		var data map[string]interface{}
		if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_body", err.Error(), nil)
			return
		}

		// Call implementation
		obj, err := admin.UpdateObject(ctx, id, data)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "update_failed", err.Error(), nil)
			return
		}

		respondJSON(w, http.StatusOK, obj)
	}
}

// handleReplace fully replaces an object
func (r *Router) handleReplace(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)
		idStr := chi.URLParam(req, "id")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_id", "Invalid ID format", nil)
			return
		}

		// Check change permission
		if !admin.HasChangePermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to change this object", nil)
			return
		}

		// TODO: Fetch object, parse request body, validate, replace object
		_ = id

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Object replaced",
		})
	}
}

// handleDelete deletes an object
func (r *Router) handleDelete(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)
		idStr := chi.URLParam(req, "id")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_id", "Invalid ID format", nil)
			return
		}

		// Check delete permission
		if !admin.HasDeletePermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to delete this object", nil)
			return
		}

		// Call implementation
		err = admin.DeleteObject(ctx, id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "delete_failed", err.Error(), nil)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// handleBulkCreate creates multiple objects
func (r *Router) handleBulkCreate(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)

		// Check add permission
		if !admin.HasAddPermission(ctx, user) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to add this model", nil)
			return
		}

		// TODO: Parse request body, validate, create objects

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"created": 0,
			"objects": []interface{}{},
		})
	}
}

// handleBulkUpdate updates multiple objects
func (r *Router) handleBulkUpdate(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)

		// Check change permission
		if !admin.HasChangePermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to change this model", nil)
			return
		}

		// TODO: Parse request body, validate, update objects

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"updated": 0,
		})
	}
}

// handleBulkDelete deletes multiple objects
func (r *Router) handleBulkDelete(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)

		// Check delete permission
		if !admin.HasDeletePermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to delete this model", nil)
			return
		}

		// TODO: Parse request body, validate, delete objects

		w.WriteHeader(http.StatusNoContent)
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

		// Call implementation
		response, err := admin.ExecuteAction(ctx, actionName, request.IDs, request.Params)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "action_failed", err.Error(), nil)
			return
		}

		respondJSON(w, http.StatusOK, response)
	}
}

// handleAutocomplete returns autocomplete suggestions
func (r *Router) handleAutocomplete(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)

		// Check view permission
		if !admin.HasViewPermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to view this model", nil)
			return
		}

		field := req.URL.Query().Get("field")
		query := req.URL.Query().Get("q")
		limit := 10

		if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		results, err := admin.Autocomplete(ctx, query, limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "autocomplete_failed", err.Error(), nil)
			return
		}

		_ = field // Currently not using field-specific autocomplete, but available for future

		respondJSON(w, http.StatusOK, core.AutocompleteResponse{
			Results: results,
		})
	}
}

// handleGlobalSearch handles global search across all models
func (r *Router) handleGlobalSearch(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	query := req.URL.Query().Get("q")

	if query == "" {
		respondError(w, http.StatusBadRequest, "missing_query", "Search query is required", nil)
		return
	}

	allAdmins := r.registry.GetAll()
	results := make([]core.SearchResultGroup, 0)

	for name, admin := range allAdmins {
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
				URL:   fmt.Sprintf("/admin/%s/%v", name, item.Value),
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

type savedView struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Filters   map[string]interface{} `json:"filters"`
	Ordering  []string               `json:"ordering"`
	Display   []string               `json:"display"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type savedViewRequest struct {
	Name     string                 `json:"name"`
	Filters  map[string]interface{} `json:"filters"`
	Ordering []string               `json:"ordering"`
	Display  []string               `json:"display"`
}

type savedViewStore struct {
	mu    sync.RWMutex
	views map[string]map[string][]savedView
}

func newSavedViewStore() *savedViewStore {
	return &savedViewStore{views: make(map[string]map[string][]savedView)}
}

func (s *savedViewStore) list(userID, model string) []savedView {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userViews, ok := s.views[userID]
	if !ok {
		return []savedView{}
	}
	modelViews := userViews[model]
	if modelViews == nil {
		return []savedView{}
	}
	return append([]savedView{}, modelViews...)
}

func (s *savedViewStore) upsert(userID, model string, request savedViewRequest) savedView {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.views[userID] == nil {
		s.views[userID] = make(map[string][]savedView)
	}

	views := s.views[userID][model]
	for i, view := range views {
		if strings.EqualFold(view.Name, request.Name) {
			updated := view
			updated.Filters = request.Filters
			updated.Ordering = request.Ordering
			updated.Display = request.Display
			updated.UpdatedAt = time.Now()
			views[i] = updated
			s.views[userID][model] = views
			return updated
		}
	}

	newView := savedView{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:      request.Name,
		Filters:   request.Filters,
		Ordering:  request.Ordering,
		Display:   request.Display,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.views[userID][model] = append(views, newView)
	return newView
}

func userKey(user interface{}) string {
	if user == nil {
		return "anonymous"
	}

	val := reflect.ValueOf(user)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.IsValid() && val.Kind() == reflect.Struct {
		for _, name := range []string{"ID", "Id", "id", "UserID", "Username", "Email"} {
			field := val.FieldByName(name)
			if field.IsValid() {
				return fmt.Sprintf("%v", field.Interface())
			}
		}
	}

	return fmt.Sprintf("%v", user)
}

func (r *Router) handleSavedViewsList(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	modelName := chi.URLParam(req, "model")
	user, _ := apicore.UserFromContext(ctx)

	admin, err := r.registry.Get(modelName)
	if err != nil {
		respondError(w, http.StatusNotFound, "model_not_found", err.Error(), nil)
		return
	}
	if !admin.HasViewPermission(ctx, user, nil) {
		respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to view this model", nil)
		return
	}

	views := r.views.list(userKey(user), modelName)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"views": views,
	})
}

func (r *Router) handleSavedViewSave(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	modelName := chi.URLParam(req, "model")
	user, _ := apicore.UserFromContext(ctx)

	admin, err := r.registry.Get(modelName)
	if err != nil {
		respondError(w, http.StatusNotFound, "model_not_found", err.Error(), nil)
		return
	}
	if !admin.HasViewPermission(ctx, user, nil) {
		respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to view this model", nil)
		return
	}

	var request savedViewRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", err.Error(), nil)
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" {
		respondError(w, http.StatusBadRequest, "missing_name", "View name is required", nil)
		return
	}

	view := r.views.upsert(userKey(user), modelName, request)
	respondJSON(w, http.StatusOK, view)
}

// Helper types and functions

func parseListParams(req *http.Request) core.ListParams {
	params := core.ListParams{
		Page:     1,
		PageSize: 25,
		Filters:  make(map[string]interface{}),
	}

	if pageStr := req.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	if pageSizeStr := req.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			params.PageSize = pageSize
		}
	}

	params.Search = req.URL.Query().Get("search")

	if ordering := req.URL.Query().Get("ordering"); ordering != "" {
		params.Ordering = strings.Split(ordering, ",")
	}

	// Parse filters from query params
	for key, values := range req.URL.Query() {
		if key == "page" || key == "page_size" || key == "search" || key == "ordering" {
			continue
		}
		if len(values) > 0 {
			params.Filters[key] = values[0]
		}
	}

	return params
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code string, message string, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(core.ErrorResponse{
		Error: core.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
