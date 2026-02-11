package orders

import (
	"context"
	"time"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers order models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// Cart admin
	admin.Register(&admin.Config[Cart]{
		Icon: "ShoppingCart",
		ListDisplay: []admin.Field{
			CartFieldsInstance.Id,
			CartFieldsInstance.CustomerId,
			CartFieldsInstance.GuestEmail,
			CartFieldsInstance.Total,
			CartFieldsInstance.Status,
			CartFieldsInstance.IsAbandoned,
			CartFieldsInstance.LastActivityAt,
			CartFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			CartFieldsInstance.Status,
			CartFieldsInstance.IsAbandoned,
		},
		SearchFields: []admin.Field{
			CartFieldsInstance.GuestEmail,
			CartFieldsInstance.CustomerId,
		},
		Fieldsets: []admin.Fieldset[Cart]{
			{
				Name: "Cart Information",
				Fields: []string{"customer_id", "session_id", "guest_email"},
			},
			{
				Name: "Pricing",
				Fields: []string{"subtotal", "discount_amount", "tax_amount", "shipping_amount", "total"},
			},
			{
				Name: "Coupon",
				Fields: []string{"coupon_id", "coupon_code"},
			},
			{
				Name: "Status",
				Fields: []string{"status", "is_abandoned"},
			},
			{
				Name: "Timestamps",
				Fields: []string{"created_at", "updated_at", "last_activity_at", "converted_at"},
			},
		},
		InlineRelations: map[string]admin.InlineRelationConfig{
			"items": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Cart Items",
				RelatedModel: "CartItem",
				RelatedField: "cart_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"product_name", "quantity", "unit_price", "total"},
					Fieldsets: []admin.Fieldset[CartItem]{
						{Name: "Product", Fields: []string{"product_id", "variant_id", "product_name", "variant_name"}},
						{Name: "Pricing", Fields: []string{"quantity", "unit_price", "discount_amount", "tax_amount", "total"}},
					},
				},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "is_abandoned", "total"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "converted_at"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Cart], user interface{}, obj *Cart) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Cart], user interface{}, obj *Cart) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Cart], user interface{}, obj *Cart) bool {
			return true
		},
		Filters: []admin.Filter[Cart]{
			{
				Name:  "abandoned_carts",
				Label: "Abandoned Carts",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Cart], value interface{}) orm.QuerySet[Cart] {
					return qs.Filter(orm.F("is_abandoned").Eq(true))
				},
			},
			{
				Name:  "active_carts",
				Label: "Active Carts",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Cart], value interface{}) orm.QuerySet[Cart] {
					return qs.Filter(orm.F("status").Eq("active"))
				},
			},
			{
				Name:  "high_value",
				Label: "High Value Carts ($100+)",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Cart], value interface{}) orm.QuerySet[Cart] {
					return qs.Filter(orm.F("total").Gte(100))
				},
			},
		},
		Actions: []admin.Action[Cart]{
			{
				Name:  "mark_abandoned",
				Label: "Mark Abandoned",
				Handler: func(ctx context.Context, instances []*Cart) error {
					for _, cart := range instances {
						cart.IsAbandoned = true
						cart.Status = "abandoned"
						if err := CartObjects.Update(ctx, cart); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "mark_active",
				Label: "Mark Active",
				Handler: func(ctx context.Context, instances []*Cart) error {
					for _, cart := range instances {
						cart.IsAbandoned = false
						cart.Status = "active"
						if err := CartObjects.Update(ctx, cart); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// CartItem admin
	admin.Register(&admin.Config[CartItem]{
		Icon: "ShoppingBag",
		ListDisplay: []admin.Field{
			CartItemFieldsInstance.CartId,
			CartItemFieldsInstance.ProductName,
			CartItemFieldsInstance.VariantName,
			CartItemFieldsInstance.Quantity,
			CartItemFieldsInstance.UnitPrice,
			CartItemFieldsInstance.Total,
		},
		ListFilter: []admin.Field{
			CartItemFieldsInstance.CartId,
			CartItemFieldsInstance.ProductId,
		},
		Fieldsets: []admin.Fieldset[CartItem]{
			{
				Name: "Item",
				Fields: []string{"cart_id", "product_id", "variant_id"},
			},
			{
				Name: "Product Snapshot",
				Fields: []string{"product_name", "variant_name", "image_url"},
			},
			{
				Name: "Pricing",
				Fields: []string{"quantity", "unit_price", "discount_amount", "tax_amount", "total"},
			},
			{
				Name:  "Timestamps",
				Fields: []string{"created_at", "updated_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"quantity", "unit_price", "discount_amount"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "total"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[CartItem], user interface{}, obj *CartItem) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[CartItem], user interface{}, obj *CartItem) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[CartItem], user interface{}, obj *CartItem) bool {
			return true
		},
		Actions: []admin.Action[CartItem]{
			{
				Name:  "reset_quantity",
				Label: "Reset Quantity to 1",
				Handler: func(ctx context.Context, instances []*CartItem) error {
					for _, item := range instances {
						item.Quantity = 1
						if err := CartItemObjects.Update(ctx, item); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "increase_quantity",
				Label: "Increase Quantity by 1",
				Handler: func(ctx context.Context, instances []*CartItem) error {
					for _, item := range instances {
						item.Quantity++
						if err := CartItemObjects.Update(ctx, item); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// Order admin
	admin.Register(&admin.Config[Order]{
		Icon: "ClipboardList",
		ListDisplay: []admin.Field{
			OrderFieldsInstance.OrderNumber,
			OrderFieldsInstance.CustomerEmail,
			OrderFieldsInstance.Total,
			OrderFieldsInstance.Status,
			OrderFieldsInstance.PaymentStatus,
			OrderFieldsInstance.FulfillmentStatus,
			OrderFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			OrderFieldsInstance.Status,
			OrderFieldsInstance.PaymentStatus,
			OrderFieldsInstance.FulfillmentStatus,
		},
		SearchFields: []admin.Field{
			OrderFieldsInstance.OrderNumber,
			OrderFieldsInstance.CustomerEmail,
			OrderFieldsInstance.CustomerFirstName,
			OrderFieldsInstance.CustomerLastName,
		},
		Fieldsets: []admin.Fieldset[Order]{
			{
				Name: "Order Information",
				Fields: []string{"order_number", "customer_id", "customer_email", "customer_first_name", "customer_last_name", "customer_phone"},
			},
			{
				Name: "Pricing",
				Fields: []string{"subtotal", "discount_amount", "tax_amount", "shipping_amount", "total"},
			},
			{
				Name: "Coupon",
				Fields: []string{"coupon_id", "coupon_code", "coupon_discount"},
			},
			{
				Name: "Status",
				Fields: []string{"status", "payment_status", "fulfillment_status"},
			},
			{
				Name: "Shipping Address",
				Fields: []string{"shipping_first_name", "shipping_last_name", "shipping_company", "shipping_address_line1", "shipping_address_line2", "shipping_city", "shipping_state", "shipping_postal_code", "shipping_country_code", "shipping_phone"},
			},
			{
				Name: "Billing Address",
				Fields: []string{"billing_first_name", "billing_last_name", "billing_company", "billing_address_line1", "billing_address_line2", "billing_city", "billing_state", "billing_postal_code", "billing_country_code"},
			},
			{
				Name: "Payment",
				Fields: []string{"payment_method", "payment_transaction_id"},
			},
			{
				Name: "Shipping",
				Fields: []string{"shipping_method", "tracking_number", "carrier"},
			},
			{
				Name: "Notes",
				Fields: []string{"customer_notes", "admin_notes"},
			},
			{
				Name: "Technical",
				Fields: []string{"ip_address", "user_agent"},
			},
			{
				Name: "Timestamps",
				Fields: []string{"created_at", "updated_at", "paid_at", "shipped_at", "delivered_at", "cancelled_at", "expires_at"},
			},
		},
		InlineRelations: map[string]admin.InlineRelationConfig{
			"items": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Order Items",
				RelatedModel: "OrderItem",
				RelatedField: "order_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"product_name", "quantity", "unit_price", "total", "fulfillment_status"},
					Fieldsets: []admin.Fieldset[OrderItem]{
						{Name: "Product", Fields: []string{"product_id", "variant_id", "product_name", "product_sku", "variant_name", "variant_sku", "image_url"}},
						{Name: "Pricing", Fields: []string{"quantity", "unit_price", "discount_amount", "tax_amount", "total"}},
						{Name: "Fulfillment", Fields: []string{"quantity_fulfilled", "quantity_refunded", "fulfillment_status"}},
					},
				},
			},
			"payments": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Payments",
				RelatedModel: "Payment",
				RelatedField: "order_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"transaction_id", "amount", "currency", "payment_method", "status"},
				},
			},
			"shipments": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Shipments",
				RelatedModel: "Shipment",
				RelatedField: "order_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"tracking_number", "carrier", "status", "shipped_at"},
				},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "payment_status", "fulfillment_status", "total", "admin_notes"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "order_number", "paid_at", "shipped_at", "delivered_at", "cancelled_at"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Order], user interface{}, obj *Order) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Order], user interface{}, obj *Order) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Order], user interface{}, obj *Order) bool {
			return false // Orders should not be deleted, only cancelled
		},
		Filters: []admin.Filter[Order]{
			{
				Name:  "pending_orders",
				Label: "Pending Orders",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Order], value interface{}) orm.QuerySet[Order] {
					return qs.Filter(orm.F("status").Eq("pending"))
				},
			},
			{
				Name:  "processing_orders",
				Label: "Processing Orders",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Order], value interface{}) orm.QuerySet[Order] {
					return qs.Filter(orm.F("status").Eq("processing"))
				},
			},
			{
				Name:  "shipped_orders",
				Label: "Shipped Orders",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Order], value interface{}) orm.QuerySet[Order] {
					return qs.Filter(orm.F("status").Eq("shipped"))
				},
			},
			{
				Name:  "completed_orders",
				Label: "Completed Orders",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Order], value interface{}) orm.QuerySet[Order] {
					return qs.Filter(orm.F("status").Eq("delivered"))
				},
			},
			{
				Name:  "cancelled_orders",
				Label: "Cancelled Orders",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Order], value interface{}) orm.QuerySet[Order] {
					return qs.Filter(orm.F("status").Eq("cancelled"))
				},
			},
			{
				Name:  "unpaid_orders",
				Label: "Unpaid Orders",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Order], value interface{}) orm.QuerySet[Order] {
					return qs.Filter(orm.F("payment_status").Eq("pending"))
				},
			},
			{
				Name:  "paid_orders",
				Label: "Paid Orders",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Order], value interface{}) orm.QuerySet[Order] {
					return qs.Filter(orm.F("payment_status").Eq("paid"))
				},
			},
			{
				Name:  "unfulfilled_orders",
				Label: "Unfulfilled Orders",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Order], value interface{}) orm.QuerySet[Order] {
					return qs.Filter(orm.F("fulfillment_status").Eq("unfulfilled"))
				},
			},
			{
				Name:  "high_value_orders",
				Label: "High Value Orders ($500+)",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Order], value interface{}) orm.QuerySet[Order] {
					return qs.Filter(orm.F("total").Gte(500))
				},
			},
			{
				Name:  "today_orders",
				Label: "Today's Orders",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Order], value interface{}) orm.QuerySet[Order] {
					return qs.Filter(orm.F("created_at").Gte(time.Now().Add(-24 * time.Hour)))
				},
			},
		},
		Actions: []admin.Action[Order]{
			{
				Name:  "mark_processing",
				Label: "Mark Processing",
				Handler: func(ctx context.Context, instances []*Order) error {
					for _, order := range instances {
						order.Status = "processing"
						if err := OrderObjects.Update(ctx, order); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "mark_shipped",
				Label: "Mark Shipped",
				Handler: func(ctx context.Context, instances []*Order) error {
					for _, order := range instances {
						order.Status = "shipped"
						order.ShippedAt = time.Now()
						if err := OrderObjects.Update(ctx, order); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "mark_delivered",
				Label: "Mark Delivered",
				Handler: func(ctx context.Context, instances []*Order) error {
					for _, order := range instances {
						order.Status = "delivered"
						order.FulfillmentStatus = "fulfilled"
						order.DeliveredAt = time.Now()
						if err := OrderObjects.Update(ctx, order); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "cancel_order",
				Label: "Cancel Order",
				Handler: func(ctx context.Context, instances []*Order) error {
					for _, order := range instances {
						order.Status = "cancelled"
						order.CancelledAt = time.Now()
						if err := OrderObjects.Update(ctx, order); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "mark_paid",
				Label: "Mark as Paid",
				Handler: func(ctx context.Context, instances []*Order) error {
					for _, order := range instances {
						order.PaymentStatus = "paid"
						order.PaidAt = time.Now()
						if err := OrderObjects.Update(ctx, order); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "export_orders",
				Label: "Export Orders",
				Handler: func(ctx context.Context, instances []*Order) error {
					return nil
				},
			},
		},
	})

	// OrderItem admin
	admin.Register(&admin.Config[OrderItem]{
		Icon: "Package",
		ListDisplay: []admin.Field{
			OrderItemFieldsInstance.OrderId,
			OrderItemFieldsInstance.ProductName,
			OrderItemFieldsInstance.Quantity,
			OrderItemFieldsInstance.UnitPrice,
			OrderItemFieldsInstance.Total,
			OrderItemFieldsInstance.FulfillmentStatus,
		},
		ListFilter: []admin.Field{
			OrderItemFieldsInstance.OrderId,
			OrderItemFieldsInstance.FulfillmentStatus,
		},
		SearchFields: []admin.Field{
			OrderItemFieldsInstance.ProductName,
			OrderItemFieldsInstance.ProductSku,
		},
		Fieldsets: []admin.Fieldset[OrderItem]{
			{
				Name: "Order Reference",
				Fields: []string{"order_id"},
			},
			{
				Name: "Product Information",
				Fields: []string{"product_id", "variant_id", "product_name", "product_sku", "variant_name", "variant_sku", "image_url"},
			},
			{
				Name: "Pricing",
				Fields: []string{"quantity", "unit_price", "discount_amount", "tax_amount", "total"},
			},
			{
				Name: "Fulfillment",
				Fields: []string{"quantity_fulfilled", "quantity_refunded", "fulfillment_status"},
			},
			{
				Name:  "Shipping",
				Fields: []string{"weight"},
			},
			{
				Name:  "Timestamps",
				Fields: []string{"created_at", "updated_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"fulfillment_status", "quantity_fulfilled"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "total"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[OrderItem], user interface{}, obj *OrderItem) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[OrderItem], user interface{}, obj *OrderItem) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[OrderItem], user interface{}, obj *OrderItem) bool {
			return true
		},
		Filters: []admin.Filter[OrderItem]{
			{
				Name:  "unfulfilled",
				Label: "Unfulfilled Items",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[OrderItem], value interface{}) orm.QuerySet[OrderItem] {
					return qs.Filter(orm.F("fulfillment_status").Eq("unfulfilled"))
				},
			},
			{
				Name:  "fulfilled",
				Label: "Fulfilled Items",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[OrderItem], value interface{}) orm.QuerySet[OrderItem] {
					return qs.Filter(orm.F("fulfillment_status").Eq("fulfilled"))
				},
			},
			{
				Name:  "partially_fulfilled",
				Label: "Partially Fulfilled",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[OrderItem], value interface{}) orm.QuerySet[OrderItem] {
					return qs.Filter(orm.F("quantity_fulfilled").Gt(0))
				},
			},
		},
		Actions: []admin.Action[OrderItem]{
			{
				Name:  "mark_fulfilled",
				Label: "Mark Fulfilled",
				Handler: func(ctx context.Context, instances []*OrderItem) error {
					for _, item := range instances {
						item.FulfillmentStatus = "fulfilled"
						if err := OrderItemObjects.Update(ctx, item); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "mark_refunded",
				Label: "Mark Refunded",
				Handler: func(ctx context.Context, instances []*OrderItem) error {
					for _, item := range instances {
						item.FulfillmentStatus = "refunded"
						item.QuantityRefunded = item.Quantity
						if err := OrderItemObjects.Update(ctx, item); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// Payment admin
	admin.Register(&admin.Config[Payment]{
		Icon: "CreditCard",
		ListDisplay: []admin.Field{
			PaymentFieldsInstance.OrderId,
			PaymentFieldsInstance.Amount,
			PaymentFieldsInstance.Currency,
			PaymentFieldsInstance.PaymentMethod,
			PaymentFieldsInstance.Status,
			PaymentFieldsInstance.TransactionId,
		},
		ListFilter: []admin.Field{
			PaymentFieldsInstance.Status,
			PaymentFieldsInstance.PaymentMethod,
		},
		SearchFields: []admin.Field{
			PaymentFieldsInstance.TransactionId,
			PaymentFieldsInstance.OrderId,
		},
		Fieldsets: []admin.Fieldset[Payment]{
			{
				Name: "Transaction",
				Fields: []string{"order_id", "transaction_id"},
			},
			{
				Name: "Amount",
				Fields: []string{"amount", "currency"},
			},
			{
				Name: "Payment Method",
				Fields: []string{"payment_method", "payment_gateway", "card_last4", "card_brand"},
			},
			{
				Name: "Status",
				Fields: []string{"status"},
			},
			{
				Name: "Gateway Response",
				Fields: []string{"gateway_response", "failure_reason"},
			},
			{
				Name: "Timestamps",
				Fields: []string{"created_at", "updated_at", "completed_at", "failed_at", "refunded_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "transaction_id"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "completed_at", "failed_at", "refunded_at"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Payment], user interface{}, obj *Payment) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Payment], user interface{}, obj *Payment) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Payment], user interface{}, obj *Payment) bool {
			return false // Payments should not be deleted
		},
		Filters: []admin.Filter[Payment]{
			{
				Name:  "pending_payments",
				Label: "Pending Payments",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Payment], value interface{}) orm.QuerySet[Payment] {
					return qs.Filter(orm.F("status").Eq("pending"))
				},
			},
			{
				Name:  "completed_payments",
				Label: "Completed Payments",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Payment], value interface{}) orm.QuerySet[Payment] {
					return qs.Filter(orm.F("status").Eq("completed"))
				},
			},
			{
				Name:  "failed_payments",
				Label: "Failed Payments",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Payment], value interface{}) orm.QuerySet[Payment] {
					return qs.Filter(orm.F("status").Eq("failed"))
				},
			},
			{
				Name:  "refunded_payments",
				Label: "Refunded Payments",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Payment], value interface{}) orm.QuerySet[Payment] {
					return qs.Filter(orm.F("status").Eq("refunded"))
				},
			},
		},
		Actions: []admin.Action[Payment]{
			{
				Name:  "mark_completed",
				Label: "Mark Completed",
				Handler: func(ctx context.Context, instances []*Payment) error {
					for _, payment := range instances {
						payment.Status = "completed"
						if err := PaymentObjects.Update(ctx, payment); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "mark_failed",
				Label: "Mark Failed",
				Handler: func(ctx context.Context, instances []*Payment) error {
					for _, payment := range instances {
						payment.Status = "failed"
						if err := PaymentObjects.Update(ctx, payment); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "refund",
				Label: "Refund Payment",
				Handler: func(ctx context.Context, instances []*Payment) error {
					for _, payment := range instances {
						payment.Status = "refunded"
						if err := PaymentObjects.Update(ctx, payment); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// Shipment admin
	admin.Register(&admin.Config[Shipment]{
		Icon: "Truck",
		ListDisplay: []admin.Field{
			ShipmentFieldsInstance.OrderId,
			ShipmentFieldsInstance.TrackingNumber,
			ShipmentFieldsInstance.Carrier,
			ShipmentFieldsInstance.Status,
			ShipmentFieldsInstance.ShippedAt,
		},
		ListFilter: []admin.Field{
			ShipmentFieldsInstance.Status,
			ShipmentFieldsInstance.Carrier,
		},
		SearchFields: []admin.Field{
			ShipmentFieldsInstance.TrackingNumber,
			ShipmentFieldsInstance.OrderId,
		},
		Fieldsets: []admin.Fieldset[Shipment]{
			{
				Name: "Shipment Details",
				Fields: []string{"order_id", "tracking_number", "carrier"},
			},
			{
				Name: "Status",
				Fields: []string{"status"},
			},
			{
				Name: "Address",
				Fields: []string{"recipient_name", "recipient_company", "address_line1", "address_line2", "city", "state", "postal_code", "country_code"},
			},
			{
				Name: "Dates",
				Fields: []string{"shipped_at", "delivered_at", "estimated_delivery_at"},
			},
			{
				Name:  "Notes",
				Fields: []string{"notes"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "tracking_number"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "shipped_at", "delivered_at"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Shipment], user interface{}, obj *Shipment) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Shipment], user interface{}, obj *Shipment) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Shipment], user interface{}, obj *Shipment) bool {
			return true
		},
		Filters: []admin.Filter[Shipment]{
			{
				Name:  "pending_shipments",
				Label: "Pending Shipments",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Shipment], value interface{}) orm.QuerySet[Shipment] {
					return qs.Filter(orm.F("status").Eq("pending"))
				},
			},
			{
				Name:  "in_transit",
				Label: "In Transit",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Shipment], value interface{}) orm.QuerySet[Shipment] {
					return qs.Filter(orm.F("status").Eq("in_transit"))
				},
			},
			{
				Name:  "delivered",
				Label: "Delivered",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Shipment], value interface{}) orm.QuerySet[Shipment] {
					return qs.Filter(orm.F("status").Eq("delivered"))
				},
			},
			{
				Name:  "has_tracking",
				Label: "Has Tracking Number",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Shipment], value interface{}) orm.QuerySet[Shipment] {
					return qs.Filter(orm.F("tracking_number").IsNotNull())
				},
			},
		},
		Actions: []admin.Action[Shipment]{
			{
				Name:  "mark_in_transit",
				Label: "Mark In Transit",
				Handler: func(ctx context.Context, instances []*Shipment) error {
					for _, shipment := range instances {
						shipment.Status = "in_transit"
						if err := ShipmentObjects.Update(ctx, shipment); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "mark_delivered",
				Label: "Mark Delivered",
				Handler: func(ctx context.Context, instances []*Shipment) error {
					for _, shipment := range instances {
						shipment.Status = "delivered"
						shipment.DeliveredAt = time.Now()
						if err := ShipmentObjects.Update(ctx, shipment); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "mark_returned",
				Label: "Mark Returned",
				Handler: func(ctx context.Context, instances []*Shipment) error {
					for _, shipment := range instances {
						shipment.Status = "returned"
						if err := ShipmentObjects.Update(ctx, shipment); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}
