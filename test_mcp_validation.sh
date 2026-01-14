#!/bin/bash

# Test MCP tool validation by sending invalid enum values
echo "Testing MCP tool validation..."

# Test invalid language in validate_code
echo "Testing invalid language in sentinel_validate_code..."
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "sentinel_validate_code", "arguments": {"code": "test", "language": "invalid_lang"}}}' | timeout 5 ./sentinel mcp-server | jq '.error.message'

# Test invalid sort_by in get_patterns  
echo "Testing invalid sort_by in sentinel_get_patterns..."
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "sentinel_get_patterns", "arguments": {"sort_by": "invalid_sort"}}}' | timeout 5 ./sentinel mcp-server | jq '.error.message'

echo "Validation tests completed."
