package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/forgego/forge/admin/core"
	apicore "github.com/forgego/forge/api/core"
	"github.com/go-chi/chi/v5"
)

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

func isFKOrOneToOne(relType string) bool {
	switch relType {
	case "ForeignKey", "OneToOne", "foreign_key", "one_to_one":
		return true
	default:
		return false
	}
}

func (r *Router) attachDisplayLabels(ctx context.Context, admin core.AdminInterface, user interface{}, response *core.PaginatedResponse) {
	if r.registry == nil || response == nil || response.Results == nil {
		return
	}
	meta, err := admin.GetMetadata(ctx, user)
	if err != nil || meta == nil || len(meta.Relations) == 0 {
		return
	}
	rows, ok := displayRows(response.Results)
	if !ok {
		return
	}
	for _, relation := range meta.Relations {
		r.attachRelationLabels(ctx, response, rows, relation)
	}
}

func displayRows(results interface{}) ([]map[string]interface{}, bool) {
	data, err := json.Marshal(results)
	if err != nil {
		return nil, false
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal(data, &rows); err != nil || len(rows) == 0 {
		return nil, false
	}
	return rows, true
}

func (r *Router) attachRelationLabels(ctx context.Context, response *core.PaginatedResponse, rows []map[string]interface{}, relation core.RelationMetadata) {
	if !isFKOrOneToOne(relation.Type) {
		return
	}
	ids := collectRelationIDs(rows, relation.Name)
	if len(ids) == 0 {
		return
	}
	labels := r.loadRelationLabels(ctx, relation.RelatedModel, ids)
	applyRelationLabels(response, relation.Name, labels)
}

func (r *Router) loadRelationLabels(ctx context.Context, relatedModel string, ids []interface{}) map[string]string {
	related, err := r.registry.Get(relatedModel)
	if err != nil || related == nil {
		return nil
	}
	resolver, ok := related.(core.LabelResolver)
	if !ok {
		return nil
	}
	labels, err := resolver.ObjectLabels(ctx, ids)
	if err != nil || len(labels) == 0 {
		return nil
	}
	return labels
}

func applyRelationLabels(response *core.PaginatedResponse, relationName string, labels map[string]string) {
	if len(labels) == 0 {
		return
	}
	if response.Display == nil {
		response.Display = make(map[string]map[string]string)
	}
	response.Display[relationName] = labels
}

func collectRelationIDs(rows []map[string]interface{}, relationName string) []interface{} {
	ids := make([]interface{}, 0)
	seen := make(map[string]bool)
	for _, row := range rows {
		value, ok := relationValue(row, relationName)
		if !ok {
			continue
		}
		value, ok = relationID(value)
		if !ok {
			continue
		}
		key := fmt.Sprint(value)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		ids = append(ids, value)
	}
	return ids
}

func relationValue(row map[string]interface{}, relationName string) (interface{}, bool) {
	for _, key := range []string{relationName, relationName + "_id"} {
		if value, ok := row[key]; ok && value != nil {
			return value, true
		}
	}
	for key, value := range row {
		if value != nil && (strings.EqualFold(key, relationName) || strings.EqualFold(key, relationName+"_id") || strings.EqualFold(key, relationName+"id")) {
			return value, true
		}
	}
	return nil, false
}

func relationID(value interface{}) (interface{}, bool) {
	if object, ok := value.(map[string]interface{}); ok {
		if id, ok := object["id"]; ok && id != nil {
			value = id
		} else if id, ok := object["ID"]; ok && id != nil {
			value = id
		} else {
			return nil, false
		}
	}
	if number, ok := value.(float64); ok && number == math.Floor(number) && !math.IsNaN(number) && !math.IsInf(number, 0) {
		return int64(number), true
	}
	return value, true
}
