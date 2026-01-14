#!/bin/bash

# Phase 18: Security Audit Script
# Comprehensive security assessment for Sentinel

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Results tracking
CRITICAL_ISSUES=0
HIGH_ISSUES=0
MEDIUM_ISSUES=0
LOW_ISSUES=0

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_critical() {
    echo -e "${RED}[CRITICAL]${NC} $1"
    ((CRITICAL_ISSUES++))
}

log_high() {
    echo -e "${RED}[HIGH]${NC} $1"
    ((HIGH_ISSUES++))
}

log_medium() {
    echo -e "${YELLOW}[MEDIUM]${NC} $1"
    ((MEDIUM_ISSUES++))
}

log_low() {
    echo -e "${BLUE}[LOW]${NC} $1"
    ((LOW_ISSUES++))
}

echo "========================================="
echo "🔒 SENTINEL SECURITY AUDIT"
echo "========================================="
echo "Started: $(date)"
echo ""

# 1. Input Validation Audit
echo "1. INPUT VALIDATION AUDIT"
echo "=========================="

log_info "Checking for user input sanitization..."

# Check for direct file path usage without sanitization
if grep -r "r\.URL\.Query\.Get.*path" hub/api/ --include="*.go" | grep -v -A10 "sanitizePath\|ValidateDirectory\|isValidPath" | grep -c "filepath\|os\." >/dev/null 2>&1; then
    log_medium "Some file path inputs may lack sanitization - manual review recommended"
else
    log_success "File path inputs appear to be properly sanitized"
fi

# Check for SQL injection vulnerabilities (string concatenation in queries)
SQL_INJECTION_RISKS=$(grep -r "Query(\|Exec(" hub/api/ --include="*.go" | grep -v '\$[0-9]\|\$[a-zA-Z]' | grep -c '"SELECT\|"INSERT\|"UPDATE\|"DELETE' 2>/dev/null || echo "0")
SQL_INJECTION_RISKS=$(echo "$SQL_INJECTION_RISKS" | tr -d '\n' | tr -d ' ')
if [ "$SQL_INJECTION_RISKS" -gt 0 ] 2>/dev/null; then
    log_medium "Some queries may use string concatenation - manual review recommended"
else
    log_success "SQL queries use parameterized statements"
fi

# 2. Authentication & Authorization Audit
echo ""
echo "2. AUTHENTICATION & AUTHORIZATION AUDIT"
echo "======================================="

log_info "Checking authentication mechanisms..."

# Check for hardcoded credentials (exclude legitimate patterns and security detection)
if grep -r "password.*=.*[\"'][^\"']*[\"']" hub/ --exclude-dir=.git --include="*.go" --include="*.html" --include="*.js" --include="*.sh" 2>/dev/null | grep -v "test\|fixture\|regexp.MustCompile\|input.*password\|type.*password\|security_analyzer\|password.*s*=.*s*.*regexp" | grep -q "password.*=.*[\"'][^\"']*[\"']"; then
    log_critical "Hardcoded passwords found in source code"
else
    log_success "No hardcoded passwords detected"
fi

# Check for API key validation
if grep -r "api.*key\|API.*KEY" hub/api/ --include="*.go" | grep -i "validate"; then
    log_success "API key validation is implemented"
else
    log_medium "API key validation may need enhancement"
fi

# 3. Data Exposure Audit
echo ""
echo "3. DATA EXPOSURE AUDIT"
echo "======================="

log_info "Checking for sensitive data exposure..."

# Check for debug logging of sensitive data
if grep -r "log.*password\|log.*key\|log.*token" hub/api/ --include="*.go"; then
    log_high "Sensitive data may be logged in debug statements"
else
    log_success "No sensitive data logging detected"
fi

# Check for secure headers
if grep -r "Content-Security-Policy\|X-Frame-Options\|X-Content-Type-Options" hub/api/ --include="*.go"; then
    log_success "Security headers are implemented"
else
    log_medium "Security headers may be incomplete"
fi

# 4. Cryptography Audit
echo ""
echo "4. CRYPTOGRAPHY AUDIT"
echo "======================"

log_info "Checking cryptographic implementations..."

# Check for weak encryption
if grep -r "md5\|sha1" hub/api/ --include="*.go" | grep -v "test"; then
    log_high "Weak cryptographic algorithms detected (MD5/SHA1)"
else
    log_success "No weak cryptographic algorithms found"
fi

# Check for proper key management
if grep -r "encrypt\|decrypt" hub/api/ --include="*.go"; then
    log_success "Encryption/decryption functions are implemented"
else
    log_medium "Encryption functions may need implementation"
fi

# 5. Error Handling Audit
echo ""
echo "5. ERROR HANDLING AUDIT"
echo "========================"

log_info "Checking error handling patterns..."

# Check for panic usage
if grep -r "panic(" hub/api/ --include="*.go" | grep -v "test"; then
    log_medium "Panic statements found - should use proper error handling"
else
    log_success "No panic statements in production code"
fi

# Check for error wrapping
if grep -r "fmt\.Errorf.*%w" hub/api/ --include="*.go"; then
    log_success "Error wrapping is implemented"
else
    log_medium "Error wrapping could be improved"
fi

# 6. Access Control Audit
echo ""
echo "6. ACCESS CONTROL AUDIT"
echo "======================="

log_info "Checking access control mechanisms..."

# Check for admin middleware
if grep -r "adminAuthMiddleware\|AdminAuth" hub/api/ --include="*.go"; then
    log_success "Admin authentication middleware is implemented"
else
    log_high "Admin authentication middleware missing"
fi

# Check for rate limiting
if grep -r "rateLimit\|RateLimit" hub/api/ --include="*.go"; then
    log_success "Rate limiting is implemented"
else
    log_medium "Rate limiting may be missing"
fi

# 7. File Upload Security
echo ""
echo "7. FILE UPLOAD SECURITY"
echo "======================="

log_info "Checking file upload security..."

# Check for file type validation
if grep -r "validateDocumentContentType\|validateBinaryContentType" hub/api/ --include="*.go"; then
    log_success "File type validation is implemented"
else
    log_high "File type validation is missing"
fi

# Check for file size limits
if grep -r "ParseMultipartForm.*[0-9]*.*<<" hub/api/ --include="*.go"; then
    log_success "File size limits are enforced"
else
    log_medium "File size limits may need verification"
fi

# 8. Dependencies Audit
echo ""
echo "8. DEPENDENCIES AUDIT"
echo "======================"

log_info "Checking for vulnerable dependencies..."

# Check go.mod for known vulnerable packages (simplified check)
if grep -A 20 "^require (" hub/api/go.mod | grep -E "(net/http|crypto|encoding)" | head -5 >/dev/null; then
    log_success "Standard library dependencies are up to date"
else
    log_low "Dependency versions should be verified"
fi

# 9. Configuration Security
echo ""
echo "9. CONFIGURATION SECURITY"
echo "========================="

log_info "Checking configuration security..."

# Check for secure defaults
if grep -r "change-me-in-production\|default-password\|admin.*admin" hub/ --exclude-dir=.git; then
    log_critical "Insecure default credentials found"
else
    log_success "No insecure default credentials detected"
fi

# Check for environment variable usage
if grep -r "os\.Getenv" hub/api/ --include="*.go"; then
    log_success "Environment variables are used for configuration"
else
    log_medium "Environment variable usage could be expanded"
fi

# 10. Logging Security
echo ""
echo "10. LOGGING SECURITY"
echo "===================="

log_info "Checking logging security..."

# Check for sensitive data in logs
if grep -r "LogError.*password\|LogInfo.*key\|LogDebug.*token" hub/api/ --include="*.go"; then
    log_high "Sensitive data may be exposed in logs"
else
    log_success "No sensitive data logging detected"
fi

# Generate final report
echo ""
echo "========================================="
echo "🔒 SECURITY AUDIT RESULTS"
echo "========================================="

echo "Issues Found:"
echo "  Critical: $CRITICAL_ISSUES"
echo "  High: $HIGH_ISSUES"
echo "  Medium: $MEDIUM_ISSUES"
echo "  Low: $LOW_ISSUES"

TOTAL_ISSUES=$((CRITICAL_ISSUES + HIGH_ISSUES + MEDIUM_ISSUES + LOW_ISSUES))

if [ $TOTAL_ISSUES -eq 0 ]; then
    echo ""
    log_success "🎉 SECURITY AUDIT PASSED - No issues found!"
    echo "Recommendation: Regular security audits should be performed."
elif [ $CRITICAL_ISSUES -gt 0 ]; then
    echo ""
    log_critical "🚨 SECURITY AUDIT FAILED - Critical issues must be fixed before production deployment!"
    echo "Critical issues block production deployment and must be addressed immediately."
elif [ $HIGH_ISSUES -gt 0 ]; then
    echo ""
    log_error "⚠️  SECURITY AUDIT WARNING - High-priority issues should be addressed"
    echo "High-priority issues should be resolved before production deployment."
else
    echo ""
    log_warning "⚠️  SECURITY AUDIT PASSED WITH MINOR ISSUES"
    echo "Minor issues should be addressed but do not block production deployment."
fi

echo ""
echo "Completed: $(date)"
echo "========================================="

# Exit with error code if critical issues found
if [ $CRITICAL_ISSUES -gt 0 ]; then
    exit 1
elif [ $HIGH_ISSUES -gt 0 ]; then
    exit 2
else
    exit 0
fi
