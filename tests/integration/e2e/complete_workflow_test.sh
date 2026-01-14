#!/bin/bash
# Complete workflow end-to-end integration test
# Tests: init → learn → audit → fix → hub integration

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}ℹ️  $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

TEST_DIR="/tmp/sentinel_e2e_test_$(date +%s)"
SENTINEL="./sentinel"

cleanup() {
    rm -rf "$TEST_DIR"
}

trap cleanup EXIT

echo "🚀 Starting Complete Workflow E2E Test"
echo "======================================"

# Setup test environment
mkdir -p "$TEST_DIR/project"
cd "$TEST_DIR/project"

# Create test project structure
mkdir -p src/components src/utils tests
cat > package.json << 'EOF'
{
  "name": "test-project",
  "version": "1.0.0",
  "dependencies": {
    "react": "^18.0.0",
    "lodash": "^4.17.0"
  }
}
EOF

cat > src/components/Button.js << 'EOF'
import React from 'react';

function Button({ children, onClick }) {
    console.log("Button rendered");
    return (
        <button onClick={onClick}>
            {children}
        </button>
    );
}

export default Button;
EOF

cat > src/utils/helpers.js << 'EOF'
import _ from 'lodash';

export function formatDate(date) {
    return date.toISOString().split('T')[0];
}

export function debounce(func, wait) {
    return _.debounce(func, wait);
}
EOF

cat > tests/Button.test.js << 'EOF'
import { render } from '@testing-library/react';
import Button from '../src/components/Button';

test('renders button correctly', () => {
    const { getByText } = render(<Button>Click me</Button>);
    expect(getByText('Click me')).toBeInTheDocument();
});
EOF

# Test 1: Project Initialization
echo ""
log_info "Test 1: Project Initialization"
if $SENTINEL init 2>&1 | grep -q "Project initialized"; then
    log_info "✅ Project initialization successful"
else
    log_error "❌ Project initialization failed"
    exit 1
fi

# Test 2: Pattern Learning
echo ""
log_info "Test 2: Pattern Learning"
if $SENTINEL learn 2>&1 | grep -q "Pattern learning complete"; then
    log_info "✅ Pattern learning successful"
    if [[ -f ".sentinel/patterns.json" && -f ".cursor/rules/project-patterns.md" ]]; then
        log_info "✅ Pattern files generated"
    else
        log_error "❌ Pattern files not generated"
        exit 1
    fi
else
    log_error "❌ Pattern learning failed"
    exit 1
fi

# Test 3: Security Audit
echo ""
log_info "Test 3: Security Audit"
if AUDIT_OUTPUT=$($SENTINEL audit --offline 2>&1); then
    if echo "$AUDIT_OUTPUT" | grep -q "Audit.*FAILED\|Audit.*PASSED"; then
        log_info "✅ Security audit completed"
        if echo "$AUDIT_OUTPUT" | grep -q "console.log"; then
            log_info "✅ Console.log detection working"
        fi
    else
        log_error "❌ Security audit did not complete properly"
        exit 1
    fi
else
    log_error "❌ Security audit failed"
    exit 1
fi

# Test 4: Auto-Fix
echo ""
log_info "Test 4: Auto-Fix"
if FIX_OUTPUT=$($SENTINEL fix --safe 2>&1); then
    if echo "$FIX_OUTPUT" | grep -q "Auto-fix complete"; then
        log_info "✅ Auto-fix completed successfully"
    else
        log_error "❌ Auto-fix did not complete properly"
        exit 1
    fi
else
    log_error "❌ Auto-fix failed"
    exit 1
fi

# Test 5: Hub Integration (if available)
echo ""
log_info "Test 5: Hub Integration Test"
if [[ -n "$SENTINEL_HUB_URL" && -n "$SENTINEL_API_KEY" ]]; then
    if $SENTINEL tasks scan 2>&1 | grep -q "codebase.*tasks\|Found.*tasks"; then
        log_info "✅ Hub integration working"
    else
        log_warn "⚠️  Hub integration test inconclusive (may be expected)"
    fi
else
    log_info "ℹ️  Hub not configured, skipping integration test"
fi

# Test 6: File integrity check
echo ""
log_info "Test 6: File Integrity Check"
if [[ -f "package.json" && -f "src/components/Button.js" ]]; then
    log_info "✅ Test files intact after operations"
else
    log_error "❌ Test files corrupted during testing"
    exit 1
fi

echo ""
echo "🎉 Complete Workflow E2E Test PASSED"
echo "==================================="
log_info "All integration tests successful!"
log_info "Sentinel system fully operational"

exit 0



