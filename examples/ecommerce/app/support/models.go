package support

import (
	"context"
	"time"

	"github.com/forgego/forge/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SupportTicket represents a customer support ticket
type SupportTicket struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	TicketNumber    string            `bson:"ticket_number" json:"ticket_number" db:"ticket_number"`
	CustomerID      primitive.ObjectID `bson:"customer_id" json:"customer_id" db:"customer_id"`

	// Ticket details
	Type            string            `bson:"type" json:"type" db:"type"`
	Subject         string            `bson:"subject" json:"subject" db:"subject"`
	Description     string            `bson:"description" json:"description" db:"description"`

	// Priority and status
	Priority        string            `bson:"priority" json:"priority" db:"priority"`
	Status         string            `bson:"status" json:"status" db:"status"`

	// Assignment
	AssignedTo      *primitive.ObjectID `bson:"assigned_to,omitempty" json:"assigned_to,omitempty" db:"assigned_to"`
	AssignedGroup   string            `bson:"assigned_group" json:"assigned_group" db:"assigned_group"`

	// Related entity
	RelatedType     string            `bson:"related_type" json:"related_type" db:"related_type"`
	RelatedID      primitive.ObjectID `bson:"related_id,omitempty" json:"related_id,omitempty" db:"related_id"`

	// Tags
	Tags            []string          `bson:"tags" json:"tags" db:"tags"`

	// Resolution
	Resolution      string            `bson:"resolution" json:"resolution" db:"resolution"`
	ResolvedAt     *time.Time        `bson:"resolved_at,omitempty" json:"resolved_at,omitempty" db:"resolved_at"`

	// Feedback
	CustomerSatisfaction *int         `bson:"customer_satisfaction,omitempty" json:"customer_satisfaction,omitempty" db:"customer_satisfaction"`
	CustomerFeedback string          `bson:"customer_feedback" json:"customer_feedback" db:"customer_feedback"`

	// Escalation
	IsEscalated    bool              `bson:"is_escalated" json:"is_escalated" db:"is_escalated"`
	EscalatedAt    *time.Time        `bson:"escalated_at,omitempty" json:"escalated_at,omitempty" db:"escalated_at"`
	EscalationReason string          `bson:"escalation_reason" json:"escalation_reason" db:"escalation_reason"`

	// Timestamps
	FirstResponseAt *time.Time       `bson:"first_response_at,omitempty" json:"first_response_at,omitempty" db:"first_response_at"`
	CreatedAt       time.Time         `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (SupportTicket) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("ticket_number", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique ticket number")),
		schema.StringField("customer_id", schema.Required(),
			schema.HelpText("Customer ID")),

		// Ticket details
		schema.StringField("type", schema.Required(),
			schema.HelpText("Ticket type: order_inquiry, product_inquiry, shipping_issue, payment_issue, return_refund, technical, account, billing, general, feedback, praise, suggestion, other")),
		schema.StringField("subject", schema.Required(), schema.MinLength(5), schema.MaxLength(200),
			schema.HelpText("Ticket subject")),
		schema.StringField("description", schema.Required(), schema.MinLength(10), schema.MaxLength(5000),
			schema.HelpText("Ticket description")),

		// Priority and status
		schema.StringField("priority", schema.Required(),
			schema.HelpText("Priority: low, medium, high, urgent")),
		schema.StringField("status", schema.Required(),
			schema.HelpText("Status: open, pending_customer_response, pending_merchant_response, resolved, closed, spam")),

		// Assignment
		schema.StringField("assigned_to", schema.Optional(),
			schema.HelpText("Assigned agent ID")),
		schema.StringField("assigned_group", schema.Optional(),
			schema.HelpText("Assigned group: general, billing, technical, shipping, management")),

		// Related entity
		schema.StringField("related_type", schema.Optional(),
			schema.HelpText("Related entity type: order, product, payment, shipment, return")),
		schema.StringField("related_id", schema.Optional(),
			schema.HelpText("Related entity ID")),

		// Tags
		schema.StringArrayField("tags", schema.Optional(),
			schema.HelpText("Ticket tags")),

		// Resolution
		schema.StringField("resolution", schema.Optional(), schema.MaxLength(2000),
			schema.HelpText("Resolution notes")),
		schema.TimeField("resolved_at", schema.Optional()),

		// Feedback
		schema.IntField("customer_satisfaction", schema.Optional(), schema.Min(1), schema.Max(5),
			schema.HelpText("Customer satisfaction rating 1-5")),
		schema.StringField("customer_feedback", schema.Optional(), schema.MaxLength(1000),
			schema.HelpText("Customer feedback")),

		// Escalation
		schema.BoolField("is_escalated", schema.Default(false)),
		schema.TimeField("escalated_at", schema.Optional()),
		schema.StringField("escalation_reason", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Escalation reason")),

		// Timestamps
		schema.TimeField("first_response_at", schema.Optional()),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (SupportTicket) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "support_tickets",
		VerboseName:       "Support Ticket",
		VerboseNamePlural: "Support Tickets",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_ticket_number", "ticket_number"),
			schema.IndexOn("idx_ticket_customer", "customer_id"),
			schema.IndexOn("idx_ticket_status", "status"),
			schema.IndexOn("idx_ticket_priority", "priority"),
			schema.IndexOn("idx_ticket_type", "type"),
			schema.IndexOn("idx_ticket_assigned", "assigned_to"),
			schema.IndexOn("idx_ticket_created", "created_at"),
		},
	}
}

func (SupportTicket) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (SupportTicket) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Auto-generate ticket number if not set
			// Validate ticket constraints
			return nil
		},
	}
}

// SupportMessage represents a message in a support ticket
type SupportMessage struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	TicketID        primitive.ObjectID `bson:"ticket_id" json:"ticket_id" db:"ticket_id"`

	// Sender
	SenderType      string            `bson:"sender_type" json:"sender_type" db:"sender_type"`
	CustomerID      *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty" db:"customer_id"`
	AgentID         *primitive.ObjectID `bson:"agent_id,omitempty" json:"agent_id,omitempty" db:"agent_id"`

	// Message
	Content         string            `bson:"content" json:"content" db:"content"`
	ContentType     string            `bson:"content_type" json:"content_type" db:"content_type"`

	// Attachments
	Attachments     []Attachment      `bson:"attachments" json:"attachments" db:"attachments"`

	// Internal note
	IsInternalNote  bool              `bson:"is_internal_note" json:"is_internal_note" db:"is_internal_note"`

	// Automation
	IsAutomated     bool              `bson:"is_automated" json:"is_automated" db:"is_automated"`
	AutomationRuleID string           `bson:"automation_rule_id" json:"automation_rule_id" db:"automation_rule_id"`

	CreatedAt       time.Time         `bson:"created_at" json:"created_at" db:"created_at"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	FileName        string            `bson:"file_name" json:"file_name" db:"file_name"`
	FileURL         string            `bson:"file_url" json:"file_url" db:"file_url"`
	FileSize        int64             `bson:"file_size" json:"file_size" db:"file_size"`
	MIMEType        string            `bson:"mime_type" json:"mime_type" db:"mime_type"`
}

func (SupportMessage) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("ticket_id", schema.Required(),
			schema.HelpText("Parent ticket ID")),

		// Sender
		schema.StringField("sender_type", schema.Required(),
			schema.HelpText("Sender type: customer, agent, system, automation")),
		schema.StringField("customer_id", schema.Optional(),
			schema.HelpText("Customer ID if sender is customer")),
		schema.StringField("agent_id", schema.Optional(),
			schema.HelpText("Agent ID if sender is agent")),

		// Message
		schema.StringField("content", schema.Required(), schema.MaxLength(10000),
			schema.HelpText("Message content")),
		schema.StringField("content_type", schema.Optional(),
			schema.HelpText("Content type: text, html, markdown")),

		// Internal note
		schema.BoolField("is_internal_note", schema.Default(false)),

		// Automation
		schema.BoolField("is_automated", schema.Default(false)),
		schema.StringField("automation_rule_id", schema.Optional(),
			schema.HelpText("Automation rule ID")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (SupportMessage) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "support_messages",
		VerboseName:       "Support Message",
		VerboseNamePlural: "Support Messages",
		OrderBy:           []string{"created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_message_ticket", "ticket_id"),
			schema.IndexOn("idx_message_sender", "sender_type"),
			schema.IndexOn("idx_message_created", "created_at"),
		},
	}
}

func (SupportMessage) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (SupportMessage) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// ReturnRequest represents a customer return request
type ReturnRequest struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	ReturnNumber    string            `bson:"return_number" json:"return_number" db:"return_number"`
	OrderID         primitive.ObjectID `bson:"order_id" json:"order_id" db:"order_id"`
	CustomerID      primitive.ObjectID `bson:"customer_id" json:"customer_id" db:"customer_id"`

	// Return details
	Reason          string            `bson:"reason" json:"reason" db:"reason"`
	ReasonDetail    string            `bson:"reason_detail" json:"reason_detail" db:"reason_detail"`

	// Items being returned
	Items           []ReturnItem      `bson:"items" json:"items" db:"items"`

	// Resolution
	ResolutionType  string            `bson:"resolution_type" json:"resolution_type" db:"resolution_type"`
	ResolutionDetail string           `bson:"resolution_detail" json:"resolution_detail" db:"resolution_detail"`

	// Refund details
	RefundAmount    float64           `bson:"refund_amount" json:"refund_amount" db:"refund_amount"`
	RefundMethod    string            `bson:"refund_method" json:"refund_method" db:"refund_method"`
	RefundStatus    string            `bson:"refund_status" json:"refund_status" db:"refund_status"`
	RefundProcessedAt *time.Time     `bson:"refund_processed_at,omitempty" json:"refund_processed_at,omitempty" db:"refund_processed_at"`

	// Return shipping
	ReturnShippingMethod string        `bson:"return_shipping_method" json:"return_shipping_method" db:"return_shipping_method"`
	ReturnShippingLabelURL string      `bson:"return_shipping_label_url" json:"return_shipping_label_url" db:"return_shipping_label_url"`
	ReturnTrackingNumber string       `bson:"return_tracking_number" json:"return_tracking_number" db:"return_tracking_number"`
	ReturnCarrier       string        `bson:"return_carrier" json:"return_carrier" db:"return_carrier"`

	// Status workflow
	Status          string            `bson:"status" json:"status" db:"status"`
	StatusHistory   []StatusChange    `bson:"status_history" json:"status_history" db:"status_history"`

	// Inspection
	InspectedAt     *time.Time        `bson:"inspected_at,omitempty" json:"inspected_at,omitempty" db:"inspected_at"`
	InspectionNotes string           `bson:"inspection_notes" json:"inspection_notes" db:"inspection_notes"`
	InspectionResult string          `bson:"inspection_result" json:"inspection_result" db:"inspection_result"`

	// Timeline
	ApprovedAt      *time.Time        `bson:"approved_at,omitempty" json:"approved_at,omitempty" db:"approved_at"`
	ReturnedAt      *time.Time        `bson:"returned_at,omitempty" json:"returned_at,omitempty" db:"returned_at"`
	CompletedAt     *time.Time        `bson:"completed_at,omitempty" json:"completed_at,omitempty" db:"completed_at"`

	CreatedAt       time.Time         `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

// ReturnItem represents an item being returned
type ReturnItem struct {
	OrderItemID     primitive.ObjectID `bson:"order_item_id" json:"order_item_id" db:"order_item_id"`
	ProductID       primitive.ObjectID `bson:"product_id" json:"product_id" db:"product_id"`
	VariantID       *primitive.ObjectID `bson:"variant_id,omitempty" json:"variant_id,omitempty" db:"variant_id"`
	Quantity        int               `bson:"quantity" json:"quantity" db:"quantity"`
	UnitPrice       float64           `bson:"unit_price" json:"unit_price" db:"unit_price"`
	Reason          string            `bson:"reason" json:"reason" db:"reason"`
	Condition       string            `bson:"condition" json:"condition" db:"condition"`
}

// StatusChange represents a status change in the return workflow
type StatusChange struct {
	Status          string            `bson:"status" json:"status" db:"status"`
	ChangedBy       primitive.ObjectID `bson:"changed_by" json:"changed_by" db:"changed_by"`
	ChangedByType   string            `bson:"changed_by_type" json:"changed_by_type" db:"changed_by_type"`
	Reason          string            `bson:"reason" json:"reason" db:"reason"`
	ChangedAt       time.Time         `bson:"changed_at" json:"changed_at" db:"changed_at"`
}

func (ReturnRequest) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("return_number", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique return number")),
		schema.StringField("order_id", schema.Required(),
			schema.HelpText("Original order ID")),
		schema.StringField("customer_id", schema.Required(),
			schema.HelpText("Customer ID")),

		// Return details
		schema.StringField("reason", schema.Required(),
			schema.HelpText("Return reason: defective, wrong_item, not_as_described, changed_mind, size_issues, quality_issues, other")),
		schema.StringField("reason_detail", schema.Optional(), schema.MaxLength(1000),
			schema.HelpText("Detailed reason")),

		// Resolution
		schema.StringField("resolution_type", schema.Required(),
			schema.HelpText("Resolution: refund, exchange, store_credit, repair, replacement")),
		schema.StringField("resolution_detail", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Resolution details")),

		// Refund details
		schema.Float64Field("refund_amount", schema.Default(0.0),
			schema.HelpText("Refund amount")),
		schema.StringField("refund_method", schema.Optional(),
			schema.HelpText("Refund method: original_payment, store_credit, bank_transfer")),
		schema.StringField("refund_status", schema.Optional(),
			schema.HelpText("Refund status: pending, processing, completed, failed")),
		schema.TimeField("refund_processed_at", schema.Optional()),

		// Return shipping
		schema.StringField("return_shipping_method", schema.Optional(),
			schema.HelpText("Return shipping method")),
		schema.StringField("return_shipping_label_url", schema.Optional(),
			schema.HelpText("Return shipping label URL")),
		schema.StringField("return_tracking_number", schema.Optional(),
			schema.HelpText("Return tracking number")),
		schema.StringField("return_carrier", schema.Optional(),
			schema.HelpText("Return carrier")),

		// Status
		schema.StringField("status", schema.Required(),
			schema.HelpText("Status: pending_approval, approved, rejected, pending_return, in_transit, received_inspected, pending_refund, refunded, exchanged, completed, cancelled")),

		// Inspection
		schema.TimeField("inspected_at", schema.Optional()),
		schema.StringField("inspection_notes", schema.Optional(), schema.MaxLength(2000),
			schema.HelpText("Inspection notes")),
		schema.StringField("inspection_result", schema.Optional(),
			schema.HelpText("Inspection result: pending, accepted, rejected, damaged")),

		// Timeline
		schema.TimeField("approved_at", schema.Optional()),
		schema.TimeField("returned_at", schema.Optional()),
		schema.TimeField("completed_at", schema.Optional()),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ReturnRequest) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "return_requests",
		VerboseName:       "Return Request",
		VerboseNamePlural: "Return Requests",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_return_number", "return_number"),
			schema.IndexOn("idx_return_order", "order_id"),
			schema.IndexOn("idx_return_customer", "customer_id"),
			schema.IndexOn("idx_return_status", "status"),
			schema.IndexOn("idx_return_created", "created_at"),
		},
	}
}

