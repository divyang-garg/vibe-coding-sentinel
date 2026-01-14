#!/bin/bash

echo "=== MCP Protocol Compliance Test Suite ==="
echo

# Test 1: MCP Initialization
echo "1. Testing MCP Initialization..."
INIT_RESPONSE=$(printf '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}\n' | ./sentinel mcp-server)
if echo "$INIT_RESPONSE" | grep -q '"jsonrpc":"2.0"'; then
    echo "✅ MCP initialization: PASS"
else
    echo "❌ MCP initialization: FAIL"
fi

# Test 2: Tools List
echo "2. Testing tools/list..."
TOOLS_RESPONSE=$(printf '{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}\n' | ./sentinel mcp-server)
TOOLS_COUNT=$(echo "$TOOLS_RESPONSE" | jq '.result.tools | length' 2>/dev/null || echo "0")
if [ "$TOOLS_COUNT" -eq 10 ]; then
    echo "✅ Tools list (10 tools): PASS"
else
    echo "❌ Tools list ($TOOLS_COUNT tools): FAIL"
fi

# Test 3: Parameter Validation
echo "3. Testing parameter validation..."
VALIDATE_RESPONSE=$(printf '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "sentinel_validate_code", "arguments": {"code": "test", "language": "invalid"}}}\n' | ./sentinel mcp-server)
if echo "$VALIDATE_RESPONSE" | grep -q '"code":-32602'; then
    echo "✅ Parameter validation: PASS"
else
    echo "❌ Parameter validation: FAIL"
fi

# Test 4: Error Handling
echo "4. Testing error handling..."
ERROR_RESPONSE=$(printf '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "nonexistent_tool"}}\n' | ./sentinel mcp-server)
if echo "$ERROR_RESPONSE" | grep -q '"code":-32601'; then
    echo "✅ Error handling: PASS"
else
    echo "❌ Error handling: FAIL"
fi

# Test 5: JSON-RPC Format
echo "5. Testing JSON-RPC format..."
JSON_RESPONSE=$(printf '{"jsonrpc": "2.0", "id": 5, "method": "tools/list", "params": {}}\n' | ./sentinel mcp-server)
if echo "$JSON_RESPONSE" | jq -e '.jsonrpc == "2.0" and .id == 5' >/dev/null 2>&1; then
    echo "✅ JSON-RPC format: PASS"
else
    echo "❌ JSON-RPC format: FAIL"
fi

echo
echo "=== MCP Compliance Test Complete ==="
