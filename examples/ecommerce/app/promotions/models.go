package promotions

import (
	"context"
	"github.com/forgego/forge/schema"
)

// Promotion represents a promotional campaign
type Promotion struct {
	schema.BaseSchema
	Id               int64   `json:"id" db:"id"`
	Name             string  `json:"name" db:"name"`
	Code             string  `json:"code" db:"code"`
	Description      string  `json:"description" db:"description"`
	DiscountType     string  `json:"discount_type" db:"discount_type"`
	DiscountValue    float64 `json:"discount_value" db:"discount_value"`
	MinPurchase      float64 `json:"min_purchase" db:"min_purchase"`
	MaxDiscount      float64 `json:"max_discount" db:"max_discount"`
	StartDate        string  `json:"start_date" db:"start_date"`
	EndDate          string  `json:"end_date" db:"end_date"`
	UsageLimit       int32   `json:"usage_limit" db:"usage_limit"`
	UsageCount       int32   `json:"usage_count" db:"usage_count"`
	PerCustomerLimit int32   `json:"per_customer_limit" db:"per_customer_limit"`
	IsActive         bool    `json:"is_active" db:"is_active"`
	IsStackable      bool    `json:"is_stackable" db:"is_stackable"`
	Priority         int32   `json:"priority" db:"priority"`
	AppliesTo        string  `json:"applies_to" db:"applies_to"`
	TargetEntityIDs  string  `json:"target_entity_ids" db:"target_entity_ids"`
	CreatedAt        string  `json:"created_at" db:"created_at"`
	UpdatedAt        string  `json:"updated_at" db:"updated_at"`
}

func (Promotion) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(200),
			schema.HelpText("Promotion name")),
		schema.StringField("code", schema.MaxLength(50), schema.Optional(), schema.Unique(),
			schema.HelpText("Promotional code (leave empty for auto-apply)")),
		schema.TextField("description", schema.Optional(),
			schema.HelpText("Promotion description")),
		schema.StringField("discount_type", schema.MaxLength(20), schema.Required(),
			schema.VerboseName("Discount Type"),
			schema.HelpText("Type: percentage, fixed_amount, free_shipping, buy_x_get_y")),
		schema.FloatField("discount_value", schema.Required(),
			schema.VerboseName("Discount Value"),
			schema.HelpText("Percentage (e.g., 15 for 15%) or fixed amount")),
		schema.FloatField("min_purchase", schema.Default(0.0),
			schema.VerboseName("Minimum Purchase"),
			schema.HelpText("Minimum purchase amount to qualify")),
		schema.FloatField("max_discount", schema.Default(0.0),
			schema.VerboseName("Maximum Discount"),
			schema.HelpText("Maximum discount amount (0 = unlimited)")),
		schema.DateTimeField("start_date", schema.Required(),
			schema.VerboseName("Start Date"),
			schema.HelpText("Promotion start date and time")),
		schema.DateTimeField("end_date", schema.Optional(),
			schema.VerboseName("End Date"),
			schema.HelpText("Promotion end date and time (null = no expiration)")),
		schema.Int32Field("usage_limit", schema.Default(0),
			schema.VerboseName("Usage Limit"),
			schema.HelpText("Total usage limit (0 = unlimited)")),
		schema.Int32Field("usage_count", schema.Default(0),
			schema.VerboseName("Usage Count"),
			schema.HelpText("Current usage count")),
		schema.Int32Field("per_customer_limit", schema.Default(0),
			schema.VerboseName("Per Customer Limit"),
			schema.HelpText("Usage limit per customer (0 = unlimited)")),
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.BoolField("is_stackable", schema.Default(false),
			schema.VerboseName("Stackable"),
			schema.HelpText("Can be combined with other promotions")),
		schema.Int32Field("priority", schema.Default(0),
			schema.HelpText("Application priority for non-stackable promotions")),
		schema.StringField("applies_to", schema.MaxLength(50), schema.Default("all"),
			schema.VerboseName("Applies To"),
			schema.HelpText("Scope: all, categories, products, customer_groups")),
		schema.TextField("target_entity_ids", schema.Optional(),
			schema.VerboseName("Target Entity IDs"),
			schema.HelpText("Comma-separated IDs of target entities")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Promotion) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "promotions",
		VerboseName:       "Promotion",
		VerboseNamePlural: "Promotions",
		OrderBy:           []string{"-priority", "-created_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_promotion_code", "code"),
			schema.IndexOn("idx_promotion_active", "is_active"),
			schema.IndexOn("idx_promotion_dates", "start_date", "end_date"),
		},
	}
}

func (Promotion) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Promotion) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate discount type and value
			// Generate code if not provided
			return nil
		},
	}
}

