# Forge Ecommerce - Complete Example

A full-featured, production-grade ecommerce system built with Forge framework, demonstrating all framework capabilities.

## Overview

This is a comprehensive ecommerce platform showcasing:

- **Complex Schema System** - Multiple models with intricate relationships
- **Advanced ORM** - Complex queries, filters, aggregations
- **Complete Admin Interface** - Full CRUD with custom actions and filters
- **REST API** - Full API for frontend applications
- **Filter System** - Advanced filtering across all models
- **Authentication & Authorization** - User management, permissions
- **Migrations** - Complete database schema management
- **Business Logic** - Inventory management, order processing, pricing

## Features

### Catalog Management
- **Categories** - Hierarchical product categories with nested subcategories
- **Products** - Full product information with multiple variants
- **Product Variants** - SKU-based variants (size, color, etc.)
- **Product Images** - Multiple images per product
- **Attributes** - Dynamic product attributes

### Customer Management
- **Customers** - Customer profiles with authentication
- **Addresses** - Multiple shipping/billing addresses
- **Wish Lists** - Customer wish lists
- **Customer Groups** - Segmentation for pricing/promotions

### Order Management
- **Shopping Cart** - Persistent shopping carts
- **Orders** - Complete order lifecycle
- **Order Items** - Line items with pricing history
- **Payments** - Payment tracking and processing
- **Shipments** - Shipping and fulfillment tracking
- **Order Status** - Complete status workflow

### Inventory Management
- **Warehouses** - Multiple warehouse support
- **Stock** - Real-time inventory tracking
- **Stock Movements** - Full inventory audit trail
- **Stock Alerts** - Low stock notifications

### Marketing & Promotion
- **Coupons** - Discount codes and promotions
- **Reviews** - Product reviews and ratings
- **Ratings** - Star ratings with aggregation

## Project Structure

```
ecommerce/
├── main.go                     # Application entry point
├── go.mod                      # Go module file
├── config/
│   └── config.yaml            # Configuration
├── app/
│   ├── catalog/               # Catalog management
│   │   ├── models.go          # Product, Category, Variant models
│   │   ├── admin.go           # Admin configuration
│   │   ├── api.go             # API viewsets
│   │   └── filters.go         # Custom filters
│   ├── customers/             # Customer management
│   │   ├── models.go          # Customer, Address models
│   │   ├── admin.go
│   │   ├── api.go
│   │   └── filters.go
│   ├── orders/                # Order management
│   │   ├── models.go          # Order, OrderItem, Payment models
│   │   ├── admin.go
│   │   ├── api.go
│   │   ├── filters.go
│   │   └── services.go        # Business logic
│   ├── inventory/             # Inventory management
│   │   ├── models.go          # Stock, Warehouse models
│   │   ├── admin.go
│   │   ├── api.go
│   │   └── services.go
│   └── marketing/             # Marketing features
│       ├── models.go          # Coupon, Review models
│       ├── admin.go
│       ├── api.go
│       └── filters.go
├── migrations/                 # Database migrations
├── static/                     # Static files
└── templates/                  # HTML templates

```

## Models Overview

### Catalog (7 models)
1. **Category** - Product categories with parent/child hierarchy
2. **Product** - Main product model
3. **ProductVariant** - Product variations (SKU level)
4. **ProductImage** - Product images with ordering
5. **ProductAttribute** - Dynamic attributes
6. **AttributeValue** - Attribute values
7. **Brand** - Product brands

### Customers (4 models)
1. **Customer** - Customer profiles
2. **Address** - Shipping/billing addresses
3. **WishList** - Customer wish lists
4. **CustomerGroup** - Customer segmentation

### Orders (5 models)
1. **Cart** - Shopping cart
2. **CartItem** - Cart line items
3. **Order** - Order master
4. **OrderItem** - Order line items
5. **Payment** - Payment transactions
6. **Shipment** - Shipping information

### Inventory (4 models)
1. **Warehouse** - Warehouse locations
2. **Stock** - Inventory levels
3. **StockMovement** - Inventory transactions
4. **StockAlert** - Low stock alerts

### Marketing (3 models)
1. **Coupon** - Discount codes
2. **Review** - Product reviews
3. **Rating** - Product ratings

