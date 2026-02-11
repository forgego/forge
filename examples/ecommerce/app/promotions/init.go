package promotions

import "github.com/forgego/forge/db"

// Init initializes the promotions package with database connection
func Init(database *db.DB) {
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

import "github.com/forgego/forge/db"

// Init initializes the promotions package with database connection
func Init(database *db.DB) {
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

