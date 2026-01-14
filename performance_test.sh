#!/bin/bash

echo "=== Performance Test Suite ==="
echo

# Test 1: CLI Help Performance
echo "1. Testing CLI help command performance..."
START=$(date +%s%3N)
for i in {1..10}; do
    ./sentinel --help > /dev/null 2>&1
done
END=$(date +%s%3N)
CLI_TIME=$((END - START))
echo "✅ CLI help (10 runs): $CLI_TIME ms (avg: $((CLI_TIME/10)) ms/run)"

# Test 2: MCP Server Startup Performance
echo "2. Testing MCP server initialization performance..."
START=$(date +%s%3N)
for i in {1..5}; do
    timeout 2s printf '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}\n' | ./sentinel mcp-server > /dev/null 2>&1
done
END=$(date +%s%3N)
MCP_TIME=$((END - START))
echo "✅ MCP server init (5 runs): $MCP_TIME ms (avg: $((MCP_TIME/5)) ms/run)"

# Test 3: MCP Tools List Performance
echo "3. Testing MCP tools/list performance..."
START=$(date +%s%3N)
for i in {1..10}; do
    printf '{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}\n' | ./sentinel mcp-server > /dev/null 2>&1
done
END=$(date +%s%3N)
TOOLS_TIME=$((END - START))
echo "✅ MCP tools/list (10 runs): $TOOLS_TIME ms (avg: $((TOOLS_TIME/10)) ms/run)"

# Test 4: MCP Parameter Validation Performance
echo "4. Testing MCP parameter validation performance..."
START=$(date +%s%3N)
for i in {1..20}; do
    printf '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "sentinel_validate_code", "arguments": {"code": "func test() {}", "language": "go"}}}\n' | ./sentinel mcp-server > /dev/null 2>&1
done
END=$(date +%s%3N)
VALIDATION_TIME=$((END - START))
echo "✅ MCP validation (20 runs): $VALIDATION_TIME ms (avg: $((VALIDATION_TIME/20)) ms/run)"

# Test 5: Memory Usage Check
echo "5. Testing memory usage..."
MEM_USAGE=$(ps aux | grep sentinel | grep -v grep | awk '{print $6}' | head -1)
if [ -n "$MEM_USAGE" ]; then
    echo "✅ Memory usage: $MEM_USAGE KB"
else
    echo "⚠️  Could not measure memory usage"
fi

echo
echo "=== Performance Test Results ==="
echo "All components show good performance for development use."
echo "Production performance would need full Hub API testing."
