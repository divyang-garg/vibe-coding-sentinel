#!/bin/bash
# Run all Sentinel tests
# Usage: ./tests/run_all_tests.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           SENTINEL TEST SUITE                              ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Clean up lock file
rm -f /tmp/sentinel.lock

# Ensure binary exists
if [[ ! -f "./sentinel" ]]; then
    echo -e "${YELLOW}Building Sentinel...${NC}"
    ./synapsevibsentinel.sh
    echo ""
fi

# Track results
UNIT_SCANNING_RESULT=0
UNIT_PATTERNS_RESULT=0
UNIT_FIX_RESULT=0
UNIT_INGEST_RESULT=0
UNIT_KNOWLEDGE_RESULT=0
UNIT_TELEMETRY_RESULT=0
UNIT_AGENT_TELEMETRY_RESULT=0
UNIT_HUB_API_RESULT=0
UNIT_MCP_RESULT=0
UNIT_AST_ANALYSIS_RESULT=0
UNIT_VIBE_ACCURACY_RESULT=0
UNIT_VIBE_COMPARISON_RESULT=0
UNIT_VIBE_FALLBACK_RESULT=0
UNIT_VIBE_DEDUPLICATION_RESULT=0
UNIT_CROSS_FILE_ANALYSIS_RESULT=0
UNIT_BUSINESS_RULES_RESULT=0
UNIT_DETECTION_METRICS_RESULT=0
UNIT_SECURITY_ANALYSIS_RESULT=0
UNIT_FILE_SIZE_RESULT=0
UNIT_CACHE_RACE_RESULT=0
UNIT_RETRY_LOGIC_RESULT=0
UNIT_DB_TIMEOUT_RESULT=0
UNIT_ERROR_RECOVERY_RESULT=0
INTEGRATION_RESULT=0
INTEGRATION_HOOK_ERROR_RESULT=0
INTEGRATION_CACHE_INVALIDATION_RESULT=0

# Run scanning unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Scanning Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if ./tests/unit/scanning_test.sh; then
    UNIT_SCANNING_RESULT=0
    echo -e "${GREEN}Scanning unit tests passed!${NC}"
else
    UNIT_SCANNING_RESULT=1
    echo -e "${RED}Scanning unit tests failed!${NC}"
fi

echo ""

# Run pattern learning unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Pattern Learning Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if ./tests/unit/pattern_learning_test.sh; then
    UNIT_PATTERNS_RESULT=0
    echo -e "${GREEN}Pattern learning tests passed!${NC}"
else
    UNIT_PATTERNS_RESULT=1
    echo -e "${RED}Pattern learning tests failed!${NC}"
fi

echo ""

# Run auto-fix unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Auto-Fix Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if ./tests/unit/fix_test.sh; then
    UNIT_FIX_RESULT=0
    echo -e "${GREEN}Auto-fix tests passed!${NC}"
else
    UNIT_FIX_RESULT=1
    echo -e "${RED}Auto-fix tests failed!${NC}"
fi

echo ""

# Run document ingestion unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Document Ingestion Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if ./tests/unit/ingest_test.sh; then
    UNIT_INGEST_RESULT=0
    echo -e "${GREEN}Document ingestion tests passed!${NC}"
else
    UNIT_INGEST_RESULT=1
    echo -e "${RED}Document ingestion tests failed!${NC}"
fi

echo ""

# Run knowledge management unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Knowledge Management Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if ./tests/unit/knowledge_test.sh; then
    UNIT_KNOWLEDGE_RESULT=0
    echo -e "${GREEN}Knowledge management tests passed!${NC}"
else
    UNIT_KNOWLEDGE_RESULT=1
    echo -e "${RED}Knowledge management tests failed!${NC}"
fi

echo ""

# Run telemetry unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Telemetry Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if ./tests/unit/telemetry_test.sh; then
    UNIT_TELEMETRY_RESULT=0
    echo -e "${GREEN}Telemetry tests passed!${NC}"
else
    UNIT_TELEMETRY_RESULT=1
    echo -e "${RED}Telemetry tests failed!${NC}"
fi

echo ""

# Run agent telemetry unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Agent Telemetry Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if ./tests/unit/agent_telemetry_test.sh; then
    UNIT_AGENT_TELEMETRY_RESULT=0
    echo -e "${GREEN}Agent telemetry tests passed!${NC}"
else
    UNIT_AGENT_TELEMETRY_RESULT=1
    echo -e "${RED}Agent telemetry tests failed!${NC}"
fi

echo ""

# Run Hub API unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Hub API Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if ./tests/unit/hub_api_test.sh; then
    UNIT_HUB_API_RESULT=0
    echo -e "${GREEN}Hub API tests passed!${NC}"
else
    UNIT_HUB_API_RESULT=1
    echo -e "${RED}Hub API tests failed!${NC}"
fi

echo ""

# Run MCP server unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running MCP Server Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if ./tests/unit/mcp_test.sh; then
    UNIT_MCP_RESULT=0
    echo -e "${GREEN}MCP server tests passed!${NC}"
else
    UNIT_MCP_RESULT=1
    echo -e "${RED}MCP server tests failed!${NC}"
fi

echo ""

# Run AST Analysis unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running AST Analysis Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if bash tests/unit/ast_analysis_test.sh; then
    UNIT_AST_ANALYSIS_RESULT=0
    echo -e "${GREEN}AST Analysis tests passed!${NC}"
else
    UNIT_AST_ANALYSIS_RESULT=1
    echo -e "${RED}AST Analysis tests failed!${NC}"
fi

echo ""

# Run Security Analysis unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Security Analysis Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if bash tests/unit/security_analysis_test.sh; then
    UNIT_SECURITY_RESULT=0
    echo -e "${GREEN}Security Analysis tests passed!${NC}"
else
    UNIT_SECURITY_RESULT=1
    echo -e "${RED}Security Analysis tests failed!${NC}"
fi

echo ""

# Run File Size Management unit tests (Phase 9)
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running File Size Management Unit Tests (Phase 9)...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if bash tests/unit/file_size_test.sh; then
    UNIT_FILE_SIZE_RESULT=0
    echo -e "${GREEN}File Size Management tests passed!${NC}"
else
    UNIT_FILE_SIZE_RESULT=1
    echo -e "${RED}File Size Management tests failed!${NC}"
fi

echo ""

# Run Business Rules Compliance unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Business Rules Compliance Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if bash tests/unit/business_rules_test.sh; then
    UNIT_BUSINESS_RULES_RESULT=0
    echo -e "${GREEN}Business Rules Compliance tests passed!${NC}"
else
    UNIT_BUSINESS_RULES_RESULT=1
    echo -e "${RED}Business Rules Compliance tests failed!${NC}"
fi

echo ""

# Run Detection Metrics unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Detection Metrics Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if bash tests/unit/detection_metrics_test.sh; then
    UNIT_DETECTION_METRICS_RESULT=0
    echo -e "${GREEN}Detection Metrics tests passed!${NC}"
else
    UNIT_DETECTION_METRICS_RESULT=1
    echo -e "${RED}Detection Metrics tests failed!${NC}"
fi

echo ""

# Run Cache Race Condition unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Cache Race Condition Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if bash tests/unit/cache_race_condition_test.sh; then
    UNIT_CACHE_RACE_RESULT=0
    echo -e "${GREEN}Cache Race Condition tests passed!${NC}"
else
    UNIT_CACHE_RACE_RESULT=1
    echo -e "${RED}Cache Race Condition tests failed!${NC}"
fi

echo ""

# Run Retry Logic unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Retry Logic Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if bash tests/unit/retry_logic_test.sh; then
    UNIT_RETRY_LOGIC_RESULT=0
    echo -e "${GREEN}Retry Logic tests passed!${NC}"
else
    UNIT_RETRY_LOGIC_RESULT=1
    echo -e "${RED}Retry Logic tests failed!${NC}"
fi

echo ""

# Run Database Timeout unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Database Timeout Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if bash tests/unit/db_timeout_test.sh; then
    UNIT_DB_TIMEOUT_RESULT=0
    echo -e "${GREEN}Database Timeout tests passed!${NC}"
else
    UNIT_DB_TIMEOUT_RESULT=1
    echo -e "${RED}Database Timeout tests failed!${NC}"
fi

echo ""

# Run Error Recovery unit tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Error Recovery Unit Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if bash tests/unit/error_recovery_test.sh; then
    UNIT_ERROR_RECOVERY_RESULT=0
    echo -e "${GREEN}Error Recovery tests passed!${NC}"
else
    UNIT_ERROR_RECOVERY_RESULT=1
    echo -e "${RED}Error Recovery tests failed!${NC}"
fi

echo ""

# Run integration tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Integration Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if ./tests/integration/workflow_test.sh; then
    INTEGRATION_RESULT=0
    echo -e "${GREEN}Integration tests passed!${NC}"
else
    INTEGRATION_RESULT=1
    echo -e "${RED}Integration tests failed!${NC}"
fi

echo ""

