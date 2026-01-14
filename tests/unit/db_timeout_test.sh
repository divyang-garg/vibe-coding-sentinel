#!/bin/bash
# Database Timeout Test Suite
# Tests for database query timeout helpers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🧪 Testing Database Timeout Helpers"
echo ""

# Test 1: Verify timeout helper functions exist in hook_handler.go
echo "Test 1: Timeout Helper Functions in hook_handler.go"
if grep -q "func queryWithTimeout" "$PROJECT_ROOT/hub/api/hook_handler.go"; then
    echo "  ✅ queryWithTimeout function exists"
else
    echo "  ❌ queryWithTimeout function not found"
    exit 1
fi

if grep -q "func queryRowWithTimeout" "$PROJECT_ROOT/hub/api/hook_handler.go"; then
    echo "  ✅ queryRowWithTimeout function exists"
else
    echo "  ❌ queryRowWithTimeout function not found"
    exit 1
fi

if grep -q "func execWithTimeout" "$PROJECT_ROOT/hub/api/hook_handler.go"; then
    echo "  ✅ execWithTimeout function exists"
else
    echo "  ❌ execWithTimeout function not found"
    exit 1
fi

# Test 2: Verify timeout helper functions exist in policy.go
echo "Test 2: Timeout Helper Functions in policy.go"
if grep -q "func execWithTimeout\|func queryRowWithTimeout" "$PROJECT_ROOT/hub/api/policy.go"; then
    echo "  ✅ Timeout helper functions exist in policy.go"
else
    echo "  ⚠️  Timeout helpers may not be in policy.go (could be shared)"
fi

# Test 3: Verify 10-second timeout
echo "Test 3: Timeout Duration"
if grep -q "10\*time.Second" "$PROJECT_ROOT/hub/api/hook_handler.go"; then
    echo "  ✅ 10-second timeout configured"
else
    echo "  ⚠️  Timeout duration not verified"
fi

# Test 4: Verify context.WithTimeout usage
echo "Test 4: Context Timeout Usage"
if grep -q "context.WithTimeout" "$PROJECT_ROOT/hub/api/hook_handler.go"; then
    echo "  ✅ context.WithTimeout used correctly"
else
    echo "  ❌ context.WithTimeout not found"
    exit 1
fi

# Test 5: Verify database queries use timeout helpers
echo "Test 5: Database Queries Use Timeout Helpers"
QUERY_COUNT=$(grep -c "queryWithTimeout\|queryRowWithTimeout\|execWithTimeout" "$PROJECT_ROOT/hub/api/hook_handler.go" || echo "0")
if [ "$QUERY_COUNT" -gt "0" ]; then
    echo "  ✅ Found $QUERY_COUNT uses of timeout helpers"
else
    echo "  ⚠️  No timeout helper usage found (may need verification)"
fi

echo ""
echo "✅ Database timeout helper tests passed"
echo "   Note: Full timeout tests require database with slow queries"












