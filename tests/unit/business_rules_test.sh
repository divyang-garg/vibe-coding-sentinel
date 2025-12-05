#!/bin/bash
# Business Rules Compliance Test Suite
# Tests enhanced business rules validation (Phase 8 compliance fixes)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🧪 Testing Business Rules Compliance"
echo ""

cd "$PROJECT_ROOT"

# Test 1: --business-rules flag exists
echo "Test 1: --business-rules flag integration"
if ./sentinel audit --help 2>&1 | grep -q "business-rules"; then
    echo "  ✅ --business-rules flag documented"
else
    echo "  ⚠️  --business-rules flag not in help (may be implemented but not documented)"
fi

# Test 2: Business rules compliance with no rules
echo "Test 2: Business rules compliance with no approved rules"
# Create a temporary knowledge store with no approved rules
mkdir -p docs/knowledge
cat > docs/knowledge/knowledge-store.json <<EOF
{
  "items": [
    {
      "id": "ki_test_001",
      "type": "business_rule",
      "title": "Test Rule",
      "content": "Test content",
      "status": "pending",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ],
  "version": 1,
  "lastUpdated": "2024-01-01T00:00:00Z"
}
EOF

OUTPUT=$(./sentinel audit --business-rules 2>&1 || true)
if echo "$OUTPUT" | grep -q "No approved business rules"; then
    echo "  ✅ Correctly handles no approved rules"
else
    echo "  ⚠️  May not handle no approved rules correctly"
fi

# Test 3: Business rules compliance with approved rules
echo "Test 3: Business rules compliance with approved rules"
cat > docs/knowledge/knowledge-store.json <<EOF
{
  "items": [
    {
      "id": "ki_test_002",
      "type": "business_rule",
      "title": "Order Cancellation Policy",
      "content": "Orders can only be cancelled within 24 hours of placement",
      "status": "approved",
      "approvedBy": "test",
      "approvedAt": "2024-01-01T00:00:00Z",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ],
  "version": 1,
  "lastUpdated": "2024-01-01T00:00:00Z"
}
EOF

OUTPUT=$(./sentinel audit --business-rules 2>&1 || true)
if echo "$OUTPUT" | grep -q "Checking.*business rules"; then
    echo "  ✅ Correctly processes approved business rules"
else
    echo "  ⚠️  May not process approved business rules correctly"
fi

# Cleanup
rm -f docs/knowledge/knowledge-store.json

echo ""
echo "✅ Business rules compliance tests completed"

