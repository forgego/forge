# Ecommerce Application - Comprehensive Testing Guide

This guide provides step-by-step instructions for manually testing all features of the Forge Ecommerce application.

## Table of Contents
1. [Setup](#setup)
2. [Database Testing](#database-testing)
3. [Admin Interface Testing](#admin-interface-testing)
4. [API Testing](#api-testing)
5. [Feature Testing](#feature-testing)
6. [UI/UX Testing](#uiux-testing)
7. [Performance Testing](#performance-testing)

## Setup

### Prerequisites
- Go 1.21+ installed
- PostgreSQL 13+ (or SQLite for quick testing)
- Git
- curl or Postman for API testing
- Web browser for admin UI testing

### Installation Steps

```bash
# Navigate to the ecommerce directory
cd examples/ecommerce

# Install dependencies
go mod download

# Configure database
cp config/config.yaml config/config.local.yaml
# Edit config.local.yaml with your database credentials

# Generate code from models
go generate

# Create and apply migrations
forge makemigrations
forge migrate

# Create a superuser
forge createsuperuser
# Username: admin
# Email: admin@example.com
# Password: admin123

# Start the server
forge runserver
# Or: go run main.go
```

The application should now be running at:
- Homepage: http://localhost:8000
- Admin: http://localhost:8000/admin/
- API: http://localhost:8000/api/v1/

## Database Testing

### 1. Verify Database Schema

```sql
-- Connect to your database and verify tables exist
\dt

-- Expected tables:
-- Core modules: categories, products, product_variants, product_images, brands
-- Customer: customers, addresses, wish_lists, customer_groups
-- Orders: carts, cart_items, orders, order_items, payments, shipments
-- Inventory: warehouses, stock, stock_movements, stock_alerts
-- Marketing: coupons, reviews, ratings
-- Commerce: shipping_methods, payment_methods, tax_rates, currencies, exchange_rates
-- Promotions: promotions, promotion_rules, banners, newsletter_subscriptions, promotion_usages
-- Engagement: recently_viewed, product_comparisons, notifications, customer_activities, 
--             abandoned_cart_reminders, user_segments, segment_rules
-- Support: support_tickets, support_messages, return_requests, live_chat_sessions, faqs,
--          attachments, return_items, status_changes, chat_messages
```

### 2. Verify Indexes

```sql
-- Check that indexes are created correctly
SELECT tablename, indexname, indexdef 
FROM pg_indexes 
WHERE schemaname = 'public' 
ORDER BY tablename, indexname;
```

### 3. Test Foreign Key Constraints

```sql
-- Try inserting data that violates foreign keys (should fail)
INSERT INTO product_variants (product_id, sku) VALUES (99999, 'TEST-SKU');
-- Expected: Foreign key constraint violation
```

## Admin Interface Testing

### 1. Login and Dashboard

✅ **Test Steps:**
1. Navigate to http://localhost:8000/admin/
2. Login with superuser credentials
3. Verify dashboard loads correctly
4. Check navigation menu shows all modules

**Expected Results:**
- Dashboard displays with no errors
- All module groups visible: Catalog, Commerce, Customers, Engagement, Inventory, Marketing, Orders, Promotions, Support
- Recent actions panel shows activity
- Quick stats display (if configured)

### 2. Catalog Module

#### Categories
```
Test Case: Create Hierarchical Categories
1. Click "Categories" in admin menu
2. Click "Add Category"
3. Fill in:
   - Name: "Electronics"
   - Slug: "electronics"
   - Description: "Electronic products"
   - Is Active: Yes
4. Save
5. Create a child category:
   - Name: "Laptops"
   - Parent: Electronics
   - Save
Expected: Both categories created, hierarchy displayed
```

#### Products
```
Test Case: Create Product with Variants
1. Click "Products" in admin menu
2. Click "Add Product"
3. Fill in:
   - Name: "MacBook Pro 16"
   - SKU: "APPLE-MBP16"
   - Category: Laptops (from above)
   - Price: 2499.99
   - Is Active: Yes
4. Add inline variant:
   - SKU: "APPLE-MBP16-SILVER"
   - Name: "Silver, 16GB RAM"
   - Price: 2499.99
5. Add inline image:
   - Image URL: "https://example.com/macbook.jpg"
   - Alt Text: "MacBook Pro"
6. Save
Expected: Product created with variant and image
```

#### Bulk Actions
```
Test Case: Bulk Activate Products
1. Go to Products list
2. Select multiple products
3. Choose "Activate Products" from action dropdown
4. Click "Go"
Expected: Products are activated, success message shown
```

### 3. Commerce Module

#### Shipping Methods
```
Test Case: Create Shipping Method
1. Click "Shipping Methods"
2. Add new:
   - Name: "Standard Shipping"
   - Code: "standard"
   - Base Price: 5.99
   - Carrier: "USPS"
   - Est. Days Min: 3
   - Est. Days Max: 7
   - Is Active: Yes
3. Save
Expected: Shipping method created successfully
```

#### Payment Methods
```
Test Case: Configure Payment Methods
1. Click "Payment Methods"
2. Add:
   - Name: "Credit Card"
   - Code: "credit_card"
   - Processor: "Stripe"
   - Supports Refund: Yes
   - Requires Auth: Yes
   - Is Active: Yes
3. Save
Expected: Payment method available for orders
```

#### Tax Rates
```
Test Case: Create Tax Rate
1. Click "Tax Rates"
2. Add:
   - Name: "California Sales Tax"
   - Code: "CA_SALES_TAX"
   - Rate: 0.0725 (7.25%)
   - Country: US
   - State: CA
   - Is Active: Yes
3. Save
Expected: Tax rate created and will be applied to CA orders
```

### 4. Customers Module

```
Test Case: Create Customer with Address
1. Click "Customers"
2. Add new customer:
   - Email: "john@example.com"
   - First Name: "John"
   - Last Name: "Doe"
   - Is Active: Yes
3. Add inline address:
   - Type: Shipping
   - Street: "123 Main St"
   - City: "San Francisco"
   - State: "CA"
   - Zip: "94102"
   - Country: "US"
4. Save
Expected: Customer created with address
```

### 5. Orders Module

```
Test Case: Create Order
1. Click "Orders"
2. Add new order:
   - Customer: John Doe
   - Status: Pending
3. Add inline order items:
   - Product Variant: MacBook Pro Silver
   - Quantity: 1
   - Unit Price: 2499.99
4. Save
Expected: Order created with items, total calculated
```

### 6. Promotions Module

```
Test Case: Create Promotion Code
1. Click "Promotions"
2. Add:
   - Name: "Summer Sale 2024"
   - Code: "SUMMER20"
   - Discount Type: Percentage
   - Discount Value: 20
   - Start Date: Today
   - End Date: +30 days
   - Is Active: Yes
   - Applies To: All
3. Save
Expected: Promotion code ready for use
```

```
Test Case: Create Banner
1. Click "Banners"
2. Add:
   - Title: "Summer Sale"
   - Image URL: "https://example.com/banner.jpg"
   - Link URL: "/sale"
   - Placement: home_hero
   - Is Active: Yes
   - Start Date: Today
3. Save
Expected: Banner created and will display on homepage
```

### 7. Engagement Module

```
Test Case: Create User Segment
1. Click "User Segments"
2. Add:
   - Name: "High Value Customers"
   - Description: "Customers with total spend > $1000"
   - Is Active: Yes
   - Is Dynamic: Yes
3. Save
4. Add Segment Rules:
   - Field: "total_spent"
   - Operator: "greater_than"
   - Value: "1000"
5. Save
Expected: Segment created and will automatically update membership
```

```
Test Case: Send Notification
1. Click "Notifications"
2. Add:
   - Customer: John Doe
   - Title: "Your order has shipped!"
   - Message: "Track your package here"
   - Type: shipping
   - Priority: high
3. Save
Expected: Notification created for customer
```

### 8. Support Module

```
Test Case: Create Support Ticket
1. Click "Support Tickets"
2. Add:
   - Customer: John Doe
   - Subject: "Product defect"
   - Description: "Laptop screen flickering"
   - Status: Open
   - Priority: High
   - Category: product_inquiry
3. Save
4. Add Support Message:
   - Ticket: (select created ticket)
   - Sender Type: Agent
   - Message: "We're sorry to hear that. Can you provide more details?"
5. Save
Expected: Ticket created with initial response
```

```
Test Case: Create Return Request
1. Click "Return Requests"
2. Add:
   - Order: (select an order)
   - Customer: John Doe
   - Reason: defective
   - Description: "Screen flickering"
   - Status: Pending
3. Add Return Items:
   - Order Item: (select item)
   - Quantity: 1
   - Condition: opened
4. Save
Expected: Return request created for review
```

### 9. Inventory Module

```
Test Case: Manage Stock
1. Click "Stock"
2. Add:
   - Product Variant: MacBook Pro Silver
   - Warehouse: Main Warehouse
   - Quantity: 50
   - Reserved: 0
3. Save
4. Add Stock Movement:
   - Stock: (select created)
   - Type: purchase
   - Quantity: 50
   - Reference: PO-001
5. Save
Expected: Stock created with audit trail
```

## API Testing

### Authentication Setup
```bash
# Most endpoints will require authentication
# For testing, you can use the admin credentials or create an API token

# Example: Get auth token (if JWT is configured)
curl -X POST http://localhost:8000/api/v1/auth/login/ \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 1. Catalog API

```bash
# List all products
curl http://localhost:8000/api/v1/products/

# Get specific product
curl http://localhost:8000/api/v1/products/1/

# Filter products by category
curl "http://localhost:8000/api/v1/products/?category_id=1"

# Search products
curl "http://localhost:8000/api/v1/products/?search=macbook"

# Filter by price range
curl "http://localhost:8000/api/v1/products/?price__gte=1000&price__lte=3000"

# Order by price
curl "http://localhost:8000/api/v1/products/?ordering=-price"

# Create product
curl -X POST http://localhost:8000/api/v1/products/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPad Pro",
    "sku": "APPLE-IPAD-PRO",
    "price": 999.99,
    "category_id": 1,
    "is_active": true
  }'
```

### 2. Commerce API

```bash
# List shipping methods
curl http://localhost:8000/api/v1/shipping-methods/

# List payment methods
curl http://localhost:8000/api/v1/payment-methods/

# Get tax rates
curl http://localhost:8000/api/v1/tax-rates/

# Filter by location
curl "http://localhost:8000/api/v1/tax-rates/?country=US&state=CA"

# List currencies
curl http://localhost:8000/api/v1/currencies/

# Get exchange rates
curl http://localhost:8000/api/v1/exchange-rates/
```

### 3. Orders API

```bash
# List orders
curl http://localhost:8000/api/v1/orders/

# Filter by status
curl "http://localhost:8000/api/v1/orders/?status=pending"

# Filter by customer
curl "http://localhost:8000/api/v1/orders/?customer_id=1"

# Get order details
curl http://localhost:8000/api/v1/orders/1/

# Create order
curl -X POST http://localhost:8000/api/v1/orders/ \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "status": "pending",
    "shipping_address_id": 1,
    "billing_address_id": 1
  }'
```

### 4. Promotions API

```bash
# List active promotions
curl "http://localhost:8000/api/v1/promotions/?is_active=true"

# Validate promotion code
curl "http://localhost:8000/api/v1/promotions/?code=SUMMER20"

# List banners
curl http://localhost:8000/api/v1/banners/

# Filter by placement
curl "http://localhost:8000/api/v1/banners/?placement=home_hero&is_active=true"

# Newsletter subscriptions
curl -X POST http://localhost:8000/api/v1/newsletter-subscriptions/ \
  -H "Content-Type: application/json" \
  -d '{
    "email": "subscriber@example.com",
    "first_name": "Jane",
    "status": "pending"
  }'
```

### 5. Engagement API

```bash
# Track recently viewed
curl -X POST http://localhost:8000/api/v1/recently-viewed/ \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "product_id": 1
  }'

# Get customer notifications
curl "http://localhost:8000/api/v1/notifications/?customer_id=1&is_read=false"

# Mark notification as read
curl -X PATCH http://localhost:8000/api/v1/notifications/1/ \
  -H "Content-Type: application/json" \
  -d '{"is_read": true}'

# Get user segments
curl http://localhost:8000/api/v1/user-segments/
```

### 6. Support API

```bash
# List support tickets
curl "http://localhost:8000/api/v1/support-tickets/?customer_id=1"

# Create ticket
curl -X POST http://localhost:8000/api/v1/support-tickets/ \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "subject": "Need help with order",
    "description": "My order hasnt arrived",
    "priority": "normal",
    "category": "order_issue"
  }'

# Add message to ticket
curl -X POST http://localhost:8000/api/v1/support-messages/ \
  -H "Content-Type: application/json" \
  -d '{
    "ticket_id": 1,
    "sender_type": "customer",
    "message": "Its been 10 days"
  }'

# List FAQs
curl "http://localhost:8000/api/v1/faqs/?is_public=true&category=shipping"

# Create return request
curl -X POST http://localhost:8000/api/v1/return-requests/ \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": 1,
    "customer_id": 1,
    "reason": "defective",
    "description": "Product arrived damaged"
  }'
