package main

import (
	"examples/ecommerce/app/dbsetup"
	"github.com/forgego/forge/db"
)

// SetupSchema creates the database schema for the ecommerce example
func SetupSchema(database *db.DB) {
	dbsetup.SetupSchema(database)
}
