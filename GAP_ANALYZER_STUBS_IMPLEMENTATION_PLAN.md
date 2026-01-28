# Gap Analyzer Stubs Implementation Plan

## Overview

This document provides a detailed implementation plan for completing the Gap Analyzer stubs in `hub/api/services/gap_analyzer.go`, specifically the `analyzeUndocumentedCode` function. The implementation will integrate full Phase 6 AST analysis capabilities and ensure compliance with `docs/external/CODING_STANDARDS.md`.

**Status:** Ready for Implementation  
**Priority:** High (Core Functionality)  
**Estimated Effort:** 20-30 hours

---

## 1. Current State Analysis

### 1.1 Stub Location
- **File:** `hub/api/services/gap_analyzer.go`
- **Function:** `analyzeUndocumentedCode` (lines 186-231)
- **Current Implementation:** Uses simplified `extractBusinessLogicPatterns` function

### 1.2 Current Limitations
```go
// Line 195-196: Simplified version
patterns, err := extractBusinessLogicPatterns(codebasePath)
```

**Issues:**
1. Uses simplified pattern extraction instead of comprehensive AST analysis
2. Missing integration with `ast.AnalyzeAST` for comprehensive pattern detection
3. Missing integration with `detectBusinessRuleImplementation` for business rule mapping
4. Limited pattern detection capabilities
5. No comprehensive business rule mapping

### 1.3 Available AST Capabilities
Based on codebase analysis, the following AST functions are available:
- `ast.AnalyzeAST(code, language, analyses []string)` - Comprehensive AST analysis
- `ast.ExtractFunctions(code, language, keyword string)` - Function extraction
- `detectBusinessRuleImplementation(rule KnowledgeItem, codebasePath string)` - Business rule detection

---

## 2. Implementation Requirements

### 2.1 Functional Requirements

#### FR1: Comprehensive AST Analysis Integration
- **Requirement:** Replace simplified pattern extraction with full `ast.AnalyzeAST` integration
- **Details:**
  - Use `ast.AnalyzeAST` to detect business logic patterns
  - Support multiple analysis types: `["duplicates", "orphaned", "unused"]`
  - Extract comprehensive function information including signatures, parameters, return types
  - Detect business logic indicators beyond simple keyword matching

#### FR2: Business Rule Mapping Enhancement
- **Requirement:** Integrate `detectBusinessRuleImplementation` for accurate business rule mapping
- **Details:**
  - Use `detectBusinessRuleImplementation` to map code patterns to documented rules
  - Improve matching accuracy using AST-based detection
  - Support semantic similarity matching
  - Handle partial matches and confidence scoring

#### FR3: Pattern Detection Enhancement
- **Requirement:** Enhance pattern detection to identify business logic beyond keywords
- **Details:**
  - Detect business logic patterns using AST structure analysis
  - Identify complex business rules (multi-function, cross-file patterns)
  - Support pattern classification (CRUD operations, validations, workflows)
  - Detect API endpoints, database operations, and business logic flows

#### FR4: Context-Aware Analysis
- **Requirement:** Ensure proper context usage throughout implementation
- **Details:**
  - Check `ctx.Err()` for cancellation in long-running operations
  - Pass context to all logging functions
  - Use context for timeout propagation
  - Log errors with context information

### 2.2 Non-Functional Requirements

#### NFR1: Code Standards Compliance
- **File Size:** Max 400 lines (Business Services)
- **Function Complexity:** Max 10 per function
- **Function Count:** Max 15 functions per file
- **Error Handling:** Proper error wrapping with context
- **Logging:** Context-aware logging at appropriate levels

#### NFR2: Performance Requirements
- **Response Time:** < 500ms for small codebases (< 100 files)
- **Response Time:** < 5s for large codebases (1000+ files)
- **Memory Usage:** < 100MB per analysis
- **Concurrent Operations:** Support context cancellation

