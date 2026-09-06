package promotions

import (
	"context"

	"github.com/forgego/forge/admin"
)

// RegisterAdmin registers all promotions models with the admin interface
func RegisterAdmin(ctx context.Context) error {
	// Promotion Admin
	_, err := admin.Register(&admin.Config[Promotion]{
		ListDisplay: []admin.Field{
			PromotionFieldsInstance.Id,
			PromotionFieldsInstance.Name,
			PromotionFieldsInstance.Code,
			PromotionFieldsInstance.DiscountType,
			PromotionFieldsInstance.DiscountValue,
			PromotionFieldsInstance.IsActive,
			PromotionFieldsInstance.TimesUsed,
			PromotionFieldsInstance.StartDate,
			PromotionFieldsInstance.EndDate,
		},
		SearchFields: []admin.Field{
			PromotionFieldsInstance.Name,
			PromotionFieldsInstance.Code,
			PromotionFieldsInstance.Description,
		},
		ListFilter: []admin.Field{
			PromotionFieldsInstance.IsActive,
			PromotionFieldsInstance.DiscountType,
			PromotionFieldsInstance.CanStack,
		},
		Ordering: []admin.Field{
			PromotionFieldsInstance.Priority,
			PromotionFieldsInstance.CreatedAt,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// PromotionRule Admin
	_, err = admin.Register(&admin.Config[PromotionRule]{
		ListDisplay: []admin.Field{
			PromotionRuleFieldsInstance.Id,
			PromotionRuleFieldsInstance.PromotionId,
			PromotionRuleFieldsInstance.RuleType,
			PromotionRuleFieldsInstance.Parameters,
			PromotionRuleFieldsInstance.Logic,
		},
		SearchFields: []admin.Field{
			PromotionRuleFieldsInstance.RuleType,
			PromotionRuleFieldsInstance.Parameters,
		},
		ListFilter: []admin.Field{
			PromotionRuleFieldsInstance.RuleType,
			PromotionRuleFieldsInstance.IsActive,
		},
		Ordering: []admin.Field{
			PromotionRuleFieldsInstance.PromotionId,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// Banner Admin
	_, err = admin.Register(&admin.Config[Banner]{
		ListDisplay: []admin.Field{
			BannerFieldsInstance.Id,
			BannerFieldsInstance.Title,
			BannerFieldsInstance.Placement,
			BannerFieldsInstance.IsActive,
			BannerFieldsInstance.ViewCount,
			BannerFieldsInstance.ClickCount,
			BannerFieldsInstance.StartDate,
			BannerFieldsInstance.EndDate,
		},
		SearchFields: []admin.Field{
			BannerFieldsInstance.Title,
			BannerFieldsInstance.Description,
		},
		ListFilter: []admin.Field{
			BannerFieldsInstance.IsActive,
			BannerFieldsInstance.Placement,
			BannerFieldsInstance.TargetGroup,
		},
		Ordering: []admin.Field{
			BannerFieldsInstance.SortOrder,
			BannerFieldsInstance.CreatedAt,
		},
		ListPerPage: 50,
	})
	if err != nil {
		return err
	}

	// NewsletterSubscription Admin
	_, err = admin.Register(&admin.Config[NewsletterSubscription]{
		ListDisplay: []admin.Field{
			NewsletterSubscriptionFieldsInstance.Id,
			NewsletterSubscriptionFieldsInstance.Email,
			NewsletterSubscriptionFieldsInstance.FirstName,
			NewsletterSubscriptionFieldsInstance.LastName,
			NewsletterSubscriptionFieldsInstance.Status,
			NewsletterSubscriptionFieldsInstance.Source,
			NewsletterSubscriptionFieldsInstance.ConfirmedAt,
			NewsletterSubscriptionFieldsInstance.CreatedAt,
		},
		SearchFields: []admin.Field{
			NewsletterSubscriptionFieldsInstance.Email,
			NewsletterSubscriptionFieldsInstance.FirstName,
			NewsletterSubscriptionFieldsInstance.LastName,
		},
		ListFilter: []admin.Field{
			NewsletterSubscriptionFieldsInstance.Status,
			NewsletterSubscriptionFieldsInstance.Source,
		},
		Ordering: []admin.Field{
			NewsletterSubscriptionFieldsInstance.CreatedAt,
		},
		ListPerPage: 100,
	})
	if err != nil {
		return err
	}

	// PromotionUsage Admin
	_, err = admin.Register(&admin.Config[PromotionUsage]{
		ListDisplay: []admin.Field{
			PromotionUsageFieldsInstance.Id,
			PromotionUsageFieldsInstance.PromotionId,
			PromotionUsageFieldsInstance.CustomerId,
			PromotionUsageFieldsInstance.OrderId,
			PromotionUsageFieldsInstance.DiscountAmount,
			PromotionUsageFieldsInstance.UsedAt,
		},
		SearchFields: []admin.Field{
			PromotionUsageFieldsInstance.PromotionId,
			PromotionUsageFieldsInstance.CustomerId,
			PromotionUsageFieldsInstance.OrderId,
		},
		ListFilter: []admin.Field{
			PromotionUsageFieldsInstance.UsedAt,
		},
		Ordering: []admin.Field{
			PromotionUsageFieldsInstance.UsedAt,
		},
		ListPerPage: 100,
	})
	if err != nil {
		return err
	}

	return nil
}
