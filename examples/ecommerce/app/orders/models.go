package orders

import (
	"context"

	"github.com/forgego/forge/registry"
	"github.com/forgego/forge/schema"
)

// Cart represents a shopping cart
type Cart struct {
	CartGenerated
}

func (Cart) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement(),
			schema.VerboseName("ID")),
		schema.Int64Field("customer_id", schema.Optional(),
			schema.HelpText("Customer (null for guest carts)")),

		// Guest cart identification
		schema.StringField("session_id", schema.MaxLength(255), schema.Optional(),
			schema.HelpText("Session ID for guest carts")),
		schema.StringField("guest_email", schema.MaxLength(255), schema.Optional(),
			schema.VerboseName("Guest Email")),

		// Pricing
		schema.Float64Field("subtotal", schema.Default(0.0),
			schema.HelpText("Cart subtotal before discounts")),
		schema.Float64Field("discount_amount", schema.Default(0.0)),
		schema.Float64Field("tax_amount", schema.Default(0.0)),
		schema.Float64Field("shipping_amount", schema.Default(0.0)),
		schema.Float64Field("total", schema.Default(0.0),
			schema.VerboseName("Total"),
			schema.HelpText("Cart total")),

		// Coupon
		schema.Int64Field("coupon_id", schema.Optional()),
		schema.StringField("coupon_code", schema.MaxLength(50), schema.Optional()),

		// Status
		schema.StringField("status", schema.Required(), schema.MaxLength(20), schema.Default("active"),
			schema.VerboseName("Status"),
			schema.HelpText("Status: active, abandoned, converted")),
		schema.BoolField("is_abandoned", schema.Default(false),
			schema.VerboseName("Abandoned")),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
		schema.TimeField("updated_at", schema.AutoNow()),
		schema.TimeField("last_activity_at", schema.Optional(),
			schema.VerboseName("Last Activity At")),
		schema.TimeField("converted_at", schema.Optional(),
			schema.HelpText("When cart was converted to order")),
	}
}

func (Cart) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "carts",
		VerboseName:       "Shopping Cart",
		VerboseNamePlural: "Shopping Carts",
		OrderBy:           []string{"-updated_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_cart_customer", "customer_id"),
			schema.IndexOn("idx_cart_session", "session_id"),
			schema.IndexOn("idx_cart_status", "status"),
			schema.IndexOn("idx_cart_abandoned", "is_abandoned"),
		},
	}
}

func (Cart) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("carts")),
		schema.ForeignKeyField("coupon_id", "Coupon",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("carts")),
	}
}

func (Cart) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		AfterSave: func(ctx context.Context, instance interface{}) error {
			// Recalculate cart totals
			// Check for abandoned cart
			return nil
		},
	}
}

// CartItem represents items in a shopping cart
type CartItem struct {
	CartItemGenerated
}

func (CartItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("cart_id", schema.Required(),
			schema.VerboseName("Cart")),
		schema.Int64Field("product_id", schema.Required()),
		schema.Int64Field("variant_id", schema.Optional()),

		// Quantity and pricing
		schema.Int32Field("quantity", schema.Required(), schema.Default(1)),
		schema.Float64Field("unit_price", schema.Required(),
			schema.VerboseName("Unit Price"),
			schema.HelpText("Price per unit at time of adding")),
		schema.Float64Field("discount_amount", schema.Default(0.0)),
		schema.Float64Field("tax_amount", schema.Default(0.0)),
		schema.Float64Field("total", schema.Required(),
			schema.VerboseName("Total"),
			schema.HelpText("Line total (quantity * unit_price - discount + tax)")),

		// Product snapshot (for price changes)
		schema.StringField("product_name", schema.MaxLength(255), schema.Optional(),
			schema.VerboseName("Product")),
		schema.StringField("variant_name", schema.MaxLength(255), schema.Optional(),
			schema.VerboseName("Variant")),
		schema.StringField("image_url", schema.MaxLength(500), schema.Optional()),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (CartItem) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "cart_items",
		VerboseName:       "Cart Item",
		VerboseNamePlural: "Cart Items",
		OrderBy:           []string{"created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_cart_item_cart", "cart_id"),
			schema.IndexOn("idx_cart_item_product", "product_id"),
			schema.IndexOn("idx_cart_item_variant", "variant_id"),
		},
		UniqueTogether: [][]string{
			{"cart_id", "product_id", "variant_id"},
		},
	}
}

