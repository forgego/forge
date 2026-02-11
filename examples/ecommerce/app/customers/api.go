package customers

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers customer API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("customer-groups", &api.ViewSetConfig{
		Model:      &CustomerGroup{},
		Queryset:   CustomerGroupObjects,
		Serializer: base,
	})

	router.Register("customers", &api.ViewSetConfig{
		Model:      &Customer{},
		Queryset:   CustomerObjects,
		Serializer: base,
	})

	router.Register("addresses", &api.ViewSetConfig{
		Model:      &Address{},
		Queryset:   AddressObjects,
		Serializer: base,
	})

	router.Register("wish-lists", &api.ViewSetConfig{
		Model:      &WishList{},
		Queryset:   WishListObjects,
		Serializer: base,
	})

	router.Register("wish-list-items", &api.ViewSetConfig{
		Model:      &WishListItem{},
		Queryset:   WishListItemObjects,
		Serializer: base,
	})
}
