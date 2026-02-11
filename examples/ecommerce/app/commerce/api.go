package commerce

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers commerce API endpoints with enhanced features
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// ShippingMethod API with enhanced features
	shippingMethodViewSet := api.NewViewSet(&ShippingMethod{})
	shippingMethodViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			commerceUserLookup,
		),
	}
	shippingMethodViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	shippingMethodViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for ShippingMethod
	shippingMethodViewSet.RegisterAction("calculate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/calculate",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleShippingCalculate,
	})

	shippingMethodViewSet.RegisterAction("available", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/available",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleShippingAvailable,
	})

	shippingMethodViewSet.RegisterAction("zones", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/zones",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleShippingZones,
	})

	router.Register("shipping-methods", &api.ViewSetConfig{
		Model:           &ShippingMethod{},
		Queryset:        ShippingMethodObjects,
		Serializer:      &ShippingMethodSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "name", "code", "carrier", "service_level", "base_price", "is_active", "is_default", "min_days", "max_days"},
		DetailFields:    []string{"id", "name", "code", "description", "carrier", "service_level", "base_price", "handling_fee", "free_shipping_threshold", "min_weight", "max_weight", "weight_unit", "min_length", "max_length", "min_width", "max_width", "min_height", "max_height", "dimension_unit", "min_days", "max_days", "countries", "excluded_countries", "is_active", "is_default", "priority", "config", "created_at", "updated_at"},
		Filterable:      []string{"is_active", "is_default", "carrier", "service_level"},
		Searchable:      []string{"name", "code", "carrier", "description"},
		Ordering:        []string{"priority", "name", "base_price"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: shippingMethodViewSet,
	})

	// PaymentMethod API with enhanced features
	paymentMethodViewSet := api.NewViewSet(&PaymentMethod{})
	paymentMethodViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			commerceUserLookup,
		),
	}
	paymentMethodViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	paymentMethodViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for PaymentMethod
	paymentMethodViewSet.RegisterAction("configure", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      true,
		URLPath:     "/configure",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handlePaymentConfigure,
	})

	paymentMethodViewSet.RegisterAction("test-connection", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/test-connection",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handlePaymentTestConnection,
	})

	router.Register("payment-methods", &api.ViewSetConfig{
		Model:           &PaymentMethod{},
		Queryset:        PaymentMethodObjects,
		Serializer:      &PaymentMethodSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "name", "code", "type", "gateway", "fixed_fee", "percentage_fee", "is_active", "is_default"},
		DetailFields:    []string{"id", "name", "code", "type", "description", "gateway", "gateway_config", "fixed_fee", "percentage_fee", "min_amount", "max_amount", "currencies", "icon", "display_order", "is_active", "is_default", "test_mode", "created_at", "updated_at"},
		Filterable:      []string{"is_active", "is_default", "type", "gateway"},
		Searchable:      []string{"name", "code", "description"},
		Ordering:        []string{"display_order", "name"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: paymentMethodViewSet,
	})

	// TaxRate API with enhanced features
	taxRateViewSet := api.NewViewSet(&TaxRate{})
	taxRateViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			commerceUserLookup,
		),
	}
	taxRateViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	taxRateViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for TaxRate
	taxRateViewSet.RegisterAction("calculate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/calculate",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleTaxCalculate,
	})

	taxRateViewSet.RegisterAction("by-location", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/by-location",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleTaxByLocation,
	})

	router.Register("tax-rates", &api.ViewSetConfig{
		Model:           &TaxRate{},
		Queryset:        TaxRateObjects,
		Serializer:      &TaxRateSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "name", "code", "rate", "country", "state", "tax_type", "is_active", "priority"},
		DetailFields:    []string{"id", "name", "code", "description", "rate", "country", "state", "zip_pattern", "city", "tax_type", "applies_to_products", "applies_to_shipping", "applies_to_services", "start_date", "end_date", "is_active", "is_compound", "priority", "created_at", "updated_at"},
		Filterable:      []string{"is_active", "country", "state", "tax_type", "is_compound"},
		Searchable:      []string{"name", "code", "description"},
		Ordering:        []string{"country", "state", "priority"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: taxRateViewSet,
	})

	// Currency API with enhanced features
	currencyViewSet := api.NewViewSet(&Currency{})
	currencyViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			commerceUserLookup,
		),
	}
	currencyViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	currencyViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for Currency
	currencyViewSet.RegisterAction("convert", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/convert",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleCurrencyConvert,
	})

	currencyViewSet.RegisterAction("latest-rates", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/latest-rates",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleLatestRates,
	})

	router.Register("currencies", &api.ViewSetConfig{
		Model:           &Currency{},
		Queryset:        CurrencyObjects,
		Serializer:      &CurrencySerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "code", "name", "symbol", "decimal_places", "is_base_currency", "exchange_rate", "is_active", "is_default"},
		DetailFields:    []string{"id", "code", "name", "symbol", "decimal_places", "decimal_separator", "thousand_separator", "symbol_position", "is_base_currency", "exchange_rate", "is_active", "is_default", "created_at", "updated_at"},
		Filterable:      []string{"is_active", "is_default", "is_base_currency"},
		Searchable:      []string{"code", "name", "symbol"},
		Ordering:        []string{"code"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: currencyViewSet,
	})

	// ExchangeRate API with enhanced features
	exchangeRateViewSet := api.NewViewSet(&ExchangeRate{})
	exchangeRateViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			commerceUserLookup,
		),
	}
	exchangeRateViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	exchangeRateViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	router.Register("exchange-rates", &api.ViewSetConfig{
		Model:           &ExchangeRate{},
		Queryset:        ExchangeRateObjects,
		Serializer:      &ExchangeRateSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "from_currency", "to_currency", "rate", "effective_from", "effective_to", "provider"},
		DetailFields:    []string{"id", "from_currency", "to_currency", "rate", "effective_from", "effective_to", "provider", "created_at", "updated_at"},
		Filterable:      []string{"from_currency", "to_currency"},
		Searchable:      []string{"from_currency", "to_currency", "provider"},
		Ordering:        []string{"from_currency", "to_currency", "-effective_from"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: exchangeRateViewSet,
	})
}

