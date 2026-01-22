#!/bin/bash
# Comprehensive Load Testing Suite for Sentinel Hub API
# Tests authentication, API endpoints, concurrent requests, and stress scenarios

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
HUB_URL="${SENTINEL_HUB_URL:-http://localhost:8080}"
API_KEY="${SENTINEL_API_KEY:-}"
DURATION=30  # seconds
CONCURRENT_USERS=10
REQUESTS_PER_SECOND=50
TIMEOUT=10

# Counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_REQUESTS=0
SUCCESSFUL_REQUESTS=0
FAILED_REQUESTS=0

# Results storage
RESULTS_DIR="/tmp/load_test_results_$(date +%s)"
mkdir -p "$RESULTS_DIR"

log_info() {
    echo -e "${BLUE}ℹ INFO:${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    ((TESTS_FAILED++))
}

log_warn() {
    echo -e "${YELLOW}⚠ WARN:${NC} $1"
}

log_section() {
    echo ""
    echo -e "${CYAN}══════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}══════════════════════════════════════════════════════════════${NC}"
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v curl >/dev/null 2>&1; then
        log_fail "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq >/dev/null 2>&1; then
        log_warn "jq not found - JSON parsing will be limited"
    fi
    
    log_success "Dependencies check passed"
}

# Make HTTP request and return metrics
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local api_key_header=$4
    
    local start_time=$(date +%s%N)
    local response
    local http_code
    
    if [ -n "$api_key_header" ]; then
        if [ -n "$data" ]; then
            response=$(curl -s -w "\n%{http_code}" -X "$method" \
                -H "Content-Type: application/json" \
                -H "X-API-Key: $api_key_header" \
                -d "$data" \
                --max-time $TIMEOUT \
                "$HUB_URL$endpoint" 2>&1)
        else
            response=$(curl -s -w "\n%{http_code}" -X "$method" \
                -H "X-API-Key: $api_key_header" \
                --max-time $TIMEOUT \
                "$HUB_URL$endpoint" 2>&1)
        fi
    else
        if [ -n "$data" ]; then
            response=$(curl -s -w "\n%{http_code}" -X "$method" \
                -H "Content-Type: application/json" \
                -d "$data" \
                --max-time $TIMEOUT \
                "$HUB_URL$endpoint" 2>&1)
        else
            response=$(curl -s -w "\n%{http_code}" -X "$method" \
                --max-time $TIMEOUT \
                "$HUB_URL$endpoint" 2>&1)
        fi
    fi
    
    local end_time=$(date +%s%N)
    local duration_ms=$(( (end_time - start_time) / 1000000 ))
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    echo "$http_code|$duration_ms|$body"
}

# Test health endpoint
test_health_endpoint() {
    log_section "Test 1: Health Endpoint Load Test"
    
    local concurrent=$1
    local requests=$2
    local success_count=0
    local fail_count=0
    local total_time=0
    local min_time=999999
    local max_time=0
    
    log_info "Sending $requests requests with $concurrent concurrent connections..."
    
    for ((i=1; i<=requests; i++)); do
        {
            result=$(make_request "GET" "/health" "" "")
            http_code=$(echo "$result" | cut -d'|' -f1)
            duration=$(echo "$result" | cut -d'|' -f2)
            
            echo "$http_code|$duration" >> "$RESULTS_DIR/health_test.txt"
            
            if [ "$http_code" = "200" ]; then
                ((success_count++))
            else
                ((fail_count++))
            fi
        } &
        
        if [ $((i % concurrent)) -eq 0 ]; then
            wait
        fi
        
        sleep 0.1
    done
    wait
    
    # Calculate statistics
    if [ -f "$RESULTS_DIR/health_test.txt" ]; then
        while IFS='|' read -r code duration; do
            total_time=$((total_time + duration))
            if [ $duration -lt $min_time ]; then
                min_time=$duration
            fi
            if [ $duration -gt $max_time ]; then
                max_time=$duration
            fi
        done < "$RESULTS_DIR/health_test.txt"
    fi
    
    local avg_time=$((total_time / requests))
    local success_rate=$((success_count * 100 / requests))
    
    echo "Results:"
    echo "  Successful: $success_count/$requests"
    echo "  Failed: $fail_count/$requests"
    echo "  Success Rate: ${success_rate}%"
    echo "  Avg Response Time: ${avg_time}ms"
    echo "  Min Response Time: ${min_time}ms"
    echo "  Max Response Time: ${max_time}ms"
    
    if [ $success_rate -ge 95 ]; then
        log_success "Health endpoint load test passed (${success_rate}% success rate)"
    else
        log_fail "Health endpoint load test failed (${success_rate}% success rate)"
    fi
}

