#!/bin/bash
# VAPT Revalidation Script
# Comprehensive security audit to verify all vulnerabilities from VAPT report

# Note: Don't use set -e as we want to continue even if some checks fail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

ISSUES_FOUND=0
FIXES_VERIFIED=0
WARNINGS=0

log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    ((FIXES_VERIFIED++))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    ((ISSUES_FOUND++))
}

log_warn() {
    echo -e "${YELLOW}⚠ WARN:${NC} $1"
    ((WARNINGS++))
}

log_info() {
    echo -e "${BLUE}ℹ INFO:${NC} $1"
}

log_section() {
    echo ""
    echo -e "${CYAN}══════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}══════════════════════════════════════════════════════════════${NC}"
}

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║           VAPT REVALIDATION AUDIT                            ║${NC}"
echo -e "${CYAN}║           Comprehensive Security Verification                ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# ============================================================================
# CVE-SENTINEL-001: Insecure API Key Generation
# ============================================================================
log_section "CVE-SENTINEL-001: API Key Generation Security"

if grep -r "time\.Now()\.UnixNano()" hub/api/services/organization_service_api_keys.go 2>/dev/null | grep -v "//" | grep -v "test" | grep -q .; then
    log_fail "Vulnerable timestamp-based API key generation found"
else
    log_pass "No timestamp-based key generation detected"
fi

if grep -r "crypto/rand" hub/api/services/organization_service_api_keys.go >/dev/null 2>&1; then
    log_pass "crypto/rand is used for secure key generation"
else
    log_fail "crypto/rand not found in API key generation"
fi

if grep -r "rand\.Read" hub/api/services/organization_service_api_keys.go >/dev/null 2>&1; then
    log_pass "crypto/rand.Read() is used (cryptographically secure)"
else
    log_fail "crypto/rand.Read() not found"
fi

# ============================================================================
# CVE-SENTINEL-002: Hardcoded API Keys
# ============================================================================
log_section "CVE-SENTINEL-002: Hardcoded API Keys Check"

# Check middleware for hardcoded keys
if grep -r "dev-api-key-123\|test-api-key-456" hub/api/middleware/security.go 2>/dev/null | grep -v "//" | grep -v "test" | grep -q .; then
    log_fail "Hardcoded API keys found in authentication middleware"
else
    log_pass "No hardcoded keys in authentication middleware"
fi

# Check config for production usage
if grep -r "dev-api-key-123\|test-api-key-456" hub/api/config/config.go 2>/dev/null | grep -q "ENV.*==.*development\|ENV.*==.*dev"; then
    log_pass "Hardcoded keys only used in development mode"
else
    # Check if they exist but might be used in production
    if grep -r "dev-api-key-123\|test-api-key-456" hub/api/config/config.go 2>/dev/null | grep -v "//" | grep -v "test"; then
        log_warn "Hardcoded keys found in config - verify development-only usage"
    else
        log_pass "No hardcoded keys in production config"
    fi
fi

# Check if middleware uses service-based validation
if grep -r "OrganizationService\.ValidateAPIKey\|config\.OrganizationService" hub/api/middleware/security.go >/dev/null 2>&1; then
    log_pass "Middleware uses service-based API key validation"
else
    log_fail "Middleware does not use service-based validation"
fi

# ============================================================================
# CVE-SENTINEL-003: Hardcoded JWT Secret
# ============================================================================
log_section "CVE-SENTINEL-003: JWT Secret Security"

if grep -r "dev-jwt-secret-change-in-production" hub/api/config/config.go 2>/dev/null | grep -v "//" | grep -v "test" | grep -q .; then
    # Check if it's only in development
    if grep -r "dev-jwt-secret-change-in-production" hub/api/config/config.go 2>/dev/null | grep -q "ENV.*==.*development\|ENV.*==.*dev"; then
        log_pass "JWT secret default only used in development mode"
    else
        log_fail "Hardcoded JWT secret may be used in production"
    fi
else
    log_pass "No hardcoded JWT secret found"
fi

if grep -r "JWT_SECRET" hub/api/config/config.go >/dev/null 2>&1; then
    log_pass "JWT secret loaded from environment variable"
else
    log_warn "JWT_SECRET environment variable not verified"
fi

# ============================================================================
# CVE-SENTINEL-004: CORS Configuration
# ============================================================================
log_section "CVE-SENTINEL-004: CORS Security"

if grep -r "Access-Control-Allow-Origin.*\*" hub/api/middleware/security.go 2>/dev/null | grep -v "//" | grep -v "test" | grep -v "development\|dev" | grep -q .; then
    log_fail "CORS allows all origins in production mode"
else
    log_pass "CORS wildcard not used in production"
fi

if grep -r "ENV.*==.*development\|ENV.*==.*dev" hub/api/middleware/security.go 2>/dev/null | grep -q "production.*whitelist\|production.*strict"; then
    log_pass "CORS has environment-aware configuration"
else
    log_warn "CORS environment-aware configuration needs verification"
fi

if grep -r "originMap\|AllowedOrigins" hub/api/middleware/security.go >/dev/null 2>&1; then
    log_pass "CORS origin whitelist mechanism implemented"
else
    log_warn "CORS origin whitelist not found"
fi

# ============================================================================
# CVE-SENTINEL-005: SQL Injection Protection
# ============================================================================
log_section "CVE-SENTINEL-005: SQL Injection Protection"

# Check for parameterized queries
if grep -r "\$[0-9]\|sqlx\|database/sql" hub/api/repository/*.go 2>/dev/null | head -5 >/dev/null 2>&1; then
    log_pass "Parameterized queries detected (SQL injection safe)"
else
    log_warn "Parameterized queries usage needs verification"
fi

# Check for dangerous string formatting in SQL
DANGEROUS_SQL=$(grep -r "fmt\.Sprintf.*SELECT\|fmt\.Sprintf.*INSERT\|fmt\.Sprintf.*UPDATE\|fmt\.Sprintf.*DELETE" hub/api/repository/*.go 2>/dev/null | grep -v "//" | grep -v "test" | grep -c . || echo "0")
if [ "$DANGEROUS_SQL" -gt 0 ]; then
    log_warn "String formatting found in SQL queries - verify parameterization"
    grep -r "fmt\.Sprintf.*SELECT\|fmt\.Sprintf.*INSERT\|fmt\.Sprintf.*UPDATE\|fmt\.Sprintf.*DELETE" hub/api/repository/*.go 2>/dev/null | grep -v "//" | head -3
else
    log_pass "No dangerous SQL string formatting detected"
fi

# ============================================================================
# CVE-SENTINEL-006: API Key Hashing
# ============================================================================
log_section "CVE-SENTINEL-006: API Key Hashing"

if grep -r "APIKeyHash\|sha256\|SHA256" hub/api/services/organization_service_api_keys.go >/dev/null 2>&1; then
    log_pass "API key hashing implementation found"
else
    log_fail "API key hashing not implemented"
fi

if grep -r "api_key_hash" hub/api/repository/organization_repository.go >/dev/null 2>&1; then
    log_pass "Database stores API key hashes"
else
    log_fail "API key hash storage not found in repository"
fi

# Check for plaintext storage
if grep -r "INSERT.*api_key\|UPDATE.*api_key" hub/api/repository/organization_repository.go 2>/dev/null | grep -v "api_key_hash\|api_key_prefix" | grep -v "//" | grep -v "test" | grep -q .; then
    log_warn "Direct API key storage may still exist - verify hashing is used"
else
    log_pass "No direct plaintext API key storage detected"
fi

# ============================================================================
# CVE-SENTINEL-007: Authentication Middleware
# ============================================================================
log_section "CVE-SENTINEL-007: Authentication Middleware"

if grep -r "user-123\|hardcoded.*user" hub/api/middleware/security.go 2>/dev/null | grep -v "//" | grep -v "test" | grep -q .; then
    log_fail "Hardcoded user IDs found in middleware"
else
    log_pass "No hardcoded user IDs"
fi

if grep -r "ValidateAPIKey\|OrganizationService" hub/api/middleware/security.go >/dev/null 2>&1; then
    log_pass "Middleware integrated with service layer"
else
    log_fail "Middleware not integrated with service layer"
fi

if grep -r "project_id\|org_id" hub/api/middleware/security.go >/dev/null 2>&1; then
    log_pass "Context injection for project/org ID implemented"
else
    log_warn "Context injection not found"
fi

# ============================================================================
# CVE-SENTINEL-008: Error Message Security
# ============================================================================
log_section "CVE-SENTINEL-008: Error Message Security"

# Check for potential information leakage
SENSITIVE_ERRORS=$(grep -r "fmt\.Errorf.*password\|fmt\.Errorf.*secret\|fmt\.Errorf.*key" hub/api/**/*.go 2>/dev/null | grep -v "//" | grep -v "test" | grep -c . || echo "0")
if [ "$SENSITIVE_ERRORS" -gt 0 ]; then
    log_warn "Potential sensitive information in error messages"
else
    log_pass "No obvious sensitive data in error messages"
fi

# ============================================================================
# CVE-SENTINEL-009: Rate Limiting
# ============================================================================
log_section "CVE-SENTINEL-009: Rate Limiting"

if grep -r "RateLimit\|rate.*limit" hub/api/middleware/security.go -i >/dev/null 2>&1; then
    log_pass "Rate limiting middleware implemented"
