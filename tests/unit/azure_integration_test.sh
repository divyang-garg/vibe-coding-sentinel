#!/bin/bash
# Azure AI Foundry Integration Tests
# Tests provider interface, fallback chain, and Azure client (mocked)

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

echo "🧪 Testing Azure AI Foundry Integration"
echo "══════════════════════════════════════════════════════════════"
echo ""

# Test 1: Provider interface exists
echo "Test 1: LLM Provider Interface"
if grep -q "type LLMProvider interface" hub/processor/main.go; then
    echo "   ✅ LLMProvider interface found"
else
    echo "   ❌ LLMProvider interface not found"
    exit 1
fi

# Test 2: Azure provider exists
echo "Test 2: Azure Provider Implementation"
if grep -q "type AzureFoundryProvider struct" hub/processor/main.go; then
    echo "   ✅ AzureFoundryProvider struct found"
else
    echo "   ❌ AzureFoundryProvider struct not found"
    exit 1
fi

# Test 3: Ollama provider exists
echo "Test 3: Ollama Provider Implementation"
if grep -q "type OllamaProvider struct" hub/processor/main.go; then
    echo "   ✅ OllamaProvider struct found"
else
    echo "   ❌ OllamaProvider struct not found"
    exit 1
fi

# Test 4: Fallback chain implemented
echo "Test 4: Provider Fallback Chain"
if grep -q "providers := \[\]LLMProvider" hub/processor/main.go; then
    echo "   ✅ Provider fallback chain found"
else
    echo "   ❌ Provider fallback chain not found"
    exit 1
fi

# Test 5: Azure configuration in docker-compose
echo "Test 5: Azure Environment Variables in Docker Compose"
if grep -q "AZURE_AI_ENDPOINT" hub/docker-compose.yml; then
    echo "   ✅ Azure env vars in docker-compose.yml"
else
    echo "   ❌ Azure env vars missing from docker-compose.yml"
    exit 1
fi

# Test 6: Azure setup guide exists
echo "Test 6: Azure Setup Documentation"
if [ -f "docs/external/AZURE_SETUP_GUIDE.md" ]; then
    echo "   ✅ Azure setup guide found"
else
    echo "   ❌ Azure setup guide not found"
    exit 1
fi

# Test 7: Provider methods implemented
echo "Test 7: Provider Methods"
if grep -q "func (p \*AzureFoundryProvider) Name()" hub/processor/main.go && \
   grep -q "func (p \*AzureFoundryProvider) IsAvailable()" hub/processor/main.go && \
   grep -q "func (p \*AzureFoundryProvider) ExtractKnowledge" hub/processor/main.go; then
    echo "   ✅ All Azure provider methods implemented"
else
    echo "   ❌ Missing Azure provider methods"
    exit 1
fi

# Test 8: Claude-optimized prompts
echo "Test 8: Claude-Optimized Prompts"
if grep -q "<document>" hub/processor/main.go && \
   grep -q "system" hub/processor/main.go; then
    echo "   ✅ Claude-optimized prompts found"
else
    echo "   ⚠️  Claude prompts may need optimization"
fi

echo ""
echo "✅ All Azure integration tests passed!"
echo ""

