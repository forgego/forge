package support

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers all support API endpoints
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// SupportTicket ViewSet
	supportTicketViewSet := api.NewModelViewSet(
		SupportTicket{},
		database,
		api.WithFilterFields("customer_id", "status", "priority", "category", "assigned_to"),
		api.WithSearchFields("ticket_number", "subject", "description"),
		api.WithOrderingFields("-created_at", "priority"),
	)
	router.RegisterViewSet("support-tickets", supportTicketViewSet)

	// SupportMessage ViewSet
	supportMessageViewSet := api.NewModelViewSet(
		SupportMessage{},
		database,
		api.WithFilterFields("ticket_id", "sender_type", "is_internal"),
		api.WithSearchFields("message", "sender_name"),
		api.WithOrderingFields("created_at"),
	)
	router.RegisterViewSet("support-messages", supportMessageViewSet)

	// ReturnRequest ViewSet
	returnRequestViewSet := api.NewModelViewSet(
		ReturnRequest{},
		database,
		api.WithFilterFields("customer_id", "order_id", "status", "reason"),
		api.WithSearchFields("return_number"),
		api.WithOrderingFields("-created_at"),
	)
	router.RegisterViewSet("return-requests", returnRequestViewSet)

	// LiveChatSession ViewSet
	liveChatSessionViewSet := api.NewModelViewSet(
		LiveChatSession{},
		database,
		api.WithFilterFields("customer_id", "agent_id", "status"),
		api.WithSearchFields("session_id"),
		api.WithOrderingFields("-started_at"),
	)
	router.RegisterViewSet("live-chat-sessions", liveChatSessionViewSet)

	// FAQ ViewSet (public endpoints)
	faqViewSet := api.NewModelViewSet(
		FAQ{},
		database,
		api.WithFilterFields("category", "is_public"),
		api.WithSearchFields("question", "answer", "tags"),
		api.WithOrderingFields("sort_order", "-view_count"),
	)
	router.RegisterViewSet("faqs", faqViewSet)

	// Attachment ViewSet
	attachmentViewSet := api.NewModelViewSet(
		Attachment{},
		database,
		api.WithFilterFields("entity_type", "entity_id", "uploader_type"),
		api.WithSearchFields("file_name"),
		api.WithOrderingFields("-created_at"),
	)
	router.RegisterViewSet("attachments", attachmentViewSet)

	// ReturnItem ViewSet
	returnItemViewSet := api.NewModelViewSet(
		ReturnItem{},
		database,
		api.WithFilterFields("return_request_id", "order_item_id", "is_restockable"),
		api.WithSearchFields("return_request_id"),
		api.WithOrderingFields("return_request_id"),
	)
	router.RegisterViewSet("return-items", returnItemViewSet)

	// StatusChange ViewSet (read-only audit log)
	statusChangeViewSet := api.NewModelViewSet(
		StatusChange{},
		database,
		api.WithFilterFields("entity_type", "entity_id", "changer_type"),
		api.WithSearchFields("entity_type"),
		api.WithOrderingFields("-created_at"),
	)
	router.RegisterViewSet("status-changes", statusChangeViewSet)

	// ChatMessage ViewSet
	chatMessageViewSet := api.NewModelViewSet(
		ChatMessage{},
		database,
		api.WithFilterFields("session_id", "sender_type", "is_read"),
		api.WithSearchFields("message"),
		api.WithOrderingFields("created_at"),
	)
	router.RegisterViewSet("chat-messages", chatMessageViewSet)
}
