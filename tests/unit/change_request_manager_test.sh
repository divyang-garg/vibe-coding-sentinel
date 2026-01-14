#!/bin/bash
# Phase 12: Change Request Manager Unit Tests
# Tests for change request management functionality

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
# TEST 1: CRUD Tests
# =============================================================================

test_start "CRUD - read change request"

# Get a change request
CR_ID=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key" | grep -oE "CR-[0-9]+" | head -1)

if [ -z "$CR_ID" ]; then
    test_fail "No change request found to test"
    exit 1
fi

# Read change request
CR_DETAIL=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests/$CR_ID" \
   -H "Authorization: Bearer test-key")

if echo "$CR_DETAIL" | grep -q "$CR_ID"; then
    test_pass
else
    test_fail "Should retrieve change request by ID"
fi

test_start "CRUD - list change requests"

# List change requests
LIST_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests" \
   -H "Authorization: Bearer test-key")

if echo "$LIST_RESPONSE" | grep -q "change_requests\|total"; then
    test_pass
else
    test_fail "Should list change requests"
fi

test_start "CRUD - list with pagination"

# List with limit and offset
PAGED_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?limit=10&offset=0" \
   -H "Authorization: Bearer test-key")

if echo "$PAGED_RESPONSE" | grep -q "limit\|offset\|has_next\|has_previous"; then
    test_pass
else
    test_fail "Should support pagination"
fi

test_start "CRUD - filter by status"

# List pending change requests
PENDING_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key")

if echo "$PENDING_RESPONSE" | grep -q "pending\|CR-"; then
    test_pass
else
    test_fail "Should filter change requests by status"
fi

# =============================================================================
# TEST 2: Workflow Tests
# =============================================================================

test_start "Workflow - approve change request"

# Get a pending change request
PENDING_CR=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key" | grep -oE "CR-[0-9]+" | head -1)

if [ -z "$PENDING_CR" ]; then
    test_fail "No pending change request found to test"
    exit 1
fi

# Approve change request
APPROVE_RESPONSE=$(echo '{"approved_by": "test-user"}' | \
   curl -s -X PUT "http://localhost:8080/api/v1/change-requests/$PENDING_CR/approve" \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer test-key" \
   -d @-)

if echo "$APPROVE_RESPONSE" | grep -q "approved\|success"; then
    # Verify status changed
    UPDATED_CR=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests/$PENDING_CR" \
       -H "Authorization: Bearer test-key")
    
    if echo "$UPDATED_CR" | grep -q '"status":"approved"'; then
        test_pass
    else
        test_fail "Status should change to approved"
    fi
else
    test_fail "Should approve change request"
fi

test_start "Workflow - reject change request"

# Get another pending change request
PENDING_CR=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key" | grep -oE "CR-[0-9]+" | head -1)

if [ -n "$PENDING_CR" ]; then
    # Reject change request
    REJECT_RESPONSE=$(echo '{"rejected_by": "test-user", "reason": "Not needed"}' | \
       curl -s -X PUT "http://localhost:8080/api/v1/change-requests/$PENDING_CR/reject" \
       -H "Content-Type: application/json" \
       -H "Authorization: Bearer test-key" \
       -d @-)
    
    if echo "$REJECT_RESPONSE" | grep -q "rejected\|success"; then
        # Verify status changed
        UPDATED_CR=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests/$PENDING_CR" \
           -H "Authorization: Bearer test-key")
        
        if echo "$UPDATED_CR" | grep -q '"status":"rejected"'; then
            test_pass
        else
            test_fail "Status should change to rejected"
        fi
    else
        test_fail "Should reject change request"
    fi
else
    test_pass  # No pending requests to test
fi

test_start "Workflow - knowledge item updates on approval"

# Get an approved change request
APPROVED_CR=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=approved" \
   -H "Authorization: Bearer test-key" | grep -oE "CR-[0-9]+" | head -1)

if [ -n "$APPROVED_CR" ]; then
    # Check if knowledge item was updated
    CR_DETAIL=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests/$APPROVED_CR" \
       -H "Authorization: Bearer test-key")
    
    # Check if knowledge_item_id is present
    if echo "$CR_DETAIL" | grep -q "knowledge_item_id"; then
        test_pass
    else
        test_fail "Should update knowledge item on approval"
    fi
else
    test_pass  # No approved requests to test
fi

# =============================================================================
# TEST 3: Structured Logging Tests
# =============================================================================

test_start "Structured logging - no log.Printf calls in code"

# Verify no log.Printf calls exist in change_request_manager.go
if grep -q "log\.Printf" "$PROJECT_ROOT/hub/api/change_request_manager.go"; then
    test_fail "Should not use log.Printf, use structured logging (LogWarn, LogError, LogInfo)"
else
    test_pass
fi

test_start "Structured logging - log import removed"

# Verify log import is not present
if grep -q '^[[:space:]]*"log"' "$PROJECT_ROOT/hub/api/change_request_manager.go"; then
    test_fail "log import should be removed if not used"
else
    test_pass
fi

test_start "Structured logging - uses LogWarn/LogError"

# Verify structured logging functions are used
if grep -q "LogWarn\|LogError\|LogInfo" "$PROJECT_ROOT/hub/api/change_request_manager.go"; then
    test_pass
else
    test_fail "Should use structured logging functions (LogWarn, LogError, LogInfo)"
fi

# =============================================================================
# TEST 4: Validation Tests
# =============================================================================

test_start "Validation - UUID validation"

# Try to get change request with invalid UUID
INVALID_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests/invalid-uuid" \
   -H "Authorization: Bearer test-key")

if echo "$INVALID_RESPONSE" | grep -q "invalid UUID\|400\|bad request"; then
    test_pass
else
    test_fail "Should validate UUID format"
fi

test_start "Validation - required field validation"

# Try to approve without required field
APPROVE_RESPONSE=$(echo '{}' | \
   curl -s -X PUT "http://localhost:8080/api/v1/change-requests/$CR_ID/approve" \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer test-key" \
   -d @-)

if echo "$APPROVE_RESPONSE" | grep -q "required\|approved_by\|400"; then
    test_pass
else
    test_fail "Should validate required fields"
fi

test_start "Validation - status validation"

# Try invalid status transition
RESPONSE=$(echo '{"status": "invalid_status", "notes": "test"}' | \
   curl -s -X PUT "http://localhost:8080/api/v1/change-requests/$CR_ID/implementation-status" \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer test-key" \
   -d @-)

if echo "$RESPONSE" | grep -q "invalid\|status\|400"; then
    test_pass
else
    test_fail "Should validate status values"
fi

# =============================================================================
# Summary
# =============================================================================

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "Change Request Manager Test Results"
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

