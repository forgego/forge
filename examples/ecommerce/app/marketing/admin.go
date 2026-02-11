package marketing

import (
	"context"
	"time"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
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
		SearchFields: []admin.Field{
			CouponFieldsInstance.Code,
			CouponFieldsInstance.Name,
		},
		Fieldsets: []admin.Fieldset[Coupon]{
			{
				Name: "Basic Information",
				Fields: []string{"code", "name", "description"},
			},
			{
				Name: "Discount Rules",
				Fields: []string{"discount_type", "discount_value", "minimum_purchase_amount", "maximum_discount_amount"},
			},
			{
				Name: "Usage Limits",
				Fields: []string{"usage_limit", "usage_limit_per_customer", "usage_count"},
			},
			{
				Name: "Applicability",
				Fields: []string{"applies_to_all_products", "product_ids", "category_ids", "excluded_product_ids"},
				Collapsed: true,
			},
			{
				Name: "Customer Restrictions",
				Fields: []string{"applies_to_all_customers", "customer_group_ids", "customer_email_list"},
				Collapsed: true,
			},
			{
				Name: "Validity",
				Fields: []string{"valid_from", "valid_until"},
			},
			{
				Name:  "Status",
				Fields: []string{"is_active", "is_public", "priority"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"code", "discount_type", "discount_value", "is_active", "usage_limit", "valid_until"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "usage_count"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Coupon], user interface{}, obj *Coupon) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Coupon], user interface{}, obj *Coupon) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Coupon], user interface{}, obj *Coupon) bool {
			return true
		},
		Filters: []admin.Filter[Coupon]{
			{
				Name:  "active_coupons",
				Label: "Active Coupons",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Coupon], value interface{}) orm.QuerySet[Coupon] {
					return qs.Filter(orm.F("is_active").Eq(true))
				},
			},
			{
				Name:  "expired_coupons",
				Label: "Expired Coupons",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Coupon], value interface{}) orm.QuerySet[Coupon] {
					return qs.Filter(orm.F("valid_until").Lt(time.Now()))
				},
			},
		},
		Actions: []admin.Action[Coupon]{
			{
				Name:  "activate",
				Label: "Activate Coupons",
				Handler: func(ctx context.Context, instances []*Coupon) error {
					for _, coupon := range instances {
						coupon.IsActive = true
						if err := CouponObjects.Update(ctx, coupon); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "deactivate",
				Label: "Deactivate Coupons",
				Handler: func(ctx context.Context, instances []*Coupon) error {
					for _, coupon := range instances {
						coupon.IsActive = false
						if err := CouponObjects.Update(ctx, coupon); err != nil {
							return err
						}
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
		Fieldsets: []admin.Fieldset[CouponUsage]{
			{
				Name: "Usage Details",
				Fields: []string{"coupon_id", "order_id", "customer_id", "discount_amount"},
			},
			{
				Name:  "Timestamp",
				Fields: []string{"created_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"discount_amount"},
		},
		ReadOnlyFields: []string{"created_at"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[CouponUsage], user interface{}, obj *CouponUsage) bool {
			return true
		},
		Actions: []admin.Action[CouponUsage]{
			{
				Name:  "zero_discount",
				Label: "Zero Discount Amount",
				Handler: func(ctx context.Context, instances []*CouponUsage) error {
					for _, usage := range instances {
						usage.DiscountAmount = 0
						if err := CouponUsageObjects.Update(ctx, usage); err != nil {
							return err
						}
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
		SearchFields: []admin.Field{
			ReviewFieldsInstance.Title,
			ReviewFieldsInstance.Content,
		},
		Fieldsets: []admin.Fieldset[Review]{
			{
				Name: "Review Content",
				Fields: []string{"product_id", "customer_id", "order_id", "title", "content", "rating"},
			},
			{
				Name: "Verification",
				Fields: []string{"is_verified_purchase"},
			},
			{
				Name: "Moderation",
				Fields: []string{"status", "is_featured"},
			},
			{
				Name: "Helpfulness",
				Fields: []string{"helpful_count", "not_helpful_count"},
				Collapsed: true,
			},
			{
				Name: "Merchant Response",
				Fields: []string{"merchant_response", "merchant_response_at"},
				Collapsed: true,
			},
			{
				Name:  "Timestamps",
				Fields: []string{"created_at", "updated_at", "approved_at"},
			},
		},
		InlineRelations: map[string]admin.InlineRelationConfig{
			"images": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Review Images",
				RelatedModel: "ReviewImage",
				RelatedField: "review_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"image_url", "alt_text", "sort_order"},
				},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "is_featured", "rating", "merchant_response"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "helpful_count", "not_helpful_count"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Review], user interface{}, obj *Review) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Review], user interface{}, obj *Review) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Review], user interface{}, obj *Review) bool {
			return true
		},
		Filters: []admin.Filter[Review]{
			{
				Name:  "pending_reviews",
				Label: "Pending Reviews",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Review], value interface{}) orm.QuerySet[Review] {
					return qs.Filter(orm.F("status").Eq("pending"))
				},
			},
			{
				Name:  "approved_reviews",
				Label: "Approved Reviews",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Review], value interface{}) orm.QuerySet[Review] {
					return qs.Filter(orm.F("status").Eq("approved"))
				},
			},
			{
				Name:  "five_star",
				Label: "5 Star Reviews",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Review], value interface{}) orm.QuerySet[Review] {
					return qs.Filter(orm.F("rating").Eq(5))
				},
			},
		},
		Actions: []admin.Action[Review]{
			{
				Name:  "approve",
				Label: "Approve Reviews",
				Handler: func(ctx context.Context, instances []*Review) error {
					for _, review := range instances {
						review.Status = "approved"
						if err := ReviewObjects.Update(ctx, review); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "reject",
				Label: "Reject Reviews",
				Handler: func(ctx context.Context, instances []*Review) error {
					for _, review := range instances {
						review.Status = "rejected"
						if err := ReviewObjects.Update(ctx, review); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "feature",
				Label: "Feature Reviews",
				Handler: func(ctx context.Context, instances []*Review) error {
					for _, review := range instances {
						review.IsFeatured = true
						if err := ReviewObjects.Update(ctx, review); err != nil {
							return err
						}
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
		Fieldsets: []admin.Fieldset[ReviewImage]{
			{
				Name: "Image Information",
				Fields: []string{"review_id", "image_url", "thumbnail_url"},
			},
			{
				Name:  "Display",
				Fields: []string{"alt_text", "sort_order"},
			},
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ReviewImage], user interface{}, obj *ReviewImage) bool {
			return true
		},
		Actions: []admin.Action[ReviewImage]{
			{
				Name:  "reset_sort",
				Label: "Reset Sort Order",
				Handler: func(ctx context.Context, instances []*ReviewImage) error {
					for _, image := range instances {
						image.SortOrder = 0
						if err := ReviewImageObjects.Update(ctx, image); err != nil {
							return err
						}
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
			ReviewHelpfulnessFieldsInstance.ReviewId,
		},
		Fieldsets: []admin.Fieldset[ReviewHelpfulness]{
			{
				Name: "Vote Details",
				Fields: []string{"review_id", "customer_id", "is_helpful"},
			},
			{
				Name:  "Metadata",
				Fields: []string{"ip_address", "created_at"},
			},
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ReviewHelpfulness], user interface{}, obj *ReviewHelpfulness) bool {
			return true
		},
		Actions: []admin.Action[ReviewHelpfulness]{
			{
				Name:  "mark_helpful",
				Label: "Mark Helpful",
				Handler: func(ctx context.Context, instances []*ReviewHelpfulness) error {
					for _, helpfulness := range instances {
						helpfulness.IsHelpful = true
						if err := ReviewHelpfulnessObjects.Update(ctx, helpfulness); err != nil {
							return err
						}
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
		SearchFields: []admin.Field{
			ProductQuestionFieldsInstance.Question,
			ProductQuestionFieldsInstance.Answer,
		},
		Fieldsets: []admin.Fieldset[ProductQuestion]{
			{
				Name: "Q&A",
				Fields: []string{"product_id", "customer_id", "question"},
			},
			{
				Name: "Answer",
				Fields: []string{"answer", "answered_at", "answered_by_user_id"},
			},
			{
				Name: "Status",
				Fields: []string{"status", "is_public"},
			},
			{
				Name:  "Timestamps",
				Fields: []string{"created_at", "updated_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "answer", "is_public"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ProductQuestion], user interface{}, obj *ProductQuestion) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ProductQuestion], user interface{}, obj *ProductQuestion) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ProductQuestion], user interface{}, obj *ProductQuestion) bool {
			return true
		},
		Filters: []admin.Filter[ProductQuestion]{
			{
				Name:  "pending",
				Label: "Pending Questions",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ProductQuestion], value interface{}) orm.QuerySet[ProductQuestion] {
					return qs.Filter(orm.F("status").Eq("pending"))
				},
			},
			{
				Name:  "answered",
				Label: "Answered Questions",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ProductQuestion], value interface{}) orm.QuerySet[ProductQuestion] {
					return qs.Filter(orm.F("status").Eq("answered"))
				},
			},
			{
				Name:  "unanswered",
				Label: "Unanswered Questions",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ProductQuestion], value interface{}) orm.QuerySet[ProductQuestion] {
					return qs.Filter(orm.F("answer").IsNull())
				},
			},
		},
		Actions: []admin.Action[ProductQuestion]{
			{
				Name:  "mark_answered",
				Label: "Mark Answered",
				Handler: func(ctx context.Context, instances []*ProductQuestion) error {
					for _, question := range instances {
						question.Status = "answered"
						if err := ProductQuestionObjects.Update(ctx, question); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "hide",
				Label: "Hide Questions",
				Handler: func(ctx context.Context, instances []*ProductQuestion) error {
					for _, question := range instances {
						question.Status = "hidden"
						if err := ProductQuestionObjects.Update(ctx, question); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}
