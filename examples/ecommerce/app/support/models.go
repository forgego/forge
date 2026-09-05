package support

import (
	"github.com/forgego/forge/schema"
)

// SupportTicket represents a customer support ticket
type SupportTicket struct {
	schema.BaseSchema
	Id           int64  `json:"id" db:"id"`
	TicketNumber string `json:"ticket_number" db:"ticket_number"`
	CustomerID   int64  `json:"customer_id" db:"customer_id"`
	Subject      string `json:"subject" db:"subject"`
	Description  string `json:"description" db:"description"`
	Status       string `json:"status" db:"status"`
	Priority     string `json:"priority" db:"priority"`
	Category     string `json:"category" db:"category"`
	AssignedTo   int64  `json:"assigned_to" db:"assigned_to"`
	OrderID      int64  `json:"order_id" db:"order_id"`
	Source       string `json:"source" db:"source"`
	Resolution   string `json:"resolution" db:"resolution"`
	ResolvedAt   string `json:"resolved_at" db:"resolved_at"`
	ClosedAt     string `json:"closed_at" db:"closed_at"`
	FirstRespAt  string `json:"first_response_at" db:"first_response_at"`
	Tags         string `json:"tags" db:"tags"`
	CreatedAt    string `json:"created_at" db:"created_at"`
	UpdatedAt    string `json:"updated_at" db:"updated_at"`
}

func (SupportTicket) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("ticket_number", schema.Required(), schema.MaxLength(50), schema.Unique(),
			schema.VerboseName("Ticket Number"),
			schema.HelpText("Auto-generated unique ticket number")),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),
		schema.StringField("subject", schema.Required(), schema.MaxLength(300),
			schema.HelpText("Ticket subject")),
		schema.TextField("description", schema.Required(),
			schema.HelpText("Detailed description of the issue")),
		schema.StringField("status", schema.MaxLength(20), schema.Default("open"),
			schema.HelpText("Status: open, in_progress, waiting_customer, resolved, closed")),
		schema.StringField("priority", schema.MaxLength(20), schema.Default("normal"),
			schema.HelpText("Priority: low, normal, high, urgent")),
		schema.StringField("category", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Category: order_issue, product_inquiry, technical, billing, etc.")),
		schema.Int64Field("assigned_to", schema.Optional(),
			schema.VerboseName("Assigned To"),
			schema.HelpText("Support agent ID")),
		schema.Int64Field("order_id", schema.Optional(),
			schema.VerboseName("Related Order"),
			schema.HelpText("Related order if applicable")),
		schema.StringField("source", schema.MaxLength(50), schema.Default("web"),
			schema.HelpText("Source: web, email, phone, chat, social")),
		schema.TextField("resolution", schema.Optional(),
			schema.HelpText("Resolution notes")),
		schema.TimeField("resolved_at", schema.Optional(),
			schema.VerboseName("Resolved At")),
		schema.TimeField("closed_at", schema.Optional(),
			schema.VerboseName("Closed At")),
		schema.TimeField("first_response_at", schema.Optional(),
			schema.VerboseName("First Response At"),
			schema.HelpText("Time of first agent response")),
		schema.TextField("tags", schema.Optional(),
			schema.HelpText("Comma-separated tags for categorization")),
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
			schema.IndexOn("idx_ticket_assigned", "assigned_to"),
			schema.IndexOn("idx_ticket_order", "order_id"),
		},
	}
}

func (SupportTicket) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("support_tickets")),
		schema.ForeignKeyField("order_id", "Order",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("support_tickets")),
	}
}

func (SupportTicket) Hooks() *schema.ModelHooks {
	return nil
}

// SupportMessage represents messages in a support ticket
type SupportMessage struct {
	schema.BaseSchema
	Id         int64  `json:"id" db:"id"`
	TicketID   int64  `json:"ticket_id" db:"ticket_id"`
	SenderType string `json:"sender_type" db:"sender_type"`
	SenderID   int64  `json:"sender_id" db:"sender_id"`
	SenderName string `json:"sender_name" db:"sender_name"`
	Message    string `json:"message" db:"message"`
	IsInternal bool   `json:"is_internal" db:"is_internal"`
	CreatedAt  string `json:"created_at" db:"created_at"`
	UpdatedAt  string `json:"updated_at" db:"updated_at"`
}

func (SupportMessage) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("ticket_id", schema.Required(),
			schema.VerboseName("Ticket")),
		schema.StringField("sender_type", schema.MaxLength(20), schema.Required(),
			schema.VerboseName("Sender Type"),
			schema.HelpText("Type: customer, agent, system")),
		schema.Int64Field("sender_id", schema.Optional(),
			schema.VerboseName("Sender ID"),
			schema.HelpText("Customer or agent ID")),
		schema.StringField("sender_name", schema.MaxLength(200), schema.Optional(),
			schema.VerboseName("Sender Name")),
		schema.TextField("message", schema.Required(),
			schema.HelpText("Message content")),
		schema.BoolField("is_internal", schema.Default(false),
			schema.VerboseName("Internal Note"),
			schema.HelpText("Internal note not visible to customer")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (SupportMessage) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "support_messages",
		VerboseName:       "Support Message",
		VerboseNamePlural: "Support Messages",
		OrderBy:           []string{"created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_support_message_ticket", "ticket_id"),
			schema.IndexOn("idx_support_message_sender", "sender_type", "sender_id"),
		},
	}
}

func (SupportMessage) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("ticket_id", "SupportTicket",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("messages")),
	}
}

func (SupportMessage) Hooks() *schema.ModelHooks {
	return nil
}

// ReturnRequest represents product return requests
type ReturnRequest struct {
	schema.BaseSchema
	Id            int64   `json:"id" db:"id"`
	ReturnNumber  string  `json:"return_number" db:"return_number"`
	OrderID       int64   `json:"order_id" db:"order_id"`
	CustomerID    int64   `json:"customer_id" db:"customer_id"`
	Reason        string  `json:"reason" db:"reason"`
	Description   string  `json:"description" db:"description"`
	Status        string  `json:"status" db:"status"`
	ReturnMethod  string  `json:"return_method" db:"return_method"`
	RefundMethod  string  `json:"refund_method" db:"refund_method"`
	RefundAmount  float64 `json:"refund_amount" db:"refund_amount"`
	RestockFee    float64 `json:"restock_fee" db:"restock_fee"`
	ShippingLabel string  `json:"shipping_label" db:"shipping_label"`
	TrackingNum   string  `json:"tracking_number" db:"tracking_number"`
	ApprovedAt    string  `json:"approved_at" db:"approved_at"`
	ApprovedBy    int64   `json:"approved_by" db:"approved_by"`
	ReceivedAt    string  `json:"received_at" db:"received_at"`
	ProcessedAt   string  `json:"processed_at" db:"processed_at"`
	RefundedAt    string  `json:"refunded_at" db:"refunded_at"`
	RejectedAt    string  `json:"rejected_at" db:"rejected_at"`
	RejectionNote string  `json:"rejection_note" db:"rejection_note"`
	CreatedAt     string  `json:"created_at" db:"created_at"`
	UpdatedAt     string  `json:"updated_at" db:"updated_at"`
}

