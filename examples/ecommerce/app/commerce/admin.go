package commerce

import (
	"context"

	"github.com/forgego/forge/admin"
	adminCore "github.com/forgego/forge/admin/core"
)

// RegisterAdmin registers all commerce models with the admin interface
func RegisterAdmin(ctx context.Context) {
	site := admin.DefaultSite

	// ShippingMethod Admin
	shippingMethodAdmin := adminCore.NewModelAdmin(
		ShippingMethod{},
		adminCore.WithListDisplay("id", "name", "code", "base_price", "carrier_name", "is_active", "sort_order"),
		adminCore.WithSearchFields("name", "code", "carrier_name"),
		adminCore.WithListFilter("is_active"),
		adminCore.WithOrdering("sort_order", "name"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, shippingMethodAdmin)

	// PaymentMethod Admin
	paymentMethodAdmin := adminCore.NewModelAdmin(
		PaymentMethod{},
		adminCore.WithListDisplay("id", "name", "code", "processor_name", "is_active", "supports_refund", "sort_order"),
		adminCore.WithSearchFields("name", "code", "processor_name"),
		adminCore.WithListFilter("is_active", "supports_refund", "requires_auth"),
		adminCore.WithOrdering("sort_order", "name"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, paymentMethodAdmin)

	// TaxRate Admin
	taxRateAdmin := adminCore.NewModelAdmin(
		TaxRate{},
		adminCore.WithListDisplay("id", "name", "code", "rate", "country", "state", "is_active", "priority"),
		adminCore.WithSearchFields("name", "code", "country", "state"),
		adminCore.WithListFilter("is_active", "country", "is_compound", "apply_to_shipping"),
		adminCore.WithOrdering("priority", "country", "state"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, taxRateAdmin)

	// Currency Admin
	currencyAdmin := adminCore.NewModelAdmin(
		Currency{},
		adminCore.WithListDisplay("id", "code", "name", "symbol", "decimal_places", "is_active", "is_default"),
		adminCore.WithSearchFields("code", "name"),
		adminCore.WithListFilter("is_active", "is_default"),
		adminCore.WithOrdering("code"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, currencyAdmin)

	// ExchangeRate Admin
	exchangeRateAdmin := adminCore.NewModelAdmin(
		ExchangeRate{},
		adminCore.WithListDisplay("id", "from_currency_id", "to_currency_id", "rate", "effective_date", "source", "is_active"),
		adminCore.WithSearchFields("source"),
		adminCore.WithListFilter("is_active", "effective_date"),
		adminCore.WithOrdering("-effective_date"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, exchangeRateAdmin)
}