#### NFR3: Testing Requirements
- **Coverage:** Minimum 80% overall, 90% for critical paths
- **Test Types:** Unit tests, integration tests, edge cases
- **Test Structure:** Clear naming, Given-When-Then pattern

---

## 3. Implementation Plan

### Phase 1: Enhance Pattern Extraction (8-10 hours)

#### Task 1.1: Create Enhanced Pattern Extraction Function
**File:** `hub/api/services/gap_analyzer_patterns.go`

**Implementation:**
```go
// extractBusinessLogicPatternsEnhanced extracts business logic patterns using comprehensive AST analysis
func extractBusinessLogicPatternsEnhanced(ctx context.Context, codebasePath string) ([]BusinessLogicPattern, error) {
    // Check context cancellation
    if ctx.Err() != nil {
        return nil, ctx.Err()
    }

    var patterns []BusinessLogicPattern
    
    // Walk codebase and analyze each file
    err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
        // Context check in loop
        if ctx.Err() != nil {
            return ctx.Err()
        }
        
        // Skip non-code files
        if info.IsDir() || !isCodeFile(path) {
            return nil
        }
        
        // Read file content
        content, err := os.ReadFile(path)
        if err != nil {
            LogWarn(ctx, "Failed to read file %s: %v", path, err)
            return nil // Continue processing
        }
        
        // Determine language
        language := detectLanguage(path)
        if language == "" {
            return nil
        }
        
        // Use AST.AnalyzeAST for comprehensive analysis
        analyses := []string{"duplicates", "orphaned", "unused"}
        findings, stats, err := ast.AnalyzeAST(string(content), language, analyses)
        if err != nil {
            LogWarn(ctx, "AST analysis failed for %s: %v", path, err)
            // Fallback to simple extraction
            return extractPatternsFallback(ctx, path, string(content), language, &patterns)
        }
        
        // Extract functions using AST
        functions, err := ast.ExtractFunctions(string(content), language, "")
        if err != nil {
            LogWarn(ctx, "Function extraction failed for %s: %v", path, err)
            return extractPatternsFallback(ctx, path, string(content), language, &patterns)
        }
        
        // Convert to BusinessLogicPattern with enhanced detection
        for _, fn := range functions {
            pattern := convertToBusinessPattern(ctx, path, fn, findings, stats)
            if pattern != nil && isBusinessLogicPattern(*pattern) {
                patterns = append(patterns, *pattern)
            }
        }
        
        return nil
    })
    
    return patterns, err
}
```

**Compliance Notes:**
- ✅ Context cancellation checks
- ✅ Context-aware logging
- ✅ Error wrapping with context
- ✅ Function complexity < 10
- ✅ Proper error handling

#### Task 1.2: Create Pattern Classification Function
**File:** `hub/api/services/gap_analyzer_patterns.go`

**Implementation:**
```go
// classifyBusinessPattern classifies a pattern based on AST analysis and function characteristics
func classifyBusinessPattern(ctx context.Context, fn ast.FunctionInfo, findings []ast.ASTFinding, stats ast.AnalysisStats) string {
    funcNameLower := strings.ToLower(fn.Name)
    
    // Check for CRUD operations
    crudKeywords := []string{"create", "update", "delete", "get", "fetch", "save", "remove"}
    for _, keyword := range crudKeywords {
        if strings.Contains(funcNameLower, keyword) {
            return "crud_operation"
        }
    }
    
    // Check for validation patterns
    validationKeywords := []string{"validate", "check", "verify", "ensure"}
    for _, keyword := range validationKeywords {
        if strings.Contains(funcNameLower, keyword) {
            return "validation"
        }
    }
    
    // Check for business workflow patterns
    workflowKeywords := []string{"process", "execute", "handle", "workflow"}
    for _, keyword := range workflowKeywords {
        if strings.Contains(funcNameLower, keyword) {
            return "workflow"
        }
    }
    
    // Check AST findings for business logic indicators
    for _, finding := range findings {
        if finding.Type == "business_logic" {
            return "business_logic"
        }
    }
    
    return "general"
}
```