# Test authentication under load
test_authentication_load() {
    log_section "Test 2: Authentication Load Test"
    
    if [ -z "$API_KEY" ]; then
        log_warn "API_KEY not set - skipping authentication tests"
        return
    fi
    
    local requests=100
    local success_count=0
    local fail_count=0
    local auth_fail_count=0
    
    log_info "Testing authentication with $requests requests..."
    
    for ((i=1; i<=requests; i++)); do
        {
            # Test with valid API key
            result=$(make_request "GET" "/api/v1/projects" "" "$API_KEY")
            http_code=$(echo "$result" | cut -d'|' -f1)
            
            if [ "$http_code" = "200" ] || [ "$http_code" = "401" ]; then
                if [ "$http_code" = "200" ]; then
                    ((success_count++))
                else
                    ((auth_fail_count++))
                fi
            else
                ((fail_count++))
            fi
            
            # Test with invalid API key
            result=$(make_request "GET" "/api/v1/projects" "" "invalid-key-$i")
            http_code=$(echo "$result" | cut -d'|' -f1)
            
            if [ "$http_code" = "401" ]; then
                ((success_count++))
            else
                ((fail_count++))
            fi
        } &
        
        if [ $((i % 10)) -eq 0 ]; then
            wait
            sleep 0.1
        fi
    done
    wait
    
    local total=$((requests * 2))
    local success_rate=$((success_count * 100 / total))
    
    echo "Results:"
    echo "  Successful auth checks: $success_count/$total"
    echo "  Auth failures (expected): $auth_fail_count"
    echo "  Unexpected failures: $fail_count"
    echo "  Success Rate: ${success_rate}%"
    
    if [ $success_rate -ge 95 ]; then
        log_success "Authentication load test passed"
    else
        log_fail "Authentication load test failed"
    fi
}

# Test rate limiting
test_rate_limiting() {
    log_section "Test 3: Rate Limiting Test"
    
    if [ -z "$API_KEY" ]; then
        log_warn "API_KEY not set - skipping rate limiting tests"
        return
    fi
    
    local burst_size=20
    local rate_limit_hit=0
    local normal_responses=0
    
    log_info "Testing rate limiting with burst of $burst_size requests..."
    
    for ((i=1; i<=burst_size; i++)); do
        result=$(make_request "GET" "/health" "" "$API_KEY")
        http_code=$(echo "$result" | cut -d'|' -f1)
        
        if [ "$http_code" = "429" ]; then
            ((rate_limit_hit++))
        elif [ "$http_code" = "200" ] || [ "$http_code" = "401" ]; then
            ((normal_responses++))
        fi
        
        sleep 0.05
    done
    
    echo "Results:"
    echo "  Normal responses: $normal_responses"
    echo "  Rate limit responses (429): $rate_limit_hit"
    
    if [ $rate_limit_hit -gt 0 ]; then
        log_success "Rate limiting is working (429 responses detected)"
    elif [ $normal_responses -gt 0 ]; then
        log_warn "Rate limiting may not be configured (no 429 responses)"
    else
        log_fail "Rate limiting test inconclusive"
    fi
}

# Test concurrent requests
test_concurrent_requests() {
    log_section "Test 4: Concurrent Request Load Test"
    
    local concurrent=$CONCURRENT_USERS
    local requests_per_user=10
    local total_requests=$((concurrent * requests_per_user))
    
    log_info "Testing $concurrent concurrent users with $requests_per_user requests each..."
    
    local success_count=0
    local fail_count=0
    local start_time=$(date +%s)
    
    for ((user=1; user<=concurrent; user++)); do
        (
            for ((req=1; req<=requests_per_user; req++)); do
                result=$(make_request "GET" "/health" "" "")
                http_code=$(echo "$result" | cut -d'|' -f1)
                duration=$(echo "$result" | cut -d'|' -f2)
                
                echo "$user|$req|$http_code|$duration" >> "$RESULTS_DIR/concurrent_test.txt"
                
                if [ "$http_code" = "200" ]; then
                    ((success_count++))
                else
                    ((fail_count++))
                fi
            done
        ) &
    done
    wait
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local throughput=$((total_requests / duration))
    local success_rate=$((success_count * 100 / total_requests))
    
    echo "Results:"
    echo "  Total Requests: $total_requests"
    echo "  Successful: $success_count"
    echo "  Failed: $fail_count"
    echo "  Success Rate: ${success_rate}%"
    echo "  Duration: ${duration}s"
    echo "  Throughput: ${throughput} requests/second"
    
    if [ $success_rate -ge 90 ]; then
        log_success "Concurrent request test passed (${success_rate}% success, ${throughput} req/s)"
    else
        log_fail "Concurrent request test failed (${success_rate}% success)"
    fi
}

