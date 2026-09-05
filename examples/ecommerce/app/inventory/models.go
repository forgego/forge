package inventory

import (
	"context"

	"github.com/forgego/forge/registry"
	"github.com/forgego/forge/schema"
)

// Warehouse represents a physical warehouse or storage location
type Warehouse struct {
	WarehouseGenerated
}

func (Warehouse) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),

		// Identification
		schema.StringField("name", schema.Required(), schema.MaxLength(200), schema.Unique()),
		schema.StringField("code", schema.Required(), schema.MaxLength(50), schema.Unique(),
			schema.HelpText("Warehouse code for reference")),

		// Contact
		schema.StringField("contact_name", schema.MaxLength(200), schema.Optional()),
		schema.StringField("contact_email", schema.MaxLength(255), schema.Optional()),
		schema.StringField("contact_phone", schema.MaxLength(20), schema.Optional()),

		// Address
		schema.StringField("address_line1", schema.Required(), schema.MaxLength(255)),
		schema.StringField("address_line2", schema.MaxLength(255), schema.Optional()),
		schema.StringField("city", schema.Required(), schema.MaxLength(100)),
		schema.StringField("state", schema.MaxLength(100), schema.Optional()),
		schema.StringField("postal_code", schema.Required(), schema.MaxLength(20)),
		schema.StringField("country_code", schema.Required(), schema.MaxLength(2),
			schema.VerboseName("Country Code")),
		schema.StringField("country_name", schema.Required(), schema.MaxLength(100)),

		// Geolocation
		schema.Float64Field("latitude", schema.Optional()),
		schema.Float64Field("longitude", schema.Optional()),

		// Settings
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.BoolField("is_primary", schema.Default(false),
			schema.VerboseName("Primary"),
			schema.HelpText("Primary warehouse for fulfillment")),
		schema.Int32Field("priority", schema.Default(0),
			schema.HelpText("Fulfillment priority (higher = preferred)")),

		// Capacity
		schema.Int32Field("total_capacity", schema.Optional(),
			schema.HelpText("Total storage capacity (units)")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Warehouse) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "warehouses",
		VerboseName:       "Warehouse",
		VerboseNamePlural: "Warehouses",
		OrderBy:           []string{"-is_primary", "-priority", "name"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_warehouse_code", "code"),
			schema.IndexOn("idx_warehouse_active", "is_active"),
			schema.IndexOn("idx_warehouse_primary", "is_primary"),
		},
	}
}

func (Warehouse) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Warehouse) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Ensure only one primary warehouse
			return nil
		},
	}
}

// Stock represents inventory levels for product variants
type Stock struct {
	StockGenerated
}

func (Stock) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("product_variant_id", schema.Required(),
			schema.VerboseName("Product Variant")),
		schema.Int64Field("warehouse_id", schema.Required(),
			schema.VerboseName("Warehouse")),

		// Quantities
		schema.Int32Field("quantity", schema.Required(), schema.Default(0),
			schema.HelpText("Available quantity")),
		schema.Int32Field("reserved_quantity", schema.Default(0),
			schema.VerboseName("Reserved Quantity"),
			schema.HelpText("Quantity reserved in pending orders")),
		schema.Int32Field("available_quantity", schema.Default(0),
			schema.HelpText("quantity - reserved_quantity")),

		// Reordering
		schema.Int32Field("reorder_point", schema.Default(10),
			schema.HelpText("Minimum quantity before reorder")),
		schema.Int32Field("reorder_quantity", schema.Default(50),
			schema.HelpText("Quantity to reorder")),

		// Location
		schema.StringField("bin_location", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("Physical location in warehouse (e.g., A-12-3)")),

		// Status
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.BoolField("allow_backorder", schema.Default(false)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
		schema.TimeField("last_counted_at", schema.Optional(),
			schema.HelpText("Last physical count")),
	}
}

func (Stock) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "stock",
		VerboseName:       "Stock",
		VerboseNamePlural: "Stock",
		OrderBy:           []string{"warehouse_id", "product_variant_id"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_stock_variant", "product_variant_id"),
			schema.IndexOn("idx_stock_warehouse", "warehouse_id"),
			schema.IndexOn("idx_stock_quantity", "quantity"),
		},
		UniqueTogether: [][]string{
			{"product_variant_id", "warehouse_id"},
		},
	}
}

func (Stock) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("product_variant_id", "ProductVariant",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("stock_records")),
		schema.ForeignKeyField("warehouse_id", "Warehouse",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("stock_records")),
	}
}

func (Stock) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Calculate available_quantity
			// Check for low stock alert
			return nil
		},
		AfterSave: func(ctx context.Context, instance interface{}) error {
			// Update product variant total stock
			// Trigger low stock alert if needed
			return nil
		},
	}
}

// StockMovement represents inventory transactions
type StockMovement struct {
	StockMovementGenerated
}

func (StockMovement) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("stock_id", schema.Required()),
		schema.Int64Field("product_variant_id", schema.Required(),
			schema.VerboseName("Product Variant")),
		schema.Int64Field("warehouse_id", schema.Required(),
			schema.VerboseName("Warehouse")),

		// Movement details
		schema.StringField("type", schema.Required(), schema.MaxLength(50),
			schema.VerboseName("Movement Type"),
			schema.HelpText("Type: purchase, sale, return, adjustment, transfer, damage, loss, count")),
		schema.Int32Field("quantity", schema.Required(),
			schema.HelpText("Quantity change (positive or negative)")),
		schema.Int32Field("quantity_before", schema.Required(),
			schema.HelpText("Quantity before movement")),
		schema.Int32Field("quantity_after", schema.Required(),
			schema.HelpText("Quantity after movement")),

		// Reference
		schema.StringField("reference_type", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Reference type: order, return, transfer, etc.")),
		schema.Int64Field("reference_id", schema.Optional(),
			schema.HelpText("ID of related record (order_id, transfer_id, etc.)")),
		schema.StringField("reference_number", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("Human-readable reference number")),

		// Transfer (if applicable)
		schema.Int64Field("from_warehouse_id", schema.Optional(),
			schema.HelpText("Source warehouse for transfers")),
		schema.Int64Field("to_warehouse_id", schema.Optional(),
			schema.HelpText("Destination warehouse for transfers")),

		// Additional information
		schema.StringField("reason", schema.MaxLength(255), schema.Optional(),
			schema.HelpText("Reason for movement")),
		schema.TextField("notes", schema.Optional()),

		// User tracking
		schema.Int64Field("user_id", schema.Optional(),
			schema.HelpText("User who performed the movement")),
		schema.StringField("user_name", schema.MaxLength(200), schema.Optional()),

		// Cost (for accounting)
		schema.Float64Field("unit_cost", schema.Optional(),
			schema.HelpText("Cost per unit at time of movement")),
		schema.Float64Field("total_cost", schema.Optional(),
			schema.HelpText("Total cost of movement")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("movement_date", schema.Required(),
			schema.VerboseName("Movement Date"),
			schema.HelpText("Date of movement (can be backdated)")),
	}
}

func (StockMovement) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "stock_movements",
		VerboseName:       "Stock Movement",
		VerboseNamePlural: "Stock Movements",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_movement_stock", "stock_id"),
			schema.IndexOn("idx_movement_variant", "product_variant_id"),
			schema.IndexOn("idx_movement_warehouse", "warehouse_id"),
			schema.IndexOn("idx_movement_type", "type"),
			schema.IndexOn("idx_movement_reference", "reference_type", "reference_id"),
			schema.IndexOn("idx_movement_date", "movement_date"),
		},
	}
}

func (StockMovement) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("stock_id", "Stock",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("movements")),
		schema.ForeignKeyField("product_variant_id", "ProductVariant",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("stock_movements")),
		schema.ForeignKeyField("warehouse_id", "Warehouse",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("stock_movements")),
		schema.ForeignKeyField("from_warehouse_id", "Warehouse",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("outbound_transfers")),
		schema.ForeignKeyField("to_warehouse_id", "Warehouse",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("inbound_transfers")),
	}
}

func (StockMovement) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeCreate: func(ctx context.Context, instance interface{}) error {
			// Capture quantity_before
			// Calculate quantity_after
			// Calculate total_cost
			return nil
		},
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// Update stock quantity
			// Create alert if low stock
			return nil
		},
	}
}

// StockAlert represents low stock alerts
type StockAlert struct {
	StockAlertGenerated
}

func (StockAlert) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("stock_id", schema.Required()),
		schema.Int64Field("product_variant_id", schema.Required(),
			schema.VerboseName("Product Variant")),
		schema.Int64Field("warehouse_id", schema.Required(),
			schema.VerboseName("Warehouse")),

		// Alert details
		schema.StringField("alert_type", schema.Required(), schema.MaxLength(50),
			schema.VerboseName("Alert Type"),
			schema.HelpText("Type: low_stock, out_of_stock, overstock")),
		schema.Int32Field("current_quantity", schema.Required(),
			schema.VerboseName("Current Quantity")),
		schema.Int32Field("threshold", schema.Required(),
			schema.HelpText("Threshold that triggered alert")),

		// Status
		schema.StringField("status", schema.Required(), schema.MaxLength(20), schema.Default("active"),
			schema.VerboseName("Status"),
			schema.HelpText("Status: active, acknowledged, resolved, dismissed")),

		// Resolution
		schema.Int64Field("resolved_by_user_id", schema.Optional()),
		schema.StringField("resolved_by_user_name", schema.MaxLength(200), schema.Optional()),
		schema.TextField("resolution_notes", schema.Optional()),

		// Notification
		schema.BoolField("notification_sent", schema.Default(false)),
		schema.TimeField("notification_sent_at", schema.Optional()),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
		schema.TimeField("acknowledged_at", schema.Optional()),
		schema.TimeField("resolved_at", schema.Optional()),
	}
}

