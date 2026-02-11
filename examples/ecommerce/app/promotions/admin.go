package promotions

import (
	"context"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers promotions models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// Promotion admin
	admin.Register(&admin.Config[Promotion]{
		Icon: "Megaphone",
		ListDisplay: []admin.Field{
			PromotionFieldsInstance.Name,
			PromotionFieldsInstance.Code,
			PromotionFieldsInstance.Type,
			PromotionFieldsInstance.DiscountValue,
			PromotionFieldsInstance.IsActive,
			PromotionFieldsInstance.StartDate,
			PromotionFieldsInstance.EndDate,
		},
		ListFilter: []admin.Field{
			PromotionFieldsInstance.Type,
			PromotionFieldsInstance.IsActive,
			PromotionFieldsInstance.AppliesTo,
		},
		SearchFields: []admin.Field{
			PromotionFieldsInstance.Name,
			PromotionFieldsInstance.Code,
			PromotionFieldsInstance.Description,
		},
		Fieldsets: []admin.Fieldset[Promotion]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "code", "description"},
			},
			{
				Name: "Discount Rules",
				Fields: []string{"type", "discount_value", "discount_type", "min_purchase", "max_discount"},
			},
			{
				Name: "Buy X Get Y",
				Fields: []string{"buy_quantity", "get_quantity", "free_product_id"},
				Collapsed: true,
			},
			{
				Name: "Free Shipping",
				Fields: []string{"free_shipping"},
			},
			{
				Name: "Applicability",
				Fields: []string{"applies_to", "product_ids", "category_ids", "brand_ids"},
				Collapsed: true,
			},
			{
				Name: "Customer Restrictions",
				Fields: []string{"new_customers_only", "customer_group_ids"},
				Collapsed: true,
			},
			{
				Name: "Usage Limits",
				Fields: []string{"total_usage_limit", "per_customer_limit", "times_used"},
			},
			{
				Name: "Validity Period",
				Fields: []string{"start_date", "end_date"},
			},
			{
				Name: "Priority & Stacking",
				Fields: []string{"priority", "can_stack", "stack_with"},
				Collapsed: true,
			},
			{
				Name: "Status",
				Fields: []string{"is_active"},
			},
		},
		InlineRelations: map[string]admin.InlineRelationConfig{
			"rules": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Promotion Rules",
				RelatedModel: "PromotionRule",
				RelatedField: "promotion_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"rule_type", "logic", "is_active"},
				},
			},
			"usage_records": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Usage Records",
				RelatedModel: "PromotionUsage",
				RelatedField: "promotion_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"order_id", "customer_id", "discount_amount", "used_at"},
				},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "code", "type", "discount_value", "is_active", "start_date", "end_date"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "times_used"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Promotion], user interface{}, obj *Promotion) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[Promotion], user interface{}, obj *Promotion) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Promotion], user interface{}, obj *Promotion) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Promotion], user interface{}, obj *Promotion) bool {
			return true
		},
		Filters: []admin.Filter[Promotion]{
			{
				Name:  "active_promotions",
				Label: "Active Promotions",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Promotion], value interface{}) orm.QuerySet[Promotion] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "expired_promotions",
				Label: "Expired Promotions",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Promotion], value interface{}) orm.QuerySet[Promotion] {
					return qs.Filter("end_date__lt", "now")
				},
			},
			{
				Name:  "upcoming_promotions",
				Label: "Upcoming Promotions",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Promotion], value interface{}) orm.QuerySet[Promotion] {
					return qs.Filter("start_date__gt", "now")
				},
			},
			{
				Name:  "percentage_discounts",
				Label: "Percentage Discounts",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Promotion], value interface{}) orm.QuerySet[Promotion] {
					return qs.Filter("discount_type", "percentage")
				},
			},
			{
				Name:  "free_shipping",
				Label: "Free Shipping",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Promotion], value interface{}) orm.QuerySet[Promotion] {
					return qs.Filter("free_shipping", true)
				},
			},
		},
		Actions: []admin.Action[Promotion]{
			{
				Name:  "activate",
				Label: "Activate Promotions",
				Handler: func(ctx context.Context, instances []*Promotion) error {
					for _, promotion := range instances {
						promotion.IsActive = true
						if err := PromotionObjects.Update(ctx, promotion); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "deactivate",
				Label: "Deactivate Promotions",
				Handler: func(ctx context.Context, instances []*Promotion) error {
					for _, promotion := range instances {
						promotion.IsActive = false
						if err := PromotionObjects.Update(ctx, promotion); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "usage-report",
				Label: "Generate Usage Report",
				Handler: func(ctx context.Context, instances []*Promotion) error {
					// Generate usage report logic
					return nil
				},
			},
			{
				Name:  "validate",
				Label: "Validate Promotions",
				Handler: func(ctx context.Context, instances []*Promotion) error {
					for _, promotion := range instances {
						// Validate promotion logic
						_ = promotion
					}
					return nil
				},
			},
		},
	})

	// PromotionRule admin
	admin.Register(&admin.Config[PromotionRule]{
		Icon: "Scroll",
		ListDisplay: []admin.Field{
			PromotionRuleFieldsInstance.ID,
			PromotionRuleFieldsInstance.PromotionID,
			PromotionRuleFieldsInstance.RuleType,
			PromotionRuleFieldsInstance.Logic,
			PromotionRuleFieldsInstance.IsActive,
		},
		ListFilter: []admin.Field{
			PromotionRuleFieldsInstance.PromotionID,
			PromotionRuleFieldsInstance.RuleType,
			PromotionRuleFieldsInstance.IsActive,
		},
		Fieldsets: []admin.Fieldset[PromotionRule]{
			{
				Name: "Rule Details",
				Fields: []string{"promotion_id", "rule_type", "logic"},
			},
			{
				Name: "Parameters",
				Fields: []string{"parameters"},
			},
			{
				Name:  "Status",
				Fields: []string{"is_active"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"rule_type", "parameters", "logic", "is_active"},
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[PromotionRule], user interface{}, obj *PromotionRule) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[PromotionRule], user interface{}, obj *PromotionRule) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[PromotionRule], user interface{}, obj *PromotionRule) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[PromotionRule], user interface{}, obj *PromotionRule) bool {
			return true
		},
	})

	// Banner admin
	admin.Register(&admin.Config[Banner]{
		Icon: "Image",
		ListDisplay: []admin.Field{
			BannerFieldsInstance.Title,
			BannerFieldsInstance.Position,
			BannerFieldsInstance.Priority,
			BannerFieldsInstance.IsActive,
			BannerFieldsInstance.StartDate,
			BannerFieldsInstance.EndDate,
		},
		ListFilter: []admin.Field{
			BannerFieldsInstance.Position,
			BannerFieldsInstance.IsActive,
			BannerFieldsInstance.DeviceTypes,
		},
		SearchFields: []admin.Field{
			BannerFieldsInstance.Title,
			BannerFieldsInstance.Subtitle,
		},
		Fieldsets: []admin.Fieldset[Banner]{
			{
				Name: "Basic Information",
				Fields: []string{"title", "subtitle"},
			},
			{
				Name: "Media",
				Fields: []string{"image_url", "mobile_image_url", "video_url"},
			},
			{
				Name: "Content",
				Fields: []string{"content", "link", "link_text"},
			},
			{
				Name: "Placement",
				Fields: []string{"position"},
			},
			{
				Name: "Styling",
				Fields: []string{"background_color", "text_color"},
			},
			{
				Name: "Validity Period",
				Fields: []string{"start_date", "end_date"},
			},
			{
				Name: "Scheduling",
				Fields: []string{"schedule"},
			},
			{
				Name: "Targeting",
				Fields: []string{"device_types", "user_types", "customer_group_ids"},
				Collapsed: true,
			},
			{
				Name: "Priority",
				Fields: []string{"priority"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"title", "image_url", "position", "is_active", "priority"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "click_count", "view_count"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Banner], user interface{}, obj *Banner) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[Banner], user interface{}, obj *Banner) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Banner], user interface{}, obj *Banner) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Banner], user interface{}, obj *Banner) bool {
			return true
		},
		Filters: []admin.Filter[Banner]{
			{
				Name:  "active_banners",
				Label: "Active Banners",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Banner], value interface{}) orm.QuerySet[Banner] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "homepage_banners",
				Label: "Homepage Banners",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Banner], value interface{}) orm.QuerySet[Banner] {
					return qs.Filter("position__in", []string{"homepage_top", "homepage_middle", "homepage_bottom"})
				},
			},
			{
				Name:  "mobile_banners",
				Label: "Mobile Banners",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Banner], value interface{}) orm.QuerySet[Banner] {
					return qs.Filter("device_types__contains", "mobile")
				},
			},
		},
		Actions: []admin.Action[Banner]{
			{
				Name:  "stats",
				Label: "View Statistics",
				Handler: func(ctx context.Context, instances []*Banner) error {
					for _, banner := range instances {
						_ = banner
					}
					return nil
				},
			},
			{
				Name:  "schedule",
				Label: "Schedule Banner",
				Handler: func(ctx context.Context, instances []*Banner) error {
					for _, banner := range instances {
						_ = banner
					}
					return nil
				},
			},
			{
				Name:  "activate",
				Label: "Activate Banners",
				Handler: func(ctx context.Context, instances []*Banner) error {
					for _, banner := range instances {
						banner.IsActive = true
						if err := BannerObjects.Update(ctx, banner); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "deactivate",
				Label: "Deactivate Banners",
				Handler: func(ctx context.Context, instances []*Banner) error {
					for _, banner := range instances {
						banner.IsActive = false
						if err := BannerObjects.Update(ctx, banner); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// NewsletterSubscription admin
	admin.Register(&admin.Config[NewsletterSubscription]{
		Icon: "Mail",
		ListDisplay: []admin.Field{
			NewsletterSubscriptionFieldsInstance.Email,
			NewsletterSubscriptionFieldsInstance.Status,
			NewsletterSubscriptionFieldsInstance.ListType,
			NewsletterSubscriptionFieldsInstance.Source,
			NewsletterSubscriptionFieldsInstance.ClickCount,
			NewsletterSubscriptionFieldsInstance.OpenCount,
			NewsletterSubscriptionFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			NewsletterSubscriptionFieldsInstance.Status,
			NewsletterSubscriptionFieldsInstance.ListType,
			NewsletterSubscriptionFieldsInstance.Source,
		},
		SearchFields: []admin.Field{
			NewsletterSubscriptionFieldsInstance.Email,
			NewsletterSubscriptionFieldsInstance.Segments,
		},
		Fieldsets: []admin.Fieldset[NewsletterSubscription]{
			{
				Name: "Subscription Details",
				Fields: []string{"email", "list_type", "source"},
			},
			{
				Name: "Preferences",
				Fields: []string{"preferences"},
			},
			{
				Name: "GDPR Consent",
				Fields: []string{"consent_given", "consent_date", "consent_ip"},
			},
			{
				Name: "Status",
				Fields: []string{"status", "subscribed_at", "unsubscribed_at"},
			},
			{
				Name: "Tracking",
				Fields: []string{"click_count", "open_count"},
			},
			{
				Name: "Segments",
				Fields: []string{"segments"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "list_type", "consent_given"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "click_count", "open_count"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[NewsletterSubscription], user interface{}, obj *NewsletterSubscription) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[NewsletterSubscription], user interface{}, obj *NewsletterSubscription) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[NewsletterSubscription], user interface{}, obj *NewsletterSubscription) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[NewsletterSubscription], user interface{}, obj *NewsletterSubscription) bool {
			return true
		},
		Filters: []admin.Filter[NewsletterSubscription]{
			{
				Name:  "subscribed",
				Label: "Subscribed",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[NewsletterSubscription], value interface{}) orm.QuerySet[NewsletterSubscription] {
					return qs.Filter("status", "subscribed")
				},
			},
			{
				Name:  "unsubscribed",
				Label: "Unsubscribed",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[NewsletterSubscription], value interface{}) orm.QuerySet[NewsletterSubscription] {
					return qs.Filter("status", "unsubscribed")
				},
			},
			{
				Name:  "pending",
				Label: "Pending Confirmation",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[NewsletterSubscription], value interface{}) orm.QuerySet[NewsletterSubscription] {
					return qs.Filter("status", "pending")
				},
			},
		},
		Actions: []admin.Action[NewsletterSubscription]{
			{
				Name:  "subscribe",
				Label: "Subscribe",
				Handler: func(ctx context.Context, instances []*NewsletterSubscription) error {
					for _, subscription := range instances {
						subscription.Status = StatusSubscribed
						if err := NewsletterSubscriptionObjects.Update(ctx, subscription); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "unsubscribe",
				Label: "Unsubscribe",
				Handler: func(ctx context.Context, instances []*NewsletterSubscription) error {
					for _, subscription := range instances {
						subscription.Status = StatusUnsubscribed
						if err := NewsletterSubscriptionObjects.Update(ctx, subscription); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "stats",
				Label: "View Statistics",
				Handler: func(ctx context.Context, instances []*NewsletterSubscription) error {
					for _, subscription := range instances {
						_ = subscription
					}
					return nil
				},
			},
		},
	})

	// PromotionUsage admin
	admin.Register(&admin.Config[PromotionUsage]{
		Icon: "BarChart2",
		ListDisplay: []admin.Field{
			PromotionUsageFieldsInstance.PromotionID,
			PromotionUsageFieldsInstance.OrderID,
			PromotionUsageFieldsInstance.CustomerID,
			PromotionUsageFieldsInstance.DiscountAmount,
			PromotionUsageFieldsInstance.UsedAt,
		},
		ListFilter: []admin.Field{
			PromotionUsageFieldsInstance.PromotionID,
			PromotionUsageFieldsInstance.OrderID,
			PromotionUsageFieldsInstance.CustomerID,
		},
		Fieldsets: []admin.Fieldset[PromotionUsage]{
			{
				Name: "Usage Details",
				Fields: []string{"promotion_id", "order_id", "customer_id", "discount_amount"},
			},
			{
				Name:  "Timestamp",
				Fields: []string{"used_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"discount_amount"},
		},
		ReadOnlyFields: []string{"used_at"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[PromotionUsage], user interface{}, obj *PromotionUsage) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[PromotionUsage], user interface{}, obj *PromotionUsage) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[PromotionUsage], user interface{}, obj *PromotionUsage) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[PromotionUsage], user interface{}, obj *PromotionUsage) bool {
			return true
		},
	})
}

import (
	"context"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers promotions models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// Promotion admin
	admin.Register(&admin.Config[Promotion]{
		Icon: "Megaphone",
		ListDisplay: []admin.Field{
			PromotionFieldsInstance.Name,
			PromotionFieldsInstance.Code,
			PromotionFieldsInstance.Type,
			PromotionFieldsInstance.DiscountValue,
			PromotionFieldsInstance.IsActive,
			PromotionFieldsInstance.StartDate,
			PromotionFieldsInstance.EndDate,
		},
		ListFilter: []admin.Field{
			PromotionFieldsInstance.Type,
			PromotionFieldsInstance.IsActive,
			PromotionFieldsInstance.AppliesTo,
		},
		SearchFields: []admin.Field{
			PromotionFieldsInstance.Name,
			PromotionFieldsInstance.Code,
			PromotionFieldsInstance.Description,
		},
		Fieldsets: []admin.Fieldset[Promotion]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "code", "description"},
			},
			{
				Name: "Discount Rules",
				Fields: []string{"type", "discount_value", "discount_type", "min_purchase", "max_discount"},
			},
			{
				Name: "Buy X Get Y",
				Fields: []string{"buy_quantity", "get_quantity", "free_product_id"},
				Collapsed: true,
			},
			{
				Name: "Free Shipping",
				Fields: []string{"free_shipping"},
			},
			{
				Name: "Applicability",
				Fields: []string{"applies_to", "product_ids", "category_ids", "brand_ids"},
				Collapsed: true,
			},
			{
				Name: "Customer Restrictions",
				Fields: []string{"new_customers_only", "customer_group_ids"},
				Collapsed: true,
			},
			{
				Name: "Usage Limits",
				Fields: []string{"total_usage_limit", "per_customer_limit", "times_used"},
			},
			{
				Name: "Validity Period",
				Fields: []string{"start_date", "end_date"},
			},
			{
				Name: "Priority & Stacking",
				Fields: []string{"priority", "can_stack", "stack_with"},
				Collapsed: true,
			},
			{
				Name: "Status",
				Fields: []string{"is_active"},
			},
		},
		InlineRelations: map[string]admin.InlineRelationConfig{
			"rules": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Promotion Rules",
				RelatedModel: "PromotionRule",
				RelatedField: "promotion_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"rule_type", "logic", "is_active"},
				},
			},
			"usage_records": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Usage Records",
				RelatedModel: "PromotionUsage",
				RelatedField: "promotion_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"order_id", "customer_id", "discount_amount", "used_at"},
				},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "code", "type", "discount_value", "is_active", "start_date", "end_date"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "times_used"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Promotion], user interface{}, obj *Promotion) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[Promotion], user interface{}, obj *Promotion) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Promotion], user interface{}, obj *Promotion) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Promotion], user interface{}, obj *Promotion) bool {
			return true
		},
		Filters: []admin.Filter[Promotion]{
			{
				Name:  "active_promotions",
				Label: "Active Promotions",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Promotion], value interface{}) orm.QuerySet[Promotion] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "expired_promotions",
				Label: "Expired Promotions",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Promotion], value interface{}) orm.QuerySet[Promotion] {
					return qs.Filter("end_date__lt", "now")
				},
			},
			{
				Name:  "upcoming_promotions",
				Label: "Upcoming Promotions",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Promotion], value interface{}) orm.QuerySet[Promotion] {
					return qs.Filter("start_date__gt", "now")
				},
			},
			{
				Name:  "percentage_discounts",
				Label: "Percentage Discounts",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Promotion], value interface{}) orm.QuerySet[Promotion] {
					return qs.Filter("discount_type", "percentage")
				},
			},
			{
				Name:  "free_shipping",
				Label: "Free Shipping",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Promotion], value interface{}) orm.QuerySet[Promotion] {
					return qs.Filter("free_shipping", true)
				},
			},
		},
		Actions: []admin.Action[Promotion]{
			{
				Name:  "activate",
				Label: "Activate Promotions",
				Handler: func(ctx context.Context, instances []*Promotion) error {
					for _, promotion := range instances {
						promotion.IsActive = true
						if err := PromotionObjects.Update(ctx, promotion); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "deactivate",
				Label: "Deactivate Promotions",
				Handler: func(ctx context.Context, instances []*Promotion) error {
					for _, promotion := range instances {
						promotion.IsActive = false
						if err := PromotionObjects.Update(ctx, promotion); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "usage-report",
				Label: "Generate Usage Report",
				Handler: func(ctx context.Context, instances []*Promotion) error {
					// Generate usage report logic
					return nil
				},
			},
			{
				Name:  "validate",
				Label: "Validate Promotions",
				Handler: func(ctx context.Context, instances []*Promotion) error {
					for _, promotion := range instances {
						// Validate promotion logic
						_ = promotion
					}
					return nil
				},
			},
		},
	})

	// PromotionRule admin
	admin.Register(&admin.Config[PromotionRule]{
		Icon: "Scroll",
		ListDisplay: []admin.Field{
			PromotionRuleFieldsInstance.ID,
			PromotionRuleFieldsInstance.PromotionID,
			PromotionRuleFieldsInstance.RuleType,
			PromotionRuleFieldsInstance.Logic,
			PromotionRuleFieldsInstance.IsActive,
		},
		ListFilter: []admin.Field{
			PromotionRuleFieldsInstance.PromotionID,
			PromotionRuleFieldsInstance.RuleType,
			PromotionRuleFieldsInstance.IsActive,
		},
		Fieldsets: []admin.Fieldset[PromotionRule]{
			{
				Name: "Rule Details",
				Fields: []string{"promotion_id", "rule_type", "logic"},
			},
			{
				Name: "Parameters",
				Fields: []string{"parameters"},
			},
			{
				Name:  "Status",
				Fields: []string{"is_active"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"rule_type", "parameters", "logic", "is_active"},
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[PromotionRule], user interface{}, obj *PromotionRule) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[PromotionRule], user interface{}, obj *PromotionRule) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[PromotionRule], user interface{}, obj *PromotionRule) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[PromotionRule], user interface{}, obj *PromotionRule) bool {
			return true
		},
	})

	// Banner admin
	admin.Register(&admin.Config[Banner]{
		Icon: "Image",
		ListDisplay: []admin.Field{
			BannerFieldsInstance.Title,
			BannerFieldsInstance.Position,
			BannerFieldsInstance.Priority,
			BannerFieldsInstance.IsActive,
			BannerFieldsInstance.StartDate,
			BannerFieldsInstance.EndDate,
		},
		ListFilter: []admin.Field{
			BannerFieldsInstance.Position,
			BannerFieldsInstance.IsActive,
			BannerFieldsInstance.DeviceTypes,
		},
		SearchFields: []admin.Field{
			BannerFieldsInstance.Title,
			BannerFieldsInstance.Subtitle,
		},
		Fieldsets: []admin.Fieldset[Banner]{
			{
				Name: "Basic Information",
				Fields: []string{"title", "subtitle"},
			},
			{
				Name: "Media",
				Fields: []string{"image_url", "mobile_image_url", "video_url"},
			},
			{
				Name: "Content",
				Fields: []string{"content", "link", "link_text"},
			},
			{
				Name: "Placement",
				Fields: []string{"position"},
			},
			{
				Name: "Styling",
				Fields: []string{"background_color", "text_color"},
			},
			{
				Name: "Validity Period",
				Fields: []string{"start_date", "end_date"},
			},
			{
				Name: "Scheduling",
				Fields: []string{"schedule"},
			},
			{
				Name: "Targeting",
				Fields: []string{"device_types", "user_types", "customer_group_ids"},
				Collapsed: true,
			},
			{
				Name: "Priority",
				Fields: []string{"priority"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"title", "image_url", "position", "is_active", "priority"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "click_count", "view_count"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Banner], user interface{}, obj *Banner) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[Banner], user interface{}, obj *Banner) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Banner], user interface{}, obj *Banner) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Banner], user interface{}, obj *Banner) bool {
			return true
		},
		Filters: []admin.Filter[Banner]{
			{
				Name:  "active_banners",
				Label: "Active Banners",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Banner], value interface{}) orm.QuerySet[Banner] {
					return qs.Filter("is_active", true)
				},
			},
			{
				Name:  "homepage_banners",
				Label: "Homepage Banners",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Banner], value interface{}) orm.QuerySet[Banner] {
					return qs.Filter("position__in", []string{"homepage_top", "homepage_middle", "homepage_bottom"})
				},
			},
			{
				Name:  "mobile_banners",
				Label: "Mobile Banners",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Banner], value interface{}) orm.QuerySet[Banner] {
					return qs.Filter("device_types__contains", "mobile")
				},
			},
		},
		Actions: []admin.Action[Banner]{
			{
				Name:  "stats",
				Label: "View Statistics",
				Handler: func(ctx context.Context, instances []*Banner) error {
					for _, banner := range instances {
						_ = banner
					}
					return nil
				},
			},
			{
				Name:  "schedule",
				Label: "Schedule Banner",
				Handler: func(ctx context.Context, instances []*Banner) error {
					for _, banner := range instances {
						_ = banner
					}
					return nil
				},
			},
			{
				Name:  "activate",
				Label: "Activate Banners",
				Handler: func(ctx context.Context, instances []*Banner) error {
					for _, banner := range instances {
						banner.IsActive = true
						if err := BannerObjects.Update(ctx, banner); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "deactivate",
				Label: "Deactivate Banners",
				Handler: func(ctx context.Context, instances []*Banner) error {
					for _, banner := range instances {
						banner.IsActive = false
						if err := BannerObjects.Update(ctx, banner); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// NewsletterSubscription admin
	admin.Register(&admin.Config[NewsletterSubscription]{
		Icon: "Mail",
		ListDisplay: []admin.Field{
			NewsletterSubscriptionFieldsInstance.Email,
			NewsletterSubscriptionFieldsInstance.Status,
			NewsletterSubscriptionFieldsInstance.ListType,
			NewsletterSubscriptionFieldsInstance.Source,
			NewsletterSubscriptionFieldsInstance.ClickCount,
			NewsletterSubscriptionFieldsInstance.OpenCount,
			NewsletterSubscriptionFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			NewsletterSubscriptionFieldsInstance.Status,
			NewsletterSubscriptionFieldsInstance.ListType,
			NewsletterSubscriptionFieldsInstance.Source,
		},
		SearchFields: []admin.Field{
			NewsletterSubscriptionFieldsInstance.Email,
			NewsletterSubscriptionFieldsInstance.Segments,
		},
		Fieldsets: []admin.Fieldset[NewsletterSubscription]{
			{
				Name: "Subscription Details",
				Fields: []string{"email", "list_type", "source"},
			},
			{
				Name: "Preferences",
				Fields: []string{"preferences"},
			},
			{
				Name: "GDPR Consent",
				Fields: []string{"consent_given", "consent_date", "consent_ip"},
			},
			{
				Name: "Status",
				Fields: []string{"status", "subscribed_at", "unsubscribed_at"},
			},
			{
				Name: "Tracking",
				Fields: []string{"click_count", "open_count"},
			},
			{
				Name: "Segments",
				Fields: []string{"segments"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "list_type", "consent_given"},
		},
		ReadOnlyFields: []string{"created_at", "updated_at", "click_count", "open_count"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[NewsletterSubscription], user interface{}, obj *NewsletterSubscription) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[NewsletterSubscription], user interface{}, obj *NewsletterSubscription) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[NewsletterSubscription], user interface{}, obj *NewsletterSubscription) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[NewsletterSubscription], user interface{}, obj *NewsletterSubscription) bool {
			return true
		},
		Filters: []admin.Filter[NewsletterSubscription]{
			{
				Name:  "subscribed",
				Label: "Subscribed",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[NewsletterSubscription], value interface{}) orm.QuerySet[NewsletterSubscription] {
					return qs.Filter("status", "subscribed")
				},
			},
			{
				Name:  "unsubscribed",
				Label: "Unsubscribed",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[NewsletterSubscription], value interface{}) orm.QuerySet[NewsletterSubscription] {
					return qs.Filter("status", "unsubscribed")
				},
			},
			{
				Name:  "pending",
				Label: "Pending Confirmation",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[NewsletterSubscription], value interface{}) orm.QuerySet[NewsletterSubscription] {
					return qs.Filter("status", "pending")
				},
			},
		},
		Actions: []admin.Action[NewsletterSubscription]{
			{
				Name:  "subscribe",
				Label: "Subscribe",
				Handler: func(ctx context.Context, instances []*NewsletterSubscription) error {
					for _, subscription := range instances {
						subscription.Status = StatusSubscribed
						if err := NewsletterSubscriptionObjects.Update(ctx, subscription); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "unsubscribe",
				Label: "Unsubscribe",
				Handler: func(ctx context.Context, instances []*NewsletterSubscription) error {
					for _, subscription := range instances {
						subscription.Status = StatusUnsubscribed
						if err := NewsletterSubscriptionObjects.Update(ctx, subscription); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "stats",
				Label: "View Statistics",
				Handler: func(ctx context.Context, instances []*NewsletterSubscription) error {
					for _, subscription := range instances {
						_ = subscription
					}
					return nil
				},
			},
		},
	})

	// PromotionUsage admin
	admin.Register(&admin.Config[PromotionUsage]{
		Icon: "BarChart2",
		ListDisplay: []admin.Field{
			PromotionUsageFieldsInstance.PromotionID,
			PromotionUsageFieldsInstance.OrderID,
			PromotionUsageFieldsInstance.CustomerID,
			PromotionUsageFieldsInstance.DiscountAmount,
			PromotionUsageFieldsInstance.UsedAt,
		},
		ListFilter: []admin.Field{
			PromotionUsageFieldsInstance.PromotionID,
			PromotionUsageFieldsInstance.OrderID,
			PromotionUsageFieldsInstance.CustomerID,
		},
		Fieldsets: []admin.Fieldset[PromotionUsage]{
			{
				Name: "Usage Details",
				Fields: []string{"promotion_id", "order_id", "customer_id", "discount_amount"},
			},
			{
				Name:  "Timestamp",
				Fields: []string{"used_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"discount_amount"},
		},
		ReadOnlyFields: []string{"used_at"},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[PromotionUsage], user interface{}, obj *PromotionUsage) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[PromotionUsage], user interface{}, obj *PromotionUsage) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[PromotionUsage], user interface{}, obj *PromotionUsage) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[PromotionUsage], user interface{}, obj *PromotionUsage) bool {
			return true
		},
	})
}