// PromotionRule represents complex promotion rules
type PromotionRule struct {
	schema.BaseSchema
	Id          int64  `json:"id" db:"id"`
	PromotionID int64  `json:"promotion_id" db:"promotion_id"`
	RuleType    string `json:"rule_type" db:"rule_type"`
	Field       string `json:"field" db:"field"`
	Operator    string `json:"operator" db:"operator"`
	Value       string `json:"value" db:"value"`
	LogicType   string `json:"logic_type" db:"logic_type"`
	SortOrder   int32  `json:"sort_order" db:"sort_order"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

func (PromotionRule) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("promotion_id", schema.Required(),
			schema.VerboseName("Promotion")),
		schema.StringField("rule_type", schema.MaxLength(50), schema.Required(),
			schema.VerboseName("Rule Type"),
			schema.HelpText("Type: eligibility, action, exclusion")),
		schema.StringField("field", schema.MaxLength(100), schema.Required(),
			schema.HelpText("Field to evaluate (e.g., cart_total, product_category)")),
		schema.StringField("operator", schema.MaxLength(20), schema.Required(),
			schema.HelpText("Operator: equals, not_equals, greater_than, less_than, in, contains")),
		schema.StringField("value", schema.MaxLength(500), schema.Required(),
			schema.HelpText("Value to compare against")),
		schema.StringField("logic_type", schema.MaxLength(10), schema.Default("AND"),
			schema.VerboseName("Logic Type"),
			schema.HelpText("Logic: AND, OR")),
		schema.Int32Field("sort_order", schema.Default(0),
			schema.VerboseName("Sort Order")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (PromotionRule) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "promotion_rules",
		VerboseName:       "Promotion Rule",
		VerboseNamePlural: "Promotion Rules",
		OrderBy:           []string{"promotion_id", "sort_order"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_promotion_rule_promotion", "promotion_id"),
		},
	}
}

func (PromotionRule) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("promotion_id", "Promotion",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("rules")),
	}
}

func (PromotionRule) Hooks() *schema.ModelHooks {
	return nil
}

// Banner represents marketing banners
type Banner struct {
	schema.BaseSchema
	Id          int64  `json:"id" db:"id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	ImageURL    string `json:"image_url" db:"image_url"`
	LinkURL     string `json:"link_url" db:"link_url"`
	Placement   string `json:"placement" db:"placement"`
	StartDate   string `json:"start_date" db:"start_date"`
	EndDate     string `json:"end_date" db:"end_date"`
	IsActive    bool   `json:"is_active" db:"is_active"`
	SortOrder   int32  `json:"sort_order" db:"sort_order"`
	ClickCount  int32  `json:"click_count" db:"click_count"`
	ViewCount   int32  `json:"view_count" db:"view_count"`
	TargetGroup string `json:"target_group" db:"target_group"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

func (Banner) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required(), schema.MaxLength(200),
			schema.HelpText("Banner title")),
		schema.TextField("description", schema.Optional(),
			schema.HelpText("Banner description")),
		schema.StringField("image_url", schema.MaxLength(500), schema.Required(),
			schema.VerboseName("Image URL"),
			schema.HelpText("Banner image URL")),
		schema.StringField("link_url", schema.MaxLength(500), schema.Optional(),
			schema.VerboseName("Link URL"),
			schema.HelpText("Click destination URL")),
		schema.StringField("placement", schema.MaxLength(50), schema.Default("home_hero"),
			schema.HelpText("Placement: home_hero, home_sidebar, category_top, product_related")),
		schema.DateTimeField("start_date", schema.Required(),
			schema.VerboseName("Start Date")),
		schema.DateTimeField("end_date", schema.Optional(),
			schema.VerboseName("End Date"),
			schema.HelpText("null = no expiration")),
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.Int32Field("sort_order", schema.Default(0),
			schema.VerboseName("Sort Order")),
		schema.Int32Field("click_count", schema.Default(0),
			schema.VerboseName("Click Count"),
			schema.HelpText("Number of clicks")),
		schema.Int32Field("view_count", schema.Default(0),
			schema.VerboseName("View Count"),
			schema.HelpText("Number of impressions")),
		schema.StringField("target_group", schema.MaxLength(50), schema.Default("all"),
			schema.VerboseName("Target Group"),
			schema.HelpText("Target audience: all, new_customers, vip, etc.")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Banner) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "banners",
		VerboseName:       "Banner",
		VerboseNamePlural: "Banners",
		OrderBy:           []string{"sort_order", "-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_banner_active", "is_active"),
			schema.IndexOn("idx_banner_placement", "placement"),
			schema.IndexOn("idx_banner_dates", "start_date", "end_date"),
		},
	}
}

func (Banner) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Banner) Hooks() *schema.ModelHooks {
	return nil
}

// NewsletterSubscription represents newsletter subscriptions
type NewsletterSubscription struct {
	schema.BaseSchema
	Id           int64  `json:"id" db:"id"`
	Email        string `json:"email" db:"email"`
	FirstName    string `json:"first_name" db:"first_name"`
	LastName     string `json:"last_name" db:"last_name"`
	CustomerID   int64  `json:"customer_id" db:"customer_id"`
	Status       string `json:"status" db:"status"`
	Source       string `json:"source" db:"source"`
	IPAddress    string `json:"ip_address" db:"ip_address"`
	ConfirmedAt  string `json:"confirmed_at" db:"confirmed_at"`
	UnsubscribedAt string `json:"unsubscribed_at" db:"unsubscribed_at"`
	Preferences  string `json:"preferences" db:"preferences"`
	CreatedAt    string `json:"created_at" db:"created_at"`
	UpdatedAt    string `json:"updated_at" db:"updated_at"`
}

func (NewsletterSubscription) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("email", schema.Required(), schema.MaxLength(255), schema.Unique(),
			schema.HelpText("Subscriber email address")),
		schema.StringField("first_name", schema.MaxLength(100), schema.Optional(),
			schema.VerboseName("First Name")),
		schema.StringField("last_name", schema.MaxLength(100), schema.Optional(),
			schema.VerboseName("Last Name")),
		schema.Int64Field("customer_id", schema.Optional(),
			schema.VerboseName("Customer"),
			schema.HelpText("Linked customer account (if applicable)")),
		schema.StringField("status", schema.MaxLength(20), schema.Default("pending"),
			schema.HelpText("Status: pending, active, unsubscribed, bounced")),
		schema.StringField("source", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Subscription source: footer, checkout, popup, import")),
		schema.StringField("ip_address", schema.MaxLength(45), schema.Optional(),
			schema.VerboseName("IP Address")),
		schema.TimeField("confirmed_at", schema.Optional(),
			schema.VerboseName("Confirmed At"),
			schema.HelpText("Double opt-in confirmation time")),
		schema.TimeField("unsubscribed_at", schema.Optional(),
			schema.VerboseName("Unsubscribed At")),
		schema.TextField("preferences", schema.Optional(),
			schema.HelpText("Subscription preferences in JSON format")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (NewsletterSubscription) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "newsletter_subscriptions",
		VerboseName:       "Newsletter Subscription",
		VerboseNamePlural: "Newsletter Subscriptions",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_newsletter_email", "email"),
			schema.IndexOn("idx_newsletter_status", "status"),
			schema.IndexOn("idx_newsletter_customer", "customer_id"),
		},
	}
}

func (NewsletterSubscription) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("newsletter_subscriptions")),
	}
}

func (NewsletterSubscription) Hooks() *schema.ModelHooks {
	return nil
}

// PromotionUsage tracks promotion usage by customers
type PromotionUsage struct {
	schema.BaseSchema
	Id          int64  `json:"id" db:"id"`
	PromotionID int64  `json:"promotion_id" db:"promotion_id"`
	CustomerID  int64  `json:"customer_id" db:"customer_id"`
	OrderID     int64  `json:"order_id" db:"order_id"`
	UsedAt      string `json:"used_at" db:"used_at"`
	DiscountAmt float64 `json:"discount_amount" db:"discount_amount"`
	CreatedAt   string `json:"created_at" db:"created_at"`
}

func (PromotionUsage) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("promotion_id", schema.Required(),
			schema.VerboseName("Promotion")),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),
		schema.Int64Field("order_id", schema.Required(),
			schema.VerboseName("Order")),
		schema.TimeField("used_at", schema.Required(),
			schema.VerboseName("Used At")),
		schema.FloatField("discount_amount", schema.Required(),
			schema.VerboseName("Discount Amount"),
			schema.HelpText("Actual discount amount applied")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (PromotionUsage) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "promotion_usages",
		VerboseName:       "Promotion Usage",
		VerboseNamePlural: "Promotion Usages",
		OrderBy:           []string{"-used_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_promotion_usage_promotion", "promotion_id"),
			schema.IndexOn("idx_promotion_usage_customer", "customer_id"),
			schema.IndexOn("idx_promotion_usage_order", "order_id"),
		},
	}
}

func (PromotionUsage) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("promotion_id", "Promotion",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("usages")),
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("promotion_usages")),
		schema.ForeignKeyField("order_id", "Order",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("promotion_usages")),
	}
}

func (PromotionUsage) Hooks() *schema.ModelHooks {
	return nil
}
