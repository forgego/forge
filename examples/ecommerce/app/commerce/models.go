package commerce

import (
	"context"
	"time"

	"github.com/forgego/forge/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShippingMethod represents a shipping method configuration
type ShippingMethod struct {
	ID                  primitive.ObjectID     `bson:"_id,omitempty" json:"id" db:"_id"`
	Name                string                 `bson:"name" json:"name" db:"name"`
	Code                string                 `bson:"code" json:"code" db:"code"`
	Description         string                 `bson:"description" json:"description" db:"description"`
	Carrier             string                 `bson:"carrier" json:"carrier" db:"carrier"`
	ServiceLevel        string                 `bson:"service_level" json:"service_level" db:"service_level"`

	// Pricing
	BasePrice           float64              `bson:"base_price" json:"base_price" db:"base_price"`
	HandlingFee         float64              `bson:"handling_fee" json:"handling_fee" db:"handling_fee"`
	FreeShippingThreshold float64            `bson:"free_shipping_threshold" json:"free_shipping_threshold" db:"free_shipping_threshold"`

	// Weight constraints
	MinWeight           float64              `bson:"min_weight" json:"min_weight" db:"min_weight"`
	MaxWeight           float64              `bson:"max_weight" json:"max_weight" db:"max_weight"`
	WeightUnit          string               `bson:"weight_unit" json:"weight_unit" db:"weight_unit"`

	// Dimensions
	MinLength           float64              `bson:"min_length" json:"min_length" db:"min_length"`
	MaxLength           float64              `bson:"max_length" json:"max_length" db:"max_length"`
	MinWidth            float64              `bson:"min_width" json:"min_width" db:"min_width"`
	MaxWidth            float64              `bson:"max_width" json:"max_width" db:"max_width"`
	MinHeight           float64              `bson:"min_height" json:"min_height" db:"min_height"`
	MaxHeight           float64              `bson:"max_height" json:"max_height" db:"max_height"`
	DimensionUnit       string               `bson:"dimension_unit" json:"dimension_unit" db:"dimension_unit"`

	// Estimated delivery
	MinDays             int                  `bson:"min_days" json:"min_days" db:"min_days"`
	MaxDays             int                  `bson:"max_days" json:"max_days" db:"max_days"`

	// Geographic restrictions
	Countries           []string             `bson:"countries" json:"countries" db:"countries"`
	ExcludedCountries   []string             `bson:"excluded_countries" json:"excluded_countries" db:"excluded_countries"`

	// Status
	IsActive            bool                 `bson:"is_active" json:"is_active" db:"is_active"`
	IsDefault           bool                 `bson:"is_default" json:"is_default" db:"is_default"`
	Priority            int                  `bson:"priority" json:"priority" db:"priority"`

	// Metadata
	Config              map[string]interface{} `bson:"config" json:"config" db:"config"`
	CreatedAt           time.Time             `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt           time.Time             `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (ShippingMethod) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Shipping method name")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique code for the shipping method")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional()),
		schema.StringField("carrier", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Carrier name")),
		schema.StringField("service_level", schema.Required(), schema.MaxLength(50),
			schema.HelpText("Service level (e.g., express, standard)")),

		// Pricing
		schema.Float64Field("base_price", schema.Default(0.0),
			schema.HelpText("Base shipping price")),
		schema.Float64Field("handling_fee", schema.Default(0.0),
			schema.HelpText("Additional handling fee")),
		schema.Float64Field("free_shipping_threshold", schema.Default(0.0),
			schema.HelpText("Order total for free shipping")),

		// Weight constraints
		schema.Float64Field("min_weight", schema.Default(0.0),
			schema.HelpText("Minimum weight allowed")),
		schema.Float64Field("max_weight", schema.Default(0.0),
			schema.HelpText("Maximum weight allowed")),
		schema.StringField("weight_unit", schema.Default("kg"),
			schema.HelpText("Weight unit: kg, lb, oz, g")),

		// Dimensions
		schema.Float64Field("min_length", schema.Default(0.0)),
		schema.Float64Field("max_length", schema.Default(0.0)),
		schema.Float64Field("min_width", schema.Default(0.0)),
		schema.Float64Field("max_width", schema.Default(0.0)),
		schema.Float64Field("min_height", schema.Default(0.0)),
		schema.Float64Field("max_height", schema.Default(0.0)),
		schema.StringField("dimension_unit", schema.Default("cm"),
			schema.HelpText("Dimension unit: cm, in, m")),

		// Estimated delivery
		schema.IntField("min_days", schema.Default(0),
			schema.HelpText("Minimum delivery days")),
		schema.IntField("max_days", schema.Default(0),
			schema.HelpText("Maximum delivery days")),

		// Geographic restrictions
		schema.StringArrayField("countries", schema.Optional(),
			schema.HelpText("Allowed countries (empty = all)")),
		schema.StringArrayField("excluded_countries", schema.Optional(),
			schema.HelpText("Excluded countries")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_default", schema.Default(false)),
		schema.IntField("priority", schema.Default(0),
			schema.HelpText("Display priority (higher = first)")),

		// Metadata
		schema.MapField("config", schema.Optional()),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ShippingMethod) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "shipping_methods",
		VerboseName:       "Shipping Method",
		VerboseNamePlural: "Shipping Methods",
		OrderBy:           []string{"priority", "name"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_shipping_code", "code"),
			schema.IndexOn("idx_shipping_carrier", "carrier"),
			schema.IndexOn("idx_shipping_active", "is_active"),
			schema.IndexOn("idx_shipping_default", "is_default"),
		},
	}
}

func (ShippingMethod) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ShippingMethod) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate weight and dimension constraints
			// Ensure only one default shipping method
			return nil
		},
	}
}