func (StockAlert) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "stock_alerts",
		VerboseName:       "Stock Alert",
		VerboseNamePlural: "Stock Alerts",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_alert_stock", "stock_id"),
			schema.IndexOn("idx_alert_variant", "product_variant_id"),
			schema.IndexOn("idx_alert_warehouse", "warehouse_id"),
			schema.IndexOn("idx_alert_status", "status"),
			schema.IndexOn("idx_alert_type", "alert_type"),
		},
	}
}

func (StockAlert) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("stock_id", "Stock",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("alerts")),
		schema.ForeignKeyField("product_variant_id", "ProductVariant",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("stock_alerts")),
		schema.ForeignKeyField("warehouse_id", "Warehouse",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("stock_alerts")),
	}
}

func (StockAlert) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// Send notification email
			return nil
		},
	}
}

// StockTransfer represents transfers between warehouses
type StockTransfer struct {
	StockTransferGenerated
}

func (StockTransfer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),

		// Transfer identification
		schema.StringField("transfer_number", schema.Required(), schema.MaxLength(50), schema.Unique(),
			schema.VerboseName("Transfer Number")),

		// Warehouses
		schema.Int64Field("from_warehouse_id", schema.Required(),
			schema.VerboseName("From Warehouse")),
		schema.Int64Field("to_warehouse_id", schema.Required(),
			schema.VerboseName("To Warehouse")),

		// Product
		schema.Int64Field("product_variant_id", schema.Required(),
			schema.VerboseName("Product Variant")),
		schema.Int32Field("quantity", schema.Required()),

		// Status
		schema.StringField("status", schema.Required(), schema.MaxLength(20), schema.Default("pending"),
			schema.VerboseName("Status"),
			schema.HelpText("Status: pending, in_transit, completed, cancelled")),

		// Tracking
		schema.StringField("tracking_number", schema.MaxLength(255), schema.Optional()),
		schema.StringField("carrier", schema.MaxLength(100), schema.Optional()),

		// Notes
		schema.TextField("notes", schema.Optional()),
		schema.TextField("reason", schema.Optional()),

		// User tracking
		schema.Int64Field("requested_by_user_id", schema.Optional()),
		schema.StringField("requested_by_user_name", schema.MaxLength(200), schema.Optional()),
		schema.Int64Field("approved_by_user_id", schema.Optional()),
		schema.StringField("approved_by_user_name", schema.MaxLength(200), schema.Optional()),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
		schema.TimeField("shipped_at", schema.Optional()),
		schema.TimeField("completed_at", schema.Optional()),
		schema.TimeField("cancelled_at", schema.Optional()),
	}
}

func (StockTransfer) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "stock_transfers",
		VerboseName:       "Stock Transfer",
		VerboseNamePlural: "Stock Transfers",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_transfer_number", "transfer_number"),
			schema.IndexOn("idx_transfer_from", "from_warehouse_id"),
			schema.IndexOn("idx_transfer_to", "to_warehouse_id"),
			schema.IndexOn("idx_transfer_variant", "product_variant_id"),
			schema.IndexOn("idx_transfer_status", "status"),
		},
	}
}

func (StockTransfer) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("from_warehouse_id", "Warehouse",
			schema.OnDelete(schema.CascadePROTECT),
			schema.RelatedName("outbound_stock_transfers")),
		schema.ForeignKeyField("to_warehouse_id", "Warehouse",
			schema.OnDelete(schema.CascadePROTECT),
			schema.RelatedName("inbound_stock_transfers")),
		schema.ForeignKeyField("product_variant_id", "ProductVariant",
			schema.OnDelete(schema.CascadePROTECT),
			schema.RelatedName("transfers")),
	}
}

func (StockTransfer) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeCreate: func(ctx context.Context, instance interface{}) error {
			// Generate transfer number
			// Validate sufficient stock in source warehouse
			return nil
		},
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// Create stock movements for both warehouses
			return nil
		},
		AfterUpdate: func(ctx context.Context, instance interface{}) error {
			// Handle status changes
			// Update stock movements
			return nil
		},
	}
}

// RegisterModels registers inventory models with the framework
func RegisterModels() {
	registry.RegisterModel(&Warehouse{})
	registry.RegisterModel(&Stock{})
	registry.RegisterModel(&StockMovement{})
	registry.RegisterModel(&StockAlert{})
	registry.RegisterModel(&StockTransfer{})
}