func (ReturnRequest) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ReturnRequest) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// LiveChatSession represents a live chat session
type LiveChatSession struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	SessionID       string            `bson:"session_id" json:"session_id" db:"session_id"`

	// Participants
	CustomerID      *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty" db:"customer_id"`
	GuestName       string            `bson:"guest_name" json:"guest_name" db:"guest_name"`
	GuestEmail      string            `bson:"guest_email" json:"guest_email" db:"guest_email"`

	// Agent
	AgentID         *primitive.ObjectID `bson:"agent_id,omitempty" json:"agent_id,omitempty" db:"agent_id"`
	AgentName       string            `bson:"agent_name" json:"agent_name" db:"agent_name"`

	// Chat details
	Status          string            `bson:"status" json:"status" db:"status"`
	Type            string            `bson:"type" json:"type" db:"type"`
	Subject         string            `bson:"subject" json:"subject" db:"subject"`

	// Messages
	Messages        []ChatMessage     `bson:"messages" json:"messages" db:"messages"`

	// End details
	EndReason       string            `bson:"end_reason" json:"end_reason" db:"end_reason"`
	EndNotes        string            `bson:"end_notes" json:"end_notes" db:"end_notes"`

	// Satisfaction
	CustomerSatisfaction *int         `bson:"customer_satisfaction,omitempty" json:"customer_satisfaction,omitempty" db:"customer_satisfaction"`

	// Duration
	WaitTimeSeconds int               `bson:"wait_time_seconds" json:"wait_time_seconds" db:"wait_time_seconds"`
	ChatDurationSeconds int           `bson:"chat_duration_seconds" json:"chat_duration_seconds" db:"chat_duration_seconds"`

	// Linked entities
	RelatedTicketID *primitive.ObjectID `bson:"related_ticket_id,omitempty" json:"related_ticket_id,omitempty" db:"related_ticket_id"`
	RelatedOrderID  *primitive.ObjectID `bson:"related_order_id,omitempty" json:"related_order_id,omitempty" db:"related_order_id"`

	StartedAt       time.Time         `bson:"started_at" json:"started_at" db:"started_at"`
	EndedAt         *time.Time        `bson:"ended_at,omitempty" json:"ended_at,omitempty" db:"ended_at"`
}

