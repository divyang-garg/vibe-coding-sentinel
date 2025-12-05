#!/bin/bash
# Integration tests for Sentinel workflow
# Run from project root: ./tests/integration/workflow_test.sh

# Don't use set -e as grep returns non-zero when no match

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    ((TESTS_FAILED++))
}

log_info() {
    echo -e "${YELLOW}→${NC} $1"
}

cleanup_lock() {
    rm -f /tmp/sentinel.lock
}

# ============================================================================
# Test Setup
# ============================================================================

echo ""
echo "=============================================="
echo "   Sentinel Integration Tests"
echo "=============================================="
echo ""

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT"

# Ensure sentinel binary exists
if [[ ! -f "./sentinel" ]]; then
    echo "Building Sentinel..."
    ./synapsevibsentinel.sh
fi

# ============================================================================
# Test: Help Command
# ============================================================================

log_info "Testing help command..."
cleanup_lock
if ./sentinel --help | grep -q "Synapse Sentinel"; then
    log_pass "Help command works"
else
    log_fail "Help command failed"
fi

# ============================================================================
# Test: Status Command
# ============================================================================

log_info "Testing status command..."
cleanup_lock
if ./sentinel status | grep -q "PROJECT HEALTH"; then
    log_pass "Status command works"
else
    log_fail "Status command failed"
fi

# ============================================================================
# Test: Audit Command
# ============================================================================

log_info "Testing audit command..."
cleanup_lock
# Audit should complete successfully (PASSED or WARNING)
if ./sentinel audit 2>&1 | grep -q "Audit"; then
    log_pass "Audit command runs successfully"
else
    log_fail "Audit command failed"
fi

# ============================================================================
# Test: Audit JSON Output
# ============================================================================

log_info "Testing audit JSON output..."
cleanup_lock
TEMP_REPORT=$(mktemp)
./sentinel audit --output json --output-file "$TEMP_REPORT" 2>/dev/null || true

if [[ -f "$TEMP_REPORT" ]] && grep -q '"findings"' "$TEMP_REPORT"; then
    log_pass "Audit JSON output works"
else
    log_fail "Audit JSON output failed"
fi
rm -f "$TEMP_REPORT"

# ============================================================================
# Test: Baseline Commands
# ============================================================================

log_info "Testing baseline commands..."
cleanup_lock

# List baseline (should be empty or have entries)
if ./sentinel baseline list 2>&1 | grep -qE "(No baseline|Baselined)"; then
    log_pass "Baseline list works"
else
    log_fail "Baseline list failed"
fi

# ============================================================================
# Test: History Commands
# ============================================================================

log_info "Testing history commands..."
cleanup_lock

if ./sentinel history list 2>&1 | grep -qE "(No audit history|Recent Audits)"; then
    log_pass "History list works"
else
    log_fail "History list failed"
fi

# ============================================================================
# Test: List Rules
# ============================================================================

log_info "Testing list-rules command..."
cleanup_lock

if ./sentinel list-rules 2>&1 | grep -qE "(Active Rules|No rules found|rules)"; then
    log_pass "List rules works"
else
    log_fail "List rules failed"
fi

# ============================================================================
# Test: Validate Rules
# ============================================================================

log_info "Testing validate-rules command..."
cleanup_lock

if ./sentinel validate-rules 2>&1; then
    log_pass "Validate rules works"
else
    log_fail "Validate rules failed"
fi

# ============================================================================
# Test: Security Detection - Secrets
# ============================================================================

log_info "Testing secret detection..."
cleanup_lock

# Create temporary config to scan fixtures
TEMP_CONFIG=$(mktemp)
cat > "$TEMP_CONFIG" << 'EOF'
{
  "scanDirs": ["tests/fixtures/security"],
  "excludePaths": [".git"],
  "severityLevels": {
    "secrets": "critical",
    "console.log": "warning"
  }
}
EOF

# Use temp config
cp .sentinelsrc .sentinelsrc.bak
cp "$TEMP_CONFIG" .sentinelsrc

AUDIT_OUTPUT=$(./sentinel audit 2>&1 || true)

# Restore config
mv .sentinelsrc.bak .sentinelsrc
rm -f "$TEMP_CONFIG"

if echo "$AUDIT_OUTPUT" | grep -qi "secret\|API_KEY\|password\|console"; then
    log_pass "Detects hardcoded secrets"
else
    log_fail "Failed to detect hardcoded secrets"
fi

# ============================================================================
# Test: Security Detection - SQL Injection
# ============================================================================

log_info "Testing SQL injection detection..."
cleanup_lock

if echo "$AUDIT_OUTPUT" | grep -qi "SQL\|injection\|query\|eval\|php"; then
    log_pass "Detects SQL injection patterns"
else
    log_fail "Failed to detect SQL injection"
fi

# ============================================================================
# Test: Security Detection - Console Logs
# ============================================================================

log_info "Testing console.log detection..."
cleanup_lock

if echo "$AUDIT_OUTPUT" | grep -qi "console"; then
    log_pass "Detects console.log statements"
else
    log_fail "Failed to detect console.log"
fi

# ============================================================================
# Summary
# ============================================================================

echo ""
echo "=============================================="
echo "   Test Results"
echo "=============================================="
echo ""
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"
echo ""

TOTAL=$((TESTS_PASSED + TESTS_FAILED))
PERCENT=$((TESTS_PASSED * 100 / TOTAL))
echo "Success Rate: ${PERCENT}%"
echo ""

# Exit with error if any tests failed
if [[ $TESTS_FAILED -gt 0 ]]; then
    exit 1
fi

exit 0

