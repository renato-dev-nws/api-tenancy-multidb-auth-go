# Settings Tests (Read/Update Only - Fixed Configurations)

.PHONY: test-setting-list test-setting-get test-setting-update test-settings-all

# Test Tenant API - Settings (Read/Update)
test-setting-list:
	@echo "Listing all settings..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/95RM301XKTJ/settings" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-setting-get:
	@echo "Getting setting by key..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET "http://localhost:8081/api/v1/95RM301XKTJ/settings/interface" \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-setting-update:
	@echo "Updating setting..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X PUT "http://localhost:8081/api/v1/95RM301XKTJ/settings/interface" \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"value":{"logo":"https://example.com/logo.png","primary_color":"#FF5733","secondary_color":"#33FF57"}}'
	@echo ""

test-settings-all:
	@echo "========================================="
	@echo "Running complete Settings test"
	@echo "========================================="
	@echo ""
	@echo "1. Listing all settings..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@teste.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -s -X GET "http://localhost:8081/api/v1/95RM301XKTJ/settings" \
		-H "Authorization: Bearer $$TOKEN"; \
	echo ""; \
	echo ""; \
	echo "2. Getting interface setting..."; \
	curl -s -X GET "http://localhost:8081/api/v1/95RM301XKTJ/settings/interface" \
		-H "Authorization: Bearer $$TOKEN"; \
	echo ""; \
	echo ""; \
	echo "3. Updating interface setting..."; \
	curl -s -X PUT "http://localhost:8081/api/v1/95RM301XKTJ/settings/interface" \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"value":{"logo":"https://example.com/new-logo.png","primary_color":"#0066CC","secondary_color":"#FFCC00"}}'; \
	echo ""; \
	echo ""; \
	echo "4. Verifying update..."; \
	curl -s -X GET "http://localhost:8081/api/v1/95RM301XKTJ/settings/interface" \
		-H "Authorization: Bearer $$TOKEN"; \
	echo ""; \
	echo ""
	@echo "========================================="
	@echo "Test completed!"
	@echo "Note: Settings are fixed configs (no create/delete)"
	@echo "========================================="
