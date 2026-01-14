#!/bin/bash
# Unit tests for Sentinel document ingestion functionality
# Run from project root: ./tests/unit/ingest_test.sh

# Don't use set -e as some commands intentionally fail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    ((TESTS_FAILED++))
}

cleanup_lock() {
    rm -f /tmp/sentinel.lock
}

echo ""
echo "=============================================="
echo "   Document Ingestion Unit Tests"
echo "=============================================="
echo ""

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SENTINEL="$PROJECT_ROOT/sentinel"
TEST_DIR=$(mktemp -d)
FIXTURES_DIR="$PROJECT_ROOT/tests/fixtures/documents"

# ============================================================================
# Test: Ingest help message
# ============================================================================

echo "Testing ingest help message..."
cleanup_lock

OUTPUT=$("$SENTINEL" ingest 2>&1)
if echo "$OUTPUT" | grep -q "Supported formats"; then
    log_pass "Shows help when no path provided"
else
    log_fail "Help message not shown"
fi

# ============================================================================
# Test: Ingest text file
# ============================================================================

echo ""
echo "Testing text file ingestion..."
cleanup_lock

cd "$TEST_DIR"
mkdir -p docs/knowledge/source-documents docs/knowledge/extracted

# Create a test text file
echo "This is a test document for ingestion." > test.txt

OUTPUT=$("$SENTINEL" ingest test.txt 2>&1)
if echo "$OUTPUT" | grep -q "Successful:.*1"; then
    log_pass "Ingests text files successfully"
else
    log_fail "Failed to ingest text file"
fi

# Check extracted file exists
if [[ -f "docs/knowledge/extracted/test.txt" ]]; then
    log_pass "Creates extracted text file"
else
    log_fail "Extracted file not created"
fi

# ============================================================================
# Test: Ingest markdown file
# ============================================================================

echo ""
echo "Testing markdown file ingestion..."
cleanup_lock

cp "$FIXTURES_DIR/sample_scope.md" "$TEST_DIR/"
OUTPUT=$("$SENTINEL" ingest sample_scope.md 2>&1)

if echo "$OUTPUT" | grep -q "Successful"; then
    log_pass "Ingests markdown files successfully"
else
    log_fail "Failed to ingest markdown file"
fi

# ============================================================================
# Test: Ingest directory
# ============================================================================

echo ""
echo "Testing directory ingestion..."
cleanup_lock

mkdir -p "$TEST_DIR/docs_to_ingest"
echo "Document 1" > "$TEST_DIR/docs_to_ingest/doc1.txt"
echo "Document 2" > "$TEST_DIR/docs_to_ingest/doc2.txt"
echo "Document 3" > "$TEST_DIR/docs_to_ingest/doc3.md"

OUTPUT=$("$SENTINEL" ingest "$TEST_DIR/docs_to_ingest" 2>&1)
if echo "$OUTPUT" | grep -q "Found 3 documents"; then
    log_pass "Finds multiple documents in directory"
else
    log_fail "Failed to find documents in directory"
fi

# ============================================================================
# Test: --list command
# ============================================================================

echo ""
echo "Testing --list command..."
cleanup_lock

OUTPUT=$("$SENTINEL" ingest --list 2>&1)
if echo "$OUTPUT" | grep -q "Ingested Documents"; then
    log_pass "Lists ingested documents"
else
    log_fail "Failed to list documents"
fi

# ============================================================================
# Test: Skip images flag
# ============================================================================

echo ""
echo "Testing --skip-images flag..."
cleanup_lock

# Create a fake image file (just for testing the skip logic)
mkdir -p "$TEST_DIR/with_images"
echo "text content" > "$TEST_DIR/with_images/readme.txt"
touch "$TEST_DIR/with_images/image.png"

OUTPUT=$("$SENTINEL" ingest "$TEST_DIR/with_images" --skip-images 2>&1)
if echo "$OUTPUT" | grep -q "Skipping image"; then
    log_pass "Skips images with --skip-images flag"
else
    log_fail "Did not skip images"
fi

# ============================================================================
# Test: Manifest creation
# ============================================================================

echo ""
echo "Testing manifest creation..."
cleanup_lock

if [[ -f "$TEST_DIR/docs/knowledge/source-documents/manifest.json" ]]; then
    log_pass "Creates manifest.json"
else
    log_fail "Manifest not created"
fi

# ============================================================================
# Test: Text file from fixtures
# ============================================================================

echo ""
echo "Testing requirements.txt parsing..."
cleanup_lock

cd "$TEST_DIR"
cp "$FIXTURES_DIR/sample_requirements.txt" "$TEST_DIR/"

OUTPUT=$("$SENTINEL" ingest sample_requirements.txt --verbose 2>&1)
if echo "$OUTPUT" | grep -q "Extracted"; then
    log_pass "Parses requirements document"
else
    log_fail "Failed to parse requirements document"
fi

# Check content was extracted
if [[ -f "docs/knowledge/extracted/sample_requirements.txt" ]]; then
    CONTENT=$(cat "docs/knowledge/extracted/sample_requirements.txt")
    if echo "$CONTENT" | grep -q "Business Rules"; then
        log_pass "Extracts business rules content"
    else
        log_fail "Content not properly extracted"
    fi
else
    log_fail "Extracted file not found"
fi

# ============================================================================
# Cleanup
# ============================================================================

rm -rf "$TEST_DIR"

# ============================================================================
# Summary
# ============================================================================

echo ""
echo "=============================================="
echo "   Document Ingestion Test Results"
echo "=============================================="
echo ""
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"

TOTAL=$((TESTS_PASSED + TESTS_FAILED))
if [[ $TOTAL -gt 0 ]]; then
    PERCENT=$((TESTS_PASSED * 100 / TOTAL))
    echo "Success Rate: ${PERCENT}%"
fi
echo ""

if [[ $TESTS_FAILED -gt 0 ]]; then
    exit 1
fi

exit 0












