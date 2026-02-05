# Forge Ecommerce - Quick Setup Guide

Get started with the Forge Ecommerce example in 5 minutes!

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Forge CLI installed

## Quick Start (5 minutes)

### 1. Clone/Navigate to Project

```bash
cd examples/ecommerce
```

### 2. Install Dependencies

```bash
make install
# or
go mod download
```

### 3. Create Database

```bash
make db-create
# or
createdb ecommerce_db
```

### 4. Configure Database

Edit `config/config.yaml` if needed (default settings should work for local PostgreSQL):

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

### 5. Generate Code

```bash
make generate
# or
forge generate
```

### 6. Run Migrations

```bash
make migrate
# or
forge migrate
```

### 7. Create Superuser

```bash
make superuser
# or
forge createsuperuser
```

### 8. Start Server

```bash
make run
# or
forge runserver
```

### 9. Access the Application

- **Homepage**: http://localhost:8000/
- **Admin Interface**: http://localhost:8000/admin/
- **REST API**: http://localhost:8000/api/v1/

## Docker Setup (Alternative)

### Using Docker Compose

```bash
# Start all services (app + database)
docker-compose up -d

# Check logs
docker-compose logs -f app

# Run migrations in container
docker-compose exec app forge migrate

# Create superuser in container
docker-compose exec app forge createsuperuser

# Stop services
docker-compose down
```

## Project Structure

```
ecommerce/
├── main.go                     # Application entry point
├── go.mod                      # Go module
├── Makefile                    # Build automation
├── Dockerfile                  # Docker image
├── docker-compose.yml          # Docker orchestration
├── config/
│   └── config.yaml            # Configuration
├── app/                       # Application code
│   ├── catalog/              # Product catalog
│   │   ├── models.go         # Model definitions
│   │   ├── admin.go          # Admin config
│   │   ├── api.go            # API endpoints
│   │   └── *_gen.go          # Generated code
│   ├── customers/            # Customer management
│   ├── orders/               # Order processing
│   ├── inventory/            # Inventory tracking
│   └── marketing/            # Marketing features
├── migrations/               # Database migrations
├── static/                   # Static files
└── templates/                # HTML templates
```

## Key Features to Explore

### 1. Admin Interface (http://localhost:8000/admin/)

**Login** with the superuser account you created.

**Explore:**
- 📦 **Catalog** - Products, Categories, Brands, Variants
- 👥 **Customers** - Customer management, Addresses, Wish Lists
- 📋 **Orders** - Order processing, Payments, Shipments
- 📊 **Inventory** - Stock levels, Warehouses, Transfers, Alerts
- ⭐ **Marketing** - Reviews, Coupons, Questions

**Try:**
- Create a category
- Add a product with variants
- Create a customer
- Place a test order
- Add a review

### 2. REST API (http://localhost:8000/api/v1/)

**Available endpoints:**
- `/api/v1/products/` - Product catalog
- `/api/v1/categories/` - Product categories
- `/api/v1/customers/` - Customer data
- `/api/v1/orders/` - Order information
- `/api/v1/reviews/` - Product reviews
- `/api/v1/coupons/` - Discount coupons

**Example API calls:**

```bash
# List products
curl http://localhost:8000/api/v1/products/

# Filter products
curl "http://localhost:8000/api/v1/products/?category__name=Electronics&price__gte=100"

# Get specific product
curl http://localhost:8000/api/v1/products/1/

# Create product (requires authentication)
curl -X POST http://localhost:8000/api/v1/products/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop",
    "sku": "LAP001",
    "price": 999.99,
    "category_id": 1
  }'
```

### 3. Complex Filtering

```bash
# Products in Electronics category with 4+ rating
curl "http://localhost:8000/api/v1/products/?category__name=Electronics&rating_average__gte=4"

# Orders from specific customer in last 30 days
curl "http://localhost:8000/api/v1/orders/?customer__email=john@example.com&created_at__gte=2024-01-01"

# Low stock alerts
curl "http://localhost:8000/api/v1/stock-alerts/?status=active&alert_type=low_stock"
```

## Development Workflow

### Making Changes

```bash
# 1. Edit models
vim app/catalog/models.go

# 2. Regenerate code
make generate

# 3. Create migration
forge makemigrations

# 4. Apply migration
make migrate

# 5. Restart server
make run
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific tests
go test ./app/catalog/...
```

### Code Quality

```bash
# Check code
make check

# Format code
make fmt

# Lint code
make lint

# Run all checks
make ci
```

## Sample Data

### Seed Database (Optional)

Create `scripts/seed.go` for sample data, then:

```bash
make seed
```

### Manual Sample Data

Use the admin interface to add:
1. Categories (Electronics, Clothing, Books)
2. Brands (Apple, Samsung, Nike)
3. Products with variants
4. Sample customers
5. Test orders

## Troubleshooting

### Database Connection Error

```bash
# Check PostgreSQL is running
pg_isadmin -h localhost -p 5432 -U postgres

# Recreate database
make db-reset
```

### Port Already in Use

```bash
# Use different port
forge runserver --port=9000
```

### Generated Code Issues

```bash
# Clean and regenerate
make clean
make generate
```

### Migration Issues

```bash
# Check status
make migrate-status

# Rollback if needed
make migrate-down

# Reset database
make db-reset
```

## Next Steps

1. **Read the Documentation**
   - `README.md` - Project overview
   - `CLI_USAGE_GUIDE.md` - Complete CLI reference
   - `../../docs/` - Framework documentation

2. **Explore the Code**
   - Review model definitions in `app/*/models.go`
   - Check admin configurations in `app/*/admin.go`
   - Examine API endpoints in `app/*/api.go`

3. **Customize**
   - Add your own models
   - Customize admin interface
   - Create custom API endpoints
   - Add business logic

4. **Deploy**
   - Build production binary: `make build`
   - Use Docker: `docker-compose up`
   - Deploy to cloud (AWS, GCP, Azure)

## Useful Commands

```bash
# Complete setup
make setup

# Development mode (auto-reload)
make dev

# Docker setup
make docker-up

# Database reset
make db-reset

# Project stats
make stats

# Help
make help
```

## Support

- **Documentation**: `../../docs/`
- **Issues**: GitHub Issues
- **Examples**: `../../examples/`

## License

MIT

---

**Happy Coding! 🚀**
