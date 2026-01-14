#!/bin/bash

# Phase 18: Performance Testing Script
# Comprehensive performance benchmarking for Sentinel Hub

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
HUB_URL="${HUB_URL:-http://localhost:8080}"
TEST_DURATION="${TEST_DURATION:-60}"  # seconds
CONCURRENT_USERS="${CONCURRENT_USERS:-10}"
API_KEY="${API_KEY:-test-api-key-12345678901234567890}"

# Results tracking
TOTAL_REQUESTS=0
SUCCESSFUL_REQUESTS=0
FAILED_REQUESTS=0
TOTAL_RESPONSE_TIME=0
MIN_RESPONSE_TIME=999999
MAX_RESPONSE_TIME=0

# Arrays to store response times for percentile calculations
declare -a RESPONSE_TIMES

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to record response time
record_response() {
    local response_time=$1
    local success=$2

    ((TOTAL_REQUESTS++))
    TOTAL_RESPONSE_TIME=$((TOTAL_RESPONSE_TIME + response_time))

    if [ "$success" = "true" ]; then
        ((SUCCESSFUL_REQUESTS++))
    else
        ((FAILED_REQUESTS++))
    fi

    RESPONSE_TIMES+=($response_time)

    if [ $response_time -lt $MIN_RESPONSE_TIME ]; then
        MIN_RESPONSE_TIME=$response_time
    fi

    if [ $response_time -gt $MAX_RESPONSE_TIME ]; then
        MAX_RESPONSE_TIME=$response_time
    fi
}