// PaymentMethod represents a payment method configuration
type PaymentMethod struct {
	ID              primitive.ObjectID        `bson:"_id,omitempty" json:"id" db:"_id"`
	Name            string                   `bson:"name" json:"name" db:"name"`
	Code            string                   `bson:"code" json:"code" db:"code"`
	Type            string                   `bson:"type" json:"type" db:"type"`
	Description     string                   `bson:"description" json:"description" db:"description"`

	// Gateway configuration
	Gateway         string                   `bson:"gateway" json:"gateway" db:"gateway"`
	GatewayConfig   map[string]interface{}   `bson:"gateway_config" json:"gateway_config" db:"gateway_config"`

	// Fees
	FixedFee        float64                  `bson:"fixed_fee" json:"fixed_fee" db:"fixed_fee"`
	PercentageFee   float64                  `bson:"percentage_fee" json:"percentage_fee" db:"percentage_fee"`

	// Limits
	MinAmount       float64                  `bson:"min_amount" json:"min_amount" db:"min_amount"`
	MaxAmount       float64                  `bson:"max_amount" json:"max_amount" db:"max_amount"`

	// Currencies
	Currencies      []string                 `bson:"currencies" json:"currencies" db:"currencies"`

	// UI
	Icon            string                   `bson:"icon" json:"icon" db:"icon"`
	DisplayOrder    int                      `bson:"display_order" json:"display_order" db:"display_order"`

	// Status
	IsActive        bool                     `bson:"is_active" json:"is_active" db:"is_active"`
	IsDefault       bool                     `bson:"is_default" json:"is_default" db:"is_default"`

	// Test mode
	TestMode        bool                     `bson:"test_mode" json:"test_mode" db:"test_mode"`

	CreatedAt       time.Time                `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time                `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (PaymentMethod) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Payment method name")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique code for the payment method")),
		schema.StringField("type", schema.Required(),
			schema.HelpText("Type: credit_card, debit_card, paypal, bank_transfer, cash_on_delivery, crypto, wallet")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional()),

		// Gateway configuration
		schema.StringField("gateway", schema.Required(), schema.MaxLength(50),
			schema.HelpText("Payment gateway name")),
		schema.MapField("gateway_config", schema.Optional(),
			schema.HelpText("Gateway-specific configuration")),

		// Fees
		schema.Float64Field("fixed_fee", schema.Default(0.0),
			schema.HelpText("Fixed fee per transaction")),
		schema.Float64Field("percentage_fee", schema.Default(0.0),
			schema.HelpText("Percentage fee (0-100)")),

		// Limits
		schema.Float64Field("min_amount", schema.Default(0.0),
			schema.HelpText("Minimum transaction amount")),
		schema.Float64Field("max_amount", schema.Default(0.0),
			schema.HelpText("Maximum transaction amount (0 = unlimited)")),

		// Currencies
		schema.StringArrayField("currencies", schema.Optional(),
			schema.HelpText("Supported currencies")),

		// UI
		schema.StringField("icon", schema.MaxLength(255), schema.Optional(),
			schema.HelpText("Icon URL or name")),
		schema.IntField("display_order", schema.Default(0),
			schema.HelpText("Display order")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_default", schema.Default(false)),
		schema.BoolField("test_mode", schema.Default(false)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (PaymentMethod) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "payment_methods",
		VerboseName:       "Payment Method",
		VerboseNamePlural: "Payment Methods",
		OrderBy:           []string{"display_order", "name"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_payment_code", "code"),
			schema.IndexOn("idx_payment_type", "type"),
			schema.IndexOn("idx_payment_active", "is_active"),
			schema.IndexOn("idx_payment_default", "is_default"),
		},
	}
}

func (PaymentMethod) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (PaymentMethod) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate fee percentages
			// Ensure only one default payment method
			return nil
		},
	}
}

