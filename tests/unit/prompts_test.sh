#!/bin/bash
# Unit tests for Phase 13 prompts
# Run from project root: ./tests/unit/prompts_test.sh

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

echo "🧪 Testing Prompts"
echo "══════════════════════════════════════════════════════════════"
echo ""

# Test 1: Business rule prompt includes required sections
echo "Test 1: Business Rule Prompt Includes Required Sections"
if grep -q "constraints" hub/processor/prompts.go && \
   grep -q "pseudocode" hub/processor/prompts.go && \
   grep -q "boundary" hub/processor/prompts.go && \
   grep -q "test_requirements" hub/processor/prompts.go && \
   grep -q "traceability" hub/processor/prompts.go; then
    echo "   ✅ Business rule prompt includes all required sections"
else
    echo "   ❌ Business rule prompt missing required sections"
    exit 1
fi

# Test 2: Prompt includes JSON schema reference
echo ""
echo "Test 2: Prompt Includes JSON Schema Reference"
if grep -q "JSON schema\|schema\|format" hub/processor/prompts.go; then
    echo "   ✅ Prompt includes schema reference"
else
    echo "   ⚠️  Prompt may need schema reference"
fi

# Test 3: Prompt includes examples
echo ""
echo "Test 3: Prompt Includes Examples"
if grep -q "example\|Example\|BR-001" hub/processor/prompts.go; then
    echo "   ✅ Prompt includes examples"
else
    echo "   ⚠️  Prompt may need examples"
fi

# Test 4: Prompt functions exist
echo ""
echo "Test 4: Prompt Functions Exist"
if grep -q "getBusinessRulePrompt\|getEntityPrompt\|getGlossaryPrompt" hub/processor/prompts.go; then
    echo "   ✅ Prompt functions found"
else
    echo "   ❌ Prompt functions not found"
    exit 1
fi

# Test 5: Boundary specification emphasized
echo ""
echo "Test 5: Boundary Specification Emphasized"
if grep -q "inclusive\|exclusive\|boundary" hub/processor/prompts.go; then
    echo "   ✅ Boundary specification emphasized"
else
    echo "   ⚠️  Boundary specification may need emphasis"
fi

echo ""
echo "✅ Prompt tests completed!"
echo ""











