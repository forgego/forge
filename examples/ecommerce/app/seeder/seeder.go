package seeder

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/db"
)

// rebind converts '?' placeholders to '$1, $2, ...' for PostgreSQL compatibility.
func rebind(driver string, query string) string {
	if driver == "postgres" || driver == "postgresql" {
		var out strings.Builder
		paramIndex := 1
		for i := 0; i < len(query); i++ {
			if query[i] == '?' {
				out.WriteString(fmt.Sprintf("$%d", paramIndex))
				paramIndex++
			} else {
				out.WriteByte(query[i])
			}
		}
		return out.String()
	}
	return query
}

func exec(ctx context.Context, database *db.DB, query string, args ...interface{}) (sql.Result, error) {
	q := rebind(database.Driver, query)
	return database.ExecContext(ctx, q, args...)
}

func queryRow(ctx context.Context, database *db.DB, query string, args ...interface{}) *sql.Row {
	q := rebind(database.Driver, query)
	return database.QueryRowContext(ctx, q, args...)
}

func idExists(ctx context.Context, database *db.DB, table string, col string, val interface{}) bool {
	var id int64
	err := queryRow(ctx, database, fmt.Sprintf("SELECT id FROM %s WHERE %s = ? LIMIT 1", table, col), val).Scan(&id)
	return err == nil && id > 0
}

func getId(ctx context.Context, database *db.DB, table string, col string, val interface{}) int64 {
	var id int64
	_ = queryRow(ctx, database, fmt.Sprintf("SELECT id FROM %s WHERE %s = ? LIMIT 1", table, col), val).Scan(&id)
	return id
}

// Seed executes full synthetic data generation across all ecommerce domains.
func Seed(ctx context.Context, database *db.DB) error {
	log.Println("🌱 Starting comprehensive ecommerce data generation...")

	catMap, err := seedCategories(ctx, database)
	if err != nil {
		return fmt.Errorf("seedCategories: %w", err)
	}

	brandMap, err := seedBrands(ctx, database)
	if err != nil {
		return fmt.Errorf("seedBrands: %w", err)
	}

	prodMap, err := seedProducts(ctx, database, catMap, brandMap)
	if err != nil {
		return fmt.Errorf("seedProducts: %w", err)
	}

	variantMap, err := seedProductVariants(ctx, database, prodMap)
	if err != nil {
		return fmt.Errorf("seedProductVariants: %w", err)
	}

	groupMap, err := seedCustomerGroups(ctx, database)
	if err != nil {
		return fmt.Errorf("seedCustomerGroups: %w", err)
	}

	custMap, addrMap, err := seedCustomersAndAddresses(ctx, database, groupMap)
	if err != nil {
		return fmt.Errorf("seedCustomersAndAddresses: %w", err)
	}

	if err := seedWarehousesAndStock(ctx, database, variantMap); err != nil {
		return fmt.Errorf("seedWarehousesAndStock: %w", err)
	}

	if err := seedCouponsAndPromotions(ctx, database); err != nil {
		return fmt.Errorf("seedCouponsAndPromotions: %w", err)
	}

	if err := seedOrdersAndPayments(ctx, database, custMap, addrMap, prodMap); err != nil {
		return fmt.Errorf("seedOrdersAndPayments: %w", err)
	}

	if err := seedReviews(ctx, database, custMap, prodMap); err != nil {
		return fmt.Errorf("seedReviews: %w", err)
	}

	if err := seedSupportAndFAQs(ctx, database, custMap); err != nil {
		return fmt.Errorf("seedSupportAndFAQs: %w", err)
	}

	if err := seedUsersAndGroups(ctx, database); err != nil {
		return fmt.Errorf("seedUsersAndGroups: %w", err)
	}

	seedAuditHistory(ctx)

	log.Println("✅ All ecommerce models successfully seeded with rich fake data!")
	return nil
}

func seedCategories(ctx context.Context, database *db.DB) (map[string]int64, error) {
	parents := []struct {
		name, slug, desc string
		sortOrder        int
	}{
		{"Electronics", "electronics", "High-performance consumer electronics, computing, and peripherals", 1},
		{"Fashion & Apparel", "fashion-apparel", "Curated designer apparel, premium footwear, and accessories", 2},
		{"Home & Living", "home-living", "Smart home devices, architectural lighting, and modern cookware", 3},
		{"Sports & Outdoors", "sports-outdoors", "Aerospace-grade training equipment, activewear, and bikes", 4},
		{"Books & Tech", "books-tech", "Definitive software engineering blueprints and technical architecture", 5},
	}

	catMap := make(map[string]int64)

	for _, p := range parents {
		if !idExists(ctx, database, "categories", "slug", p.slug) {
			_, err := exec(ctx, database, `
				INSERT INTO categories (name, slug, description, sort_order, is_active, level)
				VALUES (?, ?, ?, ?, true, 0)
			`, p.name, p.slug, p.desc, p.sortOrder)
			if err != nil {
				return nil, err
			}
		}
		catMap[p.slug] = getId(ctx, database, "categories", "slug", p.slug)
	}

	children := []struct {
		name, slug, desc, parentSlug string
		sortOrder                    int
	}{
		{"Laptops & Workstations", "laptops-workstations", "Developer and pro workstations", "electronics", 1},
		{"Audio & Headphones", "audio-headphones", "Studio-grade ANC monitors and wireless gear", "electronics", 2},
		{"Smartphones & Tablets", "smartphones-tablets", "Ultra-thin portable computation", "electronics", 3},
		{"Men's Apparel", "mens-apparel", "Technical fabrics, luxury merino wool, and tailoring", "fashion-apparel", 1},
		{"Women's Apparel", "womens-apparel", "Minimalist cuts and high-grade organic silks", "fashion-apparel", 2},
		{"Performance Footwear", "performance-footwear", "Carbon-plated road runners and city sneakers", "fashion-apparel", 3},
		{"Smart Home", "smart-home", "Matter-enabled sensors and ambient luminaires", "home-living", 1},
		{"Kitchen & Dining", "kitchen-dining", "Dual-boiler espresso machines and enameled cast iron", "home-living", 2},
	}

	for _, c := range children {
		if !idExists(ctx, database, "categories", "slug", c.slug) {
			pID := catMap[c.parentSlug]
			_, err := exec(ctx, database, `
				INSERT INTO categories (name, slug, description, parent_id, sort_order, is_active, level)
				VALUES (?, ?, ?, ?, ?, true, 1)
			`, c.name, c.slug, c.desc, pID, c.sortOrder)
			if err != nil {
				return nil, err
			}
		}
		catMap[c.slug] = getId(ctx, database, "categories", "slug", c.slug)
	}

	log.Printf("  • Seeded %d categories", len(catMap))
	return catMap, nil
}