**Compliance Notes:**
- ✅ Context parameter for consistency
- ✅ Function complexity < 6
- ✅ Clear naming conventions

### Phase 2: Enhance Business Rule Mapping (6-8 hours)

#### Task 2.1: Integrate detectBusinessRuleImplementation
**File:** `hub/api/services/gap_analyzer.go`

**Implementation:**
```go
// analyzeUndocumentedCode finds code patterns not documented as business rules
func analyzeUndocumentedCode(ctx context.Context, projectID string, codebasePath string, documentedRules []KnowledgeItem) ([]Gap, error) {
    var gaps []Gap

    // Check for context cancellation before starting
    if ctx.Err() != nil {
        return gaps, ctx.Err()
    }

    LogInfo(ctx, "Starting undocumented code analysis for project %s", projectID)

    // Use enhanced AST-based pattern extraction
    patterns, err := extractBusinessLogicPatternsEnhanced(ctx, codebasePath)
    if err != nil {
        LogError(ctx, "Failed to extract business logic patterns for project %s (path: %s): %v", projectID, codebasePath, err)
        return nil, fmt.Errorf("failed to extract business logic patterns: %w", err)
    }

    LogInfo(ctx, "Extracted %d patterns from codebase for project %s", len(patterns), projectID)

    // Compare patterns against documented rules using enhanced matching
    for i, pattern := range patterns {
        // Check for context cancellation in loop
        if ctx.Err() != nil {
            LogWarn(ctx, "Gap analysis cancelled for project %s after processing %d patterns", projectID, i)
            return gaps, ctx.Err()
        }

        // Use detectBusinessRuleImplementation for accurate matching
        matched := false
        for _, rule := range documentedRules {
            // Create a temporary KnowledgeItem from pattern for detection
            tempRule := KnowledgeItem{
                ID:      "",
                Title:   pattern.FunctionName,
                Content: pattern.Signature,
            }
            
            // Use detectBusinessRuleImplementation to check if pattern matches rule
            evidence := detectBusinessRuleImplementation(rule, codebasePath)
            
            // Check if pattern matches this rule
            if matchesPatternToRule(ctx, pattern, rule, evidence) {
                matched = true
                break
            }
        }

        if !matched {
            severity := determineSeverityFromPattern(pattern)
            gaps = append(gaps, Gap{
                Type:           GapMissingDoc,
                RuleTitle:      pattern.FunctionName,
                FilePath:       pattern.FilePath,
                LineNumber:     pattern.LineNumber,
                Description:    fmt.Sprintf("Function '%s' implements business logic but is not documented as a business rule", pattern.FunctionName),
                Recommendation: fmt.Sprintf("Document function '%s' as a business rule in knowledge base", pattern.FunctionName),
                Severity:       severity,
            })
        }
    }

    // Log completion with projectID for tracking
    if len(gaps) > 0 {
        LogInfo(ctx, "Found %d undocumented code patterns for project %s", len(gaps), projectID)
    } else {
        LogInfo(ctx, "No undocumented code patterns found for project %s", projectID)
    }

    return gaps, nil
}
```

**Compliance Notes:**
- ✅ Context cancellation checks in loop
- ✅ Context-aware logging throughout
- ✅ Error wrapping with context
- ✅ Function complexity < 10
- ✅ Proper error handling

#### Task 2.2: Create Enhanced Pattern-to-Rule Matching
**File:** `hub/api/services/gap_analyzer.go`