// TaxRate represents a tax rate configuration
type TaxRate struct {
	ID                primitive.ObjectID  `bson:"_id,omitempty" json:"id" db:"_id"`
	Name              string             `bson:"name" json:"name" db:"name"`
	Code              string             `bson:"code" json:"code" db:"code"`
	Description       string             `bson:"description" json:"description" db:"description"`

	// Rate
	Rate              float64            `bson:"rate" json:"rate" db:"rate"`

	// Geographic scope
	Country           string             `bson:"country" json:"country" db:"country"`
	State             string             `bson:"state" json:"state" db:"state"`
	ZipPattern        string             `bson:"zip_pattern" json:"zip_pattern" db:"zip_pattern"`
	City              string             `bson:"city" json:"city" db:"city"`

	// Tax type
	TaxType           string             `bson:"tax_type" json:"tax_type" db:"tax_type"`

	// Product/category applicability
	AppliesToProducts bool               `bson:"applies_to_products" json:"applies_to_products" db:"applies_to_products"`
	AppliesToShipping bool               `bson:"applies_to_shipping" json:"applies_to_shipping" db:"applies_to_shipping"`
	AppliesToServices bool               `bson:"applies_to_services" json:"applies_to_services" db:"applies_to_services"`

	// Dates
	StartDate         *time.Time         `bson:"start_date,omitempty" json:"start_date,omitempty" db:"start_date"`
	EndDate           *time.Time         `bson:"end_date,omitempty" json:"end_date,omitempty" db:"end_date"`

	// Status
	IsActive          bool               `bson:"is_active" json:"is_active" db:"is_active"`
	IsCompound        bool               `bson:"is_compound" json:"is_compound" db:"is_compound"`
	Priority          int                `bson:"priority" json:"priority" db:"priority"`

	CreatedAt         time.Time          `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (TaxRate) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Tax rate name")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique code for the tax rate")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional()),

		// Rate
		schema.Float64Field("rate", schema.Required(),
			schema.HelpText("Tax rate percentage (0-100)")),

		// Geographic scope
		schema.StringField("country", schema.Required(), schema.Length(2),
			schema.HelpText("Country code (ISO 2)")),
		schema.StringField("state", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("State/Province (optional)")),
		schema.StringField("zip_pattern", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("ZIP/Postal code pattern for matching")),
		schema.StringField("city", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("City (optional)")),

		// Tax type
		schema.StringField("tax_type",
			schema.HelpText("Type: sales, vat, gst, excised")),

		// Product/category applicability
		schema.BoolField("applies_to_products", schema.Default(true)),
		schema.BoolField("applies_to_shipping", schema.Default(true)),
		schema.BoolField("applies_to_services", schema.Default(false)),

		// Dates
		schema.TimeField("start_date", schema.Optional()),
		schema.TimeField("end_date", schema.Optional()),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_compound", schema.Default(false),
			schema.HelpText("Applied after other taxes")),
		schema.IntField("priority", schema.Default(0),
			schema.HelpText("Priority for compound tax calculation")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (TaxRate) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "tax_rates",
		VerboseName:       "Tax Rate",
		VerboseNamePlural: "Tax Rates",
		OrderBy:           []string{"country", "state", "priority"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_tax_code", "code"),
			schema.IndexOn("idx_tax_country", "country"),
			schema.IndexOn("idx_tax_state", "state"),
			schema.IndexOn("idx_tax_active", "is_active"),
			schema.IndexOn("idx_tax_compound", "is_compound"),
		},
	}
}

func (TaxRate) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (TaxRate) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate rate is within bounds
			// Validate date ranges
			return nil
		},
	}
}

// Currency represents a currency configuration
type Currency struct {
	ID                  primitive.ObjectID  `bson:"_id,omitempty" json:"id" db:"_id"`
	Code                string             `bson:"code" json:"code" db:"code"`
	Name                string             `bson:"name" json:"name" db:"name"`
	Symbol              string             `bson:"symbol" json:"symbol" db:"symbol"`

	// Formatting
	DecimalPlaces       int                `bson:"decimal_places" json:"decimal_places" db:"decimal_places"`
	DecimalSeparator     string            `bson:"decimal_separator" json:"decimal_separator" db:"decimal_separator"`
	ThousandSeparator    string            `bson:"thousand_separator" json:"thousand_separator" db:"thousand_separator"`
	SymbolPosition       string             `bson:"symbol_position" json:"symbol_position" db:"symbol_position"`

	// Exchange
	IsBaseCurrency      bool               `bson:"is_base_currency" json:"is_base_currency" db:"is_base_currency"`
	ExchangeRate         float64            `bson:"exchange_rate" json:"exchange_rate" db:"exchange_rate"`

	// Status
	IsActive            bool               `bson:"is_active" json:"is_active" db:"is_active"`
	IsDefault           bool               `bson:"is_default" json:"is_default" db:"is_default"`

	CreatedAt           time.Time         `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt           time.Time         `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (Currency) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("code", schema.Required(), schema.Length(3),
			schema.HelpText("Currency code (ISO 4217)")),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Currency name")),
		schema.StringField("symbol", schema.Required(), schema.MaxLength(5),
			schema.HelpText("Currency symbol")),

		// Formatting
		schema.IntField("decimal_places", schema.Required(), schema.Default(2),
			schema.HelpText("Number of decimal places")),
		schema.StringField("decimal_separator", schema.Default("."),
			schema.HelpText("Decimal separator: . or ,")),
		schema.StringField("thousand_separator", schema.Default(","),
			schema.HelpText("Thousand separator: , . or space")),
		schema.StringField("symbol_position",
			schema.HelpText("Symbol position: before, after, with_space, without_space")),

		// Exchange
		schema.BoolField("is_base_currency", schema.Default(false)),
		schema.Float64Field("exchange_rate", schema.Default(1.0),
			schema.HelpText("Exchange rate relative to base currency")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_default", schema.Default(false)),

		// Timestamps
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
			schema.IndexOn("idx_currency_base", "is_base_currency"),
		},
	}
}

func (Currency) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Currency) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Ensure only one base currency
			// Ensure only one default currency
			return nil
		},
	}
}

// ExchangeRate represents an exchange rate between currencies
type ExchangeRate struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id" db:"_id"`
	FromCurrency    string             `bson:"from_currency" json:"from_currency" db:"from_currency"`
	ToCurrency      string             `bson:"to_currency" json:"to_currency" db:"to_currency"`
	Rate            float64            `bson:"rate" json:"rate" db:"rate"`

	// Dates
	EffectiveFrom   time.Time          `bson:"effective_from" json:"effective_from" db:"effective_from"`
	EffectiveTo     *time.Time         `bson:"effective_to,omitempty" json:"effective_to,omitempty" db:"effective_to"`

	// Provider
	Provider        string             `bson:"provider" json:"provider" db:"provider"`

	CreatedAt       time.Time          `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (ExchangeRate) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("from_currency", schema.Required(), schema.Length(3),
			schema.HelpText("Source currency code")),
		schema.StringField("to_currency", schema.Required(), schema.Length(3),
			schema.HelpText("Target currency code")),
		schema.Float64Field("rate", schema.Required(),
			schema.HelpText("Exchange rate")),

		// Dates
		schema.TimeField("effective_from", schema.Required(),
			schema.HelpText("Start date for this rate")),
		schema.TimeField("effective_to", schema.Optional(),
			schema.HelpText("End date for this rate (null = current)")),

		// Provider
		schema.StringField("provider", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Rate provider name")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ExchangeRate) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "exchange_rates",
		VerboseName:       "Exchange Rate",
		VerboseNamePlural: "Exchange Rates",
		OrderBy:           []string{"from_currency", "to_currency", "-effective_from"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_exchange_from", "from_currency"),
			schema.IndexOn("idx_exchange_to", "to_currency"),
			schema.IndexOn("idx_exchange_date", "effective_from"),
			schema.IndexOn("idx_exchange_pair", []string{"from_currency", "to_currency"}),
		},
		UniqueTogether: [][]string{
			{"from_currency", "to_currency", "effective_from"},
		},
	}
}

func (ExchangeRate) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ExchangeRate) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate date ranges
			// Ensure no overlapping rates for same currency pair
			return nil
		},
	}
}

import (
	"context"
	"time"

	"github.com/forgego/forge/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShippingMethod represents a shipping method configuration
type ShippingMethod struct {
	ID                  primitive.ObjectID     `bson:"_id,omitempty" json:"id" db:"_id"`
	Name                string                 `bson:"name" json:"name" db:"name"`
	Code                string                 `bson:"code" json:"code" db:"code"`
	Description         string                 `bson:"description" json:"description" db:"description"`
	Carrier             string                 `bson:"carrier" json:"carrier" db:"carrier"`
	ServiceLevel        string                 `bson:"service_level" json:"service_level" db:"service_level"`

	// Pricing
	BasePrice           float64              `bson:"base_price" json:"base_price" db:"base_price"`
	HandlingFee         float64              `bson:"handling_fee" json:"handling_fee" db:"handling_fee"`
	FreeShippingThreshold float64            `bson:"free_shipping_threshold" json:"free_shipping_threshold" db:"free_shipping_threshold"`

	// Weight constraints
	MinWeight           float64              `bson:"min_weight" json:"min_weight" db:"min_weight"`
	MaxWeight           float64              `bson:"max_weight" json:"max_weight" db:"max_weight"`
	WeightUnit          string               `bson:"weight_unit" json:"weight_unit" db:"weight_unit"`

	// Dimensions
	MinLength           float64              `bson:"min_length" json:"min_length" db:"min_length"`
	MaxLength           float64              `bson:"max_length" json:"max_length" db:"max_length"`
	MinWidth            float64              `bson:"min_width" json:"min_width" db:"min_width"`
	MaxWidth            float64              `bson:"max_width" json:"max_width" db:"max_width"`
	MinHeight           float64              `bson:"min_height" json:"min_height" db:"min_height"`
	MaxHeight           float64              `bson:"max_height" json:"max_height" db:"max_height"`
	DimensionUnit       string               `bson:"dimension_unit" json:"dimension_unit" db:"dimension_unit"`

	// Estimated delivery
	MinDays             int                  `bson:"min_days" json:"min_days" db:"min_days"`
	MaxDays             int                  `bson:"max_days" json:"max_days" db:"max_days"`

	// Geographic restrictions
	Countries           []string             `bson:"countries" json:"countries" db:"countries"`
	ExcludedCountries   []string             `bson:"excluded_countries" json:"excluded_countries" db:"excluded_countries"`

	// Status
	IsActive            bool                 `bson:"is_active" json:"is_active" db:"is_active"`
	IsDefault           bool                 `bson:"is_default" json:"is_default" db:"is_default"`
	Priority            int                  `bson:"priority" json:"priority" db:"priority"`

	// Metadata
	Config              map[string]interface{} `bson:"config" json:"config" db:"config"`
	CreatedAt           time.Time             `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt           time.Time             `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (ShippingMethod) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Shipping method name")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique code for the shipping method")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional()),
		schema.StringField("carrier", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Carrier name")),
		schema.StringField("service_level", schema.Required(), schema.MaxLength(50),
			schema.HelpText("Service level (e.g., express, standard)")),

		// Pricing
		schema.Float64Field("base_price", schema.Default(0.0),
			schema.HelpText("Base shipping price")),
		schema.Float64Field("handling_fee", schema.Default(0.0),
			schema.HelpText("Additional handling fee")),
		schema.Float64Field("free_shipping_threshold", schema.Default(0.0),
			schema.HelpText("Order total for free shipping")),

		// Weight constraints
		schema.Float64Field("min_weight", schema.Default(0.0),
			schema.HelpText("Minimum weight allowed")),
		schema.Float64Field("max_weight", schema.Default(0.0),
			schema.HelpText("Maximum weight allowed")),
		schema.StringField("weight_unit", schema.Default("kg"),
			schema.HelpText("Weight unit: kg, lb, oz, g")),

		// Dimensions
		schema.Float64Field("min_length", schema.Default(0.0)),
		schema.Float64Field("max_length", schema.Default(0.0)),
		schema.Float64Field("min_width", schema.Default(0.0)),
		schema.Float64Field("max_width", schema.Default(0.0)),
		schema.Float64Field("min_height", schema.Default(0.0)),
		schema.Float64Field("max_height", schema.Default(0.0)),
		schema.StringField("dimension_unit", schema.Default("cm"),
			schema.HelpText("Dimension unit: cm, in, m")),

		// Estimated delivery
		schema.IntField("min_days", schema.Default(0),
			schema.HelpText("Minimum delivery days")),
		schema.IntField("max_days", schema.Default(0),
			schema.HelpText("Maximum delivery days")),

		// Geographic restrictions
		schema.StringArrayField("countries", schema.Optional(),
			schema.HelpText("Allowed countries (empty = all)")),
		schema.StringArrayField("excluded_countries", schema.Optional(),
			schema.HelpText("Excluded countries")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_default", schema.Default(false)),
		schema.IntField("priority", schema.Default(0),
			schema.HelpText("Display priority (higher = first)")),

		// Metadata
		schema.MapField("config", schema.Optional()),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ShippingMethod) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "shipping_methods",
		VerboseName:       "Shipping Method",
		VerboseNamePlural: "Shipping Methods",
		OrderBy:           []string{"priority", "name"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_shipping_code", "code"),
			schema.IndexOn("idx_shipping_carrier", "carrier"),
			schema.IndexOn("idx_shipping_active", "is_active"),
			schema.IndexOn("idx_shipping_default", "is_default"),
		},
	}
}

func (ShippingMethod) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ShippingMethod) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate weight and dimension constraints
			// Ensure only one default shipping method
			return nil
		},
	}
}

// PaymentMethod represents a payment method configuration
type PaymentMethod struct {
	ID              primitive.ObjectID        `bson:"_id,omitempty" json:"id" db:"_id"`
	Name            string                   `bson:"name" json:"name" db:"name"`
	Code            string                   `bson:"code" json:"code" db:"code"`
	Type            string                   `bson:"type" json:"type" db:"type"`
	Description     string                   `bson:"description" json:"description" db:"description"`

	// Gateway configuration
	Gateway         string                   `bson:"gateway" json:"gateway" db:"gateway"`
	GatewayConfig   map[string]interface{}   `bson:"gateway_config" json:"gateway_config" db:"gateway_config"`

	// Fees
	FixedFee        float64                  `bson:"fixed_fee" json:"fixed_fee" db:"fixed_fee"`
	PercentageFee   float64                  `bson:"percentage_fee" json:"percentage_fee" db:"percentage_fee"`

	// Limits
	MinAmount       float64                  `bson:"min_amount" json:"min_amount" db:"min_amount"`
	MaxAmount       float64                  `bson:"max_amount" json:"max_amount" db:"max_amount"`

	// Currencies
	Currencies      []string                 `bson:"currencies" json:"currencies" db:"currencies"`

	// UI
	Icon            string                   `bson:"icon" json:"icon" db:"icon"`
	DisplayOrder    int                      `bson:"display_order" json:"display_order" db:"display_order"`

	// Status
	IsActive        bool                     `bson:"is_active" json:"is_active" db:"is_active"`
	IsDefault       bool                     `bson:"is_default" json:"is_default" db:"is_default"`

	// Test mode
	TestMode        bool                     `bson:"test_mode" json:"test_mode" db:"test_mode"`

	CreatedAt       time.Time                `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time                `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (PaymentMethod) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Payment method name")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique code for the payment method")),
		schema.StringField("type", schema.Required(),
			schema.HelpText("Type: credit_card, debit_card, paypal, bank_transfer, cash_on_delivery, crypto, wallet")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional()),

		// Gateway configuration
		schema.StringField("gateway", schema.Required(), schema.MaxLength(50),
			schema.HelpText("Payment gateway name")),
		schema.MapField("gateway_config", schema.Optional(),
			schema.HelpText("Gateway-specific configuration")),

		// Fees
		schema.Float64Field("fixed_fee", schema.Default(0.0),
			schema.HelpText("Fixed fee per transaction")),
		schema.Float64Field("percentage_fee", schema.Default(0.0),
			schema.HelpText("Percentage fee (0-100)")),

		// Limits
		schema.Float64Field("min_amount", schema.Default(0.0),
			schema.HelpText("Minimum transaction amount")),
		schema.Float64Field("max_amount", schema.Default(0.0),
			schema.HelpText("Maximum transaction amount (0 = unlimited)")),

		// Currencies
		schema.StringArrayField("currencies", schema.Optional(),
			schema.HelpText("Supported currencies")),

		// UI
		schema.StringField("icon", schema.MaxLength(255), schema.Optional(),
			schema.HelpText("Icon URL or name")),
		schema.IntField("display_order", schema.Default(0),
			schema.HelpText("Display order")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_default", schema.Default(false)),
		schema.BoolField("test_mode", schema.Default(false)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (PaymentMethod) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "payment_methods",
		VerboseName:       "Payment Method",
		VerboseNamePlural: "Payment Methods",
		OrderBy:           []string{"display_order", "name"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_payment_code", "code"),
			schema.IndexOn("idx_payment_type", "type"),
			schema.IndexOn("idx_payment_active", "is_active"),
			schema.IndexOn("idx_payment_default", "is_default"),
		},
	}
}

func (PaymentMethod) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (PaymentMethod) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate fee percentages
			// Ensure only one default payment method
			return nil
		},
	}
}

// TaxRate represents a tax rate configuration
type TaxRate struct {
	ID                primitive.ObjectID  `bson:"_id,omitempty" json:"id" db:"_id"`
	Name              string             `bson:"name" json:"name" db:"name"`
	Code              string             `bson:"code" json:"code" db:"code"`
	Description       string             `bson:"description" json:"description" db:"description"`

	// Rate
	Rate              float64            `bson:"rate" json:"rate" db:"rate"`

	// Geographic scope
	Country           string             `bson:"country" json:"country" db:"country"`
	State             string             `bson:"state" json:"state" db:"state"`
	ZipPattern        string             `bson:"zip_pattern" json:"zip_pattern" db:"zip_pattern"`
	City              string             `bson:"city" json:"city" db:"city"`

	// Tax type
	TaxType           string             `bson:"tax_type" json:"tax_type" db:"tax_type"`

	// Product/category applicability
	AppliesToProducts bool               `bson:"applies_to_products" json:"applies_to_products" db:"applies_to_products"`
	AppliesToShipping bool               `bson:"applies_to_shipping" json:"applies_to_shipping" db:"applies_to_shipping"`
	AppliesToServices bool               `bson:"applies_to_services" json:"applies_to_services" db:"applies_to_services"`

	// Dates
	StartDate         *time.Time         `bson:"start_date,omitempty" json:"start_date,omitempty" db:"start_date"`
	EndDate           *time.Time         `bson:"end_date,omitempty" json:"end_date,omitempty" db:"end_date"`

	// Status
	IsActive          bool               `bson:"is_active" json:"is_active" db:"is_active"`
	IsCompound        bool               `bson:"is_compound" json:"is_compound" db:"is_compound"`
	Priority          int                `bson:"priority" json:"priority" db:"priority"`

	CreatedAt         time.Time          `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (TaxRate) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Tax rate name")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique code for the tax rate")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional()),

		// Rate
		schema.Float64Field("rate", schema.Required(),
			schema.HelpText("Tax rate percentage (0-100)")),

		// Geographic scope
		schema.StringField("country", schema.Required(), schema.Length(2),
			schema.HelpText("Country code (ISO 2)")),
		schema.StringField("state", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("State/Province (optional)")),
		schema.StringField("zip_pattern", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("ZIP/Postal code pattern for matching")),
		schema.StringField("city", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("City (optional)")),

		// Tax type
		schema.StringField("tax_type",
			schema.HelpText("Type: sales, vat, gst, excised")),

		// Product/category applicability
		schema.BoolField("applies_to_products", schema.Default(true)),
		schema.BoolField("applies_to_shipping", schema.Default(true)),
		schema.BoolField("applies_to_services", schema.Default(false)),

		// Dates
		schema.TimeField("start_date", schema.Optional()),
		schema.TimeField("end_date", schema.Optional()),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_compound", schema.Default(false),
			schema.HelpText("Applied after other taxes")),
		schema.IntField("priority", schema.Default(0),
			schema.HelpText("Priority for compound tax calculation")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (TaxRate) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "tax_rates",
		VerboseName:       "Tax Rate",
		VerboseNamePlural: "Tax Rates",
		OrderBy:           []string{"country", "state", "priority"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_tax_code", "code"),
			schema.IndexOn("idx_tax_country", "country"),
			schema.IndexOn("idx_tax_state", "state"),
			schema.IndexOn("idx_tax_active", "is_active"),
			schema.IndexOn("idx_tax_compound", "is_compound"),
		},
	}
}

