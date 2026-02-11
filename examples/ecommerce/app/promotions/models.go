package promotions

import (
	"context"
	"time"

	"github.com/forgego/forge/registry"
	"github.com/forgego/forge/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BannerSchedule represents scheduling for banners
type BannerSchedule struct {
	DayOfWeek int    `bson:"day_of_week" json:"day_of_week"`
	StartTime string `bson:"start_time" json:"start_time"`
	EndTime   string `bson:"end_time" json:"end_time"`
}

// Promotion represents a promotional campaign
type Promotion struct {
	PromotionGenerated
}

// PromotionRule represents rules for promotions
type PromotionRule struct {
	PromotionRuleGenerated
}

// Banner represents promotional banners
type Banner struct {
	BannerGenerated
}

// NewsletterSubscription represents newsletter subscribers
type NewsletterSubscription struct {
	NewsletterSubscriptionGenerated
}

// PromotionUsage tracks promotion usage
type PromotionUsage struct {
	PromotionUsageGenerated
}

// ==================== Promotion Fields ====================

func (Promotion) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),

		// Identification
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Promotion name")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Promotion description")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique promotion code"),
			schema.Unique()),

		// Promotion type
		schema.StringField("type", schema.Required(),
			schema.HelpText("Type: percentage_fixed, percentage_off, buy_x_get_y, bundle_discount, free_shipping, flash_sale, gift_with_purchase")),
		schema.Float64Field("discount_value", schema.Default(0.0),
			schema.HelpText("Discount value")),
		schema.StringField("discount_type",
			schema.HelpText("Type: percentage, fixed")),

		// Conditions
		schema.Float64Field("min_purchase", schema.Default(0.0),
			schema.HelpText("Minimum purchase amount")),
		schema.Float64Field("max_discount", schema.Default(0.0),
			schema.HelpText("Maximum discount amount")),

		// Buy X Get Y
		schema.IntField("buy_quantity", schema.Default(0),
			schema.HelpText("Buy quantity")),
		schema.IntField("get_quantity", schema.Default(0),
			schema.HelpText("Get quantity")),
		schema.ObjectIDField("free_product_id", schema.Optional()),

		// Free shipping
		schema.BoolField("free_shipping", schema.Default(false)),

		// Applicability
		schema.StringField("applies_to",
			schema.HelpText("Type: all, products, categories, brands, specific_products")),
		schema.ObjectIDSliceField("product_ids", schema.Optional()),
		schema.ObjectIDSliceField("category_ids", schema.Optional()),
		schema.ObjectIDSliceField("brand_ids", schema.Optional()),

		// Customer restrictions
		schema.BoolField("new_customers_only", schema.Default(false)),
		schema.ObjectIDSliceField("customer_group_ids", schema.Optional()),

		// Usage limits
		schema.IntField("total_usage_limit", schema.Default(0),
			schema.HelpText("Total usage limit (0 = unlimited)")),
		schema.IntField("per_customer_limit", schema.Default(0),
			schema.HelpText("Per customer limit (0 = unlimited)")),

		// Dates
		schema.TimeField("start_date", schema.Optional()),
		schema.TimeField("end_date", schema.Optional()),

		// Priority and stacking
		schema.IntField("priority", schema.Default(0)),
		schema.BoolField("can_stack", schema.Default(false)),
		schema.ObjectIDSliceField("stack_with", schema.Optional()),

		// Status
		schema.BoolField("is_active", schema.Default(true)),

		// Stats
		schema.IntField("times_used", schema.Default(0)),

		// Timestamps
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
			schema.IndexOn("idx_promotion_type", "type"),
			schema.IndexOn("idx_promotion_dates", "start_date", "end_date"),
		},
	}
}

func (Promotion) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("free_product_id", "Product",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("free_promotions")),
		schema.ForeignKeyField("promotion_id", "PromotionRule",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("promotion")),
		schema.ForeignKeyField("promotion_id", "PromotionUsage",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("promotion")),
	}
}

func (Promotion) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate promotion type
			// Validate dates
			// Uppercase code
			return nil
		},
	}
}

// ==================== PromotionRule Fields ====================

