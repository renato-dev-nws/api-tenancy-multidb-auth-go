# Logs Commands

.PHONY: logs logs-admin logs-tenant logs-worker

# View all logs
logs:
	@docker compose logs -f

# View Admin API logs only
logs-admin:
	@docker compose logs -f admin-api

# View Tenant API logs only
logs-tenant:
	@docker compose logs -f tenant-api

# View Worker logs only
logs-worker:
	@docker compose logs -f worker
