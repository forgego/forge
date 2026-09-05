package marketing

import (
	"context"

	"github.com/forgego/forge/registry"
	"github.com/forgego/forge/schema"
)

// Coupon represents discount codes and promotions
type Coupon struct {
	CouponGenerated
}

func (Coupon) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),

		// Identification
		schema.StringField("code", schema.Required(), schema.MaxLength(50), schema.Unique(),
			schema.HelpText("Coupon code (e.g., SAVE20)")),
		schema.StringField("name", schema.Required(), schema.MaxLength(200),
			schema.HelpText("Internal name for coupon")),
		schema.TextField("description", schema.Optional(),
			schema.HelpText("Description shown to customers")),

		// Discount type
		schema.StringField("discount_type", schema.Required(), schema.MaxLength(20),
			schema.VerboseName("Discount Type"),
			schema.HelpText("Type: percentage, fixed_amount, free_shipping, buy_x_get_y")),
		schema.Float64Field("discount_value", schema.Required(),
			schema.VerboseName("Discount Value"),
			schema.HelpText("Discount value (percentage or fixed amount)")),

		// Constraints
		schema.Float64Field("minimum_purchase_amount", schema.Optional(),
			schema.HelpText("Minimum cart value to apply coupon")),
		schema.Float64Field("maximum_discount_amount", schema.Optional(),
			schema.HelpText("Cap on discount amount for percentage coupons")),

		// Usage limits
		schema.Int32Field("usage_limit", schema.Optional(),
			schema.HelpText("Total number of times coupon can be used (null = unlimited)")),
		schema.Int32Field("usage_limit_per_customer", schema.Default(1),
			schema.HelpText("Max uses per customer")),
		schema.Int32Field("usage_count", schema.Default(0),
			schema.HelpText("Number of times coupon has been used")),

		// Applicability
		schema.BoolField("applies_to_all_products", schema.Default(true)),
		schema.StringField("product_ids", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Comma-separated product IDs (if not all products)")),
		schema.StringField("category_ids", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Comma-separated category IDs")),
		schema.StringField("excluded_product_ids", schema.MaxLength(500), schema.Optional()),

		// Customer restrictions
		schema.BoolField("applies_to_all_customers", schema.Default(true)),
		schema.StringField("customer_group_ids", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Comma-separated customer group IDs")),
		schema.TextField("customer_email_list", schema.Optional(),
			schema.HelpText("Specific email addresses (one per line)")),

		// Validity period
		schema.TimeField("valid_from", schema.Required(),
			schema.HelpText("Start date/time")),
		schema.TimeField("valid_until", schema.Optional(),
			schema.VerboseName("Valid Until"),
			schema.HelpText("End date/time (null = no expiry)")),

		// Status
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.BoolField("is_public", schema.Default(true),
			schema.HelpText("Show in promotions list")),

		// Priority (for multiple coupons)
		schema.Int32Field("priority", schema.Default(0),
			schema.HelpText("Higher priority coupons apply first")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Coupon) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "coupons",
		VerboseName:       "Coupon",
		VerboseNamePlural: "Coupons",
		OrderBy:           []string{"-priority", "-created_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_coupon_code", "code"),
			schema.IndexOn("idx_coupon_active", "is_active"),
			schema.IndexOn("idx_coupon_valid_from", "valid_from"),
			schema.IndexOn("idx_coupon_valid_until", "valid_until"),
		},
	}
}

func (Coupon) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Coupon) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate discount_type
			// Uppercase coupon code
			// Validate dates
			return nil
		},
	}
}

// CouponUsage tracks coupon usage
type CouponUsage struct {
	CouponUsageGenerated
}

func (CouponUsage) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("coupon_id", schema.Required(),
			schema.VerboseName("Coupon")),
		schema.Int64Field("order_id", schema.Required(),
			schema.VerboseName("Order")),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),

		schema.Float64Field("discount_amount", schema.Required(),
			schema.VerboseName("Discount Amount"),
			schema.HelpText("Actual discount applied")),

		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
	}
}

