package models

import (
	"context"
	"reflect"

	"github.com/forgego/forge/schema"
)

// Order represents a customer order
type Order struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Order
func (Order) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.UUID("order_number").Required().Unique().VerboseName("Order Number").Build(),
		schema.Int64("customer_id").Required().VerboseName("Customer ID").Build(),
		schema.String("status").Required().MaxLength(50).Choices(
			schema.Choice{Value: "pending", Label: "Pending"},
			schema.Choice{Value: "confirmed", Label: "Confirmed"},
			schema.Choice{Value: "processing", Label: "Processing"},
			schema.Choice{Value: "shipped", Label: "Shipped"},
			schema.Choice{Value: "delivered", Label: "Delivered"},
			schema.Choice{Value: "cancelled", Label: "Cancelled"},
			schema.Choice{Value: "refunded", Label: "Refunded"},
		).Default("pending").VerboseName("Order Status").Build(),
		schema.Decimal("subtotal").MaxDigits(12).DecimalPlaces(2).Required().VerboseName("Subtotal").Build(),
		schema.Decimal("tax_amount").MaxDigits(12).DecimalPlaces(2).Default(0.0).VerboseName("Tax Amount").Build(),
		schema.Decimal("shipping_amount").MaxDigits(12).DecimalPlaces(2).Default(0.0).VerboseName("Shipping Amount").Build(),
		schema.Decimal("discount_amount").MaxDigits(12).DecimalPlaces(2).Default(0.0).VerboseName("Discount Amount").Build(),
		schema.Decimal("total_amount").MaxDigits(12).DecimalPlaces(2).Required().VerboseName("Total Amount").Build(),
		schema.String("currency").MaxLength(3).Default("USD").VerboseName("Currency").Build(),
		schema.Int64("billing_address_id").Required().VerboseName("Billing Address ID").Build(),
		schema.Int64("shipping_address_id").Required().VerboseName("Shipping Address ID").Build(),
		schema.String("payment_status").MaxLength(50).Choices(
			schema.Choice{Value: "pending", Label: "Pending"},
			schema.Choice{Value: "paid", Label: "Paid"},
			schema.Choice{Value: "failed", Label: "Failed"},
			schema.Choice{Value: "refunded", Label: "Refunded"},
		).Default("pending").VerboseName("Payment Status").Build(),
		schema.String("shipping_status").MaxLength(50).Choices(
			schema.Choice{Value: "pending", Label: "Pending"},
			schema.Choice{Value: "processing", Label: "Processing"},
			schema.Choice{Value: "shipped", Label: "Shipped"},
			schema.Choice{Value: "delivered", Label: "Delivered"},
			schema.Choice{Value: "returned", Label: "Returned"},
		).Default("pending").VerboseName("Shipping Status").Build(),
		schema.Text("notes").Optional().VerboseName("Order Notes").Build(),
		schema.Text("customer_notes").Optional().VerboseName("Customer Notes").Build(),
		schema.JSON("metadata").Optional().VerboseName("Additional Metadata").Build(),
		schema.DateTime("placed_at").AutoNowAdd().VerboseName("Placed At").Build(),
		schema.DateTime("confirmed_at").Optional().VerboseName("Confirmed At").Build(),
		schema.DateTime("shipped_at").Optional().VerboseName("Shipped At").Build(),
		schema.DateTime("delivered_at").Optional().VerboseName("Delivered At").Build(),
		schema.DateTime("cancelled_at").Optional().VerboseName("Cancelled At").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (Order) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "orders",
		VerboseName:       "Order",
		VerboseNamePlural: "Orders",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			{Name: "idx_order_order_number", Fields: []string{"order_number"}, Unique: true},
			{Name: "idx_order_customer_id", Fields: []string{"customer_id"}, Unique: false},
			{Name: "idx_order_status", Fields: []string{"status"}, Unique: false},
			{Name: "idx_order_payment_status", Fields: []string{"payment_status"}, Unique: false},
			{Name: "idx_order_shipping_status", Fields: []string{"shipping_status"}, Unique: false},
			{Name: "idx_order_placed_at", Fields: []string{"placed_at"}, Unique: false},
		},
	}
}

