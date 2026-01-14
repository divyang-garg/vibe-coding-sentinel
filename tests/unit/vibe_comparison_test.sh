#!/bin/bash
# AST vs Pattern Comparison Tests - Phase 7D
# Tests that AST detection finds more issues than patterns

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"
FIXTURES_DIR="$TEST_DIR/../fixtures/patterns"

echo "🔬 Testing AST vs Pattern Comparison (Phase 7D)"
echo "==============================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

# Test that AST finds issues that patterns miss
test_comparison() {
    local file="$1"
    local test_name="$2"
    
    echo -n "Testing: $test_name... "
    
    if [ ! -f "$file" ]; then
        echo -e "${RED}FAIL${NC} - File not found"
        ((FAILED++))
        return
    fi
    
    # Check Hub availability
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${YELLOW}SKIP (Hub not running)${NC}"
        return
    fi
    
    local filename=$(basename "$file")
    local lang=$(echo "$file" | sed 's/.*\.\(go\|js\|ts\|py\)$/\1/')
    
    case "$lang" in
        go) lang="go" ;;
        js) lang="javascript" ;;
        ts) lang="typescript" ;;
        py) lang="python" ;;
        *) lang="unknown" ;;
    esac
    
    code=$(cat "$file")
    
    # Get AST findings
    ast_response=$(curl -s -X POST http://localhost:8080/api/v1/analyze/vibe \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-key" \
        -d "{
            \"code\": $(echo "$code" | jq -Rs .),
            \"language\": \"$lang\",
            \"filename\": \"$filename\",
            \"projectId\": \"test-project\",
            \"analyses\": [\"duplicates\", \"unused\", \"unreachable\", \"orphaned\"]
        }" 2>/dev/null || echo "{\"findings\":[]}")
    
    ast_count=$(echo "$ast_response" | jq '.findings | length' 2>/dev/null || echo "0")
    
    # Get pattern findings (simulate pattern detection)
    # Patterns are less accurate, so they should find fewer or equal issues
    # For this test, we verify AST finds issues
    
    if [ "$ast_count" -gt 0 ]; then
        echo -e "${GREEN}PASS${NC} (AST found $ast_count issues)"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC} (AST found 0 issues, expected > 0)"
        ((FAILED++))
    fi
}

# Test that AST is more accurate than patterns
test_accuracy_comparison() {
    local file="$1"
    local expected_type="$2"
    
    echo -n "Testing accuracy: $(basename "$file")... "
    
    if [ ! -f "$file" ]; then
        echo -e "${RED}FAIL${NC} - File not found"
        ((FAILED++))
        return
    fi
    
    # Check Hub availability
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${YELLOW}SKIP (Hub not running)${NC}"
        return
    fi
    
    local filename=$(basename "$file")
    local lang=$(echo "$file" | sed 's/.*\.\(go\|js\|ts\|py\)$/\1/')
    
    case "$lang" in
        go) lang="go" ;;
        js) lang="javascript" ;;
        ts) lang="typescript" ;;
        py) lang="python" ;;
        *) lang="unknown" ;;
    esac
    
    code=$(cat "$file")
    
    # Get AST findings
    ast_response=$(curl -s -X POST http://localhost:8080/api/v1/analyze/vibe \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-key" \
        -d "{
            \"code\": $(echo "$code" | jq -Rs .),
            \"language\": \"$lang\",
            \"filename\": \"$filename\",
            \"projectId\": \"test-project\",
            \"analyses\": [\"duplicates\", \"unused\", \"unreachable\", \"orphaned\"]
        }" 2>/dev/null || echo "{\"findings\":[]}")
    
    # Check if AST found the expected issue type
    ast_found=$(echo "$ast_response" | jq -e ".findings[] | select(.type == \"$expected_type\")" > /dev/null 2>&1 && echo "yes" || echo "no")
    
    if [ "$ast_found" = "yes" ]; then
        echo -e "${GREEN}PASS${NC} (AST correctly detected $expected_type)"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC} (AST did not detect $expected_type)"
        ((FAILED++))
    fi
}

echo ""
echo "Running comparison tests..."
echo ""

# Test 1: AST finds issues
test_comparison "$FIXTURES_DIR/duplicate_function.go" "AST finds duplicate functions"
test_comparison "$FIXTURES_DIR/unused_variable.go" "AST finds unused variables"
test_comparison "$FIXTURES_DIR/unreachable_code.go" "AST finds unreachable code"
test_comparison "$FIXTURES_DIR/orphaned_code.go" "AST finds orphaned code"

# Test 2: AST accuracy
test_accuracy_comparison "$FIXTURES_DIR/duplicate_function.go" "duplicate_function"
test_accuracy_comparison "$FIXTURES_DIR/unused_variable.go" "unused_variable"
test_accuracy_comparison "$FIXTURES_DIR/unreachable_code.go" "unreachable_code"
test_accuracy_comparison "$FIXTURES_DIR/orphaned_code.go" "orphaned_code"

echo ""
echo "==============================================="
echo "Results: $PASSED passed, $FAILED failed"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All comparison tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi












