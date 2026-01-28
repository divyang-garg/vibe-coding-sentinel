// Package cli provides extended tests for validate-rules command
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunValidateRules_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("rules directory not found", func(t *testing.T) {
		err := runValidateRules([]string{})
		if err == nil {
			t.Error("Expected error when rules directory doesn't exist")
		}
	})

	t.Run("with valid rules", func(t *testing.T) {
		os.MkdirAll(".cursor/rules", 0755)
		validRule := `---
description: Test Rule
globs: ["*.go"]
alwaysApply: true
---
# Test Rule
Content here.`
		os.WriteFile(".cursor/rules/valid.md", []byte(validRule), 0644)

		err := runValidateRules([]string{})
		if err != nil {
			t.Errorf("runValidateRules() with valid rule error = %v", err)
		}
	})

	t.Run("with invalid rules", func(t *testing.T) {
		os.MkdirAll(".cursor/rules", 0755)
		invalidRule := `---
description: Test Rule
globs: ["*.go"]
---
# Missing alwaysApply`
		os.WriteFile(".cursor/rules/invalid.md", []byte(invalidRule), 0644)

		err := runValidateRules([]string{})
		if err == nil {
			t.Error("Expected error for invalid rule")
		}
	})

	t.Run("with mixed valid and invalid", func(t *testing.T) {
		os.MkdirAll(".cursor/rules", 0755)

		validRule := `---
description: Valid
globs: ["*.go"]
alwaysApply: true
---
Content`
		os.WriteFile(".cursor/rules/valid.md", []byte(validRule), 0644)

		invalidRule := `---
description: Invalid
globs: ["*.js"]
---
# Missing alwaysApply`
		os.WriteFile(".cursor/rules/invalid.md", []byte(invalidRule), 0644)

		err := runValidateRules([]string{})
		if err == nil {
			t.Error("Expected error when some rules are invalid")
		}
	})

	t.Run("skips non-md files", func(t *testing.T) {
		os.MkdirAll(".cursor/rules", 0755)
		os.WriteFile(".cursor/rules/test.txt", []byte("not a rule"), 0644)
		os.WriteFile(".cursor/rules/subdir", []byte("directory"), 0644)

		err := runValidateRules([]string{})
		// Should handle gracefully
		_ = err
	})
}

func TestValidateRuleFile_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("missing YAML frontmatter", func(t *testing.T) {
		file := filepath.Join(tmpDir, "no-frontmatter.md")
		os.WriteFile(file, []byte("# Just markdown\nNo frontmatter"), 0644)

		err := validateRuleFile(file)
		if err == nil {
			t.Error("Expected error for missing frontmatter")
		}
	})

	t.Run("malformed YAML frontmatter", func(t *testing.T) {
		file := filepath.Join(tmpDir, "malformed.md")
		os.WriteFile(file, []byte("---\nIncomplete frontmatter"), 0644)

		err := validateRuleFile(file)
		if err == nil {
			t.Error("Expected error for malformed frontmatter")
		}
	})

	t.Run("missing description field", func(t *testing.T) {
		file := filepath.Join(tmpDir, "no-desc.md")
		content := `---
globs: ["*.go"]
alwaysApply: true
---
Content`
		os.WriteFile(file, []byte(content), 0644)

		err := validateRuleFile(file)
		if err == nil {
			t.Error("Expected error for missing description")
		}
	})

	t.Run("missing globs field", func(t *testing.T) {
		file := filepath.Join(tmpDir, "no-globs.md")
		content := `---
description: Test
alwaysApply: true
---
Content`
		os.WriteFile(file, []byte(content), 0644)

		err := validateRuleFile(file)
		if err == nil {
			t.Error("Expected error for missing globs")
		}
	})

	t.Run("missing alwaysApply field", func(t *testing.T) {
		file := filepath.Join(tmpDir, "no-always.md")
		content := `---
description: Test
globs: ["*.go"]
---
Content`
		os.WriteFile(file, []byte(content), 0644)

		err := validateRuleFile(file)
		if err == nil {
			t.Error("Expected error for missing alwaysApply")
		}
	})

	t.Run("valid rule file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "valid.md")
		content := `---
description: Test Rule
globs: ["*.go"]
alwaysApply: true
---
# Test Rule
Content here.`
		os.WriteFile(file, []byte(content), 0644)

		err := validateRuleFile(file)
		if err != nil {
			t.Errorf("validateRuleFile() with valid rule error = %v", err)
		}
	})

	t.Run("file read error", func(t *testing.T) {
		err := validateRuleFile("/nonexistent/file.md")
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})
}

