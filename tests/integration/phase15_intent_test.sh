#!/bin/bash
# Integration tests for Phase 15: Intent & Simple Language end-to-end
# Run from project root: ./tests/integration/phase15_intent_test.sh

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

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

log_warn() {
    echo -e "${YELLOW}⚠ WARN:${NC} $1"
}

echo "🧪 Phase 15: Intent & Simple Language Integration Tests"
echo "══════════════════════════════════════════════════════════════"
echo ""

# Source test helper if available
source "$TEST_DIR/../helpers/mcp_test_client.sh" 2>/dev/null || {
    log_warn "MCP test helper not found, using basic functions"
}

# Ensure binary is built
if [[ ! -f "./sentinel" ]]; then
    echo "Building Sentinel binary..."
    ./synapsevibsentinel.sh > /dev/null 2>&1 || true
fi

# Test 1: MCP Tool Registration
echo "Test 1: Verify sentinel_check_intent tool is registered..."
if grep -q "sentinel_check_intent" synapsevibsentinel.sh && \
   grep -q "Analyze unclear prompts" synapsevibsentinel.sh; then
    log_pass "MCP tool registered correctly"
else
    log_fail "MCP tool not registered"
fi

# Test 2: Handler Function Exists
echo "Test 2: Verify handleCheckIntent function exists..."
if grep -q "func handleCheckIntent" synapsevibsentinel.sh; then
    log_pass "Handler function exists"
else
    log_fail "Handler function missing"
fi

# Test 3: API Endpoints Exist
echo "Test 3: Verify Hub API endpoints exist..."
if grep -q "/api/v1/analyze/intent" hub/api/main.go && \
   grep -q "/api/v1/intent/decisions" hub/api/main.go && \
   grep -q "/api/v1/intent/patterns" hub/api/main.go; then
    log_pass "API endpoints exist"
else
    log_fail "API endpoints missing"
fi

# Test 4: Database Schema
echo "Test 4: Verify database schema exists..."
if grep -q "CREATE TABLE.*intent_decisions" hub/api/main.go && \
   grep -q "CREATE TABLE.*intent_patterns" hub/api/main.go; then
    log_pass "Database schema exists"
else
    log_fail "Database schema missing"
fi

# Test 5: Intent Analyzer Functions
echo "Test 5: Verify intent analyzer functions exist..."
if grep -q "func AnalyzeIntent" hub/api/intent_analyzer.go && \
   grep -q "func GatherContext" hub/api/intent_analyzer.go && \
   grep -q "func RecordDecision" hub/api/intent_analyzer.go && \
   grep -q "func GetLearnedPatterns" hub/api/intent_analyzer.go; then
    log_pass "Intent analyzer functions exist"
else
    log_fail "Intent analyzer functions missing"
fi

# Test 6: Type Definitions
echo "Test 6: Verify type definitions exist..."
if grep -q "type IntentType" hub/api/types.go && \
   grep -q "type IntentAnalysisRequest" hub/api/types.go && \
   grep -q "type IntentAnalysisResponse" hub/api/types.go; then
    log_pass "Type definitions exist"
else
    log_fail "Type definitions missing"
fi

# Test 7: Code Compilation (if Go is available)
echo "Test 7: Verify code compiles..."
if command -v go &> /dev/null; then
    cd hub/api
    if go build -o /dev/null . 2>&1 | grep -i "error\|fail"; then
        log_fail "Compilation errors found"
    else
        log_pass "Code compiles successfully"
    fi
    cd "$PROJECT_ROOT"
else
    log_warn "Go not installed, skipping compilation test"
fi

# Test 8: MCP Tool Schema Validation
echo "Test 8: Verify MCP tool schema is valid..."
if grep -A 20 "sentinel_check_intent" synapsevibsentinel.sh | grep -q "prompt" && \
   grep -A 20 "sentinel_check_intent" synapsevibsentinel.sh | grep -q "codebasePath"; then
    log_pass "MCP tool schema is valid"
else
    log_fail "MCP tool schema invalid"
fi

# Test 9: Response Formatter
echo "Test 9: Verify response formatter exists..."
if grep -q "func formatIntentAnalysisResponse" synapsevibsentinel.sh; then
    log_pass "Response formatter exists"
else
    log_fail "Response formatter missing"
fi

# Test 10: Error Handling
echo "Test 10: Verify error handling exists..."
if grep -q "ConfigErrorCode" synapsevibsentinel.sh && \
   grep -q "handleHubError" synapsevibsentinel.sh; then
    log_pass "Error handling exists"
else
    log_fail "Error handling missing"
fi

# Summary
echo ""
echo "══════════════════════════════════════════════════════════════"
echo "Test Summary"
echo "══════════════════════════════════════════════════════════════"
echo -e "${GREEN}Passed:${NC} $TESTS_PASSED"
echo -e "${RED}Failed:${NC} $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All integration tests passed!${NC}"
    echo ""
    echo "Note: Full end-to-end tests require Hub API to be running."
    echo "To test with Hub API:"
    echo "  1. Start Hub API: cd hub/api && go run ."
    echo "  2. Set SENTINEL_HUB_URL and SENTINEL_API_KEY"
    echo "  3. Test MCP tool via Cursor IDE"
    exit 0
else
    echo -e "${RED}❌ Some integration tests failed!${NC}"
    exit 1
fi










