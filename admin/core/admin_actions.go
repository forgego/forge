package core

import (
	"context"
	"fmt"

	apicore "github.com/forgego/forge/api/core"
)

func (a *Admin[T]) ExecuteAction(ctx context.Context, actionName string, ids []interface{}, params map[string]interface{}) (interface{}, error) {
	_ = params

	// Find action
	var selectedAction *Action[T]
	for _, action := range a.config.Actions {
		if action.Name == actionName {
			selectedAction = &action
			break
		}
	}

	if selectedAction == nil {
		return nil, fmt.Errorf("action %s not found", actionName)
	}

	user, _ := apicore.UserFromContext(ctx)

	// Fetch instances
	instances := make([]*T, 0, len(ids))
	actionErrors := make([]BulkActionError, 0)
	for _, id := range ids {
		intID, err := toInt64(id)
		if err != nil {
			actionErrors = append(actionErrors, BulkActionError{
				ID:      0,
				Code:    "invalid_id",
				Message: fmt.Sprintf("invalid id %v", id),
			})
			continue
		}

		instance, err := a.safeGetObjectByID(ctx, intID)
		if err != nil {
			actionErrors = append(actionErrors, BulkActionError{
				ID:      intID,
				Code:    "not_found",
				Message: "object not found",
			})
			continue
		}

		if !a.HasChangePermission(ctx, user, instance) {
			actionErrors = append(actionErrors, BulkActionError{
				ID:      intID,
				Code:    "permission_denied",
				Message: "permission denied",
			})
			continue
		}

		instances = append(instances, instance)
	}

	if len(instances) == 0 {
		response := &BulkActionResponse{
			Success:  false,
			Affected: 0,
			Message:  "No permitted objects found for action",
		}
		if len(actionErrors) > 0 {
			response.Errors = actionErrors
		}
		return response, nil
	}

	// Execute handler
	err := selectedAction.Handler(ctx, instances)
	if err != nil {
		return nil, err
	}

	response := &BulkActionResponse{
		Success:  len(actionErrors) == 0,
		Affected: len(instances),
		Message:  fmt.Sprintf("Successfully executed %s on %d objects", selectedAction.Label, len(instances)),
	}
	if len(actionErrors) > 0 {
		response.Errors = actionErrors
		response.Message = fmt.Sprintf("Executed %s on %d objects; %d skipped", selectedAction.Label, len(instances), len(actionErrors))
	}
	return response, nil
}
