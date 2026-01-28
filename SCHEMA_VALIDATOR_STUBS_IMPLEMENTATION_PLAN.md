# Schema Validator Stubs Implementation Plan

## Overview

This document provides a detailed step-by-step plan for implementing the schema validator stubs, specifically focusing on AST-based security middleware validation. The implementation will fully comply with `docs/external/CODING_STANDARDS.md`.

## Current State Analysis

### Location
- **File:** `hub/api/services/schema_validator.go`
- **Function:** `validateSecurity` (line 315-368)
- **Current Implementation:** Simplified validation that only checks `len(endpoint.Auth) > 0`

### Missing Functionality
1. **AST-based security middleware verification** - Currently only checks endpoint metadata
2. **Actual middleware detection in code** - No AST analysis to verify middleware exists
3. **Security pattern matching** - No pattern-based detection of security implementations

### Available Resources
- ✅ `ast.AnalyzeAST(code, language, analyses []string)` - Comprehensive AST analysis
- ✅ `ast.ExtractFunctions(code, language, keyword string)` - Function extraction
- ✅ Security middleware patterns in `hub/api/middleware/security.go`
- ✅ Endpoint information includes file path for AST analysis

---

## Implementation Plan

### Phase 1: Preparation and Setup (Day 1)

#### Step 1.1: Review Current Code Structure
**Objective:** Understand the current implementation and dependencies

**Tasks:**
1. Review `hub/api/services/schema_validator.go` structure
2. Identify all dependencies and imports
3. Review `hub/api/middleware/security.go` to understand security patterns
4. Review AST package API (`hub/api/ast/`) to understand available functions
5. Review `EndpointInfo` struct to understand available data

**Deliverables:**
- Document current code structure
- List all dependencies
- Identify security middleware patterns to detect

**Compliance Check:**
- ✅ File size: Current file is 486 lines (within 400 line limit for business services - **NOTE: May need file splitting**)
- ✅ Function complexity: `validateSecurity` is currently simple
- ✅ Error handling: Uses context cancellation

#### Step 1.2: Define Security Middleware Patterns
**Objective:** Create a comprehensive list of security patterns to detect

**Tasks:**
1. Analyze `hub/api/middleware/security.go` for security patterns:
   - Authentication middleware (JWT, OAuth, API keys)
   - Authorization middleware (role-based, permission-based)
   - Rate limiting middleware
   - CORS middleware
   - Input validation middleware
   - Security headers middleware
2. Document Go-specific patterns:
   - Middleware function signatures: `func(http.Handler) http.Handler`
   - Middleware registration patterns
   - Security function calls
3. Create pattern detection rules

**Deliverables:**
- Security pattern specification document
- Pattern detection rules for AST analysis

**Compliance Check:**
- ✅ Documentation standards: All patterns documented with examples
- ✅ Naming conventions: Clear, descriptive pattern names

---

### Phase 2: AST Integration Design (Day 1-2)

#### Step 2.1: Design AST Analysis Strategy
**Objective:** Design how AST analysis will be integrated into security validation

**Tasks:**
1. Design AST analysis flow:
   ```
   For each endpoint:
     1. Read source file (endpoint.File)
     2. Parse code with AST
     3. Extract functions in endpoint handler
     4. Analyze function calls for security middleware
     5. Match detected patterns against contract requirements
     6. Generate findings for missing/incomplete security
   ```

2. Define AST analysis parameters:
   - Language: Detect from file extension or endpoint metadata
   - Analyses: `["security", "middleware", "authentication", "authorization"]`
   - Scope: Handler function and middleware chain

3. Design error handling:
   - File read errors
   - AST parsing errors
   - Context cancellation
   - Missing file errors

**Deliverables:**
- AST integration design document
- Error handling strategy

**Compliance Check:**
- ✅ Error wrapping: All errors wrapped with context (`fmt.Errorf("...: %w", err)`)
- ✅ Context usage: All functions check `ctx.Err()` for cancellation
- ✅ Logging: Errors logged with context using appropriate log levels

#### Step 2.2: Create Security Pattern Detection Module
**Objective:** Create a reusable module for detecting security patterns in AST

**Tasks:**
1. Create new file: `hub/api/services/schema_validator_security_patterns.go`
   - **Rationale:** Keep file size under 400 lines (CODING_STANDARDS.md)
   - **Purpose:** Security pattern detection logic

2. Implement pattern detection functions:
   ```go
   // detectSecurityMiddleware detects security middleware in AST
   func detectSecurityMiddleware(ctx context.Context, code string, language string) ([]SecurityPattern, error)
   
   // matchSecurityScheme matches detected patterns to contract schemes
   func matchSecurityScheme(patterns []SecurityPattern, contractScheme string) bool
   
   // SecurityPattern represents a detected security pattern
   type SecurityPattern struct {
       Type        string // "authentication", "authorization", "rate_limit", etc.
       Scheme      string // "BearerAuth", "ApiKeyAuth", etc.
       Location    string // File and line number
       Confidence  float64 // 0.0-1.0 confidence score
   }
   ```

3. Implement pattern-specific detectors:
   - JWT/Bearer token detection
   - API key detection
   - OAuth detection
   - Role-based access control detection
   - Rate limiting detection

**Deliverables:**
- New file: `schema_validator_security_patterns.go`
- Pattern detection functions
- Security pattern types

**Compliance Check:**
- ✅ File size: New file should be < 400 lines
- ✅ Function size: Each function < 50 lines, complexity < 10
- ✅ Single responsibility: Each function has one clear purpose
- ✅ Naming: Clear, descriptive function names
- ✅ Documentation: All exported functions have godoc comments

---

### Phase 3: Core Implementation (Day 2-3)

#### Step 3.1: Enhance `validateSecurity` Function
**Objective:** Replace simplified validation with AST-based verification

**Tasks:**
1. Update function signature (if needed):
   ```go
   func validateSecurity(ctx context.Context, endpoint EndpointInfo, contract ContractEndpoint) []APILayerFinding
   ```
   - Keep existing signature for backward compatibility
   - Add internal helper for AST analysis

2. Implement AST-based validation:
   ```go
   func validateSecurity(ctx context.Context, endpoint EndpointInfo, contract ContractEndpoint) []APILayerFinding {
       findings := []APILayerFinding{}
       
       // Check context cancellation
       if ctx.Err() != nil {
           return findings
       }
       
       if len(contract.Security) == 0 {
           return findings
       }
       
       // Read endpoint source file
       code, err := readEndpointSource(ctx, endpoint.File)
       if err != nil {
           // Log error and return finding
           findings = append(findings, createFileReadErrorFinding(endpoint, err))
           // Fall back to metadata-based validation
           return validateSecurityMetadata(ctx, endpoint, contract, findings)
       }
       
       // Detect language
       language := detectLanguageFromFile(endpoint.File)
       
       // Perform AST analysis
       patterns, err := detectSecurityMiddleware(ctx, code, language)
       if err != nil {
           // Log error and fall back
           LogWarn(ctx, "AST analysis failed for security validation: %v", err)
           return validateSecurityMetadata(ctx, endpoint, contract, findings)
       }
       
       // Match patterns against contract requirements
       for _, contractSec := range contract.Security {
           if ctx.Err() != nil {
               return findings
           }
           
           for _, scheme := range contractSec.Schemes {
               matched := matchSecurityScheme(patterns, scheme)
               if !matched {
                   findings = append(findings, createMissingSecurityFinding(endpoint, scheme, contract))
               }
           }
       }
       
       return findings
   }
   ```

3. Implement helper functions:
   - `readEndpointSource(ctx, filePath string) (string, error)`
   - `detectLanguageFromFile(filePath string) string`
   - `validateSecurityMetadata(ctx, endpoint, contract, existingFindings) []APILayerFinding`
   - `createFileReadErrorFinding(endpoint, err) APILayerFinding`
   - `createMissingSecurityFinding(endpoint, scheme, contract) APILayerFinding`

**Deliverables:**
- Updated `validateSecurity` function
- Helper functions for file reading and error handling

**Compliance Check:**
- ✅ Function size: `validateSecurity` should remain < 80 lines
- ✅ Complexity: Cyclomatic complexity < 10
- ✅ Error handling: All errors wrapped with context
- ✅ Context usage: Checks `ctx.Err()` in loops
- ✅ Logging: Uses `LogWarn`/`LogError` with context

