#!/bin/bash
# Test Utility Functions
# Usage: source test_utils.sh

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TEST_PASSED=0
TEST_FAILED=0
TEST_TOTAL=0

# Assert functions
assert_equal() {
    local expected=$1
    local actual=$2
    local message=${3:-"Values should be equal"}
    
    TEST_TOTAL=$((TEST_TOTAL + 1))
    
    if [ "$expected" = "$actual" ]; then
        echo -e "${GREEN}✓${NC} $message"
        TEST_PASSED=$((TEST_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $message (expected: $expected, actual: $actual)"
        TEST_FAILED=$((TEST_FAILED + 1))
        return 1
    fi
}

assert_not_equal() {
    local expected=$1
    local actual=$2
    local message=${3:-"Values should not be equal"}
    
    TEST_TOTAL=$((TEST_TOTAL + 1))
    
    if [ "$expected" != "$actual" ]; then
        echo -e "${GREEN}✓${NC} $message"
        TEST_PASSED=$((TEST_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $message (expected: $expected, actual: $actual)"
        TEST_FAILED=$((TEST_FAILED + 1))
        return 1
    fi
}

assert_contains() {
    local haystack=$1
    local needle=$2
    local message=${3:-"String should contain substring"}
    
    TEST_TOTAL=$((TEST_TOTAL + 1))
    
    if [[ "$haystack" == *"$needle"* ]]; then
        echo -e "${GREEN}✓${NC} $message"
        TEST_PASSED=$((TEST_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $message (haystack: $haystack, needle: $needle)"
        TEST_FAILED=$((TEST_FAILED + 1))
        return 1
    fi
}

assert_exit_code() {
    local expected=$1
    local command=$2
    local message=${3:-"Command should exit with code $expected"}
    
    TEST_TOTAL=$((TEST_TOTAL + 1))
    
    eval "$command" > /dev/null 2>&1
    local actual=$?
    
    if [ $actual -eq $expected ]; then
        echo -e "${GREEN}✓${NC} $message"
        TEST_PASSED=$((TEST_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $message (expected: $expected, actual: $actual)"
        TEST_FAILED=$((TEST_FAILED + 1))
        return 1
    fi
}

# Test summary
print_test_summary() {
    echo ""
    echo "=========================================="
    echo "Test Summary"
    echo "=========================================="
    echo "Total:  $TEST_TOTAL"
    echo -e "${GREEN}Passed: $TEST_PASSED${NC}"
    echo -e "${RED}Failed: $TEST_FAILED${NC}"
    echo "=========================================="
    
    if [ $TEST_FAILED -eq 0 ]; then
        return 0
    else
        return 1
    fi
}

# Cleanup function
cleanup_test_env() {
    # Kill any background processes
    jobs -p | xargs -r kill 2>/dev/null
    
    # Clean up temp files
    rm -f /tmp/test_*.json /tmp/test_*.log /tmp/mock_server_*.py
    
    # Reset counters
    TEST_PASSED=0
    TEST_FAILED=0
    TEST_TOTAL=0
}