func (TaxRate) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (TaxRate) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate rate is within bounds
			// Validate date ranges
			return nil
		},
	}
}

// Currency represents a currency configuration
type Currency struct {
	ID                  primitive.ObjectID  `bson:"_id,omitempty" json:"id" db:"_id"`
	Code                string             `bson:"code" json:"code" db:"code"`
	Name                string             `bson:"name" json:"name" db:"name"`
	Symbol              string             `bson:"symbol" json:"symbol" db:"symbol"`

	// Formatting
	DecimalPlaces       int                `bson:"decimal_places" json:"decimal_places" db:"decimal_places"`
	DecimalSeparator     string            `bson:"decimal_separator" json:"decimal_separator" db:"decimal_separator"`
	ThousandSeparator    string            `bson:"thousand_separator" json:"thousand_separator" db:"thousand_separator"`
	SymbolPosition       string             `bson:"symbol_position" json:"symbol_position" db:"symbol_position"`

	// Exchange
	IsBaseCurrency      bool               `bson:"is_base_currency" json:"is_base_currency" db:"is_base_currency"`
	ExchangeRate         float64            `bson:"exchange_rate" json:"exchange_rate" db:"exchange_rate"`

	// Status
	IsActive            bool               `bson:"is_active" json:"is_active" db:"is_active"`
	IsDefault           bool               `bson:"is_default" json:"is_default" db:"is_default"`

	CreatedAt           time.Time         `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt           time.Time         `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (Currency) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("code", schema.Required(), schema.Length(3),
			schema.HelpText("Currency code (ISO 4217)")),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Currency name")),
		schema.StringField("symbol", schema.Required(), schema.MaxLength(5),
			schema.HelpText("Currency symbol")),

		// Formatting
		schema.IntField("decimal_places", schema.Required(), schema.Default(2),
			schema.HelpText("Number of decimal places")),
		schema.StringField("decimal_separator", schema.Default("."),
			schema.HelpText("Decimal separator: . or ,")),
		schema.StringField("thousand_separator", schema.Default(","),
			schema.HelpText("Thousand separator: , . or space")),
		schema.StringField("symbol_position",
			schema.HelpText("Symbol position: before, after, with_space, without_space")),

		// Exchange
		schema.BoolField("is_base_currency", schema.Default(false)),
		schema.Float64Field("exchange_rate", schema.Default(1.0),
			schema.HelpText("Exchange rate relative to base currency")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_default", schema.Default(false)),

		// Timestamps
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
			schema.IndexOn("idx_currency_base", "is_base_currency"),
		},
	}
}

func (Currency) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Currency) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Ensure only one base currency
			// Ensure only one default currency
			return nil
		},
	}
}

// ExchangeRate represents an exchange rate between currencies
type ExchangeRate struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id" db:"_id"`
	FromCurrency    string             `bson:"from_currency" json:"from_currency" db:"from_currency"`
	ToCurrency      string             `bson:"to_currency" json:"to_currency" db:"to_currency"`
	Rate            float64            `bson:"rate" json:"rate" db:"rate"`

	// Dates
	EffectiveFrom   time.Time          `bson:"effective_from" json:"effective_from" db:"effective_from"`
	EffectiveTo     *time.Time         `bson:"effective_to,omitempty" json:"effective_to,omitempty" db:"effective_to"`

	// Provider
	Provider        string             `bson:"provider" json:"provider" db:"provider"`

	CreatedAt       time.Time          `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (ExchangeRate) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("from_currency", schema.Required(), schema.Length(3),
			schema.HelpText("Source currency code")),
		schema.StringField("to_currency", schema.Required(), schema.Length(3),
			schema.HelpText("Target currency code")),
		schema.Float64Field("rate", schema.Required(),
			schema.HelpText("Exchange rate")),

		// Dates
		schema.TimeField("effective_from", schema.Required(),
			schema.HelpText("Start date for this rate")),
		schema.TimeField("effective_to", schema.Optional(),
			schema.HelpText("End date for this rate (null = current)")),

		// Provider
		schema.StringField("provider", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Rate provider name")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ExchangeRate) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "exchange_rates",
		VerboseName:       "Exchange Rate",
		VerboseNamePlural: "Exchange Rates",
		OrderBy:           []string{"from_currency", "to_currency", "-effective_from"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_exchange_from", "from_currency"),
			schema.IndexOn("idx_exchange_to", "to_currency"),
			schema.IndexOn("idx_exchange_date", "effective_from"),
			schema.IndexOn("idx_exchange_pair", []string{"from_currency", "to_currency"}),
		},
		UniqueTogether: [][]string{
			{"from_currency", "to_currency", "effective_from"},
		},
	}
}

func (ExchangeRate) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ExchangeRate) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate date ranges
			// Ensure no overlapping rates for same currency pair
			return nil
		},
	}
}


import (
	"context"
	"time"

	"github.com/forgego/forge/registry"
	"github.com/forgego/forge/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShippingMethod represents a shipping method configuration
type ShippingMethod struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id" db:"_id"`
	Name            string                 `bson:"name" json:"name" db:"name"`
	Code            string                 `bson:"code" json:"code" db:"code"`
	Description     string                 `bson:"description" json:"description" db:"description"`
	Carrier         string                 `bson:"carrier" json:"carrier" db:"carrier"`
	ServiceLevel    string                 `bson:"service_level" json:"service_level" db:"service_level"`

	// Pricing
	BasePrice           float64              `bson:"base_price" json:"base_price" db:"base_price"`
	HandlingFee         float64              `bson:"handling_fee" json:"handling_fee" db:"handling_fee"`
	FreeShippingThreshold float64            `bson:"free_shipping_threshold" json:"free_shipping_threshold" db:"free_shipping_threshold"`

	// Weight constraints
	MinWeight       float64              `bson:"min_weight" json:"min_weight" db:"min_weight"`
	MaxWeight       float64              `bson:"max_weight" json:"max_weight" db:"max_weight"`
	WeightUnit      string               `bson:"weight_unit" json:"weight_unit" db:"weight_unit"`

	// Dimensions
	MinLength       float64              `bson:"min_length" json:"min_length" db:"min_length"`
	MaxLength       float64              `bson:"max_length" json:"max_length" db:"max_length"`
	MinWidth        float64              `bson:"min_width" json:"min_width" db:"min_width"`
	MaxWidth        float64              `bson:"max_width" json:"max_width" db:"max_width"`
	MinHeight       float64              `bson:"min_height" json:"min_height" db:"min_height"`
	MaxHeight       float64              `bson:"max_height" json:"max_height" db:"max_height"`
	DimensionUnit   string               `bson:"dimension_unit" json:"dimension_unit" db:"dimension_unit"`

	// Estimated delivery
	MinDays         int                  `bson:"min_days" json:"min_days" db:"min_days"`
	MaxDays         int                  `bson:"max_days" json:"max_days" db:"max_days"`

	// Geographic restrictions
	Countries       []string             `bson:"countries" json:"countries" db:"countries"`
	ExcludedCountries []string           `bson:"excluded_countries" json:"excluded_countries" db:"excluded_countries"`

	// Status
	IsActive        bool                 `bson:"is_active" json:"is_active" db:"is_active"`
	IsDefault       bool                 `bson:"is_default" json:"is_default" db:"is_default"`
	Priority        int                  `bson:"priority" json:"priority" db:"priority"`

	// Metadata
	Config          map[string]interface{} `bson:"config" json:"config" db:"config"`
	CreatedAt       time.Time             `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time             `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (ShippingMethod) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Shipping method name")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique code for the shipping method")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional()),
		schema.StringField("carrier", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Carrier name")),
		schema.StringField("service_level", schema.Required(), schema.MaxLength(50),
			schema.HelpText("Service level (e.g., express, standard)")),

		// Pricing
		schema.Float64Field("base_price", schema.Default(0.0),
			schema.HelpText("Base shipping price")),
		schema.Float64Field("handling_fee", schema.Default(0.0),
			schema.HelpText("Additional handling fee")),
		schema.Float64Field("free_shipping_threshold", schema.Default(0.0),
			schema.HelpText("Order total for free shipping")),

		// Weight constraints
		schema.Float64Field("min_weight", schema.Default(0.0),
			schema.HelpText("Minimum weight allowed")),
		schema.Float64Field("max_weight", schema.Default(0.0),
			schema.HelpText("Maximum weight allowed")),
		schema.StringField("weight_unit", schema.Default("kg"),
			schema.HelpText("Weight unit: kg, lb, oz, g")),

		// Dimensions
		schema.Float64Field("min_length", schema.Default(0.0)),
		schema.Float64Field("max_length", schema.Default(0.0)),
		schema.Float64Field("min_width", schema.Default(0.0)),
		schema.Float64Field("max_width", schema.Default(0.0)),
		schema.Float64Field("min_height", schema.Default(0.0)),
		schema.Float64Field("max_height", schema.Default(0.0)),
		schema.StringField("dimension_unit", schema.Default("cm"),
			schema.HelpText("Dimension unit: cm, in, m")),

		// Estimated delivery
		schema.IntField("min_days", schema.Default(0),
			schema.HelpText("Minimum delivery days")),
		schema.IntField("max_days", schema.Default(0),
			schema.HelpText("Maximum delivery days")),

		// Geographic restrictions
		schema.StringArrayField("countries", schema.Optional(),
			schema.HelpText("Allowed countries (empty = all)")),
		schema.StringArrayField("excluded_countries", schema.Optional(),
			schema.HelpText("Excluded countries")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_default", schema.Default(false)),
		schema.IntField("priority", schema.Default(0),
			schema.HelpText("Display priority (higher = first)")),

		// Metadata
		schema.MapField("config", schema.Optional()),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ShippingMethod) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "shipping_methods",
		VerboseName:       "Shipping Method",
		VerboseNamePlural: "Shipping Methods",
		OrderBy:           []string{"priority", "name"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_shipping_code", "code"),
			schema.IndexOn("idx_shipping_carrier", "carrier"),
			schema.IndexOn("idx_shipping_active", "is_active"),
			schema.IndexOn("idx_shipping_default", "is_default"),
		},
	}
}

func (ShippingMethod) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ShippingMethod) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate weight and dimension constraints
			// Ensure only one default shipping method
			return nil
		},
	}
}

// PaymentMethod represents a payment method configuration
type PaymentMethod struct {
	ID            primitive.ObjectID        `bson:"_id,omitempty" json:"id" db:"_id"`
	Name          string                    `bson:"name" json:"name" db:"name"`
	Code          string                    `bson:"code" json:"code" db:"code"`
	Type          string                    `bson:"type" json:"type" db:"type"`
	Description   string                    `bson:"description" json:"description" db:"description"`

	// Gateway configuration
	Gateway       string                    `bson:"gateway" json:"gateway" db:"gateway"`
	GatewayConfig map[string]interface{}     `bson:"gateway_config" json:"gateway_config" db:"gateway_config"`

	// Fees
	FixedFee      float64                   `bson:"fixed_fee" json:"fixed_fee" db:"fixed_fee"`
	PercentageFee float64                   `bson:"percentage_fee" json:"percentage_fee" db:"percentage_fee"`

	// Limits
	MinAmount     float64                   `bson:"min_amount" json:"min_amount" db:"min_amount"`
	MaxAmount     float64                   `bson:"max_amount" json:"max_amount" db:"max_amount"`

	// Currencies
	Currencies    []string                  `bson:"currencies" json:"currencies" db:"currencies"`

	// UI
	Icon          string                    `bson:"icon" json:"icon" db:"icon"`
	DisplayOrder  int                       `bson:"display_order" json:"display_order" db:"display_order"`

	// Status
	IsActive      bool                      `bson:"is_active" json:"is_active" db:"is_active"`
	IsDefault     bool                      `bson:"is_default" json:"is_default" db:"is_default"`

	// Test mode
	TestMode      bool                      `bson:"test_mode" json:"test_mode" db:"test_mode"`

	CreatedAt     time.Time                 `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt     time.Time                 `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (PaymentMethod) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Payment method name")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique code for the payment method")),
		schema.StringField("type", schema.Required(),
			schema.HelpText("Type: credit_card, debit_card, paypal, bank_transfer, cash_on_delivery, crypto, wallet")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional()),

		// Gateway configuration
		schema.StringField("gateway", schema.Required(), schema.MaxLength(50),
			schema.HelpText("Payment gateway name")),
		schema.MapField("gateway_config", schema.Optional(),
			schema.HelpText("Gateway-specific configuration")),

		// Fees
		schema.Float64Field("fixed_fee", schema.Default(0.0),
			schema.HelpText("Fixed fee per transaction")),
		schema.Float64Field("percentage_fee", schema.Default(0.0),
			schema.HelpText("Percentage fee (0-100)")),

		// Limits
		schema.Float64Field("min_amount", schema.Default(0.0),
			schema.HelpText("Minimum transaction amount")),
		schema.Float64Field("max_amount", schema.Default(0.0),
			schema.HelpText("Maximum transaction amount (0 = unlimited)")),

		// Currencies
		schema.StringArrayField("currencies", schema.Optional(),
			schema.HelpText("Supported currencies")),

		// UI
		schema.StringField("icon", schema.MaxLength(255), schema.Optional(),
			schema.HelpText("Icon URL or name")),
		schema.IntField("display_order", schema.Default(0),
			schema.HelpText("Display order")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_default", schema.Default(false)),
		schema.BoolField("test_mode", schema.Default(false)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (PaymentMethod) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "payment_methods",
		VerboseName:       "Payment Method",
		VerboseNamePlural: "Payment Methods",
		OrderBy:           []string{"display_order", "name"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_payment_code", "code"),
			schema.IndexOn("idx_payment_type", "type"),
			schema.IndexOn("idx_payment_active", "is_active"),
			schema.IndexOn("idx_payment_default", "is_default"),
		},
	}
}

