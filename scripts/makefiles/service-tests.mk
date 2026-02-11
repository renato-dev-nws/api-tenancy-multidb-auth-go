# Service CRUD Tests

.PHONY: test-service-create test-service-list test-service-get test-service-update test-service-delete test-services-all

# Test Tenant API - Services CRUD
test-service-create:
	@echo "Creating new service..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/$$URL_CODE/services \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Consultoria TI","description":"Consultoria em tecnologia","duration_minutes":60,"price":150.00}'
	@echo ""

test-service-list:
	@echo "Listing services..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/$$URL_CODE/services?page=1&page_size=10" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-service-get:
	@echo "Getting service by ID..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/$$URL_CODE/services/$(SERVICE_ID)" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-service-update:
	@echo "Updating service..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	curl -X PUT "http://localhost:8081/api/v1/$$URL_CODE/services/$(SERVICE_ID)" \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Consultoria TI Avançada","price":200.00,"duration_minutes":90}'
	@echo ""

test-service-delete:
	@echo "Deleting service (soft delete)..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	curl -X DELETE "http://localhost:8081/api/v1/$$URL_CODE/services/$(SERVICE_ID)" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-services-all:
	@echo "========================================="
	@echo "Running complete Services CRUD test"
	@echo "========================================="
	@echo ""
	@echo "1. Creating service..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/$$URL_CODE/services \
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
		curl -s -X GET "http://localhost:8081/api/v1/$$URL_CODE/services?page=1&page_size=10" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "3. Getting service by ID..."; \
		curl -s -X GET "http://localhost:8081/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "4. Updating service..."; \
		curl -s -X PUT "http://localhost:8081/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
			-H "Content-Type: application/json" \
			-H "Authorization: Bearer $$TOKEN" \
			-d '{"name":"Consultoria TI Avançada","price":200.00,"duration_minutes":90}'; \
		echo ""; \
		echo ""; \
		echo "5. Deleting service (soft delete)..."; \
		curl -s -X DELETE "http://localhost:8081/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "6. Listing services after delete..."; \
		curl -s -X GET "http://localhost:8081/api/v1/$$URL_CODE/services?page=1&page_size=10&active=false" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
	else \
		echo "ERROR: Failed to create service"; \
	fi
	@echo ""
	@echo "========================================="
	@echo "Test completed!"
	@echo "========================================="