func seedBrands(ctx context.Context, database *db.DB) (map[string]int64, error) {
	brands := []struct {
		name, slug, desc, website string
	}{
		{"Acme Dynamics", "acme-dynamics", "Universal precision engineering and culinary hardware", "https://acme.example.com"},
		{"Stark Innovations", "stark-innovations", "Advanced silicon, computing platforms, and wearable power", "https://stark.example.com"},
		{"Apex Audio", "apex-audio", "Reference-grade studio monitoring and active noise cancellation", "https://apex.example.com"},
		{"Lumina Lifestyle", "lumina-lifestyle", "Architectural apparel crafted from regenerative fibers", "https://lumina.example.com"},
		{"Titan Athletics", "titan-athletics", "Aerospace carbon frames and endurance athletics equipment", "https://titan.example.com"},
		{"Chrono Goods", "chrono-goods", "Heritage leatherwork and horological instruments", "https://chrono.example.com"},
		{"Nordic Living", "nordic-living", "Ergonomic Scandinavian minimalism for home and studio", "https://nordic.example.com"},
		{"Nexus Core", "nexus-core", "Autonomous edge devices and smart home coordination", "https://nexus.example.com"},
	}

	brandMap := make(map[string]int64)
	for _, b := range brands {
		if !idExists(ctx, database, "brands", "slug", b.slug) {
			_, err := exec(ctx, database, `
				INSERT INTO brands (name, slug, description, website_url, is_active)
				VALUES (?, ?, ?, ?, true)
			`, b.name, b.slug, b.desc, b.website)
			if err != nil {
				return nil, err
			}
		}
		brandMap[b.slug] = getId(ctx, database, "brands", "slug", b.slug)
	}

	log.Printf("  • Seeded %d brands", len(brandMap))
	return brandMap, nil
}

