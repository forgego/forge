package catalog

import (
	"context"
	"time"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers catalog models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// Category admin
	admin.Register(&admin.Config[Category]{
		Icon: "FolderTree",
		ListDisplay: []admin.Field{
			CategoryFieldsInstance.Name,
			CategoryFieldsInstance.ParentId,
			CategoryFieldsInstance.SortOrder,
			CategoryFieldsInstance.IsActive,
			CategoryFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			CategoryFieldsInstance.IsActive,
			CategoryFieldsInstance.ParentId,
			CategoryFieldsInstance.Level,
		},
		SearchFields: []admin.Field{
			CategoryFieldsInstance.Name,
			CategoryFieldsInstance.Slug,
			CategoryFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			CategoryFieldsInstance.SortOrder,
			CategoryFieldsInstance.Name,
		},
		// Fieldsets for grouping form fields
		Fieldsets: []admin.Fieldset[Category]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "slug", "description", "image_url"},
			},
			{
				Name: "Hierarchy",
				Fields: []string{"parent_id", "level", "sort_order"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active"},
			},
		},
		// Inline relations for related models
		InlineRelations: map[string]admin.InlineRelationConfig{
			"children": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Subcategories",
				RelatedModel: "Category",
				RelatedField: "parent_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"name", "slug", "is_active", "sort_order"},
				},
			},
		},
		// History manager for audit logging
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "slug", "is_active", "parent_id"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at", "level"},
		// Custom permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Category], user interface{}, obj *Category) bool {
			return true // All authenticated users can view categories
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Category], user interface{}, obj *Category) bool {
			return true // Admins can change categories
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Category], user interface{}, obj *Category) bool {
			return true // Admins can delete categories (check for children in real implementation)
		},
		// Custom filters
		Filters: []admin.Filter[Category]{
			{
				Name:  "active_categories",
				Label: "Active Categories Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Category], value interface{}) orm.QuerySet[Category] {
					return qs.Filter(orm.F("is_active").Eq(true))
				},
			},
			{
				Name:  "root_categories",
				Label: "Root Categories",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Category], value interface{}) orm.QuerySet[Category] {
					return qs.Filter(orm.F("parent_id").IsNull())
				},
			},
		},
		// Custom bulk actions
		Actions: []admin.Action[Category]{
			{
				Name:         "activate",
				Label:        "Activate Categories",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected categories?",
				Handler: func(ctx context.Context, instances []*Category) error {
					for _, category := range instances {
						category.IsActive = true
						if err := CategoryObjects.Update(ctx, category); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Categories",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected categories?",
				Handler: func(ctx context.Context, instances []*Category) error {
					for _, category := range instances {
						category.IsActive = false
						if err := CategoryObjects.Update(ctx, category); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// Brand admin
	admin.Register(&admin.Config[Brand]{
		Icon: "Tag",
		ListDisplay: []admin.Field{
			BrandFieldsInstance.Name,
			BrandFieldsInstance.Slug,
			BrandFieldsInstance.WebsiteUrl,
			BrandFieldsInstance.IsActive,
			BrandFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			BrandFieldsInstance.IsActive,
		},
		SearchFields: []admin.Field{
			BrandFieldsInstance.Name,
			BrandFieldsInstance.Slug,
			BrandFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			BrandFieldsInstance.Name,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[Brand]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "slug", "description", "logo_url"},
			},
			{
				Name: "Contact",
				Fields: []string{"website_url"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active"},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "slug", "is_active"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Brand], user interface{}, obj *Brand) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Brand], user interface{}, obj *Brand) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Brand], user interface{}, obj *Brand) bool {
			return true
		},
		// Custom actions
		Actions: []admin.Action[Brand]{
			{
				Name:         "activate",
				Label:        "Activate Brands",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected brands?",
				Handler: func(ctx context.Context, instances []*Brand) error {
					for _, brand := range instances {
						brand.IsActive = true
						if err := BrandObjects.Update(ctx, brand); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Brands",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected brands?",
				Handler: func(ctx context.Context, instances []*Brand) error {
					for _, brand := range instances {
						brand.IsActive = false
						if err := BrandObjects.Update(ctx, brand); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// Product admin
	admin.Register(&admin.Config[Product]{
		Icon: "Package",
		ListDisplay: []admin.Field{
			ProductFieldsInstance.Name,
			ProductFieldsInstance.Sku,
			ProductFieldsInstance.Price,
			ProductFieldsInstance.StockQuantity,
			ProductFieldsInstance.IsActive,
			ProductFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			ProductFieldsInstance.IsActive,
			ProductFieldsInstance.IsFeatured,
			ProductFieldsInstance.CategoryId,
			ProductFieldsInstance.BrandId,
		},
		SearchFields: []admin.Field{
			ProductFieldsInstance.Name,
			ProductFieldsInstance.Sku,
			ProductFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			ProductFieldsInstance.CreatedAt.Desc(),
		},
		// Fieldsets for comprehensive product management
		Fieldsets: []admin.Fieldset[Product]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "slug", "sku", "description", "short_description"},
			},
			{
				Name:        "Categorization",
				Fields:      []string{"category_id", "brand_id"},
				Collapsed:   false,
				Description: "Product category and brand associations",
			},
			{
				Name:        "Pricing",
				Fields:      []string{"price", "cost_price", "compare_at_price"},
				Collapsed:   true,
				Description: "Product pricing information",
			},
			{
				Name:        "Inventory",
				Fields:      []string{"stock_quantity", "track_inventory", "allow_backorder"},
				Collapsed:   true,
				Description: "Inventory tracking settings",
			},
			{
				Name:        "Physical Attributes",
				Fields:      []string{"weight", "length", "width", "height"},
				Collapsed:   true,
				Description: "Product physical dimensions",
			},
			{
				Name:        "Status",
				Fields:      []string{"is_active", "is_featured", "is_digital"},
				Collapsed:   false,
				Description: "Product status flags",
			},
			{
				Name:        "SEO",
				Fields:      []string{"meta_title", "meta_description", "meta_keywords"},
				Collapsed:   true,
				Description: "Search engine optimization",
			},
			{
				Name:        "Statistics",
				Fields:      []string{"view_count", "order_count", "rating_average", "rating_count"},
				Collapsed:   true,
				Description: "Product statistics (read-only)",
			},
		},
		// Inline relations for variants and images
		InlineRelations: map[string]admin.InlineRelationConfig{
			"variants": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Product Variants",
				RelatedModel: "ProductVariant",
				RelatedField: "product_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"name", "sku", "price", "stock_quantity", "is_active"},
					Fieldsets: []admin.Fieldset[ProductVariant]{
						{Name: "Identification", Fields: []string{"sku", "name", "is_default"}},
						{Name: "Options", Fields: []string{"option1_name", "option1_value", "option2_name", "option2_value"}},
						{Name: "Pricing", Fields: []string{"price", "cost_price", "compare_at_price"}},
						{Name: "Inventory", Fields: []string{"stock_quantity", "track_inventory"}},
					},
				},
			},
			"images": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Product Images",
				RelatedModel: "ProductImage",
				RelatedField: "product_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"image_url", "alt_text", "is_primary", "sort_order"},
					Fieldsets: []admin.Fieldset[ProductImage]{
						{Name: "Image", Fields: []string{"image_url", "thumbnail_url", "alt_text"}},
						{Name: "Display", Fields: []string{"is_primary", "sort_order"}},
					},
				},
			},
			"attribute_values": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Attribute Values",
				RelatedModel: "ProductAttributeValue",
				RelatedField: "product_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"attribute_id", "value"},
				},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "slug", "price", "is_active", "is_featured", "stock_quantity"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at", "published_at", "view_count", "order_count", "rating_average", "rating_count"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Product], user interface{}, obj *Product) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Product], user interface{}, obj *Product) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Product], user interface{}, obj *Product) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[Product]{
			{
				Name:  "in_stock",
				Label: "In Stock",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Product], value interface{}) orm.QuerySet[Product] {
					return qs.Filter(orm.F("stock_quantity").Gt(0))
				},
			},
			{
				Name:  "out_of_stock",
				Label: "Out of Stock",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Product], value interface{}) orm.QuerySet[Product] {
					return qs.Filter(orm.F("stock_quantity").Eq(0))
				},
			},
			{
				Name:  "featured",
				Label: "Featured Products",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Product], value interface{}) orm.QuerySet[Product] {
					return qs.Filter(orm.F("is_featured").Eq(true))
				},
			},
			{
				Name:  "price_range",
				Label: "Price Range",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "0-50", Label: "Under $50"},
					{Value: "50-100", Label: "$50 - $100"},
					{Value: "100-500", Label: "$100 - $500"},
					{Value: "500+", Label: "$500+"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[Product], value interface{}) orm.QuerySet[Product] {
					switch value {
					case "0-50":
						return qs.Filter(orm.F("price").Lt(50))
					case "50-100":
						return qs.Filter(orm.F("price").Gte(50)).Filter(orm.F("price").Lte(100))
					case "100-500":
						return qs.Filter(orm.F("price").Gte(100)).Filter(orm.F("price").Lte(500))
					case "500+":
						return qs.Filter(orm.F("price").Gt(500))
					}
					return qs
				},
			},
		},
		// Custom bulk actions
		Actions: []admin.Action[Product]{
			{
				Name:         "activate",
				Label:        "Activate Products",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected products?",
				Handler: func(ctx context.Context, instances []*Product) error {
					for _, p := range instances {
						p.IsActive = true
						if err := ProductObjects.Update(ctx, p); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Products",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected products?",
				Handler: func(ctx context.Context, instances []*Product) error {
					for _, p := range instances {
						p.IsActive = false
						if err := ProductObjects.Update(ctx, p); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:         "feature",
				Label:        "Mark as Featured",
				Icon:         "Star",
				Confirmation: "Mark selected products as featured?",
				Handler: func(ctx context.Context, instances []*Product) error {
					for _, p := range instances {
						p.IsFeatured = true
						if err := ProductObjects.Update(ctx, p); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:         "unfeature",
				Label:        "Remove from Featured",
				Icon:         "StarOff",
				Confirmation: "Remove selected products from featured?",
				Handler: func(ctx context.Context, instances []*Product) error {
					for _, p := range instances {
						p.IsFeatured = false
						if err := ProductObjects.Update(ctx, p); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:         "update_price",
				Label:        "Update Price",
				Icon:         "DollarSign",
				Confirmation: "Update price for selected products?",
				Handler: func(ctx context.Context, instances []*Product) error {
					// Implementation would include price adjustment logic
					return nil
				},
			},
		},
	})

	// ProductVariant admin
	admin.Register(&admin.Config[ProductVariant]{
		Icon: "Layers",
		ListDisplay: []admin.Field{
			ProductVariantFieldsInstance.Name,
			ProductVariantFieldsInstance.Sku,
			ProductVariantFieldsInstance.Price,
			ProductVariantFieldsInstance.StockQuantity,
			ProductVariantFieldsInstance.IsActive,
		},
		ListFilter: []admin.Field{
			ProductVariantFieldsInstance.IsActive,
			ProductVariantFieldsInstance.ProductId,
		},
		SearchFields: []admin.Field{
			ProductVariantFieldsInstance.Name,
			ProductVariantFieldsInstance.Sku,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[ProductVariant]{
			{
				Name: "Identification",
				Fields: []string{"product_id", "sku", "name", "is_default"},
			},
			{
				Name: "Options",
				Fields: []string{"option1_name", "option1_value", "option2_name", "option2_value", "option3_name", "option3_value"},
			},
			{
				Name:        "Pricing",
				Fields:      []string{"price", "cost_price", "compare_at_price"},
				Collapsed:   true,
				Description: "Variant pricing (overrides product price if set)",
			},
			{
				Name:        "Inventory",
				Fields:      []string{"stock_quantity", "reserved_quantity", "track_inventory"},
				Collapsed:   false,
				Description: "Inventory levels",
			},
			{
				Name:        "Physical Attributes",
				Fields:      []string{"weight", "length", "width", "height"},
				Collapsed:   true,
				Description: "Physical dimensions",
			},
			{
				Name:  "Display",
				Fields: []string{"image_url", "sort_order", "is_active"},
			},
		},
		// Inline relations
		InlineRelations: map[string]admin.InlineRelationConfig{
			"images": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Variant Images",
				RelatedModel: "ProductImage",
				RelatedField: "variant_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"image_url", "alt_text", "is_primary", "sort_order"},
				},
			},
			"stock_records": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Stock by Warehouse",
				RelatedModel: "Stock",
				RelatedField: "product_variant_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"warehouse_id", "quantity", "reserved_quantity"},
				},
			},
			"movements": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Stock Movements",
				RelatedModel: "StockMovement",
				RelatedField: "product_variant_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"type", "quantity", "movement_date"},
				},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "sku", "price", "stock_quantity", "is_active"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at", "reserved_quantity"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ProductVariant], user interface{}, obj *ProductVariant) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ProductVariant], user interface{}, obj *ProductVariant) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ProductVariant], user interface{}, obj *ProductVariant) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[ProductVariant]{
			{
				Name:  "in_stock",
				Label: "In Stock",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ProductVariant], value interface{}) orm.QuerySet[ProductVariant] {
					return qs.Filter(orm.F("stock_quantity").Gt(0))
				},
			},
			{
				Name:  "low_stock",
				Label: "Low Stock",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ProductVariant], value interface{}) orm.QuerySet[ProductVariant] {
					return qs.Filter(orm.F("stock_quantity").Lte(10))
				},
			},
		},
		// Custom actions
		Actions: []admin.Action[ProductVariant]{
			{
				Name:         "activate",
				Label:        "Activate Variants",
				Icon:         "Check",
				Confirmation: "Are you sure you want to activate the selected variants?",
				Handler: func(ctx context.Context, instances []*ProductVariant) error {
					for _, variant := range instances {
						variant.IsActive = true
						if err := ProductVariantObjects.Update(ctx, variant); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Variants",
				Icon:         "X",
				Confirmation: "Are you sure you want to deactivate the selected variants?",
				Handler: func(ctx context.Context, instances []*ProductVariant) error {
					for _, variant := range instances {
						variant.IsActive = false
						if err := ProductVariantObjects.Update(ctx, variant); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:         "set_default",
				Label:        "Set as Default",
				Icon:         "CheckCircle",
				Confirmation: "Set selected variant as the default for its product?",
				Handler: func(ctx context.Context, instances []*ProductVariant) error {
					if len(instances) == 1 {
						variant := instances[0]
						variant.IsDefault = true
						if err := ProductVariantObjects.Update(ctx, variant); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// ProductImage admin
	admin.Register(&admin.Config[ProductImage]{
		Icon: "Image",
		ListDisplay: []admin.Field{
			ProductImageFieldsInstance.ProductId,
			ProductImageFieldsInstance.AltText,
			ProductImageFieldsInstance.IsPrimary,
			ProductImageFieldsInstance.SortOrder,
		},
		ListFilter: []admin.Field{
			ProductImageFieldsInstance.IsPrimary,
			ProductImageFieldsInstance.ProductId,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[ProductImage]{
			{
				Name: "Image Information",
				Fields: []string{"product_id", "variant_id", "image_url", "thumbnail_url"},
			},
			{
				Name: "Display",
				Fields: []string{"alt_text", "is_primary", "sort_order"},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"image_url", "is_primary", "sort_order"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ProductImage], user interface{}, obj *ProductImage) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ProductImage], user interface{}, obj *ProductImage) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ProductImage], user interface{}, obj *ProductImage) bool {
			return true
		},
		// Custom actions
		Actions: []admin.Action[ProductImage]{
			{
				Name:         "mark_primary",
				Label:        "Mark as Primary",
				Icon:         "Check",
				Confirmation: "Mark selected images as primary?",
				Handler: func(ctx context.Context, instances []*ProductImage) error {
					for _, image := range instances {
						image.IsPrimary = true
						if err := ProductImageObjects.Update(ctx, image); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:         "reset_sort",
				Label:        "Reset Sort Order",
				Icon:         "RefreshCw",
				Confirmation: "Reset sort order for selected images?",
				Handler: func(ctx context.Context, instances []*ProductImage) error {
					for _, image := range instances {
						image.SortOrder = 0
						if err := ProductImageObjects.Update(ctx, image); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// ProductAttribute admin
	admin.Register(&admin.Config[ProductAttribute]{
		Icon: "Settings",
		ListDisplay: []admin.Field{
			ProductAttributeFieldsInstance.Name,
			ProductAttributeFieldsInstance.Code,
			ProductAttributeFieldsInstance.Type,
			ProductAttributeFieldsInstance.IsFilterable,
			ProductAttributeFieldsInstance.IsVisible,
			ProductAttributeFieldsInstance.SortOrder,
		},
		ListFilter: []admin.Field{
			ProductAttributeFieldsInstance.Type,
			ProductAttributeFieldsInstance.IsFilterable,
			ProductAttributeFieldsInstance.IsVisible,
		},
		SearchFields: []admin.Field{
			ProductAttributeFieldsInstance.Name,
			ProductAttributeFieldsInstance.Code,
		},
		Ordering: []admin.Field{
			ProductAttributeFieldsInstance.SortOrder,
			ProductAttributeFieldsInstance.Name,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[ProductAttribute]{
			{
				Name: "Attribute Definition",
				Fields: []string{"name", "code", "type"},
			},
			{
				Name: "Display Settings",
				Fields: []string{"is_filterable", "is_visible", "sort_order"},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "code", "type", "is_filterable", "is_visible"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ProductAttribute], user interface{}, obj *ProductAttribute) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ProductAttribute], user interface{}, obj *ProductAttribute) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ProductAttribute], user interface{}, obj *ProductAttribute) bool {
			return true
		},
		// Custom actions
		Actions: []admin.Action[ProductAttribute]{
			{
				Name:         "toggle_filterable",
				Label:        "Toggle Filterable",
				Icon:         "Filter",
				Confirmation: "Toggle filterable status for selected attributes?",
				Handler: func(ctx context.Context, instances []*ProductAttribute) error {
					for _, attr := range instances {
						attr.IsFilterable = !attr.IsFilterable
						if err := ProductAttributeObjects.Update(ctx, attr); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:         "toggle_visible",
				Label:        "Toggle Visible",
				Icon:         "Eye",
				Confirmation: "Toggle visible status for selected attributes?",
				Handler: func(ctx context.Context, instances []*ProductAttribute) error {
					for _, attr := range instances {
						attr.IsVisible = !attr.IsVisible
						if err := ProductAttributeObjects.Update(ctx, attr); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// ProductAttributeValue admin
	admin.Register(&admin.Config[ProductAttributeValue]{
		Icon: "Tag",
		ListDisplay: []admin.Field{
			ProductAttributeValueFieldsInstance.ProductId,
			ProductAttributeValueFieldsInstance.AttributeId,
			ProductAttributeValueFieldsInstance.Value,
			ProductAttributeValueFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			ProductAttributeValueFieldsInstance.ProductId,
			ProductAttributeValueFieldsInstance.AttributeId,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[ProductAttributeValue]{
			{
				Name: "Attribute Value",
				Fields: []string{"product_id", "attribute_id", "value"},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"value"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ProductAttributeValue], user interface{}, obj *ProductAttributeValue) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ProductAttributeValue], user interface{}, obj *ProductAttributeValue) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ProductAttributeValue], user interface{}, obj *ProductAttributeValue) bool {
			return true
		},
	})
}

// Helper function to get current time
func now() time.Time {
	return time.Now()
}
