package engagement

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers engagement API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("recently-viewed", &api.ViewSetConfig{
		Model:      &RecentlyViewed{},
		Queryset:   RecentlyViewedObjects,
		Serializer: base,
	})

	router.Register("product-comparisons", &api.ViewSetConfig{
		Model:      &ProductComparison{},
		Queryset:   ProductComparisonObjects,
		Serializer: base,
	})

	router.Register("notifications", &api.ViewSetConfig{
		Model:      &Notification{},
		Queryset:   NotificationObjects,
		Serializer: base,
	})

	router.Register("customer-activities", &api.ViewSetConfig{
		Model:      &CustomerActivity{},
		Queryset:   CustomerActivityObjects,
		Serializer: base,
	})

	router.Register("abandoned-cart-reminders", &api.ViewSetConfig{
		Model:      &AbandonedCartReminder{},
		Queryset:   AbandonedCartReminderObjects,
		Serializer: base,
	})

	router.Register("user-segments", &api.ViewSetConfig{
		Model:      &UserSegment{},
		Queryset:   UserSegmentObjects,
		Serializer: base,
	})

	router.Register("segment-rules", &api.ViewSetConfig{
		Model:      &SegmentRule{},
		Queryset:   SegmentRuleObjects,
		Serializer: base,
	})
}
