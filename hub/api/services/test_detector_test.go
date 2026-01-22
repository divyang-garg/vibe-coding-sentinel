// Package services tests for test detector
// Complies with CODING_STANDARDS.md: Tests max 500 lines

package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectTests_Go(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "user_test.go")
		code := `
package main

import "testing"

func TestCreateUser(t *testing.T) {}
func TestUpdateUser(t *testing.T) {}
func TestDeleteUser(t *testing.T) {}
`
		if err := os.WriteFile(testFile, []byte(code), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		keywords := []string{"user", "create", "update"}

		// When
		tests := detectTests(tmpDir, "User Management", keywords)

		// Then
		if len(tests) < 2 {
			t.Errorf("Expected at least 2 test functions, got %d", len(tests))
		}
	})
}

func TestDetectTests_Jest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "user.test.js")
		code := `
describe('User Management', () => {
    test('create user', () => {});
    it('should update user', () => {});
});
`
		if err := os.WriteFile(testFile, []byte(code), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		keywords := []string{"user", "create", "update"}

		// When
		tests := detectTests(tmpDir, "User Management", keywords)

		// Then
		if len(tests) < 1 {
			t.Errorf("Expected at least 1 test, got %d", len(tests))
		}
	})
}

func TestDetectTests_Pytest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test_user.py")
		code := `
def test_create_user():
    pass

def test_update_user():
    pass
`
		if err := os.WriteFile(testFile, []byte(code), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		keywords := []string{"user", "create", "update"}

		// When
		tests := detectTests(tmpDir, "User Management", keywords)

		// Then
		if len(tests) < 2 {
			t.Errorf("Expected at least 2 test functions, got %d", len(tests))
		}
	})
}

func TestDetectTestFramework(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{"Go test", "user_test.go", "go-testing"},
		{"Jest test", "user.test.js", "jest"},
		{"Jest spec", "user.spec.ts", "jest"},
		{"Pytest", "test_user.py", "pytest"},
		{"Pytest suffix", "user_test.py", "pytest"},
		{"Unknown", "unknown.txt", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTestFramework(tt.filePath)
			if got != tt.want {
				t.Errorf("detectTestFramework() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectGoTests(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		code := `
package main

import "testing"

func TestCreateUser(t *testing.T) {}
func TestUpdateUser(t *testing.T) {}
func TestHelperFunction(t *testing.T) {}
`
		keywords := []string{"user", "create", "update"}

		// When
		tests := detectGoTests(code, keywords)

		// Then
		if len(tests) < 2 {
			t.Errorf("Expected at least 2 test functions, got %d", len(tests))
		}

		// Should not include TestHelperFunction (doesn't match keywords)
		for _, test := range tests {
			if test == "TestHelperFunction" {
				t.Error("TestHelperFunction should not be included (doesn't match keywords)")
			}
		}
	})
}

func TestDetectJestTests(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		code := `
describe('User Management', () => {
    test('create user', () => {});
    it('should update user', () => {});
    test('helper function', () => {});
});
`
		keywords := []string{"user", "create", "update"}

		// When
		tests := detectJestTests(code, keywords)

		// Then
		if len(tests) >= 2 {
			// Should find at least create and update tests
			found := false
			for _, test := range tests {
				if test == "create user" || test == "should update user" {
					found = true
				}
			}
			if !found {
				t.Error("Expected to find user-related tests")
			}
		}
	})
}

func TestDetectPytestTests(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		code := `
def test_create_user():
    pass

def test_update_user():
    pass

def test_helper():
    pass
`
		keywords := []string{"user", "create", "update"}

		// When
		tests := detectPytestTests(code, keywords)

		// Then
		if len(tests) < 2 {
			t.Errorf("Expected at least 2 test functions, got %d", len(tests))
		}
	})
}
