#!/bin/bash
# Unit test: Verify Cursor rules use .md extension
# Run from project root: ./tests/unit/cursor_compatibility_test.sh

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    exit 1
}

echo ""
echo "Testing Cursor Rules File Extensions"
echo "===================================="
echo ""

# Test 1: Verify code uses .md extension
echo "Test 1: Verify code uses .md extension..."
MDC_COUNT=$(grep -c 'writeFile.*\.mdc' synapsevibsentinel.sh 2>/dev/null || echo "0")
MD_COUNT=$(grep -c 'writeFile.*\.cursor/rules.*\.md' synapsevibsentinel.sh 2>/dev/null || echo "0")

# Remove any whitespace/newlines
MDC_COUNT=$(echo "$MDC_COUNT" | tr -d '[:space:]')
MD_COUNT=$(echo "$MD_COUNT" | tr -d '[:space:]')

if [ "$MDC_COUNT" -eq 0 ] && [ "$MD_COUNT" -ge 10 ]; then
    log_pass "Code uses .md extension ($MD_COUNT instances)"
else
    log_fail "Code still uses .mdc extension ($MDC_COUNT instances) or insufficient .md references ($MD_COUNT)"
fi

# Test 2: Functional test - run init and check files
echo ""
echo "Test 2: Functional test - init command creates .md files..."
TEST_DIR="/tmp/sentinel_test_$(date +%s)"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

export SENTINEL_STACK=web
export SENTINEL_DB=sql
"$PROJECT_ROOT/sentinel" init --non-interactive > /dev/null 2>&1

if [ -f ".cursor/rules/00-constitution.md" ] && \
   [ -f ".cursor/rules/01-firewall.md" ] && \
   [ -f ".cursor/rules/web.md" ] && \
   [ ! -f ".cursor/rules/00-constitution.mdc" ]; then
    log_pass "Init command creates .md files correctly"
else
    log_fail "Init command did not create .md files correctly"
fi

# Cleanup
cd "$PROJECT_ROOT"
rm -rf "$TEST_DIR"

echo ""
echo "All tests passed!"

