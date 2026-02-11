# Database Commands

.PHONY: migrate seed

# Apply Master DB migrations
migrate:
	@docker exec -i saas-postgres psql -U postgres -d master_db < migrations/master/001_initial_schema.up.sql

# Seed is no longer needed - all data is inserted via migration
seed:
	@echo "✓ All initial data created via migration (admin@teste.com / admin123)"
