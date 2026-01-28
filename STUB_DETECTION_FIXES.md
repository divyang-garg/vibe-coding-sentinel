# Stub Detection Fixes - Documented/Pending Stubs

**Date:** January 23, 2026  
**Issue:** Detection script was excluding documented/pending stubs  
**Fix:** Updated to flag documented/pending stubs while excluding only true false positives

---

## Problem Identified

The original detection script was **incorrectly excluding documented/pending stubs**, treating them as false positives. This was wrong because:

1. **Documented/pending stubs are still stubs** - they need implementation
2. **Documentation doesn't make it less of a stub** - it just documents that it's pending
3. **We want to track these** - so they don't get missed

---

## What Changed

### 1. Updated Detection Logic

**Before:**
```bash
# Excluded anything mentioning "pending", "waiting", "will be", "future"
if grep -E "$pattern" "$file" | grep -qE "(pending|waiting|will be|future)"; then
    continue  # Skip - WRONG!
fi
```

**After:**
```bash
# Check if it's a Tree-Sitter documentation comment (false positive)
# OR a real stub implementation (flag it, even if documented as pending)
if grep -E "$pattern" "$file" | grep -qE "(tree-sitter|Tree-Sitter)"; then
    # Check if this is actually a stub implementation
    if grep -E "$pattern" "$file" | grep -qE "(return nil|return.*error|return.*fmt\.Errorf)"; then
        # This IS a stub (even if documented as pending) - FLAG IT
        # Don't skip!
    else
        # Just documentation comment - skip it
        continue
    fi
fi
```

### 2. Removed Exclusion of helpers_stubs.go

**Before:**
- `helpers_stubs.go` was excluded entirely
- This could hide real production stubs

**After:**
- `helpers_stubs.go` is checked for real stubs
- Only test helper files (`helpers_stubs_test.go`) are excluded

### 3. Clarified Exclusion Criteria

**Files Excluded (True False Positives Only):**
- Test files (`*_test.go`)
- Test fixtures (`/tests/fixtures/`)
- Database operations (`task_integrations.go`) - fully implemented
- Files with Tree-Sitter integration already (`utils_business_rule.go`, etc.)

**Files NOT Excluded (Will Be Checked):**
- `helpers_stubs.go` - May contain real stubs
- Files with documented/pending stubs - These are still stubs!

---

## Examples

### ✅ NOW FLAGGED (Previously Missed):

```go
// Stub - tree-sitter integration required
func analyzeAST(code string) error {
    return fmt.Errorf("analyzeAST not implemented (tree-sitter integration required)")
}
```
**Status:** ✅ **NOW FLAGGED** - This is a real stub, even though documented as pending

```go
// NOTE: Stubbed until tree-sitter integration is complete
func getParser(language string) (*Parser, error) {
    return nil, fmt.Errorf("getParser not implemented (tree-sitter integration required)")
}
```
**Status:** ✅ **NOW FLAGGED** - Documented as pending, but still a stub

### ❌ STILL NOT FLAGGED (True False Positives):

```go
// Note: sitter import is kept for future tree-sitter integration
// This is just a comment, no stub implementation
import sitter "github.com/smacker/go-tree-sitter"
```
**Status:** ❌ **NOT FLAGGED** - Just documentation, no stub implementation

```go
// task_integrations.go - Database operations
func GetChangeRequestByID(ctx context.Context, id string) (*ChangeRequest, error) {
    query := `SELECT ... FROM change_requests WHERE id = $1`
    // ... full database implementation
}
```
**Status:** ❌ **NOT FLAGGED** - Fully implemented database operation

---

## Detection Algorithm

### Step 1: Find Stub Patterns
- Search for: `// Stub`, `not implemented`, `would be implemented`, etc.

### Step 2: Check if Real Stub
- Does it have stub implementation? (`return nil`, `return error`, empty body)
- If YES → It's a real stub (proceed to Step 3)
- If NO → It's just a comment (skip)

### Step 3: Check for False Positives
- Is it Tree-Sitter documentation with NO stub? → Skip
- Is it a database operation file? → Skip
- Is it a fully implemented file? → Skip
- **Otherwise → FLAG IT** (even if documented as pending)

### Step 4: Flag Documented/Pending Stubs
- **Documented/pending stubs are still stubs** - flag them!

---

## Key Principle

**"Documented as pending" does NOT mean "not a stub"**  
**It means "a stub that's documented as pending"**

All stubs should be flagged, regardless of documentation status:
- ✅ Undocumented stubs → Flag
- ✅ Documented stubs → Flag
- ✅ Pending stubs → Flag
- ❌ Documentation comments (no stub) → Don't flag
- ❌ Fully implemented code → Don't flag

---

## Testing

### Test 1: Documented/Pending Stubs Are Flagged
```bash
# Create test file with documented stub
echo '// Stub - tree-sitter integration required
func test() error {
    return fmt.Errorf("not implemented")
}' > test_stub.go

./scripts/detect_stubs.sh | grep test_stub.go
# Expected: test_stub.go is flagged
```

### Test 2: Documentation Comments Are Not Flagged
```bash
# Create test file with just documentation
echo '// Note: sitter import is kept for future tree-sitter integration
import sitter "github.com/smacker/go-tree-sitter"' > test_doc.go

./scripts/detect_stubs.sh | grep test_doc.go
# Expected: test_doc.go is NOT flagged
```

### Test 3: Database Operations Are Not Flagged
```bash
# task_integrations.go should not be flagged
./scripts/detect_stubs.sh | grep task_integrations.go
# Expected: task_integrations.go is NOT flagged
```

---

## Files Updated

1. ✅ `scripts/detect_stubs.sh` - Updated detection logic
2. ✅ `.githooks/pre-commit` - Updated exclusion patterns
3. ✅ `STUB_CLASSIFICATION_GUIDE.md` - Created classification guide
4. ✅ `STUB_DETECTION_FIXES.md` - This document

---

## Summary

**Before:** Documented/pending stubs were incorrectly excluded  
**After:** Documented/pending stubs are correctly flagged  
**Result:** No stubs are missed, only true false positives are excluded
