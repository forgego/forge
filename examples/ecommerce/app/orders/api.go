package orders

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers order API endpoints
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// Cart API
	router.Register("carts", &api.ViewSetConfig{
		Model:        &Cart{},
		Queryset:     CartObjects,
		Serializer:   &CartSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "customer_id", "guest_email", "subtotal", "discount_amount", "tax_amount", "shipping_amount", "total", "status"},
		DetailFields: []string{"id", "customer_id", "session_id", "guest_email", "subtotal", "discount_amount", "tax_amount", "shipping_amount", "total", "coupon_id", "coupon_code", "status", "is_abandoned", "created_at", "updated_at", "last_activity_at", "converted_at"},
		Filterable:   []string{"customer_id", "status", "is_abandoned"},
		Searchable:   []string{"guest_email", "session_id"},
		Ordering:     []string{"-updated_at", "-created_at"},
		PerPage:      20,
	})

	// CartItem API
	router.Register("cart-items", &api.ViewSetConfig{
		Model:        &CartItem{},
		Queryset:     CartItemObjects,
		Serializer:   &CartItemSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "cart_id", "product_id", "variant_id", "product_name", "quantity", "unit_price", "total"},
		DetailFields: []string{"id", "cart_id", "product_id", "variant_id", "product_name", "variant_name", "image_url", "quantity", "unit_price", "discount_amount", "tax_amount", "total", "created_at", "updated_at"},
		Filterable:   []string{"cart_id", "product_id"},
		Searchable:   []string{"product_name", "variant_name"},
		Ordering:     []string{"cart_id", "created_at"},
		PerPage:      50,
	})

	// Order API
	router.Register("orders", &api.ViewSetConfig{
		Model:        &Order{},
		Queryset:     OrderObjects,
		Serializer:   &OrderSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "order_number", "customer_id", "customer_email", "subtotal", "discount_amount", "tax_amount", "shipping_amount", "total", "status", "payment_status", "fulfillment_status", "created_at"},
		DetailFields: []string{"id", "order_number", "customer_id", "customer_email", "customer_first_name", "customer_last_name", "customer_phone", "subtotal", "discount_amount", "tax_amount", "shipping_amount", "total", "coupon_code", "coupon_discount", "status", "payment_status", "fulfillment_status", "shipping_address_id", "billing_address_id", "payment_method", "payment_transaction_id", "shipping_method", "tracking_number", "carrier", "customer_notes", "ip_address", "created_at", "updated_at", "paid_at", "shipped_at", "delivered_at", "cancelled_at", "expires_at"},
		Filterable:   []string{"customer_id", "status", "payment_status", "fulfillment_status", "created_at"},
		Searchable:   []string{"order_number", "customer_email", "customer_first_name", "customer_last_name", "tracking_number"},
		Ordering:     []string{"-created_at", "-total", "status"},
		PerPage:      20,
	})

	// OrderItem API
	router.Register("order-items", &api.ViewSetConfig{
		Model:        &OrderItem{},
		Queryset:     OrderItemObjects,
		Serializer:   &OrderItemSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "order_id", "product_name", "variant_name", "quantity", "unit_price", "total", "fulfillment_status"},
		DetailFields: []string{"id", "order_id", "product_id", "variant_id", "product_name", "product_sku", "variant_name", "variant_sku", "image_url", "quantity", "unit_price", "discount_amount", "tax_amount", "total", "quantity_fulfilled", "quantity_refunded", "fulfillment_status", "weight", "created_at", "updated_at"},
		Filterable:   []string{"order_id", "product_id", "fulfillment_status"},
		Searchable:   []string{"product_name", "product_sku", "variant_sku"},
		Ordering:     []string{"order_id", "created_at"},
		PerPage:      50,
	})

	// Payment API
	router.Register("payments", &api.ViewSetConfig{
		Model:        &Payment{},
		Queryset:     PaymentObjects,
		Serializer:   &PaymentSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "order_id", "amount", "currency", "payment_method", "status", "transaction_id", "created_at"},
		DetailFields: []string{"id", "order_id", "transaction_id", "amount", "currency", "payment_method", "payment_gateway", "card_last4", "card_brand", "status", "failure_reason", "created_at", "updated_at", "completed_at", "failed_at", "refunded_at"},
		Filterable:   []string{"order_id", "status", "payment_method", "payment_gateway"},
		Searchable:   []string{"transaction_id", "order_id"},
		Ordering:     []string{"-created_at", "status"},
		PerPage:      20,
	})

	// Shipment API
	router.Register("shipments", &api.ViewSetConfig{
		Model:        &Shipment{},
		Queryset:     ShipmentObjects,
		Serializer:   &ShipmentSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "order_id", "tracking_number", "carrier", "status", "shipping_cost", "shipped_at", "estimated_delivery_at", "delivered_at"},
		DetailFields: []string{"id", "order_id", "tracking_number", "carrier", "service_level", "recipient_name", "address_line1", "address_line2", "city", "state", "postal_code", "country_code", "country_name", "phone", "weight", "shipping_cost", "status", "tracking_events", "notes", "created_at", "updated_at", "shipped_at", "estimated_delivery_at", "delivered_at"},
		Filterable:   []string{"order_id", "status", "carrier"},
		Searchable:   []string{"tracking_number", "recipient_name", "order_id"},
		Ordering:     []string{"-created_at", "status"},
		PerPage:      20,
	})
}

// Serializers

type CartSerializer struct {
	*api.BaseSerializer
}

func (s *CartSerializer) New() api.Serializer {
	return &CartSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type CartItemSerializer struct {
	*api.BaseSerializer
}

func (s *CartItemSerializer) New() api.Serializer {
	return &CartItemSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type OrderSerializer struct {
	*api.BaseSerializer
}

func (s *OrderSerializer) New() api.Serializer {
	return &OrderSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type OrderItemSerializer struct {
	*api.BaseSerializer
}

func (s *OrderItemSerializer) New() api.Serializer {
	return &OrderItemSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type PaymentSerializer struct {
	*api.BaseSerializer
}

func (s *PaymentSerializer) New() api.Serializer {
	return &PaymentSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type ShipmentSerializer struct {
	*api.BaseSerializer
}

func (s *ShipmentSerializer) New() api.Serializer {
	return &ShipmentSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}