func seedProducts(ctx context.Context, database *db.DB, catMap, brandMap map[string]int64) (map[string]int64, error) {
	type prodSpec struct {
		name, slug, sku, desc, shortDesc, catSlug, brandSlug string
		price, cost, compareAt                               float64
		stock                                                int
		weight                                               float64
		featured                                             bool
		ratingAvg                                            float64
		ratingCount                                          int
	}

	products := []prodSpec{
		{
			name: "TitanBook Pro 16", slug: "titanbook-pro-16", sku: "LAP-TITAN-16",
			desc: "Engineered for compile speeds and complex local simulations. Features a 16-inch 120Hz Liquid Retina display, 12-core CPU, and 48-core neural accelerator.",
			shortDesc: "The ultimate workstation for engineers and creators.",
			catSlug: "laptops-workstations", brandSlug: "stark-innovations",
			price: 1999.00, cost: 1350.00, compareAt: 2199.00, stock: 50, weight: 2.1,
			featured: true, ratingAvg: 4.9, ratingCount: 42,
		},
		{
			name: "UltraPad Air 11", slug: "ultrapad-air-11", sku: "TAB-ULTRA-11",
			desc: "Featherweight 480g tablet with tandem OLED screen, M-series chip, and pencil support for architecture and sketching.",
			shortDesc: "Powerful performance packed into an impossibly thin frame.",
			catSlug: "smartphones-tablets", brandSlug: "stark-innovations",
			price: 799.00, cost: 520.00, compareAt: 849.00, stock: 75, weight: 0.48,
			featured: true, ratingAvg: 4.8, ratingCount: 28,
		},
		{
			name: "Apex ANC 900 Studio Wireless", slug: "apex-anc-900", sku: "AUD-APEX-900",
			desc: "Custom 45mm beryllium drivers delivering acoustic isolation up to -42dB. 60-hour continuous playback with lossless spatial audio support.",
			shortDesc: "Studio-grade ANC headphones with immersive spatial clarity.",
			catSlug: "audio-headphones", brandSlug: "apex-audio",
			price: 349.00, cost: 200.00, compareAt: 399.00, stock: 120, weight: 0.28,
			featured: true, ratingAvg: 4.9, ratingCount: 86,
		},
		{
			name: "SoundPulse Mini Bluetooth 5.4", slug: "soundpulse-mini", sku: "AUD-PULSE-MN",
			desc: "IP67 dust and waterproof outdoor speaker with dual passive radiators and 360-degree sound projection.",
			shortDesc: "Pocket-sized acoustic monster with all-day battery.",
			catSlug: "audio-headphones", brandSlug: "apex-audio",
			price: 89.00, cost: 42.00, compareAt: 99.00, stock: 200, weight: 0.35,
			featured: false, ratingAvg: 4.6, ratingCount: 19,
		},
		{
			name: "Lumina Merino Wool Crewneck", slug: "merino-wool-crewneck", sku: "CLO-MERINO-CRW",
			desc: "100% 17.5-micron superfine New Zealand merino wool. Odor-resistant, thermal-regulating, and tailored for effortless layering.",
			shortDesc: "All-season temperature regulating luxury knitwear.",
			catSlug: "mens-apparel", brandSlug: "lumina-lifestyle",
			price: 129.00, cost: 58.00, compareAt: 149.00, stock: 90, weight: 0.32,
			featured: true, ratingAvg: 4.7, ratingCount: 35,
		},
		{
			name: "AeroDry Athletic Performance Tee", slug: "aerodry-athletic-tee", sku: "CLO-AERODRY-TEE",
			desc: "Micro-perforated polyester blend with silver ion anti-microbial treatment. Rapid moisture dispersion for high-output cardio.",
			shortDesc: "Ultra-breathable training top for intense workouts.",
			catSlug: "mens-apparel", brandSlug: "titan-athletics",
			price: 45.00, cost: 16.00, compareAt: 55.00, stock: 160, weight: 0.14,
			featured: false, ratingAvg: 4.5, ratingCount: 17,
		},
		{
			name: "Nordic Silk Trench Coat", slug: "nordic-silk-trench", sku: "CLO-NORDIC-TRN",
			desc: "Water-repellent raw silk twill with storm flap, horn buttons, and belted waistline. Hand-finished in Milan.",
			shortDesc: "Timeless outerwear silhouette in water-resistant raw silk.",
			catSlug: "womens-apparel", brandSlug: "lumina-lifestyle",
			price: 289.00, cost: 140.00, compareAt: 349.00, stock: 35, weight: 0.85,
			featured: true, ratingAvg: 4.9, ratingCount: 24,
		},
		{
			name: "CloudRunner Pro Carbon Sneakers", slug: "cloudrunner-pro-sneakers", sku: "FTW-CLOUDRUN-PRO",
			desc: "Full-length curved carbon propulsion plate encapsulated in supercritical nitrogen-infused foam. 195 grams of pure running efficiency.",
			shortDesc: "Marathon race-day shoe with maximal energy return.",
			catSlug: "performance-footwear", brandSlug: "titan-athletics",
			price: 159.00, cost: 75.00, compareAt: 189.00, stock: 110, weight: 0.42,
			featured: true, ratingAvg: 4.8, ratingCount: 57,
		},
		{
			name: "SmartGlow Ambient Floor Lamp", slug: "smartglow-floor-lamp", sku: "HOM-SMARTGLOW-LP",
			desc: "Anodized aluminum stem with full-spectrum circadian lighting, Thread/Matter smart home integration, and capacitive touch dimmer.",
			shortDesc: "Matter-compatible minimalist lighting for productive studios.",
			catSlug: "smart-home", brandSlug: "nexus-core",
			price: 119.00, cost: 55.00, compareAt: 139.00, stock: 60, weight: 3.2,
			featured: false, ratingAvg: 4.5, ratingCount: 21,
		},
		{
			name: "BrewMaster Touch Dual-Boiler Espresso Machine", slug: "brewmaster-espresso", sku: "KIT-BREWMASTER-ESP",
			desc: "Commercial 58mm rotary pump with independent PID temperature control for brew and steam boilers. Integrated 0.1g digital scale.",
			shortDesc: "True cafe-grade espresso extraction on your home kitchen counter.",
			catSlug: "kitchen-dining", brandSlug: "acme-dynamics",
			price: 499.00, cost: 280.00, compareAt: 599.00, stock: 40, weight: 14.5,
			featured: true, ratingAvg: 4.9, ratingCount: 68,
		},
		{
			name: "ChefPro Enameled Cast Iron Dutch Oven", slug: "chefpro-dutch-oven", sku: "KIT-CHEFPRO-DUTCH",
			desc: "6.5-quart capacity with matte black enamel interior for deep browning and moisture-retaining condensation lid nodes.",
			shortDesc: "Heirloom-grade heat retention for sourdough and braises.",
			catSlug: "kitchen-dining", brandSlug: "acme-dynamics",
			price: 110.00, cost: 50.00, compareAt: 135.00, stock: 70, weight: 5.8,
			featured: false, ratingAvg: 4.8, ratingCount: 33,
		},
		{
			name: "Velocity Aero Carbon Road Bicycle", slug: "velocity-carbon-roadbike", sku: "SPT-VELOCITY-RD",
			desc: "Toray T1000 carbon monocoque frame, integrated cockpit, wireless electronic shifting, and tubeless 45mm deep-section aero wheels.",
			shortDesc: "Wind-tunnel honed racing machine with electronic shifting.",
			catSlug: "sports-outdoors", brandSlug: "titan-athletics",
			price: 1450.00, cost: 850.00, compareAt: 1699.00, stock: 15, weight: 7.4,
			featured: true, ratingAvg: 5.0, ratingCount: 14,
		},
		{
			name: "FlexGrip 50lb Quick-Select Dumbbells", slug: "flexgrip-adjustable-dumbbells", sku: "SPT-FLEXGRIP-DB",
			desc: "Replaces 16 pairs of weights. Twist-lock selector dial shifts from 5 to 50 lbs in 2.5 lb increments within 1 second.",
			shortDesc: "Compact home gym solution with precision mechanical dials.",
			catSlug: "sports-outdoors", brandSlug: "titan-athletics",
			price: 299.00, cost: 160.00, compareAt: 349.00, stock: 55, weight: 45.3,
			featured: false, ratingAvg: 4.7, ratingCount: 29,
		},
		{
			name: "The Go Architecture & Concurrency Blueprint", slug: "go-blueprint-book", sku: "BOK-GO-BLUEPRINT",
			desc: "550 pages of production-proven patterns: channels, memory models, distributed consensus, microservices, and zero-allocation performance.",
			shortDesc: "Master advanced Go systems design and concurrent programming.",
			catSlug: "books-tech", brandSlug: "acme-dynamics",
			price: 49.99, cost: 18.00, compareAt: 59.99, stock: 150, weight: 0.75,
			featured: true, ratingAvg: 5.0, ratingCount: 95,
		},
		{
			name: "Designing Resilient Cloud-Native Distributed Systems", slug: "cloudnative-systems-book", sku: "BOK-CLOUDNATIVE",
			desc: "Comprehensive guide to saga orchestration, event sourcing, circuit breakers, and zero-downtime rolling migrations in mission-critical clouds.",
			shortDesc: "Pragmatic guide to distributed architectures and resilience.",
			catSlug: "books-tech", brandSlug: "acme-dynamics",
			price: 59.99, cost: 22.00, compareAt: 69.99, stock: 110, weight: 0.85,
			featured: false, ratingAvg: 4.8, ratingCount: 47,
		},
		{
			name: "Artisan Full-Grain Veg-Tan Leather Wallet", slug: "leather-bifold-wallet", sku: "ACC-LEATHER-WLT",
			desc: "Saddle-stitched by hand with waxed Japanese poly thread. Ages to a rich patina over decades of everyday carry.",
			shortDesc: "Handmade vegetable-tanned leather bifold with RFID shielding.",
			catSlug: "mens-apparel", brandSlug: "chrono-goods",
			price: 65.00, cost: 24.00, compareAt: 80.00, stock: 130, weight: 0.09,
			featured: false, ratingAvg: 4.9, ratingCount: 38,
		},
	}

	prodMap := make(map[string]int64)
	for _, p := range products {
		if !idExists(ctx, database, "products", "sku", p.sku) {
			catID := catMap[p.catSlug]
			brandID := brandMap[p.brandSlug]
			_, err := exec(ctx, database, `
				INSERT INTO products (
					name, slug, sku, description, short_description, category_id, brand_id,
					price, cost_price, compare_at_price, stock_quantity, track_inventory,
					weight, is_active, is_featured, rating_average, rating_count,
					published_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, true, ?, true, ?, ?, ?, ?)
			`, p.name, p.slug, p.sku, p.desc, p.shortDesc, catID, brandID,
				p.price, p.cost, p.compareAt, p.stock, p.weight, p.featured,
				p.ratingAvg, p.ratingCount, time.Now().Add(-720*time.Hour))
			if err != nil {
				return nil, err
			}
		}
		prodMap[p.slug] = getId(ctx, database, "products", "sku", p.sku)
	}

	log.Printf("  • Seeded %d products", len(prodMap))
	return prodMap, nil
}

