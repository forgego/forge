package engagement

import (
	"context"
	"time"

	"github.com/forgego/forge/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RecentlyViewed tracks products that users have viewed
type RecentlyViewed struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID  *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty"`
	GuestID     string              `bson:"guest_id" json:"guest_id" validate:"max=100"`
	ProductID   primitive.ObjectID `bson:"product_id" json:"product_id" validate:"required"`
	VariantID   *primitive.ObjectID `bson:"variant_id,omitempty" json:"variant_id,omitempty"`

	// Tracking
	ViewedAt    time.Time `bson:"viewed_at" json:"viewed_at"`
	SessionID   string    `bson:"session_id" json:"session_id" validate:"max=100"`
	UserAgent   string    `bson:"user_agent" json:"user_agent" validate:"max=500"`
	IPAddress   string    `bson:"ip_address" json:"ip_address" validate:"max=45"`

	// Context
	Source      string    `bson:"source" json:"source" validate:"oneof=search browse category product recommended email social direct"`
	RefererURL  string    `bson:"referer_url" json:"referer_url" validate:"max=500"`
}

func (RecentlyViewed) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("customer_id", schema.Optional(),
			schema.HelpText("Customer ID (null for guests)")),
		schema.StringField("guest_id", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Guest session ID")),
		schema.ObjectIDField("product_id", schema.Required(),
			schema.HelpText("Viewed product ID")),
		schema.ObjectIDField("variant_id", schema.Optional(),
			schema.HelpText("Viewed variant ID")),

		// Tracking
		schema.TimeField("viewed_at", schema.Required(),
			schema.HelpText("When the product was viewed")),
		schema.StringField("session_id", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("User session ID")),
		schema.StringField("user_agent", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Browser user agent")),
		schema.StringField("ip_address", schema.Optional(), schema.MaxLength(45),
			schema.HelpText("User IP address")),

		// Context
		schema.StringField("source", schema.Optional(),
			schema.HelpText("Source: search, browse, category, product, recommended, email, social, direct")),
		schema.StringField("referer_url", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Referring URL")),
	}
}

func (RecentlyViewed) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "recently_viewed",
		VerboseName:       "Recently Viewed",
		VerboseNamePlural: "Recently Viewed",
		OrderBy:           []string{"-viewed_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_customer_viewed", "customer_id", "-viewed_at"),
			schema.IndexOn("idx_guest_viewed", "guest_id", "-viewed_at"),
			schema.IndexOn("idx_product_viewed", "product_id", "-viewed_at"),
			schema.IndexOn("idx_session_viewed", "session_id", "-viewed_at"),
		},
	}
}

func (RecentlyViewed) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (RecentlyViewed) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Auto-set viewed_at if not set
			return nil
		},
	}
}

// ProductComparison allows users to compare products
type ProductComparison struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID  *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty"`
	GuestID     string              `bson:"guest_id" json:"guest_id" validate:"max=100"`
	Name        string              `bson:"name" json:"name" validate:"required,min=2,max=100"`

	// Products being compared
	ProductIDs  []primitive.ObjectID `bson:"product_ids" json:"product_ids"`

	// Settings
	IsPublic    bool   `bson:"is_public" json:"is_public"`
	ShareToken  string `bson:"share_token" json:"share_token" validate:"max=100"`

	// Stats
	ViewCount   int       `bson:"view_count" json:"view_count"`

	// Timestamps
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

func (ProductComparison) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("customer_id", schema.Optional(),
			schema.HelpText("Customer ID (null for guest comparisons)")),
		schema.StringField("guest_id", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Guest session ID")),
		schema.StringField("name", schema.Required(), schema.MinLength(2), schema.MaxLength(100),
			schema.HelpText("Comparison name")),

		// Products
		schema.ObjectIDArrayField("product_ids",
			schema.HelpText("Products being compared")),

		// Settings
		schema.BoolField("is_public", schema.Default(false),
			schema.HelpText("Whether the comparison is publicly shareable")),
		schema.StringField("share_token", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Unique token for sharing")),

		// Stats
		schema.IntField("view_count", schema.Default(0),
			schema.HelpText("Number of times this comparison was viewed")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ProductComparison) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "product_comparisons",
		VerboseName:       "Product Comparison",
		VerboseNamePlural: "Product Comparisons",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_customer_comparison", "customer_id", "-created_at"),
			schema.IndexOn("idx_guest_comparison", "guest_id", "-created_at"),
			schema.IndexOn("idx_share_token", "share_token"),
			schema.IndexOn("idx_created_at", "-created_at"),
		},
	}
}

