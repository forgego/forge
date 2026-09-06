package promotions

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers promotions API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("promotions", &api.ViewSetConfig{
		Model:      &Promotion{},
		Queryset:   PromotionObjects,
		Serializer: base,
	})

	router.Register("promotion-rules", &api.ViewSetConfig{
		Model:      &PromotionRule{},
		Queryset:   PromotionRuleObjects,
		Serializer: base,
	})

	router.Register("banners", &api.ViewSetConfig{
		Model:      &Banner{},
		Queryset:   BannerObjects,
		Serializer: base,
	})

	router.Register("newsletter-subscriptions", &api.ViewSetConfig{
		Model:      &NewsletterSubscription{},
		Queryset:   NewsletterSubscriptionObjects,
		Serializer: base,
	})

	router.Register("promotion-usages", &api.ViewSetConfig{
		Model:      &PromotionUsage{},
		Queryset:   PromotionUsageObjects,
		Serializer: base,
	})
}
