package inventory

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers inventory API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("warehouses", &api.ViewSetConfig{
		Model:      &Warehouse{},
		Queryset:   WarehouseObjects,
		Serializer: base,
	})

	router.Register("stock", &api.ViewSetConfig{
		Model:      &Stock{},
		Queryset:   StockObjects,
		Serializer: base,
	})

	router.Register("stock-movements", &api.ViewSetConfig{
		Model:      &StockMovement{},
		Queryset:   StockMovementObjects,
		Serializer: base,
	})

	router.Register("stock-alerts", &api.ViewSetConfig{
		Model:      &StockAlert{},
		Queryset:   StockAlertObjects,
		Serializer: base,
	})

	router.Register("stock-transfers", &api.ViewSetConfig{
		Model:      &StockTransfer{},
		Queryset:   StockTransferObjects,
		Serializer: base,
	})
}
