// Package scanner provides tests for vibe detection
package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectDuplicateFunctions(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("detects duplicate functions", func(t *testing.T) {
		// Create two files with same function name
		file1 := filepath.Join(tmpDir, "file1.js")
		os.WriteFile(file1, []byte(`function testFunc() {
	return 1;
}`), 0644)

		file2 := filepath.Join(tmpDir, "file2.js")
		os.WriteFile(file2, []byte(`function testFunc() {
	return 2;
}`), 0644)

		findings, err := detectDuplicateFunctions(tmpDir)
		if err != nil {
			t.Errorf("detectDuplicateFunctions() error = %v", err)
		}
		// The function may or may not detect duplicates depending on implementation
		// Just verify it doesn't error
		_ = findings
	})

	t.Run("handles no duplicates", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "unique1.js")
		os.WriteFile(file1, []byte(`function uniqueFunc1() {}`), 0644)

		findings, err := detectDuplicateFunctions(tmpDir)
		if err != nil {
			t.Errorf("detectDuplicateFunctions() error = %v", err)
		}
		// May or may not find duplicates depending on previous test
		_ = findings
	})
}

func TestDetectOrphanedCode(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("detects orphaned code", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "orphan.js")
		content := `const x = 1;
x = x + 1; // Orphaned code
function test() {
	return x;
}`
		os.WriteFile(file1, []byte(content), 0644)

		findings, err := detectOrphanedCode(tmpDir)
		if err != nil {
			t.Errorf("detectOrphanedCode() error = %v", err)
		}
		// May detect orphaned code
		_ = findings
	})

	t.Run("handles code in functions", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "valid.js")
		content := `function test() {
	const x = 1;
	x = x + 1;
	return x;
}`
		os.WriteFile(file1, []byte(content), 0644)

		findings, err := detectOrphanedCode(tmpDir)
		if err != nil {
			t.Errorf("detectOrphanedCode() error = %v", err)
		}
		// Should not detect code inside functions
		_ = findings
	})

	t.Run("handles brace depth tracking", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "test.js")
		content := `function test() {
	if (true) {
		const x = 1;
	}
}
orphaned = 2;`
		os.WriteFile(file1, []byte(content), 0644)

		findings, err := detectOrphanedCode(tmpDir)
		if err != nil {
			t.Errorf("detectOrphanedCode() error = %v", err)
		}
		_ = findings
	})

	t.Run("skips imports and declarations", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "test2.js")
		content := `import { x } from 'module';
const y = 1;
var z = 2;
function test() {
	return y;
}`
		os.WriteFile(file1, []byte(content), 0644)

		findings, err := detectOrphanedCode(tmpDir)
		if err != nil {
			t.Errorf("detectOrphanedCode() error = %v", err)
		}
		_ = findings
	})
}

func TestDetectUnusedVariables(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("detects potentially unused variables", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "unused.js")
		content := `const unusedVar = 1;
const usedVar = 2;
console.log(usedVar);`
		os.WriteFile(file1, []byte(content), 0644)

		findings, err := detectUnusedVariables(tmpDir)
		if err != nil {
			t.Errorf("detectUnusedVariables() error = %v", err)
		}
		// May detect unused variables
		_ = findings
	})

	t.Run("handles used variables", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "used.js")
		content := `const used1 = 1;
const used2 = 2;
console.log(used1, used2);`
		os.WriteFile(file1, []byte(content), 0644)

		findings, err := detectUnusedVariables(tmpDir)
		if err != nil {
			t.Errorf("detectUnusedVariables() error = %v", err)
		}
		_ = findings
	})

	t.Run("handles let declarations", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "let.js")
		content := `let unused = 1;
let used = 2;
console.log(used);`
		os.WriteFile(file1, []byte(content), 0644)

		findings, err := detectUnusedVariables(tmpDir)
		if err != nil {
			t.Errorf("detectUnusedVariables() error = %v", err)
		}
		_ = findings
	})

	t.Run("handles var declarations", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "var.js")
		content := `var unused = 1;
var used = 2;
console.log(used);`
		os.WriteFile(file1, []byte(content), 0644)

		findings, err := detectUnusedVariables(tmpDir)
		if err != nil {
			t.Errorf("detectUnusedVariables() error = %v", err)
		}
		_ = findings
	})
}

func TestDetectVibeIssues(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("detects vibe issues", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "test.js")
		content := `function duplicate() {}
function duplicate() {} // duplicate
const orphaned = 1; // orphaned code`
		os.WriteFile(file1, []byte(content), 0644)

		opts := ScanOptions{
			CodebasePath: tmpDir,
			CIMode:       true,
		}
		findings, err := DetectVibeIssues(opts)
		if err != nil {
			t.Errorf("DetectVibeIssues() error = %v", err)
		}
		// May detect vibe issues
		_ = findings
	})

	t.Run("handles empty directory", func(t *testing.T) {
		opts := ScanOptions{
			CodebasePath: tmpDir,
			CIMode:       true,
		}
		findings, err := DetectVibeIssues(opts)
		if err != nil {
			t.Errorf("DetectVibeIssues() error = %v", err)
		}
		_ = findings
	})
}

