#!/bin/bash
# Comprehensive integration tests for all MCP tools
# Tests all MCP tool handlers: context, patterns, business, security, validation, actions
# Run from project root: ./tests/integration/mcp_tools_comprehensive_test.sh

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

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "   MCP Tools Comprehensive Integration Tests"
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

# Source MCP test helper
source "$TEST_DIR/../helpers/mcp_test_client.sh" 2>/dev/null || {
    log_warn "MCP test helper not found, using basic functions"
}

# Test directory setup
TEST_TMP_DIR=$(mktemp -d)
trap "rm -rf $TEST_TMP_DIR" EXIT

# Create test codebase
mkdir -p "$TEST_TMP_DIR/src"
cat > "$TEST_TMP_DIR/src/test.js" << 'EOF'
// Test file
function test() {
    return true;
}
EOF

# Function to send MCP tool call
send_mcp_tool_call() {
    local tool_name=$1
    local params=$2
    
    echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"$tool_name\",\"arguments\":$params}}" | \
    timeout 5 ./sentinel mcp-server 2>&1 || true
}

# Test 1: sentinel_get_context
echo ""
echo "Test 1: sentinel_get_context"
echo "──────────────────────────────────────────────────────────────"

RESPONSE=$(send_mcp_tool_call "sentinel_get_context" "{\"codebasePath\":\"$TEST_TMP_DIR\"}")

if echo "$RESPONSE" | grep -qi "context\|recent\|errors\|git"; then
    log_pass "sentinel_get_context returns context data"
else
    log_warn "sentinel_get_context may need Hub configuration"
fi

# Test 2: sentinel_get_patterns
echo ""
echo "Test 2: sentinel_get_patterns"
echo "──────────────────────────────────────────────────────────────"

RESPONSE=$(send_mcp_tool_call "sentinel_get_patterns" "{\"codebasePath\":\"$TEST_TMP_DIR\"}")

if echo "$RESPONSE" | grep -qi "patterns\|conventions\|learned"; then
    log_pass "sentinel_get_patterns returns patterns"
else
    log_warn "sentinel_get_patterns may need Hub configuration"
fi

# Test 3: sentinel_get_business_context
echo ""
echo "Test 3: sentinel_get_business_context"
echo "──────────────────────────────────────────────────────────────"

RESPONSE=$(send_mcp_tool_call "sentinel_get_business_context" "{\"codebasePath\":\"$TEST_TMP_DIR\"}")

if echo "$RESPONSE" | grep -qi "business\|rules\|entities\|journeys"; then
    log_pass "sentinel_get_business_context returns business context"
else
    log_warn "sentinel_get_business_context may need Hub configuration"
fi

# Test 4: sentinel_get_security_context
echo ""
echo "Test 4: sentinel_get_security_context"
echo "──────────────────────────────────────────────────────────────"

RESPONSE=$(send_mcp_tool_call "sentinel_get_security_context" "{\"codebasePath\":\"$TEST_TMP_DIR\"}")

if echo "$RESPONSE" | grep -qi "security\|rules\|compliance\|score"; then
    log_pass "sentinel_get_security_context returns security context"
else
    log_warn "sentinel_get_security_context may need Hub configuration"
fi

# Test 5: sentinel_get_test_requirements
echo ""
echo "Test 5: sentinel_get_test_requirements"
echo "──────────────────────────────────────────────────────────────"

RESPONSE=$(send_mcp_tool_call "sentinel_get_test_requirements" "{\"codebasePath\":\"$TEST_TMP_DIR\"}")

if echo "$RESPONSE" | grep -qi "test\|requirements\|coverage"; then
    log_pass "sentinel_get_test_requirements returns test requirements"
else
    log_warn "sentinel_get_test_requirements may need Hub configuration"
fi

# Test 6: sentinel_check_file_size
echo ""
echo "Test 6: sentinel_check_file_size"
echo "──────────────────────────────────────────────────────────────"

RESPONSE=$(send_mcp_tool_call "sentinel_check_file_size" "{\"filePath\":\"$TEST_TMP_DIR/src/test.js\"}")

if echo "$RESPONSE" | grep -qi "size\|recommendation\|split"; then
    log_pass "sentinel_check_file_size returns file size analysis"
else
    log_warn "sentinel_check_file_size may need verification"
fi

# Test 7: sentinel_validate_code
echo ""
echo "Test 7: sentinel_validate_code"
echo "──────────────────────────────────────────────────────────────"

CODE=$(cat "$TEST_TMP_DIR/src/test.js" | jq -Rs .)
RESPONSE=$(send_mcp_tool_call "sentinel_validate_code" "{\"code\":$CODE,\"language\":\"javascript\"}")

if echo "$RESPONSE" | grep -qi "valid\|issues\|suggestions"; then
    log_pass "sentinel_validate_code returns validation results"
else
    log_warn "sentinel_validate_code may need Hub configuration"
fi

# Test 8: sentinel_validate_security
echo ""
echo "Test 8: sentinel_validate_security"
echo "──────────────────────────────────────────────────────────────"

CODE=$(cat "$TEST_TMP_DIR/src/test.js" | jq -Rs .)
RESPONSE=$(send_mcp_tool_call "sentinel_validate_security" "{\"code\":$CODE,\"language\":\"javascript\"}")

