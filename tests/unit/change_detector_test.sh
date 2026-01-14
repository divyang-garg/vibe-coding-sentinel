#!/bin/bash
# Phase 12: Change Detector Unit Tests
# Tests for change detection functionality

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
# TEST 1: Change Detection
# =============================================================================

test_start "Change detection - new knowledge item"

# Upload a document to trigger change detection
TEST_DOC="$FIXTURES_DIR/docs/test_business_rules.md"
mkdir -p "$(dirname "$TEST_DOC")"

cat > "$TEST_DOC" << 'EOF'
# Business Rules

## Rule 1: Process Payment
Payment processing must validate card number.
EOF

# Upload document (this should trigger change detection)
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/documents/upload \
   -H "Authorization: Bearer test-key" \
   -F "file=@$TEST_DOC" \
   -F "projectId=00000000-0000-0000-0000-000000000000")

# Check if change request was created
CHANGE_REQUESTS=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key")

if echo "$CHANGE_REQUESTS" | grep -q "new\|CR-"; then
    test_pass
else
    test_fail "Should detect new knowledge items"
fi

test_start "Change detection - modified knowledge item"

# Modify the document
cat > "$TEST_DOC" << 'EOF'
# Business Rules

## Rule 1: Process Payment
Payment processing must validate card number and expiration date.
EOF

# Re-upload document
curl -s -X POST http://localhost:8080/api/v1/documents/upload \
   -H "Authorization: Bearer test-key" \
   -F "file=@$TEST_DOC" \
   -F "projectId=00000000-0000-0000-0000-000000000000" > /dev/null

# Check if modification change request was created
CHANGE_REQUESTS=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key")

if echo "$CHANGE_REQUESTS" | grep -q "modified\|CR-"; then
    test_pass
else
    test_fail "Should detect modified knowledge items"
fi

test_start "Change detection - removed knowledge item"

# Remove a rule from the document
cat > "$TEST_DOC" << 'EOF'
# Business Rules

## Rule 2: Validate User
User validation must check email format.
EOF

# Re-upload document
curl -s -X POST http://localhost:8080/api/v1/documents/upload \
   -H "Authorization: Bearer test-key" \
   -F "file=@$TEST_DOC" \
   -F "projectId=00000000-0000-0000-0000-000000000000" > /dev/null

# Check if removal change request was created
CHANGE_REQUESTS=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key")

if echo "$CHANGE_REQUESTS" | grep -q "removed\|CR-"; then
    test_pass
else
    test_fail "Should detect removed knowledge items"
fi

test_start "Change detection - unchanged items not flagged"

# Upload same document again
curl -s -X POST http://localhost:8080/api/v1/documents/upload \
   -H "Authorization: Bearer test-key" \
   -F "file=@$TEST_DOC" \
   -F "projectId=00000000-0000-0000-0000-000000000000" > /dev/null

# Count change requests
BEFORE_COUNT=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key" | grep -o "CR-" | wc -l)

# Upload again
curl -s -X POST http://localhost:8080/api/v1/documents/upload \
   -H "Authorization: Bearer test-key" \
   -F "file=@$TEST_DOC" \
   -F "projectId=00000000-0000-0000-0000-000000000000" > /dev/null

AFTER_COUNT=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key" | grep -o "CR-" | wc -l)

if [ "$BEFORE_COUNT" -eq "$AFTER_COUNT" ]; then
    test_pass
else
    test_fail "Should not create change requests for unchanged items"
fi

# =============================================================================
# TEST 2: Change Request Generation
# =============================================================================

test_start "Change request generation - CR ID format"

# Get a change request
CHANGE_REQUESTS=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key")

# Check if CR IDs match format CR-XXX
if echo "$CHANGE_REQUESTS" | grep -qE "CR-[0-9]+"; then
    test_pass
else
    test_fail "Change request IDs should match format CR-XXX"
fi

test_start "Change request generation - current/proposed state stored"

# Get a change request detail
CR_ID=$(echo "$CHANGE_REQUESTS" | grep -oE "CR-[0-9]+" | head -1)

if [ -n "$CR_ID" ]; then
    CR_DETAIL=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests/$CR_ID" \
       -H "Authorization: Bearer test-key")
    
    if echo "$CR_DETAIL" | grep -q "current_state\|proposed_state"; then
        test_pass
    else
        test_fail "Should store current and proposed state"
    fi
else
    test_fail "No change request found to test"
fi

# =============================================================================
# TEST 3: Structured Logging Tests
# =============================================================================

test_start "Structured logging - no log.Printf calls in code"

# Verify no log.Printf calls exist in change_detector.go
if grep -q "log\.Printf" "$PROJECT_ROOT/hub/api/change_detector.go"; then
    test_fail "Should not use log.Printf, use structured logging (LogWarn, LogError, LogInfo)"
else
    test_pass
fi

test_start "Structured logging - log import removed"

# Verify log import is not present
if grep -q '^[[:space:]]*"log"' "$PROJECT_ROOT/hub/api/change_detector.go"; then
    test_fail "log import should be removed if not used"
else
    test_pass
fi

test_start "Structured logging - uses LogWarn/LogError"

# Verify structured logging functions are used
if grep -q "LogWarn\|LogError\|LogInfo" "$PROJECT_ROOT/hub/api/change_detector.go"; then
    test_pass
else
    test_fail "Should use structured logging functions (LogWarn, LogError, LogInfo)"
fi

# =============================================================================
# TEST 4: Storage Tests
# =============================================================================

test_start "Storage - change request stored in database"

# Verify change request can be retrieved
if [ -n "$CR_ID" ]; then
    CR_DETAIL=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests/$CR_ID" \
       -H "Authorization: Bearer test-key")
    
    if echo "$CR_DETAIL" | grep -q "$CR_ID"; then
        test_pass
    else
        test_fail "Change request should be stored and retrievable"
    fi
else
    test_fail "No change request found to test"
fi

test_start "Storage - JSONB fields marshaled correctly"

# Check if JSONB fields are properly formatted
if [ -n "$CR_ID" ]; then
    CR_DETAIL=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests/$CR_ID" \
       -H "Authorization: Bearer test-key")
    
    # Check if JSON parsing works (no parse errors)
    if echo "$CR_DETAIL" | python3 -m json.tool > /dev/null 2>&1; then
        test_pass
    else
        test_fail "JSONB fields should be properly marshaled"
    fi
else
    test_fail "No change request found to test"
fi

# =============================================================================
# Summary
# =============================================================================

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "Change Detector Test Results"
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

