package commerce

import (
	"context"

	"github.com/forgego/forge/admin"
)

// RegisterAdmin registers all commerce models with the admin interface
func RegisterAdmin(ctx context.Context) error {
	// ShippingMethod Admin
	_, err := admin.Register(&admin.Config[ShippingMethod]{
		ListDisplay: []admin.Field{
			ShippingMethodFieldsInstance.Id,
			ShippingMethodFieldsInstance.Name,
			ShippingMethodFieldsInstance.Code,
			ShippingMethodFieldsInstance.BasePrice,
			ShippingMethodFieldsInstance.CarrierName,
			ShippingMethodFieldsInstance.IsActive,
			ShippingMethodFieldsInstance.SortOrder,
		},
		SearchFields: []admin.Field{
			ShippingMethodFieldsInstance.Name,
			ShippingMethodFieldsInstance.Code,
			ShippingMethodFieldsInstance.CarrierName,
		},
		ListFilter: []admin.Field{
			ShippingMethodFieldsInstance.IsActive,
		},
		Ordering: []admin.Field{
			ShippingMethodFieldsInstance.SortOrder,
			ShippingMethodFieldsInstance.Name,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// PaymentMethod Admin
	_, err = admin.Register(&admin.Config[PaymentMethod]{
		ListDisplay: []admin.Field{
			PaymentMethodFieldsInstance.Id,
			PaymentMethodFieldsInstance.Name,
			PaymentMethodFieldsInstance.Code,
			PaymentMethodFieldsInstance.ProcessorName,
			PaymentMethodFieldsInstance.IsActive,
			PaymentMethodFieldsInstance.SupportsRefund,
			PaymentMethodFieldsInstance.SortOrder,
		},
		SearchFields: []admin.Field{
			PaymentMethodFieldsInstance.Name,
			PaymentMethodFieldsInstance.Code,
			PaymentMethodFieldsInstance.ProcessorName,
		},
		ListFilter: []admin.Field{
			PaymentMethodFieldsInstance.IsActive,
			PaymentMethodFieldsInstance.SupportsRefund,
			PaymentMethodFieldsInstance.RequiresAuth,
		},
		Ordering: []admin.Field{
			PaymentMethodFieldsInstance.SortOrder,
			PaymentMethodFieldsInstance.Name,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// TaxRate Admin
	_, err = admin.Register(&admin.Config[TaxRate]{
		ListDisplay: []admin.Field{
			TaxRateFieldsInstance.Id,
			TaxRateFieldsInstance.Name,
			TaxRateFieldsInstance.Code,
			TaxRateFieldsInstance.Rate,
			TaxRateFieldsInstance.Country,
			TaxRateFieldsInstance.State,
			TaxRateFieldsInstance.IsActive,
			TaxRateFieldsInstance.Priority,
		},
		SearchFields: []admin.Field{
			TaxRateFieldsInstance.Name,
			TaxRateFieldsInstance.Code,
			TaxRateFieldsInstance.Country,
			TaxRateFieldsInstance.State,
		},
		ListFilter: []admin.Field{
			TaxRateFieldsInstance.IsActive,
			TaxRateFieldsInstance.Country,
			TaxRateFieldsInstance.IsCompound,
			TaxRateFieldsInstance.ApplyToShipping,
		},
		Ordering: []admin.Field{
			TaxRateFieldsInstance.Priority,
			TaxRateFieldsInstance.Country,
			TaxRateFieldsInstance.State,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// Currency Admin
	_, err = admin.Register(&admin.Config[Currency]{
		ListDisplay: []admin.Field{
			CurrencyFieldsInstance.Id,
			CurrencyFieldsInstance.Code,
			CurrencyFieldsInstance.Name,
			CurrencyFieldsInstance.Symbol,
			CurrencyFieldsInstance.DecimalPlaces,
			CurrencyFieldsInstance.IsActive,
			CurrencyFieldsInstance.IsDefault,
		},
		SearchFields: []admin.Field{
			CurrencyFieldsInstance.Code,
			CurrencyFieldsInstance.Name,
		},
		ListFilter: []admin.Field{
			CurrencyFieldsInstance.IsActive,
			CurrencyFieldsInstance.IsDefault,
		},
		Ordering: []admin.Field{
			CurrencyFieldsInstance.Code,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// ExchangeRate Admin
	_, err = admin.Register(&admin.Config[ExchangeRate]{
		ListDisplay: []admin.Field{
			ExchangeRateFieldsInstance.Id,
			ExchangeRateFieldsInstance.FromCurrencyId,
			ExchangeRateFieldsInstance.ToCurrencyId,
			ExchangeRateFieldsInstance.Rate,
			ExchangeRateFieldsInstance.EffectiveDate,
			ExchangeRateFieldsInstance.Source,
			ExchangeRateFieldsInstance.IsActive,
		},
		SearchFields: []admin.Field{
			ExchangeRateFieldsInstance.Source,
		},
		ListFilter: []admin.Field{
			ExchangeRateFieldsInstance.IsActive,
			ExchangeRateFieldsInstance.EffectiveDate,
		},
		Ordering: []admin.Field{
			ExchangeRateFieldsInstance.EffectiveDate,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	return nil
}
