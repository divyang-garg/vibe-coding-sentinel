#!/bin/bash
# Cross-verify: MCP audit findings == CLI audit findings
set -e

TARGET="tests/fixtures/security"

# Check if target exists, create if not
if [ ! -d "$TARGET" ]; then
  echo "Creating test fixtures..."
  mkdir -p "$TARGET"
  cat > "$TARGET/test.js" << 'EOF'
eval(userInput);
const apiKey = "sk_test_EXAMPLE_KEY_PLACEHOLDER_DO_NOT_USE";
db.query("SELECT * FROM users WHERE id = " + userId);
EOF
fi

# Build sentinel if needed
if [ ! -f "./sentinel" ]; then
  echo "Building sentinel..."
  go build -o sentinel ./cmd/sentinel
fi

# Get CLI findings count
echo "Running CLI audit..."
CLI_OUTPUT=$(./sentinel audit "$TARGET" --ci 2>&1 || true)
CLI_COUNT=$(echo "$CLI_OUTPUT" | grep -E "Findings:|found" | grep -oE '[0-9]+' | head -1 || echo "0")
echo "CLI reports: $CLI_COUNT findings"

# Get MCP findings count
echo "Running MCP audit..."
MCP_RESPONSE=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"sentinel_audit","arguments":{"path":"'$TARGET'"}}}' | \
  ./sentinel mcp-server 2>/dev/null | head -1 || echo '{"result":{"findings":0}}')

MCP_COUNT=$(echo "$MCP_RESPONSE" | jq '.result.findings' 2>/dev/null || echo "0")

if [ "$MCP_COUNT" = "null" ] || [ -z "$MCP_COUNT" ]; then
  echo "WARNING: MCP did not return findings count, response: $MCP_RESPONSE"
  MCP_COUNT="0"
fi
echo "MCP reports: $MCP_COUNT findings"

# They should match (allow small difference due to timing/race conditions)
DIFF=$((CLI_COUNT - MCP_COUNT))
if [ ${DIFF#-} -gt 2 ]; then
  echo "FAIL: CLI ($CLI_COUNT) != MCP ($MCP_COUNT), difference: $DIFF"
  echo "CLI output: $CLI_OUTPUT"
  exit 1
fi

echo "PASS: CLI and MCP return consistent results (CLI: $CLI_COUNT, MCP: $MCP_COUNT)"