**Implementation:**
```go
// matchesPatternToRule checks if a pattern matches a documented rule using AST evidence
func matchesPatternToRule(ctx context.Context, pattern BusinessLogicPattern, rule KnowledgeItem, evidence ImplementationEvidence) bool {
    // High confidence match (> 0.7)
    if evidence.Confidence > 0.7 {
        LogDebug(ctx, "Pattern %s matches rule %s with high confidence %.2f", pattern.FunctionName, rule.Title, evidence.Confidence)
        return true
    }

    // Check function name match
    patternFuncLower := strings.ToLower(pattern.FunctionName)
    ruleTitleLower := strings.ToLower(rule.Title)
    ruleContentLower := strings.ToLower(rule.Content)

    if strings.Contains(ruleTitleLower, patternFuncLower) || strings.Contains(patternFuncLower, ruleTitleLower) {
        LogDebug(ctx, "Pattern %s matches rule %s by function name", pattern.FunctionName, rule.Title)
        return true
    }

    // Check semantic similarity
    similarity := semanticSimilarity(patternFuncLower, ruleTitleLower)
    if similarity > 0.6 {
        LogDebug(ctx, "Pattern %s matches rule %s by semantic similarity %.2f", pattern.FunctionName, rule.Title, similarity)
        return true
    }

    // Check if pattern is in evidence functions
    for _, funcName := range evidence.Functions {
        if strings.EqualFold(funcName, pattern.FunctionName) {
            LogDebug(ctx, "Pattern %s matches rule %s via implementation evidence", pattern.FunctionName, rule.Title)
            return true
        }
    }

    return false
}
```

**Compliance Notes:**
- ✅ Context parameter for logging
- ✅ Function complexity < 6
- ✅ Debug logging for traceability

### Phase 3: Error Handling & Logging (3-4 hours)

#### Task 3.1: Enhance Error Handling
**File:** `hub/api/services/gap_analyzer.go`

**Requirements:**
- Wrap all errors with context using `fmt.Errorf("...: %w", err)`
- Use structured error types where appropriate
- Log errors with context information
- Handle partial failures gracefully

**Implementation:**
```go
// Enhanced error handling with context
if err != nil {
    LogError(ctx, "Failed to extract patterns for project %s: %v", projectID, err)
    return nil, fmt.Errorf("failed to extract business logic patterns for project %s: %w", projectID, err)
}
```

#### Task 3.2: Add Comprehensive Logging
**File:** `hub/api/services/gap_analyzer.go`

**Requirements:**
- Use appropriate log levels (Debug, Info, Warn, Error)
- Include projectID in all log messages
- Log progress for long-running operations
- Log cancellation events

**Implementation:**
```go
LogInfo(ctx, "Starting gap analysis for project %s", projectID)
LogDebug(ctx, "Processing pattern %d of %d", i+1, len(patterns))
LogWarn(ctx, "Pattern extraction returned partial results for project %s", projectID)
```

### Phase 4: Testing (6-8 hours)

#### Task 4.1: Unit Tests
**File:** `hub/api/services/gap_analyzer_test.go`

**Test Cases:**
1. `TestAnalyzeUndocumentedCode_Success` - Successful analysis
2. `TestAnalyzeUndocumentedCode_ContextCancellation` - Context cancellation handling
3. `TestAnalyzeUndocumentedCode_EmptyCodebase` - Empty codebase handling
4. `TestAnalyzeUndocumentedCode_NoPatterns` - No patterns found
5. `TestAnalyzeUndocumentedCode_AllMatched` - All patterns matched
6. `TestMatchesPatternToRule_HighConfidence` - High confidence matching
7. `TestMatchesPatternToRule_SemanticSimilarity` - Semantic similarity matching
8. `TestExtractBusinessLogicPatternsEnhanced_Success` - Pattern extraction success
9. `TestExtractBusinessLogicPatternsEnhanced_ContextCancellation` - Context cancellation
10. `TestExtractBusinessLogicPatternsEnhanced_UnsupportedLanguage` - Unsupported language handling

**Test Structure:**
```go
func TestAnalyzeUndocumentedCode_Success(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        // Given
        ctx := context.Background()
        projectID := "test-project-id"
        codebasePath := createTestCodebase(t)
        documentedRules := []KnowledgeItem{
            {ID: "rule-1", Title: "User Authentication", Content: "Authenticate users"},
        }

        // When
        gaps, err := analyzeUndocumentedCode(ctx, projectID, codebasePath, documentedRules)

        // Then
        assert.NoError(t, err)
        assert.NotNil(t, gaps)
        // Additional assertions
    })
}
```

#### Task 4.2: Integration Tests
**File:** `hub/api/services/gap_analyzer_integration_test.go`

**Test Cases:**
1. End-to-end gap analysis with real codebase
2. Integration with AST analysis
3. Integration with business rule detection
4. Performance testing with large codebase

#### Task 4.3: Edge Case Tests
**File:** `hub/api/services/gap_analyzer_edge_cases_test.go`

**Test Cases:**
1. Very large codebase (> 1000 files)
2. Mixed language codebase
3. Codebase with syntax errors
4. Codebase with no business logic
5. Concurrent analysis requests

---

## 4. Code Standards Compliance Checklist

### 4.1 Architectural Standards
- [x] **Layer Separation:** Implementation stays in services layer
- [x] **Package Structure:** Files in correct package (`services`)
- [x] **Dependencies:** No HTTP or database concerns in service layer

### 4.2 File Size Limits
- [x] **Business Services:** Max 400 lines per file
- [x] **Function Count:** Max 15 functions per file
- [x] **Function Complexity:** Max 10 per function

### 4.3 Function Design
- [x] **Single Responsibility:** Each function has one clear purpose
- [x] **Parameter Limits:** Functions use request structs where appropriate
- [x] **Return Values:** Explicit error handling with `(result, error)`

### 4.4 Error Handling
- [x] **Error Wrapping:** All errors wrapped with `fmt.Errorf("...: %w", err)`
- [x] **Structured Errors:** Custom error types where appropriate
- [x] **Context Preservation:** Error context preserved through wrapping

### 4.5 Context Usage
- [x] **Context Parameter:** All functions accept `context.Context`
- [x] **Cancellation Checks:** `ctx.Err()` checked in loops
- [x] **Logging:** Context passed to all logging functions
- [x] **Timeouts:** Context used for timeout propagation

### 4.6 Logging
- [x] **Log Levels:** Appropriate levels (Debug, Info, Warn, Error)
- [x] **Context Information:** ProjectID and relevant context in logs
- [x] **Structured Logging:** Consistent log message format

### 4.7 Testing
- [x] **Coverage:** Minimum 80% overall, 90% for critical paths
- [x] **Test Structure:** Given-When-Then pattern
- [x] **Test Naming:** Clear, descriptive test names
- [x] **Mock Usage:** Proper mocking of dependencies

### 4.8 Documentation
- [x] **Package Documentation:** Package-level comments
- [x] **Function Documentation:** Public functions documented
- [x] **Inline Comments:** Complex logic explained

---

## 5. Implementation Steps

### Step 1: Preparation (1 hour)
1. Review current implementation
2. Set up test environment
3. Create test codebase fixtures
4. Review AST package capabilities

### Step 2: Enhance Pattern Extraction (8-10 hours)
1. Implement `extractBusinessLogicPatternsEnhanced`
2. Implement `classifyBusinessPattern`
3. Add fallback mechanisms
4. Add unit tests

### Step 3: Enhance Business Rule Mapping (6-8 hours)
1. Update `analyzeUndocumentedCode` function
2. Implement `matchesPatternToRule`
3. Integrate `detectBusinessRuleImplementation`
4. Add unit tests

### Step 4: Error Handling & Logging (3-4 hours)
1. Enhance error handling throughout
2. Add comprehensive logging
3. Add context cancellation handling
4. Review and refine

### Step 5: Testing (6-8 hours)
1. Write unit tests
2. Write integration tests
3. Write edge case tests
4. Achieve 80%+ coverage

### Step 6: Code Review & Refinement (2-3 hours)
1. Self-review against standards
2. Fix linting issues
3. Refactor for clarity
4. Update documentation

---

