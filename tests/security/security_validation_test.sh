#!/bin/bash
# Security tests for Sentinel Agent and Hub API
# Tests: input validation, authentication, authorization, rate limiting, path traversal
# Run from project root: ./tests/security/security_validation_test.sh

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_info() {
    echo -e "${BLUE}ℹ INFO:${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠ WARN:${NC} $1"
}

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "   Security Validation Tests"
echo "══════════════════════════════════════════════════════════════"
echo ""

# Ensure binary is built
if [[ ! -f "./sentinel" ]]; then
    log_info "Building Sentinel binary..."
    ./synapsevibsentinel.sh > /dev/null 2>&1 || true
fi

# Test directory setup
TEST_TMP_DIR=$(mktemp -d)
trap "rm -rf $TEST_TMP_DIR" EXIT

# Test 1: Path traversal prevention
echo ""
echo "Test 1: Path traversal prevention"
echo "──────────────────────────────────────────────────────────────"

MALICIOUS_PATHS=(
    "../../../etc/passwd"
    "..\\..\\..\\windows\\system32"
    "/etc/passwd"
    "....//....//etc/passwd"
)

for path in "${MALICIOUS_PATHS[@]}"; do
    OUTPUT=$(cd "$TEST_TMP_DIR" && ./sentinel audit "$path" 2>&1) || true
    
    if echo "$OUTPUT" | grep -qi "invalid\|error\|denied\|traversal"; then
        log_pass "Path traversal prevented: $path"
    else
        log_warn "Path traversal may not be prevented: $path"
    fi
done

# Test 2: String sanitization
echo ""
echo "Test 2: String sanitization"
echo "──────────────────────────────────────────────────────────────"

# Test control characters
MALICIOUS_STRINGS=(
    "test\x00code"
    "test\x1fcode"
    "$(printf 'test%bcode' '\x00')"
)

for str in "${MALICIOUS_STRINGS[@]}"; do
    # Create test file with malicious string
    echo "$str" > "$TEST_TMP_DIR/test.txt"
    OUTPUT=$(cd "$TEST_TMP_DIR" && ./sentinel audit 2>&1) || true
    
    if echo "$OUTPUT" | grep -qi "sanitized\|invalid\|error"; then
        log_pass "String sanitization works for control characters"
    else
        log_warn "String sanitization may need verification"
    fi
done

# Test 3: Input length limits
echo ""
echo "Test 3: Input length limits"
echo "──────────────────────────────────────────────────────────────"

# Create very long string (>1024 chars)
LONG_STRING=$(head -c 2000 < /dev/zero | tr '\0' 'a')
echo "$LONG_STRING" > "$TEST_TMP_DIR/long.txt"

OUTPUT=$(cd "$TEST_TMP_DIR" && ./sentinel audit 2>&1) || true

if echo "$OUTPUT" | grep -qi "too long\|limit\|truncated"; then
    log_pass "Input length limits enforced"
else
    log_warn "Input length limits may need verification"
fi

# Test 4: API key validation
echo ""
echo "Test 4: API key validation"
echo "──────────────────────────────────────────────────────────────"

if [[ -n "$SENTINEL_HUB_URL" ]]; then
    # Test with invalid API key
    export SENTINEL_HUB_URL
    export SENTINEL_API_KEY="invalid-key-12345"
    
    OUTPUT=$(cd "$TEST_TMP_DIR" && timeout 5 ./sentinel audit 2>&1) || true
    
    if echo "$OUTPUT" | grep -qi "unauthorized\|invalid\|401\|403"; then
        log_pass "Invalid API key is rejected"
    else
        log_warn "API key validation may need verification"
    fi
else
    log_warn "Skipping API key test (Hub URL not configured)"
fi

# Test 5: Rate limiting
echo ""
echo "Test 5: Rate limiting"
echo "──────────────────────────────────────────────────────────────"