func (PromotionRule) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("promotion_id", schema.Required(),
			schema.HelpText("Associated promotion")),

		// Rule type
		schema.StringField("rule_type", schema.Required(),
			schema.HelpText("Rule type: product_category, product_brand, product_tag, product_price, customer_group, customer_location, day_of_week, time_range, cart_total, cart_item_count, quantity_threshold, first_purchase")),

		// Rule parameters
		schema.MapField("parameters", schema.Optional(),
			schema.HelpText("Flexible JSON for different rule types")),

		// Logic
		schema.StringField("logic",
			schema.HelpText("Logic: and, or")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (PromotionRule) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "promotion_rules",
		VerboseName:       "Promotion Rule",
		VerboseNamePlural: "Promotion Rules",
		OrderBy:           []string{"promotion_id", "created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_rule_promotion", "promotion_id"),
			schema.IndexOn("idx_rule_type", "rule_type"),
			schema.IndexOn("idx_rule_active", "is_active"),
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

// ==================== Banner Fields ====================

func (Banner) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),

		schema.StringField("title", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Banner title")),
		schema.StringField("subtitle", schema.MaxLength(200), schema.Optional(),
			schema.HelpText("Banner subtitle")),

		// Media
		schema.StringField("image_url", schema.Required(), schema.MaxLength(500),
			schema.HelpText("Banner image URL")),
		schema.StringField("mobile_image_url", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Mobile-optimized image URL")),
		schema.StringField("video_url", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Video URL")),

		// Content
		schema.TextField("content", schema.Optional(),
			schema.HelpText("Banner content")),
		schema.StringField("link", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Click-through link")),
		schema.StringField("link_text", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Link button text")),

		// Placement
		schema.StringField("position", schema.Required(),
			schema.HelpText("Position: homepage_top, homepage_middle, homepage_bottom, category_page, product_page, cart_page, checkout_page")),

		// Styling
		schema.StringField("background_color", schema.MaxLength(20), schema.Optional(),
			schema.HelpText("Background color (hex)")),
		schema.StringField("text_color", schema.MaxLength(20), schema.Optional(),
			schema.HelpText("Text color (hex)")),

		// Dates
		schema.TimeField("start_date", schema.Optional()),
		schema.TimeField("end_date", schema.Optional()),

		// Scheduling
		schema.SliceField("schedule", schema.Optional(),
			schema.HelpText("Banner schedule")),

		// Targeting
		schema.StringSliceField("device_types", schema.Optional(),
			schema.HelpText("Target devices: mobile, tablet, desktop")),
		schema.StringSliceField("user_types", schema.Optional(),
			schema.HelpText("Target users: guest, logged_in")),
		schema.ObjectIDSliceField("customer_group_ids", schema.Optional()),

		// Priority
		schema.IntField("priority", schema.Default(0)),

		// Status
		schema.BoolField("is_active", schema.Default(true)),

		// Stats
		schema.IntField("click_count", schema.Default(0)),
		schema.IntField("view_count", schema.Default(0)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Banner) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "banners",
		VerboseName:       "Banner",
		VerboseNamePlural: "Banners",
		OrderBy:           []string{"-priority", "-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_banner_active", "is_active"),
			schema.IndexOn("idx_banner_position", "position"),
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

// ==================== NewsletterSubscription Fields ====================

func (NewsletterSubscription) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),

		schema.StringField("email", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Subscriber email"),
			schema.Unique()),

		// Subscription details
		schema.StringField("list_type",
			schema.HelpText("List type: newsletter, promotions, offers, all")),
		schema.StringField("source",
			schema.HelpText("Source: website, checkout, popup, social, referral, other")),

		// Preferences
		schema.MapField("preferences", schema.Optional(),
			schema.HelpText("Subscriber preferences")),

		// GDPR
		schema.BoolField("consent_given", schema.Default(false),
			schema.HelpText("GDPR consent given")),
		schema.TimeField("consent_date", schema.Optional()),
		schema.StringField("consent_ip", schema.MaxLength(45), schema.Optional(),
			schema.HelpText("IP address when consent given")),

		// Status
		schema.StringField("status",
			schema.HelpText("Status: subscribed, unsubscribed, pending, bounced, complained")),
		schema.TimeField("subscribed_at", schema.Optional()),
		schema.TimeField("unsubscribed_at", schema.Optional()),

		// Tracking
		schema.IntField("click_count", schema.Default(0)),
		schema.IntField("open_count", schema.Default(0)),

		// Segments
		schema.StringSliceField("segments", schema.Optional(),
			schema.HelpText("Subscriber segments")),

		// Timestamps
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
			schema.UniqueIndexOn("idx_subscription_email", "email"),
			schema.IndexOn("idx_subscription_status", "status"),
			schema.IndexOn("idx_subscription_source", "source"),
		},
	}
}

func (NewsletterSubscription) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (NewsletterSubscription) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate email format
			return nil
		},
	}
}

// ==================== PromotionUsage Fields ====================

func (PromotionUsage) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("promotion_id", schema.Required(),
			schema.HelpText("Associated promotion")),
		schema.ObjectIDField("order_id", schema.Required(),
			schema.HelpText("Associated order")),
		schema.ObjectIDField("customer_id", schema.Optional(),
			schema.HelpText("Associated customer")),

		// Discount applied
		schema.Float64Field("discount_amount", schema.Default(0.0),
			schema.HelpText("Discount amount applied")),

		// Timestamps
		schema.TimeField("used_at", schema.AutoNowAdd(),
			schema.HelpText("When the promotion was used")),
	}
}

func (PromotionUsage) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "promotion_usage",
		VerboseName:       "Promotion Usage",
		VerboseNamePlural: "Promotion Usage",
		OrderBy:           []string{"-used_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_usage_promotion", "promotion_id"),
			schema.IndexOn("idx_usage_order", "order_id"),
			schema.IndexOn("idx_usage_customer", "customer_id"),
			schema.IndexOn("idx_usage_date", "used_at"),
		},
	}
}