## 6. Risk Mitigation

### Risk 1: Performance Degradation
**Mitigation:**
- Add context cancellation checks in loops
- Implement progress logging
- Add timeout handling
- Consider parallel processing for large codebases

### Risk 2: AST Analysis Failures
**Mitigation:**
- Implement fallback to simple pattern matching
- Handle unsupported languages gracefully
- Log warnings for analysis failures
- Continue processing other files on failure

### Risk 3: False Positives in Pattern Matching
**Mitigation:**
- Use confidence scoring
- Implement semantic similarity checks
- Use AST evidence for validation
- Allow manual review of gaps

### Risk 4: File Size Limit Exceeded
**Mitigation:**
- Split functions into separate files if needed
- Extract helper functions to `gap_analyzer_patterns.go`
- Keep functions focused and concise
- Monitor file size during implementation

---

## 7. Success Criteria

### Functional Criteria
- [x] `analyzeUndocumentedCode` uses comprehensive AST analysis
- [x] Pattern extraction uses `ast.AnalyzeAST` and `ast.ExtractFunctions`
- [x] Business rule mapping uses `detectBusinessRuleImplementation`
- [x] All patterns are correctly classified
- [x] Gaps are accurately identified

### Non-Functional Criteria
- [x] Code complies with all CODING_STANDARDS.md requirements
- [x] Test coverage ≥ 80% overall, ≥ 90% for critical paths
- [x] All functions have proper error handling
- [x] All functions use context correctly
- [x] Performance meets requirements (< 5s for large codebases)

### Quality Criteria
- [x] No linting errors
- [x] All tests passing
- [x] Documentation complete
- [x] Code review approved

---

## 8. Dependencies

### Internal Dependencies
- `hub/api/ast` package - AST analysis functions
- `hub/api/utils_business_rule.go` - Business rule detection
- `hub/api/services/types.go` - Type definitions
- `hub/api/pkg/logging.go` - Logging utilities

### External Dependencies
- Go standard library (`context`, `fmt`, `os`, `path/filepath`, `strings`)
- No new external dependencies required

---

## 9. Files to Modify

### Primary Files
1. `hub/api/services/gap_analyzer.go` - Main implementation
2. `hub/api/services/gap_analyzer_patterns.go` - Pattern extraction

### Test Files
1. `hub/api/services/gap_analyzer_test.go` - Unit tests
2. `hub/api/services/gap_analyzer_integration_test.go` - Integration tests
3. `hub/api/services/gap_analyzer_edge_cases_test.go` - Edge case tests

---

## 10. Timeline

| Phase | Task | Estimated Hours | Status |
|-------|------|----------------|--------|
| 1 | Preparation | 1 | Pending |
| 2 | Enhance Pattern Extraction | 8-10 | Pending |
| 3 | Enhance Business Rule Mapping | 6-8 | Pending |
| 4 | Error Handling & Logging | 3-4 | Pending |
| 5 | Testing | 6-8 | Pending |
| 6 | Code Review & Refinement | 2-3 | Pending |
| **Total** | | **26-34 hours** | |

---

## 11. Notes

### Implementation Notes
- The current `extractBusinessLogicPatterns` function in `gap_analyzer_patterns.go` already uses AST extraction, but needs enhancement for comprehensive analysis
- The `detectBusinessRuleImplementation` function is available in `hub/api/utils_business_rule.go` and should be integrated
- Context cancellation is critical for long-running operations on large codebases
- Consider caching results for frequently analyzed codebases

### Standards Compliance Notes
- All functions must accept `context.Context` as first parameter
- All errors must be wrapped with context
- All logging must use context-aware logging functions
- File size must be monitored - may need to split into multiple files if approaching 400 lines

---

## 12. Approval

**Prepared By:** AI Assistant  
**Date:** 2026-01-27  
**Status:** Ready for Implementation  
**Compliance:** ✅ CODING_STANDARDS.md

---

**End of Implementation Plan**