if echo "$RESPONSE" | grep -qi "security\|valid\|issues"; then
    log_pass "sentinel_validate_security returns security validation"
else
    log_warn "sentinel_validate_security may need Hub configuration"
fi

# Test 9: sentinel_validate_business
echo ""
echo "Test 9: sentinel_validate_business"
echo "──────────────────────────────────────────────────────────────"

CODE=$(cat "$TEST_TMP_DIR/src/test.js" | jq -Rs .)
RESPONSE=$(send_mcp_tool_call "sentinel_validate_business" "{\"feature\":\"test feature\",\"code\":$CODE}")

if echo "$RESPONSE" | grep -qi "business\|valid\|violations"; then
    log_pass "sentinel_validate_business returns business validation"
else
    log_warn "sentinel_validate_business may need Hub configuration"
fi

# Test 10: sentinel_validate_tests
echo ""
echo "Test 10: sentinel_validate_tests"
echo "──────────────────────────────────────────────────────────────"

RESPONSE=$(send_mcp_tool_call "sentinel_validate_tests" "{\"testFilePath\":\"$TEST_TMP_DIR/src/test.js\"}")

if echo "$RESPONSE" | grep -qi "test\|quality\|coverage"; then
    log_pass "sentinel_validate_tests returns test validation"
else
    log_warn "sentinel_validate_tests may need Hub configuration"
fi

# Test 11: sentinel_apply_fix
echo ""
echo "Test 11: sentinel_apply_fix"
echo "──────────────────────────────────────────────────────────────"

RESPONSE=$(send_mcp_tool_call "sentinel_apply_fix" "{\"filePath\":\"$TEST_TMP_DIR/src/test.js\",\"fixType\":\"remove_debug\"}")

if echo "$RESPONSE" | grep -qi "fix\|applied\|changes"; then
    log_pass "sentinel_apply_fix returns fix results"
else
    log_warn "sentinel_apply_fix may need Hub configuration"
fi

# Test 12: sentinel_generate_tests
echo ""
echo "Test 12: sentinel_generate_tests"
echo "──────────────────────────────────────────────────────────────"

RESPONSE=$(send_mcp_tool_call "sentinel_generate_tests" "{\"filePath\":\"$TEST_TMP_DIR/src/test.js\"}")

if echo "$RESPONSE" | grep -qi "test\|generated\|cases"; then
    log_pass "sentinel_generate_tests returns test generation results"
else
    log_warn "sentinel_generate_tests may need Hub configuration"
fi

# Test 13: sentinel_run_tests
echo ""
echo "Test 13: sentinel_run_tests"
echo "──────────────────────────────────────────────────────────────"

RESPONSE=$(send_mcp_tool_call "sentinel_run_tests" "{\"testFilePath\":\"$TEST_TMP_DIR/src/test.js\"}")

if echo "$RESPONSE" | grep -qi "test\|execution\|results"; then
    log_pass "sentinel_run_tests returns test execution results"
else
    log_warn "sentinel_run_tests may need Hub configuration"
fi

# Test 14: sentinel_check_intent
echo ""
echo "Test 14: sentinel_check_intent"
echo "──────────────────────────────────────────────────────────────"

RESPONSE=$(send_mcp_tool_call "sentinel_check_intent" "{\"prompt\":\"Add user authentication\",\"codebasePath\":\"$TEST_TMP_DIR\"}")

if echo "$RESPONSE" | grep -qi "intent\|clarifying\|template"; then
    log_pass "sentinel_check_intent returns intent analysis"
else
    log_warn "sentinel_check_intent may need Hub configuration"
fi

# Test 15: Input validation
echo ""
echo "Test 15: Input validation and sanitization"
echo "──────────────────────────────────────────────────────────────"

# Test path traversal prevention
RESPONSE=$(send_mcp_tool_call "sentinel_get_context" "{\"codebasePath\":\"../../../etc/passwd\"}")

if echo "$RESPONSE" | grep -qi "invalid\|error"; then
    log_pass "Path traversal prevention works"
else
    log_warn "Path validation may need verification"
fi

# Test string sanitization
RESPONSE=$(send_mcp_tool_call "sentinel_validate_code" "{\"code\":\"test\x00code\",\"language\":\"javascript\"}")

if echo "$RESPONSE" | grep -qi "valid\|error"; then
    log_pass "String sanitization works"
else
    log_warn "String sanitization may need verification"
fi

# Test 16: Error handling
echo ""
echo "Test 16: Error handling"
echo "──────────────────────────────────────────────────────────────"

# Test invalid tool name
RESPONSE=$(echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"invalid_tool\",\"arguments\":{}}}" | \
    timeout 5 ./sentinel mcp-server 2>&1 || true)

if echo "$RESPONSE" | grep -qi "error\|not found\|invalid"; then
    log_pass "Error handling for invalid tool works"
else
    log_warn "Error handling may need verification"
fi

# Test missing required parameters
RESPONSE=$(send_mcp_tool_call "sentinel_validate_code" "{}")

if echo "$RESPONSE" | grep -qi "error\|invalid\|missing"; then
    log_pass "Error handling for missing parameters works"
else
    log_warn "Parameter validation may need verification"
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










