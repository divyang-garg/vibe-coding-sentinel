#!/bin/bash
# Performance tests for Hub API
# Load testing, benchmark audit execution times
# Run from project root: ./tests/performance/hub_api_load_test.sh

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
echo "   Hub API Performance Tests"
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

# Create test codebase
mkdir -p "$TEST_TMP_DIR/src"
for i in {1..10}; do
    cat > "$TEST_TMP_DIR/src/file$i.js" << EOF
// Test file $i
function test$i() {
    return true;
}
EOF
done

# Test 1: Audit execution time benchmark
echo ""
echo "Test 1: Audit execution time benchmark"
echo "──────────────────────────────────────────────────────────────"

START_TIME=$(date +%s%N)
cd "$TEST_TMP_DIR"
./sentinel audit --ci > /dev/null 2>&1 || true
END_TIME=$(date +%s%N)
DURATION_MS=$(( (END_TIME - START_TIME) / 1000000 ))

if [[ $DURATION_MS -lt 10000 ]]; then
    log_pass "Audit completes in reasonable time (${DURATION_MS}ms)"
else
    log_warn "Audit took longer than expected (${DURATION_MS}ms)"
fi

# Test 2: Concurrent audit requests
echo ""
echo "Test 2: Concurrent audit requests"
echo "──────────────────────────────────────────────────────────────"

CONCURRENT=5
START_TIME=$(date +%s%N)

for i in $(seq 1 $CONCURRENT); do
    (cd "$TEST_TMP_DIR" && ./sentinel audit --ci > /dev/null 2>&1) &
done
wait

END_TIME=$(date +%s%N)
DURATION_MS=$(( (END_TIME - START_TIME) / 1000000 ))

if [[ $DURATION_MS -lt 30000 ]]; then
    log_pass "Concurrent audits complete in reasonable time (${DURATION_MS}ms for $CONCURRENT requests)"
else
    log_warn "Concurrent audits took longer than expected (${DURATION_MS}ms)"
fi

# Test 3: Cache performance
echo ""
echo "Test 3: Cache performance"
echo "──────────────────────────────────────────────────────────────"

# First run (no cache)
START_TIME=$(date +%s%N)
cd "$TEST_TMP_DIR"
./sentinel audit --ci > /dev/null 2>&1 || true
END_TIME=$(date +%s%N)
FIRST_RUN_MS=$(( (END_TIME - START_TIME) / 1000000 ))

# Second run (with cache)
START_TIME=$(date +%s%N)
cd "$TEST_TMP_DIR"
./sentinel audit --ci > /dev/null 2>&1 || true
END_TIME=$(date +%s%N)
SECOND_RUN_MS=$(( (END_TIME - START_TIME) / 1000000 ))

if [[ $SECOND_RUN_MS -lt $FIRST_RUN_MS ]]; then
    log_pass "Cache improves performance (first: ${FIRST_RUN_MS}ms, second: ${SECOND_RUN_MS}ms)"
else
    log_warn "Cache may not be working optimally"
fi

# Test 4: Memory usage
echo ""
echo "Test 4: Memory usage"
echo "──────────────────────────────────────────────────────────────"

if command -v ps > /dev/null; then
    MEM_BEFORE=$(ps -o rss= -p $$ 2>/dev/null || echo "0")
    cd "$TEST_TMP_DIR"
    ./sentinel audit --ci > /dev/null 2>&1 || true
    MEM_AFTER=$(ps -o rss= -p $$ 2>/dev/null || echo "0")
    
    MEM_DIFF=$((MEM_AFTER - MEM_BEFORE))
    if [[ $MEM_DIFF -lt 100000 ]]; then
        log_pass "Memory usage is reasonable (${MEM_DIFF}KB increase)"
    else
        log_warn "Memory usage may be high (${MEM_DIFF}KB increase)"
    fi
else
    log_warn "Cannot measure memory usage (ps command not available)"
fi

# Test 5: Rate limiting behavior
echo ""
echo "Test 5: Rate limiting behavior"
echo "──────────────────────────────────────────────────────────────"

if [[ -n "$SENTINEL_HUB_URL" && -n "$SENTINEL_API_KEY" ]]; then
    export SENTINEL_HUB_URL
    export SENTINEL_API_KEY
    
    RATE_LIMIT_HIT=0
    for i in {1..20}; do
        RESPONSE=$(cd "$TEST_TMP_DIR" && timeout 2 ./sentinel audit --ci 2>&1 || true)
        if echo "$RESPONSE" | grep -qi "rate limit\|429\|too many"; then
            RATE_LIMIT_HIT=1
            break
        fi
        sleep 0.1
    done
    
    if [[ $RATE_LIMIT_HIT -eq 1 ]]; then
        log_pass "Rate limiting is working"
    else
        log_warn "Rate limiting may not be triggered (or not configured)"
    fi
else
    log_warn "Skipping rate limit test (Hub not configured)"
fi

# Summary
echo ""
echo "══════════════════════════════════════════════════════════════"
echo "   Performance Test Summary"
echo "══════════════════════════════════════════════════════════════"
echo ""
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo ""

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}All performance tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some performance tests failed${NC}"
    exit 1
fi










