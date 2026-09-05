package promotions

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers all promotions API endpoints
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// Promotion ViewSet
	promotionViewSet := api.NewModelViewSet(
		Promotion{},
		database,
		api.WithFilterFields("is_active", "discount_type", "is_stackable", "applies_to"),
		api.WithSearchFields("name", "code", "description"),
		api.WithOrderingFields("-priority", "-created_at"),
	)
	router.RegisterViewSet("promotions", promotionViewSet)

	// PromotionRule ViewSet
	promotionRuleViewSet := api.NewModelViewSet(
		PromotionRule{},
		database,
		api.WithFilterFields("promotion_id", "rule_type", "operator"),
		api.WithSearchFields("field", "value"),
		api.WithOrderingFields("promotion_id", "sort_order"),
	)
	router.RegisterViewSet("promotion-rules", promotionRuleViewSet)

	// Banner ViewSet
	bannerViewSet := api.NewModelViewSet(
		Banner{},
		database,
		api.WithFilterFields("is_active", "placement", "target_group"),
		api.WithSearchFields("title", "description"),
		api.WithOrderingFields("sort_order", "-created_at"),
	)
	router.RegisterViewSet("banners", bannerViewSet)

	// NewsletterSubscription ViewSet
	newsletterSubscriptionViewSet := api.NewModelViewSet(
		NewsletterSubscription{},
		database,
		api.WithFilterFields("status", "source"),
		api.WithSearchFields("email", "first_name", "last_name"),
		api.WithOrderingFields("-created_at"),
	)
	router.RegisterViewSet("newsletter-subscriptions", newsletterSubscriptionViewSet)

	// PromotionUsage ViewSet (read-only for analytics)
	promotionUsageViewSet := api.NewModelViewSet(
		PromotionUsage{},
		database,
		api.WithFilterFields("promotion_id", "customer_id", "order_id"),
		api.WithSearchFields("promotion_id", "customer_id"),
		api.WithOrderingFields("-used_at"),
	)
	router.RegisterViewSet("promotion-usages", promotionUsageViewSet)
}
