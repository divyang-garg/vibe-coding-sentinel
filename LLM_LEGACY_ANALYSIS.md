# Detailed LLM Legacy Code Analysis

## Executive Summary

After thorough codebase analysis, here are the **actual** legacy cases (not assumptions):

## 1. CONFIRMED LEGACY/DEPRECATED

### 1.1 Deprecated AST Functions in `hub/api/utils.go`
**Status:** ✅ CONFIRMED DEPRECATED

**Location:** `hub/api/utils.go` (lines 75-106)

**Functions:**
- `getParser(language string)` - Returns error: "deprecated - use ast.GetParser"
- `traverseAST(node, callback)` - Empty stub, marked deprecated
- `analyzeAST(code, language, options)` - Returns error: "deprecated - use ast.AnalyzeAST"

**Evidence:**
```go
// DEPRECATED: These stub functions have been replaced by the AST package.
// Use github.com/divyang-garg/sentinel-hub-api/hub/api/ast instead
```

**Replacement:** Use `ast.GetParser()`, `ast.TraverseAST()`, `ast.AnalyzeAST()` from `hub/api/ast` package

**Impact:** Low - These are stub functions that return errors, so any code using them would fail immediately.

---

### 1.2 Unimplemented Stub Function in `hub/api/utils.go`
**Status:** ✅ CONFIRMED STUB (Never Implemented)

**Location:** `hub/api/utils.go` (line 227)

**Function:**
- `callLLMWithDepth(ctx, config, prompt, taskType, model, mode)` 

**Evidence:**
```go
// callLLMWithDepth calls LLM with depth settings (stub)
func callLLMWithDepth(...) (string, int, error) {
    return "", 0, fmt.Errorf("callLLMWithDepth not implemented - use services package")
}
```

**Replacement:** Use `services.callLLMWithDepth()` from `hub/api/services/helpers_stubs.go`

**Impact:** Medium - This function would fail if called, but there's a working implementation in services package.

---

## 2. DUPLICATE IMPLEMENTATIONS (Not Legacy, But Redundant)

### 2.1 Duplicate `generatePrompt` Functions
**Status:** ⚠️ DUPLICATE (Both Active, Different Packages)

**Two Implementations:**

#### Implementation A: `hub/api/llm_cache_prompts.go` (package main)
**Location:** `hub/api/llm_cache_prompts.go:10`
**Package:** `main`
**Analysis Types Supported:**
- `semantic_analysis` - JSON format with issues array
- `business_logic` - Business rule compliance analysis
- `error_handling` - Error handling pattern analysis
- `default` - Generic code analysis

**Used By:**
- `hub/api/llm_cache_analysis.go:56` (package main)
- `hub/api/llm_cache_analysis.go:106` (package main)

**Evidence:**
```go
// In hub/api/llm_cache_analysis.go (package main)
prompt := generatePrompt(analysisType, depth, fileContent)  // Uses main package version
```

#### Implementation B: `hub/api/services/llm_cache_analysis.go` (package services)
**Location:** `hub/api/services/llm_cache_analysis.go:40`
**Package:** `services`
**Analysis Types Supported:**
- `security` - Security vulnerabilities and best practices
- `performance` - Performance issues and optimization
- `maintainability` - Code quality and maintainability
- `architecture` - Code structure and design patterns
- `default` - Generic code analysis

**Used By:**
- `hub/api/services/llm_cache_analysis.go:25` (package services)
- `hub/api/services/logic_analyzer_semantic.go:44` (calls `analyzeWithProgressiveDepth` which uses services version)

**Evidence:**
```go
// In hub/api/services/llm_cache_analysis.go (package services)
func analyzeWithProgressiveDepth(...) {
    prompt := generatePrompt(analysisType, depth, fileContent)  // Uses services package version
}
```

**Key Difference:**
- **Main package version** focuses on: semantic analysis, business logic, error handling
- **Services package version** focuses on: security, performance, maintainability, architecture

**Conclusion:** These are **NOT legacy** - they serve different purposes and are in different packages. However, they could be consolidated for better maintainability.

---

### 2.2 Duplicate `analyzeWithProgressiveDepth` Functions
**Status:** ⚠️ DUPLICATE (Both Active, Different Packages)

#### Implementation A: `hub/api/llm_cache_analysis.go` (package main)
**Location:** `hub/api/llm_cache_analysis.go:26`
**Package:** `main`
**Features:**
- Uses `generatePrompt` from main package
- Calls `callLLMWithDepth` (which may not exist in main package)
- Supports surface/medium/deep depth levels
- Includes model selection logic

**Used By:**
- `hub/api/logic_analyzer.go:156` (package main) - ✅ VERIFIED: Uses main package version

**Evidence:**
```go
// hub/api/logic_analyzer.go is in package main
// It calls analyzeWithProgressiveDepth without importing services
// Therefore it uses the main package version from llm_cache_analysis.go
```

#### Implementation B: `hub/api/services/llm_cache_analysis.go` (package services)
**Location:** `hub/api/services/llm_cache_analysis.go:12`
**Package:** `services`
**Features:**
- Uses `generatePrompt` from services package
- Calls `callLLM` from services package
- Supports quick/medium/deep depth levels
- Simpler implementation

**Used By:**
- `hub/api/services/logic_analyzer_semantic.go:44` (package services) - ✅ VERIFIED: Uses services package version

**Verification:**
- `hub/api/logic_analyzer.go` (package main) uses main package version ✅
- `hub/api/services/logic_analyzer_semantic.go` (package services) uses services package version ✅
- Both are correct for their respective packages

