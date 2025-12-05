#!/bin/bash
# Unit tests for Sentinel auto-fix functionality
# Run from project root: ./tests/unit/fix_test.sh

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
echo "   Auto-Fix Unit Tests"
echo "=============================================="
echo ""

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SENTINEL="$PROJECT_ROOT/sentinel"
TEST_DIR=$(mktemp -d)

# ============================================================================
# Setup: Create test files
# ============================================================================

echo "Setting up test files..."

# Create test JS file with console.log
cat > "$TEST_DIR/test.js" << 'EOF'
// Test file
function hello() {
    console.log("debug message");
    return "hello";
}
console.log("another debug");
debugger;
EOF

# Create test file with trailing whitespace
cat > "$TEST_DIR/whitespace.js" << 'EOF'
function test() {   
    return true;    
}   
EOF

# Create test file without EOF newline (use printf to avoid adding newline)
printf 'function noNewline() { return true; }' > "$TEST_DIR/no-eof.js"

# ============================================================================
# Test: Dry-run mode doesn't modify files
# ============================================================================

echo ""
echo "Testing dry-run mode..."
cleanup_lock

cd "$TEST_DIR"
ORIGINAL_CONTENT=$(cat test.js)
"$SENTINEL" fix --safe --dry-run > /dev/null 2>&1
NEW_CONTENT=$(cat test.js)

if [[ "$ORIGINAL_CONTENT" == "$NEW_CONTENT" ]]; then
    log_pass "Dry-run mode doesn't modify files"
else
    log_fail "Dry-run mode modified files"
fi

# ============================================================================
# Test: Fix detects console.log
# ============================================================================

echo ""
echo "Testing console.log detection..."
cleanup_lock

OUTPUT=$("$SENTINEL" fix --safe --dry-run 2>&1)
if echo "$OUTPUT" | grep -q "Remove console.log"; then
    log_pass "Detects console.log statements"
else
    log_fail "Failed to detect console.log"
fi

# ============================================================================
# Test: Fix detects debugger
# ============================================================================

echo ""
echo "Testing debugger detection..."
cleanup_lock

if echo "$OUTPUT" | grep -q "Remove debugger"; then
    log_pass "Detects debugger statements"
else
    log_fail "Failed to detect debugger"
fi

# ============================================================================
# Test: Fix detects trailing whitespace
# ============================================================================

echo ""
echo "Testing trailing whitespace detection..."
cleanup_lock

if echo "$OUTPUT" | grep -q "trailing whitespace"; then
    log_pass "Detects trailing whitespace"
else
    log_fail "Failed to detect trailing whitespace"
fi

# ============================================================================
# Test: Safe fix actually removes console.log
# ============================================================================

echo ""
echo "Testing actual console.log removal..."
cleanup_lock

# Create fresh test file
cat > "$TEST_DIR/consolelog.js" << 'EOF'
function test() {
    console.log("debug");
    return true;
}
EOF

"$SENTINEL" fix --safe --yes > /dev/null 2>&1

if grep -q "console.log" "$TEST_DIR/consolelog.js" 2>/dev/null; then
    log_fail "console.log was not removed"
else
    log_pass "console.log was removed"
fi

# ============================================================================
# Test: Backup is created
# ============================================================================

echo ""
echo "Testing backup creation..."
cleanup_lock

if [[ -d ".sentinel/backups" ]] && [[ $(ls -A .sentinel/backups 2>/dev/null) ]]; then
    log_pass "Backup directory created"
else
    log_fail "Backup directory not created"
fi

# ============================================================================
# Test: Fix history is saved
# ============================================================================

echo ""
echo "Testing fix history..."
cleanup_lock

if [[ -f ".sentinel/fix-history.json" ]]; then
    log_pass "Fix history file created"
else
    log_fail "Fix history file not created"
fi

# ============================================================================
# Test: Help shows fix command
# ============================================================================

echo ""
echo "Testing help output..."
cleanup_lock

cd "$PROJECT_ROOT"
HELP_OUTPUT=$("$SENTINEL" --help 2>&1)
if echo "$HELP_OUTPUT" | grep -q "fix.*--safe"; then
    log_pass "Help shows fix command"
else
    log_fail "Help doesn't show fix command"
fi

# ============================================================================
# Test: Import sorting fix definition exists
# ============================================================================

echo ""
echo "Testing import sorting fix exists..."
cleanup_lock

if grep -q "sort-imports" "$PROJECT_ROOT/synapsevibsentinel.sh" 2>/dev/null; then
    log_pass "Import sorting fix is defined"
else
    log_fail "Import sorting fix not found"
fi

# ============================================================================
# Test: Unused imports fix definition exists
# ============================================================================

echo ""
echo "Testing unused imports fix exists..."
cleanup_lock

if grep -q "remove-unused-imports" "$PROJECT_ROOT/synapsevibsentinel.sh" 2>/dev/null; then
    log_pass "Unused imports fix is defined"
else
    log_fail "Unused imports fix not found"
fi

# ============================================================================
# Test: Import sorting helper functions exist
# ============================================================================

echo ""
echo "Testing import sorting helpers exist..."
cleanup_lock

if grep -q "func sortImportsInFile" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "func findImportBlock" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "func categorizeImports" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    log_pass "Import sorting helper functions exist"
else
    log_fail "Import sorting helpers missing"
fi

# ============================================================================
# Test: Unused imports helper functions exist
# ============================================================================

echo ""
echo "Testing unused imports helpers exist..."
cleanup_lock

if grep -q "func findUnusedImports" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "func extractImportedNames" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "func isNameUsedInFile" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "func removeUnusedImports" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    log_pass "Unused imports helper functions exist"
else
    log_fail "Unused imports helpers missing"
fi

# ============================================================================
# Test: Fix handles import sorting in application logic
# ============================================================================

echo ""
echo "Testing import sorting integration..."
cleanup_lock

if grep -q 'f.fix.ID == "sort-imports"' "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    log_pass "Import sorting integrated in fix application"
else
    log_fail "Import sorting not integrated"
fi

# ============================================================================
# Test: Fix handles unused imports in application logic
# ============================================================================

echo ""
echo "Testing unused imports integration..."
cleanup_lock

if grep -q 'f.fix.ID == "remove-unused-imports"' "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    log_pass "Unused imports integrated in fix application"
else
    log_fail "Unused imports not integrated"
fi

# ============================================================================
# Cleanup
# ============================================================================

rm -rf "$TEST_DIR"

# ============================================================================
# Summary
# ============================================================================

echo ""
echo "=============================================="
echo "   Auto-Fix Test Results"
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

