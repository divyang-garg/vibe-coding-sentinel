#!/bin/bash
# AST Analysis Tests - Phase 6D
# Tests for Hub AST analysis functionality

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"
FIXTURES_DIR="$TEST_DIR/../fixtures/patterns"

echo "🧪 Testing AST Analysis (Phase 6D)"
echo "=================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

# Test function
test_case() {
    local name="$1"
    local file="$2"
    local expected_type="$3"
    
    echo -n "Testing: $name... "
    
    # Check if Hub is running (optional - skip if not available)
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${YELLOW}SKIP (Hub not running)${NC}"
        return
    fi
    
    # Read file content
    if [ ! -f "$file" ]; then
        echo -e "${RED}FAIL${NC} - File not found: $file"
        ((FAILED++))
        return
    fi
    
    code=$(cat "$file")
    lang=$(echo "$file" | sed 's/.*\.\(go\|js\|ts\|py\)$/\1/')
    
    # Map file extension to language name
    case "$lang" in
        go) lang="go" ;;
        js) lang="javascript" ;;
        ts) lang="typescript" ;;
        py) lang="python" ;;
        *) lang="unknown" ;;
    esac
    
    # Send to Hub AST endpoint
    response=$(curl -s -X POST http://localhost:8080/api/v1/analyze/vibe \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-key" \
        -d "{
            \"code\": $(echo "$code" | jq -Rs .),
            \"language\": \"$lang\",
            \"filename\": \"$(basename "$file")\",
            \"projectId\": \"test-project\",
            \"analyses\": [\"duplicates\", \"unused\", \"unreachable\", \"orphaned\"]
        }" 2>/dev/null || echo "{}")
    
    # Check if response contains expected finding type
    if echo "$response" | jq -e ".findings[] | select(.type == \"$expected_type\")" > /dev/null 2>&1; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC} - Expected finding type '$expected_type' not found"
        echo "Response: $response" | head -5
        ((FAILED++))
    fi
}

# Run tests
echo ""
echo "Running AST analysis tests..."
echo ""

# Go tests
test_case "Duplicate function (Go)" "$FIXTURES_DIR/duplicate_function.go" "duplicate_function"
test_case "Unused variable (Go)" "$FIXTURES_DIR/unused_variable.go" "unused_variable"
test_case "Unreachable code (Go)" "$FIXTURES_DIR/unreachable_code.go" "unreachable_code"
test_case "Orphaned code (Go)" "$FIXTURES_DIR/orphaned_code.go" "orphaned_code"

# JavaScript tests
test_case "Duplicate function (JS)" "$FIXTURES_DIR/duplicate_function.js" "duplicate_function"
test_case "Unused variable (JS)" "$FIXTURES_DIR/unused_variable.js" "unused_variable"
test_case "Unreachable code (JS)" "$FIXTURES_DIR/unreachable_code.js" "unreachable_code"

# TypeScript tests
test_case "Duplicate function (TS)" "$FIXTURES_DIR/duplicate_function.ts" "duplicate_function"
test_case "Orphaned code (TS)" "$FIXTURES_DIR/orphaned_code.ts" "orphaned_code"

# Python tests
test_case "Duplicate function (Python)" "$FIXTURES_DIR/duplicate_function.py" "duplicate_function"
test_case "Unused variable (Python)" "$FIXTURES_DIR/unused_variable.py" "unused_variable"

# Additional edge case tests
echo ""
echo "Testing edge cases..."

# Test empty file (should not crash)
if [ -f "$FIXTURES_DIR/empty.go" ] 2>/dev/null; then
    test_case "Empty file handling" "$FIXTURES_DIR/empty.go" "duplicate_function"
fi

# Test syntax error handling (if we have a fixture)
if [ -f "$FIXTURES_DIR/syntax_error.js" ] 2>/dev/null; then
    test_case "Syntax error handling" "$FIXTURES_DIR/syntax_error.js" "duplicate_function"
fi

# Summary
echo ""
echo "=================================="
echo "Results: $PASSED passed, $FAILED failed"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All AST analysis tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi

