#!/bin/bash
# Agent Telemetry Tests
# Tests telemetry client functionality in the Agent

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

echo "🧪 Testing Agent Telemetry Client"
echo "══════════════════════════════════════════════════════════════"
echo ""

PASSED=0
FAILED=0

# Test 1: Telemetry types exist
echo "Test 1: Telemetry Types"
if grep -q "type TelemetryConfig struct" synapsevibsentinel.sh && \
   grep -q "type TelemetryEvent struct" synapsevibsentinel.sh && \
   grep -q "type TelemetryQueue struct" synapsevibsentinel.sh; then
    echo "   ✅ Telemetry types found"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Telemetry types missing"
    FAILED=$((FAILED + 1))
fi

# Test 2: Telemetry functions exist
echo "Test 2: Telemetry Functions"
if grep -q "func getTelemetryConfig" synapsevibsentinel.sh && \
   grep -q "func sendTelemetry" synapsevibsentinel.sh && \
   grep -q "func queueTelemetryEvent" synapsevibsentinel.sh && \
   grep -q "func loadTelemetryQueue" synapsevibsentinel.sh && \
   grep -q "func saveTelemetryQueue" synapsevibsentinel.sh && \
   grep -q "func flushTelemetryQueue" synapsevibsentinel.sh; then
    echo "   ✅ Telemetry functions found"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Telemetry functions missing"
    FAILED=$((FAILED + 1))
fi

# Test 3: Telemetry helper functions
echo "Test 3: Telemetry Helper Functions"
if grep -q "func sendAuditTelemetry" synapsevibsentinel.sh && \
   grep -q "func sendFixTelemetry" synapsevibsentinel.sh && \
   grep -q "func sendPatternTelemetry" synapsevibsentinel.sh && \
   grep -q "func calculateCompliance" synapsevibsentinel.sh; then
    echo "   ✅ Telemetry helper functions found"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Telemetry helper functions missing"
    FAILED=$((FAILED + 1))
fi

# Test 4: Integration in runAudit
echo "Test 4: Integration in runAudit"
if grep -q "sendAuditTelemetry" synapsevibsentinel.sh; then
    echo "   ✅ Telemetry integrated in runAudit"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Telemetry not integrated in runAudit"
    FAILED=$((FAILED + 1))
fi

# Test 5: Integration in runFix
echo "Test 5: Integration in runFix"
if grep -q "sendFixTelemetry" synapsevibsentinel.sh; then
    echo "   ✅ Telemetry integrated in runFix"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Telemetry not integrated in runFix"
    FAILED=$((FAILED + 1))
fi

# Test 6: Integration in runLearn
echo "Test 6: Integration in runLearn"
if grep -q "sendPatternTelemetry" synapsevibsentinel.sh; then
    echo "   ✅ Telemetry integrated in runLearn"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Telemetry not integrated in runLearn"
    FAILED=$((FAILED + 1))
fi

# Test 7: Offline queue file path
echo "Test 7: Offline Queue File Path"
if grep -q "telemetry-queue.json" synapsevibsentinel.sh; then
    echo "   ✅ Queue file path defined"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Queue file path missing"
    FAILED=$((FAILED + 1))
fi

# Test 8: Agent ID generation
echo "Test 8: Agent ID Generation"
if grep -q "func getAgentID" synapsevibsentinel.sh; then
    echo "   ✅ Agent ID function found"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Agent ID function missing"
    FAILED=$((FAILED + 1))
fi

# Test 9: Payload sanitization (client-side check)
echo "Test 9: Event Structure"
if grep -A 10 "type TelemetryEvent struct" synapsevibsentinel.sh | grep -q "Metrics"; then
    echo "   ✅ Event structure includes Metrics"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Event structure incomplete"
    FAILED=$((FAILED + 1))
fi

# Test 10: Hub endpoint construction
echo "Test 10: Hub Endpoint Construction"
if grep -q "/api/v1/telemetry" synapsevibsentinel.sh; then
    echo "   ✅ Hub endpoint path correct"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Hub endpoint path missing"
    FAILED=$((FAILED + 1))
fi

echo ""
echo "=============================================="
echo "   Agent Telemetry Test Results"
echo "=============================================="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "✅ All agent telemetry tests passed!"
    exit 0
else
    echo "❌ Some tests failed"
    exit 1
fi

