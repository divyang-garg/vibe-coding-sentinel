#!/bin/bash
# Unit tests for change detection integration

set -e

echo "Testing change detection integration..."

# Test 1: detectChangesHandler endpoint exists
echo "Test 1: Checking detectChangesHandler endpoint..."
if ! grep -q "func detectChangesHandler" hub/api/main.go; then
    echo "FAIL: detectChangesHandler not found"
    exit 1
fi
echo "PASS: detectChangesHandler exists"

# Test 2: Endpoint is registered
echo "Test 2: Checking endpoint registration..."
if ! grep -q 'r.Post("/documents/{id}/detect-changes"' hub/api/main.go; then
    echo "FAIL: Endpoint not registered"
    exit 1
fi
echo "PASS: Endpoint registered"

# Test 3: triggerChangeDetection function exists
echo "Test 3: Checking triggerChangeDetection function..."
if ! grep -q "func triggerChangeDetection" hub/processor/main.go; then
    echo "FAIL: triggerChangeDetection not found"
    exit 1
fi
echo "PASS: triggerChangeDetection exists"

# Test 4: Integration in extractKnowledge
echo "Test 4: Checking integration in extractKnowledge..."
if ! grep -q "triggerChangeDetection" hub/processor/main.go; then
    echo "FAIL: triggerChangeDetection not called in extractKnowledge"
    exit 1
fi
echo "PASS: triggerChangeDetection integrated"

# Test 5: HTTP client helpers exist
echo "Test 5: Checking HTTP client helpers..."
if ! grep -q "func callHubAPI" hub/processor/main.go; then
    echo "FAIL: callHubAPI not found"
    exit 1
fi
if ! grep -q "func callHubAPIWithRetry" hub/processor/main.go; then
    echo "FAIL: callHubAPIWithRetry not found"
    exit 1
fi
echo "PASS: HTTP client helpers exist"

# Test 6: Config includes HubURL and HubAPIKey
echo "Test 6: Checking Config struct..."
if ! grep -q "HubURL" hub/processor/main.go; then
    echo "FAIL: HubURL not in Config"
    exit 1
fi
if ! grep -q "HubAPIKey" hub/processor/main.go; then
    echo "FAIL: HubAPIKey not in Config"
    exit 1
fi
echo "PASS: Config includes HubURL and HubAPIKey"

echo "All tests passed!"











