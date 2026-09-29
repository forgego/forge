package rest

import (
	"context"
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

// SavedView is a named list view (filters, ordering, displayed columns) a
// user saved for a model.
type SavedView struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Filters   map[string]interface{} `json:"filters"`
	Ordering  []string               `json:"ordering"`
	Display   []string               `json:"display"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// SavedViewRequest is the body of a save request.
type SavedViewRequest struct {
	Name     string                 `json:"name"`
	Filters  map[string]interface{} `json:"filters"`
	Ordering []string               `json:"ordering"`
	Display  []string               `json:"display"`
}

// SavedViewStore keeps saved views per user and model. Saving a view under
// the name (compared case-insensitively) of an existing view of the same
// user and model replaces its filters, ordering and display. The default
// keeps views in process memory; stores/adminstore keeps them in the
// database.
type SavedViewStore interface {
	ListSavedViews(ctx context.Context, userKey, model string) ([]SavedView, error)
	SaveView(ctx context.Context, userKey, model string, request SavedViewRequest) (view SavedView, created bool, err error)
	DeleteSavedView(ctx context.Context, userKey, model, id string) (bool, error)
}

type savedViewStore struct {
	mu    sync.RWMutex
	views map[string]map[string][]SavedView
}

func newSavedViewStore() *savedViewStore {
	return &savedViewStore{views: make(map[string]map[string][]SavedView)}
}

func (s *savedViewStore) list(userID, model string) []SavedView {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userViews, ok := s.views[userID]
	if !ok {
		return []SavedView{}
	}
	modelViews := userViews[model]
	if modelViews == nil {
		return []SavedView{}
	}
	return append([]SavedView{}, modelViews...)
}

func (s *savedViewStore) upsert(userID, model string, request SavedViewRequest) (SavedView, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.views[userID] == nil {
		s.views[userID] = make(map[string][]SavedView)
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

	newView := SavedView{
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

// ListSavedViews implements SavedViewStore.
func (s *savedViewStore) ListSavedViews(_ context.Context, userKey, model string) ([]SavedView, error) {
	return s.list(userKey, model), nil
}

// SaveView implements SavedViewStore.
func (s *savedViewStore) SaveView(_ context.Context, userKey, model string, request SavedViewRequest) (SavedView, bool, error) {
	view, created := s.upsert(userKey, model, request)
	return view, created, nil
}

// DeleteSavedView implements SavedViewStore.
func (s *savedViewStore) DeleteSavedView(_ context.Context, userKey, model, id string) (bool, error) {
	return s.delete(userKey, model, id), nil
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

	views, err := r.savedViews.ListSavedViews(ctx, userKey(user), modelName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "saved_views_unavailable", "Could not load saved views", nil)
		return
	}
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

	var request SavedViewRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", err.Error(), nil)
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" {
		respondError(w, http.StatusBadRequest, "missing_name", "View name is required", nil)
		return
	}

	view, created, err := r.savedViews.SaveView(ctx, userKey(user), modelName, request)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "saved_views_unavailable", "Could not save the view", nil)
		return
	}
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

	deleted, err := r.savedViews.DeleteSavedView(ctx, userKey(user), modelName, viewID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "saved_views_unavailable", "Could not delete the view", nil)
		return
	}
	if !deleted {
		respondError(w, http.StatusNotFound, "view_not_found", "Saved view not found", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