// Relations returns all relationship definitions
func (Order) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("customer_id", "Customer").Required().OnDelete(schema.CascadePROTECT).RelatedName("orders").Build(),
		schema.ForeignKey("billing_address_id", "Address").Required().OnDelete(schema.CascadePROTECT).RelatedName("billing_orders").Build(),
		schema.ForeignKey("shipping_address_id", "Address").Required().OnDelete(schema.CascadePROTECT).RelatedName("shipping_orders").Build(),
		schema.OneToMany("items", "OrderItem", "order_id").CascadeOnDelete().Build(),
		schema.OneToMany("payments", "Payment", "order_id").CascadeOnDelete().Build(),
		schema.OneToOne("shipping", "Shipping").Optional().OnDelete(schema.CascadeCASCADE).RelatedName("order").Build(),
	}
}

// Hooks returns model lifecycle hooks
// Note: These hooks use reflection to work before code generation.
// After code generation, you can update them to use direct field access for better performance.
func (Order) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Calculate total from items if not set
			// Uses reflection to work before/after code generation
			val := reflect.ValueOf(instance).Elem()

			// Get total_amount field
			totalAmountField := val.FieldByName("TotalAmount")
			if !totalAmountField.IsValid() {
				return nil // Field doesn't exist yet (before code generation)
			}

			// Check if total is zero
			if totalAmountField.MethodByName("IsZero").Call(nil)[0].Bool() {
				// Calculate from subtotal + tax + shipping - discount
				subtotal := val.FieldByName("Subtotal")
				taxAmount := val.FieldByName("TaxAmount")
				shippingAmount := val.FieldByName("ShippingAmount")
				discountAmount := val.FieldByName("DiscountAmount")

				if subtotal.IsValid() {
					// Use reflection to call Add/Sub methods on decimal.Decimal
					total := reflect.ValueOf(subtotal.Interface())
					if taxAmount.IsValid() {
						total = total.MethodByName("Add").Call([]reflect.Value{taxAmount})[0]
					}
					if shippingAmount.IsValid() {
						total = total.MethodByName("Add").Call([]reflect.Value{shippingAmount})[0]
					}
					if discountAmount.IsValid() {
						total = total.MethodByName("Sub").Call([]reflect.Value{discountAmount})[0]
					}
					totalAmountField.Set(total)
				}
			}

			return nil
		},
		BeforeCreate: func(ctx context.Context, instance interface{}) error {
			// Order number and placed_at are handled by field defaults (UUID and AutoNowAdd)
			return nil
		},
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// After order is created, we could send notifications, update customer stats, etc.
			// This will be implemented after code generation when we can access related models
			return nil
		},
		BeforeUpdate: func(ctx context.Context, instance interface{}) error {
			// Recalculate totals if needed
			val := reflect.ValueOf(instance).Elem()
			totalAmountField := val.FieldByName("TotalAmount")
			if !totalAmountField.IsValid() {
				return nil
			}

			if totalAmountField.MethodByName("IsZero").Call(nil)[0].Bool() {
				subtotal := val.FieldByName("Subtotal")
				if subtotal.IsValid() {
					total := reflect.ValueOf(subtotal.Interface())
					if taxAmount := val.FieldByName("TaxAmount"); taxAmount.IsValid() {
						total = total.MethodByName("Add").Call([]reflect.Value{taxAmount})[0]
					}
					if shippingAmount := val.FieldByName("ShippingAmount"); shippingAmount.IsValid() {
						total = total.MethodByName("Add").Call([]reflect.Value{shippingAmount})[0]
					}
					if discountAmount := val.FieldByName("DiscountAmount"); discountAmount.IsValid() {
						total = total.MethodByName("Sub").Call([]reflect.Value{discountAmount})[0]
					}
					totalAmountField.Set(total)
				}
			}

			return nil
		},
	}
}
