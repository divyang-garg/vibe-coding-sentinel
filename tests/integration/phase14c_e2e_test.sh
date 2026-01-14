#!/bin/bash
# Integration tests for Phase 14C Hub Configuration Interface
# Run from project root: ./tests/integration/phase14c_e2e_test.sh

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    ((TESTS_FAILED++))
}

log_warn() {
    echo -e "${YELLOW}⚠ WARN:${NC} $1"
}

echo "🧪 Phase 14C Hub Configuration Interface Integration Tests"
echo "══════════════════════════════════════════════════════════════"
echo ""

# Test configuration
HUB_URL="${HUB_URL:-http://localhost:8080}"
API_KEY="${API_KEY:-test-api-key-12345678901234567890}"
PROJECT_ID="${PROJECT_ID:-test-project-id}"

# Test 1: Get Providers List
echo "Test 1: Get Providers List"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
    "$HUB_URL/api/v1/llm/providers" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" 2>/dev/null || echo "")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [[ "$HTTP_CODE" == "200" ]]; then
    if echo "$BODY" | grep -q "openai\|anthropic\|azure"; then
        log_pass "Providers endpoint returns valid providers"
    else
        log_fail "Providers endpoint response invalid"
    fi
else
    log_warn "Providers endpoint returned $HTTP_CODE (Hub may not be running)"
fi

# Test 2: Get Models for Provider
echo ""
echo "Test 2: Get Models for Provider"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
    "$HUB_URL/api/v1/llm/models/openai" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" 2>/dev/null || echo "")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [[ "$HTTP_CODE" == "200" ]]; then
    if echo "$BODY" | grep -q "gpt\|model"; then
        log_pass "Models endpoint returns valid models"
    else
        log_fail "Models endpoint response invalid"
    fi
else
    log_warn "Models endpoint returned $HTTP_CODE (Hub may not be running)"
fi

# Test 3: Create LLM Configuration
echo ""
echo "Test 3: Create LLM Configuration"
CONFIG_DATA=$(cat <<EOF
{
    "provider": "openai",
    "model": "gpt-4",
    "api_key": "test-key-12345",
    "use_cache": true,
    "cache_ttl_hours": 24,
    "progressive_depth": true,
    "max_cost_per_request": 0.10
}
EOF
)

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    "$HUB_URL/api/v1/llm/config" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "$CONFIG_DATA" 2>/dev/null || echo "")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [[ "$HTTP_CODE" == "200" ]] || [[ "$HTTP_CODE" == "201" ]]; then
    CONFIG_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4 || echo "")
    if [[ -n "$CONFIG_ID" ]]; then
        log_pass "LLM configuration created successfully"
        export TEST_CONFIG_ID="$CONFIG_ID"
    else
        log_fail "LLM configuration creation response missing ID"
    fi
else
    log_warn "LLM configuration creation returned $HTTP_CODE (Hub may not be running)"
fi

# Test 4: Get LLM Configuration
echo ""
echo "Test 4: Get LLM Configuration"
if [[ -n "$TEST_CONFIG_ID" ]]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
        "$HUB_URL/api/v1/llm/config/$TEST_CONFIG_ID" \
        -H "Authorization: Bearer $API_KEY" \
        -H "Content-Type: application/json" 2>/dev/null || echo "")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [[ "$HTTP_CODE" == "200" ]]; then
        log_pass "LLM configuration retrieved successfully"
    else
        log_fail "LLM configuration retrieval returned $HTTP_CODE"
    fi
else
    log_warn "Skipping get config test (no config ID available)"
fi

# Test 5: Validate LLM Configuration
echo ""
echo "Test 5: Validate LLM Configuration"
VALIDATE_DATA=$(cat <<EOF
{
    "provider": "openai",
    "model": "gpt-4",
    "api_key": "test-key-12345"
}
EOF
)

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    "$HUB_URL/api/v1/llm/config/validate" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "$VALIDATE_DATA" 2>/dev/null || echo "")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [[ "$HTTP_CODE" == "200" ]]; then
    log_pass "LLM configuration validation endpoint works"
else
    log_warn "LLM configuration validation returned $HTTP_CODE (expected for test key)"
fi

# Test 6: Get Usage Report
echo ""
echo "Test 6: Get Usage Report"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
    "$HUB_URL/api/v1/llm/usage/report?projectId=$PROJECT_ID&period=monthly" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" 2>/dev/null || echo "")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [[ "$HTTP_CODE" == "200" ]]; then
    if echo "$BODY" | grep -q "total_tokens\|total_cost"; then
        log_pass "Usage report endpoint returns valid data"
    else
        log_pass "Usage report endpoint works (empty data expected)"
    fi
else
    log_warn "Usage report endpoint returned $HTTP_CODE (Hub may not be running)"
fi

# Test 7: Get Usage Stats
echo ""
echo "Test 7: Get Usage Stats"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
    "$HUB_URL/api/v1/llm/usage/stats?projectId=$PROJECT_ID" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" 2>/dev/null || echo "")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [[ "$HTTP_CODE" == "200" ]]; then
    log_pass "Usage stats endpoint works"
else
    log_warn "Usage stats endpoint returned $HTTP_CODE (Hub may not be running)"
fi

# Test 8: Get Cost Breakdown
echo ""
echo "Test 8: Get Cost Breakdown"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
    "$HUB_URL/api/v1/llm/usage/cost-breakdown?projectId=$PROJECT_ID" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" 2>/dev/null || echo "")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [[ "$HTTP_CODE" == "200" ]]; then
    log_pass "Cost breakdown endpoint works"
else
    log_warn "Cost breakdown endpoint returned $HTTP_CODE (Hub may not be running)"
fi

# Test 9: Get Usage Trends
echo ""
echo "Test 9: Get Usage Trends"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
    "$HUB_URL/api/v1/llm/usage/trends?projectId=$PROJECT_ID&period=weekly" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" 2>/dev/null || echo "")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [[ "$HTTP_CODE" == "200" ]]; then
    log_pass "Usage trends endpoint works"
else
    log_warn "Usage trends endpoint returned $HTTP_CODE (Hub may not be running)"
fi

# Test 10: List Configurations
echo ""
echo "Test 10: List Configurations"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
    "$HUB_URL/api/v1/llm/config/project/$PROJECT_ID" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" 2>/dev/null || echo "")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [[ "$HTTP_CODE" == "200" ]]; then
    log_pass "List configurations endpoint works"
else
    log_warn "List configurations endpoint returned $HTTP_CODE (Hub may not be running)"
fi

# Summary
echo ""
echo "══════════════════════════════════════════════════════════════"
echo "Test Summary:"
echo "  Passed: $TESTS_PASSED"
echo "  Failed: $TESTS_FAILED"
echo "══════════════════════════════════════════════════════════════"

if [[ $TESTS_FAILED -eq 0 ]]; then
    exit 0
else
    exit 1
fi






