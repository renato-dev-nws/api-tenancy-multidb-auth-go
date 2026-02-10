# Service CRUD Tests

.PHONY: test-service-create test-service-list test-service-get test-service-update test-service-delete test-services-all

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
