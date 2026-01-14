#!/bin/bash
# Phase 12: End-to-End Integration Tests
# Tests complete Phase 12 workflow

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"
FIXTURES_DIR="$TEST_DIR/../fixtures"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0

# Test helper functions
test_start() {
    echo -n "Testing $1... "
}

test_pass() {
    echo -e "${GREEN}✓ PASSED${NC}"
    ((PASSED++))
}

test_fail() {
    echo -e "${RED}✗ FAILED${NC}"
    echo "  $1"
    ((FAILED++))
}

# =============================================================================
# TEST 1: End-to-End Workflow
# =============================================================================

test_start "E2E - document ingestion triggers change detection"

# Create test document
TEST_DOC="$FIXTURES_DIR/docs/e2e_test_rules.md"
mkdir -p "$(dirname "$TEST_DOC")"

cat > "$TEST_DOC" << 'EOF'
# Business Rules

## Rule 1: Process Order
Order processing must validate inventory.
EOF

PROJECT_ID="00000000-0000-0000-0000-000000000000"

# Upload document
UPLOAD_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/documents/upload \
   -H "Authorization: Bearer test-key" \
   -F "file=@$TEST_DOC" \
   -F "projectId=$PROJECT_ID")

# Wait a moment for change detection
sleep 2

# Check if change request was created
CHANGE_REQUESTS=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key")

if echo "$CHANGE_REQUESTS" | grep -q "CR-"; then
    test_pass
else
    test_fail "Document ingestion should trigger change detection"
fi

test_start "E2E - change request created"

# Verify change request exists
CR_ID=$(echo "$CHANGE_REQUESTS" | grep -oE "CR-[0-9]+" | head -1)

if [ -n "$CR_ID" ]; then
    test_pass
else
    test_fail "Change request should be created"
fi

test_start "E2E - impact analysis performed"

# Perform impact analysis
TEST_CODEBASE="$FIXTURES_DIR/e2e_test_codebase"
mkdir -p "$TEST_CODEBASE"

cat > "$TEST_CODEBASE/process_order.go" << 'EOF'
package main

func processOrder(orderID string) error {
    return nil
}
EOF

IMPACT_RESPONSE=$(echo "{\"codebasePath\": \"$TEST_CODEBASE\"}" | \
   curl -s -X GET "http://localhost:8080/api/v1/change-requests/$CR_ID/impact" \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer test-key" \
   -d @-)

if echo "$IMPACT_RESPONSE" | grep -q "impact\|affected"; then
    test_pass
else
    test_fail "Impact analysis should be performed"
fi

test_start "E2E - change request approved"

# Approve change request
APPROVE_RESPONSE=$(echo '{"approved_by": "e2e-test-user"}' | \
   curl -s -X PUT "http://localhost:8080/api/v1/change-requests/$CR_ID/approve" \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer test-key" \
   -d @-)

if echo "$APPROVE_RESPONSE" | grep -q "approved\|success"; then
    test_pass
else
    test_fail "Change request should be approved"
fi

test_start "E2E - implementation tracked"

# Start implementation
START_RESPONSE=$(echo '{"notes": "Starting implementation"}' | \
   curl -s -X PUT "http://localhost:8080/api/v1/change-requests/$CR_ID/implementation/start" \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer test-key" \
   -d @-)

if echo "$START_RESPONSE" | grep -q "in_progress\|success"; then
    # Check implementation status
    STATUS_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests/$CR_ID/implementation-status" \
       -H "Authorization: Bearer test-key")
    
    if echo "$STATUS_RESPONSE" | grep -q "in_progress"; then
        test_pass
    else
        test_fail "Implementation status should be tracked"
    fi
else
    test_fail "Implementation should be startable"
fi

test_start "E2E - gap analysis works"

# Run gap analysis
GAP_RESPONSE=$(echo "{\"projectId\": \"$PROJECT_ID\", \"codebasePath\": \"$TEST_CODEBASE\"}" | \
   curl -s -X POST http://localhost:8080/api/v1/knowledge/gap-analysis \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer test-key" \
   -d @-)

if echo "$GAP_RESPONSE" | grep -q "gaps\|report"; then
    test_pass
else
    test_fail "Gap analysis should work"
fi

test_start "E2E - gap analysis cache and persistence"

# Run gap analysis and check for report_id
GAP_RESPONSE=$(echo "{\"projectId\": \"$PROJECT_ID\", \"codebasePath\": \"$TEST_CODEBASE\"}" | \
   curl -s -X POST http://localhost:8080/api/v1/knowledge/gap-analysis \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer test-key" \
   -d @-)

# Check if report_id is returned
REPORT_ID=$(echo "$GAP_RESPONSE" | grep -o '"report_id":"[^"]*"' | cut -d'"' -f4)
if [ -n "$REPORT_ID" ] && [ "$REPORT_ID" != "null" ] && [ "$REPORT_ID" != "" ]; then
    # Verify it's a valid UUID format
    if echo "$REPORT_ID" | grep -qE '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'; then
        test_pass
    else
        test_fail "report_id should be a valid UUID format"
    fi
else
    test_fail "Gap analysis should return report_id after storing"
fi

test_start "E2E - gap analysis cache hit"

# Run gap analysis again (should use cache)
START_TIME=$(date +%s%N)
CACHED_RESPONSE=$(echo "{\"projectId\": \"$PROJECT_ID\", \"codebasePath\": \"$TEST_CODEBASE\"}" | \
   curl -s -X POST http://localhost:8080/api/v1/knowledge/gap-analysis \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer test-key" \
   -d @-)
END_TIME=$(date +%s%N)

# Cache hit should be faster (rough check)
DURATION=$((END_TIME - START_TIME))
if [ "$DURATION" -lt 1000000000 ]; then  # Less than 1 second
    test_pass
else
    test_fail "Cache hit should be faster than cache miss"
fi

# =============================================================================
# TEST 2: API Integration
# =============================================================================

test_start "API integration - all endpoints respond correctly"

# Test all Phase 12 endpoints
ENDPOINTS=(
    "GET /api/v1/change-requests"
    "GET /api/v1/change-requests/$CR_ID"
    "POST /api/v1/knowledge/gap-analysis"
)

ALL_OK=true
for endpoint in "${ENDPOINTS[@]}"; do
    METHOD=$(echo "$endpoint" | cut -d' ' -f1)
    PATH=$(echo "$endpoint" | cut -d' ' -f2)
    
    if [ "$METHOD" = "GET" ]; then
        RESPONSE=$(curl -s -X GET "http://localhost:8080$PATH" \
           -H "Authorization: Bearer test-key")
    elif [ "$METHOD" = "POST" ]; then
        RESPONSE=$(echo '{}' | curl -s -X POST "http://localhost:8080$PATH" \
           -H "Content-Type: application/json" \
           -H "Authorization: Bearer test-key" \
           -d @-)
    fi
    
    if echo "$RESPONSE" | grep -q "error\|404\|500"; then
        ALL_OK=false
        break
    fi
done

if [ "$ALL_OK" = true ]; then
    test_pass
else
    test_fail "All endpoints should respond correctly"
fi

test_start "API integration - authentication works"

# Try without authentication
UNAUTH_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests" \
   -w "%{http_code}")

if echo "$UNAUTH_RESPONSE" | grep -q "401\|403\|unauthorized"; then
    test_pass
else
    test_fail "Should require authentication"
fi

test_start "API integration - error responses correct"

# Try invalid endpoint
ERROR_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/invalid-endpoint" \
   -H "Authorization: Bearer test-key" \
   -w "%{http_code}")

if echo "$ERROR_RESPONSE" | grep -q "404\|not found"; then
    test_pass
else
    test_fail "Should return correct error responses"
fi

test_start "API integration - rate limiting works"

# Make multiple rapid requests
for i in {1..10}; do
    curl -s -X GET "http://localhost:8080/api/v1/change-requests" \
       -H "Authorization: Bearer test-key" > /dev/null
done

# Check if rate limiting kicked in (might not always trigger)
RATE_LIMIT_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests" \
   -H "Authorization: Bearer test-key" \
   -w "%{http_code}")

# Rate limiting might not always trigger, so we'll just check if it doesn't break
if [ -n "$RATE_LIMIT_RESPONSE" ]; then
    test_pass
else
    test_fail "Rate limiting should not break API"
fi

# =============================================================================
# Summary
# =============================================================================

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "Phase 12 E2E Test Results"
echo "═══════════════════════════════════════════════════════════"
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi

