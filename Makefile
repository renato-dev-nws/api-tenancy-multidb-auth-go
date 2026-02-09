.PHONY: help setup reset start stop restart logs logs-admin logs-tenant logs-worker migrate seed test-admin-login test-tenant-login test-tenant clean test-product-create test-product-list test-product-get test-product-update test-product-delete test-products-all test-service-create test-service-list test-service-get test-service-update test-service-delete test-services-all

# Default target
help:
	@echo "=== SaaS Multi-Tenant API - Makefile ==="
	@echo ""
	@echo "Quick Start:"
	@echo "  make setup           - Complete initial setup (build + migrate + seed)"
	@echo "  make reset           - Reset everything (down -v + setup)"
	@echo "  make start           - Start all services"
	@echo "  make stop            - Stop all services"
	@echo "  make restart         - Restart all services"
	@echo ""
	@echo "Development:"
	@echo "  make logs            - View all logs (tail -f)"
	@echo "  make logs-admin      - View Admin API logs only"
	@echo "  make logs-tenant     - View Tenant API logs only"
	@echo "  make logs-worker     - View Worker logs only"
	@echo "  make migrate         - Apply Master DB migrations"
	@echo "  make seed            - Create admin user (admin@teste.com / admin123)"
	@echo ""
	@echo "Testing:"
	@echo "  make test-admin-login      - Test Admin API login"
	@echo "  make test-login            - Test Tenant API login (with interface)"
	@echo "  make test-switch-tenant    - Test tenant switching"
	@echo "  make test-subscription     - Test subscription (public signup)"
	@echo ""
	@echo "Product CRUD (use: wsl make <command>):"
	@echo "  make test-product-create   - Create a test product"
	@echo "  make test-product-list     - List all products"
	@echo "  make test-product-get      - Get product by ID (PRODUCT_ID=uuid)"
	@echo "  make test-product-update   - Update product (PRODUCT_ID=uuid)"
	@echo "  make test-product-delete   - Delete product (PRODUCT_ID=uuid)"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean           - Clean volumes and rebuild"
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

# View all logs
logs:
	@docker compose logs -f

# View Admin API logs only
logs-admin:
	@docker compose logs -f admin-api

# View Tenant API logs only
logs-tenant:
	@docker compose logs -f tenant-api

# View Worker logs only
logs-worker:
	@docker compose logs -f worker

# Apply Master DB migrations
migrate:
	@docker exec -i saas-postgres psql -U postgres -d master_db < migrations/master/001_initial_schema.up.sql

# Create admin user (idempotent)
seed:
	@docker exec saas-postgres psql -U postgres -d master_db -c "DELETE FROM user_profiles WHERE user_id IN (SELECT id FROM users WHERE email = 'admin@teste.com');" > /dev/null 2>&1 || true
	@docker exec saas-postgres psql -U postgres -d master_db -c "DELETE FROM users WHERE email = 'admin@teste.com';" > /dev/null 2>&1 || true
	@docker exec saas-postgres psql -U postgres -d master_db -c "INSERT INTO users (email, password_hash) VALUES ('admin@teste.com', '\$$2a\$$10\$$AlfQHB81zVpyRCL8x4NTeurxmM9skCihPZiACtivFcKV2hAiRy8M.');" > /dev/null
	@docker exec saas-postgres psql -U postgres -d master_db -c "INSERT INTO user_profiles (user_id, full_name) SELECT id, 'Admin User' FROM users WHERE email = 'admin@teste.com';" > /dev/null
	@echo "Admin user ready: admin@teste.com / admin123"

# Test Admin API login
test-admin-login:
	@curl -X POST http://localhost:8080/api/v1/admin/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@teste.com","password":"admin123"}'
	@echo ""

# Test Tenant API login
test-tenant-login:
	@curl -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@teste.com","password":"admin123"}'
	@echo ""

# Create test tenant via Admin API
test-tenant:
	@echo "Creating test tenant via Admin API..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/admin/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@teste.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8080/api/v1/admin/tenants \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Empresa Teste","url_code":"teste","plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","company_name":"Empresa Teste Ltda","is_company":true}'
	@echo ""
	@echo "Check worker logs: wsl make logs-worker"

# Test subscription endpoint (cadastro público de assinante)
test-subscription:
	@echo "Testing subscription endpoint..."
	@curl -X POST http://localhost:8081/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","name":"Empresa João Silva","subdomain":"joao","full_name":"João Silva","email":"joao@teste.com","password":"senha12345","is_company":false}'
	@echo ""

# Test tenant user login (returns complete config of last_tenant_id)
test-login:
	@echo "Testing login with interface config..."
	@curl -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'
	@echo ""

# Test switching active tenant
test-switch-tenant:
	@echo "Testing tenant switch..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/auth/switch-tenant \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"url_code":"joao"}'
	@echo ""

# Test Admin API - Plans CRUD
test-plans-list:
	@echo "Listing all plans..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/admin/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@teste.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/admin/plans \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-plans-create:
	@echo "Creating new plan..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/admin/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@teste.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8080/api/v1/admin/plans \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Enterprise Plan","description":"Full access plan","price":99.99,"feature_ids":["aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"]}'
	@echo ""

# Test Admin API - Features CRUD
test-features-list:
	@echo "Listing all features..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/admin/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@teste.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/admin/features \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-features-create:
	@echo "Creating new feature..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/admin/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@teste.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8080/api/v1/admin/features \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"title":"Users","slug":"users","code":"user","description":"User management module","is_active":true}'
	@echo ""

