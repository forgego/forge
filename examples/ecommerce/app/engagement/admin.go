package engagement

import (
	"context"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers all engagement models with the admin interface
func RegisterAdmin(ctx context.Context) error {
	// RecentlyViewed Admin
	_, err := admin.Register(&admin.Config[RecentlyViewed]{
		ListDisplay: []admin.Field{
			RecentlyViewedFieldsInstance.ID,
			RecentlyViewedFieldsInstance.CustomerID,
			RecentlyViewedFieldsInstance.ProductID,
			RecentlyViewedFieldsInstance.ViewedAt,
			RecentlyViewedFieldsInstance.ViewCount,
		},
		SearchFields: []admin.Field{
			RecentlyViewedFieldsInstance.CustomerID,
			RecentlyViewedFieldsInstance.ProductID,
			RecentlyViewedFieldsInstance.SessionID,
		},
		ListFilter: []admin.Field{
			RecentlyViewedFieldsInstance.ViewedAt,
		},
		Ordering: []admin.Field{
			RecentlyViewedFieldsInstance.ViewedAt,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// ProductComparison Admin
	_, err = admin.Register(&admin.Config[ProductComparison]{
		ListDisplay: []admin.Field{
			ProductComparisonFieldsInstance.ID,
			ProductComparisonFieldsInstance.CustomerID,
			ProductComparisonFieldsInstance.Name,
			ProductComparisonFieldsInstance.IsPublic,
			ProductComparisonFieldsInstance.CreatedAt,
		},
		SearchFields: []admin.Field{
			ProductComparisonFieldsInstance.Name,
			ProductComparisonFieldsInstance.CustomerID,
		},
		ListFilter: []admin.Field{
			ProductComparisonFieldsInstance.IsPublic,
		},
		Ordering: []admin.Field{
			ProductComparisonFieldsInstance.CreatedAt,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// Notification Admin
	_, err = admin.Register(&admin.Config[Notification]{
		ListDisplay: []admin.Field{
			NotificationFieldsInstance.ID,
			NotificationFieldsInstance.CustomerID,
			NotificationFieldsInstance.Title,
			NotificationFieldsInstance.Type,
			NotificationFieldsInstance.Priority,
			NotificationFieldsInstance.IsRead,
			NotificationFieldsInstance.CreatedAt,
		},
		SearchFields: []admin.Field{
			NotificationFieldsInstance.Title,
			NotificationFieldsInstance.Message,
			NotificationFieldsInstance.CustomerID,
		},
		ListFilter: []admin.Field{
			NotificationFieldsInstance.Type,
			NotificationFieldsInstance.Priority,
			NotificationFieldsInstance.IsRead,
		},
		Ordering: []admin.Field{
			NotificationFieldsInstance.CreatedAt,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// CustomerActivity Admin
	_, err = admin.Register(&admin.Config[CustomerActivity]{
		ListDisplay: []admin.Field{
			CustomerActivityFieldsInstance.ID,
			CustomerActivityFieldsInstance.CustomerID,
			CustomerActivityFieldsInstance.ActivityType,
			CustomerActivityFieldsInstance.EntityType,
			CustomerActivityFieldsInstance.EntityID,
			CustomerActivityFieldsInstance.CreatedAt,
		},
		SearchFields: []admin.Field{
			CustomerActivityFieldsInstance.CustomerID,
			CustomerActivityFieldsInstance.ActivityType,
			CustomerActivityFieldsInstance.Description,
		},
		ListFilter: []admin.Field{
			CustomerActivityFieldsInstance.ActivityType,
			CustomerActivityFieldsInstance.EntityType,
		},
		Ordering: []admin.Field{
			CustomerActivityFieldsInstance.CreatedAt,
		},
		ListPerPage: 100,
	})
	if err != nil {
		return err
	}

	// AbandonedCartReminder Admin
	_, err = admin.Register(&admin.Config[AbandonedCartReminder]{
		ListDisplay: []admin.Field{
			AbandonedCartReminderFieldsInstance.ID,
			AbandonedCartReminderFieldsInstance.CartID,
			AbandonedCartReminderFieldsInstance.CustomerID,
			AbandonedCartReminderFieldsInstance.ReminderType,
			AbandonedCartReminderFieldsInstance.Status,
			AbandonedCartReminderFieldsInstance.Converted,
			AbandonedCartReminderFieldsInstance.SentAt,
		},
		SearchFields: []admin.Field{
			AbandonedCartReminderFieldsInstance.CartID,
			AbandonedCartReminderFieldsInstance.CustomerID,
			AbandonedCartReminderFieldsInstance.EmailAddress,
		},
		ListFilter: []admin.Field{
			AbandonedCartReminderFieldsInstance.Status,
			AbandonedCartReminderFieldsInstance.Converted,
			AbandonedCartReminderFieldsInstance.ReminderType,
		},
		Ordering: []admin.Field{
			AbandonedCartReminderFieldsInstance.SentAt,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// UserSegment Admin
	_, err = admin.Register(&admin.Config[UserSegment]{
		ListDisplay: []admin.Field{
			UserSegmentFieldsInstance.ID,
			UserSegmentFieldsInstance.Name,
			UserSegmentFieldsInstance.IsActive,
			UserSegmentFieldsInstance.IsDynamic,
			UserSegmentFieldsInstance.MemberCount,
			UserSegmentFieldsInstance.Priority,
			UserSegmentFieldsInstance.CreatedAt,
		},
		SearchFields: []admin.Field{
			UserSegmentFieldsInstance.Name,
			UserSegmentFieldsInstance.Description,
		},
		ListFilter: []admin.Field{
			UserSegmentFieldsInstance.IsActive,
			UserSegmentFieldsInstance.IsDynamic,
		},
		Ordering: []admin.Field{
			UserSegmentFieldsInstance.Priority,
			UserSegmentFieldsInstance.Name,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// SegmentRule Admin
	_, err = admin.Register(&admin.Config[SegmentRule]{
		ListDisplay: []admin.Field{
			SegmentRuleFieldsInstance.ID,
			SegmentRuleFieldsInstance.SegmentID,
			SegmentRuleFieldsInstance.Field,
			SegmentRuleFieldsInstance.Operator,
			SegmentRuleFieldsInstance.Value,
			SegmentRuleFieldsInstance.LogicType,
			SegmentRuleFieldsInstance.SortOrder,
		},
		SearchFields: []admin.Field{
			SegmentRuleFieldsInstance.Field,
			SegmentRuleFieldsInstance.Value,
		},
		ListFilter: []admin.Field{
			SegmentRuleFieldsInstance.Operator,
			SegmentRuleFieldsInstance.LogicType,
		},
		Ordering: []admin.Field{
			SegmentRuleFieldsInstance.SegmentID,
			SegmentRuleFieldsInstance.SortOrder,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	return nil
}
