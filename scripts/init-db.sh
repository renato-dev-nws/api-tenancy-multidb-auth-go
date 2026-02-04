#!/bin/bash
set -e

# Create API user
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER saas_api WITH PASSWORD 'saas_api_password';
    GRANT CONNECT ON DATABASE master_db TO saas_api;
    GRANT USAGE ON SCHEMA public TO saas_api;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO saas_api;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO saas_api;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO saas_api;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO saas_api;
EOSQL

echo "Database initialization completed"
