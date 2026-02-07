.PHONY: help setup reset start stop restart logs logs-admin logs-tenant logs-worker migrate seed test-admin-login test-tenant-login test-tenant clean

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
	@echo "  make test-tenant-login     - Test Tenant API login"
	@echo "  make test-tenant           - Create test tenant via Admin API"
	@echo "  make test-subscription     - Test subscription (public signup)"
	@echo "  make test-login-to-tenant  - Test login to tenant (features + permissions + config)"
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
