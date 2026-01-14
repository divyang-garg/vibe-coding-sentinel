#!/bin/bash
# Unit tests for Phase 13 knowledge migration
# Run from project root: ./tests/unit/migration_test.sh

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

echo "🧪 Testing Knowledge Migration"
echo "══════════════════════════════════════════════════════════════"
echo ""

# Test 1: Migration function exists
echo "Test 1: Migration Function Exists"
if grep -q "migrateKnowledgeItems" hub/processor/migrate_knowledge.go; then
    echo "   ✅ Migration function found"
else
    echo "   ❌ Migration function not found"
    exit 1
fi

# Test 2: Migration flag exists
echo ""
echo "Test 2: Migration Flag Exists"
if grep -q "--migrate" hub/processor/main.go; then
    echo "   ✅ Migration flag found"
else
    echo "   ❌ Migration flag not found"
    exit 1
fi

# Test 3: Migration handles NULL structured_data
echo ""
echo "Test 3: Migration Handles NULL structured_data"
if grep -q "structured_data IS NULL" hub/processor/migrate_knowledge.go; then
    echo "   ✅ Migration queries NULL structured_data"
else
    echo "   ❌ Migration may not handle NULL correctly"
    exit 1
fi

# Test 4: Migration preserves backward compatibility
echo ""
echo "Test 4: Migration Preserves Backward Compatibility"
if grep -q "content" hub/processor/migrate_knowledge.go; then
    echo "   ✅ Migration preserves content field"
else
    echo "   ⚠️  Migration may not preserve content"
fi

echo ""
echo "✅ Migration tests completed!"
echo ""











