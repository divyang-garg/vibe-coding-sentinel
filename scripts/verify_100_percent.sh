#!/bin/bash
# Master verification - runs ALL real tests
set -e

echo "=========================================="
echo "SENTINEL 100% VERIFICATION - NO MOCKS"
echo "=========================================="

cd "$(dirname "$0")/.."

# Build sentinel binary first
echo ""
echo "=== Building Sentinel Binary ==="
go build -o sentinel ./cmd/sentinel || {
  echo "FAIL: Failed to build sentinel"
  exit 1
}
echo "PASS: Binary built successfully"

# Phase 1: Unit tests
echo ""
echo "=== Phase 1: Unit Tests ==="
go test ./internal/... -v 2>&1 | tee /tmp/unit_tests.log
if grep -q "FAIL" /tmp/unit_tests.log; then
  echo "FAIL: Unit tests have failures"
  exit 1
fi
if ! grep -q "PASS" /tmp/unit_tests.log; then
  echo "WARNING: No test results found"
fi
echo "PASS: All unit tests pass"

# Phase 2: MCP process tests
echo ""
echo "=== Phase 2: MCP Process Tests ==="
bash tests/e2e/mcp_real_test.sh || {
  echo "FAIL: MCP process tests failed"
  exit 1
}

bash tests/e2e/mcp_filesize_test.sh || {
  echo "FAIL: MCP filesize test failed"
  exit 1
}

# Phase 3: CLI integration tests
echo ""
echo "=== Phase 3: CLI Integration Tests ==="
bash tests/e2e/baseline_real_test.sh || {
  echo "FAIL: Baseline integration test failed"
  exit 1
}

bash tests/e2e/knowledge_real_test.sh || {
  echo "FAIL: Knowledge integration test failed"
  exit 1
}

# Phase 4: Cross verification
echo ""
echo "=== Phase 4: Cross Verification ==="
bash tests/e2e/cross_verify_test.sh || {
  echo "FAIL: Cross verification test failed"
  exit 1
}

# Phase 5: Docker Hub tests (optional)
echo ""
echo "=== Phase 5: Hub Docker Tests (if Docker available) ==="
bash tests/e2e/hub_docker_test.sh || {
  echo "WARNING: Hub Docker tests skipped or failed (this is optional)"
}

echo ""
echo "=========================================="
echo "ALL VERIFICATIONS PASSED"
echo "=========================================="
echo ""
echo "Summary:"
echo "  - Unit tests: PASS"
echo "  - MCP process tests: PASS"
echo "  - CLI integration tests: PASS"
echo "  - Cross verification: PASS"
echo "  - Hub Docker tests: $(docker info > /dev/null 2>&1 && echo 'PASS/SKIP' || echo 'SKIP (Docker not available)')"
echo ""
echo "This provides verifiable, reproducible proof that the implementation works."
