package core

import (
	"context"
	apicore "github.com/forgego/forge/api/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type actionPermissionChecker struct{ approve, objectApprove bool }

func (p *actionPermissionChecker) HasPermission(_ context.Context, _ any, permission string) bool {
	return permission != "approve" || p.approve
}
func (p *actionPermissionChecker) HasObjectPermission(_ context.Context, _ any, _ any, permission string) bool {
	return permission != "approve" || p.objectApprove
}
func TestActionsEnforceDeclaredPermissionsBeforeHandler(t *testing.T) {
	checker := &actionPermissionChecker{}
	calls := 0
	cfg := &Config[writeFieldsTestItem]{PermissionChecker: checker, Actions: []Action[writeFieldsTestItem]{NewAction("approve", "Approve", func(context.Context, []*writeFieldsTestItem) error { calls++; return nil }).WithPermissions("approve")}}
	admin, _, id, _ := setupWriteFieldsDB(t, cfg)
	ctx := apicore.WithUser(context.Background(), "operator")
	for _, stage := range []struct {
		name           string
		global, object bool
		affected       int
	}{
		{"global denial", false, true, 0}, {"object denial", true, false, 0}, {"grant", true, true, 1},
	} {
		t.Run(stage.name, func(t *testing.T) {
			checker.approve, checker.objectApprove = stage.global, stage.object
			result, err := admin.ExecuteAction(ctx, "approve", []any{id}, nil)
			require.NoError(t, err)
			response := result.(*BulkActionResponse)
			assert.Equal(t, stage.affected, response.Affected)
			if stage.affected == 0 {
				assert.Zero(t, calls)
				assert.Len(t, response.Errors, 1)
			}
		})
	}
	assert.Equal(t, 1, calls)
}
