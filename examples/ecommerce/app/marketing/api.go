package marketing

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers marketing API endpoints
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// Coupon API
	router.Register("coupons", &api.ViewSetConfig{
		Model:        &Coupon{},
		Queryset:     CouponObjects,
		Serializer:   &CouponSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "code", "name", "discount_type", "discount_value", "usage_count", "usage_limit", "valid_from", "valid_until", "is_active"},
		DetailFields: []string{"id", "code", "name", "description", "discount_type", "discount_value", "minimum_purchase_amount", "maximum_discount_amount", "usage_limit", "usage_limit_per_customer", "usage_count", "applies_to_all_products", "product_ids", "category_ids", "excluded_product_ids", "applies_to_all_customers", "customer_group_ids", "customer_email_list", "valid_from", "valid_until", "is_active", "is_public", "priority", "created_at", "updated_at"},
		Filterable:   []string{"discount_type", "is_active", "is_public", "valid_from", "valid_until"},
		Searchable:   []string{"code", "name", "description"},
		Ordering:     []string{"-priority", "-created_at", "code"},
		PerPage:      20,
	})

	// CouponUsage API
	router.Register("coupon-usage", &api.ViewSetConfig{
		Model:        &CouponUsage{},
		Queryset:     CouponUsageObjects,
		Serializer:   &CouponUsageSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "coupon_id", "order_id", "customer_id", "discount_amount", "created_at"},
		DetailFields: []string{"id", "coupon_id", "order_id", "customer_id", "discount_amount", "created_at"},
		Filterable:   []string{"coupon_id", "order_id", "customer_id"},
		Searchable:   []string{},
		Ordering:     []string{"-created_at"},
		PerPage:      20,
	})

	// Review API
	router.Register("reviews", &api.ViewSetConfig{
		Model:        &Review{},
		Queryset:     ReviewObjects,
		Serializer:   &ReviewSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "product_id", "customer_id", "title", "rating", "is_verified_purchase", "status", "helpful_count", "is_featured", "created_at"},
		DetailFields: []string{"id", "product_id", "customer_id", "order_id", "title", "content", "rating", "is_verified_purchase", "status", "is_featured", "helpful_count", "not_helpful_count", "merchant_response", "merchant_response_at", "merchant_response_by_user_id", "report_count", "report_reasons", "created_at", "updated_at", "approved_at"},
		Filterable:   []string{"product_id", "customer_id", "rating", "status", "is_verified_purchase", "is_featured"},
		Searchable:   []string{"title", "content"},
		Ordering:     []string{"-created_at", "-helpful_count", "rating"},
		PerPage:      20,
	})

	// ReviewImage API
	router.Register("review-images", &api.ViewSetConfig{
		Model:        &ReviewImage{},
		Queryset:     ReviewImageObjects,
		Serializer:   &ReviewImageSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "review_id", "image_url", "thumbnail_url", "alt_text", "sort_order"},
		DetailFields: []string{"id", "review_id", "image_url", "thumbnail_url", "alt_text", "sort_order", "created_at"},
		Filterable:   []string{"review_id"},
		Searchable:   []string{"alt_text"},
		Ordering:     []string{"review_id", "sort_order"},
		PerPage:      50,
	})

	// ReviewHelpfulness API (usually not exposed, just for admin)
	router.Register("review-helpfulness", &api.ViewSetConfig{
		Model:        &ReviewHelpfulness{},
		Queryset:     ReviewHelpfulnessObjects,
		Serializer:   &ReviewHelpfulnessSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "review_id", "customer_id", "is_helpful", "created_at"},
		DetailFields: []string{"id", "review_id", "customer_id", "is_helpful", "ip_address", "created_at"},
		Filterable:   []string{"review_id", "customer_id", "is_helpful"},
		Searchable:   []string{},
		Ordering:     []string{"-created_at"},
		PerPage:      20,
	})

	// ProductQuestion API
	router.Register("product-questions", &api.ViewSetConfig{
		Model:        &ProductQuestion{},
		Queryset:     ProductQuestionObjects,
		Serializer:   &ProductQuestionSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "product_id", "customer_id", "question", "status", "is_public", "answered_at", "created_at"},
		DetailFields: []string{"id", "product_id", "customer_id", "question", "answer", "answered_at", "answered_by_user_id", "answered_by_user_name", "status", "is_public", "helpful_count", "created_at", "updated_at"},
		Filterable:   []string{"product_id", "customer_id", "status", "is_public"},
		Searchable:   []string{"question", "answer"},
		Ordering:     []string{"-created_at", "-helpful_count"},
		PerPage:      20,
	})
}

// Serializers

type CouponSerializer struct {
	*api.BaseSerializer
}

func (s *CouponSerializer) New() api.Serializer {
	return &CouponSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type CouponUsageSerializer struct {
	*api.BaseSerializer
}

func (s *CouponUsageSerializer) New() api.Serializer {
	return &CouponUsageSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type ReviewSerializer struct {
	*api.BaseSerializer
}

func (s *ReviewSerializer) New() api.Serializer {
	return &ReviewSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type ReviewImageSerializer struct {
	*api.BaseSerializer
}

func (s *ReviewImageSerializer) New() api.Serializer {
	return &ReviewImageSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type ReviewHelpfulnessSerializer struct {
	*api.BaseSerializer
}

func (s *ReviewHelpfulnessSerializer) New() api.Serializer {
	return &ReviewHelpfulnessSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type ProductQuestionSerializer struct {
	*api.BaseSerializer
}

func (s *ProductQuestionSerializer) New() api.Serializer {
	return &ProductQuestionSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}