func (ProductComparison) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ProductComparison) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// Notification represents user notifications
type Notification struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID    primitive.ObjectID `bson:"customer_id" json:"customer_id" validate:"required"`

	// Notification details
	Type          string    `bson:"type" json:"type" validate:"required,oneof=order_confirmation order_shipped order_delivered order_cancelled payment_received payment_failed review_reminder wishlist_price_drop back_in_stock abandoned_cart coupon_available account_security system"`
	Title         string    `bson:"title" json:"title" validate:"required,min=2,max=200"`
	Message       string    `bson:"message" json:"message" validate:"required,max=1000"`

	// Action
	ActionURL     string    `bson:"action_url" json:"action_url" validate:"max=500"`
	ActionText    string    `bson:"action_text" json:"action_text" validate:"max=50"`

	// Related entity
	RelatedType   string               `bson:"related_type" json:"related_type" validate:"oneof=order product review cart coupon"`
	RelatedID     *primitive.ObjectID  `bson:"related_id,omitempty" json:"related_id,omitempty"`

	// Status
	IsRead        bool       `bson:"is_read" json:"is_read"`
	ReadAt        *time.Time `bson:"read_at,omitempty" json:"read_at,omitempty"`

	// Channels
	PushEnabled   bool `bson:"push_enabled" json:"push_enabled"`
	EmailEnabled  bool `bson:"email_enabled" json:"email_enabled"`
	SMSEnabled    bool `bson:"sms_enabled" json:"sms_enabled"`

	// Scheduling
	ScheduledFor  *time.Time `bson:"scheduled_for,omitempty" json:"scheduled_for,omitempty"`
	SentAt        *time.Time `bson:"sent_at,omitempty" json:"sent_at,omitempty"`

	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
}

func (Notification) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("customer_id", schema.Required(),
			schema.HelpText("Customer receiving the notification")),

		// Details
		schema.StringField("type", schema.Required(),
			schema.HelpText("Notification type")),
		schema.StringField("title", schema.Required(), schema.MinLength(2), schema.MaxLength(200),
			schema.HelpText("Notification title")),
		schema.StringField("message", schema.Required(), schema.MaxLength(1000),
			schema.HelpText("Notification message")),

		// Action
		schema.StringField("action_url", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Action URL")),
		schema.StringField("action_text", schema.Optional(), schema.MaxLength(50),
			schema.HelpText("Action button text")),

		// Related entity
		schema.StringField("related_type", schema.Optional(),
			schema.HelpText("Related entity type")),
		schema.ObjectIDField("related_id", schema.Optional(),
			schema.HelpText("Related entity ID")),

		// Status
		schema.BoolField("is_read", schema.Default(false)),
		schema.TimeField("read_at", schema.Optional()),

		// Channels
		schema.BoolField("push_enabled", schema.Default(true)),
		schema.BoolField("email_enabled", schema.Default(true)),
		schema.BoolField("sms_enabled", schema.Default(false)),

		// Scheduling
		schema.TimeField("scheduled_for", schema.Optional()),
		schema.TimeField("sent_at", schema.Optional()),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Notification) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "notifications",
		VerboseName:       "Notification",
		VerboseNamePlural: "Notifications",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_customer_notification", "customer_id", "-created_at"),
			schema.IndexOn("idx_customer_unread", "customer_id", "is_read", "-created_at"),
			schema.IndexOn("idx_type", "type"),
			schema.IndexOn("idx_scheduled", "scheduled_for"),
			schema.IndexOn("idx_created_at", "-created_at"),
		},
	}
}