func (CartItem) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("cart_id", "Cart",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("items")),
		schema.ForeignKeyField("product_id", "Product",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("cart_items")),
		schema.ForeignKeyField("variant_id", "ProductVariant",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("cart_items")),
	}
}

func (CartItem) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Calculate line total
			// Capture product snapshot
			return nil
		},
		AfterSave: func(ctx context.Context, instance interface{}) error {
			// Update cart totals
			return nil
		},
	}
}

// Order represents a customer order
type Order struct {
	OrderGenerated
}

func (Order) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),

		// Order identification
		schema.StringField("order_number", schema.Required(), schema.MaxLength(50), schema.Unique(),
			schema.VerboseName("Order Number"),
			schema.HelpText("Human-readable order number")),

		// Customer
		schema.Int64Field("customer_id", schema.Required()),
		schema.StringField("customer_email", schema.Required(), schema.MaxLength(255),
			schema.VerboseName("Customer Email")),
		schema.StringField("customer_first_name", schema.Required(), schema.MaxLength(100)),
		schema.StringField("customer_last_name", schema.Required(), schema.MaxLength(100)),
		schema.StringField("customer_phone", schema.MaxLength(20), schema.Optional()),

		// Pricing
		schema.Float64Field("subtotal", schema.Required(),
			schema.HelpText("Order subtotal before discounts")),
		schema.Float64Field("discount_amount", schema.Default(0.0)),
		schema.Float64Field("tax_amount", schema.Required()),
		schema.Float64Field("shipping_amount", schema.Required()),
		schema.Float64Field("total", schema.Required(),
			schema.HelpText("Order total")),

		// Coupon
		schema.Int64Field("coupon_id", schema.Optional()),
		schema.StringField("coupon_code", schema.MaxLength(50), schema.Optional()),
		schema.Float64Field("coupon_discount", schema.Default(0.0)),

		// Status
		schema.StringField("status", schema.Required(), schema.MaxLength(20), schema.Default("pending"),
			schema.VerboseName("Status"),
			schema.HelpText("Status: pending, processing, shipped, delivered, cancelled, refunded")),
		schema.StringField("payment_status", schema.Required(), schema.MaxLength(20), schema.Default("pending"),
			schema.VerboseName("Payment Status"),
			schema.HelpText("Payment status: pending, paid, failed, refunded")),
		schema.StringField("fulfillment_status", schema.Required(), schema.MaxLength(20), schema.Default("unfulfilled"),
			schema.VerboseName("Fulfillment Status"),
			schema.HelpText("Fulfillment: unfulfilled, partial, fulfilled")),

		// Addresses
		schema.Int64Field("shipping_address_id", schema.Optional()),
		schema.Int64Field("billing_address_id", schema.Optional()),

		// Shipping address snapshot
		schema.StringField("shipping_first_name", schema.MaxLength(100), schema.Optional()),
		schema.StringField("shipping_last_name", schema.MaxLength(100), schema.Optional()),
		schema.StringField("shipping_company", schema.MaxLength(200), schema.Optional()),
		schema.StringField("shipping_address_line1", schema.MaxLength(255), schema.Optional()),
		schema.StringField("shipping_address_line2", schema.MaxLength(255), schema.Optional()),
		schema.StringField("shipping_city", schema.MaxLength(100), schema.Optional()),
		schema.StringField("shipping_state", schema.MaxLength(100), schema.Optional()),
		schema.StringField("shipping_postal_code", schema.MaxLength(20), schema.Optional()),
		schema.StringField("shipping_country_code", schema.MaxLength(2), schema.Optional()),
		schema.StringField("shipping_country_name", schema.MaxLength(100), schema.Optional()),
		schema.StringField("shipping_phone", schema.MaxLength(20), schema.Optional()),

		// Billing address snapshot
		schema.StringField("billing_first_name", schema.MaxLength(100), schema.Optional()),
		schema.StringField("billing_last_name", schema.MaxLength(100), schema.Optional()),
		schema.StringField("billing_company", schema.MaxLength(200), schema.Optional()),
		schema.StringField("billing_address_line1", schema.MaxLength(255), schema.Optional()),
		schema.StringField("billing_address_line2", schema.MaxLength(255), schema.Optional()),
		schema.StringField("billing_city", schema.MaxLength(100), schema.Optional()),
		schema.StringField("billing_state", schema.MaxLength(100), schema.Optional()),
		schema.StringField("billing_postal_code", schema.MaxLength(20), schema.Optional()),
		schema.StringField("billing_country_code", schema.MaxLength(2), schema.Optional()),
		schema.StringField("billing_country_name", schema.MaxLength(100), schema.Optional()),

		// Payment
		schema.StringField("payment_method", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Payment method: credit_card, paypal, stripe, etc.")),
		schema.StringField("payment_transaction_id", schema.MaxLength(255), schema.Optional()),

		// Shipping
		schema.StringField("shipping_method", schema.MaxLength(100), schema.Optional()),
		schema.StringField("tracking_number", schema.MaxLength(255), schema.Optional()),
		schema.StringField("carrier", schema.MaxLength(100), schema.Optional()),

		// Additional information
		schema.TextField("customer_notes", schema.Optional(),
			schema.HelpText("Notes from customer")),
		schema.TextField("admin_notes", schema.Optional(),
			schema.HelpText("Internal notes")),

		// IP and user agent (for fraud detection)
		schema.StringField("ip_address", schema.MaxLength(45), schema.Optional()),
		schema.StringField("user_agent", schema.MaxLength(500), schema.Optional()),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
		schema.TimeField("updated_at", schema.AutoNow()),
		schema.TimeField("paid_at", schema.Optional()),
		schema.TimeField("shipped_at", schema.Optional()),
		schema.TimeField("delivered_at", schema.Optional()),
		schema.TimeField("cancelled_at", schema.Optional()),
		schema.TimeField("expires_at", schema.Optional(),
			schema.HelpText("Order expiry for pending orders")),
	}
}

