#!/bin/bash
# Test Performance Monitoring and Alerting
# Monitors test execution times and alerts on performance regressions
# Run from project root: ./tests/monitoring/test_performance_monitor.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MONITORING_DIR="$PROJECT_ROOT/tests/monitoring"
REPORTS_DIR="$MONITORING_DIR/reports"
BASELINE_DIR="$MONITORING_DIR/baselines"

# Performance thresholds (seconds)
UNIT_TEST_TIMEOUT=30
INTEGRATION_TEST_TIMEOUT=120
E2E_TEST_TIMEOUT=300
PERFORMANCE_TEST_TIMEOUT=600

# Regression thresholds (percentage increase)
REGRESSION_WARNING=50   # 50% slower
REGRESSION_CRITICAL=100 # 100% slower (2x slower)

# Create directories
mkdir -p "$REPORTS_DIR" "$BASELINE_DIR"

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_header() {
    echo -e "${PURPLE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}═══════════════════════════════════════════════════════════════${NC}"
}

# Function to run tests with timing
run_timed_tests() {
    local test_type="$1"
    local test_pattern="$2"
    local timeout="$3"
    local report_file="$REPORTS_DIR/${test_type}_timing_$(date '+%Y%m%d_%H%M%S').json"

    log_info "Running $test_type tests with timeout ${timeout}s..."

    # Start timing
    local start_time=$(date +%s.%3N)

    # Run tests with timeout and capture output
    if timeout "$timeout" bash -c "$test_pattern" 2>&1 | tee "$REPORTS_DIR/${test_type}_output.log"; then
        local exit_code=0
        log_success "$test_type tests completed successfully"
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            log_error "$test_type tests timed out after ${timeout}s"
        else
            log_error "$test_type tests failed with exit code $exit_code"
        fi
    fi

    # Calculate duration
    local end_time=$(date +%s.%3N)
    local duration=$(echo "$end_time - $start_time" | bc)

    # Generate timing report
    generate_timing_report "$test_type" "$duration" "$exit_code" "$timeout" "$report_file"

    # Check for regressions
    check_performance_regression "$test_type" "$duration"

    return $exit_code
}

# Function to generate timing report
generate_timing_report() {
    local test_type="$1"
    local duration="$2"
    local exit_code="$3"
    local timeout="$4"
    local report_file="$5"

    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    cat > "$report_file" << EOF
{
  "timestamp": "$timestamp",
  "test_type": "$test_type",
  "duration_seconds": $duration,
  "exit_code": $exit_code,
  "timeout_seconds": $timeout,
  "status": "$( [ $exit_code -eq 0 ] && echo "PASSED" || echo "FAILED" )",
  "timeout_exceeded": $( [ $exit_code -eq 124 ] && echo "true" || echo "false" ),
  "performance_threshold_seconds": $timeout,
  "performance_compliant": $( [ $(echo "$duration < $timeout" | bc) -eq 1 ] && echo "true" || echo "false" )
}
EOF

    log_success "Timing report generated: $report_file"
}

# Function to check performance regression
check_performance_regression() {
    local test_type="$1"
    local current_duration="$2"
    local baseline_file="$BASELINE_DIR/${test_type}_baseline.json"

    # Load baseline if it exists
    if [ -f "$baseline_file" ]; then
        local baseline_duration=$(jq -r '.duration_seconds' "$baseline_file" 2>/dev/null || echo "0")

        if [ "$baseline_duration" != "0" ] && (( $(echo "$baseline_duration > 0" | bc -l) )); then
            # Calculate regression percentage
            local regression_percent=$(echo "scale=2; (($current_duration - $baseline_duration) / $baseline_duration) * 100" | bc)

            log_info "Performance comparison for $test_type:"
            echo -e "${CYAN}  Baseline:${NC} ${baseline_duration}s"
            echo -e "${CYAN}  Current: ${NC} ${current_duration}s"
            echo -e "${CYAN}  Change:  ${NC} ${regression_percent}%"

            # Check regression thresholds
            if (( $(echo "$regression_percent >= $REGRESSION_CRITICAL" | bc -l) )); then
                log_error "CRITICAL PERFORMANCE REGRESSION: ${regression_percent}% slower than baseline"
                return 1
            elif (( $(echo "$regression_percent >= $REGRESSION_WARNING" | bc -l) )); then
                log_warning "Performance regression warning: ${regression_percent}% slower than baseline"
                return 0
            else
                log_success "Performance within acceptable range"
                return 0
            fi
        fi
    fi

    # Create/update baseline
    update_baseline "$test_type" "$current_duration"
    log_info "Baseline updated for $test_type: ${current_duration}s"
    return 0
}

