package marketing

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers marketing API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("coupons", &api.ViewSetConfig{
		Model:      &Coupon{},
		Queryset:   CouponObjects,
		Serializer: base,
	})

	router.Register("coupon-usage", &api.ViewSetConfig{
		Model:      &CouponUsage{},
		Queryset:   CouponUsageObjects,
		Serializer: base,
	})

	router.Register("reviews", &api.ViewSetConfig{
		Model:      &Review{},
		Queryset:   ReviewObjects,
		Serializer: base,
	})

	router.Register("review-images", &api.ViewSetConfig{
		Model:      &ReviewImage{},
		Queryset:   ReviewImageObjects,
		Serializer: base,
	})

	router.Register("review-helpfulness", &api.ViewSetConfig{
		Model:      &ReviewHelpfulness{},
		Queryset:   ReviewHelpfulnessObjects,
		Serializer: base,
	})

	router.Register("product-questions", &api.ViewSetConfig{
		Model:      &ProductQuestion{},
		Queryset:   ProductQuestionObjects,
		Serializer: base,
	})
}
