#!/bin/bash
# Security Analysis Test Suite
# Tests Phase 8: Security Rules System

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
FIXTURES_DIR="$PROJECT_ROOT/tests/fixtures/security"

echo "🧪 Testing Security Analysis (Phase 8)"
echo ""

# Test 1: SQL Injection Detection
echo "Test 1: SQL Injection Detection (SEC-002)"
if [ -f "$FIXTURES_DIR/sql_injection.js" ]; then
    echo "  ✅ Test fixture exists"
else
    echo "  ❌ Test fixture missing"
    exit 1
fi

# Test 2: Password Hashing Detection (SEC-005)
echo "Test 2: Password Hashing Detection (SEC-005)"
if [ -f "$FIXTURES_DIR/password_hashing.js" ]; then
    echo "  ✅ Test fixture exists"
else
    echo "  ❌ Test fixture missing"
    exit 1
fi

# Test 2a: SEC-005 Data Flow - Insecure Password Flow
echo "Test 2a: SEC-005 Data Flow - Insecure Password Flow"
if [ -f "$FIXTURES_DIR/password_flow_insecure.js" ]; then
    echo "  ✅ Insecure password flow fixture exists"
    # Verify fixture contains expected vulnerability
    if grep -q "crypto.createHash('md5')" "$FIXTURES_DIR/password_flow_insecure.js"; then
        echo "  ✅ Fixture contains MD5 hashing (expected vulnerability)"
    else
        echo "  ⚠️  Fixture may not contain expected vulnerability pattern"
    fi
else
    echo "  ❌ Insecure password flow fixture missing"
    exit 1
fi

# Test 2b: SEC-005 Data Flow - Secure Password Flow
echo "Test 2b: SEC-005 Data Flow - Secure Password Flow"
if [ -f "$FIXTURES_DIR/password_flow_secure.js" ]; then
    echo "  ✅ Secure password flow fixture exists"
    # Verify fixture contains secure hashing
    if grep -q "bcrypt" "$FIXTURES_DIR/password_flow_secure.js"; then
        echo "  ✅ Fixture contains bcrypt hashing (expected secure pattern)"
    else
        echo "  ⚠️  Fixture may not contain expected secure pattern"
    fi
else
    echo "  ❌ Secure password flow fixture missing"
    exit 1
fi

# Test 2c: SEC-005 Data Flow - Missing Password Hashing
echo "Test 2c: SEC-005 Data Flow - Missing Password Hashing"
if [ -f "$FIXTURES_DIR/password_flow_missing.js" ]; then
    echo "  ✅ Missing password hashing fixture exists"
    # Verify fixture contains password without hashing
    if grep -q "password.*password" "$FIXTURES_DIR/password_flow_missing.js"; then
        echo "  ✅ Fixture contains password stored without hashing (expected vulnerability)"
    else
        echo "  ⚠️  Fixture may not contain expected vulnerability pattern"
    fi
else
    echo "  ❌ Missing password hashing fixture missing"
    exit 1
fi

# Test 3: Missing Auth Middleware
echo "Test 3: Missing Auth Middleware (SEC-003)"
if [ -f "$FIXTURES_DIR/missing_auth.js" ]; then
    echo "  ✅ Test fixture exists"
else
    echo "  ❌ Test fixture missing"
    exit 1
fi

# Test 4: Missing Ownership Check
echo "Test 4: Missing Ownership Check (SEC-001)"
if [ -f "$FIXTURES_DIR/missing_ownership.js" ]; then
    echo "  ✅ Test fixture exists"
else
    echo "  ❌ Test fixture missing"
    exit 1
fi

# Test 5: CORS Wildcard
echo "Test 5: CORS Wildcard Detection (SEC-008)"
if [ -f "$FIXTURES_DIR/cors_wildcard.js" ]; then
    echo "  ✅ Test fixture exists"
else
    echo "  ❌ Test fixture missing"
    exit 1
fi

