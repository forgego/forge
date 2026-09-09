package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx := context.Background()
	log.Println("🌱 Seeding Forge Ecommerce Database...")

	cfg := config.NewConfig()
	sqlitePath := cfg.GetString("database.sqlite_path", filepath.Join(".", "ecommerce.sqlite"))
	driver := cfg.GetDriver()

	var dsn string
	if driver == "postgres" || driver == "postgresql" {
		dbHost := cfg.GetString("database.host", "localhost")
		dbPort := cfg.GetInt("database.port", 5432)
		dbUser := cfg.GetString("database.user", "postgres")
		dbPassword := cfg.GetString("database.password", "")
		dbSSLMode := cfg.GetString("database.sslmode", "disable")
		dbName := cfg.GetString("database.name", "forge_ecommerce")
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	} else {
		dsn = sqlitePath
	}

	database, err := db.NewDB(dsn)
	if err != nil {
		log.Printf("Falling back to SQLite at %s: %v", sqlitePath, err)
		database, err = db.NewDB(sqlitePath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
	}
	defer database.Close()

	// Seed categories
	categories := []struct {
		name, slug, description string
		sortOrder               int
	}{
		{"Electronics", "electronics", "Electronic devices and gadgets", 1},
		{"Clothing", "clothing", "Men and Women apparel", 2},
		{"Home & Kitchen", "home-kitchen", "Furniture, cookware, and appliances", 3},
		{"Books", "books", "Physical and digital reading material", 4},
	}

	for _, c := range categories {
		_, err := database.ExecContext(ctx, `
			INSERT INTO categories (name, slug, description, sort_order, is_active)
			VALUES ($1, $2, $3, $4, true)
			ON CONFLICT (slug) DO NOTHING;
		`, c.name, c.slug, c.description, c.sortOrder)
		if err != nil && !strings.Contains(err.Error(), "syntax error") {
			_, _ = database.ExecContext(ctx, `
				INSERT OR IGNORE INTO categories (name, slug, description, sort_order, is_active)
				VALUES (?, ?, ?, ?, 1);
			`, c.name, c.slug, c.description, c.sortOrder)
		}
	}

	// Seed brands
	brands := []struct {
		name, slug, description string
	}{
		{"Acme Corp", "acme-corp", "Universal supplier of roadrunner gear"},
		{"Stark Tech", "stark-tech", "Advanced consumer electronics"},
		{"Wayne Enterprises", "wayne-enterprises", "High reliability lifestyle goods"},
	}

	for _, b := range brands {
		_, err := database.ExecContext(ctx, `
			INSERT INTO brands (name, slug, description, is_active)
			VALUES ($1, $2, $3, true)
			ON CONFLICT (slug) DO NOTHING;
		`, b.name, b.slug, b.description)
		if err != nil && !strings.Contains(err.Error(), "syntax error") {
			_, _ = database.ExecContext(ctx, `
				INSERT OR IGNORE INTO brands (name, slug, description, is_active)
				VALUES (?, ?, ?, 1);
			`, b.name, b.slug, b.description)
		}
	}

	log.Println("✅ Database seeded successfully!")
}
