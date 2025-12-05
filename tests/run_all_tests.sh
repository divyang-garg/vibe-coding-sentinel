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
INTEGRATION_RESULT=0

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

if [[ "${INTEGRATION_RESULT}" -eq 0 ]]; then
    echo -e "Integration Tests: ${GREEN}✓ PASSED${NC}"
else
    echo -e "Integration Tests: ${RED}✗ FAILED${NC}"
fi

echo ""

# Overall result
if [[ "${UNIT_SCANNING_RESULT}" -eq 0 && "${UNIT_PATTERNS_RESULT}" -eq 0 && "${UNIT_FIX_RESULT}" -eq 0 && "${UNIT_INGEST_RESULT}" -eq 0 && "${UNIT_KNOWLEDGE_RESULT}" -eq 0 && "${UNIT_TELEMETRY_RESULT}" -eq 0 && "${UNIT_AGENT_TELEMETRY_RESULT}" -eq 0 && "${UNIT_HUB_API_RESULT}" -eq 0 && "${UNIT_MCP_RESULT}" -eq 0 && "${UNIT_AST_ANALYSIS_RESULT}" -eq 0 && "${UNIT_VIBE_ACCURACY_RESULT}" -eq 0 && "${UNIT_VIBE_COMPARISON_RESULT}" -eq 0 && "${UNIT_VIBE_FALLBACK_RESULT}" -eq 0 && "${UNIT_VIBE_DEDUPLICATION_RESULT}" -eq 0 && "${UNIT_CROSS_FILE_ANALYSIS_RESULT}" -eq 0 && "${UNIT_SECURITY_RESULT}" -eq 0 && "${UNIT_BUSINESS_RULES_RESULT}" -eq 0 && "${UNIT_DETECTION_METRICS_RESULT}" -eq 0 && "${UNIT_FILE_SIZE_RESULT}" -eq 0 && "${INTEGRATION_RESULT}" -eq 0 ]]; then
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

