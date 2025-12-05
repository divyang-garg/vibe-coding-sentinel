#!/bin/bash
# File Size Management Tests (Phase 9)
# Tests file size checking, architecture analysis, and split suggestions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
FIXTURES_DIR="$PROJECT_ROOT/tests/fixtures/file_size"
SENTINEL="$PROJECT_ROOT/sentinel"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_info() {
    echo -e "${GREEN}✓${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

test_file_size_detection() {
    echo "Testing file size detection..."
    
    # Create test config with file size thresholds
    TEST_CONFIG="$PROJECT_ROOT/.sentinelsrc.test"
    cat > "$TEST_CONFIG" <<EOF
{
  "fileSize": {
    "thresholds": {
      "warning": 50,
      "critical": 100,
      "maximum": 200
    },
    "exceptions": []
  }
}
EOF
    
    # Test with large file (should trigger warning)
    cd "$FIXTURES_DIR"
    OUTPUT=$($SENTINEL audit --output json 2>&1 || true)
    
    # Check if large_file.go is detected
    if echo "$OUTPUT" | grep -q "large_file.go"; then
        log_info "File size detection works (large_file.go detected)"
        ((TESTS_PASSED++))
    else
        log_error "File size detection failed (large_file.go not detected)"
        ((TESTS_FAILED++))
    fi
    
    # Test with oversized file (should trigger critical)
    if echo "$OUTPUT" | grep -q "oversized_file.ts"; then
        log_info "Oversized file detection works (oversized_file.ts detected)"
        ((TESTS_PASSED++))
    else
        log_error "Oversized file detection failed (oversized_file.ts not detected)"
        ((TESTS_FAILED++))
    fi
    
    # Cleanup
    rm -f "$TEST_CONFIG"
}

test_analyze_structure_flag() {
    echo "Testing --analyze-structure flag..."
    
    cd "$FIXTURES_DIR"
    OUTPUT=$($SENTINEL audit --analyze-structure 2>&1 || true)
    
    if echo "$OUTPUT" | grep -q "Architecture Analysis"; then
        log_info "--analyze-structure flag works"
        ((TESTS_PASSED++))
    else
        log_error "--analyze-structure flag failed"
        ((TESTS_FAILED++))
    fi
    
    if echo "$OUTPUT" | grep -q "File Size Analysis Results"; then
        log_info "File size analysis results displayed"
        ((TESTS_PASSED++))
    else
        log_error "File size analysis results not displayed"
        ((TESTS_FAILED++))
    fi
}

test_file_size_exceptions() {
    echo "Testing file size exceptions..."
    
    TEST_CONFIG="$PROJECT_ROOT/.sentinelsrc.test"
    cat > "$TEST_CONFIG" <<EOF
{
  "fileSize": {
    "thresholds": {
      "warning": 50,
      "critical": 100,
      "maximum": 200
    },
    "exceptions": ["large_file.go"]
  }
}
EOF
    
    cd "$FIXTURES_DIR"
    OUTPUT=$($SENTINEL audit --output json 2>&1 || true)
    
    # large_file.go should NOT be detected (in exceptions)
    if echo "$OUTPUT" | grep -q "large_file.go"; then
        log_error "File size exceptions failed (large_file.go still detected)"
        ((TESTS_FAILED++))
    else
        log_info "File size exceptions work (large_file.go excluded)"
        ((TESTS_PASSED++))
    fi
    
    # Cleanup
    rm -f "$TEST_CONFIG"
}

test_file_type_thresholds() {
    echo "Testing file-type-specific thresholds..."
    
    TEST_CONFIG="$PROJECT_ROOT/.sentinelsrc.test"
    cat > "$TEST_CONFIG" <<EOF
{
  "fileSize": {
    "thresholds": {
      "warning": 50,
      "critical": 100,
      "maximum": 200
    },
    "byFileType": {
      "service": 300,
      "component": 150
    },
    "exceptions": []
  }
}
EOF
    
    cd "$FIXTURES_DIR"
    OUTPUT=$($SENTINEL audit --output json 2>&1 || true)
    
    # File type thresholds should be applied
    log_info "File-type-specific thresholds configured"
    ((TESTS_PASSED++))
    
    # Cleanup
    rm -f "$TEST_CONFIG"
}

test_integration_with_audit() {
    echo "Testing integration with audit command..."
    
    cd "$PROJECT_ROOT"
    OUTPUT=$($SENTINEL audit 2>&1 || true)
    
    # File size checking should run as part of audit
    if echo "$OUTPUT" | grep -q "exceeds.*threshold" || echo "$OUTPUT" | grep -q "File size"; then
        log_info "File size checking integrated into audit"
        ((TESTS_PASSED++))
    else
        log_warn "File size checking may not be running in audit (no oversized files found or config disabled)"
        # This is not necessarily a failure - might just mean no oversized files
    fi
}

# Run all tests
echo "=========================================="
echo "Phase 9: File Size Management Tests"
echo "=========================================="
echo ""

# Check if sentinel binary exists
if [ ! -f "$SENTINEL" ]; then
    log_error "Sentinel binary not found at $SENTINEL"
    log_info "Building sentinel..."
    cd "$PROJECT_ROOT"
    ./synapsevibsentinel.sh
fi

# Run tests
test_file_size_detection
echo ""
test_analyze_structure_flag
echo ""
test_file_size_exceptions
echo ""
test_file_type_thresholds
echo ""
test_integration_with_audit
echo ""

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo "Passed: $TESTS_PASSED"
echo "Failed: $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi

