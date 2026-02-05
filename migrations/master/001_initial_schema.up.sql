-- Master DB Schema (Control Plane)

-- Create UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    last_tenant_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- User profiles
CREATE TABLE user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    full_name VARCHAR(255),
    avatar_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Features table
CREATE TABLE features (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Plans table
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Plan features junction table
CREATE TABLE plan_features (
    plan_id UUID REFERENCES plans(id) ON DELETE CASCADE,
    feature_id UUID REFERENCES features(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (plan_id, feature_id)
);

-- Tenants table
CREATE TYPE tenant_status AS ENUM ('provisioning', 'active', 'suspended');

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    db_code UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    url_code VARCHAR(11) UNIQUE NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id),
    status tenant_status NOT NULL DEFAULT 'provisioning',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tenant profiles
CREATE TABLE tenant_profiles (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    logo_url TEXT,
    custom_settings JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Roles table (tenant-scoped or global)
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, slug)
);

-- Permissions table
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Role permissions junction table
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- Tenant members (users can belong to multiple tenants)
CREATE TABLE tenant_members (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, tenant_id)
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_tenants_url_code ON tenants(url_code);
CREATE INDEX idx_tenants_db_code ON tenants(db_code);
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenant_members_user_id ON tenant_members(user_id);
CREATE INDEX idx_tenant_members_tenant_id ON tenant_members(tenant_id);
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);

-- Insert default data
-- Default features
INSERT INTO features (title, slug, description) VALUES
    ('Products', 'products', 'Product management module'),
    ('Services', 'services', 'Service management module');

-- Default plans
INSERT INTO plans (name, description, price) VALUES
    ('Products Plan', 'Plan with product management only', 19.99),
    ('Services Plan', 'Plan with service management only', 19.99),
    ('Premium Plan', 'Full plan with products and services', 39.99);

-- Link features to plans
-- Products Plan: only products
INSERT INTO plan_features (plan_id, feature_id)
SELECT p.id, f.id FROM plans p, features f
WHERE p.name = 'Products Plan' AND f.slug = 'products';

-- Services Plan: only services
INSERT INTO plan_features (plan_id, feature_id)
SELECT p.id, f.id FROM plans p, features f
WHERE p.name = 'Services Plan' AND f.slug = 'services';

-- Premium Plan: all features
INSERT INTO plan_features (plan_id, feature_id)
SELECT p.id, f.id FROM plans p, features f
WHERE p.name = 'Premium Plan';

-- Default permissions
INSERT INTO permissions (name, slug, description) VALUES
    ('Create Product', 'create_product', 'Can create products'),
    ('Read Product', 'read_product', 'Can read products'),
    ('Update Product', 'update_product', 'Can update products'),
    ('Delete Product', 'delete_product', 'Can delete products'),
    ('Create Service', 'create_service', 'Can create services'),
    ('Read Service', 'read_service', 'Can read services'),
    ('Update Service', 'update_service', 'Can update services'),
    ('Delete Service', 'delete_service', 'Can delete services'),
    ('Manage Users', 'manage_users', 'Can manage tenant users'),
    ('Manage Settings', 'manage_settings', 'Can manage tenant settings');

-- Default global roles
INSERT INTO roles (tenant_id, name, slug) VALUES
    (NULL, 'Global Admin', 'global_admin'),
    (NULL, 'Owner', 'owner'),
    (NULL, 'Admin', 'admin'),
    (NULL, 'Member', 'member');

-- Link all permissions to global admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.slug = 'global_admin';
