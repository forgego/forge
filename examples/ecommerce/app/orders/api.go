package orders

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers order API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("carts", &api.ViewSetConfig{
		Model:      &Cart{},
		Queryset:   CartObjects,
		Serializer: base,
	})

	router.Register("cart-items", &api.ViewSetConfig{
		Model:      &CartItem{},
		Queryset:   CartItemObjects,
		Serializer: base,
	})

	router.Register("orders", &api.ViewSetConfig{
		Model:      &Order{},
		Queryset:   OrderObjects,
		Serializer: base,
	})

	router.Register("order-items", &api.ViewSetConfig{
		Model:      &OrderItem{},
		Queryset:   OrderItemObjects,
		Serializer: base,
	})

	router.Register("payments", &api.ViewSetConfig{
		Model:      &Payment{},
		Queryset:   PaymentObjects,
		Serializer: base,
	})

	router.Register("shipments", &api.ViewSetConfig{
		Model:      &Shipment{},
		Queryset:   ShipmentObjects,
		Serializer: base,
	})
}
