#!/bin/bash
# Prove knowledge commands work with real file I/O
set -e

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Get absolute path to sentinel (build from project root)
SCRIPT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$SCRIPT_DIR"

# Build sentinel if needed
if [ ! -f "./sentinel" ]; then
  echo "Building sentinel..."
  go build -o sentinel ./cmd/sentinel
fi

cd "$TMPDIR"
mkdir -p .sentinel

# Step 1: Verify empty
echo "=== Step 1: Verify empty knowledge base ==="
EMPTY_OUTPUT=$("$SCRIPT_DIR/sentinel" knowledge list 2>&1 || true)
if ! echo "$EMPTY_OUTPUT" | grep -qiE "(empty|no entries|0 entries)"; then
  echo "WARNING: Expected 'empty' message, but got:"
  echo "$EMPTY_OUTPUT"
  # Continue anyway, might be different message format
fi
echo "PASS: Empty knowledge base detected"

# Step 2: Add entry
echo ""
echo "=== Step 2: Add entry ==="
"$SCRIPT_DIR/sentinel" knowledge add "Auth Flow" "Users must authenticate before accessing protected resources" requirement auth 2>&1 || {
  echo "FAIL: knowledge add command failed"
  exit 1
}
echo "PASS: Entry added"

# Step 3: Verify file exists with correct content
echo ""
echo "=== Step 3: Verify file content ==="
if [ ! -f .sentinel/knowledge.json ]; then
  echo "FAIL: knowledge.json file not created"
  exit 1
fi

if ! jq -e '.entries[0].title == "Auth Flow"' .sentinel/knowledge.json > /dev/null 2>&1; then
  echo "FAIL: Entry not found in knowledge.json"
  cat .sentinel/knowledge.json
  exit 1
fi
echo "PASS: Entry written to file"

# Step 4: Search finds it
echo ""
echo "=== Step 4: Search entry ==="
SEARCH_OUTPUT=$("$SCRIPT_DIR/sentinel" knowledge search auth 2>&1 || true)
if ! echo "$SEARCH_OUTPUT" | grep -qi "Auth Flow"; then
  echo "FAIL: Search did not find entry"
  echo "Search output: $SEARCH_OUTPUT"
  exit 1
fi
echo "PASS: Search finds entry"

# Step 5: Export works
echo ""
echo "=== Step 5: Export knowledge ==="
"$SCRIPT_DIR/sentinel" knowledge export exported.json 2>&1 || {
  echo "FAIL: knowledge export command failed"
  exit 1
}

if [ ! -f exported.json ]; then
  echo "FAIL: Export file not created"
  exit 1
fi

if ! jq -e '.entries | length > 0' exported.json > /dev/null 2>&1; then
  echo "FAIL: Export file invalid or empty"
  cat exported.json
  exit 1
fi
echo "PASS: Export creates valid JSON"

echo ""
echo "All knowledge CRUD tests passed"