// commerceUserLookup looks up a user from JWT claims
func commerceUserLookup(claims authentication.JWTClaims) (interface{}, error) {
	// Extract user ID from subject claim
	userID, ok := claims["subject"].(string)
	if !ok {
		if subject, ok := claims["subject"].(float64); ok {
			userID = strconv.FormatFloat(subject, 'f', -1, 64)
		} else {
			return nil, nil
		}
	}
	// In a real implementation, you would query the database
	return nil, nil
}

// Custom action handlers for ShippingMethod

func handleShippingCalculate(w http.ResponseWriter, r *http.Request) {
	// Extract request body
	var req struct {
		ShippingMethodID string  `json:"shipping_method_id"`
		Weight          float64 `json:"weight"`
		Subtotal        float64 `json:"subtotal"`
		Country         string  `json:"country"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate shipping cost (placeholder implementation)
	response := map[string]interface{}{
		"shipping_method_id": req.ShippingMethodID,
		"base_price":        10.0,
		"handling_fee":      2.0,
		"weight_fee":        req.Weight * 0.5,
		"subtotal":          req.Subtotal,
		"free_shipping":     req.Subtotal >= 100.0,
		"total":             12.0 + req.Weight*0.5,
	}

	if req.Subtotal >= 100.0 {
		response["total"] = 0.0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleShippingAvailable(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	weight, _ := strconv.ParseFloat(r.URL.Query().Get("weight"), 64)

	// Return available shipping methods (placeholder)
	response := map[string]interface{}{
		"country": country,
		"weight": weight,
		"methods": []interface{}{
			map[string]interface{}{
				"id":           "1",
				"name":         "Standard Shipping",
				"carrier":      "Generic Carrier",
				"base_price":   10.0,
				"min_days":     5,
				"max_days":     7,
			},
			map[string]interface{}{
				"id":           "2",
				"name":         "Express Shipping",
				"carrier":      "Generic Carrier",
				"base_price":   25.0,
				"min_days":     1,
				"max_days":     3,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleShippingZones(w http.ResponseWriter, r *http.Request) {
	// Return shipping zones (placeholder)
	response := map[string]interface{}{
		"zones": []interface{}{
			map[string]interface{}{
				"id":         "1",
				"name":       "Domestic",
				"countries":  []string{"US", "CA"},
				"methods":    []string{"standard", "express"},
			},
			map[string]interface{}{
				"id":         "2",
				"name":       "International",
				"countries":  []string{"UK", "DE", "FR"},
				"methods":    []string{"international"},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for PaymentMethod

func handlePaymentConfigure(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	paymentMethodID := vars["id"]

	// Return gateway configuration (placeholder)
	response := map[string]interface{}{
		"payment_method_id": paymentMethodID,
		"gateway": map[string]interface{}{
			"name":      "stripe",
			"api_key":   "sk_test_****",
			"mode":      "test",
			"webhook":   "https://example.com/webhook",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePaymentTestConnection(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	paymentMethodID := vars["id"]

	// Test gateway connection (placeholder)
	response := map[string]interface{}{
		"payment_method_id": paymentMethodID,
		"status":            "success",
		"message":           "Gateway connection successful",
		"latency":           "125ms",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for TaxRate

func handleTaxCalculate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Country  string  `json:"country"`
		State    string  `json:"state"`
		Subtotal float64 `json:"subtotal"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate tax (placeholder)
	taxRate := 0.08 // 8% default
	if req.Country == "US" && req.State == "CA" {
		taxRate = 0.0725 // California sales tax
	} else if req.Country == "UK" {
		taxRate = 0.20 // UK VAT
	}

	taxAmount := req.Subtotal * taxRate

	response := map[string]interface{}{
		"country":     req.Country,
		"state":       req.State,
		"subtotal":    req.Subtotal,
		"tax_rate":    taxRate,
		"tax_amount":  taxAmount,
		"total":       req.Subtotal + taxAmount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTaxByLocation(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	state := r.URL.Query().Get("state")
	zipCode := r.URL.Query().Get("zip")

	// Return tax rates for location (placeholder)
	response := map[string]interface{}{
		"country":  country,
		"state":    state,
		"zip":      zipCode,
		"tax_rates": []interface{}{
			map[string]interface{}{
				"id":    "1",
				"name":  "Standard Sales Tax",
				"rate":  0.08,
				"type":  "sales",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for Currency

func handleCurrencyConvert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromCurrency string  `json:"from_currency"`
		ToCurrency   string  `json:"to_currency"`
		Amount       float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert currency (placeholder - would use actual exchange rates)
	rate := 1.0
	if req.FromCurrency != req.ToCurrency {
		// Simple conversion rates for demo
		rates := map[string]float64{
			"USD": 1.0,
			"EUR": 0.92,
			"GBP": 0.79,
			"JPY": 149.50,
		}

		fromRate := rates[req.FromCurrency]
		toRate := rates[req.ToCurrency]

		if fromRate > 0 && toRate > 0 {
			rate = toRate / fromRate
		}
	}

	convertedAmount := req.Amount * rate

	response := map[string]interface{}{
		"from_currency":   req.FromCurrency,
		"to_currency":     req.ToCurrency,
		"original_amount": req.Amount,
		"rate":           rate,
		"converted_amount": convertedAmount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleLatestRates(w http.ResponseWriter, r *http.Request) {
	baseCurrency := r.URL.Query().Get("base")
	if baseCurrency == "" {
		baseCurrency = "USD"
	}

	// Return latest exchange rates (placeholder)
	response := map[string]interface{}{
		"base_currency": baseCurrency,
		"date":          time.Now().Format("2006-01-02"),
		"rates": map[string]interface{}{
			"USD": 1.0,
			"EUR": 0.92,
			"GBP": 0.79,
			"JPY": 149.50,
			"CAD": 1.36,
			"AUD": 1.53,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Serializers with validation

type ShippingMethodSerializer struct {
	*api.BaseSerializer
}

func (s *ShippingMethodSerializer) New() api.Serializer {
	return &ShippingMethodSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *ShippingMethodSerializer) Validate() error {
	errors := make(map[string]string)

	if s.Data["name"] == "" {
		errors["name"] = "Name is required"
	}
	if s.Data["code"] == "" {
		errors["code"] = "Code is required"
	}
	if s.Data["carrier"] == "" {
		errors["carrier"] = "Carrier is required"
	}

	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type PaymentMethodSerializer struct {
	*api.BaseSerializer
}

func (s *PaymentMethodSerializer) New() api.Serializer {
	return &PaymentMethodSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *PaymentMethodSerializer) Validate() error {
	errors := make(map[string]string)

	if s.Data["name"] == "" {
		errors["name"] = "Name is required"
	}
	if s.Data["code"] == "" {
		errors["code"] = "Code is required"
	}
	if s.Data["type"] == "" {
		errors["type"] = "Type is required"
	}

	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type TaxRateSerializer struct {
	*api.BaseSerializer
}

func (s *TaxRateSerializer) New() api.Serializer {
	return &TaxRateSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *TaxRateSerializer) Validate() error {
	errors := make(map[string]string)

	if s.Data["name"] == "" {
		errors["name"] = "Name is required"
	}
	if s.Data["code"] == "" {
		errors["code"] = "Code is required"
	}
	if s.Data["country"] == "" {
		errors["country"] = "Country is required"
	}

	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type CurrencySerializer struct {
	*api.BaseSerializer
}

func (s *CurrencySerializer) New() api.Serializer {
	return &CurrencySerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *CurrencySerializer) Validate() error {
	errors := make(map[string]string)

	if s.Data["code"] == "" {
		errors["code"] = "Code is required"
	}
	if s.Data["name"] == "" {
		errors["name"] = "Name is required"
	}
	if s.Data["symbol"] == "" {
		errors["symbol"] = "Symbol is required"
	}

	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type ExchangeRateSerializer struct {
	*api.BaseSerializer
}

func (s *ExchangeRateSerializer) New() api.Serializer {
	return &ExchangeRateSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *ExchangeRateSerializer) Validate() error {
	errors := make(map[string]string)

	if s.Data["from_currency"] == "" {
		errors["from_currency"] = "From currency is required"
	}
	if s.Data["to_currency"] == "" {
		errors["to_currency"] = "To currency is required"
	}
	if s.Data["rate"] == nil || s.Data["rate"].(float64) <= 0 {
		errors["rate"] = "Rate must be greater than 0"
	}

	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers commerce API endpoints with enhanced features
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// ShippingMethod API with enhanced features
	shippingMethodViewSet := api.NewViewSet(&ShippingMethod{})
	shippingMethodViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			commerceUserLookup,
		),
	}
	shippingMethodViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	shippingMethodViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for ShippingMethod
	shippingMethodViewSet.RegisterAction("calculate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/calculate",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleShippingCalculate,
	})

	shippingMethodViewSet.RegisterAction("available", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/available",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleShippingAvailable,
	})

	shippingMethodViewSet.RegisterAction("zones", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/zones",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleShippingZones,
	})

	router.Register("shipping-methods", &api.ViewSetConfig{
		Model:           &ShippingMethod{},
		Queryset:        ShippingMethodObjects,
		Serializer:      &ShippingMethodSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "name", "code", "carrier", "service_level", "base_price", "is_active", "is_default", "min_days", "max_days"},
		DetailFields:    []string{"id", "name", "code", "description", "carrier", "service_level", "base_price", "handling_fee", "free_shipping_threshold", "min_weight", "max_weight", "weight_unit", "min_length", "max_length", "min_width", "max_width", "min_height", "max_height", "dimension_unit", "min_days", "max_days", "countries", "excluded_countries", "is_active", "is_default", "priority", "config", "created_at", "updated_at"},
		Filterable:      []string{"is_active", "is_default", "carrier", "service_level"},
		Searchable:      []string{"name", "code", "carrier", "description"},
		Ordering:        []string{"priority", "name", "base_price"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: shippingMethodViewSet,
	})

	// PaymentMethod API with enhanced features
	paymentMethodViewSet := api.NewViewSet(&PaymentMethod{})
	paymentMethodViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			commerceUserLookup,
		),
	}
	paymentMethodViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	paymentMethodViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for PaymentMethod
	paymentMethodViewSet.RegisterAction("configure", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      true,
		URLPath:     "/configure",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handlePaymentConfigure,
	})

	paymentMethodViewSet.RegisterAction("test-connection", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/test-connection",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handlePaymentTestConnection,
	})

	router.Register("payment-methods", &api.ViewSetConfig{
		Model:           &PaymentMethod{},
		Queryset:        PaymentMethodObjects,
		Serializer:      &PaymentMethodSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "name", "code", "type", "gateway", "fixed_fee", "percentage_fee", "is_active", "is_default"},
		DetailFields:    []string{"id", "name", "code", "type", "description", "gateway", "gateway_config", "fixed_fee", "percentage_fee", "min_amount", "max_amount", "currencies", "icon", "display_order", "is_active", "is_default", "test_mode", "created_at", "updated_at"},
		Filterable:      []string{"is_active", "is_default", "type", "gateway"},
		Searchable:      []string{"name", "code", "description"},
		Ordering:        []string{"display_order", "name"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: paymentMethodViewSet,
	})

	// TaxRate API with enhanced features
	taxRateViewSet := api.NewViewSet(&TaxRate{})
	taxRateViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			commerceUserLookup,
		),
	}
	taxRateViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	taxRateViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for TaxRate
	taxRateViewSet.RegisterAction("calculate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/calculate",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleTaxCalculate,
	})

	taxRateViewSet.RegisterAction("by-location", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/by-location",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleTaxByLocation,
	})

	router.Register("tax-rates", &api.ViewSetConfig{
		Model:           &TaxRate{},
		Queryset:        TaxRateObjects,
		Serializer:      &TaxRateSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "name", "code", "rate", "country", "state", "tax_type", "is_active", "priority"},
		DetailFields:    []string{"id", "name", "code", "description", "rate", "country", "state", "zip_pattern", "city", "tax_type", "applies_to_products", "applies_to_shipping", "applies_to_services", "start_date", "end_date", "is_active", "is_compound", "priority", "created_at", "updated_at"},
		Filterable:      []string{"is_active", "country", "state", "tax_type", "is_compound"},
		Searchable:      []string{"name", "code", "description"},
		Ordering:        []string{"country", "state", "priority"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: taxRateViewSet,
	})

	// Currency API with enhanced features
	currencyViewSet := api.NewViewSet(&Currency{})
	currencyViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			commerceUserLookup,
		),
	}
	currencyViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	currencyViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for Currency
	currencyViewSet.RegisterAction("convert", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/convert",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleCurrencyConvert,
	})

	currencyViewSet.RegisterAction("latest-rates", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/latest-rates",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleLatestRates,
	})

	router.Register("currencies", &api.ViewSetConfig{
		Model:           &Currency{},
		Queryset:        CurrencyObjects,
		Serializer:      &CurrencySerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "code", "name", "symbol", "decimal_places", "is_base_currency", "exchange_rate", "is_active", "is_default"},
		DetailFields:    []string{"id", "code", "name", "symbol", "decimal_places", "decimal_separator", "thousand_separator", "symbol_position", "is_base_currency", "exchange_rate", "is_active", "is_default", "created_at", "updated_at"},
		Filterable:      []string{"is_active", "is_default", "is_base_currency"},
		Searchable:      []string{"code", "name", "symbol"},
		Ordering:        []string{"code"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: currencyViewSet,
	})

	// ExchangeRate API with enhanced features
	exchangeRateViewSet := api.NewViewSet(&ExchangeRate{})
	exchangeRateViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			commerceUserLookup,
		),
	}
	exchangeRateViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	exchangeRateViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	router.Register("exchange-rates", &api.ViewSetConfig{
		Model:           &ExchangeRate{},
		Queryset:        ExchangeRateObjects,
		Serializer:      &ExchangeRateSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "from_currency", "to_currency", "rate", "effective_from", "effective_to", "provider"},
		DetailFields:    []string{"id", "from_currency", "to_currency", "rate", "effective_from", "effective_to", "provider", "created_at", "updated_at"},
		Filterable:      []string{"from_currency", "to_currency"},
		Searchable:      []string{"from_currency", "to_currency", "provider"},
		Ordering:        []string{"from_currency", "to_currency", "-effective_from"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: exchangeRateViewSet,
	})
}