#### Step 3.2: Implement File Reading with Error Handling
**Objective:** Safely read endpoint source files

**Tasks:**
1. Implement `readEndpointSource`:
   ```go
   func readEndpointSource(ctx context.Context, filePath string) (string, error) {
       if ctx.Err() != nil {
           return "", ctx.Err()
       }
       
       // Resolve file path (handle relative/absolute)
       absPath, err := resolveFilePath(filePath)
       if err != nil {
           return "", fmt.Errorf("failed to resolve file path %s: %w", filePath, err)
       }
       
       // Check if file exists
       if _, err := os.Stat(absPath); os.IsNotExist(err) {
           return "", fmt.Errorf("endpoint source file not found: %s", absPath)
       }
       
       // Read file with context cancellation support
       data, err := readFileWithContext(ctx, absPath)
       if err != nil {
           return "", fmt.Errorf("failed to read endpoint source file %s: %w", absPath, err)
       }
       
       return string(data), nil
   }
   ```

2. Implement `readFileWithContext`:
   ```go
   func readFileWithContext(ctx context.Context, filePath string) ([]byte, error) {
       // Use goroutine with context cancellation
       var data []byte
       var readErr error
       done := make(chan struct{})
       
       go func() {
           defer close(done)
           data, readErr = os.ReadFile(filePath)
       }()
       
       select {
       case <-ctx.Done():
           return nil, ctx.Err()
       case <-done:
           return data, readErr
       }
   }
   ```

3. Implement `resolveFilePath`:
   ```go
   func resolveFilePath(filePath string) (string, error) {
       if filepath.IsAbs(filePath) {
           return filePath, nil
       }
       
       // Try to resolve relative to project root
       // This may need project root detection logic
       absPath, err := filepath.Abs(filePath)
       if err != nil {
           return "", fmt.Errorf("failed to resolve relative path: %w", err)
       }
       
       return absPath, nil
   }
   ```

**Deliverables:**
- File reading functions with context support
- Path resolution logic

**Compliance Check:**
- ✅ Error wrapping: All errors wrapped with context
- ✅ Context cancellation: Properly handles context cancellation
- ✅ Resource cleanup: No resource leaks

#### Step 3.3: Implement Language Detection
**Objective:** Detect programming language from file path

**Tasks:**
1. Implement `detectLanguageFromFile`:
   ```go
   func detectLanguageFromFile(filePath string) string {
       ext := filepath.Ext(filePath)
       switch ext {
       case ".go":
           return "go"
       case ".js", ".jsx":
           return "javascript"
       case ".ts", ".tsx":
           return "typescript"
       case ".py":
           return "python"
       case ".java":
           return "java"
       default:
           // Default to Go for hub/api codebase
           return "go"
       }
   }
   ```

**Deliverables:**
- Language detection function

**Compliance Check:**
- ✅ Function simplicity: Single responsibility, clear logic

---

### Phase 4: Pattern Detection Implementation (Day 3-4)

#### Step 4.1: Implement Core Pattern Detection
**Objective:** Implement AST-based security pattern detection

**Tasks:**
1. Implement `detectSecurityMiddleware`:
   ```go
   func detectSecurityMiddleware(ctx context.Context, code string, language string) ([]SecurityPattern, error) {
       if ctx.Err() != nil {
           return nil, ctx.Err()
       }
       
       patterns := []SecurityPattern{}
       
       // Use AST analysis for security patterns
       analyses := []string{"security", "middleware", "authentication", "authorization"}
       findings, _, err := ast.AnalyzeAST(code, language, analyses)
       if err != nil {
           return nil, fmt.Errorf("AST analysis failed: %w", err)
       }
       
       // Extract security patterns from AST findings
       for _, finding := range findings {
           if ctx.Err() != nil {
               return patterns, ctx.Err()
           }
           
           pattern := extractSecurityPatternFromFinding(finding)
           if pattern != nil {
               patterns = append(patterns, *pattern)
           }
       }
       
       // Also extract functions to find middleware registrations
       functions, err := ast.ExtractFunctions(code, language, "")
       if err != nil {
           // Log but don't fail - we have findings from AnalyzeAST
           LogWarn(ctx, "Function extraction failed: %v", err)
       } else {
           // Analyze functions for middleware patterns
           middlewarePatterns := detectMiddlewareInFunctions(ctx, functions, code, language)
           patterns = append(patterns, middlewarePatterns...)
       }
       
       return patterns, nil
   }
   ```

