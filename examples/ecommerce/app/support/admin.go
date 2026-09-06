package support

import (
	"context"

	"github.com/forgego/forge/admin"
)

// RegisterAdmin registers all support models with the admin interface
func RegisterAdmin(ctx context.Context) error {
	// SupportTicket Admin
	_, err := admin.Register(&admin.Config[SupportTicket]{
		ListDisplay: []admin.Field{
			SupportTicketFieldsInstance.Id,
			SupportTicketFieldsInstance.TicketNumber,
			SupportTicketFieldsInstance.CustomerId,
			SupportTicketFieldsInstance.Subject,
			SupportTicketFieldsInstance.Status,
			SupportTicketFieldsInstance.Priority,
			SupportTicketFieldsInstance.Type,
			SupportTicketFieldsInstance.AssignedTo,
			SupportTicketFieldsInstance.CreatedAt,
		},
		SearchFields: []admin.Field{
			SupportTicketFieldsInstance.TicketNumber,
			SupportTicketFieldsInstance.Subject,
			SupportTicketFieldsInstance.Description,
		},
		ListFilter: []admin.Field{
			SupportTicketFieldsInstance.Status,
			SupportTicketFieldsInstance.Priority,
			SupportTicketFieldsInstance.Type,
			SupportTicketFieldsInstance.RelatedType,
		},
		Ordering: []admin.Field{
			SupportTicketFieldsInstance.CreatedAt,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// SupportMessage Admin
	_, err = admin.Register(&admin.Config[SupportMessage]{
		ListDisplay: []admin.Field{
			SupportMessageFieldsInstance.Id,
			SupportMessageFieldsInstance.TicketId,
			SupportMessageFieldsInstance.SenderType,
			SupportMessageFieldsInstance.CustomerId,
			SupportMessageFieldsInstance.IsInternalNote,
			SupportMessageFieldsInstance.CreatedAt,
		},
		SearchFields: []admin.Field{
			SupportMessageFieldsInstance.Content,
			SupportMessageFieldsInstance.CustomerId,
		},
		ListFilter: []admin.Field{
			SupportMessageFieldsInstance.SenderType,
			SupportMessageFieldsInstance.IsInternalNote,
		},
		Ordering: []admin.Field{
			SupportMessageFieldsInstance.CreatedAt,
		},
		ListPerPage: 100,
	})
	if err != nil {
		return err
	}

	// ReturnRequest Admin
	_, err = admin.Register(&admin.Config[ReturnRequest]{
		ListDisplay: []admin.Field{
			ReturnRequestFieldsInstance.Id,
			ReturnRequestFieldsInstance.ReturnNumber,
			ReturnRequestFieldsInstance.OrderId,
			ReturnRequestFieldsInstance.CustomerId,
			ReturnRequestFieldsInstance.Reason,
			ReturnRequestFieldsInstance.Status,
			ReturnRequestFieldsInstance.RefundAmount,
			ReturnRequestFieldsInstance.CreatedAt,
		},
		SearchFields: []admin.Field{
			ReturnRequestFieldsInstance.ReturnNumber,
			ReturnRequestFieldsInstance.OrderId,
			ReturnRequestFieldsInstance.CustomerId,
		},
		ListFilter: []admin.Field{
			ReturnRequestFieldsInstance.Status,
			ReturnRequestFieldsInstance.Reason,
			ReturnRequestFieldsInstance.RefundMethod,
		},
		Ordering: []admin.Field{
			ReturnRequestFieldsInstance.CreatedAt,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// LiveChatSession Admin
	_, err = admin.Register(&admin.Config[LiveChatSession]{
		ListDisplay: []admin.Field{
			LiveChatSessionFieldsInstance.Id,
			LiveChatSessionFieldsInstance.SessionId,
			LiveChatSessionFieldsInstance.CustomerId,
			LiveChatSessionFieldsInstance.AgentId,
			LiveChatSessionFieldsInstance.Status,
			LiveChatSessionFieldsInstance.Duration,
			LiveChatSessionFieldsInstance.Rating,
			LiveChatSessionFieldsInstance.StartedAt,
		},
		SearchFields: []admin.Field{
			LiveChatSessionFieldsInstance.SessionId,
			LiveChatSessionFieldsInstance.CustomerId,
			LiveChatSessionFieldsInstance.AgentId,
		},
		ListFilter: []admin.Field{
			LiveChatSessionFieldsInstance.Status,
			LiveChatSessionFieldsInstance.Rating,
		},
		Ordering: []admin.Field{
			LiveChatSessionFieldsInstance.StartedAt,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// FAQ Admin
	_, err = admin.Register(&admin.Config[FAQ]{
		ListDisplay: []admin.Field{
			FAQFieldsInstance.Id,
			FAQFieldsInstance.Question,
			FAQFieldsInstance.Category,
			FAQFieldsInstance.IsPublic,
			FAQFieldsInstance.ViewCount,
			FAQFieldsInstance.HelpfulYes,
			FAQFieldsInstance.HelpfulNo,
			FAQFieldsInstance.SortOrder,
		},
		SearchFields: []admin.Field{
			FAQFieldsInstance.Question,
			FAQFieldsInstance.Answer,
			FAQFieldsInstance.Tags,
		},
		ListFilter: []admin.Field{
			FAQFieldsInstance.IsPublic,
			FAQFieldsInstance.Category,
		},
		Ordering: []admin.Field{
			FAQFieldsInstance.SortOrder,
			FAQFieldsInstance.ViewCount,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	return nil
}
