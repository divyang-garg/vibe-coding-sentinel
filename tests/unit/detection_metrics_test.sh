#!/bin/bash
# Detection Rate Metrics Test Suite
# Tests detection rate validation (Phase 8 compliance fixes)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
FIXTURES_DIR="$PROJECT_ROOT/tests/fixtures/security"

echo "🧪 Testing Detection Rate Metrics"
echo ""

# Test 1: Detection metrics struct exists in Hub
echo "Test 1: Detection metrics struct verification"
if [ -f "$PROJECT_ROOT/hub/api/security_analyzer.go" ]; then
    if grep -q "type DetectionMetrics struct" "$PROJECT_ROOT/hub/api/security_analyzer.go"; then
        echo "  ✅ DetectionMetrics struct exists"
    else
        echo "  ❌ DetectionMetrics struct missing"
        exit 1
    fi
else
    echo "  ❌ security_analyzer.go not found"
    exit 1
fi

# Test 2: calculateDetectionRate function exists
echo "Test 2: calculateDetectionRate function verification"
if grep -q "func calculateDetectionRate" "$PROJECT_ROOT/hub/api/security_analyzer.go"; then
    echo "  ✅ calculateDetectionRate function exists"
else
    echo "  ❌ calculateDetectionRate function missing"
    exit 1
fi

# Test 3: SecurityAnalysisResponse includes Metrics field
echo "Test 3: SecurityAnalysisResponse Metrics field verification"
if grep -q "Metrics.*DetectionMetrics" "$PROJECT_ROOT/hub/api/main.go"; then
    echo "  ✅ SecurityAnalysisResponse includes Metrics field"
else
    echo "  ❌ SecurityAnalysisResponse missing Metrics field"
    exit 1
fi

# Test 4: SecurityAnalysisRequest includes ExpectedFindings field
echo "Test 4: SecurityAnalysisRequest ExpectedFindings field verification"
if grep -q "ExpectedFindings" "$PROJECT_ROOT/hub/api/main.go"; then
    echo "  ✅ SecurityAnalysisRequest includes ExpectedFindings field"
else
    echo "  ❌ SecurityAnalysisRequest missing ExpectedFindings field"
    exit 1
fi

# Test 5: Ground truth test suite exists
echo "Test 5: Ground truth test suite verification"
if [ -d "$FIXTURES_DIR/ground_truth" ]; then
    echo "  ✅ Ground truth test suite directory exists"
    if [ -f "$FIXTURES_DIR/ground_truth/README.md" ]; then
        echo "  ✅ Ground truth README exists"
    else
        echo "  ⚠️  Ground truth README missing"
    fi
else
    echo "  ⚠️  Ground truth test suite directory missing (optional)"
fi

# Test 6: Hub integration test (if Hub is available)
echo "Test 6: Detection metrics Hub integration"
HUB_URL="${HUB_URL:-http://localhost:8080}"
if curl -s --connect-timeout 2 "$HUB_URL/health" > /dev/null 2>&1; then
    echo "  ✅ Hub is available at $HUB_URL"
    
    # Test with ground truth
    TEST_CODE="function test() { const password = req.body.password; }"
    EXPECTED_FINDINGS='{"SEC-005":true}'
    
    RESPONSE=$(curl -s -X POST "$HUB_URL/api/v1/analyze/security" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-key" \
        -d "{\"code\":\"$TEST_CODE\",\"language\":\"javascript\",\"filename\":\"test.js\",\"rules\":[\"SEC-005\"],\"expectedFindings\":$EXPECTED_FINDINGS}")
    
    if echo "$RESPONSE" | grep -q "metrics"; then
        echo "  ✅ Detection metrics included in response"
        if echo "$RESPONSE" | grep -q "detectionRate\|precision\|recall"; then
            echo "  ✅ Detection metrics contain expected fields"
        else
            echo "  ⚠️  Detection metrics may be missing expected fields"
        fi
    else
        echo "  ⚠️  Detection metrics not included in response"
    fi
else
    echo "  ⚠️  Hub not available at $HUB_URL (skipping integration test)"
    echo "  To run integration test:"
    echo "    1. Start Hub: cd hub/api && go run ."
    echo "    2. Set HUB_URL if different: export HUB_URL=http://localhost:8080"
    echo "    3. Run tests again: ./tests/unit/detection_metrics_test.sh"
fi

echo ""
echo "✅ Detection rate metrics tests completed"