func (PaymentMethod) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (PaymentMethod) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate fee percentages
			// Ensure only one default payment method
			return nil
		},
	}
}

// TaxRate represents a tax rate configuration
type TaxRate struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id" db:"_id"`
	Name            string             `bson:"name" json:"name" db:"name"`
	Code            string             `bson:"code" json:"code" db:"code"`
	Description     string             `bson:"description" json:"description" db:"description"`

	// Rate
	Rate            float64            `bson:"rate" json:"rate" db:"rate"`

	// Geographic scope
	Country         string             `bson:"country" json:"country" db:"country"`
	State           string             `bson:"state" json:"state" db:"state"`
	ZipPattern      string             `bson:"zip_pattern" json:"zip_pattern" db:"zip_pattern"`
	City            string             `bson:"city" json:"city" db:"city"`

	// Tax type
	TaxType         string             `bson:"tax_type" json:"tax_type" db:"tax_type"`

	// Product/category applicability
	AppliesToProducts bool             `bson:"applies_to_products" json:"applies_to_products" db:"applies_to_products"`
	AppliesToShipping bool             `bson:"applies_to_shipping" json:"applies_to_shipping" db:"applies_to_shipping"`
	AppliesToServices bool             `bson:"applies_to_services" json:"applies_to_services" db:"applies_to_services"`

	// Dates
	StartDate       *time.Time         `bson:"start_date,omitempty" json:"start_date,omitempty" db:"start_date"`
	EndDate         *time.Time         `bson:"end_date,omitempty" json:"end_date,omitempty" db:"end_date"`

	// Status
	IsActive        bool               `bson:"is_active" json:"is_active" db:"is_active"`
	IsCompound      bool               `bson:"is_compound" json:"is_compound" db:"is_compound"`
	Priority        int                `bson:"priority" json:"priority" db:"priority"`

	CreatedAt       time.Time          `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (TaxRate) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Tax rate name")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique code for the tax rate")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional()),

		// Rate
		schema.Float64Field("rate", schema.Required(),
			schema.HelpText("Tax rate percentage (0-100)")),

		// Geographic scope
		schema.StringField("country", schema.Required(), schema.Length(2),
			schema.HelpText("Country code (ISO 2)")),
		schema.StringField("state", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("State/Province (optional)")),
		schema.StringField("zip_pattern", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("ZIP/Postal code pattern for matching")),
		schema.StringField("city", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("City (optional)")),

		// Tax type
		schema.StringField("tax_type",
			schema.HelpText("Type: sales, vat, gst, excised")),

		// Product/category applicability
		schema.BoolField("applies_to_products", schema.Default(true)),
		schema.BoolField("applies_to_shipping", schema.Default(true)),
		schema.BoolField("applies_to_services", schema.Default(false)),

		// Dates
		schema.TimeField("start_date", schema.Optional()),
		schema.TimeField("end_date", schema.Optional()),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_compound", schema.Default(false),
			schema.HelpText("Applied after other taxes")),
		schema.IntField("priority", schema.Default(0),
			schema.HelpText("Priority for compound tax calculation")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (TaxRate) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "tax_rates",
		VerboseName:       "Tax Rate",
		VerboseNamePlural: "Tax Rates",
		OrderBy:           []string{"country", "state", "priority"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_tax_code", "code"),
			schema.IndexOn("idx_tax_country", "country"),
			schema.IndexOn("idx_tax_state", "state"),
			schema.IndexOn("idx_tax_active", "is_active"),
			schema.IndexOn("idx_tax_compound", "is_compound"),
		},
	}
}

func (TaxRate) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (TaxRate) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate rate is within bounds
			// Validate date ranges
			return nil
		},
	}
}

// Currency represents a currency configuration
type Currency struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id" db:"_id"`
	Code            string             `bson:"code" json:"code" db:"code"`
	Name            string             `bson:"name" json:"name" db:"name"`
	Symbol          string             `bson:"symbol" json:"symbol" db:"symbol"`

	// Formatting
	DecimalPlaces   int                `bson:"decimal_places" json:"decimal_places" db:"decimal_places"`
	DecimalSeparator string            `bson:"decimal_separator" json:"decimal_separator" db:"decimal_separator"`
	ThousandSeparator string           `bson:"thousand_separator" json:"thousand_separator" db:"thousand_separator"`
	SymbolPosition  string             `bson:"symbol_position" json:"symbol_position" db:"symbol_position"`

	// Exchange
	IsBaseCurrency  bool               `bson:"is_base_currency" json:"is_base_currency" db:"is_base_currency"`
	ExchangeRate    float64            `bson:"exchange_rate" json:"exchange_rate" db:"exchange_rate"`

	// Status
	IsActive        bool               `bson:"is_active" json:"is_active" db:"is_active"`
	IsDefault       bool               `bson:"is_default" json:"is_default" db:"is_default"`

	CreatedAt       time.Time         `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (Currency) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("code", schema.Required(), schema.Length(3),
			schema.HelpText("Currency code (ISO 4217)")),
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Currency name")),
		schema.StringField("symbol", schema.Required(), schema.MaxLength(5),
			schema.HelpText("Currency symbol")),

		// Formatting
		schema.IntField("decimal_places", schema.Required(), schema.Default(2),
			schema.HelpText("Number of decimal places")),
		schema.StringField("decimal_separator", schema.Default("."),
			schema.HelpText("Decimal separator: . or ,")),
		schema.StringField("thousand_separator", schema.Default(","),
			schema.HelpText("Thousand separator: , . or space")),
		schema.StringField("symbol_position",
			schema.HelpText("Symbol position: before, after, with_space, without_space")),

		// Exchange
		schema.BoolField("is_base_currency", schema.Default(false)),
		schema.Float64Field("exchange_rate", schema.Default(1.0),
			schema.HelpText("Exchange rate relative to base currency")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_default", schema.Default(false)),

		// Timestamps
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
			schema.IndexOn("idx_currency_base", "is_base_currency"),
		},
	}
}

func (Currency) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Currency) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Ensure only one base currency
			// Ensure only one default currency
			return nil
		},
	}
}

// ExchangeRate represents an exchange rate between currencies
type ExchangeRate struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id" db:"_id"`
	FromCurrency    string             `bson:"from_currency" json:"from_currency" db:"from_currency"`
	ToCurrency      string             `bson:"to_currency" json:"to_currency" db:"to_currency"`
	Rate            float64            `bson:"rate" json:"rate" db:"rate"`

	// Dates
	EffectiveFrom   time.Time          `bson:"effective_from" json:"effective_from" db:"effective_from"`
	EffectiveTo     *time.Time         `bson:"effective_to,omitempty" json:"effective_to,omitempty" db:"effective_to"`

	// Provider
	Provider        string             `bson:"provider" json:"provider" db:"provider"`

	CreatedAt       time.Time          `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (ExchangeRate) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("from_currency", schema.Required(), schema.Length(3),
			schema.HelpText("Source currency code")),
		schema.StringField("to_currency", schema.Required(), schema.Length(3),
			schema.HelpText("Target currency code")),
		schema.Float64Field("rate", schema.Required(),
			schema.HelpText("Exchange rate")),

		// Dates
		schema.TimeField("effective_from", schema.Required(),
			schema.HelpText("Start date for this rate")),
		schema.TimeField("effective_to", schema.Optional(),
			schema.HelpText("End date for this rate (null = current)")),

		// Provider
		schema.StringField("provider", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Rate provider name")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ExchangeRate) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "exchange_rates",
		VerboseName:       "Exchange Rate",
		VerboseNamePlural: "Exchange Rates",
		OrderBy:           []string{"from_currency", "to_currency", "-effective_from"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_exchange_from", "from_currency"),
			schema.IndexOn("idx_exchange_to", "to_currency"),
			schema.IndexOn("idx_exchange_date", "effective_from"),
			schema.IndexOn("idx_exchange_pair", []string{"from_currency", "to_currency"}),
		},
		UniqueTogether: [][]string{
			{"from_currency", "to_currency", "effective_from"},
		},
	}
}

func (ExchangeRate) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ExchangeRate) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate date ranges
			// Ensure no overlapping rates for same currency pair
			return nil
		},
	}
}

