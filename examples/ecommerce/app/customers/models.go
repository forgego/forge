package customers

import (
	"context"

	"github.com/forgego/forge/registry"
	"github.com/forgego/forge/schema"
)

// CustomerGroup represents customer segmentation for pricing/promotions
type CustomerGroup struct {
	CustomerGroupGenerated
}

func (CustomerGroup) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(200), schema.Unique(),
			schema.HelpText("Group name (e.g., 'VIP', 'Wholesale')")),
		schema.StringField("code", schema.Required(), schema.MaxLength(50), schema.Unique()),
		schema.TextField("description", schema.Optional()),

		// Pricing
		schema.Float64Field("discount_percentage", schema.Default(0.0),
			schema.VerboseName("Discount %"),
			schema.HelpText("Default discount percentage for this group")),

		// Status
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (CustomerGroup) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "customer_groups",
		VerboseName:       "Customer Group",
		VerboseNamePlural: "Customer Groups",
		OrderBy:           []string{"name"},
	}
}

func (CustomerGroup) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (CustomerGroup) Hooks() *schema.ModelHooks {
	return nil
}

// Customer represents a customer account
type Customer struct {
	CustomerGenerated
}

func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),

		// Authentication (if using built-in auth)
		schema.StringField("email", schema.Required(), schema.MaxLength(255), schema.Unique(),
			schema.HelpText("Customer email address")),
		schema.StringField("password_hash", schema.Required(), schema.MaxLength(255),
			schema.HelpText("Hashed password")),

		// Personal information
		schema.StringField("first_name", schema.Required(), schema.MaxLength(100),
			schema.VerboseName("First Name")),
		schema.StringField("last_name", schema.Required(), schema.MaxLength(100),
			schema.VerboseName("Last Name")),
		schema.StringField("phone", schema.MaxLength(20), schema.Optional()),
		schema.DateField("date_of_birth", schema.Optional()),
		schema.StringField("gender", schema.MaxLength(20), schema.Optional(),
			schema.HelpText("Gender: male, female, other, prefer_not_to_say")),

		// Business (for B2B)
		schema.StringField("company_name", schema.MaxLength(200), schema.Optional()),
		schema.StringField("tax_id", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Tax ID / VAT number")),

		// Customer group
		schema.Int64Field("customer_group_id", schema.Optional()),

		// Status
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active"),
			schema.HelpText("Is account active")),
		schema.BoolField("is_verified", schema.Default(false),
			schema.VerboseName("Verified"),
			schema.HelpText("Is email verified")),
		schema.BoolField("accepts_marketing", schema.Default(false),
			schema.HelpText("Accepts marketing emails")),

		// Account security
		schema.StringField("verification_token", schema.MaxLength(255), schema.Optional()),
		schema.StringField("reset_password_token", schema.MaxLength(255), schema.Optional()),
		schema.TimeField("reset_password_expires_at", schema.Optional()),
		schema.TimeField("last_login_at", schema.Optional()),
		schema.StringField("last_login_ip", schema.MaxLength(45), schema.Optional()),

		// Statistics
		schema.Int32Field("total_orders", schema.Default(0),
			schema.HelpText("Total number of orders")),
		schema.Float64Field("total_spent", schema.Default(0.0),
			schema.HelpText("Total amount spent")),
		schema.Float64Field("average_order_value", schema.Default(0.0)),

		// Preferences
		schema.StringField("preferred_language", schema.MaxLength(10), schema.Default("en")),
		schema.StringField("preferred_currency", schema.MaxLength(3), schema.Default("USD")),

		// Notes
		schema.TextField("notes", schema.Optional(),
			schema.HelpText("Admin notes about customer")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Customer) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "customers",
		VerboseName:       "Customer",
		VerboseNamePlural: "Customers",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_customer_email", "email"),
			schema.IndexOn("idx_customer_group", "customer_group_id"),
			schema.IndexOn("idx_customer_active", "is_active"),
			schema.IndexOn("idx_customer_verified", "is_verified"),
		},
	}
}

func (Customer) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_group_id", "CustomerGroup",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("customers")),
	}
}

func (Customer) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeCreate: func(ctx context.Context, instance interface{}) error {
			// Hash password before storing
			// Generate verification token
			return nil
		},
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate email format
			// Validate phone format
			return nil
		},
	}
}

// Address represents shipping/billing addresses
type Address struct {
	AddressGenerated
}

