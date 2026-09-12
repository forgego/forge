package models

import (
	"github.com/forgego/forge/schema"
)

// Inventory represents inventory tracking for products
type Inventory struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Inventory
func (Inventory) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Int64("product_id").Required().VerboseName("Product ID").Build(),
		schema.Int64("variant_id").Optional().VerboseName("Variant ID").Build(),
		schema.Int64("warehouse_id").Required().VerboseName("Warehouse ID").Build(),
		schema.Int32("quantity").Default(0).VerboseName("Quantity").Build(),
		schema.Int32("reserved_quantity").Default(0).VerboseName("Reserved Quantity").Build(),
		schema.Int32("available_quantity").Default(0).VerboseName("Available Quantity").Build(),
		schema.Decimal("cost_per_unit").MaxDigits(12).DecimalPlaces(2).Optional().VerboseName("Cost Per Unit").Build(),
		schema.String("location").MaxLength(100).VerboseName("Warehouse Location").Build(),
		schema.Bool("is_active").Default(true).VerboseName("Is Active").Build(),
		schema.DateTime("last_restocked_at").Optional().VerboseName("Last Restocked At").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (Inventory) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "inventory",
		VerboseName:       "Inventory",
		VerboseNamePlural: "Inventory Items",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			{Name: "idx_inventory_product_id", Fields: []string{"product_id"}, Unique: false},
			{Name: "idx_inventory_variant_id", Fields: []string{"variant_id"}, Unique: false},
			{Name: "idx_inventory_warehouse_id", Fields: []string{"warehouse_id"}, Unique: false},
			{Name: "idx_inventory_available_quantity", Fields: []string{"available_quantity"}, Unique: false},
		},
		UniqueTogether: [][]string{
			{"product_id", "variant_id", "warehouse_id"},
		},
	}
}

// Relations returns all relationship definitions
func (Inventory) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("product_id", "Product").Required().OnDelete(schema.CascadeCASCADE).RelatedName("inventory_items").Build(),
		schema.ForeignKey("variant_id", "ProductVariant").Optional().OnDelete(schema.CascadeCASCADE).RelatedName("inventory").Build(),
		schema.ForeignKey("warehouse_id", "Warehouse").Required().OnDelete(schema.CascadePROTECT).RelatedName("inventory").Build(),
	}
}

// Hooks returns model lifecycle hooks
func (Inventory) Hooks() *schema.ModelHooks {
	return nil
}
