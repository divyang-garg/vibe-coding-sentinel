#!/bin/bash

# Test Database Setup Script
# Initializes test database for integration tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
DB_NAME="${TEST_DB_NAME:-sentinel_test}"
DB_USER="${TEST_DB_USER:-sentinel}"
DB_PASSWORD="${TEST_DB_PASSWORD:-sentinel}"
DB_HOST="${TEST_DB_HOST:-localhost}"
DB_PORT="${TEST_DB_PORT:-5432}"

echo -e "${GREEN}Setting up test database...${NC}"

# Check if PostgreSQL is available
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql command not found. Please install PostgreSQL client tools.${NC}"
    exit 1
fi

# Create database if it doesn't exist
echo -e "${YELLOW}Creating test database if it doesn't exist...${NC}"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || \
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME"

# Set connection string
export TEST_DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

echo -e "${GREEN}Test database setup complete!${NC}"
echo -e "${YELLOW}Database URL: $TEST_DATABASE_URL${NC}"
echo ""
echo -e "${YELLOW}Note: Run migrations separately using the Hub API migrations.${NC}"
echo -e "${YELLOW}The test database will be populated with schema when tests run.${NC}"









