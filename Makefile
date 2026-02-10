# SaaS Multi-Tenant API - Makefile
# Modular structure: infrastructure commands here, features in scripts/makefiles/

# Import modular Makefiles
include scripts/makefiles/logs.mk
include scripts/makefiles/database.mk
include scripts/makefiles/admin-tests.mk
include scripts/makefiles/tenant-tests.mk
include scripts/makefiles/product-tests.mk
include scripts/makefiles/service-tests.mk
include scripts/makefiles/setting-tests.mk
include scripts/makefiles/image-tests.mk

# Infrastructure commands
.PHONY: help setup reset start stop restart clean

# Default target
help:
	@echo "========================================="
	@echo "SaaS Multi-Tenant API - Makefile"
	@echo "========================================="
	@echo ""
	@echo "Quick Start:"
	@echo "  make setup           - Complete initial setup (build + migrate + seed)"
	@echo "  make reset           - Reset everything (down -v + setup)"
	@echo "  make start           - Start all services"
	@echo "  make stop            - Stop all services"
	@echo "  make restart         - Restart all services"
	@echo "  make clean           - Clean volumes and rebuild"
	@echo ""
	@echo "Database:"
	@echo "  make migrate         - Apply Master DB migrations"
	@echo "  make seed            - Create admin user (admin@teste.com / admin123)"
	@echo ""
	@echo "Logs:"
	@echo "  make logs            - View all logs (tail -f)"
	@echo "  make logs-admin      - View Admin API logs only"
	@echo "  make logs-tenant     - View Tenant API logs only"
	@echo "  make logs-worker     - View Worker logs only"
	@echo ""
	@echo "Admin API Tests:"
	@echo "  make test-admin-login      - Test Admin API login"
	@echo "  make test-plans-list       - List all plans"
	@echo "  make test-plans-create     - Create new plan"
	@echo "  make test-features-list    - List all features"
	@echo "  make test-features-create  - Create new feature"
	@echo "  make test-sysusers-list    - List all sys users"
	@echo "  make test-sysusers-create  - Create new sys user"
	@echo ""
	@echo "Tenant API Tests:"
	@echo "  make test-login            - Test Tenant API login (with interface)"
	@echo "  make test-tenant           - Create test tenant via Admin API"
	@echo "  make test-subscription     - Test subscription (public signup)"
	@echo "  make test-new-tenant       - Test new tenant + settings validation"
	@echo "  make test-testenovo        - Test existing testenovo tenant"
	@echo "  make test-switch-tenant    - Test tenant switching"
	@echo ""
	@echo "Product CRUD Tests:"
	@echo "  make test-product-create   - Create a test product"
	@echo "  make test-product-list     - List all products"
	@echo "  make test-product-get      - Get product by ID (PRODUCT_ID=uuid)"
	@echo "  make test-product-update   - Update product (PRODUCT_ID=uuid)"
	@echo "  make test-product-delete   - Delete product (PRODUCT_ID=uuid)"
	@echo "  make test-products-all     - Run complete Products CRUD test"
	@echo ""
	@echo "Service CRUD Tests:"
	@echo "  make test-service-create   - Create a test service"
	@echo "  make test-service-list     - List all services"
	@echo "  make test-service-get      - Get service by ID (SERVICE_ID=uuid)"
	@echo "  make test-service-update   - Update service (SERVICE_ID=uuid)"
	@echo "  make test-service-delete   - Delete service (SERVICE_ID=uuid)"
	@echo "  make test-services-all     - Run complete Services CRUD test"
	@echo ""
	@echo "Settings Tests:"
	@echo "  make test-setting-list     - List all settings"
	@echo "  make test-setting-get      - Get interface setting"
	@echo "  make test-setting-update   - Update interface setting"
	@echo "  make test-settings-all     - Run complete Settings test"
	@echo ""
	@echo "Image Tests:"
	@echo "  make test-image-upload     - Upload test image (requires test-image.jpg)"
	@echo "  make test-image-list       - List images for entity"
	@echo "  make test-image-get        - Get image by ID (IMAGE_ID=uuid)"
	@echo "  make test-image-update     - Update image metadata (IMAGE_ID=uuid)"
	@echo "  make test-image-delete     - Delete image (IMAGE_ID=uuid)"
	@echo "  make test-images-all       - Run complete Images workflow"
	@echo ""
	@echo "========================================="
	@echo ""

# Complete setup: build, start, migrate, seed
setup:
	@echo "==> Building services..."
	@docker compose build
	@echo ""
	@echo "==> Starting services..."
	@docker compose up -d
	@echo ""
	@echo "==> Waiting for database..."
	@sleep 8
	@echo ""
	@echo "==> Applying migrations..."
	@$(MAKE) migrate
	@echo ""
	@echo "==> Creating admin user..."
	@$(MAKE) seed
	@echo ""
	@echo "===================================="
	@echo "Setup complete!"
	@echo "Admin: admin@teste.com / admin123"
	@echo "API: http://localhost:8080"
	@echo "===================================="

# Reset everything (clean volumes + setup)
reset:
	@echo "==> Stopping and removing all containers and volumes..."
	@docker compose down -v
	@echo ""
	@$(MAKE) setup

# Start all services
start:
	@docker compose up -d
	@echo "Services started!"

# Stop all services
stop:
	@docker compose down
	@echo "Services stopped!"

# Restart all services
restart:
	@docker compose restart
	@echo "Services restarted!"

# Clean everything
clean:
	@docker compose down -v
	@docker system prune -f
	@echo "All cleaned!"
