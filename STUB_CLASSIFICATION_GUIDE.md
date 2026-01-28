# Stub Classification Guide

**Purpose:** Clarify what types of stubs exist and how they should be handled by detection scripts.

---

## Stub Categories

### 1. ✅ **Real Stubs (MUST BE FLAGGED)**

These are actual stub implementations that need to be completed:

#### A. Documented/Pending Stubs
**Definition:** Stubs that are documented as pending implementation but are still stubs.

**Examples:**
```go
// Stub - tree-sitter integration required
func analyzeAST(code string) error {
    return fmt.Errorf("analyzeAST not implemented (tree-sitter integration required)")
}

// NOTE: Stubbed until tree-sitter integration is complete
func getParser(language string) (*Parser, error) {
    return nil, fmt.Errorf("getParser not implemented (tree-sitter integration required)")
}
```

**Detection:** ✅ **SHOULD BE FLAGGED** - These are real stubs that need implementation, even if documented as pending.

**Why:** Documentation doesn't make it less of a stub - it still needs to be implemented.

---

#### B. Undocumented Stubs
**Definition:** Stubs without documentation that need implementation.

**Examples:**
```go
// Stub
func processData(data []byte) error {
    return nil
}

func calculateMetrics() float64 {
    // not implemented
    return 0.0
}
```

**Detection:** ✅ **SHOULD BE FLAGGED** - These are real stubs.

---

### 2. ❌ **False Positives (SHOULD NOT BE FLAGGED)**

These are NOT stubs - they're documentation comments or fully implemented code:

#### A. Tree-Sitter Documentation Comments
**Definition:** Comments that mention Tree-Sitter but are NOT actual stub implementations.

**Examples:**
```go
// Note: sitter import is kept for future tree-sitter integration
// This is NOT a stub - it's documentation

// Tree-sitter parsers are initialized in ast/parsers.go
// This is NOT a stub - it's a comment about where code exists
```

**Detection:** ❌ **SHOULD NOT BE FLAGGED** - These are documentation, not stubs.

**How to Distinguish:**
- If the comment is about Tree-Sitter but there's NO actual stub implementation (no `return nil`, no `fmt.Errorf("not implemented")`, no empty function body)
- If it's just explaining where Tree-Sitter code exists or will exist

---

#### B. Database Operations
**Definition:** Files that perform database operations (not code analysis).

**Examples:**
```go
// task_integrations.go - Database operations
func GetChangeRequestByID(ctx context.Context, id string) (*ChangeRequest, error) {
    // Database query - fully implemented
    query := `SELECT ... FROM change_requests WHERE id = $1`
    // ... full implementation
}
```

**Detection:** ❌ **SHOULD NOT BE FLAGGED** - These are fully implemented database operations.

**Why:** Database operations don't need Tree-Sitter or AST parsing - they're complete as-is.

---

#### C. Fully Implemented Files with Tree-Sitter
**Definition:** Files that already have Tree-Sitter integration implemented.

**Examples:**
```go
// dependency_detector_helpers.go - Has Tree-Sitter integration
func extractSymbolsFromAST(code string, language string) map[string]bool {
    parser, err := GetParser(language)  // ✅ Uses Tree-Sitter
    tree, err := parser.ParseCtx(ctx, nil, []byte(code))  // ✅ Full implementation
    // ... full AST parsing implementation
}
```

**Detection:** ❌ **SHOULD NOT BE FLAGGED** - These are fully implemented.

---

### 3. ⚠️ **Intentional Stubs (NEEDS CLARIFICATION)**

**Question:** What are "intentional stubs"?

**Possible Meanings:**

#### A. Test Helper Stubs
**Definition:** Stubs created specifically for testing purposes.

**Example:**
```go
// helpers_stubs.go - Test helper stubs
func selectModelWithDepth(ctx context.Context, projectID string, config *LLMConfig, mode string, depth int, feature string) (string, error) {
    return config.Model, nil  // Simplified for testing
}
```

**Decision:** 
- If these are **test-only helpers** that are meant to remain simple: ❌ Don't flag
- If these are **production code stubs** that need implementation: ✅ Flag