func (Notification) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Notification) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// CustomerActivity tracks user activities on the site
type CustomerActivity struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID    primitive.ObjectID `bson:"customer_id" json:"customer_id" validate:"required"`

	// Activity details
	ActivityType  string `bson:"activity_type" json:"activity_type" validate:"required,oneof=page_view product_view add_to_cart remove_from_cart checkout_start checkout_complete search wishlist_add wishlist_remove review_submit coupon_apply login logout password_change profile_update address_add address_change payment_method_add subscription_start subscription_end"`

	// Target entity
	EntityType    string               `bson:"entity_type" json:"entity_type" validate:"oneof=product category cart order search query review coupon address payment subscription"`
	EntityID      *primitive.ObjectID  `bson:"entity_id,omitempty" json:"entity_id,omitempty"`

	// Activity data
	Data          map[string]interface{} `bson:"data" json:"data"`

	// Context
	SessionID     string    `bson:"session_id" json:"session_id" validate:"max=100"`
	UserAgent     string    `bson:"user_agent" json:"user_agent" validate:"max=500"`
	IPAddress     string    `bson:"ip_address" json:"ip_address" validate:"max=45"`

	Timestamp     time.Time `bson:"timestamp" json:"timestamp"`
}

func (CustomerActivity) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("customer_id", schema.Required(),
			schema.HelpText("Customer who performed the activity")),

		// Activity type
		schema.StringField("activity_type", schema.Required(),
			schema.HelpText("Type of activity")),

		// Target entity
		schema.StringField("entity_type", schema.Optional(),
			schema.HelpText("Entity type involved")),
		schema.ObjectIDField("entity_id", schema.Optional(),
			schema.HelpText("Entity ID involved")),

		// Activity data
		schema.MapField("data", schema.Optional(),
			schema.HelpText("Additional activity data")),

		// Context
		schema.StringField("session_id", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Session ID")),
		schema.StringField("user_agent", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Browser user agent")),
		schema.StringField("ip_address", schema.Optional(), schema.MaxLength(45),
			schema.HelpText("User IP address")),

		// Timestamp
		schema.TimeField("timestamp", schema.Required(),
			schema.HelpText("When the activity occurred")),
	}
}

func (CustomerActivity) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "customer_activities",
		VerboseName:       "Customer Activity",
		VerboseNamePlural: "Customer Activities",
		OrderBy:           []string{"-timestamp"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_customer_activity", "customer_id", "-timestamp"),
			schema.IndexOn("idx_activity_type", "activity_type", "-timestamp"),
			schema.IndexOn("idx_entity", "entity_type", "entity_id"),
			schema.IndexOn("idx_session", "session_id", "-timestamp"),
			schema.IndexOn("idx_timestamp", "-timestamp"),
		},
	}
}

func (CustomerActivity) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (CustomerActivity) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// AbandonedCartReminder tracks abandoned cart email reminders
type AbandonedCartReminder struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CartID          primitive.ObjectID `bson:"cart_id" json:"cart_id" validate:"required"`
	CustomerID      *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty"`
	GuestEmail      string              `bson:"guest_email" json:"guest_email" validate:"email,max=100"`

	// Reminder details
	ReminderNumber  int        `bson:"reminder_number" json:"reminder_number"`
	ReminderSentAt  *time.Time `bson:"reminder_sent_at,omitempty" json:"reminder_sent_at,omitempty"`
	ReminderOpenedAt *time.Time `bson:"reminder_opened_at,omitempty" json:"reminder_opened_at,omitempty"`
	ReminderClickedAt *time.Time `bson:"reminder_clicked_at,omitempty" json:"reminder_clicked_at,omitempty"`

	// Recovery
	RecoveredAt     *time.Time `bson:"recovered_at,omitempty" json:"recovered_at,omitempty"`
	RecoveredOrderID *primitive.ObjectID `bson:"recovered_order_id,omitempty" json:"recovered_order_id,omitempty"`

	// Status
	Status          string    `bson:"status" json:"status" validate:"oneof=pending sent opened clicked recovered cancelled"`

	CreatedAt       time.Time `bson:"created_at" json:"created_at"`
}