// ChatMessage represents a message in a chat session
type ChatMessage struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	SenderType      string            `bson:"sender_type" json:"sender_type" db:"sender_type"`
	SenderID        primitive.ObjectID `bson:"sender_id" json:"sender_id" db:"sender_id"`
	SenderName      string            `bson:"sender_name" json:"sender_name" db:"sender_name"`
	Content         string            `bson:"content" json:"content" db:"content"`
	ContentType     string            `bson:"content_type" json:"content_type" db:"content_type"`

	// Quick replies
	QuickReplies    []string          `bson:"quick_replies" json:"quick_replies" db:"quick_replies"`

	// Read status
	IsRead          bool              `bson:"is_read" json:"is_read" db:"is_read"`

	Timestamp       time.Time         `bson:"timestamp" json:"timestamp" db:"timestamp"`
}

func (LiveChatSession) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("session_id", schema.Required(), schema.MaxLength(50),
			schema.HelpText("Unique session ID")),

		// Participants
		schema.StringField("customer_id", schema.Optional(),
			schema.HelpText("Customer ID")),
		schema.StringField("guest_name", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Guest name")),
		schema.StringField("guest_email", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Guest email")),

		// Agent
		schema.StringField("agent_id", schema.Optional(),
			schema.HelpText("Agent ID")),
		schema.StringField("agent_name", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Agent name")),

		// Chat details
		schema.StringField("status", schema.Required(),
			schema.HelpText("Status: waiting, active, transferred, ended, abandoned")),
		schema.StringField("type", schema.Optional(),
			schema.HelpText("Chat type: sales, support, billing, technical")),
		schema.StringField("subject", schema.Optional(), schema.MaxLength(200),
			schema.HelpText("Chat subject")),

		// End details
		schema.StringField("end_reason", schema.Optional(),
			schema.HelpText("End reason: customer_left, agent_resolved, transferred, escalated, timeout, other")),
		schema.StringField("end_notes", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("End notes")),

		// Satisfaction
		schema.IntField("customer_satisfaction", schema.Optional(), schema.Min(1), schema.Max(5),
			schema.HelpText("Customer satisfaction rating 1-5")),

		// Duration
		schema.IntField("wait_time_seconds", schema.Default(0),
			schema.HelpText("Wait time in seconds")),
		schema.IntField("chat_duration_seconds", schema.Default(0),
			schema.HelpText("Chat duration in seconds")),

		// Linked entities
		schema.StringField("related_ticket_id", schema.Optional(),
			schema.HelpText("Related ticket ID")),
		schema.StringField("related_order_id", schema.Optional(),
			schema.HelpText("Related order ID")),

		// Timestamps
		schema.TimeField("started_at", schema.AutoNowAdd()),
		schema.TimeField("ended_at", schema.Optional()),
	}
}

func (LiveChatSession) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "live_chat_sessions",
		VerboseName:       "Live Chat Session",
		VerboseNamePlural: "Live Chat Sessions",
		OrderBy:           []string{"-started_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_chat_session", "session_id"),
			schema.IndexOn("idx_chat_customer", "customer_id"),
			schema.IndexOn("idx_chat_agent", "agent_id"),
			schema.IndexOn("idx_chat_status", "status"),
			schema.IndexOn("idx_chat_started", "started_at"),
		},
	}
}

