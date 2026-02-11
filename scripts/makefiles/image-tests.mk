# Image Upload Tests - WSL/Bash Version

.PHONY: create-test-image test-image-upload test-image-list test-image-get test-image-update test-image-delete test-images-complete test-full-setup

# Complete setup from scratch (reset + provision + test)
test-full-setup:
	@echo "========================================="
	@echo "Complete Setup + Image Upload Test"
	@echo "========================================="
	@echo ""
	@echo "[1/10] Resetting environment..."
	@docker compose down -v > /dev/null 2>&1
	@echo "✓ Environment reset"
	@echo ""
	@echo "[2/10] Building services..."
	@docker compose build > /dev/null 2>&1
	@echo "✓ Services built"
	@echo ""
	@echo "[3/10] Starting services..."
	@docker compose up -d
	@echo "✓ Services started"
	@echo ""
	@echo "[4/10] Waiting for database..."
	@sleep 10
	@echo "✓ Database ready"
	@echo ""
	@echo "[5/10] Applying Master DB migrations..."
	@docker exec -i saas-postgres psql -U postgres -d master_db < migrations/master/001_initial_schema.up.sql > /dev/null 2>&1
	@echo "✓ Master DB migrated"
	@echo ""
	@echo "[6/10] Creating admin user..."
	@docker exec saas-postgres psql -U postgres -d master_db -c "DELETE FROM user_profiles WHERE user_id IN (SELECT id FROM users WHERE email = 'admin@teste.com');" > /dev/null 2>&1 || true
	@docker exec saas-postgres psql -U postgres -d master_db -c "DELETE FROM users WHERE email = 'admin@teste.com';" > /dev/null 2>&1 || true
	@docker exec saas-postgres psql -U postgres -d master_db -c "INSERT INTO users (email, password_hash) VALUES ('admin@teste.com', '\$$2a\$$10\$$AlfQHB81zVpyRCL8x4NTeurxmM9skCihPZiACtivFcKV2hAiRy8M.');" > /dev/null 2>&1
	@docker exec saas-postgres psql -U postgres -d master_db -c "INSERT INTO user_profiles (user_id, full_name) SELECT id, 'Admin User' FROM users WHERE email = 'admin@teste.com';" > /dev/null 2>&1
	@echo "✓ Admin user created"
	@echo ""
	@echo "[7/10] Cleaning previous tenant data..."
	@docker exec saas-postgres psql -U postgres -d master_db -c "DELETE FROM tenant_members WHERE user_id IN (SELECT id FROM users WHERE email = 'joao@teste.com');" > /dev/null 2>&1 || true
	@docker exec saas-postgres psql -U postgres -d master_db -c "DELETE FROM user_profiles WHERE user_id IN (SELECT id FROM users WHERE email = 'joao@teste.com');" > /dev/null 2>&1 || true
	@docker exec saas-postgres psql -U postgres -d master_db -c "DELETE FROM users WHERE email = 'joao@teste.com';" > /dev/null 2>&1 || true
	@docker exec saas-postgres psql -U postgres -d master_db -c "DELETE FROM subscriptions WHERE tenant_id IN (SELECT id FROM tenants WHERE subdomain = 'joao');" > /dev/null 2>&1 || true
	@docker exec saas-postgres psql -U postgres -d master_db -c "DELETE FROM tenants WHERE subdomain = 'joao';" > /dev/null 2>&1 || true
	@echo "✓ Previous data cleaned"
	@echo ""
	@echo "[8/10] Creating tenant..."
	@RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","name":"Empresa João Silva","subdomain":"joao","full_name":"João Silva","email":"joao@teste.com","password":"senha12345","is_company":false}'); \
	URL_CODE=$$(echo "$$RESPONSE" | grep -o '"url_code":"[^"]*"' | cut -d'"' -f4); \
	if [ -z "$$URL_CODE" ]; then \
		echo "✗ Tenant creation failed"; \
		echo "Response: $$RESPONSE"; \
		exit 1; \
	fi; \
	echo "✓ Tenant created: $$URL_CODE"; \
	echo ""; \
	echo "[9/10] Waiting for Worker to provision database..."; \
	sleep 12; \
	echo "✓ Database provisioned"; \
	echo ""; \
	echo "[10/10] Creating product and testing image upload..."; \
	TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4); \
	if [ -z "$$TOKEN" ]; then \
		echo "✗ Login failed"; \
		exit 1; \
	fi; \
	SKU="PROD-$$(date +%s)"; \
	PRODUCT_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/$$URL_CODE/products \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"Test Product\",\"sku\":\"$$SKU\",\"price\":99.90,\"stock_quantity\":10,\"is_active\":true}"); \
	PRODUCT_ID=$$(echo "$$PRODUCT_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4); \
	if [ -z "$$PRODUCT_ID" ]; then \
		echo "✗ Product creation failed"; \
		echo "Response: $$PRODUCT_RESPONSE"; \
		exit 1; \
	fi; \
	echo "✓ Product created: $$PRODUCT_ID"; \
	echo ""; \
	if [ ! -f test-image.jpg ]; then \
		printf '%s' '/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAA/9k=' | base64 -d > test-image.jpg 2>/dev/null; \
	fi; \
	UPLOAD_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/$$URL_CODE/images \
		-H "Authorization: Bearer $$TOKEN" \
		-F "imageable_type=product" \
		-F "imageable_id=$$PRODUCT_ID" \
		-F "files=@test-image.jpg" \
		-F "titles=Product Image" \
		-F "alt_texts=Test image"); \
	UPLOADED_COUNT=$$(echo "$$UPLOAD_RESPONSE" | grep -o '"uploaded":[0-9]*' | cut -d':' -f2); \
	if [ -z "$$UPLOADED_COUNT" ] || [ "$$UPLOADED_COUNT" = "0" ]; then \
		echo "✗ Image upload failed"; \
		echo "Response: $$UPLOAD_RESPONSE"; \
		exit 1; \
	fi; \
	IMAGE_ID=$$(echo "$$UPLOAD_RESPONSE" | grep -o '"id":"[a-f0-9-]*"' | head -1 | cut -d'"' -f4); \
	echo "✓ Image uploaded: $$IMAGE_ID"; \
	echo ""; \
	sleep 3; \
	LIST_RESPONSE=$$(curl -s http://localhost:8081/api/v1/$$URL_CODE/images?imageable_type=product&imageable_id=$$PRODUCT_ID \
		-H "Authorization: Bearer $$TOKEN"); \
	TOTAL=$$(echo "$$LIST_RESPONSE" | grep -o '"id":' | wc -l); \
	echo "✓ Total images (with variants): $$TOTAL"; \
	echo ""; \
	echo "========================================="; \
	echo "✓ SETUP COMPLETE!"; \
	echo "========================================="; \
	echo ""; \
	echo "Environment ready:"; \
	echo "  URL Code: $$URL_CODE"; \
	echo "  User: joao@teste.com / senha12345"; \
	echo "  Product ID: $$PRODUCT_ID"; \
	echo "  Image ID: $$IMAGE_ID"; \
	echo ""; \
	echo "Test commands:"; \
	echo "  make test-image-list PRODUCT_ID=$$PRODUCT_ID"; \
	echo "  make test-image-get IMAGE_ID=$$IMAGE_ID"

# Create a test image (1x1 pixel JPEG)
create-test-image:
	@echo '/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAA/9k=' | base64 -d > test-image.jpg
	@echo "✓ Test image created: test-image.jpg"

# Complete end-to-end image workflow test
test-images-complete:
	@echo "========================================="
	@echo "Image Upload - Complete Workflow Test"
	@echo "========================================="
	@echo ""
	@bash scripts/test-images.sh

# Quick upload test (requires PRODUCT_ID variable)
test-image-upload:
	@if [ ! -f test-image.jpg ]; then \
		echo "ERROR: test-image.jpg not found. Run: make create-test-image"; \
		exit 1; \
	fi
	@if [ -z "$(PRODUCT_ID)" ]; then \
		echo "ERROR: PRODUCT_ID not set. Usage: make test-image-upload PRODUCT_ID=<uuid>"; \
		exit 1; \
	fi
	@echo "Uploading test image..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/95RM301XKTJ/images \
		-H "Authorization: Bearer $$TOKEN" \
		-F "imageable_type=product" \
		-F "imageable_id=$(PRODUCT_ID)" \
		-F "files=@test-image.jpg" \
		-F "titles=Test Product Image" \
		-F "alt_texts=Test image description"

# List images for a product
test-image-list:
	@if [ -z "$(PRODUCT_ID)" ]; then \
		echo "ERROR: PRODUCT_ID not set. Usage: make test-image-list PRODUCT_ID=<uuid>"; \
		exit 1; \
	fi
	@echo "Listing images..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4); \
	curl -s "http://localhost:8081/api/v1/95RM301XKTJ/images?imageable_type=product&imageable_id=$(PRODUCT_ID)" \
		-H "Authorization: Bearer $$TOKEN"

# Get single image by ID
test-image-get:
	@if [ -z "$(IMAGE_ID)" ]; then \
		echo "ERROR: IMAGE_ID not set. Usage: make test-image-get IMAGE_ID=<uuid>"; \
		exit 1; \
	fi
	@echo "Getting image..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4); \
	curl -s "http://localhost:8081/api/v1/95RM301XKTJ/images/$(IMAGE_ID)" \
		-H "Authorization: Bearer $$TOKEN"

# Update image metadata
test-image-update:
	@if [ -z "$(IMAGE_ID)" ]; then \
		echo "ERROR: IMAGE_ID not set. Usage: make test-image-update IMAGE_ID=<uuid>"; \
		exit 1; \
	fi
	@echo "Updating image metadata..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4); \
	curl -s -X PUT "http://localhost:8081/api/v1/95RM301XKTJ/images/$(IMAGE_ID)" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"title":"Updated Title","alt_text":"Updated description","display_order":1}'

# Delete image
test-image-delete:
	@if [ -z "$(IMAGE_ID)" ]; then \
		echo "ERROR: IMAGE_ID not set. Usage: make test-image-delete IMAGE_ID=<uuid>"; \
		exit 1; \
	fi
	@echo "Deleting image..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4); \
	curl -s -X DELETE "http://localhost:8081/api/v1/95RM301XKTJ/images/$(IMAGE_ID)" \
		-H "Authorization: Bearer $$TOKEN"
	@echo "✓ Image deleted successfully"
