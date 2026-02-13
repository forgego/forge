package rest

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
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
	sessions *adminSessionStore
}

// NewRouter creates a new admin API router
func NewRouter(registry *core.Registry) *Router {
	return &Router{
		registry: registry,
		prefix:   "/api",
		views:    newSavedViewStore(),
		sessions: newAdminSessionStore(),
	}
}

// RegisterRoutes registers all admin routes
func (r *Router) RegisterRoutes(router chi.Router) {
	router.Route(r.prefix, func(sub chi.Router) {
		// Middleware
		sub.Use(middleware.Logger)
		sub.Use(middleware.Recoverer)
		sub.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"}, // Admin panel may be served from any origin; restrict via reverse proxy in production
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
		basePath := fmt.Sprintf("/%s", name)

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

// handleHistory returns change history for an object
func (r *Router) handleHistory(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)
		idStr := chi.URLParam(req, "id")

		if idStr == "" {
			respondError(w, http.StatusBadRequest, "invalid_id", "Missing ID", nil)
			return
		}

		if !admin.HasViewPermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to view this object", nil)
			return
		}

		history, err := admin.GetHistory(ctx, idStr)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "history_failed", err.Error(), nil)
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"entries": history,
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
		"environment": adminEnvironment(),
		"dashboard":   core.GetDashboard(ctx),
	}

	respondJSON(w, http.StatusOK, config)
}

// handleMetaList returns list of all models
func (r *Router) handleMetaList(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	user, _ := apicore.UserFromContext(ctx)

	allAdmins := r.registry.GetAll()
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

		count := modelCountFromList(ctx, admin)

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

func adminEnvironment() string {
	for _, key := range []string{"FORGE_ENV", "APP_ENV", "GO_ENV", "ENV"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return "development"
}

func modelCountFromList(ctx context.Context, admin core.AdminInterface) int64 {
	response, err := admin.ListObjects(ctx, core.ListParams{
		Page:     1,
		PageSize: 1,
		Filters:  map[string]interface{}{},
	})
	if err != nil || response == nil {
		return 0
	}
	return response.Count
}

// handleLogin handles admin login
func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", "Invalid login payload", nil)
		return
	}
	if decoder.More() {
		respondError(w, http.StatusBadRequest, "invalid_body", "Invalid login payload", nil)
		return
	}

	payload.Username = strings.TrimSpace(payload.Username)
	payload.Password = strings.TrimSpace(payload.Password)
	if payload.Username == "" || payload.Password == "" {
		respondError(w, http.StatusBadRequest, "invalid_credentials", "Username and password are required", nil)
		return
	}
	expectedUsername, expectedPassword := adminCredentials()
	if !secureEqual(payload.Username, expectedUsername) || !secureEqual(payload.Password, expectedPassword) {
		respondError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password", nil)
		return
	}

	token, err := r.sessions.Issue(payload.Username, 24*time.Hour)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "login_failed", "Could not create session token", nil)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]string{
			"name": payload.Username,
			"role": "superuser",
		},
		"expires_at": time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
	})
}

func adminCredentials() (string, string) {
	username := strings.TrimSpace(os.Getenv("FORGE_ADMIN_USERNAME"))
	if username == "" {
		username = "admin"
	}

	password := strings.TrimSpace(os.Getenv("FORGE_ADMIN_PASSWORD"))
	if password == "" {
		password = "secret"
	}

	return username, password
}

func secureEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// handleMetaDetail returns detailed metadata for a model
func (r *Router) handleMetaDetail(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	modelName := chi.URLParam(req, "model")
	user, _ := apicore.UserFromContext(ctx)

	admin, err := r.registry.Get(modelName)
	if err != nil {
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

		// Check view permission
		if !admin.HasViewPermission(ctx, user, nil) {
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

// handleExport returns list results as CSV or JSON using current filters/search.
func (r *Router) handleExport(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)

		if !admin.HasViewPermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to export this model", nil)
			return
		}

		meta, err := admin.GetMetadata(ctx, user)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		listFields, labelMap := exportListFields(meta)
		if len(listFields) == 0 {
			respondError(w, http.StatusBadRequest, "export_failed", "No fields available for export", nil)
			return
		}

		params := parseListParams(req)
		maxPageSize := meta.Pagination.MaxPageSize
		if maxPageSize <= 0 {
			maxPageSize = 100
		}
		params.PageSize = maxPageSize
		params.Page = 1

		results := make([]map[string]interface{}, 0)
		for {
			response, err := admin.ListObjects(ctx, params)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "export_failed", err.Error(), nil)
				return
			}

			items := interfaceSlice(response.Results)
			if len(items) == 0 {
				break
			}

			for _, item := range items {
				rowMap := make(map[string]interface{}, len(listFields))
				itemMap := objectToMap(item)
				for _, field := range listFields {
					rowMap[field] = getMapValue(itemMap, field)
				}
				results = append(results, rowMap)
			}

			if int64(len(results)) >= response.Count || len(items) < params.PageSize {
				break
			}
			params.Page++
		}

		format := strings.ToLower(req.URL.Query().Get("format"))
		if format == "" {
			format = "json"
		}

		switch format {
		case "csv":
			if err := writeCSVExport(w, admin.ModelName(), listFields, labelMap, results); err != nil {
				respondError(w, http.StatusInternalServerError, "export_failed", err.Error(), nil)
			}
		case "json":
			respondJSON(w, http.StatusOK, results)
		default:
			respondError(w, http.StatusBadRequest, "invalid_format", "Unsupported export format", nil)
		}
	}
}

// handleDetail returns a single object
func (r *Router) handleDetail(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)
		idStr := chi.URLParam(req, "id")

		id, err := normalizePathID(idStr)
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

		id, err := normalizePathID(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_id", "Invalid ID format", nil)
			return
		}

		// Check change permission
		if !admin.HasChangePermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to change this object", nil)
			return
		}

		existing, err := admin.GetObject(ctx, id)
		if err != nil {
			respondError(w, http.StatusNotFound, "not_found", "Object not found", nil)
			return
		}
		if !admin.HasChangePermission(ctx, user, existing) {
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

		id, err := normalizePathID(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_id", "Invalid ID format", nil)
			return
		}

		// Check change permission
		if !admin.HasChangePermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to change this object", nil)
			return
		}

		existing, err := admin.GetObject(ctx, id)
		if err != nil {
			respondError(w, http.StatusNotFound, "not_found", "Object not found", nil)
			return
		}

		if !admin.HasChangePermission(ctx, user, existing) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to change this object", nil)
			return
		}

		var data map[string]interface{}
		if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_body", err.Error(), nil)
			return
		}
		if len(data) == 0 {
			respondError(w, http.StatusBadRequest, "invalid_body", "Request body must include at least one field", nil)
			return
		}

		obj, err := admin.UpdateObject(ctx, id, data)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "update_failed", err.Error(), nil)
			return
		}

		respondJSON(w, http.StatusOK, obj)
	}
}

