package engagement

import (
	"context"

	"github.com/forgego/forge/admin"
	adminCore "github.com/forgego/forge/admin/core"
)

// RegisterAdmin registers all engagement models with the admin interface
func RegisterAdmin(ctx context.Context) {
	site := admin.DefaultSite

	// RecentlyViewed Admin
	recentlyViewedAdmin := adminCore.NewModelAdmin(
		RecentlyViewed{},
		adminCore.WithListDisplay("id", "customer_id", "product_id", "viewed_at", "view_count"),
		adminCore.WithSearchFields("customer_id", "product_id", "session_id"),
		adminCore.WithListFilter("viewed_at"),
		adminCore.WithOrdering("-viewed_at"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, recentlyViewedAdmin)

	// ProductComparison Admin
	productComparisonAdmin := adminCore.NewModelAdmin(
		ProductComparison{},
		adminCore.WithListDisplay("id", "customer_id", "name", "is_public", "created_at"),
		adminCore.WithSearchFields("name", "customer_id"),
		adminCore.WithListFilter("is_public"),
		adminCore.WithOrdering("-created_at"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, productComparisonAdmin)

	// Notification Admin
	notificationAdmin := adminCore.NewModelAdmin(
		Notification{},
		adminCore.WithListDisplay("id", "customer_id", "title", "type", "priority", "is_read", "created_at"),
		adminCore.WithSearchFields("title", "message", "customer_id"),
		adminCore.WithListFilter("type", "priority", "is_read"),
		adminCore.WithOrdering("-created_at"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, notificationAdmin)

	// CustomerActivity Admin
	customerActivityAdmin := adminCore.NewModelAdmin(
		CustomerActivity{},
		adminCore.WithListDisplay("id", "customer_id", "activity_type", "entity_type", "entity_id", "created_at"),
		adminCore.WithSearchFields("customer_id", "activity_type", "description"),
		adminCore.WithListFilter("activity_type", "entity_type"),
		adminCore.WithOrdering("-created_at"),
		adminCore.WithListPerPage(100),
	)
	site.RegisterModel(ctx, customerActivityAdmin)

	// AbandonedCartReminder Admin
	abandonedCartReminderAdmin := adminCore.NewModelAdmin(
		AbandonedCartReminder{},
		adminCore.WithListDisplay("id", "cart_id", "customer_id", "reminder_type", "status", "converted", "sent_at"),
		adminCore.WithSearchFields("cart_id", "customer_id", "email_address"),
		adminCore.WithListFilter("status", "converted", "reminder_type"),
		adminCore.WithOrdering("-sent_at"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, abandonedCartReminderAdmin)

	// UserSegment Admin
	userSegmentAdmin := adminCore.NewModelAdmin(
		UserSegment{},
		adminCore.WithListDisplay("id", "name", "is_active", "is_dynamic", "member_count", "priority", "created_at"),
		adminCore.WithSearchFields("name", "description"),
		adminCore.WithListFilter("is_active", "is_dynamic"),
		adminCore.WithOrdering("-priority", "name"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, userSegmentAdmin)

	// SegmentRule Admin
	segmentRuleAdmin := adminCore.NewModelAdmin(
		SegmentRule{},
		adminCore.WithListDisplay("id", "segment_id", "field", "operator", "value", "logic_type", "sort_order"),
		adminCore.WithSearchFields("field", "value"),
		adminCore.WithListFilter("operator", "logic_type"),
		adminCore.WithOrdering("segment_id", "sort_order"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, segmentRuleAdmin)
}
