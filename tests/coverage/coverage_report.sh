#!/bin/bash
# Automated Test Coverage Reporting
# Generates comprehensive coverage reports with CODING_STANDARDS.md compliance
# Run from project root: ./tests/coverage/coverage_report.sh

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
COVERAGE_DIR="$PROJECT_ROOT/tests/coverage"
REPORTS_DIR="$COVERAGE_DIR/reports"
THRESHOLD_OVERALL=80
THRESHOLD_CRITICAL=90
MINIMUM_LINES=1000

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

# Function to check if Go is available
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    log_success "Go $(go version) found"
}

# Function to run tests with coverage
run_coverage_tests() {
    log_info "Running tests with coverage analysis..."

    # Clean previous coverage files
    rm -f *.out coverage.html

    # Run unit tests with coverage
    log_info "Running unit tests..."
    if go test ./... -coverprofile=coverage_unit.out -covermode=atomic -v | tee test_output.log; then
        log_success "Unit tests completed"
    else
        log_error "Unit tests failed"
        return 1
    fi

    # Run integration tests (if they exist)
    if [ -d "tests/integration" ]; then
        log_info "Running integration tests..."
        # Note: Integration tests would need to be converted to Go tests or use a different approach
        log_warning "Integration test coverage not yet implemented"
    fi

    return 0
}

# Function to generate coverage reports
generate_reports() {
    log_info "Generating coverage reports..."

    # Merge coverage profiles if multiple exist
    if [ -f coverage_unit.out ]; then
        cp coverage_unit.out coverage.out
        log_success "Coverage profile ready"
    else
        log_error "No coverage profile found"
        return 1
    fi

    # Generate HTML report
    go tool cover -html=coverage.out -o "$REPORTS_DIR/coverage.html"
    log_success "HTML coverage report generated: $REPORTS_DIR/coverage.html"

    # Generate text report
    go tool cover -func=coverage.out > "$REPORTS_DIR/coverage.txt"
    log_success "Text coverage report generated: $REPORTS_DIR/coverage.txt"
}

# Function to analyze coverage results
analyze_coverage() {
    log_header "COVERAGE ANALYSIS REPORT"

    if [ ! -f "$REPORTS_DIR/coverage.txt" ]; then
        log_error "Coverage report not found"
        return 1
    fi

    # Extract overall coverage percentage
    OVERALL_COVERAGE=$(grep "total:" "$REPORTS_DIR/coverage.txt" | awk '{print $3}' | sed 's/%//')

    if [ -z "$OVERALL_COVERAGE" ]; then
        log_error "Could not extract overall coverage percentage"
        return 1
    fi

    echo -e "${CYAN}Overall Coverage: ${OVERALL_COVERAGE}%${NC}"

    # Check against CODING_STANDARDS.md thresholds
    if (( $(echo "$OVERALL_COVERAGE >= $THRESHOLD_OVERALL" | bc -l) )); then
        log_success "Overall coverage meets minimum threshold (${THRESHOLD_OVERALL}%)"
        OVERALL_PASS=true
    else
        log_error "Overall coverage below minimum threshold (${THRESHOLD_OVERALL}%)"
        OVERALL_PASS=false
    fi

    # Analyze critical path coverage (business logic, handlers, services)
    log_info "Analyzing critical path coverage..."

    CRITICAL_COVERAGE=$(grep -E "(handlers|services|repository|models)" "$REPORTS_DIR/coverage.txt" | awk '{sum+=$3; count++} END {if(count>0) print sum/count; else print 0}')

    if (( $(echo "$CRITICAL_COVERAGE >= $THRESHOLD_CRITICAL" | bc -l) )); then
        log_success "Critical path coverage meets threshold (${THRESHOLD_CRITICAL}%): ${CRITICAL_COVERAGE}%"
        CRITICAL_PASS=true
    else
        log_warning "Critical path coverage below threshold (${THRESHOLD_CRITICAL}%): ${CRITICAL_COVERAGE}%"
        CRITICAL_PASS=false
    fi

    # Check for uncovered critical functions
    log_info "Checking for uncovered critical functions..."
    grep -E "(handlers|services|repository)" "$REPORTS_DIR/coverage.txt" | grep "0.0%" | head -5 | while read -r line; do
        log_warning "Uncovered critical function: $line"
    done

    # Generate summary report
    generate_summary_report "$OVERALL_COVERAGE" "$CRITICAL_COVERAGE" "$OVERALL_PASS" "$CRITICAL_PASS"
}

# Function to generate summary report
generate_summary_report() {
    local overall=$1
    local critical=$2
    local overall_pass=$3
    local critical_pass=$4

    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local report_file="$REPORTS_DIR/coverage_summary_$(date '+%Y%m%d_%H%M%S').json"

    cat > "$report_file" << EOF
{
  "timestamp": "$timestamp",
  "overall_coverage": $overall,
  "critical_coverage": $critical,
  "thresholds": {
    "overall_minimum": $THRESHOLD_OVERALL,
    "critical_minimum": $THRESHOLD_CRITICAL
  },
  "compliance": {
    "overall_pass": $overall_pass,
    "critical_pass": $critical_pass,
    "coding_standards_compliant": $([ "$overall_pass" = "true" ] && [ "$critical_pass" = "true" ] && echo "true" || echo "false")
  },
  "files_analyzed": $(find . -name "*.go" -not -path "./vendor/*" | wc -l),
  "test_files": $(find . -name "*_test.go" | wc -l),
  "report_files": [
    "$REPORTS_DIR/coverage.html",
    "$REPORTS_DIR/coverage.txt",
    "$report_file"
  ]
}
EOF

    log_success "Summary report generated: $report_file"

    # Display results
    echo ""
    log_header "COVERAGE SUMMARY"
    echo -e "${CYAN}Overall Coverage:${NC} $overall% (Threshold: ${THRESHOLD_OVERALL}%)"
    echo -e "${CYAN}Critical Path Coverage:${NC} $critical% (Threshold: ${THRESHOLD_CRITICAL}%)"

    if [ "$overall_pass" = "true" ] && [ "$critical_pass" = "true" ]; then
        log_success "🎉 ALL COVERAGE THRESHOLDS MET - CODING_STANDARDS.md COMPLIANT"
        return 0
    else
        log_error "❌ COVERAGE THRESHOLDS NOT MET"
        echo ""
        log_info "Recommendations:"
        echo "  • Focus on testing critical business logic (handlers, services, repository)"
        echo "  • Add unit tests for uncovered functions"
        echo "  • Consider integration tests for complex workflows"
        echo "  • Review test exclusions and ensure critical code is covered"
        return 1
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Automated Test Coverage Reporting for Sentinel"
    echo ""
    echo "OPTIONS:"
    echo "  --help          Show this help message"
    echo "  --threshold N   Set overall coverage threshold (default: $THRESHOLD_OVERALL)"
    echo "  --critical N    Set critical path threshold (default: $THRESHOLD_CRITICAL)"
    echo "  --html-only     Generate HTML report only (skip analysis)"
    echo "  --ci            CI/CD mode - exit with error code on threshold failure"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                           # Run full coverage analysis"
    echo "  $0 --threshold 85           # Use 85% overall threshold"
    echo "  $0 --html-only              # Generate HTML report only"
    echo "  $0 --ci                     # CI/CD mode with strict checking"
    echo ""
    echo "REPORTS GENERATED:"
    echo "  • $REPORTS_DIR/coverage.html    - Interactive HTML coverage report"
    echo "  • $REPORTS_DIR/coverage.txt     - Text-based coverage data"
    echo "  • $REPORTS_DIR/coverage_summary_TIMESTAMP.json - JSON summary"
    echo ""
    echo "CODING_STANDARDS.md REQUIREMENTS:"
    echo "  • Overall Coverage: ≥${THRESHOLD_OVERALL}%"
    echo "  • Critical Path: ≥${THRESHOLD_CRITICAL}% (handlers, services, repository, models)"
}

# Parse command line arguments
CI_MODE=false
HTML_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            show_usage
            exit 0
            ;;
        --threshold)
            THRESHOLD_OVERALL="$2"
            shift 2
            ;;
        --critical)
            THRESHOLD_CRITICAL="$2"
            shift 2
            ;;
        --html-only)
            HTML_ONLY=true
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

    log_header "SENTINEL TEST COVERAGE ANALYSIS"
    log_info "CODING_STANDARDS.md Compliance: Coverage ≥${THRESHOLD_OVERALL}%, Critical ≥${THRESHOLD_CRITICAL}%"
    echo ""

    # Pre-flight checks
    check_go

    # Run coverage tests
    if [ "$HTML_ONLY" = "false" ]; then
        if ! run_coverage_tests; then
            log_error "Test execution failed"
            exit 1
        fi
    else
        log_info "HTML-only mode: skipping test execution"
    fi

    # Generate reports
    if ! generate_reports; then
        log_error "Report generation failed"
        exit 1
    fi

    # Analyze coverage (unless HTML-only)
    if [ "$HTML_ONLY" = "false" ]; then
        if analyze_coverage; then
            log_success "Coverage analysis completed successfully"
            exit 0
        else
            if [ "$CI_MODE" = "true" ]; then
                log_error "CI mode: Coverage thresholds not met - failing build"
                exit 1
            else
                log_warning "Coverage thresholds not met - review recommendations above"
                exit 0
            fi
        fi
    else
        log_success "HTML coverage report generated successfully"
        exit 0
    fi
}

# Run main function
main "$@"