# Function to update baseline
update_baseline() {
    local test_type="$1"
    local duration="$2"
    local baseline_file="$BASELINE_DIR/${test_type}_baseline.json"

    cat > "$baseline_file" << EOF
{
  "test_type": "$test_type",
  "duration_seconds": $duration,
  "updated_at": "$(date '+%Y-%m-%d %H:%M:%S')",
  "note": "Auto-generated baseline - update manually to lock specific performance targets"
}
EOF
}

# Function to run all test suites with monitoring
run_all_tests_with_monitoring() {
    log_header "TEST PERFORMANCE MONITORING SESSION"

    local total_start=$(date +%s.%3N)
    local results=()

    # Unit Tests
    log_info "Starting unit test monitoring..."
    if run_timed_tests "unit" "./tests/run_all_tests.sh" "$UNIT_TEST_TIMEOUT"; then
        results+=("unit:PASSED")
    else
        results+=("unit:FAILED")
    fi

    echo ""

    # Integration Tests
    log_info "Starting integration test monitoring..."
    if run_timed_tests "integration" "find ./tests/integration -name '*.sh' -exec bash {} \;" "$INTEGRATION_TEST_TIMEOUT"; then
        results+=("integration:PASSED")
    else
        results+=("integration:FAILED")
    fi

    echo ""

    # E2E Tests
    log_info "Starting E2E test monitoring..."
    if run_timed_tests "e2e" "find ./tests/e2e -name '*.sh' -exec bash {} \;" "$E2E_TEST_TIMEOUT"; then
        results+=("e2e:PASSED")
    else
        results+=("e2e:FAILED")
    fi

    echo ""

    # Performance Tests
    log_info "Starting performance test monitoring..."
    if run_timed_tests "performance" "find ./tests/performance -name '*.sh' -exec bash {} \;" "$PERFORMANCE_TEST_TIMEOUT"; then
        results+=("performance:PASSED")
    else
        results+=("performance:FAILED")
    fi

    # Calculate total duration
    local total_end=$(date +%s.%3N)
    local total_duration=$(echo "$total_end - $total_start" | bc)

    # Generate session summary
    generate_session_summary "${results[@]}" "$total_duration"
}

