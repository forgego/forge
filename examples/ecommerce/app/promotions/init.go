package promotions

import (
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/registry"
)

// Init initializes the promotions module
func Init(database *db.DB) {
	// Register models with schema registry
	registry.RegisterModel(Promotion{})
	registry.RegisterModel(PromotionRule{})
	registry.RegisterModel(Banner{})
	registry.RegisterModel(NewsletterSubscription{})
	registry.RegisterModel(PromotionUsage{})

	if PromotionObjects != nil {
		PromotionObjects.SetDB(database)
	}
	if PromotionRuleObjects != nil {
		PromotionRuleObjects.SetDB(database)
	}
	if BannerObjects != nil {
		BannerObjects.SetDB(database)
	}
	if NewsletterSubscriptionObjects != nil {
		NewsletterSubscriptionObjects.SetDB(database)
	}
	if PromotionUsageObjects != nil {
		PromotionUsageObjects.SetDB(database)
	}
}