func seedProductVariants(ctx context.Context, database *db.DB, prodMap map[string]int64) (map[string]int64, error) {
	type varSpec struct {
		prodSlug, sku, name string
		price, cost         float64
		stock               int
		isDefault           bool
	}

	variants := []varSpec{
		// Default / base variants for all products
		{"titanbook-pro-16", "LAP-TITAN-16-512", "512GB SSD / 32GB RAM / Space Gray", 1999.00, 1350.00, 20, true},
		{"titanbook-pro-16", "LAP-TITAN-16-1TB", "1TB SSD / 64GB RAM / Silver", 2299.00, 1550.00, 18, false},
		{"titanbook-pro-16", "LAP-TITAN-16-2TB", "2TB SSD / 96GB RAM / Dark Titanium", 2699.00, 1800.00, 12, false},

		{"ultrapad-air-11", "TAB-ULTRA-11-256", "256GB / Wi-Fi / Space Gray", 799.00, 520.00, 45, true},
		{"ultrapad-air-11", "TAB-ULTRA-11-512", "512GB / Cellular / Starlight", 999.00, 650.00, 30, false},

		{"apex-anc-900", "AUD-APEX-900-BLK", "Midnight Black", 349.00, 200.00, 50, true},
		{"apex-anc-900", "AUD-APEX-900-SLV", "Matte Silver", 349.00, 200.00, 45, false},
		{"apex-anc-900", "AUD-APEX-900-SND", "Desert Sand", 349.00, 200.00, 25, false},

		{"soundpulse-mini", "AUD-PULSE-MN-STD", "Standard Carbon Black", 89.00, 42.00, 200, true},

		{"merino-wool-crewneck", "CLO-MERINO-M-CHR", "Medium / Charcoal Heather", 129.00, 58.00, 30, true},
		{"merino-wool-crewneck", "CLO-MERINO-L-NVY", "Large / Deep Navy", 129.00, 58.00, 35, false},
		{"merino-wool-crewneck", "CLO-MERINO-XL-FOR", "X-Large / Forest Pine", 129.00, 58.00, 25, false},

		{"aerodry-athletic-tee", "CLO-AERODRY-M-BLK", "Medium / Stealth Black", 45.00, 16.00, 80, true},
		{"aerodry-athletic-tee", "CLO-AERODRY-L-BLK", "Large / Stealth Black", 45.00, 16.00, 80, false},

		{"nordic-silk-trench", "CLO-NORDIC-S-BEI", "Small / Oat Beige", 289.00, 140.00, 15, true},
		{"nordic-silk-trench", "CLO-NORDIC-M-BEI", "Medium / Oat Beige", 289.00, 140.00, 20, false},

		{"cloudrunner-pro-sneakers", "FTW-CLOUDRUN-M9", "US 9 / Slate Gray", 159.00, 75.00, 35, true},
		{"cloudrunner-pro-sneakers", "FTW-CLOUDRUN-M10", "US 10 / Obsidian Black", 159.00, 75.00, 45, false},
		{"cloudrunner-pro-sneakers", "FTW-CLOUDRUN-M11", "US 11 / White Pulse", 159.00, 75.00, 30, false},

		{"smartglow-floor-lamp", "HOM-SMARTGLOW-STD", "Matte Aluminum Standard", 119.00, 55.00, 60, true},
		{"brewmaster-espresso", "KIT-BREWMASTER-STD", "Brushed Stainless Steel", 499.00, 280.00, 40, true},
		{"chefpro-dutch-oven", "KIT-CHEFPRO-STD", "Enamelled Matte Black", 110.00, 50.00, 70, true},
		{"velocity-carbon-roadbike", "SPT-VELOCITY-M54", "Size 54cm Medium / Gloss Carbon", 1450.00, 850.00, 15, true},
		{"flexgrip-adjustable-dumbbells", "SPT-FLEXGRIP-STD", "50lb Pair Standard", 299.00, 160.00, 55, true},
		{"go-blueprint-book", "BOK-GO-PRINT", "Hardcover Print Edition", 49.99, 18.00, 150, true},
		{"cloudnative-systems-book", "BOK-CLOUD-PRINT", "Hardcover Print Edition", 59.99, 22.00, 110, true},
		{"leather-bifold-wallet", "ACC-LEATHER-BRN", "Whiskey Brown Veg-Tan", 65.00, 24.00, 130, true},
	}

	variantMap := make(map[string]int64)
	count := 0
	for _, v := range variants {
		if !idExists(ctx, database, "product_variants", "sku", v.sku) {
			prodID := prodMap[v.prodSlug]
			if prodID > 0 {
				_, err := exec(ctx, database, `
					INSERT INTO product_variants (
						product_id, sku, name, price, cost_price, stock_quantity,
						is_active, is_default
					) VALUES (?, ?, ?, ?, ?, ?, true, ?)
				`, prodID, v.sku, v.name, v.price, v.cost, v.stock, v.isDefault)
				if err != nil {
					return nil, err
				}
				count++
			}
		}
		variantMap[v.sku] = getId(ctx, database, "product_variants", "sku", v.sku)
	}

	log.Printf("  • Seeded %d product variants", len(variantMap))
	return variantMap, nil
}

func seedCustomerGroups(ctx context.Context, database *db.DB) (map[string]int64, error) {
	groups := []struct {
		name, code, desc string
		discount         float64
	}{
		{"VIP Diamond Circle", "VIP", "Top-tier loyal customers with premium concierge service", 15.0},
		{"Executive Gold Tier", "GOLD", "Frequent shoppers with priority shipping", 10.0},
		{"Standard Retail", "RETAIL", "General retail registered shoppers", 0.0},
		{"Wholesale Commercial", "WHOLESALE", "Enterprise bulk purchase and institutional buyers", 20.0},
	}

	groupMap := make(map[string]int64)
	for _, g := range groups {
		if !idExists(ctx, database, "customer_groups", "code", g.code) {
			_, err := exec(ctx, database, `
				INSERT INTO customer_groups (name, code, description, discount_percentage, is_active)
				VALUES (?, ?, ?, ?, true)
			`, g.name, g.code, g.desc, g.discount)
			if err != nil {
				return nil, err
			}
		}
		groupMap[g.code] = getId(ctx, database, "customer_groups", "code", g.code)
	}

	log.Printf("  • Seeded %d customer groups", len(groupMap))
	return groupMap, nil
}

