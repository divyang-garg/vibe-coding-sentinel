#!/bin/bash
# Unit tests for Sentinel pattern learning functionality
# Run from project root: ./tests/unit/pattern_learning_test.sh

# Don't use set -e as grep returns non-zero when no match

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
echo "   Pattern Learning Unit Tests"
echo "=============================================="
echo ""

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SENTINEL="$PROJECT_ROOT/sentinel"

# ============================================================================
# Test: Learn command on JavaScript project
# ============================================================================

echo "Testing learn command on JavaScript project..."
cleanup_lock

cd "$PROJECT_ROOT/tests/fixtures/projects/javascript"
OUTPUT=$("$SENTINEL" learn 2>&1 || true)

# Check that it detects JavaScript
if echo "$OUTPUT" | grep -q "Primary language: JavaScript"; then
    log_pass "Detects JavaScript as primary language"
else
    log_fail "Failed to detect JavaScript"
fi

# Check that it detects React
if echo "$OUTPUT" | grep -q "Framework: React"; then
    log_pass "Detects React framework"
else
    log_fail "Failed to detect React framework"
fi

# Check that it detects camelCase
if echo "$OUTPUT" | grep -q "Functions:.*camelCase"; then
    log_pass "Detects camelCase for functions"
else
    log_fail "Failed to detect camelCase functions"
fi

# Check that patterns.json was created
if [[ -f ".sentinel/patterns.json" ]]; then
    log_pass "Creates patterns.json file"
else
    log_fail "Failed to create patterns.json"
fi

# Check that Cursor rules were created
if [[ -f ".cursor/rules/project-patterns.md" ]]; then
    log_pass "Creates project-patterns.md rule"
else
    log_fail "Failed to create project-patterns.md"
fi

# Cleanup
rm -rf .sentinel .cursor

# ============================================================================
# Test: Learn command on Python project
# ============================================================================

echo ""
echo "Testing learn command on Python project..."
cleanup_lock

cd "$PROJECT_ROOT/tests/fixtures/projects/python"
OUTPUT=$("$SENTINEL" learn 2>&1 || true)

# Check that it detects Python
if echo "$OUTPUT" | grep -q "Primary language: Python"; then
    log_pass "Detects Python as primary language"
else
    log_fail "Failed to detect Python"
fi

# Check that it detects FastAPI
if echo "$OUTPUT" | grep -q "Framework: FastAPI"; then
    log_pass "Detects FastAPI framework"
else
    log_fail "Failed to detect FastAPI framework"
fi

# Check that it detects snake_case
if echo "$OUTPUT" | grep -q "Functions:.*snake_case"; then
    log_pass "Detects snake_case for Python functions"
else
    log_fail "Failed to detect snake_case functions"
fi

# Check folder structure detection
if echo "$OUTPUT" | grep -q "Source root: src/"; then
    log_pass "Detects src/ as source root"
else
    log_fail "Failed to detect source root"
fi

# Cleanup
rm -rf .sentinel .cursor

# ============================================================================
# Test: Learn command on Shell project
# ============================================================================

echo ""
echo "Testing learn command on Shell project..."
cleanup_lock

cd "$PROJECT_ROOT/tests/fixtures/projects/shell"
OUTPUT=$("$SENTINEL" learn 2>&1 || true)

# Check that it detects Shell
if echo "$OUTPUT" | grep -q "Primary language: Shell"; then
    log_pass "Detects Shell as primary language"
else
    log_fail "Failed to detect Shell"
fi

# Check that it detects naming patterns (shell may have limited samples)
if echo "$OUTPUT" | grep -q "Functions:.*\(snake_case\|unknown\)"; then
    log_pass "Processes Shell naming patterns"
else
    log_fail "Failed to process Shell naming patterns"
fi

# Cleanup
rm -rf .sentinel .cursor

# ============================================================================
# Test: Learn command with --naming flag
# ============================================================================

echo ""
echo "Testing learn command with --naming flag..."
cleanup_lock

cd "$PROJECT_ROOT/tests/fixtures/projects/javascript"
OUTPUT=$("$SENTINEL" learn --naming 2>&1 || true)

# Check that naming is learned
if echo "$OUTPUT" | grep -q "Learning naming conventions"; then
    log_pass "Learns naming conventions with flag"
else
    log_fail "Failed with --naming flag"
fi

# Check that imports are NOT learned (since we only specified naming)
if echo "$OUTPUT" | grep -q "Learning import patterns"; then
    log_fail "Should not learn imports with --naming flag only"
else
    log_pass "Correctly skips imports with --naming flag"
fi

# Cleanup
rm -rf .sentinel .cursor

# ============================================================================
# Test: Learn command generates valid JSON
# ============================================================================

echo ""
echo "Testing learn generates valid JSON..."
cleanup_lock

cd "$PROJECT_ROOT/tests/fixtures/projects/javascript"
"$SENTINEL" learn >/dev/null 2>&1

if [[ -f ".sentinel/patterns.json" ]]; then
    # Try to parse with Python (if available) or basic validation
    if command -v python3 &> /dev/null; then
        if python3 -c "import json; json.load(open('.sentinel/patterns.json'))" 2>/dev/null; then
            log_pass "Generates valid JSON"
        else
            log_fail "Generated JSON is invalid"
        fi
    else
        # Basic validation - check it starts with { and ends with }
        if head -1 .sentinel/patterns.json | grep -q "^{" && tail -1 .sentinel/patterns.json | grep -q "}$"; then
            log_pass "Generates valid JSON (basic check)"
        else
            log_fail "Generated JSON is invalid (basic check)"
        fi
    fi
else
    log_fail "patterns.json not created"
fi

# Cleanup
rm -rf .sentinel .cursor

# ============================================================================
# Test: Learn command generates valid Cursor rule
# ============================================================================

echo ""
echo "Testing learn generates valid Cursor rule..."
cleanup_lock

cd "$PROJECT_ROOT/tests/fixtures/projects/javascript"
"$SENTINEL" learn >/dev/null 2>&1

if [[ -f ".cursor/rules/project-patterns.md" ]]; then
    # Check YAML frontmatter
    if head -5 .cursor/rules/project-patterns.md | grep -q "description:"; then
        log_pass "Cursor rule has valid frontmatter"
    else
        log_fail "Cursor rule missing frontmatter"
    fi
    
    # Check content sections
    if grep -q "## Naming Conventions" .cursor/rules/project-patterns.md; then
        log_pass "Cursor rule has naming conventions section"
    else
        log_fail "Missing naming conventions in Cursor rule"
    fi
else
    log_fail "project-patterns.md not created"
fi

# Cleanup
rm -rf .sentinel .cursor

# Return to project root
cd "$PROJECT_ROOT"

# ============================================================================
# Summary
# ============================================================================

echo ""
echo "=============================================="
echo "   Pattern Learning Test Results"
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