# Function to generate session summary
generate_session_summary() {
    local results=("$@")
    local total_duration="${results[-1]}"
    unset 'results[-1]'

    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local summary_file="$REPORTS_DIR/performance_summary_$(date '+%Y%m%d_%H%M%S').json"

    # Count results
    local passed=0
    local failed=0
    local test_results=()

    for result in "${results[@]}"; do
        local test_type=$(echo "$result" | cut -d: -f1)
        local status=$(echo "$result" | cut -d: -f2)

        test_results+=("{\"type\":\"$test_type\",\"status\":\"$status\"}")

        if [ "$status" = "PASSED" ]; then
            ((passed++))
        else
            ((failed++))
        fi
    done

    # Join test results
    local test_results_json=$(IFS=,; echo "[${test_results[*]}]")

    cat > "$summary_file" << EOF
{
  "timestamp": "$timestamp",
  "session_duration_seconds": $total_duration,
  "tests_run": $((${#results[@]})),
  "tests_passed": $passed,
  "tests_failed": $failed,
  "overall_status": "$( [ $failed -eq 0 ] && echo "SUCCESS" || echo "FAILURE" )",
  "test_results": $test_results_json,
  "performance_thresholds": {
    "unit_test_timeout_seconds": $UNIT_TEST_TIMEOUT,
    "integration_test_timeout_seconds": $INTEGRATION_TEST_TIMEOUT,
    "e2e_test_timeout_seconds": $E2E_TEST_TIMEOUT,
    "performance_test_timeout_seconds": $PERFORMANCE_TEST_TIMEOUT
  },
  "regression_thresholds": {
    "warning_percent": $REGRESSION_WARNING,
    "critical_percent": $REGRESSION_CRITICAL
  },
  "report_files": [
    "$REPORTS_DIR/unit_timing_*.json",
    "$REPORTS_DIR/integration_timing_*.json",
    "$REPORTS_DIR/e2e_timing_*.json",
    "$REPORTS_DIR/performance_timing_*.json"
  ]
}
EOF

    # Display summary
    log_header "PERFORMANCE MONITORING SUMMARY"
    echo -e "${CYAN}Session Duration:${NC} ${total_duration}s"
    echo -e "${CYAN}Tests Run:${NC} ${#results[@]}"
    echo -e "${CYAN}Tests Passed:${NC} $passed"
    echo -e "${CYAN}Tests Failed:${NC} $failed"
    echo -e "${CYAN}Status:${NC} $( [ $failed -eq 0 ] && echo "✅ SUCCESS" || echo "❌ FAILURE" )"
    echo ""
    echo -e "${CYAN}Report saved:${NC} $summary_file"

    # Show individual results
    echo ""
    log_info "Individual test results:"
    for result in "${results[@]}"; do
        local test_type=$(echo "$result" | cut -d: -f1)
        local status=$(echo "$result" | cut -d: -f2)
        if [ "$status" = "PASSED" ]; then
            echo -e "  ${GREEN}✅ $test_type${NC}"
        else
            echo -e "  ${RED}❌ $test_type${NC}"
        fi
    done

    return $failed
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Test Performance Monitoring and Alerting for Sentinel"
    echo ""
    echo "OPTIONS:"
    echo "  --help              Show this help message"
    echo "  --unit-only         Run unit tests only"
    echo "  --integration-only  Run integration tests only"
    echo "  --e2e-only          Run E2E tests only"
    echo "  --performance-only  Run performance tests only"
    echo "  --update-baselines  Force update all performance baselines"
    echo "  --ci                CI/CD mode - strict performance checking"
    echo ""
    echo "THRESHOLDS:"
    echo "  Unit Tests: ${UNIT_TEST_TIMEOUT}s timeout"
    echo "  Integration: ${INTEGRATION_TEST_TIMEOUT}s timeout"
    echo "  E2E Tests: ${E2E_TEST_TIMEOUT}s timeout"
    echo "  Performance: ${PERFORMANCE_TEST_TIMEOUT}s timeout"
    echo "  Regression Warning: ${REGRESSION_WARNING}% slower"
    echo "  Regression Critical: ${REGRESSION_CRITICAL}% slower"
    echo ""
    echo "REPORTS GENERATED:"
    echo "  • $REPORTS_DIR/*_timing_*.json         - Individual test timing"
    echo "  • $REPORTS_DIR/performance_summary_*.json - Session summary"
    echo "  • $BASELINE_DIR/*_baseline.json        - Performance baselines"
    echo ""
    echo "CODING_STANDARDS.md COMPLIANCE:"
    echo "  • Performance monitoring with automated regression detection"
    echo "  • SLA tracking for test execution times"
    echo "  • Automated alerting for performance violations"
}

# Parse command line arguments
UNIT_ONLY=false
INTEGRATION_ONLY=false
E2E_ONLY=false
PERFORMANCE_ONLY=false
UPDATE_BASELINES=false
CI_MODE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            show_usage
            exit 0
            ;;
        --unit-only)
            UNIT_ONLY=true
            shift
            ;;
        --integration-only)
            INTEGRATION_ONLY=true
            shift
            ;;
        --e2e-only)
            E2E_ONLY=true
            shift
            ;;
        --performance-only)
            PERFORMANCE_ONLY=true
            shift
            ;;
        --update-baselines)
            UPDATE_BASELINES=true
            shift
            ;;
        --ci)
            CI_MODE=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    cd "$PROJECT_ROOT"

    log_header "SENTINEL TEST PERFORMANCE MONITORING"
    echo ""

    # Handle single test type runs
    if [ "$UNIT_ONLY" = "true" ]; then
        run_timed_tests "unit" "./tests/run_all_tests.sh" "$UNIT_TEST_TIMEOUT"
        exit $?
    fi

    if [ "$INTEGRATION_ONLY" = "true" ]; then
        run_timed_tests "integration" "find ./tests/integration -name '*.sh' -exec bash {} \;" "$INTEGRATION_TEST_TIMEOUT"
        exit $?
    fi

    if [ "$E2E_ONLY" = "true" ]; then
        run_timed_tests "e2e" "find ./tests/e2e -name '*.sh' -exec bash {} \;" "$E2E_TEST_TIMEOUT"
        exit $?
    fi

    if [ "$PERFORMANCE_ONLY" = "true" ]; then
        run_timed_tests "performance" "find ./tests/performance -name '*.sh' -exec bash {} \;" "$PERFORMANCE_TEST_TIMEOUT"
        exit $?
    fi

    # Run full monitoring session
    if run_all_tests_with_monitoring; then
        log_success "Performance monitoring completed successfully"
        exit 0
    else
        if [ "$CI_MODE" = "true" ]; then
            log_error "CI mode: Performance issues detected - failing build"
            exit 1
        else
            log_warning "Performance issues detected - review reports"
            exit 0
        fi
    fi
}

# Check for jq dependency
if ! command -v jq &> /dev/null; then
    log_warning "jq not found - JSON processing will be limited. Install jq for full functionality."
fi

# Run main function
main "$@"