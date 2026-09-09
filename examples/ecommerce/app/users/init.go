package users

import "github.com/forgego/forge/db"

// Init initializes the users package with database connection
func Init(database *db.DB) {
	if UserObjects != nil {
		UserObjects.SetDB(database)
	}
	if GroupObjects != nil {
		GroupObjects.SetDB(database)
	}
}
