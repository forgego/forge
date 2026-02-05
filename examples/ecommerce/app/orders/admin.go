package orders

import (
	"context"

	"github.com/forgego/forge/admin"
)

// RegisterAdmin registers order models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// Cart admin
	admin.Register(&admin.Config[Cart]{
		Icon: "ShoppingCart",
		ListDisplay: []admin.Field{
			CartFieldsInstance.Id,
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
		},
	})
}
