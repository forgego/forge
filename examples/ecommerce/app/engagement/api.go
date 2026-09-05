package engagement

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers all engagement API endpoints
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// RecentlyViewed ViewSet
	recentlyViewedViewSet := api.NewModelViewSet(
		RecentlyViewed{},
		database,
		api.WithFilterFields("customer_id", "product_id", "viewed_at"),
		api.WithSearchFields("session_id"),
		api.WithOrderingFields("-viewed_at", "view_count"),
	)
	router.RegisterViewSet("recently-viewed", recentlyViewedViewSet)

	// ProductComparison ViewSet
	productComparisonViewSet := api.NewModelViewSet(
		ProductComparison{},
		database,
		api.WithFilterFields("customer_id", "is_public"),
		api.WithSearchFields("name"),
		api.WithOrderingFields("-created_at"),
	)
	router.RegisterViewSet("product-comparisons", productComparisonViewSet)

	// Notification ViewSet
	notificationViewSet := api.NewModelViewSet(
		Notification{},
		database,
		api.WithFilterFields("customer_id", "type", "priority", "is_read"),
		api.WithSearchFields("title", "message"),
		api.WithOrderingFields("-created_at"),
	)
	router.RegisterViewSet("notifications", notificationViewSet)

	// CustomerActivity ViewSet (read-only for analytics)
	customerActivityViewSet := api.NewModelViewSet(
		CustomerActivity{},
		database,
		api.WithFilterFields("customer_id", "activity_type", "entity_type"),
		api.WithSearchFields("description"),
		api.WithOrderingFields("-created_at"),
	)
	router.RegisterViewSet("customer-activities", customerActivityViewSet)

	// AbandonedCartReminder ViewSet
	abandonedCartReminderViewSet := api.NewModelViewSet(
		AbandonedCartReminder{},
		database,
		api.WithFilterFields("customer_id", "cart_id", "status", "converted"),
		api.WithSearchFields("email_address"),
		api.WithOrderingFields("-sent_at"),
	)
	router.RegisterViewSet("abandoned-cart-reminders", abandonedCartReminderViewSet)

	// UserSegment ViewSet
	userSegmentViewSet := api.NewModelViewSet(
		UserSegment{},
		database,
		api.WithFilterFields("is_active", "is_dynamic"),
		api.WithSearchFields("name", "description"),
		api.WithOrderingFields("-priority", "name"),
	)
	router.RegisterViewSet("user-segments", userSegmentViewSet)

	// SegmentRule ViewSet
	segmentRuleViewSet := api.NewModelViewSet(
		SegmentRule{},
		database,
		api.WithFilterFields("segment_id", "operator", "logic_type"),
		api.WithSearchFields("field", "value"),
		api.WithOrderingFields("segment_id", "sort_order"),
	)
	router.RegisterViewSet("segment-rules", segmentRuleViewSet)
}
