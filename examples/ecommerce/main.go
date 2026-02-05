package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/server"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq" // Import postgres driver

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
    
    // 1. Initialize Database
    // Ensure Postgres is running and database 'forge_ecommerce' exists
    // We try to create it if it doesn't exist
    defaultDSN := "postgres://postgres:123@127.0.0.1:5432/postgres?sslmode=disable"
    if defaultDB, err := sql.Open("postgres", defaultDSN); err == nil {
        defer defaultDB.Close()
        var exists bool
        defaultDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'forge_ecommerce')").Scan(&exists)
        if !exists {
            if _, err := defaultDB.Exec("CREATE DATABASE forge_ecommerce"); err != nil {
                log.Printf("Warning: Failed to create database: %v", err)
            } else {
                log.Println("Created database forge_ecommerce")
            }
        }
    }

    dsn := "postgres://postgres:123@127.0.0.1:5432/forge_ecommerce?sslmode=disable"
	database, err := db.NewDB(dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
        log.Fatal("Make sure Postgres is running and database 'forge_ecommerce' exists.")
	}
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
	
	port := ":8003"
	fmt.Printf("\n✨ Forge Ecommerce is Alive ✨\n")
	fmt.Printf("------------------------------\n")
	fmt.Printf("🏠 Homepage: http://localhost%s\n", port)
	fmt.Printf("🛠️  Premium Admin: http://localhost%s/admin/\n", port)
	fmt.Printf("------------------------------\n\n")
	
	log.Fatal(http.ListenAndServe(port, r))
}
