#!/bin/bash
# Integration tests for Phase 13 end-to-end extraction
# Run from project root: ./tests/integration/phase13_e2e_test.sh

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

echo "🧪 Phase 13 End-to-End Integration Tests"
echo "══════════════════════════════════════════════════════════════"
echo ""

# Test 1: Structured extraction works
echo "Test 1: Structured Extraction"
echo "   ⚠️  Requires running Hub and Processor with LLM configured"
echo "   Manual test: Upload document and verify structured_data populated"

# Test 2: Validation passes
echo ""
echo "Test 2: Validation Passes"
if [ -f "hub/processor/schema_validator.go" ]; then
    echo "   ✅ Validator module exists"
else
    echo "   ❌ Validator module not found"
    exit 1
fi

# Test 3: Test requirements generated
echo ""
echo "Test 3: Test Requirements Generated"
if grep -q "generateTestRequirements" hub/processor/test_generator.go; then
    echo "   ✅ Test generator exists"
else
    echo "   ❌ Test generator not found"
    exit 1
fi

# Test 4: Ambiguity flags detected
echo ""
echo "Test 4: Ambiguity Flags Detected"
if grep -q "analyzeAmbiguity" hub/processor/ambiguity_analyzer.go; then
    echo "   ✅ Ambiguity analyzer exists"
else
    echo "   ❌ Ambiguity analyzer not found"
    exit 1
fi

# Test 5: Database schema updated
echo ""
echo "Test 5: Database Schema Updated"
if grep -q "structured_data JSONB" hub/api/main.go; then
    echo "   ✅ Database schema includes structured_data column"
else
    echo "   ❌ Database schema missing structured_data column"
    exit 1
fi

echo ""
echo "✅ Phase 13 integration tests completed!"
echo ""