# Function to calculate percentile
calculate_percentile() {
    local percentile=$1
    local sorted_times=($(printf '%s\n' "${RESPONSE_TIMES[@]}" | sort -n))
    local index=$(( (percentile * ${#sorted_times[@]}) / 100 ))
    if [ $index -ge ${#sorted_times[@]} ]; then
        index=$(( ${#sorted_times[@]} - 1 ))
    fi
    echo "${sorted_times[$index]}"
}

# Function to test endpoint with timing
test_endpoint() {
    local endpoint=$1
    local method=${2:-GET}
    local data=${3:-}

    local start_time=$(date +%s000)

    local response
    if [ "$method" = "POST" ]; then
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
            -H "Authorization: Bearer $API_KEY" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$HUB_URL$endpoint" 2>/dev/null)
    else
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -H "Authorization: Bearer $API_KEY" \
            "$HUB_URL$endpoint" 2>/dev/null)
    fi

    local end_time=$(date +%s000)
    local response_time=$((end_time - start_time))

    local http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

    local success=false
    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ] || [ "$http_code" = "202" ]; then
        success=true
    fi

    record_response $response_time $success

    echo "$response_time:$success:$http_code"
}

# Function to run concurrent load test
run_load_test() {
    local endpoint=$1
    local method=${2:-GET}
    local data=${3:-}
    local requests_per_user=$4

    log_info "Running load test: $endpoint ($CONCURRENT_USERS users, $requests_per_user requests each)"

    local pids=()

    for ((i=1; i<=CONCURRENT_USERS; i++)); do
        (
            for ((j=1; j<=requests_per_user; j++)); do
                test_endpoint "$endpoint" "$method" "$data" > /dev/null
                sleep 0.1  # Small delay between requests
            done
        ) &
        pids+=($!)
    done

    # Wait for all background processes to complete
    for pid in "${pids[@]}"; do
        wait $pid
    done
}

echo "========================================="
echo "⚡ SENTINEL PERFORMANCE TEST SUITE"
echo "========================================="
echo "Hub URL: $HUB_URL"
echo "Duration: $TEST_DURATION seconds"
echo "Concurrent Users: $CONCURRENT_USERS"
echo "Started: $(date)"
echo ""

# Check if Hub is running
log_info "Checking if Sentinel Hub is accessible..."
if ! curl -s -f "$HUB_URL/health" > /dev/null 2>&1; then
    log_error "Sentinel Hub is not accessible at $HUB_URL"
    log_error "Please ensure the Hub is running before executing performance tests"
    exit 1
fi
log_success "Sentinel Hub is accessible"

echo ""

# Test 1: Health Check Endpoint
echo "1. HEALTH CHECK PERFORMANCE"
echo "==========================="
run_load_test "/health" "GET" "" 20
echo ""

# Test 2: API Status Endpoint
echo "2. API STATUS PERFORMANCE"
echo "========================="
run_load_test "/api/v1/status" "GET" "" 15
echo ""

# Test 3: Knowledge Base Query (light load)
echo "3. KNOWLEDGE BASE QUERY PERFORMANCE"
echo "===================================="
run_load_test "/api/v1/projects/test-project/knowledge?limit=10" "GET" "" 10
echo ""

# Test 4: MCP Tools List (frequent IDE operation)
echo "4. MCP TOOLS LIST PERFORMANCE"
echo "============================="
run_load_test "/api/v1/mcp/tools" "GET" "" 30
echo ""

# Test 5: MCP Tool Call (simulated)
echo "5. MCP TOOL CALL PERFORMANCE"
echo "============================"
MCP_PAYLOAD='{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "sentinel_analyze_feature_comprehensive",
        "arguments": {
            "feature_description": "test feature",
            "project_context": "test context"
        }
    }
}'
run_load_test "/api/v1/mcp" "POST" "$MCP_PAYLOAD" 5
echo ""

# Test 6: Task Management (CRUD operations)
echo "6. TASK MANAGEMENT PERFORMANCE"
echo "=============================="
TASK_PAYLOAD='{
    "title": "Performance Test Task",
    "description": "Testing task creation performance",
    "priority": "medium",
    "status": "pending"
}'
run_load_test "/api/v1/tasks" "POST" "$TASK_PAYLOAD" 5
run_load_test "/api/v1/tasks?limit=10" "GET" "" 10
echo ""

# Calculate final statistics
if [ $TOTAL_REQUESTS -gt 0 ]; then
    AVERAGE_RESPONSE_TIME=$((TOTAL_RESPONSE_TIME / TOTAL_REQUESTS))
    SUCCESS_RATE=$(( (SUCCESSFUL_REQUESTS * 100) / TOTAL_REQUESTS ))
    ERROR_RATE=$(( (FAILED_REQUESTS * 100) / TOTAL_REQUESTS ))

    P50_RESPONSE_TIME=$(calculate_percentile 50)
    P95_RESPONSE_TIME=$(calculate_percentile 95)
    P99_RESPONSE_TIME=$(calculate_percentile 99)
else
    AVERAGE_RESPONSE_TIME=0
    SUCCESS_RATE=0
    ERROR_RATE=0
    P50_RESPONSE_TIME=0
    P95_RESPONSE_TIME=0
    P99_RESPONSE_TIME=0
fi

echo "========================================="
echo "📊 PERFORMANCE TEST RESULTS"
echo "========================================="

echo "Test Configuration:"
echo "  Hub URL: $HUB_URL"
echo "  Concurrent Users: $CONCURRENT_USERS"
echo "  Test Duration: ~$TEST_DURATION seconds"
echo ""

echo "Request Statistics:"
echo "  Total Requests: $TOTAL_REQUESTS"
echo "  Successful Requests: $SUCCESSFUL_REQUESTS"
echo "  Failed Requests: $FAILED_REQUESTS"
echo ""

echo "Response Time Statistics (ms):"
echo "  Average: $AVERAGE_RESPONSE_TIME"
echo "  Minimum: $MIN_RESPONSE_TIME"
echo "  Maximum: $MAX_RESPONSE_TIME"
echo "  50th Percentile (P50): $P50_RESPONSE_TIME"
echo "  95th Percentile (P95): $P95_RESPONSE_TIME"
echo "  99th Percentile (P99): $P99_RESPONSE_TIME"
echo ""

echo "Success Rates:"
echo "  Success Rate: ${SUCCESS_RATE}%"
echo "  Error Rate: ${ERROR_RATE}%"
echo ""

# Performance Assessment
echo "Performance Assessment:"
if [ $AVERAGE_RESPONSE_TIME -lt 500 ]; then
    log_success "✅ Average response time is excellent (< 500ms)"
elif [ $AVERAGE_RESPONSE_TIME -lt 1000 ]; then
    log_success "✅ Average response time is good (< 1000ms)"
elif [ $AVERAGE_RESPONSE_TIME -lt 2000 ]; then
    log_warning "⚠️  Average response time is acceptable (< 2000ms)"
else
    log_error "❌ Average response time is poor (> 2000ms)"
fi

if [ $SUCCESS_RATE -ge 95 ]; then
    log_success "✅ Success rate is excellent (≥ 95%)"
elif [ $SUCCESS_RATE -ge 90 ]; then
    log_success "✅ Success rate is good (≥ 90%)"
elif [ $SUCCESS_RATE -ge 80 ]; then
    log_warning "⚠️  Success rate is acceptable (≥ 80%)"
else
    log_error "❌ Success rate needs improvement (< 80%)"
fi

if [ $P95_RESPONSE_TIME -lt 2000 ]; then
    log_success "✅ P95 response time is excellent (< 2000ms)"
elif [ $P95_RESPONSE_TIME -lt 5000 ]; then
    log_warning "⚠️  P95 response time is acceptable (< 5000ms)"
else
    log_error "❌ P95 response time needs optimization (> 5000ms)"
fi

echo ""
echo "Recommendations:"
if [ $AVERAGE_RESPONSE_TIME -gt 1000 ]; then
    echo "  - Consider database query optimization"
    echo "  - Review caching strategies"
    echo "  - Check for N+1 query problems"
fi

if [ $SUCCESS_RATE -lt 95 ]; then
    echo "  - Investigate error patterns in logs"
    echo "  - Check rate limiting configuration"
    echo "  - Verify API key validation"
fi

if [ $P95_RESPONSE_TIME -gt 3000 ]; then
    echo "  - Implement response time monitoring"
    echo "  - Consider load balancing for high traffic"
    echo "  - Optimize slowest endpoints first"
fi

echo ""
echo "Completed: $(date)"
echo "========================================="

# Exit with error if performance is critically poor
if [ $SUCCESS_RATE -lt 80 ] || [ $AVERAGE_RESPONSE_TIME -gt 5000 ]; then
    log_error "Performance test failed - critical issues detected"
    exit 1
elif [ $SUCCESS_RATE -lt 90 ] || [ $AVERAGE_RESPONSE_TIME -gt 2000 ]; then
    log_warning "Performance test completed with warnings"
    exit 0
else
    log_success "Performance test passed successfully"
    exit 0
fi