```

## Feature Testing

### 1. Complete Purchase Flow

**Steps:**
1. Browse products → Add to cart → Proceed to checkout
2. Enter shipping address
3. Select shipping method
4. Enter payment details
5. Apply coupon code
6. Review order
7. Complete purchase
8. Verify order confirmation
9. Check email notification

**Expected Results:**
- Cart updates correctly
- Prices calculated with tax
- Coupon discount applied
- Inventory decremented
- Order created with correct total
- Customer receives confirmation

### 2. Return Flow

**Steps:**
1. Customer initiates return
2. Admin reviews return request
3. Admin approves return
4. Customer ships item
5. Admin receives and inspects
6. Admin processes refund
7. Inventory restocked

**Expected Results:**
- Return request tracked through all statuses
- Status changes logged
- Refund processed
- Stock updated

### 3. Support Flow

**Steps:**
1. Customer creates support ticket
2. Agent responds
3. Ticket status updated
4. Resolution provided
5. Ticket closed
6. Customer satisfaction recorded

**Expected Results:**
- All messages threaded correctly
- Status changes logged
- First response time tracked
- Resolution documented

## UI/UX Testing

### 1. Admin Interface

**Check:**
- ✅ Navigation is intuitive
- ✅ Forms are clear and well-organized
- ✅ Validation messages are helpful
- ✅ List views are sortable and filterable
- ✅ Search works correctly
- ✅ Bulk actions function properly
- ✅ Inline forms save correctly
- ✅ Date pickers work
- ✅ File uploads function
- ✅ Responsive design works on tablet/mobile

### 2. Design Consistency

**Check:**
- ✅ Consistent color scheme
- ✅ Typography is readable
- ✅ Icons are meaningful
- ✅ Spacing is consistent
- ✅ Buttons have hover states
- ✅ Loading states are shown
- ✅ Error states are clear

### 3. Accessibility

**Check:**
- ✅ Keyboard navigation works
- ✅ Forms have proper labels
- ✅ Color contrast is sufficient
- ✅ Screen reader compatible
- ✅ Focus indicators visible

## Performance Testing

### 1. Database Query Performance

```sql
-- Enable query logging
SET log_statement = 'all';