func (AbandonedCartReminder) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("cart_id", schema.Required(),
			schema.HelpText("Abandoned cart ID")),
		schema.ObjectIDField("customer_id", schema.Optional(),
			schema.HelpText("Customer ID (null for guest carts)")),
		schema.StringField("guest_email", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Guest email address")),

		// Reminder details
		schema.IntField("reminder_number", schema.Required(),
			schema.HelpText("1st, 2nd, 3rd reminder")),
		schema.TimeField("reminder_sent_at", schema.Optional()),
		schema.TimeField("reminder_opened_at", schema.Optional()),
		schema.TimeField("reminder_clicked_at", schema.Optional()),

		// Recovery
		schema.TimeField("recovered_at", schema.Optional()),
		schema.ObjectIDField("recovered_order_id", schema.Optional(),
			schema.HelpText("Order that recovered the cart")),

		// Status
		schema.StringField("status", schema.Required(),
			schema.HelpText("Status: pending, sent, opened, clicked, recovered, cancelled")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (AbandonedCartReminder) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "abandoned_cart_reminders",
		VerboseName:       "Abandoned Cart Reminder",
		VerboseNamePlural: "Abandoned Cart Reminders",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_cart_reminder", "cart_id", "reminder_number"),
			schema.IndexOn("idx_customer_reminder", "customer_id", "-created_at"),
			schema.IndexOn("idx_status", "status"),
			schema.IndexOn("idx_sent_at", "reminder_sent_at"),
			schema.IndexOn("idx_created_at", "-created_at"),
		},
	}
}

func (AbandonedCartReminder) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (AbandonedCartReminder) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// UserSegment represents user segmentation rules
type UserSegment struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string              `bson:"name" json:"name" validate:"required,min=2,max=100"`
	Description   string              `bson:"description" json:"description" validate:"max=500"`

	// Segment type
	Type          string `bson:"type" json:"type" validate:"required,oneof=static dynamic behavioral predictive"`

	// Criteria (flexible JSON for different segment types)
	Criteria      map[string]interface{} `bson:"criteria" json:"criteria"`

	// Rules
	Rules         []SegmentRule `bson:"rules" json:"rules"`

	// Membership
	CustomerIDs   []primitive.ObjectID `bson:"customer_ids" json:"customer_ids"`

	// Stats
	CustomerCount int `bson:"customer_count" json:"customer_count"`

	// Status
	IsActive      bool `bson:"is_active" json:"is_active"`

	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updated_at"`
}

// SegmentRule defines a rule for user segmentation
type SegmentRule struct {
	Field          string      `bson:"field" json:"field"`
	Operator       string      `bson:"operator" json:"operator" validate:"oneof=equals not_equals greater_than less_than contains not_contains in not_in between is_set is_not_set"`
	Value          interface{} `bson:"value" json:"value"`
	LogicalOperator string     `bson:"logical_operator" json:"logical_operator" validate:"oneof=and or"`
}

func (UserSegment) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("name", schema.Required(), schema.MinLength(2), schema.MaxLength(100),
			schema.HelpText("Segment name")),
		schema.StringField("description", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Segment description")),

		// Segment type
		schema.StringField("type", schema.Required(),
			schema.HelpText("Segment type: static, dynamic, behavioral, predictive")),

		// Criteria
		schema.MapField("criteria", schema.Optional(),
			schema.HelpText("Flexible criteria for segment matching")),

		// Rules
		schema.JSONField("rules", schema.Optional(),
			schema.HelpText("Segment rules")),

		// Membership
		schema.ObjectIDArrayField("customer_ids",
			schema.HelpText("Customer IDs in this segment (for static segments)")),

		// Stats
		schema.IntField("customer_count", schema.Default(0),
			schema.HelpText("Number of customers in segment")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (UserSegment) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "user_segments",
		VerboseName:       "User Segment",
		VerboseNamePlural: "User Segments",
		OrderBy:           []string{"name"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_segment_type", "type"),
			schema.IndexOn("idx_segment_active", "is_active"),
			schema.IndexOn("idx_segment_count", "-customer_count"),
			schema.IndexOn("idx_created_at", "-created_at"),
		},
	}
}

func (UserSegment) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (UserSegment) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
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

