#!/bin/bash
# Deployment Readiness Validation Script
# Run from project root: ./validate_deployment_readiness.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

echo -e "${PURPLE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${PURPLE}🔍 SENTINEL HUB DEPLOYMENT READINESS VALIDATION${NC}"
echo -e "${PURPLE}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Initialize counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0

check_result() {
    local name="$1"
    local status="$2"
    local details="$3"

    ((TOTAL_CHECKS++))

    case $status in
        "PASS")
            echo -e "${GREEN}✅ $name${NC}"
            ((PASSED_CHECKS++))
            ;;
        "FAIL")
            echo -e "${RED}❌ $name${NC}"
            if [ -n "$details" ]; then
                echo -e "${RED}   └─ $details${NC}"
            fi
            ((FAILED_CHECKS++))
            ;;
        "WARN")
            echo -e "${YELLOW}⚠️  $name${NC}"
            if [ -n "$details" ]; then
                echo -e "${YELLOW}   └─ $details${NC}"
            fi
            ((WARNINGS++))
            ;;
    esac
}

echo -e "${BLUE}📁 ARCHITECTURAL COMPLIANCE${NC}"
echo "────────────────────────────────"

# Check main.go size
MAIN_LINES=$(wc -l < hub/api/main.go)
if [ "$MAIN_LINES" -le 50 ]; then
    check_result "Entry Point Size" "PASS" "$MAIN_LINES lines (≤50 limit)"
else
    check_result "Entry Point Size" "FAIL" "$MAIN_LINES lines (>50 limit)"
fi

# Check for files in wrong package
MAIN_FILES=$(find hub/api -name "*.go" -exec grep -l "^package main$" {} \; | grep -v "/main.go$" | wc -l)
if [ "$MAIN_FILES" -eq 0 ]; then
    check_result "Package Structure" "PASS" "No files in wrong package"
else
    check_result "Package Structure" "FAIL" "$MAIN_FILES files still in 'package main'"
fi

# Check file sizes
LARGE_FILES=$(find hub/api -name "*.go" -exec wc -l {} \; | awk '$1 > 500 {print $2 ": " $1 " lines"}' | wc -l)
if [ "$LARGE_FILES" -eq 0 ]; then
    check_result "File Size Limits" "PASS" "All files within limits"
else
    check_result "File Size Limits" "FAIL" "$LARGE_FILES files exceed 500-line limit"
fi

echo ""
echo -e "${BLUE}🔨 BUILD SYSTEM${NC}"
echo "────────────────────────────────"

# Check if Go modules are valid
if cd hub/api && go mod tidy >/dev/null 2>&1; then
    check_result "Go Modules" "PASS" "Modules are valid"
else
    check_result "Go Modules" "FAIL" "Go modules have issues"
fi

# Check if code compiles
if cd hub/api && go build -o /tmp/sentinel-test . >/dev/null 2>&1; then
    check_result "Build Success" "PASS" "Code compiles successfully"
else
    BUILD_ERROR=$(cd hub/api && go build -o /tmp/sentinel-test . 2>&1 | head -1)
    check_result "Build Success" "FAIL" "$BUILD_ERROR"
fi

echo ""
echo -e "${BLUE}🗄️ DATABASE${NC}"
echo "────────────────────────────────"

# Check if migrations directory has files
MIGRATION_COUNT=$(find hub/migrations -name "*.sql" 2>/dev/null | wc -l)
if [ "$MIGRATION_COUNT" -gt 0 ]; then
    check_result "Database Migrations" "PASS" "$MIGRATION_COUNT migration files found"
else
    check_result "Database Migrations" "FAIL" "No migration files found"
fi

# Check init script
if [ -f "hub/init-test-db.sql" ] && [ -s "hub/init-test-db.sql" ]; then
    check_result "Database Init Script" "PASS" "Init script exists and not empty"
else
    check_result "Database Init Script" "FAIL" "Init script missing or empty"
fi

echo ""
echo -e "${BLUE}🐳 DEPLOYMENT${NC}"
echo "────────────────────────────────"

# Check Docker files
if [ -f "hub/api/Dockerfile" ]; then
    check_result "Dockerfile" "PASS" "API Dockerfile exists"
else
    check_result "Dockerfile" "FAIL" "API Dockerfile missing"
fi

if [ -f "hub/docker-compose.yml" ]; then
    check_result "Docker Compose" "PASS" "Docker Compose config exists"
else
    check_result "Docker Compose" "FAIL" "Docker Compose config missing"
