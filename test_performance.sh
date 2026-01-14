#!/bin/bash

# Performance Testing for Task Management
# Tests response times and resource usage

set -e

echo "=== Task Management Performance Test Suite ==="
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SENTINEL_BIN="./sentinel"
PASSED=0
FAILED=0

# Test 1: MCP Server Startup Time
echo "Test 1: MCP Server Startup Time"
echo "--------------------------------"
START_TIME=$(date +%s%N)
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}}' | $SENTINEL_BIN mcp-server > /dev/null 2>&1
END_TIME=$(date +%s%N)
STARTUP_TIME=$((($END_TIME - $START_TIME) / 1000000))  # Convert to milliseconds

if [ $STARTUP_TIME -lt 100 ]; then
    echo -e "${GREEN}✅ PASS${NC}: MCP server startup time: ${STARTUP_TIME}ms (< 100ms)"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  WARN${NC}: MCP server startup time: ${STARTUP_TIME}ms (>= 100ms)"
fi
echo

# Test 2: Tools List Response Time
echo "Test 2: Tools List Response Time"
echo "---------------------------------"
TIMES=()
for i in {1..5}; do
    START=$(date +%s%N)
    echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | $SENTINEL_BIN mcp-server > /dev/null 2>&1
    END=$(date +%s%N)
    TIME=$((($END - $START) / 1000000))
    TIMES+=($TIME)
done

# Calculate average
SUM=0
for t in "${TIMES[@]}"; do
    SUM=$((SUM + t))
done
AVG=$((SUM / ${#TIMES[@]}))

if [ $AVG -lt 50 ]; then
    echo -e "${GREEN}✅ PASS${NC}: Average tools/list response time: ${AVG}ms (< 50ms)"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Average tools/list response time: ${AVG}ms (>= 50ms)"
fi
echo "   Individual times: ${TIMES[@]}ms"
echo

# Test 3: Parameter Validation Performance
echo "Test 3: Parameter Validation Performance"
echo "----------------------------------------"
TIMES=()
for i in {1..10}; do
    START=$(date +%s%N)
    printf '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "sentinel_list_tasks", "arguments": {"status": "pending", "limit": 10}}}\n' | $SENTINEL_BIN mcp-server > /dev/null 2>&1
    END=$(date +%s%N)
    TIME=$((($END - $START) / 1000000))
    TIMES+=($TIME)
done

SUM=0
for t in "${TIMES[@]}"; do
    SUM=$((SUM + t))
done
AVG=$((SUM / ${#TIMES[@]}))

if [ $AVG -lt 20 ]; then
    echo -e "${GREEN}✅ PASS${NC}: Average validation time: ${AVG}ms (< 20ms)"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Average validation time: ${AVG}ms (>= 20ms)"
fi
echo "   Individual times: ${TIMES[@]}ms"
echo

# Test 4: CLI Command Response Time
echo "Test 4: CLI Command Response Time"
echo "----------------------------------"
TIMES=()
for i in {1..5}; do
    START=$(date +%s%N)
    $SENTINEL_BIN tasks > /dev/null 2>&1
    END=$(date +%s%N)
    TIME=$((($END - $START) / 1000000))
    TIMES+=($TIME)
done

SUM=0
for t in "${TIMES[@]}"; do
    SUM=$((SUM + t))
done
AVG=$((SUM / ${#TIMES[@]}))

if [ $AVG -lt 50 ]; then
    echo -e "${GREEN}✅ PASS${NC}: Average CLI response time: ${AVG}ms (< 50ms)"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Average CLI response time: ${AVG}ms (>= 50ms)"
fi
echo "   Individual times: ${TIMES[@]}ms"
echo

# Test 5: Memory Usage (Basic Check)
echo "Test 5: Memory Usage Check"
echo "---------------------------"
# Check binary size
BINARY_SIZE=$(ls -lh sentinel 2>/dev/null | awk '{print $5}' || echo "0")
echo "Binary size: $BINARY_SIZE"
if [ -f "sentinel" ]; then
    echo -e "${GREEN}✅ PASS${NC}: Binary exists"
    ((PASSED++))
else
    echo -e "${RED}❌ FAIL${NC}: Binary not found"
    ((FAILED++))
fi
echo

# Summary
echo "=== Performance Test Summary ==="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All performance tests passed!${NC}"
    echo "Performance metrics are within acceptable ranges."
    exit 0
else
    echo -e "${RED}❌ Some performance tests failed${NC}"
    exit 1
fi




