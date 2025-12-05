#!/bin/bash
# Unit tests for Sentinel Hub API endpoints
# Run from project root: ./tests/unit/hub_api_test.sh

# Don't use set -e as some commands intentionally fail

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
echo "   Hub API Unit Tests"
echo "=============================================="
echo ""

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT"

# ============================================================================
# Test: AST Analysis endpoint exists (Phase 6)
# ============================================================================

echo "Testing AST analysis endpoint definition..."
cleanup_lock

if grep -q "func astAnalysisHandler" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "AST analysis handler function exists"
else
    log_fail "AST analysis handler not found"
fi

if grep -q "/analyze/ast" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "AST analysis endpoint route is registered"
else
    log_fail "AST analysis endpoint route not found"
fi

# ============================================================================
# Test: AST Analysis types exist (Phase 6)
# ============================================================================

echo ""
echo "Testing AST analysis types..."
cleanup_lock

if grep -q "type ASTAnalysisRequest struct" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "ASTAnalysisRequest type exists"
else
    log_fail "ASTAnalysisRequest type not found"
fi

if grep -q "type ASTAnalysisResponse struct" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "ASTAnalysisResponse type exists"
else
    log_fail "ASTAnalysisResponse type not found"
fi

if grep -q "type ASTFinding struct" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "ASTFinding type exists"
else
    log_fail "ASTFinding type not found"
fi

# ============================================================================
# Test: Vibe Analysis endpoint exists (Phase 7)
# ============================================================================

echo ""
echo "Testing vibe analysis endpoint..."
cleanup_lock

if grep -q "func vibeAnalysisHandler" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "Vibe analysis handler function exists"
else
    log_fail "Vibe analysis handler not found"
fi

if grep -q "/analyze/vibe" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "Vibe analysis endpoint route is registered"
else
    log_fail "Vibe analysis endpoint route not found"
fi

# ============================================================================
# Test: Cross-file Analysis endpoint exists (Phase 6)
# ============================================================================

echo ""
echo "Testing cross-file analysis endpoint..."
cleanup_lock

if grep -q "func crossFileAnalysisHandler" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "Cross-file analysis handler function exists"
else
    log_fail "Cross-file analysis handler not found"
fi

if grep -q "/analyze/cross-file" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "Cross-file analysis endpoint route is registered"
else
    log_fail "Cross-file analysis endpoint route not found"
fi

# ============================================================================
# Test: Security Analysis endpoint exists (Phase 8)
# ============================================================================

echo ""
echo "Testing security analysis endpoint..."
cleanup_lock

if grep -q "func securityAnalysisHandler" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "Security analysis handler function exists"
else
    log_fail "Security analysis handler not found"
fi

if grep -q "/analyze/security" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "Security analysis endpoint route is registered"
else
    log_fail "Security analysis endpoint route not found"
fi

# ============================================================================
# Test: Security Rules definitions exist (Phase 8)
# ============================================================================

echo ""
echo "Testing security rules definitions..."
cleanup_lock

if grep -q "SEC-001\|SEC-002\|SEC-003\|SEC-004\|SEC-005\|SEC-006\|SEC-007\|SEC-008" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "Security rule IDs (SEC-001 to SEC-008) are defined"
else
    log_fail "Security rule IDs not found"
fi

if grep -q "var securityRules\|securityRules.*=" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "Security rules map/variable exists"
else
    log_fail "Security rules variable not found"
fi

# ============================================================================
# Test: Security Analysis types exist (Phase 8)
# ============================================================================

echo ""
echo "Testing security analysis types..."
cleanup_lock

if grep -q "type SecurityAnalysisRequest struct" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "SecurityAnalysisRequest type exists"
else
    log_fail "SecurityAnalysisRequest type not found"
fi

if grep -q "type SecurityAnalysisResponse struct" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "SecurityAnalysisResponse type exists"
else
    log_fail "SecurityAnalysisResponse type not found"
fi

if grep -q "type SecurityFinding struct" "$PROJECT_ROOT/hub/api/main.go"; then
    log_pass "SecurityFinding type exists"
else
    log_fail "SecurityFinding type not found"
fi

# ============================================================================
# Test: Tree-sitter dependency added (Phase 6)
# ============================================================================

echo ""
echo "Testing Tree-sitter dependency..."
cleanup_lock

if grep -q "tree-sitter\|go-tree-sitter" "$PROJECT_ROOT/hub/api/go.mod"; then
    log_pass "Tree-sitter dependency is in go.mod"
else
    log_fail "Tree-sitter dependency not found in go.mod"
fi

# ============================================================================
# Summary
# ============================================================================

echo ""
echo "=============================================="
echo "   Hub API Test Results"
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

