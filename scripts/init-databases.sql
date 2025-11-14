-- Initialize databases for all ToxicToastGo services
-- This script runs automatically when PostgreSQL container starts

-- Create databases
CREATE DATABASE blog_service;
CREATE DATABASE foodfolio_service;
CREATE DATABASE link_service;
CREATE DATABASE notification_service;
CREATE DATABASE webhook_service;
CREATE DATABASE twitchbot_service;
CREATE DATABASE keycloak;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE blog_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE foodfolio_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE link_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE notification_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE webhook_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE twitchbot_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE keycloak TO postgres;

-- Print confirmation
\echo 'All databases created successfully'
