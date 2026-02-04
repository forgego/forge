package customers

import (
	"context"

	"github.com/forgego/forge/admin"
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
		Actions: []admin.Action[CustomerGroup]{
			{
				Name:  "activate",
				Label: "Activate Groups",
				Handler: func(ctx context.Context, instances []*CustomerGroup) error {
					for _, group := range instances {
						group.IsActive = true
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
		},
		SearchFields: []admin.Field{
			CustomerFieldsInstance.Email,
			CustomerFieldsInstance.FirstName,
			CustomerFieldsInstance.LastName,
		},
		Actions: []admin.Action[Customer]{
			{
				Name:  "activate",
				Label: "Activate Customers",
				Handler: func(ctx context.Context, instances []*Customer) error {
					for _, customer := range instances {
						customer.IsActive = true
					}
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
		},
		ListFilter: []admin.Field{
			AddressFieldsInstance.AddressType,
			AddressFieldsInstance.CountryCode,
		},
		Actions: []admin.Action[Address]{
			{
				Name:  "set_default_shipping",
				Label: "Set as Default Shipping",
				Handler: func(ctx context.Context, instances []*Address) error {
					for _, address := range instances {
						address.IsDefaultShipping = true
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
		Actions: []admin.Action[WishList]{
			{
				Name:  "make_public",
				Label: "Make Public",
				Handler: func(ctx context.Context, instances []*WishList) error {
					for _, wishList := range instances {
						wishList.IsPublic = true
					}
					return nil
				},
			},
		},
	})
}
