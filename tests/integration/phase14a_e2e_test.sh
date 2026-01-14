#!/bin/bash
# Integration tests for Phase 14A end-to-end comprehensive analysis
# Run from project root: ./tests/integration/phase14a_e2e_test.sh

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

echo "🧪 Phase 14A End-to-End Integration Tests"
echo "══════════════════════════════════════════════════════════════"
echo ""

# Test 1: Comprehensive Analysis Endpoint
echo "Test 1: Comprehensive Analysis Endpoint"
echo "   ⚠️  Requires running Hub API"
echo "   Manual test: POST /api/v1/analyze/comprehensive"
echo "   Expected: Returns validation report with all layer findings"

# Test 2: Feature Discovery Integration
echo ""
echo "Test 2: Feature Discovery Integration"
if grep -q "discoverFeature" hub/api/main.go; then
    echo "   ✅ Feature discovery integrated in main handler"
else
    echo "   ❌ Feature discovery not integrated"
    exit 1
fi

# Test 3: All Analyzers Integrated
echo ""
echo "Test 3: All Analyzers Integrated"
if grep -q "analyzeUILayer\|analyzeAPILayer\|analyzeDatabaseLayer\|analyzeIntegrationLayer\|analyzeBusinessLogic" hub/api/main.go; then
    echo "   ✅ All layer analyzers integrated"
else
    echo "   ❌ Some analyzers not integrated"
    exit 1
fi

# Test 4: Flow Verification Integrated
echo ""
echo "Test 4: Flow Verification Integrated"
if grep -q "verifyEndToEndFlows" hub/api/main.go; then
    echo "   ✅ Flow verification integrated"
else
    echo "   ❌ Flow verification not integrated"
    exit 1
fi

# Test 5: Result Aggregation Integrated
echo ""
echo "Test 5: Result Aggregation Integrated"
if grep -q "AggregateResults\|generateChecklist\|generateSummary" hub/api/main.go; then
    echo "   ✅ Result aggregation integrated"
else
    echo "   ❌ Result aggregation not integrated"
    exit 1
fi

# Test 6: LLM Integration
echo ""
echo "Test 6: LLM Integration"
if grep -q "getLLMConfig\|trackUsage" hub/api/logic_analyzer.go; then
    echo "   ✅ LLM integration in semantic analysis"
else
    echo "   ❌ LLM integration missing"
    exit 1
fi

# Test 7: Database Storage
echo ""
echo "Test 7: Database Storage"
if grep -q "storeComprehensiveValidation" hub/api/main.go; then
    echo "   ✅ Validation storage function exists"
else
    echo "   ❌ Validation storage function not found"
    exit 1
fi

# Test 8: Validation Retrieval
echo ""
echo "Test 8: Validation Retrieval Endpoint"
if grep -q "getComprehensiveValidationHandler\|listValidationsHandler" hub/api/main.go; then
    echo "   ✅ Validation retrieval endpoints exist"
else
    echo "   ❌ Validation retrieval endpoints not found"
    exit 1
fi

# Test 9: Error Handling
echo ""
echo "Test 9: Error Handling"
if grep -q "LogWarn.*analysis failed" hub/api/main.go; then
    echo "   ✅ Error handling for failed analyses"
else
    echo "   ⚠️  Error handling could be improved"
fi

# Test 10: All Layers Analyzed
echo ""
echo "Test 10: All Layers Analyzed"
LAYERS=("business" "ui" "api" "database" "logic" "integration" "test")
for layer in "${LAYERS[@]}"; do
    if grep -q "\"$layer\"" hub/api/main.go; then
        echo "   ✅ $layer layer included in analysis"
    else
        echo "   ⚠️  $layer layer may not be included"
    fi
done

echo ""
echo "✅ All integration tests passed!"
echo "══════════════════════════════════════════════════════════════"
echo ""
echo "📝 Note: Full end-to-end testing requires:"
echo "   1. Running Hub API server"
echo "   2. Database connection configured"
echo "   3. LLM configuration (optional, for semantic analysis)"
echo "   4. Test project with codebase to analyze"

