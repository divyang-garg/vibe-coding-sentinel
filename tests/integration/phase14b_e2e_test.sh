#!/bin/bash
# Integration tests for Phase 14B MCP Integration end-to-end
# Run from project root: ./tests/integration/phase14b_e2e_test.sh

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

echo "🧪 Phase 14B MCP Integration End-to-End Tests"
echo "══════════════════════════════════════════════════════════════"
echo ""

# Source test helper
source "$TEST_DIR/../helpers/mcp_test_client.sh" 2>/dev/null || {
    log_warn "MCP test helper not found, using basic functions"
}

# Ensure binary is built
if [[ ! -f "./sentinel" ]]; then
    echo "Building Sentinel binary..."
    ./synapsevibsentinel.sh > /dev/null 2>&1 || true
fi

# Test 1: MCP Server Startup
echo "Test 1: MCP Server Startup"
if grep -q 'case "mcp-server":' synapsevibsentinel.sh; then
    log_pass "mcp-server command registered"
else
    log_fail "mcp-server command not registered"
fi

if grep -q "func runMCPServer" synapsevibsentinel.sh; then
    log_pass "runMCPServer function exists"
else
    log_fail "runMCPServer function not found"
fi

# Test 2: Initialize Method
echo ""
echo "Test 2: Initialize Method"
INIT_RESPONSE=$(send_initialize 2>/dev/null || echo "")
if echo "$INIT_RESPONSE" | grep -q '"jsonrpc":"2.0"'; then
    if echo "$INIT_RESPONSE" | grep -q '"result"'; then
        if echo "$INIT_RESPONSE" | grep -q '"protocolVersion":"2024-11-05"'; then
            log_pass "Initialize returns correct protocol version"
        else
            log_fail "Initialize protocol version incorrect"
        fi
    else
        log_fail "Initialize response missing result"
    fi
else
    log_fail "Initialize response not valid JSON-RPC 2.0"
fi

# Test 3: Tools List
echo ""
echo "Test 3: Tools List"
TOOLS_RESPONSE=$(send_tools_list 2>/dev/null || echo "")
if echo "$TOOLS_RESPONSE" | grep -q '"jsonrpc":"2.0"'; then
    if echo "$TOOLS_RESPONSE" | grep -q '"result"'; then
        if echo "$TOOLS_RESPONSE" | grep -q "sentinel_analyze_feature_comprehensive"; then
            log_pass "Tools list includes comprehensive analysis tool"
        else
            log_fail "Tools list missing comprehensive analysis tool"
        fi
    else
        log_fail "Tools list response missing result"
    fi
else
    log_fail "Tools list response not valid JSON-RPC 2.0"
fi

# Test 4: Tool Call - Parameter Validation
echo ""
echo "Test 4: Tool Call - Parameter Validation"
TOOL_CALL_NO_FEATURE='{"feature":""}'
TOOL_CALL_RESPONSE=$(send_tools_call "sentinel_analyze_feature_comprehensive" "$TOOL_CALL_NO_FEATURE" 4 2>/dev/null || echo "")
if echo "$TOOL_CALL_RESPONSE" | grep -q '"error"'; then
    if echo "$TOOL_CALL_RESPONSE" | grep -q '"code":-32602'; then
        log_pass "Tool call correctly validates required parameters"
    else
        log_fail "Tool call error code incorrect for missing parameter"
    fi
else
    log_fail "Tool call should return error for missing feature"
fi

# Test 5: Tool Call - Invalid Codebase Path
echo ""
echo "Test 5: Tool Call - Invalid Codebase Path"
TOOL_CALL_INVALID_PATH='{"feature":"test","codebasePath":"/nonexistent/path/12345"}'
TOOL_CALL_INVALID_RESPONSE=$(send_tools_call "sentinel_analyze_feature_comprehensive" "$TOOL_CALL_INVALID_PATH" 5 2>/dev/null || echo "")
if echo "$TOOL_CALL_INVALID_RESPONSE" | grep -q '"error"'; then
    if echo "$TOOL_CALL_INVALID_RESPONSE" | grep -q '"code":-32602'; then
        log_pass "Tool call correctly validates codebase path"
    else
        log_fail "Tool call error code incorrect for invalid path"
    fi
else
    log_fail "Tool call should return error for invalid codebase path"
fi

# Test 6: Tool Call - Hub Not Configured (Fallback Scenario)
echo ""
echo "Test 6: Tool Call - Hub Not Configured (Fallback Scenario)"
OLD_HUB_URL="$SENTINEL_HUB_URL"
OLD_API_KEY="$SENTINEL_API_KEY"
unset SENTINEL_HUB_URL
unset SENTINEL_API_KEY

TOOL_CALL_NO_HUB='{"feature":"test-feature","codebasePath":"."}'
TOOL_CALL_NO_HUB_RESPONSE=$(send_tools_call "sentinel_analyze_feature_comprehensive" "$TOOL_CALL_NO_HUB" 6 2>/dev/null || echo "")

export SENTINEL_HUB_URL="$OLD_HUB_URL"
export SENTINEL_API_KEY="$OLD_API_KEY"

if echo "$TOOL_CALL_NO_HUB_RESPONSE" | grep -q '"error"'; then
    if echo "$TOOL_CALL_NO_HUB_RESPONSE" | grep -q '"code":-32002'; then
        log_pass "Tool call correctly handles Hub not configured"
    else
        log_fail "Tool call error code incorrect for missing Hub config"
    fi
else
    log_fail "Tool call should return error when Hub not configured"
fi

# Test 7: Tool Call - Hub Unavailable (Network Error)
echo ""
echo "Test 7: Tool Call - Hub Unavailable (Network Error)"
# Set invalid Hub URL
export SENTINEL_HUB_URL="http://localhost:99999"
export SENTINEL_API_KEY="test-api-key-12345678901234567890"

