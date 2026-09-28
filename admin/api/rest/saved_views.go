package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	apicore "github.com/forgego/forge/api/core"
	"github.com/go-chi/chi/v5"
)

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

func (s *savedViewStore) upsert(userID, model string, request savedViewRequest) (savedView, bool) {
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
			return updated, false
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
	return newView, true
}

// delete removes a saved view by ID. It reports whether a view was removed.
func (s *savedViewStore) delete(userID, model, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	userViews, ok := s.views[userID]
	if !ok {
		return false
	}
	views := userViews[model]
	for i, view := range views {
		if view.ID == id {
			userViews[model] = append(views[:i], views[i+1:]...)
			return true
		}
	}
	return false
}

func userKey(user interface{}) string {
	if user == nil {
		return "anonymous"
	}

	// The admin auth middleware stores the user as a map.
	if m, ok := user.(map[string]interface{}); ok {
		for _, key := range []string{"username", "name", "id", "email"} {
			if v, exists := m[key]; exists {
				if s := fmt.Sprintf("%v", v); s != "" && s != "<nil>" {
					return s
				}
			}
		}
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

	view, created := r.views.upsert(userKey(user), modelName, request)
	if created {
		respondJSON(w, http.StatusCreated, view)
		return
	}
	respondJSON(w, http.StatusOK, view)
}

func (r *Router) handleSavedViewDelete(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	modelName := chi.URLParam(req, "model")
	viewID := chi.URLParam(req, "id")
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

	if !r.views.delete(userKey(user), modelName, viewID) {
		respondError(w, http.StatusNotFound, "view_not_found", "Saved view not found", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
