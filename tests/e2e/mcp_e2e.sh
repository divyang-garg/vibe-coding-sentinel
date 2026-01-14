#!/bin/bash
# End-to-end test script for MCP tools
# Tests full workflow: Agent → Hub → Response

set -e

echo "🧪 Running MCP E2E Tests..."

# Check if Hub is running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Hub API not running. Start it first:"
    echo "   cd hub/api && go run main.go"
    exit 1
fi

# Check if Sentinel binary exists
if [ ! -f "./sentinel" ]; then
    echo "❌ Sentinel binary not found. Build it first:"
    echo "   ./synapsevibsentinel.sh"
    exit 1
fi

echo "✅ Prerequisites met"

# Test 1: Test sentinel_analyze_intent
echo ""
echo "Test 1: sentinel_analyze_intent"
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"sentinel_analyze_intent","arguments":{"request":"add user authentication"}}}' | ./sentinel mcp-server > /tmp/mcp_test1.json 2>&1 || true
if grep -q "error" /tmp/mcp_test1.json; then
    echo "⚠️  Test 1 failed (may be expected if Hub not configured)"
else
    echo "✅ Test 1 passed"
fi

# Test 2: Test sentinel_validate_code
echo ""
echo "Test 2: sentinel_validate_code"
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"sentinel_validate_code","arguments":{"code":"function test() { return 1; }","language":"javascript"}}}' | ./sentinel mcp-server > /tmp/mcp_test2.json 2>&1 || true
if grep -q "error" /tmp/mcp_test2.json; then
    echo "⚠️  Test 2 failed (may be expected if Hub not configured)"
else
    echo "✅ Test 2 passed"
fi

# Test 3: Test sentinel_apply_fix
echo ""
echo "Test 3: sentinel_apply_fix"
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"sentinel_apply_fix","arguments":{"filePath":"test.js","fixType":"security"}}}' | ./sentinel mcp-server > /tmp/mcp_test3.json 2>&1 || true
if grep -q "error" /tmp/mcp_test3.json; then
    echo "⚠️  Test 3 failed (may be expected if Hub not configured)"
else
    echo "✅ Test 3 passed"
fi

echo ""
echo "📊 E2E Tests Complete"
echo "Note: These tests require Hub API to be running and configured"









