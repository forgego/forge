package marketing

import "github.com/forgego/forge/db"

// Init initializes the marketing package with database connection
func Init(database *db.DB) {
	if CouponObjects != nil {
		CouponObjects.SetDB(database)
	}
	if CouponUsageObjects != nil {
		CouponUsageObjects.SetDB(database)
	}
	if ReviewObjects != nil {
		ReviewObjects.SetDB(database)
	}
	if ReviewImageObjects != nil {
		ReviewImageObjects.SetDB(database)
	}
	if ReviewHelpfulnessObjects != nil {
		ReviewHelpfulnessObjects.SetDB(database)
	}
	if ProductQuestionObjects != nil {
		ProductQuestionObjects.SetDB(database)
	}
}
