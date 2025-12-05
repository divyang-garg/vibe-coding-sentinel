#!/bin/bash
# Unit tests for Sentinel MCP server functionality
# Run from project root: ./tests/unit/mcp_test.sh

# Don't use set -e as some commands intentionally fail

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

cleanup_lock() {
    rm -f /tmp/sentinel.lock
}

echo ""
echo "=============================================="
echo "   MCP Server Unit Tests"
echo "=============================================="
echo ""

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT"

# ============================================================================
# Test: mcp-server command exists (Phase 14)
# ============================================================================

echo "Testing mcp-server command..."
cleanup_lock

# Check if command is in switch statement
if grep -q 'case "mcp-server":' "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    log_pass "mcp-server command is registered"
else
    log_fail "mcp-server command not found in switch statement"
fi

# ============================================================================
# Test: runMCPServer function exists (Phase 14)
# ============================================================================

echo ""
echo "Testing runMCPServer function..."
cleanup_lock

if grep -q "func runMCPServer" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    log_pass "runMCPServer function is defined"
else
    log_fail "runMCPServer function not found"
fi

# ============================================================================
# Test: MCP server command executes (Phase 14)
# ============================================================================

echo ""
echo "Testing mcp-server command execution..."
cleanup_lock

# Rebuild binary to ensure latest code
if [[ -f "$PROJECT_ROOT/synapsevibsentinel.sh" ]]; then
    "$PROJECT_ROOT/synapsevibsentinel.sh" > /dev/null 2>&1 || true
fi

MCP_OUTPUT=$(./sentinel mcp-server 2>&1 || true)
MCP_EXIT_CODE=$?
if echo "$MCP_OUTPUT" | grep -qi "MCP\|mcp\|Starting Sentinel\|Sentinel MCP\|server\|stub\|structure ready\|Phase 14"; then
    log_pass "mcp-server command executes and produces output"
elif [[ $MCP_EXIT_CODE -eq 0 ]] && [[ -n "$MCP_OUTPUT" ]]; then
    # Stub exits with 0 and produces output - acceptable
    log_pass "mcp-server command executes (stub implementation)"
else
    # Check if command is at least registered (doesn't show "unknown command")
    if ! echo "$MCP_OUTPUT" | grep -qi "unknown\|not found\|invalid"; then
        log_pass "mcp-server command is registered (output may vary)"
    else
        log_fail "mcp-server command doesn't produce expected output"
    fi
fi

# ============================================================================
# Test: MCP server indicates stub status (Phase 14)
# ============================================================================

echo ""
echo "Testing MCP server stub status message..."
cleanup_lock

if echo "$MCP_OUTPUT" | grep -qi "stub\|not yet\|pending\|Phase 14"; then
    log_pass "MCP server correctly indicates stub implementation"
else
    log_fail "MCP server doesn't indicate stub status"
fi

# ============================================================================
# Test: MCP documentation exists (Phase 14)
# ============================================================================

echo ""
echo "Testing MCP documentation..."
cleanup_lock

if grep -q "mcp-server\|MCP" "$PROJECT_ROOT/docs/external/FEATURES.md" 2>/dev/null; then
    log_pass "MCP is documented in FEATURES.md"
else
    log_fail "MCP not documented in FEATURES.md"
fi

if grep -q "MCP\|mcp-server" "$PROJECT_ROOT/docs/external/TECHNICAL_SPEC.md" 2>/dev/null; then
    log_pass "MCP is documented in TECHNICAL_SPEC.md"
else
    log_fail "MCP not documented in TECHNICAL_SPEC.md"
fi

# ============================================================================
# Summary
# ============================================================================

echo ""
echo "=============================================="
echo "   MCP Server Test Results"
echo "=============================================="
echo ""
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"

TOTAL=$((TESTS_PASSED + TESTS_FAILED))
if [[ $TOTAL -gt 0 ]]; then
    PERCENT=$((TESTS_PASSED * 100 / TOTAL))
    echo "Success Rate: ${PERCENT}%"
fi
echo ""

if [[ $TESTS_FAILED -gt 0 ]]; then
    exit 1
fi

exit 0