func (PromotionUsage) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("promotion_id", "Promotion",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("usage_records")),
		schema.ForeignKeyField("order_id", "Order",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("promotion_usage")),
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("promotion_usage")),
	}
}

func (PromotionUsage) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// Increment promotion times_used
			return nil
		},
	}
}

// ==================== Helper Types ====================

// PromotionType constants
const (
	PromotionTypePercentageFixed    = "percentage_fixed"
	PromotionTypePercentageOff      = "percentage_off"
	PromotionTypeBuyXGetY          = "buy_x_get_y"
	PromotionTypeBundleDiscount     = "bundle_discount"
	PromotionTypeFreeShipping       = "free_shipping"
	PromotionTypeFlashSale          = "flash_sale"
	PromotionTypeGiftWithPurchase   = "gift_with_purchase"
)

// DiscountType constants
const (
	DiscountTypePercentage = "percentage"
	DiscountTypeFixed      = "fixed"
)

// AppliesTo constants
const (
	AppliesToAll              = "all"
	AppliesToProducts         = "products"
	AppliesToCategories       = "categories"
	AppliesToBrands           = "brands"
	AppliesToSpecificProducts = "specific_products"
)

// ListType constants
const (
	ListTypeNewsletter = "newsletter"
	ListTypePromotions = "promotions"
	ListTypeOffers     = "offers"
	ListTypeAll        = "all"
)

// Source constants
const (
	SourceWebsite  = "website"
	SourceCheckout = "checkout"
	SourcePopup    = "popup"
	SourceSocial   = "social"
	SourceReferral = "referral"
	SourceOther    = "other"
)

// Status constants
const (
	StatusSubscribed   = "subscribed"
	StatusUnsubscribed = "unsubscribed"
	StatusPending      = "pending"
	StatusBounced      = "bounced"
	StatusComplained   = "complained"
)

// Position constants
const (
	PositionHomepageTop    = "homepage_top"
	PositionHomepageMiddle = "homepage_middle"
	PositionHomepageBottom = "homepage_bottom"
	PositionCategoryPage   = "category_page"
	PositionProductPage    = "product_page"
	PositionCartPage       = "cart_page"
	PositionCheckoutPage   = "checkout_page"
)

// RuleType constants
const (
	RuleTypeProductCategory   = "product_category"
	RuleTypeProductBrand      = "product_brand"
	RuleTypeProductTag        = "product_tag"
	RuleTypeProductPrice      = "product_price"
	RuleTypeCustomerGroup     = "customer_group"
	RuleTypeCustomerLocation  = "customer_location"
	RuleTypeDayOfWeek         = "day_of_week"
	RuleTypeTimeRange         = "time_range"
	RuleTypeCartTotal         = "cart_total"
	RuleTypeCartItemCount     = "cart_item_count"
	RuleTypeQuantityThreshold  = "quantity_threshold"
	RuleTypeFirstPurchase     = "first_purchase"
)

// Logic constants
const (
	LogicAnd = "and"
	LogicOr  = "or"
)

// RegisterModels registers promotions models with the framework
func RegisterModels() {
	registry.RegisterModel(&Promotion{})
	registry.RegisterModel(&PromotionRule{})
	registry.RegisterModel(&Banner{})
	registry.RegisterModel(&NewsletterSubscription{})
	registry.RegisterModel(&PromotionUsage{})
}