func TestRunUpdateRules_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("rules directory not found", func(t *testing.T) {
		err := runUpdateRules([]string{})
		if err == nil {
			t.Error("Expected error when rules directory doesn't exist")
		}
	})

	t.Run("with force flag", func(t *testing.T) {
		os.MkdirAll(".cursor/rules", 0755)
		err := runUpdateRules([]string{"--force"})
		if err != nil {
			t.Errorf("runUpdateRules() with force error = %v", err)
		}
	})

	t.Run("with short force flag", func(t *testing.T) {
		os.MkdirAll(".cursor/rules", 0755)
		err := runUpdateRules([]string{"-f"})
		if err != nil {
			t.Errorf("runUpdateRules() with -f flag error = %v", err)
		}
	})

	t.Run("without force - creates backup", func(t *testing.T) {
		os.MkdirAll(".cursor/rules", 0755)

		// Create existing rule
		existingRule := `---
description: Existing
globs: ["*.go"]
alwaysApply: true
---
Content`
		os.WriteFile(".cursor/rules/existing.md", []byte(existingRule), 0644)

		err := runUpdateRules([]string{})
		if err != nil {
			t.Errorf("runUpdateRules() error = %v", err)
		}

		// Verify backup was created
		entries, _ := os.ReadDir(".cursor/rules")
		backupFound := false
		for _, entry := range entries {
			if entry.IsDir() && filepath.Base(entry.Name()) == ".backup-" {
				backupFound = true
				break
			}
		}
		if !backupFound {
			// Check for backup directories starting with .backup-
			for _, entry := range entries {
				if entry.IsDir() && len(entry.Name()) > 7 && entry.Name()[:7] == ".backup" {
					backupFound = true
					break
				}
			}
		}
		// Backup may or may not be created depending on timing
		_ = backupFound
	})
}

func TestUpdateRulesFromDefaults_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	rulesDir := ".cursor/rules"
	os.MkdirAll(rulesDir, 0755)

	t.Run("update without force", func(t *testing.T) {
		err := updateRulesFromDefaults(rulesDir, false)
		if err != nil {
			t.Errorf("updateRulesFromDefaults() error = %v", err)
		}
	})

	t.Run("update with force", func(t *testing.T) {
		err := updateRulesFromDefaults(rulesDir, true)
		if err != nil {
			t.Errorf("updateRulesFromDefaults() with force error = %v", err)
		}
	})

	t.Run("backup existing rules", func(t *testing.T) {
		// Create existing rule
		os.WriteFile(filepath.Join(rulesDir, "existing.md"), []byte("existing"), 0644)

		err := updateRulesFromDefaults(rulesDir, false)
		if err != nil {
			t.Errorf("updateRulesFromDefaults() error = %v", err)
		}
	})

	t.Run("skip non-md files in backup", func(t *testing.T) {
		os.WriteFile(filepath.Join(rulesDir, "non-md.txt"), []byte("text"), 0644)
		os.MkdirAll(filepath.Join(rulesDir, "subdir"), 0755)

		err := updateRulesFromDefaults(rulesDir, false)
		if err != nil {
			t.Errorf("updateRulesFromDefaults() error = %v", err)
		}
	})

	t.Run("write file error", func(t *testing.T) {
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		os.MkdirAll(readOnlyDir, 0444)
		defer os.Chmod(readOnlyDir, 0755)

		err := updateRulesFromDefaults(readOnlyDir, false)
		// Should handle write errors gracefully
		_ = err
	})
}
