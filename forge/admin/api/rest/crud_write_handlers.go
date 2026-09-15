package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/forgego/forge/admin/core"
	apicore "github.com/forgego/forge/api/core"
	validation "github.com/forgego/forge/validate"
	"github.com/go-chi/chi/v5"
)

type bulkItemError struct {
	Index   int    `json:"index"`
	Code    string `json:"code"`
	Message string `json:"message"`
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
			if isValidationError(err) {
				respondError(w, http.StatusBadRequest, "validation_error", err.Error(), validationDetails(err))
				return
			}
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
			if isValidationError(err) {
				respondError(w, http.StatusBadRequest, "validation_error", err.Error(), validationDetails(err))
				return
			}
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
			if isValidationError(err) {
				respondError(w, http.StatusBadRequest, "validation_error", err.Error(), validationDetails(err))
				return
			}
			respondError(w, http.StatusInternalServerError, "update_failed", err.Error(), nil)
			return
		}

		respondJSON(w, http.StatusOK, obj)
	}
}

// bulkFailure responds for a bulk operation in which every item failed.
// Client-caused failures (bad input, not found, permission denied,
// validation) map to 4xx; only unexpected storage errors stay 500.
func bulkFailure(w http.ResponseWriter, code, message string, errs []bulkItemError) {
	status := http.StatusBadRequest
	allPermissionDenied := len(errs) > 0
	for _, e := range errs {
		switch e.Code {
		case "permission_denied":
			continue
		case "invalid_item", "invalid_id", "not_found":
			allPermissionDenied = false
			continue
		default:
			allPermissionDenied = false
			if !isValidationError(errors.New(e.Message)) {
				status = http.StatusInternalServerError
			}
		}
	}
	if allPermissionDenied {
		status = http.StatusForbidden
	}
	respondError(w, status, code, message, map[string]interface{}{
		"errors": errs,
	})
}

// validationDetails converts field validation errors into the per-field
// details map the admin UI renders next to each input.
func validationDetails(err error) map[string]interface{} {
	var verrs *validation.ValidationErrors
	if !errors.As(err, &verrs) {
		return nil
	}
	grouped := make(map[string][]string, len(verrs.Errors))
	for _, e := range verrs.Errors {
		field := e.Field
		if field == "" {
			field = "non_field_errors"
		}
		grouped[field] = append(grouped[field], e.Message)
	}
	details := make(map[string]interface{}, len(grouped))
	for field, messages := range grouped {
		details[field] = messages
	}
	return details
}

func isValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "validation") ||
		strings.Contains(msg, "not null constraint") ||
		strings.Contains(msg, "violates not-null") ||
		strings.Contains(msg, "required") ||
		strings.Contains(msg, "invalid input") ||
		strings.Contains(msg, "cannot be null") ||
		strings.Contains(msg, "is required")
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
			bulkFailure(w, "create_failed", "Failed to create any objects", errors)
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
			bulkFailure(w, "update_failed", "Failed to update any objects", errors)
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
			bulkFailure(w, "delete_failed", "Failed to delete any objects", errors)
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
