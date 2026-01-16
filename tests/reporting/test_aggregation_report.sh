#!/bin/bash
# Unified Test Reporting and Aggregation
# Aggregates results from all test suites into comprehensive reports
# Run from project root: ./tests/reporting/test_aggregation_report.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REPORTS_DIR="$PROJECT_ROOT/tests/reporting"
AGGREGATE_DIR="$REPORTS_DIR/aggregated"
HISTORY_DIR="$REPORTS_DIR/history"

# CODING_STANDARDS.md thresholds
THRESHOLD_COVERAGE_OVERALL=80
THRESHOLD_COVERAGE_CRITICAL=90
THRESHOLD_TEST_SUCCESS=95

# Create directories
mkdir -p "$AGGREGATE_DIR" "$HISTORY_DIR"

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

# Function to discover all test reports
discover_reports() {
    local report_files=()

    # Coverage reports
    while IFS= read -r -d '' file; do
        report_files+=("$file")
    done < <(find "$PROJECT_ROOT/tests/coverage/reports" -name "*.json" -print0 2>/dev/null)

    # Performance reports
    while IFS= read -r -d '' file; do
        report_files+=("$file")
    done < <(find "$PROJECT_ROOT/tests/monitoring/reports" -name "*.json" -print0 2>/dev/null)

    # CI execution reports
    while IFS= read -r -d '' file; do
        report_files+=("$file")
    done < <(find "$PROJECT_ROOT/tests/ci/reports" -name "*.json" -print0 2>/dev/null)

    echo "${report_files[@]}"
}