// RecentlyViewed tracks products that users have viewed
type RecentlyViewed struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID  *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty"`
	GuestID     string              `bson:"guest_id" json:"guest_id" validate:"max=100"`
	ProductID   primitive.ObjectID `bson:"product_id" json:"product_id" validate:"required"`
	VariantID   *primitive.ObjectID `bson:"variant_id,omitempty" json:"variant_id,omitempty"`

	// Tracking
	ViewedAt    time.Time `bson:"viewed_at" json:"viewed_at"`
	SessionID   string    `bson:"session_id" json:"session_id" validate:"max=100"`
	UserAgent   string    `bson:"user_agent" json:"user_agent" validate:"max=500"`
	IPAddress   string    `bson:"ip_address" json:"ip_address" validate:"max=45"`

	// Context
	Source      string    `bson:"source" json:"source" validate:"oneof=search browse category product recommended email social direct"`
	RefererURL  string    `bson:"referer_url" json:"referer_url" validate:"max=500"`
}

func (RecentlyViewed) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("customer_id", schema.Optional(),
			schema.HelpText("Customer ID (null for guests)")),
		schema.StringField("guest_id", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Guest session ID")),
		schema.ObjectIDField("product_id", schema.Required(),
			schema.HelpText("Viewed product ID")),
		schema.ObjectIDField("variant_id", schema.Optional(),
			schema.HelpText("Viewed variant ID")),

		// Tracking
		schema.TimeField("viewed_at", schema.Required(),
			schema.HelpText("When the product was viewed")),
		schema.StringField("session_id", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("User session ID")),
		schema.StringField("user_agent", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Browser user agent")),
		schema.StringField("ip_address", schema.Optional(), schema.MaxLength(45),
			schema.HelpText("User IP address")),

		// Context
		schema.StringField("source", schema.Optional(),
			schema.HelpText("Source: search, browse, category, product, recommended, email, social, direct")),
		schema.StringField("referer_url", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Referring URL")),
	}
}

func (RecentlyViewed) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "recently_viewed",
		VerboseName:       "Recently Viewed",
		VerboseNamePlural: "Recently Viewed",
		OrderBy:           []string{"-viewed_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_customer_viewed", "customer_id", "-viewed_at"),
			schema.IndexOn("idx_guest_viewed", "guest_id", "-viewed_at"),
			schema.IndexOn("idx_product_viewed", "product_id", "-viewed_at"),
			schema.IndexOn("idx_session_viewed", "session_id", "-viewed_at"),
		},
	}
}

func (RecentlyViewed) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (RecentlyViewed) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Auto-set viewed_at if not set
			return nil
		},
	}
}

// ProductComparison allows users to compare products
type ProductComparison struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID  *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty"`
	GuestID     string              `bson:"guest_id" json:"guest_id" validate:"max=100"`
	Name        string              `bson:"name" json:"name" validate:"required,min=2,max=100"`

	// Products being compared
	ProductIDs  []primitive.ObjectID `bson:"product_ids" json:"product_ids"`

	// Settings
	IsPublic    bool   `bson:"is_public" json:"is_public"`
	ShareToken  string `bson:"share_token" json:"share_token" validate:"max=100"`

	// Stats
	ViewCount   int       `bson:"view_count" json:"view_count"`

	// Timestamps
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

func (ProductComparison) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("customer_id", schema.Optional(),
			schema.HelpText("Customer ID (null for guest comparisons)")),
		schema.StringField("guest_id", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Guest session ID")),
		schema.StringField("name", schema.Required(), schema.MinLength(2), schema.MaxLength(100),
			schema.HelpText("Comparison name")),

		// Products
		schema.ObjectIDArrayField("product_ids",
			schema.HelpText("Products being compared")),

		// Settings
		schema.BoolField("is_public", schema.Default(false),
			schema.HelpText("Whether the comparison is publicly shareable")),
		schema.StringField("share_token", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Unique token for sharing")),

		// Stats
		schema.IntField("view_count", schema.Default(0),
			schema.HelpText("Number of times this comparison was viewed")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ProductComparison) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "product_comparisons",
		VerboseName:       "Product Comparison",
		VerboseNamePlural: "Product Comparisons",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_customer_comparison", "customer_id", "-created_at"),
			schema.IndexOn("idx_guest_comparison", "guest_id", "-created_at"),
			schema.IndexOn("idx_share_token", "share_token"),
			schema.IndexOn("idx_created_at", "-created_at"),
		},
	}
}