2. Implement `extractSecurityPatternFromFinding`:
   ```go
   func extractSecurityPatternFromFinding(finding ast.ASTFinding) *SecurityPattern {
       // Map AST finding types to security patterns
       switch finding.Type {
       case "jwt_auth", "bearer_token":
           return &SecurityPattern{
               Type:       "authentication",
               Scheme:     "BearerAuth",
               Location:   finding.Location,
               Confidence: 0.9,
           }
       case "api_key":
           return &SecurityPattern{
               Type:       "authentication",
               Scheme:     "ApiKeyAuth",
               Location:   finding.Location,
               Confidence: 0.8,
           }
       case "oauth":
           return &SecurityPattern{
               Type:       "authentication",
               Scheme:     "OAuth2",
               Location:   finding.Location,
               Confidence: 0.85,
           }
       case "rbac", "role_based":
           return &SecurityPattern{
               Type:       "authorization",
               Scheme:     "RBAC",
               Location:   finding.Location,
               Confidence: 0.9,
           }
       default:
           return nil
       }
   }
   ```

**Deliverables:**
- Core pattern detection function
- Pattern extraction from AST findings

**Compliance Check:**
- ✅ AST integration: Uses `ast.AnalyzeAST` and `ast.ExtractFunctions`
- ✅ Error handling: Proper error wrapping
- ✅ Context usage: Checks context cancellation

#### Step 4.2: Implement Middleware Detection in Functions
**Objective:** Detect middleware patterns in function definitions

**Tasks:**
1. Implement `detectMiddlewareInFunctions`:
   ```go
   func detectMiddlewareInFunctions(ctx context.Context, functions []ast.FunctionInfo, code string, language string) []SecurityPattern {
       patterns := []SecurityPattern{}
       
       // Common middleware function name patterns
       middlewarePatterns := []struct {
           namePattern string
           scheme      string
           patternType string
       }{
           {"Auth", "BearerAuth", "authentication"},
           {"Authenticate", "BearerAuth", "authentication"},
           {"JWT", "BearerAuth", "authentication"},
           {"APIKey", "ApiKeyAuth", "authentication"},
           {"OAuth", "OAuth2", "authentication"},
           {"Authorize", "RBAC", "authorization"},
           {"RBAC", "RBAC", "authorization"},
           {"RateLimit", "RateLimit", "rate_limit"},
           {"CORS", "CORS", "cors"},
       }
       
       for _, fn := range functions {
           if ctx.Err() != nil {
               return patterns
           }
           
           fnName := strings.ToLower(fn.Name)
           
           for _, pattern := range middlewarePatterns {
               if strings.Contains(fnName, strings.ToLower(pattern.namePattern)) {
                   // Check if function has middleware signature
                   if isMiddlewareFunction(fn, code, language) {
                       patterns = append(patterns, SecurityPattern{
                           Type:       pattern.patternType,
                           Scheme:     pattern.scheme,
                           Location:   fn.Location,
                           Confidence: 0.7, // Lower confidence for name-based detection
                       })
                   }
               }
           }
       }
       
       return patterns
   }
   ```

2. Implement `isMiddlewareFunction`:
   ```go
   func isMiddlewareFunction(fn ast.FunctionInfo, code string, language string) bool {
       // For Go: Check if function signature matches middleware pattern
       // func(http.Handler) http.Handler
       if language == "go" {
           // Parse function signature from code
           // This is a simplified check - full implementation would parse AST
           return strings.Contains(fn.Signature, "http.Handler") && 
                  strings.Contains(fn.Signature, "func")
       }
       
       // For other languages, implement language-specific checks
       return false
   }
   ```

**Deliverables:**
- Middleware detection in functions
- Middleware signature validation

**Compliance Check:**
- ✅ Function complexity: Each function < 50 lines
- ✅ Pattern matching: Clear, maintainable patterns

#### Step 4.3: Implement Scheme Matching
**Objective:** Match detected patterns to contract security schemes

