// Package services tests for helper stubs
package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"sentinel-hub-api/models"
)

func TestVerifyTask_EmptyInputs(t *testing.T) {
	testCases := []struct {
		name         string
		taskID       string
		codebasePath string
		wantError    bool
	}{
		{
			name:         "empty task ID",
			taskID:       "",
			codebasePath: "/tmp/test",
			wantError:    true,
		},
		{
			name:         "empty codebase path",
			taskID:       "task-123",
			codebasePath: "",
			wantError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := VerifyTask(context.Background(), tc.taskID, tc.codebasePath, false)
			if (err != nil) != tc.wantError {
				t.Errorf("VerifyTask() error = %v, wantError %v", err, tc.wantError)
			}
		})
	}
}

func TestAnalyzeTaskCompletion_FileExists(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	codebasePath := tmpDir

	// Create a test file
	testFile := filepath.Join(codebasePath, "test.go")
	err := os.WriteFile(testFile, []byte("package main\n\nfunc testFunction() {\n\t// Test implementation\n}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create task with file path
	task := &Task{
		ID:          "task-123",
		Title:       "Implement test function",
		Description: "Add test function implementation",
		FilePath:    "test.go",
	}

	confidence, evidence := analyzeTaskCompletion(context.Background(), task, codebasePath)

	// Should have some confidence if file exists
	if confidence == 0.0 {
		t.Error("Expected confidence > 0 when file exists")
	}

	// Check evidence
	if evidence["file_exists"] != true {
		t.Error("Expected file_exists to be true")
	}

	if evidence["file_path"] != "test.go" {
		t.Errorf("Expected file_path to be 'test.go', got %v", evidence["file_path"])
	}

	// Should have keywords extracted
	keywords, ok := evidence["keywords"].([]string)
	if !ok || len(keywords) == 0 {
		t.Error("Expected keywords to be extracted")
	}

	t.Logf("Confidence: %.2f", confidence)
	t.Logf("Evidence: %+v", evidence)
}

func TestAnalyzeTaskCompletion_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	codebasePath := tmpDir

	task := &Task{
		ID:          "task-123",
		Title:       "Implement missing function",
		Description: "Add missing function",
		FilePath:    "nonexistent.go",
	}

	confidence, evidence := analyzeTaskCompletion(context.Background(), task, codebasePath)

	// Should have lower confidence when file doesn't exist
	if evidence["file_exists"] != false {
		t.Error("Expected file_exists to be false")
	}

	// But might still have some confidence from keyword matching
	t.Logf("Confidence: %.2f", confidence)
	t.Logf("Evidence: %+v", evidence)
}

func TestAnalyzeTaskCompletion_KeywordMatching(t *testing.T) {
	tmpDir := t.TempDir()
	codebasePath := tmpDir

	// Create file with matching keywords
	testFile := filepath.Join(codebasePath, "auth.go")
	content := `package main

func authenticateUser(username, password string) bool {
	// Authentication implementation
	return true
}`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	task := &Task{
		ID:          "task-123",
		Title:       "Implement user authentication",
		Description: "Add authentication function",
		FilePath:    "auth.go",
	}

	confidence, evidence := analyzeTaskCompletion(context.Background(), task, codebasePath)

	// Should have higher confidence with keyword matches
	keywordScore, ok := evidence["keyword_score"].(float64)
	if ok && keywordScore > 0 {
		if confidence < 0.5 {
			t.Errorf("Expected higher confidence with keyword matches, got %.2f", confidence)
		}
	}

	t.Logf("Confidence: %.2f", confidence)
	t.Logf("Keyword Score: %.2f", keywordScore)
}

func TestDetermineVerificationStatus(t *testing.T) {
	testCases := []struct {
		name       string
		confidence float64
		expected   models.VerificationStatus
	}{
		{
			name:       "high confidence",
			confidence: 0.9,
			expected:   models.VerificationStatusVerified,
		},
		{
			name:       "medium confidence",
			confidence: 0.7,
			expected:   models.VerificationStatusPending,
		},
		{
			name:       "low confidence",
			confidence: 0.3,
			expected:   models.VerificationStatusPending,
		},
		{
			name:       "exactly 0.8",
			confidence: 0.8,
			expected:   models.VerificationStatusVerified,
		},
		{
			name:       "exactly 0.5",
			confidence: 0.5,
			expected:   models.VerificationStatusPending,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := determineVerificationStatus(tc.confidence)
			if result != tc.expected {
				t.Errorf("determineVerificationStatus(%.2f) = %v, want %v", tc.confidence, result, tc.expected)
			}
		})
	}
}

func TestFindTestFile(t *testing.T) {
	testCases := []struct {
		name           string
		filePath       string
		createTestFile bool
		testFileName   string
		wantFound      bool
	}{
		{
			name:           "Go test file in same directory",
			filePath:       "service.go",
			createTestFile: true,
			testFileName:   "service_test.go",
			wantFound:      true,
		},
		{
			name:           "No test file",
			filePath:       "nonexistent.go",
			createTestFile: false,
			wantFound:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			codebasePath := tmpDir

			// Create main file
			mainFile := filepath.Join(codebasePath, tc.filePath)
			err := os.WriteFile(mainFile, []byte("package main"), 0644)
			if err != nil {
				t.Fatalf("Failed to create main file: %v", err)
			}

			if tc.createTestFile {
				testFile := filepath.Join(codebasePath, tc.testFileName)
				err := os.WriteFile(testFile, []byte("package main\n\nfunc TestService(t *testing.T) {}"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			result := findTestFile(codebasePath, tc.filePath)
			if (result != "") != tc.wantFound {
				t.Errorf("findTestFile() = %q, wantFound %v", result, tc.wantFound)
			}
		})
	}
}

func TestSearchCodebaseForKeywords(t *testing.T) {
	tmpDir := t.TempDir()
	codebasePath := tmpDir

	// Create files with keywords
	files := map[string]string{
		"auth.go":    "package main\nfunc authenticate() {}",
		"user.go":    "package main\nfunc getUser() {}",
		"service.go": "package main\nfunc process() {}",
	}

	for fileName, content := range files {
		filePath := filepath.Join(codebasePath, fileName)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", fileName, err)
		}
	}

	keywords := []string{"auth", "user"}
	matches := searchCodebaseForKeywords(context.Background(), codebasePath, keywords)

	// Should find at least 2 files (auth.go and user.go)
	if matches < 2 {
		t.Errorf("Expected at least 2 matches, got %d", matches)
	}

	t.Logf("Found %d matches for keywords %v", matches, keywords)
}
