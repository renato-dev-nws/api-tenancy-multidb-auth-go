# Product CRUD Tests

.PHONY: test-product-create test-product-list test-product-get test-product-update test-product-delete test-products-all

# Test Tenant API - Products CRUD
test-product-create:
	@echo "Creating new product..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/$$URL_CODE/products \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Notebook Dell","description":"Intel i7, 16GB RAM","price":3500.00,"sku":"NB-DELL-001","stock":10}'
	@echo ""

test-product-list:
	@echo "Listing products..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/$$URL_CODE/products?page=1&page_size=10" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-product-get:
	@echo "Getting product by ID..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/$$URL_CODE/products/$(PRODUCT_ID)" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-product-update:
	@echo "Updating product..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	curl -X PUT "http://localhost:8081/api/v1/$$URL_CODE/products/$(PRODUCT_ID)" \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Notebook Dell Atualizado","price":3200.00,"stock":15}'
	@echo ""

test-product-delete:
	@echo "Deleting product (soft delete)..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	curl -X DELETE "http://localhost:8081/api/v1/$$URL_CODE/products/$(PRODUCT_ID)" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-products-all:
	@echo "========================================="
	@echo "Running complete Products CRUD test"
	@echo "========================================="
	@echo ""
	@echo "1. Creating product..."
	@LOGIN_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN_RESPONSE" | grep -o '"current_tenant"[^}]*"url_code":"[^"]*' | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/$$URL_CODE/products \
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
		curl -s -X GET "http://localhost:8081/api/v1/$$URL_CODE/products?page=1&page_size=10" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "3. Getting product by ID..."; \
		curl -s -X GET "http://localhost:8081/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "4. Updating product..."; \
		curl -s -X PUT "http://localhost:8081/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
			-H "Content-Type: application/json" \
			-H "Authorization: Bearer $$TOKEN" \
			-d '{"name":"Notebook Dell XPS","price":4000.00,"stock":8}'; \
		echo ""; \
		echo ""; \
		echo "5. Deleting product (soft delete)..."; \
		curl -s -X DELETE "http://localhost:8081/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
		echo ""; \
		echo "6. Listing products after delete..."; \
		curl -s -X GET "http://localhost:8081/api/v1/$$URL_CODE/products?page=1&page_size=10&active=false" \
			-H "Authorization: Bearer $$TOKEN"; \
		echo ""; \
	else \
		echo "ERROR: Failed to create product"; \
	fi
	@echo ""
	@echo "========================================="
	@echo "Test completed!"
	@echo "========================================="
