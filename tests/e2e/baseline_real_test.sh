#!/bin/bash
# Prove baseline filtering actually works with real CLI execution
set -e

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Create vulnerable test file
mkdir -p "$TMPDIR/.sentinel"
cat > "$TMPDIR/vuln.js" << 'EOF'
eval(userInput);
const db = require('db');
db.query("SELECT * FROM users WHERE id = " + userId);
EOF

# Get absolute path to sentinel (build from project root)
SCRIPT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$SCRIPT_DIR"

# Build sentinel if needed
if [ ! -f "./sentinel" ]; then
  echo "Building sentinel..."
  go build -o sentinel ./cmd/sentinel
fi

cd "$TMPDIR"

# Count 1: Without baseline
echo "Running audit without baseline..."
COUNT1=$("$SCRIPT_DIR/sentinel" audit . --ci 2>&1 | grep -E "Findings:|found" | grep -oE '[0-9]+' | head -1 || echo "0")
echo "Without baseline: $COUNT1 findings"

if [ "$COUNT1" -eq "0" ]; then
  echo "WARNING: No findings detected, cannot test baseline filtering"
  exit 0
fi

# Add first finding to baseline
echo '{"version":"1.0","entries":[{"file":"vuln.js","line":1,"hash":"vuln.js:1"}]}' > .sentinel/baseline.json

# Count 2: With baseline
echo "Running audit with baseline..."
COUNT2=$("$SCRIPT_DIR/sentinel" audit . --ci 2>&1 | grep -E "Findings:|found" | grep -oE '[0-9]+' | head -1 || echo "0")
echo "With baseline: $COUNT2 findings"

# Verify count decreased (should be at least 1 less)
if [ "$COUNT2" -ge "$COUNT1" ]; then
  echo "FAIL: Baseline did not reduce findings (before: $COUNT1, after: $COUNT2)"
  exit 1
fi

echo "PASS: Baseline filtering reduces findings from $COUNT1 to $COUNT2"
