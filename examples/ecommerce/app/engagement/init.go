package engagement

import (
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/registry"
)

// Init initializes the engagement module
func Init(database *db.DB) {
	// Register models with schema registry
	registry.RegisterModel(RecentlyViewed{})
	registry.RegisterModel(ProductComparison{})
	registry.RegisterModel(Notification{})
	registry.RegisterModel(CustomerActivity{})
	registry.RegisterModel(AbandonedCartReminder{})
	registry.RegisterModel(UserSegment{})
	registry.RegisterModel(SegmentRule{})
}
