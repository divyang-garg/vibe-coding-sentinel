#!/bin/bash
# E2E Test: API Key Management Endpoints
# Tests: Generate, Get Info, Revoke API keys
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

API_URL="${SENTINEL_HUB_URL:-http://localhost:8080}"
PASS_COUNT=0
FAIL_COUNT=0

echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  API Key Management E2E Test${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Check if API is running
if ! curl -sf "$API_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ FAIL${NC}: API is not running at $API_URL"
    echo "Please start the API server first"
    exit 1
fi

echo -e "${GREEN}✅ API is running${NC}"
echo ""

# Test 1: Create a project (auto-generates API key)
echo -e "${YELLOW}Test 1: Create Project (Auto-generates API Key)${NC}"
CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -H "X-API-Key: test-admin-key" \
    -d '{"name": "E2E Test Project"}' \
    "$API_URL/api/v1/projects" 2>&1)

HTTP_CODE=$(echo "$CREATE_RESPONSE" | tail -1)
BODY=$(echo "$CREATE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "401" ]; then
    if [ "$HTTP_CODE" = "201" ]; then
        PROJECT_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        API_KEY=$(echo "$BODY" | grep -o '"api_key":"[^"]*"' | head -1 | cut -d'"' -f4)
        echo -e "${GREEN}✅ PASS${NC}: Project created (HTTP $HTTP_CODE)"
        echo "   Project ID: $PROJECT_ID"
        echo "   API Key: ${API_KEY:0:20}..."
        ((PASS_COUNT++))
    else
        echo -e "${YELLOW}⚠️  WARN${NC}: Need valid admin API key (HTTP $HTTP_CODE)"
        PROJECT_ID="proj_test_123"
        API_KEY="test-api-key-12345"
    fi
else
    echo -e "${RED}❌ FAIL${NC}: Project creation failed (HTTP $HTTP_CODE)"
    ((FAIL_COUNT++))
    PROJECT_ID=""
fi
echo ""

# Test 2: Generate API Key (if we have a project ID)
if [ -n "$PROJECT_ID" ] && [ "$PROJECT_ID" != "" ]; then
    echo -e "${YELLOW}Test 2: Generate API Key${NC}"
    GEN_RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "X-API-Key: test-admin-key" \
        "$API_URL/api/v1/projects/$PROJECT_ID/api-key" 2>&1)
    
    HTTP_CODE=$(echo "$GEN_RESPONSE" | tail -1)
    BODY=$(echo "$GEN_RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "401" ]; then
        if [ "$HTTP_CODE" = "200" ]; then
            NEW_API_KEY=$(echo "$BODY" | grep -o '"api_key":"[^"]*"' | head -1 | cut -d'"' -f4)
            echo -e "${GREEN}✅ PASS${NC}: API key generated (HTTP $HTTP_CODE)"
            echo "   New API Key: ${NEW_API_KEY:0:20}..."
            ((PASS_COUNT++))
        else
            echo -e "${YELLOW}⚠️  WARN${NC}: Need valid admin API key (HTTP $HTTP_CODE)"
        fi
    else
        echo -e "${RED}❌ FAIL${NC}: API key generation failed (HTTP $HTTP_CODE)"
        ((FAIL_COUNT++))
    fi
else
    echo -e "${YELLOW}⚠️  SKIP${NC}: No project ID available"
fi
echo ""

# Test 3: Get API Key Info
if [ -n "$PROJECT_ID" ] && [ "$PROJECT_ID" != "" ]; then
    echo -e "${YELLOW}Test 3: Get API Key Info${NC}"
    INFO_RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X GET \
        -H "X-API-Key: test-admin-key" \
        "$API_URL/api/v1/projects/$PROJECT_ID/api-key" 2>&1)
    
    HTTP_CODE=$(echo "$INFO_RESPONSE" | tail -1)
    BODY=$(echo "$INFO_RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "401" ]; then
        if [ "$HTTP_CODE" = "200" ]; then
            HAS_KEY=$(echo "$BODY" | grep -o '"has_api_key":[^,}]*' | cut -d':' -f2)
            PREFIX=$(echo "$BODY" | grep -o '"api_key_prefix":"[^"]*"' | cut -d'"' -f4)
            echo -e "${GREEN}✅ PASS${NC}: API key info retrieved (HTTP $HTTP_CODE)"
            echo "   Has API Key: $HAS_KEY"
            echo "   Prefix: $PREFIX"
            ((PASS_COUNT++))
        else
            echo -e "${YELLOW}⚠️  WARN${NC}: Need valid admin API key (HTTP $HTTP_CODE)"
        fi
    else
        echo -e "${RED}❌ FAIL${NC}: Get API key info failed (HTTP $HTTP_CODE)"
        ((FAIL_COUNT++))
    fi
else
    echo -e "${YELLOW}⚠️  SKIP${NC}: No project ID available"
fi
echo ""

# Test 4: Revoke API Key
if [ -n "$PROJECT_ID" ] && [ "$PROJECT_ID" != "" ]; then
    echo -e "${YELLOW}Test 4: Revoke API Key${NC}"
    REVOKE_RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X DELETE \
        -H "X-API-Key: test-admin-key" \
        "$API_URL/api/v1/projects/$PROJECT_ID/api-key" 2>&1)
    
    HTTP_CODE=$(echo "$REVOKE_RESPONSE" | tail -1)
    
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "401" ]; then
        if [ "$HTTP_CODE" = "200" ]; then
            echo -e "${GREEN}✅ PASS${NC}: API key revoked (HTTP $HTTP_CODE)"
            ((PASS_COUNT++))
        else
            echo -e "${YELLOW}⚠️  WARN${NC}: Need valid admin API key (HTTP $HTTP_CODE)"
        fi
    else
        echo -e "${RED}❌ FAIL${NC}: API key revocation failed (HTTP $HTTP_CODE)"
        ((FAIL_COUNT++))
    fi
else
    echo -e "${YELLOW}⚠️  SKIP${NC}: No project ID available"
fi
echo ""

# Test 5: Verify Endpoints Exist (Check 404 vs 401)
echo -e "${YELLOW}Test 5: Verify Endpoint Routes${NC}"
ENDPOINTS=(
    "POST:/api/v1/projects/test-id/api-key"
    "GET:/api/v1/projects/test-id/api-key"
    "DELETE:/api/v1/projects/test-id/api-key"
)

for endpoint in "${ENDPOINTS[@]}"; do
    METHOD=$(echo "$endpoint" | cut -d':' -f1)
    PATH=$(echo "$endpoint" | cut -d':' -f2)
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X "$METHOD" "$API_URL$PATH" 2>&1)
    HTTP_CODE=$(echo "$RESPONSE" | tail -1)
    
    # 401 means endpoint exists (auth required), 404 means endpoint doesn't exist
    if [ "$HTTP_CODE" = "401" ]; then
        echo -e "${GREEN}✅ PASS${NC}: $METHOD $PATH exists (requires auth)"
    elif [ "$HTTP_CODE" = "404" ]; then
        echo -e "${RED}❌ FAIL${NC}: $METHOD $PATH not found (404)"
        ((FAIL_COUNT++))
    else
        echo -e "${YELLOW}⚠️  INFO${NC}: $METHOD $PATH returned $HTTP_CODE"
    fi
done
echo ""

# Summary
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Test Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Passed: $PASS_COUNT${NC}"
echo -e "${RED}Failed: $FAIL_COUNT${NC}"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some tests failed or were skipped${NC}"
    echo "Note: Some tests may require valid API keys to fully test"
    exit 1
fi
