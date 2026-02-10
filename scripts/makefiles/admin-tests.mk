# Admin API Tests

.PHONY: test-admin-login test-plans-list test-plans-create test-features-list test-features-create test-sysusers-list test-sysusers-create

# Test Admin API login
test-admin-login:
	@curl -X POST http://localhost:8080/api/v1/admin/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@teste.com","password":"admin123"}'
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
