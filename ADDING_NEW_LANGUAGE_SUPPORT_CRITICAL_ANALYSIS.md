# Critical Analysis: Adding New Language Support

## Implementation Status (January 2026)

**Registry-based language support is now complete for Go, JavaScript, TypeScript, and Python.**

- **Go:** Full detector, extractor, support; 100% test coverage (`go_detector_test.go`).
- **JavaScript / TypeScript:** `JsDetector`, `JsExtractor`, `JsLanguageSupport` / `TsLanguageSupport`; registered in `language_init.go`; all detection entry points use registry-first pattern.
- **Python:** `PythonDetector`, `PythonExtractor`, `PythonLanguageSupport`; registered in `language_init.go`.
- **All detection functions** (`detectUnusedVariables`, `detectDuplicateFunctions`, `detectUnreachableCode`, `detectMissingAwait`, `detectSQLInjection`, `detectXSS`, `detectCommandInjection`, `detectInsecureCrypto`, `detectSecurityMiddleware`) use **registry-first** with fallback to switch for backward compatibility.

See **CRITICAL_ANALYSIS_IMPLEMENTATION.md** for production readiness and verification details.

---

## Executive Summary

Adding support for a new language requires changes across **15+ files** and **20+ functions**. The architecture is **highly coupled** with language-specific logic scattered throughout the codebase. This analysis provides a comprehensive guide with critical considerations.

---

## Architecture Overview

### Current Language Support
- **Go** (golang)
- **JavaScript** (javascript, js, jsx)
- **TypeScript** (typescript, ts, tsx)
- **Python** (python, py)

### Language Support Components

The codebase has **three layers** of language support:

1. **Parser Layer** - Tree-sitter parser initialization
2. **Detection Layer** - Language-specific pattern detection
3. **Extraction Layer** - Language-specific code extraction

---

## Critical Analysis: Required Changes

### 🔴 **CRITICAL: Parser Registration** (Must Do)

**Files:** `hub/api/ast/parsers.go`

**Changes Required:**

#### 1. Add Tree-sitter Language Binding

**Dependency:** Add to `go.mod`
```go
require (
    github.com/smacker/go-tree-sitter/java v0.0.0-... // Example for Java
    // OR
    github.com/smacker/go-tree-sitter/rust v0.0.0-... // Example for Rust
)
```

**Challenge:** 
- ⚠️ Not all languages have official Tree-sitter bindings
- ⚠️ May need to use community bindings or build custom
- ⚠️ Version compatibility issues

#### 2. Import Language Package

**File:** `parsers.go`
```go
import (
    // ... existing imports
    "github.com/smacker/go-tree-sitter/java" // Example
)
```

#### 3. Initialize Parser

**Function:** `initParsers()`
```go
func initParsers() {
    // ... existing parsers
    
    // NEW LANGUAGE: Java example
    javaParser := sitter.NewParser()
    javaParser.SetLanguage(java.GetLanguage())
    parsers["java"] = javaParser
    parsers["jav"] = javaParser // Optional: alias
}
```

**Critical Considerations:**
- ✅ Must add to `parsers` map
- ✅ Consider aliases (e.g., `jav` for `java`)
- ✅ Thread-safety: Uses `sync.Once` (safe)

#### 4. Normalize Language Name

**Function:** `normalizeLanguage()`
```go
func normalizeLanguage(lang string) string {
    lang = strings.ToLower(lang)
    switch lang {
    // ... existing cases
    case "java", "jav":
        return "java"
    default:
        return lang
    }
}
```

#### 5. Create Parser Function

**Function:** `createParserForLanguage()`
```go
func createParserForLanguage(language string) (*sitter.Parser, error) {
    lang := normalizeLanguage(language)
    var parser *sitter.Parser
    switch lang {
    // ... existing cases
    case "java", "jav":
        parser = sitter.NewParser()
        parser.SetLanguage(java.GetLanguage())
    default:
        return nil, fmt.Errorf("unsupported language: %s (supported: go, javascript, typescript, python, java)", language)
    }
    return parser, nil
}
```

