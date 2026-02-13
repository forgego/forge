package engagement

import "github.com/forgego/forge/db"

// Init initializes the engagement package with database connection
func Init(database *db.DB) {
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
}