**Total: 23 Models with complex relationships**

## Database Schema

### Key Relationships

```
Category (self-referential)
    ├── Products (one-to-many)
    
Product
    ├── Category (foreign key)
    ├── Brand (foreign key)
    ├── ProductVariants (one-to-many)
    ├── ProductImages (one-to-many)
    ├── Reviews (one-to-many)
    └── Attributes (many-to-many)

ProductVariant
    ├── Product (foreign key)
    └── Stock (one-to-many)

Customer
    ├── Addresses (one-to-many)
    ├── WishLists (one-to-many)
    ├── CustomerGroup (foreign key)
    ├── Orders (one-to-many)
    └── Reviews (one-to-many)

Order
    ├── Customer (foreign key)
    ├── OrderItems (one-to-many)
    ├── Payments (one-to-many)
    ├── Shipment (one-to-one)
    └── Coupon (foreign key)

OrderItem
    ├── Order (foreign key)
    └── ProductVariant (foreign key)

Stock
    ├── ProductVariant (foreign key)
    ├── Warehouse (foreign key)
    └── StockMovements (one-to-many)
```

## Quick Start

### 1. Create the Project

This example is already created, but to create a similar project:

```bash
# Navigate to examples directory
cd examples

# The ecommerce project is ready to use
cd ecommerce
```

### 2. Configure Database

Edit `config/config.yaml`:

```yaml
database:
  driver: postgres
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: ecommerce_db
  sslmode: disable
```

Create the database:

```sql
CREATE DATABASE ecommerce_db;
```

### 3. Generate Code

Generate type-safe code from models:

```bash
forge generate
```

This generates:
- Model structs with all fields
- Type-safe field expressions
- Managers with CRUD operations
- QuerySets for filtering and queries

### 4. Create Migrations

Generate migrations from models:

```bash
forge makemigrations
```

Apply migrations:

```bash
forge migrate
```

### 5. Create Superuser

Create an admin user:

```bash
forge createsuperuser
```

### 6. Run Server

Start the development server:

```bash
forge runserver
```

Or directly:

```bash
go run main.go
```

### 7. Access Interfaces

- **Admin Interface**: http://localhost:8000/admin/
- **REST API**: http://localhost:8000/api/v1/
- **API Docs**: http://localhost:8000/api/v1/docs/

## CLI Commands Demonstrated

### Project Management
```bash
# Create new project
forge new ecommerce --template=simple --database=postgres

# Add new app
forge add app marketing

# Add new model
forge add model Product --app=catalog
```

### Code Generation
```bash
# Generate all code
forge generate

# Generate for specific app
forge generate --models=./app/catalog

# Preview generation
forge generate --dry-run
```

### Migration Management
```bash
# Create new migration
forge makemigrations

# Apply migrations
forge migrate

# Check migration status
forge migrate status

# Rollback migration
forge migrate rollback

# Show migration SQL
forge migrate show 001_initial

# Lint migrations
forge migrate lint
```

### Development
```bash
# Run development server
forge runserver

# Run with custom host/port
forge runserver --host=0.0.0.0 --port=9000

# Run tests
forge test

# Check project
forge check

# Interactive shell
forge shell
```

### Admin
```bash
# Create superuser
forge createsuperuser

# Create regular user
forge createsuperuser --no-superuser
```

## API Examples

### Products API

```bash
# List products
curl http://localhost:8000/api/v1/products/

# Filter products
curl "http://localhost:8000/api/v1/products/?category__name=Electronics&price__gte=100"

# Get product
curl http://localhost:8000/api/v1/products/1/

# Create product
curl -X POST http://localhost:8000/api/v1/products/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop",
    "description": "High-performance laptop",
    "price": 999.99,
    "category_id": 1
  }'

# Update product
curl -X PATCH http://localhost:8000/api/v1/products/1/ \
  -H "Content-Type: application/json" \
  -d '{"price": 899.99}'

# Delete product
curl -X DELETE http://localhost:8000/api/v1/products/1/
```

### Orders API

