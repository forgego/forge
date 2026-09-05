package orders

import "github.com/forgego/forge/db"

// Init initializes the orders package with database connection
func Init(database *db.DB) {
	if CartObjects != nil {
		CartObjects.SetDB(database)
	}
	if CartItemObjects != nil {
		CartItemObjects.SetDB(database)
	}
	if OrderObjects != nil {
		OrderObjects.SetDB(database)
	}
	if OrderItemObjects != nil {
		OrderItemObjects.SetDB(database)
	}
	if PaymentObjects != nil {
		PaymentObjects.SetDB(database)
	}
	if ShipmentObjects != nil {
		ShipmentObjects.SetDB(database)
	}
}
