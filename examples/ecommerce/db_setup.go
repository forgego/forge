package main

import (
	"context"
	"log"

	"examples/ecommerce/app/dbsetup"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/stores"
)

// SetupSchema creates the database schema for the ecommerce example
func SetupSchema(database *db.DB) {
	dbsetup.SetupSchema(database)
	if err := stores.Migrate(context.Background(), database); err != nil {
		log.Printf("Warning: framework store migrations: %v", err)
	}
}
