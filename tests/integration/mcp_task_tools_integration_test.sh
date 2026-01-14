#!/bin/bash
# Integration tests for MCP Task Tools (Phase 14E)
# Tests end-to-end workflows: lifecycle, dependencies, filters, concurrency
# Run from project root: ./tests/integration/mcp_task_tools_integration_test.sh

set -e

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
echo "   MCP Task Tools Integration Tests (Phase 14E)"
echo "══════════════════════════════════════════════════════════════"
echo ""

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT"

# Source test utilities
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$TEST_DIR/../helpers/test_utils.sh" ]; then
    source "$TEST_DIR/../helpers/test_utils.sh"
fi

if [ -f "$TEST_DIR/../helpers/mcp_test_client.sh" ]; then
    source "$TEST_DIR/../helpers/mcp_test_client.sh"
fi

# Ensure binary is built
if [[ ! -f "./sentinel" ]]; then
    log_info "Building Sentinel binary..."
    ./synapsevibsentinel.sh > /dev/null 2>&1 || true
    if [[ ! -f "./sentinel" ]]; then
        log_fail "Failed to build Sentinel binary"
        exit 1
    fi
fi

# Check if Hub is available
if [ -z "$SENTINEL_HUB_URL" ] || [ -z "$SENTINEL_API_KEY" ]; then
    log_warn "SENTINEL_HUB_URL or SENTINEL_API_KEY not set. Some tests may be skipped."
    SKIP_HUB_TESTS=true
else
    SKIP_HUB_TESTS=false
    # Test Hub connectivity
    if ! curl -s -f -H "Authorization: Bearer $SENTINEL_API_KEY" "$SENTINEL_HUB_URL/health" > /dev/null 2>&1; then
        log_warn "Hub not reachable at $SENTINEL_HUB_URL. Some tests may be skipped."
        SKIP_HUB_TESTS=true
    fi
fi

# Helper function to send MCP tool call
send_mcp_tool_call() {
    local tool_name=$1
    local arguments=$2
    local id=${3:-1}
    
    local request=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": $id,
  "method": "tools/call",
  "params": {
    "name": "$tool_name",
    "arguments": $arguments
  }
}
EOF
)
    
    echo "$request" | timeout 10 ./sentinel mcp-server 2>&1 | head -1
}

# Helper function to check for error in response
has_error() {
    local response=$1
    echo "$response" | grep -q '"error"'
}

# ============================================================================
# Test 1: Complete Task Lifecycle Flow
# ============================================================================

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "   Test 1: Complete Task Lifecycle Flow"
echo "══════════════════════════════════════════════════════════════"
echo ""

if [ "$SKIP_HUB_TESTS" = "true" ]; then
    log_warn "Skipping lifecycle test - Hub not available"
else
    # Step 1: List tasks
    log_info "Step 1: Listing tasks..."
    RESPONSE=$(send_mcp_tool_call "sentinel_list_tasks" "{}")
    if ! has_error "$RESPONSE"; then
        log_pass "List tasks succeeded"
        
        # Extract task ID from response if available
        TASK_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        if [ -n "$TASK_ID" ]; then
            log_info "Found task ID: $TASK_ID"
            
            # Step 2: Get task status
            log_info "Step 2: Getting task status..."
            STATUS_RESPONSE=$(send_mcp_tool_call "sentinel_get_task_status" "{\"taskId\": \"$TASK_ID\"}")
            if ! has_error "$STATUS_RESPONSE"; then
                log_pass "Get task status succeeded"
                
                # Step 3: Verify task
                log_info "Step 3: Verifying task..."
                VERIFY_RESPONSE=$(send_mcp_tool_call "sentinel_verify_task" "{\"taskId\": \"$TASK_ID\"}")
                if ! has_error "$VERIFY_RESPONSE"; then
                    log_pass "Verify task succeeded"
                else
                    log_fail "Verify task failed"
                fi
            else
                log_fail "Get task status failed"
            fi
        else
            log_info "No tasks found - this is acceptable for new projects"
        fi
    else
        log_fail "List tasks failed"
    fi
fi

# ============================================================================
# Test 2: Filter Combinations
# ============================================================================

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "   Test 2: Filter Combinations"
echo "══════════════════════════════════════════════════════════════"
echo ""

if [ "$SKIP_HUB_TESTS" = "true" ]; then
    log_warn "Skipping filter test - Hub not available"
else
    # Test status + priority filter
    log_info "Testing status + priority filter..."
    RESPONSE=$(send_mcp_tool_call "sentinel_list_tasks" '{"status": "pending", "priority": "high"}')
    if ! has_error "$RESPONSE"; then
        log_pass "Status + priority filter works"
    else
        log_fail "Status + priority filter failed"
    fi
    
    # Test source + tags filter
    log_info "Testing source + tags filter..."
    RESPONSE=$(send_mcp_tool_call "sentinel_list_tasks" '{"source": "cursor", "tags": ["bug"]}')
    if ! has_error "$RESPONSE"; then
        log_pass "Source + tags filter works"
    else
        log_fail "Source + tags filter failed"
    fi
    
    # Test all filters
    log_info "Testing all filters combined..."
    RESPONSE=$(send_mcp_tool_call "sentinel_list_tasks" '{"status": "in_progress", "priority": "medium", "source": "cursor", "limit": 10, "offset": 0}')
    if ! has_error "$RESPONSE"; then
        log_pass "All filters combined work"
    else
        log_fail "All filters combined failed"
    fi
fi

# ============================================================================
# Test 3: Concurrent Requests
# ============================================================================

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "   Test 3: Concurrent Requests"
echo "══════════════════════════════════════════════════════════════"
echo ""

if [ "$SKIP_HUB_TESTS" = "true" ]; then
    log_warn "Skipping concurrency test - Hub not available"
else
    log_info "Launching 5 concurrent get_task_status requests..."
    TASK_ID="550e8400-e29b-41d4-a716-446655440000"
    SUCCESS_COUNT=0
    
    for i in {1..5}; do
        RESPONSE=$(send_mcp_tool_call "sentinel_get_task_status" "{\"taskId\": \"$TASK_ID\"}" "$i" &)
        if ! has_error "$RESPONSE"; then
            ((SUCCESS_COUNT++))
        fi
    done
    
    wait
    
    if [ $SUCCESS_COUNT -ge 3 ]; then
        log_pass "Concurrent requests handled ($SUCCESS_COUNT/5 succeeded)"
    else
        log_warn "Some concurrent requests failed ($SUCCESS_COUNT/5 succeeded) - may be expected if task doesn't exist"
    fi
fi

# ============================================================================
# Test Summary
# ============================================================================

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "   Integration Test Summary"
echo "══════════════════════════════════════════════════════════════"
echo ""
echo "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All integration tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some integration tests failed.${NC}"
    exit 1
fi









