-- Tenant Database Schema Template

-- Create UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Products table
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    sku VARCHAR(100),
    stock INTEGER NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Services table
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    duration_minutes INTEGER,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Settings table
CREATE TABLE settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert default interface settings
INSERT INTO settings (key, value) VALUES 
('interface', '{"logo": null, "primary_color": "#003388", "secondary_color": "#DDDDDD"}');

-- Images table (Polymorphic Association)
CREATE TYPE media_type AS ENUM ('image', 'video', 'document');
CREATE TYPE image_variant AS ENUM ('original', 'medium', 'small', 'thumb');

CREATE TABLE images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    imageable_type VARCHAR(50) NOT NULL,
    imageable_id UUID NOT NULL,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255),
    title VARCHAR(255),
    alt_text VARCHAR(255),
    media_type media_type NOT NULL DEFAULT 'image',
    mime_type VARCHAR(100) NOT NULL,
    extension VARCHAR(10) NOT NULL,
    variant image_variant NOT NULL DEFAULT 'original',
    parent_id UUID REFERENCES images(id) ON DELETE CASCADE,
    width INTEGER,
    height INTEGER,
    file_size BIGINT,
    storage_driver VARCHAR(20) NOT NULL DEFAULT 'local',
    storage_path TEXT NOT NULL,
    public_url TEXT,
    processing_status VARCHAR(20) DEFAULT 'pending',
    processed_at TIMESTAMP,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_active ON products(active);
CREATE INDEX idx_services_active ON services(active);
CREATE INDEX idx_images_imageable ON images(imageable_type, imageable_id);
CREATE INDEX idx_images_variant ON images(variant);
CREATE INDEX idx_images_parent ON images(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_images_status ON images(processing_status);
CREATE INDEX idx_images_display_order ON images(imageable_type, imageable_id, display_order);