# Function to aggregate coverage data
aggregate_coverage_data() {
    log_info "Aggregating coverage data..."

    local coverage_files=()
    while IFS= read -r -d '' file; do
        coverage_files+=("$file")
    done < <(find "$PROJECT_ROOT/tests/coverage/reports" -name "*coverage_summary*.json" -print0 2>/dev/null)

    if [ ${#coverage_files[@]} -eq 0 ]; then
        log_warning "No coverage reports found"
        return 1
    fi

    # Get latest coverage report
    local latest_coverage=""
    local latest_timestamp=0

    for file in "${coverage_files[@]}"; do
        local timestamp=$(jq -r '.timestamp' "$file" 2>/dev/null || echo "")
        if [ -n "$timestamp" ]; then
            local file_timestamp=$(date -d "$timestamp" +%s 2>/dev/null || echo "0")
            if [ "$file_timestamp" -gt "$latest_timestamp" ]; then
                latest_timestamp=$file_timestamp
                latest_coverage="$file"
            fi
        fi
    done

    if [ -n "$latest_coverage" ]; then
        log_success "Latest coverage report: $latest_coverage"

        # Extract coverage metrics
        local overall=$(jq -r '.overall_coverage' "$latest_coverage" 2>/dev/null || echo "0")
        local critical=$(jq -r '.critical_coverage' "$latest_coverage" 2>/dev/null || echo "0")
        local overall_pass=$(jq -r '.compliance.overall_pass' "$latest_coverage" 2>/dev/null || echo "false")
        local critical_pass=$(jq -r '.compliance.critical_pass' "$latest_coverage" 2>/dev/null || echo "false")

        cat << EOF
{
  "coverage_overall_percent": $overall,
  "coverage_critical_percent": $critical,
  "coverage_overall_pass": $overall_pass,
  "coverage_critical_pass": $critical_pass,
  "coverage_report": "$latest_coverage"
}
EOF
    else
        log_warning "No valid coverage reports found"
        echo "{}"
    fi
}

# Function to aggregate performance data
aggregate_performance_data() {
    log_info "Aggregating performance data..."

    local performance_files=()
    while IFS= read -r -d '' file; do
        performance_files+=("$file")
    done < <(find "$PROJECT_ROOT/tests/monitoring/reports" -name "*performance_summary*.json" -print0 2>/dev/null)

    if [ ${#performance_files[@]} -eq 0 ]; then
        log_warning "No performance reports found"
        return 1
    fi

    # Get latest performance report
    local latest_performance=""
    local latest_timestamp=0

    for file in "${performance_files[@]}"; do
        local timestamp=$(jq -r '.timestamp' "$file" 2>/dev/null || echo "")
        if [ -n "$timestamp" ]; then
            local file_timestamp=$(date -d "$timestamp" +%s 2>/dev/null || echo "0")
            if [ "$file_timestamp" -gt "$latest_timestamp" ]; then
                latest_timestamp=$file_timestamp
                latest_performance="$file"
            fi
        fi
    done

    if [ -n "$latest_performance" ]; then
        log_success "Latest performance report: $latest_performance"

        # Extract performance metrics
        local total_duration=$(jq -r '.session_duration_seconds' "$latest_performance" 2>/dev/null || echo "0")
        local tests_passed=$(jq -r '.tests_passed' "$latest_performance" 2>/dev/null || echo "0")
        local tests_failed=$(jq -r '.tests_failed' "$latest_performance" 2>/dev/null || echo "0")
        local tests_run=$(jq -r '.tests_run' "$latest_performance" 2>/dev/null || echo "0")

        # Calculate success rate
        local success_rate=0
        if [ "$tests_run" -gt 0 ]; then
            success_rate=$(echo "scale=2; ($tests_passed / $tests_run) * 100" | bc 2>/dev/null || echo "0")
        fi

        cat << EOF
{
  "performance_session_duration_seconds": $total_duration,
  "performance_tests_run": $tests_run,
  "performance_tests_passed": $tests_passed,
  "performance_tests_failed": $tests_failed,
  "performance_success_rate_percent": $success_rate,
  "performance_report": "$latest_performance"
}
EOF
    else
        log_warning "No valid performance reports found"
        echo "{}"
    fi
}

# Function to aggregate CI execution data
aggregate_ci_execution_data() {
    log_info "Aggregating CI execution data..."

    local ci_files=()
    while IFS= read -r -d '' file; do
        ci_files+=("$file")
    done < <(find "$PROJECT_ROOT/tests/ci/reports" -name "*execution_summary*.json" -print0 2>/dev/null)

    if [ ${#ci_files[@]} -eq 0 ]; then
        log_warning "No CI execution reports found"
        return 1
    fi

    # Get latest CI report
    local latest_ci=""
    local latest_timestamp=0

    for file in "${ci_files[@]}"; do
        local timestamp=$(jq -r '.timestamp' "$file" 2>/dev/null || echo "")
        if [ -n "$timestamp" ]; then
            local file_timestamp=$(date -d "$timestamp" +%s 2>/dev/null || echo "0")
            if [ "$file_timestamp" -gt "$latest_timestamp" ]; then
                latest_timestamp=$file_timestamp
                latest_ci="$file"
            fi
        fi
    done

    if [ -n "$latest_ci" ]; then
        log_success "Latest CI report: $latest_ci"

        # Extract CI metrics
        local total_duration=$(jq -r '.total_duration_seconds' "$latest_ci" 2>/dev/null || echo "0")
        local tests_run=$(jq -r '.tests_run' "$latest_ci" 2>/dev/null || echo "0")
        local tests_passed=$(jq -r '.tests_passed' "$latest_ci" 2>/dev/null || echo "0")
        local tests_failed=$(jq -r '.tests_failed' "$latest_ci" 2>/dev/null || echo "0")
        local parallel_jobs=$(jq -r '.configuration.parallel_jobs' "$latest_ci" 2>/dev/null || echo "1")

        # Calculate metrics
        local success_rate=0
        local avg_duration=0
        if [ "$tests_run" -gt 0 ]; then
            success_rate=$(echo "scale=2; ($tests_passed / $tests_run) * 100" | bc 2>/dev/null || echo "0")
            avg_duration=$(echo "scale=2; $total_duration / $tests_run" | bc 2>/dev/null || echo "0")
        fi

        cat << EOF
{
  "ci_total_duration_seconds": $total_duration,
  "ci_tests_run": $tests_run,
  "ci_tests_passed": $tests_passed,
  "ci_tests_failed": $tests_failed,
  "ci_success_rate_percent": $success_rate,
  "ci_average_duration_seconds": $avg_duration,
  "ci_parallel_jobs": $parallel_jobs,
  "ci_report": "$latest_ci"
}
EOF
    else
        log_warning "No valid CI execution reports found"
        echo "{}"
    fi
}

# Function to generate comprehensive aggregated report
generate_aggregated_report() {
    log_header "GENERATING AGGREGATED TEST REPORT"

    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local report_file="$AGGREGATE_DIR/comprehensive_test_report_$(date '+%Y%m%d_%H%M%S').json"

    # Aggregate data from different sources
    local coverage_data=$(aggregate_coverage_data)
    local performance_data=$(aggregate_performance_data)
    local ci_data=$(aggregate_ci_execution_data)

    # Extract individual metrics
    local coverage_overall=$(echo "$coverage_data" | jq -r '.coverage_overall_percent' 2>/dev/null || echo "0")
    local coverage_critical=$(echo "$coverage_data" | jq -r '.coverage_critical_percent' 2>/dev/null || echo "0")
    local coverage_overall_pass=$(echo "$coverage_data" | jq -r '.coverage_overall_pass' 2>/dev/null || echo "false")
    local coverage_critical_pass=$(echo "$coverage_data" | jq -r '.coverage_critical_pass' 2>/dev/null || echo "false")

    local perf_duration=$(echo "$performance_data" | jq -r '.performance_session_duration_seconds' 2>/dev/null || echo "0")
    local perf_tests_run=$(echo "$performance_data" | jq -r '.performance_tests_run' 2>/dev/null || echo "0")
    local perf_success_rate=$(echo "$performance_data" | jq -r '.performance_success_rate_percent' 2>/dev/null || echo "0")

    local ci_duration=$(echo "$ci_data" | jq -r '.ci_total_duration_seconds' 2>/dev/null || echo "0")
    local ci_tests_run=$(echo "$ci_data" | jq -r '.ci_tests_run' 2>/dev/null || echo "0")
    local ci_success_rate=$(echo "$ci_data" | jq -r '.ci_success_rate_percent' 2>/dev/null || echo "0")
    local ci_parallel_jobs=$(echo "$ci_data" | jq -r '.ci_parallel_jobs' 2>/dev/null || echo "1")

    # Calculate overall metrics
    local total_tests_run=$((perf_tests_run + ci_tests_run))
    local total_duration=$((perf_duration + ci_duration))

    local overall_success_rate=0
    if [ "$total_tests_run" -gt 0 ]; then
        # Weighted average of success rates
        local perf_weighted=$((perf_tests_run * perf_success_rate / 100))
        local ci_weighted=$((ci_tests_run * ci_success_rate / 100))
        overall_success_rate=$(echo "scale=2; (($perf_weighted + $ci_weighted) / $total_tests_run) * 100" | bc 2>/dev/null || echo "0")
    fi

    # Determine overall compliance status
    local overall_compliant="false"
    if [ "$coverage_overall_pass" = "true" ] && [ "$coverage_critical_pass" = "true" ] && (( $(echo "$overall_success_rate >= $THRESHOLD_TEST_SUCCESS" | bc -l) )); then
        overall_compliant="true"
    fi

    # Generate comprehensive JSON report
    cat > "$report_file" << EOF
{
  "timestamp": "$timestamp",
  "report_type": "comprehensive_test_aggregation",
  "coding_standards_compliance": {
    "overall_compliant": $overall_compliant,
    "coverage_overall_threshold": $THRESHOLD_COVERAGE_OVERALL,
    "coverage_critical_threshold": $THRESHOLD_COVERAGE_CRITICAL,
    "test_success_threshold": $THRESHOLD_TEST_SUCCESS
  },
  "coverage_metrics": {
    "overall_percent": $coverage_overall,
    "critical_percent": $coverage_critical,
    "overall_pass": $coverage_overall_pass,
    "critical_pass": $coverage_critical_pass
  },
  "performance_metrics": {
    "session_duration_seconds": $perf_duration,
    "tests_run": $perf_tests_run,
    "success_rate_percent": $perf_success_rate
  },
  "ci_execution_metrics": {
    "total_duration_seconds": $ci_duration,
    "tests_run": $ci_tests_run,
    "success_rate_percent": $ci_success_rate,
    "parallel_jobs": $ci_parallel_jobs,
    "average_duration_seconds": $(echo "$ci_data" | jq -r '.ci_average_duration_seconds' 2>/dev/null || echo "0")
  },
  "aggregated_metrics": {
    "total_tests_run": $total_tests_run,
    "total_duration_seconds": $total_duration,
    "overall_success_rate_percent": $overall_success_rate,
    "efficiency_score": $([ "$ci_parallel_jobs" -gt 1 ] && echo "scale=2; $total_duration / $ci_parallel_jobs" | bc 2>/dev/null || echo "0")
  },
  "recommendations": $(generate_recommendations "$coverage_overall_pass" "$coverage_critical_pass" "$overall_success_rate"),
  "source_reports": {
    "coverage": $(echo "$coverage_data" | jq -r '.coverage_report' 2>/dev/null || echo "null"),
    "performance": $(echo "$performance_data" | jq -r '.performance_report' 2>/dev/null || echo "null"),
    "ci_execution": $(echo "$ci_data" | jq -r '.ci_report' 2>/dev/null || echo "null")
  },
  "metadata": {
    "generated_by": "test_aggregation_report.sh",
    "project_root": "$PROJECT_ROOT",
    "test_suites_discovered": $(discover_reports | wc -w),
    "report_version": "1.0"
  }
}
EOF

    # Display comprehensive summary
    display_comprehensive_summary "$report_file" "$overall_compliant" "$overall_success_rate" "$coverage_overall" "$coverage_critical"

    log_success "Comprehensive aggregated report generated: $report_file"
}

# Function to generate recommendations
generate_recommendations() {
    local coverage_overall_pass="$1"
    local coverage_critical_pass="$2"
    local overall_success_rate="$3"

    local recommendations=()

    if [ "$coverage_overall_pass" != "true" ]; then
        recommendations+=("Increase overall code coverage - focus on untested functions")
    fi

    if [ "$coverage_critical_pass" != "true" ]; then
        recommendations+=("Improve critical path coverage - handlers, services, repository need more tests")
    fi

    if (( $(echo "$overall_success_rate < $THRESHOLD_TEST_SUCCESS" | bc -l) )); then
        recommendations+=("Improve test reliability - address flaky tests and failures")
    fi

    if [ ${#recommendations[@]} -eq 0 ]; then
        recommendations+=("All metrics meeting standards - maintain current quality levels")
    fi

    # Format as JSON array
    local json_array="["
    for i in "${!recommendations[@]}"; do
        if [ $i -gt 0 ]; then
            json_array+=","
        fi
        json_array+="\"${recommendations[$i]}\""
    done
    json_array+="]"

    echo "$json_array"
}

# Function to display comprehensive summary
display_comprehensive_summary() {
    local report_file="$1"
    local overall_compliant="$2"
    local overall_success_rate="$3"
    local coverage_overall="$4"
    local coverage_critical="$5"

    log_header "COMPREHENSIVE TEST SUMMARY"

    echo -e "${WHITE}CODING STANDARDS COMPLIANCE${NC}"
    if [ "$overall_compliant" = "true" ]; then
        echo -e "${GREEN}✅ OVERALL STATUS: COMPLIANT${NC}"
    else
        echo -e "${RED}❌ OVERALL STATUS: NON-COMPLIANT${NC}"
    fi
    echo ""

    echo -e "${CYAN}Coverage Metrics:${NC}"
    echo -e "  • Overall: ${coverage_overall}% (Threshold: ${THRESHOLD_COVERAGE_OVERALL}%)"
    echo -e "  • Critical: ${coverage_critical}% (Threshold: ${THRESHOLD_COVERAGE_CRITICAL}%)"
    echo ""

    echo -e "${CYAN}Test Success Rate:${NC} ${overall_success_rate}% (Threshold: ${THRESHOLD_TEST_SUCCESS}%)"
    echo ""

    echo -e "${CYAN}Report Details:${NC}"
    echo -e "  • Generated: $(date)"
    echo -e "  • Location: $report_file"
    echo ""

    # Show recommendations
    local recommendations=$(jq -r '.recommendations[]' "$report_file" 2>/dev/null)
    if [ -n "$recommendations" ]; then
        echo -e "${YELLOW}Recommendations:${NC}"
        echo "$recommendations" | while read -r rec; do
            echo -e "  • $rec"
        done
    fi
}

# Function to archive historical reports
archive_reports() {
    log_info "Archiving historical reports..."

    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local archive_file="$HISTORY_DIR/test_reports_archive_$timestamp.tar.gz"

    # Find all report files from the last 30 days
    local report_files=()
    while IFS= read -r -d '' file; do
        # Only archive files older than 7 days
        if [ $(find "$file" -mtime +7 2>/dev/null | wc -l) -gt 0 ]; then
            report_files+=("$file")
        fi
    done < <(find "$PROJECT_ROOT/tests" -name "*.json" -o -name "*.log" -o -name "*.html" | grep -E "(coverage|monitoring|ci|reporting)" | head -100 | tr '\n' '\0')

    if [ ${#report_files[@]} -gt 0 ]; then
        # Create archive (relative paths)
        cd "$PROJECT_ROOT"
        tar -czf "$archive_file" --files-from <(printf '%s\n' "${report_files[@]}" | sed "s|$PROJECT_ROOT/||g") 2>/dev/null || true
        log_success "Archived ${#report_files[@]} historical reports: $archive_file"

        # Clean up archived files (keep last 7 days)
        for file in "${report_files[@]}"; do
            rm -f "$file"
        done
        log_info "Cleaned up ${#report_files[@]} old report files"
    else
        log_info "No historical reports to archive"
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Unified Test Reporting and Aggregation for Sentinel"
    echo ""
    echo "OPTIONS:"
    echo "  --help              Show this help message"
    echo "  --archive           Archive historical reports (>7 days old)"
    echo "  --no-display        Skip console output, only generate reports"
    echo "  --ci                CI/CD mode - exit with error code on failures"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                           # Generate comprehensive aggregated report"
    echo "  $0 --archive                # Generate report and archive old files"
    echo "  $0 --no-display             # Background report generation"
    echo "  $0 --ci                     # CI/CD mode with strict validation"
    echo ""
    echo "REPORTS GENERATED:"
    echo "  • $AGGREGATE_DIR/comprehensive_test_report_*.json - Main aggregated report"
    echo "  • $HISTORY_DIR/test_reports_archive_*.tar.gz      - Historical archives"
    echo ""
    echo "AGGREGATED METRICS:"
    echo "  • Coverage: Overall and critical path percentages"
    echo "  • Performance: Test execution times and success rates"
    echo "  • CI Execution: Parallel execution efficiency"
    echo "  • Compliance: CODING_STANDARDS.md validation"
    echo ""
    echo "CODING STANDARDS COMPLIANCE:"
    echo "  • Coverage: ≥${THRESHOLD_COVERAGE_OVERALL}% overall, ≥${THRESHOLD_COVERAGE_CRITICAL}% critical"
    echo "  • Test Success: ≥${THRESHOLD_TEST_SUCCESS}% pass rate"
    echo "  • Automated reporting and alerting"
}

# Parse command line arguments
ARCHIVE=false
NO_DISPLAY=false
CI_MODE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            show_usage
            exit 0
            ;;
        --archive)
            ARCHIVE=true
            shift
            ;;
        --no-display)
            NO_DISPLAY=true
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

    log_header "SENTINEL UNIFIED TEST REPORTING"
    log_info "Aggregating results from all test suites..."
    echo ""

    # Generate comprehensive report
    generate_aggregated_report

    # Archive if requested
    if [ "$ARCHIVE" = "true" ]; then
        echo ""
        archive_reports
    fi

    # Check compliance for CI mode
    if [ "$CI_MODE" = "true" ]; then
        local latest_report=$(find "$AGGREGATE_DIR" -name "comprehensive_test_report_*.json" | sort | tail -1)
        if [ -f "$latest_report" ]; then
            local compliant=$(jq -r '.coding_standards_compliance.overall_compliant' "$latest_report" 2>/dev/null || echo "false")
            if [ "$compliant" != "true" ]; then
                log_error "CI mode: CODING_STANDARDS.md compliance check failed"
                exit 1
            fi
        fi
    fi

    log_success "Test aggregation and reporting completed successfully"
}

# Check for jq dependency
if ! command -v jq &> /dev/null && [ "$NO_DISPLAY" = "false" ]; then
    log_warning "jq not found - JSON processing limited. Install jq for full functionality."
fi

# Run main function
main "$@"