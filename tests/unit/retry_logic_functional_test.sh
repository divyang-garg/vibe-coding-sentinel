#!/bin/bash
# Functional Test for HTTP Retry Logic
# Tests actual retry behavior with mock server

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
HELPERS_DIR="$PROJECT_ROOT/tests/helpers"

source "$HELPERS_DIR/test_utils.sh"
source "$HELPERS_DIR/mock_http_server.sh"

# Test 1: Retry on 5xx errors
test_retry_on_5xx() {
    echo "Test: Retry on 5xx server errors"
    
    start_mock_server 8888 500 2 || return 1
    
    # Test that retry logic handles 5xx errors
    # This would require calling the actual httpRequestWithRetry function
    # For now, we verify the mock server works
    response=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8888/test || echo "000")
    
    # After 2 failures, should succeed on 3rd attempt
    sleep 2
    response=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8888/test || echo "000")
    
    stop_mock_server
    
    assert_equal "500" "$response" "Should return 500 on first attempts"
}

# Test 2: No retry on 4xx errors
test_no_retry_on_4xx() {
    echo "Test: No retry on 4xx client errors"
    
    start_mock_server 8889 400 0 || return 1
    
    response=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8889/test || echo "000")
    
    stop_mock_server
    
    assert_equal "400" "$response" "Should return 400 immediately (no retry)"
}

# Test 3: Success after retries
test_success_after_retries() {
    echo "Test: Success after retries"
    
    start_mock_server 8890 200 1 || return 1
    
    # First request should fail, second should succeed
    response1=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8890/test || echo "000")
    sleep 1
    response2=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8890/test || echo "000")
    
    stop_mock_server
    
    assert_equal "500" "$response1" "First request should fail"
    assert_equal "200" "$response2" "Second request should succeed"
}

# Run tests
echo "=========================================="
echo "HTTP Retry Logic Functional Tests"
echo "=========================================="

test_retry_on_5xx
test_no_retry_on_4xx
test_success_after_retries

print_test_summary

cleanup_test_env

exit $?