#### 6. Update Error Messages

**Locations:**
- `GetParser()` - Line 69
- `createParserForLanguage()` - Line 110

**Change:**
```go
// Before
return nil, fmt.Errorf("unsupported language: %s (supported: go, javascript, typescript, python)", language)

// After
return nil, fmt.Errorf("unsupported language: %s (supported: go, javascript, typescript, python, java)", language)
```

**Impact:** 🔴 **CRITICAL** - Without this, parser won't work

---

### 🟡 **HIGH PRIORITY: Security Middleware Detection** (Must Do)

**File:** `hub/api/ast/detection_security_middleware.go`

**Changes Required:**

#### 1. Add Language Case

**Function:** `detectSecurityMiddleware()`
```go
func detectSecurityMiddleware(root *sitter.Node, code string, language string) []ASTFinding {
    findings := []ASTFinding{}
    
    switch language {
    case "go":
        findings = append(findings, detectSecurityMiddlewareGo(root, code)...)
    case "javascript", "typescript":
        findings = append(findings, detectSecurityMiddlewareJS(root, code)...)
    case "python":
        findings = append(findings, detectSecurityMiddlewarePython(root, code)...)
    case "java": // NEW
        findings = append(findings, detectSecurityMiddlewareJava(root, code)...)
    default:
        findings = append(findings, detectSecurityMiddlewareGeneric(root, code, language)...)
    }
    
    return findings
}
```

#### 2. Implement Language-Specific Detection

**New Function:** `detectSecurityMiddlewareJava()`
```go
// detectSecurityMiddlewareJava detects security middleware in Java code
func detectSecurityMiddlewareJava(root *sitter.Node, code string) []ASTFinding {
    findings := []ASTFinding{}
    codeLower := strings.ToLower(code)
    
    // Traverse AST to find method declarations
    TraverseAST(root, func(node *sitter.Node) bool {
        // Java-specific node types
        if node.Type() == "method_declaration" || node.Type() == "method_invocation" {
            methodName, methodCode := extractFunctionInfoJava(node, code)
            if methodName == "" {
                return true
            }
            
            methodNameLower := strings.ToLower(methodName)
            methodCodeLower := strings.ToLower(methodCode)
            
            // Check for JWT/Bearer patterns
            if containsJWTBearerPattern(methodNameLower, methodCodeLower) {
                finding := createMiddlewareFinding(node, code, "jwt_middleware", "BearerAuth", methodName)
                findings = append(findings, finding)
            }
            
            // Check for API key patterns
            if containsAPIKeyPattern(methodNameLower, methodCodeLower) {
                finding := createMiddlewareFinding(node, code, "apikey_middleware", "ApiKeyAuth", methodName)
                findings = append(findings, finding)
            }
            
            // ... other patterns
        }
        return true
    })
    
    // Also check for security patterns in code (not just methods)
    if strings.Contains(codeLower, "authorization") && strings.Contains(codeLower, "bearer") {
        finding := ASTFinding{
            Type:       "jwt_middleware",
            Severity:   "info",
            Line:       1,
            Column:     1,
            Message:    "JWT/Bearer token authentication detected",
            Code:       extractCodeSnippet(code, "bearer", "authorization"),
            Suggestion: "Security middleware pattern detected",
            Confidence: 0.85,
        }
        findings = append(findings, finding)
    }
    
    return findings
}
```

#### 3. Language-Specific Helper Functions

**New Function:** `extractFunctionInfoJava()`
```go
// extractFunctionInfoJava extracts method name and code from Java AST node
func extractFunctionInfoJava(node *sitter.Node, code string) (string, string) {
    methodName := ""
    methodCode := safeSlice(code, node.StartByte(), node.EndByte())
    
    // Java-specific AST traversal
    // Method name is typically in an "identifier" child node
    for i := 0; i < int(node.ChildCount()); i++ {
        child := node.Child(i)
        if child != nil {
            // Java method declaration structure:
            // modifiers type_identifier identifier (parameters) { body }
            if child.Type() == "identifier" {
                // Check if this is the method name (not a type)
                parent := child.Parent()
                if parent != nil && parent.Type() == "method_declaration" {
                    // This is likely the method name
                    methodName = safeSlice(code, child.StartByte(), child.EndByte())
                    break
                }
            }
        }
    }
    
    return methodName, methodCode
}
```

