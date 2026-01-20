#!/bin/bash
# Real MCP test - spawns actual process, sends real JSON, reads real output
set -e

SENTINEL="./sentinel"
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Build fresh binary
echo "Building sentinel binary..."
go build -o "$SENTINEL" ./cmd/sentinel

# Test 1: Initialize request
echo ""
echo "=== Test 1: Initialize ==="
# MCP server reads from stdin until EOF, then responds
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | \
  "$SENTINEL" mcp-server 2>/dev/null > "$TMPDIR/init_response.json" || true

# Verify response is valid JSON with expected fields
if ! jq -e '.result.protocolVersion' "$TMPDIR/init_response.json" > /dev/null 2>&1; then
  echo "FAIL: Initialize did not return protocolVersion"
  echo "Response content:"
  cat "$TMPDIR/init_response.json"
  exit 1
fi
echo "PASS: Initialize returns valid response"

# Test 2: Tools list
echo ""
echo "=== Test 2: Tools List ==="
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | \
  "$SENTINEL" mcp-server 2>/dev/null > "$TMPDIR/tools_response.json" || true

# Verify at least 7 tools returned
TOOL_COUNT=$(jq '.result.tools | length' "$TMPDIR/tools_response.json" 2>/dev/null || echo "0")
if [ "$TOOL_COUNT" -lt 7 ]; then
  echo "FAIL: Expected at least 7 tools, got $TOOL_COUNT"
  cat "$TMPDIR/tools_response.json"
  exit 1
fi
echo "PASS: Tools list returns $TOOL_COUNT tools"

# Test 3: Audit tool (verify real scan happens)
echo ""
echo "=== Test 3: Audit Tool ==="
cat > "$TMPDIR/test.js" << 'EOF'
const secret = "sk_live_1234567890abcdef";
eval(userInput);
EOF

echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"sentinel_audit","arguments":{"path":"'$TMPDIR'"}}}' | \
  "$SENTINEL" mcp-server 2>/dev/null > "$TMPDIR/audit_response.json" || true

# Verify findings > 0
FINDINGS=$(jq '.result.findings' "$TMPDIR/audit_response.json" 2>/dev/null || echo "0")
if [ "$FINDINGS" -eq 0 ] || [ "$FINDINGS" = "null" ]; then
  echo "FAIL: Audit returned 0 findings for vulnerable file"
  cat "$TMPDIR/audit_response.json"
  exit 1
fi
echo "PASS: Audit detected $FINDINGS findings"

echo ""
echo "All MCP process tests passed"