// commerceUserLookup looks up a user from JWT claims
func commerceUserLookup(claims authentication.JWTClaims) (interface{}, error) {
	// Extract user ID from subject claim
	userID, ok := claims["subject"].(string)
	if !ok {
		if subject, ok := claims["subject"].(float64); ok {
			userID = strconv.FormatFloat(subject, 'f', -1, 64)
		} else {
			return nil, nil
		}
	}
	// In a real implementation, you would query the database
	return nil, nil
}

// Custom action handlers for ShippingMethod

func handleShippingCalculate(w http.ResponseWriter, r *http.Request) {
	// Extract request body
	var req struct {
		ShippingMethodID string  `json:"shipping_method_id"`
		Weight          float64 `json:"weight"`
		Subtotal        float64 `json:"subtotal"`
		Country         string  `json:"country"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate shipping cost (placeholder implementation)
	response := map[string]interface{}{
		"shipping_method_id": req.ShippingMethodID,
		"base_price":        10.0,
		"handling_fee":      2.0,
		"weight_fee":        req.Weight * 0.5,
		"subtotal":          req.Subtotal,
		"free_shipping":     req.Subtotal >= 100.0,
		"total":             12.0 + req.Weight*0.5,
	}

	if req.Subtotal >= 100.0 {
		response["total"] = 0.0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleShippingAvailable(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	weight, _ := strconv.ParseFloat(r.URL.Query().Get("weight"), 64)

	// Return available shipping methods (placeholder)
	response := map[string]interface{}{
		"country": country,
		"weight": weight,
		"methods": []interface{}{
			map[string]interface{}{
				"id":           "1",
				"name":         "Standard Shipping",
				"carrier":      "Generic Carrier",
				"base_price":   10.0,
				"min_days":     5,
				"max_days":     7,
			},
			map[string]interface{}{
				"id":           "2",
				"name":         "Express Shipping",
				"carrier":      "Generic Carrier",
				"base_price":   25.0,
				"min_days":     1,
				"max_days":     3,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleShippingZones(w http.ResponseWriter, r *http.Request) {
	// Return shipping zones (placeholder)
	response := map[string]interface{}{
		"zones": []interface{}{
			map[string]interface{}{
				"id":         "1",
				"name":       "Domestic",
				"countries":  []string{"US", "CA"},
				"methods":    []string{"standard", "express"},
			},
			map[string]interface{}{
				"id":         "2",
				"name":       "International",
				"countries":  []string{"UK", "DE", "FR"},
				"methods":    []string{"international"},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for PaymentMethod

func handlePaymentConfigure(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	paymentMethodID := vars["id"]

	// Return gateway configuration (placeholder)
	response := map[string]interface{}{
		"payment_method_id": paymentMethodID,
		"gateway": map[string]interface{}{
			"name":      "stripe",
			"api_key":   "sk_test_****",
			"mode":      "test",
			"webhook":   "https://example.com/webhook",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePaymentTestConnection(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	paymentMethodID := vars["id"]

	// Test gateway connection (placeholder)
	response := map[string]interface{}{
		"payment_method_id": paymentMethodID,
		"status":            "success",
		"message":           "Gateway connection successful",
		"latency":           "125ms",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for TaxRate

func handleTaxCalculate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Country  string  `json:"country"`
		State    string  `json:"state"`
		Subtotal float64 `json:"subtotal"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate tax (placeholder)
	taxRate := 0.08 // 8% default
	if req.Country == "US" && req.State == "CA" {
		taxRate = 0.0725 // California sales tax
	} else if req.Country == "UK" {
		taxRate = 0.20 // UK VAT
	}

	taxAmount := req.Subtotal * taxRate

	response := map[string]interface{}{
		"country":     req.Country,
		"state":       req.State,
		"subtotal":    req.Subtotal,
		"tax_rate":    taxRate,
		"tax_amount":  taxAmount,
		"total":       req.Subtotal + taxAmount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTaxByLocation(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	state := r.URL.Query().Get("state")
	zipCode := r.URL.Query().Get("zip")

	// Return tax rates for location (placeholder)
	response := map[string]interface{}{
		"country":  country,
		"state":    state,
		"zip":      zipCode,
		"tax_rates": []interface{}{
			map[string]interface{}{
				"id":    "1",
				"name":  "Standard Sales Tax",
				"rate":  0.08,
				"type":  "sales",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for Currency

func handleCurrencyConvert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromCurrency string  `json:"from_currency"`
		ToCurrency   string  `json:"to_currency"`
		Amount       float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert currency (placeholder - would use actual exchange rates)
	rate := 1.0
	if req.FromCurrency != req.ToCurrency {
		// Simple conversion rates for demo
		rates := map[string]float64{
			"USD": 1.0,
			"EUR": 0.92,
			"GBP": 0.79,
			"JPY": 149.50,
		}

		fromRate := rates[req.FromCurrency]
		toRate := rates[req.ToCurrency]

		if fromRate > 0 && toRate > 0 {
			rate = toRate / fromRate
		}
	}

	convertedAmount := req.Amount * rate

	response := map[string]interface{}{
		"from_currency":   req.FromCurrency,
		"to_currency":     req.ToCurrency,
		"original_amount": req.Amount,
		"rate":           rate,
		"converted_amount": convertedAmount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleLatestRates(w http.ResponseWriter, r *http.Request) {
	baseCurrency := r.URL.Query().Get("base")
	if baseCurrency == "" {
		baseCurrency = "USD"
	}

	// Return latest exchange rates (placeholder)
	response := map[string]interface{}{
		"base_currency": baseCurrency,
		"date":          time.Now().Format("2006-01-02"),
		"rates": map[string]interface{}{
			"USD": 1.0,
			"EUR": 0.92,
			"GBP": 0.79,
			"JPY": 149.50,
			"CAD": 1.36,
			"AUD": 1.53,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Serializers with validation

type ShippingMethodSerializer struct {
	*api.BaseSerializer
}

func (s *ShippingMethodSerializer) New() api.Serializer {
	return &ShippingMethodSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *ShippingMethodSerializer) Validate() error {
	errors := make(map[string]string)

	if s.Data["name"] == "" {
		errors["name"] = "Name is required"
	}
	if s.Data["code"] == "" {
		errors["code"] = "Code is required"
	}
	if s.Data["carrier"] == "" {
		errors["carrier"] = "Carrier is required"
	}

	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type PaymentMethodSerializer struct {
	*api.BaseSerializer
}

func (s *PaymentMethodSerializer) New() api.Serializer {
	return &PaymentMethodSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *PaymentMethodSerializer) Validate() error {
	errors := make(map[string]string)

	if s.Data["name"] == "" {
		errors["name"] = "Name is required"
	}
	if s.Data["code"] == "" {
		errors["code"] = "Code is required"
	}
	if s.Data["type"] == "" {
		errors["type"] = "Type is required"
	}

	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type TaxRateSerializer struct {
	*api.BaseSerializer
}

func (s *TaxRateSerializer) New() api.Serializer {
	return &TaxRateSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *TaxRateSerializer) Validate() error {
	errors := make(map[string]string)

	if s.Data["name"] == "" {
		errors["name"] = "Name is required"
	}
	if s.Data["code"] == "" {
		errors["code"] = "Code is required"
	}
	if s.Data["country"] == "" {
		errors["country"] = "Country is required"
	}

	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type CurrencySerializer struct {
	*api.BaseSerializer
}

func (s *CurrencySerializer) New() api.Serializer {
	return &CurrencySerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *CurrencySerializer) Validate() error {
	errors := make(map[string]string)

	if s.Data["code"] == "" {
		errors["code"] = "Code is required"
	}
	if s.Data["name"] == "" {
		errors["name"] = "Name is required"
	}
	if s.Data["symbol"] == "" {
		errors["symbol"] = "Symbol is required"
	}

	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type ExchangeRateSerializer struct {
	*api.BaseSerializer
}

func (s *ExchangeRateSerializer) New() api.Serializer {
	return &ExchangeRateSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *ExchangeRateSerializer) Validate() error {
	errors := make(map[string]string)

	if s.Data["from_currency"] == "" {
		errors["from_currency"] = "From currency is required"
	}
	if s.Data["to_currency"] == "" {
		errors["to_currency"] = "To currency is required"
	}
	if s.Data["rate"] == nil || s.Data["rate"].(float64) <= 0 {
		errors["rate"] = "Rate must be greater than 0"
	}

	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

