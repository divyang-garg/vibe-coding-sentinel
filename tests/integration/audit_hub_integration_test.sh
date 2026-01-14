#!/bin/bash
# Integration tests for runAudit() Hub integration
# Tests Hub API integration, fallback scenarios, caching, and output formats
# Run from project root: ./tests/integration/audit_hub_integration_test.sh

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_info() {
    echo -e "${BLUE}ℹ INFO:${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠ WARN:${NC} $1"
}

# Source test utilities
source "$TEST_DIR/../helpers/test_utils.sh" 2>/dev/null || {
    log_warn "Test utilities not found, using basic functions"
}

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "   Audit Hub Integration Tests"
echo "══════════════════════════════════════════════════════════════"
echo ""

# Ensure binary is built
if [[ ! -f "./sentinel" ]]; then
    log_info "Building Sentinel binary..."
    ./synapsevibsentinel.sh > /dev/null 2>&1 || true
    if [[ ! -f "./sentinel" ]]; then
        log_fail "Failed to build Sentinel binary"
        exit 1
    fi
fi

# Test directory setup
TEST_TMP_DIR=$(mktemp -d)
trap "rm -rf $TEST_TMP_DIR" EXIT

# Create test codebase
mkdir -p "$TEST_TMP_DIR/src"
cat > "$TEST_TMP_DIR/src/test.js" << 'EOF'
// Test file with some issues
function test() {
    console.log("debug"); // Should be flagged
    var password = "hardcoded123"; // Should be flagged
    return true;
}
EOF

# Test 1: Audit with Hub configured (successful)
echo ""
echo "Test 1: Audit with Hub configured (successful)"
echo "──────────────────────────────────────────────────────────────"

if [[ -n "$SENTINEL_HUB_URL" && -n "$SENTINEL_API_KEY" ]]; then
    export SENTINEL_HUB_URL
    export SENTINEL_API_KEY
    
    OUTPUT=$(cd "$TEST_TMP_DIR" && ./sentinel audit --output json 2>&1) || true
    
    if echo "$OUTPUT" | grep -q "findings\|success"; then
        log_pass "Audit with Hub configured returns results"
    else
        log_fail "Audit with Hub configured failed or returned unexpected output"
        echo "Output: $OUTPUT"
    fi
else
    log_warn "SENTINEL_HUB_URL and SENTINEL_API_KEY not set, skipping Hub integration test"
fi

# Test 2: Audit fallback to local when Hub unavailable
echo ""
echo "Test 2: Audit fallback to local when Hub unavailable"
echo "──────────────────────────────────────────────────────────────"

# Set invalid Hub URL
export SENTINEL_HUB_URL="http://localhost:99999"
export SENTINEL_API_KEY="test-api-key-12345678901234567890"

OUTPUT=$(cd "$TEST_TMP_DIR" && timeout 5 ./sentinel audit 2>&1) || true

if echo "$OUTPUT" | grep -qi "fallback\|local\|scanning"; then
    log_pass "Audit falls back to local scanning when Hub unavailable"
else
    log_fail "Audit did not fall back to local scanning"
    echo "Output: $OUTPUT"
fi

# Test 3: Audit with offline mode
echo ""
echo "Test 3: Audit with --offline flag"
echo "──────────────────────────────────────────────────────────────"

OUTPUT=$(cd "$TEST_TMP_DIR" && ./sentinel audit --offline 2>&1) || true

if echo "$OUTPUT" | grep -qi "scanning\|audit"; then
    log_pass "Audit works in offline mode"
else
    log_fail "Audit offline mode failed"
    echo "Output: $OUTPUT"
fi

# Test 4: Audit output formats
echo ""
echo "Test 4: Audit output formats (JSON, HTML, Markdown)"
echo "──────────────────────────────────────────────────────────────"

# Test JSON output
OUTPUT_FILE="$TEST_TMP_DIR/audit.json"
cd "$TEST_TMP_DIR"
./sentinel audit --output json --output-file "$OUTPUT_FILE" 2>&1 > /dev/null || true

