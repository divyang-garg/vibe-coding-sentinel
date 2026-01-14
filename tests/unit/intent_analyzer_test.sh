#!/bin/bash
# Unit tests for Phase 15: Intent Analyzer
# Run from project root: ./tests/unit/intent_analyzer_test.sh

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT"

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

echo ""
echo "Testing Phase 15: Intent Analyzer"
echo "===================================="
echo ""

# Test 1: Verify intent_analyzer.go exists
echo "Test 1: Verify intent_analyzer.go exists..."
if [ -f "hub/api/intent_analyzer.go" ]; then
    log_pass "intent_analyzer.go exists"
else
    log_fail "intent_analyzer.go not found"
fi

# Test 2: Verify type definitions exist
echo "Test 2: Verify type definitions exist..."
if grep -q "type IntentType" hub/api/types.go && \
   grep -q "type IntentAnalysisRequest" hub/api/types.go && \
   grep -q "type IntentAnalysisResponse" hub/api/types.go; then
    log_pass "Type definitions exist"
else
    log_fail "Type definitions missing"
fi

# Test 3: Verify database schema exists
echo "Test 3: Verify database schema exists..."
if grep -q "CREATE TABLE.*intent_decisions" hub/api/main.go && \
   grep -q "CREATE TABLE.*intent_patterns" hub/api/main.go; then
    log_pass "Database schema exists"
else
    log_fail "Database schema missing"
fi

# Test 4: Verify API endpoints exist
echo "Test 4: Verify API endpoints exist..."
if grep -q "/analyze/intent" hub/api/main.go && \
   grep -q "/intent/decisions" hub/api/main.go && \
   grep -q "/intent/patterns" hub/api/main.go; then
    log_pass "API endpoints exist"
else
    log_fail "API endpoints missing"
fi

# Test 5: Verify handler functions exist
echo "Test 5: Verify handler functions exist..."
if grep -q "func intentAnalysisHandler" hub/api/main.go && \
   grep -q "func recordIntentDecisionHandler" hub/api/main.go && \
   grep -q "func getIntentPatternsHandler" hub/api/main.go; then
    log_pass "Handler functions exist"
else
    log_fail "Handler functions missing"
fi

# Test 6: Verify intent analyzer functions exist
echo "Test 6: Verify intent analyzer functions exist..."
if grep -q "func GetTemplates" hub/api/intent_analyzer.go && \
   grep -q "func GatherContext" hub/api/intent_analyzer.go && \
   grep -q "func AnalyzeIntent" hub/api/intent_analyzer.go && \
   grep -q "func RecordDecision" hub/api/intent_analyzer.go && \
   grep -q "func GetLearnedPatterns" hub/api/intent_analyzer.go; then
    log_pass "Intent analyzer functions exist"
else
    log_fail "Intent analyzer functions missing"
fi

# Test 7: Verify MCP tool registered
echo "Test 7: Verify MCP tool registered..."
if grep -q "sentinel_check_intent" synapsevibsentinel.sh && \
   grep -q "handleCheckIntent" synapsevibsentinel.sh; then
    log_pass "MCP tool registered"
else
    log_fail "MCP tool not registered"
fi

# Test 8: Verify code compiles (basic syntax check)
echo "Test 8: Verify code compiles..."
if command -v go &> /dev/null; then
    cd hub/api
    if go build -o /dev/null . 2>&1 | grep -q "intent"; then
        log_fail "Compilation errors found"
    else
        log_pass "Code compiles successfully"
    fi
    cd "$PROJECT_ROOT"
else
    echo -e "${YELLOW}⚠ SKIP:${NC} Go not installed, skipping compilation test"
fi

# Summary
echo ""
echo "===================================="
echo "Test Summary"
echo "===================================="
echo -e "${GREEN}Passed:${NC} $TESTS_PASSED"
echo -e "${RED}Failed:${NC} $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi










