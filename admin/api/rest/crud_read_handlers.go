package rest

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/forgego/forge/admin/core"
	apicore "github.com/forgego/forge/api/core"
	"github.com/go-chi/chi/v5"
)

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

		r.attachDisplayLabels(ctx, admin, user, response)

		respondJSON(w, http.StatusOK, response)
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

func normalizePathID(rawID string) (interface{}, error) {
	return normalizeBulkID(rawID)
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
		// Clamp so a bad `limit` can neither blow up memory nor kill results.
		const maxAutocompleteLimit = 50
		if limit < 1 {
			limit = 10
		}
		if limit > maxAutocompleteLimit {
			limit = maxAutocompleteLimit
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
