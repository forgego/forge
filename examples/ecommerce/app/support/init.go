package support

import (
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/registry"
)

// Init initializes the support module
func Init(database *db.DB) {
	// Register models with schema registry
	registry.RegisterModel(SupportTicket{})
	registry.RegisterModel(SupportMessage{})
	registry.RegisterModel(ReturnRequest{})
	registry.RegisterModel(LiveChatSession{})
	registry.RegisterModel(FAQ{})
	registry.RegisterModel(Attachment{})
	registry.RegisterModel(ReturnItem{})
	registry.RegisterModel(StatusChange{})
	registry.RegisterModel(ChatMessage{})

	if SupportTicketObjects != nil {
		SupportTicketObjects.SetDB(database)
	}
	if SupportMessageObjects != nil {
		SupportMessageObjects.SetDB(database)
	}
	if ReturnRequestObjects != nil {
		ReturnRequestObjects.SetDB(database)
	}
	if LiveChatSessionObjects != nil {
		LiveChatSessionObjects.SetDB(database)
	}
	if FAQObjects != nil {
		FAQObjects.SetDB(database)
	}
	if AttachmentObjects != nil {
		AttachmentObjects.SetDB(database)
	}
	if ReturnItemObjects != nil {
		ReturnItemObjects.SetDB(database)
	}
	if StatusChangeObjects != nil {
		StatusChangeObjects.SetDB(database)
	}
	if ChatMessageObjects != nil {
		ChatMessageObjects.SetDB(database)
	}
}