func (LiveChatSession) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (LiveChatSession) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// FAQ represents a frequently asked question
type FAQ struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	Question        string            `bson:"question" json:"question" db:"question"`
	Answer          string            `bson:"answer" json:"answer" db:"answer"`

	// Categorization
	Category        string            `bson:"category" json:"category" db:"category"`

	// Keywords for search
	Keywords        []string          `bson:"keywords" json:"keywords" db:"keywords"`

	// Display
	DisplayOrder    int               `bson:"display_order" json:"display_order" db:"display_order"`
	IsVisible       bool              `bson:"is_visible" json:"is_visible" db:"is_visible"`

	// Featured
	IsFeatured      bool              `bson:"is_featured" json:"is_featured" db:"is_featured"`

	// Stats
	ViewCount       int               `bson:"view_count" json:"view_count" db:"view_count"`
	HelpfulCount    int               `bson:"helpful_count" json:"helpful_count" db:"helpful_count"`
	NotHelpfulCount int               `bson:"not_helpful_count" json:"not_helpful_count" db:"not_helpful_count"`

	CreatedAt       time.Time         `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (FAQ) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("question", schema.Required(), schema.MinLength(5), schema.MaxLength(500),
			schema.HelpText("FAQ question")),
		schema.StringField("answer", schema.Required(), schema.MinLength(10), schema.MaxLength(10000),
			schema.HelpText("FAQ answer")),

		// Categorization
		schema.StringField("category", schema.Required(),
			schema.HelpText("Category: orders, shipping, payments, returns, products, account, technical, general")),

		// Keywords for search
		schema.StringArrayField("keywords", schema.Optional(),
			schema.HelpText("Search keywords")),

		// Display
		schema.IntField("display_order", schema.Default(0),
			schema.HelpText("Display order")),
		schema.BoolField("is_visible", schema.Default(true)),

		// Featured
		schema.BoolField("is_featured", schema.Default(false)),

		// Stats
		schema.IntField("view_count", schema.Default(0),
			schema.HelpText("View count")),
		schema.IntField("helpful_count", schema.Default(0),
			schema.HelpText("Helpful votes count")),
		schema.IntField("not_helpful_count", schema.Default(0),
			schema.HelpText("Not helpful votes count")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (FAQ) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "faqs",
		VerboseName:       "FAQ",
		VerboseNamePlural: "FAQs",
		OrderBy:           []string{"display_order", "category"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_faq_category", "category"),
			schema.IndexOn("idx_faq_visible", "is_visible"),
			schema.IndexOn("idx_faq_featured", "is_featured"),
			schema.IndexOn("idx_faq_display", "display_order"),
		},
	}
}

func (FAQ) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (FAQ) Hooks() *schema.ModelHooks {
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

// SupportTicket represents a customer support ticket
type SupportTicket struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	TicketNumber    string            `bson:"ticket_number" json:"ticket_number" db:"ticket_number"`
	CustomerID      primitive.ObjectID `bson:"customer_id" json:"customer_id" db:"customer_id"`

	// Ticket details
	Type            string            `bson:"type" json:"type" db:"type"`
	Subject         string            `bson:"subject" json:"subject" db:"subject"`
	Description     string            `bson:"description" json:"description" db:"description"`

	// Priority and status
	Priority        string            `bson:"priority" json:"priority" db:"priority"`
	Status         string            `bson:"status" json:"status" db:"status"`

	// Assignment
	AssignedTo      *primitive.ObjectID `bson:"assigned_to,omitempty" json:"assigned_to,omitempty" db:"assigned_to"`
	AssignedGroup   string            `bson:"assigned_group" json:"assigned_group" db:"assigned_group"`

	// Related entity
	RelatedType     string            `bson:"related_type" json:"related_type" db:"related_type"`
	RelatedID      primitive.ObjectID `bson:"related_id,omitempty" json:"related_id,omitempty" db:"related_id"`

	// Tags
	Tags            []string          `bson:"tags" json:"tags" db:"tags"`

	// Resolution
	Resolution      string            `bson:"resolution" json:"resolution" db:"resolution"`
	ResolvedAt     *time.Time        `bson:"resolved_at,omitempty" json:"resolved_at,omitempty" db:"resolved_at"`

	// Feedback
	CustomerSatisfaction *int         `bson:"customer_satisfaction,omitempty" json:"customer_satisfaction,omitempty" db:"customer_satisfaction"`
	CustomerFeedback string          `bson:"customer_feedback" json:"customer_feedback" db:"customer_feedback"`

	// Escalation
	IsEscalated    bool              `bson:"is_escalated" json:"is_escalated" db:"is_escalated"`
	EscalatedAt    *time.Time        `bson:"escalated_at,omitempty" json:"escalated_at,omitempty" db:"escalated_at"`
	EscalationReason string          `bson:"escalation_reason" json:"escalation_reason" db:"escalation_reason"`

	// Timestamps
	FirstResponseAt *time.Time       `bson:"first_response_at,omitempty" json:"first_response_at,omitempty" db:"first_response_at"`
	CreatedAt       time.Time         `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (SupportTicket) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("ticket_number", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique ticket number")),
		schema.StringField("customer_id", schema.Required(),
			schema.HelpText("Customer ID")),

		// Ticket details
		schema.StringField("type", schema.Required(),
			schema.HelpText("Ticket type: order_inquiry, product_inquiry, shipping_issue, payment_issue, return_refund, technical, account, billing, general, feedback, praise, suggestion, other")),
		schema.StringField("subject", schema.Required(), schema.MinLength(5), schema.MaxLength(200),
			schema.HelpText("Ticket subject")),
		schema.StringField("description", schema.Required(), schema.MinLength(10), schema.MaxLength(5000),
			schema.HelpText("Ticket description")),

		// Priority and status
		schema.StringField("priority", schema.Required(),
			schema.HelpText("Priority: low, medium, high, urgent")),
		schema.StringField("status", schema.Required(),
			schema.HelpText("Status: open, pending_customer_response, pending_merchant_response, resolved, closed, spam")),

		// Assignment
		schema.StringField("assigned_to", schema.Optional(),
			schema.HelpText("Assigned agent ID")),
		schema.StringField("assigned_group", schema.Optional(),
			schema.HelpText("Assigned group: general, billing, technical, shipping, management")),

		// Related entity
		schema.StringField("related_type", schema.Optional(),
			schema.HelpText("Related entity type: order, product, payment, shipment, return")),
		schema.StringField("related_id", schema.Optional(),
			schema.HelpText("Related entity ID")),

		// Tags
		schema.StringArrayField("tags", schema.Optional(),
			schema.HelpText("Ticket tags")),

		// Resolution
		schema.StringField("resolution", schema.Optional(), schema.MaxLength(2000),
			schema.HelpText("Resolution notes")),
		schema.TimeField("resolved_at", schema.Optional()),

		// Feedback
		schema.IntField("customer_satisfaction", schema.Optional(), schema.Min(1), schema.Max(5),
			schema.HelpText("Customer satisfaction rating 1-5")),
		schema.StringField("customer_feedback", schema.Optional(), schema.MaxLength(1000),
			schema.HelpText("Customer feedback")),

		// Escalation
		schema.BoolField("is_escalated", schema.Default(false)),
		schema.TimeField("escalated_at", schema.Optional()),
		schema.StringField("escalation_reason", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Escalation reason")),

		// Timestamps
		schema.TimeField("first_response_at", schema.Optional()),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (SupportTicket) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "support_tickets",
		VerboseName:       "Support Ticket",
		VerboseNamePlural: "Support Tickets",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_ticket_number", "ticket_number"),
			schema.IndexOn("idx_ticket_customer", "customer_id"),
			schema.IndexOn("idx_ticket_status", "status"),
			schema.IndexOn("idx_ticket_priority", "priority"),
			schema.IndexOn("idx_ticket_type", "type"),
			schema.IndexOn("idx_ticket_assigned", "assigned_to"),
			schema.IndexOn("idx_ticket_created", "created_at"),
		},
	}
}

