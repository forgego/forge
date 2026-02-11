package commerce

import (
	"context"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers commerce models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// ShippingMethod admin
	admin.Register(&admin.Config[ShippingMethod]{
		Icon: "Truck",
		ListDisplay: []admin.Field{
			ShippingMethodFieldsInstance.Name,
			ShippingMethodFieldsInstance.Code,
			ShippingMethodFieldsInstance.Carrier,
			ShippingMethodFieldsInstance.ServiceLevel,
			ShippingMethodFieldsInstance.BasePrice,
			ShippingMethodFieldsInstance.IsActive,
			ShippingMethodFieldsInstance.IsDefault,
			ShippingMethodFieldsInstance.Priority,
		},
		ListFilter: []admin.Field{
			ShippingMethodFieldsInstance.IsActive,
			ShippingMethodFieldsInstance.Carrier,
			ShippingMethodFieldsInstance.IsDefault,
		},
		SearchFields: []admin.Field{
			ShippingMethodFieldsInstance.Name,
			ShippingMethodFieldsInstance.Code,
			ShippingMethodFieldsInstance.Carrier,
			ShippingMethodFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			ShippingMethodFieldsInstance.Priority,
			ShippingMethodFieldsInstance.Name,
		},
		// Fieldsets for grouping form fields
		Fieldsets: []admin.Fieldset[ShippingMethod]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "code", "description", "carrier", "service_level"},
			},
			{
				Name: "Pricing",
				Fields: []string{"base_price", "handling_fee", "free_shipping_threshold"},
			},
			{
				Name: "Weight Constraints",
				Fields: []string{"min_weight", "max_weight", "weight_unit"},
			},
			{
				Name: "Dimensions",
				Fields: []string{"min_length", "max_length", "min_width", "max_width", "min_height", "max_height", "dimension_unit"},
			},
			{
				Name: "Delivery",
				Fields: []string{"min_days", "max_days"},
			},
			{
				Name: "Geographic Restrictions",
				Fields: []string{"countries", "excluded_countries"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active", "is_default", "priority"},
			},
		},
		// History manager for audit logging
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "code", "base_price", "is_active", "is_default"},
		},
		// Read-only fields
		ReadOnlyFields: []admin.Field{
			ShippingMethodFieldsInstance.CreatedAt,
			ShippingMethodFieldsInstance.UpdatedAt,
		},
		// Custom permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}, obj *ShippingMethod) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}, obj *ShippingMethod) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}, obj *ShippingMethod) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}, obj *ShippingMethod) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[ShippingMethod]{
			{
				Name:  "active_methods",
				Label: "Active Methods Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ShippingMethod], value interface{}) orm.QuerySet[ShippingMethod] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "default_method",
				Label: "Default Method",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ShippingMethod], value interface{}) orm.QuerySet[ShippingMethod] {
					return qs.Filter("is_default", true)
				},
			},
			{
				Name:  "by_carrier",
				Label: "By Carrier",
				Type:  admin.FilterTypeString,
				Handler: func(ctx context.Context, qs orm.QuerySet[ShippingMethod], value interface{}) orm.QuerySet[ShippingMethod] {
					return qs.Filter("carrier", value)
				},
			},
		},
		// Custom bulk actions
		Actions: []admin.Action[ShippingMethod]{
			{
				Name:         "activate",
				Label:        "Activate Methods",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected shipping methods?",
				Handler: func(ctx context.Context, admin *admin.Admin[ShippingMethod], ids []interface{}) error {
					for _, id := range ids {
						method, err := ShippingMethodObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						method.IsActive = true
						if err := ShippingMethodObjects.Update(ctx, method); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Methods",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected shipping methods?",
				Handler: func(ctx context.Context, admin *admin.Admin[ShippingMethod], ids []interface{}) error {
					for _, id := range ids {
						method, err := ShippingMethodObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						method.IsActive = false
						if err := ShippingMethodObjects.Update(ctx, method); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "set_default",
				Label:        "Set as Default",
				Icon:         "Star",
				Confirmation: "Set selected shipping method as the default?",
				Handler: func(ctx context.Context, admin *admin.Admin[ShippingMethod], ids []interface{}) error {
					for _, id := range ids {
						method, err := ShippingMethodObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						method.IsDefault = true
						if err := ShippingMethodObjects.Update(ctx, method); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}) bool {
					return true
				},
			},
		},
	})

	// PaymentMethod admin
	admin.Register(&admin.Config[PaymentMethod]{
		Icon: "CreditCard",
		ListDisplay: []admin.Field{
			PaymentMethodFieldsInstance.Name,
			PaymentMethodFieldsInstance.Code,
			PaymentMethodFieldsInstance.Type,
			PaymentMethodFieldsInstance.Gateway,
			PaymentMethodFieldsInstance.FixedFee,
			PaymentMethodFieldsInstance.PercentageFee,
			PaymentMethodFieldsInstance.IsActive,
			PaymentMethodFieldsInstance.IsDefault,
		},
		ListFilter: []admin.Field{
			PaymentMethodFieldsInstance.IsActive,
			PaymentMethodFieldsInstance.Type,
			PaymentMethodFieldsInstance.Gateway,
			PaymentMethodFieldsInstance.IsDefault,
		},
		SearchFields: []admin.Field{
			PaymentMethodFieldsInstance.Name,
			PaymentMethodFieldsInstance.Code,
			PaymentMethodFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			PaymentMethodFieldsInstance.DisplayOrder,
			PaymentMethodFieldsInstance.Name,
		},
		Fieldsets: []admin.Fieldset[PaymentMethod]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "code", "type", "description"},
			},
			{
				Name: "Gateway Configuration",
				Fields: []string{"gateway", "gateway_config"},
			},
			{
				Name: "Fees",
				Fields: []string{"fixed_fee", "percentage_fee"},
			},
			{
				Name: "Limits",
				Fields: []string{"min_amount", "max_amount"},
			},
			{
				Name: "Currencies",
				Fields: []string{"currencies"},
			},
			{
				Name: "UI Settings",
				Fields: []string{"icon", "display_order"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active", "is_default", "test_mode"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "code", "is_active", "is_default", "fixed_fee", "percentage_fee"},
		},
		ReadOnlyFields: []admin.Field{
			PaymentMethodFieldsInstance.CreatedAt,
			PaymentMethodFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}, obj *PaymentMethod) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}, obj *PaymentMethod) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}, obj *PaymentMethod) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}, obj *PaymentMethod) bool {
			return true
		},
		Filters: []admin.Filter[PaymentMethod]{
			{
				Name:  "active_methods",
				Label: "Active Methods Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[PaymentMethod], value interface{}) orm.QuerySet[PaymentMethod] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "by_type",
				Label: "By Type",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "credit_card", Label: "Credit Card"},
					{Value: "debit_card", Label: "Debit Card"},
					{Value: "paypal", Label: "PayPal"},
					{Value: "bank_transfer", Label: "Bank Transfer"},
					{Value: "cash_on_delivery", Label: "Cash on Delivery"},
					{Value: "crypto", Label: "Cryptocurrency"},
					{Value: "wallet", Label: "Digital Wallet"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[PaymentMethod], value interface{}) orm.QuerySet[PaymentMethod] {
					return qs.Filter("type", value)
				},
			},
		},
		Actions: []admin.Action[PaymentMethod]{
			{
				Name:         "activate",
				Label:        "Activate Methods",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected payment methods?",
				Handler: func(ctx context.Context, admin *admin.Admin[PaymentMethod], ids []interface{}) error {
					for _, id := range ids {
						method, err := PaymentMethodObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						method.IsActive = true
						if err := PaymentMethodObjects.Update(ctx, method); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Methods",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected payment methods?",
				Handler: func(ctx context.Context, admin *admin.Admin[PaymentMethod], ids []interface{}) error {
					for _, id := range ids {
						method, err := PaymentMethodObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						method.IsActive = false
						if err := PaymentMethodObjects.Update(ctx, method); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}) bool {
					return true
				},
			},
		},
	})

	// TaxRate admin
	admin.Register(&admin.Config[TaxRate]{
		Icon: "Calculator",
		ListDisplay: []admin.Field{
			TaxRateFieldsInstance.Name,
			TaxRateFieldsInstance.Code,
			TaxRateFieldsInstance.Rate,
			TaxRateFieldsInstance.Country,
			TaxRateFieldsInstance.State,
			TaxRateFieldsInstance.TaxType,
			TaxRateFieldsInstance.IsActive,
			TaxRateFieldsInstance.Priority,
		},
		ListFilter: []admin.Field{
			TaxRateFieldsInstance.IsActive,
			TaxRateFieldsInstance.Country,
			TaxRateFieldsInstance.TaxType,
			TaxRateFieldsInstance.IsCompound,
		},
		SearchFields: []admin.Field{
			TaxRateFieldsInstance.Name,
			TaxRateFieldsInstance.Code,
			TaxRateFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			TaxRateFieldsInstance.Country,
			TaxRateFieldsInstance.State,
			TaxRateFieldsInstance.Priority,
		},
		Fieldsets: []admin.Fieldset[TaxRate]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "code", "description"},
			},
			{
				Name: "Rate",
				Fields: []string{"rate", "tax_type"},
			},
			{
				Name: "Geographic Scope",
				Fields: []string{"country", "state", "zip_pattern", "city"},
			},
			{
				Name: "Applicability",
				Fields: []string{"applies_to_products", "applies_to_shipping", "applies_to_services"},
			},
			{
				Name: "Date Range",
				Fields: []string{"start_date", "end_date"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active", "is_compound", "priority"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "code", "rate", "is_active", "is_compound"},
		},
		ReadOnlyFields: []admin.Field{
			TaxRateFieldsInstance.CreatedAt,
			TaxRateFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}, obj *TaxRate) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}, obj *TaxRate) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}, obj *TaxRate) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}, obj *TaxRate) bool {
			return true
		},
		Filters: []admin.Filter[TaxRate]{
			{
				Name:  "active_rates",
				Label: "Active Rates Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[TaxRate], value interface{}) orm.QuerySet[TaxRate] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "compound_rates",
				Label: "Compound Rates",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[TaxRate], value interface{}) orm.QuerySet[TaxRate] {
					return qs.Filter("is_compound", true)
				},
			},
			{
				Name:  "by_country",
				Label: "By Country",
				Type:  admin.FilterTypeString,
				Handler: func(ctx context.Context, qs orm.QuerySet[TaxRate], value interface{}) orm.QuerySet[TaxRate] {
					return qs.Filter("country", value)
				},
			},
		},
		Actions: []admin.Action[TaxRate]{
			{
				Name:         "activate",
				Label:        "Activate Rates",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected tax rates?",
				Handler: func(ctx context.Context, admin *admin.Admin[TaxRate], ids []interface{}) error {
					for _, id := range ids {
						rate, err := TaxRateObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						rate.IsActive = true
						if err := TaxRateObjects.Update(ctx, rate); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Rates",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected tax rates?",
				Handler: func(ctx context.Context, admin *admin.Admin[TaxRate], ids []interface{}) error {
					for _, id := range ids {
						rate, err := TaxRateObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						rate.IsActive = false
						if err := TaxRateObjects.Update(ctx, rate); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}) bool {
					return true
				},
			},
		},
	})

	// Currency admin
	admin.Register(&admin.Config[Currency]{
		Icon: "DollarSign",
		ListDisplay: []admin.Field{
			CurrencyFieldsInstance.Code,
			CurrencyFieldsInstance.Name,
			CurrencyFieldsInstance.Symbol,
			CurrencyFieldsInstance.DecimalPlaces,
			CurrencyFieldsInstance.IsBaseCurrency,
			CurrencyFieldsInstance.ExchangeRate,
			CurrencyFieldsInstance.IsActive,
			CurrencyFieldsInstance.IsDefault,
		},
		ListFilter: []admin.Field{
			CurrencyFieldsInstance.IsActive,
			CurrencyFieldsInstance.IsDefault,
			CurrencyFieldsInstance.IsBaseCurrency,
		},
		SearchFields: []admin.Field{
			CurrencyFieldsInstance.Code,
			CurrencyFieldsInstance.Name,
			CurrencyFieldsInstance.Symbol,
		},
		Ordering: []admin.Field{
			CurrencyFieldsInstance.Code,
		},
		Fieldsets: []admin.Fieldset[Currency]{
			{
				Name: "Basic Information",
				Fields: []string{"code", "name", "symbol"},
			},
			{
				Name: "Formatting",
				Fields: []string{"decimal_places", "decimal_separator", "thousand_separator", "symbol_position"},
			},
			{
				Name: "Exchange",
				Fields: []string{"is_base_currency", "exchange_rate"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active", "is_default"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"code", "name", "is_active", "is_default", "exchange_rate"},
		},
		ReadOnlyFields: []admin.Field{
			CurrencyFieldsInstance.CreatedAt,
			CurrencyFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}, obj *Currency) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}, obj *Currency) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}, obj *Currency) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}, obj *Currency) bool {
			return true
		},
		Filters: []admin.Filter[Currency]{
			{
				Name:  "active_currencies",
				Label: "Active Currencies Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Currency], value interface{}) orm.QuerySet[Currency] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "base_currency",
				Label: "Base Currency",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Currency], value interface{}) orm.QuerySet[Currency] {
					return qs.Filter("is_base_currency", true)
				},
			},
		},
		Actions: []admin.Action[Currency]{
			{
				Name:         "activate",
				Label:        "Activate Currencies",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected currencies?",
				Handler: func(ctx context.Context, admin *admin.Admin[Currency], ids []interface{}) error {
					for _, id := range ids {
						currency, err := CurrencyObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						currency.IsActive = true
						if err := CurrencyObjects.Update(ctx, currency); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Currencies",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected currencies?",
				Handler: func(ctx context.Context, admin *admin.Admin[Currency], ids []interface{}) error {
					for _, id := range ids {
						currency, err := CurrencyObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						currency.IsActive = false
						if err := CurrencyObjects.Update(ctx, currency); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}) bool {
					return true
				},
			},
		},
	})

	// ExchangeRate admin
	admin.Register(&admin.Config[ExchangeRate]{
		Icon: "ArrowsUpDown",
		ListDisplay: []admin.Field{
			ExchangeRateFieldsInstance.FromCurrency,
			ExchangeRateFieldsInstance.ToCurrency,
			ExchangeRateFieldsInstance.Rate,
			ExchangeRateFieldsInstance.EffectiveFrom,
			ExchangeRateFieldsInstance.EffectiveTo,
			ExchangeRateFieldsInstance.Provider,
		},
		ListFilter: []admin.Field{
			ExchangeRateFieldsInstance.FromCurrency,
			ExchangeRateFieldsInstance.ToCurrency,
		},
		SearchFields: []admin.Field{
			ExchangeRateFieldsInstance.FromCurrency,
			ExchangeRateFieldsInstance.ToCurrency,
			ExchangeRateFieldsInstance.Provider,
		},
		Ordering: []admin.Field{
			ExchangeRateFieldsInstance.FromCurrency,
			ExchangeRateFieldsInstance.ToCurrency,
			ExchangeRateFieldsInstance.EffectiveFrom,
		},
		Fieldsets: []admin.Fieldset[ExchangeRate]{
			{
				Name: "Rate Information",
				Fields: []string{"from_currency", "to_currency", "rate"},
			},
			{
				Name: "Effective Dates",
				Fields: []string{"effective_from", "effective_to"},
			},
			{
				Name: "Provider",
				Fields: []string{"provider"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"from_currency", "to_currency", "rate", "effective_from"},
		},
		ReadOnlyFields: []admin.Field{
			ExchangeRateFieldsInstance.CreatedAt,
			ExchangeRateFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ExchangeRate], user interface{}, obj *ExchangeRate) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[ExchangeRate], user interface{}, obj *ExchangeRate) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ExchangeRate], user interface{}, obj *ExchangeRate) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ExchangeRate], user interface{}, obj *ExchangeRate) bool {
			return true
		},
		Filters: []admin.Filter[ExchangeRate]{
			{
				Name:  "current_rates",
				Label: "Current Rates",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ExchangeRate], value interface{}) orm.QuerySet[ExchangeRate] {
					return qs.Filter("effective_to", nil)
				},
			},
			{
				Name:  "by_currency_pair",
				Label: "By Currency Pair",
				Type:  admin.FilterTypeString,
				Handler: func(ctx context.Context, qs orm.QuerySet[ExchangeRate], value interface{}) orm.QuerySet[ExchangeRate] {
					return qs.Filter("from_currency", value)
				},
			},
		},
		Actions: []admin.Action[ExchangeRate]{
			{
				Name:         "update_rates",
				Label:        "Update Exchange Rates",
				Icon:         "Refresh",
				Confirmation: "Fetch latest exchange rates from provider?",
				Handler: func(ctx context.Context, admin *admin.Admin[ExchangeRate], ids []interface{}) error {
					// Implementation would fetch latest rates from provider
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ExchangeRate], user interface{}) bool {
					return true
				},
			},
		},
	})
}

