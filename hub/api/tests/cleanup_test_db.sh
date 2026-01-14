#!/bin/bash

# Test Database Cleanup Script
# Cleans test data from test database

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

echo -e "${YELLOW}Cleaning test database...${NC}"

# Check if PostgreSQL is available
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql command not found. Please install PostgreSQL client tools.${NC}"
    exit 1
fi

# Tables to clean (in order to respect foreign keys)
TABLES=(
    "task_links"
    "task_verifications"
    "task_dependencies"
    "tasks"
    "change_requests"
    "test_requirements"
    "comprehensive_validations"
    "knowledge_items"
    "documents"
    "projects"
)

# Truncate tables (faster than DELETE, resets sequences)
echo -e "${YELLOW}Truncating test tables...${NC}"
for table in "${TABLES[@]}"; do
    echo -e "  Truncating $table..."
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "TRUNCATE TABLE $table CASCADE;" 2>/dev/null || \
        echo -e "  ${YELLOW}Warning: Table $table does not exist or cannot be truncated${NC}"
done

echo -e "${GREEN}Test database cleanup complete!${NC}"