func (SupportTicket) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (SupportTicket) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Auto-generate ticket number if not set
			// Validate ticket constraints
			return nil
		},
	}
}

// SupportMessage represents a message in a support ticket
type SupportMessage struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	TicketID        primitive.ObjectID `bson:"ticket_id" json:"ticket_id" db:"ticket_id"`

	// Sender
	SenderType      string            `bson:"sender_type" json:"sender_type" db:"sender_type"`
	CustomerID      *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty" db:"customer_id"`
	AgentID         *primitive.ObjectID `bson:"agent_id,omitempty" json:"agent_id,omitempty" db:"agent_id"`

	// Message
	Content         string            `bson:"content" json:"content" db:"content"`
	ContentType     string            `bson:"content_type" json:"content_type" db:"content_type"`

	// Attachments
	Attachments     []Attachment      `bson:"attachments" json:"attachments" db:"attachments"`

	// Internal note
	IsInternalNote  bool              `bson:"is_internal_note" json:"is_internal_note" db:"is_internal_note"`

	// Automation
	IsAutomated     bool              `bson:"is_automated" json:"is_automated" db:"is_automated"`
	AutomationRuleID string           `bson:"automation_rule_id" json:"automation_rule_id" db:"automation_rule_id"`

	CreatedAt       time.Time         `bson:"created_at" json:"created_at" db:"created_at"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	FileName        string            `bson:"file_name" json:"file_name" db:"file_name"`
	FileURL         string            `bson:"file_url" json:"file_url" db:"file_url"`
	FileSize        int64             `bson:"file_size" json:"file_size" db:"file_size"`
	MIMEType        string            `bson:"mime_type" json:"mime_type" db:"mime_type"`
}

func (SupportMessage) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("ticket_id", schema.Required(),
			schema.HelpText("Parent ticket ID")),

		// Sender
		schema.StringField("sender_type", schema.Required(),
			schema.HelpText("Sender type: customer, agent, system, automation")),
		schema.StringField("customer_id", schema.Optional(),
			schema.HelpText("Customer ID if sender is customer")),
		schema.StringField("agent_id", schema.Optional(),
			schema.HelpText("Agent ID if sender is agent")),

		// Message
		schema.StringField("content", schema.Required(), schema.MaxLength(10000),
			schema.HelpText("Message content")),
		schema.StringField("content_type", schema.Optional(),
			schema.HelpText("Content type: text, html, markdown")),

		// Internal note
		schema.BoolField("is_internal_note", schema.Default(false)),

		// Automation
		schema.BoolField("is_automated", schema.Default(false)),
		schema.StringField("automation_rule_id", schema.Optional(),
			schema.HelpText("Automation rule ID")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (SupportMessage) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "support_messages",
		VerboseName:       "Support Message",
		VerboseNamePlural: "Support Messages",
		OrderBy:           []string{"created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_message_ticket", "ticket_id"),
			schema.IndexOn("idx_message_sender", "sender_type"),
			schema.IndexOn("idx_message_created", "created_at"),
		},
	}
}

func (SupportMessage) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (SupportMessage) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// ReturnRequest represents a customer return request
type ReturnRequest struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	ReturnNumber    string            `bson:"return_number" json:"return_number" db:"return_number"`
	OrderID         primitive.ObjectID `bson:"order_id" json:"order_id" db:"order_id"`
	CustomerID      primitive.ObjectID `bson:"customer_id" json:"customer_id" db:"customer_id"`

	// Return details
	Reason          string            `bson:"reason" json:"reason" db:"reason"`
	ReasonDetail    string            `bson:"reason_detail" json:"reason_detail" db:"reason_detail"`

	// Items being returned
	Items           []ReturnItem      `bson:"items" json:"items" db:"items"`

	// Resolution
	ResolutionType  string            `bson:"resolution_type" json:"resolution_type" db:"resolution_type"`
	ResolutionDetail string           `bson:"resolution_detail" json:"resolution_detail" db:"resolution_detail"`

	// Refund details
	RefundAmount    float64           `bson:"refund_amount" json:"refund_amount" db:"refund_amount"`
	RefundMethod    string            `bson:"refund_method" json:"refund_method" db:"refund_method"`
	RefundStatus    string            `bson:"refund_status" json:"refund_status" db:"refund_status"`
	RefundProcessedAt *time.Time     `bson:"refund_processed_at,omitempty" json:"refund_processed_at,omitempty" db:"refund_processed_at"`

	// Return shipping
	ReturnShippingMethod string        `bson:"return_shipping_method" json:"return_shipping_method" db:"return_shipping_method"`
	ReturnShippingLabelURL string      `bson:"return_shipping_label_url" json:"return_shipping_label_url" db:"return_shipping_label_url"`
	ReturnTrackingNumber string       `bson:"return_tracking_number" json:"return_tracking_number" db:"return_tracking_number"`
	ReturnCarrier       string        `bson:"return_carrier" json:"return_carrier" db:"return_carrier"`

	// Status workflow
	Status          string            `bson:"status" json:"status" db:"status"`
	StatusHistory   []StatusChange    `bson:"status_history" json:"status_history" db:"status_history"`

	// Inspection
	InspectedAt     *time.Time        `bson:"inspected_at,omitempty" json:"inspected_at,omitempty" db:"inspected_at"`
	InspectionNotes string           `bson:"inspection_notes" json:"inspection_notes" db:"inspection_notes"`
	InspectionResult string          `bson:"inspection_result" json:"inspection_result" db:"inspection_result"`

	// Timeline
	ApprovedAt      *time.Time        `bson:"approved_at,omitempty" json:"approved_at,omitempty" db:"approved_at"`
	ReturnedAt      *time.Time        `bson:"returned_at,omitempty" json:"returned_at,omitempty" db:"returned_at"`
	CompletedAt     *time.Time        `bson:"completed_at,omitempty" json:"completed_at,omitempty" db:"completed_at"`

	CreatedAt       time.Time         `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

