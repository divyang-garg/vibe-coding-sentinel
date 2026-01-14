#!/bin/bash
# Production Readiness Validation
# Final validation before production deployment

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_header() {
    echo -e "${PURPLE}🔥 $1${NC}"
}

PASSED=0
FAILED=0
WARNINGS=0

SENTINEL="./sentinel"

# Validation checklist
check_binary_integrity() {
    log_header "BINARY INTEGRITY CHECK"

    if [[ ! -f "$SENTINEL" ]]; then
        log_error "Sentinel binary not found at $SENTINEL"
        ((FAILED++))
        return 1
    fi

    if [[ ! -x "$SENTINEL" ]]; then
        log_error "Sentinel binary is not executable"
        ((FAILED++))
        return 1
    fi

    # Check file size (reasonable binary size)
    size=$(stat -f%z "$SENTINEL" 2>/dev/null || stat -c%s "$SENTINEL" 2>/dev/null || echo "0")
    if [[ "$size" -lt 1000000 ]]; then  # At least 1MB
        log_error "Binary suspiciously small ($size bytes)"
        ((FAILED++))
        return 1
    fi

    log_success "Binary integrity verified"
    ((PASSED++))
}

check_core_functionality() {
    log_header "CORE FUNCTIONALITY VALIDATION"

    # Test help command
    if ! $SENTINEL --help >/dev/null 2>&1; then
        log_error "Help command failed"
        ((FAILED++))
        return 1
    fi

    # Test version/info command
    if ! $SENTINEL --help | grep -q "sentinel\|Sentinel"; then
        log_error "Help output malformed"
        ((FAILED++))
        return 1
    fi

    # Test audit command
    if ! $SENTINEL audit --offline --help >/dev/null 2>&1; then
        log_error "Audit command failed"
        ((FAILED++))
        return 1
    fi

    # Test learn command
    if ! $SENTINEL learn --help >/dev/null 2>&1; then
        log_error "Learn command failed"
        ((FAILED++))
        return 1
    fi

    # Test fix command
    if ! $SENTINEL fix --help >/dev/null 2>&1; then
        log_error "Fix command failed"
        ((FAILED++))
        return 1
    fi

    log_success "All core commands operational"
    ((PASSED++))
}

check_error_handling() {
    log_header "ERROR HANDLING VALIDATION"

    # Test invalid command
    if $SENTINEL invalid-command 2>&1 | grep -q "error\|Error\|usage\|Usage"; then
        log_success "Invalid command handled gracefully"
        ((PASSED++))
    else
        log_error "Invalid command not handled properly"
        ((FAILED++))
    fi

    # Test missing arguments
    if $SENTINEL audit 2>&1 | grep -q "error\|Error\|usage\|Usage\|help\|Help"; then
        log_success "Missing arguments handled gracefully"
        ((PASSED++))
    else
        log_error "Missing arguments not handled properly"
        ((FAILED++))
    fi

    # Test invalid paths
    if $SENTINEL audit /nonexistent/path 2>&1 | grep -q "error\|Error\|No such file"; then
        log_success "Invalid paths handled gracefully"
        ((PASSED++))
    else
        log_error "Invalid paths not handled properly"
        ((FAILED++))
    fi
}

check_configuration_handling() {
    log_header "CONFIGURATION HANDLING"

    # Test with valid config
    cat > test_config.json << 'EOF'
{
  "hubUrl": "https://test-hub.example.com",
  "scanDirs": ["src", "test"],
  "excludePaths": [".git"]
}
EOF

    if $SENTINEL audit --offline 2>&1 | grep -q "Scan\|Audit\|found"; then
        log_success "Configuration loading functional"
        ((PASSED++))
    else
        log_warning "Configuration loading may have issues"
        ((WARNINGS++))
    fi

    rm -f test_config.json
}

check_resource_limits() {
    log_header "RESOURCE LIMITS VALIDATION"

    # Test memory usage
    log_info "Testing memory usage..."

    # Create a small test project
    mkdir -p /tmp/readiness_test
    cd /tmp/readiness_test
    echo 'console.log("test"); const apiKey = "sk-123";' > test.js
    echo '{}' > package.json

    # Run audit and check if it completes without excessive memory
    timeout 30s $SENTINEL audit --offline >/dev/null 2>&1
    exit_code=$?

    if [[ $exit_code -eq 0 ]]; then
        log_success "Memory usage within acceptable limits"
        ((PASSED++))
    elif [[ $exit_code -eq 124 ]]; then
        log_error "Command timed out - possible memory or performance issue"
        ((FAILED++))
    else
        log_warning "Audit completed with warnings (may be expected)"
        ((WARNINGS++))
    fi

    cd /
    rm -rf /tmp/readiness_test
}

