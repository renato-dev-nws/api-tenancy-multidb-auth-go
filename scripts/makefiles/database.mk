# Database Commands

.PHONY: migrate seed

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
