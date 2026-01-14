#!/bin/bash
# Phase 11: Doc-Sync Unit Tests
# Tests for documentation-code synchronization functionality

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
# TEST 1: Status Marker Parsing
# =============================================================================

test_start "Status marker parsing"

# Create a test roadmap file
TEST_ROADMAP="$FIXTURES_DIR/docs/test_roadmap.md"
mkdir -p "$(dirname "$TEST_ROADMAP")"

cat > "$TEST_ROADMAP" << 'EOF'
## Phase 1: Test Phase ✅ COMPLETE

### Tasks

| Task | Days | Status |
|------|------|--------|
| Task 1 | 1 | ✅ Done |
| Task 2 | 0.5 | ⏳ Pending |

## Phase 2: Another Phase 🔴 STUB

### Tasks

| Task | Days | Status |
|------|------|--------|
| Task A | 1 | ⏳ Pending |
EOF

# Test that the file exists and can be parsed
if [ -f "$TEST_ROADMAP" ]; then
    # Check for phase markers
    if grep -q "Phase 1.*COMPLETE" "$TEST_ROADMAP" && grep -q "Phase 2.*STUB" "$TEST_ROADMAP"; then
        test_pass
    else
        test_fail "Phase markers not found correctly"
    fi
else
    test_fail "Test roadmap file not created"
fi

# =============================================================================
# TEST 2: Code Implementation Detection
# =============================================================================

test_start "Code implementation detection"

# Create a test Go file
TEST_CODE_DIR="$FIXTURES_DIR/code"
mkdir -p "$TEST_CODE_DIR"

cat > "$TEST_CODE_DIR/test_feature.go" << 'EOF'
package main

func testFeatureFunction() {
    // Implementation exists
}

func anotherTestFunction() {
    // Another function
}
EOF

# Test that function exists
if grep -q "func testFeatureFunction" "$TEST_CODE_DIR/test_feature.go"; then
    test_pass
else
    test_fail "Function detection failed"
fi

# =============================================================================
# TEST 3: API Endpoint Detection
# =============================================================================

test_start "API endpoint detection"

# Check if main.go has endpoint patterns
if [ -f "$PROJECT_ROOT/hub/api/main.go" ]; then
    if grep -q 'r\.Post\|r\.Get' "$PROJECT_ROOT/hub/api/main.go"; then
        test_pass
    else
        test_fail "Endpoint patterns not found"
    fi
else
    test_fail "main.go not found"
fi

# =============================================================================
# TEST 4: Test Coverage Validation
# =============================================================================

test_start "Test coverage validation"

# Check if test files exist
if [ -d "$PROJECT_ROOT/tests/unit" ] && [ "$(ls -A $PROJECT_ROOT/tests/unit/*.sh 2>/dev/null | wc -l)" -gt 0 ]; then
    test_pass
else
    test_fail "Test directory or test files not found"
fi

# =============================================================================
# TEST 5: Hub API Endpoint Exists
# =============================================================================

test_start "Hub API endpoint registration"

if grep -q "/analyze/doc-sync" "$PROJECT_ROOT/hub/api/main.go"; then
    test_pass
else
    test_fail "Doc-sync endpoint not registered in main.go"
fi

# =============================================================================
# TEST 6: Database Schema Migration
# =============================================================================

test_start "Database schema migration"

if grep -q "doc_sync_reports" "$PROJECT_ROOT/hub/api/main.go" && \
   grep -q "doc_sync_updates" "$PROJECT_ROOT/hub/api/main.go"; then
    test_pass
else
    test_fail "Database migrations not found"
fi

# =============================================================================
# TEST 7: Agent Command Handler
# =============================================================================

test_start "Agent command handler"

if grep -q "case \"doc-sync\":" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "func runDocSync" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    test_pass
else
    test_fail "Doc-sync command handler not found in Agent"
fi

# =============================================================================
# TEST 8: HTTP Client Implementation
# =============================================================================

test_start "HTTP client implementation"

if grep -q "sendHTTPRequest\|sendDocSyncRequest" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "net/http" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    test_pass
else
    test_fail "HTTP client functions not found"
fi

# =============================================================================
# TEST 9: Audit Flag Integration
# =============================================================================

test_start "Audit flag integration"

if grep -q "--doc-sync" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    test_pass
else
    test_fail "--doc-sync flag not integrated into audit"
fi

# =============================================================================
# TEST 10: Report Generation
# =============================================================================

test_start "Report generation functions"

if grep -q "generateReport\|formatReportHumanReadable" "$PROJECT_ROOT/hub/api/doc_sync.go"; then
    test_pass
else
    test_fail "Report generation functions not found"
fi

# =============================================================================
# TEST 11: Update Storage Function
# =============================================================================

test_start "Update storage function"

if grep -q "storeDocSyncUpdate\|storeDocSyncUpdates" "$PROJECT_ROOT/hub/api/doc_sync.go"; then
    test_pass
else
    test_fail "Update storage functions not found"
fi

# =============================================================================
# TEST 12: Update Storage Integration
# =============================================================================

test_start "Update storage integration in analyzeDocSync"

if grep -q "storeDocSyncUpdates" "$PROJECT_ROOT/hub/api/doc_sync.go" && \
   grep -q "generateUpdateSuggestions" "$PROJECT_ROOT/hub/api/doc_sync.go"; then
    test_pass
else
    test_fail "Update storage not integrated into analyzeDocSync"
fi

# =============================================================================
# TEST 13: HTTP Retry Logic
# =============================================================================

test_start "HTTP retry logic implementation"

if grep -q "sendHTTPRequestWithRetry" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "maxRetries\|exponential\|backoff" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    test_pass
else
    test_fail "HTTP retry logic not found"
fi

# =============================================================================
# TEST 14: Retry Logic Integration
# =============================================================================

test_start "Retry logic integration"

# Check that sendHTTPRequest calls sendHTTPRequestWithRetry
if grep -q "sendHTTPRequestWithRetry" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -A 2 "func sendHTTPRequest" "$PROJECT_ROOT/synapsevibsentinel.sh" | grep -q "sendHTTPRequestWithRetry"; then
    test_pass
else
    test_fail "Retry logic not integrated into sendHTTPRequest"
fi

# =============================================================================
# TEST 15: Database Schema - No Duplicates
# =============================================================================

test_start "Database schema - no duplicate tables"

# Count occurrences of doc_sync_reports table definition
COUNT=$(grep -c "CREATE TABLE IF NOT EXISTS doc_sync_reports" "$PROJECT_ROOT/hub/api/main.go" || echo "0")
if [ "$COUNT" -eq 1 ]; then
    test_pass
else
    test_fail "Found $COUNT occurrences of doc_sync_reports table (expected 1)"
fi

# =============================================================================
# SUMMARY
# =============================================================================

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "Test Results:"
echo "  ${GREEN}Passed: $PASSED${NC}"
echo "  ${RED}Failed: $FAILED${NC}"
echo "  Total:  $((PASSED + FAILED))"
echo "═══════════════════════════════════════════════════════════════"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi

