package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"examples/ecommerce/app/dbsetup"
	"examples/ecommerce/app/seeder"
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

	// Ensure schema exists first
	dbsetup.SetupSchema(database)

	// Run full data seeder
	if err := seeder.Seed(ctx, database); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("🎉 Database seeded successfully with all models and relations!")
}