func (CouponUsage) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "coupon_usage",
		VerboseName:       "Coupon Usage",
		VerboseNamePlural: "Coupon Usage",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_usage_coupon", "coupon_id"),
			schema.IndexOn("idx_usage_order", "order_id"),
			schema.IndexOn("idx_usage_customer", "customer_id"),
		},
	}
}

func (CouponUsage) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("coupon_id", "Coupon",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("usage_records")),
		schema.ForeignKeyField("order_id", "Order",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("coupon_usage")),
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("coupon_usage")),
	}
}

func (CouponUsage) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// Increment coupon usage_count
			return nil
		},
	}
}

// Review represents product reviews
type Review struct {
	ReviewGenerated
}

func (Review) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("product_id", schema.Required(),
			schema.VerboseName("Product")),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),
		schema.Int64Field("order_id", schema.Optional(),
			schema.HelpText("Order where product was purchased (verified purchase)")),

		// Review content
		schema.StringField("title", schema.Required(), schema.MaxLength(255),
			schema.HelpText("Review title/headline")),
		schema.TextField("content", schema.Required(),
			schema.HelpText("Review body")),
		schema.Int32Field("rating", schema.Required(),
			schema.HelpText("Rating from 1 to 5")),

		// Verification
		schema.BoolField("is_verified_purchase", schema.Default(false),
			schema.HelpText("Customer purchased this product")),

		// Moderation
		schema.StringField("status", schema.Required(), schema.MaxLength(20), schema.Default("pending"),
			schema.VerboseName("Status"),
			schema.HelpText("Status: pending, approved, rejected, flagged")),
		schema.BoolField("is_featured", schema.Default(false),
			schema.VerboseName("Featured"),
			schema.HelpText("Feature this review")),

		// Helpfulness
		schema.Int32Field("helpful_count", schema.Default(0),
			schema.HelpText("Number of users who found this helpful")),
		schema.Int32Field("not_helpful_count", schema.Default(0)),

		// Response
		schema.TextField("merchant_response", schema.Optional(),
			schema.HelpText("Response from store owner")),
		schema.TimeField("merchant_response_at", schema.Optional()),
		schema.Int64Field("merchant_response_by_user_id", schema.Optional()),

		// Reporting
		schema.Int32Field("report_count", schema.Default(0),
			schema.HelpText("Number of times review was reported")),
		schema.TextField("report_reasons", schema.Optional(),
			schema.HelpText("Reasons for reports (JSON)")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
		schema.TimeField("updated_at", schema.AutoNow()),
		schema.TimeField("approved_at", schema.Optional()),
	}
}

func (Review) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "reviews",
		VerboseName:       "Product Review",
		VerboseNamePlural: "Product Reviews",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_review_product", "product_id"),
			schema.IndexOn("idx_review_customer", "customer_id"),
			schema.IndexOn("idx_review_order", "order_id"),
			schema.IndexOn("idx_review_status", "status"),
			schema.IndexOn("idx_review_rating", "rating"),
			schema.IndexOn("idx_review_featured", "is_featured"),
		},
		UniqueTogether: [][]string{
			{"product_id", "customer_id", "order_id"},
		},
	}
}

func (Review) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("product_id", "Product",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("reviews")),
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("reviews")),
		schema.ForeignKeyField("order_id", "Order",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("reviews")),
	}
}

func (Review) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Validate rating (1-5)
			// Check for duplicate review
			return nil
		},
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// Update product rating statistics
			// Send notification to merchant
			return nil
		},
		AfterUpdate: func(ctx context.Context, instance interface{}) error {
			// Update product rating statistics
			// Notify customer of status change
			return nil
		},
	}
}

// ReviewImage represents images attached to reviews
type ReviewImage struct {
	ReviewImageGenerated
}