func (ReturnRequest) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("return_number", schema.Required(), schema.MaxLength(50), schema.Unique(),
			schema.VerboseName("Return Number"),
			schema.HelpText("Auto-generated unique return number")),
		schema.Int64Field("order_id", schema.Required(),
			schema.VerboseName("Order")),
		schema.Int64Field("customer_id", schema.Required(),
			schema.VerboseName("Customer")),
		schema.StringField("reason", schema.MaxLength(50), schema.Required(),
			schema.HelpText("Reason: defective, wrong_item, not_as_described, changed_mind, etc.")),
		schema.TextField("description", schema.Optional(),
			schema.HelpText("Detailed description")),
		schema.StringField("status", schema.MaxLength(20), schema.Default("pending"),
			schema.HelpText("Status: pending, approved, rejected, received, processed, refunded")),
		schema.StringField("return_method", schema.MaxLength(50), schema.Optional(),
			schema.VerboseName("Return Method"),
			schema.HelpText("How item will be returned: mail, drop_off, pickup")),
		schema.StringField("refund_method", schema.MaxLength(50), schema.Default("original"),
			schema.VerboseName("Refund Method"),
			schema.HelpText("Method: original, store_credit, exchange")),
		schema.FloatField("refund_amount", schema.Default(0.0),
			schema.VerboseName("Refund Amount")),
		schema.FloatField("restock_fee", schema.Default(0.0),
			schema.VerboseName("Restock Fee")),
		schema.StringField("shipping_label", schema.MaxLength(500), schema.Optional(),
			schema.VerboseName("Shipping Label URL")),
		schema.StringField("tracking_number", schema.MaxLength(100), schema.Optional(),
			schema.VerboseName("Tracking Number")),
		schema.TimeField("approved_at", schema.Optional(),
			schema.VerboseName("Approved At")),
		schema.Int64Field("approved_by", schema.Optional(),
			schema.VerboseName("Approved By"),
			schema.HelpText("Staff member who approved")),
		schema.TimeField("received_at", schema.Optional(),
			schema.VerboseName("Received At")),
		schema.TimeField("processed_at", schema.Optional(),
			schema.VerboseName("Processed At")),
		schema.TimeField("refunded_at", schema.Optional(),
			schema.VerboseName("Refunded At")),
		schema.TimeField("rejected_at", schema.Optional(),
			schema.VerboseName("Rejected At")),
		schema.TextField("rejection_note", schema.Optional(),
			schema.VerboseName("Rejection Note")),
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
		},
	}
}

func (ReturnRequest) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("order_id", "Order",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("return_requests")),
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("return_requests")),
	}
}

func (ReturnRequest) Hooks() *schema.ModelHooks {
	return nil
}