// ReturnItem represents an item being returned
type ReturnItem struct {
	OrderItemID     primitive.ObjectID `bson:"order_item_id" json:"order_item_id" db:"order_item_id"`
	ProductID       primitive.ObjectID `bson:"product_id" json:"product_id" db:"product_id"`
	VariantID       *primitive.ObjectID `bson:"variant_id,omitempty" json:"variant_id,omitempty" db:"variant_id"`
	Quantity        int               `bson:"quantity" json:"quantity" db:"quantity"`
	UnitPrice       float64           `bson:"unit_price" json:"unit_price" db:"unit_price"`
	Reason          string            `bson:"reason" json:"reason" db:"reason"`
	Condition       string            `bson:"condition" json:"condition" db:"condition"`
}

// StatusChange represents a status change in the return workflow
type StatusChange struct {
	Status          string            `bson:"status" json:"status" db:"status"`
	ChangedBy       primitive.ObjectID `bson:"changed_by" json:"changed_by" db:"changed_by"`
	ChangedByType   string            `bson:"changed_by_type" json:"changed_by_type" db:"changed_by_type"`
	Reason          string            `bson:"reason" json:"reason" db:"reason"`
	ChangedAt       time.Time         `bson:"changed_at" json:"changed_at" db:"changed_at"`
}