# Test Admin API - SysUsers CRUD
test-sysusers-list:
	@echo "Listing all sys users..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/admin/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@teste.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/admin/sys-users \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-sysusers-create:
	@echo "Creating new sys user..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/admin/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@teste.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8080/api/v1/admin/sys-users \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"email":"manager@teste.com","password":"manager123","full_name":"Manager User"}'
	@echo ""

# Clean everything
clean:
	@docker compose down -v
	@docker system prune -f
	@echo "All cleaned!"

# Test Tenant API - Products CRUD
test-product-create:
	@echo "Creating new product..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/95RM301XKTJ/products \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Notebook Dell","description":"Intel i7, 16GB RAM","price":3500.00,"sku":"NB-DELL-001","stock":10}'
	@echo ""

test-product-list:
	@echo "Listing products..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/95RM301XKTJ/products?page=1&page_size=10" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-product-get:
	@echo "Getting product by ID..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/95RM301XKTJ/products/$(PRODUCT_ID)" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-product-update:
	@echo "Updating product..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X PUT "http://localhost:8081/api/v1/95RM301XKTJ/products/$(PRODUCT_ID)" \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Notebook Dell Atualizado","price":3200.00,"stock":15}'
	@echo ""

test-product-delete:
	@echo "Deleting product (soft delete)..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X DELETE "http://localhost:8081/api/v1/95RM301XKTJ/products/$(PRODUCT_ID)" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-products-all:
	@echo "========================================="
	@echo "Running complete Products CRUD test"
	@echo "========================================="
	@echo ""
	@echo "1. Creating product..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/95RM301XKTJ/products \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Notebook Dell","description":"Intel i7, 16GB RAM","price":3500.00,"sku":"NB-DELL-001","stock":10}'); \
	echo "$$RESPONSE"; \
	PRODUCT_ID=$$(echo "$$RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; \
	if [ -n "$$PRODUCT_ID" ]; then \
		echo "Product created with ID: $$PRODUCT_ID"; \
		echo ""; \
		echo "2. Listing products..."; \
		curl -s -X GET "http://localhost:8081/api/v1/95RM301XKTJ/products?page=1&page_size=10" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "3. Getting product by ID..."; \
		curl -s -X GET "http://localhost:8081/api/v1/95RM301XKTJ/products/$$PRODUCT_ID" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "4. Updating product..."; \
		curl -s -X PUT "http://localhost:8081/api/v1/95RM301XKTJ/products/$$PRODUCT_ID" \
			-H "Content-Type: application/json" \
			-H "Authorization: Bearer $$TOKEN" \
			-d '{"name":"Notebook Dell XPS","price":4000.00,"stock":8}'; \
		echo ""; \
		echo ""; \
		echo "5. Deleting product (soft delete)..."; \
		curl -s -X DELETE "http://localhost:8081/api/v1/95RM301XKTJ/products/$$PRODUCT_ID" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "6. Listing products after delete..."; \
		curl -s -X GET "http://localhost:8081/api/v1/95RM301XKTJ/products?page=1&page_size=10&active=false" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
	else \
		echo "ERROR: Failed to create product"; \
	fi
	@echo ""
	@echo "========================================="
	@echo "Test completed!"
	@echo "========================================="

# Test Tenant API - Services CRUD
test-service-create:
	@echo "Creating new service..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/95RM301XKTJ/services \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Consultoria TI","description":"Consultoria em tecnologia","duration_minutes":60,"price":150.00}'
	@echo ""

test-service-list:
	@echo "Listing services..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/95RM301XKTJ/services?page=1&page_size=10" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-service-get:
	@echo "Getting service by ID..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/95RM301XKTJ/services/$(SERVICE_ID)" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-service-update:
	@echo "Updating service..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X PUT "http://localhost:8081/api/v1/95RM301XKTJ/services/$(SERVICE_ID)" \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Consultoria TI Avançada","price":200.00,"duration_minutes":90}'
	@echo ""

test-service-delete:
	@echo "Deleting service (soft delete)..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X DELETE "http://localhost:8081/api/v1/95RM301XKTJ/services/$(SERVICE_ID)" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-services-all:
	@echo "========================================="
	@echo "Running complete Services CRUD test"
	@echo "========================================="
	@echo ""
	@echo "1. Creating service..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/95RM301XKTJ/services \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Consultoria TI","description":"Consultoria em tecnologia","duration_minutes":60,"price":150.00}'); \
	echo "$$RESPONSE"; \
	SERVICE_ID=$$(echo "$$RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; \
	if [ -n "$$SERVICE_ID" ]; then \
		echo "Service created with ID: $$SERVICE_ID"; \
		echo ""; \
		echo "2. Listing services..."; \
		curl -s -X GET "http://localhost:8081/api/v1/95RM301XKTJ/services?page=1&page_size=10" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "3. Getting service by ID..."; \
		curl -s -X GET "http://localhost:8081/api/v1/95RM301XKTJ/services/$$SERVICE_ID" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "4. Updating service..."; \
		curl -s -X PUT "http://localhost:8081/api/v1/95RM301XKTJ/services/$$SERVICE_ID" \
			-H "Content-Type: application/json" \
			-H "Authorization: Bearer $$TOKEN" \
			-d '{"name":"Consultoria TI Avançada","price":200.00,"duration_minutes":90}'; \
		echo ""; \
		echo ""; \
		echo "5. Deleting service (soft delete)..."; \
		curl -s -X DELETE "http://localhost:8081/api/v1/95RM301XKTJ/services/$$SERVICE_ID" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "6. Listing services after delete..."; \
		curl -s -X GET "http://localhost:8081/api/v1/95RM301XKTJ/services?page=1&page_size=10&active=false" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
	else \
		echo "ERROR: Failed to create service"; \
	fi
	@echo ""
	@echo "========================================="
	@echo "Test completed!"
	@echo "========================================="
