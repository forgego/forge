package users

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers users API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("users", &api.ViewSetConfig{
		Model:      &User{},
		Queryset:   UserObjects,
		Serializer: base,
	})

	router.Register("groups", &api.ViewSetConfig{
		Model:      &Group{},
		Queryset:   GroupObjects,
		Serializer: base,
	})
}
