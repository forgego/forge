package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/api"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/server"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"examples/ecommerce/app/catalog"
	"examples/ecommerce/app/commerce"
	"examples/ecommerce/app/customers"
	"examples/ecommerce/app/engagement"
	"examples/ecommerce/app/inventory"
	"examples/ecommerce/app/marketing"
	"examples/ecommerce/app/orders"
	"examples/ecommerce/app/promotions"
	"examples/ecommerce/app/seeder"
	"examples/ecommerce/app/support"
	"examples/ecommerce/app/users"
)

//go:generate forge generate

func main() {
	ctx := context.Background()
	fmt.Println("Forge Ecommerce System")
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

	// Initialize Database. If postgres is unavailable, fall back to sqlite.
	driver := cfg.GetDriver()
	sqlitePath := cfg.GetString("database.sqlite_path", filepath.Join(".", "ecommerce.sqlite"))
	defaultDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbSSLMode)
	if driver == "postgres" || driver == "postgresql" {
		if defaultDB, err := sql.Open("postgres", defaultDSN); err == nil && defaultDB.Ping() == nil {
			defer defaultDB.Close()
			var exists bool
			if err := defaultDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists); err != nil {
				log.Printf("Warning: failed to check database existence: %v", err)
			} else if !exists {
				// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
				if _, err := defaultDB.Exec("CREATE DATABASE " + pq.QuoteIdentifier(dbName)); err != nil {
					log.Printf("Warning: failed to create database: %v", err)
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
		if strings.Contains(dsn, "?") {
			dsn += "&_journal=WAL&_busy_timeout=5000"
		} else {
			dsn += "?_journal=WAL&_busy_timeout=5000"
		}
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

	if os.Getenv("FORGE_ADMIN_USERNAME") == "" {
		if u := cfg.GetString("admin.username", "admin"); u != "" {
			_ = os.Setenv("FORGE_ADMIN_USERNAME", u)
		}
	}
	if os.Getenv("FORGE_ADMIN_PASSWORD") == "" {
		if p := cfg.GetString("admin.password", "admin123"); p != "" {
			_ = os.Setenv("FORGE_ADMIN_PASSWORD", p)
		}
	}

	if hasSeedFlag() {
		log.Println("🌱 --seed flag detected, generating sample data...")
		if err := seeder.Seed(ctx, database); err != nil {
			log.Printf("Warning: seed error: %v", err)
		}
	}

	r := buildEcommerceRouter(ctx, cfg, database)

	adminPath := normalizePath(cfg.GetString("admin.path", "/admin"), "/admin")
	serverHost := cfg.GetString("server.host", "localhost")
	serverPort := cfg.GetString("server.port", "8020")
	listenAddr := fmt.Sprintf("%s:%s", serverHost, serverPort)
	readTimeout := time.Duration(cfg.GetInt("server.read_timeout", 30)) * time.Second
	writeTimeout := time.Duration(cfg.GetInt("server.write_timeout", 30)) * time.Second
	idleTimeout := time.Duration(cfg.GetInt("server.idle_timeout", 120)) * time.Second
	maxHeaderBytes := cfg.GetInt("server.max_header_bytes", 1048576)

	fmt.Printf("\nForge Ecommerce is alive\n")
	fmt.Printf("------------------------------\n")
	fmt.Printf("Homepage: http://%s\n", listenAddr)
	fmt.Printf("Admin: http://%s%s/\n", listenAddr, adminPath)
	fmt.Printf("------------------------------\n\n")

	httpServer := &http.Server{
		Addr:           listenAddr,
		Handler:        r,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    idleTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
	log.Fatal(httpServer.ListenAndServe())
}

func buildEcommerceRouter(ctx context.Context, cfg *config.Config, database *db.DB) *server.Router {
	adminPath := normalizePath(cfg.GetString("admin.path", "/admin"), "/admin")
	apiPath := normalizePath(cfg.GetString("api.path", "/api/v1"), "/api/v1")

	SetupSchema(database)

	// Initialize all modules
	catalog.Init(database)
	commerce.Init(database)
	customers.Init(database)
	engagement.Init(database)
	inventory.Init(database)
	marketing.Init(database)
	orders.Init(database)
	promotions.Init(database)
	support.Init(database)
	users.Init(database)

	adminSite := admin.DefaultSite
	adminSite.Title = "Forge Ecommerce Admin"
	uiConfig := adminSite.GetUIConfig()
	uiConfig.Prefix = adminPath

	// Configure Admin UI source. If static directory exists, prefer it so users don't need `-tags embed`
	adminStaticDir := cfg.GetString("admin.static_dir", "")
	if adminStaticDir != "" {
		if st, err := os.Stat(filepath.Join(adminStaticDir, "index.html")); err != nil || st.IsDir() {
			adminStaticDir = ""
		}
	}
	if adminStaticDir == "" {
		candidates := []string{
			"../../forge/admin/ui/dist",
			"../forge/admin/ui/dist",
			"./dist",
			"static/admin",
		}
		for _, c := range candidates {
			if st, err := os.Stat(filepath.Join(c, "index.html")); err == nil && !st.IsDir() {
				adminStaticDir = c
				break
			}
		}
	}
	if adminStaticDir != "" {
		uiConfig.Source = admin.UISourceStatic
		uiConfig.StaticDir = adminStaticDir
	}
	adminSite.WithUIConfig(uiConfig)
	adminSite.SetDB(database)

	// Register all models with admin
	catalog.RegisterAdmin(ctx)
	commerce.RegisterAdmin(ctx)
	customers.RegisterAdmin(ctx)
	engagement.RegisterAdmin(ctx)
	inventory.RegisterAdmin(ctx)
	marketing.RegisterAdmin(ctx)
	orders.RegisterAdmin(ctx)
	promotions.RegisterAdmin(ctx)
	support.RegisterAdmin(ctx)
	users.RegisterAdmin(ctx)

	_ = adminSite.RegisterPlugin(ctx, &ReportsPlugin{})
	SetupDashboard()

	r := server.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.GetStringSlice("cors.allowed_origins", nil),
		AllowedMethods:   cfg.GetStringSlice("cors.allowed_methods", nil),
		AllowedHeaders:   cfg.GetStringSlice("cors.allowed_headers", nil),
		ExposedHeaders:   cfg.GetStringSlice("cors.exposed_headers", nil),
		AllowCredentials: cfg.GetBool("cors.allow_credentials", false),
		MaxAge:           cfg.GetInt("cors.max_age", 0),
	}))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	if cfg.GetBool("api.enabled", true) {
		api.Initialize()
		apiRouter := api.NewRouter(apiPath)
		// Register all API endpoints
		catalog.RegisterAPI(ctx, apiRouter, database)
		commerce.RegisterAPI(ctx, apiRouter, database)
		customers.RegisterAPI(ctx, apiRouter, database)
		engagement.RegisterAPI(ctx, apiRouter, database)
		inventory.RegisterAPI(ctx, apiRouter, database)
		marketing.RegisterAPI(ctx, apiRouter, database)
		orders.RegisterAPI(ctx, apiRouter, database)
		promotions.RegisterAPI(ctx, apiRouter, database)
		support.RegisterAPI(ctx, apiRouter, database)
		users.RegisterAPI(ctx, apiRouter, database)
		apiRouter.RegisterRoutes(r)
	}

	r.Mount(adminPath, adminSite.Handler())

	if adminPath != "/" {
		r.Get(adminPath, func(w http.ResponseWriter, req *http.Request) {
			http.Redirect(w, req, adminPath+"/", http.StatusPermanentRedirect)
		})
	}

	r.Get("/login", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, adminPath+"/login", http.StatusTemporaryRedirect)
	})

	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","message":"Forge Ecommerce Example is running"}`)
	})

	r.Get("/api/openapi.json", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		spec := map[string]any{
			"openapi": "3.0.3",
			"info": map[string]any{
				"title":       "Forge Ecommerce Reference API",
				"version":     "1.0.0",
				"description": "Auto-generated REST API powered by Forge ModelViewSets, QuerySet, and Gorilla CSRF.",
			},
			"servers": []map[string]any{
				{"url": "/", "description": "Current Server"},
			},
			"paths": map[string]any{
				"/health": map[string]any{
					"get": map[string]any{
						"summary":     "Health Check Probe",
						"description": "Returns system health and server heartbeat",
						"responses": map[string]any{
							"200": map[string]any{"description": "Healthy response"},
						},
					},
				},
				"/api/v1/catalog/stats": map[string]any{
					"get": map[string]any{
						"summary":     "Catalog Aggregated Statistics",
						"description": "Aggregations computed via Forge ORM (Count, Avg, Min, Max)",
						"responses": map[string]any{
							"200": map[string]any{"description": "Catalog statistics JSON"},
						},
					},
				},
				"/api/v1/catalog/search": map[string]any{
					"get": map[string]any{
						"summary":     "Faceted Product Search",
						"description": "Multi-field search filtering by keyword, price bounds, category, and stock",
						"parameters": []map[string]any{
							{"name": "q", "in": "query", "schema": map[string]string{"type": "string"}, "description": "Search keyword"},
							{"name": "category_id", "in": "query", "schema": map[string]string{"type": "integer"}, "description": "Category filter"},
							{"name": "min_price", "in": "query", "schema": map[string]string{"type": "number"}, "description": "Minimum price"},
							{"name": "max_price", "in": "query", "schema": map[string]string{"type": "number"}, "description": "Maximum price"},
							{"name": "in_stock", "in": "query", "schema": map[string]string{"type": "boolean"}, "description": "Filter in-stock items"},
						},
						"responses": map[string]any{
							"200": map[string]any{"description": "Search results list"},
						},
					},
				},
				"/api/v1/products/": map[string]any{
					"get": map[string]any{
						"summary":     "List Products",
						"description": "Paginated product listings from Forge ModelViewSet",
						"responses": map[string]any{
							"200": map[string]any{"description": "Paginated products list"},
						},
					},
				},
				"/api/v1/orders/summary": map[string]any{
					"get": map[string]any{
						"summary":     "Orders Analytics Summary",
						"description": "Order counts and revenue totals grouped by fulfillment status",
						"responses": map[string]any{
							"200": map[string]any{"description": "Orders summary JSON"},
						},
					},
				},
				"/api/v1/orders/checkout": map[string]any{
					"post": map[string]any{
						"summary":     "Checkout Order Placement",
						"description": "Places order, triggers BeforeCreate lifecycle hooks, and calculates totals",
						"responses": map[string]any{
							"201": map[string]any{"description": "Order created successfully"},
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(spec)
	})

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := strings.ReplaceAll(storefrontHTML, "{{ADMIN_PATH}}", adminPath)
		html = strings.ReplaceAll(html, "{{API_PATH}}", apiPath)
		_, _ = w.Write([]byte(html))
	})

	return r
}

func normalizePath(value string, fallback string) string {
	path := strings.TrimSpace(value)
	if path == "" {
		path = fallback
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = strings.TrimRight(path, "/")
	if path == "" {
		return "/"
	}
	return path
}

func hasSeedFlag() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--seed" || arg == "-seed" || arg == "seed" {
			return true
		}
	}
	return false
}