func (ReturnRequest) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("return_number", schema.Required(), schema.MaxLength(20),
			schema.HelpText("Unique return number")),
		schema.StringField("order_id", schema.Required(),
			schema.HelpText("Original order ID")),
		schema.StringField("customer_id", schema.Required(),
			schema.HelpText("Customer ID")),

		// Return details
		schema.StringField("reason", schema.Required(),
			schema.HelpText("Return reason: defective, wrong_item, not_as_described, changed_mind, size_issues, quality_issues, other")),
		schema.StringField("reason_detail", schema.Optional(), schema.MaxLength(1000),
			schema.HelpText("Detailed reason")),

		// Resolution
		schema.StringField("resolution_type", schema.Required(),
			schema.HelpText("Resolution: refund, exchange, store_credit, repair, replacement")),
		schema.StringField("resolution_detail", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("Resolution details")),

		// Refund details
		schema.Float64Field("refund_amount", schema.Default(0.0),
			schema.HelpText("Refund amount")),
		schema.StringField("refund_method", schema.Optional(),
			schema.HelpText("Refund method: original_payment, store_credit, bank_transfer")),
		schema.StringField("refund_status", schema.Optional(),
			schema.HelpText("Refund status: pending, processing, completed, failed")),
		schema.TimeField("refund_processed_at", schema.Optional()),

		// Return shipping
		schema.StringField("return_shipping_method", schema.Optional(),
			schema.HelpText("Return shipping method")),
		schema.StringField("return_shipping_label_url", schema.Optional(),
			schema.HelpText("Return shipping label URL")),
		schema.StringField("return_tracking_number", schema.Optional(),
			schema.HelpText("Return tracking number")),
		schema.StringField("return_carrier", schema.Optional(),
			schema.HelpText("Return carrier")),

		// Status
		schema.StringField("status", schema.Required(),
			schema.HelpText("Status: pending_approval, approved, rejected, pending_return, in_transit, received_inspected, pending_refund, refunded, exchanged, completed, cancelled")),

		// Inspection
		schema.TimeField("inspected_at", schema.Optional()),
		schema.StringField("inspection_notes", schema.Optional(), schema.MaxLength(2000),
			schema.HelpText("Inspection notes")),
		schema.StringField("inspection_result", schema.Optional(),
			schema.HelpText("Inspection result: pending, accepted, rejected, damaged")),

		// Timeline
		schema.TimeField("approved_at", schema.Optional()),
		schema.TimeField("returned_at", schema.Optional()),
		schema.TimeField("completed_at", schema.Optional()),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ReturnRequest) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "return_requests",
		VerboseName:       "Return Request",
		VerboseNamePlural: "Return Requests",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_return_number", "return_number"),
			schema.IndexOn("idx_return_order", "order_id"),
			schema.IndexOn("idx_return_customer", "customer_id"),
			schema.IndexOn("idx_return_status", "status"),
			schema.IndexOn("idx_return_created", "created_at"),
		},
	}
}

func (ReturnRequest) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ReturnRequest) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// LiveChatSession represents a live chat session
type LiveChatSession struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	SessionID       string            `bson:"session_id" json:"session_id" db:"session_id"`

	// Participants
	CustomerID      *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty" db:"customer_id"`
	GuestName       string            `bson:"guest_name" json:"guest_name" db:"guest_name"`
	GuestEmail      string            `bson:"guest_email" json:"guest_email" db:"guest_email"`

	// Agent
	AgentID         *primitive.ObjectID `bson:"agent_id,omitempty" json:"agent_id,omitempty" db:"agent_id"`
	AgentName       string            `bson:"agent_name" json:"agent_name" db:"agent_name"`

	// Chat details
	Status          string            `bson:"status" json:"status" db:"status"`
	Type            string            `bson:"type" json:"type" db:"type"`
	Subject         string            `bson:"subject" json:"subject" db:"subject"`

	// Messages
	Messages        []ChatMessage     `bson:"messages" json:"messages" db:"messages"`

	// End details
	EndReason       string            `bson:"end_reason" json:"end_reason" db:"end_reason"`
	EndNotes        string            `bson:"end_notes" json:"end_notes" db:"end_notes"`

	// Satisfaction
	CustomerSatisfaction *int         `bson:"customer_satisfaction,omitempty" json:"customer_satisfaction,omitempty" db:"customer_satisfaction"`

	// Duration
	WaitTimeSeconds int               `bson:"wait_time_seconds" json:"wait_time_seconds" db:"wait_time_seconds"`
	ChatDurationSeconds int           `bson:"chat_duration_seconds" json:"chat_duration_seconds" db:"chat_duration_seconds"`

	// Linked entities
	RelatedTicketID *primitive.ObjectID `bson:"related_ticket_id,omitempty" json:"related_ticket_id,omitempty" db:"related_ticket_id"`
	RelatedOrderID  *primitive.ObjectID `bson:"related_order_id,omitempty" json:"related_order_id,omitempty" db:"related_order_id"`

	StartedAt       time.Time         `bson:"started_at" json:"started_at" db:"started_at"`
	EndedAt         *time.Time        `bson:"ended_at,omitempty" json:"ended_at,omitempty" db:"ended_at"`
}