func (Order) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "orders",
		VerboseName:       "Order",
		VerboseNamePlural: "Orders",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_order_number", "order_number"),
			schema.IndexOn("idx_order_customer", "customer_id"),
			schema.IndexOn("idx_order_status", "status"),
			schema.IndexOn("idx_order_payment_status", "payment_status"),
			schema.IndexOn("idx_order_created", "created_at"),
			schema.IndexOn("idx_order_email", "customer_email"),
		},
	}
}

func (Order) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer",
			schema.OnDelete(schema.CascadePROTECT),
			schema.RelatedName("orders")),
		schema.ForeignKeyField("coupon_id", "Coupon",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("orders")),
		schema.ForeignKeyField("shipping_address_id", "Address",
			schema.OnDelete(schema.CascadePROTECT),
			schema.RelatedName("shipping_orders")),
		schema.ForeignKeyField("billing_address_id", "Address",
			schema.OnDelete(schema.CascadePROTECT),
			schema.RelatedName("billing_orders")),
	}
}

func (Order) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeCreate: func(ctx context.Context, instance interface{}) error {
			// Generate order number
			// Snapshot customer info
			// Snapshot addresses
			// Calculate totals
			// Set expiry date
			return nil
		},
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// Send order confirmation email
			// Reserve inventory
			// Update customer stats
			return nil
		},
		BeforeUpdate: func(ctx context.Context, instance interface{}) error {
			// Handle status changes
			// Update timestamps based on status
			return nil
		},
	}
}

// OrderItem represents line items in an order
type OrderItem struct {
	OrderItemGenerated
}

