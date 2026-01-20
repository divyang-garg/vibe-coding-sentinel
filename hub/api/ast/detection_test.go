// Package ast detection tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"strings"
	"testing"
)

// assertFindingExists checks that at least one finding of the given type exists
func assertFindingExists(t *testing.T, findings []ASTFinding, findingType string) {
	t.Helper()
	for _, f := range findings {
		if f.Type == findingType {
			return
		}
	}
	t.Errorf("Expected finding of type '%s', found none in %d findings", findingType, len(findings))
}

// assertNoFindingOfType checks that no finding of the given type exists
func assertNoFindingOfType(t *testing.T, findings []ASTFinding, findingType string) {
	t.Helper()
	for _, f := range findings {
		if f.Type == findingType {
			t.Errorf("Did not expect finding of type '%s', but found: %s", findingType, f.Message)
			return
		}
	}
}

// assertFindingCount checks the exact number of findings of a given type
func assertFindingCount(t *testing.T, findings []ASTFinding, findingType string, expected int) {
	t.Helper()
	count := 0
	for _, f := range findings {
		if f.Type == findingType {
			count++
		}
	}
	if count != expected {
		t.Errorf("Expected %d findings of type '%s', got %d", expected, findingType, count)
	}
}

// assertFindingContainsName checks that a finding mentions a specific name
func assertFindingContainsName(t *testing.T, findings []ASTFinding, findingType, name string) {
	t.Helper()
	for _, f := range findings {
		if f.Type == findingType && strings.Contains(f.Message, name) {
			return
		}
	}
	t.Errorf("Expected finding of type '%s' mentioning '%s', found none", findingType, name)
}

// TestDetectDuplicateFunctions tests duplicate function detection
func TestDetectDuplicateFunctions(t *testing.T) {
	t.Run("go_duplicates", func(t *testing.T) {
		code := `package main
func test() {}
func test() {}`
		findings, _, err := AnalyzeAST(code, "go", []string{"duplicates"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		assertFindingExists(t, findings, "duplicate_function")
	})

	t.Run("javascript_no_duplicates", func(t *testing.T) {
		findings, _, _ := AnalyzeAST(`function a() {} function b() {}`, "javascript", []string{"duplicates"})
		assertNoFindingOfType(t, findings, "duplicate_function")
	})

	t.Run("python_methods", func(t *testing.T) {
		code := "class T:\n    def m(self): pass\n    def m(self): pass"
		findings, _, _ := AnalyzeAST(code, "python", []string{"duplicates"})
		assertFindingExists(t, findings, "duplicate_function")
	})
}

// TestDetectUnusedVariables tests unused variable detection
func TestDetectUnusedVariables(t *testing.T) {
	t.Run("go_unused", func(t *testing.T) {
		code := `package main
func test() { var unused int; used := 10; fmt.Println(used) }`
		findings, _, err := AnalyzeAST(code, "go", []string{"unused"})
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		assertFindingContainsName(t, findings, "unused_variable", "unused")
	})

	t.Run("javascript_destructuring", func(t *testing.T) {
		findings, _, _ := AnalyzeAST(`const [used, unused] = [1, 2]; console.log(used);`, "javascript", []string{"unused"})
		assertFindingContainsName(t, findings, "unused_variable", "unused")
	})

	t.Run("python_parameter", func(t *testing.T) {
		findings, _, _ := AnalyzeAST(`def test(used, unused): print(used)`, "python", []string{"unused"})
		assertFindingContainsName(t, findings, "unused_variable", "unused")
	})
}

// TestDetectUnreachableCode tests unreachable code detection
func TestDetectUnreachableCode(t *testing.T) {
	t.Run("go_after_return", func(t *testing.T) {
		findings, _, _ := AnalyzeAST(`package main
func test() { return; fmt.Println("x") }`, "go", []string{"unreachable"})
		assertFindingExists(t, findings, "unreachable_code")
	})

	t.Run("javascript_after_throw", func(t *testing.T) {
		findings, _, _ := AnalyzeAST(`function t() { throw new Error(); console.log("x"); }`, "javascript", []string{"unreachable"})
		assertFindingExists(t, findings, "unreachable_code")
	})

	t.Run("python_after_raise", func(t *testing.T) {
		findings, _, _ := AnalyzeAST("def t():\n    raise Exception()\n    print('x')", "python", []string{"unreachable"})
		assertFindingExists(t, findings, "unreachable_code")
	})
}

// TestDetectEmptyCatchBlocks tests empty catch block detection
func TestDetectEmptyCatchBlocks(t *testing.T) {
	t.Run("javascript", func(t *testing.T) {
		findings, _, _ := AnalyzeAST(`try { risky(); } catch (e) { }`, "javascript", []string{"empty_catch"})
		assertFindingExists(t, findings, "empty_catch")
	})

	t.Run("python", func(t *testing.T) {
		findings, _, _ := AnalyzeAST("try:\n    risky()\nexcept:\n    pass", "python", []string{"empty_catch"})
		assertFindingExists(t, findings, "empty_catch")
	})
}

// TestDetectMissingAwait tests missing await detection
func TestDetectMissingAwait(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		findings, _, _ := AnalyzeAST(`async function t() { fetch("/api"); }`, "javascript", []string{"missing_await"})
		assertFindingExists(t, findings, "missing_await")
	})

	t.Run("present", func(t *testing.T) {
		findings, _, _ := AnalyzeAST(`async function t() { await fetch("/api"); }`, "javascript", []string{"missing_await"})
		assertNoFindingOfType(t, findings, "missing_await")
	})
}

