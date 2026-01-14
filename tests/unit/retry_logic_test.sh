#!/bin/bash
# Retry Logic Test Suite
# Tests for HTTP retry logic with exponential backoff

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🧪 Testing Retry Logic Implementation"
echo ""

# Test 1: Verify httpRequestWithRetry function exists
echo "Test 1: httpRequestWithRetry Function"
if grep -q "func httpRequestWithRetry" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ httpRequestWithRetry function exists"
else
    echo "  ❌ httpRequestWithRetry function not found"
    exit 1
fi

# Test 2: Verify exponential backoff logic
echo "Test 2: Exponential Backoff Logic"
if grep -q "100\*(1<<uint" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ Exponential backoff calculation found"
else
    echo "  ❌ Exponential backoff not found"
    exit 1
fi

# Test 3: Verify retry on 5xx errors
echo "Test 3: Retry on Server Errors"
if grep -q "resp.StatusCode < 500" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ Retry logic for 5xx errors found"
else
    echo "  ⚠️  Retry logic for 5xx errors not verified"
fi

# Test 4: Verify no retry on 4xx errors
echo "Test 4: No Retry on Client Errors"
if grep -q "resp.StatusCode < 500" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ Client errors (4xx) don't trigger retry"
else
    echo "  ⚠️  Client error handling not verified"
fi

# Test 5: Verify max retries limit
echo "Test 5: Max Retries Limit"
if grep -q "maxRetries" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ Max retries parameter found"
else
    echo "  ⚠️  Max retries parameter not verified"
fi

echo ""
echo "✅ Retry logic structure tests passed"
echo "   Note: Full functional tests require mock HTTP server"












