package operations

import (
	"context"
	"fmt"

	"github.com/forgego/forge/operation"
)

const ApproveOrderPermission = "orders.approve_order"

type ApproveOrderInput struct {
	OrderID int64 `json:"order_id"`
}

type ApproveOrderResult struct {
	OrderID    int64  `json:"order_id"`
	Status     string `json:"status"`
	ApprovedBy string `json:"approved_by"`
}

// ApproveOrder conditionally transitions one pending order and stores its
// history in the runner's transaction.
var ApproveOrder = operation.Operation[ApproveOrderInput, ApproveOrderResult]{
	Name:       "orders.approve_order",
	Permission: ApproveOrderPermission,
	Audit: operation.AuditDetails{
		Model:  "orders",
		Source: "reference_app",
	},
	Run: func(ctx context.Context, tx operation.Transaction, actor operation.Actor, input ApproveOrderInput) (ApproveOrderResult, []operation.Effect, error) {
		result, err := tx.ExecContext(ctx, `UPDATE orders SET status = 'approved', approved_by = ? WHERE id = ? AND status = 'pending'`, actor.ID(), input.OrderID)
		if err != nil {
			return ApproveOrderResult{}, nil, fmt.Errorf("update pending order: %w", err)
		}
		changed, err := result.RowsAffected()
		if err != nil {
			return ApproveOrderResult{}, nil, fmt.Errorf("count updated orders: %w", err)
		}
		if changed == 0 {
			return ApproveOrderResult{}, nil, fmt.Errorf("order %d is not pending: %w", input.OrderID, operation.ErrConflict)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO order_history (order_id, status, changed_by) VALUES (?, 'approved', ?)`, input.OrderID, actor.ID()); err != nil {
			return ApproveOrderResult{}, nil, fmt.Errorf("record order history: %w", err)
		}
		return ApproveOrderResult{OrderID: input.OrderID, Status: "approved", ApprovedBy: actor.ID()}, nil, nil
	},
}
