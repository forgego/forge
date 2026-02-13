package engagement

import (
	"github.com/forgego/forge/schema"
)

// RecentlyViewed tracks recently viewed products by customers
type RecentlyViewed struct {
	schema.BaseSchema
	Id          int64  `json:"id" db:"id"`
	CustomerID  int64  `json:"customer_id" db:"customer_id"`
	ProductID   int64  `json:"product_id" db:"product_id"`
	ViewedAt    string `json:"viewed_at" db:"viewed_at"`
	ViewCount   int32  `json:"view_count" db:"view_count"`
	SessionID   string `json:"session_id" db:"session_id"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

func (RecentlyViewed) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),
		schema.Int64Field("product_id", schema.Required(),
			schema.VerboseName("Product")),
		schema.TimeField("viewed_at", schema.Required(),
			schema.VerboseName("Last Viewed At"),
			schema.HelpText("Timestamp of last view")),
		schema.Int32Field("view_count", schema.Default(1),
			schema.VerboseName("View Count"),
			schema.HelpText("Number of times viewed")),
		schema.StringField("session_id", schema.MaxLength(100), schema.Optional(),
			schema.VerboseName("Session ID"),
			schema.HelpText("Browser session ID")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (RecentlyViewed) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "recently_viewed",
		VerboseName:       "Recently Viewed Product",
		VerboseNamePlural: "Recently Viewed Products",
		OrderBy:           []string{"-viewed_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_recently_viewed_customer", "customer_id", "viewed_at"),
			schema.IndexOn("idx_recently_viewed_product", "product_id"),
			schema.UniqueIndexOn("idx_recently_viewed_unique", "customer_id", "product_id"),
		},
	}
}

func (RecentlyViewed) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("recently_viewed")),
		schema.ForeignKeyField("product_id", "Product",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("viewed_by")),
	}
}

func (RecentlyViewed) Hooks() *schema.ModelHooks {
	return nil
}

// ProductComparison allows customers to compare products
type ProductComparison struct {
	schema.BaseSchema
	Id         int64  `json:"id" db:"id"`
	CustomerID int64  `json:"customer_id" db:"customer_id"`
	Name       string `json:"name" db:"name"`
	IsPublic   bool   `json:"is_public" db:"is_public"`
	CreatedAt  string `json:"created_at" db:"created_at"`
	UpdatedAt  string `json:"updated_at" db:"updated_at"`
}

func (ProductComparison) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),
		schema.StringField("name", schema.MaxLength(200), schema.Optional(),
			schema.HelpText("Comparison list name")),
		schema.BoolField("is_public", schema.Default(false),
			schema.VerboseName("Public"),
			schema.HelpText("Can be shared with others")),
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
			schema.IndexOn("idx_product_comparison_customer", "customer_id"),
		},
	}
}

func (ProductComparison) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("comparisons")),
		// Many-to-many with Product through ComparisonItem (implement separately if needed)
	}
}

func (ProductComparison) Hooks() *schema.ModelHooks {
	return nil
}

// Notification represents customer notifications
type Notification struct {
	schema.BaseSchema
	Id           int64  `json:"id" db:"id"`
	CustomerID   int64  `json:"customer_id" db:"customer_id"`
	Title        string `json:"title" db:"title"`
	Message      string `json:"message" db:"message"`
	Type         string `json:"type" db:"type"`
	Priority     string `json:"priority" db:"priority"`
	IsRead       bool   `json:"is_read" db:"is_read"`
	ReadAt       string `json:"read_at" db:"read_at"`
	ActionURL    string `json:"action_url" db:"action_url"`
	ActionLabel  string `json:"action_label" db:"action_label"`
	RelatedType  string `json:"related_type" db:"related_type"`
	RelatedID    int64  `json:"related_id" db:"related_id"`
	ExpiresAt    string `json:"expires_at" db:"expires_at"`
	CreatedAt    string `json:"created_at" db:"created_at"`
	UpdatedAt    string `json:"updated_at" db:"updated_at"`
}

func (Notification) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),
		schema.StringField("title", schema.Required(), schema.MaxLength(200),
			schema.HelpText("Notification title")),
		schema.TextField("message", schema.Required(),
			schema.HelpText("Notification message")),
		schema.StringField("type", schema.MaxLength(50), schema.Default("info"),
			schema.HelpText("Type: info, success, warning, error, order, shipping")),
		schema.StringField("priority", schema.MaxLength(20), schema.Default("normal"),
			schema.HelpText("Priority: low, normal, high, urgent")),
		schema.BoolField("is_read", schema.Default(false),
			schema.VerboseName("Read")),
		schema.TimeField("read_at", schema.Optional(),
			schema.VerboseName("Read At")),
		schema.StringField("action_url", schema.MaxLength(500), schema.Optional(),
			schema.VerboseName("Action URL"),
			schema.HelpText("URL to navigate when clicked")),
		schema.StringField("action_label", schema.MaxLength(100), schema.Optional(),
			schema.VerboseName("Action Label"),
			schema.HelpText("Button label for action")),
		schema.StringField("related_type", schema.MaxLength(50), schema.Optional(),
			schema.VerboseName("Related Type"),
			schema.HelpText("Related entity type (order, product, etc.)")),
		schema.Int64Field("related_id", schema.Optional(),
			schema.VerboseName("Related ID"),
			schema.HelpText("Related entity ID")),
		schema.TimeField("expires_at", schema.Optional(),
			schema.VerboseName("Expires At"),
			schema.HelpText("Notification expiration time")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Notification) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "notifications",
		VerboseName:       "Notification",
		VerboseNamePlural: "Notifications",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_notification_customer", "customer_id", "is_read"),
			schema.IndexOn("idx_notification_type", "type"),
			schema.IndexOn("idx_notification_created", "created_at"),
		},
	}
}

func (Notification) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("notifications")),
	}
}

func (Notification) Hooks() *schema.ModelHooks {
	return nil
}

// CustomerActivity tracks customer activities for analytics
type CustomerActivity struct {
	schema.BaseSchema
	Id           int64  `json:"id" db:"id"`
	CustomerID   int64  `json:"customer_id" db:"customer_id"`
	ActivityType string `json:"activity_type" db:"activity_type"`
	Description  string `json:"description" db:"description"`
	EntityType   string `json:"entity_type" db:"entity_type"`
	EntityID     int64  `json:"entity_id" db:"entity_id"`
	IPAddress    string `json:"ip_address" db:"ip_address"`
	UserAgent    string `json:"user_agent" db:"user_agent"`
	SessionID    string `json:"session_id" db:"session_id"`
	Metadata     string `json:"metadata" db:"metadata"`
	CreatedAt    string `json:"created_at" db:"created_at"`
}

func (CustomerActivity) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("customer_id", schema.Optional(),
			schema.VerboseName("Customer"),
			schema.HelpText("Null for anonymous users")),
		schema.StringField("activity_type", schema.Required(), schema.MaxLength(50),
			schema.VerboseName("Activity Type"),
			schema.HelpText("Type: page_view, product_view, add_to_cart, purchase, etc.")),
		schema.TextField("description", schema.Optional(),
			schema.HelpText("Activity description")),
		schema.StringField("entity_type", schema.MaxLength(50), schema.Optional(),
			schema.VerboseName("Entity Type"),
			schema.HelpText("Related entity type")),
		schema.Int64Field("entity_id", schema.Optional(),
			schema.VerboseName("Entity ID"),
			schema.HelpText("Related entity ID")),
		schema.StringField("ip_address", schema.MaxLength(45), schema.Optional(),
			schema.VerboseName("IP Address"),
			schema.HelpText("IPv4 or IPv6 address")),
		schema.StringField("user_agent", schema.MaxLength(500), schema.Optional(),
			schema.VerboseName("User Agent")),
		schema.StringField("session_id", schema.MaxLength(100), schema.Optional(),
			schema.VerboseName("Session ID")),
		schema.TextField("metadata", schema.Optional(),
			schema.HelpText("Additional data in JSON format")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (CustomerActivity) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "customer_activities",
		VerboseName:       "Customer Activity",
		VerboseNamePlural: "Customer Activities",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_customer_activity_customer", "customer_id", "created_at"),
			schema.IndexOn("idx_customer_activity_type", "activity_type"),
			schema.IndexOn("idx_customer_activity_entity", "entity_type", "entity_id"),
			schema.IndexOn("idx_customer_activity_session", "session_id"),
		},
	}
}

func (CustomerActivity) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("activities")),
	}
}

func (CustomerActivity) Hooks() *schema.ModelHooks {
	return nil
}

// AbandonedCartReminder tracks reminders sent for abandoned carts
type AbandonedCartReminder struct {
	schema.BaseSchema
	Id           int64  `json:"id" db:"id"`
	CartID       int64  `json:"cart_id" db:"cart_id"`
	CustomerID   int64  `json:"customer_id" db:"customer_id"`
	ReminderType string `json:"reminder_type" db:"reminder_type"`
	SentAt       string `json:"sent_at" db:"sent_at"`
	Status       string `json:"status" db:"status"`
	EmailAddress string `json:"email_address" db:"email_address"`
	Converted    bool   `json:"converted" db:"converted"`
	ConvertedAt  string `json:"converted_at" db:"converted_at"`
	CreatedAt    string `json:"created_at" db:"created_at"`
	UpdatedAt    string `json:"updated_at" db:"updated_at"`
}

func (AbandonedCartReminder) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("cart_id", schema.Required(),
			schema.VerboseName("Cart")),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),
		schema.StringField("reminder_type", schema.MaxLength(50), schema.Default("email"),
			schema.VerboseName("Reminder Type"),
			schema.HelpText("Type: email, sms, push")),
		schema.TimeField("sent_at", schema.Required(),
			schema.VerboseName("Sent At")),
		schema.StringField("status", schema.MaxLength(20), schema.Default("sent"),
			schema.HelpText("Status: sent, delivered, opened, clicked, failed")),
		schema.StringField("email_address", schema.MaxLength(255), schema.Optional(),
			schema.VerboseName("Email Address")),
		schema.BoolField("converted", schema.Default(false),
			schema.VerboseName("Converted"),
			schema.HelpText("Customer completed purchase")),
		schema.TimeField("converted_at", schema.Optional(),
			schema.VerboseName("Converted At")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (AbandonedCartReminder) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "abandoned_cart_reminders",
		VerboseName:       "Abandoned Cart Reminder",
		VerboseNamePlural: "Abandoned Cart Reminders",
		OrderBy:           []string{"-sent_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_abandoned_cart_reminder_cart", "cart_id"),
			schema.IndexOn("idx_abandoned_cart_reminder_customer", "customer_id"),
			schema.IndexOn("idx_abandoned_cart_reminder_status", "status"),
			schema.IndexOn("idx_abandoned_cart_reminder_converted", "converted"),
		},
	}
}

func (AbandonedCartReminder) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("cart_id", "Cart",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("reminders")),
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("cart_reminders")),
	}
}

func (AbandonedCartReminder) Hooks() *schema.ModelHooks {
	return nil
}

// UserSegment represents customer segments for targeted marketing
type UserSegment struct {
	schema.BaseSchema
	Id          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Conditions  string `json:"conditions" db:"conditions"`
	IsActive    bool   `json:"is_active" db:"is_active"`
	IsDynamic   bool   `json:"is_dynamic" db:"is_dynamic"`
	Priority    int32  `json:"priority" db:"priority"`
	MemberCount int32  `json:"member_count" db:"member_count"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

func (UserSegment) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(200),
			schema.HelpText("Segment name (e.g., High Value Customers)")),
		schema.TextField("description", schema.Optional(),
			schema.HelpText("Segment description")),
		schema.TextField("conditions", schema.Optional(),
			schema.HelpText("Segment conditions in JSON format")),
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.BoolField("is_dynamic", schema.Default(true),
			schema.VerboseName("Dynamic"),
			schema.HelpText("Automatically update membership based on conditions")),
		schema.Int32Field("priority", schema.Default(0),
			schema.HelpText("Segment priority for overlapping segments")),
		schema.Int32Field("member_count", schema.Default(0),
			schema.VerboseName("Member Count"),
			schema.HelpText("Cached count of members")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (UserSegment) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "user_segments",
		VerboseName:       "User Segment",
		VerboseNamePlural: "User Segments",
		OrderBy:           []string{"-priority", "name"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_user_segment_active", "is_active"),
		},
	}
}

