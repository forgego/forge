package inventory

import "github.com/forgego/forge/db"

// Init initializes the inventory package with database connection
func Init(database *db.DB) {
	if WarehouseObjects != nil {
		WarehouseObjects.SetDB(database)
	}
	if StockObjects != nil {
		StockObjects.SetDB(database)
	}
	if StockMovementObjects != nil {
		StockMovementObjects.SetDB(database)
	}
	if StockAlertObjects != nil {
		StockAlertObjects.SetDB(database)
	}
	if StockTransferObjects != nil {
		StockTransferObjects.SetDB(database)
	}
}
