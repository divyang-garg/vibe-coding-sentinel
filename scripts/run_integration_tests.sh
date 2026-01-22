#!/bin/bash

# Integration Test Runner
# Runs integration tests with Docker database setup
# Complies with CODING_STANDARDS.md: Scripts for build/deployment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HUB_DIR="$PROJECT_ROOT/hub"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Integration Test Runner ===${NC}"
echo

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

# Check if test database is running
if ! docker ps | grep -q "sentinel-test-db\|test-db"; then
    echo -e "${YELLOW}⚠️  Test database not running. Starting test database...${NC}"
    cd "$HUB_DIR"
    
    # Start test database
    docker-compose -f docker-compose.yml -f docker-compose.test.yml up -d test-db
    
    # Wait for database to be ready
    echo "Waiting for test database to be ready..."
    timeout=30
    elapsed=0
    while [ $elapsed -lt $timeout ]; do
        if docker exec sentinel-postgres-test-db pg_isready -U sentinel -d sentinel_test > /dev/null 2>&1 || \
           docker exec test-db pg_isready -U sentinel -d sentinel_test > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Test database is ready${NC}"
            break
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done
    
    if [ $elapsed -ge $timeout ]; then
        echo -e "${RED}❌ Test database failed to start${NC}"
        exit 1
    fi
fi

# Set test environment variables
export TEST_DB_HOST="${TEST_DB_HOST:-localhost}"
export TEST_DB_PORT="${TEST_DB_PORT:-5433}"
export TEST_DB_USER="${TEST_DB_USER:-sentinel}"
export TEST_DB_PASSWORD="${TEST_DB_PASSWORD:-sentinel}"
export TEST_DB_NAME="${TEST_DB_NAME:-sentinel_test}"
export TEST_DB_SSLMODE="${TEST_DB_SSLMODE:-disable}"

echo -e "${GREEN}Running integration tests...${NC}"
echo "Database: $TEST_DB_HOST:$TEST_DB_PORT/$TEST_DB_NAME"
echo

# Run integration tests
cd "$PROJECT_ROOT"

# Test hub/api integration tests
if [ -d "hub/api" ]; then
    echo -e "${GREEN}Running hub/api integration tests...${NC}"
    cd hub/api
    
    # Run services integration test
    if [ -f "services/integration_test.go" ]; then
        go test -v -tags=integration ./services -run TestIntegrationSuite 2>&1 | tee /tmp/integration_test_output.log
    fi
    
    cd "$PROJECT_ROOT"
fi

# Test internal integration tests
if [ -d "internal/extraction" ]; then
    echo -e "${GREEN}Running internal/extraction integration tests...${NC}"
    go test -v -tags=integration ./internal/extraction -run TestLiveLLMExtraction 2>&1 | tee -a /tmp/integration_test_output.log || true
fi

# Test tests/integration
if [ -d "tests/integration" ]; then
    echo -e "${GREEN}Running tests/integration tests...${NC}"
    go test -v -tags=integration ./tests/integration 2>&1 | tee -a /tmp/integration_test_output.log || true
fi

echo
echo -e "${GREEN}=== Integration Test Summary ===${NC}"
echo "Test output saved to: /tmp/integration_test_output.log"

# Check if tests passed
if grep -q "PASS" /tmp/integration_test_output.log; then
    echo -e "${GREEN}✅ Some integration tests passed${NC}"
fi

if grep -q "FAIL" /tmp/integration_test_output.log; then
    echo -e "${RED}❌ Some integration tests failed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Integration test runner completed${NC}"
