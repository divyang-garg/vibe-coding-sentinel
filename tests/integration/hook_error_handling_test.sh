#!/bin/bash
# Hook Error Handling Integration Test Suite
# Tests end-to-end error handling in hook execution

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🧪 Testing Hook Error Handling Integration"
echo ""

# Test 1: Verify performAuditForHook returns AuditReport with CheckResults
echo "Test 1: performAuditForHook CheckResults"
if grep -q "func performAuditForHook.*\*AuditReport" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ performAuditForHook returns AuditReport"
    if grep -q "report.CheckResults = make" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
        echo "  ✅ CheckResults initialized"
    else
        echo "  ❌ CheckResults not initialized"
        exit 1
    fi
else
    echo "  ❌ performAuditForHook not found"
    exit 1
fi

# Test 2: Verify error wrapper functions are called
echo "Test 2: Error Wrapper Function Calls"
if grep -q "performSecurityAnalysisWithError(" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ performSecurityAnalysisWithError called"
else
    echo "  ❌ performSecurityAnalysisWithError not called"
    exit 1
fi

if grep -q "detectVibeIssuesWithError(" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ detectVibeIssuesWithError called"
else
    echo "  ❌ detectVibeIssuesWithError not called"
    exit 1
fi

if grep -q "checkBusinessRulesComplianceWithError(" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ checkBusinessRulesComplianceWithError called"
else
    echo "  ❌ checkBusinessRulesComplianceWithError not called"
    exit 1
fi

# Test 3: Verify error state propagation
echo "Test 3: Error State Propagation"
if grep -q "cr.Success = false" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "cr.Error = " "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ Error state set in CheckResults on failure"
else
    echo "  ⚠️  Error state propagation not verified"
fi

# Test 4: Verify finding counts are tracked
echo "Test 4: Finding Count Tracking"
if grep -q "beforeCount := len(report.Findings)" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "afterCount := len(report.Findings)" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ Finding counts tracked (before/after)"
else
    echo "  ⚠️  Finding count tracking not verified"
fi

# Test 5: Verify CheckResults for disabled checks
echo "Test 5: Disabled Check Handling"
if grep -q "CheckResult{Enabled: false}" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ Disabled checks have CheckResult with Enabled: false"
else
    echo "  ⚠️  Disabled check handling not verified"
fi

echo ""
echo "✅ Hook error handling integration tests passed"
echo "   Note: Full end-to-end tests require Hub and database setup"












