# Consolidated Implementation Plan

## Executive Summary

This plan consolidates three critical improvements:
1. **Schema Validator Improvements** - Partial parsing support & enhanced generic detection
2. **Language Registry Pattern** - Dynamic language integration
3. **New Language Support** - Framework for adding languages easily

**Total Effort:** 4-6 weeks
**Priority:** High
**Impact:** 40-50% reduction in fallbacks, 3-5 files for new languages (vs 15+)

---

## Phase Overview

| Phase | Focus | Duration | Priority | Impact |
|-------|-------|----------|----------|--------|
| **Phase 1** | Schema Validator Improvements | 1 week | 🔴 Critical | High |
| **Phase 2** | Language Registry Foundation | 1 week | 🔴 Critical | High |
| **Phase 3** | Refactor Existing Languages | 1 week | 🟡 High | Medium |
| **Phase 4** | Enhanced Generic Detection | 3 days | 🟡 High | Medium |
| **Phase 5** | Testing & Documentation | 1 week | 🟡 High | High |
| **Phase 6** | Future: Plugin System (Optional) | 2 weeks | 🟢 Low | High |

---

## Phase 1: Schema Validator Improvements

**Goal:** Reduce fallback frequency by 30-40% through partial parsing support

**Duration:** 1 week (5 days)
**Priority:** 🔴 **CRITICAL** - Immediate impact

### 1.1 Partial Parsing Support

**Files to Modify:**
- `hub/api/ast/analysis.go`

**Implementation:**

```go
// Current (line 70-73)
tree, err := parser.ParseCtx(ctx, nil, []byte(code))
if err != nil {
    return nil, AnalysisStats{}, fmt.Errorf("parse error: %w", err)
}

// Improved: Partial parsing support
tree, err := parser.ParseCtx(ctx, nil, []byte(code))
if err != nil {
    // Tree-sitter may create partial AST even with syntax errors
    // Check if we have a usable tree before giving up
    if tree != nil {
        rootNode := tree.RootNode()
        if rootNode != nil && rootNode.ChildCount() > 0 {
            // Partial parsing succeeded - log warning but continue
            pkg.LogWarn(ctx, "Partial AST parsing (syntax errors present): %v", err)
            // Continue with partial tree - don't return error
        } else {
            // No usable tree - fall back
            return nil, AnalysisStats{}, fmt.Errorf("parse error: %w", err)
        }
    } else {
        // No tree at all - fall back
        return nil, AnalysisStats{}, fmt.Errorf("parse error: %w", err)
    }
}
```

**Tasks:**
- [ ] Modify `analyzeASTInternal()` in `analysis.go`
- [ ] Add partial tree validation logic
- [ ] Add warning logging for partial parsing
- [ ] Update error handling to continue with partial trees
- [ ] Test with syntax error scenarios

**Test Cases:**
```go
func TestPartialParsing_SyntaxError(t *testing.T) {
    code := `
    package main
    func AuthMiddleware(next http.Handler) http.Handler {
        auth := r.Header.Get("Authorization")
        if strings.HasPrefix(auth, "Bearer ") {
            // Missing closing brace
    `
    // Should use partial AST, not fallback
}

func TestPartialParsing_ValidCode(t *testing.T) {
    // Should work normally
}

func TestPartialParsing_NoUsableTree(t *testing.T) {
    // Should fallback
}
```

**Success Criteria:**
- ✅ Partial parsing works for syntax errors
- ✅ Fallback only when no usable tree
- ✅ 30-40% reduction in fallback frequency
- ✅ All existing tests pass

---

### 1.2 Enhanced Generic Detection

**Files to Modify:**
- `hub/api/ast/detection_security_middleware.go`

**Implementation:**

```go
// Current: Minimal generic detection
func detectSecurityMiddlewareGeneric(root *sitter.Node, code string, language string) []ASTFinding {
    findings := []ASTFinding{}
    codeLower := strings.ToLower(code)
    
    // Only checks Bearer token
    if strings.Contains(codeLower, "bearer") && strings.Contains(codeLower, "authorization") {
        // ... minimal detection
    }
    return findings
}

// Improved: Comprehensive generic detection
func detectSecurityMiddlewareGeneric(root *sitter.Node, code string, language string) []ASTFinding {
    findings := []ASTFinding{}
    codeLower := strings.ToLower(code)
    
    // Use same comprehensive pattern detection as code-based fallback
    // JWT/Bearer detection
    if containsJWTBearerPattern("", codeLower) {
        finding := createGenericMiddlewareFinding(code, "jwt_middleware", "BearerAuth", codeLower)
        findings = append(findings, finding)
    }
    
    // API key detection
    if containsAPIKeyPattern("", codeLower) {
        finding := createGenericMiddlewareFinding(code, "apikey_middleware", "ApiKeyAuth", codeLower)
        findings = append(findings, finding)
    }
    
    // OAuth detection
    if containsOAuthPattern("", codeLower) {
        finding := createGenericMiddlewareFinding(code, "oauth_middleware", "OAuth2", codeLower)
        findings = append(findings, finding)
    }
    
    // RBAC detection
    if containsRBACPattern("", codeLower) {
        finding := createGenericMiddlewareFinding(code, "rbac_middleware", "RBAC", codeLower)
        findings = append(findings, finding)
    }
    
    // Rate limiting detection
    if containsRateLimitPattern("", codeLower) {
        finding := createGenericMiddlewareFinding(code, "ratelimit_middleware", "RateLimit", codeLower)
        findings = append(findings, finding)
    }
    
    // CORS detection
    if containsCORSPattern("", codeLower) {
        finding := createGenericMiddlewareFinding(code, "cors_middleware", "CORS", codeLower)
        findings = append(findings, finding)
    }
    
    return findings
}

// Helper function
func createGenericMiddlewareFinding(code, findingType, scheme, codeLower string) ASTFinding {
    // Find line number for the pattern
    lines := strings.Split(code, "\n")
    lineNum := 1
    for i, line := range lines {
        lineLower := strings.ToLower(line)
        if strings.Contains(lineLower, scheme) || 
           strings.Contains(lineLower, "bearer") ||
           strings.Contains(lineLower, "apikey") {
            lineNum = i + 1
            break
        }
    }
    
    return ASTFinding{
        Type:       findingType,
        Severity:   "info",
        Line:       lineNum,
        Column:     1,
        Message:    fmt.Sprintf("Security middleware detected: %s", scheme),
        Code:       extractCodeSnippet(code, scheme, ""),
        Suggestion: fmt.Sprintf("Middleware implements %s security scheme", scheme),
        Confidence: 0.75, // Lower confidence for generic detection
        Reasoning:  fmt.Sprintf("Pattern detected via generic analysis for unsupported language"),
    }
}
```

**Tasks:**
- [ ] Enhance `detectSecurityMiddlewareGeneric()`
- [ ] Add all pattern detection functions (reuse from code-based detection)
- [ ] Implement `createGenericMiddlewareFinding()` helper
- [ ] Test with unsupported languages (Java, Rust, etc.)

**Test Cases:**
```go
func TestGenericDetection_Java(t *testing.T) {
    code := `
    public class AuthMiddleware {
        public void authenticate(HttpServletRequest request) {
            String auth = request.getHeader("Authorization");
            if (auth != null && auth.startsWith("Bearer ")) {
                // JWT validation
            }
        }
    }
    `
    // Should detect JWT pattern
}

func TestGenericDetection_AllPatterns(t *testing.T) {
    // Test all security patterns
}
```

**Success Criteria:**
- ✅ All security patterns detected for unsupported languages
- ✅ 80-85% accuracy (vs 60% current)
- ✅ Consistent with code-based fallback patterns

---

## Phase 2: Language Registry Foundation

**Goal:** Create registry pattern infrastructure for dynamic language support

**Duration:** 1 week (5 days)
**Priority:** 🔴 **CRITICAL** - Foundation for future

### 2.1 Define Language Interfaces

**New File:** `hub/api/ast/language_interfaces.go`

**Implementation:**

```go
package ast

import (
    sitter "github.com/smacker/go-tree-sitter"
)

// LanguageDetector defines language-specific detection capabilities
type LanguageDetector interface {
    // Security middleware detection
    DetectSecurityMiddleware(root *sitter.Node, code string) []ASTFinding
    
    // Other detections
    DetectUnused(root *sitter.Node, code string) []ASTFinding
    DetectDuplicates(root *sitter.Node, code string) []ASTFinding
    DetectSQLInjection(root *sitter.Node, code string) []SecurityVulnerability
    DetectXSS(root *sitter.Node, code string) []SecurityVulnerability
    DetectCommandInjection(root *sitter.Node, code string) []SecurityVulnerability
    DetectCrypto(root *sitter.Node, code string) []SecurityVulnerability
    DetectUnreachable(root *sitter.Node, code string) []ASTFinding
    DetectAsync(root *sitter.Node, code string) []ASTFinding
}

// LanguageExtractor defines language-specific extraction capabilities
type LanguageExtractor interface {
    ExtractFunctions(code, keyword string) ([]FunctionInfo, error)
    ExtractImports(code string) ([]ImportInfo, error)
    ExtractSymbols(root *sitter.Node, code string) (map[string]*Symbol, error)
}

// LanguageNodeTypes defines language-specific AST node types
type LanguageNodeTypes struct {
    FunctionDeclaration []string
    MethodDeclaration   []string
    VariableDeclaration []string
    ClassDeclaration    []string
    ImportStatement     []string
}

// LanguageSupport provides complete language support
type LanguageSupport interface {
    GetLanguage() string
    GetDetector() LanguageDetector
    GetExtractor() LanguageExtractor
    GetNodeTypes() LanguageNodeTypes
}
```

**Tasks:**
- [ ] Create `language_interfaces.go`
- [ ] Define all interfaces
- [ ] Document interface contracts
- [ ] Review with team

---

### 2.2 Create Language Registry

**New File:** `hub/api/ast/language_registry.go`

**Implementation:**

```go
package ast

import (
    "fmt"
    "sync"
)

var (
    languageRegistry = make(map[string]LanguageSupport)
    registryMutex    sync.RWMutex
)

// RegisterLanguageSupport registers a language implementation
func RegisterLanguageSupport(support LanguageSupport) error {
    if support == nil {
        return fmt.Errorf("language support cannot be nil")
    }
    
    lang := support.GetLanguage()
    if lang == "" {
        return fmt.Errorf("language name cannot be empty")
    }
    
    registryMutex.Lock()
    defer registryMutex.Unlock()
    
    if _, exists := languageRegistry[lang]; exists {
        return fmt.Errorf("language %s already registered", lang)
    }
    
    languageRegistry[lang] = support
    return nil
}

// GetLanguageSupport retrieves language support by name
func GetLanguageSupport(language string) LanguageSupport {
    registryMutex.RLock()
    defer registryMutex.RUnlock()
    return languageRegistry[language]
}

// GetLanguageDetector retrieves detector for language
func GetLanguageDetector(language string) LanguageDetector {
    support := GetLanguageSupport(language)
    if support == nil {
        return nil
    }
    return support.GetDetector()
}

// GetLanguageExtractor retrieves extractor for language
func GetLanguageExtractor(language string) LanguageExtractor {
    support := GetLanguageSupport(language)
    if support == nil {
        return nil
    }
    return support.GetExtractor()
}

// GetSupportedLanguages returns list of registered languages
func GetSupportedLanguages() []string {
    registryMutex.RLock()
    defer registryMutex.RUnlock()
    
    languages := make([]string, 0, len(languageRegistry))
    for lang := range languageRegistry {
        languages = append(languages, lang)
    }
    return languages
}
```

**Tasks:**
- [ ] Create `language_registry.go`
- [ ] Implement registry functions
- [ ] Add thread-safety (mutex)
- [ ] Add error handling
- [ ] Write unit tests

**Test Cases:**
```go
func TestLanguageRegistry_Register(t *testing.T) {
    support := &GoLanguageSupport{}
    err := RegisterLanguageSupport(support)
    assert.NoError(t, err)
}

func TestLanguageRegistry_Duplicate(t *testing.T) {
    support := &GoLanguageSupport{}
    RegisterLanguageSupport(support)
    err := RegisterLanguageSupport(support)
    assert.Error(t, err)
}

func TestLanguageRegistry_GetDetector(t *testing.T) {
    detector := GetLanguageDetector("go")
    assert.NotNil(t, detector)
}
```

**Success Criteria:**
- ✅ Registry works correctly
- ✅ Thread-safe
- ✅ Error handling complete
- ✅ Tests pass

---

### 2.3 Create Base Language Support

**New File:** `hub/api/ast/language_base.go`

**Implementation:**

```go
package ast

// BaseLanguageSupport provides default implementations
type BaseLanguageSupport struct {
    Language string
    Detector LanguageDetector
    Extractor LanguageExtractor
    NodeTypes LanguageNodeTypes
}

func (b *BaseLanguageSupport) GetLanguage() string {
    return b.Language
}

func (b *BaseLanguageSupport) GetDetector() LanguageDetector {
    return b.Detector
}

func (b *BaseLanguageSupport) GetExtractor() LanguageExtractor {
    return b.Extractor
}

func (b *BaseLanguageSupport) GetNodeTypes() LanguageNodeTypes {
    return b.NodeTypes
}
```

**Tasks:**
- [ ] Create base implementation
- [ ] Document usage
- [ ] Test base functionality

---

## Phase 3: Refactor Existing Languages

**Goal:** Migrate Go, JavaScript, TypeScript, Python to registry pattern

**Duration:** 1 week (5 days)
**Priority:** 🟡 **HIGH** - Enables dynamic support

### 3.1 Create Go Language Support

**New File:** `hub/api/ast/languages/go_support.go`

**Implementation:**

```go
package ast

import (
    sitter "github.com/smacker/go-tree-sitter"
)

type GoLanguageSupport struct {
    *BaseLanguageSupport
}

func NewGoLanguageSupport() *GoLanguageSupport {
    return &GoLanguageSupport{
        BaseLanguageSupport: &BaseLanguageSupport{
            Language: "go",
            Detector: &GoDetector{},
            Extractor: &GoExtractor{},
            NodeTypes: LanguageNodeTypes{
                FunctionDeclaration: []string{"function_declaration"},
                MethodDeclaration:   []string{"method_declaration"},
                VariableDeclaration: []string{"var_declaration", "short_var_declaration"},
            },
        },
    }
}

type GoDetector struct{}

func (d *GoDetector) DetectSecurityMiddleware(root *sitter.Node, code string) []ASTFinding {
    // Move existing detectSecurityMiddlewareGo() logic here
    return detectSecurityMiddlewareGo(root, code)
}

func (d *GoDetector) DetectUnused(root *sitter.Node, code string) []ASTFinding {
    return detectUnusedVariablesGo(root, code)
}

// ... implement other detection methods

type GoExtractor struct{}

func (e *GoExtractor) ExtractFunctions(code, keyword string) ([]FunctionInfo, error) {
    // Move existing extraction logic here
    return extractFunctionsGo(code, keyword)
}

// ... implement other extraction methods
```

**Tasks:**
- [ ] Create `languages/go_support.go`
- [ ] Move Go-specific detection logic
- [ ] Move Go-specific extraction logic
- [ ] Register in init()
- [ ] Test thoroughly

---

### 3.2 Create JavaScript/TypeScript Support

**New File:** `hub/api/ast/languages/javascript_support.go`

**Similar structure to Go support**

**Tasks:**
- [ ] Create JavaScript support
- [ ] Create TypeScript support (or combine)
- [ ] Move JS/TS-specific logic
- [ ] Register
- [ ] Test

---

### 3.3 Create Python Support

**New File:** `hub/api/ast/languages/python_support.go`

**Similar structure to Go support**

**Tasks:**
- [ ] Create Python support
- [ ] Move Python-specific logic
- [ ] Register
- [ ] Test

---

### 3.4 Refactor Detection Functions

**Files to Modify:** 10+ detection files

**Pattern:**

```go
// Before
func detectSecurityMiddleware(root *sitter.Node, code string, language string) []ASTFinding {
    switch language {
    case "go":
        return detectSecurityMiddlewareGo(root, code)
    case "javascript", "typescript":
        return detectSecurityMiddlewareJS(root, code)
    case "python":
        return detectSecurityMiddlewarePython(root, code)
    default:
        return detectSecurityMiddlewareGeneric(root, code, language)
    }
}

// After
func detectSecurityMiddleware(root *sitter.Node, code string, language string) []ASTFinding {
    detector := GetLanguageDetector(language)
    if detector == nil {
        return detectSecurityMiddlewareGeneric(root, code, language)
    }
    return detector.DetectSecurityMiddleware(root, code)
}
```

**Files to Refactor:**
- [ ] `detection_security_middleware.go`
- [ ] `detection_unused.go`
- [ ] `detection_duplicates.go`
- [ ] `detection_sql_injection.go`
- [ ] `detection_xss.go`
- [ ] `detection_command_injection.go`
- [ ] `detection_crypto.go`
- [ ] `detection_unreachable.go`
- [ ] `detection_async.go`
- [ ] `extraction.go`
- [ ] `extraction_helpers.go`

**Tasks:**
- [ ] Refactor each detection function
- [ ] Remove language switch statements
- [ ] Use registry instead
- [ ] Test each refactored function
- [ ] Ensure backward compatibility

**Success Criteria:**
- ✅ All detection functions use registry
- ✅ No language switch statements remain
- ✅ All tests pass
- ✅ Performance maintained or improved

---

## Phase 4: Enhanced Generic Detection

**Goal:** Improve detection for unsupported languages

**Duration:** 3 days
**Priority:** 🟡 **HIGH** - Better user experience

### 4.1 Enhance Generic Detection

**Already covered in Phase 1.2, but ensure integration with registry:**

```go
// In detectSecurityMiddlewareGeneric()
// Use comprehensive pattern detection
// Reuse pattern functions from code-based detection
```

**Tasks:**
- [ ] Ensure generic detection uses all patterns
- [ ] Integrate with registry (fallback when no detector)
- [ ] Test with multiple unsupported languages
- [ ] Document generic detection capabilities

---

## Phase 5: Testing & Documentation

**Goal:** Comprehensive testing and documentation

**Duration:** 1 week (5 days)
**Priority:** 🟡 **HIGH** - Quality assurance

### 5.1 Unit Tests

**Test Files:**
- [ ] `language_registry_test.go` - Registry tests
- [ ] `go_support_test.go` - Go language tests
- [ ] `javascript_support_test.go` - JS/TS tests
- [ ] `python_support_test.go` - Python tests
- [ ] `partial_parsing_test.go` - Partial parsing tests
- [ ] `generic_detection_test.go` - Generic detection tests

**Coverage Goals:**
- ✅ Registry: 95%+
- ✅ Language support: 90%+
- ✅ Partial parsing: 90%+
- ✅ Generic detection: 85%+

---

### 5.2 Integration Tests

**Test Scenarios:**
- [ ] Real-world code samples
- [ ] Multiple languages
- [ ] Syntax error scenarios
- [ ] Unsupported languages
- [ ] Performance benchmarks

---

### 5.3 Documentation

**Documents to Create/Update:**
- [ ] `docs/LANGUAGE_REGISTRY.md` - How to use registry
- [ ] `docs/ADDING_NEW_LANGUAGE.md` - Step-by-step guide
- [ ] `docs/PARTIAL_PARSING.md` - Partial parsing behavior
- [ ] `docs/GENERIC_DETECTION.md` - Generic detection capabilities
- [ ] Update main README

---

## Phase 6: Future Enhancement - Plugin System (Optional)

**Goal:** Fully dynamic plugin-based system

**Duration:** 2 weeks
**Priority:** 🟢 **LOW** - Future enhancement

### 6.1 Plugin Interface

**New File:** `hub/api/ast/plugin_interface.go`

```go
type LanguagePlugin interface {
    GetLanguage() string
    GetParser() (*sitter.Parser, error)
    GetSupport() LanguageSupport
    Init() error
}
```

### 6.2 Auto-Registration

**New File:** `hub/api/ast/plugins/plugins.go`

```go
package plugins

import _ "sentinel-hub-api/ast/plugins/go"
import _ "sentinel-hub-api/ast/plugins/java"
// ... auto-register on import
```

**Tasks:**
- [ ] Design plugin interface
- [ ] Implement auto-registration
- [ ] Convert existing languages to plugins
- [ ] Test plugin system
- [ ] Document plugin development

---

## Implementation Timeline

### Week 1: Schema Validator Improvements
- **Day 1-2:** Partial parsing support
- **Day 3-4:** Enhanced generic detection
- **Day 5:** Testing & bug fixes

### Week 2: Language Registry Foundation
- **Day 1:** Define interfaces
- **Day 2:** Create registry
- **Day 3:** Base language support
- **Day 4-5:** Testing & refinement

### Week 3: Refactor Existing Languages
- **Day 1:** Go language support
- **Day 2:** JavaScript/TypeScript support
- **Day 3:** Python support
- **Day 4-5:** Refactor detection functions

### Week 4: Enhanced Generic & Testing
- **Day 1-2:** Enhanced generic detection integration
- **Day 3-5:** Comprehensive testing

### Week 5: Documentation & Polish
- **Day 1-3:** Documentation
- **Day 4-5:** Final testing & bug fixes

---

## Risk Assessment & Mitigation

### Risk 1: Breaking Changes
**Probability:** Medium
**Impact:** High
**Mitigation:**
- Maintain backward compatibility
- Gradual migration
- Comprehensive testing
- Feature flags if needed

### Risk 2: Performance Regression
**Probability:** Low
**Impact:** Medium
**Mitigation:**
- Benchmark before/after
- Profile critical paths
- Optimize registry lookups
- Cache language support

### Risk 3: Incomplete Migration
**Probability:** Medium
**Impact:** Medium
**Mitigation:**
- Checklist for each file
- Code review
- Automated tests
- Integration tests

---

## Success Metrics

### Quantitative
- ✅ **Fallback Reduction:** 30-40% (from 25-30% to 10-15%)
- ✅ **Accuracy Improvement:** 10-15% (in fallback scenarios)
- ✅ **Files for New Language:** 3-5 (from 15+)
- ✅ **Test Coverage:** 90%+ for new code
- ✅ **Performance:** No regression (<5% overhead)

### Qualitative
- ✅ **Maintainability:** Easier to add languages
- ✅ **Code Quality:** Centralized language logic
- ✅ **Documentation:** Complete guides
- ✅ **Developer Experience:** Clear patterns

---

## Rollback Plan

If issues arise:

1. **Phase 1 (Partial Parsing):** Can disable with feature flag
2. **Phase 2-3 (Registry):** Can keep old code path, migrate gradually
3. **Phase 4 (Generic):** Can revert to minimal generic detection

**Rollback Strategy:**
- Keep old code paths initially
- Feature flags for new code
- Gradual rollout
- Monitor metrics

---

## Dependencies

### External
- Tree-sitter library (already in use)
- No new dependencies required

### Internal
- Existing AST package
- Schema validator service
- Test infrastructure

---

## Team Requirements

### Skills Needed
- Go programming (advanced)
- AST/tree-sitter knowledge
- Testing expertise
- Documentation skills

### Roles
- **Lead Developer:** Architecture & implementation
- **Developer:** Language support implementation
- **QA Engineer:** Testing & validation
- **Technical Writer:** Documentation

---

## Next Steps

1. **Review Plan:** Team review and approval
2. **Create Issues:** Break down into GitHub issues
3. **Set Up Branch:** Create feature branch
4. **Start Phase 1:** Begin with partial parsing
5. **Daily Standups:** Track progress

---

## Appendix: File Structure After Implementation

```
hub/api/ast/
├── analysis.go                    # Generic (no language switches)
├── parsers.go                     # Generic parser access
├── language_interfaces.go         # NEW: Interfaces
├── language_registry.go           # NEW: Registry
├── language_base.go               # NEW: Base implementation
├── languages/
│   ├── go_support.go              # NEW: Go implementation
│   ├── javascript_support.go      # NEW: JS/TS implementation
│   └── python_support.go          # NEW: Python implementation
├── detection_security_middleware.go  # Generic (uses registry)
├── detection_unused.go            # Generic (uses registry)
├── detection_duplicates.go        # Generic (uses registry)
├── ... (other detections)         # All generic
├── extraction.go                  # Generic (uses registry)
└── ... (other files)
```

---

## Conclusion

This consolidated plan provides:
- ✅ **Immediate Impact:** Partial parsing reduces fallbacks by 30-40%
- ✅ **Long-term Value:** Registry pattern enables easy language addition
- ✅ **Better UX:** Enhanced generic detection for unsupported languages
- ✅ **Maintainability:** Centralized, testable architecture

**Total Effort:** 4-6 weeks
**Expected Outcome:** Production-ready, maintainable, extensible language support system
