.PHONY: help setup dev migrate-up migrate-down docker-up docker-down docker-rebuild logs test clean

# Default target
help:
	@echo "Available commands:"
	@echo "  make setup           - Setup development environment"
	@echo "  make dev             - Run API locally (requires Docker services)"
	@echo "  make migrate-up      - Run database migrations"
	@echo "  make migrate-down    - Rollback database migrations"
	@echo "  make docker-up       - Start all Docker services"
	@echo "  make docker-down     - Stop all Docker services"
	@echo "  make docker-rebuild  - Rebuild and restart Docker services"
	@echo "  make logs            - View Docker logs"
	@echo "  make test            - Run tests"
	@echo "  make clean           - Clean build artifacts"

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@if not exist .env copy .env.example .env
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Setup complete! Edit .env file with your configuration."

# Run API locally (requires Docker services to be running)
dev:
	@echo "Running API locally..."
	go run cmd/api/main.go

# Run database migrations
migrate-up:
	@echo "Running Master DB migrations..."
	@docker exec -i saas-postgres psql -U postgres -d master_db < migrations/master/001_initial_schema.up.sql
	@echo "Master DB migrations completed!"

# Rollback database migrations
migrate-down:
	@echo "Rolling back Master DB migrations..."
	@docker exec -i saas-postgres psql -U postgres -d master_db < migrations/master/001_initial_schema.down.sql
	@echo "Master DB migrations rolled back!"

# Start all Docker services
docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d postgres redis pgbouncer
	@echo "Waiting for services to be ready..."
	@timeout /t 5 /nobreak > nul
	@echo "Docker services started!"

# Stop all Docker services
docker-down:
	@echo "Stopping Docker services..."
	docker-compose down
	@echo "Docker services stopped!"

# Rebuild and restart Docker services
docker-rebuild:
	@echo "Rebuilding Docker services..."
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d
	@echo "Docker services rebuilt and started!"

# View Docker logs
logs:
	docker-compose logs -f

# View API logs only
logs-api:
	docker-compose logs -f api

# View Postgres logs only
logs-db:
	docker-compose logs -f postgres

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@if exist bin rmdir /s /q bin
	@echo "Clean complete!"

# Build the application
build:
	@echo "Building application..."
	@if not exist bin mkdir bin
	go build -o bin/api.exe cmd/api/main.go
	@echo "Build complete! Binary at bin/api.exe"

# Run full development stack
full-dev: docker-up migrate-up dev

# Reset database (WARNING: This will delete all data!)
reset-db:
	@echo "WARNING: This will delete all data!"
	@set /p confirm="Are you sure? (yes/no): "
	@if "$(confirm)"=="yes" (
		$(MAKE) docker-down
		docker volume rm saas-multi-database-api_postgres_data || echo Volume already removed
		$(MAKE) docker-up
		$(MAKE) migrate-up
		@echo "Database reset complete!"
	) else (
		@echo "Database reset cancelled."
	)
