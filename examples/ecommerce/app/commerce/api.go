package commerce

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers all commerce API endpoints
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// ShippingMethod ViewSet
	shippingMethodViewSet := api.NewModelViewSet(
		ShippingMethod{},
		database,
		api.WithFilterFields("is_active", "carrier_name"),
		api.WithSearchFields("name", "code", "carrier_name"),
		api.WithOrderingFields("sort_order", "name", "base_price"),
	)
	router.RegisterViewSet("shipping-methods", shippingMethodViewSet)

	// PaymentMethod ViewSet
	paymentMethodViewSet := api.NewModelViewSet(
		PaymentMethod{},
		database,
		api.WithFilterFields("is_active", "processor_name", "supports_refund"),
		api.WithSearchFields("name", "code", "processor_name"),
		api.WithOrderingFields("sort_order", "name"),
	)
	router.RegisterViewSet("payment-methods", paymentMethodViewSet)

	// TaxRate ViewSet
	taxRateViewSet := api.NewModelViewSet(
		TaxRate{},
		database,
		api.WithFilterFields("is_active", "country", "state", "is_compound"),
		api.WithSearchFields("name", "code", "country", "state"),
		api.WithOrderingFields("priority", "country", "state"),
	)
	router.RegisterViewSet("tax-rates", taxRateViewSet)

	// Currency ViewSet
	currencyViewSet := api.NewModelViewSet(
		Currency{},
		database,
		api.WithFilterFields("is_active", "is_default"),
		api.WithSearchFields("code", "name"),
		api.WithOrderingFields("code"),
	)
	router.RegisterViewSet("currencies", currencyViewSet)

	// ExchangeRate ViewSet
	exchangeRateViewSet := api.NewModelViewSet(
		ExchangeRate{},
		database,
		api.WithFilterFields("is_active", "from_currency_id", "to_currency_id"),
		api.WithSearchFields("source"),
		api.WithOrderingFields("-effective_date"),
	)
	router.RegisterViewSet("exchange-rates", exchangeRateViewSet)
}
