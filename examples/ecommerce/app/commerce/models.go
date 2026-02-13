package commerce

import (
	"github.com/forgego/forge/schema"
)

// ShippingMethod represents a shipping method
type ShippingMethod struct {
	schema.BaseSchema
	Id                int64   `json:"id" db:"id"`
	Name              string  `json:"name" db:"name"`
	Code              string  `json:"code" db:"code"`
	Description       string  `json:"description" db:"description"`
	BasePrice         float64 `json:"base_price" db:"base_price"`
	PricePerKg        float64 `json:"price_per_kg" db:"price_per_kg"`
	EstimatedDaysMin  int32   `json:"estimated_days_min" db:"estimated_days_min"`
	EstimatedDaysMax  int32   `json:"estimated_days_max" db:"estimated_days_max"`
	IsActive          bool    `json:"is_active" db:"is_active"`
	SortOrder         int32   `json:"sort_order" db:"sort_order"`
	CarrierName       string  `json:"carrier_name" db:"carrier_name"`
	TrackingURLFormat string  `json:"tracking_url_format" db:"tracking_url_format"`
	CreatedAt         string  `json:"created_at" db:"created_at"`
	UpdatedAt         string  `json:"updated_at" db:"updated_at"`
}

func (ShippingMethod) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(200),
			schema.HelpText("Shipping method name (e.g., Standard Shipping)")),
		schema.StringField("code", schema.Required(), schema.MaxLength(50), schema.Unique(),
			schema.HelpText("Unique code (e.g., standard, express)")),
		schema.TextField("description", schema.Optional(),
			schema.HelpText("Detailed description")),
		schema.FloatField("base_price", schema.Default(0.0),
			schema.HelpText("Base shipping price")),
		schema.FloatField("price_per_kg", schema.Default(0.0),
			schema.HelpText("Additional price per kilogram")),
		schema.Int32Field("estimated_days_min", schema.Default(1),
			schema.VerboseName("Min Delivery Days"),
			schema.HelpText("Minimum estimated delivery days")),
		schema.Int32Field("estimated_days_max", schema.Default(7),
			schema.VerboseName("Max Delivery Days"),
			schema.HelpText("Maximum estimated delivery days")),
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active"),
			schema.HelpText("Is shipping method available")),
		schema.Int32Field("sort_order", schema.Default(0),
			schema.VerboseName("Sort Order")),
		schema.StringField("carrier_name", schema.MaxLength(200), schema.Optional(),
			schema.VerboseName("Carrier Name"),
			schema.HelpText("Shipping carrier (e.g., UPS, FedEx)")),
		schema.StringField("tracking_url_format", schema.MaxLength(500), schema.Optional(),
			schema.VerboseName("Tracking URL Format"),
			schema.HelpText("URL format for tracking (use {tracking_number})")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ShippingMethod) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "shipping_methods",
		VerboseName:       "Shipping Method",
		VerboseNamePlural: "Shipping Methods",
		OrderBy:           []string{"sort_order", "name"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_shipping_method_code", "code"),
			schema.IndexOn("idx_shipping_method_active", "is_active"),
		},
	}
}

func (ShippingMethod) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ShippingMethod) Hooks() *schema.ModelHooks {
	return nil
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	schema.BaseSchema
	Id             int64  `json:"id" db:"id"`
	Name           string `json:"name" db:"name"`
	Code           string `json:"code" db:"code"`
	Description    string `json:"description" db:"description"`
	ProcessorName  string `json:"processor_name" db:"processor_name"`
	IsActive       bool   `json:"is_active" db:"is_active"`
	SortOrder      int32  `json:"sort_order" db:"sort_order"`
	RequiresAuth   bool   `json:"requires_auth" db:"requires_auth"`
	SupportsRefund bool   `json:"supports_refund" db:"supports_refund"`
	IconURL        string `json:"icon_url" db:"icon_url"`
	CreatedAt      string `json:"created_at" db:"created_at"`
	UpdatedAt      string `json:"updated_at" db:"updated_at"`
}

func (PaymentMethod) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(200),
			schema.HelpText("Payment method name (e.g., Credit Card, PayPal)")),
		schema.StringField("code", schema.Required(), schema.MaxLength(50), schema.Unique(),
			schema.HelpText("Unique code (e.g., credit_card, paypal)")),
		schema.TextField("description", schema.Optional(),
			schema.HelpText("Detailed description")),
		schema.StringField("processor_name", schema.MaxLength(100), schema.Optional(),
			schema.VerboseName("Payment Processor"),
			schema.HelpText("Payment gateway (e.g., Stripe, PayPal, Square)")),
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.Int32Field("sort_order", schema.Default(0),
			schema.VerboseName("Sort Order")),
		schema.BoolField("requires_auth", schema.Default(false),
			schema.VerboseName("Requires Auth"),
			schema.HelpText("Requires 3D Secure or similar authentication")),
		schema.BoolField("supports_refund", schema.Default(true),
			schema.VerboseName("Supports Refund"),
			schema.HelpText("Can process refunds")),
		schema.StringField("icon_url", schema.MaxLength(500), schema.Optional(),
			schema.VerboseName("Icon URL")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (PaymentMethod) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "payment_methods",
		VerboseName:       "Payment Method",
		VerboseNamePlural: "Payment Methods",
		OrderBy:           []string{"sort_order", "name"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_payment_method_code", "code"),
			schema.IndexOn("idx_payment_method_active", "is_active"),
		},
	}
}

func (PaymentMethod) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (PaymentMethod) Hooks() *schema.ModelHooks {
	return nil
}

// TaxRate represents a tax rate configuration
type TaxRate struct {
	schema.BaseSchema
	Id           int64   `json:"id" db:"id"`
	Name         string  `json:"name" db:"name"`
	Code         string  `json:"code" db:"code"`
	Rate         float64 `json:"rate" db:"rate"`
	Country      string  `json:"country" db:"country"`
	State        string  `json:"state" db:"state"`
	City         string  `json:"city" db:"city"`
	ZipCode      string  `json:"zip_code" db:"zip_code"`
	IsCompound   bool    `json:"is_compound" db:"is_compound"`
	IsActive     bool    `json:"is_active" db:"is_active"`
	Priority     int32   `json:"priority" db:"priority"`
	ApplyToShip  bool    `json:"apply_to_shipping" db:"apply_to_shipping"`
	IncludedInPrice bool `json:"included_in_price" db:"included_in_price"`
	CreatedAt    string  `json:"created_at" db:"created_at"`
	UpdatedAt    string  `json:"updated_at" db:"updated_at"`
}

func (TaxRate) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(200),
			schema.HelpText("Tax rate name (e.g., California Sales Tax)")),
		schema.StringField("code", schema.Required(), schema.MaxLength(50), schema.Unique(),
			schema.HelpText("Unique code (e.g., CA_SALES_TAX)")),
		schema.FloatField("rate", schema.Required(),
			schema.HelpText("Tax rate as decimal (e.g., 0.0725 for 7.25%)")),
		schema.StringField("country", schema.MaxLength(2), schema.Optional(),
			schema.HelpText("ISO 3166-1 alpha-2 country code")),
		schema.StringField("state", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("State or province")),
		schema.StringField("city", schema.MaxLength(100), schema.Optional()),
		schema.StringField("zip_code", schema.MaxLength(20), schema.Optional(),
			schema.VerboseName("ZIP Code")),
		schema.BoolField("is_compound", schema.Default(false),
			schema.VerboseName("Compound Tax"),
			schema.HelpText("Tax calculated on top of other taxes")),
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.Int32Field("priority", schema.Default(0),
			schema.HelpText("Application order for compound taxes")),
		schema.BoolField("apply_to_shipping", schema.Default(false),
			schema.VerboseName("Apply to Shipping"),
			schema.HelpText("Apply tax to shipping costs")),
		schema.BoolField("included_in_price", schema.Default(false),
			schema.VerboseName("Included in Price"),
			schema.HelpText("Tax already included in product price")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (TaxRate) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "tax_rates",
		VerboseName:       "Tax Rate",
		VerboseNamePlural: "Tax Rates",
		OrderBy:           []string{"priority", "country", "state"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_tax_rate_code", "code"),
			schema.IndexOn("idx_tax_rate_location", "country", "state", "city"),
			schema.IndexOn("idx_tax_rate_active", "is_active"),
		},
	}
}