import (
	"context"
	"time"

	"github.com/forgego/forge/registry"
	"github.com/forgego/forge/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BannerSchedule represents scheduling for banners
type BannerSchedule struct {
	DayOfWeek int    `bson:"day_of_week" json:"day_of_week"`
	StartTime string `bson:"start_time" json:"start_time"`
	EndTime   string `bson:"end_time" json:"end_time"`
}

// Promotion represents a promotional campaign
type Promotion struct {
	PromotionGenerated
}

// PromotionRule represents rules for promotions
type PromotionRule struct {
	PromotionRuleGenerated
}

// Banner represents promotional banners
type Banner struct {
	BannerGenerated
}

// NewsletterSubscription represents newsletter subscribers
type NewsletterSubscription struct {
	NewsletterSubscriptionGenerated
}

// PromotionUsage tracks promotion usage
type PromotionUsage struct {
	PromotionUsageGenerated
}

// ==================== Promotion Fields ====================

func (Promotion) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),

		// Identification
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Promotion name")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Promotion description")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique promotion code"),
			schema.Unique()),

		// Promotion type
		schema.StringField("type", schema.Required(),
			schema.HelpText("Type: percentage_fixed, percentage_off, buy_x_get_y, bundle_discount, free_shipping, flash_sale, gift_with_purchase")),
		schema.Float64Field("discount_value", schema.Default(0.0),
			schema.HelpText("Discount value")),
		schema.StringField("discount_type",
			schema.HelpText("Type: percentage, fixed")),

		// Conditions
		schema.Float64Field("min_purchase", schema.Default(0.0),
			schema.HelpText("Minimum purchase amount")),
		schema.Float64Field("max_discount", schema.Default(0.0),
			schema.HelpText("Maximum discount amount")),

		// Buy X Get Y
		schema.IntField("buy_quantity", schema.Default(0),
			schema.HelpText("Buy quantity")),
		schema.IntField("get_quantity", schema.Default(0),
			schema.HelpText("Get quantity")),
		schema.ObjectIDField("free_product_id", schema.Optional()),

		// Free shipping
		schema.BoolField("free_shipping", schema.Default(false)),

		// Applicability
		schema.StringField("applies_to",
			schema.HelpText("Type: all, products, categories, brands, specific_products")),
		schema.ObjectIDSliceField("product_ids", schema.Optional()),
		schema.ObjectIDSliceField("category_ids", schema.Optional()),
		schema.ObjectIDSliceField("brand_ids", schema.Optional()),

		// Customer restrictions
		schema.BoolField("new_customers_only", schema.Default(false)),
		schema.ObjectIDSliceField("customer_group_ids", schema.Optional()),

		// Usage limits
		schema.IntField("total_usage_limit", schema.Default(0),
			schema.HelpText("Total usage limit (0 = unlimited)")),
		schema.IntField("per_customer_limit", schema.Default(0),
			schema.HelpText("Per customer limit (0 = unlimited)")),

		// Dates
		schema.TimeField("start_date", schema.Optional()),
		schema.TimeField("end_date", schema.Optional()),

		// Priority and stacking
		schema.IntField("priority", schema.Default(0)),
		schema.BoolField("can_stack", schema.Default(false)),
		schema.ObjectIDSliceField("stack_with", schema.Optional()),

		// Status
		schema.BoolField("is_active", schema.Default(true)),

		// Stats
		schema.IntField("times_used", schema.Default(0)),

		// Timestamps
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
			schema.IndexOn("idx_promotion_type", "type"),
			schema.IndexOn("idx_promotion_dates", "start_date", "end_date"),
		},
	}
}

func (Promotion) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("free_product_id", "Product",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("free_promotions")),
		schema.ForeignKeyField("promotion_id", "PromotionRule",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("promotion")),
		schema.ForeignKeyField("promotion_id", "PromotionUsage",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("promotion")),
	}
}

func (Promotion) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate promotion type
			// Validate dates
			// Uppercase code
			return nil
		},
	}
}

// ==================== PromotionRule Fields ====================

func (PromotionRule) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("promotion_id", schema.Required(),
			schema.HelpText("Associated promotion")),

		// Rule type
		schema.StringField("rule_type", schema.Required(),
			schema.HelpText("Rule type: product_category, product_brand, product_tag, product_price, customer_group, customer_location, day_of_week, time_range, cart_total, cart_item_count, quantity_threshold, first_purchase")),

		// Rule parameters
		schema.MapField("parameters", schema.Optional(),
			schema.HelpText("Flexible JSON for different rule types")),

		// Logic
		schema.StringField("logic",
			schema.HelpText("Logic: and, or")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (PromotionRule) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "promotion_rules",
		VerboseName:       "Promotion Rule",
		VerboseNamePlural: "Promotion Rules",
		OrderBy:           []string{"promotion_id", "created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_rule_promotion", "promotion_id"),
			schema.IndexOn("idx_rule_type", "rule_type"),
			schema.IndexOn("idx_rule_active", "is_active"),
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

// ==================== Banner Fields ====================

func (Banner) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),

		schema.StringField("title", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Banner title")),
		schema.StringField("subtitle", schema.MaxLength(200), schema.Optional(),
			schema.HelpText("Banner subtitle")),

		// Media
		schema.StringField("image_url", schema.Required(), schema.MaxLength(500),
			schema.HelpText("Banner image URL")),
		schema.StringField("mobile_image_url", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Mobile-optimized image URL")),
		schema.StringField("video_url", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Video URL")),

		// Content
		schema.TextField("content", schema.Optional(),
			schema.HelpText("Banner content")),
		schema.StringField("link", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Click-through link")),
		schema.StringField("link_text", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Link button text")),

		// Placement
		schema.StringField("position", schema.Required(),
			schema.HelpText("Position: homepage_top, homepage_middle, homepage_bottom, category_page, product_page, cart_page, checkout_page")),

		// Styling
		schema.StringField("background_color", schema.MaxLength(20), schema.Optional(),
			schema.HelpText("Background color (hex)")),
		schema.StringField("text_color", schema.MaxLength(20), schema.Optional(),
			schema.HelpText("Text color (hex)")),

		// Dates
		schema.TimeField("start_date", schema.Optional()),
		schema.TimeField("end_date", schema.Optional()),

		// Scheduling
		schema.SliceField("schedule", schema.Optional(),
			schema.HelpText("Banner schedule")),

		// Targeting
		schema.StringSliceField("device_types", schema.Optional(),
			schema.HelpText("Target devices: mobile, tablet, desktop")),
		schema.StringSliceField("user_types", schema.Optional(),
			schema.HelpText("Target users: guest, logged_in")),
		schema.ObjectIDSliceField("customer_group_ids", schema.Optional()),

		// Priority
		schema.IntField("priority", schema.Default(0)),

		// Status
		schema.BoolField("is_active", schema.Default(true)),

		// Stats
		schema.IntField("click_count", schema.Default(0)),
		schema.IntField("view_count", schema.Default(0)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Banner) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "banners",
		VerboseName:       "Banner",
		VerboseNamePlural: "Banners",
		OrderBy:           []string{"-priority", "-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_banner_active", "is_active"),
			schema.IndexOn("idx_banner_position", "position"),
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

// ==================== NewsletterSubscription Fields ====================

func (NewsletterSubscription) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),

		schema.StringField("email", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Subscriber email"),
			schema.Unique()),

		// Subscription details
		schema.StringField("list_type",
			schema.HelpText("List type: newsletter, promotions, offers, all")),
		schema.StringField("source",
			schema.HelpText("Source: website, checkout, popup, social, referral, other")),

		// Preferences
		schema.MapField("preferences", schema.Optional(),
			schema.HelpText("Subscriber preferences")),

		// GDPR
		schema.BoolField("consent_given", schema.Default(false),
			schema.HelpText("GDPR consent given")),
		schema.TimeField("consent_date", schema.Optional()),
		schema.StringField("consent_ip", schema.MaxLength(45), schema.Optional(),
			schema.HelpText("IP address when consent given")),

		// Status
		schema.StringField("status",
			schema.HelpText("Status: subscribed, unsubscribed, pending, bounced, complained")),
		schema.TimeField("subscribed_at", schema.Optional()),
		schema.TimeField("unsubscribed_at", schema.Optional()),

		// Tracking
		schema.IntField("click_count", schema.Default(0)),
		schema.IntField("open_count", schema.Default(0)),

		// Segments
		schema.StringSliceField("segments", schema.Optional(),
			schema.HelpText("Subscriber segments")),

		// Timestamps
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
			schema.UniqueIndexOn("idx_subscription_email", "email"),
			schema.IndexOn("idx_subscription_status", "status"),
			schema.IndexOn("idx_subscription_source", "source"),
		},
	}
}

func (NewsletterSubscription) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (NewsletterSubscription) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate email format
			return nil
		},
	}
}