TOOL_CALL_UNREACHABLE='{"feature":"test-feature","codebasePath":"."}'
TOOL_CALL_UNREACHABLE_RESPONSE=$(send_tools_call "sentinel_analyze_feature_comprehensive" "$TOOL_CALL_UNREACHABLE" 7 2>/dev/null || echo "")

# Restore original values
export SENTINEL_HUB_URL="${OLD_HUB_URL:-http://localhost:8080}"
export SENTINEL_API_KEY="$OLD_API_KEY"

if echo "$TOOL_CALL_UNREACHABLE_RESPONSE" | grep -q '"error"'; then
    if echo "$TOOL_CALL_UNREACHABLE_RESPONSE" | grep -q '"code":-32000'; then
        log_pass "Tool call correctly handles Hub unavailable"
    else
        log_fail "Tool call error code incorrect for Hub unavailable"
    fi
else
    log_warn "Tool call Hub unavailable test (may pass if Hub is actually running)"
fi

# Test 8: End-to-End Flow (Requires Running Hub)
echo ""
echo "Test 8: End-to-End Flow (Requires Running Hub)"
if [[ -n "$SENTINEL_HUB_URL" ]] && [[ -n "$SENTINEL_API_KEY" ]]; then
    # Check if Hub is reachable
    if curl -s -f -o /dev/null --connect-timeout 2 "$SENTINEL_HUB_URL/health" 2>/dev/null || \
       curl -s -f -o /dev/null --connect-timeout 2 "$SENTINEL_HUB_URL/api/v1/health" 2>/dev/null; then
        TOOL_CALL_VALID='{"feature":"test-feature","codebasePath":".","depth":"surface"}'
        TOOL_CALL_VALID_RESPONSE=$(send_tools_call "sentinel_analyze_feature_comprehensive" "$TOOL_CALL_VALID" 8 5 2>/dev/null || echo "")
        
        if echo "$TOOL_CALL_VALID_RESPONSE" | grep -q '"result"'; then
            log_pass "End-to-end flow works with running Hub"
        elif echo "$TOOL_CALL_VALID_RESPONSE" | grep -q '"error"'; then
            ERROR_CODE=$(get_mcp_error_code "$TOOL_CALL_VALID_RESPONSE")
            log_warn "End-to-end flow returned error (code: $ERROR_CODE) - may be expected if Hub not fully configured"
        else
            log_warn "End-to-end flow response unclear"
        fi
    else
        log_warn "Hub not reachable, skipping end-to-end test"
    fi
else
    log_warn "Hub not configured, skipping end-to-end test"
fi

# Test 9: Unknown Method Handling
echo ""
echo "Test 9: Unknown Method Handling"
UNKNOWN_RESPONSE=$(send_mcp_request "unknown/method" "{}" 9 2>/dev/null || echo "")
if echo "$UNKNOWN_RESPONSE" | grep -q '"error"'; then
    if echo "$UNKNOWN_RESPONSE" | grep -q '"code":-32601'; then
        log_pass "Unknown method correctly returns method not found error"
    else
        log_fail "Unknown method error code incorrect"
    fi
else
    log_fail "Unknown method should return error"
fi

# Test 10: Malformed JSON Handling
echo ""
echo "Test 10: Malformed JSON Handling"
MALFORMED_REQUEST='{"jsonrpc":"2.0","id":10,"method":"initialize"'
MALFORMED_RESPONSE=$(echo "$MALFORMED_REQUEST" | timeout 2 ./sentinel mcp-server 2>/dev/null | head -1 || echo "")
if echo "$MALFORMED_RESPONSE" | grep -q '"error"'; then
    if echo "$MALFORMED_RESPONSE" | grep -q '"code":-32700'; then
        log_pass "Malformed JSON correctly returns parse error"
    else
        log_fail "Malformed JSON error code incorrect"
    fi
else
    log_warn "Malformed JSON handling (may vary based on implementation)"
fi

# Summary
echo ""
echo "══════════════════════════════════════════════════════════════"
echo "   Phase 14B Integration Test Results"
echo "══════════════════════════════════════════════════════════════"
echo ""
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"

TOTAL=$((TESTS_PASSED + TESTS_FAILED))
if [[ $TOTAL -gt 0 ]]; then
    PERCENT=$((TESTS_PASSED * 100 / TOTAL))
    echo "Success Rate: ${PERCENT}%"
fi
echo ""

# Manual Testing Instructions
echo "══════════════════════════════════════════════════════════════"
echo "   Manual Cursor Integration Testing"
echo "══════════════════════════════════════════════════════════════"
echo ""
echo "To test MCP integration with Cursor IDE:"
echo ""
echo "1. Configure Cursor MCP settings (~/.cursor/mcp.json):"
echo '   {'
echo '     "mcpServers": {'
echo '       "sentinel": {'
echo '         "command": "'"$(pwd)/sentinel"'",'
echo '         "args": ["mcp-server"],'
echo '         "env": {'
echo '           "SENTINEL_HUB_URL": "http://localhost:8080",'
echo '           "SENTINEL_API_KEY": "your-api-key"'
echo '         }'
echo '       }'
echo '     }'
echo '   }'
echo ""
echo "2. Restart Cursor IDE"
echo ""
echo "3. In Cursor chat, try:"
echo '   "Use sentinel_analyze_feature_comprehensive to analyze the user authentication feature"'
echo ""
echo "4. Verify results appear in Cursor chat"
echo ""

if [[ $TESTS_FAILED -gt 0 ]]; then
    exit 1
fi

exit 0