**Critical Considerations:**
- ⚠️ **AST Node Types Vary by Language** - Must understand Tree-sitter grammar
- ⚠️ **Function/Method Structure Differs** - Java methods vs Go functions vs Python functions
- ⚠️ **Naming Conventions** - Java camelCase vs Go PascalCase vs Python snake_case
- ⚠️ **Signature Patterns** - Different middleware patterns per language

**Resources Needed:**
- Tree-sitter grammar documentation for the language
- Understanding of language-specific AST node types
- Examples of security middleware in that language

---

### 🟡 **HIGH PRIORITY: Other Detection Modules** (Should Do)

**Files Affected:** 10+ detection files

Each detection module needs language-specific support:

1. **`detection_unused.go`** - Unused variable detection
2. **`detection_duplicates.go`** - Duplicate function detection
3. **`detection_sql_injection.go`** - SQL injection detection
4. **`detection_xss.go`** - XSS detection
5. **`detection_command_injection.go`** - Command injection detection
6. **`detection_crypto.go`** - Insecure crypto detection
7. **`detection_unreachable.go`** - Unreachable code detection
8. **`detection_async.go`** - Missing await detection
9. **`detection_secrets.go`** - Secrets detection
10. **`detection_syntax.go`** - Syntax error detection

**Pattern:**
```go
func detectXXX(root *sitter.Node, code string, language string) []ASTFinding {
    switch language {
    case "go":
        return detectXXXGo(root, code)
    case "javascript", "typescript":
        return detectXXXJS(root, code)
    case "python":
        return detectXXXPython(root, code)
    case "java": // NEW
        return detectXXXJava(root, code)
    }
    return []ASTFinding{}
}
```

**Impact:** 🟡 **HIGH** - Without this, new language won't have full detection capabilities

**Effort:** High - Each detection needs language-specific implementation

---

### 🟢 **MEDIUM PRIORITY: Extraction Modules** (Should Do)

**Files:** 
- `extraction.go` - Function extraction
- `extraction_helpers.go` - Helper functions

**Changes Required:**

#### 1. Function Extraction

**File:** `extraction.go`

**Function:** `ExtractFunctions()`
```go
func ExtractFunctions(code, language, keyword string) ([]FunctionInfo, error) {
    // ... existing code
    
    switch language {
    case "go":
        // ... existing
    case "javascript", "typescript":
        // ... existing
    case "python":
        // ... existing
    case "java": // NEW
        return extractFunctionsJava(code, keyword)
    default:
        return nil, fmt.Errorf("unsupported language for function extraction: %s", language)
    }
}
```

**New Function:** `extractFunctionsJava()`
```go
func extractFunctionsJava(code, keyword string) ([]FunctionInfo, error) {
    parser, err := GetParser("java")
    if err != nil {
        return nil, err
    }
    
    tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
    if err != nil {
        return nil, err
    }
    defer tree.Close()
    
    rootNode := tree.RootNode()
    functions := []FunctionInfo{}
    
    TraverseAST(rootNode, func(node *sitter.Node) bool {
        if node.Type() == "method_declaration" {
            // Extract method information
            methodName := extractMethodNameJava(node, code)
            if keyword == "" || strings.Contains(strings.ToLower(methodName), strings.ToLower(keyword)) {
                functions = append(functions, FunctionInfo{
                    Name:   methodName,
                    Line:   int(node.StartPoint().Row) + 1,
                    Column: int(node.StartPoint().Column) + 1,
                })
            }
        }
        return true
    })
    
    return functions, nil
}
```

**Impact:** 🟢 **MEDIUM** - Affects function-based middleware detection

