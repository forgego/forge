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
}