// ==================== PromotionUsage Fields ====================

func (PromotionUsage) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("promotion_id", schema.Required(),
			schema.HelpText("Associated promotion")),
		schema.ObjectIDField("order_id", schema.Required(),
			schema.HelpText("Associated order")),
		schema.ObjectIDField("customer_id", schema.Optional(),
			schema.HelpText("Associated customer")),

		// Discount applied
		schema.Float64Field("discount_amount", schema.Default(0.0),
			schema.HelpText("Discount amount applied")),

		// Timestamps
		schema.TimeField("used_at", schema.AutoNowAdd(),
			schema.HelpText("When the promotion was used")),
	}
}

func (PromotionUsage) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "promotion_usage",
		VerboseName:       "Promotion Usage",
		VerboseNamePlural: "Promotion Usage",
		OrderBy:           []string{"-used_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_usage_promotion", "promotion_id"),
			schema.IndexOn("idx_usage_order", "order_id"),
			schema.IndexOn("idx_usage_customer", "customer_id"),
			schema.IndexOn("idx_usage_date", "used_at"),
		},
	}
}

func (PromotionUsage) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("promotion_id", "Promotion",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("usage_records")),
		schema.ForeignKeyField("order_id", "Order",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("promotion_usage")),
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("promotion_usage")),
	}
}

func (PromotionUsage) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// Increment promotion times_used
			return nil
		},
	}
}

// ==================== Helper Types ====================

// PromotionType constants
const (
	PromotionTypePercentageFixed    = "percentage_fixed"
	PromotionTypePercentageOff      = "percentage_off"
	PromotionTypeBuyXGetY          = "buy_x_get_y"
	PromotionTypeBundleDiscount     = "bundle_discount"
	PromotionTypeFreeShipping       = "free_shipping"
	PromotionTypeFlashSale          = "flash_sale"
	PromotionTypeGiftWithPurchase   = "gift_with_purchase"
)

// DiscountType constants
const (
	DiscountTypePercentage = "percentage"
	DiscountTypeFixed      = "fixed"
)

// AppliesTo constants
const (
	AppliesToAll              = "all"
	AppliesToProducts         = "products"
	AppliesToCategories       = "categories"
	AppliesToBrands           = "brands"
	AppliesToSpecificProducts = "specific_products"
)

// ListType constants
const (
	ListTypeNewsletter = "newsletter"
	ListTypePromotions = "promotions"
	ListTypeOffers     = "offers"
	ListTypeAll        = "all"
)

// Source constants
const (
	SourceWebsite  = "website"
	SourceCheckout = "checkout"
	SourcePopup    = "popup"
	SourceSocial   = "social"
	SourceReferral = "referral"
	SourceOther    = "other"
)

// Status constants
const (
	StatusSubscribed   = "subscribed"
	StatusUnsubscribed = "unsubscribed"
	StatusPending      = "pending"
	StatusBounced      = "bounced"
	StatusComplained   = "complained"
)

// Position constants
const (
	PositionHomepageTop    = "homepage_top"
	PositionHomepageMiddle = "homepage_middle"
	PositionHomepageBottom = "homepage_bottom"
	PositionCategoryPage   = "category_page"
	PositionProductPage    = "product_page"
	PositionCartPage       = "cart_page"
	PositionCheckoutPage   = "checkout_page"
)

// RuleType constants
const (
	RuleTypeProductCategory   = "product_category"
	RuleTypeProductBrand      = "product_brand"
	RuleTypeProductTag        = "product_tag"
	RuleTypeProductPrice      = "product_price"
	RuleTypeCustomerGroup     = "customer_group"
	RuleTypeCustomerLocation  = "customer_location"
	RuleTypeDayOfWeek         = "day_of_week"
	RuleTypeTimeRange         = "time_range"
	RuleTypeCartTotal         = "cart_total"
	RuleTypeCartItemCount     = "cart_item_count"
	RuleTypeQuantityThreshold  = "quantity_threshold"
	RuleTypeFirstPurchase     = "first_purchase"
)

// Logic constants
const (
	LogicAnd = "and"
	LogicOr  = "or"
)

// RegisterModels registers promotions models with the framework
func RegisterModels() {
	registry.RegisterModel(&Promotion{})
	registry.RegisterModel(&PromotionRule{})
	registry.RegisterModel(&Banner{})
	registry.RegisterModel(&NewsletterSubscription{})
	registry.RegisterModel(&PromotionUsage{})
}

	"context"
	"time"

	"github.com/forgego/forge/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BannerSchedule represents scheduling for banners
type BannerSchedule struct {
	DayOfWeek int    `bson:"day_of_week" json:"day_of_week"` // 0-6
	StartTime string `bson:"start_time" json:"start_time"`   // HH:MM
	EndTime   string `bson:"end_time" json:"end_time"`       // HH:MM
}

// Promotion represents a promotional campaign
type Promotion struct {
	PromotionGenerated
}

// PromotionRule represents rules for promotions
type PromotionRule struct {
	PromotionRuleGenerated
}

// Banner represents promotional banners
type Banner struct {
	BannerGenerated
}

// NewsletterSubscription represents newsletter subscribers
type NewsletterSubscription struct {
	NewsletterSubscriptionGenerated
}

// PromotionUsage tracks promotion usage
type PromotionUsage struct {
	PromotionUsageGenerated
}

// ==================== Promotion Fields ====================

func (Promotion) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),

		// Identification
		schema.StringField("name", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Promotion name")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Promotion description")),
		schema.StringField("code", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique promotion code"),
			schema.Unique()),

		// Promotion type
		schema.StringField("type", schema.Required(),
			schema.HelpText("Type: percentage_fixed, percentage_off, buy_x_get_y, bundle_discount, free_shipping, flash_sale, gift_with_purchase")),
		schema.Float64Field("discount_value", schema.Default(0.0),
			schema.HelpText("Discount value")),
		schema.StringField("discount_type",
			schema.HelpText("Type: percentage, fixed")),

		// Conditions
		schema.Float64Field("min_purchase", schema.Default(0.0),
			schema.HelpText("Minimum purchase amount")),
		schema.Float64Field("max_discount", schema.Default(0.0),
			schema.HelpText("Maximum discount amount")),

		// Buy X Get Y
		schema.IntField("buy_quantity", schema.Default(0),
			schema.HelpText("Buy quantity")),
		schema.IntField("get_quantity", schema.Default(0),
			schema.HelpText("Get quantity")),
		schema.ObjectIDField("free_product_id", schema.Optional()),

		// Free shipping
		schema.BoolField("free_shipping", schema.Default(false)),

		// Applicability
		schema.StringField("applies_to",
			schema.HelpText("Type: all, products, categories, brands, specific_products")),
		schema.ObjectIDSliceField("product_ids", schema.Optional()),
		schema.ObjectIDSliceField("category_ids", schema.Optional()),
		schema.ObjectIDSliceField("brand_ids", schema.Optional()),

		// Customer restrictions
		schema.BoolField("new_customers_only", schema.Default(false)),
		schema.ObjectIDSliceField("customer_group_ids", schema.Optional()),

		// Usage limits
		schema.IntField("total_usage_limit", schema.Default(0),
			schema.HelpText("Total usage limit (0 = unlimited)")),
		schema.IntField("per_customer_limit", schema.Default(0),
			schema.HelpText("Per customer limit (0 = unlimited)")),

		// Dates
		schema.TimeField("start_date", schema.Optional()),
		schema.TimeField("end_date", schema.Optional()),

		// Priority and stacking
		schema.IntField("priority", schema.Default(0)),
		schema.BoolField("can_stack", schema.Default(false)),
		schema.ObjectIDSliceField("stack_with", schema.Optional()),

		// Status
		schema.BoolField("is_active", schema.Default(true)),

		// Stats
		schema.IntField("times_used", schema.Default(0)),

		// Timestamps
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
			schema.IndexOn("idx_promotion_type", "type"),
			schema.IndexOn("idx_promotion_dates", "start_date", "end_date"),
		},
	}
}

func (Promotion) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("free_product_id", "Product",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("free_promotions")),
		schema.ForeignKeyField("promotion_id", "PromotionRule",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("promotion")),
		schema.ForeignKeyField("promotion_id", "PromotionUsage",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("promotion")),
	}
}

func (Promotion) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate promotion type
			// Validate dates
			// Uppercase code
			return nil
		},
	}
}

// ==================== PromotionRule Fields ====================

func (PromotionRule) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("promotion_id", schema.Required(),
			schema.HelpText("Associated promotion")),

		// Rule type
		schema.StringField("rule_type", schema.Required(),
			schema.HelpText("Rule type: product_category, product_brand, product_tag, product_price, customer_group, customer_location, day_of_week, time_range, cart_total, cart_item_count, quantity_threshold, first_purchase")),

		// Rule parameters
		schema.MapField("parameters", schema.Optional(),
			schema.HelpText("Flexible JSON for different rule types")),

		// Logic
		schema.StringField("logic",
			schema.HelpText("Logic: and, or")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (PromotionRule) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "promotion_rules",
		VerboseName:       "Promotion Rule",
		VerboseNamePlural: "Promotion Rules",
		OrderBy:           []string{"promotion_id", "created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_rule_promotion", "promotion_id"),
			schema.IndexOn("idx_rule_type", "rule_type"),
			schema.IndexOn("idx_rule_active", "is_active"),
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

// ==================== Banner Fields ====================

func (Banner) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),

		schema.StringField("title", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Banner title")),
		schema.StringField("subtitle", schema.MaxLength(200), schema.Optional(),
			schema.HelpText("Banner subtitle")),

		// Media
		schema.StringField("image_url", schema.Required(), schema.MaxLength(500),
			schema.HelpText("Banner image URL")),
		schema.StringField("mobile_image_url", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Mobile-optimized image URL")),
		schema.StringField("video_url", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Video URL")),

		// Content
		schema.TextField("content", schema.Optional(),
			schema.HelpText("Banner content")),
		schema.StringField("link", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Click-through link")),
		schema.StringField("link_text", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Link button text")),

		// Placement
		schema.StringField("position", schema.Required(),
			schema.HelpText("Position: homepage_top, homepage_middle, homepage_bottom, category_page, product_page, cart_page, checkout_page")),

		// Styling
		schema.StringField("background_color", schema.MaxLength(20), schema.Optional(),
			schema.HelpText("Background color (hex)")),
		schema.StringField("text_color", schema.MaxLength(20), schema.Optional(),
			schema.HelpText("Text color (hex)")),

		// Dates
		schema.TimeField("start_date", schema.Optional()),
		schema.TimeField("end_date", schema.Optional()),

		// Scheduling
		schema.SliceField("schedule", schema.Optional(),
			schema.HelpText("Banner schedule")),

		// Targeting
		schema.StringSliceField("device_types", schema.Optional(),
			schema.HelpText("Target devices: mobile, tablet, desktop")),
		schema.StringSliceField("user_types", schema.Optional(),
			schema.HelpText("Target users: guest, logged_in")),
		schema.ObjectIDSliceField("customer_group_ids", schema.Optional()),

		// Priority
		schema.IntField("priority", schema.Default(0)),

		// Status
		schema.BoolField("is_active", schema.Default(true)),

		// Stats
		schema.IntField("click_count", schema.Default(0)),
		schema.IntField("view_count", schema.Default(0)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Banner) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "banners",
		VerboseName:       "Banner",
		VerboseNamePlural: "Banners",
		OrderBy:           []string{"-priority", "-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_banner_active", "is_active"),
			schema.IndexOn("idx_banner_position", "position"),
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

// ==================== NewsletterSubscription Fields ====================

func (NewsletterSubscription) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),

		schema.StringField("email", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Subscriber email"),
			schema.Unique()),

		// Subscription details
		schema.StringField("list_type",
			schema.HelpText("List type: newsletter, promotions, offers, all")),
		schema.StringField("source",
			schema.HelpText("Source: website, checkout, popup, social, referral, other")),

		// Preferences
		schema.MapField("preferences", schema.Optional(),
			schema.HelpText("Subscriber preferences")),

		// GDPR
		schema.BoolField("consent_given", schema.Default(false),
			schema.HelpText("GDPR consent given")),
		schema.TimeField("consent_date", schema.Optional()),
		schema.StringField("consent_ip", schema.MaxLength(45), schema.Optional(),
			schema.HelpText("IP address when consent given")),

		// Status
		schema.StringField("status",
			schema.HelpText("Status: subscribed, unsubscribed, pending, bounced, complained")),
		schema.TimeField("subscribed_at", schema.Optional()),
		schema.TimeField("unsubscribed_at", schema.Optional()),

		// Tracking
		schema.IntField("click_count", schema.Default(0)),
		schema.IntField("open_count", schema.Default(0)),

		// Segments
		schema.StringSliceField("segments", schema.Optional(),
			schema.HelpText("Subscriber segments")),

		// Timestamps
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
			schema.UniqueIndexOn("idx_subscription_email", "email"),
			schema.IndexOn("idx_subscription_status", "status"),
			schema.IndexOn("idx_subscription_source", "source"),
		},
	}
}

func (NewsletterSubscription) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (NewsletterSubscription) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate email format
			return nil
		},
	}
}

