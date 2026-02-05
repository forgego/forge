package customers

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers customer API endpoints
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// CustomerGroup API
	router.Register("customer-groups", &api.ViewSetConfig{
		Model:        &CustomerGroup{},
		Queryset:     CustomerGroupObjects,
		Serializer:   &CustomerGroupSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "name", "code", "discount_percentage", "is_active"},
		DetailFields: []string{"id", "name", "code", "description", "discount_percentage", "is_active", "created_at", "updated_at"},
		Filterable:   []string{"is_active"},
		Searchable:   []string{"name", "code"},
		Ordering:     []string{"name", "created_at"},
		PerPage:      20,
	})

	// Customer API
	router.Register("customers", &api.ViewSetConfig{
		Model:        &Customer{},
		Queryset:     CustomerObjects,
		Serializer:   &CustomerSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "email", "first_name", "last_name", "phone", "customer_group_id", "total_orders", "total_spent", "is_active", "is_verified"},
		DetailFields: []string{"id", "email", "first_name", "last_name", "phone", "date_of_birth", "gender", "company_name", "tax_id", "customer_group_id", "is_active", "is_verified", "accepts_marketing", "total_orders", "total_spent", "average_order_value", "preferred_language", "preferred_currency", "last_login_at", "created_at", "updated_at"},
		Filterable:   []string{"is_active", "is_verified", "customer_group_id", "accepts_marketing"},
		Searchable:   []string{"email", "first_name", "last_name", "company_name"},
		Ordering:     []string{"first_name", "last_name", "created_at", "-total_spent"},
		PerPage:      20,
	})

	// Address API
	router.Register("addresses", &api.ViewSetConfig{
		Model:        &Address{},
		Queryset:     AddressObjects,
		Serializer:   &AddressSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "customer_id", "address_type", "first_name", "last_name", "address_line1", "city", "country_code", "is_default_shipping", "is_default_billing"},
		DetailFields: []string{"id", "customer_id", "address_type", "first_name", "last_name", "company_name", "phone", "address_line1", "address_line2", "city", "state_province", "postal_code", "country_code", "country_name", "latitude", "longitude", "is_default_shipping", "is_default_billing", "delivery_instructions", "created_at", "updated_at"},
		Filterable:   []string{"customer_id", "address_type", "country_code", "is_default_shipping", "is_default_billing"},
		Searchable:   []string{"first_name", "last_name", "address_line1", "city", "postal_code"},
		Ordering:     []string{"customer_id", "-is_default_shipping", "-is_default_billing"},
		PerPage:      20,
	})

	// WishList API
	router.Register("wish-lists", &api.ViewSetConfig{
		Model:        &WishList{},
		Queryset:     WishListObjects,
		Serializer:   &WishListSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "customer_id", "name", "is_public", "is_default", "created_at"},
		DetailFields: []string{"id", "customer_id", "name", "description", "is_public", "share_token", "is_default", "created_at", "updated_at"},
		Filterable:   []string{"customer_id", "is_public", "is_default"},
		Searchable:   []string{"name", "description"},
		Ordering:     []string{"customer_id", "-is_default", "-created_at"},
		PerPage:      20,
	})

	// WishListItem API
	router.Register("wish-list-items", &api.ViewSetConfig{
		Model:        &WishListItem{},
		Queryset:     WishListItemObjects,
		Serializer:   &WishListItemSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:   []string{"id", "wish_list_id", "product_id", "variant_id", "desired_quantity", "priority", "created_at"},
		DetailFields: []string{"id", "wish_list_id", "product_id", "variant_id", "desired_quantity", "price_when_added", "notes", "priority", "created_at", "updated_at"},
		Filterable:   []string{"wish_list_id", "product_id", "priority"},
		Searchable:   []string{"notes"},
		Ordering:     []string{"wish_list_id", "-priority", "-created_at"},
		PerPage:      50,
	})
}

// Serializers

type CustomerGroupSerializer struct {
	*api.BaseSerializer
}

func (s *CustomerGroupSerializer) New() api.Serializer {
	return &CustomerGroupSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type CustomerSerializer struct {
	*api.BaseSerializer
}

func (s *CustomerSerializer) New() api.Serializer {
	return &CustomerSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type AddressSerializer struct {
	*api.BaseSerializer
}

func (s *AddressSerializer) New() api.Serializer {
	return &AddressSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type WishListSerializer struct {
	*api.BaseSerializer
}

func (s *WishListSerializer) New() api.Serializer {
	return &WishListSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type WishListItemSerializer struct {
	*api.BaseSerializer
}

func (s *WishListItemSerializer) New() api.Serializer {
	return &WishListItemSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}