---

### 🟢 **MEDIUM PRIORITY: Utility Modules** (Nice to Have)

**Files:**
- `utils.go` - Language utilities
- `search_patterns.go` - Search patterns

**Changes Required:**

#### 1. Language Detection

**File:** `utils.go`

**Function:** `DetectLanguageFromFile()`
```go
func DetectLanguageFromFile(filePath string) string {
    ext := filepath.Ext(filePath)
    switch ext {
    // ... existing cases
    case ".java":
        return "java"
    default:
        return "go" // default
    }
}
```

#### 2. Search Patterns

**File:** `search_patterns.go`

**Function:** `GetSearchPatterns()`
```go
func GetSearchPatterns(language string) []string {
    switch language {
    // ... existing cases
    case "java":
        return []string{
            "method_declaration",
            "class_declaration",
            "interface_declaration",
        }
    default:
        return []string{}
    }
}
```

**Impact:** 🟢 **MEDIUM** - Affects search and utility functions

---

## Step-by-Step Implementation Guide

### Phase 1: Parser Setup (Foundation)

**Step 1.1:** Add Tree-sitter binding dependency
```bash
cd hub/api
go get github.com/smacker/go-tree-sitter/java@latest
# OR for other languages
go get github.com/smacker/go-tree-sitter/rust@latest
go get github.com/smacker/go-tree-sitter/cpp@latest
```

**Step 1.2:** Update `parsers.go`
- Add import
- Add to `initParsers()`
- Add to `normalizeLanguage()`
- Add to `createParserForLanguage()`
- Update error messages

**Step 1.3:** Test parser initialization
```go
// Test
parser, err := GetParser("java")
if err != nil {
    t.Fatalf("Failed to get Java parser: %v", err)
}
```

**Verification:**
```bash
go test ./ast -run TestGetParser
```

---

### Phase 2: Security Middleware Detection (Core Feature)

**Step 2.1:** Study Tree-sitter grammar
- Review language grammar: https://tree-sitter.github.io/tree-sitter/
- Understand AST node types for the language
- Identify function/method declaration patterns

**Step 2.2:** Implement `detectSecurityMiddlewareJava()`
- Copy structure from `detectSecurityMiddlewareGo()`
- Adapt to language-specific AST nodes
- Test with real code examples

**Step 2.3:** Implement helper functions
- `extractFunctionInfoJava()` or language-specific equivalent
- Adapt pattern detection functions if needed

**Step 2.4:** Add to switch statement
- Update `detectSecurityMiddleware()` switch

**Step 2.5:** Test thoroughly
```go
func TestDetectSecurityMiddlewareJava(t *testing.T) {
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
    // Test detection
}
```

---

### Phase 3: Other Detection Modules (Optional but Recommended)

**For each detection module:**
1. Add case to switch statement
2. Implement language-specific function
3. Test with language examples

**Priority Order:**
1. `detection_unused.go` - High usage
2. `detection_duplicates.go` - High usage
3. `detection_sql_injection.go` - Security critical
4. `detection_xss.go` - Security critical
5. Others as needed

---

### Phase 4: Extraction Modules (Recommended)

**Step 4.1:** Update `ExtractFunctions()`
- Add case for new language
- Implement extraction function

**Step 4.2:** Update `extraction_helpers.go`
- Add language-specific helpers if needed

---

### Phase 5: Utilities (Nice to Have)

**Step 5.1:** Update `utils.go`
- Add file extension mapping
- Add language detection

**Step 5.2:** Update `search_patterns.go`
- Add language-specific patterns

---

## Critical Challenges & Considerations

### 🔴 **Challenge 1: Tree-sitter Grammar Understanding**

**Problem:** Each language has different AST node types

**Example:**
- Go: `function_declaration`, `method_declaration`
- Java: `method_declaration`, `constructor_declaration`
- Python: `function_definition`, `decorated_definition`
- JavaScript: `function_declaration`, `function`, `arrow_function`