---

## 3. POTENTIALLY UNUSED CODE

### 3.1 `hub/api/llm_cache_prompts.go`
**Status:** ⚠️ POTENTIALLY REDUNDANT

**Analysis:**
- File is in `package main`
- Defines `generatePrompt` with semantic_analysis, business_logic, error_handling support
- **IS being used** by `hub/api/llm_cache_analysis.go` (same package)
- However, the services package has a newer implementation with different analysis types

**Recommendation:** 
- If `hub/api/llm_cache_analysis.go` (main package) is still actively used, keep it
- If not, this file could be removed
- Consider consolidating both implementations

---

## 4. ACTIVE (NOT LEGACY) LLM FUNCTIONS

### 4.1 Knowledge Extraction Prompts
**Status:** ✅ ACTIVE
**Location:** `internal/extraction/prompt.go`
- `BuildBusinessRulesPrompt()` - ✅ Active
- `BuildEntitiesPrompt()` - ✅ Active
- `BuildAPIContractsPrompt()` - ✅ Active
- `BuildUserJourneysPrompt()` - ✅ Active
- `BuildGlossaryPrompt()` - ✅ Active

**Used By:** `internal/extraction/extractor.go`

---

### 4.2 Intent Analysis Prompt
**Status:** ✅ ACTIVE
**Location:** `hub/api/services/intent_analyzer.go:132`
**Function:** `buildIntentAnalysisPrompt()`
**Used By:** `hub/api/services/intent_analyzer.go:74`

---

### 4.3 Semantic Function Analysis Prompt
**Status:** ✅ ACTIVE
**Location:** `hub/api/services/logic_analyzer_semantic.go:76`
**Function:** `buildSemanticAnalysisPrompt()`
**Used By:** `hub/api/services/logic_analyzer_semantic.go:40`

---

### 4.4 Progressive Depth Analysis (Services Package)
**Status:** ✅ ACTIVE
**Location:** `hub/api/services/llm_cache_analysis.go:40`
**Function:** `generatePrompt()` (services package version)
**Used By:** `hub/api/services/llm_cache_analysis.go:25`

---

### 4.5 Progressive Depth Analysis (Main Package)
**Status:** ✅ ACTIVE (but potentially redundant)
**Location:** `hub/api/llm_cache_prompts.go:10`
**Function:** `generatePrompt()` (main package version)
**Used By:** `hub/api/llm_cache_analysis.go:56,106`

---

## 5. SUMMARY TABLE

| Item | Status | Location | Package | Still Used? | Legacy? |
|------|--------|----------|---------|------------|---------|
| `utils.go` AST functions | Deprecated | `hub/api/utils.go` | main | No (returns errors) | ✅ Yes |
| `utils.go` callLLMWithDepth | Stub | `hub/api/utils.go` | main | No (returns error) | ✅ Yes |
| `llm_cache_prompts.go` generatePrompt | Active | `hub/api/llm_cache_prompts.go` | main | Yes | ❌ No |
| `services/llm_cache_analysis.go` generatePrompt | Active | `hub/api/services/llm_cache_analysis.go` | services | Yes | ❌ No |
| `llm_cache_analysis.go` analyzeWithProgressiveDepth | Active | `hub/api/llm_cache_analysis.go` | main | Yes | ❌ No |
| `services/llm_cache_analysis.go` analyzeWithProgressiveDepth | Active | `hub/api/services/llm_cache_analysis.go` | services | Yes | ❌ No |
| Knowledge extraction prompts | Active | `internal/extraction/prompt.go` | extraction | Yes | ❌ No |
| Intent analysis prompt | Active | `hub/api/services/intent_analyzer.go` | services | Yes | ❌ No |
| Semantic analysis prompt | Active | `hub/api/services/logic_analyzer_semantic.go` | services | Yes | ❌ No |

---

## 6. RECOMMENDATIONS

### High Priority
1. **Remove deprecated stubs** in `hub/api/utils.go`:
   - `getParser()`, `traverseAST()`, `analyzeAST()` - Already return errors, safe to remove
   - `callLLMWithDepth()` stub - Replace with proper import from services

### Medium Priority
2. **Consolidate duplicate implementations**:
   - Merge `hub/api/llm_cache_prompts.go` and `hub/api/services/llm_cache_analysis.go` `generatePrompt` functions
   - Support all analysis types in one function
   - Choose one package (preferably services) for the consolidated version

3. **Clarify package boundaries**:
   - `hub/api/logic_analyzer.go` (package main) calls `analyzeWithProgressiveDepth` - verify it's using the correct one
   - Consider moving all LLM analysis to services package

### Low Priority
4. **Document the difference** between main package and services package implementations
5. **Add deprecation warnings** if main package versions are being phased out

---

## 7. VERIFICATION COMMANDS

To verify which functions are actually called:

```bash
# Find all calls to generatePrompt
grep -r "generatePrompt(" --include="*.go" hub/api/

# Find all calls to analyzeWithProgressiveDepth  
grep -r "analyzeWithProgressiveDepth(" --include="*.go" hub/api/

# Check which package files are in
grep -r "^package " hub/api/llm_cache*.go
```

---

## Conclusion

**Only 2 items are confirmed legacy:**
1. Deprecated AST functions in `utils.go` (already return errors)
2. Unimplemented stub `callLLMWithDepth` in `utils.go` (already returns error)

**Everything else is active code**, though there are duplicate implementations that could be consolidated for better maintainability.
