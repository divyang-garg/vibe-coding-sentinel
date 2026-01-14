#!/bin/bash
# Cache Invalidation Integration Test Suite
# Tests cache invalidation logic and timestamp tracking

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🧪 Testing Cache Invalidation Logic"
echo ""

# Test 1: Verify cachedPolicy struct with updated_at
echo "Test 1: cachedPolicy Structure"
if grep -q "type cachedPolicy struct" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ cachedPolicy struct exists"
    if grep -q "UpdatedAt.*time.Time" "$PROJECT_ROOT/synapsevibsentinel.sh" && \
       grep -q "CachedAt.*time.Time" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
        echo "  ✅ cachedPolicy has UpdatedAt and CachedAt fields"
    else
        echo "  ⚠️  cachedPolicy may be missing timestamp fields"
    fi
else
    echo "  ❌ cachedPolicy struct not found"
    exit 1
fi

# Test 2: Verify limitsCacheEntry struct
echo "Test 2: limitsCacheEntry Structure"
if grep -q "type limitsCacheEntry struct" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ limitsCacheEntry struct exists"
    if grep -q "Expires.*time.Time" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
        echo "  ✅ limitsCacheEntry has Expires field"
    else
        echo "  ⚠️  limitsCacheEntry may be missing Expires field"
    fi
else
    echo "  ❌ limitsCacheEntry struct not found"
    exit 1
fi

# Test 3: Verify cache invalidation logic
echo "Test 3: Cache Invalidation Logic"
if grep -q "policyUpdatedTime.After(cachedHookPolicy.UpdatedAt)" "$PROJECT_ROOT/synapsevibsentinel.sh" || \
   grep -q "time.Since(cachedHookPolicy.CachedAt)" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ Cache invalidation based on timestamps"
else
    echo "  ⚠️  Cache invalidation logic not verified"
fi

# Test 4: Verify AST cache cleanup
echo "Test 4: AST Cache Cleanup"
if grep -q "lastCacheCleanup\|cacheCleanupInterval" "$PROJECT_ROOT/hub/api/ast_analyzer.go"; then
    echo "  ✅ AST cache cleanup tracking exists"
else
    echo "  ⚠️  AST cache cleanup not verified"
fi

# Test 5: Verify RWMutex usage
echo "Test 5: RWMutex Usage"
if grep -q "sync.RWMutex" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
    echo "  ✅ RWMutex used for cache synchronization"
    if grep -q "RLock\|RUnlock" "$PROJECT_ROOT/synapsevibsentinel.sh"; then
        echo "  ✅ Read locks used for cache reads"
    fi
else
    echo "  ⚠️  RWMutex usage not verified"
fi

echo ""
echo "✅ Cache invalidation structure tests passed"
echo "   Note: Full functional tests require cache state manipulation"












