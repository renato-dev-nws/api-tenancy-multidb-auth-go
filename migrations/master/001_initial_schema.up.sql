-- Master DB Schema (Control Plane)

-- Create UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ====================================
-- SECTION 1: SaaS Administrator Tables (Control Plane users)
-- ====================================

-- System users table (SaaS administrators)
CREATE TABLE sys_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'inactive')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- System roles table
CREATE TABLE sys_roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- System permissions table
CREATE TABLE sys_permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- System user roles junction table
CREATE TABLE sys_user_roles (
    sys_user_id UUID REFERENCES sys_users(id) ON DELETE CASCADE,
    sys_role_id INTEGER REFERENCES sys_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (sys_user_id, sys_role_id)
);

-- System role permissions junction table
CREATE TABLE sys_role_permissions (
    sys_role_id INTEGER REFERENCES sys_roles(id) ON DELETE CASCADE,
    sys_permission_id INTEGER REFERENCES sys_permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (sys_role_id, sys_permission_id)
);

-- ====================================
-- SECTION 2: Tenant Users Tables (Data Plane users)
-- ====================================

-- Tenant users table (users that belong to tenants)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    last_tenant_logged VARCHAR(11),
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
    code VARCHAR(10) UNIQUE NOT NULL, -- Código curto para permissões (ex: prod, serv)
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
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
CREATE TYPE billing_cycle AS ENUM ('monthly', 'quarterly', 'semiannual', 'annual');

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    db_code UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    url_code VARCHAR(11) UNIQUE NOT NULL,
    subdomain VARCHAR(50) UNIQUE NOT NULL,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL, -- Nullable: admin pode criar tenant sem owner
    plan_id UUID NOT NULL REFERENCES plans(id),
    billing_cycle billing_cycle NOT NULL DEFAULT 'monthly',
    status tenant_status NOT NULL DEFAULT 'provisioning',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tenant profiles
CREATE TABLE tenant_profiles (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    company_name VARCHAR(255),
    is_company BOOLEAN NOT NULL DEFAULT false,
    custom_domain VARCHAR(255), -- Domain customizado do cliente (ex: app.empresa.com)
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
CREATE INDEX idx_sys_users_email ON sys_users(email);
CREATE INDEX idx_sys_users_status ON sys_users(status);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_last_tenant_logged ON users(last_tenant_logged);
CREATE INDEX idx_tenants_url_code ON tenants(url_code);
CREATE INDEX idx_tenants_db_code ON tenants(db_code);
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenant_members_user_id ON tenant_members(user_id);
CREATE INDEX idx_tenant_members_tenant_id ON tenant_members(tenant_id);
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);

-- ====================================
-- SECTION 3: Seed Data
-- ====================================

-- System roles (for SaaS administrators)
INSERT INTO sys_roles (name, slug, description) VALUES
    ('Super Admin', 'super_admin', 'Full system access with all permissions'),
    ('Admin', 'admin', 'Administrative access to most features'),
    ('Support', 'support', 'Customer support access with limited permissions'),
    ('Viewer', 'viewer', 'Read-only access to system data');

-- System permissions (for SaaS administrators)
INSERT INTO sys_permissions (name, slug, description) VALUES
    ('Create Tenant', 'create_tenant', 'Can create new tenants'),
    ('View Tenants', 'view_tenants', 'Can view tenant list and details'),
    ('Update Tenant', 'update_tenant', 'Can update tenant information'),
    ('Delete Tenant', 'delete_tenant', 'Can delete tenants'),
    ('Manage Plans', 'manage_plans', 'Can manage subscription plans'),
    ('Manage System Users', 'manage_sys_users', 'Can manage SaaS administrator users'),
    ('View Analytics', 'view_analytics', 'Can view system analytics'),
    ('Manage Billing', 'manage_billing', 'Can manage billing and payments'),
    ('Access Support Tools', 'access_support_tools', 'Can access customer support tools'),
    ('View Audit Logs', 'view_audit_logs', 'Can view system audit logs');

-- Assign permissions to system roles
-- Super Admin: all permissions
INSERT INTO sys_role_permissions (sys_role_id, sys_permission_id)
SELECT 1, id FROM sys_permissions;

-- Admin: all except delete_tenant and manage_billing
INSERT INTO sys_role_permissions (sys_role_id, sys_permission_id)
SELECT 2, id FROM sys_permissions
WHERE slug NOT IN ('delete_tenant', 'manage_billing');

-- Support: view and support tools only
INSERT INTO sys_role_permissions (sys_role_id, sys_permission_id)
SELECT 3, id FROM sys_permissions
WHERE slug IN ('view_tenants', 'view_analytics', 'access_support_tools', 'view_audit_logs');

-- Viewer: read-only access
INSERT INTO sys_role_permissions (sys_role_id, sys_permission_id)
SELECT 4, id FROM sys_permissions
WHERE slug IN ('view_tenants', 'view_analytics');

-- Default SaaS administrator (initial super admin)
-- Password: admin123 (bcrypt hash with cost 12)
INSERT INTO sys_users (email, password_hash, full_name, status) VALUES
    ('admin@teste.com', '$2a$12$6qRbnes1LBvu5vXM9UGYHuifduRsnykrK.E/T.o/B0E3.n8OAuOhy', 'System Administrator', 'active');

-- Assign super_admin role to the default admin
INSERT INTO sys_user_roles (sys_user_id, sys_role_id)
SELECT id, 1 FROM sys_users WHERE email = 'admin@teste.com';

-- Default features (UUIDs fixos para facilitar)
INSERT INTO features (id, title, slug, code, description, is_active) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Products', 'products', 'prod', 'Product management module', true),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Services', 'services', 'serv', 'Service management module', true);

-- Default plans (UUIDs fixos para facilitar)
INSERT INTO plans (id, name, description, price) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Products Plan', 'Plan with product management only', 19.99),
    ('22222222-2222-2222-2222-222222222222', 'Services Plan', 'Plan with service management only', 19.99),
    ('33333333-3333-3333-3333-333333333333', 'Premium Plan', 'Full plan with products and services', 39.99);

-- Link features to plans (usando UUIDs fixos)
-- Products Plan: only products
INSERT INTO plan_features (plan_id, feature_id) VALUES
    ('11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa');

-- Services Plan: only services
INSERT INTO plan_features (plan_id, feature_id) VALUES
    ('22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');

-- Premium Plan: all features
INSERT INTO plan_features (plan_id, feature_id) VALUES
    ('33333333-3333-3333-3333-333333333333', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
    ('33333333-3333-3333-3333-333333333333', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');

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
