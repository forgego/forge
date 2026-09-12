package models

import (
	"github.com/forgego/forge/schema"
)

// ProductVariant represents a variant of a product (size, color, etc.)
type ProductVariant struct {
	schema.BaseSchema
}

// Fields returns all field definitions for ProductVariant
func (ProductVariant) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Int64("product_id").Required().VerboseName("Product ID").Build(),
		schema.UUID("sku").Required().Unique().VerboseName("Variant SKU").Build(),
		schema.String("name").Required().MaxLength(200).VerboseName("Variant Name").Build(),
		schema.String("option1").MaxLength(100).VerboseName("Option 1 (e.g., Size)").Build(),
		schema.String("option2").MaxLength(100).VerboseName("Option 2 (e.g., Color)").Build(),
		schema.String("option3").MaxLength(100).VerboseName("Option 3").Build(),
		schema.Decimal("price").MaxDigits(12).DecimalPlaces(2).Optional().VerboseName("Variant Price").Build(),
		schema.Decimal("compare_at_price").MaxDigits(12).DecimalPlaces(2).Optional().VerboseName("Compare At Price").Build(),
		schema.Int32("stock_quantity").Default(0).VerboseName("Stock Quantity").Build(),
		schema.Bool("is_active").Default(true).VerboseName("Is Active").Build(),
		schema.Bool("is_default").Default(false).VerboseName("Is Default Variant").Build(),
		schema.URL("image_url").Optional().MaxLength(500).VerboseName("Variant Image URL").Build(),
		schema.Float64("weight").Optional().VerboseName("Weight (kg)").Build(),
		schema.JSON("metadata").Optional().VerboseName("Additional Metadata").Build(),
		schema.Int32("position").Default(0).VerboseName("Position").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (ProductVariant) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "product_variants",
		VerboseName:       "Product Variant",
		VerboseNamePlural: "Product Variants",
		OrderBy:           []string{"position", "name"},
		Indexes: []schema.Index{
			{Name: "idx_variant_product_id", Fields: []string{"product_id"}, Unique: false},
			{Name: "idx_variant_sku", Fields: []string{"sku"}, Unique: true},
			{Name: "idx_variant_is_active", Fields: []string{"is_active"}, Unique: false},
		},
		UniqueTogether: [][]string{
			{"product_id", "option1", "option2", "option3"},
		},
	}
}

// Relations returns all relationship definitions
func (ProductVariant) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("product_id", "Product").Required().OnDelete(schema.CascadeCASCADE).RelatedName("variants").Build(),
	}
}

// Hooks returns model lifecycle hooks
func (ProductVariant) Hooks() *schema.ModelHooks {
	return nil
}
