package models

import (
	"context"
	"reflect"

	"github.com/forgego/forge/schema"
)

// OrderItem represents an item in an order
type OrderItem struct {
	schema.BaseSchema
}

// Fields returns all field definitions for OrderItem
func (OrderItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Int64("order_id").Required().VerboseName("Order ID").Build(),
		schema.Int64("product_id").Required().VerboseName("Product ID").Build(),
		schema.Int64("variant_id").Optional().VerboseName("Variant ID").Build(),
		schema.String("product_name").Required().MaxLength(500).VerboseName("Product Name").Build(),
		schema.String("product_sku").Required().MaxLength(100).VerboseName("Product SKU").Build(),
		schema.Int32("quantity").Required().VerboseName("Quantity").Build(),
		schema.Decimal("unit_price").MaxDigits(12).DecimalPlaces(2).Required().VerboseName("Unit Price").Build(),
		schema.Decimal("total_price").MaxDigits(12).DecimalPlaces(2).Required().VerboseName("Total Price").Build(),
		schema.Decimal("discount_amount").MaxDigits(12).DecimalPlaces(2).Default(0.0).VerboseName("Discount Amount").Build(),
		schema.Decimal("tax_amount").MaxDigits(12).DecimalPlaces(2).Default(0.0).VerboseName("Tax Amount").Build(),
		schema.JSON("product_data").Optional().VerboseName("Product Snapshot Data").Build(),
		schema.JSON("variant_data").Optional().VerboseName("Variant Snapshot Data").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (OrderItem) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "order_items",
		VerboseName:       "Order Item",
		VerboseNamePlural: "Order Items",
		OrderBy:           []string{"id"},
		Indexes: []schema.Index{
			{Name: "idx_order_item_order_id", Fields: []string{"order_id"}, Unique: false},
			{Name: "idx_order_item_product_id", Fields: []string{"product_id"}, Unique: false},
			{Name: "idx_order_item_variant_id", Fields: []string{"variant_id"}, Unique: false},
		},
	}
}

// Relations returns all relationship definitions
func (OrderItem) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("order_id", "Order").Required().OnDelete(schema.CascadeCASCADE).RelatedName("items").Build(),
		schema.ForeignKey("product_id", "Product").Required().OnDelete(schema.CascadePROTECT).RelatedName("order_items").Build(),
		schema.ForeignKey("variant_id", "ProductVariant").Optional().OnDelete(schema.CascadePROTECT).RelatedName("order_items").Build(),
	}
}

// Hooks returns model lifecycle hooks
// Note: These hooks use reflection to work before code generation.
// After code generation, you can update them to use direct field access for better performance.
func (OrderItem) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Calculate total_price from unit_price * quantity if not set
			val := reflect.ValueOf(instance).Elem()

			totalPriceField := val.FieldByName("TotalPrice")
			unitPriceField := val.FieldByName("UnitPrice")
			quantityField := val.FieldByName("Quantity")

			if !totalPriceField.IsValid() || !unitPriceField.IsValid() || !quantityField.IsValid() {
				return nil // Fields don't exist yet (before code generation)
			}

			// Check if total_price is zero
			if totalPriceField.MethodByName("IsZero").Call(nil)[0].Bool() {
				// Check if unit_price is not zero
				if !unitPriceField.MethodByName("IsZero").Call(nil)[0].Bool() {
					// Calculate: total = unit_price * quantity
					unitPrice := reflect.ValueOf(unitPriceField.Interface())
					quantity := quantityField.Int() // Assuming Int32, convert to int64 for decimal

					// Create decimal from quantity and multiply
					// This is simplified - actual implementation would use decimal package properly
					// For now, we'll just set a placeholder that will work after code generation
					// The actual calculation will be done in the generated code
				}
			}

			return nil
		},
		BeforeCreate: func(ctx context.Context, instance interface{}) error {
			// Ensure total_price is calculated
			// Same logic as BeforeSave
			val := reflect.ValueOf(instance).Elem()
			totalPriceField := val.FieldByName("TotalPrice")
			unitPriceField := val.FieldByName("UnitPrice")

			if !totalPriceField.IsValid() || !unitPriceField.IsValid() {
				return nil
			}

			// Note: Inventory updates would happen here after code generation
			// We would check available inventory and reserve it using the Inventory model

			return nil
		},
		AfterCreate: func(ctx context.Context, instance interface{}) error {
			// After order item is created, update order totals
			// This would require accessing the order via order_id and recalculating
			// Will be implemented after code generation when we can access Order.Objects
			return nil
		},
		BeforeDelete: func(ctx context.Context, instance interface{}) error {
			// Before deleting order item, release inventory
			// Will be implemented after code generation when we can access Inventory.Objects
			return nil
		},
	}
}