**Solution:**
- Study Tree-sitter grammar documentation
- Use AST playground: https://tree-sitter.github.io/tree-sitter/playground
- Test with sample code to understand node structure

**Tool:**
```bash
# Install tree-sitter CLI
npm install -g tree-sitter

# Parse and inspect
tree-sitter parse code.java
```

---

### 🔴 **Challenge 2: Language-Specific Patterns**

**Problem:** Security patterns differ by language

**Examples:**

**JWT Detection:**
- Go: `r.Header.Get("Authorization")` + `strings.HasPrefix(auth, "Bearer ")`
- Java: `request.getHeader("Authorization")` + `auth.startsWith("Bearer ")`
- Python: `request.headers.get("Authorization")` + `auth.startswith("Bearer ")`
- JavaScript: `req.headers.authorization` + `auth.startsWith("Bearer ")`

**Solution:**
- Language-specific pattern detection
- Use AST to find method calls, not just string matching
- Understand language's HTTP framework conventions

---

### 🔴 **Challenge 3: Dependency Management**

**Problem:** Tree-sitter bindings may not exist or be outdated

**Solutions:**
1. **Official Bindings:** Use `github.com/smacker/go-tree-sitter/<lang>`
2. **Community Bindings:** May need to fork or use alternatives
3. **Custom Bindings:** Build from Tree-sitter grammar (complex)

**Check Availability:**
```bash
# Search for bindings
go list -m -versions github.com/smacker/go-tree-sitter/java
```

**Fallback:** If no binding exists, language falls back to generic detection

---

### 🟡 **Challenge 4: Testing Complexity**

**Problem:** Need comprehensive tests for each language

**Required Tests:**
1. Parser initialization
2. AST parsing
3. Security middleware detection
4. Function extraction
5. Edge cases (syntax errors, partial code, etc.)

**Test Structure:**
```go
func TestSecurityMiddlewareJava_RealWorld(t *testing.T) {
    testCases := []struct {
        name     string
        code     string
        expected []string // Expected patterns
    }{
        {
            name: "Spring Security JWT",
            code: `@Component
public class JwtAuthenticationFilter extends OncePerRequestFilter {
    @Override
    protected void doFilterInternal(HttpServletRequest request, ...) {
        String authHeader = request.getHeader("Authorization");
        if (authHeader != null && authHeader.startsWith("Bearer ")) {
            // JWT validation
        }
    }
}`,
            expected: []string{"BearerAuth"},
        },
        // ... more test cases
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test detection
        })
    }
}
```

---

### 🟡 **Challenge 5: Maintenance Burden**

**Problem:** Each new language multiplies maintenance

**Impact:**
- 15+ files need updates
- 20+ functions need language cases
- Tests multiply
- Documentation needs updates

**Mitigation:**
- Use code generation for boilerplate
- Create language abstraction layer
- Document patterns clearly
- Use generic detection as fallback

---

## Architecture Improvements (Future)

### Current Architecture Issues

1. **High Coupling:** Language logic scattered across many files
2. **Code Duplication:** Similar patterns repeated for each language
3. **Maintenance Burden:** Adding language requires many changes

### Proposed Improvements

#### Option 1: Language Registry Pattern

```go
// Language-specific implementations
type LanguageDetector interface {
    DetectSecurityMiddleware(root *sitter.Node, code string) []ASTFinding
    ExtractFunctions(code, keyword string) ([]FunctionInfo, error)
    GetNodeTypes() LanguageNodeTypes
}

type LanguageNodeTypes struct {
    FunctionDeclaration []string
    MethodDeclaration   []string
    VariableDeclaration []string
}

// Registry
var languageDetectors = map[string]LanguageDetector{
    "go":       &GoDetector{},
    "java":     &JavaDetector{},
    "python":   &PythonDetector{},
    // ...
}

func detectSecurityMiddleware(root *sitter.Node, code string, language string) []ASTFinding {
    detector, ok := languageDetectors[language]
    if !ok {
        return detectSecurityMiddlewareGeneric(root, code, language)
    }
    return detector.DetectSecurityMiddleware(root, code)
}
```

