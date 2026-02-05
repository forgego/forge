package marketing

import (
	"context"

	"github.com/forgego/forge/admin"
)

// RegisterAdmin registers marketing models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// Coupon admin
	admin.Register(&admin.Config[Coupon]{
		Icon: "Ticket",
		ListDisplay: []admin.Field{
			CouponFieldsInstance.Code,
			CouponFieldsInstance.Name,
			CouponFieldsInstance.DiscountType,
			CouponFieldsInstance.DiscountValue,
			CouponFieldsInstance.IsActive,
			CouponFieldsInstance.ValidUntil,
		},
		ListFilter: []admin.Field{
			CouponFieldsInstance.DiscountType,
			CouponFieldsInstance.IsActive,
		},
		Actions: []admin.Action[Coupon]{
			{
				Name:  "activate",
				Label: "Activate Coupons",
				Handler: func(ctx context.Context, instances []*Coupon) error {
					for _, coupon := range instances {
						coupon.IsActive = true
					}
					return nil
				},
			},
		},
	})

	// CouponUsage admin
	admin.Register(&admin.Config[CouponUsage]{
		Icon: "BarChart3",
		ListDisplay: []admin.Field{
			CouponUsageFieldsInstance.CouponId,
			CouponUsageFieldsInstance.OrderId,
			CouponUsageFieldsInstance.CustomerId,
			CouponUsageFieldsInstance.DiscountAmount,
			CouponUsageFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			CouponUsageFieldsInstance.CouponId,
			CouponUsageFieldsInstance.OrderId,
			CouponUsageFieldsInstance.CustomerId,
		},
		Actions: []admin.Action[CouponUsage]{
			{
				Name:  "zero_discount",
				Label: "Zero Discount Amount",
				Handler: func(ctx context.Context, instances []*CouponUsage) error {
					for _, usage := range instances {
						usage.DiscountAmount = 0
					}
					return nil
				},
			},
		},
	})

	// Review admin
	admin.Register(&admin.Config[Review]{
		Icon: "Star",
		ListDisplay: []admin.Field{
			ReviewFieldsInstance.ProductId,
			ReviewFieldsInstance.CustomerId,
			ReviewFieldsInstance.Rating,
			ReviewFieldsInstance.Status,
			ReviewFieldsInstance.IsFeatured,
			ReviewFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			ReviewFieldsInstance.Status,
			ReviewFieldsInstance.Rating,
			ReviewFieldsInstance.IsFeatured,
		},
		Actions: []admin.Action[Review]{
			{
				Name:  "approve",
				Label: "Approve Reviews",
				Handler: func(ctx context.Context, instances []*Review) error {
					for _, review := range instances {
						review.Status = "approved"
					}
					return nil
				},
			},
		},
	})

	// ReviewImage admin
	admin.Register(&admin.Config[ReviewImage]{
		Icon: "Image",
		ListDisplay: []admin.Field{
			ReviewImageFieldsInstance.ReviewId,
			ReviewImageFieldsInstance.AltText,
			ReviewImageFieldsInstance.SortOrder,
		},
		ListFilter: []admin.Field{
			ReviewImageFieldsInstance.ReviewId,
		},
		Actions: []admin.Action[ReviewImage]{
			{
				Name:  "reset_sort",
				Label: "Reset Sort Order",
				Handler: func(ctx context.Context, instances []*ReviewImage) error {
					for _, image := range instances {
						image.SortOrder = 0
					}
					return nil
				},
			},
		},
	})

	// ReviewHelpfulness admin
	admin.Register(&admin.Config[ReviewHelpfulness]{
		Icon: "ThumbsUp",
		ListDisplay: []admin.Field{
			ReviewHelpfulnessFieldsInstance.ReviewId,
			ReviewHelpfulnessFieldsInstance.CustomerId,
			ReviewHelpfulnessFieldsInstance.IsHelpful,
		},
		ListFilter: []admin.Field{
			ReviewHelpfulnessFieldsInstance.IsHelpful,
		},
		Actions: []admin.Action[ReviewHelpfulness]{
			{
				Name:  "mark_helpful",
				Label: "Mark Helpful",
				Handler: func(ctx context.Context, instances []*ReviewHelpfulness) error {
					for _, helpfulness := range instances {
						helpfulness.IsHelpful = true
					}
					return nil
				},
			},
		},
	})

	// ProductQuestion admin
	admin.Register(&admin.Config[ProductQuestion]{
		Icon: "HelpCircle",
		ListDisplay: []admin.Field{
			ProductQuestionFieldsInstance.ProductId,
			ProductQuestionFieldsInstance.CustomerId,
			ProductQuestionFieldsInstance.Status,
			ProductQuestionFieldsInstance.IsPublic,
			ProductQuestionFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			ProductQuestionFieldsInstance.Status,
			ProductQuestionFieldsInstance.IsPublic,
		},
		Actions: []admin.Action[ProductQuestion]{
			{
				Name:  "mark_answered",
				Label: "Mark Answered",
				Handler: func(ctx context.Context, instances []*ProductQuestion) error {
					for _, question := range instances {
						question.Status = "answered"
					}
					return nil
				},
			},
		},
	})
}