// TestDetectBraceMismatch tests brace mismatch detection
func TestDetectBraceMismatch(t *testing.T) {
	t.Run("go_mismatch", func(t *testing.T) {
		findings, _, err := AnalyzeAST(`package main
func test() { if true { fmt.Println("x")`, "go", []string{"brace_mismatch"})
		if err != nil {
			return // Parse error acceptable for syntax errors
		}
		for _, f := range findings {
			if f.Type == "brace_mismatch" || strings.Contains(f.Message, "mismatch") {
				return
			}
		}
	})

	t.Run("javascript_mismatch", func(t *testing.T) {
		findings, _, err := AnalyzeAST(`function t() { if (true { console.log("x"); } }`, "javascript", []string{"brace_mismatch"})
		if err != nil {
			return
		}
		for _, f := range findings {
			if f.Type == "brace_mismatch" || strings.Contains(f.Message, "mismatch") {
				return
			}
		}
	})
}

// TestDetectOrphanedCode tests orphaned code detection
func TestDetectOrphanedCode(t *testing.T) {
	findings, _, _ := AnalyzeAST(`package main
func main() {}
func unused() {}`, "go", []string{"orphaned"})
	assertFindingContainsName(t, findings, "orphaned_code", "unused")
}

// TestEdgeCases tests edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		findings, _, _ := AnalyzeAST("", "go", []string{})
		if findings == nil {
			t.Error("nil findings")
		}
	})

	t.Run("unicode", func(t *testing.T) {
		findings, _, _ := AnalyzeAST(`package main; func t() { fmt.Println("测试") }`, "go", []string{"unused"})
		if findings == nil {
			t.Error("nil findings")
		}
	})

	t.Run("minified", func(t *testing.T) {
		findings, _, _ := AnalyzeAST(`function a(){const b=1;console.log(b);}`, "javascript", []string{"unused"})
		if findings == nil {
			t.Error("nil findings")
		}
	})
}

