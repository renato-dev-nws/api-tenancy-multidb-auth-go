-- Drop tenant tables (reverse order due to foreign keys)
DROP TABLE IF EXISTS images CASCADE;
DROP TABLE IF EXISTS settings CASCADE;
DROP TABLE IF EXISTS services CASCADE;
DROP TABLE IF EXISTS products CASCADE;

-- Drop ENUMs
DROP TYPE IF EXISTS image_variant;
DROP TYPE IF EXISTS media_type;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
