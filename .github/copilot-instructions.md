# AI Coding Agent Instructions - Multi-Tenant SaaS API

## Project Overview
Go-based multi-tenant SaaS API with **database-per-tenant isolation** (`db_tenant_{uuid}`). Central Control Plane manages subscriptions, feature-based plans, and RBAC across isolated tenant databases.

## Architecture Principles

### Database Strategy
- **Master DB (Control Plane)**: Users, tenants, plans, features, RBAC (`users`, `tenants`, `plans`, `features`, `roles`, `permissions`)
- **Tenant DBs**: Isolated per tenant as `db_tenant_{db_code}` (products, services, settings)
- **Connection Pooling**: PgBouncer on port 6432 in transaction mode, dynamically routing to tenant databases
- **Driver**: Use `pgx/v5` with `pgxpool.Pool` for connection management

### Tenant Resolution Flow
1. Extract `:url_code` from route/subdomain via middleware
2. Check Redis cache for `db_code` mapping (format: `tenant:urlcode:{url_code}`)
3. Validate user access via `tenant_members` table
4. Query plan features from `plan_features` join
5. Inject `*pgxpool.Pool` and active features into `context.Context` as:
   - `ctx.Value("tenant_pool")`: database connection pool
   - `ctx.Value("features")`: `[]string` of feature slugs

### Provisioning Workflow
- **Synchronous**: Create tenant record in Master DB with `status='provisioning'`
- **Asynchronous**: Publish event to Redis queue (`tenant:provision:{db_code}`)
- **Worker**: Consume event → `CREATE DATABASE` → Apply schema migrations → Update status to `active`

## Code Conventions

### Project Structure (Expected)
```
cmd/
  api/         # Main API server
  worker/      # Async provisioning worker
internal/
  cache/       # Redis cache wrapper
  config/      # Configuration management
  database/    # Dynamic pool manager
  middleware/  # Tenant resolution, auth, RBAC
  models/      # Database models (Master & Tenant schemas)
  services/    # Business logic (TenantService, FeatureService)
  repository/  # Data access layer
  utils/       # Utilities (JWT, password hashing, slug generation)
migrations/
  master/      # Control Plane schema
  tenant/      # Tenant DB schema template
```

### Middleware Pattern
```go
// Tenant resolution must:
// 1. Extract url_code from gin.Context param
// 2. Get db_code from Redis (cache aside pattern)
// 3. Retrieve or create pgxpool.Pool for tenant DB
// 4. Inject pool into context: c.Set("tenant_pool", pool)
// 5. Inject features: c.Set("features", []string{"products", "services"})
```

### Feature-Based Authorization
```go
// Controllers must check feature availability before execution:
func CreateProduct(c *gin.Context) {
    features := c.MustGet("features").([]string)
    if !slices.Contains(features, "products") {
        c.JSON(403, gin.H{"error": "feature_disabled"})
        return
    }
    // ... proceed with tenant DB pool
}
```

### Frontend Config Endpoint
`GET /adm/:url_code/config` must return JSON:
```json
{
  "features": ["products", "services"],
  "permissions": ["create_product", "delete_user"]
}
```

## Critical Patterns

### Connection Pool Management
- **Never create pools per request** - cache pools by `db_code` in memory (sync.Map or similar)
- Use context timeout for pool acquisition: `pgxpool.AcquireConnection(ctx, 5*time.Second)`
- Close pools on tenant deletion/suspension

### Security
- API user must NOT be superuser - grant only `CONNECT` on tenant databases
- SQL injection: Always use parameterized queries (`$1`, `$2`)
- Password hashing: Use `bcrypt` with cost 12+

### Testing Strategy
- Unit tests: Mock `pgxpool.Pool` using interfaces
- Integration tests: Use `testcontainers-go` for Postgres instances
- Tenant isolation tests: Verify cross-tenant data leakage prevention

## Common Tasks

### Adding a New Feature Module
1. Add row to `features` table (slug: `feature_name`)
2. Link to plans via `plan_features`
3. Create migration in `migrations/tenant/` for new tables
4. Update worker schema application logic
5. Add feature check in relevant controllers

### Database Migrations
- Master DB: Versioned migrations in `migrations/master/`
- Tenant DBs: Template migrations in `migrations/tenant/` applied by worker
- Use migration tool: `golang-migrate/migrate` or `pressly/goose`

### Local Development
- Start infrastructure: `docker-compose up -d` (Postgres, Redis, PgBouncer)
- Run migrations: `make migrate-master` / `make migrate-tenant`
- Start API: `make run-api` (port 8080)
- Start worker: `make run-worker`

## Key Files to Reference
- `SPEC.md`: Complete architectural specification
- `internal/middleware/tenant.go`: Tenant resolution logic (when created)
- `internal/database/manager.go`: Dynamic pool management (when created)
- `internal/utils/auth.go`: JWT and authentication utilities
- `cmd/worker/main.go`: Async provisioning flow (when created)

## Anti-Patterns to Avoid
- ❌ Creating database connections per request
- ❌ Hardcoding tenant database names
- ❌ Skipping feature checks in controllers
- ❌ Using ORM migrations for tenant DBs (use raw SQL for worker execution)
- ❌ Storing tenant data in Master DB (except metadata)
