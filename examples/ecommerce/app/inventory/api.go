package inventory

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers inventory API endpoints
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// Warehouse API
	router.Register("warehouses", &api.ViewSetConfig{
		Model:        &Warehouse{},
		Queryset:     WarehouseObjects,
		Serializer:   &WarehouseSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "name", "code", "city", "country_code", "is_active", "is_primary", "priority"},
		DetailFields: []string{"id", "name", "code", "contact_name", "contact_email", "contact_phone", "address_line1", "address_line2", "city", "state", "postal_code", "country_code", "country_name", "latitude", "longitude", "is_active", "is_primary", "priority", "total_capacity", "created_at", "updated_at"},
		Filterable:   []string{"is_active", "is_primary", "country_code"},
		Searchable:   []string{"name", "code", "city"},
		Ordering:     []string{"-is_primary", "-priority", "name"},
		PerPage:      20,
	})

	// Stock API
	router.Register("stock", &api.ViewSetConfig{
		Model:        &Stock{},
		Queryset:     StockObjects,
		Serializer:   &StockSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "product_variant_id", "warehouse_id", "quantity", "reserved_quantity", "available_quantity", "reorder_point", "is_active"},
		DetailFields: []string{"id", "product_variant_id", "warehouse_id", "quantity", "reserved_quantity", "available_quantity", "reorder_point", "reorder_quantity", "bin_location", "is_active", "allow_backorder", "created_at", "updated_at", "last_counted_at"},
		Filterable:   []string{"product_variant_id", "warehouse_id", "is_active"},
		Searchable:   []string{"bin_location"},
		Ordering:     []string{"warehouse_id", "quantity"},
		PerPage:      20,
	})

	// StockMovement API
	router.Register("stock-movements", &api.ViewSetConfig{
		Model:        &StockMovement{},
		Queryset:     StockMovementObjects,
		Serializer:   &StockMovementSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "product_variant_id", "warehouse_id", "type", "quantity", "quantity_before", "quantity_after", "reference_number", "movement_date"},
		DetailFields: []string{"id", "stock_id", "product_variant_id", "warehouse_id", "type", "quantity", "quantity_before", "quantity_after", "reference_type", "reference_id", "reference_number", "from_warehouse_id", "to_warehouse_id", "reason", "notes", "user_id", "user_name", "unit_cost", "total_cost", "created_at", "movement_date"},
		Filterable:   []string{"product_variant_id", "warehouse_id", "type", "movement_date"},
		Searchable:   []string{"reference_number", "reason", "notes"},
		Ordering:     []string{"-movement_date", "-created_at"},
		PerPage:      20,
	})

	// StockAlert API
	router.Register("stock-alerts", &api.ViewSetConfig{
		Model:        &StockAlert{},
		Queryset:     StockAlertObjects,
		Serializer:   &StockAlertSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "product_variant_id", "warehouse_id", "alert_type", "current_quantity", "threshold", "status", "created_at"},
		DetailFields: []string{"id", "stock_id", "product_variant_id", "warehouse_id", "alert_type", "current_quantity", "threshold", "status", "resolved_by_user_id", "resolved_by_user_name", "resolution_notes", "notification_sent", "notification_sent_at", "created_at", "updated_at", "acknowledged_at", "resolved_at"},
		Filterable:   []string{"product_variant_id", "warehouse_id", "alert_type", "status"},
		Searchable:   []string{"resolution_notes"},
		Ordering:     []string{"-created_at", "status"},
		PerPage:      20,
	})

	// StockTransfer API
	router.Register("stock-transfers", &api.ViewSetConfig{
		Model:        &StockTransfer{},
		Queryset:     StockTransferObjects,
		Serializer:   &StockTransferSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "transfer_number", "from_warehouse_id", "to_warehouse_id", "product_variant_id", "quantity", "status", "created_at"},
		DetailFields: []string{"id", "transfer_number", "from_warehouse_id", "to_warehouse_id", "product_variant_id", "quantity", "status", "tracking_number", "carrier", "notes", "reason", "requested_by_user_id", "requested_by_user_name", "approved_by_user_id", "approved_by_user_name", "created_at", "updated_at", "shipped_at", "completed_at", "cancelled_at"},
		Filterable:   []string{"from_warehouse_id", "to_warehouse_id", "product_variant_id", "status"},
		Searchable:   []string{"transfer_number", "tracking_number"},
		Ordering:     []string{"-created_at", "status"},
		PerPage:      20,
	})
}

// Serializers

type WarehouseSerializer struct {
	*api.BaseSerializer
}

func (s *WarehouseSerializer) New() api.Serializer {
	return &WarehouseSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type StockSerializer struct {
	*api.BaseSerializer
}

func (s *StockSerializer) New() api.Serializer {
	return &StockSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type StockMovementSerializer struct {
	*api.BaseSerializer
}

func (s *StockMovementSerializer) New() api.Serializer {
	return &StockMovementSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type StockAlertSerializer struct {
	*api.BaseSerializer
}

func (s *StockAlertSerializer) New() api.Serializer {
	return &StockAlertSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type StockTransferSerializer struct {
	*api.BaseSerializer
}

func (s *StockTransferSerializer) New() api.Serializer {
	return &StockTransferSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}
