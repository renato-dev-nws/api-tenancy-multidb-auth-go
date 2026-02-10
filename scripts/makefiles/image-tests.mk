# Image Upload Tests

.PHONY: test-image-upload test-image-list test-image-get test-image-update test-image-delete test-images-all

# Test image upload (requires a test image file)
test-image-upload:
	@echo "Uploading test image..."
	@if (-not (Test-Path "test-image.jpg")) { \
		Write-Host "ERROR: test-image.jpg not found. Create a test image first."; \
		exit 1; \
	}
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/95RM301XKTJ/images \
		-H "Authorization: Bearer $$TOKEN" \
		-F "imageable_type=product" \
		-F "imageable_id=PRODUCT_ID_HERE" \
		-F "files=@test-image.jpg" \
		-F "titles=Test Product Image" \
		-F "alt_texts=A beautiful test image"
	@echo ""

# Test listing images for a product
test-image-list:
	@echo "Listing images for product..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/95RM301XKTJ/images?imageable_type=product&imageable_id=PRODUCT_ID_HERE" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test getting a single image
test-image-get:
	@echo "Getting image by ID..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/95RM301XKTJ/images/$(IMAGE_ID)" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test updating image metadata
test-image-update:
	@echo "Updating image metadata..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X PUT "http://localhost:8081/api/v1/95RM301XKTJ/images/$(IMAGE_ID)" \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"title":"Updated Image Title","alt_text":"Updated alt text","display_order":1}'
	@echo ""

# Test deleting an image
test-image-delete:
	@echo "Deleting image..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X DELETE "http://localhost:8081/api/v1/95RM301XKTJ/images/$(IMAGE_ID)" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Complete image workflow test
test-images-all:
	@echo "========================================="
	@echo "Running complete Images workflow test"
	@echo "========================================="
	@echo ""
	@echo "1. Creating a test product first..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	PRODUCT_RESPONSE=$$(curl -s -X POST http://localhost:8081/api/v1/95RM301XKTJ/products \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Product with Images","description":"Test product for image upload","price":99.99,"sku":"IMG-TEST-001","stock":5}'); \
	echo "$$PRODUCT_RESPONSE"; \
	PRODUCT_ID=$$(echo "$$PRODUCT_RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; \
	if [ -n "$$PRODUCT_ID" ]; then \
		echo "Product created with ID: $$PRODUCT_ID"; \
		echo ""; \
		echo "Note: To complete this test, you need:"; \
		echo "1. Create a test image file: test-image.jpg"; \
		echo "2. Run: make test-image-upload PRODUCT_ID=$$PRODUCT_ID"; \
		echo "3. Run: make test-image-list PRODUCT_ID=$$PRODUCT_ID"; \
	else \
		echo "ERROR: Failed to create product"; \
	fi
	@echo ""
	@echo "========================================="
	@echo "Test info displayed"
	@echo "========================================="
