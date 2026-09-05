package promotions

import (
	"context"

	"github.com/forgego/forge/admin"
	adminCore "github.com/forgego/forge/admin/core"
)

// RegisterAdmin registers all promotions models with the admin interface
func RegisterAdmin(ctx context.Context) {
	site := admin.DefaultSite

	// Promotion Admin
	promotionAdmin := adminCore.NewModelAdmin(
		Promotion{},
		adminCore.WithListDisplay("id", "name", "code", "discount_type", "discount_value", "is_active", "usage_count", "start_date", "end_date"),
		adminCore.WithSearchFields("name", "code", "description"),
		adminCore.WithListFilter("is_active", "discount_type", "is_stackable"),
		adminCore.WithOrdering("-priority", "-created_at"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, promotionAdmin)

	// PromotionRule Admin
	promotionRuleAdmin := adminCore.NewModelAdmin(
		PromotionRule{},
		adminCore.WithListDisplay("id", "promotion_id", "rule_type", "field", "operator", "value", "logic_type"),
		adminCore.WithSearchFields("field", "value"),
		adminCore.WithListFilter("rule_type", "operator", "logic_type"),
		adminCore.WithOrdering("promotion_id", "sort_order"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, promotionRuleAdmin)

	// Banner Admin
	bannerAdmin := adminCore.NewModelAdmin(
		Banner{},
		adminCore.WithListDisplay("id", "title", "placement", "is_active", "view_count", "click_count", "start_date", "end_date"),
		adminCore.WithSearchFields("title", "description"),
		adminCore.WithListFilter("is_active", "placement", "target_group"),
		adminCore.WithOrdering("sort_order", "-created_at"),
		adminCore.WithListPerPage(50),
	)
	site.RegisterModel(ctx, bannerAdmin)

	// NewsletterSubscription Admin
	newsletterSubscriptionAdmin := adminCore.NewModelAdmin(
		NewsletterSubscription{},
		adminCore.WithListDisplay("id", "email", "first_name", "last_name", "status", "source", "confirmed_at", "created_at"),
		adminCore.WithSearchFields("email", "first_name", "last_name"),
		adminCore.WithListFilter("status", "source"),
		adminCore.WithOrdering("-created_at"),
		adminCore.WithListPerPage(100),
	)
	site.RegisterModel(ctx, newsletterSubscriptionAdmin)

	// PromotionUsage Admin
	promotionUsageAdmin := adminCore.NewModelAdmin(
		PromotionUsage{},
		adminCore.WithListDisplay("id", "promotion_id", "customer_id", "order_id", "discount_amount", "used_at"),
		adminCore.WithSearchFields("promotion_id", "customer_id", "order_id"),
		adminCore.WithListFilter("used_at"),
		adminCore.WithOrdering("-used_at"),
		adminCore.WithListPerPage(100),
	)
	site.RegisterModel(ctx, promotionUsageAdmin)
}