import (
	"context"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers commerce models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// ShippingMethod admin
	admin.Register(&admin.Config[ShippingMethod]{
		Icon: "Truck",
		ListDisplay: []admin.Field{
			ShippingMethodFieldsInstance.Name,
			ShippingMethodFieldsInstance.Code,
			ShippingMethodFieldsInstance.Carrier,
			ShippingMethodFieldsInstance.ServiceLevel,
			ShippingMethodFieldsInstance.BasePrice,
			ShippingMethodFieldsInstance.IsActive,
			ShippingMethodFieldsInstance.IsDefault,
			ShippingMethodFieldsInstance.Priority,
		},
		ListFilter: []admin.Field{
			ShippingMethodFieldsInstance.IsActive,
			ShippingMethodFieldsInstance.Carrier,
			ShippingMethodFieldsInstance.IsDefault,
		},
		SearchFields: []admin.Field{
			ShippingMethodFieldsInstance.Name,
			ShippingMethodFieldsInstance.Code,
			ShippingMethodFieldsInstance.Carrier,
			ShippingMethodFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			ShippingMethodFieldsInstance.Priority,
			ShippingMethodFieldsInstance.Name,
		},
		// Fieldsets for grouping form fields
		Fieldsets: []admin.Fieldset[ShippingMethod]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "code", "description", "carrier", "service_level"},
			},
			{
				Name: "Pricing",
				Fields: []string{"base_price", "handling_fee", "free_shipping_threshold"},
			},
			{
				Name: "Weight Constraints",
				Fields: []string{"min_weight", "max_weight", "weight_unit"},
			},
			{
				Name: "Dimensions",
				Fields: []string{"min_length", "max_length", "min_width", "max_width", "min_height", "max_height", "dimension_unit"},
			},
			{
				Name: "Delivery",
				Fields: []string{"min_days", "max_days"},
			},
			{
				Name: "Geographic Restrictions",
				Fields: []string{"countries", "excluded_countries"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active", "is_default", "priority"},
			},
		},
		// History manager for audit logging
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "code", "base_price", "is_active", "is_default"},
		},
		// Read-only fields
		ReadOnlyFields: []admin.Field{
			ShippingMethodFieldsInstance.CreatedAt,
			ShippingMethodFieldsInstance.UpdatedAt,
		},
		// Custom permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}, obj *ShippingMethod) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}, obj *ShippingMethod) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}, obj *ShippingMethod) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}, obj *ShippingMethod) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[ShippingMethod]{
			{
				Name:  "active_methods",
				Label: "Active Methods Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ShippingMethod], value interface{}) orm.QuerySet[ShippingMethod] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "default_method",
				Label: "Default Method",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ShippingMethod], value interface{}) orm.QuerySet[ShippingMethod] {
					return qs.Filter("is_default", true)
				},
			},
			{
				Name:  "by_carrier",
				Label: "By Carrier",
				Type:  admin.FilterTypeString,
				Handler: func(ctx context.Context, qs orm.QuerySet[ShippingMethod], value interface{}) orm.QuerySet[ShippingMethod] {
					return qs.Filter("carrier", value)
				},
			},
		},
		// Custom bulk actions
		Actions: []admin.Action[ShippingMethod]{
			{
				Name:         "activate",
				Label:        "Activate Methods",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected shipping methods?",
				Handler: func(ctx context.Context, admin *admin.Admin[ShippingMethod], ids []interface{}) error {
					for _, id := range ids {
						method, err := ShippingMethodObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						method.IsActive = true
						if err := ShippingMethodObjects.Update(ctx, method); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Methods",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected shipping methods?",
				Handler: func(ctx context.Context, admin *admin.Admin[ShippingMethod], ids []interface{}) error {
					for _, id := range ids {
						method, err := ShippingMethodObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						method.IsActive = false
						if err := ShippingMethodObjects.Update(ctx, method); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "set_default",
				Label:        "Set as Default",
				Icon:         "Star",
				Confirmation: "Set selected shipping method as the default?",
				Handler: func(ctx context.Context, admin *admin.Admin[ShippingMethod], ids []interface{}) error {
					for _, id := range ids {
						method, err := ShippingMethodObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						method.IsDefault = true
						if err := ShippingMethodObjects.Update(ctx, method); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ShippingMethod], user interface{}) bool {
					return true
				},
			},
		},
	})

	// PaymentMethod admin
	admin.Register(&admin.Config[PaymentMethod]{
		Icon: "CreditCard",
		ListDisplay: []admin.Field{
			PaymentMethodFieldsInstance.Name,
			PaymentMethodFieldsInstance.Code,
			PaymentMethodFieldsInstance.Type,
			PaymentMethodFieldsInstance.Gateway,
			PaymentMethodFieldsInstance.FixedFee,
			PaymentMethodFieldsInstance.PercentageFee,
			PaymentMethodFieldsInstance.IsActive,
			PaymentMethodFieldsInstance.IsDefault,
		},
		ListFilter: []admin.Field{
			PaymentMethodFieldsInstance.IsActive,
			PaymentMethodFieldsInstance.Type,
			PaymentMethodFieldsInstance.Gateway,
			PaymentMethodFieldsInstance.IsDefault,
		},
		SearchFields: []admin.Field{
			PaymentMethodFieldsInstance.Name,
			PaymentMethodFieldsInstance.Code,
			PaymentMethodFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			PaymentMethodFieldsInstance.DisplayOrder,
			PaymentMethodFieldsInstance.Name,
		},
		Fieldsets: []admin.Fieldset[PaymentMethod]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "code", "type", "description"},
			},
			{
				Name: "Gateway Configuration",
				Fields: []string{"gateway", "gateway_config"},
			},
			{
				Name: "Fees",
				Fields: []string{"fixed_fee", "percentage_fee"},
			},
			{
				Name: "Limits",
				Fields: []string{"min_amount", "max_amount"},
			},
			{
				Name: "Currencies",
				Fields: []string{"currencies"},
			},
			{
				Name: "UI Settings",
				Fields: []string{"icon", "display_order"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active", "is_default", "test_mode"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "code", "is_active", "is_default", "fixed_fee", "percentage_fee"},
		},
		ReadOnlyFields: []admin.Field{
			PaymentMethodFieldsInstance.CreatedAt,
			PaymentMethodFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}, obj *PaymentMethod) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}, obj *PaymentMethod) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}, obj *PaymentMethod) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}, obj *PaymentMethod) bool {
			return true
		},
		Filters: []admin.Filter[PaymentMethod]{
			{
				Name:  "active_methods",
				Label: "Active Methods Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[PaymentMethod], value interface{}) orm.QuerySet[PaymentMethod] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "by_type",
				Label: "By Type",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "credit_card", Label: "Credit Card"},
					{Value: "debit_card", Label: "Debit Card"},
					{Value: "paypal", Label: "PayPal"},
					{Value: "bank_transfer", Label: "Bank Transfer"},
					{Value: "cash_on_delivery", Label: "Cash on Delivery"},
					{Value: "crypto", Label: "Cryptocurrency"},
					{Value: "wallet", Label: "Digital Wallet"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[PaymentMethod], value interface{}) orm.QuerySet[PaymentMethod] {
					return qs.Filter("type", value)
				},
			},
		},
		Actions: []admin.Action[PaymentMethod]{
			{
				Name:         "activate",
				Label:        "Activate Methods",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected payment methods?",
				Handler: func(ctx context.Context, admin *admin.Admin[PaymentMethod], ids []interface{}) error {
					for _, id := range ids {
						method, err := PaymentMethodObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						method.IsActive = true
						if err := PaymentMethodObjects.Update(ctx, method); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Methods",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected payment methods?",
				Handler: func(ctx context.Context, admin *admin.Admin[PaymentMethod], ids []interface{}) error {
					for _, id := range ids {
						method, err := PaymentMethodObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						method.IsActive = false
						if err := PaymentMethodObjects.Update(ctx, method); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[PaymentMethod], user interface{}) bool {
					return true
				},
			},
		},
	})

	// TaxRate admin
	admin.Register(&admin.Config[TaxRate]{
		Icon: "Calculator",
		ListDisplay: []admin.Field{
			TaxRateFieldsInstance.Name,
			TaxRateFieldsInstance.Code,
			TaxRateFieldsInstance.Rate,
			TaxRateFieldsInstance.Country,
			TaxRateFieldsInstance.State,
			TaxRateFieldsInstance.TaxType,
			TaxRateFieldsInstance.IsActive,
			TaxRateFieldsInstance.Priority,
		},
		ListFilter: []admin.Field{
			TaxRateFieldsInstance.IsActive,
			TaxRateFieldsInstance.Country,
			TaxRateFieldsInstance.TaxType,
			TaxRateFieldsInstance.IsCompound,
		},
		SearchFields: []admin.Field{
			TaxRateFieldsInstance.Name,
			TaxRateFieldsInstance.Code,
			TaxRateFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			TaxRateFieldsInstance.Country,
			TaxRateFieldsInstance.State,
			TaxRateFieldsInstance.Priority,
		},
		Fieldsets: []admin.Fieldset[TaxRate]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "code", "description"},
			},
			{
				Name: "Rate",
				Fields: []string{"rate", "tax_type"},
			},
			{
				Name: "Geographic Scope",
				Fields: []string{"country", "state", "zip_pattern", "city"},
			},
			{
				Name: "Applicability",
				Fields: []string{"applies_to_products", "applies_to_shipping", "applies_to_services"},
			},
			{
				Name: "Date Range",
				Fields: []string{"start_date", "end_date"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active", "is_compound", "priority"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "code", "rate", "is_active", "is_compound"},
		},
		ReadOnlyFields: []admin.Field{
			TaxRateFieldsInstance.CreatedAt,
			TaxRateFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}, obj *TaxRate) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}, obj *TaxRate) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}, obj *TaxRate) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}, obj *TaxRate) bool {
			return true
		},
		Filters: []admin.Filter[TaxRate]{
			{
				Name:  "active_rates",
				Label: "Active Rates Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[TaxRate], value interface{}) orm.QuerySet[TaxRate] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "compound_rates",
				Label: "Compound Rates",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[TaxRate], value interface{}) orm.QuerySet[TaxRate] {
					return qs.Filter("is_compound", true)
				},
			},
			{
				Name:  "by_country",
				Label: "By Country",
				Type:  admin.FilterTypeString,
				Handler: func(ctx context.Context, qs orm.QuerySet[TaxRate], value interface{}) orm.QuerySet[TaxRate] {
					return qs.Filter("country", value)
				},
			},
		},
		Actions: []admin.Action[TaxRate]{
			{
				Name:         "activate",
				Label:        "Activate Rates",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected tax rates?",
				Handler: func(ctx context.Context, admin *admin.Admin[TaxRate], ids []interface{}) error {
					for _, id := range ids {
						rate, err := TaxRateObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						rate.IsActive = true
						if err := TaxRateObjects.Update(ctx, rate); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Rates",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected tax rates?",
				Handler: func(ctx context.Context, admin *admin.Admin[TaxRate], ids []interface{}) error {
					for _, id := range ids {
						rate, err := TaxRateObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						rate.IsActive = false
						if err := TaxRateObjects.Update(ctx, rate); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[TaxRate], user interface{}) bool {
					return true
				},
			},
		},
	})

	// Currency admin
	admin.Register(&admin.Config[Currency]{
		Icon: "DollarSign",
		ListDisplay: []admin.Field{
			CurrencyFieldsInstance.Code,
			CurrencyFieldsInstance.Name,
			CurrencyFieldsInstance.Symbol,
			CurrencyFieldsInstance.DecimalPlaces,
			CurrencyFieldsInstance.IsBaseCurrency,
			CurrencyFieldsInstance.ExchangeRate,
			CurrencyFieldsInstance.IsActive,
			CurrencyFieldsInstance.IsDefault,
		},
		ListFilter: []admin.Field{
			CurrencyFieldsInstance.IsActive,
			CurrencyFieldsInstance.IsDefault,
			CurrencyFieldsInstance.IsBaseCurrency,
		},
		SearchFields: []admin.Field{
			CurrencyFieldsInstance.Code,
			CurrencyFieldsInstance.Name,
			CurrencyFieldsInstance.Symbol,
		},
		Ordering: []admin.Field{
			CurrencyFieldsInstance.Code,
		},
		Fieldsets: []admin.Fieldset[Currency]{
			{
				Name: "Basic Information",
				Fields: []string{"code", "name", "symbol"},
			},
			{
				Name: "Formatting",
				Fields: []string{"decimal_places", "decimal_separator", "thousand_separator", "symbol_position"},
			},
			{
				Name: "Exchange",
				Fields: []string{"is_base_currency", "exchange_rate"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active", "is_default"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"code", "name", "is_active", "is_default", "exchange_rate"},
		},
		ReadOnlyFields: []admin.Field{
			CurrencyFieldsInstance.CreatedAt,
			CurrencyFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}, obj *Currency) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}, obj *Currency) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}, obj *Currency) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}, obj *Currency) bool {
			return true
		},
		Filters: []admin.Filter[Currency]{
			{
				Name:  "active_currencies",
				Label: "Active Currencies Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Currency], value interface{}) orm.QuerySet[Currency] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "base_currency",
				Label: "Base Currency",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Currency], value interface{}) orm.QuerySet[Currency] {
					return qs.Filter("is_base_currency", true)
				},
			},
		},
		Actions: []admin.Action[Currency]{
			{
				Name:         "activate",
				Label:        "Activate Currencies",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected currencies?",
				Handler: func(ctx context.Context, admin *admin.Admin[Currency], ids []interface{}) error {
					for _, id := range ids {
						currency, err := CurrencyObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						currency.IsActive = true
						if err := CurrencyObjects.Update(ctx, currency); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Currencies",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected currencies?",
				Handler: func(ctx context.Context, admin *admin.Admin[Currency], ids []interface{}) error {
					for _, id := range ids {
						currency, err := CurrencyObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						currency.IsActive = false
						if err := CurrencyObjects.Update(ctx, currency); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[Currency], user interface{}) bool {
					return true
				},
			},
		},
	})

	// ExchangeRate admin
	admin.Register(&admin.Config[ExchangeRate]{
		Icon: "ArrowsUpDown",
		ListDisplay: []admin.Field{
			ExchangeRateFieldsInstance.FromCurrency,
			ExchangeRateFieldsInstance.ToCurrency,
			ExchangeRateFieldsInstance.Rate,
			ExchangeRateFieldsInstance.EffectiveFrom,
			ExchangeRateFieldsInstance.EffectiveTo,
			ExchangeRateFieldsInstance.Provider,
		},
		ListFilter: []admin.Field{
			ExchangeRateFieldsInstance.FromCurrency,
			ExchangeRateFieldsInstance.ToCurrency,
		},
		SearchFields: []admin.Field{
			ExchangeRateFieldsInstance.FromCurrency,
			ExchangeRateFieldsInstance.ToCurrency,
			ExchangeRateFieldsInstance.Provider,
		},
		Ordering: []admin.Field{
			ExchangeRateFieldsInstance.FromCurrency,
			ExchangeRateFieldsInstance.ToCurrency,
			ExchangeRateFieldsInstance.EffectiveFrom,
		},
		Fieldsets: []admin.Fieldset[ExchangeRate]{
			{
				Name: "Rate Information",
				Fields: []string{"from_currency", "to_currency", "rate"},
			},
			{
				Name: "Effective Dates",
				Fields: []string{"effective_from", "effective_to"},
			},
			{
				Name: "Provider",
				Fields: []string{"provider"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"from_currency", "to_currency", "rate", "effective_from"},
		},
		ReadOnlyFields: []admin.Field{
			ExchangeRateFieldsInstance.CreatedAt,
			ExchangeRateFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ExchangeRate], user interface{}, obj *ExchangeRate) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[ExchangeRate], user interface{}, obj *ExchangeRate) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ExchangeRate], user interface{}, obj *ExchangeRate) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ExchangeRate], user interface{}, obj *ExchangeRate) bool {
			return true
		},
		Filters: []admin.Filter[ExchangeRate]{
			{
				Name:  "current_rates",
				Label: "Current Rates",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ExchangeRate], value interface{}) orm.QuerySet[ExchangeRate] {
					return qs.Filter("effective_to", nil)
				},
			},
			{
				Name:  "by_currency_pair",
				Label: "By Currency Pair",
				Type:  admin.FilterTypeString,
				Handler: func(ctx context.Context, qs orm.QuerySet[ExchangeRate], value interface{}) orm.QuerySet[ExchangeRate] {
					return qs.Filter("from_currency", value)
				},
			},
		},
		Actions: []admin.Action[ExchangeRate]{
			{
				Name:         "update_rates",
				Label:        "Update Exchange Rates",
				Icon:         "Refresh",
				Confirmation: "Fetch latest exchange rates from provider?",
				Handler: func(ctx context.Context, admin *admin.Admin[ExchangeRate], ids []interface{}) error {
					// Implementation would fetch latest rates from provider
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ExchangeRate], user interface{}) bool {
					return true
				},
			},
		},
	})
}

