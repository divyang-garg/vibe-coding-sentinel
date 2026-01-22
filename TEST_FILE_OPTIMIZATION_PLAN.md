# Test File Optimization Plan

**Date:** January 20, 2026  
**Status:** ⚠️ **ACTION REQUIRED**

---

## Summary

4 test files exceed the CODING_STANDARDS.md limit of 500 lines. These files need to be split into smaller, logically organized files.

---

## Files Requiring Split

### 1. `ast/extraction_test.go` (1004 lines → Split into 4 files)

**Current Structure:**
- `TestExtractFunctions_Go` (lines 10-111) - 102 lines
- `TestExtractFunctions_JavaScript` (lines 113-329) - 217 lines
- `TestExtractFunctions_TypeScript` (lines 331-422) - 92 lines
- `TestExtractFunctions_Python` (lines 424-451) - 28 lines
- `TestExtractFunctions_ErrorHandling` (lines 453-505) - 53 lines
- `TestExtractFunctionByName` (lines 507-549) - 43 lines
- `TestExtractFunctions_Visibility` (lines 551-637) - 87 lines
- `TestExtractFunctions_ClassMethods` (lines 639-707) - 69 lines
- `TestExtractFunctions_Parameters` (lines 709-836) - 128 lines
- `TestExtractFunctions_ReturnTypes` (lines 838-928) - 91 lines
- `TestExtractFunctions_Documentation` (lines 930-1004) - 75 lines

**Split Plan:**
1. **`extraction_go_test.go`** (~250 lines)
   - `TestExtractFunctions_Go`
   - Go-specific tests

2. **`extraction_js_ts_test.go`** (~350 lines)
   - `TestExtractFunctions_JavaScript`
   - `TestExtractFunctions_TypeScript`
   - JS/TS-specific tests

3. **`extraction_python_test.go`** (~200 lines)
   - `TestExtractFunctions_Python`
   - Python-specific tests

4. **`extraction_helpers_test.go`** (~450 lines)
   - `TestExtractFunctions_ErrorHandling`
   - `TestExtractFunctionByName`
   - `TestExtractFunctions_Visibility`
   - `TestExtractFunctions_ClassMethods`
   - `TestExtractFunctions_Parameters`
   - `TestExtractFunctions_ReturnTypes`
   - `TestExtractFunctions_Documentation`

**Note:** The helpers file will be ~450 lines, which is under 500, so it's compliant.

### 2. `handlers/ast_handler_e2e_test.go` (651 lines → Split into 2 files)

**Split Plan:**
1. **`ast_handler_e2e_analyze_test.go`** (~400 lines)
   - `TestASTEndToEnd_AnalyzeAST`
   - `TestASTEndToEnd_AnalyzeMultiFile`
   - `TestASTEndToEnd_AnalyzeSecurity`
   - `TestASTEndToEnd_AnalyzeCrossFile`
   - `setupTestRouter()` helper function

2. **`ast_handler_e2e_support_test.go`** (~250 lines)
   - `TestASTEndToEnd_GetSupportedAnalyses`
   - `TestASTEndToEnd_RealWorldScenarios`

### 3. `ast/symbol_table_coverage_test.go` (646 lines → Split into 2 files)

**Split Plan:**
1. **`symbol_extraction_test.go`** (~400 lines)
   - `TestExtractJSSymbol`
   - `TestExtractPythonSymbol`
   - `TestExtractTypeName`
   - `TestExtractExportName`
   - `TestExtractClassName`
   - `TestExtractSymbolsFromFile_JavaScript`
   - `TestExtractSymbolsFromFile_Python`

2. **`symbol_table_operations_test.go`** (~250 lines)
   - `TestScopeStack`
   - `TestSymbolTable_AddReference`
   - `TestSymbolTable_GetFileSymbols`

### 4. `ast/coverage_additional_test.go` (619 lines → Split into 2 files)

**Split Plan:**
1. **`parser_dependency_coverage_test.go`** (~350 lines)
   - `TestAnalyzeAST_PanicRecovery`
   - `TestGetParser_UnsupportedLanguage`
   - `TestCreateParserForLanguage`
   - `TestDependencyGraph_GetDependents`
   - `TestDependencyGraph_ExtractJSImport`
   - `TestDependencyGraph_ExtractPythonImport`
   - `TestContainsString`
   - `TestDetectCrossFileDuplicates`
   - `TestAnalyzeMultiFile`

2. **`search_pattern_coverage_test.go`** (~270 lines)
   - `TestSearchCodebaseCached`
   - `TestCountReferences`
   - `TestFindImports`
   - `TestBuildImportPattern`
   - `TestBuildExportPattern`
   - `TestValidatePattern`
   - `TestIsValidIdentifier`
   - `TestExtractLanguageFromPath`
   - `TestSearchWithTimeout`
   - `TestHasTemplateLiteral`
   - `TestHasStringFormatting`
   - `TestHasUnescapedUserInput`
   - `TestIsSubprocessCall`
   - `TestDetectGenericReflection`
   - `TestExtractClassNameFromParent`
   - `TestExtractRestParameter`

---

## Implementation Steps

1. **Backup Original Files**
   ```bash
   cd hub/api
   cp ast/extraction_test.go ast/extraction_test.go.backup
   cp handlers/ast_handler_e2e_test.go handlers/ast_handler_e2e_test.go.backup
   cp ast/symbol_table_coverage_test.go ast/symbol_table_coverage_test.go.backup
   cp ast/coverage_additional_test.go ast/coverage_additional_test.go.backup
   ```

2. **Create Split Files** (see detailed file contents below)

3. **Delete Original Files** (after verification)

4. **Run Tests** to verify all tests still pass:
   ```bash
   go test ./ast/... -v
   go test ./handlers/... -v
   ```

5. **Verify Line Counts**:
   ```bash
   find . -name "*_test.go" -exec wc -l {} + | sort -rn
   ```

---

## Expected Results

After splitting:
- ✅ All test files ≤ 500 lines
- ✅ All tests passing
- ✅ Better code organization
- ✅ Easier maintenance
- ✅ Compliance with CODING_STANDARDS.md

---

**Next Steps:** Implement file splits as detailed above.
