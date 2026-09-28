package testutils

import (
	"database/sql"
	"testing"
)

// CreateSchema executes the given SQL schema
func CreateSchema(t *testing.T, database *sql.DB, schemaSQL string) {
	_, err := database.Exec(schemaSQL)
	if err != nil {
		t.Fatalf("Failed to create schema: %v\nSQL: %s", err, schemaSQL)
	}
}

// Common schemas for testing
const (
	UserSchema = `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL UNIQUE,
			email VARCHAR(254) NOT NULL UNIQUE,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	CategorySchema = `
		CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			parent_id INTEGER REFERENCES categories(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	ProductSchema = `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			category_id INTEGER REFERENCES categories(id),
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	TestModelSchema = `
        CREATE TABLE IF NOT EXISTS test_models (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT,
            price DECIMAL(10, 2) DEFAULT 0.0,
            available BOOLEAN DEFAULT TRUE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );
    `
)

// Fixture helpers

func CreateTestTable(t *testing.T, database *sql.DB) {
	CreateSchema(t, database, TestModelSchema)
}

func CreateUserTable(t *testing.T, database *sql.DB) {
	CreateSchema(t, database, UserSchema)
}

func CreateCategoryTable(t *testing.T, database *sql.DB) {
	CreateSchema(t, database, CategorySchema)
}

func CreateProductTable(t *testing.T, database *sql.DB) {
	CreateSchema(t, database, ProductSchema)
}

func CreateEcommerceTables(t *testing.T, database *sql.DB) {
	CreateCategoryTable(t, database)
	CreateProductTable(t, database)
}
