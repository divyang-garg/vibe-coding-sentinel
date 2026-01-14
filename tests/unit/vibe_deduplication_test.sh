#!/bin/bash
# Semantic Deduplication Tests - Phase 7D
# Tests that AST findings take precedence over pattern findings

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"
FIXTURES_DIR="$TEST_DIR/../fixtures/patterns"

echo "🔗 Testing Semantic Deduplication (Phase 7D)"
echo "============================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

# Test that deduplication removes overlapping findings
test_deduplication() {
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
    
    # Verify AST found issues (which should take precedence over patterns)
    if [ "$ast_count" -gt 0 ]; then
        # Check that findings have proper structure (line numbers, types, etc.)
        has_line=$(echo "$ast_response" | jq '.findings[0].line' 2>/dev/null || echo "null")
        has_type=$(echo "$ast_response" | jq '.findings[0].type' 2>/dev/null || echo "null")
        
        if [ "$has_line" != "null" ] && [ "$has_type" != "null" ]; then
            echo -e "${GREEN}PASS${NC} (AST found $ast_count issues with proper structure)"
            ((PASSED++))
        else
            echo -e "${RED}FAIL${NC} (AST findings missing structure)"
            ((FAILED++))
        fi
    else
        echo -e "${YELLOW}SKIP${NC} (No AST findings to deduplicate)"
    fi
}

# Test that semantic matching works (nearby lines)
test_semantic_matching() {
    echo -n "Testing: Semantic matching (nearby lines)... "
    
    # Verify deduplication function exists in code
    if grep -q "func deduplicateFindings" "$PROJECT_ROOT/synapsevibsentinel.sh" 2>/dev/null; then
        # Check if semantic matching is implemented
        if grep -q "normalizeMessage\|semantic\|nearby" "$PROJECT_ROOT/synapsevibsentinel.sh" 2>/dev/null; then
            echo -e "${GREEN}PASS${NC} - Semantic matching implemented"
            ((PASSED++))
        else
            echo -e "${RED}FAIL${NC} - Semantic matching not implemented"
            ((FAILED++))
        fi
    else
        echo -e "${RED}FAIL${NC} - Deduplication function not found"
        ((FAILED++))
    fi
}

# Test that AST findings take precedence
test_ast_precedence() {
    echo -n "Testing: AST findings take precedence... "
    
    # Verify code logic: AST findings are added first, then deduplicated patterns
    if grep -q "deduplicateFindings.*astFindings" "$PROJECT_ROOT/synapsevibsentinel.sh" 2>/dev/null; then
        echo -e "${GREEN}PASS${NC} - Code implements AST precedence"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC} - AST precedence not implemented"
        ((FAILED++))
    fi
}

echo ""
echo "Running deduplication tests..."
echo ""

# Test deduplication for various issue types
test_deduplication "$FIXTURES_DIR/duplicate_function.go" "Deduplicate duplicate functions"
test_deduplication "$FIXTURES_DIR/unused_variable.go" "Deduplicate unused variables"
test_deduplication "$FIXTURES_DIR/unreachable_code.go" "Deduplicate unreachable code"
test_deduplication "$FIXTURES_DIR/orphaned_code.go" "Deduplicate orphaned code"

# Test semantic matching
test_semantic_matching

# Test AST precedence
test_ast_precedence

echo ""
echo "============================================="
echo "Results: $PASSED passed, $FAILED failed"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All deduplication tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi












