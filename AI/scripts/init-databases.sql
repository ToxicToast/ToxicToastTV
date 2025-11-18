-- Initialize database for ToxicToastGo monorepo
-- This script runs automatically when PostgreSQL container starts

-- Create single shared database
CREATE DATABASE toxictoast;
CREATE DATABASE keycloak;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE toxictoast TO postgres;
GRANT ALL PRIVILEGES ON DATABASE keycloak TO postgres;

-- Print confirmation
\echo 'Databases created successfully'
\echo 'All services will use toxictoast database with table prefixes'
