package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/schema"
)

func hiddenResponseFields(admin core.AdminInterface) []string {
	modelType := admin.ModelType()
	if modelType == nil {
		return nil
	}
	for modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	return schema.NonSerializableFieldNames(reflect.New(modelType).Interface())
}

func safeAdminObject(admin core.AdminInterface, obj any) (any, error) {
	hidden := hiddenResponseFields(admin)
	if len(hidden) == 0 || obj == nil {
		return obj, nil
	}
	var result map[string]any
	if original, ok := obj.(map[string]any); ok {
		result = make(map[string]any, len(original))
		for key, value := range original {
			result[key] = value
		}
	} else {
		raw, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber() // Never round integer primary keys through float64.
		if err := decoder.Decode(&result); err != nil {
			return nil, err
		}
	}
	for _, name := range hidden {
		delete(result, name)
	}
	return result, nil
}

func respondAdminObject(w http.ResponseWriter, status int, admin core.AdminInterface, obj any) {
	data, err := safeAdminObject(admin, obj)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "serialization_failed", "Unable to serialize object", nil)
		return
	}
	respondJSON(w, status, data)
}

func filterAdminList(admin core.AdminInterface, response *core.PaginatedResponse) error {
	if len(hiddenResponseFields(admin)) == 0 {
		return nil
	}
	items := interfaceSlice(response.Results)
	results := make([]any, 0, len(items))
	for _, item := range items {
		safe, err := safeAdminObject(admin, item)
		if err != nil {
			return err
		}
		results = append(results, safe)
	}
	response.Results = results
	return nil
}

// visibleAdminRows applies the same object check used by detail and history
// before a row can reach a list response or export.
func visibleAdminRows(ctx context.Context, admin core.AdminInterface, user any, results any) []any {
	items := interfaceSlice(results)
	visible := make([]any, 0, len(items))
	for _, item := range items {
		if item != nil && admin.HasViewPermission(ctx, user, item) {
			visible = append(visible, item)
		}
	}
	return visible
}

// visibleAutocompleteItems resolves suggestions before exposing their labels.
func visibleAutocompleteItems(ctx context.Context, admin core.AdminInterface, user any, items []core.AutocompleteItem) []core.AutocompleteItem {
	visible := make([]core.AutocompleteItem, 0, len(items))
	for _, item := range items {
		obj, err := admin.GetObject(ctx, item.Value)
		if err != nil || obj == nil || !admin.HasViewPermission(ctx, user, obj) {
			continue
		}
		visible = append(visible, item)
	}
	return visible
}
