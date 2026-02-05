package catalog

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers catalog API endpoints
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// Category API
	router.Register("categories", &api.ViewSetConfig{
		Model:        &Category{},
		Queryset:     CategoryObjects,
		Serializer:   &CategorySerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "name", "slug", "description", "parent_id", "level", "is_active"},
		DetailFields: []string{"id", "name", "slug", "description", "parent_id", "image_url", "sort_order", "is_active", "level", "created_at", "updated_at"},
		Filterable:   []string{"is_active", "parent_id", "level"},
		Searchable:   []string{"name", "slug", "description"},
		Ordering:     []string{"sort_order", "name", "created_at"},
		PerPage:      20,
	})

	// Brand API
	router.Register("brands", &api.ViewSetConfig{
		Model:        &Brand{},
		Queryset:     BrandObjects,
		Serializer:   &BrandSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "name", "slug", "logo_url", "is_active"},
		DetailFields: []string{"id", "name", "slug", "description", "logo_url", "website_url", "is_active", "created_at", "updated_at"},
		Filterable:   []string{"is_active"},
		Searchable:   []string{"name", "slug", "description"},
		Ordering:     []string{"name", "created_at"},
		PerPage:      20,
	})

	// Product API
	router.Register("products", &api.ViewSetConfig{
		Model:        &Product{},
		Queryset:     ProductObjects,
		Serializer:   &ProductSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "name", "slug", "sku", "short_description", "price", "compare_at_price", "category_id", "brand_id", "is_active", "is_featured", "rating_average", "rating_count"},
		DetailFields: []string{"id", "name", "slug", "sku", "description", "short_description", "price", "cost_price", "compare_at_price", "category_id", "brand_id", "stock_quantity", "weight", "length", "width", "height", "is_active", "is_featured", "is_digital", "meta_title", "meta_description", "view_count", "order_count", "rating_average", "rating_count", "created_at", "updated_at", "published_at"},
		Filterable:   []string{"is_active", "is_featured", "category_id", "brand_id", "price", "rating_average"},
		Searchable:   []string{"name", "sku", "description", "short_description"},
		Ordering:     []string{"name", "price", "created_at", "-rating_average", "-order_count"},
		PerPage:      20,
	})

	// ProductVariant API
	router.Register("product-variants", &api.ViewSetConfig{
		Model:        &ProductVariant{},
		Queryset:     ProductVariantObjects,
		Serializer:   &ProductVariantSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "product_id", "sku", "name", "price", "stock_quantity", "is_active"},
		DetailFields: []string{"id", "product_id", "sku", "name", "option1_name", "option1_value", "option2_name", "option2_value", "option3_name", "option3_value", "price", "compare_at_price", "cost_price", "stock_quantity", "reserved_quantity", "weight", "length", "width", "height", "is_active", "is_default", "image_url", "sort_order", "created_at", "updated_at"},
		Filterable:   []string{"is_active", "product_id"},
		Searchable:   []string{"name", "sku"},
		Ordering:     []string{"product_id", "sort_order", "name"},
		PerPage:      20,
	})

	// ProductImage API
	router.Register("product-images", &api.ViewSetConfig{
		Model:        &ProductImage{},
		Queryset:     ProductImageObjects,
		Serializer:   &ProductImageSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "product_id", "variant_id", "image_url", "thumbnail_url", "alt_text", "is_primary", "sort_order"},
		DetailFields: []string{"id", "product_id", "variant_id", "image_url", "thumbnail_url", "alt_text", "is_primary", "sort_order", "created_at", "updated_at"},
		Filterable:   []string{"product_id", "variant_id", "is_primary"},
		Searchable:   []string{"alt_text"},
		Ordering:     []string{"product_id", "sort_order"},
		PerPage:      50,
	})

	// ProductAttribute API
	router.Register("product-attributes", &api.ViewSetConfig{
		Model:        &ProductAttribute{},
		Queryset:     ProductAttributeObjects,
		Serializer:   &ProductAttributeSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "name", "code", "type", "is_filterable", "is_visible", "sort_order"},
		DetailFields: []string{"id", "name", "code", "type", "is_filterable", "is_visible", "sort_order", "created_at"},
		Filterable:   []string{"type", "is_filterable", "is_visible"},
		Searchable:   []string{"name", "code"},
		Ordering:     []string{"sort_order", "name"},
		PerPage:      50,
	})

	// ProductAttributeValue API
	router.Register("product-attribute-values", &api.ViewSetConfig{
		Model:        &ProductAttributeValue{},
		Queryset:     ProductAttributeValueObjects,
		Serializer:   &ProductAttributeValueSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "product_id", "attribute_id", "value"},
		DetailFields: []string{"id", "product_id", "attribute_id", "value", "created_at"},
		Filterable:   []string{"product_id", "attribute_id"},
		Searchable:   []string{"value"},
		Ordering:     []string{"product_id", "attribute_id"},
		PerPage:      50,
	})
}

// Serializers

type CategorySerializer struct {
	*api.BaseSerializer
}

func (s *CategorySerializer) New() api.Serializer {
	return &CategorySerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type BrandSerializer struct {
	*api.BaseSerializer
}

func (s *BrandSerializer) New() api.Serializer {
	return &BrandSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type ProductSerializer struct {
	*api.BaseSerializer
}

func (s *ProductSerializer) New() api.Serializer {
	return &ProductSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type ProductVariantSerializer struct {
	*api.BaseSerializer
}

func (s *ProductVariantSerializer) New() api.Serializer {
	return &ProductVariantSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type ProductImageSerializer struct {
	*api.BaseSerializer
}

func (s *ProductImageSerializer) New() api.Serializer {
	return &ProductImageSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type ProductAttributeSerializer struct {
	*api.BaseSerializer
}

func (s *ProductAttributeSerializer) New() api.Serializer {
	return &ProductAttributeSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type ProductAttributeValueSerializer struct {
	*api.BaseSerializer
}

func (s *ProductAttributeValueSerializer) New() api.Serializer {
	return &ProductAttributeValueSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}
