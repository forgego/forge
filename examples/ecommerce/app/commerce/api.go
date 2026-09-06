package commerce

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers commerce API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("shipping-methods", &api.ViewSetConfig{
		Model:      &ShippingMethod{},
		Queryset:   ShippingMethodObjects,
		Serializer: base,
	})

	router.Register("payment-methods", &api.ViewSetConfig{
		Model:      &PaymentMethod{},
		Queryset:   PaymentMethodObjects,
		Serializer: base,
	})

	router.Register("tax-rates", &api.ViewSetConfig{
		Model:      &TaxRate{},
		Queryset:   TaxRateObjects,
		Serializer: base,
	})

	router.Register("currencies", &api.ViewSetConfig{
		Model:      &Currency{},
		Queryset:   CurrencyObjects,
		Serializer: base,
	})

	router.Register("exchange-rates", &api.ViewSetConfig{
		Model:      &ExchangeRate{},
		Queryset:   ExchangeRateObjects,
		Serializer: base,
	})
}
