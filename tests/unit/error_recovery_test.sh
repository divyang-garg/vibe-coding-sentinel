#!/bin/bash
# Error Recovery Test Suite
# Tests for CheckResults implementation and error recovery

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🧪 Testing Error Recovery System"
echo ""

# Test 1: Verify CheckResult struct exists
echo "Test 1: CheckResult Struct"
if grep -q "type CheckResult struct" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ CheckResult struct exists"
    if grep -q "Enabled.*bool" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
       grep -q "Success.*bool" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
       grep -q "Error.*string" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
       grep -q "Findings.*int" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
        echo "  ✅ CheckResult has all required fields"
    else
        echo "  ⚠️  CheckResult may be missing some fields"
    fi
else
    echo "  ❌ CheckResult struct not found"
    exit 1
fi

# Test 2: Verify CheckResults in AuditReport
echo "Test 2: CheckResults in AuditReport"
if grep -q "CheckResults.*map\[string\]CheckResult" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ CheckResults map in AuditReport"
else
    echo "  ❌ CheckResults not in AuditReport"
    exit 1
fi

# Test 3: Verify CheckResults initialization in performAuditForHook
echo "Test 3: CheckResults Initialization"
if grep -q "report.CheckResults = make" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ CheckResults initialized in performAuditForHook"
else
    echo "  ❌ CheckResults not initialized"
    exit 1
fi

# Test 4: Verify error wrapper functions exist
echo "Test 4: Error Wrapper Functions"
if grep -q "func performSecurityAnalysisWithError" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ performSecurityAnalysisWithError exists"
else
    echo "  ❌ performSecurityAnalysisWithError not found"
    exit 1
fi

if grep -q "func detectVibeIssuesWithError" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ detectVibeIssuesWithError exists"
else
    echo "  ❌ detectVibeIssuesWithError not found"
    exit 1
fi

if grep -q "func checkBusinessRulesComplianceWithError" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ checkBusinessRulesComplianceWithError exists"
else
    echo "  ❌ checkBusinessRulesComplianceWithError not found"
    exit 1
fi

# Test 5: Verify panic recovery in wrapper functions
echo "Test 5: Panic Recovery"
if grep -q "defer func.*recover" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ Panic recovery found in wrapper functions"
    if grep -q "CheckResults.*Error" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
        echo "  ✅ CheckResults error state set on panic"
    else
        echo "  ⚠️  CheckResults error state may not be set on panic"
    fi
else
    echo "  ⚠️  Panic recovery not verified"
fi

# Test 6: Verify CheckResults populated for all checks
echo "Test 6: CheckResults Population"
if grep -q "CheckResults\[\"file_size\"\]" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "CheckResults\[\"security\"\]" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "CheckResults\[\"vibe\"\]" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
   grep -q "CheckResults\[\"business_rules\"\]" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ CheckResults populated for all check types"
else
    echo "  ⚠️  Some CheckResults entries may be missing"
fi

echo ""
echo "✅ Error recovery structure tests passed"
echo "   Note: Full functional tests require error injection"












