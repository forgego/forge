package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/api"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/server"
	"github.com/go-chi/cors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"examples/ecommerce/app/catalog"
	"examples/ecommerce/app/customers"
	"examples/ecommerce/app/inventory"
	"examples/ecommerce/app/marketing"
	"examples/ecommerce/app/orders"
)

//go:generate forge generate

func main() {
	ctx := context.Background()
	fmt.Println("🛒 Forge Ecommerce System")
	fmt.Println("=" + string(make([]byte, 50)))

	cfg := config.NewConfig()
	dbHost := cfg.GetString("database.host", "localhost")
	dbPort := cfg.GetInt("database.port", 5432)
	dbUser := cfg.GetString("database.user", "postgres")
	dbPassword := cfg.GetString("database.password", "")
	dbSSLMode := cfg.GetString("database.sslmode", "disable")
	dbName := cfg.GetString("database.name", "")
	if dbName == "" {
		dbName = cfg.GetString("database.dbname", "forge_ecommerce")
	}

	// 1. Initialize Database
	// Ensure Postgres is running and database exists; fall back to SQLite if needed.
	driver := cfg.GetDriver()
	sqlitePath := cfg.GetString("database.sqlite_path", filepath.Join(".", "ecommerce.sqlite"))
	defaultDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbSSLMode)
	if driver == "postgres" || driver == "postgresql" {
		if defaultDB, err := sql.Open("postgres", defaultDSN); err == nil && defaultDB.Ping() == nil {
			defer defaultDB.Close()
			var exists bool
			if err := defaultDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists); err != nil {
				log.Printf("Warning: Failed to check database existence: %v", err)
			} else if !exists {
				if _, err := defaultDB.Exec("CREATE DATABASE " + pq.QuoteIdentifier(dbName)); err != nil {
					log.Printf("Warning: Failed to create database: %v", err)
				} else {
					log.Printf("Created database %s", dbName)
				}
			}
		} else {
			log.Printf("Postgres not reachable at %s. Falling back to SQLite at %s", defaultDSN, sqlitePath)
			driver = "sqlite3"
		}
	}

	var dsn string
	if driver == "sqlite" || driver == "sqlite3" {
		dsn = sqlitePath
	} else {
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	}
	database, err := db.NewDB(dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		log.Fatal("Make sure Postgres is running and the configured database exists.")
	}
	database.SetMaxOpenConns(cfg.GetInt("database.max_open_conns", 0))
	database.SetMaxIdleConns(cfg.GetInt("database.max_idle_conns", 0))
	database.SetConnMaxLifetime(time.Duration(cfg.GetInt("database.conn_max_lifetime", 0)) * time.Second)
	defer database.Close()

	// 2. Setup Schema
	SetupSchema(database)

	// 3. Initialize Apps
	catalog.Init(database)
	customers.Init(database)
	inventory.Init(database)
	marketing.Init(database)
	orders.Init(database)

	// 4. Initialize Admin Site
	adminSite := admin.DefaultSite
	adminSite.Title = "Forge Ecommerce Admin"

	// Set UI Prefix
	uiConfig := adminSite.GetUIConfig()
	uiConfig.Prefix = "/admin"
	adminSite.WithUIConfig(uiConfig)

	adminSite.SetDB(database)

	// Register application admins
	catalog.RegisterAdmin(ctx)
	customers.RegisterAdmin(ctx)
	inventory.RegisterAdmin(ctx)
	marketing.RegisterAdmin(ctx)
	orders.RegisterAdmin(ctx)

	// Register Plugins
	adminSite.RegisterPlugin(ctx, &ReportsPlugin{})

	// Setup Dashboard
	SetupDashboard()

	// 5. Setup API Router
	apiRouter := api.NewRouter("/api/v1")
	catalog.RegisterAPI(ctx, apiRouter, database)
	customers.RegisterAPI(ctx, apiRouter, database)
	inventory.RegisterAPI(ctx, apiRouter, database)
	marketing.RegisterAPI(ctx, apiRouter, database)
	orders.RegisterAPI(ctx, apiRouter, database)

	// 6. Create Server Router
	r := server.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.GetStringSlice("cors.allowed_origins"),
		AllowedMethods:   cfg.GetStringSlice("cors.allowed_methods"),
		AllowedHeaders:   cfg.GetStringSlice("cors.allowed_headers"),
		ExposedHeaders:   cfg.GetStringSlice("cors.exposed_headers"),
		AllowCredentials: cfg.GetBool("cors.allow_credentials", false),
		MaxAge:           cfg.GetInt("cors.max_age", 0),
	}))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Register API routes
	apiRouter.RegisterRoutes(r)

	// Admin handler (handles both REST API and UI)
	r.Mount("/admin", adminSite.Handler())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","message":"Forge Ecommerce Example is running"}`)
	})

	// Homepage
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
	<title>Forge Ecommerce</title>
	<style>
		body { font-family: Inter, system-ui, sans-serif; max-width: 900px; margin: 50px auto; padding: 20px; background: #0f172a; color: #f8fafc; }
		h1 { color: #38bdf8; border-bottom: 2px solid #334155; padding-bottom: 10px; }
		a { color: #38bdf8; text-decoration: none; font-weight: 500; }
		a:hover { color: #7dd3fc; }
		.btn { display: inline-block; background: #0ea5e9; color: white; padding: 12px 24px; border-radius: 8px; margin-right: 10px; }
	</style>
</head>
<body>
	<h1>🛒 Forge Ecommerce</h1>
	<p>Example application showcasing the Forge Framework.</p>
    <p>
        <a href="/admin/" class="btn">Admin Panel</a>
        <a href="/api/v1/categories" class="btn">Categories API</a>
        <a href="/api/v1/products" class="btn">Products API</a>
        <a href="/api/v1/orders" class="btn">Orders API</a>
    </p>
</body>
</html>
`)
	})

	serverHost := cfg.GetString("server.host", "localhost")
	serverPort := cfg.GetString("server.port", "8000")
	listenAddr := fmt.Sprintf("%s:%s", serverHost, serverPort)
	fmt.Printf("\n✨ Forge Ecommerce is Alive ✨\n")
	fmt.Printf("------------------------------\n")
	fmt.Printf("🏠 Homepage: http://%s\n", listenAddr)
	fmt.Printf("🛠️  Premium Admin: http://%s/admin/\n", listenAddr)
	fmt.Printf("------------------------------\n\n")

	log.Fatal(http.ListenAndServe(listenAddr, r))
}