// LiveChatSession represents live chat sessions
type LiveChatSession struct {
	schema.BaseSchema
	Id          int64  `json:"id" db:"id"`
	SessionID   string `json:"session_id" db:"session_id"`
	CustomerID  int64  `json:"customer_id" db:"customer_id"`
	AgentID     int64  `json:"agent_id" db:"agent_id"`
	Status      string `json:"status" db:"status"`
	StartedAt   string `json:"started_at" db:"started_at"`
	EndedAt     string `json:"ended_at" db:"ended_at"`
	Duration    int32  `json:"duration" db:"duration"`
	MsgCount    int32  `json:"message_count" db:"message_count"`
	Rating      int32  `json:"rating" db:"rating"`
	Feedback    string `json:"feedback" db:"feedback"`
	IPAddress   string `json:"ip_address" db:"ip_address"`
	UserAgent   string `json:"user_agent" db:"user_agent"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

func (LiveChatSession) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("session_id", schema.Required(), schema.MaxLength(100), schema.Unique(),
			schema.VerboseName("Session ID"),
			schema.HelpText("Unique session identifier")),
		schema.Int64Field("customer_id", schema.Optional(),
			schema.VerboseName("Customer"),
			schema.HelpText("null for anonymous users")),
		schema.Int64Field("agent_id", schema.Optional(),
			schema.VerboseName("Agent"),
			schema.HelpText("Assigned support agent")),
		schema.StringField("status", schema.MaxLength(20), schema.Default("waiting"),
			schema.HelpText("Status: waiting, active, ended, abandoned")),
		schema.TimeField("started_at", schema.Required(),
			schema.VerboseName("Started At")),
		schema.TimeField("ended_at", schema.Optional(),
			schema.VerboseName("Ended At")),
		schema.Int32Field("duration", schema.Default(0),
			schema.HelpText("Session duration in seconds")),
		schema.Int32Field("message_count", schema.Default(0),
			schema.VerboseName("Message Count")),
		schema.Int32Field("rating", schema.Optional(),
			schema.HelpText("Customer rating (1-5)")),
		schema.TextField("feedback", schema.Optional(),
			schema.HelpText("Customer feedback")),
		schema.StringField("ip_address", schema.MaxLength(45), schema.Optional(),
			schema.VerboseName("IP Address")),
		schema.StringField("user_agent", schema.MaxLength(500), schema.Optional(),
			schema.VerboseName("User Agent")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (LiveChatSession) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "live_chat_sessions",
		VerboseName:       "Live Chat Session",
		VerboseNamePlural: "Live Chat Sessions",
		OrderBy:           []string{"-started_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_chat_session_id", "session_id"),
			schema.IndexOn("idx_chat_customer", "customer_id"),
			schema.IndexOn("idx_chat_agent", "agent_id"),
			schema.IndexOn("idx_chat_status", "status"),
		},
	}
}

func (LiveChatSession) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("chat_sessions")),
	}
}

func (LiveChatSession) Hooks() *schema.ModelHooks {
	return nil
}

// FAQ represents frequently asked questions
type FAQ struct {
	schema.BaseSchema
	Id         int64  `json:"id" db:"id"`
	Question   string `json:"question" db:"question"`
	Answer     string `json:"answer" db:"answer"`
	Category   string `json:"category" db:"category"`
	IsPublic   bool   `json:"is_public" db:"is_public"`
	ViewCount  int32  `json:"view_count" db:"view_count"`
	HelpfulYes int32  `json:"helpful_yes" db:"helpful_yes"`
	HelpfulNo  int32  `json:"helpful_no" db:"helpful_no"`
	SortOrder  int32  `json:"sort_order" db:"sort_order"`
	Tags       string `json:"tags" db:"tags"`
	CreatedAt  string `json:"created_at" db:"created_at"`
	UpdatedAt  string `json:"updated_at" db:"updated_at"`
}

func (FAQ) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.TextField("question", schema.Required(),
			schema.HelpText("FAQ question")),
		schema.TextField("answer", schema.Required(),
			schema.HelpText("FAQ answer (supports HTML)")),
		schema.StringField("category", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("Category: orders, shipping, returns, products, account, etc.")),
		schema.BoolField("is_public", schema.Default(true),
			schema.VerboseName("Public"),
			schema.HelpText("Visible to customers")),
		schema.Int32Field("view_count", schema.Default(0),
			schema.VerboseName("View Count")),
		schema.Int32Field("helpful_yes", schema.Default(0),
			schema.VerboseName("Helpful (Yes)"),
			schema.HelpText("Number of 'helpful' votes")),
		schema.Int32Field("helpful_no", schema.Default(0),
			schema.VerboseName("Helpful (No)"),
			schema.HelpText("Number of 'not helpful' votes")),
		schema.Int32Field("sort_order", schema.Default(0),
			schema.VerboseName("Sort Order")),
		schema.TextField("tags", schema.Optional(),
			schema.HelpText("Comma-separated tags for search")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (FAQ) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "faqs",
		VerboseName:       "FAQ",
		VerboseNamePlural: "FAQs",
		OrderBy:           []string{"sort_order", "-view_count"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_faq_category", "category"),
			schema.IndexOn("idx_faq_public", "is_public"),
		},
	}
}

func (FAQ) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (FAQ) Hooks() *schema.ModelHooks {
	return nil
}

// Attachment represents file attachments for tickets and returns
type Attachment struct {
	schema.BaseSchema
	Id           int64  `json:"id" db:"id"`
	EntityType   string `json:"entity_type" db:"entity_type"`
	EntityID     int64  `json:"entity_id" db:"entity_id"`
	FileName     string `json:"file_name" db:"file_name"`
	FileURL      string `json:"file_url" db:"file_url"`
	FileSize     int64  `json:"file_size" db:"file_size"`
	MimeType     string `json:"mime_type" db:"mime_type"`
	UploadedBy   int64  `json:"uploaded_by" db:"uploaded_by"`
	UploaderType string `json:"uploader_type" db:"uploader_type"`
	CreatedAt    string `json:"created_at" db:"created_at"`
}

func (Attachment) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("entity_type", schema.MaxLength(50), schema.Required(),
			schema.VerboseName("Entity Type"),
			schema.HelpText("Type: support_ticket, support_message, return_request")),
		schema.Int64Field("entity_id", schema.Required(),
			schema.VerboseName("Entity ID")),
		schema.StringField("file_name", schema.MaxLength(255), schema.Required(),
			schema.VerboseName("File Name")),
		schema.StringField("file_url", schema.MaxLength(500), schema.Required(),
			schema.VerboseName("File URL")),
		schema.Int64Field("file_size", schema.Default(0),
			schema.VerboseName("File Size"),
			schema.HelpText("File size in bytes")),
		schema.StringField("mime_type", schema.MaxLength(100), schema.Optional(),
			schema.VerboseName("MIME Type")),
		schema.Int64Field("uploaded_by", schema.Optional(),
			schema.VerboseName("Uploaded By")),
		schema.StringField("uploader_type", schema.MaxLength(20), schema.Default("customer"),
			schema.VerboseName("Uploader Type"),
			schema.HelpText("Type: customer, agent, system")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Attachment) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "attachments",
		VerboseName:       "Attachment",
		VerboseNamePlural: "Attachments",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_attachment_entity", "entity_type", "entity_id"),
			schema.IndexOn("idx_attachment_uploader", "uploaded_by"),
		},
	}
}

func (Attachment) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Attachment) Hooks() *schema.ModelHooks {
	return nil
}

// ReturnItem represents individual items in a return request
type ReturnItem struct {
	schema.BaseSchema
	Id             int64   `json:"id" db:"id"`
	ReturnReqID    int64   `json:"return_request_id" db:"return_request_id"`
	OrderItemID    int64   `json:"order_item_id" db:"order_item_id"`
	Quantity       int32   `json:"quantity" db:"quantity"`
	Reason         string  `json:"reason" db:"reason"`
	Condition      string  `json:"condition" db:"condition"`
	RefundAmount   float64 `json:"refund_amount" db:"refund_amount"`
	IsRestockable  bool    `json:"is_restockable" db:"is_restockable"`
	InspectionNote string  `json:"inspection_note" db:"inspection_note"`
	CreatedAt      string  `json:"created_at" db:"created_at"`
	UpdatedAt      string  `json:"updated_at" db:"updated_at"`
}

func (ReturnItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("return_request_id", schema.Required(),
			schema.VerboseName("Return Request")),
		schema.Int64Field("order_item_id", schema.Required(),
			schema.VerboseName("Order Item")),
		schema.Int32Field("quantity", schema.Required(),
			schema.HelpText("Quantity being returned")),
		schema.StringField("reason", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Item-specific reason")),
		schema.StringField("condition", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Condition: unopened, opened, used, damaged")),
		schema.FloatField("refund_amount", schema.Default(0.0),
			schema.VerboseName("Refund Amount")),
		schema.BoolField("is_restockable", schema.Default(true),
			schema.VerboseName("Restockable"),
			schema.HelpText("Can be returned to inventory")),
		schema.TextField("inspection_note", schema.Optional(),
			schema.VerboseName("Inspection Note")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ReturnItem) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "return_items",
		VerboseName:       "Return Item",
		VerboseNamePlural: "Return Items",
		OrderBy:           []string{"return_request_id"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_return_item_request", "return_request_id"),
			schema.IndexOn("idx_return_item_order_item", "order_item_id"),
		},
	}
}

func (ReturnItem) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("return_request_id", "ReturnRequest",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("items")),
		schema.ForeignKeyField("order_item_id", "OrderItem",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("return_items")),
	}
}

func (ReturnItem) Hooks() *schema.ModelHooks {
	return nil
}

// StatusChange tracks status changes for tickets and returns
type StatusChange struct {
	schema.BaseSchema
	Id         int64  `json:"id" db:"id"`
	EntityType string `json:"entity_type" db:"entity_type"`
	EntityID   int64  `json:"entity_id" db:"entity_id"`
	FromStatus string `json:"from_status" db:"from_status"`
	ToStatus   string `json:"to_status" db:"to_status"`
	ChangedBy  int64  `json:"changed_by" db:"changed_by"`
	ChangerType string `json:"changer_type" db:"changer_type"`
	Note       string `json:"note" db:"note"`
	CreatedAt  string `json:"created_at" db:"created_at"`
}

func (StatusChange) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("entity_type", schema.MaxLength(50), schema.Required(),
			schema.VerboseName("Entity Type"),
			schema.HelpText("Type: support_ticket, return_request")),
		schema.Int64Field("entity_id", schema.Required(),
			schema.VerboseName("Entity ID")),
		schema.StringField("from_status", schema.MaxLength(20), schema.Optional(),
			schema.VerboseName("From Status")),
		schema.StringField("to_status", schema.MaxLength(20), schema.Required(),
			schema.VerboseName("To Status")),
		schema.Int64Field("changed_by", schema.Optional(),
			schema.VerboseName("Changed By")),
		schema.StringField("changer_type", schema.MaxLength(20), schema.Default("system"),
			schema.VerboseName("Changer Type"),
			schema.HelpText("Type: customer, agent, system")),
		schema.TextField("note", schema.Optional(),
			schema.HelpText("Optional note about the change")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (StatusChange) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "status_changes",
		VerboseName:       "Status Change",
		VerboseNamePlural: "Status Changes",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_status_change_entity", "entity_type", "entity_id"),
		},
	}
}

func (StatusChange) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (StatusChange) Hooks() *schema.ModelHooks {
	return nil
}

// ChatMessage represents messages in live chat sessions
type ChatMessage struct {
	schema.BaseSchema
	Id         int64  `json:"id" db:"id"`
	SessionID  int64  `json:"session_id" db:"session_id"`
	SenderType string `json:"sender_type" db:"sender_type"`
	SenderID   int64  `json:"sender_id" db:"sender_id"`
	Message    string `json:"message" db:"message"`
	IsRead     bool   `json:"is_read" db:"is_read"`
	ReadAt     string `json:"read_at" db:"read_at"`
	CreatedAt  string `json:"created_at" db:"created_at"`
}

func (ChatMessage) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("session_id", schema.Required(),
			schema.VerboseName("Session")),
		schema.StringField("sender_type", schema.MaxLength(20), schema.Required(),
			schema.VerboseName("Sender Type"),
			schema.HelpText("Type: customer, agent")),
		schema.Int64Field("sender_id", schema.Optional(),
			schema.VerboseName("Sender ID")),
		schema.TextField("message", schema.Required(),
			schema.HelpText("Message content")),
		schema.BoolField("is_read", schema.Default(false),
			schema.VerboseName("Read")),
		schema.TimeField("read_at", schema.Optional(),
			schema.VerboseName("Read At")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (ChatMessage) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "chat_messages",
		VerboseName:       "Chat Message",
		VerboseNamePlural: "Chat Messages",
		OrderBy:           []string{"created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_chat_message_session", "session_id"),
		},
	}
}

func (ChatMessage) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("session_id", "LiveChatSession",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("messages")),
	}
}

func (ChatMessage) Hooks() *schema.ModelHooks {
	return nil
}
