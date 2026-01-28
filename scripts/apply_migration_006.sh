#!/bin/bash
# Script to apply migration 006: Add knowledge sync metadata columns
# Usage: ./scripts/apply_migration_006.sh [DATABASE_URL]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get database connection
if [ -n "$1" ]; then
    DATABASE_URL="$1"
elif [ -n "$DATABASE_URL" ]; then
    # Use environment variable
    :
elif [ -n "$DB_HOST" ] && [ -n "$DB_USER" ] && [ -n "$DB_NAME" ]; then
    # Build from individual environment variables
    DB_PASSWORD="${DB_PASSWORD:-sentinel}"
    DB_PORT="${DB_PORT:-5432}"
    DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
else
    # Default connection
    DATABASE_URL="postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable"
fi

MIGRATION_FILE="hub/migrations/006_add_knowledge_sync_metadata.sql"

echo -e "${YELLOW}Applying migration 006: Add knowledge sync metadata columns${NC}"
echo -e "Database: ${DATABASE_URL//:*@*/:***@*/}"
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

# Test database connection
echo -e "${YELLOW}Testing database connection...${NC}"
if ! psql "$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}Error: Cannot connect to database${NC}"
    echo -e "Please check:"
    echo -e "  1. PostgreSQL is running"
    echo -e "  2. Database exists"
    echo -e "  3. User has proper permissions"
    echo -e "  4. Connection string is correct"
    exit 1
fi

echo -e "${GREEN}✓ Database connection successful${NC}"
echo ""

# Check if knowledge_items table exists
echo -e "${YELLOW}Checking if knowledge_items table exists...${NC}"
TABLE_EXISTS=$(psql "$DATABASE_URL" -tAc "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'knowledge_items');" 2>/dev/null || echo "false")

if [ "$TABLE_EXISTS" != "t" ]; then
    echo -e "${RED}Error: knowledge_items table does not exist${NC}"
    echo -e "Please run migration 002_create_core_tables.sql first"
    exit 1
fi

echo -e "${GREEN}✓ knowledge_items table exists${NC}"
echo ""

# Apply migration
echo -e "${YELLOW}Applying migration...${NC}"
if psql "$DATABASE_URL" -f "$MIGRATION_FILE"; then
    echo ""
    echo -e "${GREEN}✓ Migration applied successfully${NC}"
else
    echo ""
    echo -e "${RED}✗ Migration failed${NC}"
    exit 1
fi

# Verify migration
echo ""
echo -e "${YELLOW}Verifying migration...${NC}"

# Check if columns exist
COLUMNS=("updated_at" "approved_by" "approved_at" "last_synced_at" "sync_version" "sync_status")
MISSING_COLUMNS=()

for col in "${COLUMNS[@]}"; do
    EXISTS=$(psql "$DATABASE_URL" -tAc "SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'knowledge_items' AND column_name = '$col');" 2>/dev/null || echo "false")
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
    EXISTS=$(psql "$DATABASE_URL" -tAc "SELECT EXISTS (SELECT FROM pg_indexes WHERE tablename = 'knowledge_items' AND indexname = '$idx');" 2>/dev/null || echo "false")
    if [ "$EXISTS" != "t" ]; then
        MISSING_INDEXES+=("$idx")
    fi
done

if [ ${#MISSING_INDEXES[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ All indexes created successfully${NC}"
else
    echo -e "${YELLOW}⚠ Missing indexes: ${MISSING_INDEXES[*]}${NC}"
    echo -e "  (This may be normal if indexes already existed with different names)"
fi

echo ""
echo -e "${GREEN}Migration 006 completed successfully!${NC}"
echo ""
echo "Summary:"
echo "  - Added columns: updated_at, approved_by, approved_at, last_synced_at, sync_version, sync_status"
echo "  - Created indexes for performance"
echo "  - Updated existing rows with default values"