func (ProductComparison) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ProductComparison) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// Notification represents user notifications
type Notification struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID    primitive.ObjectID `bson:"customer_id" json:"customer_id" validate:"required"`

	// Notification details
	Type          string    `bson:"type" json:"type" validate:"required,oneof=order_confirmation order_shipped order_delivered order_cancelled payment_received payment_failed review_reminder wishlist_price_drop back_in_stock abandoned_cart coupon_available account_security system"`
	Title         string    `bson:"title" json:"title" validate:"required,min=2,max=200"`
	Message       string    `bson:"message" json:"message" validate:"required,max=1000"`

	// Action
	ActionURL     string    `bson:"action_url" json:"action_url" validate:"max=500"`
	ActionText    string    `bson:"action_text" json:"action_text" validate:"max=50"`

	// Related entity
	RelatedType   string               `bson:"related_type" json:"related_type" validate:"oneof=order product review cart coupon"`
	RelatedID     *primitive.ObjectID  `bson:"related_id,omitempty" json:"related_id,omitempty"`

	// Status
	IsRead        bool       `bson:"is_read" json:"is_read"`
	ReadAt        *time.Time `bson:"read_at,omitempty" json:"read_at,omitempty"`

	// Channels
	PushEnabled   bool `bson:"push_enabled" json:"push_enabled"`
	EmailEnabled  bool `bson:"email_enabled" json:"email_enabled"`
	SMSEnabled    bool `bson:"sms_enabled" json:"sms_enabled"`

	// Scheduling
	ScheduledFor  *time.Time `bson:"scheduled_for,omitempty" json:"scheduled_for,omitempty"`
	SentAt        *time.Time `bson:"sent_at,omitempty" json:"sent_at,omitempty"`

	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
}

func (Notification) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("customer_id", schema.Required(),
			schema.HelpText("Customer receiving the notification")),

		// Details
		schema.StringField("type", schema.Required(),
			schema.HelpText("Notification type")),
		schema.StringField("title", schema.Required(), schema.MinLength(2), schema.MaxLength(200),
			schema.HelpText("Notification title")),
		schema.StringField("message", schema.Required(), schema.MaxLength(1000),
			schema.HelpText("Notification message")),

		// Action
		schema.StringField("action_url", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Action URL")),
		schema.StringField("action_text", schema.Optional(), schema.MaxLength(50),
			schema.HelpText("Action button text")),

		// Related entity
		schema.StringField("related_type", schema.Optional(),
			schema.HelpText("Related entity type")),
		schema.ObjectIDField("related_id", schema.Optional(),
			schema.HelpText("Related entity ID")),

		// Status
		schema.BoolField("is_read", schema.Default(false)),
		schema.TimeField("read_at", schema.Optional()),

		// Channels
		schema.BoolField("push_enabled", schema.Default(true)),
		schema.BoolField("email_enabled", schema.Default(true)),
		schema.BoolField("sms_enabled", schema.Default(false)),

		// Scheduling
		schema.TimeField("scheduled_for", schema.Optional()),
		schema.TimeField("sent_at", schema.Optional()),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Notification) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "notifications",
		VerboseName:       "Notification",
		VerboseNamePlural: "Notifications",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_customer_notification", "customer_id", "-created_at"),
			schema.IndexOn("idx_customer_unread", "customer_id", "is_read", "-created_at"),
			schema.IndexOn("idx_type", "type"),
			schema.IndexOn("idx_scheduled", "scheduled_for"),
			schema.IndexOn("idx_created_at", "-created_at"),
		},
	}
}

