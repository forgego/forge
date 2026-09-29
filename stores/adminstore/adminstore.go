// Package adminstore keeps the admin's saved views and change history in
// the framework store tables (see package stores), so that they survive a
// restart and are shared by every instance using the same database.
//
// admin.Site.UseDatabaseStores wires these, together with stores.AdminTokens
// and stores.LoginAttempts, into a site.
package adminstore

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/forgego/forge/admin/api/rest"
	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/stores"
	"github.com/google/uuid"
)

// History is a core.HistoryManager backed by the forge_admin_log table.
// Entries are kept until you delete them.
type History struct {
	stores *stores.Stores
}

// NewHistory returns the change history kept in s.
func NewHistory(s *stores.Stores) *History { return &History{stores: s} }

var _ core.HistoryManager = (*History)(nil)

// LogAction implements core.HistoryManager.
func (h *History) LogAction(ctx context.Context, entry core.LogEntry) error {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = h.stores.Now()
	}
	_, err := h.stores.DB().ExecContext(ctx,
		`INSERT INTO forge_admin_log
			(action_time, user_id, user_name, model_name, object_id, object_repr, action, change_stats)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		entry.Timestamp.UTC(), entry.UserID, entry.UserName, entry.ModelName, entry.ObjectID,
		entry.ObjectRepr, string(entry.Action), entry.ChangeStats,
	)
	return err
}

// GetHistory implements core.HistoryManager. An empty modelName or objectID
// matches every value; the model name is compared case-insensitively.
// Entries come newest first, at most 1000 of them.
func (h *History) GetHistory(ctx context.Context, modelName string, objectID string) ([]core.LogEntry, error) {
	rows, err := h.stores.DB().QueryContext(ctx,
		`SELECT id, action_time, user_id, user_name, model_name, object_id, object_repr, action, change_stats
		FROM forge_admin_log
		WHERE ($1 = '' OR lower(model_name) = $1) AND ($2 = '' OR object_id = $2)
		ORDER BY id DESC
		LIMIT 1000`,
		strings.ToLower(modelName), objectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]core.LogEntry, 0)
	for rows.Next() {
		var entry core.LogEntry
		var action string
		if err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.UserID, &entry.UserName, &entry.ModelName,
			&entry.ObjectID, &entry.ObjectRepr, &action, &entry.ChangeStats); err != nil {
			return nil, err
		}
		entry.Action = core.ActionType(action)
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// SavedViews is a rest.SavedViewStore backed by forge_admin_saved_views.
type SavedViews struct {
	stores *stores.Stores
}

// NewSavedViews returns the saved views kept in s.
func NewSavedViews(s *stores.Stores) *SavedViews { return &SavedViews{stores: s} }

var (
	_ rest.SavedViewStore    = (*SavedViews)(nil)
	_ rest.TokenStore        = (*stores.AdminTokens)(nil)
	_ rest.LoginAttemptStore = (*stores.LoginAttempts)(nil)
)

// ListSavedViews implements rest.SavedViewStore, oldest first.
func (v *SavedViews) ListSavedViews(ctx context.Context, userKey, model string) ([]rest.SavedView, error) {
	rows, err := v.stores.DB().QueryContext(ctx,
		`SELECT id, name, filters, ordering, display, created_at, updated_at
		FROM forge_admin_saved_views WHERE user_key = $1 AND model = $2
		ORDER BY created_at, id`,
		userKey, model,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	views := make([]rest.SavedView, 0)
	for rows.Next() {
		view, err := scanView(rows)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, rows.Err()
}

type scanner interface{ Scan(dest ...any) error }

func scanView(row scanner) (rest.SavedView, error) {
	var view rest.SavedView
	var filters, ordering, display string
	if err := row.Scan(&view.ID, &view.Name, &filters, &ordering, &display, &view.CreatedAt, &view.UpdatedAt); err != nil {
		return rest.SavedView{}, err
	}
	for _, field := range []struct {
		raw string
		dst any
	}{{filters, &view.Filters}, {ordering, &view.Ordering}, {display, &view.Display}} {
		if err := json.Unmarshal([]byte(field.raw), field.dst); err != nil {
			return rest.SavedView{}, err
		}
	}
	return view, nil
}

// SaveView implements rest.SavedViewStore. It inserts the view, or updates
// the view of the same user and model whose name matches case-insensitively,
// in one statement.
func (v *SavedViews) SaveView(ctx context.Context, userKey, model string, request rest.SavedViewRequest) (rest.SavedView, bool, error) {
	filters, err := json.Marshal(request.Filters)
	if err != nil {
		return rest.SavedView{}, false, err
	}
	ordering, err := json.Marshal(request.Ordering)
	if err != nil {
		return rest.SavedView{}, false, err
	}
	display, err := json.Marshal(request.Display)
	if err != nil {
		return rest.SavedView{}, false, err
	}
	id := uuid.NewString()
	now := v.stores.Now().UTC()
	row := v.stores.DB().QueryRowContext(ctx,
		`INSERT INTO forge_admin_saved_views
			(id, user_key, model, name, name_key, filters, ordering, display, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		ON CONFLICT (user_key, model, name_key) DO UPDATE SET
			filters = excluded.filters, ordering = excluded.ordering,
			display = excluded.display, updated_at = excluded.updated_at
		RETURNING id, name, filters, ordering, display, created_at, updated_at`,
		id, userKey, model, request.Name, strings.ToLower(request.Name),
		string(filters), string(ordering), string(display), now,
	)
	view, err := scanView(row)
	if err != nil {
		return rest.SavedView{}, false, err
	}
	return view, view.ID == id, nil
}

// DeleteSavedView implements rest.SavedViewStore.
func (v *SavedViews) DeleteSavedView(ctx context.Context, userKey, model, id string) (bool, error) {
	result, err := v.stores.DB().ExecContext(ctx,
		"DELETE FROM forge_admin_saved_views WHERE id = $1 AND user_key = $2 AND model = $3",
		id, userKey, model,
	)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