// ==================== PromotionUsage Fields ====================

func (PromotionUsage) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("promotion_id", schema.Required(),
			schema.HelpText("Associated promotion")),
		schema.ObjectIDField("order_id", schema.Required(),
			schema.HelpText("Associated order")),
		schema.ObjectIDField("customer_id", schema.Optional(),
			schema.HelpText("Associated customer")),

		// Discount applied
		schema.Float64Field("discount_amount", schema.Default(0.0),
			schema.HelpText("Discount amount applied")),

		// Timestamps
		schema.TimeField("used_at", schema.AutoNowAdd(),
			schema.HelpText("When the promotion was used")),
	}
}

func (PromotionUsage) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "promotion_usage",
		VerboseName:       "Promotion Usage",
		VerboseNamePlural: "Promotion Usage",
		OrderBy:           []string{"-used_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_usage_promotion", "promotion_id"),
			schema.IndexOn("idx_usage_order", "order_id"),
			schema.IndexOn("idx_usage_customer", "customer_id"),
			schema.IndexOn("idx_usage_date", "used_at"),
		],
	}
}

func (PromotionUsage) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("promotion_id", "Promotion",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("usage_records")),
		schema.ForeignKeyField("order_id", "Order",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("promotion_usage")),
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("promotion_usage")),
	}
}

func (PromotionUsage) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// Increment promotion times_used
			return nil
		},
	}
}

// ==================== Helper Types ====================

// PromotionType constants
const (
	PromotionTypePercentageFixed     = "percentage_fixed"
	PromotionTypePercentageOff        = "percentage_off"
	PromotionTypeBuyXGetY             = "buy_x_get_y"
	PromotionTypeBundleDiscount       = "bundle_discount"
	PromotionTypeFreeShipping         = "free_shipping"
	PromotionTypeFlashSale            = "flash_sale"
	PromotionTypeGiftWithPurchase     = "gift_with_purchase"
)

// DiscountType constants
const (
	DiscountTypePercentage = "percentage"
	DiscountTypeFixed      = "fixed"
)

// AppliesTo constants
const (
	AppliesToAll            = "all"
	AppliesToProducts       = "products"
	AppliesToCategories     = "categories"
	AppliesToBrands         = "brands"
	AppliesToSpecificProducts = "specific_products"
)

// ListType constants
const (
	ListTypeNewsletter = "newsletter"
	ListTypePromotions = "promotions"
	ListTypeOffers     = "offers"
	ListTypeAll        = "all"
)

// Source constants
const (
	SourceWebsite  = "website"
	SourceCheckout = "checkout"
	SourcePopup    = "popup"
	SourceSocial   = "social"
	SourceReferral = "referral"
	SourceOther    = "other"
)

// Status constants
const (
	StatusSubscribed   = "subscribed"
	StatusUnsubscribed  = "unsubscribed"
	StatusPending      = "pending"
	StatusBounced      = "bounced"
	StatusComplained   = "complained"
)

// Position constants
const (
	PositionHomepageTop    = "homepage_top"
	PositionHomepageMiddle = "homepage_middle"
	PositionHomepageBottom = "homepage_bottom"
	PositionCategoryPage   = "category_page"
	PositionProductPage    = "product_page"
	PositionCartPage       = "cart_page"
	PositionCheckoutPage   = "checkout_page"
)

// RuleType constants
const (
	RuleTypeProductCategory    = "product_category"
	RuleTypeProductBrand       = "product_brand"
	RuleTypeProductTag         = "product_tag"
	RuleTypeProductPrice       = "product_price"
	RuleTypeCustomerGroup      = "customer_group"
	RuleTypeCustomerLocation   = "customer_location"
	RuleTypeDayOfWeek          = "day_of_week"
	RuleTypeTimeRange          = "time_range"
	RuleTypeCartTotal          = "cart_total"
	RuleTypeCartItemCount      = "cart_item_count"
	RuleTypeQuantityThreshold  = "quantity_threshold"
	RuleTypeFirstPurchase      = "first_purchase"
)

// Logic constants
const (
	LogicAnd = "and"
	LogicOr  = "or"
)

// RegisterModels registers promotions models with the framework
func RegisterModels() {
	registry.RegisterModel(&Promotion{})
	registry.RegisterModel(&PromotionRule{})
	registry.RegisterModel(&Banner{})
	registry.RegisterModel(&NewsletterSubscription{})
	registry.RegisterModel(&PromotionUsage{})
}

