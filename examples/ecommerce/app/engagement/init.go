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

	if RecentlyViewedObjects != nil {
		RecentlyViewedObjects.SetDB(database)
	}
	if ProductComparisonObjects != nil {
		ProductComparisonObjects.SetDB(database)
	}
	if NotificationObjects != nil {
		NotificationObjects.SetDB(database)
	}
	if CustomerActivityObjects != nil {
		CustomerActivityObjects.SetDB(database)
	}
	if AbandonedCartReminderObjects != nil {
		AbandonedCartReminderObjects.SetDB(database)
	}
	if UserSegmentObjects != nil {
		UserSegmentObjects.SetDB(database)
	}
	if SegmentRuleObjects != nil {
		SegmentRuleObjects.SetDB(database)
	}
}