func seedCustomersAndAddresses(ctx context.Context, database *db.DB, groupMap map[string]int64) (map[string]int64, map[string]int64, error) {
	type custSpec struct {
		first, last, email, phone, groupCode string
		spent                                float64
		orders                               int
		city, state, zip, countryCode, cName string
		street                               string
	}

	customers := []custSpec{
		{"Alice", "Walker", "alice.walker@example.com", "+1-212-555-0143", "VIP", 3850.00, 7, "New York", "NY", "10001", "US", "United States", "350 5th Avenue, Suite 4400"},
		{"Marcus", "Vance", "marcus.vance@example.com", "+1-310-555-0182", "GOLD", 1420.00, 4, "Los Angeles", "CA", "90012", "US", "United States", "1200 Wilshire Blvd"},
		{"Sophia", "Chen", "sophia.chen@example.com", "+1-206-555-0199", "RETAIL", 480.00, 2, "Seattle", "WA", "98101", "US", "United States", "705 Pike Street"},
		{"David", "Miller", "david.miller@example.com", "+1-312-555-0174", "RETAIL", 240.00, 1, "Chicago", "IL", "60601", "US", "United States", "233 S Wacker Drive"},
		{"Elena", "Rostova", "elena.rostova@example.com", "+1-512-555-0115", "VIP", 4650.00, 9, "Austin", "TX", "78701", "US", "United States", "500 W 2nd St"},
		{"Jordan", "Taylor", "jordan.taylor@example.com", "+1-415-555-0168", "WHOLESALE", 14500.00, 12, "San Francisco", "CA", "94105", "US", "United States", "415 Mission Street"},
		{"Noah", "Kim", "noah.kim@example.com", "+1-617-555-0129", "RETAIL", 159.00, 1, "Boston", "MA", "02110", "US", "United States", "100 Federal Street"},
		{"Olivia", "Martinez", "olivia.martinez@example.com", "+44-20-7946-0912", "GOLD", 2100.00, 5, "London", "Greater London", "EC2N 4AY", "GB", "United Kingdom", "100 Bishopsgate"},
	}

	custMap := make(map[string]int64)
	addrMap := make(map[string]int64)

	for _, c := range customers {
		if !idExists(ctx, database, "customers", "email", c.email) {
			grpID := groupMap[c.groupCode]
			_, err := exec(ctx, database, `
				INSERT INTO customers (
					first_name, last_name, email, password_hash, phone, customer_group_id,
					is_active, is_verified, total_orders, total_spent
				) VALUES (?, ?, ?, 'argon2id$development_hash', ?, ?, true, true, ?, ?)
			`, c.first, c.last, c.email, c.phone, grpID, c.orders, c.spent)
			if err != nil {
				return nil, nil, err
			}
		}
		cID := getId(ctx, database, "customers", "email", c.email)
		custMap[c.email] = cID

		// Insert Default Shipping address
		if !idExists(ctx, database, "addresses", "customer_id", cID) {
			_, err := exec(ctx, database, `
				INSERT INTO addresses (
					customer_id, address_type, first_name, last_name, address_line1,
					city, state_province, postal_code, country_code, country_name, phone,
					is_default_shipping, is_default_billing
				) VALUES (?, 'shipping', ?, ?, ?, ?, ?, ?, ?, ?, ?, true, true)
			`, cID, c.first, c.last, c.street, c.city, c.state, c.zip, c.countryCode, c.cName, c.phone)
			if err != nil {
				return nil, nil, err
			}
		}
		addrID := getId(ctx, database, "addresses", "customer_id", cID)
		addrMap[c.email] = addrID
	}

	log.Printf("  • Seeded %d customers with addresses", len(custMap))
	return custMap, addrMap, nil
}

func seedWarehousesAndStock(ctx context.Context, database *db.DB, variantMap map[string]int64) error {
	warehouses := []struct {
		name, code, street, city, state, zip, cCode, cName string
		isPrimary                                          bool
	}{
		{"East Coast Distribution Hub", "HUB-EAST", "400 Port Street", "Newark", "NJ", "07114", "US", "United States", true},
		{"West Coast Logistics Center", "HUB-WEST", "1900 E 7th St", "Los Angeles", "CA", "90021", "US", "United States", false},
		{"European Fulfillment Base", "HUB-EU", "Flughafenstrasse 15", "Frankfurt", "HE", "60549", "DE", "Germany", false},
	}

	whMap := make(map[string]int64)
	for _, w := range warehouses {
		if !idExists(ctx, database, "warehouses", "code", w.code) {
			_, err := exec(ctx, database, `
				INSERT INTO warehouses (
					name, code, address_line1, city, state, postal_code, country_code, country_name,
					is_active, is_primary
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, true, ?)
			`, w.name, w.code, w.street, w.city, w.state, w.zip, w.cCode, w.cName, w.isPrimary)
			if err != nil {
				return err
			}
		}
		whMap[w.code] = getId(ctx, database, "warehouses", "code", w.code)
	}

	hubEastID := whMap["HUB-EAST"]
	hubWestID := whMap["HUB-WEST"]

	stockCount := 0
	for _, variantID := range variantMap {
		for _, whID := range []int64{hubEastID, hubWestID} {
			var stockID int64
			err := queryRow(ctx, database, "SELECT id FROM stock WHERE product_variant_id = ? AND warehouse_id = ?", variantID, whID).Scan(&stockID)
			if err != nil || stockID == 0 {
				qty := 45
				if whID == hubEastID {
					qty = 75
				}
				_, err := exec(ctx, database, `
					INSERT INTO stock (
						product_variant_id, warehouse_id, quantity, reserved_quantity,
						available_quantity, reorder_point, reorder_quantity, is_active
					) VALUES (?, ?, ?, 5, ?, 10, 50, true)
				`, variantID, whID, qty, qty-5)
				if err != nil {
					return err
				}
				_ = queryRow(ctx, database, "SELECT id FROM stock WHERE product_variant_id = ? AND warehouse_id = ?", variantID, whID).Scan(&stockID)

				if stockID > 0 {
					_, _ = exec(ctx, database, `
						INSERT INTO stock_movements (
							stock_id, product_variant_id, warehouse_id, type,
							quantity, quantity_before, quantity_after, movement_date, notes
						) VALUES (?, ?, ?, 'inbound', ?, 0, ?, ?, 'Initial warehouse replenishment')
					`, stockID, variantID, whID, qty, qty, time.Now())
					stockCount++
				}
			}
		}
	}

	log.Printf("  • Seeded %d warehouses and %d stock records", len(whMap), stockCount)
	return nil
}

func seedCouponsAndPromotions(ctx context.Context, database *db.DB) error {
	coupons := []struct {
		code, name, desc, discType string
		discVal, minPurchase       float64
	}{
		{"FORGE2026", "20% Framework Launch Discount", "20% off entire store catalog", "percentage", 20.0, 50.0},
		{"SAVE50", "$50 Off Executive Cart", "Flat $50 off orders exceeding $250", "fixed", 50.0, 250.0},
		{"FREESHIP", "Complimentary Express Delivery", "100% shipping waiver on all items", "shipping", 10.0, 0.0},
		{"VIPCLUB", "VIP Exclusive 15% Reward", "Member reward for loyalty circle", "percentage", 15.0, 75.0},
	}

	now := time.Now()
	for _, c := range coupons {
		if !idExists(ctx, database, "coupons", "code", c.code) {
			_, err := exec(ctx, database, `
				INSERT INTO coupons (
					code, name, description, discount_type, discount_value, minimum_purchase_amount,
					valid_from, valid_until, is_active, is_public
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, true, true)
			`, c.code, c.name, c.desc, c.discType, c.discVal, c.minPurchase,
				now.Add(-48*time.Hour), now.Add(365*24*time.Hour))
			if err != nil {
				return err
			}
		}
	}

	banners := []struct {
		title, desc, image, link, placement string
	}{
		{"Experience Pro Computing with TitanBook", "Unrivaled engineering speed and display fidelity", "https://images.unsplash.com/photo-1517336714731-489689fd1ca8", "/products/titanbook-pro-16", "home_hero"},
		{"Sound Without Compromise - Apex ANC 900", "Beryllium drivers with active spatial depth", "https://images.unsplash.com/photo-1505740420928-5e560c06d30e", "/products/apex-anc-900", "home_hero"},
		{"Spring Lifestyle Apparel Collection", "Clean cuts, superfine merino, and Scandinavian silk", "https://images.unsplash.com/photo-1441986300917-64674bd600d8", "/categories/fashion-apparel", "home_banner"},
	}

	for _, b := range banners {
		if !idExists(ctx, database, "banners", "title", b.title) {
			_, _ = exec(ctx, database, `
				INSERT INTO banners (title, description, image_url, link_url, placement, start_date, is_active)
				VALUES (?, ?, ?, ?, ?, ?, true)
			`, b.title, b.desc, b.image, b.link, b.placement, now.Add(-24*time.Hour))
		}
	}

	log.Printf("  • Seeded %d coupons and %d store banners", len(coupons), len(banners))
	return nil
}

func seedOrdersAndPayments(ctx context.Context, database *db.DB, custMap, addrMap, prodMap map[string]int64) error {
	type orderDef struct {
		num, custEmail, status, payStatus, fulfillStatus string
		subtotal, discount, tax, shipping, total         float64
		payMethod, carrier, tracking                     string
		prodSlug                                         string
		qty                                              int
		unitPrice                                        float64
	}

	orders := []orderDef{
		{
			num: "ORD-2026-1001", custEmail: "alice.walker@example.com",
			status: "delivered", payStatus: "paid", fulfillStatus: "fulfilled",
			subtotal: 1999.00, discount: 0.0, tax: 160.00, shipping: 0.0, total: 2159.00,
			payMethod: "stripe", carrier: "FedEx Express", tracking: "FDX-8823-9011",
			prodSlug: "titanbook-pro-16", qty: 1, unitPrice: 1999.00,
		},
		{
			num: "ORD-2026-1002", custEmail: "marcus.vance@example.com",
			status: "shipped", payStatus: "paid", fulfillStatus: "shipped",
			subtotal: 499.00, discount: 50.0, tax: 40.00, shipping: 10.0, total: 499.00,
			payMethod: "paypal", carrier: "UPS Ground", tracking: "UPS-1Z-999-4321",
			prodSlug: "brewmaster-espresso", qty: 1, unitPrice: 499.00,
		},
		{
			num: "ORD-2026-1003", custEmail: "sophia.chen@example.com",
			status: "processing", payStatus: "paid", fulfillStatus: "unfulfilled",
			subtotal: 204.00, discount: 0.0, tax: 16.32, shipping: 10.0, total: 230.32,
			payMethod: "credit_card", carrier: "USPS Priority", tracking: "USPS-9400-1122",
			prodSlug: "cloudrunner-pro-sneakers", qty: 1, unitPrice: 159.00,
		},
		{
			num: "ORD-2026-1004", custEmail: "david.miller@example.com",
			status: "pending", payStatus: "pending", fulfillStatus: "unfulfilled",
			subtotal: 109.98, discount: 10.0, tax: 8.80, shipping: 5.0, total: 113.78,
			payMethod: "stripe", carrier: "Standard Ground", tracking: "",
			prodSlug: "go-blueprint-book", qty: 2, unitPrice: 49.99,
		},
		{
			num: "ORD-2026-1005", custEmail: "elena.rostova@example.com",
			status: "delivered", payStatus: "paid", fulfillStatus: "fulfilled",
			subtotal: 1450.00, discount: 0.0, tax: 116.00, shipping: 0.0, total: 1566.00,
			payMethod: "wire_transfer", carrier: "DHL Global", tracking: "DHL-4421-9988",
			prodSlug: "velocity-carbon-roadbike", qty: 1, unitPrice: 1450.00,
		},
		{
			num: "ORD-2026-1006", custEmail: "jordan.taylor@example.com",
			status: "shipped", payStatus: "paid", fulfillStatus: "shipped",
			subtotal: 9995.00, discount: 1999.00, tax: 639.68, shipping: 120.0, total: 8755.68,
			payMethod: "net30_invoice", carrier: "Freight Logistics", tracking: "FRT-9900-1100",
			prodSlug: "titanbook-pro-16", qty: 5, unitPrice: 1999.00,
		},
		{
			num: "ORD-2026-1007", custEmail: "noah.kim@example.com",
			status: "cancelled", payStatus: "refunded", fulfillStatus: "cancelled",
			subtotal: 119.00, discount: 0.0, tax: 9.52, shipping: 10.0, total: 138.52,
			payMethod: "stripe", carrier: "Standard Ground", tracking: "",
			prodSlug: "smartglow-floor-lamp", qty: 1, unitPrice: 119.00,
		},
		{
			num: "ORD-2026-1008", custEmail: "olivia.martinez@example.com",
			status: "delivered", payStatus: "paid", fulfillStatus: "fulfilled",
			subtotal: 289.00, discount: 0.0, tax: 23.12, shipping: 0.0, total: 312.12,
			payMethod: "apple_pay", carrier: "Royal Mail Special", tracking: "RM-8812-4411",
			prodSlug: "nordic-silk-trench", qty: 1, unitPrice: 289.00,
		},
	}

	count := 0
	for _, o := range orders {
		if !idExists(ctx, database, "orders", "order_number", o.num) {
			cID := custMap[o.custEmail]
			pID := prodMap[o.prodSlug]
			addrID := addrMap[o.custEmail]

			if cID > 0 && pID > 0 {
				_, err := exec(ctx, database, `
					INSERT INTO orders (
						order_number, customer_id, customer_email, customer_first_name, customer_last_name,
						subtotal, discount_amount, tax_amount, shipping_amount, total,
						status, payment_status, fulfillment_status, shipping_address_id, billing_address_id,
						payment_method, shipping_method, created_at, updated_at
					) VALUES (?, ?, ?, 'Customer', 'Account', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, o.num, cID, o.custEmail, o.subtotal, o.discount, o.tax, o.shipping, o.total,
					o.status, o.payStatus, o.fulfillStatus, addrID, addrID,
					o.payMethod, o.carrier,
					time.Now().Add(-120*time.Hour), time.Now().Add(-24*time.Hour))
				if err != nil {
					return err
				}

				orderID := getId(ctx, database, "orders", "order_number", o.num)

				// Insert Order Item
				_, _ = exec(ctx, database, `
					INSERT INTO order_items (
						order_id, product_id, product_name, product_sku, quantity,
						unit_price, total
					) VALUES (?, ?, ?, ?, ?, ?, ?)
				`, orderID, pID, o.prodSlug, "SKU-"+o.prodSlug, o.qty, o.unitPrice, o.subtotal)

				// Insert Payment
				_, _ = exec(ctx, database, `
					INSERT INTO payments (
						order_id, transaction_id, amount, currency, payment_method, status
					) VALUES (?, ?, ?, 'USD', ?, ?)
				`, orderID, "TXN-"+o.num, o.total, o.payMethod, o.payStatus)

				// Insert Shipment if shipped/delivered
				if o.tracking != "" {
					_, _ = exec(ctx, database, `
						INSERT INTO shipments (
							order_id, tracking_number, carrier, recipient_name,
							address_line1, city, postal_code, country_code, country_name,
							shipping_cost, status
						) VALUES (?, ?, ?, 'Customer Account', '100 Main St', 'Metropolis', '10001', 'US', 'United States', ?, ?)
					`, orderID, o.tracking, o.carrier, o.shipping, o.fulfillStatus)
				}

				count++
			}
		}
	}

	log.Printf("  • Seeded %d orders with line items, payments, and shipments", count)
	return nil
}

func seedReviews(ctx context.Context, database *db.DB, custMap, prodMap map[string]int64) error {
	reviews := []struct {
		prodSlug, email, title, content string
		rating                          int
	}{
		{"titanbook-pro-16", "alice.walker@example.com", "An absolute beast for software engineering", "Compiling massive microservices takes a fraction of the time compared to my prior laptop. Incredible cooling and zero thermal throttling.", 5},
		{"apex-anc-900", "marcus.vance@example.com", "Best acoustic isolation in the game", "The noise cancellation creates absolute silence in busy coffee shops and airport terminals. The soundstage is remarkably balanced.", 5},
		{"brewmaster-espresso", "david.miller@example.com", "Commercial cafe quality right on the counter", "The dual-boiler steam wand has immense pressure. Microfoam for latte art is silky smooth. PID temperature stability is pinpoint.", 5},
		{"cloudrunner-pro-sneakers", "sophia.chen@example.com", "Lightweight, responsive race day comfort", "The curved carbon plate delivers noticeable forward energy propulsion. PR on my half-marathon with zero blisters.", 4},
		{"go-blueprint-book", "jordan.taylor@example.com", "Must-read for every senior backend architect", "The concurrency models and memory optimization chapters alone paid for the book 100 times over. Deep, thorough, and highly practical.", 5},
		{"velocity-carbon-roadbike", "elena.rostova@example.com", "Precision aerodynamic masterpiece", "Stiff bottom bracket transfers every watt of power immediately. Electronic shifting is crisp, intuitive, and silent.", 5},
	}

	count := 0
	for _, r := range reviews {
		pID := prodMap[r.prodSlug]
		cID := custMap[r.email]
		if pID > 0 && cID > 0 {
			var exists bool
			_ = queryRow(ctx, database, "SELECT true FROM reviews WHERE product_id = ? AND customer_id = ?", pID, cID).Scan(&exists)
			if !exists {
				_, _ = exec(ctx, database, `
					INSERT INTO reviews (
						product_id, customer_id, title, content, rating,
						is_verified_purchase, status, is_featured, helpful_count
					) VALUES (?, ?, ?, ?, ?, true, 'approved', true, 12)
				`, pID, cID, r.title, r.content, r.rating)
				count++
			}
		}
	}

	log.Printf("  • Seeded %d customer reviews", count)
	return nil
}

func seedSupportAndFAQs(ctx context.Context, database *db.DB, custMap map[string]int64) error {
	tickets := []struct {
		num, email, subject, desc, status, priority, category string
		msgText                                               string
	}{
		{
			num: "TICK-2026-001", email: "alice.walker@example.com",
			subject: "Express delivery confirmation for Order #ORD-2026-1001",
			desc: "Requesting tracking verification and delivery instructions for FedEx signature drop-off.",
			status: "closed", priority: "high", category: "Shipping",
			msgText: "Your FedEx priority shipment has been flagged with direct signature required. Tracking number: FDX-8823-9011.",
		},
		{
			num: "TICK-2026-002", email: "marcus.vance@example.com",
			subject: "Firmware setup guidance for BrewMaster Touch",
			desc: "How do I calibrate the digital brew scale offset via the touch interface?",
			status: "closed", priority: "normal", category: "Technical",
			msgText: "Navigate to Settings > Calibration > Scale Tare and hold the power button for 3 seconds with the 100g test weight placed on the tray.",
		},
		{
			num: "TICK-2026-003", email: "sophia.chen@example.com",
			subject: "Sizing exchange for CloudRunner Pro Carbon Sneakers",
			desc: "Looking to verify availability for US 9.5 exchange before dispatching return package.",
			status: "in_progress", priority: "normal", category: "Returns",
			msgText: "Our East Coast warehouse currently has 15 units of US 9.5 reserved for exchange requests. Prepaid return shipping label has been dispatched to your email.",
		},
	}

	count := 0
	for _, t := range tickets {
		if !idExists(ctx, database, "support_tickets", "ticket_number", t.num) {
			cID := custMap[t.email]
			if cID > 0 {
				_, err := exec(ctx, database, `
					INSERT INTO support_tickets (
						ticket_number, customer_id, subject, description,
						status, priority, category
					) VALUES (?, ?, ?, ?, ?, ?, ?)
				`, t.num, cID, t.subject, t.desc, t.status, t.priority, t.category)
				if err != nil {
					return err
				}

				tID := getId(ctx, database, "support_tickets", "ticket_number", t.num)
				_, _ = exec(ctx, database, `
					INSERT INTO support_messages (ticket_id, sender_type, sender_name, message)
					VALUES (?, 'agent', 'Senior Support Specialist', ?)
				`, tID, t.msgText)
				count++
			}
		}
	}

	faqs := []struct {
		question, answer, cat string
		order                 int
	}{
		{"What is the international delivery timeline?", "Orders dispatched via express international air typically arrive within 2-4 business days with full door-to-door tracking.", "Shipping", 1},
		{"What is your standard return and exchange window?", "We provide a 30-day hassle-free return and exchange policy for all unworn apparel and unaltered hardware products.", "Returns", 2},
		{"How does the 2-year hardware warranty work?", "All computing hardware, audio monitors, and kitchen machinery include a comprehensive 24-month replacement warranty against manufacturing defects.", "Warranty", 3},
		{"Can I purchase with commercial invoice or Net 30 terms?", "Wholesale partners and enterprise accounts can request Net 30 invoicing through the wholesale checkout portal.", "Billing", 4},
	}

	for _, f := range faqs {
		var exists bool
		_ = queryRow(ctx, database, "SELECT true FROM faqs WHERE question = ?", f.question).Scan(&exists)
		if !exists {
			_, _ = exec(ctx, database, `
				INSERT INTO faqs (question, answer, category, is_public, sort_order)
				VALUES (?, ?, ?, true, ?)
			`, f.question, f.answer, f.cat, f.order)
		}
	}

	log.Printf("  • Seeded %d support tickets with agent messages and %d FAQs", count, len(faqs))
	return nil
}

func seedUsersAndGroups(ctx context.Context, database *db.DB) error {
	groups := []struct {
		name, desc, perms string
	}{
		{"Super Administrators", "Full unrestricted administrative access to all models and settings", `["*"]`},
		{"Store Managers", "Catalog management, order fulfillment, and promotions administration", `["catalog.*", "orders.*", "promotions.*"]`},
		{"Support Specialists", "Customer records, support ticketing, and return processing", `["customers.*", "support.*"]`},
	}

	for _, g := range groups {
		if !idExists(ctx, database, "auth_group", "name", g.name) {
			_, _ = exec(ctx, database, `
				INSERT INTO auth_group (name, description, permissions)
				VALUES (?, ?, ?)
			`, g.name, g.desc, g.perms)
		}
	}

	users := []struct {
		username, email, first, last string
		isStaff, isSuper             bool
	}{
		{"admin", "admin@forge-ecommerce.local", "System", "Administrator", true, true},
		{"store_manager", "manager@forge-ecommerce.local", "Jordan", "RetailOps", true, false},
		{"support_lead", "support@forge-ecommerce.local", "Devon", "CustomerCare", true, false},
	}

	for _, u := range users {
		if !idExists(ctx, database, "users_user", "username", u.username) {
			_, _ = exec(ctx, database, `
				INSERT INTO users_user (
					username, email, password_hash, first_name, last_name,
					is_active, is_staff, is_superuser
				) VALUES (?, ?, 'argon2id$development_hash', ?, ?, true, ?, ?)
			`, u.username, u.email, u.first, u.last, u.isStaff, u.isSuper)
		}
	}

	log.Printf("  • Seeded administrative users and permission groups")
	return nil
}

// seedAuditHistory generates realistic admin audit log entries in the DefaultSite
// so the new Audit History Viewer shows instant timeline activity out of the box.
func seedAuditHistory(ctx context.Context) {
	site := admin.DefaultSite
	if site == nil {
		return
	}

	entries := []struct {
		modelName string
		objID     string
		repr      string
		action    core.ActionType
		changes   string
		user      string
	}{
		{"products", "1", "TitanBook Pro 16", core.ActionAdd, `{"name":"TitanBook Pro 16","price":1999.00,"sku":"LAP-TITAN-16","status":"active"}`, "admin"},
		{"products", "1", "TitanBook Pro 16", core.ActionChange, `{"stock_quantity":{"old":40,"new":50},"is_featured":{"old":false,"new":true}}`, "admin"},
		{"products", "3", "Apex ANC 900 Studio Wireless", core.ActionAdd, `{"name":"Apex ANC 900","price":349.00,"sku":"AUD-APEX-900"}`, "admin"},
		{"orders", "1", "Order #ORD-2026-1001", core.ActionChange, `{"status":{"old":"processing","new":"shipped"},"tracking_number":"FDX-8823-9011"}`, "store_manager"},
		{"orders", "1", "Order #ORD-2026-1001", core.ActionChange, `{"status":{"old":"shipped","new":"delivered"}}`, "system"},
		{"warehouses", "1", "East Coast Distribution Hub", core.ActionChange, `{"is_primary":{"old":false,"new":true}}`, "admin"},
		{"categories", "1", "Electronics", core.ActionChange, `{"sort_order":{"old":0,"new":1}}`, "admin"},
	}

	for _, e := range entries {
		adminModel, err := site.GetRegistry().Get(e.modelName)
		if err == nil && adminModel != nil {
			_ = adminModel.LogAction(ctx, e.user, e.objID, e.repr, e.action, e.changes)
		}
	}

	log.Printf("  • Pre-populated audit log timeline with realistic changes")
}