// handleDelete deletes an object
func (r *Router) handleDelete(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)
		idStr := chi.URLParam(req, "id")

		id, err := normalizePathID(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_id", "Invalid ID format", nil)
			return
		}

		// Check delete permission
		if !admin.HasDeletePermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to delete this object", nil)
			return
		}

		existing, err := admin.GetObject(ctx, id)
		if err != nil {
			respondError(w, http.StatusNotFound, "not_found", "Object not found", nil)
			return
		}
		if !admin.HasDeletePermission(ctx, user, existing) {
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

		var rawBody interface{}
		if err := json.NewDecoder(req.Body).Decode(&rawBody); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_body", err.Error(), nil)
			return
		}

		var rawObjects []interface{}
		switch payload := rawBody.(type) {
		case []interface{}:
			rawObjects = payload
		case map[string]interface{}:
			objectsValue, exists := payload["objects"]
			if !exists {
				respondError(w, http.StatusBadRequest, "invalid_body", "Request body must be an array of objects or include an 'objects' array", nil)
				return
			}
			objectsArray, ok := objectsValue.([]interface{})
			if !ok {
				respondError(w, http.StatusBadRequest, "invalid_body", "'objects' must be an array", nil)
				return
			}
			rawObjects = objectsArray
		default:
			respondError(w, http.StatusBadRequest, "invalid_body", "Request body must be an array of objects or include an 'objects' array", nil)
			return
		}

		if len(rawObjects) == 0 {
			respondError(w, http.StatusBadRequest, "invalid_body", "Request body must include at least one object", nil)
			return
		}

		objects := make([]interface{}, 0, len(rawObjects))
		errors := make([]bulkItemError, 0)
		hasCreateFailure := false

		for i, rawObject := range rawObjects {
			data, ok := rawObject.(map[string]interface{})
			if !ok {
				errors = append(errors, bulkItemError{
					Index:   i,
					Code:    "invalid_item",
					Message: "Each item must be a JSON object",
				})
				continue
			}
			if len(data) == 0 {
				errors = append(errors, bulkItemError{
					Index:   i,
					Code:    "invalid_item",
					Message: "Each item must include at least one field",
				})
				continue
			}

			obj, err := admin.CreateObject(ctx, data)
			if err != nil {
				hasCreateFailure = true
				errors = append(errors, bulkItemError{
					Index:   i,
					Code:    "create_failed",
					Message: err.Error(),
				})
				continue
			}
			objects = append(objects, obj)
		}

		if len(objects) == 0 {
			if hasCreateFailure {
				respondError(w, http.StatusInternalServerError, "create_failed", "Failed to create any objects", map[string]interface{}{
					"errors": errors,
				})
				return
			}
			respondError(w, http.StatusBadRequest, "invalid_body", "No valid objects to create", map[string]interface{}{
				"errors": errors,
			})
			return
		}

		status := http.StatusCreated
		if len(errors) > 0 {
			status = http.StatusMultiStatus
		}

		response := map[string]interface{}{
			"created": len(objects),
			"objects": objects,
		}
		if len(errors) > 0 {
			response["errors"] = errors
		}

		respondJSON(w, status, response)
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

		var payload struct {
			IDs  []interface{}          `json:"ids"`
			Data map[string]interface{} `json:"data"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_body", err.Error(), nil)
			return
		}
		if len(payload.IDs) == 0 {
			respondError(w, http.StatusBadRequest, "invalid_body", "Request body must include at least one ID", nil)
			return
		}
		if len(payload.Data) == 0 {
			respondError(w, http.StatusBadRequest, "invalid_body", "Request body must include update data", nil)
			return
		}

		objects := make([]interface{}, 0, len(payload.IDs))
		errors := make([]bulkItemError, 0)
		hasUpdateFailure := false

		for i, rawID := range payload.IDs {
			id, err := normalizeBulkID(rawID)
			if err != nil {
				errors = append(errors, bulkItemError{
					Index:   i,
					Code:    "invalid_id",
					Message: err.Error(),
				})
				continue
			}

			existing, err := admin.GetObject(ctx, id)
			if err != nil {
				errors = append(errors, bulkItemError{
					Index:   i,
					Code:    "not_found",
					Message: "Object not found",
				})
				continue
			}
			if !admin.HasChangePermission(ctx, user, existing) {
				errors = append(errors, bulkItemError{
					Index:   i,
					Code:    "permission_denied",
					Message: "You don't have permission to change this object",
				})
				continue
			}

			obj, err := admin.UpdateObject(ctx, id, payload.Data)
			if err != nil {
				hasUpdateFailure = true
				errors = append(errors, bulkItemError{
					Index:   i,
					Code:    "update_failed",
					Message: err.Error(),
				})
				continue
			}
			objects = append(objects, obj)
		}

		if len(objects) == 0 {
			if hasUpdateFailure {
				respondError(w, http.StatusInternalServerError, "update_failed", "Failed to update any objects", map[string]interface{}{
					"errors": errors,
				})
				return
			}
			respondError(w, http.StatusBadRequest, "invalid_body", "No valid objects to update", map[string]interface{}{
				"errors": errors,
			})
			return
		}

		status := http.StatusOK
		if len(errors) > 0 {
			status = http.StatusMultiStatus
		}

		response := map[string]interface{}{
			"updated": len(objects),
			"objects": objects,
		}
		if len(errors) > 0 {
			response["errors"] = errors
		}

		respondJSON(w, status, response)
	}
}

func normalizeBulkID(rawID interface{}) (interface{}, error) {
	switch id := rawID.(type) {
	case float64:
		if id != float64(int64(id)) {
			return nil, fmt.Errorf("ID must be an integer")
		}
		return int64(id), nil
	case int:
		return int64(id), nil
	case int64:
		return id, nil
	case string:
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			return nil, fmt.Errorf("ID cannot be empty")
		}
		if parsed, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return parsed, nil
		}
		return trimmed, nil
	default:
		return nil, fmt.Errorf("ID must be a string or number")
	}
}

func normalizePathID(rawID string) (interface{}, error) {
	return normalizeBulkID(rawID)
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

		var payload struct {
			IDs []interface{} `json:"ids"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_body", err.Error(), nil)
			return
		}
		if len(payload.IDs) == 0 {
			respondError(w, http.StatusBadRequest, "invalid_body", "Request body must include at least one ID", nil)
			return
		}

		deleted := 0
		errors := make([]bulkItemError, 0)
		hasDeleteFailure := false

		for i, rawID := range payload.IDs {
			id, err := normalizeBulkID(rawID)
			if err != nil {
				errors = append(errors, bulkItemError{
					Index:   i,
					Code:    "invalid_id",
					Message: err.Error(),
				})
				continue
			}

			existing, err := admin.GetObject(ctx, id)
			if err != nil {
				errors = append(errors, bulkItemError{
					Index:   i,
					Code:    "not_found",
					Message: "Object not found",
				})
				continue
			}
			if !admin.HasDeletePermission(ctx, user, existing) {
				errors = append(errors, bulkItemError{
					Index:   i,
					Code:    "permission_denied",
					Message: "You don't have permission to delete this object",
				})
				continue
			}

			if err := admin.DeleteObject(ctx, id); err != nil {
				hasDeleteFailure = true
				errors = append(errors, bulkItemError{
					Index:   i,
					Code:    "delete_failed",
					Message: err.Error(),
				})
				continue
			}

			deleted++
		}

		if deleted == 0 {
			if hasDeleteFailure {
				respondError(w, http.StatusInternalServerError, "delete_failed", "Failed to delete any objects", map[string]interface{}{
					"errors": errors,
				})
				return
			}
			respondError(w, http.StatusBadRequest, "invalid_body", "No valid objects to delete", map[string]interface{}{
				"errors": errors,
			})
			return
		}

		if len(errors) > 0 {
			respondJSON(w, http.StatusMultiStatus, map[string]interface{}{
				"deleted": deleted,
				"errors":  errors,
			})
			return
		}

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