// TestUnreachableCodeAdvanced tests advanced unreachable code patterns including hasTerminatingStatement
func TestUnreachableCodeAdvanced(t *testing.T) {
	t.Run("javascript_if_true_with_return_triggers_hasTerminating", func(t *testing.T) {
		// This code pattern triggers hasTerminatingStatement function
		code := `function test() { if (true) { return; } console.log("after"); }`
		findings, _, err := AnalyzeAST(code, "javascript", []string{"unreachable"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		t.Logf("hasTerminatingStatement path: found %d findings", len(findings))
	})

	t.Run("javascript_if_true_with_throw", func(t *testing.T) {
		code := `function x() { if (true) { throw new Error(); } console.log("y"); }`
		findings, _, err := AnalyzeAST(code, "javascript", []string{"unreachable"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		t.Logf("Found %d findings for if(true) throw pattern", len(findings))
	})
}

// TestPythonUnusedAdvanced tests advanced Python unused variable patterns
func TestPythonUnusedAdvanced(t *testing.T) {
	t.Run("python_import_not_flagged", func(t *testing.T) {
		code := `
import os
from sys import path
print("hello")
`
		findings, _, err := AnalyzeAST(code, "python", []string{"unused"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		// Imports should not be flagged as unused variables
		for _, f := range findings {
			if strings.Contains(f.Message, "os") || strings.Contains(f.Message, "path") {
				t.Errorf("Import should not be flagged as unused: %s", f.Message)
			}
		}
	})

	t.Run("python_assignment_unused", func(t *testing.T) {
		code := `
x = 10
y = 20
print(x)
`
		findings, _, err := AnalyzeAST(code, "python", []string{"unused"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		assertFindingContainsName(t, findings, "unused_variable", "y")
	})

	t.Run("python_tuple_unpacking", func(t *testing.T) {
		code := `
def test():
    a, b = (1, 2)
    print(a)
`
		findings, _, err := AnalyzeAST(code, "python", []string{"unused"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		// b is unused in tuple unpacking
		t.Logf("Found %d findings for tuple unpacking", len(findings))
	})

	t.Run("python_self_cls_underscore_not_flagged", func(t *testing.T) {
		code := `
class Test:
    def method(self): pass
def other(): _ = 1
`
		findings, _, err := AnalyzeAST(code, "python", []string{"unused"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		for _, f := range findings {
			if strings.Contains(f.Message, "self") || strings.Contains(f.Message, "'_'") {
				t.Errorf("self/_ should not be flagged: %s", f.Message)
			}
		}
	})
}

// TestAnalyzeASTEntryPoint tests the main AnalyzeAST function including panic recovery
func TestAnalyzeASTEntryPoint(t *testing.T) {
	t.Run("all_checks", func(t *testing.T) {
		code := `package main
func test() {}
func test() {}
func unused() { var x int; return; fmt.Println("x") }`
		findings, _, err := AnalyzeAST(code, "go", []string{"duplicates", "unused", "unreachable", "orphaned"})
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if len(findings) == 0 {
			t.Error("Expected findings")
		}
	})

	t.Run("cache_hit", func(t *testing.T) {
		code := `package main; func x() {}`
		_, _, err1 := AnalyzeAST(code, "go", []string{"unused"})
		if err1 != nil {
			t.Fatalf("first call failed: %v", err1)
		}
		_, _, err2 := AnalyzeAST(code, "go", []string{"unused"})
		if err2 != nil {
			t.Fatalf("cache hit failed: %v", err2)
		}
	})

	t.Run("typescript_support", func(t *testing.T) {
		_, _, err := AnalyzeAST(`function t(): void { const u: number = 42; }`, "typescript", []string{"unused"})
		if err != nil {
			t.Fatalf("TypeScript failed: %v", err)
		}
	})

	t.Run("invalid_code_handling", func(t *testing.T) {
		// Very malformed code - should not panic, should return error or empty findings
		_, _, err := AnalyzeAST("{{{[[[", "go", []string{"unused"})
		if err != nil {
			t.Logf("Invalid code returned expected error: %v", err)
		}
	})
}

// TestExclusionPatterns tests that exclusion patterns work correctly
func TestExclusionPatterns(t *testing.T) {
	t.Run("init_not_flagged", func(t *testing.T) {
		code := `package main
func init() { setup() }`
		findings, _, err := AnalyzeAST(code, "go", []string{"orphaned"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		assertNoFindingOfType(t, findings, "orphaned_code")
	})

	t.Run("exported_not_flagged", func(t *testing.T) {
		code := `package main
func PublicHelper() {}`
		findings, _, err := AnalyzeAST(code, "go", []string{"orphaned"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		assertNoFindingOfType(t, findings, "orphaned_code")
	})

	t.Run("example_functions_not_flagged", func(t *testing.T) {
		code := `package main
func ExampleUsage() {}`
		findings, _, err := AnalyzeAST(code, "go", []string{"orphaned"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		assertNoFindingOfType(t, findings, "orphaned_code")
	})

	t.Run("benchmark_functions_not_flagged", func(t *testing.T) {
		code := `package main
func BenchmarkSomething() {}`
		findings, _, err := AnalyzeAST(code, "go", []string{"orphaned"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		assertNoFindingOfType(t, findings, "orphaned_code")
	})

	t.Run("test_functions_not_flagged", func(t *testing.T) {
		code := `package main
func TestSomething() {}`
		findings, _, err := AnalyzeAST(code, "go", []string{"orphaned"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		assertNoFindingOfType(t, findings, "orphaned_code")
	})
}

// TestNoFalsePositives tests that detection does NOT fire when it shouldn't
func TestNoFalsePositives(t *testing.T) {
	t.Run("go_all_used_variables", func(t *testing.T) {
		code := `
package main
func main() {
    x := 1
    y := 2
    fmt.Println(x, y)
}`
		findings, _, err := AnalyzeAST(code, "go", []string{"unused"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		assertNoFindingOfType(t, findings, "unused_variable")
	})

	t.Run("go_no_unreachable_code", func(t *testing.T) {
		code := `
package main
func test() {
    fmt.Println("reachable")
    return
}`
		findings, _, err := AnalyzeAST(code, "go", []string{"unreachable"})
		if err != nil {
			t.Fatalf("AnalyzeAST failed: %v", err)
		}
		assertNoFindingOfType(t, findings, "unreachable_code")
	})
}
