package commerce

import (
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/registry"
)

// Manager instances
var (
	ShippingMethodManager *orm.Manager[ShippingMethod]
	PaymentMethodManager  *orm.Manager[PaymentMethod]
	TaxRateManager        *orm.Manager[TaxRate]
	CurrencyManager       *orm.Manager[Currency]
	ExchangeRateManager   *orm.Manager[ExchangeRate]
)

// Init initializes the commerce module
func Init(database *db.DB) {
	// Register models with schema registry
	registry.RegisterModel(ShippingMethod{})
	registry.RegisterModel(PaymentMethod{})
	registry.RegisterModel(TaxRate{})
	registry.RegisterModel(Currency{})
	registry.RegisterModel(ExchangeRate{})

	// Initialize managers
	var err error
	ShippingMethodManager, err = orm.NewManager[ShippingMethod]("shipping_methods")
	if err != nil {
		panic(err)
	}
	ShippingMethodManager.SetDB(database)

	PaymentMethodManager, err = orm.NewManager[PaymentMethod]("payment_methods")
	if err != nil {
		panic(err)
	}
	PaymentMethodManager.SetDB(database)

	TaxRateManager, err = orm.NewManager[TaxRate]("tax_rates")
	if err != nil {
		panic(err)
	}
	TaxRateManager.SetDB(database)

	CurrencyManager, err = orm.NewManager[Currency]("currencies")
	if err != nil {
		panic(err)
	}
	CurrencyManager.SetDB(database)

	ExchangeRateManager, err = orm.NewManager[ExchangeRate]("exchange_rates")
	if err != nil {
		panic(err)
	}
	ExchangeRateManager.SetDB(database)
}
