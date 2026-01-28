#!/bin/bash
# Script to apply migration 006 to Docker database
# Usage: ./scripts/apply_migration_006_docker.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

MIGRATION_FILE="hub/migrations/006_add_knowledge_sync_metadata.sql"

# Get password from .env file
if [ -f "hub/.env" ]; then
    DB_PASSWORD=$(grep "^DB_PASSWORD=" hub/.env | cut -d '=' -f2- | tr -d '"' | tr -d "'")
elif [ -f ".env" ]; then
    DB_PASSWORD=$(grep "^DB_PASSWORD=" .env | cut -d '=' -f2- | tr -d '"' | tr -d "'")
else
    echo -e "${YELLOW}Warning: No .env file found, using default password${NC}"
    DB_PASSWORD="password"
fi

# Docker database connection
DB_HOST="127.0.0.1"
DB_PORT="5433"  # Docker port from hub/docker-compose.yml
DB_USER="sentinel"
DB_NAME="sentinel"

echo -e "${YELLOW}Applying migration 006 to Docker database${NC}"
echo -e "Container: hub-db-1 (postgres:15-alpine)"
echo -e "Connection: ${DB_HOST}:${DB_PORT}"
echo -e "Database: ${DB_NAME}"
echo -e "Migration file: ${MIGRATION_FILE}"
echo ""

# Check if migration file exists
if [ ! -f "$MIGRATION_FILE" ]; then
    echo -e "${RED}Error: Migration file not found: ${MIGRATION_FILE}${NC}"
    exit 1
fi

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql command not found. Please install PostgreSQL client tools.${NC}"
    exit 1
fi

# Check if Docker container is running
if ! docker ps --format "{{.Names}}" | grep -q "hub-db-1"; then
    echo -e "${RED}Error: Docker container 'hub-db-1' is not running${NC}"
    echo -e "Start it with: cd hub && docker-compose up -d db"
    exit 1
fi

echo -e "${GREEN}✓ Docker container is running${NC}"
echo ""

# Test database connection
echo -e "${YELLOW}Testing database connection...${NC}"
if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}Error: Cannot connect to database${NC}"
    echo -e "Please check:"
    echo -e "  1. Docker container is running: docker ps | grep hub-db-1"
    echo -e "  2. Password in .env file is correct"
    echo -e "  3. Port 5433 is accessible"
    exit 1
fi

echo -e "${GREEN}✓ Database connection successful${NC}"
echo ""

# Check if knowledge_items table exists
echo -e "${YELLOW}Checking if knowledge_items table exists...${NC}"
TABLE_EXISTS=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'knowledge_items');" 2>/dev/null || echo "false")

if [ "$TABLE_EXISTS" != "t" ]; then
    echo -e "${RED}Error: knowledge_items table does not exist${NC}"
    echo -e "Please run migration 002_create_core_tables.sql first"
    exit 1
fi

echo -e "${GREEN}✓ knowledge_items table exists${NC}"
echo ""

# Apply migration
echo -e "${YELLOW}Applying migration...${NC}"
if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATION_FILE"; then
    echo ""
    echo -e "${GREEN}✓ Migration applied successfully${NC}"
else
    echo ""
    echo -e "${RED}✗ Migration failed${NC}"
    exit 1
fi

# Fix the type_status index (column is 'type' not 'item_type')
echo -e "${YELLOW}Fixing type_status index...${NC}"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "CREATE INDEX IF NOT EXISTS idx_knowledge_type_status ON knowledge_items(type, status);" > /dev/null 2>&1
echo -e "${GREEN}✓ Index fixed${NC}"
echo ""

# Verify migration
echo -e "${YELLOW}Verifying migration...${NC}"

# Check if columns exist
COLUMNS=("updated_at" "approved_by" "approved_at" "last_synced_at" "sync_version" "sync_status")
MISSING_COLUMNS=()

for col in "${COLUMNS[@]}"; do
    EXISTS=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'knowledge_items' AND column_name = '$col');" 2>/dev/null || echo "false")
    if [ "$EXISTS" != "t" ]; then
        MISSING_COLUMNS+=("$col")
    fi
done

if [ ${#MISSING_COLUMNS[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ All columns created successfully${NC}"
else
    echo -e "${RED}✗ Missing columns: ${MISSING_COLUMNS[*]}${NC}"
    exit 1
fi

# Check if indexes exist
INDEXES=("idx_knowledge_status" "idx_knowledge_type_status" "idx_knowledge_sync_status" "idx_knowledge_last_synced")
MISSING_INDEXES=()

for idx in "${INDEXES[@]}"; do
    EXISTS=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT EXISTS (SELECT FROM pg_indexes WHERE tablename = 'knowledge_items' AND indexname = '$idx');" 2>/dev/null || echo "false")
    if [ "$EXISTS" != "t" ]; then
        MISSING_INDEXES+=("$idx")
    fi
done

if [ ${#MISSING_INDEXES[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ All indexes created successfully${NC}"
else
    echo -e "${YELLOW}⚠ Missing indexes: ${MISSING_INDEXES[*]}${NC}"
fi

echo ""
echo -e "${GREEN}Migration 006 completed successfully!${NC}"
echo ""
echo "Summary:"
echo "  - Added columns: updated_at, approved_by, approved_at, last_synced_at, sync_version, sync_status"
echo "  - Created indexes for performance"
echo "  - Updated existing rows with default values"
echo ""
echo "Database: Docker container 'hub-db-1' on port 5433"