type bulkItemError struct {
	Index   int    `json:"index"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type savedViewStore struct {
	mu    sync.RWMutex
	views map[string]map[string][]savedView
}

type adminSession struct {
	Username  string
	ExpiresAt time.Time
}

type adminSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]adminSession
}

func newSavedViewStore() *savedViewStore {
	return &savedViewStore{views: make(map[string]map[string][]savedView)}
}

func newAdminSessionStore() *adminSessionStore {
	return &adminSessionStore{
		sessions: make(map[string]adminSession),
	}
}

func (s *adminSessionStore) Issue(username string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		return "", errors.New("invalid session ttl")
	}

	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(randomBytes)
	s.mu.Lock()
	s.sessions[token] = adminSession{
		Username:  username,
		ExpiresAt: time.Now().Add(ttl),
	}
	s.mu.Unlock()

	return token, nil
}

func (s *adminSessionStore) Validate(token string) (adminSession, bool) {
	s.mu.RLock()
	session, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok {
		return adminSession{}, false
	}

	if time.Now().After(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessions, token)
		s.mu.Unlock()
		return adminSession{}, false
	}

	return session, true
}

func bearerToken(authorizationHeader string) (string, error) {
	header := strings.TrimSpace(authorizationHeader)
	if header == "" {
		return "", errors.New("missing authorization header")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("invalid authorization scheme")
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", errors.New("missing bearer token")
	}
	return token, nil
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
		if key == "page" || key == "page_size" || key == "search" || key == "ordering" || key == "format" {
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

func exportListFields(meta *core.Metadata) ([]string, map[string]string) {
	listFields := meta.ListDisplay
	if len(listFields) == 0 {
		listFields = make([]string, 0, len(meta.Fields))
		for _, field := range meta.Fields {
			listFields = append(listFields, field.Name)
		}
	}

	labelMap := make(map[string]string, len(meta.Fields))
	for _, field := range meta.Fields {
		labelMap[field.Name] = field.Label
	}

	return listFields, labelMap
}

func interfaceSlice(results interface{}) []interface{} {
	if results == nil {
		return nil
	}
	val := reflect.ValueOf(results)
	if val.Kind() != reflect.Slice {
		return nil
	}
	items := make([]interface{}, val.Len())
	for i := 0; i < val.Len(); i++ {
		items[i] = val.Index(i).Interface()
	}
	return items
}

func objectToMap(obj interface{}) map[string]interface{} {
	if obj == nil {
		return map[string]interface{}{}
	}
	if mapObj, ok := obj.(map[string]interface{}); ok {
		return mapObj
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return map[string]interface{}{}
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return map[string]interface{}{}
	}
	return result
}

func getMapValue(data map[string]interface{}, field string) interface{} {
	if data == nil {
		return nil
	}
	parts := strings.Split(field, ".")
	var current interface{} = data
	for _, part := range parts {
		next, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current, ok = next[part]
		if !ok {
			return nil
		}
	}
	return current
}

func writeCSVExport(
	w http.ResponseWriter,
	modelName string,
	fields []string,
	labelMap map[string]string,
	rows []map[string]interface{},
) error {
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	headers := make([]string, len(fields))
	for i, field := range fields {
		if label, ok := labelMap[field]; ok && label != "" {
			headers[i] = label
		} else {
			headers[i] = field
		}
	}

	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, row := range rows {
		record := make([]string, len(fields))
		for i, field := range fields {
			record[i] = formatExportValue(row[field])
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s-export.csv", modelName)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(buf.Bytes())
	return err
}

func formatExportValue(value interface{}) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case time.Time:
		return typed.Format(time.RFC3339)
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(value)
	}
}
