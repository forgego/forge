package orders

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers order API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("carts", &api.ViewSetConfig{
		Model:      &Cart{},
		Queryset:   CartObjects,
		Serializer: base,
	})

	router.Register("cart-items", &api.ViewSetConfig{
		Model:      &CartItem{},
		Queryset:   CartItemObjects,
		Serializer: base,
	})

	router.Register("orders", &api.ViewSetConfig{
		Model:      &Order{},
		Queryset:   OrderObjects,
		Serializer: base,
	})

	router.Register("order-items", &api.ViewSetConfig{
		Model:      &OrderItem{},
		Queryset:   OrderItemObjects,
		Serializer: base,
	})

	router.Register("payments", &api.ViewSetConfig{
		Model:      &Payment{},
		Queryset:   PaymentObjects,
		Serializer: base,
	})

	router.Register("shipments", &api.ViewSetConfig{
		Model:      &Shipment{},
		Queryset:   ShipmentObjects,
		Serializer: base,
	})

	// GET /orders/summary - Aggregated orders statistics
	router.Get("/orders/summary", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var totalOrders, pendingOrders, processingOrders, shippedOrders, deliveredOrders, cancelledOrders int64
		var totalRevenue float64

		row := database.QueryRowContext(ctx, `
			SELECT
				COUNT(*),
				COALESCE(SUM(CASE WHEN payment_status = 'paid' OR status = 'delivered' THEN total ELSE 0 END), 0.0),
				COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN status = 'processing' THEN 1 ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN status = 'shipped' THEN 1 ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END), 0)
			FROM orders
		`)
		if err := row.Scan(&totalOrders, &totalRevenue, &pendingOrders, &processingOrders, &shippedOrders, &deliveredOrders, &cancelledOrders); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"total_orders":      totalOrders,
			"total_revenue":     totalRevenue,
			"pending_orders":    pendingOrders,
			"processing_orders": processingOrders,
			"shipped_orders":    shippedOrders,
			"delivered_orders":  deliveredOrders,
			"cancelled_orders":  cancelledOrders,
		})
	})

	// POST /orders/checkout - Programmatic order creation showcasing Forge ORM Create and hooks
	type CheckoutItemInput struct {
		ProductID   int64   `json:"product_id"`
		ProductName string  `json:"product_name"`
		SKU         string  `json:"sku"`
		Quantity    int32   `json:"quantity"`
		UnitPrice   float64 `json:"unit_price"`
	}

	type CheckoutInput struct {
		CustomerID        int64               `json:"customer_id"`
		CustomerEmail     string              `json:"customer_email"`
		CustomerFirstName string              `json:"customer_first_name"`
		CustomerLastName  string              `json:"customer_last_name"`
		CustomerPhone     string              `json:"customer_phone"`
		ShippingCity      string              `json:"shipping_city"`
		ShippingState     string              `json:"shipping_state"`
		Items             []CheckoutItemInput `json:"items"`
	}

	router.Post("/orders/checkout", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var input CheckoutInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if len(input.Items) == 0 {
			http.Error(w, "order must contain at least 1 item", http.StatusBadRequest)
			return
		}

		var subtotal float64
		for _, it := range input.Items {
			qty := it.Quantity
			if qty <= 0 {
				qty = 1
			}
			subtotal += it.UnitPrice * float64(qty)
		}
		tax := subtotal * 0.08
		shipping := 12.50

		email := input.CustomerEmail
		if email == "" {
			email = "shopper@example.com"
		}
		fn := input.CustomerFirstName
		if fn == "" {
			fn = "Jane"
		}
		ln := input.CustomerLastName
		if ln == "" {
			ln = "Doe"
		}

		custID := input.CustomerID
		if custID <= 0 {
			custID = 1
		}

		order := Order{
			OrderGenerated: OrderGenerated{
				CustomerId:        custID,
				CustomerEmail:     email,
				CustomerFirstName: fn,
				CustomerLastName:  ln,
				CustomerPhone:     input.CustomerPhone,
				Subtotal:          subtotal,
				TaxAmount:         tax,
				ShippingAmount:    shipping,
				ShippingCity:      input.ShippingCity,
				ShippingState:     input.ShippingState,
				Status:            "pending",
				PaymentStatus:     "pending",
				FulfillmentStatus: "unfulfilled",
			},
		}

		// Order.Hooks().BeforeCreate executes here, populating OrderNumber, Total, and Expiry!
		if err := OrderObjects.Create(ctx, &order); err != nil {
			http.Error(w, "failed to create order: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert order items
		for _, it := range input.Items {
			qty := it.Quantity
			if qty <= 0 {
				qty = 1
			}
			lineTotal := it.UnitPrice * float64(qty)
			name := it.ProductName
			if name == "" {
				name = "Product Item"
			}
			sku := it.SKU
			if sku == "" {
				sku = "SKU-ITEM"
			}

			_, _ = database.ExecContext(ctx, `
				INSERT INTO order_items (order_id, product_id, product_name, product_sku, quantity, unit_price, total, fulfillment_status)
				VALUES (?, ?, ?, ?, ?, ?, ?, 'unfulfilled')
			`, order.Id, it.ProductID, name, sku, qty, it.UnitPrice, lineTotal)

			// Deduct inventory stock
			if it.ProductID > 0 {
				_, _ = database.ExecContext(ctx, `
					UPDATE products SET stock_quantity = MAX(0, stock_quantity - ?) WHERE id = ?
				`, qty, it.ProductID)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":       "success",
			"message":      "Order placed successfully",
			"order_id":     order.Id,
			"order_number": order.OrderNumber,
			"order_status": order.Status,
			"subtotal":     order.Subtotal,
			"tax":          order.TaxAmount,
			"shipping":     order.ShippingAmount,
			"total":        order.Total,
			"total_amount": order.Total,
		})
	})
}
