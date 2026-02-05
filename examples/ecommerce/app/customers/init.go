package customers

import "github.com/forgego/forge/db"

// Init initializes the customers package with database connection
func Init(database *db.DB) {
	if CustomerGroupObjects != nil {
		CustomerGroupObjects.SetDB(database)
	}
	if CustomerObjects != nil {
		CustomerObjects.SetDB(database)
	}
	if AddressObjects != nil {
		AddressObjects.SetDB(database)
	}
	if WishListObjects != nil {
		WishListObjects.SetDB(database)
	}
	if WishListItemObjects != nil {
		WishListItemObjects.SetDB(database)
	}
}
