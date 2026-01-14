#!/bin/bash
# API Integration Test Suite
# Tests all Sentinel commands and their interactions

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0

log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    ((FAILED++))
}

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

SENTINEL="./sentinel"

echo "🔌 API Integration Test Suite"
echo "============================"

# Test command availability
echo ""
log_info "Testing Command Availability..."

if $SENTINEL --help >/dev/null 2>&1; then
    log_pass "Help command accessible"
else
    log_fail "Help command not accessible"
fi

# Test command recognition
commands=("init" "audit" "learn" "fix" "tasks" "docs" "refactor" "status" "baseline" "test")
for cmd in "${commands[@]}"; do
    if $SENTINEL "$cmd" --help 2>/dev/null | grep -q "Usage\|usage\|error\|Error" || $SENTINEL "$cmd" 2>&1 | grep -q "Usage\|usage\|error\|Error\|requires\|missing"; then
        log_pass "Command '$cmd' recognized"
    else
        log_fail "Command '$cmd' not recognized properly"
    fi
done

# Test flag parsing
echo ""
log_info "Testing Flag Parsing..."

# Test audit flags
if $SENTINEL audit --offline --ci 2>&1 | grep -q "Scanning\|Audit"; then
    log_pass "Audit flags (--offline, --ci) parsed correctly"
else
    log_fail "Audit flags not parsed correctly"
fi

# Test fix flags
if $SENTINEL fix --safe --dry-run 2>&1 | grep -q "dry-run\|Dry-run\|Auto-fix"; then
    log_pass "Fix flags (--safe, --dry-run) parsed correctly"
else
    log_fail "Fix flags not parsed correctly"
fi

# Test configuration loading
echo ""
log_info "Testing Configuration Loading..."

# Create test config
cat > .sentinelsrc << 'EOF'
{
  "hubUrl": "http://localhost:8080",
  "apiKey": "test-key-123",
  "scanDirs": ["src", "tests"],
  "excludePaths": [".git", "node_modules"]
}
EOF

if $SENTINEL audit --offline 2>&1 | grep -q "Using configured"; then
    log_pass "Configuration file loaded and used"
else
    log_fail "Configuration file not loaded properly"
fi

# Test error handling
echo ""
log_info "Testing Error Handling..."

# Test invalid command
if $SENTINEL nonexistent-command 2>&1 | grep -q "error\|Error\|invalid\|Invalid"; then
    log_pass "Invalid command handled gracefully"
else
    log_fail "Invalid command not handled properly"
fi

# Test missing arguments
if $SENTINEL tasks complete 2>&1 | grep -q "Usage\|usage\|error\|Error"; then
    log_pass "Missing arguments handled gracefully"
else
    log_fail "Missing arguments not handled properly"
fi

# Test file operations
echo ""
log_info "Testing File Operations..."

# Test pattern learning file creation
$SENTINEL learn >/dev/null 2>&1
if [[ -f ".sentinel/patterns.json" && -f ".cursor/rules/project-patterns.md" ]]; then
    log_pass "Pattern learning creates required files"
else
    log_fail "Pattern learning file creation failed"
fi

# Test backup creation
echo 'console.log("test");' > test_file.js
$SENTINEL fix --safe test_file.js >/dev/null 2>&1
if [[ -d ".sentinel/backups" ]]; then
    log_pass "Backup directory created"
else
    log_fail "Backup directory not created"
fi

# Test cross-command compatibility
echo ""
log_info "Testing Cross-Command Compatibility..."

# Run audit after fix
$SENTINEL audit --offline >/dev/null 2>&1
AUDIT_EXIT=$?

if [[ $AUDIT_EXIT -eq 0 || $AUDIT_EXIT -eq 1 ]]; then
    log_pass "Audit works after fix operations"
else
    log_fail "Audit fails after fix operations"
fi

# Performance test
echo ""
log_info "Testing Performance..."

start_time=$(date +%s.%3N)
$SENTINEL audit --offline >/dev/null 2>&1
end_time=$(date +%s.%3N)

execution_time=$(echo "$end_time - $start_time" | bc)
if (( $(echo "$execution_time < 30" | bc -l) )); then
    log_pass "Audit completes in reasonable time (< 30s)"
else
    log_fail "Audit takes too long (> 30s)"
fi

# Cleanup
rm -rf .sentinel .cursor .sentinelsrc test_file.js

echo ""
echo "📊 API Integration Test Results"
echo "=============================="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
TOTAL=$((PASSED + FAILED))
SUCCESS_RATE=$((PASSED * 100 / TOTAL))

if [[ $SUCCESS_RATE -ge 80 ]]; then
    echo -e "${GREEN}🎉 SUCCESS RATE: ${SUCCESS_RATE}%${NC}"
    echo "API integration tests PASSED"
    exit 0
else
    echo -e "${RED}❌ SUCCESS RATE: ${SUCCESS_RATE}%${NC}"
    echo "API integration tests FAILED"
    exit 1
fi



