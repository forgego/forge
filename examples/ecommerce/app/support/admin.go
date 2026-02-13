package support

import (
	"context"

	"github.com/forgego/forge/admin"
	adminCore "github.com/forgego/forge/admin/core"
)

// RegisterAdmin registers all support models with the admin interface
func RegisterAdmin(ctx context.Context) {
	site := admin.DefaultSite

	// SupportTicket Admin
	supportTicketAdmin := adminCore.NewModelAdmin(
		SupportTicket{},
		adminCore.WithListDisplay("id", "ticket_number", "customer_id", "subject", "status", "priority", "category", "assigned_to", "created_at"),
		adminCore.WithSearchFields("ticket_number", "subject", "description"),
		adminCore.WithListFilter("status", "priority", "category", "source"),
		adminCore.WithOrdering("-created_at"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, supportTicketAdmin)

	// SupportMessage Admin
	supportMessageAdmin := adminCore.NewModelAdmin(
		SupportMessage{},
		adminCore.WithListDisplay("id", "ticket_id", "sender_type", "sender_name", "is_internal", "created_at"),
		adminCore.WithSearchFields("message", "sender_name"),
		adminCore.WithListFilter("sender_type", "is_internal"),
		adminCore.WithOrdering("created_at"),
		adminCore.WithListPerPage(100),
	)
	site.RegisterModel(ctx, supportMessageAdmin)

	// ReturnRequest Admin
	returnRequestAdmin := adminCore.NewModelAdmin(
		ReturnRequest{},
		adminCore.WithListDisplay("id", "return_number", "order_id", "customer_id", "reason", "status", "refund_amount", "created_at"),
		adminCore.WithSearchFields("return_number", "order_id", "customer_id"),
		adminCore.WithListFilter("status", "reason", "refund_method"),
		adminCore.WithOrdering("-created_at"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, returnRequestAdmin)

	// LiveChatSession Admin
	liveChatSessionAdmin := adminCore.NewModelAdmin(
		LiveChatSession{},
		adminCore.WithListDisplay("id", "session_id", "customer_id", "agent_id", "status", "duration", "rating", "started_at"),
		adminCore.WithSearchFields("session_id", "customer_id", "agent_id"),
		adminCore.WithListFilter("status", "rating"),
		adminCore.WithOrdering("-started_at"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, liveChatSessionAdmin)

	// FAQ Admin
	faqAdmin := adminCore.NewModelAdmin(
		FAQ{},
		adminCore.WithListDisplay("id", "question", "category", "is_public", "view_count", "helpful_yes", "helpful_no", "sort_order"),
		adminCore.WithSearchFields("question", "answer", "tags"),
		adminCore.WithListFilter("is_public", "category"),
		adminCore.WithOrdering("sort_order", "-view_count"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, faqAdmin)

	// Attachment Admin
	attachmentAdmin := adminCore.NewModelAdmin(
		Attachment{},
		adminCore.WithListDisplay("id", "entity_type", "entity_id", "file_name", "file_size", "uploader_type", "created_at"),
		adminCore.WithSearchFields("file_name", "entity_type", "entity_id"),
		adminCore.WithListFilter("entity_type", "uploader_type"),
		adminCore.WithOrdering("-created_at"),
		adminCore.WithListPerPage(100),
	)
	site.RegisterModel(ctx, attachmentAdmin)

	// ReturnItem Admin
	returnItemAdmin := adminCore.NewModelAdmin(
		ReturnItem{},
		adminCore.WithListDisplay("id", "return_request_id", "order_item_id", "quantity", "condition", "refund_amount", "is_restockable"),
		adminCore.WithSearchFields("return_request_id", "order_item_id"),
		adminCore.WithListFilter("condition", "is_restockable"),
		adminCore.WithOrdering("return_request_id"),
		adminCore.WithListPerPage(100),
	)
	site.RegisterModel(ctx, returnItemAdmin)

	// StatusChange Admin
	statusChangeAdmin := adminCore.NewModelAdmin(
		StatusChange{},
		adminCore.WithListDisplay("id", "entity_type", "entity_id", "from_status", "to_status", "changer_type", "created_at"),
		adminCore.WithSearchFields("entity_type", "entity_id"),
		adminCore.WithListFilter("entity_type", "changer_type"),
		adminCore.WithOrdering("-created_at"),
		adminCore.WithListPerPage(100),
	)
	site.RegisterModel(ctx, statusChangeAdmin)

	// ChatMessage Admin
	chatMessageAdmin := adminCore.NewModelAdmin(
		ChatMessage{},
		adminCore.WithListDisplay("id", "session_id", "sender_type", "is_read", "created_at"),
		adminCore.WithSearchFields("message", "session_id"),
		adminCore.WithListFilter("sender_type", "is_read"),
		adminCore.WithOrdering("created_at"),
		adminCore.WithListPerPage(100),
	)
	site.RegisterModel(ctx, chatMessageAdmin)
}