```bash
# List orders
curl http://localhost:8000/api/v1/orders/

# Filter by status
curl "http://localhost:8000/api/v1/orders/?status=pending"

# Filter by customer
curl "http://localhost:8000/api/v1/orders/?customer__email=john@example.com"

# Create order
curl -X POST http://localhost:8000/api/v1/orders/ \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "items": [
      {"product_variant_id": 1, "quantity": 2, "price": 99.99}
    ]
  }'
```

### Advanced Filtering

```bash
# Complex filters with deep relations
curl "http://localhost:8000/api/v1/products/?category__parent__name=Electronics&reviews__rating__gte=4"

# Multiple conditions
curl "http://localhost:8000/api/v1/orders/?status__in=pending,processing&created_at__gte=2024-01-01&total__gte=100"

# Search
curl "http://localhost:8000/api/v1/products/?search=laptop"

# Ordering
curl "http://localhost:8000/api/v1/products/?ordering=-created_at,price"
```

## ORM Usage Examples

### Simple Queries

```go
import (
    "context"
    "examples/ecommerce/app/catalog"
)

ctx := context.Background()

// Get all products
products, err := catalog.Product.Objects.All(ctx)

// Get single product
product, err := catalog.Product.Objects.Get(ctx, 1)

// Filter products
products, err := catalog.Product.Objects.
    Filter(catalog.Product.Fields.Price.GreaterThan(100)).
    OrderBy("-created_at").
    All(ctx)

// Count products
count, err := catalog.Product.Objects.
    Filter(catalog.Product.Fields.IsActive.Equals(true)).
    Count(ctx)
```

### Complex Queries

```go
// Filter with multiple conditions
products, err := catalog.Product.Objects.
    Filter(
        catalog.Product.Fields.Price.Between(50, 200).
            And(catalog.Product.Fields.IsActive.Equals(true)).
            And(catalog.Product.Fields.Stock.GreaterThan(0)),
    ).
    OrderBy("-created_at").
    Limit(10).
    All(ctx)

// Deep relation filtering
products, err := catalog.Product.Objects.
    Filter(
        catalog.Product.Fields.Category.Name.Equals("Electronics").
            And(catalog.Product.Fields.Reviews.Rating.GreaterThanOrEqual(4)),
    ).
    All(ctx)

// Aggregations
avgPrice, err := catalog.Product.Objects.
    Filter(catalog.Product.Fields.Category.ID.Equals(1)).
    Aggregate(catalog.Product.Fields.Price.Avg())
```

### Create, Update, Delete

```go
// Create
product := &catalog.Product{
    Name:        "New Product",
    Description: "Product description",
    Price:       99.99,
    CategoryID:  1,
    IsActive:    true,
}
err := catalog.Product.Objects.Create(ctx, product)

// Update
product.Price = 89.99
err := catalog.Product.Objects.Update(ctx, product)

// Delete
err := catalog.Product.Objects.Delete(ctx, product)

// Bulk operations
err := catalog.Product.Objects.
    Filter(catalog.Product.Fields.IsActive.Equals(false)).
    Delete(ctx)
```

## Filter System Examples

### Declarative Filters

```go
// Product filter
type ProductFilter struct {
    *filter.FilterSet[catalog.Product]
    Name     *filters.CharFilter[catalog.Product]
    Price    *filters.NumberFilter[catalog.Product]
    Category *filters.RelatedFilter[catalog.Product, catalog.Category]
}

func NewProductFilter() *ProductFilter {
    fs, _ := filter.NewFilterSet[catalog.Product]()
    return &ProductFilter{
        FilterSet: fs,
        Name:      filters.NewCharFilter[catalog.Product]("name").IContains(),
        Price:     filters.NewNumberFilter[catalog.Product]("price").Range(),
        Category:  filters.NewRelatedFilter[catalog.Product, catalog.Category]("category"),
    }
}
```

### Using Filters

```go
// Create filter
pf := NewProductFilter()

// Apply filters
pf.Where("name").Contains("laptop")
pf.Where("price").Between(500, 2000)
pf.Where("category__name").Equals("Electronics")

// Get results
products, err := pf.ApplyAST(ctx, pf.GetAST())
```

## Admin Interface

The admin interface provides:

### List Views
- Sortable columns
- Pagination
- Bulk actions
- Search
- Filters sidebar
- Export to CSV/JSON

### Detail Views
- View all fields
- Related objects
- Audit trail
- Action buttons

### Form Views
- Validation
- Rich widgets
- File uploads
- Related object selection
- Inline editing

### Custom Actions
- Bulk status updates
- Export operations
- Custom business logic

## Business Logic Examples

### Order Processing

```go
// app/orders/services.go
func ProcessOrder(ctx context.Context, orderID int64) error {
    // Get order
    order, err := orders.Order.Objects.Get(ctx, orderID)
    if err != nil {
        return err
    }
    
    // Validate stock
    for _, item := range order.Items {
        stock, err := inventory.Stock.Objects.
            Filter(inventory.Stock.Fields.ProductVariantID.Equals(item.ProductVariantID)).
            First(ctx)
        if err != nil {
            return err
        }
        if stock.Quantity < item.Quantity {
            return errors.New("insufficient stock")
        }
    }
    
    // Process payment
    payment := &orders.Payment{
        OrderID: order.ID,
        Amount:  order.Total,
        Status:  "pending",
    }
    err = orders.Payment.Objects.Create(ctx, payment)
    if err != nil {
        return err
    }
    
    // Update stock
    for _, item := range order.Items {
        err = inventory.UpdateStock(ctx, item.ProductVariantID, -item.Quantity, "order")
        if err != nil {
            return err
        }
    }
    
    // Update order status
    order.Status = "processing"
    err = orders.Order.Objects.Update(ctx, order)
    
    return err
}
```

### Inventory Management

```go
// app/inventory/services.go
func UpdateStock(ctx context.Context, variantID int64, quantity int, reason string) error {
    stock, err := inventory.Stock.Objects.
        Filter(inventory.Stock.Fields.ProductVariantID.Equals(variantID)).
        First(ctx)
    if err != nil {
        return err
    }
    
    // Create movement record
    movement := &inventory.StockMovement{
        StockID:   stock.ID,
        Quantity:  quantity,
        Type:      reason,
        CreatedAt: time.Now(),
    }
    err = inventory.StockMovement.Objects.Create(ctx, movement)
    if err != nil {
        return err
    }
    
    // Update stock
    stock.Quantity += quantity
    err = inventory.Stock.Objects.Update(ctx, stock)
    if err != nil {
        return err
    }
    
    // Check for low stock alert
    if stock.Quantity <= stock.LowStockThreshold {
        alert := &inventory.StockAlert{
            StockID:   stock.ID,
            Message:   "Low stock alert",
            Threshold: stock.LowStockThreshold,
            CreatedAt: time.Now(),
        }
        inventory.StockAlert.Objects.Create(ctx, alert)
    }
    
    return nil
}
```

## Testing

```bash
# Run all tests
forge test

# Run specific package
forge test ./app/orders

# Run with coverage
forge test -cover

# Run with verbose output
forge test -v
```

## Deployment

### Docker

```bash
# Build image
docker build -t ecommerce:latest .

# Run with docker-compose
docker-compose up -d
```

### Production Checklist

- [ ] Change secret keys in config
- [ ] Set up production database
- [ ] Configure CORS for API
- [ ] Set up SSL/TLS
- [ ] Configure logging
- [ ] Set up monitoring
- [ ] Configure backups
- [ ] Set up rate limiting
- [ ] Review security settings

## Architecture Highlights

### Modular Design
- Each app is independent and reusable
- Clear separation of concerns
- Easy to extend and maintain

### Type Safety
- Compile-time type checking
- IDE autocomplete support
- No runtime reflection overhead

### Performance
- Connection pooling
- Query optimization
- Efficient JOIN generation
- Caching strategies

### Security
- SQL injection prevention
- CSRF protection
- XSS prevention
- Input validation
- Permission system

## Learn More

- **Framework Docs**: `../../docs/`
- **API Reference**: `../../docs/API_REFERENCE.md`
- **Schema Guide**: `../../docs/SCHEMA_REFERENCE.md`
- **ORM Guide**: `../../docs/orm/README.md`
- **Admin Guide**: `../../docs/pkg_admin_README.md`

## License

MIT
