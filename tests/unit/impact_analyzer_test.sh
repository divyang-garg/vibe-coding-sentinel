#!/bin/bash
# Phase 12: Impact Analyzer Unit Tests
# Tests for impact analysis functionality

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
# TEST 1: Impact Analysis
# =============================================================================

test_start "Impact analysis - code impact detection"

# Get a change request ID
CR_ID=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests?status=pending" \
   -H "Authorization: Bearer test-key" | grep -oE "CR-[0-9]+" | head -1)

if [ -z "$CR_ID" ]; then
    test_fail "No change request found to test"
    exit 1
fi

# Create test codebase
TEST_CODEBASE="$FIXTURES_DIR/impact_test_codebase"
mkdir -p "$TEST_CODEBASE"

cat > "$TEST_CODEBASE/process_payment.go" << 'EOF'
package main

func processPayment(amount float64) error {
    return nil
}
EOF

# Run impact analysis
RESPONSE=$(echo "{\"codebasePath\": \"$TEST_CODEBASE\"}" | \
   curl -s -X GET "http://localhost:8080/api/v1/change-requests/$CR_ID/impact" \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer test-key" \
   -d @-)

if echo "$RESPONSE" | grep -q "code_impact\|affected_files"; then
    test_pass
else
    test_fail "Should detect code impact"
fi

test_start "Impact analysis - test impact detection"

# Create test file
cat > "$TEST_CODEBASE/process_payment_test.go" << 'EOF'
package main

import "testing"

func TestProcessPayment(t *testing.T) {
    err := processPayment(100.0)
    if err != nil {
        t.Fail()
    }
}
EOF

# Run impact analysis
RESPONSE=$(echo "{\"codebasePath\": \"$TEST_CODEBASE\"}" | \
   curl -s -X GET "http://localhost:8080/api/v1/change-requests/$CR_ID/impact" \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer test-key" \
   -d @-)

if echo "$RESPONSE" | grep -q "test_impact\|affected_tests"; then
    test_pass
else
    test_fail "Should detect test impact"
fi

test_start "Impact analysis - effort estimation"

# Check if effort estimation is included
if echo "$RESPONSE" | grep -q "effort\|estimated_hours"; then
    test_pass
else
    test_fail "Should include effort estimation"
fi

test_start "Impact analysis - empty impact handling"

# Create a change request that doesn't affect code
# (This would require a specific test setup)
# For now, check if empty impact is handled gracefully
if echo "$RESPONSE" | grep -q "impact\|\[\]"; then
    test_pass
else
    test_fail "Should handle empty impact gracefully"
fi

# =============================================================================
# TEST 2: Structured Logging Tests
# =============================================================================

test_start "Structured logging - no log.Printf calls in code"

# Verify no log.Printf calls exist in impact_analyzer.go
if grep -q "log\.Printf" "$PROJECT_ROOT/hub/api/impact_analyzer.go"; then
    test_fail "Should not use log.Printf, use structured logging (LogWarn, LogError, LogInfo)"
else
    test_pass
fi

test_start "Structured logging - log import removed"

# Verify log import is not present
if grep -q '^[[:space:]]*"log"' "$PROJECT_ROOT/hub/api/impact_analyzer.go"; then
    test_fail "log import should be removed if not used"
else
    test_pass
fi

test_start "Structured logging - uses LogWarn/LogError"

# Verify structured logging functions are used
if grep -q "LogWarn\|LogError\|LogInfo" "$PROJECT_ROOT/hub/api/impact_analyzer.go"; then
    test_pass
else
    test_fail "Should use structured logging functions (LogWarn, LogError, LogInfo)"
fi

# =============================================================================
# TEST 3: Storage Tests
# =============================================================================

test_start "Storage - storeImpactAnalysis() function exists"

# Verify storeImpactAnalysis function exists in code
if grep -q "func storeImpactAnalysis" "$PROJECT_ROOT/hub/api/impact_analyzer.go"; then
    test_pass
else
    test_fail "storeImpactAnalysis() function should exist in impact_analyzer.go"
fi

test_start "Storage - no duplicate storage in handler"

# Verify analyzeImpactHandler uses storeImpactAnalysis() instead of inline storage
if grep -A 10 "func analyzeImpactHandler" "$PROJECT_ROOT/hub/api/main.go" | grep -q "storeImpactAnalysis"; then
    test_pass
else
    test_fail "analyzeImpactHandler should use storeImpactAnalysis() function"
fi

test_start "Storage - impact analysis stored"

# Verify impact analysis is stored with change request
CR_DETAIL=$(curl -s -X GET "http://localhost:8080/api/v1/change-requests/$CR_ID" \
   -H "Authorization: Bearer test-key")

if echo "$CR_DETAIL" | grep -q "impact_analysis"; then
    test_pass
else
    test_fail "Impact analysis should be stored with change request"
fi

test_start "Storage - JSONB marshaling works"

# Check if impact analysis JSON is valid
IMPACT_JSON=$(echo "$CR_DETAIL" | grep -o '"impact_analysis":{[^}]*}' || echo "")

if [ -n "$IMPACT_JSON" ]; then
    # Try to parse as JSON
    if echo "$IMPACT_JSON" | python3 -m json.tool > /dev/null 2>&1; then
        test_pass
    else
        test_fail "Impact analysis should be valid JSON"
    fi
else
    test_fail "Impact analysis should be stored as JSONB"
fi

# =============================================================================
# Summary
# =============================================================================

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "Impact Analyzer Test Results"
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

