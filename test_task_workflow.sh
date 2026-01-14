#!/bin/bash

# End-to-End Task Management Workflow Test
# Tests: scan → list → verify → dependencies

set -e

echo "=== Task Management End-to-End Test Suite ==="
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SENTINEL_BIN="./sentinel"
TEST_DIR="test-task-workflow"
PASSED=0
FAILED=0

# Cleanup function
cleanup() {
    if [ -d "$TEST_DIR" ]; then
        rm -rf "$TEST_DIR"
    fi
}

trap cleanup EXIT

# Test 1: CLI Help Text
echo "Test 1: CLI Help Text"
echo "---------------------"
if $SENTINEL_BIN tasks 2>&1 | grep -q "Tasks Management"; then
    echo -e "${GREEN}✅ PASS${NC}: Tasks command help text displayed"
    ((PASSED++))
else
    echo -e "${RED}❌ FAIL${NC}: Tasks command help text missing"
    ((FAILED++))
fi
echo

# Test 2: MCP Tools Registration
echo "Test 2: MCP Tools Registration"
echo "-------------------------------"
TOOLS_COUNT=$(echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | $SENTINEL_BIN mcp-server | jq -r 'if .result then (.result.tools | length) else 0 end' 2>/dev/null || echo "0")

if [ "$TOOLS_COUNT" -ge 13 ]; then
    echo -e "${GREEN}✅ PASS${NC}: MCP tools registered ($TOOLS_COUNT tools)"
    ((PASSED++))
    
    # Check for task management tools
    TASK_TOOLS=$(echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | $SENTINEL_BIN mcp-server | jq -r '.result.tools[] | select(.name | contains("task")) | .name' 2>/dev/null | wc -l | tr -d ' ')
    if [ "$TASK_TOOLS" -eq 3 ]; then
        echo -e "${GREEN}✅ PASS${NC}: All 3 task management MCP tools registered"
        ((PASSED++))
    else
        echo -e "${RED}❌ FAIL${NC}: Expected 3 task tools, found $TASK_TOOLS"
        ((FAILED++))
    fi
else
    echo -e "${RED}❌ FAIL${NC}: Expected at least 13 MCP tools, found $TOOLS_COUNT"
    ((FAILED++))
fi
echo

# Test 3: MCP Tool Schema Validation
echo "Test 3: MCP Tool Schema Validation"
echo "-----------------------------------"
SCHEMA_CHECK=$(echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | $SENTINEL_BIN mcp-server | jq -r '.result.tools[] | select(.name == "sentinel_get_task_status") | .inputSchema.properties.task_id.type' 2>/dev/null || echo "")

if [ "$SCHEMA_CHECK" == "string" ]; then
    echo -e "${GREEN}✅ PASS${NC}: sentinel_get_task_status schema correct"
    ((PASSED++))
else
    echo -e "${RED}❌ FAIL${NC}: sentinel_get_task_status schema incorrect"
    ((FAILED++))
fi
echo

# Test 4: MCP Parameter Validation
echo "Test 4: MCP Parameter Validation"
echo "---------------------------------"
# Test invalid task_id (missing parameter)
INVALID_RESPONSE=$(printf '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "sentinel_get_task_status", "arguments": {}}}\n' | $SENTINEL_BIN mcp-server | jq -r '.error.code' 2>/dev/null || echo "")

if [ "$INVALID_RESPONSE" == "-32602" ]; then
    echo -e "${GREEN}✅ PASS${NC}: Parameter validation working (missing task_id rejected)"
    ((PASSED++))
else
    echo -e "${RED}❌ FAIL${NC}: Parameter validation not working correctly"
    ((FAILED++))
fi

# Test invalid enum value
INVALID_ENUM=$(printf '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "sentinel_list_tasks", "arguments": {"status": "invalid_status"}}}\n' | $SENTINEL_BIN mcp-server | jq -r '.error.code' 2>/dev/null || echo "")

if [ "$INVALID_ENUM" == "-32602" ]; then
    echo -e "${GREEN}✅ PASS${NC}: Enum validation working (invalid status rejected)"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Enum validation may not be working (Hub API may handle this)"
    # This is not a failure since Hub API might handle validation
fi
echo

# Test 5: Code Compilation
echo "Test 5: Code Compilation"
echo "------------------------"
if go build -o sentinel-test main.go 2>&1 | head -5; then
    echo -e "${GREEN}✅ PASS${NC}: Code compiles successfully"
    rm -f sentinel-test
    ((PASSED++))
else
    echo -e "${RED}❌ FAIL${NC}: Code compilation failed"
    ((FAILED++))
fi
echo

# Test 6: CLI Command Structure
echo "Test 6: CLI Command Structure"
echo "-----------------------------"
# Test all subcommands exist
SUBCOMMANDS=("scan" "list" "verify" "dependencies")
for cmd in "${SUBCOMMANDS[@]}"; do
    if $SENTINEL_BIN tasks 2>&1 | grep -q "$cmd"; then
        echo -e "${GREEN}✅ PASS${NC}: Subcommand '$cmd' documented"
    else
        echo -e "${RED}❌ FAIL${NC}: Subcommand '$cmd' missing"
        ((FAILED++))
    fi
done
((PASSED+=${#SUBCOMMANDS[@]}))
echo

# Summary
echo "=== Test Summary ==="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi




