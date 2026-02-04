package catalog

import (
	"context"

	"github.com/forgego/forge/admin"
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
		Actions: []admin.Action[Category]{
			{
				Name:  "activate",
				Label: "Activate Categories",
				Handler: func(ctx context.Context, instances []*Category) error {
					for _, category := range instances {
						category.IsActive = true
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
		Actions: []admin.Action[Brand]{
			{
				Name:  "activate",
				Label: "Activate Brands",
				Handler: func(ctx context.Context, instances []*Brand) error {
					for _, brand := range instances {
						brand.IsActive = true
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
		Actions: []admin.Action[Product]{
			{
				Name:  "activate",
				Label: "Activate Products",
				Handler: func(ctx context.Context, instances []*Product) error {
					for _, p := range instances {
						p.IsActive = true
						// Save would be handled by a manager call or similar depending on ORM context
					}
					return nil
				},
			},
			{
				Name:  "deactivate",
				Label: "Deactivate Products",
				Handler: func(ctx context.Context, instances []*Product) error {
					for _, p := range instances {
						p.IsActive = false
					}
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
		Actions: []admin.Action[ProductVariant]{
			{
				Name:  "activate",
				Label: "Activate Variants",
				Handler: func(ctx context.Context, instances []*ProductVariant) error {
					for _, variant := range instances {
						variant.IsActive = true
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
		Actions: []admin.Action[ProductImage]{
			{
				Name:  "mark_primary",
				Label: "Mark as Primary",
				Handler: func(ctx context.Context, instances []*ProductImage) error {
					for _, image := range instances {
						image.IsPrimary = true
					}
					return nil
				},
			},
		},
	})
}
