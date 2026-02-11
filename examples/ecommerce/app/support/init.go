package support

import "github.com/forgego/forge/db"

// Init initializes the support package with database connection
func Init(database *db.DB) {
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
}

import "github.com/forgego/forge/db"

// Init initializes the support package with database connection
func Init(database *db.DB) {
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
}

