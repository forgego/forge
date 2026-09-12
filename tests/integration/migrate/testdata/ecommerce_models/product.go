package models

import (
	"github.com/forgego/forge/schema"
)

// Product represents a product in the ecommerce system
type Product struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Product
func (Product) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.UUID("sku").Required().Unique().VerboseName("SKU").Build(),
		schema.String("name").Required().MaxLength(500).VerboseName("Product Name").Build(),
		schema.String("slug").Required().Unique().MaxLength(500).DBIndex().VerboseName("URL Slug").Build(),
		schema.Text("description").Optional().VerboseName("Description").Build(),
		schema.Text("short_description").Optional().MaxLength(500).VerboseName("Short Description").Build(),
		schema.Decimal("price").MaxDigits(12).DecimalPlaces(2).Required().VerboseName("Price").Build(),
		schema.Decimal("compare_at_price").MaxDigits(12).DecimalPlaces(2).Optional().VerboseName("Compare At Price").Build(),
		schema.Decimal("cost_price").MaxDigits(12).DecimalPlaces(2).Optional().VerboseName("Cost Price").Build(),
		schema.String("currency").MaxLength(3).Default("USD").VerboseName("Currency").Build(),
		schema.Int32("stock_quantity").Default(0).VerboseName("Stock Quantity").Build(),
		schema.Int32("low_stock_threshold").Default(10).VerboseName("Low Stock Threshold").Build(),
		schema.Bool("track_inventory").Default(true).VerboseName("Track Inventory").Build(),
		schema.Bool("is_active").Default(true).VerboseName("Is Active").Build(),
		schema.Bool("is_featured").Default(false).VerboseName("Is Featured").Build(),
		schema.Bool("is_digital").Default(false).VerboseName("Is Digital Product").Build(),
		schema.Bool("requires_shipping").Default(true).VerboseName("Requires Shipping").Build(),
		schema.Float64("weight").Optional().VerboseName("Weight (kg)").Build(),
		schema.Float64("length").Optional().VerboseName("Length (cm)").Build(),
		schema.Float64("width").Optional().VerboseName("Width (cm)").Build(),
		schema.Float64("height").Optional().VerboseName("Height (cm)").Build(),
		schema.String("status").MaxLength(50).Choices(
			schema.Choice{Value: "draft", Label: "Draft"},
			schema.Choice{Value: "active", Label: "Active"},
			schema.Choice{Value: "archived", Label: "Archived"},
			schema.Choice{Value: "discontinued", Label: "Discontinued"},
		).Default("draft").VerboseName("Status").Build(),
		schema.JSON("attributes").Optional().VerboseName("Product Attributes").Build(),
		schema.JSON("images").Optional().VerboseName("Product Images").Build(),
		schema.JSON("seo_data").Optional().VerboseName("SEO Data").Build(),
		schema.Int64("brand_id").Optional().VerboseName("Brand ID").Build(),
		schema.Int64("supplier_id").Optional().VerboseName("Supplier ID").Build(),
		schema.DateTime("published_at").Optional().VerboseName("Published At").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (Product) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "products",
		VerboseName:       "Product",
		VerboseNamePlural: "Products",
		OrderBy:           []string{"-created_at", "name"},
		Indexes: []schema.Index{
			{Name: "idx_product_sku", Fields: []string{"sku"}, Unique: true},
			{Name: "idx_product_slug", Fields: []string{"slug"}, Unique: true},
			{Name: "idx_product_status", Fields: []string{"status"}, Unique: false},
			{Name: "idx_product_is_active", Fields: []string{"is_active"}, Unique: false},
			{Name: "idx_product_is_featured", Fields: []string{"is_featured"}, Unique: false},
			{Name: "idx_product_brand_id", Fields: []string{"brand_id"}, Unique: false},
			{Name: "idx_product_price", Fields: []string{"price"}, Unique: false},
		},
		UniqueTogether: [][]string{
			{"sku"},
			{"slug"},
		},
	}
}

// Relations returns all relationship definitions
func (Product) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("brand_id", "Brand").Optional().OnDelete(schema.CascadeSET_NULL).RelatedName("products").Build(),
		schema.ForeignKey("supplier_id", "Supplier").Optional().OnDelete(schema.CascadeSET_NULL).RelatedName("products").Build(),
		schema.ManyToMany("categories", "Category").RelatedName("products").Build(),
		schema.OneToMany("variants", "ProductVariant", "product_id").CascadeOnDelete().Build(),
		schema.OneToMany("inventory_items", "Inventory", "product_id").CascadeOnDelete().Build(),
		schema.OneToMany("reviews", "Review", "product_id").CascadeOnDelete().Build(),
	}
}

// Hooks returns model lifecycle hooks
func (Product) Hooks() *schema.ModelHooks {
	return nil
}