# Run Hook Error Handling integration tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Hook Error Handling Integration Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if bash tests/integration/hook_error_handling_test.sh; then
    INTEGRATION_HOOK_ERROR_RESULT=0
    echo -e "${GREEN}Hook Error Handling tests passed!${NC}"
else
    INTEGRATION_HOOK_ERROR_RESULT=1
    echo -e "${RED}Hook Error Handling tests failed!${NC}"
fi

echo ""

# Run Cache Invalidation integration tests
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Running Cache Invalidation Integration Tests...${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

rm -f /tmp/sentinel.lock
if bash tests/integration/cache_invalidation_test.sh; then
    INTEGRATION_CACHE_INVALIDATION_RESULT=0
    echo -e "${GREEN}Cache Invalidation tests passed!${NC}"
else
    INTEGRATION_CACHE_INVALIDATION_RESULT=1
    echo -e "${RED}Cache Invalidation tests failed!${NC}"
fi

echo ""

# Summary
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           TEST SUMMARY                                     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

if [[ "${UNIT_SCANNING_RESULT}" -eq 0 ]]; then
    echo -e "Scanning Tests:    ${GREEN}✓ PASSED${NC}"
else
    echo -e "Scanning Tests:    ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_PATTERNS_RESULT}" -eq 0 ]]; then
    echo -e "Pattern Tests:     ${GREEN}✓ PASSED${NC}"
else
    echo -e "Pattern Tests:     ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_FIX_RESULT}" -eq 0 ]]; then
    echo -e "Auto-Fix Tests:    ${GREEN}✓ PASSED${NC}"
else
    echo -e "Auto-Fix Tests:    ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_INGEST_RESULT}" -eq 0 ]]; then
    echo -e "Ingest Tests:      ${GREEN}✓ PASSED${NC}"
else
    echo -e "Ingest Tests:      ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_KNOWLEDGE_RESULT}" -eq 0 ]]; then
    echo -e "Knowledge Tests:   ${GREEN}✓ PASSED${NC}"
else
    echo -e "Knowledge Tests:   ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_TELEMETRY_RESULT}" -eq 0 ]]; then
    echo -e "Telemetry Tests:   ${GREEN}✓ PASSED${NC}"
else
    echo -e "Telemetry Tests:   ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_AGENT_TELEMETRY_RESULT}" -eq 0 ]]; then
    echo -e "Agent Telemetry:   ${GREEN}✓ PASSED${NC}"
else
    echo -e "Agent Telemetry:   ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_HUB_API_RESULT}" -eq 0 ]]; then
    echo -e "Hub API Tests:     ${GREEN}✓ PASSED${NC}"
else
    echo -e "Hub API Tests:     ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_MCP_RESULT}" -eq 0 ]]; then
    echo -e "MCP Server Tests:  ${GREEN}✓ PASSED${NC}"
else
    echo -e "MCP Server Tests:  ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_AST_ANALYSIS_RESULT}" -eq 0 ]]; then
    echo -e "AST Analysis Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "AST Analysis Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_VIBE_ACCURACY_RESULT}" -eq 0 ]]; then
    echo -e "Vibe Accuracy Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Vibe Accuracy Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_VIBE_COMPARISON_RESULT}" -eq 0 ]]; then
    echo -e "Vibe Comparison Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Vibe Comparison Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_VIBE_FALLBACK_RESULT}" -eq 0 ]]; then
    echo -e "Vibe Fallback Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Vibe Fallback Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_VIBE_DEDUPLICATION_RESULT}" -eq 0 ]]; then
    echo -e "Vibe Deduplication Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Vibe Deduplication Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_FILE_SIZE_RESULT}" -eq 0 ]]; then
    echo -e "File Size Management Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "File Size Management Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_CACHE_RACE_RESULT}" -eq 0 ]]; then
    echo -e "Cache Race Condition Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Cache Race Condition Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_RETRY_LOGIC_RESULT}" -eq 0 ]]; then
    echo -e "Retry Logic Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Retry Logic Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_DB_TIMEOUT_RESULT}" -eq 0 ]]; then
    echo -e "Database Timeout Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Database Timeout Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_ERROR_RECOVERY_RESULT}" -eq 0 ]]; then
    echo -e "Error Recovery Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Error Recovery Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${INTEGRATION_RESULT}" -eq 0 ]]; then
    echo -e "Integration Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Integration Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${INTEGRATION_HOOK_ERROR_RESULT}" -eq 0 ]]; then
    echo -e "Hook Error Handling Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Hook Error Handling Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${INTEGRATION_CACHE_INVALIDATION_RESULT}" -eq 0 ]]; then
    echo -e "Cache Invalidation Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Cache Invalidation Tests: ${RED}✗ FAILED${NC}"