# Stress test
stress_test() {
    log_section "Test 5: Stress Test"
    
    local duration=$DURATION
    local rate=$REQUESTS_PER_SECOND
    local total_requests=$((duration * rate))
    
    log_info "Stress test: $rate requests/second for $duration seconds..."
    
    local success_count=0
    local fail_count=0
    local start_time=$(date +%s)
    local request_count=0
    
    while [ $(($(date +%s) - start_time)) -lt $duration ]; do
        for ((i=1; i<=rate; i++)); do
            {
                result=$(make_request "GET" "/health" "" "")
                http_code=$(echo "$result" | cut -d'|' -f1)
                
                if [ "$http_code" = "200" ]; then
                    ((success_count++))
                else
                    ((fail_count++))
                fi
                ((request_count++))
            } &
        done
        
        sleep 1
    done
    wait
    
    local actual_duration=$(($(date +%s) - start_time))
    local actual_rate=$((request_count / actual_duration))
    local success_rate=$((success_count * 100 / request_count))
    
    echo "Results:"
    echo "  Total Requests: $request_count"
    echo "  Successful: $success_count"
    echo "  Failed: $fail_count"
    echo "  Success Rate: ${success_rate}%"
    echo "  Actual Rate: ${actual_rate} requests/second"
    echo "  Duration: ${actual_duration}s"
    
    if [ $success_rate -ge 80 ]; then
        log_success "Stress test passed (${success_rate}% success under load)"
    else
        log_fail "Stress test failed (${success_rate}% success rate too low)"
    fi
}

# Test response time percentiles
test_response_times() {
    log_section "Test 6: Response Time Analysis"
    
    local requests=100
    local times=()
    
    log_info "Collecting $requests requests for response time analysis..."
    
    for ((i=1; i<=requests; i++)); do
        result=$(make_request "GET" "/health" "" "")
        duration=$(echo "$result" | cut -d'|' -f2)
        times+=($duration)
        
        echo "$duration" >> "$RESULTS_DIR/response_times.txt"
        sleep 0.1
    done
    
    if [ -f "$RESULTS_DIR/response_times.txt" ]; then
        sort -n "$RESULTS_DIR/response_times.txt" > "$RESULTS_DIR/response_times_sorted.txt"
        
        local p50=$(sed -n '50p' "$RESULTS_DIR/response_times_sorted.txt")
        local p95=$(sed -n '95p' "$RESULTS_DIR/response_times_sorted.txt")
        local p99=$(sed -n '99p' "$RESULTS_DIR/response_times_sorted.txt")
        local avg=$(awk '{sum+=$1} END {print int(sum/NR)}' "$RESULTS_DIR/response_times_sorted.txt")
        local min=$(head -n1 "$RESULTS_DIR/response_times_sorted.txt")
        local max=$(tail -n1 "$RESULTS_DIR/response_times_sorted.txt")
        
        echo "Response Time Statistics:"
        echo "  Average: ${avg}ms"
        echo "  Min: ${min}ms"
        echo "  Max: ${max}ms"
        echo "  p50 (Median): ${p50}ms"
        echo "  p95: ${p95}ms"
        echo "  p99: ${p99}ms"
        
        if [ ${p95:-999999} -lt 1000 ]; then
            log_success "Response times acceptable (p95: ${p95}ms)"
        else
            log_warn "Response times may be slow (p95: ${p95}ms)"
        fi
    fi
}

# Main execution
main() {
    echo ""
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║           COMPREHENSIVE LOAD TESTING SUITE                   ║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    echo "Configuration:"
    echo "  Hub URL: $HUB_URL"
    echo "  API Key: ${API_KEY:0:10}..."
    echo "  Concurrent Users: $CONCURRENT_USERS"
    echo "  Test Duration: ${DURATION}s"
    echo "  Request Rate: ${REQUESTS_PER_SECOND} req/s"
    echo ""
    
    check_dependencies
    
    # Check if Hub is accessible
    log_info "Checking Hub availability..."
    if curl -s --max-time 5 "$HUB_URL/health" >/dev/null 2>&1; then
        log_success "Hub is accessible"
    else
        log_fail "Hub is not accessible at $HUB_URL"
        echo "Please ensure the Hub API is running and accessible"
        exit 1
    fi
    
    # Run tests
    test_health_endpoint 5 50
    test_authentication_load
    test_rate_limiting
    test_concurrent_requests
    test_response_times
    stress_test
    
    # Summary
    log_section "Load Testing Summary"
    
    echo ""
    echo "Test Results:"
    echo "  Passed: $TESTS_PASSED"
    echo "  Failed: $TESTS_FAILED"
    echo "  Success Rate: $((TESTS_PASSED * 100 / (TESTS_PASSED + TESTS_FAILED)))%"
    echo ""
    echo "Detailed results saved to: $RESULTS_DIR"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✓✓✓ ALL LOAD TESTS PASSED ✓✓✓${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some load tests failed${NC}"
        exit 1
    fi
}

main
