package models

import (
	"github.com/forgego/forge/schema"
)

// Payment represents a payment for an order
type Payment struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Payment
func (Payment) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Int64("order_id").Required().VerboseName("Order ID").Build(),
		schema.UUID("transaction_id").Required().Unique().VerboseName("Transaction ID").Build(),
		schema.String("payment_method").Required().MaxLength(50).Choices(
			schema.Choice{Value: "credit_card", Label: "Credit Card"},
			schema.Choice{Value: "debit_card", Label: "Debit Card"},
			schema.Choice{Value: "paypal", Label: "PayPal"},
			schema.Choice{Value: "stripe", Label: "Stripe"},
			schema.Choice{Value: "bank_transfer", Label: "Bank Transfer"},
			schema.Choice{Value: "cash_on_delivery", Label: "Cash on Delivery"},
			schema.Choice{Value: "gift_card", Label: "Gift Card"},
		).VerboseName("Payment Method").Build(),
		schema.String("status").Required().MaxLength(50).Choices(
			schema.Choice{Value: "pending", Label: "Pending"},
			schema.Choice{Value: "processing", Label: "Processing"},
			schema.Choice{Value: "completed", Label: "Completed"},
			schema.Choice{Value: "failed", Label: "Failed"},
			schema.Choice{Value: "refunded", Label: "Refunded"},
			schema.Choice{Value: "cancelled", Label: "Cancelled"},
		).Default("pending").VerboseName("Payment Status").Build(),
		schema.Decimal("amount").MaxDigits(12).DecimalPlaces(2).Required().VerboseName("Amount").Build(),
		schema.String("currency").MaxLength(3).Default("USD").VerboseName("Currency").Build(),
		schema.String("gateway").MaxLength(100).VerboseName("Payment Gateway").Build(),
		schema.String("gateway_transaction_id").MaxLength(255).VerboseName("Gateway Transaction ID").Build(),
		schema.String("gateway_response").MaxLength(50).VerboseName("Gateway Response Code").Build(),
		schema.Text("gateway_message").Optional().VerboseName("Gateway Message").Build(),
		schema.JSON("gateway_data").Optional().VerboseName("Gateway Response Data").Build(),
		schema.String("card_last4").MaxLength(4).VerboseName("Card Last 4 Digits").Build(),
		schema.String("card_brand").MaxLength(50).VerboseName("Card Brand").Build(),
		schema.Date("card_expiry").Optional().VerboseName("Card Expiry Date").Build(),
		schema.Bool("is_refunded").Default(false).VerboseName("Is Refunded").Build(),
		schema.Decimal("refund_amount").MaxDigits(12).DecimalPlaces(2).Default(0.0).VerboseName("Refund Amount").Build(),
		schema.DateTime("paid_at").Optional().VerboseName("Paid At").Build(),
		schema.DateTime("refunded_at").Optional().VerboseName("Refunded At").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (Payment) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "payments",
		VerboseName:       "Payment",
		VerboseNamePlural: "Payments",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			{Name: "idx_payment_order_id", Fields: []string{"order_id"}, Unique: false},
			{Name: "idx_payment_transaction_id", Fields: []string{"transaction_id"}, Unique: true},
			{Name: "idx_payment_status", Fields: []string{"status"}, Unique: false},
			{Name: "idx_payment_gateway_transaction_id", Fields: []string{"gateway_transaction_id"}, Unique: false},
		},
	}
}

// Relations returns all relationship definitions
func (Payment) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("order_id", "Order").Required().OnDelete(schema.CascadeCASCADE).RelatedName("payments").Build(),
	}
}

// Hooks returns model lifecycle hooks
func (Payment) Hooks() *schema.ModelHooks {
	return nil
}
