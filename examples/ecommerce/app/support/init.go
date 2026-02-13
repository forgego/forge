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
}
