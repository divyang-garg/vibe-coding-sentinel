#!/bin/bash
# Test Script: API Key Generation and Authentication Security
# Tests the security remediation implementation

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  API Key Security Testing - Hash-Based Storage Verification${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Configuration
DB_CONTAINER="hub-db-1"
API_CONTAINER="hub-api-1"
DB_NAME="sentinel"
DB_USER="sentinel"

# Test 1: Verify Database Schema
echo -e "${YELLOW}Test 1: Verifying Database Schema${NC}"
echo "───────────────────────────────────────────────────────────────"

SCHEMA_CHECK=$(docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -t -c "
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'projects' 
            AND column_name = 'api_key_hash'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as hash_column,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'projects' 
            AND column_name = 'api_key_prefix'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as prefix_column,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM pg_indexes 
            WHERE indexname = 'idx_projects_api_key_hash'
        ) THEN 'PASS'
        ELSE 'FAIL'
    END as hash_index
;" 2>&1)

if echo "$SCHEMA_CHECK" | grep -q "PASS.*PASS.*PASS"; then
    echo -e "${GREEN}✅ PASS${NC}: Database schema correct (api_key_hash, api_key_prefix, indexes)"
else
    echo -e "${RED}❌ FAIL${NC}: Database schema incomplete"
    echo "$SCHEMA_CHECK"
    exit 1
fi
echo ""

# Test 2: Check for Existing Projects
echo -e "${YELLOW}Test 2: Checking Existing Projects${NC}"
echo "───────────────────────────────────────────────────────────────"

PROJECT_COUNT=$(docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM projects;" 2>&1 | tr -d ' ')

echo "Found $PROJECT_COUNT projects in database"

if [ "$PROJECT_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Checking existing API keys...${NC}"
    
    EXISTING_KEYS=$(docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -t -c "
    SELECT 
        COUNT(*) FILTER (WHERE api_key IS NOT NULL AND api_key != '') as plaintext_count,
        COUNT(*) FILTER (WHERE api_key_hash IS NOT NULL) as hashed_count
    FROM projects;
    " 2>&1 | tr -d ' ')
    
    echo "Existing keys: $EXISTING_KEYS"
fi
echo ""

# Test 3: Verify Service Code Implementation
echo -e "${YELLOW}Test 3: Verifying Service Implementation${NC}"
echo "───────────────────────────────────────────────────────────────"

if grep -q "hashAPIKey" hub/api/services/organization_service_api_keys.go; then
    echo -e "${GREEN}✅ PASS${NC}: hashAPIKey() method exists"
else
    echo -e "${RED}❌ FAIL${NC}: hashAPIKey() method not found"
    exit 1
fi

if grep -q "FindByAPIKeyHash" hub/api/repository/organization_repository.go; then
    echo -e "${GREEN}✅ PASS${NC}: FindByAPIKeyHash() repository method exists"
else
    echo -e "${RED}❌ FAIL${NC}: FindByAPIKeyHash() method not found"
    exit 1
fi

if grep -q "APIKeyHash" hub/api/models/organization.go; then
    echo -e "${GREEN}✅ PASS${NC}: APIKeyHash field in Project model"
else
    echo -e "${RED}❌ FAIL${NC}: APIKeyHash field not found in model"
    exit 1
fi
echo ""

# Test 4: Check Middleware Integration
echo -e "${YELLOW}Test 4: Verifying Middleware Integration${NC}"
echo "───────────────────────────────────────────────────────────────"

if grep -q "OrganizationService" hub/api/middleware/security.go; then
    echo -e "${GREEN}✅ PASS${NC}: Middleware uses OrganizationService"
else
    echo -e "${RED}❌ FAIL${NC}: Middleware not integrated with service"
    exit 1
fi

if grep -q "ValidateAPIKey" hub/api/middleware/security.go; then
    echo -e "${GREEN}✅ PASS${NC}: Middleware calls ValidateAPIKey()"
else
    echo -e "${RED}❌ FAIL${NC}: Middleware doesn't validate API keys"
    exit 1
fi
echo ""

# Test 5: Check API Container Logs
echo -e "${YELLOW}Test 5: Checking API Container Status${NC}"
echo "───────────────────────────────────────────────────────────────"

if docker ps | grep -q "$API_CONTAINER.*healthy"; then
    echo -e "${GREEN}✅ PASS${NC}: API container is running and healthy"
else
    echo -e "${YELLOW}⚠️  WARN${NC}: API container status unclear"
    docker ps | grep "$API_CONTAINER" || echo "Container not found"
fi

# Check for recent errors
RECENT_ERRORS=$(docker logs $API_CONTAINER --since 5m 2>&1 | grep -i "error\|panic\|fatal" | wc -l | tr -d ' ')
if [ "$RECENT_ERRORS" -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC}: No recent errors in API logs"
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Found $RECENT_ERRORS recent errors in logs"
    echo "Recent errors:"
    docker logs $API_CONTAINER --since 5m 2>&1 | grep -i "error\|panic\|fatal" | tail -5
fi
echo ""

# Test 6: Database Index Verification
echo -e "${YELLOW}Test 6: Verifying Database Indexes${NC}"
echo "───────────────────────────────────────────────────────────────"

INDEXES=$(docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -t -c "
SELECT indexname 
FROM pg_indexes 
WHERE tablename = 'projects' 
AND indexname IN ('idx_projects_api_key_hash', 'idx_projects_api_key_prefix');
" 2>&1 | tr -d ' ')

if echo "$INDEXES" | grep -q "idx_projects_api_key_hash" && echo "$INDEXES" | grep -q "idx_projects_api_key_prefix"; then
    echo -e "${GREEN}✅ PASS${NC}: Both indexes exist"
    echo "$INDEXES"
else
    echo -e "${RED}❌ FAIL${NC}: Missing indexes"
    echo "$INDEXES"
    exit 1
fi
echo ""

# Test 7: Verify No Plaintext Keys in New Records
echo -e "${YELLOW}Test 7: Security Verification${NC}"
echo "───────────────────────────────────────────────────────────────"

# Check if any new records have plaintext but no hash (security issue)
INSECURE_RECORDS=$(docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -t -c "
SELECT COUNT(*) 
FROM projects 
WHERE api_key IS NOT NULL 
  AND api_key != '' 
  AND (api_key_hash IS NULL OR api_key_hash = '');
" 2>&1 | tr -d ' ')

if [ "$INSECURE_RECORDS" -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC}: No insecure records (all keys have hashes or are empty)"
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Found $INSECURE_RECORDS records with plaintext keys but no hash"
    echo "These may be legacy records that need migration"
fi
echo ""

# Summary
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Test Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ Database schema: Verified${NC}"
echo -e "${GREEN}✅ Service implementation: Verified${NC}"
echo -e "${GREEN}✅ Middleware integration: Verified${NC}"
echo -e "${GREEN}✅ Database indexes: Verified${NC}"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Create a test project via API"
echo "2. Generate an API key for the project"
echo "3. Verify key is stored as hash in database"
echo "4. Test authentication with the generated key"
echo "5. Monitor logs during authentication"
echo ""
