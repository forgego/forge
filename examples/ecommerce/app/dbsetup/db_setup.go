package dbsetup

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

	// --- USERS APP ---

	// Users
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS users_user (
			id SERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL UNIQUE,
			email VARCHAR(254) NOT NULL,
			password_hash VARCHAR(255),
			first_name VARCHAR(150),
			last_name VARCHAR(150),
			is_active BOOLEAN DEFAULT TRUE,
			is_staff BOOLEAN DEFAULT FALSE,
			is_superuser BOOLEAN DEFAULT FALSE,
			avatar VARCHAR(500),
			last_login TIMESTAMP,
			date_joined TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_user_username ON users_user(username);
		CREATE INDEX IF NOT EXISTS idx_user_email ON users_user(email);
	`))
	if err != nil {
		log.Fatalf("Failed to create users_user table: %v", err)
	}

	// Auth Groups
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS auth_group (
			id SERIAL PRIMARY KEY,
			name VARCHAR(150) NOT NULL UNIQUE,
			description VARCHAR(500),
			permissions TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_auth_group_name ON auth_group(name);
	`))
	if err != nil {
		log.Fatalf("Failed to create auth_group table: %v", err)
	}

	// --- PROMOTIONS APP ---

	// Promotions
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS promotions (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			code VARCHAR(50) UNIQUE,
			description TEXT,
			discount_type VARCHAR(20) NOT NULL,
			discount_value NUMERIC(10, 2) NOT NULL,
			min_purchase NUMERIC(10, 2) DEFAULT 0.0,
			max_discount NUMERIC(10, 2) DEFAULT 0.0,
			start_date TIMESTAMP NOT NULL,
			end_date TIMESTAMP,
			usage_limit INTEGER DEFAULT 0,
			usage_count INTEGER DEFAULT 0,
			per_customer_limit INTEGER DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			is_stackable BOOLEAN DEFAULT FALSE,
			priority INTEGER DEFAULT 0,
			applies_to VARCHAR(50) DEFAULT 'all',
			target_entity_ids TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_promotion_code ON promotions(code);
		CREATE INDEX IF NOT EXISTS idx_promotion_active ON promotions(is_active);
	`))
	if err != nil {
		log.Fatalf("Failed to create promotions table: %v", err)
	}

	// Promotion Rules
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS promotion_rules (
			id SERIAL PRIMARY KEY,
			promotion_id INTEGER NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
			rule_type VARCHAR(50) NOT NULL,
			field VARCHAR(100) NOT NULL,
			operator VARCHAR(20) NOT NULL,
			value VARCHAR(500) NOT NULL,
			logic_type VARCHAR(10) DEFAULT 'AND',
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_promotion_rule_promotion ON promotion_rules(promotion_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create promotion_rules table: %v", err)
	}

	// Banners
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS banners (
			id SERIAL PRIMARY KEY,
			title VARCHAR(200) NOT NULL,
			description TEXT,
			image_url VARCHAR(500) NOT NULL,
			link_url VARCHAR(500),
			placement VARCHAR(50) DEFAULT 'home_hero',
			start_date TIMESTAMP NOT NULL,
			end_date TIMESTAMP,
			is_active BOOLEAN DEFAULT TRUE,
			sort_order INTEGER DEFAULT 0,
			click_count INTEGER DEFAULT 0,
			view_count INTEGER DEFAULT 0,
			target_group VARCHAR(50) DEFAULT 'all',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_banner_active ON banners(is_active);
		CREATE INDEX IF NOT EXISTS idx_banner_placement ON banners(placement);
	`))
	if err != nil {
		log.Fatalf("Failed to create banners table: %v", err)
	}

	// Newsletter Subscriptions
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS newsletter_subscriptions (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			first_name VARCHAR(100),
			last_name VARCHAR(100),
			customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
			status VARCHAR(20) DEFAULT 'pending',
			source VARCHAR(50),
			ip_address VARCHAR(45),
			confirmed_at TIMESTAMP,
			unsubscribed_at TIMESTAMP,
			preferences TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_newsletter_email ON newsletter_subscriptions(email);
		CREATE INDEX IF NOT EXISTS idx_newsletter_status ON newsletter_subscriptions(status);
	`))
	if err != nil {
		log.Fatalf("Failed to create newsletter_subscriptions table: %v", err)
	}

	// Promotion Usages
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS promotion_usages (
			id SERIAL PRIMARY KEY,
			promotion_id INTEGER NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			used_at TIMESTAMP NOT NULL,
			discount_amount NUMERIC(10, 2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_promotion_usage_promotion ON promotion_usages(promotion_id);
		CREATE INDEX IF NOT EXISTS idx_promotion_usage_customer ON promotion_usages(customer_id);
		CREATE INDEX IF NOT EXISTS idx_promotion_usage_order ON promotion_usages(order_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create promotion_usages table: %v", err)
	}

	// --- SUPPORT APP ---

	// Support Tickets
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS support_tickets (
			id SERIAL PRIMARY KEY,
			ticket_number VARCHAR(50) NOT NULL UNIQUE,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			subject VARCHAR(300) NOT NULL,
			description TEXT NOT NULL,
			status VARCHAR(20) DEFAULT 'open',
			priority VARCHAR(20) DEFAULT 'normal',
			category VARCHAR(50),
			assigned_to INTEGER,
			order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
			source VARCHAR(50) DEFAULT 'web',
			resolution TEXT,
			resolved_at TIMESTAMP,
			closed_at TIMESTAMP,
			first_response_at TIMESTAMP,
			tags TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_ticket_number ON support_tickets(ticket_number);
		CREATE INDEX IF NOT EXISTS idx_ticket_customer ON support_tickets(customer_id);
		CREATE INDEX IF NOT EXISTS idx_ticket_status ON support_tickets(status);
	`))
	if err != nil {
		log.Fatalf("Failed to create support_tickets table: %v", err)
	}

	// Support Messages
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS support_messages (
			id SERIAL PRIMARY KEY,
			ticket_id INTEGER NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
			sender_type VARCHAR(20) NOT NULL,
			sender_id INTEGER,
			sender_name VARCHAR(200),
			message TEXT NOT NULL,
			is_internal BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_support_message_ticket ON support_messages(ticket_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create support_messages table: %v", err)
	}

	// Return Requests
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS return_requests (
			id SERIAL PRIMARY KEY,
			return_number VARCHAR(50) NOT NULL UNIQUE,
			order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			reason VARCHAR(50) NOT NULL,
			description TEXT,
			status VARCHAR(20) DEFAULT 'pending',
			return_method VARCHAR(50),
			refund_method VARCHAR(50) DEFAULT 'original',
			refund_amount NUMERIC(10, 2) DEFAULT 0.0,
			restock_fee NUMERIC(10, 2) DEFAULT 0.0,
			shipping_label VARCHAR(500),
			tracking_number VARCHAR(100),
			approved_at TIMESTAMP,
			approved_by INTEGER,
			received_at TIMESTAMP,
			processed_at TIMESTAMP,
			refunded_at TIMESTAMP,
			rejected_at TIMESTAMP,
			rejection_note TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_return_number ON return_requests(return_number);
		CREATE INDEX IF NOT EXISTS idx_return_order ON return_requests(order_id);
		CREATE INDEX IF NOT EXISTS idx_return_customer ON return_requests(customer_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create return_requests table: %v", err)
	}

	// Return Items
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS return_items (
			id SERIAL PRIMARY KEY,
			return_request_id INTEGER NOT NULL REFERENCES return_requests(id) ON DELETE CASCADE,
			order_item_id INTEGER NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
			quantity INTEGER NOT NULL,
			reason VARCHAR(50),
			condition VARCHAR(50),
			refund_amount NUMERIC(10, 2) DEFAULT 0.0,
			is_restockable BOOLEAN DEFAULT TRUE,
			inspection_note TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_return_item_request ON return_items(return_request_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create return_items table: %v", err)
	}

	// Live Chat Sessions
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS live_chat_sessions (
			id SERIAL PRIMARY KEY,
			session_id VARCHAR(100) NOT NULL UNIQUE,
			customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
			agent_id INTEGER,
			status VARCHAR(20) DEFAULT 'waiting',
			started_at TIMESTAMP NOT NULL,
			ended_at TIMESTAMP,
			duration INTEGER DEFAULT 0,
			message_count INTEGER DEFAULT 0,
			rating INTEGER,
			feedback TEXT,
			ip_address VARCHAR(45),
			user_agent VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_chat_session_id ON live_chat_sessions(session_id);
		CREATE INDEX IF NOT EXISTS idx_chat_customer ON live_chat_sessions(customer_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create live_chat_sessions table: %v", err)
	}

	// Chat Messages
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS chat_messages (
			id SERIAL PRIMARY KEY,
			session_id INTEGER NOT NULL REFERENCES live_chat_sessions(id) ON DELETE CASCADE,
			sender_type VARCHAR(20) NOT NULL,
			sender_id INTEGER,
			message TEXT NOT NULL,
			is_read BOOLEAN DEFAULT FALSE,
			read_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_chat_message_session ON chat_messages(session_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create chat_messages table: %v", err)
	}

	// FAQs
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS faqs (
			id SERIAL PRIMARY KEY,
			question TEXT NOT NULL,
			answer TEXT NOT NULL,
			category VARCHAR(100),
			is_public BOOLEAN DEFAULT TRUE,
			view_count INTEGER DEFAULT 0,
			helpful_yes INTEGER DEFAULT 0,
			helpful_no INTEGER DEFAULT 0,
			sort_order INTEGER DEFAULT 0,
			tags TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_faq_category ON faqs(category);
		CREATE INDEX IF NOT EXISTS idx_faq_public ON faqs(is_public);
	`))
	if err != nil {
		log.Fatalf("Failed to create faqs table: %v", err)
	}

	// Attachments
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS attachments (
			id SERIAL PRIMARY KEY,
			entity_type VARCHAR(50) NOT NULL,
			entity_id INTEGER NOT NULL,
			file_name VARCHAR(255) NOT NULL,
			file_url VARCHAR(500) NOT NULL,
			file_size INTEGER DEFAULT 0,
			mime_type VARCHAR(100),
			uploaded_by INTEGER,
			uploader_type VARCHAR(20) DEFAULT 'customer',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_attachment_entity ON attachments(entity_type, entity_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create attachments table: %v", err)
	}

	// Status Changes
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS status_changes (
			id SERIAL PRIMARY KEY,
			entity_type VARCHAR(50) NOT NULL,
			entity_id INTEGER NOT NULL,
			from_status VARCHAR(20),
			to_status VARCHAR(20) NOT NULL,
			changed_by INTEGER,
			changer_type VARCHAR(20) DEFAULT 'system',
			note TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_status_change_entity ON status_changes(entity_type, entity_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create status_changes table: %v", err)
	}

	// --- ENGAGEMENT APP ---

	// Recently Viewed
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS recently_viewed (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			viewed_at TIMESTAMP NOT NULL,
			view_count INTEGER DEFAULT 1,
			session_id VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(customer_id, product_id)
		);
		CREATE INDEX IF NOT EXISTS idx_recently_viewed_customer ON recently_viewed(customer_id, viewed_at);
		CREATE INDEX IF NOT EXISTS idx_recently_viewed_product ON recently_viewed(product_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create recently_viewed table: %v", err)
	}

	// Product Comparisons
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS product_comparisons (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			name VARCHAR(200),
			is_public BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_product_comparison_customer ON product_comparisons(customer_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create product_comparisons table: %v", err)
	}

	// Notifications
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS notifications (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			title VARCHAR(200) NOT NULL,
			message TEXT NOT NULL,
			type VARCHAR(50) DEFAULT 'info',
			priority VARCHAR(20) DEFAULT 'normal',
			is_read BOOLEAN DEFAULT FALSE,
			read_at TIMESTAMP,
			action_url VARCHAR(500),
			action_label VARCHAR(100),
			related_type VARCHAR(50),
			related_id INTEGER,
			expires_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_notification_customer ON notifications(customer_id, is_read);
		CREATE INDEX IF NOT EXISTS idx_notification_type ON notifications(type);
	`))
	if err != nil {
		log.Fatalf("Failed to create notifications table: %v", err)
	}

	// Customer Activities
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS customer_activities (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
			activity_type VARCHAR(50) NOT NULL,
			description TEXT,
			entity_type VARCHAR(50),
			entity_id INTEGER,
			ip_address VARCHAR(45),
			user_agent VARCHAR(500),
			session_id VARCHAR(100),
			metadata TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_customer_activity_customer ON customer_activities(customer_id, created_at);
		CREATE INDEX IF NOT EXISTS idx_customer_activity_type ON customer_activities(activity_type);
	`))
	if err != nil {
		log.Fatalf("Failed to create customer_activities table: %v", err)
	}

	// Abandoned Cart Reminders
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS abandoned_cart_reminders (
			id SERIAL PRIMARY KEY,
			cart_id INTEGER NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			reminder_type VARCHAR(50) DEFAULT 'email',
			sent_at TIMESTAMP NOT NULL,
			status VARCHAR(20) DEFAULT 'sent',
			email_address VARCHAR(255),
			converted BOOLEAN DEFAULT FALSE,
			converted_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_abandoned_cart_reminder_cart ON abandoned_cart_reminders(cart_id);
		CREATE INDEX IF NOT EXISTS idx_abandoned_cart_reminder_customer ON abandoned_cart_reminders(customer_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create abandoned_cart_reminders table: %v", err)
	}

	// User Segments
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS user_segments (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			description TEXT,
			conditions TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			is_dynamic BOOLEAN DEFAULT TRUE,
			priority INTEGER DEFAULT 0,
			member_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_user_segment_active ON user_segments(is_active);
	`))
	if err != nil {
		log.Fatalf("Failed to create user_segments table: %v", err)
	}

	// Segment Rules
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS segment_rules (
			id SERIAL PRIMARY KEY,
			segment_id INTEGER NOT NULL REFERENCES user_segments(id) ON DELETE CASCADE,
			field VARCHAR(100) NOT NULL,
			operator VARCHAR(20) NOT NULL,
			value VARCHAR(500) NOT NULL,
			logic_type VARCHAR(10) DEFAULT 'AND',
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_segment_rule_segment ON segment_rules(segment_id);
	`))
	if err != nil {
		log.Fatalf("Failed to create segment_rules table: %v", err)
	}

	// Shipping Methods
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS shipping_methods (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			code VARCHAR(50) NOT NULL UNIQUE,
			description TEXT,
			base_price NUMERIC(10, 2) DEFAULT 0.0,
			price_per_kg NUMERIC(10, 2) DEFAULT 0.0,
			estimated_days_min INTEGER DEFAULT 1,
			estimated_days_max INTEGER DEFAULT 7,
			is_active BOOLEAN DEFAULT TRUE,
			sort_order INTEGER DEFAULT 0,
			carrier_name VARCHAR(200),
			tracking_url_format VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_method_code ON shipping_methods(code);
		CREATE INDEX IF NOT EXISTS idx_shipping_method_active ON shipping_methods(is_active);
	`))
	if err != nil {
		log.Fatalf("Failed to create shipping_methods table: %v", err)
	}

	// Payment Methods
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS payment_methods (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			code VARCHAR(50) NOT NULL UNIQUE,
			description TEXT,
			processor_name VARCHAR(100),
			is_active BOOLEAN DEFAULT TRUE,
			sort_order INTEGER DEFAULT 0,
			requires_auth BOOLEAN DEFAULT FALSE,
			supports_refund BOOLEAN DEFAULT TRUE,
			icon_url VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_method_code ON payment_methods(code);
		CREATE INDEX IF NOT EXISTS idx_payment_method_active ON payment_methods(is_active);
	`))
	if err != nil {
		log.Fatalf("Failed to create payment_methods table: %v", err)
	}

	// Tax Rates
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS tax_rates (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			code VARCHAR(50) NOT NULL UNIQUE,
			rate NUMERIC(10, 4) NOT NULL,
			country VARCHAR(2),
			state VARCHAR(100),
			city VARCHAR(100),
			zip_code VARCHAR(20),
			is_compound BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			priority INTEGER DEFAULT 0,
			apply_to_shipping BOOLEAN DEFAULT FALSE,
			included_in_price BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_tax_rate_code ON tax_rates(code);
		CREATE INDEX IF NOT EXISTS idx_tax_rate_location ON tax_rates(country, state, city);
		CREATE INDEX IF NOT EXISTS idx_tax_rate_active ON tax_rates(is_active);
	`))
	if err != nil {
		log.Fatalf("Failed to create tax_rates table: %v", err)
	}

	// Currencies
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS currencies (
			id SERIAL PRIMARY KEY,
			code VARCHAR(3) NOT NULL UNIQUE,
			name VARCHAR(100) NOT NULL,
			symbol VARCHAR(10),
			decimal_places INTEGER DEFAULT 2,
			is_active BOOLEAN DEFAULT TRUE,
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_currency_code ON currencies(code);
		CREATE INDEX IF NOT EXISTS idx_currency_active ON currencies(is_active);
		CREATE INDEX IF NOT EXISTS idx_currency_default ON currencies(is_default);
	`))
	if err != nil {
		log.Fatalf("Failed to create currencies table: %v", err)
	}

	// Exchange Rates
	_, err = database.ExecContext(ctx, adaptDDL(database.Driver, `
		CREATE TABLE IF NOT EXISTS exchange_rates (
			id SERIAL PRIMARY KEY,
			from_currency_id INTEGER NOT NULL REFERENCES currencies(id) ON DELETE CASCADE,
			to_currency_id INTEGER NOT NULL REFERENCES currencies(id) ON DELETE CASCADE,
			rate NUMERIC(12, 6) NOT NULL,
			effective_date DATE NOT NULL,
			source VARCHAR(100),
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_exchange_rate_currencies ON exchange_rates(from_currency_id, to_currency_id, effective_date);
		CREATE INDEX IF NOT EXISTS idx_exchange_rate_active ON exchange_rates(is_active);
	`))
	if err != nil {
		log.Fatalf("Failed to create exchange_rates table: %v", err)
	}

	log.Println("✅ Database schema setup complete")
}