func (TaxRate) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (TaxRate) Hooks() *schema.ModelHooks {
	return nil
}

// Currency represents a currency
type Currency struct {
	schema.BaseSchema
	Id           int64  `json:"id" db:"id"`
	Code         string `json:"code" db:"code"`
	Name         string `json:"name" db:"name"`
	Symbol       string `json:"symbol" db:"symbol"`
	DecimalPlaces int32 `json:"decimal_places" db:"decimal_places"`
	IsActive     bool   `json:"is_active" db:"is_active"`
	IsDefault    bool   `json:"is_default" db:"is_default"`
	CreatedAt    string `json:"created_at" db:"created_at"`
	UpdatedAt    string `json:"updated_at" db:"updated_at"`
}

func (Currency) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("code", schema.Required(), schema.MaxLength(3), schema.Unique(),
			schema.HelpText("ISO 4217 currency code (e.g., USD, EUR)")),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Currency name (e.g., US Dollar)")),
		schema.StringField("symbol", schema.MaxLength(10), schema.Optional(),
			schema.HelpText("Currency symbol (e.g., $, €)")),
		schema.Int32Field("decimal_places", schema.Default(2),
			schema.VerboseName("Decimal Places"),
			schema.HelpText("Number of decimal places")),
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.BoolField("is_default", schema.Default(false),
			schema.VerboseName("Default Currency"),
			schema.HelpText("Default currency for the store")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Currency) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "currencies",
		VerboseName:       "Currency",
		VerboseNamePlural: "Currencies",
		OrderBy:           []string{"code"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_currency_code", "code"),
			schema.IndexOn("idx_currency_active", "is_active"),
			schema.IndexOn("idx_currency_default", "is_default"),
		},
	}
}

func (Currency) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Currency) Hooks() *schema.ModelHooks {
	return nil
}

// ExchangeRate represents currency exchange rates
type ExchangeRate struct {
	schema.BaseSchema
	Id             int64   `json:"id" db:"id"`
	FromCurrencyID int64   `json:"from_currency_id" db:"from_currency_id"`
	ToCurrencyID   int64   `json:"to_currency_id" db:"to_currency_id"`
	Rate           float64 `json:"rate" db:"rate"`
	EffectiveDate  string  `json:"effective_date" db:"effective_date"`
	Source         string  `json:"source" db:"source"`
	IsActive       bool    `json:"is_active" db:"is_active"`
	CreatedAt      string  `json:"created_at" db:"created_at"`
	UpdatedAt      string  `json:"updated_at" db:"updated_at"`
}

func (ExchangeRate) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("from_currency_id", schema.Required(),
			schema.VerboseName("From Currency")),
		schema.Int64Field("to_currency_id", schema.Required(),
			schema.VerboseName("To Currency")),
		schema.FloatField("rate", schema.Required(),
			schema.HelpText("Exchange rate (1 from_currency = rate * to_currency)")),
		schema.DateField("effective_date", schema.Required(),
			schema.VerboseName("Effective Date"),
			schema.HelpText("Date from which this rate is effective")),
		schema.StringField("source", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("Source of exchange rate (e.g., manual, API)")),
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ExchangeRate) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "exchange_rates",
		VerboseName:       "Exchange Rate",
		VerboseNamePlural: "Exchange Rates",
		OrderBy:           []string{"-effective_date"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_exchange_rate_currencies", "from_currency_id", "to_currency_id", "effective_date"),
			schema.IndexOn("idx_exchange_rate_active", "is_active"),
		},
	}
}

func (ExchangeRate) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("from_currency_id", "Currency",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("exchange_rates_from")),
		schema.ForeignKeyField("to_currency_id", "Currency",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("exchange_rates_to")),
	}
}

func (ExchangeRate) Hooks() *schema.ModelHooks {
	return nil
}
