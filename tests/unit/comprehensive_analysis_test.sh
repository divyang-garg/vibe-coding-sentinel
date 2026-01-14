#!/bin/bash
# Unit tests for Phase 14A Comprehensive Feature Analysis
# Run from project root: ./tests/unit/comprehensive_analysis_test.sh

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

echo "🧪 Phase 14A Comprehensive Analysis Unit Tests"
echo "══════════════════════════════════════════════════════════════"
echo ""

# Test 1: Feature Discovery Module
echo "Test 1: Feature Discovery Module"
if [ -f "hub/api/feature_discovery.go" ]; then
    echo "   ✅ Feature discovery module exists"
    if grep -q "discoverFeature\|OrchestrateFeatureDiscovery" hub/api/feature_discovery.go; then
        echo "   ✅ Feature discovery function exists"
    else
        echo "   ❌ Feature discovery function not found"
        exit 1
    fi
else
    echo "   ❌ Feature discovery module not found"
    exit 1
fi

# Test 2: UI Analyzer
echo ""
echo "Test 2: UI Layer Analyzer"
if [ -f "hub/api/ui_analyzer.go" ]; then
    echo "   ✅ UI analyzer module exists"
    if grep -q "analyzeUILayer" hub/api/ui_analyzer.go; then
        echo "   ✅ UI analysis function exists"
    else
        echo "   ❌ UI analysis function not found"
        exit 1
    fi
    if grep -q "checkAccessibility" hub/api/ui_analyzer.go; then
        echo "   ✅ Accessibility checks exist"
    else
        echo "   ❌ Accessibility checks not found"
        exit 1
    fi
else
    echo "   ❌ UI analyzer module not found"
    exit 1
fi

# Test 3: API Analyzer
echo ""
echo "Test 3: API Layer Analyzer"
if [ -f "hub/api/api_analyzer.go" ]; then
    echo "   ✅ API analyzer module exists"
    if grep -q "analyzeAPILayer\|AnalyzeAPILayer" hub/api/api_analyzer.go; then
        echo "   ✅ API analysis function exists"
    else
        echo "   ❌ API analysis function not found"
        exit 1
    fi
else
    echo "   ❌ API analyzer module not found"
    exit 1
fi

# Test 4: Database Analyzer
echo ""
echo "Test 4: Database Layer Analyzer"
if [ -f "hub/api/database_analyzer.go" ]; then
    echo "   ✅ Database analyzer module exists"
    if grep -q "analyzeDatabaseLayer" hub/api/database_analyzer.go; then
        echo "   ✅ Database analysis function exists"
    else
        echo "   ❌ Database analysis function not found"
        exit 1
    fi
else
    echo "   ❌ Database analyzer module not found"
    exit 1
fi

# Test 5: Integration Analyzer
echo ""
echo "Test 5: Integration Layer Analyzer"
if [ -f "hub/api/integration_analyzer.go" ]; then
    echo "   ✅ Integration analyzer module exists"
    if grep -q "analyzeIntegrationLayer" hub/api/integration_analyzer.go; then
        echo "   ✅ Integration analysis function exists"
    else
        echo "   ❌ Integration analysis function not found"
        exit 1
    fi
else
    echo "   ❌ Integration analyzer module not found"
    exit 1
fi

# Test 6: Logic Analyzer with LLM
echo ""
echo "Test 6: Business Logic Analyzer with LLM"
if [ -f "hub/api/logic_analyzer.go" ]; then
    echo "   ✅ Logic analyzer module exists"
    if grep -q "semanticAnalysis" hub/api/logic_analyzer.go; then
        echo "   ✅ Semantic analysis function exists"
    else
        echo "   ❌ Semantic analysis function not found"
        exit 1
    fi
    if grep -q "getLLMConfig" hub/api/logic_analyzer.go; then
        echo "   ✅ LLM integration exists"
    else
        echo "   ❌ LLM integration not found"
        exit 1
    fi
