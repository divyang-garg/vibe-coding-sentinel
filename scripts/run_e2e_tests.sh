#!/bin/bash
# E2E Test Runner Script
# Runs all end-to-end tests for Sentinel Hub API
# Complies with CODING_STANDARDS.md: Scripts should be executable and well-documented

set -e

echo "=== Running Sentinel Hub API E2E Tests ==="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results tracking
PASSED=0
FAILED=0
SKIPPED=0

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${YELLOW}Warning: Docker is not running. Some E2E tests may be skipped.${NC}"
fi

# Check if test database is available
if ! docker-compose -f hub/docker-compose.yml -f hub/docker-compose.test.yml ps test-db 2>/dev/null | grep -q "Up"; then
    echo -e "${YELLOW}Starting test database...${NC}"
    docker-compose -f hub/docker-compose.yml -f hub/docker-compose.test.yml up -d test-db
    echo "Waiting for database to be ready..."
    sleep 5
fi

# Function to run a test script
run_test() {
    local test_file=$1
    local test_name=$(basename "$test_file" .sh)
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Running: $test_name"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if [ ! -f "$test_file" ]; then
        echo -e "${YELLOW}SKIP: Test file not found: $test_file${NC}"
        SKIPPED=$((SKIPPED + 1))
        return
    fi
    
    if [ ! -x "$test_file" ]; then
        chmod +x "$test_file"
    fi
    
    if bash "$test_file"; then
        echo -e "${GREEN}✓ PASSED: $test_name${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAILED: $test_name${NC}"
        FAILED=$((FAILED + 1))
    fi
}

# Run all E2E tests
echo "Discovering E2E test files..."
E2E_TESTS=(
    "tests/e2e/mcp_e2e.sh"
    "tests/e2e/hub_api_security_e2e_test.sh"
    "tests/e2e/api_key_management_e2e_test.sh"
    "tests/e2e/document_processing_e2e_test.sh"
    "tests/e2e/mcp_toolchain_e2e_test.sh"
)

for test in "${E2E_TESTS[@]}"; do
    if [ -f "$test" ]; then
        run_test "$test"
    else
        echo -e "${YELLOW}SKIP: Test file not found: $test${NC}"
        SKIPPED=$((SKIPPED + 1))
    fi
done

# Run integration E2E tests
echo ""
echo "Running integration E2E tests..."
INTEGRATION_E2E_TESTS=(
    "tests/integration/phase12_e2e_test.sh"
    "tests/integration/phase13_e2e_test.sh"
    "tests/integration/phase14a_e2e_test.sh"
    "tests/integration/phase14b_e2e_test.sh"
    "tests/integration/phase14c_e2e_test.sh"
    "tests/integration/phase14d_e2e_test.sh"
    "tests/integration/test_enforcement_e2e_test.sh"
)

for test in "${INTEGRATION_E2E_TESTS[@]}"; do
    if [ -f "$test" ]; then
        run_test "$test"
    fi
done

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "=== E2E Test Summary ==="
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}Passed:  $PASSED${NC}"
echo -e "${RED}Failed:  $FAILED${NC}"
echo -e "${YELLOW}Skipped: $SKIPPED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All E2E tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some E2E tests failed.${NC}"
    exit 1
fi
