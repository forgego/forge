package catalog

import (
	"context"

	"github.com/forgego/forge/registry"
	"github.com/forgego/forge/schema"
)

// Category represents a product category with hierarchical support
type Category struct {
	schema.BaseSchema
	Id          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Slug        string `json:"slug" db:"slug"`
	Description string `json:"description" db:"description"`
	ParentID    int64  `json:"parent_id" db:"parent_id"`
	ImageURL    string `json:"image_url" db:"image_url"`
	SortOrder   int32  `json:"sort_order" db:"sort_order"`
	IsActive    bool   `json:"is_active" db:"is_active"`
	Level       int32  `json:"level" db:"level"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

func (Category) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(200), schema.Unique(),
			schema.HelpText("Category name")),
		schema.StringField("slug", schema.Required(), schema.MaxLength(200), schema.Unique(),
			schema.HelpText("URL-friendly identifier")),
		schema.TextField("description", schema.Optional(),
			schema.HelpText("Category description")),
		schema.Int64Field("parent_id", schema.Optional(),
			schema.VerboseName("Parent Category"),
			schema.HelpText("Parent category for hierarchy")),
		schema.StringField("image_url", schema.MaxLength(500), schema.Optional(),
			schema.HelpText("Category image")),
		schema.Int32Field("sort_order", schema.Default(0),
			schema.VerboseName("Sort Order"),
			schema.HelpText("Display order")),
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active"),
			schema.HelpText("Is category active")),
		schema.Int32Field("level", schema.Default(0),
			schema.VerboseName("Level"),
			schema.HelpText("Hierarchy level (0=root)")),
		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Category) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "categories",
		VerboseName:       "Category",
		VerboseNamePlural: "Categories",
		OrderBy:           []string{"sort_order", "name"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_category_slug", "slug"),
			schema.IndexOn("idx_category_parent", "parent_id"),
			schema.IndexOn("idx_category_active", "is_active"),
		},
	}
}

func (Category) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("parent_id", "Category",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("children")),
	}
}

func (Category) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Auto-generate slug from name if not provided
			// Calculate hierarchy level
			return nil
		},
	}
}

// Brand represents a product brand
type Brand struct {
	BrandGenerated
}

func (Brand) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(200), schema.Unique()),
		schema.StringField("slug", schema.Required(), schema.MaxLength(200), schema.Unique()),
		schema.TextField("description", schema.Optional()),
		schema.StringField("logo_url", schema.MaxLength(500), schema.Optional()),
		schema.StringField("website_url", schema.MaxLength(500), schema.Optional(),
			schema.VerboseName("Website URL")),
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Brand) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "brands",
		VerboseName:       "Brand",
		VerboseNamePlural: "Brands",
		OrderBy:           []string{"name"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_brand_slug", "slug"),
		},
	}
}

func (Brand) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Brand) Hooks() *schema.ModelHooks {
	return nil
}

// Product represents a product in the catalog
type Product struct {
	ProductGenerated
}

func (Product) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(255),
			schema.HelpText("Product name")),
		schema.StringField("slug", schema.Required(), schema.MaxLength(255), schema.Unique(),
			schema.HelpText("URL-friendly identifier")),
		schema.StringField("sku", schema.Required(), schema.MaxLength(100), schema.Unique(),
			schema.VerboseName("SKU"),
			schema.HelpText("Stock Keeping Unit")),
		schema.TextField("description", schema.Required(),
			schema.HelpText("Full product description")),
		schema.TextField("short_description", schema.Optional(),
			schema.HelpText("Brief description for listings")),

		// Relationships
		schema.Int64Field("category_id", schema.Required(),
			schema.VerboseName("Category"),
			schema.HelpText("Product category")),
		schema.Int64Field("brand_id", schema.Optional(),
			schema.VerboseName("Brand"),
			schema.HelpText("Product brand")),

		// Pricing
		schema.Float64Field("price", schema.Required(),
			schema.HelpText("Base price")),
		schema.Float64Field("cost_price", schema.Optional(),
			schema.HelpText("Cost to seller")),
		schema.Float64Field("compare_at_price", schema.Optional(),
			schema.HelpText("Original price for discounts")),

		// Inventory (base level)
		schema.Int32Field("stock_quantity", schema.Default(0),
			schema.VerboseName("Stock Quantity"),
			schema.HelpText("Total stock across all warehouses")),
		schema.BoolField("track_inventory", schema.Default(true),
			schema.HelpText("Whether to track inventory")),
		schema.BoolField("allow_backorder", schema.Default(false),
			schema.HelpText("Allow orders when out of stock")),

		// Physical attributes
		schema.Float64Field("weight", schema.Optional(),
			schema.HelpText("Weight in kg")),
		schema.Float64Field("length", schema.Optional(),
			schema.HelpText("Length in cm")),
		schema.Float64Field("width", schema.Optional(),
			schema.HelpText("Width in cm")),
		schema.Float64Field("height", schema.Optional(),
			schema.HelpText("Height in cm")),

		// Status
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.BoolField("is_featured", schema.Default(false),
			schema.VerboseName("Featured")),
		schema.BoolField("is_digital", schema.Default(false),
			schema.HelpText("Digital product (no shipping)")),

		// SEO
		schema.StringField("meta_title", schema.MaxLength(255), schema.Optional()),
		schema.TextField("meta_description", schema.Optional()),
		schema.StringField("meta_keywords", schema.MaxLength(500), schema.Optional()),

		// Stats
		schema.Int32Field("view_count", schema.Default(0)),
		schema.Int32Field("order_count", schema.Default(0)),
		schema.Float64Field("rating_average", schema.Default(0.0)),
		schema.Int32Field("rating_count", schema.Default(0)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd(),
			schema.VerboseName("Created At")),
		schema.TimeField("updated_at", schema.AutoNow()),
		schema.TimeField("published_at", schema.Optional()),
	}
}

func (Product) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "products",
		VerboseName:       "Product",
		VerboseNamePlural: "Products",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_product_slug", "slug"),
			schema.UniqueIndexOn("idx_product_sku", "sku"),
			schema.IndexOn("idx_product_category", "category_id"),
			schema.IndexOn("idx_product_brand", "brand_id"),
			schema.IndexOn("idx_product_active", "is_active"),
			schema.IndexOn("idx_product_featured", "is_featured"),
			schema.IndexOn("idx_product_price", "price"),
		},
	}
}

func (Product) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("category_id", "Category",
			schema.OnDelete(schema.CascadePROTECT),
			schema.RelatedName("products")),
		schema.ForeignKeyField("brand_id", "Brand",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("products")),
	}
}

func (Product) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Auto-generate slug if not provided
			// Validate price > 0
			// Set published_at if becoming active
			return nil
		},
		AfterSave: func(ctx context.Context, instance interface{}) error {
			// Update search index
			// Clear cache
			return nil
		},
	}
}

// ProductVariant represents a product variant (size, color, etc.)
type ProductVariant struct {
	ProductVariantGenerated
}

func (ProductVariant) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("product_id", schema.Required(),
			schema.VerboseName("Product")),

		// Variant identification
		schema.StringField("sku", schema.Required(), schema.MaxLength(100), schema.Unique(),
			schema.VerboseName("SKU"),
			schema.HelpText("Unique SKU for this variant")),
		schema.StringField("name", schema.Required(), schema.MaxLength(255),
			schema.HelpText("Variant name (e.g., 'Large Red')")),

		// Variant options
		schema.StringField("option1_name", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("First option name (e.g., 'Size')")),
		schema.StringField("option1_value", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("First option value (e.g., 'Large')")),
		schema.StringField("option2_name", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("Second option name (e.g., 'Color')")),
		schema.StringField("option2_value", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("Second option value (e.g., 'Red')")),
		schema.StringField("option3_name", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("Third option name")),
		schema.StringField("option3_value", schema.MaxLength(100), schema.Optional(),
			schema.HelpText("Third option value")),

		// Pricing (can override product price)
		schema.Float64Field("price", schema.Optional(),
			schema.HelpText("Override price (uses product price if null)")),
		schema.Float64Field("compare_at_price", schema.Optional()),
		schema.Float64Field("cost_price", schema.Optional()),

		// Inventory
		schema.Int32Field("stock_quantity", schema.Default(0),
			schema.VerboseName("Stock Quantity")),
		schema.Int32Field("reserved_quantity", schema.Default(0),
			schema.HelpText("Quantity in pending orders")),
		schema.BoolField("track_inventory", schema.Default(true)),

		// Physical attributes (can override product)
		schema.Float64Field("weight", schema.Optional()),
		schema.Float64Field("length", schema.Optional()),
		schema.Float64Field("width", schema.Optional()),
		schema.Float64Field("height", schema.Optional()),

		// Status
		schema.BoolField("is_active", schema.Default(true),
			schema.VerboseName("Active")),
		schema.BoolField("is_default", schema.Default(false),
			schema.HelpText("Default variant for product")),

		// Display
		schema.StringField("image_url", schema.MaxLength(500), schema.Optional()),
		schema.Int32Field("sort_order", schema.Default(0)),

		// Timestamps
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ProductVariant) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "product_variants",
		VerboseName:       "Product Variant",
		VerboseNamePlural: "Product Variants",
		OrderBy:           []string{"product_id", "sort_order"},
		Indexes: []schema.Index{
			schema.UniqueIndexOn("idx_variant_sku", "sku"),
			schema.IndexOn("idx_variant_product", "product_id"),
			schema.IndexOn("idx_variant_active", "is_active"),
		},
		UniqueTogether: [][]string{
			{"product_id", "option1_value", "option2_value", "option3_value"},
		},
	}
}

func (ProductVariant) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("product_id", "Product",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("variants")),
	}
}

func (ProductVariant) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Ensure only one default variant per product
			// Validate SKU uniqueness
			return nil
		},
		AfterSave: func(ctx context.Context, instance interface{}) error {
			// Update product stock_quantity
			return nil
		},
	}
}

// ProductImage represents product images
type ProductImage struct {
	ProductImageGenerated
}

func (ProductImage) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("product_id", schema.Required(),
			schema.VerboseName("Product")),
		schema.Int64Field("variant_id", schema.Optional(),
			schema.HelpText("Optional: Associate image with specific variant")),

		schema.StringField("image_url", schema.Required(), schema.MaxLength(500)),
		schema.StringField("thumbnail_url", schema.MaxLength(500), schema.Optional()),
		schema.StringField("alt_text", schema.MaxLength(255), schema.Optional(),
			schema.VerboseName("Alt Text"),
			schema.HelpText("Alternative text for accessibility")),

		schema.Int32Field("sort_order", schema.Default(0),
			schema.VerboseName("Sort Order")),
		schema.BoolField("is_primary", schema.Default(false),
			schema.VerboseName("Primary"),
			schema.HelpText("Primary product image")),

		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (ProductImage) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "product_images",
		VerboseName:       "Product Image",
		VerboseNamePlural: "Product Images",
		OrderBy:           []string{"product_id", "sort_order"},
		Indexes: []schema.Index{
			schema.IndexOn("idx_image_product", "product_id"),
			schema.IndexOn("idx_image_variant", "variant_id"),
		},
	}
}

func (ProductImage) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("product_id", "Product",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("images")),
		schema.ForeignKeyField("product_image_id", "ProductImage",
			schema.OnDelete(schema.CascadeSET_NULL),
			schema.RelatedName("images")),
	}
}

func (ProductImage) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			// Ensure only one primary image per product
			return nil
		},
	}
}

// ProductAttribute represents dynamic product attributes
type ProductAttribute struct {
	ProductAttributeGenerated
}

func (ProductAttribute) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(200), schema.Unique(),
			schema.HelpText("Attribute name (e.g., 'Color', 'Size')")),
		schema.StringField("code", schema.Required(), schema.MaxLength(100), schema.Unique(),
			schema.HelpText("Code for programmatic access")),
		schema.StringField("type", schema.Required(), schema.MaxLength(50),
			schema.HelpText("Type: text, number, select, multiselect")),
		schema.BoolField("is_filterable", schema.Default(true),
			schema.HelpText("Show in filters")),
		schema.BoolField("is_visible", schema.Default(true),
			schema.HelpText("Show on product page")),
		schema.Int32Field("sort_order", schema.Default(0)),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (ProductAttribute) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "product_attributes",
		VerboseName:       "Product Attribute",
		VerboseNamePlural: "Product Attributes",
		OrderBy:           []string{"sort_order", "name"},
	}
}

func (ProductAttribute) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (ProductAttribute) Hooks() *schema.ModelHooks {
	return nil
}

// ProductAttributeValue represents attribute values for products
type ProductAttributeValue struct {
	ProductAttributeValueGenerated
}

func (ProductAttributeValue) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("product_id", schema.Required()),
		schema.Int64Field("attribute_id", schema.Required()),
		schema.StringField("value", schema.Required(), schema.MaxLength(500)),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (ProductAttributeValue) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "product_attribute_values",
		VerboseName:       "Product Attribute Value",
		VerboseNamePlural: "Product Attribute Values",
		Indexes: []schema.Index{
			schema.IndexOn("idx_attr_value_product", "product_id"),
			schema.IndexOn("idx_attr_value_attribute", "attribute_id"),
		},
		UniqueTogether: [][]string{
			{"product_id", "attribute_id"},
		},
	}
}

func (ProductAttributeValue) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("product_id", "Product",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("attribute_values")),
		schema.ForeignKeyField("attribute_id", "ProductAttribute",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("values")),
	}
}

func (ProductAttributeValue) Hooks() *schema.ModelHooks {
	return nil
}

// RegisterModels registers catalog models with the framework
func RegisterModels() {
	registry.RegisterModel(&Category{})
	registry.RegisterModel(&Brand{})
	registry.RegisterModel(&Product{})
	registry.RegisterModel(&ProductVariant{})
	registry.RegisterModel(&ProductImage{})
	registry.RegisterModel(&ProductAttribute{})
	registry.RegisterModel(&ProductAttributeValue{})
}
