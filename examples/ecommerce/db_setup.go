package main

import (
	"context"
	"log"
	"strings"

	"github.com/forgego/forge/db"
)

// adaptDDL rewrites PostgreSQL-specific DDL for SQLite, which requires
// INTEGER PRIMARY KEY (rowid alias) for autoincrement instead of SERIAL.
func adaptDDL(driver, stmt string) string {
	if driver == "sqlite3" || driver == "sqlite" {
		stmt = strings.ReplaceAll(stmt, "SERIAL PRIMARY KEY", "INTEGER PRIMARY KEY AUTOINCREMENT")
	}
	return stmt
}

// SetupSchema creates the database schema for the ecommerce example
func SetupSchema(database *db.DB) {
	ctx := context.Background()
	log.Println("🛠️  Setting up database schema...")

	// Categories
	_, err := database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			slug VARCHAR(200) NOT NULL UNIQUE,
			description TEXT,
			parent_id INTEGER REFERENCES categories(id),
			image_url VARCHAR(500),
			sort_order INTEGER DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			level INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_category_slug ON categories(slug);
		CREATE INDEX IF NOT EXISTS idx_category_parent ON categories(parent_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create categories table: %v", err)
	}

	// Brands
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS brands (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			slug VARCHAR(200) NOT NULL UNIQUE,
			description TEXT,
			logo_url VARCHAR(500),
			website_url VARCHAR(500),
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_brand_slug ON brands(slug);
	`))
	if err != nil {
		log.Fatalf("Failed to create brands table: %v", err)
	}
	
	// Products
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL UNIQUE,
			sku VARCHAR(100) NOT NULL UNIQUE,
			description TEXT NOT NULL,
			short_description TEXT,
			category_id INTEGER NOT NULL REFERENCES categories(id),
			brand_id INTEGER REFERENCES brands(id),
			price NUMERIC(10, 2) NOT NULL,
			cost_price NUMERIC(10, 2),
			compare_at_price NUMERIC(10, 2),
			stock_quantity INTEGER DEFAULT 0,
			track_inventory BOOLEAN DEFAULT TRUE,
			allow_backorder BOOLEAN DEFAULT FALSE,
			weight NUMERIC(10, 2),
			length NUMERIC(10, 2),
			width NUMERIC(10, 2),
			height NUMERIC(10, 2),
			is_active BOOLEAN DEFAULT TRUE,
			is_featured BOOLEAN DEFAULT FALSE,
			is_digital BOOLEAN DEFAULT FALSE,
			meta_title VARCHAR(255),
			meta_description TEXT,
			meta_keywords VARCHAR(500),
			view_count INTEGER DEFAULT 0,
			order_count INTEGER DEFAULT 0,
			rating_average NUMERIC(3, 2) DEFAULT 0.0,
			rating_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			published_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_product_slug ON products(slug);
		CREATE INDEX IF NOT EXISTS idx_product_sku ON products(sku);
		CREATE INDEX IF NOT EXISTS idx_product_category ON products(category_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create products table: %v", err)
	}

	// Product Variants
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS product_variants (
			id SERIAL PRIMARY KEY,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			sku VARCHAR(100) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			option1_name VARCHAR(100),
			option1_value VARCHAR(100),
			option2_name VARCHAR(100),
			option2_value VARCHAR(100),
			option3_name VARCHAR(100),
			option3_value VARCHAR(100),
			price NUMERIC(10, 2),
			compare_at_price NUMERIC(10, 2),
			cost_price NUMERIC(10, 2),
			stock_quantity INTEGER DEFAULT 0,
			reserved_quantity INTEGER DEFAULT 0,
			track_inventory BOOLEAN DEFAULT TRUE,
			weight NUMERIC(10, 2),
			length NUMERIC(10, 2),
			width NUMERIC(10, 2),
			height NUMERIC(10, 2),
			is_active BOOLEAN DEFAULT TRUE,
			is_default BOOLEAN DEFAULT FALSE,
			image_url VARCHAR(500),
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_variant_sku ON product_variants(sku);
		CREATE INDEX IF NOT EXISTS idx_variant_product ON product_variants(product_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create product_variants table: %v", err)
	}

	// Product Images
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS product_images (
			id SERIAL PRIMARY KEY,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
			image_url VARCHAR(500) NOT NULL,
			thumbnail_url VARCHAR(500),
			alt_text VARCHAR(255),
			sort_order INTEGER DEFAULT 0,
			is_primary BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_image_product ON product_images(product_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create product_images table: %v", err)
	}

	// Product Attributes
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS product_attributes (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL UNIQUE,
			code VARCHAR(100) NOT NULL UNIQUE,
			type VARCHAR(50) NOT NULL,
			is_filterable BOOLEAN DEFAULT TRUE,
			is_visible BOOLEAN DEFAULT TRUE,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`))
	if err != nil {
		log.Fatalf("Failed to create product_attributes table: %v", err)
	}

	// Product Attribute Values
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS product_attribute_values (
			id SERIAL PRIMARY KEY,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			attribute_id INTEGER NOT NULL REFERENCES product_attributes(id) ON DELETE CASCADE,
			value VARCHAR(500) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(product_id, attribute_id)
		);
		CREATE INDEX IF NOT EXISTS idx_attr_value_product ON product_attribute_values(product_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create product_attribute_values table: %v", err)
	}

	// --- CUSTOMERS APP ---

	// Customer Groups
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS customer_groups (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL UNIQUE,
			code VARCHAR(50) NOT NULL UNIQUE,
			description TEXT,
			discount_percentage NUMERIC(5, 2) DEFAULT 0.0,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`))
	if err != nil {
		log.Fatalf("Failed to create customer_groups table: %v", err)
	}

	// Customers
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS customers (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			phone VARCHAR(20),
			date_of_birth DATE,
			gender VARCHAR(20),
			company_name VARCHAR(200),
			tax_id VARCHAR(50),
			customer_group_id INTEGER REFERENCES customer_groups(id) ON DELETE SET NULL,
			is_active BOOLEAN DEFAULT TRUE,
			is_verified BOOLEAN DEFAULT FALSE,
			accepts_marketing BOOLEAN DEFAULT FALSE,
			verification_token VARCHAR(255),
			reset_password_token VARCHAR(255),
			reset_password_expires_at TIMESTAMP,
			last_login_at TIMESTAMP,
			last_login_ip VARCHAR(45),
			total_orders INTEGER DEFAULT 0,
			total_spent NUMERIC(10, 2) DEFAULT 0.0,
			average_order_value NUMERIC(10, 2) DEFAULT 0.0,
			preferred_language VARCHAR(10) DEFAULT 'en',
			preferred_currency VARCHAR(3) DEFAULT 'USD',
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_customer_email ON customers(email);
		CREATE INDEX IF NOT EXISTS idx_customer_group ON customers(customer_group_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create customers table: %v", err)
	}

	// Addresses
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS addresses (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			address_type VARCHAR(20) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			company_name VARCHAR(200),
			phone VARCHAR(20),
			address_line1 VARCHAR(255) NOT NULL,
			address_line2 VARCHAR(255),
			city VARCHAR(100) NOT NULL,
			state_province VARCHAR(100),
			postal_code VARCHAR(20) NOT NULL,
			country_code VARCHAR(2) NOT NULL,
			country_name VARCHAR(100) NOT NULL,
			latitude NUMERIC(10, 6),
			longitude NUMERIC(10, 6),
			is_default_shipping BOOLEAN DEFAULT FALSE,
			is_default_billing BOOLEAN DEFAULT FALSE,
			delivery_instructions TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_address_customer ON addresses(customer_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create addresses table: %v", err)
	}

	// Wish Lists
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS wish_lists (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			name VARCHAR(200) NOT NULL,
			description TEXT,
			is_public BOOLEAN DEFAULT FALSE,
			share_token VARCHAR(100),
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_wishlist_customer ON wish_lists(customer_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create wish_lists table: %v", err)
	}

	// Wish List Items
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS wish_list_items (
			id SERIAL PRIMARY KEY,
			wish_list_id INTEGER NOT NULL REFERENCES wish_lists(id) ON DELETE CASCADE,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
			desired_quantity INTEGER DEFAULT 1,
			price_when_added NUMERIC(10, 2),
			notes TEXT,
			priority INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(wish_list_id, product_id, variant_id)
		);
		CREATE INDEX IF NOT EXISTS idx_wishlist_item_list ON wish_list_items(wish_list_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create wish_list_items table: %v", err)
	}

	// --- MARKETING APP ---

	// Coupons
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS coupons (
			id SERIAL PRIMARY KEY,
			code VARCHAR(50) NOT NULL UNIQUE,
			name VARCHAR(200) NOT NULL,
			description TEXT,
			discount_type VARCHAR(20) NOT NULL,
			discount_value NUMERIC(10, 2) NOT NULL,
			minimum_purchase_amount NUMERIC(10, 2),
			maximum_discount_amount NUMERIC(10, 2),
			usage_limit INTEGER,
			usage_limit_per_customer INTEGER DEFAULT 1,
			usage_count INTEGER DEFAULT 0,
			applies_to_all_products BOOLEAN DEFAULT TRUE,
			product_ids VARCHAR(500),
			category_ids VARCHAR(500),
			excluded_product_ids VARCHAR(500),
			applies_to_all_customers BOOLEAN DEFAULT TRUE,
			customer_group_ids VARCHAR(500),
			customer_email_list TEXT,
			valid_from TIMESTAMP NOT NULL,
			valid_until TIMESTAMP,
			is_active BOOLEAN DEFAULT TRUE,
			is_public BOOLEAN DEFAULT TRUE,
			priority INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_coupon_code ON coupons(code);
	`))
	if err != nil {
		log.Fatalf("Failed to create coupons table: %v", err)
	}

	// Coupon Usage
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS coupon_usage (
			id SERIAL PRIMARY KEY,
			coupon_id INTEGER NOT NULL REFERENCES coupons(id) ON DELETE CASCADE,
			order_id INTEGER,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			discount_amount NUMERIC(10, 2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_usage_coupon ON coupon_usage(coupon_id);
		CREATE INDEX IF NOT EXISTS idx_usage_customer ON coupon_usage(customer_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create coupon_usage table: %v", err)
	}

	// Reviews
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS reviews (
			id SERIAL PRIMARY KEY,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			order_id INTEGER,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			rating INTEGER NOT NULL,
			is_verified_purchase BOOLEAN DEFAULT FALSE,
			status VARCHAR(20) DEFAULT 'pending',
			is_featured BOOLEAN DEFAULT FALSE,
			helpful_count INTEGER DEFAULT 0,
			not_helpful_count INTEGER DEFAULT 0,
			merchant_response TEXT,
			merchant_response_at TIMESTAMP,
			merchant_response_by_user_id INTEGER,
			report_count INTEGER DEFAULT 0,
			report_reasons TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			approved_at TIMESTAMP,
			UNIQUE(product_id, customer_id, order_id)
		);
		CREATE INDEX IF NOT EXISTS idx_review_product ON reviews(product_id);
		CREATE INDEX IF NOT EXISTS idx_review_customer ON reviews(customer_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create reviews table: %v", err)
	}

	// Review Images
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS review_images (
			id SERIAL PRIMARY KEY,
			review_id INTEGER NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
			image_url VARCHAR(500) NOT NULL,
			thumbnail_url VARCHAR(500),
			alt_text VARCHAR(255),
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_review_image_review ON review_images(review_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create review_images table: %v", err)
	}

	// Review Helpfulness
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS review_helpfulness (
			id SERIAL PRIMARY KEY,
			review_id INTEGER NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
			customer_id INTEGER REFERENCES customers(id) ON DELETE CASCADE,
			is_helpful BOOLEAN NOT NULL,
			ip_address VARCHAR(45),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(review_id, customer_id),
			UNIQUE(review_id, ip_address)
		);
		CREATE INDEX IF NOT EXISTS idx_helpfulness_review ON review_helpfulness(review_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create review_helpfulness table: %v", err)
	}

	// Product Questions
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS product_questions (
			id SERIAL PRIMARY KEY,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			question TEXT NOT NULL,
			answer TEXT,
			answered_at TIMESTAMP,
			answered_by_user_id INTEGER,
			answered_by_user_name VARCHAR(200),
			status VARCHAR(20) DEFAULT 'pending',
			is_public BOOLEAN DEFAULT TRUE,
			helpful_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_question_product ON product_questions(product_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create product_questions table: %v", err)
	}

	// --- ORDERS APP ---

	// Carts
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS carts (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER REFERENCES customers(id) ON DELETE CASCADE,
			session_id VARCHAR(255),
			guest_email VARCHAR(255),
			subtotal NUMERIC(10, 2) DEFAULT 0.0,
			discount_amount NUMERIC(10, 2) DEFAULT 0.0,
			tax_amount NUMERIC(10, 2) DEFAULT 0.0,
			shipping_amount NUMERIC(10, 2) DEFAULT 0.0,
			total NUMERIC(10, 2) DEFAULT 0.0,
			coupon_id INTEGER REFERENCES coupons(id) ON DELETE SET NULL,
			coupon_code VARCHAR(50),
			status VARCHAR(20) DEFAULT 'active',
			is_abandoned BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_activity_at TIMESTAMP,
			converted_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_cart_customer ON carts(customer_id);
		CREATE INDEX IF NOT EXISTS idx_cart_session ON carts(session_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create carts table: %v", err)
	}

	// Cart Items
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS cart_items (
			id SERIAL PRIMARY KEY,
			cart_id INTEGER NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
			quantity INTEGER DEFAULT 1,
			unit_price NUMERIC(10, 2) NOT NULL,
			discount_amount NUMERIC(10, 2) DEFAULT 0.0,
			tax_amount NUMERIC(10, 2) DEFAULT 0.0,
			total NUMERIC(10, 2) NOT NULL,
			product_name VARCHAR(255),
			variant_name VARCHAR(255),
			image_url VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(cart_id, product_id, variant_id)
		);
		CREATE INDEX IF NOT EXISTS idx_cart_item_cart ON cart_items(cart_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create cart_items table: %v", err)
	}

	// Orders
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			order_number VARCHAR(50) NOT NULL UNIQUE,
			customer_id INTEGER NOT NULL REFERENCES customers(id),
			customer_email VARCHAR(255) NOT NULL,
			customer_first_name VARCHAR(100) NOT NULL,
			customer_last_name VARCHAR(100) NOT NULL,
			customer_phone VARCHAR(20),
			subtotal NUMERIC(10, 2) NOT NULL,
			discount_amount NUMERIC(10, 2) DEFAULT 0.0,
			tax_amount NUMERIC(10, 2) NOT NULL,
			shipping_amount NUMERIC(10, 2) NOT NULL,
			total NUMERIC(10, 2) NOT NULL,
			coupon_id INTEGER REFERENCES coupons(id) ON DELETE SET NULL,
			coupon_code VARCHAR(50),
			coupon_discount NUMERIC(10, 2) DEFAULT 0.0,
			status VARCHAR(20) DEFAULT 'pending',
			payment_status VARCHAR(20) DEFAULT 'pending',
			fulfillment_status VARCHAR(20) DEFAULT 'unfulfilled',
			shipping_address_id INTEGER REFERENCES addresses(id),
			billing_address_id INTEGER REFERENCES addresses(id),
			shipping_first_name VARCHAR(100),
			shipping_last_name VARCHAR(100),
			shipping_company VARCHAR(200),
			shipping_address_line1 VARCHAR(255),
			shipping_address_line2 VARCHAR(255),
			shipping_city VARCHAR(100),
			shipping_state VARCHAR(100),
			shipping_postal_code VARCHAR(20),
			shipping_country_code VARCHAR(2),
			shipping_country_name VARCHAR(100),
			shipping_phone VARCHAR(20),
			billing_first_name VARCHAR(100),
			billing_last_name VARCHAR(100),
			billing_company VARCHAR(200),
			billing_address_line1 VARCHAR(255),
			billing_address_line2 VARCHAR(255),
			billing_city VARCHAR(100),
			billing_state VARCHAR(100),
			billing_postal_code VARCHAR(20),
			billing_country_code VARCHAR(2),
			billing_country_name VARCHAR(100),
			payment_method VARCHAR(50),
			payment_transaction_id VARCHAR(255),
			shipping_method VARCHAR(100),
			tracking_number VARCHAR(255),
			carrier VARCHAR(100),
			customer_notes TEXT,
			admin_notes TEXT,
			ip_address VARCHAR(45),
			user_agent VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			paid_at TIMESTAMP,
			shipped_at TIMESTAMP,
			delivered_at TIMESTAMP,
			cancelled_at TIMESTAMP,
			expires_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_order_number ON orders(order_number);
		CREATE INDEX IF NOT EXISTS idx_order_customer ON orders(customer_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create orders table: %v", err)
	}

	// Update coupon_usage and reviews order_id foreign key (since we couldn't add FK before order existed)
	// Actually, we can just create the FK constraint now
	_, _ = database.ExecContext(ctx, adaptDDL(database.Driver, `
		ALTER TABLE coupon_usage ADD CONSTRAINT fk_coupon_usage_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE;
		ALTER TABLE reviews ADD CONSTRAINT fk_reviews_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE SET NULL;
	`))

	// Order Items
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_id INTEGER NOT NULL REFERENCES products(id),
			variant_id INTEGER REFERENCES product_variants(id),
			product_name VARCHAR(255) NOT NULL,
			product_sku VARCHAR(100) NOT NULL,
			variant_name VARCHAR(255),
			variant_sku VARCHAR(100),
			image_url VARCHAR(500),
			quantity INTEGER DEFAULT 1,
			unit_price NUMERIC(10, 2) NOT NULL,
			discount_amount NUMERIC(10, 2) DEFAULT 0.0,
			tax_amount NUMERIC(10, 2) DEFAULT 0.0,
			total NUMERIC(10, 2) NOT NULL,
			quantity_fulfilled INTEGER DEFAULT 0,
			quantity_refunded INTEGER DEFAULT 0,
			fulfillment_status VARCHAR(20) DEFAULT 'unfulfilled',
			weight NUMERIC(10, 2),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_order_item_order ON order_items(order_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create order_items table: %v", err)
	}

	// Payments
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS payments (
			id SERIAL PRIMARY KEY,
			order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			transaction_id VARCHAR(255),
			amount NUMERIC(10, 2) NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD',
			payment_method VARCHAR(50) NOT NULL,
			payment_gateway VARCHAR(50),
			card_last4 VARCHAR(4),
			card_brand VARCHAR(50),
			status VARCHAR(20) DEFAULT 'pending',
			gateway_response TEXT,
			failure_reason VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP,
			failed_at TIMESTAMP,
			refunded_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_payment_order ON payments(order_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create payments table: %v", err)
	}

	// Shipments
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS shipments (
			id SERIAL PRIMARY KEY,
			order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			tracking_number VARCHAR(255) NOT NULL,
			carrier VARCHAR(100) NOT NULL,
			service_level VARCHAR(100),
			recipient_name VARCHAR(200) NOT NULL,
			address_line1 VARCHAR(255) NOT NULL,
			address_line2 VARCHAR(255),
			city VARCHAR(100) NOT NULL,
			state VARCHAR(100),
			postal_code VARCHAR(20) NOT NULL,
			country_code VARCHAR(2) NOT NULL,
			country_name VARCHAR(100) NOT NULL,
			phone VARCHAR(20),
			weight NUMERIC(10, 2),
			shipping_cost NUMERIC(10, 2) NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			tracking_events TEXT,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			shipped_at TIMESTAMP,
			estimated_delivery_at TIMESTAMP,
			delivered_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_shipment_order ON shipments(order_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create shipments table: %v", err)
	}

	// --- INVENTORY APP ---

	// Warehouses
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS warehouses (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			code VARCHAR(50) NOT NULL UNIQUE,
			contact_name VARCHAR(200),
			contact_email VARCHAR(255),
			contact_phone VARCHAR(20),
			address_line1 VARCHAR(255) NOT NULL,
			address_line2 VARCHAR(255),
			city VARCHAR(100) NOT NULL,
			state VARCHAR(100),
			postal_code VARCHAR(20) NOT NULL,
			country_code VARCHAR(2) NOT NULL,
			country_name VARCHAR(100) NOT NULL,
			latitude NUMERIC(10, 6),
			longitude NUMERIC(10, 6),
			is_active BOOLEAN DEFAULT TRUE,
			is_primary BOOLEAN DEFAULT FALSE,
			priority INTEGER DEFAULT 0,
			total_capacity INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_warehouse_code ON warehouses(code);
	`))
	if err != nil {
		log.Fatalf("Failed to create warehouses table: %v", err)
	}

	// Stock
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS stock (
			id SERIAL PRIMARY KEY,
			product_variant_id INTEGER NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
			warehouse_id INTEGER NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
			quantity INTEGER DEFAULT 0,
			reserved_quantity INTEGER DEFAULT 0,
			available_quantity INTEGER DEFAULT 0,
			reorder_point INTEGER DEFAULT 10,
			reorder_quantity INTEGER DEFAULT 50,
			bin_location VARCHAR(100),
			is_active BOOLEAN DEFAULT TRUE,
			allow_backorder BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_counted_at TIMESTAMP,
			UNIQUE(product_variant_id, warehouse_id)
		);
		CREATE INDEX IF NOT EXISTS idx_stock_variant ON stock(product_variant_id);
		CREATE INDEX IF NOT EXISTS idx_stock_warehouse ON stock(warehouse_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create stock table: %v", err)
	}

	// Stock Movements
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS stock_movements (
			id SERIAL PRIMARY KEY,
			stock_id INTEGER NOT NULL REFERENCES stock(id) ON DELETE CASCADE,
			product_variant_id INTEGER NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
			warehouse_id INTEGER NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			quantity INTEGER NOT NULL,
			quantity_before INTEGER NOT NULL,
			quantity_after INTEGER NOT NULL,
			reference_type VARCHAR(50),
			reference_id INTEGER,
			reference_number VARCHAR(100),
			from_warehouse_id INTEGER REFERENCES warehouses(id) ON DELETE SET NULL,
			to_warehouse_id INTEGER REFERENCES warehouses(id) ON DELETE SET NULL,
			reason VARCHAR(255),
			notes TEXT,
			user_id INTEGER,
			user_name VARCHAR(200),
			unit_cost NUMERIC(10, 2),
			total_cost NUMERIC(10, 2),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			movement_date TIMESTAMP NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_movement_stock ON stock_movements(stock_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create stock_movements table: %v", err)
	}

	// Stock Alerts
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS stock_alerts (
			id SERIAL PRIMARY KEY,
			stock_id INTEGER NOT NULL REFERENCES stock(id) ON DELETE CASCADE,
			product_variant_id INTEGER NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
			warehouse_id INTEGER NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
			alert_type VARCHAR(50) NOT NULL,
			current_quantity INTEGER NOT NULL,
			threshold INTEGER NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			resolved_by_user_id INTEGER,
			resolved_by_user_name VARCHAR(200),
			resolution_notes TEXT,
			notification_sent BOOLEAN DEFAULT FALSE,
			notification_sent_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			acknowledged_at TIMESTAMP,
			resolved_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_alert_stock ON stock_alerts(stock_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create stock_alerts table: %v", err)
	}

	// Stock Transfers
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS stock_transfers (
			id SERIAL PRIMARY KEY,
			transfer_number VARCHAR(50) NOT NULL UNIQUE,
			from_warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
			to_warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
			product_variant_id INTEGER NOT NULL REFERENCES product_variants(id),
			quantity INTEGER NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			tracking_number VARCHAR(255),
			carrier VARCHAR(100),
			notes TEXT,
			reason TEXT,
			requested_by_user_id INTEGER,
			requested_by_user_name VARCHAR(200),
			approved_by_user_id INTEGER,
			approved_by_user_name VARCHAR(200),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			shipped_at TIMESTAMP,
			completed_at TIMESTAMP,
			cancelled_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_transfer_number ON stock_transfers(transfer_number);
	`))
	if err != nil {
		log.Fatalf("Failed to create stock_transfers table: %v", err)
	}

	log.Println("✅ Database schema setup complete")
}