fi

# Phase 10: Test Enforcement System Tests
echo ""
echo -e "${YELLOW}Phase 10: Test Enforcement System Tests${NC}"
echo ""

echo "Running Mutation Engine Tests..."
UNIT_MUTATION_ENGINE_RESULT=0
if ! bash tests/unit/mutation_engine_test.sh; then
    UNIT_MUTATION_ENGINE_RESULT=1
fi

echo "Running Test Sandbox Tests..."
UNIT_TEST_SANDBOX_RESULT=0
if ! bash tests/unit/test_sandbox_test.sh; then
    UNIT_TEST_SANDBOX_RESULT=1
fi

echo "Running Agent Test Commands Tests..."
UNIT_TEST_AGENT_COMMANDS_RESULT=0
if ! bash tests/unit/test_agent_commands_test.sh; then
    UNIT_TEST_AGENT_COMMANDS_RESULT=1
fi

echo "Running Test Enforcement E2E Integration Tests..."
INTEGRATION_TEST_ENFORCEMENT_RESULT=0
if ! bash tests/integration/test_enforcement_e2e_test.sh; then
    INTEGRATION_TEST_ENFORCEMENT_RESULT=1
fi

echo ""
echo -e "${YELLOW}Phase 10 Test Results:${NC}"
if [[ "${UNIT_MUTATION_ENGINE_RESULT}" -eq 0 ]]; then
    echo -e "Mutation Engine Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Mutation Engine Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_TEST_SANDBOX_RESULT}" -eq 0 ]]; then
    echo -e "Test Sandbox Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Test Sandbox Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${UNIT_TEST_AGENT_COMMANDS_RESULT}" -eq 0 ]]; then
    echo -e "Agent Test Commands Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Agent Test Commands Tests: ${RED}✗ FAILED${NC}"
fi

if [[ "${INTEGRATION_TEST_ENFORCEMENT_RESULT}" -eq 0 ]]; then
    echo -e "Test Enforcement E2E Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Test Enforcement E2E Tests: ${RED}✗ FAILED${NC}"
fi

echo ""

# Overall result
if [[ "${UNIT_SCANNING_RESULT}" -eq 0 && "${UNIT_PATTERNS_RESULT}" -eq 0 && "${UNIT_FIX_RESULT}" -eq 0 && "${UNIT_INGEST_RESULT}" -eq 0 && "${UNIT_KNOWLEDGE_RESULT}" -eq 0 && "${UNIT_TELEMETRY_RESULT}" -eq 0 && "${UNIT_AGENT_TELEMETRY_RESULT}" -eq 0 && "${UNIT_HUB_API_RESULT}" -eq 0 && "${UNIT_MCP_RESULT}" -eq 0 && "${UNIT_AST_ANALYSIS_RESULT}" -eq 0 && "${UNIT_VIBE_ACCURACY_RESULT}" -eq 0 && "${UNIT_VIBE_COMPARISON_RESULT}" -eq 0 && "${UNIT_VIBE_FALLBACK_RESULT}" -eq 0 && "${UNIT_VIBE_DEDUPLICATION_RESULT}" -eq 0 && "${UNIT_CROSS_FILE_ANALYSIS_RESULT}" -eq 0 && "${UNIT_SECURITY_RESULT}" -eq 0 && "${UNIT_BUSINESS_RULES_RESULT}" -eq 0 && "${UNIT_DETECTION_METRICS_RESULT}" -eq 0 && "${UNIT_FILE_SIZE_RESULT}" -eq 0 && "${UNIT_CACHE_RACE_RESULT}" -eq 0 && "${UNIT_RETRY_LOGIC_RESULT}" -eq 0 && "${UNIT_DB_TIMEOUT_RESULT}" -eq 0 && "${UNIT_ERROR_RECOVERY_RESULT}" -eq 0 && "${INTEGRATION_RESULT}" -eq 0 && "${INTEGRATION_HOOK_ERROR_RESULT}" -eq 0 && "${INTEGRATION_CACHE_INVALIDATION_RESULT}" -eq 0 && "${UNIT_MUTATION_ENGINE_RESULT}" -eq 0 && "${UNIT_TEST_SANDBOX_RESULT}" -eq 0 && "${UNIT_TEST_AGENT_COMMANDS_RESULT}" -eq 0 && "${INTEGRATION_TEST_ENFORCEMENT_RESULT}" -eq 0 ]]; then
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║           ALL TESTS PASSED!                                ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║           SOME TESTS FAILED!                               ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi

