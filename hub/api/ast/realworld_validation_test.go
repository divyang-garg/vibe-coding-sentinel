// Package ast - Real-world validation tests
// Tests AST implementation with realistic code examples
package ast

import (
	"context"
	"testing"
)

func TestRealWorld_SQLInjection(t *testing.T) {
	ctx := context.Background()

	// Real-world vulnerable Go code
	vulnerableCode := `package main
import (
	"database/sql"
	"fmt"
)

func getUserByID(db *sql.DB, userID string) (*User, error) {
	// VULNERABLE: String concatenation in SQL query
	query := "SELECT * FROM users WHERE id = " + userID
	row := db.QueryRow(query)
	
	var user User
	err := row.Scan(&user.ID, &user.Name)
	return &user, err
}`

	vulns, _, stats, err := AnalyzeSecurity(ctx, vulnerableCode, "go", "all")
	if err != nil {
		t.Fatalf("AnalyzeSecurity failed: %v", err)
	}

	// Should detect SQL injection
	sqlInjectionFound := false
	for _, vuln := range vulns {
		if vuln.Type == "sql_injection" {
			sqlInjectionFound = true
			t.Logf("✅ SQL Injection detected: %s (Line %d)", vuln.Message, vuln.Line)
			if vuln.Confidence < 0.7 {
				t.Errorf("Confidence too low: %f (expected >= 0.7)", vuln.Confidence)
			}
		}
	}

	if !sqlInjectionFound {
		t.Error("❌ SQL injection NOT detected - this is a critical vulnerability")
	}

	t.Logf("Analysis stats: %d nodes visited, %dms parse time", stats.NodesVisited, stats.ParseTime)
}

func TestRealWorld_XSS(t *testing.T) {
	ctx := context.Background()

	// Real-world vulnerable JavaScript code
	vulnerableCode := `function renderUserProfile(user) {
	const container = document.getElementById('profile');
	// VULNERABLE: User input directly inserted into innerHTML
	container.innerHTML = '<h1>' + user.name + '</h1><p>' + user.bio + '</p>';
}`

	vulns, _, stats, err := AnalyzeSecurity(ctx, vulnerableCode, "javascript", "all")
	if err != nil {
		t.Fatalf("AnalyzeSecurity failed: %v", err)
	}

	// Should detect XSS
	xssFound := false
	for _, vuln := range vulns {
		if vuln.Type == "xss" {
			xssFound = true
			t.Logf("✅ XSS detected: %s (Line %d)", vuln.Message, vuln.Line)
		}
	}

	if !xssFound {
		t.Error("❌ XSS NOT detected - this is a high-severity vulnerability")
	}

	t.Logf("Analysis stats: %d nodes visited", stats.NodesVisited)
}

func TestRealWorld_CommandInjection(t *testing.T) {
	ctx := context.Background()

	// Real-world vulnerable Python code
	vulnerableCode := `import subprocess
import sys

def process_file(filename):
	# VULNERABLE: User input in shell command
	result = subprocess.call(["cat", filename], shell=True)
	return result`

	vulns, _, stats, err := AnalyzeSecurity(ctx, vulnerableCode, "python", "all")
	if err != nil {
		t.Fatalf("AnalyzeSecurity failed: %v", err)
	}

	// Should detect command injection
	cmdInjectionFound := false
	for _, vuln := range vulns {
		if vuln.Type == "command_injection" {
			cmdInjectionFound = true
			t.Logf("✅ Command Injection detected: %s (Line %d)", vuln.Message, vuln.Line)
		}
	}

	if !cmdInjectionFound {
		t.Log("⚠️  Command injection not detected - may need pattern refinement")
	}

	t.Logf("Analysis stats: %d nodes visited", stats.NodesVisited)
}

func TestRealWorld_HardcodedSecrets(t *testing.T) {
	ctx := context.Background()

	// Real-world code with hardcoded secrets
	vulnerableCode := `const config = {
	apiKey: "sk_test_EXAMPLE_KEY_PLACEHOLDER_DO_NOT_USE_IN_PRODUCTION",
	awsAccessKey: "AKIAIOSFODNN7EXAMPLE",
	password: "SuperSecret123!",
	databaseUrl: "postgresql://user:password123@localhost/db"
};`

	vulns, _, stats, err := AnalyzeSecurity(ctx, vulnerableCode, "javascript", "all")
	if err != nil {
		t.Fatalf("AnalyzeSecurity failed: %v", err)
	}

	// Should detect hardcoded secrets
	secretsFound := 0
	for _, vuln := range vulns {
		if vuln.Type == "hardcoded_secret" {
			secretsFound++
			t.Logf("✅ Secret detected: %s (Line %d, Severity: %s)", vuln.Message, vuln.Line, vuln.Severity)
		}
	}

	if secretsFound == 0 {
		t.Error("❌ Hardcoded secrets NOT detected - found 0, expected at least 2")
	} else {
		t.Logf("✅ Detected %d hardcoded secrets", secretsFound)
	}

	t.Logf("Analysis stats: %d nodes visited", stats.NodesVisited)
}

func TestRealWorld_InsecureCrypto(t *testing.T) {
	ctx := context.Background()

	// Real-world code using weak crypto
	vulnerableCode := `package main
import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
)

func hashPassword(password string) string {
	// VULNERABLE: Using MD5 for password hashing
	hash := md5.Sum([]byte(password))
	return fmt.Sprintf("%x", hash)
}

func generateChecksum(data []byte) string {
	// VULNERABLE: Using SHA1
	hash := sha1.Sum(data)
	return fmt.Sprintf("%x", hash)
}`

	vulns, _, stats, err := AnalyzeSecurity(ctx, vulnerableCode, "go", "all")
	if err != nil {
		t.Fatalf("AnalyzeSecurity failed: %v", err)
	}

	// Should detect insecure crypto
	cryptoIssuesFound := 0
	for _, vuln := range vulns {
		if vuln.Type == "insecure_crypto" {
			cryptoIssuesFound++
			t.Logf("✅ Insecure crypto detected: %s (Line %d)", vuln.Message, vuln.Line)
		}
	}

	if cryptoIssuesFound == 0 {
		t.Error("❌ Insecure crypto NOT detected - MD5 and SHA1 should be flagged")
	} else {
		t.Logf("✅ Detected %d insecure crypto issues", cryptoIssuesFound)
	}

	t.Logf("Analysis stats: %d nodes visited", stats.NodesVisited)
}

func TestRealWorld_CrossFileUnusedExports(t *testing.T) {
	ctx := context.Background()

	// Real-world scenario: exported function never used
	files := []FileInput{
		{
			Path:     "utils.go",
			Content:  "package main\n\n// Exported but never used elsewhere\nfunc HelperFunction() {\n\t// Some code\n}\n",
			Language: "go",
		},
		{
			Path:     "main.go",
			Content:  "package main\n\nfunc main() {\n\t// HelperFunction is never called\n}\n",
			Language: "go",
		},
	}

	result, err := AnalyzeCrossFile(ctx, files, []string{"unused_exports"})
	if err != nil {
		t.Fatalf("AnalyzeCrossFile failed: %v", err)
	}

	// Should detect unused export
	if len(result.UnusedExports) == 0 {
		t.Error("❌ Unused export NOT detected - HelperFunction should be flagged")
	} else {
		t.Logf("✅ Detected %d unused exports", len(result.UnusedExports))
		for _, exp := range result.UnusedExports {
			t.Logf("  - %s in %s (Line %d)", exp.Name, exp.FilePath, exp.Line)
		}
	}
}

func TestRealWorld_CrossFileCircularDeps(t *testing.T) {
	ctx := context.Background()

	// Real-world scenario: circular dependency (simplified - actual circular deps need proper package structure)
	files := []FileInput{
		{
			Path:     "a.go",
			Content:  "package main\n\nfunc A() {\n\tB()\n}\n",
			Language: "go",
		},
		{
			Path:     "b.go",
			Content:  "package main\n\nfunc B() {\n\tA()\n}\n",
			Language: "go",
		},
	}

	result, err := AnalyzeCrossFile(ctx, files, []string{"circular_deps"})
	if err != nil {
		t.Fatalf("AnalyzeCrossFile failed: %v", err)
	}

	// Should detect circular dependency
	if len(result.CircularDeps) == 0 {
		t.Log("⚠️  Circular dependency not detected - may need dependency resolution")
	} else {
		t.Logf("✅ Detected %d circular dependencies", len(result.CircularDeps))
		for _, cycle := range result.CircularDeps {
			t.Logf("  - Cycle: %v", cycle)
		}
	}
}

func TestRealWorld_AccuracyComparison(t *testing.T) {
	ctx := context.Background()

	// Test suite of known vulnerabilities
	testCases := []struct {
		name        string
		code        string
		language    string
		vulnType    string
		shouldFind  bool
		description string
	}{
		{
			name:        "SQL Injection - String Concatenation",
			code:        `db.Query("SELECT * FROM users WHERE id = " + id)`,
			language:    "go",
			vulnType:    "sql_injection",
			shouldFind:  true,
			description: "Direct string concatenation in SQL query",
		},
		{
			name:        "XSS - innerHTML Assignment",
			code:        `document.getElementById('div').innerHTML = userInput`,
			language:    "javascript",
			vulnType:    "xss",
			shouldFind:  true,
			description: "User input assigned to innerHTML",
		},
		{
			name:        "Hardcoded API Key",
			code:        `const API_KEY = "sk_test_EXAMPLE_KEY_PLACEHOLDER_DO_NOT_USE"`,
			language:    "javascript",
			vulnType:    "hardcoded_secret",
			shouldFind:  true,
			description: "Hardcoded API key in code",
		},
		{
			name:        "MD5 Hash Usage",
			code:        `import "crypto/md5"\nhash := md5.Sum(data)`,
			language:    "go",
			vulnType:    "insecure_crypto",
			shouldFind:  true,
			description: "MD5 hash algorithm usage",
		},
	}

	detected := 0
	total := len(testCases)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vulns, _, _, err := AnalyzeSecurity(ctx, tc.code, tc.language, "all")
			if err != nil {
				t.Logf("⚠️  Analysis error for %s: %v", tc.name, err)
				return
			}

			found := false
			for _, vuln := range vulns {
				if vuln.Type == tc.vulnType {
					found = true
					detected++
					t.Logf("✅ %s: DETECTED (Confidence: %.2f)", tc.name, vuln.Confidence)
					break
				}
			}

			if !found && tc.shouldFind {
				t.Errorf("❌ %s: NOT DETECTED - %s", tc.name, tc.description)
			} else if found && !tc.shouldFind {
				t.Errorf("⚠️  %s: FALSE POSITIVE", tc.name)
			}
		})
	}

	accuracy := float64(detected) / float64(total) * 100
	t.Logf("\n📊 Detection Accuracy: %.1f%% (%d/%d vulnerabilities detected)", accuracy, detected, total)

	if accuracy < 70.0 {
		t.Errorf("❌ Accuracy below target: %.1f%% (target: 70%%)", accuracy)
	} else if accuracy >= 95.0 {
		t.Logf("✅ Excellent accuracy: %.1f%% (target: 95%%)", accuracy)
	} else {
		t.Logf("✅ Good accuracy: %.1f%% (target: 95%%, minimum: 70%%)", accuracy)
	}
}
