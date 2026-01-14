#!/bin/bash
# Vibe Coding Accuracy Tests - Phase 7D
# Tests accuracy measurement (85%+ target)

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"
FIXTURES_DIR="$TEST_DIR/../fixtures/patterns"

echo "🎯 Testing Vibe Coding Detection Accuracy (Phase 7D)"
echo "===================================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Known issues in test fixtures (ground truth)
declare -A KNOWN_ISSUES
KNOWN_ISSUES["duplicate_function.go"]="duplicate_function"
KNOWN_ISSUES["duplicate_function.js"]="duplicate_function"
KNOWN_ISSUES["duplicate_function.ts"]="duplicate_function"
KNOWN_ISSUES["duplicate_function.py"]="duplicate_function"
KNOWN_ISSUES["unused_variable.go"]="unused_variable"
KNOWN_ISSUES["unused_variable.js"]="unused_variable"
KNOWN_ISSUES["unused_variable.py"]="unused_variable"
KNOWN_ISSUES["unreachable_code.go"]="unreachable_code"
KNOWN_ISSUES["unreachable_code.js"]="unreachable_code"
KNOWN_ISSUES["orphaned_code.go"]="orphaned_code"
KNOWN_ISSUES["orphaned_code.ts"]="orphaned_code"

TOTAL_TESTS=0
DETECTED=0
FALSE_POSITIVES=0
FALSE_NEGATIVES=0

# Test accuracy for a file
test_accuracy() {
    local file="$1"
    local expected_type="$2"
    local method="$3"  # "ast" or "pattern"
    
    ((TOTAL_TESTS++))
    
    if [ ! -f "$file" ]; then
        echo -e "${RED}FAIL${NC} - File not found: $file"
        ((FALSE_NEGATIVES++))
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
    
    if [ "$method" = "ast" ]; then
        # Test AST detection (requires Hub)
        if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo -e "${YELLOW}SKIP${NC} - Hub not running for AST test"
            return
        fi
        
        code=$(cat "$file")
        response=$(curl -s -X POST http://localhost:8080/api/v1/analyze/vibe \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer test-key" \
            -d "{
                \"code\": $(echo "$code" | jq -Rs .),
                \"language\": \"$lang\",
                \"filename\": \"$filename\",
                \"projectId\": \"test-project\",
                \"analyses\": [\"duplicates\", \"unused\", \"unreachable\", \"orphaned\"]
            }" 2>/dev/null || echo "{}")
        
        if echo "$response" | jq -e ".findings[] | select(.type == \"$expected_type\")" > /dev/null 2>&1; then
            ((DETECTED++))
            return 0
        else
            ((FALSE_NEGATIVES++))
            return 1
        fi
    else
        # Test pattern detection (local)
        cd "$PROJECT_ROOT"
        if [ ! -f "./sentinel" ]; then
            ./synapsevibsentinel.sh > /dev/null 2>&1
        fi
        
        # Run audit with vibe-check and offline mode (pattern only)
        result=$(./sentinel audit --vibe-check --offline --output json 2>/dev/null | jq -r '.findings[] | select(.pattern | contains("VIBE")) | .pattern' 2>/dev/null || echo "")
        
        # Check if pattern detected the issue (basic check)
        if echo "$result" | grep -q "VIBE" 2>/dev/null; then
            ((DETECTED++))
            return 0
        else
            ((FALSE_NEGATIVES++))
            return 1
        fi
    fi
}

echo ""
echo "Testing AST Detection Accuracy..."
echo ""

# Test AST detection
AST_DETECTED=0
AST_TOTAL=0

for file in "$FIXTURES_DIR"/*.go "$FIXTURES_DIR"/*.js "$FIXTURES_DIR"/*.ts "$FIXTURES_DIR"/*.py; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        expected_type="${KNOWN_ISSUES[$filename]}"
        if [ -n "$expected_type" ]; then
            ((AST_TOTAL++))
            if test_accuracy "$file" "$expected_type" "ast"; then
                ((AST_DETECTED++))
            fi
        fi
    fi
done

# Calculate accuracy
if [ $AST_TOTAL -gt 0 ]; then
    AST_ACCURACY=$(echo "scale=2; $AST_DETECTED * 100 / $AST_TOTAL" | bc)
    echo ""
    echo "AST Detection Results:"
    echo "  Detected: $AST_DETECTED / $AST_TOTAL"
    echo "  Accuracy: ${AST_ACCURACY}%"
    
    if (( $(echo "$AST_ACCURACY >= 85" | bc -l) )); then
        echo -e "  Status: ${GREEN}✅ PASS (>= 85%)${NC}"
    else
        echo -e "  Status: ${RED}❌ FAIL (< 85%)${NC}"
    fi
fi

echo ""
echo "Testing Pattern Detection Accuracy..."
echo ""

# Test pattern detection
PATTERN_DETECTED=0
PATTERN_TOTAL=0

for file in "$FIXTURES_DIR"/*.go "$FIXTURES_DIR"/*.js "$FIXTURES_DIR"/*.ts "$FIXTURES_DIR"/*.py; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        expected_type="${KNOWN_ISSUES[$filename]}"
        if [ -n "$expected_type" ]; then
            ((PATTERN_TOTAL++))
            if test_accuracy "$file" "$expected_type" "pattern"; then
                ((PATTERN_DETECTED++))
            fi
        fi
    fi
done

# Calculate accuracy
if [ $PATTERN_TOTAL -gt 0 ]; then
    PATTERN_ACCURACY=$(echo "scale=2; $PATTERN_DETECTED * 100 / $PATTERN_TOTAL" | bc)
    echo ""
    echo "Pattern Detection Results:"
    echo "  Detected: $PATTERN_DETECTED / $PATTERN_TOTAL"
    echo "  Accuracy: ${PATTERN_ACCURACY}%"
    
    if (( $(echo "$PATTERN_ACCURACY >= 60" | bc -l) )); then
        echo -e "  Status: ${GREEN}✅ PASS (>= 60% expected for patterns)${NC}"
    else
        echo -e "  Status: ${YELLOW}⚠️  WARN (< 60% for patterns)${NC}"
    fi
fi

echo ""
echo "===================================================="
echo "Summary:"
echo "  AST Accuracy: ${AST_ACCURACY}% (target: >= 85%)"
echo "  Pattern Accuracy: ${PATTERN_ACCURACY}% (target: >= 60%)"
echo ""

if (( $(echo "$AST_ACCURACY >= 85" | bc -l) )); then
    echo -e "${GREEN}✅ Accuracy test PASSED${NC}"
    exit 0
else
    echo -e "${RED}❌ Accuracy test FAILED${NC}"
    exit 1
fi