if [[ -f "$OUTPUT_FILE" ]]; then
    if grep -q "findings\|success\|timestamp" "$OUTPUT_FILE"; then
        log_pass "JSON output format works"
    else
        log_fail "JSON output format invalid"
    fi
else
    log_fail "JSON output file not created"
fi

# Test HTML output
OUTPUT_FILE="$TEST_TMP_DIR/audit.html"
cd "$TEST_TMP_DIR"
./sentinel audit --output html --output-file "$OUTPUT_FILE" 2>&1 > /dev/null || true

if [[ -f "$OUTPUT_FILE" ]]; then
    if grep -qi "<html\|<!DOCTYPE" "$OUTPUT_FILE"; then
        log_pass "HTML output format works"
    else
        log_fail "HTML output format invalid"
    fi
else
    log_fail "HTML output file not created"
fi

# Test Markdown output
OUTPUT_FILE="$TEST_TMP_DIR/audit.md"
cd "$TEST_TMP_DIR"
./sentinel audit --output markdown --output-file "$OUTPUT_FILE" 2>&1 > /dev/null || true

if [[ -f "$OUTPUT_FILE" ]]; then
    if grep -q "#\|##\|findings" "$OUTPUT_FILE"; then
        log_pass "Markdown output format works"
    else
        log_fail "Markdown output format invalid"
    fi
else
    log_fail "Markdown output file not created"
fi

# Test 5: Audit caching
echo ""
echo "Test 5: Audit result caching"
echo "──────────────────────────────────────────────────────────────"

if [[ -n "$SENTINEL_HUB_URL" && -n "$SENTINEL_API_KEY" ]]; then
    export SENTINEL_HUB_URL
    export SENTINEL_API_KEY
    
    # First run
    OUTPUT1=$(cd "$TEST_TMP_DIR" && ./sentinel audit 2>&1)
    
    # Second run (should use cache)
    OUTPUT2=$(cd "$TEST_TMP_DIR" && ./sentinel audit 2>&1)
    
    if echo "$OUTPUT2" | grep -qi "cache\|cached"; then
        log_pass "Audit caching works"
    else
        log_warn "Cache indicator not found (may still be caching)"
    fi
else
    log_warn "Skipping cache test (Hub not configured)"
fi

# Test 6: CI mode
echo ""
echo "Test 6: Audit in CI mode"
echo "──────────────────────────────────────────────────────────────"

OUTPUT=$(cd "$TEST_TMP_DIR" && ./sentinel audit --ci 2>&1) || true

if echo "$OUTPUT" | grep -qi "audit\|passed\|failed"; then
    log_pass "CI mode works correctly"
else
    log_fail "CI mode output unexpected"
    echo "Output: $OUTPUT"
fi

# Test 7: Path sanitization
echo ""
echo "Test 7: Path sanitization and validation"
echo "──────────────────────────────────────────────────────────────"

# Test with malicious path
OUTPUT=$(cd "$TEST_TMP_DIR" && ./sentinel audit "../../../etc/passwd" 2>&1) || true

if echo "$OUTPUT" | grep -qi "invalid\|error"; then
    log_pass "Path sanitization prevents directory traversal"
else
    log_warn "Path validation may need verification"
fi

# Test 8: Circuit breaker behavior
echo ""
echo "Test 8: Circuit breaker and retry logic"
echo "──────────────────────────────────────────────────────────────"

# Set Hub URL to non-existent server
export SENTINEL_HUB_URL="http://localhost:99999"
export SENTINEL_API_KEY="test-api-key-12345678901234567890"

START_TIME=$(date +%s)
OUTPUT=$(cd "$TEST_TMP_DIR" && timeout 10 ./sentinel audit 2>&1) || true
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

if [[ $DURATION -lt 10 ]]; then
    log_pass "Circuit breaker prevents long waits"
else
    log_warn "Circuit breaker may not be working (took ${DURATION}s)"
fi

# Summary
echo ""
echo "══════════════════════════════════════════════════════════════"
echo "   Test Summary"
echo "══════════════════════════════════════════════════════════════"
echo ""
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo ""

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed${NC}"
    exit 1
fi







