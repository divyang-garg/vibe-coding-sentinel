// Package ast - Comprehensive accuracy validation tests
// Tests improved SQL injection and command injection detection
package ast

import (
	"context"
	"testing"
)

func TestAccuracy_SQLInjection_ComplexCases(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name     string
		code     string
		language string
		shouldFind bool
	}{
		{
			name: "Simple string concatenation",
			code: `db.Query("SELECT * FROM users WHERE id = " + id)`,
			language: "go",
			shouldFind: true,
		},
		{
			name: "Multi-line query construction",
			code: `package main
func test(id string) {
	query := "SELECT * FROM users WHERE id = " + id
	db.Query(query)
}`,
			language: "go",
			shouldFind: true,
		},
		{
			name: "Query variable with concatenation",
			code: `package main
func getUser(id string) {
	var query string
	query = "SELECT * FROM users WHERE id = " + id
	db.QueryRow(query)
}`,
			language: "go",
			shouldFind: true,
		},
		{
			name: "Safe parameterized query",
			code: `db.Query("SELECT * FROM users WHERE id = ?", id)`,
			language: "go",
			shouldFind: false,
		},
	}

	detected := 0
	total := len(testCases)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vulns, _, _, err := AnalyzeSecurity(ctx, tc.code, tc.language, "all")
			if err != nil {
				t.Logf("⚠️  Analysis error: %v", err)
				return
			}

			found := false
			for _, vuln := range vulns {
				if vuln.Type == "sql_injection" {
					found = true
					break
				}
			}

			if found && tc.shouldFind {
				detected++
				t.Logf("✅ Correctly detected SQL injection")
			} else if !found && !tc.shouldFind {
				detected++
				t.Logf("✅ Correctly identified safe code")
			} else if found && !tc.shouldFind {
				t.Errorf("❌ False positive: Safe code flagged as vulnerable")
			} else {
				t.Errorf("❌ False negative: Vulnerable code not detected")
			}
		})
	}

	accuracy := float64(detected) / float64(total) * 100
	t.Logf("\n📊 SQL Injection Detection Accuracy: %.1f%% (%d/%d)", accuracy, detected, total)
	
	if accuracy >= 90.0 {
		t.Logf("✅ Excellent SQL injection detection accuracy")
	} else if accuracy >= 75.0 {
		t.Logf("✅ Good SQL injection detection accuracy")
	} else {
		t.Errorf("❌ SQL injection detection accuracy below target: %.1f%%", accuracy)
	}
}

func TestAccuracy_CommandInjection_PythonCases(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name     string
		code     string
		language string
		shouldFind bool
	}{
		{
			name: "os.system with user input",
			code: `import os
os.system("cat " + filename)`,
			language: "python",
			shouldFind: true,
		},
		{
			name: "subprocess.call with shell=True",
			code: `import subprocess
subprocess.call(["sh", "-c", user_input], shell=True)`,
			language: "python",
			shouldFind: true,
		},
		{
			name: "subprocess.run with list (safe)",
			code: `import subprocess
subprocess.run(["cat", filename], check=True)`,
			language: "python",
			shouldFind: false,
		},
		{
			name: "subprocess with string command",
			code: `import subprocess
subprocess.call("cat " + filename)`,
			language: "python",
			shouldFind: true,
		},
	}

	detected := 0
	total := len(testCases)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vulns, _, _, err := AnalyzeSecurity(ctx, tc.code, tc.language, "all")
			if err != nil {
				t.Logf("⚠️  Analysis error: %v", err)
				return
			}

			found := false
			for _, vuln := range vulns {
				if vuln.Type == "command_injection" {
					found = true
					break
				}
			}

			if found && tc.shouldFind {
				detected++
				t.Logf("✅ Correctly detected command injection")
			} else if !found && !tc.shouldFind {
				detected++
				t.Logf("✅ Correctly identified safe code")
			} else if found && !tc.shouldFind {
				t.Errorf("❌ False positive: Safe code flagged as vulnerable")
			} else {
				t.Errorf("❌ False negative: Vulnerable code not detected")
			}
		})
	}

	accuracy := float64(detected) / float64(total) * 100
	t.Logf("\n📊 Command Injection Detection Accuracy: %.1f%% (%d/%d)", accuracy, detected, total)
	
	if accuracy >= 90.0 {
		t.Logf("✅ Excellent command injection detection accuracy")
	} else if accuracy >= 75.0 {
		t.Logf("✅ Good command injection detection accuracy")
	} else {
		t.Errorf("❌ Command injection detection accuracy below target: %.1f%%", accuracy)
	}
}

