package commerce

import (
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/registry"
)

// Manager instances
var (
	ShippingMethodManager *ShippingMethodManagerImpl
	PaymentMethodManager  *PaymentMethodManagerImpl
	TaxRateManager        *TaxRateManagerImpl
	CurrencyManager       *CurrencyManagerImpl
	ExchangeRateManager   *ExchangeRateManagerImpl
)

// Init initializes the commerce module
func Init(database *db.DB) {
	// Register models with schema registry
	registry.RegisterModel(ShippingMethod{})
	registry.RegisterModel(PaymentMethod{})
	registry.RegisterModel(TaxRate{})
	registry.RegisterModel(Currency{})
	registry.RegisterModel(ExchangeRate{})

	// Initialize managers (these will be generated)
	// ShippingMethodManager = NewShippingMethodManager(database)
	// PaymentMethodManager = NewPaymentMethodManager(database)
	// TaxRateManager = NewTaxRateManager(database)
	// CurrencyManager = NewCurrencyManager(database)
	// ExchangeRateManager = NewExchangeRateManager(database)
}