func (Notification) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Notification) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// CustomerActivity tracks user activities on the site
type CustomerActivity struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID    primitive.ObjectID `bson:"customer_id" json:"customer_id" validate:"required"`

	// Activity details
	ActivityType  string `bson:"activity_type" json:"activity_type" validate:"required,oneof=page_view product_view add_to_cart remove_from_cart checkout_start checkout_complete search wishlist_add wishlist_remove review_submit coupon_apply login logout password_change profile_update address_add address_change payment_method_add subscription_start subscription_end"`

	// Target entity
	EntityType    string               `bson:"entity_type" json:"entity_type" validate:"oneof=product category cart order search query review coupon address payment subscription"`
	EntityID      *primitive.ObjectID  `bson:"entity_id,omitempty" json:"entity_id,omitempty"`

	// Activity data
	Data          map[string]interface{} `bson:"data" json:"data"`

	// Context
	SessionID     string    `bson:"session_id" json:"session_id" validate:"max=100"`
	UserAgent     string    `bson:"user_agent" json:"user_agent" validate:"max=500"`
	IPAddress     string    `bson:"ip_address" json:"ip_address" validate:"max=45"`

	Timestamp     time.Time `bson:"timestamp" json:"timestamp"`
}

func (CustomerActivity) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("customer_id", schema.Required(),
			schema.HelpText("Customer who performed the activity")),

		// Activity type
		schema.StringField("activity_type", schema.Required(),
			schema.HelpText("Type of activity")),

		// Target entity
		schema.StringField("entity_type", schema.Optional(),
			schema.HelpText("Entity type involved")),
		schema.ObjectIDField("entity_id", schema.Optional(),
			schema.HelpText("Entity ID involved")),

		// Activity data
		schema.MapField("data", schema.Optional(),
			schema.HelpText("Additional activity data")),

		// Context
		schema.StringField("session_id", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Session ID")),
		schema.StringField("user_agent", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Browser user agent")),
		schema.StringField("ip_address", schema.Optional(), schema.MaxLength(45),
			schema.HelpText("User IP address")),

		// Timestamp
		schema.TimeField("timestamp", schema.Required(),
			schema.HelpText("When the activity occurred")),
	}
}

func (CustomerActivity) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "customer_activities",
		VerboseName:       "Customer Activity",
		VerboseNamePlural: "Customer Activities",
		OrderBy:           []string{"-timestamp"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_customer_activity", "customer_id", "-timestamp"),
			schema.IndexOn("idx_activity_type", "activity_type", "-timestamp"),
			schema.IndexOn("idx_entity", "entity_type", "entity_id"),
			schema.IndexOn("idx_session", "session_id", "-timestamp"),
			schema.IndexOn("idx_timestamp", "-timestamp"),
		},
	}
}

func (CustomerActivity) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (CustomerActivity) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// AbandonedCartReminder tracks abandoned cart email reminders
type AbandonedCartReminder struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CartID          primitive.ObjectID `bson:"cart_id" json:"cart_id" validate:"required"`
	CustomerID      *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty"`
	GuestEmail      string              `bson:"guest_email" json:"guest_email" validate:"email,max=100"`

	// Reminder details
	ReminderNumber  int        `bson:"reminder_number" json:"reminder_number"`
	ReminderSentAt  *time.Time `bson:"reminder_sent_at,omitempty" json:"reminder_sent_at,omitempty"`
	ReminderOpenedAt *time.Time `bson:"reminder_opened_at,omitempty" json:"reminder_opened_at,omitempty"`
	ReminderClickedAt *time.Time `bson:"reminder_clicked_at,omitempty" json:"reminder_clicked_at,omitempty"`

	// Recovery
	RecoveredAt     *time.Time `bson:"recovered_at,omitempty" json:"recovered_at,omitempty"`
	RecoveredOrderID *primitive.ObjectID `bson:"recovered_order_id,omitempty" json:"recovered_order_id,omitempty"`

	// Status
	Status          string    `bson:"status" json:"status" validate:"oneof=pending sent opened clicked recovered cancelled"`

	CreatedAt       time.Time `bson:"created_at" json:"created_at"`
}

func (AbandonedCartReminder) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.ObjectIDField("cart_id", schema.Required(),
			schema.HelpText("Abandoned cart ID")),
		schema.ObjectIDField("customer_id", schema.Optional(),
			schema.HelpText("Customer ID (null for guest carts)")),
		schema.StringField("guest_email", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Guest email address")),

		// Reminder details
		schema.IntField("reminder_number", schema.Required(),
			schema.HelpText("1st, 2nd, 3rd reminder")),
		schema.TimeField("reminder_sent_at", schema.Optional()),
		schema.TimeField("reminder_opened_at", schema.Optional()),
		schema.TimeField("reminder_clicked_at", schema.Optional()),

		// Recovery
		schema.TimeField("recovered_at", schema.Optional()),
		schema.ObjectIDField("recovered_order_id", schema.Optional(),
			schema.HelpText("Order that recovered the cart")),

		// Status
		schema.StringField("status", schema.Required(),
			schema.HelpText("Status: pending, sent, opened, clicked, recovered, cancelled")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (AbandonedCartReminder) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "abandoned_cart_reminders",
		VerboseName:       "Abandoned Cart Reminder",
		VerboseNamePlural: "Abandoned Cart Reminders",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_cart_reminder", "cart_id", "reminder_number"),
			schema.IndexOn("idx_customer_reminder", "customer_id", "-created_at"),
			schema.IndexOn("idx_status", "status"),
			schema.IndexOn("idx_sent_at", "reminder_sent_at"),
			schema.IndexOn("idx_created_at", "-created_at"),
		},
	}
}

func (AbandonedCartReminder) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (AbandonedCartReminder) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// UserSegment represents user segmentation rules
type UserSegment struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string              `bson:"name" json:"name" validate:"required,min=2,max=100"`
	Description   string              `bson:"description" json:"description" validate:"max=500"`

	// Segment type
	Type          string `bson:"type" json:"type" validate:"required,oneof=static dynamic behavioral predictive"`

	// Criteria (flexible JSON for different segment types)
	Criteria      map[string]interface{} `bson:"criteria" json:"criteria"`

	// Rules
	Rules         []SegmentRule `bson:"rules" json:"rules"`

	// Membership
	CustomerIDs   []primitive.ObjectID `bson:"customer_ids" json:"customer_ids"`

	// Stats
	CustomerCount int `bson:"customer_count" json:"customer_count"`

	// Status
	IsActive      bool `bson:"is_active" json:"is_active"`

	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updated_at"`
}

// SegmentRule defines a rule for user segmentation
type SegmentRule struct {
	Field          string      `bson:"field" json:"field"`
	Operator       string      `bson:"operator" json:"operator" validate:"oneof=equals not_equals greater_than less_than contains not_contains in not_in between is_set is_not_set"`
	Value          interface{} `bson:"value" json:"value"`
	LogicalOperator string     `bson:"logical_operator" json:"logical_operator" validate:"oneof=and or"`
}

func (UserSegment) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("name", schema.Required(), schema.MinLength(2), schema.MaxLength(100),
			schema.HelpText("Segment name")),
		schema.StringField("description", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Segment description")),

		// Segment type
		schema.StringField("type", schema.Required(),
			schema.HelpText("Segment type: static, dynamic, behavioral, predictive")),

		// Criteria
		schema.MapField("criteria", schema.Optional(),
			schema.HelpText("Flexible criteria for segment matching")),

		// Rules
		schema.JSONField("rules", schema.Optional(),
			schema.HelpText("Segment rules")),

		// Membership
		schema.ObjectIDArrayField("customer_ids",
			schema.HelpText("Customer IDs in this segment (for static segments)")),

		// Stats
		schema.IntField("customer_count", schema.Default(0),
			schema.HelpText("Number of customers in segment")),

		// Status
		schema.BoolField("is_active", schema.Default(true)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (UserSegment) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "user_segments",
		VerboseName:       "User Segment",
		VerboseNamePlural: "User Segments",
		OrderBy:           []string{"name"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_segment_type", "type"),
			schema.IndexOn("idx_segment_active", "is_active"),
			schema.IndexOn("idx_segment_count", "-customer_count"),
			schema.IndexOn("idx_created_at", "-created_at"),
		},
	}
}

func (UserSegment) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (UserSegment) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