func (Address) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),

		// Address type
		schema.StringField("address_type", schema.Required(), schema.MaxLength(20),
			schema.VerboseName("Address Type"),
			schema.HelpText("Type: shipping, billing, both")),

		// Contact
		schema.StringField("first_name", schema.Required(), schema.MaxLength(100)),
		schema.StringField("last_name", schema.Required(), schema.MaxLength(100)),
		schema.StringField("company_name", schema.MaxLength(200), schema.Optional()),
		schema.StringField("phone", schema.MaxLength(20), schema.Optional()),

		// Address
		schema.StringField("address_line1", schema.Required(), schema.MaxLength(255),
			schema.HelpText("Street address, P.O. box")),
		schema.StringField("address_line2", schema.MaxLength(255), schema.Optional(),
			schema.HelpText("Apartment, suite, unit, building, floor, etc.")),
		schema.StringField("city", schema.Required(), schema.MaxLength(100)),
		schema.StringField("state_province", schema.MaxLength(100), schema.Optional(),
			schema.VerboseName("State/Province"),
			schema.HelpText("State, province, region")),
		schema.StringField("postal_code", schema.Required(), schema.MaxLength(20)),
		schema.StringField("country_code", schema.Required(), schema.MaxLength(2),
			schema.VerboseName("Country Code"),
			schema.HelpText("ISO 3166-1 alpha-2 country code")),
		schema.StringField("country_name", schema.Required(), schema.MaxLength(100)),

		// Geolocation (optional)
		schema.Float64Field("latitude", schema.Optional()),
		schema.Float64Field("longitude", schema.Optional()),

		// Preferences
		schema.BoolField("is_default_shipping", schema.Default(false),
			schema.VerboseName("Default Shipping")),
		schema.BoolField("is_default_billing", schema.Default(false),
			schema.VerboseName("Default Billing")),

		// Special instructions
		schema.TextField("delivery_instructions", schema.Optional(),
			schema.HelpText("Delivery notes, gate codes, etc.")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Address) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "addresses",
		VerboseName:       "Address",
		VerboseNamePlural: "Addresses",
		OrderBy:           []string{"-is_default_shipping", "-is_default_billing", "-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_address_customer", "customer_id"),
			schema.IndexOn("idx_address_type", "address_type"),
			schema.IndexOn("idx_address_country", "country_code"),
		},
	}
}

func (Address) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("addresses")),
	}
}

func (Address) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Ensure only one default shipping address per customer
			// Ensure only one default billing address per customer
			// Validate country code
			// Validate postal code format
			return nil
		},
	}
}

// WishList represents customer wish lists
type WishList struct {
	WishListGenerated
}

func (WishList) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),

		schema.StringField("name", schema.Required(), schema.MaxLength(200),
			schema.HelpText("Wish list name (e.g., 'Birthday Gifts')")),
		schema.TextField("description", schema.Optional()),

		// Visibility
		schema.BoolField("is_public", schema.Default(false),
			schema.VerboseName("Public"),
			schema.HelpText("Share with others")),
		schema.StringField("share_token", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("Token for sharing wish list")),

		// Status
		schema.BoolField("is_default", schema.Default(false),
			schema.VerboseName("Default")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (WishList) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "wish_lists",
		VerboseName:       "Wish List",
		VerboseNamePlural: "Wish Lists",
		OrderBy:           []string{"-is_default", "-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_wishlist_customer", "customer_id"),
			schema.IndexOn("idx_wishlist_token", "share_token"),
		},
	}
}

func (WishList) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("wish_lists")),
	}
}

func (WishList) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeCreate: func(ctx context.Context, instance interface{}) error {
			// Generate share token if public
			return nil
		},
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Ensure only one default wish list per customer
			return nil
		},
	}
}

// WishListItem represents items in a wish list
type WishListItem struct {
	WishListItemGenerated
}

func (WishListItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("wish_list_id", schema.Required()),
		schema.Int64Field("product_id", schema.Required()),
		schema.Int64Field("variant_id", schema.Optional(),
			schema.HelpText("Specific product variant")),

		schema.Int32Field("desired_quantity", schema.Default(1)),
		schema.Float64Field("price_when_added", schema.Optional(),
			schema.HelpText("Price when added to wish list")),

		schema.TextField("notes", schema.Optional(),
			schema.HelpText("Personal notes about the item")),
		schema.Int32Field("priority", schema.Default(0),
			schema.HelpText("Priority: 0=normal, 1=high, etc.")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (WishListItem) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "wish_list_items",
		VerboseName:       "Wish List Item",
		VerboseNamePlural: "Wish List Items",
		OrderBy:           []string{"-priority", "-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_wishlist_item_list", "wish_list_id"),
			schema.IndexOn("idx_wishlist_item_product", "product_id"),
			schema.IndexOn("idx_wishlist_item_variant", "variant_id"),
		},
		UniqueTogether: [][]string{
			{"wish_list_id", "product_id", "variant_id"},
		},
	}
}

func (WishListItem) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("wish_list_id", "WishList",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("items")),
		schema.ForeignKeyField("product_id", "Product",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("wish_list_items")),
		schema.ForeignKeyField("variant_id", "ProductVariant",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("wish_list_items")),
	}
}

func (WishListItem) Hooks() *schema.ModelHooks {
	return nil
}

// RegisterModels registers customer models with the framework
func RegisterModels() {
	registry.RegisterModel(&CustomerGroup{})
	registry.RegisterModel(&Customer{})
	registry.RegisterModel(&Address{})
	registry.RegisterModel(&WishList{})
	registry.RegisterModel(&WishListItem{})
}
