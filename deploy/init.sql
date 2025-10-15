-- Initialize Bazaruto database
-- This script runs when the PostgreSQL container starts for the first time

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create database if it doesn't exist (this is handled by POSTGRES_DB)
-- The database 'bazaruto' is created automatically by the postgres container

-- Set timezone
SET timezone = 'UTC';

-- Create a read-only user for monitoring
CREATE USER bazaruto_readonly WITH PASSWORD 'readonly_password';
GRANT CONNECT ON DATABASE bazaruto TO bazaruto_readonly;
GRANT USAGE ON SCHEMA public TO bazaruto_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO bazaruto_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO bazaruto_readonly;