func (ReviewImage) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("review_id", schema.Required(),
			schema.VerboseName("Review")),

		schema.StringField("image_url", schema.Required(), schema.MaxLength(500)),
		schema.StringField("thumbnail_url", schema.MaxLength(500), schema.Optional()),
		schema.StringField("alt_text", schema.MaxLength(255), schema.Optional(),
			schema.VerboseName("Alt Text")),

		schema.Int32Field("sort_order", schema.Default(0),
			schema.VerboseName("Sort Order")),

		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (ReviewImage) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "review_images",
		VerboseName:       "Review Image",
		VerboseNamePlural: "Review Images",
		OrderBy:           []string{"review_id", "sort_order"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_review_image_review", "review_id"),
		},
	}
}

func (ReviewImage) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("review_id", "Review",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("images")),
	}
}

func (ReviewImage) Hooks() *schema.ModelHooks {
	return nil
}

// ReviewHelpfulness tracks if users found reviews helpful
type ReviewHelpfulness struct {
	ReviewHelpfulnessGenerated
}

func (ReviewHelpfulness) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("review_id", schema.Required(),
			schema.VerboseName("Review")),
		schema.Int64Field("customer_id", schema.Optional(),
			schema.VerboseName("Customer")),

		schema.BoolField("is_helpful", schema.Required(),
			schema.VerboseName("Helpful"),
			schema.HelpText("True = helpful, False = not helpful")),

		schema.StringField("ip_address", schema.MaxLength(45), schema.Optional(),
			schema.HelpText("For anonymous votes")),

		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (ReviewHelpfulness) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "review_helpfulness",
		VerboseName:       "Review Helpfulness",
		VerboseNamePlural: "Review Helpfulness",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_helpfulness_review", "review_id"),
			schema.IndexOn("idx_helpfulness_customer", "customer_id"),
		},
		UniqueTogether: [][]string{
			{"review_id", "customer_id"},
			{"review_id", "ip_address"},
		},
	}
}

func (ReviewHelpfulness) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("review_id", "Review",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("helpfulness_votes")),
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("review_helpfulness_votes")),
	}
}

func (ReviewHelpfulness) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// Update review helpful_count or not_helpful_count
			return nil
		},
	}
}

// ProductQuestion represents Q&A for products
type ProductQuestion struct {
	ProductQuestionGenerated
}

func (ProductQuestion) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("product_id", schema.Required(),
			schema.VerboseName("Product")),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),

		// Question
		schema.TextField("question", schema.Required()),

		// Answer
		schema.TextField("answer", schema.Optional()),
		schema.TimeField("answered_at", schema.Optional()),
		schema.Int64Field("answered_by_user_id", schema.Optional()),
		schema.StringField("answered_by_user_name", schema.MaxLength(200), schema.Optional()),

		// Status
		schema.StringField("status", schema.Required(), schema.MaxLength(20), schema.Default("pending"),
			schema.VerboseName("Status"),
			schema.HelpText("Status: pending, answered, hidden")),
		schema.BoolField("is_public", schema.Default(true),
			schema.VerboseName("Public")),

		// Helpfulness
		schema.Int32Field("helpful_count", schema.Default(0)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ProductQuestion) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "product_questions",
		VerboseName:       "Product Question",
		VerboseNamePlural: "Product Questions",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_question_product", "product_id"),
			schema.IndexOn("idx_question_customer", "customer_id"),
			schema.IndexOn("idx_question_status", "status"),
		},
	}
}

func (ProductQuestion) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("product_id", "Product",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("questions")),
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("questions")),
	}
}

func (ProductQuestion) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		AfterUpdate: func(ctx context.Context, instance interface{}) error {
			// Send notification when answered
			return nil
		},
	}
}

// RegisterModels registers marketing models with the framework
func RegisterModels() {
	registry.RegisterModel(&Coupon{})
	registry.RegisterModel(&CouponUsage{})
	registry.RegisterModel(&Review{})
	registry.RegisterModel(&ReviewImage{})
	registry.RegisterModel(&ReviewHelpfulness{})
	registry.RegisterModel(&ProductQuestion{})
}