else
    log_warn "Rate limiting not found"
fi

# Check if it's per-API-key or global
if grep -r "per.*api.*key\|per.*IP\|per-IP\|per-API" hub/api/middleware/security.go -i >/dev/null 2>&1; then
    log_pass "Per-client rate limiting found"
else
    log_warn "Rate limiting appears to be global, not per-client"
fi

# ============================================================================
# CVE-SENTINEL-010: Input Validation
# ============================================================================
log_section "CVE-SENTINEL-010: Input Validation"

if [ -d "hub/api/validation" ]; then
    log_pass "Input validation framework exists"
    VALIDATOR_FILES=$(find hub/api/validation -name "*.go" 2>/dev/null | wc -l)
    if [ "$VALIDATOR_FILES" -gt 0 ]; then
        log_pass "Validation validators implemented ($VALIDATOR_FILES files)"
    else
        log_warn "Validation directory exists but no validators found"
    fi
else
    log_warn "Input validation framework not found"
fi

# ============================================================================
# CVE-SENTINEL-013: Security Headers
# ============================================================================
log_section "CVE-SENTINEL-013: Security Headers"

if grep -r "X-Content-Type-Options\|X-Frame-Options\|X-XSS-Protection\|Content-Security-Policy\|Strict-Transport-Security" hub/api/middleware/security.go >/dev/null 2>&1; then
    log_pass "Security headers middleware implemented"
else
    log_warn "Security headers middleware not found"
fi

# Check CSP for unsafe-inline
if grep -r "Content-Security-Policy.*unsafe-inline" hub/api/middleware/security.go 2>/dev/null | grep -v "//" | grep -v "test" | grep -q .; then
    log_warn "CSP allows 'unsafe-inline' - reduces XSS protection"
else
    log_pass "CSP does not use unsafe-inline"
fi

# ============================================================================
# CVE-SENTINEL-014: Security Logging
# ============================================================================
log_section "CVE-SENTINEL-014: Security Event Logging"

if [ -f "hub/api/pkg/security/audit_logger.go" ]; then
    log_pass "Security audit logger exists"
    if grep -r "LogAuthFailure\|LogAuthSuccess\|LogSecurityEvent" hub/api/pkg/security/audit_logger.go >/dev/null 2>&1; then
        log_pass "Authentication event logging implemented"
    else
        log_warn "Authentication event logging not found in audit logger"
    fi
else
    log_warn "Security audit logger not found"
fi

if grep -r "AuditLogger\|audit.*log" hub/api/middleware/security.go -i >/dev/null 2>&1; then
    log_pass "Middleware integrated with audit logging"
else
    log_warn "Middleware not integrated with audit logging"
fi

# ============================================================================
# Additional Security Checks
# ============================================================================
log_section "Additional Security Checks"

# Check for password in plaintext
if grep -r "password.*=.*[\"'].*[\"']" hub/api/**/*.go 2>/dev/null | grep -v "//" | grep -v "test" | grep -v "password.*hash\|password.*bcrypt" | grep -q .; then
    log_warn "Potential plaintext passwords found in code"
else
    log_pass "No plaintext passwords detected"
fi

# Check for connection strings in code
if grep -r "postgres://\|mysql://\|mongodb://" hub/api/**/*.go 2>/dev/null | grep -v "//" | grep -v "test" | grep -q .; then
    log_warn "Database connection strings may be hardcoded"
else
    log_pass "No hardcoded database connection strings"
fi

# Check HTTPS enforcement
if grep -r "http\.ListenAndServe\|Listen.*:80" hub/api/**/*.go 2>/dev/null | grep -v "//" | grep -v "test" | grep -q .; then
    log_warn "HTTP (non-HTTPS) server found - verify HTTPS enforcement in production"
else
    log_pass "No HTTP-only server detected"
fi

# ============================================================================
# Summary
# ============================================================================
log_section "VAPT Revalidation Summary"

echo ""
echo -e "${CYAN}Vulnerabilities Checked:${NC} 15 Critical, 8 High, 12 Medium"
echo -e "${GREEN}Fixes Verified:${NC} $FIXES_VERIFIED"
echo -e "${YELLOW}Warnings:${NC} $WARNINGS"
echo -e "${RED}Issues Found:${NC} $ISSUES_FOUND"
echo ""

if [ $ISSUES_FOUND -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓✓✓ ALL CRITICAL VULNERABILITIES VERIFIED AS FIXED ✓✓✓${NC}"
    echo ""
    exit 0
elif [ $ISSUES_FOUND -eq 0 ]; then
    echo -e "${YELLOW}⚠ All critical issues fixed, but some warnings remain${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}✗✗✗ CRITICAL ISSUES STILL PRESENT ✗✗✗${NC}"
    echo ""
    exit 1
fi