fi

# Check environment files
if [ -f "hub/.env" ]; then
    check_result "Environment Config" "PASS" "Environment file exists"
else
    check_result "Environment Config" "WARN" "Environment file missing (may be gitignored)"
fi

ENV_EXAMPLE_EXISTS=false
if [ -f "hub/.env.example" ] || [ -f "hub/env.example" ]; then
    ENV_EXAMPLE_EXISTS=true
fi

if [ "$ENV_EXAMPLE_EXISTS" = true ]; then
    check_result "Environment Example" "PASS" "Environment example file exists"
else
    check_result "Environment Example" "FAIL" "No environment example file found"
fi

echo ""
echo -e "${BLUE}🔒 SECURITY${NC}"
echo "────────────────────────────────"

# Check for security-related files
if [ -d "hub/api/middleware" ] && [ -f "hub/api/middleware/security.go" ]; then
    check_result "Security Middleware" "PASS" "Security middleware exists"
else
    check_result "Security Middleware" "WARN" "Security middleware may be incomplete"
fi

# Check for authentication
if grep -r "JWT\|jwt\|auth\|Auth" hub/api/handlers/ >/dev/null 2>&1; then
    check_result "Authentication" "PASS" "Authentication code found"
else
    check_result "Authentication" "FAIL" "No authentication implementation found"
fi

echo ""
echo -e "${BLUE}🧪 TESTING${NC}"
echo "────────────────────────────────"

# Check test files
TEST_COUNT=$(find hub/api -name "*_test.go" | wc -l)
if [ "$TEST_COUNT" -gt 0 ]; then
    check_result "Test Files" "PASS" "$TEST_COUNT test files found"
else
    check_result "Test Files" "FAIL" "No test files found"
fi

# Check if tests run (if build works)
if [ -f "/tmp/sentinel-test" ]; then
    if cd hub/api && go test -run=^$ ./... >/dev/null 2>&1; then
        check_result "Test Execution" "PASS" "Tests can be executed"
    else
        check_result "Test Execution" "WARN" "Tests may have issues"
    fi
else
    check_result "Test Execution" "FAIL" "Cannot test due to build failure"
fi

echo ""
echo -e "${BLUE}📚 DOCUMENTATION${NC}"
echo "────────────────────────────────"

# Check documentation files
if [ -f "docs/external/CODING_STANDARDS.md" ]; then
    check_result "Coding Standards" "PASS" "CODING_STANDARDS.md exists"
else
    check_result "Coding Standards" "FAIL" "CODING_STANDARDS.md missing"
fi

if [ -f "README.md" ]; then
    check_result "README" "PASS" "README.md exists"
else
    check_result "README" "WARN" "README.md missing"
fi

if [ -f "hub/README.md" ]; then
    check_result "Hub README" "PASS" "Hub README exists"
else
    check_result "Hub README" "WARN" "Hub README missing"
fi

echo ""
echo -e "${PURPLE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${PURPLE}📊 VALIDATION SUMMARY${NC}"
echo -e "${PURPLE}═══════════════════════════════════════════════════════════════${NC}"

SUCCESS_RATE=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
echo -e "${BLUE}Total Checks:${NC} $TOTAL_CHECKS"
echo -e "${GREEN}Passed:${NC} $PASSED_CHECKS"
echo -e "${RED}Failed:${NC} $FAILED_CHECKS"
echo -e "${YELLOW}Warnings:${NC} $WARNINGS"
echo -e "${BLUE}Success Rate:${NC} ${SUCCESS_RATE}%"

echo ""

if [ $FAILED_CHECKS -eq 0 ]; then
    echo -e "${GREEN}🎉 ALL CHECKS PASSED - DEPLOYMENT READY${NC}"
    exit 0
elif [ $SUCCESS_RATE -ge 80 ]; then
    echo -e "${YELLOW}⚠️ MOSTLY READY - MINOR ISSUES TO FIX${NC}"
    exit 1
else
    echo -e "${RED}🚫 NOT READY FOR DEPLOYMENT - CRITICAL ISSUES${NC}"
    echo ""
    echo -e "${RED}📋 IMMEDIATE ACTION REQUIRED:${NC}"
    echo -e "${RED}1. Fix build failures${NC}"
    echo -e "${RED}2. Complete package restructuring${NC}"
    echo -e "${RED}3. Add database schema${NC}"
    echo -e "${RED}4. Implement missing security features${NC}"
    exit 1
fi