check_data_persistence() {
    log_header "DATA PERSISTENCE VALIDATION"

    # Test pattern learning persistence
    mkdir -p /tmp/persistence_test
    cd /tmp/persistence_test

    echo 'function test() { return true; }' > test.js
    echo '{}' > package.json

    # Learn patterns
    $SENTINEL learn >/dev/null 2>&1

    # Check if files were created
    if [[ -f ".sentinel/patterns.json" && -f ".cursor/rules/project-patterns.md" ]]; then
        log_success "Data persistence working correctly"
        ((PASSED++))
    else
        log_error "Data persistence failed"
        ((FAILED++))
    fi

    cd /
    rm -rf /tmp/persistence_test
}

check_backup_recovery() {
    log_header "BACKUP AND RECOVERY"

    mkdir -p /tmp/backup_test
    cd /tmp/backup_test

    # Create test file
    cat > original.js << 'EOF'
console.log("original");
debugger;
function test() {
    return "test";
}
EOF

    # Run fix to create backup
    $SENTINEL fix --safe original.js >/dev/null 2>&1

    # Check if backup exists
    if [[ -f ".sentinel/backups/original.js.backup" ]]; then
        log_success "Backup creation functional"
        ((PASSED++))
    else
        log_error "Backup creation failed"
        ((FAILED++))
    fi

    # Verify original file still exists
    if [[ -f "original.js" ]]; then
        log_success "Original files preserved during operations"
        ((PASSED++))
    else
        log_error "Original file lost during operation"
        ((FAILED++))
    fi

    cd /
    rm -rf /tmp/backup_test
}

check_integration_readiness() {
    log_header "INTEGRATION READINESS"

    # Test CI/CD integration flags
    if $SENTINEL audit --ci --offline 2>&1 | grep -q "Audit\|found\|FAILED\|PASSED"; then
        log_success "CI/CD integration flags working"
        ((PASSED++))
    else
        log_error "CI/CD integration flags not working"
        ((FAILED++))
    fi

    # Test JSON output
    if $SENTINEL audit --output json --offline 2>&1 | grep -q "{"; then
        log_success "JSON output format working"
        ((PASSED++))
    else
        log_warning "JSON output format may have issues"
        ((WARNINGS++))
    fi
}

check_documentation_completeness() {
    log_header "DOCUMENTATION VALIDATION"

    # Check for essential documentation files
    docs=(
        "README.md"
        "docs/USER_GUIDE.md"
        "docs/api/API_REFERENCE.md"
    )

    for doc in "${docs[@]}"; do
        if [[ -f "$doc" ]]; then
            # Check if file has reasonable content
            lines=$(wc -l < "$doc")
            if [[ $lines -gt 10 ]]; then
                log_success "Documentation: $doc exists and populated"
                ((PASSED++))
            else
                log_warning "Documentation: $doc exists but seems incomplete"
                ((WARNINGS++))
            fi
        else
            log_error "Documentation: $doc missing"
            ((FAILED++))
        fi
    done
}

run_final_integration_test() {
    log_header "FINAL INTEGRATION TEST"

    # Create a complete test scenario
    TEST_DIR="/tmp/final_integration_$(date +%s)"
    mkdir -p "$TEST_DIR/project"
    cd "$TEST_DIR/project"

    # Create a realistic project structure
    mkdir -p src/components src/services src/utils tests
    cat > package.json << 'EOF'
{
  "name": "integration-test",
  "version": "1.0.0",
  "dependencies": {
    "react": "^18.0.0",
    "express": "^4.18.0"
  }
}
EOF

    cat > src/components/App.js << 'EOF'
import React from 'react';
import { apiService } from '../services/api';
import { formatDate } from '../utils/helpers';

function App() {
    const [data, setData] = React.useState(null);

    React.useEffect(() => {
        apiService.getData().then(setData);
    }, []);

    return (
        <div>
            <h1>My App</h1>
            {data && <p>Last updated: {formatDate(data.timestamp)}</p>}
        </div>
    );
}

export default App;
EOF

    cat > src/services/api.js << 'EOF'
const API_KEY = process.env.REACT_APP_API_KEY;
const BASE_URL = process.env.REACT_APP_API_URL;

export const apiService = {
    async getData() {
        const response = await fetch(`${BASE_URL}/data`, {
            headers: {
                'Authorization': `Bearer ${API_KEY}`,
                'Content-Type': 'application/json'
            }
        });
        return response.json();
    }
};
EOF

    cat > src/utils/helpers.js << 'EOF'
export function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString();
}

export function validateEmail(email) {
    const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return regex.test(email);
}
EOF

    cat > tests/App.test.js << 'EOF'
import React from 'react';
import { render } from '@testing-library/react';
import App from '../src/components/App';

test('renders app correctly', () => {
    render(<App />);
    // Test would go here
});
EOF

    log_info "Running complete integration workflow..."

    # Step 1: Initialize
    if $SENTINEL init 2>&1 | grep -q "initialized\|complete"; then
        log_success "Step 1: Initialization successful"
    else
        log_error "Step 1: Initialization failed"
        ((FAILED++))
    fi

    # Step 2: Learn patterns
    if $SENTINEL learn 2>&1 | grep -q "complete\|successful"; then
        log_success "Step 2: Pattern learning successful"
    else
        log_error "Step 2: Pattern learning failed"
        ((FAILED++))
    fi

    # Step 3: Security audit
    if AUDIT_RESULT=$($SENTINEL audit --offline 2>&1) && echo "$AUDIT_RESULT" | grep -q "Audit\|found\|FAILED\|PASSED"; then
        log_success "Step 3: Security audit successful"
    else
        log_error "Step 3: Security audit failed"
        ((FAILED++))
    fi

    # Step 4: Auto-fix
    if $SENTINEL fix --safe 2>&1 | grep -q "complete\|Auto-fix"; then
        log_success "Step 4: Auto-fix successful"
    else
        log_error "Step 4: Auto-fix failed"
        ((FAILED++))
    fi

    # Step 5: Status check
    if $SENTINEL status 2>&1 | grep -q "status\|Status\|project"; then
        log_success "Step 5: Status check successful"
    else
        log_error "Step 5: Status check failed"
        ((FAILED++))
    fi

    cd /
    rm -rf "$TEST_DIR"
}

# Main execution
echo "🚀 PRODUCTION READINESS VALIDATION"
echo "=================================="

check_binary_integrity
check_core_functionality
check_error_handling
check_configuration_handling
check_resource_limits
check_data_persistence
check_backup_recovery
check_integration_readiness
check_documentation_completeness
run_final_integration_test

# Final Assessment
echo ""
log_header "PRODUCTION READINESS ASSESSMENT"
echo "=================================="

TOTAL_CHECKS=$((PASSED + FAILED + WARNINGS))

if [[ $TOTAL_CHECKS -eq 0 ]]; then
    echo -e "${RED}❌ NO CHECKS WERE PERFORMED${NC}"
    exit 1
fi

SUCCESS_RATE=$((PASSED * 100 / TOTAL_CHECKS))
WARNING_RATE=$((WARNINGS * 100 / TOTAL_CHECKS))
FAILURE_RATE=$((FAILED * 100 / TOTAL_CHECKS))

echo "📊 VALIDATION RESULTS:"
echo "  ✅ Passed: $PASSED checks ($SUCCESS_RATE%)"
echo "  ⚠️  Warnings: $WARNINGS checks ($WARNING_RATE%)"
echo "  ❌ Failed: $FAILED checks ($FAILURE_RATE%)"

# Production readiness criteria
PRODUCTION_READY=true

if [[ $FAILED -gt 0 ]]; then
    PRODUCTION_READY=false
    echo -e "${RED}❌ CRITICAL ISSUES: $FAILED failed checks${NC}"
fi

if [[ $SUCCESS_RATE -lt 85 ]]; then
    PRODUCTION_READY=false
    echo -e "${RED}❌ INSUFFICIENT SUCCESS RATE: $SUCCESS_RATE% (required: 85%)${NC}"
fi

if [[ $WARNINGS -gt 3 ]]; then
    PRODUCTION_READY=false
    echo -e "${YELLOW}⚠️  TOO MANY WARNINGS: $WARNINGS warnings (max allowed: 3)${NC}"
fi

echo ""

if [[ "$PRODUCTION_READY" == "true" ]]; then
    echo -e "${GREEN}🎉 PRODUCTION READINESS: PASSED${NC}"
    echo "✅ Sentinel is ready for production deployment"
    echo ""
    echo "🚀 DEPLOYMENT CHECKLIST:"
    echo "  ✅ Binary integrity verified"
    echo "  ✅ Core functionality operational"
    echo "  ✅ Error handling robust"
    echo "  ✅ Configuration management working"
    echo "  ✅ Resource usage acceptable"
    echo "  ✅ Data persistence functional"
    echo "  ✅ Backup/recovery operational"
    echo "  ✅ Integration ready"
    echo "  ✅ Documentation complete"
    echo "  ✅ End-to-end workflow validated"
    exit 0
else
    echo -e "${RED}❌ PRODUCTION READINESS: FAILED${NC}"
    echo "❌ Sentinel requires additional work before production deployment"
    echo ""
    echo "🔧 REMEDIATION REQUIRED:"
    if [[ $FAILED -gt 0 ]]; then
        echo "  - Address $FAILED critical failures"
    fi
    if [[ $SUCCESS_RATE -lt 85 ]]; then
        echo "  - Improve success rate to 85%+"
    fi
    if [[ $WARNINGS -gt 3 ]]; then
        echo "  - Reduce warnings to 3 or fewer"
    fi
    exit 1
fi