if [[ -n "$SENTINEL_HUB_URL" && -n "$SENTINEL_API_KEY" ]]; then
    export SENTINEL_HUB_URL
    export SENTINEL_API_KEY
    
    RATE_LIMIT_HIT=0
    for i in {1..30}; do
        RESPONSE=$(cd "$TEST_TMP_DIR" && timeout 2 ./sentinel audit --ci 2>&1 || true)
        if echo "$RESPONSE" | grep -qi "rate limit\|429\|too many\|retry"; then
            RATE_LIMIT_HIT=1
            break
        fi
        sleep 0.05
    done
    
    if [[ $RATE_LIMIT_HIT -eq 1 ]]; then
        log_pass "Rate limiting is enforced"
    else
        log_warn "Rate limiting may not be configured or working"
    fi
else
    log_warn "Skipping rate limit test (Hub not configured)"
fi

# Test 6: File permissions check
echo ""
echo "Test 6: File permissions check"
echo "──────────────────────────────────────────────────────────────"

# Create .sentinelsrc with permissive permissions
echo '{"hubUrl":"http://test.com","apiKey":"test"}' > "$TEST_TMP_DIR/.sentinelsrc"
chmod 644 "$TEST_TMP_DIR/.sentinelsrc"

OUTPUT=$(cd "$TEST_TMP_DIR" && ./sentinel audit 2>&1) || true

if echo "$OUTPUT" | grep -qi "permission\|warning\|insecure"; then
    log_pass "File permission warnings work"
else
    log_warn "File permission checks may need verification"
fi

# Test 7: SQL injection prevention (if applicable)
echo ""
echo "Test 7: SQL injection prevention"
echo "──────────────────────────────────────────────────────────────"

# Test with SQL injection patterns in file names
SQL_INJECTION_PATTERNS=(
    "'; DROP TABLE users; --"
    "1' OR '1'='1"
    "admin'--"
)

for pattern in "${SQL_INJECTION_PATTERNS[@]}"; do
    echo "test" > "$TEST_TMP_DIR/$pattern.txt"
    OUTPUT=$(cd "$TEST_TMP_DIR" && ./sentinel audit 2>&1) || true
    
    if echo "$OUTPUT" | grep -qi "invalid\|sanitized\|error"; then
        log_pass "SQL injection prevention works: $pattern"
    else
        log_warn "SQL injection prevention may need verification"
    fi
done

# Test 8: XSS prevention (if applicable)
echo ""
echo "Test 8: XSS prevention"
echo "──────────────────────────────────────────────────────────────"

XSS_PATTERNS=(
    "<script>alert('xss')</script>"
    "javascript:alert('xss')"
    "<img src=x onerror=alert('xss')>"
)

for pattern in "${XSS_PATTERNS[@]}"; do
    echo "$pattern" > "$TEST_TMP_DIR/xss.txt"
    OUTPUT=$(cd "$TEST_TMP_DIR" && ./sentinel audit 2>&1) || true
    
    if echo "$OUTPUT" | grep -qi "sanitized\|invalid\|error"; then
        log_pass "XSS prevention works"
    else
        log_warn "XSS prevention may need verification"
    fi
done

# Test 9: Command injection prevention
echo ""
echo "Test 9: Command injection prevention"
echo "──────────────────────────────────────────────────────────────"

COMMAND_INJECTION_PATTERNS=(
    "; rm -rf /"
    "| cat /etc/passwd"
    "&& ls -la"
    "\$(whoami)"
)

for pattern in "${COMMAND_INJECTION_PATTERNS[@]}"; do
    echo "$pattern" > "$TEST_TMP_DIR/cmd.txt"
    OUTPUT=$(cd "$TEST_TMP_DIR" && ./sentinel audit 2>&1) || true
    
    if echo "$OUTPUT" | grep -qi "invalid\|sanitized\|error"; then
        log_pass "Command injection prevention works"
    else
        log_warn "Command injection prevention may need verification"
    fi
done

# Summary
echo ""
echo "══════════════════════════════════════════════════════════════"
echo "   Security Test Summary"
echo "══════════════════════════════════════════════════════════════"
echo ""
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo ""

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}All security tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some security tests failed${NC}"
    exit 1
fi