**Tasks:**
1. Implement `matchSecurityScheme`:
   ```go
   func matchSecurityScheme(patterns []SecurityPattern, contractScheme string) bool {
       // Normalize scheme names for comparison
       normalizedContract := normalizeSchemeName(contractScheme)
       
       for _, pattern := range patterns {
           normalizedPattern := normalizeSchemeName(pattern.Scheme)
           
           // Exact match
           if normalizedPattern == normalizedContract {
               return true
           }
           
           // Partial match (e.g., "Bearer" matches "BearerAuth")
           if strings.Contains(normalizedContract, normalizedPattern) ||
              strings.Contains(normalizedPattern, normalizedContract) {
               // Check confidence threshold
               if pattern.Confidence >= 0.7 {
                   return true
               }
           }
       }
       
       return false
   }
   ```

2. Implement `normalizeSchemeName`:
   ```go
   func normalizeSchemeName(scheme string) string {
       // Remove common suffixes/prefixes
       normalized := strings.ToLower(scheme)
       normalized = strings.TrimSuffix(normalized, "auth")
       normalized = strings.TrimSuffix(normalized, "authentication")
       normalized = strings.TrimPrefix(normalized, "api")
       return strings.TrimSpace(normalized)
   }
   ```

**Deliverables:**
- Scheme matching logic
- Scheme name normalization

**Compliance Check:**
- ✅ Function simplicity: Clear, testable logic

---

### Phase 5: Error Handling and Fallback (Day 4)

#### Step 5.1: Implement Metadata-Based Fallback
**Objective:** Provide fallback validation when AST analysis fails

**Tasks:**
1. Implement `validateSecurityMetadata`:
   ```go
   func validateSecurityMetadata(ctx context.Context, endpoint EndpointInfo, contract ContractEndpoint, existingFindings []APILayerFinding) []APILayerFinding {
       findings := existingFindings
       
       // Use existing simplified validation as fallback
       hasSecurity := len(endpoint.Auth) > 0
       if !hasSecurity && len(contract.Security) > 0 {
           findings = append(findings, APILayerFinding{
               Type:         "contract_mismatch",
               Location:     endpoint.File,
               Issue:        fmt.Sprintf("Security requirements defined in contract but not found in endpoint metadata for %s %s", endpoint.Method, endpoint.Path),
               Severity:     "critical",
               ContractPath: fmt.Sprintf("#/paths/%s/%s/security", normalizePathForJSONPath(endpoint.Path), strings.ToLower(endpoint.Method)),
               SuggestedFix: fmt.Sprintf("Add security implementation to endpoint %s %s", endpoint.Method, endpoint.Path),
               Details: map[string]string{
                   "validation_method": "metadata_fallback",
                   "note": "AST analysis unavailable, using metadata-based validation",
               },
           })
       }
       
       // Validate security schemes match (existing logic)
       for _, contractSec := range contract.Security {
           if ctx.Err() != nil {
               return findings
           }
           
           for _, scheme := range contractSec.Schemes {
               schemeFound := false
               for _, endpointAuth := range endpoint.Auth {
                   if strings.EqualFold(endpointAuth, scheme) {
                       schemeFound = true
                       break
                   }
               }
               
               if !schemeFound {
                   findings = append(findings, APILayerFinding{
                       Type:         "contract_mismatch",
                       Location:     endpoint.File,
                       Issue:        fmt.Sprintf("Security scheme '%s' required by contract but not found in endpoint %s %s", scheme, endpoint.Method, endpoint.Path),
                       Severity:     "critical",
                       ContractPath: fmt.Sprintf("#/paths/%s/%s/security/%s", normalizePathForJSONPath(endpoint.Path), strings.ToLower(endpoint.Method), scheme),
                       SuggestedFix: fmt.Sprintf("Add security scheme '%s' to endpoint %s %s", scheme, endpoint.Method, endpoint.Path),
                   })
               }
           }
       }
       
       return findings
   }
   ```

**Deliverables:**
- Metadata-based fallback validation

**Compliance Check:**
- ✅ Backward compatibility: Maintains existing behavior as fallback
- ✅ Error handling: Graceful degradation

#### Step 5.2: Implement Error Finding Creation
**Objective:** Create appropriate findings for various error conditions

