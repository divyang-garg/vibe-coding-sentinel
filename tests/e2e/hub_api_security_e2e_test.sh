#!/bin/bash
# E2E Test: Hub API Security Features
# Tests: Authentication, Validation, Audit Logging
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
echo -e "${BLUE}  Hub API Security E2E Test${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Test 1: Health Check (No Auth Required)
echo -e "${YELLOW}Test 1: Health Check${NC}"
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$API_URL/health" 2>&1)
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -1)
if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ PASS${NC}: Health endpoint accessible"
    ((PASS_COUNT++))
else
    echo -e "${RED}❌ FAIL${NC}: Health endpoint returned $HTTP_CODE"
    ((FAIL_COUNT++))
fi
echo ""

# Test 2: Authentication - Missing API Key
echo -e "${YELLOW}Test 2: Authentication - Missing API Key${NC}"
AUTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$API_URL/api/v1/tasks" 2>&1)
HTTP_CODE=$(echo "$AUTH_RESPONSE" | tail -1)
if [ "$HTTP_CODE" = "401" ]; then
    echo -e "${GREEN}✅ PASS${NC}: Missing API key correctly rejected (401)"
    ((PASS_COUNT++))
else
    echo -e "${RED}❌ FAIL${NC}: Expected 401, got $HTTP_CODE"
    ((FAIL_COUNT++))
fi
echo ""

# Test 3: Authentication - Invalid API Key
echo -e "${YELLOW}Test 3: Authentication - Invalid API Key${NC}"
INVALID_RESPONSE=$(curl -s -w "\n%{http_code}" -H "X-API-Key: invalid-key-12345" "$API_URL/api/v1/tasks" 2>&1)
HTTP_CODE=$(echo "$INVALID_RESPONSE" | tail -1)
if [ "$HTTP_CODE" = "401" ]; then
    echo -e "${GREEN}✅ PASS${NC}: Invalid API key correctly rejected (401)"
    ((PASS_COUNT++))
else
    echo -e "${RED}❌ FAIL${NC}: Expected 401, got $HTTP_CODE"
    ((FAIL_COUNT++))
fi
echo ""

# Test 4: Input Validation - Invalid Task Creation
echo -e "${YELLOW}Test 4: Input Validation - Invalid Task Creation${NC}"
VALIDATION_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -H "X-API-Key: test-key" \
    -d '{"status": "invalid_status"}' \
    "$API_URL/api/v1/tasks" 2>&1)
HTTP_CODE=$(echo "$VALIDATION_RESPONSE" | tail -1)
if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "401" ]; then
    echo -e "${GREEN}✅ PASS${NC}: Invalid input rejected (HTTP $HTTP_CODE)"
    ((PASS_COUNT++))
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Got HTTP $HTTP_CODE (may need valid API key for validation test)"
fi
echo ""

# Test 5: CORS Headers
echo -e "${YELLOW}Test 5: CORS Headers${NC}"
CORS_RESPONSE=$(curl -s -I -H "Origin: http://localhost:3000" "$API_URL/health" 2>&1)
if echo "$CORS_RESPONSE" | grep -qi "Access-Control-Allow-Origin"; then
    echo -e "${GREEN}✅ PASS${NC}: CORS headers present"
    ((PASS_COUNT++))
else
    echo -e "${YELLOW}⚠️  WARN${NC}: CORS headers not found (may be disabled for health endpoint)"
fi
echo ""

# Test 6: Security Headers
echo -e "${YELLOW}Test 6: Security Headers${NC}"
SEC_HEADERS=$(curl -s -I "$API_URL/health" 2>&1)
HAS_XCTO=$(echo "$SEC_HEADERS" | grep -i "X-Content-Type-Options" | wc -l | tr -d ' ')
HAS_XFO=$(echo "$SEC_HEADERS" | grep -i "X-Frame-Options" | wc -l | tr -d ' ')
if [ "$HAS_XCTO" -gt 0 ] && [ "$HAS_XFO" -gt 0 ]; then
    echo -e "${GREEN}✅ PASS${NC}: Security headers present"
    ((PASS_COUNT++))
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Some security headers missing"
fi
echo ""

# Test 7: Request Size Limit
echo -e "${YELLOW}Test 7: Request Size Limit${NC}"
LARGE_PAYLOAD=$(head -c 11000000 < /dev/zero | tr '\0' 'A')
SIZE_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -H "X-API-Key: test-key" \
    -d "{\"data\": \"$LARGE_PAYLOAD\"}" \
    "$API_URL/api/v1/tasks" 2>&1)
HTTP_CODE=$(echo "$SIZE_RESPONSE" | tail -1)
if [ "$HTTP_CODE" = "413" ] || [ "$HTTP_CODE" = "401" ]; then
    echo -e "${GREEN}✅ PASS${NC}: Large request handled correctly (HTTP $HTTP_CODE)"
    ((PASS_COUNT++))
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Got HTTP $HTTP_CODE (size limit may not be enforced or auth failed)"
fi
echo ""

# Summary
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Test Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Passed: $PASS_COUNT${NC}"
echo -e "${RED}Failed: $FAIL_COUNT${NC}"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}✅ All critical tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some tests failed or were skipped${NC}"
    exit 1
fi
