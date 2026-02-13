# API Endpoints Reference

## Admin API (Port 8080)

### Authentication
```
POST /api/v1/admin/register  - Register new admin user
POST /api/v1/admin/login     - Admin login
GET  /api/v1/admin/me        - Get current admin user (protected)
```

### Tenants Management (Protected)
```
POST   /api/v1/admin/tenants         - Create new tenant
GET    /api/v1/admin/tenants         - List my tenants
GET    /api/v1/admin/tenants/:id     - Get tenant details
PUT    /api/v1/admin/tenants/:id     - Update tenant
DELETE /api/v1/admin/tenants/:id     - Delete tenant
```

### Plans Management (Protected)
```
GET    /api/v1/admin/plans           - List all plans
GET    /api/v1/admin/plans/:id       - Get plan details
POST   /api/v1/admin/plans           - Create new plan
PUT    /api/v1/admin/plans/:id       - Update plan
DELETE /api/v1/admin/plans/:id       - Delete plan
```

### Features Management (Protected)
```
GET    /api/v1/admin/features        - List all features
GET    /api/v1/admin/features/:id    - Get feature details
POST   /api/v1/admin/features        - Create new feature
PUT    /api/v1/admin/features/:id    - Update feature
DELETE /api/v1/admin/features/:id    - Delete feature
```

### System Users (Protected)
```
GET    /api/v1/admin/sys-users       - List all system users
GET    /api/v1/admin/sys-users/:id   - Get system user details
POST   /api/v1/admin/sys-users       - Create system user
PUT    /api/v1/admin/sys-users/:id   - Update system user
DELETE /api/v1/admin/sys-users/:id   - Delete system user
```

---

## Tenant API (Port 8081)

### Public Endpoints (No authentication required)

#### Authentication
```
POST /api/v1/auth/register      - Register tenant user
POST /api/v1/auth/login        - Tenant user login
POST /api/v1/subscription      - Create new subscription (self-service)
```

#### Plans (Public - for registration)
```
GET  /api/v1/plans              - List all available plans
```

**Example Response:**
```json
{
  "plans": [
    {
      "id": "uuid",
      "name": "Basic",
      "description": "Basic plan for small teams",
      "price": 29.99,
      "features": [
        {
          "id": "uuid",
          "name": "Products Module",
          "slug": "products",
          "description": "Manage products"
        },
        {
          "id": "uuid",
          "name": "Services Module",
          "slug": "services",
          "description": "Manage services"
        }
      ],
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-01-15T10:00:00Z"
    }
  ],
  "total": 1
}
```

### Protected Endpoints (Requires authentication)

#### User Management
```
GET  /api/v1/auth/me              - Get current user
POST /api/v1/auth/switch-tenant  - Switch active tenant
GET  /api/v1/tenants              - List user's tenants
```

### Tenant-Scoped Endpoints (Requires tenant context)

All routes use the pattern: `/api/v1/:url_code/...`

#### Configuration
```
GET  /api/v1/:url_code/config    - Get tenant configuration (features, permissions, layout)
```

**Example Response:**
```json
{
  "features": ["products", "services"],
  "permissions": ["create_product", "delete_product", "view_reports"],
  "layout": {
    "logo_url": "https://cdn.example.com/uploads/logo.png",
    "primary_color": "#3B82F6",
    "secondary_color": "#10B981",
    "company_name": "My Company"
  }
}
```

#### Products (Feature: products)
```
GET    /api/v1/:url_code/products        - List products
GET    /api/v1/:url_code/products/:id    - Get product details
POST   /api/v1/:url_code/products        - Create product
PUT    /api/v1/:url_code/products/:id    - Update product
DELETE /api/v1/:url_code/products/:id    - Delete product
```

#### Services (Feature: services)
```
GET    /api/v1/:url_code/services        - List services
GET    /api/v1/:url_code/services/:id    - Get service details
POST   /api/v1/:url_code/services        - Create service
PUT    /api/v1/:url_code/services/:id    - Update service
DELETE /api/v1/:url_code/services/:id    - Delete service
```

#### Settings
```
GET    /api/v1/:url_code/settings        - Get tenant settings
PUT    /api/v1/:url_code/settings        - Update tenant settings
```

#### User Profile Uploads
```
POST   /api/v1/:url_code/profile/avatar  - Upload user avatar (200x200)
DELETE /api/v1/:url_code/profile/avatar  - Delete user avatar
```

#### Tenant Profile Uploads
```
POST   /api/v1/:url_code/tenant/logo     - Upload tenant logo (max 120x60, SVG allowed)
DELETE /api/v1/:url_code/tenant/logo     - Delete tenant logo
```

#### System User Profile Uploads
```
POST   /api/v1/:url_code/sys-users/:user_id/avatar  - Upload sys user avatar (200x200)
DELETE /api/v1/:url_code/sys-users/:user_id/avatar  - Delete sys user avatar
```

---

## Common Response Formats

### Success Response
```json
{
  "data": { ... },
  "message": "Success message"
}
```

### Error Response
```json
{
  "error": "Error message",
  "details": "Detailed error information"
}
```

### Validation Error
```json
{
  "error": "Validation failed",
  "fields": {
    "email": "Invalid email format",
    "password": "Password must be at least 8 characters"
  }
}
```

---

## Authentication

### JWT Token in Header
```
Authorization: Bearer <token>
```

### Token Payload (Tenant User)
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "exp": 1234567890
}
```

### Token Payload (Admin User)
```json
{
  "user_id": "uuid",
  "email": "admin@example.com",
  "is_admin": true,
  "exp": 1234567890
}
```

---

## CORS Configuration

### Development Origins
- `http://localhost:3000` (React/Next.js)
- `http://localhost:5173` (Vite default)
- `http://localhost:5174` (Vite alternative)
- `http://localhost:8080` (Vue CLI)

### Allowed Methods
- `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`

### Allowed Headers
- `Origin`, `Content-Type`, `Accept`, `Authorization`, `X-Requested-With`

---

## Example Usage

### 1. Register New Subscription (Self-Service)
```bash
POST http://localhost:8081/api/v1/subscription
Content-Type: application/json

{
  "plan_id": "uuid-of-plan",
  "billing_cycle": "monthly",
  "name": "João Silva",
  "email": "joao@example.com",
  "password": "secure123",
  "subdomain": "joao"
}
```

Response includes:
- JWT Token
- User data
- Tenant data (with url_code for admin panel)

### 2. Login
```bash
POST http://localhost:8081/api/v1/auth/login
Content-Type: application/json

{
  "email": "joao@example.com",
  "password": "secure123"
}
```

Response includes:
- JWT Token
- User data
- All user's tenants with permissions

### 3. Get Tenant Configuration
```bash
GET http://localhost:8081/api/v1/27PCKWWWN3F/config
Authorization: Bearer <token>
```

### 4. Create Product (if feature enabled)
```bash
POST http://localhost:8081/api/v1/27PCKWWWN3F/products
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Premium Widget",
  "price": 99.99,
  "description": "High-quality widget"
}
```

---

## Testing

### Health Checks
```bash
# Admin API
curl http://localhost:8080/health

# Tenant API
curl http://localhost:8081/health
```

### Get Available Plans (for registration)
```bash
curl http://localhost:8081/api/v1/plans
```

### Test CORS
Open `docs/cors-test.html` in browser or use the test script:
```bash
./scripts/test-cors.sh
```
