-- Drop tenant tables
DROP TABLE IF EXISTS settings CASCADE;
DROP TABLE IF EXISTS services CASCADE;
DROP TABLE IF EXISTS products CASCADE;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