**Benefits:**
- ✅ Centralized language logic
- ✅ Easier to add new languages
- ✅ Better testability
- ✅ Reduced code duplication

**Effort:** High (refactoring required)

---

#### Option 2: Code Generation

**Generate language-specific code from templates:**

```go
// Template
{{range .Languages}}
func detectSecurityMiddleware{{.Name}}(root *sitter.Node, code string) []ASTFinding {
    // Generated code
}
{{end}}
```

**Benefits:**
- ✅ Consistent implementation
- ✅ Less manual work
- ✅ Easier maintenance

**Effort:** Medium (setup generation)

---

## Testing Strategy

### Unit Tests

**File:** `detection_security_middleware_test.go`

```go
func TestDetectSecurityMiddlewareJava_JWT(t *testing.T) {
    code := `
    public class AuthFilter implements Filter {
        public void doFilter(ServletRequest request, ...) {
            HttpServletRequest httpReq = (HttpServletRequest) request;
            String auth = httpReq.getHeader("Authorization");
            if (auth != null && auth.startsWith("Bearer ")) {
                // JWT validation
            }
        }
    }
    `
    
    parser, _ := GetParser("java")
    tree, _ := parser.ParseCtx(context.Background(), nil, []byte(code))
    rootNode := tree.RootNode()
    
    findings := detectSecurityMiddlewareJava(rootNode, code)
    
    // Assert JWT pattern detected
    found := false
    for _, f := range findings {
        if f.Type == "jwt_middleware" {
            found = true
            break
        }
    }
    assert.True(t, found, "JWT middleware should be detected")
}
```

### Integration Tests

**File:** `detection_security_middleware_integration_test.go`

```go
func TestSecurityMiddlewareJava_RealWorldFrameworks(t *testing.T) {
    frameworks := []struct {
        name string
        code string
    }{
        {
            name: "Spring Security",
            code: springSecurityExample,
        },
        {
            name: "JAX-RS",
            code: jaxrsExample,
        },
        {
            name: "Servlet Filter",
            code: servletFilterExample,
        },
    }
    
    for _, fw := range frameworks {
        t.Run(fw.name, func(t *testing.T) {
            // Test detection
        })
    }
}
```

### Performance Tests

```go
func BenchmarkDetectSecurityMiddlewareJava(b *testing.B) {
    code := loadLargeJavaFile() // 10k+ lines
    parser, _ := GetParser("java")
    tree, _ := parser.ParseCtx(context.Background(), nil, []byte(code))
    rootNode := tree.RootNode()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        detectSecurityMiddlewareJava(rootNode, code)
    }
}
```

---

## Checklist: Adding New Language Support

### Phase 1: Foundation
- [ ] Add Tree-sitter binding to `go.mod`
- [ ] Import language package in `parsers.go`
- [ ] Add to `initParsers()`
- [ ] Add to `normalizeLanguage()`
- [ ] Add to `createParserForLanguage()`
- [ ] Update error messages
- [ ] Test parser initialization

### Phase 2: Core Detection
- [ ] Study Tree-sitter grammar
- [ ] Implement `detectSecurityMiddleware<Lang>()`
- [ ] Implement language-specific helpers
- [ ] Add to `detectSecurityMiddleware()` switch
- [ ] Write unit tests
- [ ] Write integration tests

### Phase 3: Additional Detections (Optional)
- [ ] `detection_unused.go`
- [ ] `detection_duplicates.go`
- [ ] `detection_sql_injection.go`
- [ ] `detection_xss.go`
- [ ] Others as needed

### Phase 4: Extraction
- [ ] Update `ExtractFunctions()`
- [ ] Implement `extractFunctions<Lang>()`
- [ ] Update `extraction_helpers.go` if needed

### Phase 5: Utilities
- [ ] Update `utils.go`
- [ ] Update `search_patterns.go`

### Phase 6: Documentation
- [ ] Update README with new language
- [ ] Document language-specific patterns
- [ ] Add examples

---

## Example: Adding Java Support (Complete)

### Step 1: Dependencies

**`go.mod`:**
```go
require (
    github.com/smacker/go-tree-sitter/java v0.20.0
)
```

### Step 2: Parser Setup

**`parsers.go`:**
```go
import (
    "github.com/smacker/go-tree-sitter/java"
)

func initParsers() {
    // ... existing parsers
    
    javaParser := sitter.NewParser()
    javaParser.SetLanguage(java.GetLanguage())
    parsers["java"] = javaParser
}

func normalizeLanguage(lang string) string {
    lang = strings.ToLower(lang)
    switch lang {
    // ... existing cases
    case "java", "jav":
        return "java"
    default:
        return lang
    }
}
```

### Step 3: Security Middleware Detection

**`detection_security_middleware.go`:**
```go
func detectSecurityMiddleware(root *sitter.Node, code string, language string) []ASTFinding {
    switch language {
    // ... existing cases
    case "java":
        findings = append(findings, detectSecurityMiddlewareJava(root, code)...)
    default:
        findings = append(findings, detectSecurityMiddlewareGeneric(root, code, language)...)
    }
    return findings
}

func detectSecurityMiddlewareJava(root *sitter.Node, code string) []ASTFinding {
    findings := []ASTFinding{}
    codeLower := strings.ToLower(code)
    
    TraverseAST(root, func(node *sitter.Node) bool {
        // Java method declarations
        if node.Type() == "method_declaration" {
            methodName, methodCode := extractMethodInfoJava(node, code)
            methodNameLower := strings.ToLower(methodName)
            methodCodeLower := strings.ToLower(methodCode)
            
            // JWT detection
            if containsJWTBearerPattern(methodNameLower, methodCodeLower) {
                finding := createMiddlewareFinding(node, code, "jwt_middleware", "BearerAuth", methodName)
                findings = append(findings, finding)
            }
            
            // API key detection
            if containsAPIKeyPattern(methodNameLower, methodCodeLower) {
                finding := createMiddlewareFinding(node, code, "apikey_middleware", "ApiKeyAuth", methodName)
                findings = append(findings, finding)
            }
        }
        return true
    })
    
    return findings
}

func extractMethodInfoJava(node *sitter.Node, code string) (string, string) {
    methodName := ""
    methodCode := safeSlice(code, node.StartByte(), node.EndByte())
    
    // Java method structure: modifiers type identifier (parameters) { body }
    for i := 0; i < int(node.ChildCount()); i++ {
        child := node.Child(i)
        if child != nil && child.Type() == "identifier" {
            // Check if this is the method name
            parent := child.Parent()
            if parent != nil && parent.Type() == "method_declaration" {
                methodName = safeSlice(code, child.StartByte(), child.EndByte())
                break
            }
        }
    }
    
    return methodName, methodCode
}
```

---

## Summary

### Critical Requirements
1. ✅ **Parser Registration** - Foundation (must do)
2. ✅ **Security Middleware Detection** - Core feature (must do)
3. ⚠️ **Other Detections** - Full support (should do)
4. ⚠️ **Extraction** - Function-based detection (should do)
5. ⚠️ **Utilities** - Nice to have (optional)

### Effort Estimation
- **Minimal (Parser + Security):** 4-6 hours
- **Full Support (All modules):** 20-30 hours
- **With Tests:** +10-15 hours

### Risk Assessment
- **Low Risk:** Parser setup (well-documented)
- **Medium Risk:** AST node type understanding (requires study)
- **High Risk:** Language-specific patterns (requires domain knowledge)

### Recommendations
1. **Start Small:** Parser + Security middleware only
2. **Test Thoroughly:** Real-world code examples
3. **Document Patterns:** Language-specific security patterns
4. **Consider Refactoring:** Language registry pattern for future

The architecture is **coupled but manageable**. Adding a new language requires systematic changes across multiple files, but the pattern is clear and repeatable.
