# Forge Ecommerce Makefile

.PHONY: help install generate migrate migrate-down migrate-status run test clean docker-up docker-down seed

# Default target
help:
	@echo "Forge Ecommerce - Available Commands:"
	@echo ""
	@echo "  make install       - Install dependencies and setup project"
	@echo "  make generate      - Generate code from models"
	@echo "  make migrate       - Run database migrations"
	@echo "  make migrate-down  - Rollback last migration"
	@echo "  make migrate-status - Show migration status"
	@echo "  make run           - Start development server"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean generated files and build artifacts"
	@echo "  make docker-up     - Start Docker containers"
	@echo "  make docker-down   - Stop Docker containers"
	@echo "  make seed          - Seed database with sample data"
	@echo "  make all           - Full setup (install, generate, migrate, run)"
	@echo ""

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

# Generate code from models
generate:
	@echo "🔧 Generating code..."
	forge generate
	@echo "✅ Code generated"

# Run migrations
migrate:
	@echo "🗄️  Running migrations..."
	forge migrate
	@echo "✅ Migrations applied"

# Rollback migrations
migrate-down:
	@echo "↩️  Rolling back migration..."
	forge migrate rollback
	@echo "✅ Migration rolled back"

# Show migration status
migrate-status:
	@echo "📊 Migration status:"
	forge migrate status

# Start development server
run:
	@echo "🚀 Starting server..."
	forge runserver

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...
	@echo "✅ Tests complete"

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Clean generated files
clean:
	@echo "🧹 Cleaning..."
	find . -name "*_gen.go" -delete
	find . -name "*.test" -delete
	rm -f ecommerce
	rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

# Docker commands
docker-up:
	@echo "🐳 Starting Docker containers..."
	docker-compose up -d
	@echo "✅ Containers started"

docker-down:
	@echo "🐳 Stopping Docker containers..."
	docker-compose down
	@echo "✅ Containers stopped"

docker-logs:
	docker-compose logs -f

# Database commands
db-create:
	@echo "🗄️  Creating database..."
	createdb ecommerce_db || echo "Database already exists"

db-drop:
	@echo "⚠️  Dropping database..."
	dropdb ecommerce_db || echo "Database doesn't exist"

db-reset: db-drop db-create migrate
	@echo "✅ Database reset complete"

# Seed database with sample data
seed:
	@echo "🌱 Seeding database..."
	go run scripts/seed.go
	@echo "✅ Database seeded"

# Create superuser
superuser:
	@echo "👤 Creating superuser..."
	forge createsuperuser

# Build for production
build:
	@echo "🏗️  Building production binary..."
	go build -o ecommerce -ldflags="-s -w" main.go
	@echo "✅ Build complete: ./ecommerce"

# Full setup (for new developers)
setup: install db-create generate migrate
	@echo "✅ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Create a superuser: make superuser"
	@echo "  2. Start the server: make run"
	@echo "  3. Visit http://localhost:8000/admin/"

# Complete workflow
all: setup seed run

# Development workflow (watch for changes)
dev:
	@echo "🔄 Starting development mode..."
	forge runserver --reload

# Lint code
lint:
	@echo "🔍 Running linters..."
	golangci-lint run ./...
	@echo "✅ Linting complete"

# Format code
fmt:
	@echo "✨ Formatting code..."
	go fmt ./...
	gofmt -s -w .
	@echo "✅ Formatting complete"

# Check code
check:
	@echo "✅ Running checks..."
	forge check
	go vet ./...
	@echo "✅ Checks complete"

# Run everything (check, test, build)
ci: check test build
	@echo "✅ CI pipeline complete"

# Show project stats
stats:
	@echo "📊 Project Statistics:"
	@echo ""
	@echo "Models:"
	@find app -name "models.go" -exec wc -l {} + | tail -1
	@echo ""
	@echo "Generated Code:"
	@find app -name "*_gen.go" -exec wc -l {} + | tail -1 || echo "0 (run 'make generate')"
	@echo ""
	@echo "Tests:"
	@find . -name "*_test.go" -exec wc -l {} + | tail -1 || echo "0"
	@echo ""
	@echo "Total Lines of Code:"
	@find . -name "*.go" -not -name "*_gen.go" -not -path "*/vendor/*" -exec wc -l {} + | tail -1