func TestAccuracy_OverallSecurityDetection(t *testing.T) {
	ctx := context.Background()

	allTests := []struct {
		name     string
		code     string
		language string
		vulnType string
		shouldFind bool
	}{
		// SQL Injection
		{"SQL: Simple concat", `db.Query("SELECT * FROM users WHERE id = " + id)`, "go", "sql_injection", true},
		{"SQL: Multi-line", `query := "SELECT * FROM users WHERE id = " + id\ndb.Query(query)`, "go", "sql_injection", true},
		{"SQL: Safe parameterized", `db.Query("SELECT * FROM users WHERE id = ?", id)`, "go", "sql_injection", false},
		
		// XSS
		{"XSS: innerHTML", `document.getElementById('div').innerHTML = userInput`, "javascript", "xss", true},
		{"XSS: textContent (safe)", `document.getElementById('div').textContent = userInput`, "javascript", "xss", false},
		
		// Command Injection
		{"CMD: os.system", `os.system("cat " + filename)`, "python", "command_injection", true},
		{"CMD: subprocess shell=True", `subprocess.call(["sh", "-c", user_input], shell=True)`, "python", "command_injection", true},
		{"CMD: subprocess list (safe)", `subprocess.run(["cat", filename])`, "python", "command_injection", false},
		
		// Secrets
		{"SECRET: API key", `const API_KEY = "sk_live_1234567890"`, "javascript", "hardcoded_secret", true},
		{"SECRET: Password", `password = "SuperSecret123!"`, "python", "hardcoded_secret", true},
		
		// Crypto
		{"CRYPTO: MD5", `import "crypto/md5"\nhash := md5.Sum(data)`, "go", "insecure_crypto", true},
		{"CRYPTO: SHA256 (safe)", `import "crypto/sha256"\nhash := sha256.Sum256(data)`, "go", "insecure_crypto", false},
	}

	detected := 0
	total := len(allTests)
	byType := make(map[string]int)
	byTypeTotal := make(map[string]int)

	for _, tc := range allTests {
		byTypeTotal[tc.vulnType]++
		
		vulns, _, _, err := AnalyzeSecurity(ctx, tc.code, tc.language, "all")
		if err != nil {
			continue
		}

		found := false
		for _, vuln := range vulns {
			if vuln.Type == tc.vulnType {
				found = true
				break
			}
		}

		if (found && tc.shouldFind) || (!found && !tc.shouldFind) {
			detected++
			if tc.shouldFind {
				byType[tc.vulnType]++
			}
		}
	}

	overallAccuracy := float64(detected) / float64(total) * 100
	t.Logf("\n📊 Overall Security Detection Accuracy: %.1f%% (%d/%d)", overallAccuracy, detected, total)
	
	// Accuracy by type
	for vulnType, total := range byTypeTotal {
		accuracy := float64(byType[vulnType]) / float64(total) * 100
		t.Logf("  %s: %.1f%%", vulnType, accuracy)
	}

	if overallAccuracy >= 95.0 {
		t.Logf("✅ EXCELLENT: Overall accuracy exceeds 95%% target")
	} else if overallAccuracy >= 85.0 {
		t.Logf("✅ GOOD: Overall accuracy meets minimum requirements")
	} else {
		t.Errorf("❌ Overall accuracy below target: %.1f%%", overallAccuracy)
	}
}
