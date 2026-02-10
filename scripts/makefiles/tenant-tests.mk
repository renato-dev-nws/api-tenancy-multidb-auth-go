# Tenant API Tests

.PHONY: test-tenant-login test-tenant test-subscription test-new-tenant test-testenovo test-login test-switch-tenant

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

# Test new tenant creation with settings validation
test-new-tenant:
	@echo "========================================="
	@echo "Testing new tenant creation + settings"
	@echo "========================================="
	@echo ""
	@echo "1. Creating brand new tenant..."
	@RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","name":"Nova Empresa","subdomain":"novaempresa","full_name":"Usuario Novo","email":"novo@empresa.com","password":"senha12345","is_company":false}'); \
	echo "$$RESPONSE"; \
	URL_CODE=$$(echo "$$RESPONSE" | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	echo ""; \
	if [ -n "$$URL_CODE" ]; then \
		echo "Tenant created with URL code: $$URL_CODE"; \
		echo ""; \
		echo "2. Waiting 10 seconds for Worker to provision database..."; \
		sleep 10; \
		echo ""; \
		echo "3. Logging in with new tenant..."; \
		TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
			-H "Content-Type: application/json" \
			-d '{"email":"novo@empresa.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
		echo "Login successful"; \
		echo ""; \
		echo "4. Checking settings from new tenant database..."; \
		curl -s -X GET "http://localhost:8081/api/v1/$$URL_CODE/settings" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "5. Getting interface setting details..."; \
		curl -s -X GET "http://localhost:8081/api/v1/$$URL_CODE/settings/interface" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
	else \
		echo "ERROR: Failed to create tenant"; \
	fi
	@echo ""
	@echo "========================================="
	@echo "Test completed!"
	@echo "========================================="

# Test existing testenovo tenant settings
test-testenovo:
	@echo "Testing existing testenovo tenant settings..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"teste@novo.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	echo "Settings list:"; \
	curl -s -X GET "http://localhost:8081/api/v1/DR9AKNEZV2P/settings" \
		-H "Authorization: Bearer $$TOKEN"; \
	echo ""; \
	echo ""; \
	echo "Interface setting:"; \
	curl -s -X GET "http://localhost:8081/api/v1/DR9AKNEZV2P/settings/interface" \
		-H "Authorization: Bearer $$TOKEN"
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
