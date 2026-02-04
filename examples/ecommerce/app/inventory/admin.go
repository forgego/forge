package inventory

import (
	"context"

	"github.com/forgego/forge/admin"
)

// RegisterAdmin registers inventory models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// Warehouse admin
	admin.Register(&admin.Config[Warehouse]{
		Icon: "Warehouse",
		ListDisplay: []admin.Field{
			WarehouseFieldsInstance.Name,
			WarehouseFieldsInstance.Code,
			WarehouseFieldsInstance.City,
			WarehouseFieldsInstance.CountryCode,
			WarehouseFieldsInstance.IsActive,
			WarehouseFieldsInstance.IsPrimary,
		},
		ListFilter: []admin.Field{
			WarehouseFieldsInstance.IsActive,
			WarehouseFieldsInstance.IsPrimary,
			WarehouseFieldsInstance.CountryCode,
		},
		Actions: []admin.Action[Warehouse]{
			{
				Name:  "activate",
				Label: "Activate Warehouses",
				Handler: func(ctx context.Context, instances []*Warehouse) error {
					for _, warehouse := range instances {
						warehouse.IsActive = true
					}
					return nil
				},
			},
		},
	})

	// Stock admin
	admin.Register(&admin.Config[Stock]{
		Icon: "Activity",
		ListDisplay: []admin.Field{
			StockFieldsInstance.ProductVariantId,
			StockFieldsInstance.WarehouseId,
			StockFieldsInstance.Quantity,
			StockFieldsInstance.ReservedQuantity,
			StockFieldsInstance.IsActive,
		},
		ListFilter: []admin.Field{
			StockFieldsInstance.WarehouseId,
			StockFieldsInstance.IsActive,
		},
		Actions: []admin.Action[Stock]{
			{
				Name:  "activate",
				Label: "Activate Stock Records",
				Handler: func(ctx context.Context, instances []*Stock) error {
					for _, stock := range instances {
						stock.IsActive = true
					}
					return nil
				},
			},
		},
	})

	// StockMovement admin
	admin.Register(&admin.Config[StockMovement]{
		Icon: "TrendingUp",
		ListDisplay: []admin.Field{
			StockMovementFieldsInstance.ProductVariantId,
			StockMovementFieldsInstance.WarehouseId,
			StockMovementFieldsInstance.Type,
			StockMovementFieldsInstance.Quantity,
			StockMovementFieldsInstance.MovementDate,
		},
		ListFilter: []admin.Field{
			StockMovementFieldsInstance.Type,
			StockMovementFieldsInstance.WarehouseId,
			StockMovementFieldsInstance.MovementDate,
		},
		Actions: []admin.Action[StockMovement]{
			{
				Name:  "mark_adjustment",
				Label: "Mark as Adjustment",
				Handler: func(ctx context.Context, instances []*StockMovement) error {
					for _, movement := range instances {
						movement.Type = "adjustment"
					}
					return nil
				},
			},
		},
	})

	// StockAlert admin
	admin.Register(&admin.Config[StockAlert]{
		Icon: "AlertTriangle",
		ListDisplay: []admin.Field{
			StockAlertFieldsInstance.ProductVariantId,
			StockAlertFieldsInstance.WarehouseId,
			StockAlertFieldsInstance.AlertType,
			StockAlertFieldsInstance.CurrentQuantity,
			StockAlertFieldsInstance.Status,
		},
		ListFilter: []admin.Field{
			StockAlertFieldsInstance.AlertType,
			StockAlertFieldsInstance.Status,
			StockAlertFieldsInstance.WarehouseId,
		},
		Actions: []admin.Action[StockAlert]{
			{
				Name:  "resolve",
				Label: "Mark Resolved",
				Handler: func(ctx context.Context, instances []*StockAlert) error {
					for _, alert := range instances {
						alert.Status = "resolved"
					}
					return nil
				},
			},
		},
	})

	// StockTransfer admin
	admin.Register(&admin.Config[StockTransfer]{
		Icon: "Repeat",
		ListDisplay: []admin.Field{
			StockTransferFieldsInstance.TransferNumber,
			StockTransferFieldsInstance.FromWarehouseId,
			StockTransferFieldsInstance.ToWarehouseId,
			StockTransferFieldsInstance.Quantity,
			StockTransferFieldsInstance.Status,
		},
		ListFilter: []admin.Field{
			StockTransferFieldsInstance.Status,
			StockTransferFieldsInstance.FromWarehouseId,
			StockTransferFieldsInstance.ToWarehouseId,
		},
		Actions: []admin.Action[StockTransfer]{
			{
				Name:  "complete",
				Label: "Mark Completed",
				Handler: func(ctx context.Context, instances []*StockTransfer) error {
					for _, transfer := range instances {
						transfer.Status = "completed"
					}
					return nil
				},
			},
		},
	})
}
