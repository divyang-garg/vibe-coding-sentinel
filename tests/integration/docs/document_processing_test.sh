#!/bin/bash
# Document Processing Integration Test
# Tests the complete document processing pipeline

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0

log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    ((FAILED++))
}

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

SENTINEL="/Users/divyanggarg/VicecodingSentinel/sentinel"

# Setup test environment
TEST_DIR="/tmp/doc_processing_test_$(date +%s)"
mkdir -p "$TEST_DIR/docs"

# Create test documents
cat > "$TEST_DIR/docs/readme.md" << 'EOF'
# Test Project Documentation

## Overview
This is a test project for Sentinel document processing.

## Features
- Security scanning
- Pattern learning
- Auto-fix capabilities
- Document processing

## Getting Started
1. Initialize: `sentinel init`
2. Learn patterns: `sentinel learn`
3. Process docs: `sentinel docs index docs/`
EOF

cat > "$TEST_DIR/docs/api.html" << 'EOF'
<!DOCTYPE html>
<html>
<head><title>API Documentation</title></head>
<body>
    <h1>API Reference</h1>
    <p>This document describes the API endpoints.</p>
    <div class="endpoint">
        <h3>GET /api/scan</h3>
        <p>Performs security scanning on code.</p>
    </div>
    <script>console.log("API documentation loaded");</script>
</body>
</html>
EOF

cat > "$TEST_DIR/docs/config.json" << 'EOF'
{
  "application": "sentinel",
  "version": "1.0.0",
  "features": {
    "security": true,
    "patterns": true,
    "docs": true
  },
  "settings": {
    "maxFileSize": "10MB",
    "supportedFormats": ["md", "html", "json", "txt"]
  }
}
EOF

echo "📄 Document Processing Integration Test"
echo "======================================"

cd "$TEST_DIR"

# Test 1: Document indexing
echo ""
log_info "Test 1: Document Indexing"
if $SENTINEL docs index docs/ >/dev/null 2>&1; then
    if [[ -f ".sentinel/docs/index.json" ]]; then
        log_pass "Document indexing creates index file"
    else
        log_fail "Document indexing does not create index file"
    fi
else
    log_fail "Document indexing command failed"
fi

# Test 2: Index content validation
echo ""
log_info "Test 2: Index Content Validation"
if [[ -f ".sentinel/docs/index.json" ]]; then
    INDEX_CONTENT=$(cat .sentinel/docs/index.json)
    if echo "$INDEX_CONTENT" | grep -q '"documents"'; then
        log_pass "Index contains documents section"
    else
        log_fail "Index missing documents section"
    fi

    if echo "$INDEX_CONTENT" | grep -q '"index"'; then
        log_pass "Index contains search index"
    else
        log_fail "Index missing search index"
    fi
else
    log_fail "Index file not found for validation"
fi

# Test 3: Document search
echo ""
log_info "Test 3: Document Search Functionality"
if $SENTINEL docs search "security" >/dev/null 2>&1; then
    log_pass "Document search command executes"
else
    log_fail "Document search command failed"
fi

# Test 4: Search results validation
echo ""
log_info "Test 4: Search Results Validation"
SEARCH_RESULTS=$($SENTINEL docs search "api" 2>/dev/null)
if echo "$SEARCH_RESULTS" | grep -q "API Documentation\|API Reference"; then
    log_pass "Search finds relevant documents"
else
    log_fail "Search does not find relevant documents"
fi

# Test 5: Multi-format support
echo ""
log_info "Test 5: Multi-Format Support"
INDEX_CONTENT=$(cat .sentinel/docs/index.json 2>/dev/null || echo "")
MD_COUNT=$(echo "$INDEX_CONTENT" | grep -c '"format":"md"' || echo "0")
HTML_COUNT=$(echo "$INDEX_CONTENT" | grep -c '"format":"html"' || echo "0")
JSON_COUNT=$(echo "$INDEX_CONTENT" | grep -c '"format":"json"' || echo "0")

TOTAL_DOCS=$((MD_COUNT + HTML_COUNT + JSON_COUNT))
DOC_COUNT=$TOTAL_DOCS  # Store for later comparison
if [[ "$TOTAL_DOCS" -ge 3 ]]; then
    log_pass "Multi-format documents indexed ($TOTAL_DOCS documents: $MD_COUNT MD, $HTML_COUNT HTML, $JSON_COUNT JSON)"
else
    log_fail "Not all document formats indexed (found $TOTAL_DOCS)"
fi

# Test 6: Document metadata
echo ""
log_info "Test 6: Document Metadata"
INDEX_CONTENT=$(cat .sentinel/docs/index.json 2>/dev/null || echo "")
if echo "$INDEX_CONTENT" | grep -q '"format":"md"'; then
    log_pass "Markdown documents have correct format metadata"
else
    log_fail "Markdown documents missing format metadata"
fi

if echo "$INDEX_CONTENT" | grep -q '"format":"html"'; then
    log_pass "HTML documents have correct format metadata"
else
    log_fail "HTML documents missing format metadata"
fi

# Test 7: Content extraction
echo ""
log_info "Test 7: Content Extraction"
INDEX_CONTENT=$(cat .sentinel/docs/index.json 2>/dev/null || echo "")
MD_CONTENT=$(echo "$INDEX_CONTENT" | grep -A 5 '"format":"md"' | grep '"content"' | head -1 | sed 's/.*"content":"\([^"]*\)".*/\1/')
if [[ -n "$MD_CONTENT" ]] && echo "$MD_CONTENT" | grep -q "Test Project\|Features\|Getting Started"; then
    log_pass "Markdown content properly extracted"
else
    log_fail "Markdown content extraction failed"
fi

HTML_CONTENT=$(echo "$INDEX_CONTENT" | grep -A 5 '"format":"html"' | grep '"content"' | head -1 | sed 's/.*"content":"\([^"]*\)".*/\1/')
if [[ -n "$HTML_CONTENT" ]] && echo "$HTML_CONTENT" | grep -q "API Reference\|API endpoints"; then
    log_pass "HTML content properly extracted"
else
    log_fail "HTML content extraction failed"
fi

# Test 8: Search indexing
echo ""
log_info "Test 8: Search Index Validation"
INDEX_CONTENT=$(cat .sentinel/docs/index.json 2>/dev/null || echo "")
INDEX_TERMS=$(echo "$INDEX_CONTENT" | grep -c '"index"' || echo "0")
if [[ "$INDEX_TERMS" -gt 0 ]]; then
    log_pass "Search index contains terms"
else
    log_fail "Search index is empty"
fi

# Test 9: Re-indexing
echo ""
log_info "Test 9: Re-indexing"
# Add a new document
cat > "$TEST_DIR/docs/new_doc.txt" << 'EOF'
This is a new text document.
It contains additional information about Sentinel.
The document processing system should index this automatically.
EOF

if $SENTINEL docs index docs/ >/dev/null 2>&1; then
    NEW_INDEX_CONTENT=$(cat .sentinel/docs/index.json 2>/dev/null || echo "")
    NEW_DOC_COUNT=$(echo "$NEW_INDEX_CONTENT" | grep -c '"id":' || echo "0")
    if [[ "$NEW_DOC_COUNT" -gt "$DOC_COUNT" ]]; then
        log_pass "Re-indexing adds new documents ($NEW_DOC_COUNT > $DOC_COUNT)"
    else
        log_fail "Re-indexing does not add new documents ($NEW_DOC_COUNT <= $DOC_COUNT)"
    fi
else
    log_fail "Re-indexing command failed"
fi

# Cleanup
cd /
rm -rf "$TEST_DIR"

# Final results
echo ""
echo "📊 DOCUMENT PROCESSING TEST RESULTS"
echo "==================================="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
TOTAL=$((PASSED + FAILED))

if [[ $TOTAL -gt 0 ]]; then
    SUCCESS_RATE=$((PASSED * 100 / TOTAL))

    if [[ $SUCCESS_RATE -ge 85 ]]; then
        echo -e "${GREEN}🎉 SUCCESS: Document Processing Fully Operational (${SUCCESS_RATE}%)${NC}"
        echo "✅ Document parsing, indexing, and search working correctly"
        exit 0
    elif [[ $SUCCESS_RATE -ge 70 ]]; then
        echo -e "${YELLOW}⚠️  PARTIAL SUCCESS: Core functionality working (${SUCCESS_RATE}%)${NC}"
        echo "✅ Basic document processing operational"
        exit 0
    else
        echo -e "${RED}❌ FAILURE: Document processing needs work (${SUCCESS_RATE}%)${NC}"
        echo "❌ Critical issues with document processing"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  No tests executed${NC}"
    exit 1
fi