func (UserSegment) Relations() []schema.Relation {
	return []schema.Relation{
		// Many-to-many with Customer through SegmentMembership (implement separately if needed)
	}
}

func (UserSegment) Hooks() *schema.ModelHooks {
	return nil
}

// SegmentRule represents individual rules for user segments (placeholder for future implementation)
type SegmentRule struct {
	schema.BaseSchema
	Id         int64  `json:"id" db:"id"`
	SegmentID  int64  `json:"segment_id" db:"segment_id"`
	Field      string `json:"field" db:"field"`
	Operator   string `json:"operator" db:"operator"`
	Value      string `json:"value" db:"value"`
	LogicType  string `json:"logic_type" db:"logic_type"`
	SortOrder  int32  `json:"sort_order" db:"sort_order"`
	CreatedAt  string `json:"created_at" db:"created_at"`
	UpdatedAt  string `json:"updated_at" db:"updated_at"`
}

func (SegmentRule) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("segment_id", schema.Required(),
			schema.VerboseName("Segment")),
		schema.StringField("field", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Field to evaluate (e.g., total_spent, order_count)")),
		schema.StringField("operator", schema.MaxLength(20), schema.Required(),
			schema.HelpText("Operator: equals, not_equals, greater_than, less_than, contains, etc.")),
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

func (SegmentRule) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "segment_rules",
		VerboseName:       "Segment Rule",
		VerboseNamePlural: "Segment Rules",
		OrderBy:           []string{"segment_id", "sort_order"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_segment_rule_segment", "segment_id"),
		},
	}
}

func (SegmentRule) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("segment_id", "UserSegment",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("rules")),
	}
}

func (SegmentRule) Hooks() *schema.ModelHooks {
	return nil
}
