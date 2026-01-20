#!/bin/bash
# Verify check_file_size returns REAL line count
set -e

SENTINEL="./sentinel"
TESTFILE="internal/cli/audit.go"

# Get real line count using wc -l (counts newlines)
WC_LINES=$(wc -l < "$TESTFILE" | tr -d ' ')

# Get MCP response
echo "Testing check_file_size for $TESTFILE (wc -l: $WC_LINES)"
RESPONSE=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"sentinel_check_file_size","arguments":{"file":"'$TESTFILE'"}}}' | \
  "$SENTINEL" mcp-server 2>/dev/null | head -1)

MCP_LINES=$(echo "$RESPONSE" | jq '.result.lines' 2>/dev/null || echo "null")

if [ "$MCP_LINES" = "null" ] || [ -z "$MCP_LINES" ]; then
  echo "FAIL: MCP did not return line count"
  echo "Response: $RESPONSE"
  exit 1
fi

# MCP counts lines (starts at 1, increments per newline), wc -l counts newlines
# So MCP should be wc -l + 1 (if file ends with newline) or equal (if it doesn't)
# The difference should be 0 or 1
DIFF=$((MCP_LINES - WC_LINES))
if [ ${DIFF#-} -gt 1 ]; then
  echo "FAIL: MCP lines ($MCP_LINES) differs too much from wc -l ($WC_LINES), diff=$DIFF"
  echo "Response: $RESPONSE"
  exit 1
fi

echo "PASS: MCP check_file_size reports $MCP_LINES lines (wc -l: $WC_LINES, diff: $DIFF)"
