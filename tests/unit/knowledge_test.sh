#!/bin/bash
# Unit tests for Sentinel knowledge management functionality
# Run from project root: ./tests/unit/knowledge_test.sh

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    ((TESTS_FAILED++))
}

cleanup_lock() {
    rm -f /tmp/sentinel.lock
}

echo ""
echo "=============================================="
echo "   Knowledge Management Unit Tests"
echo "=============================================="
echo ""

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SENTINEL="$PROJECT_ROOT/sentinel"
TEST_DIR=$(mktemp -d)

# ============================================================================
# Setup: Create test knowledge store
# ============================================================================

echo "Setting up test data..."

mkdir -p "$TEST_DIR/docs/knowledge"
cat > "$TEST_DIR/docs/knowledge/knowledge-store.json" << 'EOF'
{
  "items": [
    {
      "id": "ki_test001",
      "type": "business_rule",
      "title": "Test Rule",
      "content": "This is a test business rule",
      "source": "test.txt",
      "confidence": 0.95,
      "status": "pending",
      "createdAt": "2024-01-01T00:00:00Z"
    },
    {
      "id": "ki_test002",
      "type": "entity",
      "title": "TestEntity",
      "content": "This is a test entity",
      "source": "test.txt",
      "confidence": 0.85,
      "status": "pending",
      "createdAt": "2024-01-01T00:00:00Z"
    },
    {
      "id": "ki_test003",
      "type": "glossary",
      "title": "TestTerm",
      "content": "This is a test glossary term",
      "source": "test.txt",
      "confidence": 0.92,
      "status": "approved",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ],
  "lastUpdated": "2024-01-01T00:00:00Z",
  "version": 1
}
EOF

cd "$TEST_DIR"

# ============================================================================
# Test: Knowledge help command
# ============================================================================

echo ""
echo "Testing knowledge help..."
cleanup_lock

OUTPUT=$("$SENTINEL" knowledge 2>&1)
if echo "$OUTPUT" | grep -q "Usage: sentinel knowledge"; then
    log_pass "Help shows usage"
else
    log_fail "Help doesn't show usage"
fi

# ============================================================================
# Test: Knowledge list command
# ============================================================================

echo ""
echo "Testing knowledge list..."
cleanup_lock

OUTPUT=$("$SENTINEL" knowledge list 2>&1)
if echo "$OUTPUT" | grep -q "Business Rule"; then
    log_pass "Lists business rules"
else
    log_fail "Doesn't list business rules"
fi

if echo "$OUTPUT" | grep -q "Test Rule"; then
    log_pass "Shows item title"
else
    log_fail "Doesn't show item title"
fi

# ============================================================================
# Test: Knowledge list --pending
# ============================================================================

echo ""
echo "Testing knowledge list --pending..."
cleanup_lock

OUTPUT=$("$SENTINEL" knowledge list --pending 2>&1)
if echo "$OUTPUT" | grep -q "Test Rule"; then
    log_pass "Shows pending items"
else
    log_fail "Doesn't filter pending items"
fi

# ============================================================================
# Test: Knowledge stats command
# ============================================================================

echo ""
echo "Testing knowledge stats..."
cleanup_lock

OUTPUT=$("$SENTINEL" knowledge stats 2>&1)
if echo "$OUTPUT" | grep -q "Business Rules"; then
    log_pass "Shows business rules count"
else
    log_fail "Doesn't show stats"
fi

if echo "$OUTPUT" | grep -q "Average Confidence"; then
    log_pass "Shows average confidence"
else
    log_fail "Doesn't show confidence"
fi

# ============================================================================
# Test: Knowledge approve command
# ============================================================================

echo ""
echo "Testing knowledge approve..."
cleanup_lock

OUTPUT=$("$SENTINEL" knowledge approve ki_test001 2>&1)
if echo "$OUTPUT" | grep -q "Approved"; then
    log_pass "Approves item"
else
    log_fail "Failed to approve item"
fi

# Verify status changed
OUTPUT=$("$SENTINEL" knowledge list --approved 2>&1)
if echo "$OUTPUT" | grep -q "Test Rule"; then
    log_pass "Item status changed to approved"
else
    log_fail "Item status not changed"
fi

# ============================================================================
# Test: Knowledge reject command
# ============================================================================

echo ""
echo "Testing knowledge reject..."
cleanup_lock

OUTPUT=$("$SENTINEL" knowledge reject ki_test002 2>&1)
if echo "$OUTPUT" | grep -q "Rejected"; then
    log_pass "Rejects item"
else
    log_fail "Failed to reject item"
fi

# ============================================================================
# Test: Knowledge activate command
# ============================================================================

echo ""
echo "Testing knowledge activate..."
cleanup_lock

OUTPUT=$("$SENTINEL" knowledge activate 2>&1)
if echo "$OUTPUT" | grep -q "Knowledge activated"; then
    log_pass "Activates knowledge"
else
    log_fail "Failed to activate"
fi

if [[ -f ".cursor/rules/business-knowledge.md" ]]; then
    log_pass "Creates Cursor rule file"
else
    log_fail "Doesn't create rule file"
fi

# Check rule file content
if grep -q "Business Knowledge" .cursor/rules/business-knowledge.md 2>/dev/null; then
    log_pass "Rule file has content"
else
    log_fail "Rule file is empty"
fi

# ============================================================================
# Test: Knowledge approve --all
# ============================================================================

echo ""
echo "Testing knowledge approve --all..."
cleanup_lock

# Reset test data
cat > "$TEST_DIR/docs/knowledge/knowledge-store.json" << 'EOF'
{
  "items": [
    {
      "id": "ki_high001",
      "type": "business_rule",
      "title": "High Confidence Rule",
      "content": "Test",
      "source": "test.txt",
      "confidence": 0.95,
      "status": "pending",
      "createdAt": "2024-01-01T00:00:00Z"
    },
    {
      "id": "ki_low001",
      "type": "business_rule",
      "title": "Low Confidence Rule",
      "content": "Test",
      "source": "test.txt",
      "confidence": 0.75,
      "status": "pending",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ],
  "version": 1
}
EOF

OUTPUT=$("$SENTINEL" knowledge approve --all 2>&1)
if echo "$OUTPUT" | grep -q "Auto-approved 1"; then
    log_pass "Auto-approves only high confidence items"
else
    log_fail "Auto-approve doesn't filter by confidence"
fi

# ============================================================================
# Cleanup
# ============================================================================

rm -rf "$TEST_DIR"

# ============================================================================
# Summary
# ============================================================================

echo ""
echo "=============================================="
echo "   Knowledge Management Test Results"
echo "=============================================="
echo ""
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"

TOTAL=$((TESTS_PASSED + TESTS_FAILED))
if [[ $TOTAL -gt 0 ]]; then
    PERCENT=$((TESTS_PASSED * 100 / TOTAL))
    echo "Success Rate: ${PERCENT}%"
fi
echo ""

if [[ $TESTS_FAILED -gt 0 ]]; then
    exit 1
fi

exit 0