**Tasks:**
1. Implement `createFileReadErrorFinding`:
   ```go
   func createFileReadErrorFinding(endpoint EndpointInfo, err error) APILayerFinding {
       return APILayerFinding{
           Type:     "validation_error",
           Location: endpoint.File,
           Issue:    fmt.Sprintf("Failed to read endpoint source file for security validation: %v", err),
           Severity: "medium",
           Details: map[string]string{
               "error_type": "file_read_error",
               "endpoint":   fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path),
           },
       }
   }
   ```

2. Implement `createMissingSecurityFinding`:
   ```go
   func createMissingSecurityFinding(endpoint EndpointInfo, scheme string, contract ContractEndpoint) APILayerFinding {
       return APILayerFinding{
           Type:         "contract_mismatch",
           Location:     endpoint.File,
           Issue:        fmt.Sprintf("Security scheme '%s' required by contract but not detected in code for endpoint %s %s", scheme, endpoint.Method, endpoint.Path),
           Severity:     "critical",
           ContractPath: fmt.Sprintf("#/paths/%s/%s/security/%s", normalizePathForJSONPath(endpoint.Path), strings.ToLower(endpoint.Method), scheme),
           SuggestedFix: fmt.Sprintf("Add security scheme '%s' implementation to endpoint %s %s", scheme, endpoint.Method, endpoint.Path),
           Details: map[string]string{
               "validation_method": "ast_analysis",
               "scheme":            scheme,
           },
       }
   }
   ```

**Deliverables:**
- Error finding creation functions

**Compliance Check:**
- ✅ Error messages: Clear, actionable error messages
- ✅ Finding structure: Consistent with existing findings

---

### Phase 6: Testing (Day 5)

#### Step 6.1: Unit Tests for Pattern Detection
**Objective:** Test security pattern detection functions

**Tasks:**
1. Create test file: `hub/api/services/schema_validator_security_patterns_test.go`

2. Test cases:
   ```go
   func TestDetectSecurityMiddleware_JWT(t *testing.T)
   func TestDetectSecurityMiddleware_APIKey(t *testing.T)
   func TestDetectSecurityMiddleware_OAuth(t *testing.T)
   func TestDetectSecurityMiddleware_RBAC(t *testing.T)
   func TestMatchSecurityScheme_ExactMatch(t *testing.T)
   func TestMatchSecurityScheme_PartialMatch(t *testing.T)
   func TestMatchSecurityScheme_NoMatch(t *testing.T)
   ```

3. Test data:
   - Sample Go code with JWT middleware
   - Sample Go code with API key middleware
   - Sample Go code with OAuth middleware
   - Sample Go code with RBAC middleware
   - Sample Go code without security

**Deliverables:**
- Test file with comprehensive test cases
- Test coverage > 90% (CODING_STANDARDS.md requirement)

**Compliance Check:**
- ✅ Test coverage: > 90% for new code
- ✅ Test structure: Clear test naming and structure
- ✅ Test data: Realistic test scenarios

#### Step 6.2: Integration Tests for `validateSecurity`
**Objective:** Test the complete security validation flow

**Tasks:**
1. Update existing test file: `hub/api/services/schema_validator_test.go`

2. Add test cases:
   ```go
   func TestValidateSecurity_ASTBased_JWTFound(t *testing.T)
   func TestValidateSecurity_ASTBased_JWTMissing(t *testing.T)
   func TestValidateSecurity_ASTBased_MultipleSchemes(t *testing.T)
   func TestValidateSecurity_Fallback_FileNotFound(t *testing.T)
   func TestValidateSecurity_Fallback_ASTError(t *testing.T)
   func TestValidateSecurity_ContextCancellation(t *testing.T)
   ```

3. Test scenarios:
   - Endpoint with matching security (should pass)
   - Endpoint with missing security (should fail)
   - Endpoint with partial security (should report missing)
   - File not found (should fall back to metadata)
   - AST parsing error (should fall back to metadata)
   - Context cancellation (should return early)

**Deliverables:**
- Updated test file with integration tests
- Test coverage > 90%

**Compliance Check:**
- ✅ Test coverage: > 90%
- ✅ Edge cases: Context cancellation, errors, fallbacks
- ✅ Mock usage: Proper mocking of dependencies

#### Step 6.3: Performance Tests
**Objective:** Ensure AST analysis doesn't significantly impact performance

**Tasks:**
1. Create benchmark tests:
   ```go
   func BenchmarkValidateSecurity_ASTBased(b *testing.B)
   func BenchmarkDetectSecurityMiddleware(b *testing.B)
   ```

2. Performance targets (from CODING_STANDARDS.md):
   - Simple CRUD: < 100ms target, < 500ms max
   - Complex queries: < 500ms target, < 2s max

**Deliverables:**
- Benchmark tests
- Performance metrics

**Compliance Check:**
- ✅ Performance: Meets response time requirements

---

### Phase 7: Documentation and Code Review (Day 5-6)

#### Step 7.1: Code Documentation
**Objective:** Document all new functions and types

**Tasks:**
1. Add godoc comments to all exported functions:
   ```go
   // detectSecurityMiddleware performs AST-based analysis to detect security
   // middleware patterns in the provided code.
   //
   // It uses ast.AnalyzeAST and ast.ExtractFunctions to identify authentication,
   // authorization, and other security-related patterns. Returns a slice of
   // SecurityPattern structs representing detected security implementations.
   //
   // The function respects context cancellation and returns an error if AST
   // analysis fails or context is cancelled.
   func detectSecurityMiddleware(ctx context.Context, code string, language string) ([]SecurityPattern, error)
   ```

2. Document types:
   ```go
   // SecurityPattern represents a detected security implementation pattern
   // in the analyzed code.
   type SecurityPattern struct {
       // Type indicates the category of security pattern (e.g., "authentication", "authorization")
       Type string
       
       // Scheme is the security scheme name (e.g., "BearerAuth", "ApiKeyAuth")
       Scheme string
       
       // Location is the file path and line number where the pattern was detected
       Location string
       
       // Confidence is a score from 0.0 to 1.0 indicating detection confidence
       Confidence float64
   }
   ```

3. Update package documentation:
   ```go
   // Package services provides deep schema validation for OpenAPI contracts.
   //
   // Security validation uses AST-based analysis to verify that security
   // middleware is actually implemented in the code, not just declared in
   // endpoint metadata. This provides more accurate validation than
   // metadata-only checks.
   ```

**Deliverables:**
- Complete godoc documentation
- Type documentation

**Compliance Check:**
- ✅ Documentation standards: All exported functions documented
- ✅ Code comments: Inline comments for complex logic

#### Step 7.2: Update Implementation Documentation
**Objective:** Update STUB_FUNCTIONALITY_ANALYSIS.md

**Tasks:**
1. Update section 7.1 in `STUB_FUNCTIONALITY_ANALYSIS.md`:
   ```markdown
   ### 7. Schema Validator Stubs
   
   **Location:** `hub/api/services/schema_validator.go`
   
   #### 7.1 Security Middleware Validation
   **Function:** `validateSecurity` (line 315)
   - **Status:** ✅ **IMPLEMENTED** - Now uses AST-based security middleware verification
   - **Implementation:**
     - Uses `ast.AnalyzeAST` for comprehensive security pattern detection
     - Uses `ast.ExtractFunctions` for middleware function detection
     - Detects JWT, API key, OAuth, RBAC, and other security patterns
     - Falls back to metadata-based validation if AST analysis fails
   - **Files Added:**
     - `hub/api/services/schema_validator_security_patterns.go` - Pattern detection logic
   ```

**Deliverables:**
- Updated stub analysis document

**Compliance Check:**
- ✅ Documentation: Accurate status reporting

#### Step 7.3: Code Review Checklist
**Objective:** Ensure compliance with CODING_STANDARDS.md

**Review Checklist:**
- [ ] **File Size:** All files < 400 lines (business services)
- [ ] **Function Size:** All functions < 50 lines, complexity < 10
- [ ] **Error Handling:** All errors wrapped with context (`fmt.Errorf("...: %w", err)`)
- [ ] **Context Usage:** All functions check `ctx.Err()` in loops
- [ ] **Logging:** Uses `LogWarn`/`LogError` with context
- [ ] **Naming:** Clear, descriptive names (no abbreviations)
- [ ] **Documentation:** All exported functions have godoc comments
- [ ] **Testing:** Test coverage > 90%
- [ ] **Performance:** Meets response time requirements
- [ ] **Dependencies:** Proper dependency injection
- [ ] **Layer Separation:** No HTTP concerns in service layer
- [ ] **Security:** No hardcoded secrets, input validation