func (OrderItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("order_id", schema.Required(),
			schema.VerboseName("Order")),
		schema.Int64Field("product_id", schema.Required()),
		schema.Int64Field("variant_id", schema.Optional()),

		// Product snapshot (prices can change)
		schema.StringField("product_name", schema.Required(), schema.MaxLength(255)),
		schema.StringField("product_sku", schema.Required(), schema.MaxLength(100)),
		schema.StringField("variant_name", schema.MaxLength(255), schema.Optional()),
		schema.StringField("variant_sku", schema.MaxLength(100), schema.Optional()),
		schema.StringField("image_url", schema.MaxLength(500), schema.Optional()),

		// Quantity and pricing
		schema.Int32Field("quantity", schema.Required(), schema.Default(1)),
		schema.Float64Field("unit_price", schema.Required(),
			schema.VerboseName("Unit Price"),
			schema.HelpText("Price per unit at time of order")),
		schema.Float64Field("discount_amount", schema.Default(0.0)),
		schema.Float64Field("tax_amount", schema.Default(0.0)),
		schema.Float64Field("total", schema.Required(),
			schema.VerboseName("Total"),
			schema.HelpText("Line total")),

		// Fulfillment
		schema.Int32Field("quantity_fulfilled", schema.Default(0)),
		schema.Int32Field("quantity_refunded", schema.Default(0)),
		schema.StringField("fulfillment_status", schema.Required(), schema.MaxLength(20), schema.Default("unfulfilled"),
			schema.VerboseName("Fulfillment Status"),
			schema.HelpText("Status: unfulfilled, fulfilled, refunded")),

		// Physical attributes (for shipping calculations)
		schema.Float64Field("weight", schema.Optional()),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (OrderItem) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "order_items",
		VerboseName:       "Order Item",
		VerboseNamePlural: "Order Items",
		OrderBy:           []string{"created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_order_item_order", "order_id"),
			schema.IndexOn("idx_order_item_product", "product_id"),
			schema.IndexOn("idx_order_item_variant", "variant_id"),
		},
	}
}

func (OrderItem) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("order_id", "Order",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("items")),
		schema.ForeignKeyField("product_id", "Product",
			schema.OnDelete(schema.CascadePROTECT),
			schema.RelatedName("order_items")),
		schema.ForeignKeyField("variant_id", "ProductVariant",
			schema.OnDelete(schema.CascadePROTECT),
			schema.RelatedName("order_items")),
	}
}

func (OrderItem) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Calculate line total
			// Capture product snapshot
			return nil
		},
		AfterSave: func(ctx context.Context, instance interface{}) error {
			// Update order totals
			// Update product order_count
			return nil
		},
	}
}

// Payment represents payment transactions
type Payment struct {
	PaymentGenerated
}

func (Payment) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("order_id", schema.Required(),
			schema.VerboseName("Order")),

		// Payment identification
		schema.StringField("transaction_id", schema.MaxLength(255), schema.Optional(),
			schema.VerboseName("Transaction ID"),
			schema.HelpText("External payment gateway transaction ID")),

		// Amount
		schema.Float64Field("amount", schema.Required(),
			schema.HelpText("Payment amount")),
		schema.StringField("currency", schema.Required(), schema.MaxLength(3), schema.Default("USD"),
			schema.VerboseName("Currency")),

		// Payment method
		schema.StringField("payment_method", schema.Required(), schema.MaxLength(50),
			schema.VerboseName("Payment Method"),
			schema.HelpText("Method: credit_card, paypal, stripe, bank_transfer, etc.")),
		schema.StringField("payment_gateway", schema.MaxLength(50), schema.Optional(),
			schema.HelpText("Gateway used: stripe, paypal, square, etc.")),

		// Card details (last 4 digits only for security)
		schema.StringField("card_last4", schema.MaxLength(4), schema.Optional()),
		schema.StringField("card_brand", schema.MaxLength(50), schema.Optional()),

		// Status
		schema.StringField("status", schema.Required(), schema.MaxLength(20), schema.Default("pending"),
			schema.VerboseName("Status"),
			schema.HelpText("Status: pending, completed, failed, refunded, cancelled")),

		// Additional information
		schema.TextField("gateway_response", schema.Optional(),
			schema.HelpText("Full gateway response (JSON)")),
		schema.StringField("failure_reason", schema.MaxLength(500), schema.Optional()),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
		schema.TimeField("completed_at", schema.Optional()),
		schema.TimeField("failed_at", schema.Optional()),
		schema.TimeField("refunded_at", schema.Optional()),
	}
}

func (Payment) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "payments",
		VerboseName:       "Payment",
		VerboseNamePlural: "Payments",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_payment_order", "order_id"),
			schema.IndexOn("idx_payment_transaction", "transaction_id"),
			schema.IndexOn("idx_payment_status", "status"),
		},
	}
}

func (Payment) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("order_id", "Order",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("payments")),
	}
}

func (Payment) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		AfterUpdate: func(ctx context.Context, instance interface{}) error {
			// Update order payment status
			// Send payment notification
			return nil
		},
	}
}

// Shipment represents order shipments
type Shipment struct {
	ShipmentGenerated
}

func (Shipment) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("order_id", schema.Required(),
			schema.VerboseName("Order")),

		// Tracking
		schema.StringField("tracking_number", schema.Required(), schema.MaxLength(255),
			schema.VerboseName("Tracking Number")),
		schema.StringField("carrier", schema.Required(), schema.MaxLength(100),
			schema.HelpText("Carrier: USPS, UPS, FedEx, DHL, etc.")),
		schema.StringField("service_level", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("Service: Standard, Express, Overnight, etc.")),

		// Shipping address (snapshot)
		schema.StringField("recipient_name", schema.Required(), schema.MaxLength(200)),
		schema.StringField("address_line1", schema.Required(), schema.MaxLength(255)),
		schema.StringField("address_line2", schema.MaxLength(255), schema.Optional()),
		schema.StringField("city", schema.Required(), schema.MaxLength(100)),
		schema.StringField("state", schema.MaxLength(100), schema.Optional()),
		schema.StringField("postal_code", schema.Required(), schema.MaxLength(20)),
		schema.StringField("country_code", schema.Required(), schema.MaxLength(2)),
		schema.StringField("country_name", schema.Required(), schema.MaxLength(100)),
		schema.StringField("phone", schema.MaxLength(20), schema.Optional()),

		// Shipping details
		schema.Float64Field("weight", schema.Optional(),
			schema.HelpText("Total weight in kg")),
		schema.Float64Field("shipping_cost", schema.Required()),

		// Status
		schema.StringField("status", schema.Required(), schema.MaxLength(20), schema.Default("pending"),
			schema.VerboseName("Status"),
			schema.HelpText("Status: pending, in_transit, delivered, failed, returned")),

		// Tracking events
		schema.TextField("tracking_events", schema.Optional(),
			schema.HelpText("JSON array of tracking events")),

		// Notes
		schema.TextField("notes", schema.Optional()),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
		schema.TimeField("shipped_at", schema.Optional(),
			schema.VerboseName("Shipped At")),
		schema.TimeField("estimated_delivery_at", schema.Optional()),
		schema.TimeField("delivered_at", schema.Optional()),
	}
}

func (Shipment) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "shipments",
		VerboseName:       "Shipment",
		VerboseNamePlural: "Shipments",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_shipment_order", "order_id"),
			schema.IndexOn("idx_shipment_tracking", "tracking_number"),
			schema.IndexOn("idx_shipment_status", "status"),
		},
	}
}

func (Shipment) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("order_id", "Order",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("shipments")),
	}
}

func (Shipment) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// Send shipping notification
			// Update order status
			return nil
		},
		AfterUpdate: func(ctx context.Context, instance interface{}) error {
			// Send tracking updates
			// Update order delivery status
			return nil
		},
	}
}

// RegisterModels registers order models with the framework
func RegisterModels() {
	registry.RegisterModel(&Cart{})
	registry.RegisterModel(&CartItem{})
	registry.RegisterModel(&Order{})
	registry.RegisterModel(&OrderItem{})
	registry.RegisterModel(&Payment{})
	registry.RegisterModel(&Shipment{})
}
