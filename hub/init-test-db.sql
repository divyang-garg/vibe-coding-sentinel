-- Initialize test database with proper users and permissions
CREATE ROLE sentinel WITH LOGIN PASSWORD 'sentinel';
ALTER ROLE sentinel CREATEDB;
GRANT ALL PRIVILEGES ON DATABASE sentinel_test TO sentinel;