**Deliverables:**
- Code review checklist completed
- All compliance items verified

---

### Phase 8: File Size Management (If Needed)

#### Step 8.1: Assess File Size
**Objective:** Check if file splitting is needed

**Tasks:**
1. Count lines in `schema_validator.go` after implementation
2. If > 400 lines, plan file split:
   - Keep core validation in `schema_validator.go`
   - Move security pattern detection to `schema_validator_security_patterns.go`
   - Move helper functions to `schema_validator_helpers.go` (if needed)

**Deliverables:**
- File size assessment
- File splitting plan (if needed)

**Compliance Check:**
- ✅ File size: All files < 400 lines (CODING_STANDARDS.md)

---

## Implementation Summary

### Files to Create/Modify

1. **Modify:**
   - `hub/api/services/schema_validator.go` - Enhance `validateSecurity` function

2. **Create:**
   - `hub/api/services/schema_validator_security_patterns.go` - Pattern detection logic

3. **Modify:**
   - `hub/api/services/schema_validator_test.go` - Add integration tests

4. **Create:**
   - `hub/api/services/schema_validator_security_patterns_test.go` - Unit tests for pattern detection

5. **Update:**
   - `STUB_FUNCTIONALITY_ANALYSIS.md` - Mark section 7.1 as implemented

### Key Implementation Points

1. **AST Integration:**
   - Use `ast.AnalyzeAST(code, language, ["security", "middleware", "authentication", "authorization"])`
   - Use `ast.ExtractFunctions(code, language, "")` for function extraction
   - Map AST findings to security patterns

2. **Error Handling:**
   - Wrap all errors with context
   - Provide fallback to metadata-based validation
   - Log errors with appropriate levels

3. **Context Usage:**
   - Check `ctx.Err()` in all loops
   - Support context cancellation in file operations
   - Pass context to all function calls

4. **Performance:**
   - Cache AST analysis results if possible
   - Use efficient pattern matching
   - Optimize file reading operations

5. **Testing:**
   - Unit tests for pattern detection (> 90% coverage)
   - Integration tests for validation flow
   - Performance benchmarks

### Compliance Verification

All implementation will comply with `docs/external/CODING_STANDARDS.md`:

- ✅ **Architectural Standards:** Service layer, no HTTP concerns
- ✅ **File Size Limits:** Business services < 400 lines
- ✅ **Function Design:** Single responsibility, < 50 lines, complexity < 10
- ✅ **Error Handling:** Error wrapping, structured errors
- ✅ **Context Usage:** Proper context propagation and cancellation
- ✅ **Naming Conventions:** Clear, descriptive names
- ✅ **Testing Standards:** > 90% coverage, proper test structure
- ✅ **Documentation:** Complete godoc comments
- ✅ **Performance:** Meets response time requirements

---

## Timeline Estimate

- **Phase 1:** 1 day (Preparation)
- **Phase 2:** 1 day (Design)
- **Phase 3:** 2 days (Core Implementation)
- **Phase 4:** 2 days (Pattern Detection)
- **Phase 5:** 1 day (Error Handling)
- **Phase 6:** 1 day (Testing)
- **Phase 7:** 1 day (Documentation)
- **Phase 8:** 0.5 days (File Management, if needed)

**Total:** ~9.5 days

---

## Success Criteria

1. ✅ `validateSecurity` uses AST analysis for security middleware verification
2. ✅ Detects JWT, API key, OAuth, RBAC, and other security patterns
3. ✅ Falls back gracefully to metadata-based validation when AST fails
4. ✅ All code complies with CODING_STANDARDS.md
5. ✅ Test coverage > 90%
6. ✅ Performance meets requirements (< 500ms for complex validation)
7. ✅ Complete documentation (godoc comments)
8. ✅ No breaking changes to existing API

---

## Notes

- The implementation maintains backward compatibility by keeping the existing function signature
- Fallback to metadata-based validation ensures the feature works even when AST analysis fails
- File splitting may be necessary if `schema_validator.go` exceeds 400 lines
- Consider caching AST analysis results for performance optimization in future iterations
