# LLM Architecture Documentation

## Overview

The Sentinel Hub API uses LLM (Large Language Model) calls for various code analysis tasks. The implementation is split across two packages to serve different use cases:

1. **Main Package** (`hub/api/`) - Cost-optimized analysis with Phase 14D features
2. **Services Package** (`hub/api/services/`) - General-purpose code analysis

## Main Package Implementation

### Location
- `hub/api/llm_cache_analysis.go` - Progressive depth analysis with cost optimization
- `hub/api/llm_cache_prompts.go` - Prompt generation (wrapper to unified builder)

### Purpose
Implements **Phase 14D cost optimization** to reduce LLM API costs by up to 40% while maintaining analysis quality.

### Key Features

#### 1. Progressive Depth Analysis
- **surface**: No LLM calls (AST/pattern matching only) - **$0 cost**
- **medium**: Cheaper models (gpt-3.5-turbo, claude-3-haiku) - **Low cost**
- **deep**: Expensive models (gpt-4, claude-3-opus) - **High cost**

#### 2. Intelligent Model Selection
The `selectModelWithDepth()` function:
- Selects models based on analysis depth and cost limits
- Estimates token usage for cost calculation
- Enforces maximum cost per request limits
- Automatically downgrades to cheaper models when cost limits are exceeded

#### 3. Analysis Types
- `semantic_analysis` - Logic errors, edge cases, bugs
- `business_logic` - Business rule compliance
- `error_handling` - Error handling patterns

#### 4. Cost Tracking
- Tracks usage with `ValidationID` for detailed cost analysis
- Records token usage and estimated costs
- Supports cost limit enforcement

### Usage Example

```go
result, err := analyzeWithProgressiveDepth(
    ctx,
    config,
    fileContent,
    "semantic_analysis",
    "medium",  // Use cheaper models
    projectID,
    validationID,
)
```

## Services Package Implementation

### Location
- `hub/api/services/llm_cache_analysis.go` - Progressive depth analysis
- `hub/api/services/prompt_builder.go` - Unified prompt generation

### Purpose
Provides general-purpose code quality analysis without complex cost optimization.

### Key Features

#### 1. Progressive Depth Analysis
- **quick**: Brief summary, 3-5 findings - Fast feedback
- **medium**: Moderate analysis with examples - Standard reviews
- **deep**: Comprehensive with all details - Thorough analysis

#### 2. Model Selection
- Uses configured model (no dynamic selection)
- Simpler implementation
- Relies on user's model configuration

#### 3. Analysis Types
- `security` - Security vulnerabilities
- `performance` - Performance issues
- `maintainability` - Code quality
- `architecture` - Design patterns

### Usage Example

```go
result, err := analyzeWithProgressiveDepth(
    ctx,
    config,
    fileContent,
    "security",
    "deep",  // Comprehensive analysis
    projectID,
    validationID,
)
```

## Unified Prompt Generation

### Location
- `hub/api/services/prompt_builder.go` - `GeneratePrompt()` function

### Purpose
Consolidates prompt generation for all analysis types across both packages.

### Supported Analysis Types
- Main package: `semantic_analysis`, `business_logic`, `error_handling`
- Services package: `security`, `performance`, `maintainability`, `architecture`

### Depth Level Mapping
- `surface` (main) → `quick` (services) - Brief analysis
- `medium` (both) - Detailed analysis
- `deep` (both) - Comprehensive analysis

### Usage

Both packages use the unified prompt builder:
- Main package: `generatePrompt()` → `services.GeneratePrompt()`
- Services package: `generatePrompt()` → `GeneratePrompt()`

## When to Use Each Implementation

### Use Main Package When:
- Cost optimization is important (Phase 14D features)
- You need intelligent model selection
- You want to enforce cost limits
- You're doing semantic analysis, business logic, or error handling

### Use Services Package When:
- You need general code quality analysis
- Cost optimization is not a priority
- You're doing security, performance, maintainability, or architecture analysis
- You want simpler, more straightforward implementation

## Cost Optimization Strategy

### Phase 14D Features (Main Package Only)

1. **Surface Depth**: Skip LLM entirely for quick checks
2. **Model Selection**: Automatically choose cheaper models for medium depth
3. **Cost Limits**: Enforce maximum cost per request
4. **Caching**: Comprehensive caching to avoid redundant calls
5. **Token Estimation**: Estimate costs before making calls

### Cost Savings

- **Surface depth**: 100% savings (no LLM calls)
- **Medium depth**: ~20x cheaper (gpt-3.5 vs gpt-4)
- **Smart caching**: Avoids redundant calls
- **Overall**: Up to 40% cost reduction

## Function Signatures

### Main Package

```go
// Model selection with cost optimization
func selectModelWithDepth(
    ctx context.Context,
    analysisType string,
    config *LLMConfig,
    depth string,           // "surface", "medium", "deep"
    estimatedTokens int,
    projectID string,
) (string, error)

// LLM call with depth-aware settings
func callLLMWithDepth(
    ctx context.Context,
    config *LLMConfig,
    prompt string,
    analysisType string,
    depth string,           // "surface", "medium", "deep"
    projectID string,
) (string, int, error)
```

### Services Package

```go
// Progressive depth analysis
func analyzeWithProgressiveDepth(
    ctx context.Context,
    config *LLMConfig,
    fileContent string,
    analysisType string,
    depth string,           // "quick", "medium", "deep"
    projectID string,
    validationID string,
) (string, error)

// Unified prompt generation
func GeneratePrompt(
    analysisType string,
    depth string,
    fileContent string,
) string
```

## Migration Guide

### From Deprecated Functions

If you were using deprecated functions from `utils.go`:

1. **AST Functions**: Use `ast` package directly
   - `getParser()` → `ast.GetParser()`
   - `traverseAST()` → `ast.TraverseAST()`
   - `analyzeAST()` → `ast.AnalyzeAST()`

2. **LLM Functions**: Use package-specific implementations
   - Main package: Use `selectModelWithDepth()` and `callLLMWithDepth()` from `llm_cache_analysis.go`
   - Services package: Use `callLLMWithDepth()` from `helpers_stubs.go`

## Best Practices

1. **Choose the right package** based on your needs (cost optimization vs simplicity)
2. **Use appropriate depth levels** - surface/quick for CI/CD, deep for critical analysis
3. **Enable caching** to reduce costs and improve performance
4. **Set cost limits** in main package to prevent budget overruns
5. **Track usage** with ValidationID for detailed cost analysis

## Related Documentation

- [Phase 14D Cost Optimization Guide](../external/PHASE_14D_GUIDE.md)
- [LLM Implementation Purpose Analysis](../../LLM_IMPLEMENTATION_PURPOSE_ANALYSIS.md)
- [LLM Legacy Analysis](../../LLM_LEGACY_ANALYSIS.md)