-- Test common queries
EXPLAIN ANALYZE SELECT * FROM products WHERE is_active = true AND category_id = 1;
EXPLAIN ANALYZE SELECT * FROM orders WHERE customer_id = 1 ORDER BY created_at DESC;

-- Check for missing indexes
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE schemaname = 'public'
ORDER BY abs(correlation) DESC;
```

### 2. API Response Times

```bash
# Use curl with timing
curl -w "@curl-format.txt" -o /dev/null -s "http://localhost:8000/api/v1/products/"

# curl-format.txt content:
# time_namelookup: %{time_namelookup}\n
# time_connect: %{time_connect}\n
# time_total: %{time_total}\n
```

**Expected Response Times:**
- List endpoints: < 200ms
- Detail endpoints: < 100ms
- Create/Update: < 300ms
- Complex queries: < 500ms

### 3. Load Testing

```bash
# Install Apache Bench
# apt-get install apache2-utils

# Test with 100 concurrent requests
ab -n 1000 -c 100 http://localhost:8000/api/v1/products/

# Expected results:
# - No failed requests
# - Consistent response times
# - Server remains stable
```

## Checklist Summary

### Models & Schema
- [ ] All 40+ models created
- [ ] All fields defined with proper types
- [ ] All relations configured
- [ ] All indexes created
- [ ] All constraints working

### Admin Interface
- [ ] All models registered
- [ ] List displays configured
- [ ] Filters working
- [ ] Search functional
- [ ] Bulk actions operational
- [ ] Inline editing works
- [ ] Custom actions functional

### API Endpoints
- [ ] All ViewSets registered
- [ ] CRUD operations work
- [ ] Filtering functional
- [ ] Search works
- [ ] Ordering works
- [ ] Pagination configured
- [ ] Authentication working

### Business Logic
- [ ] Order calculation correct
- [ ] Tax calculation accurate
- [ ] Shipping cost calculated
- [ ] Discount application works
- [ ] Inventory tracking accurate
- [ ] Stock alerts triggering

### UI/UX
- [ ] Design is consistent
- [ ] Forms are intuitive
- [ ] Error messages helpful
- [ ] Loading states shown
- [ ] Responsive design works
- [ ] Accessibility standards met

### Performance
- [ ] Database queries optimized
- [ ] API responses fast
- [ ] No N+1 queries
- [ ] Indexes effective
- [ ] Caching configured

## Known Issues

Document any issues found during testing here:

1. **Issue**: [Description]
   - **Impact**: [High/Medium/Low]
   - **Steps to Reproduce**: [...]
   - **Expected**: [...]
   - **Actual**: [...]
   - **Status**: [Open/Fixed/Won't Fix]

## Next Steps

After completing this testing guide:

1. **Fix Critical Issues**: Address any critical bugs found
2. **Performance Optimization**: Optimize slow queries
3. **Security Audit**: Review authentication and authorization
4. **Documentation**: Update API documentation
5. **Deployment**: Prepare for production deployment

## Additional Resources

- [Architecture Documentation](ARCHITECTURE.md)
- [API Documentation](docs/api.md)
- [Deployment Guide](docs/deployment.md)
- [Contributing Guide](../../CONTRIBUTING.md)