// ChatMessage represents a message in a chat session
type ChatMessage struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	SenderType      string            `bson:"sender_type" json:"sender_type" db:"sender_type"`
	SenderID        primitive.ObjectID `bson:"sender_id" json:"sender_id" db:"sender_id"`
	SenderName      string            `bson:"sender_name" json:"sender_name" db:"sender_name"`
	Content         string            `bson:"content" json:"content" db:"content"`
	ContentType     string            `bson:"content_type" json:"content_type" db:"content_type"`

	// Quick replies
	QuickReplies    []string          `bson:"quick_replies" json:"quick_replies" db:"quick_replies"`

	// Read status
	IsRead          bool              `bson:"is_read" json:"is_read" db:"is_read"`

	Timestamp       time.Time         `bson:"timestamp" json:"timestamp" db:"timestamp"`
}

func (LiveChatSession) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("session_id", schema.Required(), schema.MaxLength(50),
			schema.HelpText("Unique session ID")),

		// Participants
		schema.StringField("customer_id", schema.Optional(),
			schema.HelpText("Customer ID")),
		schema.StringField("guest_name", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Guest name")),
		schema.StringField("guest_email", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Guest email")),

		// Agent
		schema.StringField("agent_id", schema.Optional(),
			schema.HelpText("Agent ID")),
		schema.StringField("agent_name", schema.Optional(), schema.MaxLength(100),
			schema.HelpText("Agent name")),

		// Chat details
		schema.StringField("status", schema.Required(),
			schema.HelpText("Status: waiting, active, transferred, ended, abandoned")),
		schema.StringField("type", schema.Optional(),
			schema.HelpText("Chat type: sales, support, billing, technical")),
		schema.StringField("subject", schema.Optional(), schema.MaxLength(200),
			schema.HelpText("Chat subject")),

		// End details
		schema.StringField("end_reason", schema.Optional(),
			schema.HelpText("End reason: customer_left, agent_resolved, transferred, escalated, timeout, other")),
		schema.StringField("end_notes", schema.Optional(), schema.MaxLength(500),
			schema.HelpText("End notes")),

		// Satisfaction
		schema.IntField("customer_satisfaction", schema.Optional(), schema.Min(1), schema.Max(5),
			schema.HelpText("Customer satisfaction rating 1-5")),

		// Duration
		schema.IntField("wait_time_seconds", schema.Default(0),
			schema.HelpText("Wait time in seconds")),
		schema.IntField("chat_duration_seconds", schema.Default(0),
			schema.HelpText("Chat duration in seconds")),

		// Linked entities
		schema.StringField("related_ticket_id", schema.Optional(),
			schema.HelpText("Related ticket ID")),
		schema.StringField("related_order_id", schema.Optional(),
			schema.HelpText("Related order ID")),

		// Timestamps
		schema.TimeField("started_at", schema.AutoNowAdd()),
		schema.TimeField("ended_at", schema.Optional()),
	}
}

func (LiveChatSession) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "live_chat_sessions",
		VerboseName:       "Live Chat Session",
		VerboseNamePlural: "Live Chat Sessions",
		OrderBy:           []string{"-started_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_chat_session", "session_id"),
			schema.IndexOn("idx_chat_customer", "customer_id"),
			schema.IndexOn("idx_chat_agent", "agent_id"),
			schema.IndexOn("idx_chat_status", "status"),
			schema.IndexOn("idx_chat_started", "started_at"),
		},
	}
}

func (LiveChatSession) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (LiveChatSession) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

// FAQ represents a frequently asked question
type FAQ struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id" db:"_id"`
	Question        string            `bson:"question" json:"question" db:"question"`
	Answer          string            `bson:"answer" json:"answer" db:"answer"`

	// Categorization
	Category        string            `bson:"category" json:"category" db:"category"`

	// Keywords for search
	Keywords        []string          `bson:"keywords" json:"keywords" db:"keywords"`

	// Display
	DisplayOrder    int               `bson:"display_order" json:"display_order" db:"display_order"`
	IsVisible       bool              `bson:"is_visible" json:"is_visible" db:"is_visible"`

	// Featured
	IsFeatured      bool              `bson:"is_featured" json:"is_featured" db:"is_featured"`

	// Stats
	ViewCount       int               `bson:"view_count" json:"view_count" db:"view_count"`
	HelpfulCount    int               `bson:"helpful_count" json:"helpful_count" db:"helpful_count"`
	NotHelpfulCount int               `bson:"not_helpful_count" json:"not_helpful_count" db:"not_helpful_count"`

	CreatedAt       time.Time         `bson:"created_at" json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `bson:"updated_at" json:"updated_at" db:"updated_at"`
}

func (FAQ) Fields() []schema.Field {
	return []schema.Field{
		schema.ObjectIDField("_id", schema.Primary(), schema.Auto()),
		schema.StringField("question", schema.Required(), schema.MinLength(5), schema.MaxLength(500),
			schema.HelpText("FAQ question")),
		schema.StringField("answer", schema.Required(), schema.MinLength(10), schema.MaxLength(10000),
			schema.HelpText("FAQ answer")),

		// Categorization
		schema.StringField("category", schema.Required(),
			schema.HelpText("Category: orders, shipping, payments, returns, products, account, technical, general")),

		// Keywords for search
		schema.StringArrayField("keywords", schema.Optional(),
			schema.HelpText("Search keywords")),

		// Display
		schema.IntField("display_order", schema.Default(0),
			schema.HelpText("Display order")),
		schema.BoolField("is_visible", schema.Default(true)),

		// Featured
		schema.BoolField("is_featured", schema.Default(false)),

		// Stats
		schema.IntField("view_count", schema.Default(0),
			schema.HelpText("View count")),
		schema.IntField("helpful_count", schema.Default(0),
			schema.HelpText("Helpful votes count")),
		schema.IntField("not_helpful_count", schema.Default(0),
			schema.HelpText("Not helpful votes count")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (FAQ) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "faqs",
		VerboseName:       "FAQ",
		VerboseNamePlural: "FAQs",
		OrderBy:           []string{"display_order", "category"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_faq_category", "category"),
			schema.IndexOn("idx_faq_visible", "is_visible"),
			schema.IndexOn("idx_faq_featured", "is_featured"),
			schema.IndexOn("idx_faq_display", "display_order"),
		},
	}
}

func (FAQ) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (FAQ) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			return nil
		},
	}
}

