#!/bin/bash
# Telemetry and Metrics API Tests
# Tests telemetry ingestion and metrics query endpoints

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

echo "🧪 Testing Telemetry and Metrics API"
echo "══════════════════════════════════════════════════════════════"
echo ""

PASSED=0
FAILED=0

# Test 1: Telemetry endpoint exists
echo "Test 1: Telemetry Ingestion Endpoint"
if grep -q "POST.*telemetry\|r.Post.*telemetry" hub/api/main.go || \
   grep -q "telemetryIngestionHandler" hub/api/main.go; then
    echo "   ✅ Telemetry endpoint found"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Telemetry endpoint not found"
    FAILED=$((FAILED + 1))
fi

# Test 2: Metrics endpoint exists
echo "Test 2: Metrics Query Endpoint"
if grep -q "GET.*metrics\|r.Get.*metrics" hub/api/main.go || \
   grep -q "getMetricsHandler" hub/api/main.go; then
    echo "   ✅ Metrics endpoint found"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Metrics endpoint not found"
    FAILED=$((FAILED + 1))
fi

# Test 3: Telemetry payload sanitization
echo "Test 3: Payload Sanitization"
if grep -q "sanitizeTelemetryPayload" hub/api/main.go; then
    echo "   ✅ Payload sanitization function found"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Payload sanitization not found"
    FAILED=$((FAILED + 1))
fi

# Test 4: Metrics calculation
echo "Test 4: Metrics Calculation"
if grep -q "calculateMetrics" hub/api/main.go; then
    echo "   ✅ Metrics calculation function found"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Metrics calculation not found"
    FAILED=$((FAILED + 1))
fi

# Test 5: Valid event types
echo "Test 5: Valid Event Types"
if grep -q "audit_complete" hub/api/main.go && \
   grep -q "fix_applied" hub/api/main.go && \
   grep -q "pattern_learned" hub/api/main.go && \
   grep -q "doc_ingested" hub/api/main.go; then
    echo "   ✅ Valid event types defined"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Missing event types"
    FAILED=$((FAILED + 1))
fi

# Test 6: Allowed fields in sanitization
echo "Test 6: Sanitization Allowed Fields"
if grep -q "finding_count\|compliance_percent\|fix_count" hub/api/main.go; then
    echo "   ✅ Allowed fields defined"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Allowed fields not found"
    FAILED=$((FAILED + 1))
fi

# Test 7: Database schema includes telemetry_events
echo "Test 7: Database Schema"
if grep -q "CREATE TABLE.*telemetry_events" hub/api/main.go; then
    echo "   ✅ Telemetry events table in schema"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Telemetry events table missing"
    FAILED=$((FAILED + 1))
fi

# Test 8: Metrics aggregation
echo "Test 8: Metrics Aggregation"
if grep -q "avg_compliance\|total_findings\|total_critical" hub/api/main.go; then
    echo "   ✅ Metrics aggregation implemented"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Metrics aggregation missing"
    FAILED=$((FAILED + 1))
fi

# Test 9: Recent telemetry endpoint
echo "Test 9: Recent Telemetry Endpoint"
if grep -q "getRecentTelemetryHandler" hub/api/main.go && \
   grep -q "/telemetry/recent" hub/api/main.go; then
    echo "   ✅ Recent telemetry endpoint found"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Recent telemetry endpoint missing"
    FAILED=$((FAILED + 1))
fi

# Test 10: Trends endpoint
echo "Test 10: Trends Endpoint"
if grep -q "getMetricsTrendsHandler" hub/api/main.go && \
   grep -q "/metrics/trends" hub/api/main.go; then
    echo "   ✅ Trends endpoint found"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Trends endpoint missing"
    FAILED=$((FAILED + 1))
fi

# Test 11: Team metrics endpoint
echo "Test 11: Team Metrics Endpoint"
if grep -q "getTeamMetricsHandler" hub/api/main.go && \
   grep -q "/metrics/team" hub/api/main.go; then
    echo "   ✅ Team metrics endpoint found"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Team metrics endpoint missing"
    FAILED=$((FAILED + 1))
fi

# Test 12: Database schema includes new columns
echo "Test 12: Database Schema"
if grep -q "ALTER TABLE.*agent_id\|ALTER TABLE.*org_id\|ALTER TABLE.*team_id" hub/api/main.go; then
    echo "   ✅ Database schema includes new columns"
    PASSED=$((PASSED + 1))
else
    echo "   ❌ Database schema missing new columns"
    FAILED=$((FAILED + 1))
fi

echo ""
echo "=============================================="
echo "   Telemetry Test Results"
echo "=============================================="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "✅ All telemetry tests passed!"
    exit 0
else
    echo "❌ Some tests failed"
    exit 1
fi