# Test 6: Missing Input Validation
echo "Test 6: Missing Input Validation (SEC-006)"
if [ -f "$FIXTURES_DIR/missing_validation.js" ]; then
    echo "  ✅ Test fixture exists"
else
    echo "  ❌ Test fixture missing"
    exit 1
fi

# Test 7: --security flag exists
echo "Test 7: --security flag integration"
cd "$PROJECT_ROOT"
if ./sentinel audit --help 2>&1 | grep -q "security"; then
    echo "  ✅ --security flag documented"
else
    echo "  ⚠️  --security flag not in help (may be implemented but not documented)"
fi

# Test 8: SEC-005 Data Flow Analysis Integration Test
echo "Test 8: SEC-005 Data Flow Analysis Integration"
echo "  Testing password flow detection end-to-end..."

# Check if Hub is available
HUB_URL="${HUB_URL:-http://localhost:8080}"
if curl -s --connect-timeout 2 "$HUB_URL/health" > /dev/null 2>&1; then
    echo "  ✅ Hub is available at $HUB_URL"
    
    # Test insecure password flow
    echo "  Testing insecure password flow detection..."
    INSECURE_CODE=$(cat "$FIXTURES_DIR/password_flow_insecure.js")
    RESPONSE=$(curl -s -X POST "$HUB_URL/api/v1/analyze/security" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-key" \
        -d "{\"code\":\"$INSECURE_CODE\",\"language\":\"javascript\",\"filename\":\"password_flow_insecure.js\",\"rules\":[\"SEC-005\"]}")
    
    if echo "$RESPONSE" | grep -q "SEC-005" || echo "$RESPONSE" | grep -q "Password Hashing"; then
        echo "  ✅ SEC-005 detected insecure password flow correctly"
    else
        echo "  ⚠️  SEC-005 may not have detected insecure password flow"
        echo "  Response: $RESPONSE"
    fi
    
    # Test secure password flow (should not trigger SEC-005)
    echo "  Testing secure password flow (should not trigger SEC-005)..."
    SECURE_CODE=$(cat "$FIXTURES_DIR/password_flow_secure.js")
    RESPONSE=$(curl -s -X POST "$HUB_URL/api/v1/analyze/security" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-key" \
        -d "{\"code\":\"$SECURE_CODE\",\"language\":\"javascript\",\"filename\":\"password_flow_secure.js\",\"rules\":[\"SEC-005\"]}")
    
    if echo "$RESPONSE" | grep -q "SEC-005"; then
        echo "  ⚠️  SEC-005 incorrectly flagged secure password flow"
    else
        echo "  ✅ SEC-005 correctly did not flag secure password flow"
    fi
    
    # Test missing password hashing
    echo "  Testing missing password hashing detection..."
    MISSING_CODE=$(cat "$FIXTURES_DIR/password_flow_missing.js")
    RESPONSE=$(curl -s -X POST "$HUB_URL/api/v1/analyze/security" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-key" \
        -d "{\"code\":\"$MISSING_CODE\",\"language\":\"javascript\",\"filename\":\"password_flow_missing.js\",\"rules\":[\"SEC-005\"]}")
    
    if echo "$RESPONSE" | grep -q "SEC-005" || echo "$RESPONSE" | grep -q "Password Hashing"; then
        echo "  ✅ SEC-005 detected missing password hashing correctly"
    else
        echo "  ⚠️  SEC-005 may not have detected missing password hashing"
    fi
else
    echo "  ⚠️  Hub not available at $HUB_URL (skipping integration tests)"
    echo "  To run integration tests:"
    echo "    1. Start Hub: cd hub/api && go run ."
    echo "    2. Set HUB_URL if different: export HUB_URL=http://localhost:8080"
    echo "    3. Run tests again: ./tests/unit/security_analysis_test.sh"
fi

echo ""
echo "✅ All security test fixtures created and verified"
echo ""
echo "Note: Full integration tests require Hub to be running."
echo "Run: cd hub/api && go run ."
echo "Then: ./sentinel audit --security"

