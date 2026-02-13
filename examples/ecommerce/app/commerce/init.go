package commerce

import "github.com/forgego/forge/db"

// Init initializes the commerce package with database connection
func Init(database *db.DB) {
	if ShippingMethodObjects != nil {
		ShippingMethodObjects.SetDB(database)
	}
	if PaymentMethodObjects != nil {
		PaymentMethodObjects.SetDB(database)
	}
	if TaxRateObjects != nil {
		TaxRateObjects.SetDB(database)
	}
	if CurrencyObjects != nil {
		CurrencyObjects.SetDB(database)
	}
	if ExchangeRateObjects != nil {
		ExchangeRateObjects.SetDB(database)
	}
}

