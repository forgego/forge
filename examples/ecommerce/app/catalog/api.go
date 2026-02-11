package catalog

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers catalog API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("categories", &api.ViewSetConfig{
		Model:      &Category{},
		Queryset:   CategoryObjects,
		Serializer: base,
	})

	router.Register("brands", &api.ViewSetConfig{
		Model:      &Brand{},
		Queryset:   BrandObjects,
		Serializer: base,
	})

	router.Register("products", &api.ViewSetConfig{
		Model:      &Product{},
		Queryset:   ProductObjects,
		Serializer: base,
	})

	router.Register("product-variants", &api.ViewSetConfig{
		Model:      &ProductVariant{},
		Queryset:   ProductVariantObjects,
		Serializer: base,
	})

	router.Register("product-images", &api.ViewSetConfig{
		Model:      &ProductImage{},
		Queryset:   ProductImageObjects,
		Serializer: base,
	})

	router.Register("product-attributes", &api.ViewSetConfig{
		Model:      &ProductAttribute{},
		Queryset:   ProductAttributeObjects,
		Serializer: base,
	})

	router.Register("product-attribute-values", &api.ViewSetConfig{
		Model:      &ProductAttributeValue{},
		Queryset:   ProductAttributeValueObjects,
		Serializer: base,
	})
}
