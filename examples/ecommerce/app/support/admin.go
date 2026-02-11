package support

import (
	"context"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers support models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// SupportTicket admin
	admin.Register(&admin.Config[SupportTicket]{
		Icon: "Ticket",
		ListDisplay: []admin.Field{
			SupportTicketFieldsInstance.TicketNumber,
			SupportTicketFieldsInstance.Type,
			SupportTicketFieldsInstance.Priority,
			SupportTicketFieldsInstance.Status,
			SupportTicketFieldsInstance.AssignedGroup,
			SupportTicketFieldsInstance.CustomerID,
			SupportTicketFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			SupportTicketFieldsInstance.Status,
			SupportTicketFieldsInstance.Priority,
			SupportTicketFieldsInstance.Type,
			SupportTicketFieldsInstance.AssignedGroup,
			SupportTicketFieldsInstance.IsEscalated,
		},
		SearchFields: []admin.Field{
			SupportTicketFieldsInstance.TicketNumber,
			SupportTicketFieldsInstance.Subject,
			SupportTicketFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			SupportTicketFieldsInstance.Priority,
			SupportTicketFieldsInstance.CreatedAt,
		},
		Fieldsets: []admin.Fieldset[SupportTicket]{
			{
				Name: "Basic Information",
				Fields: []string{"ticket_number", "customer_id", "type", "subject", "description"},
			},
			{
				Name: "Status & Priority",
				Fields: []string{"priority", "status", "is_escalated"},
			},
			{
				Name: "Assignment",
				Fields: []string{"assigned_to", "assigned_group"},
			},
			{
				Name: "Related Entity",
				Fields: []string{"related_type", "related_id", "tags"},
			},
			{
				Name: "Resolution",
				Fields: []string{"resolution", "resolved_at", "customer_satisfaction", "customer_feedback"},
			},
			{
				Name: "Escalation",
				Fields: []string{"escalation_reason", "escalated_at"},
			},
			{
				Name: "Internal",
				Fields: []string{"first_response_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "priority", "assigned_to", "resolution"},
		},
		ReadOnlyFields: []admin.Field{
			SupportTicketFieldsInstance.TicketNumber,
			SupportTicketFieldsInstance.CreatedAt,
			SupportTicketFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}, obj *SupportTicket) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}, obj *SupportTicket) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}, obj *SupportTicket) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}, obj *SupportTicket) bool {
			return true
		},
		Filters: []admin.Filter[SupportTicket]{
			{
				Name:  "open_tickets",
				Label: "Open Tickets",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("status", "open")
				},
			},
			{
				Name:  "urgent_tickets",
				Label: "Urgent Tickets",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("priority", "urgent")
				},
			},
			{
				Name:  "escalated_tickets",
				Label: "Escalated Tickets",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("is_escalated", true)
				},
			},
			{
				Name:  "by_status",
				Label: "By Status",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "open", Label: "Open"},
					{Value: "pending_customer_response", Label: "Pending Customer Response"},
					{Value: "pending_merchant_response", Label: "Pending Merchant Response"},
					{Value: "resolved", Label: "Resolved"},
					{Value: "closed", Label: "Closed"},
					{Value: "spam", Label: "Spam"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("status", value)
				},
			},
			{
				Name:  "by_priority",
				Label: "By Priority",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "low", Label: "Low"},
					{Value: "medium", Label: "Medium"},
					{Value: "high", Label: "High"},
					{Value: "urgent", Label: "Urgent"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("priority", value)
				},
			},
			{
				Name:  "by_type",
				Label: "By Type",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "order_inquiry", Label: "Order Inquiry"},
					{Value: "product_inquiry", Label: "Product Inquiry"},
					{Value: "shipping_issue", Label: "Shipping Issue"},
					{Value: "payment_issue", Label: "Payment Issue"},
					{Value: "return_refund", Label: "Return/Refund"},
					{Value: "technical", Label: "Technical"},
					{Value: "account", Label: "Account"},
					{Value: "billing", Label: "Billing"},
					{Value: "general", Label: "General"},
					{Value: "feedback", Label: "Feedback"},
					{Value: "praise", Label: "Praise"},
					{Value: "suggestion", Label: "Suggestion"},
					{Value: "other", Label: "Other"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("type", value)
				},
			},
		},
		Actions: []admin.Action[SupportTicket]{
			{
				Name:         "assign",
				Label:        "Assign Ticket",
				Icon:         "User",
				Confirmation: "Assign the selected tickets to an agent?",
				Handler: func(ctx context.Context, admin *admin.Admin[SupportTicket], ids []interface{}) error {
					for _, id := range ids {
						ticket, err := SupportTicketObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						// Assignment logic would go here
						if err := SupportTicketObjects.Update(ctx, ticket); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "escalate",
				Label:        "Escalate Ticket",
				Icon:         "ArrowUp",
				Confirmation: "Escalate the selected tickets?",
				Handler: func(ctx context.Context, admin *admin.Admin[SupportTicket], ids []interface{}) error {
					for _, id := range ids {
						ticket, err := SupportTicketObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						ticket.IsEscalated = true
						if err := SupportTicketObjects.Update(ctx, ticket); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "resolve",
				Label:        "Resolve Ticket",
				Icon:         "Check",
				Confirmation: "Mark the selected tickets as resolved?",
				Handler: func(ctx context.Context, admin *admin.Admin[SupportTicket], ids []interface{}) error {
					for _, id := range ids {
						ticket, err := SupportTicketObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						ticket.Status = "resolved"
						if err := SupportTicketObjects.Update(ctx, ticket); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "reopen",
				Label:        "Reopen Ticket",
				Icon:         "Refresh",
				Confirmation: "Reopen the selected tickets?",
				Handler: func(ctx context.Context, admin *admin.Admin[SupportTicket], ids []interface{}) error {
					for _, id := range ids {
						ticket, err := SupportTicketObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						ticket.Status = "open"
						if err := SupportTicketObjects.Update(ctx, ticket); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}) bool {
					return true
				},
			},
		},
	})

	// ReturnRequest admin
	admin.Register(&admin.Config[ReturnRequest]{
		Icon: "RotateCcw",
		ListDisplay: []admin.Field{
			ReturnRequestFieldsInstance.ReturnNumber,
			ReturnRequestFieldsInstance.OrderID,
			ReturnRequestFieldsInstance.Reason,
			ReturnRequestFieldsInstance.ResolutionType,
			ReturnRequestFieldsInstance.Status,
			ReturnRequestFieldsInstance.RefundAmount,
			ReturnRequestFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			ReturnRequestFieldsInstance.Status,
			ReturnRequestFieldsInstance.Reason,
			ReturnRequestFieldsInstance.ResolutionType,
			ReturnRequestFieldsInstance.RefundStatus,
		},
		SearchFields: []admin.Field{
			ReturnRequestFieldsInstance.ReturnNumber,
		},
		Ordering: []admin.Field{
			ReturnRequestFieldsInstance.CreatedAt,
		},
		Fieldsets: []admin.Fieldset[ReturnRequest]{
			{
				Name: "Basic Information",
				Fields: []string{"return_number", "order_id", "customer_id", "reason", "reason_detail"},
			},
			{
				Name: "Items",
				Fields: []string{"items"},
			},
			{
				Name: "Resolution",
				Fields: []string{"resolution_type", "resolution_detail"},
			},
			{
				Name: "Refund Details",
				Fields: []string{"refund_amount", "refund_method", "refund_status", "refund_processed_at"},
			},
			{
				Name: "Return Shipping",
				Fields: []string{"return_shipping_method", "return_shipping_label_url", "return_tracking_number", "return_carrier"},
			},
			{
				Name: "Status",
				Fields: []string{"status", "status_history"},
			},
			{
				Name: "Inspection",
				Fields: []string{"inspected_at", "inspection_notes", "inspection_result"},
			},
			{
				Name: "Timeline",
				Fields: []string{"approved_at", "returned_at", "completed_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "resolution_type", "refund_amount", "refund_status"},
		},
		ReadOnlyFields: []admin.Field{
			ReturnRequestFieldsInstance.ReturnNumber,
			ReturnRequestFieldsInstance.CreatedAt,
			ReturnRequestFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}, obj *ReturnRequest) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}, obj *ReturnRequest) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}, obj *ReturnRequest) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}, obj *ReturnRequest) bool {
			return true
		},
		Filters: []admin.Filter[ReturnRequest]{
			{
				Name:  "pending_approval",
				Label: "Pending Approval",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ReturnRequest], value interface{}) orm.QuerySet[ReturnRequest] {
					return qs.Filter("status", "pending_approval")
				},
			},
			{
				Name:  "in_progress",
				Label: "In Progress",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ReturnRequest], value interface{}) orm.QuerySet[ReturnRequest] {
					return qs.Filter("status__in", []string{"approved", "pending_return", "in_transit", "received_inspected"})
				},
			},
			{
				Name:  "completed",
				Label: "Completed",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ReturnRequest], value interface{}) orm.QuerySet[ReturnRequest] {
					return qs.Filter("status__in", []string{"refunded", "exchanged", "completed"})
				},
			},
			{
				Name:  "by_status",
				Label: "By Status",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "pending_approval", Label: "Pending Approval"},
					{Value: "approved", Label: "Approved"},
					{Value: "rejected", Label: "Rejected"},
					{Value: "pending_return", Label: "Pending Return"},
					{Value: "in_transit", Label: "In Transit"},
					{Value: "received_inspected", Label: "Received & Inspected"},
					{Value: "pending_refund", Label: "Pending Refund"},
					{Value: "refunded", Label: "Refunded"},
					{Value: "exchanged", Label: "Exchanged"},
					{Value: "completed", Label: "Completed"},
					{Value: "cancelled", Label: "Cancelled"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[ReturnRequest], value interface{}) orm.QuerySet[ReturnRequest] {
					return qs.Filter("status", value)
				},
			},
			{
				Name:  "by_reason",
				Label: "By Reason",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "defective", Label: "Defective"},
					{Value: "wrong_item", Label: "Wrong Item"},
					{Value: "not_as_described", Label: "Not As Described"},
					{Value: "changed_mind", Label: "Changed Mind"},
					{Value: "size_issues", Label: "Size Issues"},
					{Value: "quality_issues", Label: "Quality Issues"},
					{Value: "other", Label: "Other"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[ReturnRequest], value interface{}) orm.QuerySet[ReturnRequest] {
					return qs.Filter("reason", value)
				},
			},
		},
		Actions: []admin.Action[ReturnRequest]{
			{
				Name:         "approve",
				Label:        "Approve Return",
				Icon:         "Check",
				Confirmation: "Approve the selected returns?",
				Handler: func(ctx context.Context, admin *admin.Admin[ReturnRequest], ids []interface{}) error {
					for _, id := range ids {
						returnReq, err := ReturnRequestObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						returnReq.Status = "approved"
						if err := ReturnRequestObjects.Update(ctx, returnReq); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "reject",
				Label:        "Reject Return",
				Icon:         "X",
				Confirmation: "Reject the selected returns?",
				Handler: func(ctx context.Context, admin *admin.Admin[ReturnRequest], ids []interface{}) error {
					for _, id := range ids {
						returnReq, err := ReturnRequestObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						returnReq.Status = "rejected"
						if err := ReturnRequestObjects.Update(ctx, returnReq); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "process_refund",
				Label:        "Process Refund",
				Icon:         "DollarSign",
				Confirmation: "Process refunds for selected returns?",
				Handler: func(ctx context.Context, admin *admin.Admin[ReturnRequest], ids []interface{}) error {
					for _, id := range ids {
						returnReq, err := ReturnRequestObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						returnReq.RefundStatus = "processing"
						if err := ReturnRequestObjects.Update(ctx, returnReq); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}) bool {
					return true
				},
			},
		},
	})

	// LiveChatSession admin
	admin.Register(&admin.Config[LiveChatSession]{
		Icon: "MessageSquare",
		ListDisplay: []admin.Field{
			LiveChatSessionFieldsInstance.SessionID,
			LiveChatSessionFieldsInstance.Status,
			LiveChatSessionFieldsInstance.Type,
			LiveChatSessionFieldsInstance.CustomerID,
			LiveChatSessionFieldsInstance.AgentID,
			LiveChatSessionFieldsInstance.WaitTimeSeconds,
			LiveChatSessionFieldsInstance.ChatDurationSeconds,
			LiveChatSessionFieldsInstance.StartedAt,
		},
		ListFilter: []admin.Field{
			LiveChatSessionFieldsInstance.Status,
			LiveChatSessionFieldsInstance.Type,
			LiveChatSessionFieldsInstance.AgentID,
		},
		SearchFields: []admin.Field{
			LiveChatSessionFieldsInstance.SessionID,
			LiveChatSessionFieldsInstance.GuestName,
			LiveChatSessionFieldsInstance.GuestEmail,
		},
		Ordering: []admin.Field{
			LiveChatSessionFieldsInstance.StartedAt,
		},
		Fieldsets: []admin.Fieldset[LiveChatSession]{
			{
				Name: "Basic Information",
				Fields: []string{"session_id", "status", "type", "subject"},
			},
			{
				Name: "Customer",
				Fields: []string{"customer_id", "guest_name", "guest_email"},
			},
			{
				Name: "Agent",
				Fields: []string{"agent_id", "agent_name"},
			},
			{
				Name: "Duration",
				Fields: []string{"wait_time_seconds", "chat_duration_seconds"},
			},
			{
				Name: "End Details",
				Fields: []string{"end_reason", "end_notes", "customer_satisfaction"},
			},
			{
				Name: "Linked Entities",
				Fields: []string{"related_ticket_id", "related_order_id"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "agent_id", "end_reason"},
		},
		ReadOnlyFields: []admin.Field{
			LiveChatSessionFieldsInstance.SessionID,
			LiveChatSessionFieldsInstance.StartedAt,
			LiveChatSessionFieldsInstance.EndedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}, obj *LiveChatSession) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}, obj *LiveChatSession) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}, obj *LiveChatSession) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}, obj *LiveChatSession) bool {
			return true
		},
		Filters: []admin.Filter[LiveChatSession]{
			{
				Name:  "active_chats",
				Label: "Active Chats",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[LiveChatSession], value interface{}) orm.QuerySet[LiveChatSession] {
					return qs.Filter("status__in", []string{"waiting", "active"})
				},
			},
			{
				Name:  "ended_chats",
				Label: "Ended Chats",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[LiveChatSession], value interface{}) orm.QuerySet[LiveChatSession] {
					return qs.Filter("status__in", []string{"ended", "abandoned"})
				},
			},
			{
				Name:  "by_status",
				Label: "By Status",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "waiting", Label: "Waiting"},
					{Value: "active", Label: "Active"},
					{Value: "transferred", Label: "Transferred"},
					{Value: "ended", Label: "Ended"},
					{Value: "abandoned", Label: "Abandoned"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[LiveChatSession], value interface{}) orm.QuerySet[LiveChatSession] {
					return qs.Filter("status", value)
				},
			},
		},
		Actions: []admin.Action[LiveChatSession]{
			{
				Name:         "start",
				Label:        "Start Chat",
				Icon:         "Play",
				Confirmation: "Start a new chat session?",
				Handler: func(ctx context.Context, admin *admin.Admin[LiveChatSession], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "end",
				Label:        "End Chat",
				Icon:         "Square",
				Confirmation: "End the selected chat sessions?",
				Handler: func(ctx context.Context, admin *admin.Admin[LiveChatSession], ids []interface{}) error {
					for _, id := range ids {
						session, err := LiveChatSessionObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						session.Status = "ended"
						if err := LiveChatSessionObjects.Update(ctx, session); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "transfer",
				Label:        "Transfer Chat",
				Icon:         "ArrowRight",
				Confirmation: "Transfer the selected chats to another agent?",
				Handler: func(ctx context.Context, admin *admin.Admin[LiveChatSession], ids []interface{}) error {
					for _, id := range ids {
						session, err := LiveChatSessionObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						session.Status = "transferred"
						if err := LiveChatSessionObjects.Update(ctx, session); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}) bool {
					return true
				},
			},
		},
	})

	// FAQ admin
	admin.Register(&admin.Config[FAQ]{
		Icon: "HelpCircle",
		ListDisplay: []admin.Field{
			FAQFieldsInstance.Question,
			FAQFieldsInstance.Category,
			FAQFieldsInstance.IsFeatured,
			FAQFieldsInstance.ViewCount,
			FAQFieldsInstance.HelpfulCount,
			FAQFieldsInstance.IsVisible,
			FAQFieldsInstance.DisplayOrder,
		},
		ListFilter: []admin.Field{
			FAQFieldsInstance.Category,
			FAQFieldsInstance.IsVisible,
			FAQFieldsInstance.IsFeatured,
		},
		SearchFields: []admin.Field{
			FAQFieldsInstance.Question,
			FAQFieldsInstance.Answer,
			FAQFieldsInstance.Keywords,
		},
		Ordering: []admin.Field{
			FAQFieldsInstance.DisplayOrder,
			FAQFieldsInstance.Category,
		},
		Fieldsets: []admin.Fieldset[FAQ]{
			{
				Name: "Content",
				Fields: []string{"question", "answer", "category", "keywords"},
			},
			{
				Name: "Display",
				Fields: []string{"display_order", "is_visible", "is_featured"},
			},
			{
				Name: "Statistics",
				Fields: []string{"view_count", "helpful_count", "not_helpful_count"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"question", "answer", "is_visible", "is_featured"},
		},
		ReadOnlyFields: []admin.Field{
			FAQFieldsInstance.ViewCount,
			FAQFieldsInstance.HelpfulCount,
			FAQFieldsInstance.NotHelpfulCount,
			FAQFieldsInstance.CreatedAt,
			FAQFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}, obj *FAQ) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}, obj *FAQ) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}, obj *FAQ) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}, obj *FAQ) bool {
			return true
		},
		Filters: []admin.Filter[FAQ]{
			{
				Name:  "visible",
				Label: "Visible FAQs",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[FAQ], value interface{}) orm.QuerySet[FAQ] {
					return qs.Filter("is_visible", true)
				},
			},
			{
				Name:  "featured",
				Label: "Featured FAQs",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[FAQ], value interface{}) orm.QuerySet[FAQ] {
					return qs.Filter("is_featured", true)
				},
			},
			{
				Name:  "by_category",
				Label: "By Category",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "orders", Label: "Orders"},
					{Value: "shipping", Label: "Shipping"},
					{Value: "payments", Label: "Payments"},
					{Value: "returns", Label: "Returns"},
					{Value: "products", Label: "Products"},
					{Value: "account", Label: "Account"},
					{Value: "technical", Label: "Technical"},
					{Value: "general", Label: "General"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[FAQ], value interface{}) orm.QuerySet[FAQ] {
					return qs.Filter("category", value)
				},
			},
		},
		Actions: []admin.Action[FAQ]{
			{
				Name:         "toggle_visibility",
				Label:        "Toggle Visibility",
				Icon:         "Eye",
				Confirmation: "Toggle visibility for selected FAQs?",
				Handler: func(ctx context.Context, admin *admin.Admin[FAQ], ids []interface{}) error {
					for _, id := range ids {
						faq, err := FAQObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						faq.IsVisible = !faq.IsVisible
						if err := FAQObjects.Update(ctx, faq); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "toggle_featured",
				Label:        "Toggle Featured",
				Icon:         "Star",
				Confirmation: "Toggle featured status for selected FAQs?",
				Handler: func(ctx context.Context, admin *admin.Admin[FAQ], ids []interface{}) error {
					for _, id := range ids {
						faq, err := FAQObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						faq.IsFeatured = !faq.IsFeatured
						if err := FAQObjects.Update(ctx, faq); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}) bool {
					return true
				},
			},
		},
	})
}

import (
	"context"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers support models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// SupportTicket admin
	admin.Register(&admin.Config[SupportTicket]{
		Icon: "Ticket",
		ListDisplay: []admin.Field{
			SupportTicketFieldsInstance.TicketNumber,
			SupportTicketFieldsInstance.Type,
			SupportTicketFieldsInstance.Priority,
			SupportTicketFieldsInstance.Status,
			SupportTicketFieldsInstance.AssignedGroup,
			SupportTicketFieldsInstance.CustomerID,
			SupportTicketFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			SupportTicketFieldsInstance.Status,
			SupportTicketFieldsInstance.Priority,
			SupportTicketFieldsInstance.Type,
			SupportTicketFieldsInstance.AssignedGroup,
			SupportTicketFieldsInstance.IsEscalated,
		},
		SearchFields: []admin.Field{
			SupportTicketFieldsInstance.TicketNumber,
			SupportTicketFieldsInstance.Subject,
			SupportTicketFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			SupportTicketFieldsInstance.Priority,
			SupportTicketFieldsInstance.CreatedAt,
		},
		Fieldsets: []admin.Fieldset[SupportTicket]{
			{
				Name: "Basic Information",
				Fields: []string{"ticket_number", "customer_id", "type", "subject", "description"},
			},
			{
				Name: "Status & Priority",
				Fields: []string{"priority", "status", "is_escalated"},
			},
			{
				Name: "Assignment",
				Fields: []string{"assigned_to", "assigned_group"},
			},
			{
				Name: "Related Entity",
				Fields: []string{"related_type", "related_id", "tags"},
			},
			{
				Name: "Resolution",
				Fields: []string{"resolution", "resolved_at", "customer_satisfaction", "customer_feedback"},
			},
			{
				Name: "Escalation",
				Fields: []string{"escalation_reason", "escalated_at"},
			},
			{
				Name: "Internal",
				Fields: []string{"first_response_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "priority", "assigned_to", "resolution"},
		},
		ReadOnlyFields: []admin.Field{
			SupportTicketFieldsInstance.TicketNumber,
			SupportTicketFieldsInstance.CreatedAt,
			SupportTicketFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}, obj *SupportTicket) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}, obj *SupportTicket) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}, obj *SupportTicket) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}, obj *SupportTicket) bool {
			return true
		},
		Filters: []admin.Filter[SupportTicket]{
			{
				Name:  "open_tickets",
				Label: "Open Tickets",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("status", "open")
				},
			},
			{
				Name:  "urgent_tickets",
				Label: "Urgent Tickets",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("priority", "urgent")
				},
			},
			{
				Name:  "escalated_tickets",
				Label: "Escalated Tickets",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("is_escalated", true)
				},
			},
			{
				Name:  "by_status",
				Label: "By Status",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "open", Label: "Open"},
					{Value: "pending_customer_response", Label: "Pending Customer Response"},
					{Value: "pending_merchant_response", Label: "Pending Merchant Response"},
					{Value: "resolved", Label: "Resolved"},
					{Value: "closed", Label: "Closed"},
					{Value: "spam", Label: "Spam"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("status", value)
				},
			},
			{
				Name:  "by_priority",
				Label: "By Priority",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "low", Label: "Low"},
					{Value: "medium", Label: "Medium"},
					{Value: "high", Label: "High"},
					{Value: "urgent", Label: "Urgent"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("priority", value)
				},
			},
			{
				Name:  "by_type",
				Label: "By Type",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "order_inquiry", Label: "Order Inquiry"},
					{Value: "product_inquiry", Label: "Product Inquiry"},
					{Value: "shipping_issue", Label: "Shipping Issue"},
					{Value: "payment_issue", Label: "Payment Issue"},
					{Value: "return_refund", Label: "Return/Refund"},
					{Value: "technical", Label: "Technical"},
					{Value: "account", Label: "Account"},
					{Value: "billing", Label: "Billing"},
					{Value: "general", Label: "General"},
					{Value: "feedback", Label: "Feedback"},
					{Value: "praise", Label: "Praise"},
					{Value: "suggestion", Label: "Suggestion"},
					{Value: "other", Label: "Other"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[SupportTicket], value interface{}) orm.QuerySet[SupportTicket] {
					return qs.Filter("type", value)
				},
			},
		},
		Actions: []admin.Action[SupportTicket]{
			{
				Name:         "assign",
				Label:        "Assign Ticket",
				Icon:         "User",
				Confirmation: "Assign the selected tickets to an agent?",
				Handler: func(ctx context.Context, admin *admin.Admin[SupportTicket], ids []interface{}) error {
					for _, id := range ids {
						ticket, err := SupportTicketObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						// Assignment logic would go here
						if err := SupportTicketObjects.Update(ctx, ticket); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "escalate",
				Label:        "Escalate Ticket",
				Icon:         "ArrowUp",
				Confirmation: "Escalate the selected tickets?",
				Handler: func(ctx context.Context, admin *admin.Admin[SupportTicket], ids []interface{}) error {
					for _, id := range ids {
						ticket, err := SupportTicketObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						ticket.IsEscalated = true
						if err := SupportTicketObjects.Update(ctx, ticket); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "resolve",
				Label:        "Resolve Ticket",
				Icon:         "Check",
				Confirmation: "Mark the selected tickets as resolved?",
				Handler: func(ctx context.Context, admin *admin.Admin[SupportTicket], ids []interface{}) error {
					for _, id := range ids {
						ticket, err := SupportTicketObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						ticket.Status = "resolved"
						if err := SupportTicketObjects.Update(ctx, ticket); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "reopen",
				Label:        "Reopen Ticket",
				Icon:         "Refresh",
				Confirmation: "Reopen the selected tickets?",
				Handler: func(ctx context.Context, admin *admin.Admin[SupportTicket], ids []interface{}) error {
					for _, id := range ids {
						ticket, err := SupportTicketObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						ticket.Status = "open"
						if err := SupportTicketObjects.Update(ctx, ticket); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[SupportTicket], user interface{}) bool {
					return true
				},
			},
		},
	})

	// ReturnRequest admin
	admin.Register(&admin.Config[ReturnRequest]{
		Icon: "RotateCcw",
		ListDisplay: []admin.Field{
			ReturnRequestFieldsInstance.ReturnNumber,
			ReturnRequestFieldsInstance.OrderID,
			ReturnRequestFieldsInstance.Reason,
			ReturnRequestFieldsInstance.ResolutionType,
			ReturnRequestFieldsInstance.Status,
			ReturnRequestFieldsInstance.RefundAmount,
			ReturnRequestFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			ReturnRequestFieldsInstance.Status,
			ReturnRequestFieldsInstance.Reason,
			ReturnRequestFieldsInstance.ResolutionType,
			ReturnRequestFieldsInstance.RefundStatus,
		},
		SearchFields: []admin.Field{
			ReturnRequestFieldsInstance.ReturnNumber,
		},
		Ordering: []admin.Field{
			ReturnRequestFieldsInstance.CreatedAt,
		},
		Fieldsets: []admin.Fieldset[ReturnRequest]{
			{
				Name: "Basic Information",
				Fields: []string{"return_number", "order_id", "customer_id", "reason", "reason_detail"},
			},
			{
				Name: "Items",
				Fields: []string{"items"},
			},
			{
				Name: "Resolution",
				Fields: []string{"resolution_type", "resolution_detail"},
			},
			{
				Name: "Refund Details",
				Fields: []string{"refund_amount", "refund_method", "refund_status", "refund_processed_at"},
			},
			{
				Name: "Return Shipping",
				Fields: []string{"return_shipping_method", "return_shipping_label_url", "return_tracking_number", "return_carrier"},
			},
			{
				Name: "Status",
				Fields: []string{"status", "status_history"},
			},
			{
				Name: "Inspection",
				Fields: []string{"inspected_at", "inspection_notes", "inspection_result"},
			},
			{
				Name: "Timeline",
				Fields: []string{"approved_at", "returned_at", "completed_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "resolution_type", "refund_amount", "refund_status"},
		},
		ReadOnlyFields: []admin.Field{
			ReturnRequestFieldsInstance.ReturnNumber,
			ReturnRequestFieldsInstance.CreatedAt,
			ReturnRequestFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}, obj *ReturnRequest) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}, obj *ReturnRequest) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}, obj *ReturnRequest) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}, obj *ReturnRequest) bool {
			return true
		},
		Filters: []admin.Filter[ReturnRequest]{
			{
				Name:  "pending_approval",
				Label: "Pending Approval",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ReturnRequest], value interface{}) orm.QuerySet[ReturnRequest] {
					return qs.Filter("status", "pending_approval")
				},
			},
			{
				Name:  "in_progress",
				Label: "In Progress",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ReturnRequest], value interface{}) orm.QuerySet[ReturnRequest] {
					return qs.Filter("status__in", []string{"approved", "pending_return", "in_transit", "received_inspected"})
				},
			},
			{
				Name:  "completed",
				Label: "Completed",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ReturnRequest], value interface{}) orm.QuerySet[ReturnRequest] {
					return qs.Filter("status__in", []string{"refunded", "exchanged", "completed"})
				},
			},
			{
				Name:  "by_status",
				Label: "By Status",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "pending_approval", Label: "Pending Approval"},
					{Value: "approved", Label: "Approved"},
					{Value: "rejected", Label: "Rejected"},
					{Value: "pending_return", Label: "Pending Return"},
					{Value: "in_transit", Label: "In Transit"},
					{Value: "received_inspected", Label: "Received & Inspected"},
					{Value: "pending_refund", Label: "Pending Refund"},
					{Value: "refunded", Label: "Refunded"},
					{Value: "exchanged", Label: "Exchanged"},
					{Value: "completed", Label: "Completed"},
					{Value: "cancelled", Label: "Cancelled"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[ReturnRequest], value interface{}) orm.QuerySet[ReturnRequest] {
					return qs.Filter("status", value)
				},
			},
			{
				Name:  "by_reason",
				Label: "By Reason",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "defective", Label: "Defective"},
					{Value: "wrong_item", Label: "Wrong Item"},
					{Value: "not_as_described", Label: "Not As Described"},
					{Value: "changed_mind", Label: "Changed Mind"},
					{Value: "size_issues", Label: "Size Issues"},
					{Value: "quality_issues", Label: "Quality Issues"},
					{Value: "other", Label: "Other"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[ReturnRequest], value interface{}) orm.QuerySet[ReturnRequest] {
					return qs.Filter("reason", value)
				},
			},
		},
		Actions: []admin.Action[ReturnRequest]{
			{
				Name:         "approve",
				Label:        "Approve Return",
				Icon:         "Check",
				Confirmation: "Approve the selected returns?",
				Handler: func(ctx context.Context, admin *admin.Admin[ReturnRequest], ids []interface{}) error {
					for _, id := range ids {
						returnReq, err := ReturnRequestObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						returnReq.Status = "approved"
						if err := ReturnRequestObjects.Update(ctx, returnReq); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "reject",
				Label:        "Reject Return",
				Icon:         "X",
				Confirmation: "Reject the selected returns?",
				Handler: func(ctx context.Context, admin *admin.Admin[ReturnRequest], ids []interface{}) error {
					for _, id := range ids {
						returnReq, err := ReturnRequestObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						returnReq.Status = "rejected"
						if err := ReturnRequestObjects.Update(ctx, returnReq); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "process_refund",
				Label:        "Process Refund",
				Icon:         "DollarSign",
				Confirmation: "Process refunds for selected returns?",
				Handler: func(ctx context.Context, admin *admin.Admin[ReturnRequest], ids []interface{}) error {
					for _, id := range ids {
						returnReq, err := ReturnRequestObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						returnReq.RefundStatus = "processing"
						if err := ReturnRequestObjects.Update(ctx, returnReq); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ReturnRequest], user interface{}) bool {
					return true
				},
			},
		},
	})

	// LiveChatSession admin
	admin.Register(&admin.Config[LiveChatSession]{
		Icon: "MessageSquare",
		ListDisplay: []admin.Field{
			LiveChatSessionFieldsInstance.SessionID,
			LiveChatSessionFieldsInstance.Status,
			LiveChatSessionFieldsInstance.Type,
			LiveChatSessionFieldsInstance.CustomerID,
			LiveChatSessionFieldsInstance.AgentID,
			LiveChatSessionFieldsInstance.WaitTimeSeconds,
			LiveChatSessionFieldsInstance.ChatDurationSeconds,
			LiveChatSessionFieldsInstance.StartedAt,
		},
		ListFilter: []admin.Field{
			LiveChatSessionFieldsInstance.Status,
			LiveChatSessionFieldsInstance.Type,
			LiveChatSessionFieldsInstance.AgentID,
		},
		SearchFields: []admin.Field{
			LiveChatSessionFieldsInstance.SessionID,
			LiveChatSessionFieldsInstance.GuestName,
			LiveChatSessionFieldsInstance.GuestEmail,
		},
		Ordering: []admin.Field{
			LiveChatSessionFieldsInstance.StartedAt,
		},
		Fieldsets: []admin.Fieldset[LiveChatSession]{
			{
				Name: "Basic Information",
				Fields: []string{"session_id", "status", "type", "subject"},
			},
			{
				Name: "Customer",
				Fields: []string{"customer_id", "guest_name", "guest_email"},
			},
			{
				Name: "Agent",
				Fields: []string{"agent_id", "agent_name"},
			},
			{
				Name: "Duration",
				Fields: []string{"wait_time_seconds", "chat_duration_seconds"},
			},
			{
				Name: "End Details",
				Fields: []string{"end_reason", "end_notes", "customer_satisfaction"},
			},
			{
				Name: "Linked Entities",
				Fields: []string{"related_ticket_id", "related_order_id"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "agent_id", "end_reason"},
		},
		ReadOnlyFields: []admin.Field{
			LiveChatSessionFieldsInstance.SessionID,
			LiveChatSessionFieldsInstance.StartedAt,
			LiveChatSessionFieldsInstance.EndedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}, obj *LiveChatSession) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}, obj *LiveChatSession) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}, obj *LiveChatSession) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}, obj *LiveChatSession) bool {
			return true
		},
		Filters: []admin.Filter[LiveChatSession]{
			{
				Name:  "active_chats",
				Label: "Active Chats",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[LiveChatSession], value interface{}) orm.QuerySet[LiveChatSession] {
					return qs.Filter("status__in", []string{"waiting", "active"})
				},
			},
			{
				Name:  "ended_chats",
				Label: "Ended Chats",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[LiveChatSession], value interface{}) orm.QuerySet[LiveChatSession] {
					return qs.Filter("status__in", []string{"ended", "abandoned"})
				},
			},
			{
				Name:  "by_status",
				Label: "By Status",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "waiting", Label: "Waiting"},
					{Value: "active", Label: "Active"},
					{Value: "transferred", Label: "Transferred"},
					{Value: "ended", Label: "Ended"},
					{Value: "abandoned", Label: "Abandoned"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[LiveChatSession], value interface{}) orm.QuerySet[LiveChatSession] {
					return qs.Filter("status", value)
				},
			},
		},
		Actions: []admin.Action[LiveChatSession]{
			{
				Name:         "start",
				Label:        "Start Chat",
				Icon:         "Play",
				Confirmation: "Start a new chat session?",
				Handler: func(ctx context.Context, admin *admin.Admin[LiveChatSession], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "end",
				Label:        "End Chat",
				Icon:         "Square",
				Confirmation: "End the selected chat sessions?",
				Handler: func(ctx context.Context, admin *admin.Admin[LiveChatSession], ids []interface{}) error {
					for _, id := range ids {
						session, err := LiveChatSessionObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						session.Status = "ended"
						if err := LiveChatSessionObjects.Update(ctx, session); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "transfer",
				Label:        "Transfer Chat",
				Icon:         "ArrowRight",
				Confirmation: "Transfer the selected chats to another agent?",
				Handler: func(ctx context.Context, admin *admin.Admin[LiveChatSession], ids []interface{}) error {
					for _, id := range ids {
						session, err := LiveChatSessionObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						session.Status = "transferred"
						if err := LiveChatSessionObjects.Update(ctx, session); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[LiveChatSession], user interface{}) bool {
					return true
				},
			},
		},
	})

	// FAQ admin
	admin.Register(&admin.Config[FAQ]{
		Icon: "HelpCircle",
		ListDisplay: []admin.Field{
			FAQFieldsInstance.Question,
			FAQFieldsInstance.Category,
			FAQFieldsInstance.IsFeatured,
			FAQFieldsInstance.ViewCount,
			FAQFieldsInstance.HelpfulCount,
			FAQFieldsInstance.IsVisible,
			FAQFieldsInstance.DisplayOrder,
		},
		ListFilter: []admin.Field{
			FAQFieldsInstance.Category,
			FAQFieldsInstance.IsVisible,
			FAQFieldsInstance.IsFeatured,
		},
		SearchFields: []admin.Field{
			FAQFieldsInstance.Question,
			FAQFieldsInstance.Answer,
			FAQFieldsInstance.Keywords,
		},
		Ordering: []admin.Field{
			FAQFieldsInstance.DisplayOrder,
			FAQFieldsInstance.Category,
		},
		Fieldsets: []admin.Fieldset[FAQ]{
			{
				Name: "Content",
				Fields: []string{"question", "answer", "category", "keywords"},
			},
			{
				Name: "Display",
				Fields: []string{"display_order", "is_visible", "is_featured"},
			},
			{
				Name: "Statistics",
				Fields: []string{"view_count", "helpful_count", "not_helpful_count"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"question", "answer", "is_visible", "is_featured"},
		},
		ReadOnlyFields: []admin.Field{
			FAQFieldsInstance.ViewCount,
			FAQFieldsInstance.HelpfulCount,
			FAQFieldsInstance.NotHelpfulCount,
			FAQFieldsInstance.CreatedAt,
			FAQFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}, obj *FAQ) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}, obj *FAQ) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}, obj *FAQ) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}, obj *FAQ) bool {
			return true
		},
		Filters: []admin.Filter[FAQ]{
			{
				Name:  "visible",
				Label: "Visible FAQs",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[FAQ], value interface{}) orm.QuerySet[FAQ] {
					return qs.Filter("is_visible", true)
				},
			},
			{
				Name:  "featured",
				Label: "Featured FAQs",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[FAQ], value interface{}) orm.QuerySet[FAQ] {
					return qs.Filter("is_featured", true)
				},
			},
			{
				Name:  "by_category",
				Label: "By Category",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "orders", Label: "Orders"},
					{Value: "shipping", Label: "Shipping"},
					{Value: "payments", Label: "Payments"},
					{Value: "returns", Label: "Returns"},
					{Value: "products", Label: "Products"},
					{Value: "account", Label: "Account"},
					{Value: "technical", Label: "Technical"},
					{Value: "general", Label: "General"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[FAQ], value interface{}) orm.QuerySet[FAQ] {
					return qs.Filter("category", value)
				},
			},
		},
		Actions: []admin.Action[FAQ]{
			{
				Name:         "toggle_visibility",
				Label:        "Toggle Visibility",
				Icon:         "Eye",
				Confirmation: "Toggle visibility for selected FAQs?",
				Handler: func(ctx context.Context, admin *admin.Admin[FAQ], ids []interface{}) error {
					for _, id := range ids {
						faq, err := FAQObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						faq.IsVisible = !faq.IsVisible
						if err := FAQObjects.Update(ctx, faq); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "toggle_featured",
				Label:        "Toggle Featured",
				Icon:         "Star",
				Confirmation: "Toggle featured status for selected FAQs?",
				Handler: func(ctx context.Context, admin *admin.Admin[FAQ], ids []interface{}) error {
					for _, id := range ids {
						faq, err := FAQObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						faq.IsFeatured = !faq.IsFeatured
						if err := FAQObjects.Update(ctx, faq); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[FAQ], user interface{}) bool {
					return true
				},
			},
		},
	})
}