**Current Status:** `helpers_stubs.go` appears to have production code stubs, not just test helpers.

---

#### B. Deprecated Stubs
**Definition:** Stubs that are marked as deprecated and will be removed.

**Example:**
```go
// Deprecated: Use newFunction() instead
func oldFunction() error {
    // Stub - deprecated
    return nil
}
```

**Decision:** ✅ **SHOULD BE FLAGGED** - Even if deprecated, they should be removed or replaced.

---

## Detection Rules

### ✅ FLAG These (Real Stubs):

1. **Any function with stub pattern + actual stub implementation:**
   - `return nil` with stub comment
   - `return fmt.Errorf("not implemented")`
   - Empty function body with stub comment
   - `// Stub` comment with minimal/no implementation

2. **Documented/Pending Stubs:**
   - Comments like "Stub - tree-sitter integration required"
   - Comments like "NOTE: Stubbed until X is complete"
   - **Even if documented as pending, these are still stubs!**

3. **Functions with "Stub" in name:**
   - `func somethingStub()`
   - Unless it's a test helper file

### ❌ DON'T FLAG These (False Positives):

1. **Tree-Sitter Documentation Comments:**
   - Comments that mention Tree-Sitter but have NO stub implementation
   - Comments explaining where Tree-Sitter code exists
   - Import comments about Tree-Sitter

2. **Database Operations:**
   - Files like `task_integrations.go` that do database queries
   - These are fully implemented, just don't use Tree-Sitter (and don't need to)

3. **Fully Implemented Files:**
   - Files that already use Tree-Sitter
   - Files with complete implementations

4. **Test Files:**
   - `*_test.go` files
   - Test fixtures

---

## Updated Detection Logic

### Key Principle:
**"Documented as pending" does NOT mean "not a stub" - it means "a stub that's documented as pending"**

### Detection Algorithm:

1. **Find stub patterns** in code
2. **Check if it's a real stub:**
   - Does it have `return nil`, `return error`, or empty body?
   - If YES → It's a real stub (FLAG IT)
   - If NO → It's just a comment (DON'T FLAG)

3. **Check for false positives:**
   - Is it a Tree-Sitter documentation comment with NO stub implementation? → Don't flag
   - Is it a database operation file? → Don't flag
   - Is it a fully implemented file? → Don't flag

4. **Flag documented/pending stubs:**
   - Even if comment says "pending" or "will be implemented"
   - These are still stubs that need to be completed

---

## Examples

### ✅ SHOULD FLAG (Real Stub):
```go
// Stub - tree-sitter integration required
func analyzeAST(code string) error {
    return fmt.Errorf("analyzeAST not implemented (tree-sitter integration required)")
}
```

### ❌ SHOULD NOT FLAG (False Positive):
```go
// Note: sitter import is kept for future tree-sitter integration
// This is just a comment, no stub implementation
import sitter "github.com/smacker/go-tree-sitter"
```

### ✅ SHOULD FLAG (Documented/Pending Stub):
```go
// NOTE: Stubbed until tree-sitter integration is complete
func getParser(language string) (*Parser, error) {
    return nil, fmt.Errorf("getParser not implemented (tree-sitter integration required)")
}
```

### ❌ SHOULD NOT FLAG (Database Operation):
```go
// task_integrations.go
func GetChangeRequestByID(ctx context.Context, id string) (*ChangeRequest, error) {
    query := `SELECT ... FROM change_requests WHERE id = $1`
    // ... full database implementation
}
```

---

## Summary

| Type | Flag? | Reason |
|------|-------|--------|
| Documented/Pending Stub | ✅ YES | Still a stub, needs implementation |
| Undocumented Stub | ✅ YES | Real stub |
| Tree-Sitter Doc Comment | ❌ NO | Just documentation, no stub |
| Database Operation | ❌ NO | Fully implemented |
| Fully Implemented | ❌ NO | Not a stub |
| Test Helper Stub | ⚠️ MAYBE | Depends on if production code |

**Key Point:** Documentation status doesn't change whether something is a stub - if it's a stub implementation, it should be flagged regardless of documentation.
