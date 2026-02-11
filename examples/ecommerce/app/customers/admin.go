package customers

import (
	"context"
	"time"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers customer models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// CustomerGroup admin
	admin.Register(&admin.Config[CustomerGroup]{
		Icon: "Users",
		ListDisplay: []admin.Field{
			CustomerGroupFieldsInstance.Name,
			CustomerGroupFieldsInstance.Code,
			CustomerGroupFieldsInstance.DiscountPercentage,
			CustomerGroupFieldsInstance.IsActive,
			CustomerGroupFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			CustomerGroupFieldsInstance.IsActive,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[CustomerGroup]{
			{
				Name: "Group Information",
				Fields: []string{"name", "code", "description"},
			},
			{
				Name: "Pricing",
				Fields: []string{"discount_percentage"},
			},
			{
				Name:  "Status",
				Fields: []string{"is_active"},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "code", "discount_percentage", "is_active"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[CustomerGroup], user interface{}, obj *CustomerGroup) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[CustomerGroup], user interface{}, obj *CustomerGroup) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[CustomerGroup], user interface{}, obj *CustomerGroup) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[CustomerGroup]{
			{
				Name:  "active_groups",
				Label: "Active Groups Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[CustomerGroup], value interface{}) orm.QuerySet[CustomerGroup] {
					return qs.Filter(orm.F("is_active").Eq(true))
				},
			},
		},
		// Custom actions
		Actions: []admin.Action[CustomerGroup]{
			{
				Name:  "activate",
				Label: "Activate Groups",
				Handler: func(ctx context.Context, instances []*CustomerGroup) error {
					for _, group := range instances {
						group.IsActive = true
						if err := CustomerGroupObjects.Update(ctx, group); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "deactivate",
				Label: "Deactivate Groups",
				Handler: func(ctx context.Context, instances []*CustomerGroup) error {
					for _, group := range instances {
						group.IsActive = false
						if err := CustomerGroupObjects.Update(ctx, group); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// Customer admin
	admin.Register(&admin.Config[Customer]{
		Icon: "User",
		ListDisplay: []admin.Field{
			CustomerFieldsInstance.Email,
			CustomerFieldsInstance.FirstName,
			CustomerFieldsInstance.LastName,
			CustomerFieldsInstance.IsActive,
			CustomerFieldsInstance.IsVerified,
			CustomerFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			CustomerFieldsInstance.IsActive,
			CustomerFieldsInstance.IsVerified,
			CustomerFieldsInstance.CustomerGroupId,
		},
		SearchFields: []admin.Field{
			CustomerFieldsInstance.Email,
			CustomerFieldsInstance.FirstName,
			CustomerFieldsInstance.LastName,
			CustomerFieldsInstance.Phone,
		},
		// Fieldsets for comprehensive customer management
		Fieldsets: []admin.Fieldset[Customer]{
			{
				Name: "Personal Information",
				Fields: []string{"email", "first_name", "last_name", "phone", "date_of_birth", "gender"},
			},
			{
				Name: "Account",
				Fields: []string{"customer_group_id", "is_active", "is_verified", "accepts_marketing"},
			},
			{
				Name: "Business Information",
				Fields: []string{"company_name", "tax_id"},
				Collapsed: true,
			},
			{
				Name: "Security",
				Fields: []string{"last_login_at", "last_login_ip"},
				Collapsed: true,
				Description: "Security and access information (read-only)",
			},
			{
				Name: "Statistics",
				Fields: []string{"total_orders", "total_spent", "average_order_value"},
				Collapsed: true,
				Description: "Customer statistics (read-only)",
			},
			{
				Name:  "Preferences",
				Fields: []string{"preferred_language", "preferred_currency"},
			},
			{
				Name:  "Notes",
				Fields: []string{"notes"},
			},
		},
		// Inline relations for addresses and wishlists
		InlineRelations: map[string]admin.InlineRelationConfig{
			"addresses": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Customer Addresses",
				RelatedModel: "Address",
				RelatedField: "customer_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"address_type", "city", "state_province", "country_code", "is_default_shipping", "is_default_billing"},
					Fieldsets: []admin.Fieldset[Address]{
						{Name: "Contact", Fields: []string{"first_name", "last_name", "company_name", "phone"}},
						{Name: "Address", Fields: []string{"address_line1", "address_line2", "city", "state_province", "postal_code", "country_code"}},
						{Name: "Preferences", Fields: []string{"is_default_shipping", "is_default_billing", "address_type"}},
					},
				},
			},
			"wish_lists": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Wish Lists",
				RelatedModel: "WishList",
				RelatedField: "customer_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"name", "is_public", "is_default", "created_at"},
					Fieldsets: []admin.Fieldset[WishList]{
						{Name: "Wish List", Fields: []string{"name", "description", "is_public", "is_default"}},
					},
				},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"email", "first_name", "last_name", "is_active", "is_verified", "customer_group_id"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at", "password_hash", "verification_token", "reset_password_token", "last_login_at", "last_login_ip", "total_orders", "total_spent", "average_order_value"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Customer], user interface{}, obj *Customer) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Customer], user interface{}, obj *Customer) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Customer], user interface{}, obj *Customer) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[Customer]{
			{
				Name:  "verified",
				Label: "Verified Customers",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Customer], value interface{}) orm.QuerySet[Customer] {
					return qs.Filter(orm.F("is_verified").Eq(true))
				},
			},
			{
				Name:  "active_customers",
				Label: "Active Customers",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Customer], value interface{}) orm.QuerySet[Customer] {
					return qs.Filter(orm.F("is_active").Eq(true))
				},
			},
			{
				Name:  "inactive_customers",
				Label: "Inactive Customers",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Customer], value interface{}) orm.QuerySet[Customer] {
					return qs.Filter(orm.F("is_active").Eq(false))
				},
			},
			{
				Name:  "has_orders",
				Label: "Has Placed Orders",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Customer], value interface{}) orm.QuerySet[Customer] {
					return qs.Filter(orm.F("total_orders").Gt(0))
				},
			},
			{
				Name:  "vip_customers",
				Label: "VIP Customers (High Value)",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Customer], value interface{}) orm.QuerySet[Customer] {
					return qs.Filter(orm.F("total_spent").Gte(1000))
				},
			},
			{
				Name:  "marketing_subscribers",
				Label: "Marketing Subscribers",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Customer], value interface{}) orm.QuerySet[Customer] {
					return qs.Filter(orm.F("accepts_marketing").Eq(true))
				},
			},
			{
				Name:  "recent_login",
				Label: "Recently Logged In (30 days)",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Customer], value interface{}) orm.QuerySet[Customer] {
					return qs.Filter(orm.F("last_login_at").Gte(time.Now().AddDate(0, 0, -30)))
				},
			},
			{
				Name:  "never_logged_in",
				Label: "Never Logged In",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Customer], value interface{}) orm.QuerySet[Customer] {
					return qs.Filter(orm.F("last_login_at").IsNull())
				},
			},
		},
		// Custom actions
		Actions: []admin.Action[Customer]{
			{
				Name:  "activate",
				Label: "Activate Customers",
				Handler: func(ctx context.Context, instances []*Customer) error {
					for _, customer := range instances {
						customer.IsActive = true
						if err := CustomerObjects.Update(ctx, customer); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "deactivate",
				Label: "Deactivate Customers",
				Handler: func(ctx context.Context, instances []*Customer) error {
					for _, customer := range instances {
						customer.IsActive = false
						if err := CustomerObjects.Update(ctx, customer); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "verify",
				Label: "Verify Customers",
				Handler: func(ctx context.Context, instances []*Customer) error {
					for _, customer := range instances {
						customer.IsVerified = true
						if err := CustomerObjects.Update(ctx, customer); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "unverify",
				Label: "Unverify Customers",
				Handler: func(ctx context.Context, instances []*Customer) error {
					for _, customer := range instances {
						customer.IsVerified = false
						if err := CustomerObjects.Update(ctx, customer); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "toggle_marketing",
				Label: "Toggle Marketing Consent",
				Handler: func(ctx context.Context, instances []*Customer) error {
					for _, customer := range instances {
						customer.AcceptsMarketing = !customer.AcceptsMarketing
						if err := CustomerObjects.Update(ctx, customer); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "export_data",
				Label: "Export Customer Data",
				Handler: func(ctx context.Context, instances []*Customer) error {
					// Export logic would go here
					return nil
				},
			},
		},
	})

	// Address admin
	admin.Register(&admin.Config[Address]{
		Icon: "MapPin",
		ListDisplay: []admin.Field{
			AddressFieldsInstance.CustomerId,
			AddressFieldsInstance.AddressType,
			AddressFieldsInstance.City,
			AddressFieldsInstance.StateProvince,
			AddressFieldsInstance.CountryCode,
			AddressFieldsInstance.IsDefaultShipping,
			AddressFieldsInstance.IsDefaultBilling,
		},
		ListFilter: []admin.Field{
			AddressFieldsInstance.AddressType,
			AddressFieldsInstance.CountryCode,
			AddressFieldsInstance.IsDefaultShipping,
			AddressFieldsInstance.IsDefaultBilling,
		},
		SearchFields: []admin.Field{
			AddressFieldsInstance.FirstName,
			AddressFieldsInstance.LastName,
			AddressFieldsInstance.City,
			AddressFieldsInstance.PostalCode,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[Address]{
			{
				Name: "Customer",
				Fields: []string{"customer_id"},
			},
			{
				Name: "Contact Information",
				Fields: []string{"first_name", "last_name", "company_name", "phone"},
			},
			{
				Name: "Address",
				Fields: []string{"address_line1", "address_line2", "city", "state_province", "postal_code", "country_code", "country_name"},
			},
			{
				Name: "Geolocation",
				Fields: []string{"latitude", "longitude"},
				Collapsed: true,
			},
			{
				Name: "Preferences",
				Fields: []string{"is_default_shipping", "is_default_billing", "address_type"},
			},
			{
				Name:  "Delivery",
				Fields: []string{"delivery_instructions"},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"first_name", "last_name", "address_line1", "city", "country_code", "is_default_shipping", "is_default_billing"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Address], user interface{}, obj *Address) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Address], user interface{}, obj *Address) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Address], user interface{}, obj *Address) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[Address]{
			{
				Name:  "shipping_addresses",
				Label: "Shipping Addresses",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "any", Label: "Any"},
					{Value: "default", Label: "Default Only"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[Address], value interface{}) orm.QuerySet[Address] {
					if value == "default" {
						return qs.Filter(orm.F("is_default_shipping").Eq(true))
					}
					return qs
				},
			},
			{
				Name:  "billing_addresses",
				Label: "Billing Addresses",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "any", Label: "Any"},
					{Value: "default", Label: "Default Only"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[Address], value interface{}) orm.QuerySet[Address] {
					if value == "default" {
						return qs.Filter(orm.F("is_default_billing").Eq(true))
					}
					return qs
				},
			},
			{
				Name:  "by_country",
				Label: "By Country",
				Type:  admin.FilterTypeText,
				Handler: func(ctx context.Context, qs orm.QuerySet[Address], value interface{}) orm.QuerySet[Address] {
					return qs.Filter(orm.F("country_code").Eq(value))
				},
			},
		},
		// Custom actions
		Actions: []admin.Action[Address]{
			{
				Name:  "set_default_shipping",
				Label: "Set as Default Shipping",
				Handler: func(ctx context.Context, instances []*Address) error {
					for _, address := range instances {
						address.IsDefaultShipping = true
						if err := AddressObjects.Update(ctx, address); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "set_default_billing",
				Label: "Set as Default Billing",
				Handler: func(ctx context.Context, instances []*Address) error {
					for _, address := range instances {
						address.IsDefaultBilling = true
						if err := AddressObjects.Update(ctx, address); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "change_type_shipping",
				Label: "Change to Shipping",
				Handler: func(ctx context.Context, instances []*Address) error {
					for _, address := range instances {
						address.AddressType = "shipping"
						if err := AddressObjects.Update(ctx, address); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "change_type_billing",
				Label: "Change to Billing",
				Handler: func(ctx context.Context, instances []*Address) error {
					for _, address := range instances {
						address.AddressType = "billing"
						if err := AddressObjects.Update(ctx, address); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// WishList admin
	admin.Register(&admin.Config[WishList]{
		Icon: "Heart",
		ListDisplay: []admin.Field{
			WishListFieldsInstance.CustomerId,
			WishListFieldsInstance.Name,
			WishListFieldsInstance.IsPublic,
			WishListFieldsInstance.IsDefault,
			WishListFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			WishListFieldsInstance.IsPublic,
			WishListFieldsInstance.IsDefault,
		},
		SearchFields: []admin.Field{
			WishListFieldsInstance.Name,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[WishList]{
			{
				Name: "Wish List",
				Fields: []string{"customer_id", "name", "description"},
			},
			{
				Name: "Visibility",
				Fields: []string{"is_public", "is_default", "share_token"},
			},
		},
		// Inline relations for wishlist items
		InlineRelations: map[string]admin.InlineRelationConfig{
			"items": {
				Type:         admin.InlineTypeOneToMany,
				Label:        "Wish List Items",
				RelatedModel: "WishListItem",
				RelatedField: "wish_list_id",
				InlineConfig: admin.InlineConfig{
					ListDisplay: []string{"product_id", "variant_id", "desired_quantity", "priority", "price_when_added"},
					Fieldsets: []admin.Fieldset[WishListItem]{
						{Name: "Product", Fields: []string{"product_id", "variant_id"}},
						{Name: "Details", Fields: []string{"desired_quantity", "priority", "price_when_added", "notes"}},
					},
				},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "description", "is_public", "is_default"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at", "share_token"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[WishList], user interface{}, obj *WishList) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[WishList], user interface{}, obj *WishList) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[WishList], user interface{}, obj *WishList) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[WishList]{
			{
				Name:  "public_lists",
				Label: "Public Wish Lists",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[WishList], value interface{}) orm.QuerySet[WishList] {
					return qs.Filter(orm.F("is_public").Eq(true))
				},
			},
			{
				Name:  "default_lists",
				Label: "Default Wish Lists",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[WishList], value interface{}) orm.QuerySet[WishList] {
					return qs.Filter(orm.F("is_default").Eq(true))
				},
			},
			{
				Name:  "shared_lists",
				Label: "Shared Wish Lists",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[WishList], value interface{}) orm.QuerySet[WishList] {
					return qs.Filter(orm.F("share_token").IsNotNull())
				},
			},
		},
		// Custom actions
		Actions: []admin.Action[WishList]{
			{
				Name:  "make_public",
				Label: "Make Public",
				Handler: func(ctx context.Context, instances []*WishList) error {
					for _, wishList := range instances {
						wishList.IsPublic = true
						if err := WishListObjects.Update(ctx, wishList); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "make_private",
				Label: "Make Private",
				Handler: func(ctx context.Context, instances []*WishList) error {
					for _, wishList := range instances {
						wishList.IsPublic = false
						if err := WishListObjects.Update(ctx, wishList); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "set_default",
				Label: "Set as Default",
				Handler: func(ctx context.Context, instances []*WishList) error {
					for _, wishList := range instances {
						wishList.IsDefault = true
						if err := WishListObjects.Update(ctx, wishList); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "generate_share_token",
				Label: "Generate Share Token",
				Handler: func(ctx context.Context, instances []*WishList) error {
					for _, wishList := range instances {
						// Generate share token logic
						if err := WishListObjects.Update(ctx, wishList); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})

	// WishListItem admin
	admin.Register(&admin.Config[WishListItem]{
		Icon: "ShoppingBag",
		ListDisplay: []admin.Field{
			WishListItemFieldsInstance.WishListId,
			WishListItemFieldsInstance.ProductId,
			WishListItemFieldsInstance.VariantId,
			WishListItemFieldsInstance.DesiredQuantity,
			WishListItemFieldsInstance.Priority,
			WishListItemFieldsInstance.PriceWhenAdded,
		},
		ListFilter: []admin.Field{
			WishListItemFieldsInstance.WishListId,
			WishListItemFieldsInstance.ProductId,
			WishListItemFieldsInstance.Priority,
		},
		// Fieldsets
		Fieldsets: []admin.Fieldset[WishListItem]{
			{
				Name: "Item",
				Fields: []string{"wish_list_id", "product_id", "variant_id"},
			},
			{
				Name: "Details",
				Fields: []string{"desired_quantity", "priority", "price_when_added", "notes"},
			},
		},
		// History manager
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"desired_quantity", "priority"},
		},
		// Read-only fields
		ReadOnlyFields: []string{"created_at", "updated_at", "price_when_added"},
		// Permission checkers
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[WishListItem], user interface{}, obj *WishListItem) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[WishListItem], user interface{}, obj *WishListItem) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[WishListItem], user interface{}, obj *WishListItem) bool {
			return true
		},
		// Custom filters
		Filters: []admin.Filter[WishListItem]{
			{
				Name:  "high_priority",
				Label: "High Priority Items",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[WishListItem], value interface{}) orm.QuerySet[WishListItem] {
					return qs.Filter(orm.F("priority").Gt(0))
				},
			},
		},
		// Custom actions
		Actions: []admin.Action[WishListItem]{
			{
				Name:  "increase_priority",
				Label: "Increase Priority",
				Handler: func(ctx context.Context, instances []*WishListItem) error {
					for _, item := range instances {
						item.Priority++
						if err := WishListItemObjects.Update(ctx, item); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "reset_priority",
				Label: "Reset Priority",
				Handler: func(ctx context.Context, instances []*WishListItem) error {
					for _, item := range instances {
						item.Priority = 0
						if err := WishListItemObjects.Update(ctx, item); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}
