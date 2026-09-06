package support

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers support API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("support-tickets", &api.ViewSetConfig{
		Model:      &SupportTicket{},
		Queryset:   SupportTicketObjects,
		Serializer: base,
	})

	router.Register("support-messages", &api.ViewSetConfig{
		Model:      &SupportMessage{},
		Queryset:   SupportMessageObjects,
		Serializer: base,
	})

	router.Register("return-requests", &api.ViewSetConfig{
		Model:      &ReturnRequest{},
		Queryset:   ReturnRequestObjects,
		Serializer: base,
	})

	router.Register("live-chat-sessions", &api.ViewSetConfig{
		Model:      &LiveChatSession{},
		Queryset:   LiveChatSessionObjects,
		Serializer: base,
	})

	router.Register("faqs", &api.ViewSetConfig{
		Model:      &FAQ{},
		Queryset:   FAQObjects,
		Serializer: base,
	})

	router.Register("attachments", &api.ViewSetConfig{
		Model:      &Attachment{},
		Queryset:   AttachmentObjects,
		Serializer: base,
	})

	router.Register("return-items", &api.ViewSetConfig{
		Model:      &ReturnItem{},
		Queryset:   ReturnItemObjects,
		Serializer: base,
	})

	router.Register("status-changes", &api.ViewSetConfig{
		Model:      &StatusChange{},
		Queryset:   StatusChangeObjects,
		Serializer: base,
	})

	router.Register("chat-messages", &api.ViewSetConfig{
		Model:      &ChatMessage{},
		Queryset:   ChatMessageObjects,
		Serializer: base,
	})
}