else
    echo "   ❌ Logic analyzer module not found"
    exit 1
fi

# Test 7: Flow Verifier
echo ""
echo "Test 7: End-to-End Flow Verifier"
if [ -f "hub/api/flow_verifier.go" ]; then
    echo "   ✅ Flow verifier module exists"
    if grep -q "verifyEndToEndFlows" hub/api/flow_verifier.go; then
        echo "   ✅ Flow verification function exists"
    else
        echo "   ❌ Flow verification function not found"
        exit 1
    fi
    if grep -q "identifyBreakpoints" hub/api/flow_verifier.go; then
        echo "   ✅ Breakpoint identification exists"
    else
        echo "   ❌ Breakpoint identification not found"
        exit 1
    fi
else
    echo "   ❌ Flow verifier module not found"
    exit 1
fi

# Test 8: Result Aggregator
echo ""
echo "Test 8: Result Aggregator"
if [ -f "hub/api/result_aggregator.go" ]; then
    echo "   ✅ Result aggregator module exists"
    if grep -q "generateChecklist\|generateSummary\|formatReport" hub/api/result_aggregator.go; then
        echo "   ✅ Aggregation functions exist"
    else
        echo "   ❌ Aggregation functions not found"
        exit 1
    fi
else
    echo "   ❌ Result aggregator module not found"
    exit 1
fi

# Test 9: LLM Integration
echo ""
echo "Test 9: LLM Integration"
if [ -f "hub/api/llm_integration.go" ]; then
    echo "   ✅ LLM integration module exists"
    if grep -q "getLLMConfig" hub/api/llm_integration.go; then
        echo "   ✅ LLM config retrieval exists"
    else
        echo "   ❌ LLM config retrieval not found"
        exit 1
    fi
    if grep -q "trackUsage" hub/api/llm_integration.go; then
        echo "   ✅ Usage tracking exists"
    else
        echo "   ❌ Usage tracking not found"
        exit 1
    fi
else
    echo "   ❌ LLM integration module not found"
    exit 1
fi

# Test 10: LLM Cache
echo ""
echo "Test 10: LLM Cache"
if [ -f "hub/api/llm_cache.go" ]; then
    echo "   ✅ LLM cache module exists"
    if grep -q "analyzeWithProgressiveDepth" hub/api/llm_cache.go; then
        echo "   ✅ Progressive depth analysis exists"
    else
        echo "   ❌ Progressive depth analysis not found"
        exit 1
    fi
    if grep -q "getCachedLLMResponse" hub/api/llm_cache.go; then
        echo "   ✅ Cache retrieval exists"
    else
        echo "   ❌ Cache retrieval not found"
        exit 1
    fi
else
    echo "   ❌ LLM cache module not found"
    exit 1
fi

# Test 11: API Endpoints
echo ""
echo "Test 11: API Endpoints"
if grep -q "comprehensiveAnalysisHandler" hub/api/main.go; then
    echo "   ✅ Comprehensive analysis endpoint exists"
else
    echo "   ❌ Comprehensive analysis endpoint not found"
    exit 1
fi
if grep -q "getComprehensiveValidationHandler\|listValidationsHandler" hub/api/main.go; then
    echo "   ✅ Validation retrieval endpoints exist"
else
    echo "   ❌ Validation retrieval endpoints not found"
    exit 1
fi

# Test 12: Database Schema
echo ""
echo "Test 12: Database Schema"
if grep -q "comprehensive_validations" hub/api/main.go; then
    echo "   ✅ Comprehensive validations table exists"
else
    echo "   ❌ Comprehensive validations table not found"
    exit 1
fi
if grep -q "llm_usage" hub/api/main.go; then
    echo "   ✅ LLM usage table exists"
else
    echo "   ❌ LLM usage table not found"
    exit 1
fi

echo ""
echo "✅ All unit tests passed!"
echo "══════════════════════════════════════════════════════════════"