const storefrontHTML = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Forge Framework — Reference Ecommerce Showcase</title>
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
	<style>
		:root {
			--bg-dark: #090d16;
			--bg-card: #0f172a;
			--bg-card-hover: #17233f;
			--border: #1e293b;
			--border-highlight: #334155;
			--text-main: #f8fafc;
			--text-muted: #94a3b8;
			--primary: #38bdf8;
			--primary-glow: rgba(56, 189, 248, 0.2);
			--secondary: #818cf8;
			--accent-green: #10b981;
			--accent-amber: #f59e0b;
			--accent-rose: #f43f5e;
			--code-bg: #030712;
		}
		* { box-sizing: border-box; margin: 0; padding: 0; }
		body {
			background-color: var(--bg-dark);
			color: var(--text-main);
			font-family: 'Inter', system-ui, -apple-system, sans-serif;
			line-height: 1.6;
			padding-bottom: 80px;
		}
		a { color: inherit; text-decoration: none; }
		.container { max-width: 1240px; margin: 0 auto; padding: 0 24px; }
		
		/* Header / Navbar */
		header {
			position: sticky;
			top: 0;
			z-index: 50;
			background: rgba(9, 13, 22, 0.85);
			backdrop-filter: blur(12px);
			border-bottom: 1px solid var(--border);
			padding: 16px 0;
		}
		.nav-content {
			display: flex;
			align-items: center;
			justify-content: space-between;
		}
		.brand {
			display: flex;
			align-items: center;
			gap: 12px;
			font-weight: 800;
			font-size: 1.25rem;
			letter-spacing: -0.02em;
			color: #fff;
		}
		.brand-badge {
			background: linear-gradient(135deg, var(--primary), var(--secondary));
			padding: 4px 10px;
			border-radius: 6px;
			font-size: 0.75rem;
			font-weight: 700;
			color: #030712;
			text-transform: uppercase;
		}
		.nav-links {
			display: flex;
			align-items: center;
			gap: 20px;
		}
		.nav-link {
			color: var(--text-muted);
			font-size: 0.9rem;
			font-weight: 500;
			transition: color 0.2s;
		}
		.nav-link:hover { color: var(--primary); }
		.btn {
			display: inline-flex;
			align-items: center;
			gap: 8px;
			padding: 8px 18px;
			border-radius: 8px;
			font-size: 0.875rem;
			font-weight: 600;
			cursor: pointer;
			transition: all 0.2s;
			border: 1px solid transparent;
		}
		.btn-primary {
			background: linear-gradient(135deg, #0284c7, #2563eb);
			color: #fff;
			box-shadow: 0 4px 14px var(--primary-glow);
		}
		.btn-primary:hover {
			transform: translateY(-1px);
			box-shadow: 0 6px 20px rgba(56, 189, 248, 0.35);
		}
		.btn-outline {
			background: var(--bg-card);
			border-color: var(--border);
			color: var(--text-main);
		}
		.btn-outline:hover {
			border-color: var(--border-highlight);
			background: var(--bg-card-hover);
		}

		/* Hero Section */
		.hero {
			padding: 60px 0 40px;
			text-align: center;
		}
		.hero-badge {
			display: inline-flex;
			align-items: center;
			gap: 8px;
			background: rgba(56, 189, 248, 0.1);
			border: 1px solid rgba(56, 189, 248, 0.3);
			padding: 6px 14px;
			border-radius: 9999px;
			font-size: 0.8rem;
			font-weight: 600;
			color: var(--primary);
			margin-bottom: 20px;
		}
		.hero-badge .dot {
			width: 8px;
			height: 8px;
			border-radius: 50%;
			background: var(--accent-green);
			box-shadow: 0 0 8px var(--accent-green);
		}
		.hero h1 {
			font-size: 3rem;
			font-weight: 800;
			letter-spacing: -0.03em;
			line-height: 1.15;
			margin-bottom: 16px;
			background: linear-gradient(135deg, #ffffff 0%, #cbd5e1 50%, #7dd3fc 100%);
			-webkit-background-clip: text;
			-webkit-text-fill-color: transparent;
		}
		.hero p {
			font-size: 1.125rem;
			color: var(--text-muted);
			max-width: 720px;
			margin: 0 auto 30px;
		}
		.hero-actions {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 16px;
			margin-bottom: 40px;
		}

		/* Live Metrics Bar */
		.metrics-bar {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
			gap: 16px;
			background: var(--bg-card);
			border: 1px solid var(--border);
			border-radius: 12px;
			padding: 20px;
			margin-bottom: 50px;
		}
		.metric-item {
			text-align: center;
			padding: 10px;
		}
		.metric-val {
			font-size: 1.75rem;
			font-weight: 800;
			color: var(--primary);
			font-family: 'JetBrains Mono', monospace;
		}
		.metric-lbl {
			font-size: 0.8rem;
			color: var(--text-muted);
			text-transform: uppercase;
			letter-spacing: 0.05em;
			margin-top: 4px;
		}

		/* Framework Pillars Grid */
		.section-title {
			font-size: 1.75rem;
			font-weight: 700;
			letter-spacing: -0.02em;
			margin-bottom: 8px;
			display: flex;
			align-items: center;
			gap: 10px;
		}
		.section-desc {
			color: var(--text-muted);
			margin-bottom: 24px;
			font-size: 0.95rem;
		}
		.pillars-grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
			gap: 20px;
			margin-bottom: 60px;
		}
		.pillar-card {
			background: var(--bg-card);
			border: 1px solid var(--border);
			border-radius: 12px;
			padding: 24px;
			transition: all 0.2s;
		}
		.pillar-card:hover {
			border-color: var(--border-highlight);
			transform: translateY(-2px);
			box-shadow: 0 10px 25px rgba(0,0,0,0.3);
		}
		.pillar-icon {
			font-size: 1.5rem;
			margin-bottom: 12px;
		}
		.pillar-card h3 {
			font-size: 1.1rem;
			font-weight: 700;
			margin-bottom: 8px;
			color: #fff;
		}
		.pillar-card p {
			font-size: 0.875rem;
			color: var(--text-muted);
			line-height: 1.5;
		}

		/* Interactive Catalog */
		.catalog-section {
			margin-bottom: 60px;
		}
		.catalog-controls {
			display: flex;
			flex-wrap: wrap;
			gap: 16px;
			align-items: center;
			justify-content: space-between;
			margin-bottom: 24px;
		}
		.search-box {
			flex: 1;
			min-width: 280px;
			max-width: 460px;
			position: relative;
		}
		.search-box input {
			width: 100%;
			background: var(--bg-card);
			border: 1px solid var(--border);
			border-radius: 8px;
			padding: 10px 16px;
			color: #fff;
			font-size: 0.9rem;
			outline: none;
			transition: border-color 0.2s;
		}
		.search-box input:focus {
			border-color: var(--primary);
		}
		.category-chips {
			display: flex;
			flex-wrap: wrap;
			gap: 8px;
		}
		.chip {
			background: var(--bg-card);
			border: 1px solid var(--border);
			color: var(--text-muted);
			padding: 6px 14px;
			border-radius: 9999px;
			font-size: 0.8rem;
			font-weight: 500;
			cursor: pointer;
			transition: all 0.2s;
		}
		.chip:hover, .chip.active {
			background: rgba(56, 189, 248, 0.15);
			color: var(--primary);
			border-color: var(--primary);
		}

		/* Product Grid */
		.product-grid {
			display: grid;
			grid-template-columns: repeat(auto-fill, minmax(270px, 1fr));
			gap: 20px;
		}
		.product-card {
			background: var(--bg-card);
			border: 1px solid var(--border);
			border-radius: 12px;
			padding: 20px;
			display: flex;
			flex-direction: column;
			justify-content: space-between;
			transition: all 0.2s;
		}
		.product-card:hover {
			border-color: var(--border-highlight);
			transform: translateY(-2px);
		}
		.product-img {
			width: 100%;
			height: 150px;
			background: linear-gradient(135deg, #1e293b, #0f172a);
			border-radius: 8px;
			display: flex;
			align-items: center;
			justify-content: center;
			font-size: 2.5rem;
			margin-bottom: 16px;
			border: 1px solid rgba(255,255,255,0.05);
		}
		.product-meta {
			display: flex;
			align-items: center;
			justify-content: space-between;
			margin-bottom: 8px;
		}
		.product-sku {
			font-family: 'JetBrains Mono', monospace;
			font-size: 0.75rem;
			color: var(--text-muted);
		}
		.product-badge {
			font-size: 0.7rem;
			font-weight: 600;
			padding: 2px 8px;
			border-radius: 4px;
			background: rgba(16, 185, 129, 0.15);
			color: var(--accent-green);
		}
		.product-title {
			font-size: 1.05rem;
			font-weight: 700;
			margin-bottom: 6px;
			color: #fff;
		}
		.product-desc {
			font-size: 0.825rem;
			color: var(--text-muted);
			margin-bottom: 16px;
			display: -webkit-box;
			-webkit-line-clamp: 2;
			-webkit-box-orient: vertical;
			overflow: hidden;
		}
		.product-footer {
			display: flex;
			align-items: center;
			justify-content: space-between;
			padding-top: 12px;
			border-top: 1px solid var(--border);
		}
		.product-price {
			font-size: 1.25rem;
			font-weight: 800;
			color: #fff;
			font-family: 'JetBrains Mono', monospace;
		}
		.btn-buy {
			background: rgba(56, 189, 248, 0.15);
			color: var(--primary);
			border: 1px solid var(--primary);
			padding: 6px 12px;
			border-radius: 6px;
			font-size: 0.8rem;
			font-weight: 600;
			cursor: pointer;
			transition: all 0.2s;
		}
		.btn-buy:hover {
			background: var(--primary);
			color: #030712;
		}

		/* API Console */
		.api-console-section {
			background: var(--bg-card);
			border: 1px solid var(--border);
			border-radius: 12px;
			padding: 28px;
			margin-bottom: 60px;
		}
		.api-quick-buttons {
			display: flex;
			flex-wrap: wrap;
			gap: 10px;
			margin-bottom: 20px;
		}
		.api-btn {
			background: #1e293b;
			color: var(--text-main);
			border: 1px solid var(--border-highlight);
			padding: 8px 14px;
			border-radius: 6px;
			font-size: 0.825rem;
			font-weight: 600;
			cursor: pointer;
			transition: all 0.2s;
			font-family: 'JetBrains Mono', monospace;
		}
		.api-btn:hover {
			background: #334155;
			border-color: var(--primary);
			color: var(--primary);
		}
		.api-bar {
			display: flex;
			gap: 10px;
			margin-bottom: 16px;
		}
		.api-method {
			background: #1e293b;
			border: 1px solid var(--border-highlight);
			color: var(--accent-green);
			font-family: 'JetBrains Mono', monospace;
			font-weight: 700;
			padding: 10px 16px;
			border-radius: 6px;
			font-size: 0.85rem;
		}
		.api-input {
			flex: 1;
			background: var(--code-bg);
			border: 1px solid var(--border);
			color: #fff;
			padding: 10px 16px;
			border-radius: 6px;
			font-family: 'JetBrains Mono', monospace;
			font-size: 0.9rem;
			outline: none;
		}
		.api-input:focus {
			border-color: var(--primary);
		}
		.api-response-wrapper {
			background: var(--code-bg);
			border: 1px solid var(--border);
			border-radius: 8px;
			overflow: hidden;
		}
		.api-response-header {
			background: #0d1321;
			padding: 10px 16px;
			border-bottom: 1px solid var(--border);
			display: flex;
			align-items: center;
			justify-content: space-between;
			font-size: 0.8rem;
			font-family: 'JetBrains Mono', monospace;
		}
		.status-badge {
			padding: 2px 8px;
			border-radius: 4px;
			font-weight: 700;
		}
		.status-200 { background: rgba(16, 185, 129, 0.2); color: var(--accent-green); }
		.status-err { background: rgba(244, 63, 94, 0.2); color: var(--accent-rose); }
		.api-response-body {
			padding: 16px;
			max-height: 380px;
			overflow-y: auto;
			font-family: 'JetBrains Mono', monospace;
			font-size: 0.85rem;
			color: #38bdf8;
			white-space: pre-wrap;
			word-break: break-all;
		}

		/* Toast Notification */
		#toast {
			position: fixed;
			bottom: 30px;
			right: 30px;
			background: #1e293b;
			border: 1px solid var(--accent-green);
			color: #fff;
			padding: 14px 20px;
			border-radius: 8px;
			box-shadow: 0 10px 30px rgba(0,0,0,0.5);
			z-index: 100;
			display: none;
			font-size: 0.9rem;
			align-items: center;
			gap: 10px;
			animation: slideIn 0.3s ease-out;
		}
		@keyframes slideIn {
			from { transform: translateY(20px); opacity: 0; }
			to { transform: translateY(0); opacity: 1; }
		}

		/* Footer */
		footer {
			border-top: 1px solid var(--border);
			padding: 40px 0;
			text-align: center;
			color: var(--text-muted);
			font-size: 0.875rem;
		}
		footer a { color: var(--primary); }
	</style>
</head>
<body>
	<header>
		<div class="container nav-content">
			<div class="brand">
				<span>⚡ Forge</span>
				<span class="brand-badge">Ecommerce</span>
			</div>
			<nav class="nav-links">
				<a href="#catalog" class="nav-link">Storefront</a>
				<a href="#api-console" class="nav-link">Live API</a>
				<a href="/api/openapi.json" target="_blank" class="nav-link">OpenAPI 3.0</a>
				<a href="{{ADMIN_PATH}}/" class="btn btn-primary">Enter Admin Console</a>
			</nav>
		</div>
	</header>

	<main class="container">
		<!-- Hero -->
		<section class="hero">
			<div class="hero-badge">
				<div class="dot"></div>
				<span>Forge Reference Application & Framework Showcase</span>
			</div>
			<h1>Full-Stack Go Framework<br>Engineered for High Velocity</h1>
			<p>
				Showcasing Schema DSL with lifecycle hooks, type-safe ORM QuerySets, an embedded React 19 Admin SPA, and production-grade REST ModelViewSets with Gorilla CSRF.
			</p>
			<div class="hero-actions">
				<a href="{{ADMIN_PATH}}/" class="btn btn-primary">Open Admin Console &rarr;</a>
				<a href="#api-console" class="btn btn-outline">Explore Live API &darr;</a>
				<a href="/api/openapi.json" target="_blank" class="btn btn-outline">OpenAPI Specs</a>
			</div>
		</section>

		<!-- Live Metrics -->
		<section class="metrics-bar">
			<div class="metric-item">
				<div class="metric-val" id="metricProducts">--</div>
				<div class="metric-lbl">Total Products</div>
			</div>
			<div class="metric-item">
				<div class="metric-val" id="metricCategories">--</div>
				<div class="metric-lbl">Active Categories</div>
			</div>
			<div class="metric-item">
				<div class="metric-val" id="metricOrders">--</div>
				<div class="metric-lbl">Total Orders Placed</div>
			</div>
			<div class="metric-item">
				<div class="metric-val" id="metricRevenue">--</div>
				<div class="metric-lbl">Total Sales Revenue</div>
			</div>
			<div class="metric-item">
				<div class="metric-val" style="color: var(--accent-green);">&lt; 15ms</div>
				<div class="metric-lbl">P95 ViewSet Latency</div>
			</div>
		</section>

		<!-- Framework Pillars -->
		<div class="section-title">🏗️ Built on Forge Core Architecture</div>
		<div class="section-desc">Four unified layers providing complete full-stack web development without redundant glue code.</div>
		<section class="pillars-grid">
			<div class="pillar-card">
				<div class="pillar-icon">📐</div>
				<h3>Schema DSL & Hooks</h3>
				<p>Declarative schema definitions in Go with BeforeCreate and BeforeSave hooks, foreign keys, cascade rules, and generated columns.</p>
			</div>
			<div class="pillar-card">
				<div class="pillar-icon">⚡</div>
				<h3>Type-Safe ORM</h3>
				<p>Generic QuerySet[T] with compile-time field lookups, complex orm.Q boolean tree expressions, and aggregate computations.</p>
			</div>
			<div class="pillar-card">
				<div class="pillar-icon">🎛️</div>
				<h3>Modern Admin Console</h3>
				<p>React 19 single-page application with TanStack router, faceted filters, saved views, custom bulk actions, and KPI widgets.</p>
			</div>
			<div class="pillar-card">
				<div class="pillar-icon">🛡️</div>
				<h3>Hardened REST API</h3>
				<p>Automatic ModelViewSets with pagination, sliding-window throttling, Gorilla CSRF defense, and SHA-256 session management.</p>
			</div>
		</section>

		<!-- Live Catalog Section -->
		<section class="catalog-section" id="catalog">
			<div class="section-title">🛍️ Live Interactive Storefront</div>
			<div class="section-desc">All products queried in real-time from Forge ORM with faceted filtering and instant checkout hooks.</div>

			<div class="catalog-controls">
				<div class="search-box">
					<input type="text" id="searchInput" placeholder="Search catalog (e.g., 'Pro', 'Audio', 'Merino')...">
				</div>
				<div class="category-chips" id="categoryChips">
					<button class="chip active" onclick="filterCategory(0, this)">All Categories</button>
				</div>
			</div>

			<div class="product-grid" id="productGrid">
				<div style="grid-column: 1/-1; text-align: center; padding: 40px; color: var(--text-muted);">Loading catalog...</div>
			</div>
		</section>

		<!-- Live API Console -->
		<section class="api-console-section" id="api-console">
			<div class="section-title">🧪 Live Interactive API Console</div>
			<div class="section-desc">Execute real HTTP requests against the running Forge server and inspect responses in real-time.</div>

			<div class="api-quick-buttons">
				<button class="api-btn" onclick="runPreset('GET', '{{API_PATH}}/catalog/stats')">📊 Catalog Stats</button>
				<button class="api-btn" onclick="runPreset('GET', '{{API_PATH}}/catalog/search?q=pro')">🔍 Faceted Search (?q=pro)</button>
				<button class="api-btn" onclick="runPreset('GET', '{{API_PATH}}/orders/summary')">📦 Orders Summary</button>
				<button class="api-btn" onclick="runPreset('GET', '{{API_PATH}}/products/')">💻 Products ViewSet</button>
				<button class="api-btn" onclick="runPreset('GET', '{{API_PATH}}/categories/')">📁 Categories ViewSet</button>
				<button class="api-btn" onclick="runPreset('GET', '/health')">🩺 Health Check</button>
				<button class="api-btn" onclick="runPreset('GET', '/api/openapi.json')">📜 OpenAPI 3.0 Spec</button>
			</div>

			<div class="api-bar">
				<span class="api-method" id="apiMethod">GET</span>
				<input type="text" class="api-input" id="apiEndpoint" value="{{API_PATH}}/catalog/stats">
				<button class="btn btn-primary" onclick="executeApiCall()">Execute Query</button>
			</div>

			<div class="api-response-wrapper">
				<div class="api-response-header">
					<span>Response Payload</span>
					<div>
						<span id="apiStatus" class="status-badge status-200">200 OK</span>
						<span id="apiLatency" style="margin-left: 12px; color: var(--text-muted);">12ms</span>
					</div>
				</div>
				<pre class="api-response-body" id="apiOutput">Click 'Execute Query' or any quick preset button above...</pre>
			</div>
		</section>
	</main>

	<div id="toast">
		<span>✅</span>
		<span id="toastMsg">Order placed successfully!</span>
	</div>

	<footer>
		<div class="container">
			<p>Forge Framework — Enterprise Go Full-Stack Framework. Licensed under Apache 2.0.</p>
			<p style="margin-top: 8px;">Explore the full documentation on our <a href="{{ADMIN_PATH}}/">Admin Console</a> or read the guides on the docs site.</p>
		</div>
	</footer>

	<script>
		const API_PATH = '{{API_PATH}}';
		let currentCategoryId = 0;
		let searchDebounceTimer = null;

		async function loadStats() {
			try {
				const resCatalog = await fetch(API_PATH + '/catalog/stats');
				if (resCatalog.ok) {
					const data = await resCatalog.json();
					document.getElementById('metricProducts').innerText = data.total_products || 0;
					document.getElementById('metricCategories').innerText = data.total_categories || 0;
				}
				const resOrders = await fetch(API_PATH + '/orders/summary');
				if (resOrders.ok) {
					const oData = await resOrders.json();
					document.getElementById('metricOrders').innerText = oData.total_orders || 0;
					document.getElementById('metricRevenue').innerText = '$' + Number(oData.total_revenue || 0).toLocaleString('en-US', {minimumFractionDigits: 2, maximumFractionDigits: 2});
				}
			} catch (e) {
				console.error('Failed to load stats', e);
			}
		}

		async function loadCategories() {
			try {
				const res = await fetch(API_PATH + '/categories/');
				if (!res.ok) return;
				const data = await res.json();
				const cats = data.results || [];
				const container = document.getElementById('categoryChips');
				cats.slice(0, 7).forEach(cat => {
					const btn = document.createElement('button');
					btn.className = 'chip';
					btn.innerText = cat.name;
					btn.onclick = () => filterCategory(cat.id, btn);
					container.appendChild(btn);
				});
			} catch (e) {
				console.error('Failed to load categories', e);
			}
		}

		function filterCategory(catId, btn) {
			currentCategoryId = catId;
			document.querySelectorAll('.chip').forEach(c => c.classList.remove('active'));
			btn.classList.add('active');
			fetchProducts();
		}

		async function fetchProducts() {
			const query = document.getElementById('searchInput').value.trim();
			let url = API_PATH + '/catalog/search?';
			const params = new URLSearchParams();
			if (query) params.append('q', query);
			if (currentCategoryId > 0) params.append('category_id', currentCategoryId);
			url += params.toString();

			try {
				const res = await fetch(url);
				if (!res.ok) return;
				const data = await res.json();
				renderProducts(data.results || []);
			} catch (e) {
				console.error('Failed to fetch products', e);
			}
		}

		const iconMap = {
			1: '💻', 2: '🎧', 3: '📱', 4: '🧥', 5: '👗', 6: '👟', 7: '🏠', 8: '☕'
		};

		function renderProducts(products) {
			var container = document.getElementById('productGrid');
			if (!products || products.length === 0) {
				container.innerHTML = '<div style="grid-column: 1/-1; text-align: center; padding: 40px; color: var(--text-muted);">No products found matching your search.</div>';
				return;
			}
			var html = '';
			for (var i = 0; i < products.length; i++) {
				var p = products[i];
				var icon = iconMap[p.category_id] || '📦';
				var stockText = p.stock_quantity > 0 ? 'In Stock (' + p.stock_quantity + ')' : 'Out of Stock';
				var desc = p.description || 'Premium engineered product.';
				var priceStr = Number(p.price).toFixed(2);
				html += '<div class="product-card">' +
					'<div>' +
						'<div class="product-img">' + icon + '</div>' +
						'<div class="product-meta">' +
							'<span class="product-sku">' + p.sku + '</span>' +
							'<span class="product-badge">' + stockText + '</span>' +
						'</div>' +
						'<div class="product-title">' + p.name + '</div>' +
						'<div class="product-desc">' + desc + '</div>' +
					'</div>' +
					'<div class="product-footer">' +
						'<div class="product-price">$' + priceStr + '</div>' +
						'<button class="btn-buy" onclick="quickCheckout(' + p.id + ', \'' + escapeHtml(p.name) + '\', \'' + p.sku + '\', ' + p.price + ')">⚡ Instant Buy</button>' +
					'</div>' +
				'</div>';
			}
			container.innerHTML = html;
		}

		function escapeHtml(str) {
			return str.replace(/'/g, "\\'");
		}

		async function quickCheckout(prodId, name, sku, price) {
			try {
				const res = await fetch(API_PATH + '/orders/checkout', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({
						customer_email: 'customer@forgego.dev',
						customer_first_name: 'Alex',
						customer_last_name: 'Dev',
						customer_phone: '+1-555-0199',
						shipping_city: 'San Francisco',
						shipping_state: 'CA',
						items: [
							{ product_id: prodId, product_name: name, sku: sku, quantity: 1, unit_price: price }
						]
					})
				});
				if (res.ok) {
					const data = await res.json();
					showToast('Order ' + data.order_number + ' placed! Total: $' + Number(data.total).toFixed(2));
					loadStats();
					fetchProducts();
				} else {
					const err = await res.text();
					showToast('Checkout notice: ' + err);
				}
			} catch (e) {
				showToast('Checkout failed: ' + e.message);
			}
		}

		function showToast(msg) {
			const toast = document.getElementById('toast');
			document.getElementById('toastMsg').innerText = msg;
			toast.style.display = 'flex';
			setTimeout(() => { toast.style.display = 'none'; }, 4000);
		}

		function runPreset(method, endpoint) {
			document.getElementById('apiMethod').innerText = method;
			document.getElementById('apiEndpoint').value = endpoint;
			executeApiCall();
		}

		async function executeApiCall() {
			const endpoint = document.getElementById('apiEndpoint').value;
			const method = document.getElementById('apiMethod').innerText;
			const output = document.getElementById('apiOutput');
			const statusBadge = document.getElementById('apiStatus');
			const latencyBadge = document.getElementById('apiLatency');

			output.innerText = 'Executing query...';
			const start = performance.now();
			try {
				const res = await fetch(endpoint, { method: method });
				const duration = Math.round(performance.now() - start);
				latencyBadge.innerText = duration + 'ms';
				statusBadge.innerText = res.status + ' ' + res.statusText;
				if (res.ok) {
					statusBadge.className = 'status-badge status-200';
				} else {
					statusBadge.className = 'status-badge status-err';
				}
				const contentType = res.headers.get('content-type') || '';
				if (contentType.includes('json')) {
					const json = await res.json();
					output.innerText = JSON.stringify(json, null, 2);
				} else {
					const text = await res.text();
					output.innerText = text;
				}
			} catch (err) {
				statusBadge.innerText = 'ERR';
				statusBadge.className = 'status-badge status-err';
				output.innerText = 'Network error: ' + err.message;
			}
		}

		document.getElementById('searchInput').addEventListener('input', () => {
			clearTimeout(searchDebounceTimer);
			searchDebounceTimer = setTimeout(fetchProducts, 250);
		});

		window.addEventListener('DOMContentLoaded', () => {
			loadStats();
			loadCategories();
			fetchProducts();
			executeApiCall();
		});
	</script>
</body>
</html>
`
