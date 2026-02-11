package inventory

import (
	"context"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
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
		SearchFields: []admin.Field{
			WarehouseFieldsInstance.Name,
			WarehouseFieldsInstance.Code,
			WarehouseFieldsInstance.City,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[Warehouse]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "code", "contact_name", "contact_email", "contact_phone"},
			},
			{
				Name: "Address",
				Fields: []string{"address_line1", "address_line2", "city", "state", "postal_code", "country_code", "country_name"},
			},
			{
				Name: "Geolocation",
				Fields: []string{"latitude", "longitude"},
				Collapsed: true,
			},
			{
				Name: "Settings",
				Fields: []string{"is_active", "is_primary", "priority", "total_capacity"},
			},
		},
		// Inline relations
		InlineRelations: map[string]admin.InlineRelationConfig{
			"stock_records": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Stock Records",
				RelatedModel: "Stock",
				RelatedField: "warehouse_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"product_variant_id", "quantity", "reserved_quantity"},
				},
			},
			"stock_movements": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Stock Movements",
				RelatedModel: "StockMovement",
				RelatedField: "warehouse_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"product_variant_id", "type", "quantity", "movement_date"},
				},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "code", "is_active", "is_primary", "priority"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Warehouse], user interface{}, obj *Warehouse) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Warehouse], user interface{}, obj *Warehouse) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Warehouse], user interface{}, obj *Warehouse) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[Warehouse]{
			{
				Name:  "active_warehouses",
				Label: "Active Warehouses",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Warehouse], value interface{}) orm.QuerySet[Warehouse] {
					return qs.Filter(orm.F("is_active").Eq(true))
				},
			},
			{
				Name:  "primary_warehouse",
				Label: "Primary Warehouse",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Warehouse], value interface{}) orm.QuerySet[Warehouse] {
					return qs.Filter(orm.F("is_primary").Eq(true))
				},
			},
			{
				Name:  "by_country",
				Label: "By Country",
				Type:  admin.FilterTypeText,
				Handler: func(ctx context.Context, qs orm.QuerySet[Warehouse], value interface{}) orm.QuerySet[Warehouse] {
					return qs.Filter(orm.F("country_code").Eq(value))
				},
			},
		},
		// Custom actions
		Actions: []admin.Action[Warehouse]{
			{
				Name:  "activate",
				Label: "Activate Warehouses",
				Handler: func(ctx context.Context, instances []*Warehouse) error {
					for _, warehouse := range instances {
						warehouse.IsActive = true
						if err := WarehouseObjects.Update(ctx, warehouse); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "deactivate",
				Label: "Deactivate Warehouses",
				Handler: func(ctx context.Context, instances []*Warehouse) error {
					for _, warehouse := range instances {
						warehouse.IsActive = false
						if err := WarehouseObjects.Update(ctx, warehouse); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "set_primary",
				Label: "Set as Primary",
				Handler: func(ctx context.Context, instances []*Warehouse) error {
					for _, warehouse := range instances {
						warehouse.IsPrimary = true
						if err := WarehouseObjects.Update(ctx, warehouse); err != nil {
							return err
						}
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
			StockFieldsInstance.AvailableQuantity,
			StockFieldsInstance.IsActive,
		},
		ListFilter: []admin.Field{
			StockFieldsInstance.WarehouseId,
			StockFieldsInstance.IsActive,
		},
		SearchFields: []admin.Field{
			StockFieldsInstance.ProductVariantId,
			StockFieldsInstance.BinLocation,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[Stock]{
			{
				Name: "Stock Levels",
				Fields: []string{"product_variant_id", "warehouse_id", "quantity", "reserved_quantity", "available_quantity"},
			},
			{
				Name: "Thresholds",
				Fields: []string{"reorder_point", "reorder_quantity"},
				Collapsed: true,
			},
			{
				Name: "Location",
				Fields: []string{"bin_location"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active", "allow_backorder"},
			},
		},
		// Inline relations
		InlineRelations: map[string]admin.InlineRelationConfig{
			"movements": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Stock Movements",
				RelatedModel: "StockMovement",
				RelatedField: "stock_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"type", "quantity", "movement_date"},
				},
			},
			"alerts": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Stock Alerts",
				RelatedModel: "StockAlert",
				RelatedField: "stock_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"alert_type", "current_quantity", "status"},
				},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"quantity", "reserved_quantity", "reorder_point", "is_active"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at", "last_counted_at", "available_quantity"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Stock], user interface{}, obj *Stock) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Stock], user interface{}, obj *Stock) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Stock], user interface{}, obj *Stock) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[Stock]{
			{
				Name:  "low_stock",
				Label: "Low Stock",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Stock], value interface{}) orm.QuerySet[Stock] {
					return qs.Filter(orm.F("quantity").Lte(10))
				},
			},
			{
				Name:  "out_of_stock",
				Label: "Out of Stock",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Stock], value interface{}) orm.QuerySet[Stock] {
					return qs.Filter(orm.F("quantity").Eq(0))
				},
			},
			{
				Name:  "in_stock",
				Label: "In Stock",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Stock], value interface{}) orm.QuerySet[Stock] {
					return qs.Filter(orm.F("quantity").Gt(10))
				},
			},
			{
				Name:  "needs_reorder",
				Label: "Needs Reorder",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Stock], value interface{}) orm.QuerySet[Stock] {
					return qs.Filter(orm.F("quantity").Lte(10))
				},
			},
			{
				Name:  "has_reservations",
				Label: "Has Reservations",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Stock], value interface{}) orm.QuerySet[Stock] {
					return qs.Filter(orm.F("reserved_quantity").Gt(0))
				},
			},
		},
		// Custom actions
		Actions: []admin.Action[Stock]{
			{
				Name:  "activate",
				Label: "Activate Stock Records",
				Handler: func(ctx context.Context, instances []*Stock) error {
					for _, stock := range instances {
						stock.IsActive = true
						if err := StockObjects.Update(ctx, stock); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "deactivate",
				Label: "Deactivate Stock Records",
				Handler: func(ctx context.Context, instances []*Stock) error {
					for _, stock := range instances {
						stock.IsActive = false
						if err := StockObjects.Update(ctx, stock); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "adjust_quantity",
				Label: "Adjust Quantity",
				Handler: func(ctx context.Context, instances []*Stock) error {
					// Adjustment logic
					return nil
				},
			},
			{
				Name:  "update_reorder_point",
				Label: "Update Reorder Point",
				Handler: func(ctx context.Context, instances []*Stock) error {
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
		SearchFields: []admin.Field{
			StockMovementFieldsInstance.ReferenceNumber,
			StockMovementFieldsInstance.Reason,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[StockMovement]{
			{
				Name: "Movement Details",
				Fields: []string{"product_variant_id", "warehouse_id", "type", "quantity", "quantity_before", "quantity_after"},
			},
			{
				Name: "Reference",
				Fields: []string{"reference_type", "reference_id", "reference_number", "from_warehouse_id", "to_warehouse_id"},
			},
			{
				Name: "Additional Information",
				Fields: []string{"reason", "notes", "user_id", "user_name"},
				Collapsed: true,
			},
			{
				Name: "Cost Information",
				Fields: []string{"unit_cost", "total_cost"},
				Collapsed: true,
			},
			{
				Name: "Date",
				Fields: []string{"movement_date", "created_at"},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"type", "quantity", "reason"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "quantity_before", "quantity_after"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[StockMovement], user interface{}, obj *StockMovement) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[StockMovement], user interface{}, obj *StockMovement) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[StockMovement], user interface{}, obj *StockMovement) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[StockMovement]{
			{
				Name:  "incoming",
				Label: "Incoming Stock",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockMovement], value interface{}) orm.QuerySet[StockMovement] {
					return qs.Filter(orm.F("quantity").Gt(0))
				},
			},
			{
				Name:  "outgoing",
				Label: "Outgoing Stock",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockMovement], value interface{}) orm.QuerySet[StockMovement] {
					return qs.Filter(orm.F("quantity").Lt(0))
				},
			},
			{
				Name:  "by_type",
				Label: "By Movement Type",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "purchase", Label: "Purchase"},
					{Value: "sale", Label: "Sale"},
					{Value: "return", Label: "Return"},
					{Value: "adjustment", Label: "Adjustment"},
					{Value: "transfer", Label: "Transfer"},
					{Value: "damage", Label: "Damage"},
					{Value: "loss", Label: "Loss"},
					{Value: "count", Label: "Count"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[StockMovement], value interface{}) orm.QuerySet[StockMovement] {
					return qs.Filter(orm.F("type").Eq(value))
				},
			},
			{
				Name:  "date_range",
				Label: "Date Range",
				Type:  admin.FilterTypeDateRange,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockMovement], value interface{}) orm.QuerySet[StockMovement] {
					return qs.Filter(orm.F("movement_date").Gte(value))
				},
			},
		},
		// Custom actions
		Actions: []admin.Action[StockMovement]{
			{
				Name:  "mark_adjustment",
				Label: "Mark as Adjustment",
				Handler: func(ctx context.Context, instances []*StockMovement) error {
					for _, movement := range instances {
						movement.Type = "adjustment"
						if err := StockMovementObjects.Update(ctx, movement); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "mark_damage",
				Label: "Mark as Damage",
				Handler: func(ctx context.Context, instances []*StockMovement) error {
					for _, movement := range instances {
						movement.Type = "damage"
						if err := StockMovementObjects.Update(ctx, movement); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "export_movements",
				Label: "Export Movements",
				Handler: func(ctx context.Context, instances []*StockMovement) error {
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
			StockAlertFieldsInstance.Threshold,
			StockAlertFieldsInstance.Status,
		},
		ListFilter: []admin.Field{
			StockAlertFieldsInstance.AlertType,
			StockAlertFieldsInstance.Status,
			StockAlertFieldsInstance.WarehouseId,
		},
		SearchFields: []admin.Field{
			StockAlertFieldsInstance.ProductVariantId,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[StockAlert]{
			{
				Name: "Alert Details",
				Fields: []string{"product_variant_id", "warehouse_id", "alert_type", "current_quantity", "threshold"},
			},
			{
				Name: "Status",
				Fields: []string{"status", "notification_sent"},
			},
			{
				Name: "Resolution",
				Fields: []string{"resolved_by_user_id", "resolved_by_user_name", "resolution_notes", "resolved_at"},
				Collapsed: true,
			},
			{
				Name: "Notification",
				Fields: []string{"notification_sent_at"},
				Collapsed: true,
			},
			{
				Name: "Timestamps",
				Fields: []string{"created_at", "updated_at", "acknowledged_at"},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "resolution_notes"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at", "product_variant_id", "warehouse_id", "current_quantity", "threshold"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[StockAlert], user interface{}, obj *StockAlert) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[StockAlert], user interface{}, obj *StockAlert) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[StockAlert], user interface{}, obj *StockAlert) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[StockAlert]{
			{
				Name:  "active_alerts",
				Label: "Active Alerts",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockAlert], value interface{}) orm.QuerySet[StockAlert] {
					return qs.Filter(orm.F("status").Eq("active"))
				},
			},
			{
				Name:  "low_stock",
				Label: "Low Stock Alerts",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockAlert], value interface{}) orm.QuerySet[StockAlert] {
					return qs.Filter(orm.F("alert_type").Eq("low_stock"))
				},
			},
			{
				Name:  "out_of_stock",
				Label: "Out of Stock Alerts",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockAlert], value interface{}) orm.QuerySet[StockAlert] {
					return qs.Filter(orm.F("alert_type").Eq("out_of_stock"))
				},
			},
			{
				Name:  "unresolved",
				Label: "Unresolved",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockAlert], value interface{}) orm.QuerySet[StockAlert] {
					return qs.Filter(orm.F("status").In("active", "acknowledged"))
				},
			},
			{
				Name:  "notification_sent",
				Label: "Notification Sent",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockAlert], value interface{}) orm.QuerySet[StockAlert] {
					return qs.Filter(orm.F("notification_sent").Eq(true))
				},
			},
		},
		// Custom actions
		Actions: []admin.Action[StockAlert]{
			{
				Name:  "resolve",
				Label: "Mark Resolved",
				Handler: func(ctx context.Context, instances []*StockAlert) error {
					for _, alert := range instances {
						alert.Status = "resolved"
						if err := StockAlertObjects.Update(ctx, alert); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "acknowledge",
				Label: "Mark Acknowledged",
				Handler: func(ctx context.Context, instances []*StockAlert) error {
					for _, alert := range instances {
						alert.Status = "acknowledged"
						if err := StockAlertObjects.Update(ctx, alert); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "dismiss",
				Label: "Dismiss Alert",
				Handler: func(ctx context.Context, instances []*StockAlert) error {
					for _, alert := range instances {
						alert.Status = "dismissed"
						if err := StockAlertObjects.Update(ctx, alert); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "send_notification",
				Label: "Send Notification",
				Handler: func(ctx context.Context, instances []*StockAlert) error {
					for _, alert := range instances {
						alert.NotificationSent = true
						if err := StockAlertObjects.Update(ctx, alert); err != nil {
							return err
						}
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
		SearchFields: []admin.Field{
			StockTransferFieldsInstance.TransferNumber,
			StockTransferFieldsInstance.TrackingNumber,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[StockTransfer]{
			{
				Name: "Transfer Details",
				Fields: []string{"transfer_number", "from_warehouse_id", "to_warehouse_id", "product_variant_id", "quantity"},
			},
			{
				Name: "Status Tracking",
				Fields: []string{"status", "tracking_number", "carrier"},
			},
			{
				Name: "Notes",
				Fields: []string{"notes", "reason"},
			},
			{
				Name: "User Tracking",
				Fields: []string{"requested_by_user_id", "requested_by_user_name", "approved_by_user_id", "approved_by_user_name"},
				Collapsed: true,
			},
			{
				Name: "Timestamps",
				Fields: []string{"created_at", "updated_at", "shipped_at", "completed_at", "cancelled_at"},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "quantity", "tracking_number"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at", "transfer_number", "requested_by_user_name", "approved_by_user_name"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[StockTransfer], user interface{}, obj *StockTransfer) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[StockTransfer], user interface{}, obj *StockTransfer) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[StockTransfer], user interface{}, obj *StockTransfer) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[StockTransfer]{
			{
				Name:  "pending",
				Label: "Pending Transfers",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockTransfer], value interface{}) orm.QuerySet[StockTransfer] {
					return qs.Filter(orm.F("status").Eq("pending"))
				},
			},
			{
				Name:  "in_transit",
				Label: "In Transit",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockTransfer], value interface{}) orm.QuerySet[StockTransfer] {
					return qs.Filter(orm.F("status").Eq("in_transit"))
				},
			},
			{
				Name:  "completed",
				Label: "Completed Transfers",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockTransfer], value interface{}) orm.QuerySet[StockTransfer] {
					return qs.Filter(orm.F("status").Eq("completed"))
				},
			},
			{
				Name:  "cancelled",
				Label: "Cancelled Transfers",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockTransfer], value interface{}) orm.QuerySet[StockTransfer] {
					return qs.Filter(orm.F("status").Eq("cancelled"))
				},
			},
			{
				Name:  "has_tracking",
				Label: "Has Tracking Number",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[StockTransfer], value interface{}) orm.QuerySet[StockTransfer] {
					return qs.Filter(orm.F("tracking_number").IsNotNull())
				},
			},
		},
		// Custom actions
		Actions: []admin.Action[StockTransfer]{
			{
				Name:  "complete",
				Label: "Mark Completed",
				Handler: func(ctx context.Context, instances []*StockTransfer) error {
					for _, transfer := range instances {
						transfer.Status = "completed"
						if err := StockTransferObjects.Update(ctx, transfer); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "ship",
				Label: "Mark Shipped",
				Handler: func(ctx context.Context, instances []*StockTransfer) error {
					for _, transfer := range instances {
						transfer.Status = "in_transit"
						if err := StockTransferObjects.Update(ctx, transfer); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "cancel",
				Label: "Cancel Transfer",
				Handler: func(ctx context.Context, instances []*StockTransfer) error {
					for _, transfer := range instances {
						transfer.Status = "cancelled"
						if err := StockTransferObjects.Update(ctx, transfer); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "approve",
				Label: "Approve Transfer",
				Handler: func(ctx context.Context, instances []*StockTransfer) error {
					for _, transfer := range instances {
						if transfer.Status == "pending" {
							transfer.Status = "approved"
							if err := StockTransferObjects.Update(ctx, transfer); err != nil {
								return err
							}
						}
					}
					return nil
				},
			},
		},
	})
}
