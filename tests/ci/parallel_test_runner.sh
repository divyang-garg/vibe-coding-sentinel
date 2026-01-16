#!/bin/bash
# Parallel Test Execution Framework
# Optimizes CI/CD pipeline performance through parallel test execution
# Run from project root: ./tests/ci/parallel_test_runner.sh

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
CI_DIR="$PROJECT_ROOT/tests/ci"
REPORTS_DIR="$CI_DIR/reports"
PARALLEL_JOBS=${PARALLEL_JOBS:-$(nproc 2>/dev/null || echo 4)}  # Default to CPU count or 4
TIMEOUT_DEFAULT=600  # 10 minutes default timeout

# Test categories with their priorities and dependencies
declare -A TEST_PRIORITIES=(
    ["unit"]="HIGH"
    ["integration"]="MEDIUM"
    ["e2e"]="LOW"
    ["performance"]="LOW"
    ["security"]="HIGH"
    ["compatibility"]="MEDIUM"
)

declare -A TEST_DEPENDENCIES=(
    ["integration"]="unit"
    ["e2e"]="integration"
    ["performance"]=""
    ["security"]=""
    ["compatibility"]=""
)

# Create directories
mkdir -p "$REPORTS_DIR"

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

# Function to discover available test suites
discover_test_suites() {
    log_info "Discovering available test suites..."

    local test_suites=()

    # Unit tests
    if [ -f "./tests/run_all_tests.sh" ] || find ./tests/unit -name "*.sh" | grep -q .; then
        test_suites+=("unit:./tests/run_all_tests.sh")
    fi

    # Integration tests
    if find ./tests/integration -name "*.sh" | grep -q .; then
        test_suites+=("integration:find ./tests/integration -name '*.sh' -exec bash {} \;")
    fi

    # E2E tests
    if find ./tests/e2e -name "*.sh" | grep -q .; then
        test_suites+=("e2e:find ./tests/e2e -name '*.sh' -exec bash {} \;")
    fi

    # Performance tests
    if find ./tests/performance -name "*.sh" | grep -q .; then
        test_suites+=("performance:find ./tests/performance -name '*.sh' -exec bash {} \;")
    fi

    # Security tests
    if find ./tests/security -name "*.sh" | grep -q .; then
        test_suites+=("security:find ./tests/security -name '*.sh' -exec bash {} \;")
    fi

    # Go tests (if available)
    if find . -name "*_test.go" | grep -q .; then
        test_suites+=("go_unit:go test ./... -v")
        test_suites+=("go_integration:go test ./... -tags=integration -v")
        test_suites+=("go_benchmark:go test ./... -bench=. -benchmem")
    fi

    echo "${test_suites[@]}"
}

# Function to validate test dependencies
validate_dependencies() {
    local test_suites=("$@")
    local failed_deps=()

    for suite_info in "${test_suites[@]}"; do
        local suite_name=$(echo "$suite_info" | cut -d: -f1)
        local dependency=${TEST_DEPENDENCIES[$suite_name]}

        if [ -n "$dependency" ]; then
            # Check if dependency is in the test list
            local dep_found=false
            for check_suite in "${test_suites[@]}"; do
                local check_name=$(echo "$check_suite" | cut -d: -f1)
                if [ "$check_name" = "$dependency" ]; then
                    dep_found=true
                    break
                fi
            done

            if [ "$dep_found" = "false" ]; then
                failed_deps+=("$suite_name requires $dependency")
            fi
        fi
    done

    if [ ${#failed_deps[@]} -gt 0 ]; then
        log_error "Dependency validation failed:"
        for failed_dep in "${failed_deps[@]}"; do
            log_error "  • $failed_dep"
        done
        return 1
    fi

    return 0
}

# Function to run test suite with timeout and logging
run_test_suite() {
    local suite_name="$1"
    local suite_command="$2"
    local timeout="${3:-$TIMEOUT_DEFAULT}"
    local log_file="$REPORTS_DIR/${suite_name}_$(date '+%Y%m%d_%H%M%S').log"

    log_info "Starting $suite_name tests (timeout: ${timeout}s)..."

    local start_time=$(date +%s)

    # Run test with timeout
    if timeout "$timeout" bash -c "$suite_command" > "$log_file" 2>&1; then
        local exit_code=0
        local status="PASSED"
        log_success "$suite_name tests completed successfully"
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            local status="TIMEOUT"
            log_error "$suite_name tests timed out after ${timeout}s"
        else
            local status="FAILED"
            log_error "$suite_name tests failed with exit code $exit_code"
        fi
    fi

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Generate result JSON
    cat > "${log_file%.log}.json" << EOF
{
  "suite_name": "$suite_name",
  "status": "$status",
  "exit_code": $exit_code,
  "duration_seconds": $duration,
  "timeout_seconds": $timeout,
  "start_time": $(date -d "@$start_time" '+%s'),
  "end_time": $(date -d "@$end_time" '+%s'),
  "log_file": "$log_file",
  "command": "$suite_command"
}
EOF

    # Return exit code for parallel processing
    return $exit_code
}

# Function to execute tests in parallel with dependency management
run_parallel_tests() {
    local test_suites=("$@")
    local results=()
    local pids=()

    log_header "PARALLEL TEST EXECUTION"
    log_info "Running ${#test_suites[@]} test suites with $PARALLEL_JOBS parallel jobs"

    # Execute tests in parallel
    local running_jobs=0

    for suite_info in "${test_suites[@]}"; do
        local suite_name=$(echo "$suite_info" | cut -d: -f1)
        local suite_command=$(echo "$suite_info" | cut -d: -f2-)

        # Wait for available job slot
        while [ $running_jobs -ge $PARALLEL_JOBS ]; do
            # Check for completed jobs
            local completed_pid=""
            for pid_info in "${pids[@]}"; do
                local pid=$(echo "$pid_info" | cut -d: -f1)
                local name=$(echo "$pid_info" | cut -d: -f2)

                if ! kill -0 "$pid" 2>/dev/null; then
                    completed_pid="$pid"
                    wait "$pid" 2>/dev/null
                    local exit_code=$?

                    if [ $exit_code -eq 0 ]; then
                        results+=("$name:PASSED")
                    else
                        results+=("$name:FAILED")
                    fi
                    break
                fi
            done

            if [ -n "$completed_pid" ]; then
                # Remove completed job from tracking
                local new_pids=()
                for pid_info in "${pids[@]}"; do
                    if [ "$(echo "$pid_info" | cut -d: -f1)" != "$completed_pid" ]; then
                        new_pids+=("$pid_info")
                    fi
                done
                pids=("${new_pids[@]}")
                ((running_jobs--))
            else
                sleep 1
            fi
        done

        # Start new job
        run_test_suite "$suite_name" "$suite_command" &
        local pid=$!
        pids+=("$pid:$suite_name")
        ((running_jobs++))

        log_info "Started $suite_name (PID: $pid, running: $running_jobs/$PARALLEL_JOBS)"
    done

    # Wait for all remaining jobs to complete
    log_info "Waiting for remaining $running_jobs jobs to complete..."
    for pid_info in "${pids[@]}"; do
        local pid=$(echo "$pid_info" | cut -d: -f1)
        local name=$(echo "$pid_info" | cut -d: -f2)

        wait "$pid" 2>/dev/null
        local exit_code=$?

        if [ $exit_code -eq 0 ]; then
            results+=("$name:PASSED")
        else
            results+=("$name:FAILED")
        fi
    done

    # Generate execution summary
    generate_execution_summary "${results[@]}"
}

# Function to generate execution summary
generate_execution_summary() {
    local results=("$@")
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local summary_file="$REPORTS_DIR/parallel_execution_summary_$(date '+%Y%m%d_%H%M%S').json"

    # Count results
    local passed=0
    local failed=0
    local total_duration=0
    local test_results=()

    for result in "${results[@]}"; do
        local test_name=$(echo "$result" | cut -d: -f1)
        local status=$(echo "$result" | cut -d: -f2)

        # Find the JSON file for this test to get duration
        local json_file=$(find "$REPORTS_DIR" -name "${test_name}_*.json" | head -1)
        local duration=0

        if [ -f "$json_file" ]; then
            duration=$(jq -r '.duration_seconds' "$json_file" 2>/dev/null || echo "0")
            total_duration=$((total_duration + duration))
        fi

        test_results+=("{\"name\":\"$test_name\",\"status\":\"$status\",\"duration_seconds\":$duration}")

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
  "execution_mode": "parallel",
  "parallel_jobs": $PARALLEL_JOBS,
  "tests_run": $((${#results[@]})),
  "tests_passed": $passed,
  "tests_failed": $failed,
  "total_duration_seconds": $total_duration,
  "average_duration_seconds": $(echo "scale=2; $total_duration / ${#results[@]}" | bc 2>/dev/null || echo "0"),
  "overall_status": "$( [ $failed -eq 0 ] && echo "SUCCESS" || echo "FAILURE" )",
  "test_results": $test_results_json,
  "configuration": {
    "parallel_jobs": $PARALLEL_JOBS,
    "timeout_default_seconds": $TIMEOUT_DEFAULT,
    "project_root": "$PROJECT_ROOT"
  },
  "report_files": [
    "$REPORTS_DIR/*_*.log",
    "$REPORTS_DIR/*_*.json"
  ]
}
EOF

    # Display summary
    log_header "PARALLEL EXECUTION SUMMARY"
    echo -e "${CYAN}Parallel Jobs:${NC} $PARALLEL_JOBS"
    echo -e "${CYAN}Tests Run:${NC} ${#results[@]}"
    echo -e "${CYAN}Tests Passed:${NC} $passed"
    echo -e "${CYAN}Tests Failed:${NC} $failed"
    echo -e "${CYAN}Total Duration:${NC} ${total_duration}s"
    echo -e "${CYAN}Status:${NC} $( [ $failed -eq 0 ] && echo "✅ SUCCESS" || echo "❌ FAILURE" )"
    echo ""
    echo -e "${CYAN}Report saved:${NC} $summary_file"

    return $failed
}

# Function to run sequential tests (fallback mode)
run_sequential_tests() {
    local test_suites=("$@")
    local results=()
    local total_start=$(date +%s)

    log_header "SEQUENTIAL TEST EXECUTION"
    log_info "Running ${#test_suites[@]} test suites sequentially"

    for suite_info in "${test_suites[@]}"; do
        local suite_name=$(echo "$suite_info" | cut -d: -f1)
        local suite_command=$(echo "$suite_info" | cut -d: -f2-)

        if run_test_suite "$suite_name" "$suite_command"; then
            results+=("$suite_name:PASSED")
        else
            results+=("$suite_name:FAILED")
        fi
    done

    local total_end=$(date +%s)
    local total_duration=$((total_end - total_start))

    log_header "SEQUENTIAL EXECUTION SUMMARY"
    echo -e "${CYAN}Total Duration:${NC} ${total_duration}s"
    echo -e "${CYAN}Tests Run:${NC} ${#test_suites[@]}"

    # Count results
    local passed=0
    local failed=0
    for result in "${results[@]}"; do
        local status=$(echo "$result" | cut -d: -f2)
        if [ "$status" = "PASSED" ]; then
            ((passed++))
        else
            ((failed++))
        fi
    done

    echo -e "${CYAN}Tests Passed:${NC} $passed"
    echo -e "${CYAN}Tests Failed:${NC} $failed"

    return $failed
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Parallel Test Execution Framework for Sentinel"
    echo ""
    echo "OPTIONS:"
    echo "  --help              Show this help message"
    echo "  --jobs N            Number of parallel jobs (default: $PARALLEL_JOBS)"
    echo "  --sequential        Run tests sequentially instead of parallel"
    echo "  --unit-only         Run unit tests only"
    echo "  --integration-only  Run integration tests only"
    echo "  --e2e-only          Run E2E tests only"
    echo "  --timeout N         Default timeout in seconds (default: $TIMEOUT_DEFAULT)"
    echo "  --ci                CI/CD mode - strict error handling"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                           # Run all tests in parallel"
    echo "  $0 --jobs 8                 # Use 8 parallel jobs"
    echo "  $0 --sequential             # Run tests sequentially"
    echo "  $0 --unit-only              # Run only unit tests"
    echo "  $0 --ci                     # CI/CD mode"
    echo ""
    echo "TEST CATEGORIES:"
    echo "  • unit: Unit tests (highest priority)"
    echo "  • integration: Integration tests (depends on unit)"
    echo "  • e2e: End-to-end tests (depends on integration)"
    echo "  • performance: Performance benchmarks"
    echo "  • security: Security validation tests"
    echo "  • go_unit: Go unit tests"
    echo "  • go_integration: Go integration tests"
    echo "  • go_benchmark: Go benchmark tests"
    echo ""
    echo "REPORTS GENERATED:"
    echo "  • $REPORTS_DIR/*_*.log       - Individual test logs"
    echo "  • $REPORTS_DIR/*_*.json      - Individual test results"
    echo "  • $REPORTS_DIR/*_summary_*.json - Execution summary"
    echo ""
    echo "PERFORMANCE OPTIMIZATION:"
    echo "  • Parallel execution reduces CI/CD time"
    echo "  • Dependency validation prevents test conflicts"
    echo "  • Timeout protection prevents hanging tests"
    echo "  • Resource utilization monitoring"
}

# Parse command line arguments
SEQUENTIAL=false
UNIT_ONLY=false
INTEGRATION_ONLY=false
E2E_ONLY=false
CI_MODE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            show_usage
            exit 0
            ;;
        --jobs)
            PARALLEL_JOBS="$2"
            shift 2
            ;;
        --sequential)
            SEQUENTIAL=true
            shift
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
        --timeout)
            TIMEOUT_DEFAULT="$2"
            shift 2
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

    log_header "SENTINEL PARALLEL TEST RUNNER"
    log_info "CODING_STANDARDS.md Compliance: Parallel execution framework"
    echo ""

    # Discover available test suites
    local test_suites_raw
    IFS=' ' read -r -a test_suites_raw <<< "$(discover_test_suites)"

    if [ ${#test_suites_raw[@]} -eq 0 ]; then
        log_error "No test suites found"
        exit 1
    fi

    log_success "Discovered ${#test_suites_raw[@]} test suites"

    # Filter test suites based on options
    local test_suites=()

    for suite_info in "${test_suites_raw[@]}"; do
        local suite_name=$(echo "$suite_info" | cut -d: -f1)

        if [ "$UNIT_ONLY" = "true" ] && [[ "$suite_name" != unit* ]]; then
            continue
        fi

        if [ "$INTEGRATION_ONLY" = "true" ] && [[ "$suite_name" != integration* ]]; then
            continue
        fi

        if [ "$E2E_ONLY" = "true" ] && [[ "$suite_name" != e2e* ]]; then
            continue
        fi

        test_suites+=("$suite_info")
    done

    if [ ${#test_suites[@]} -eq 0 ]; then
        log_error "No test suites match the specified filters"
        exit 1
    fi

    log_info "Selected ${#test_suites[@]} test suites for execution"

    # Validate dependencies
    if ! validate_dependencies "${test_suites[@]}"; then
        log_error "Dependency validation failed"
        exit 1
    fi

    # Execute tests
    local exit_code
    if [ "$SEQUENTIAL" = "true" ]; then
        if run_sequential_tests "${test_suites[@]}"; then
            exit_code=0
        else
            exit_code=1
        fi
    else
        if run_parallel_tests "${test_suites[@]}"; then
            exit_code=0
        else
            exit_code=1
        fi
    fi

    # Final status
    if [ $exit_code -eq 0 ]; then
        log_success "All tests completed successfully"
    else
        log_error "Some tests failed"
        if [ "$CI_MODE" = "true" ]; then
            log_error "CI mode: Test failures detected - failing build"
        fi
    fi

    exit $exit_code
}

# Run main function
main "$@"