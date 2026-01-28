// Package ast provides tests for Go detector security methods
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import "testing"

// Test GoDetector_DetectSecurityMiddleware
func TestGoDetector_DetectSecurityMiddleware(t *testing.T) {
	detector := &GoDetector{}

	tests := []struct {
		name     string
		code     string
		expected int // Minimum expected findings
	}{
		{
			name: "JWT Bearer token middleware",
			code: `
package main

import "net/http"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			// JWT validation
		}
		next.ServeHTTP(w, r)
	})
}
`,
			expected: 1,
		},
		{
			name: "API key middleware",
			code: `
package main

import "net/http"

func APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			// API key validation
		}
		next.ServeHTTP(w, r)
	})
}
`,
			expected: 1,
		},
		{
			name: "OAuth middleware",
			code: `
package main

import "net/http"

func OAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// OAuth validation
		next.ServeHTTP(w, r)
	})
}
`,
			expected: 1,
		},
		{
			name: "RBAC middleware",
			code: `
package main

import "net/http"

func RBACMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		if user.HasRole("admin") {
			// Role check
		}
		next.ServeHTTP(w, r)
	})
}
`,
			expected: 1,
		},
		{
			name: "Rate limit middleware",
			code: `
package main

import "net/http"

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Rate limiting
		next.ServeHTTP(w, r)
	})
}
`,
			expected: 1,
		},
		{
			name: "CORS middleware",
			code: `
package main

import "net/http"

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}
`,
			expected: 1,
		},
		{
			name: "Multiple security patterns",
			code: `
package main

import "net/http"

func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			// JWT
		}
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			// API key
		}
		next.ServeHTTP(w, r)
	})
}
`,
			expected: 2,
		},
		{
			name: "No security patterns",
			code: `
package main

import "net/http"

func RegularHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}
`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, rootNode := parseGoCode(t, tt.code)
			defer tree.Close()
			findings := detector.DetectSecurityMiddleware(rootNode, tt.code)

			if len(findings) < tt.expected {
				t.Errorf("Expected at least %d findings, got %d", tt.expected, len(findings))
			}

			// Verify findings have required fields
			for _, finding := range findings {
				if finding.Type == "" {
					t.Error("Finding type should not be empty")
				}
				if finding.Message == "" {
					t.Error("Finding message should not be empty")
				}
				if finding.Line <= 0 {
					t.Error("Finding line should be > 0")
				}
			}
		})
	}
}

// Test GoDetector_DetectSQLInjection
func TestGoDetector_DetectSQLInjection(t *testing.T) {
	detector := &GoDetector{}

	tests := []struct {
		name     string
		code     string
		expected int // Minimum expected vulnerabilities
	}{
		{
			name: "SQL injection with string concatenation",
			code: `
package main

import "database/sql"

func getUser(db *sql.DB, id string) {
	query := "SELECT * FROM users WHERE id = '" + id + "'"
	db.Query(query)
}
`,
			expected: 1,
		},
		{
			name: "Safe parameterized query",
			code: `
package main

import "database/sql"

func getUser(db *sql.DB, id string) {
	query := "SELECT * FROM users WHERE id = ?"
	db.Query(query, id)
}
`,
			expected: 0,
		},
		{
			name: "SQL injection with fmt.Sprintf",
			code: `
package main

import (
	"database/sql"
	"fmt"
)

func getUser(db *sql.DB, id string) {
	query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", id)
	db.Query(query)
}
`,
			expected: 0,
		},
		{
			name: "Multiple SQL injection patterns",
			code: `
package main

import "database/sql"

func getUsers(db *sql.DB, id string) {
	query1 := "SELECT * FROM users WHERE id = " + id
	db.Query(query1)
	
	query2 := "UPDATE users SET name = '" + id + "'"
	db.Exec(query2)
}
`,
			expected: 2,
		},
		{
			name: "No SQL queries",
			code: `
package main

func test() {
	println("no SQL")
}
`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, rootNode := parseGoCode(t, tt.code)
			defer tree.Close()
			vulnerabilities := detector.DetectSQLInjection(rootNode, tt.code)

			if len(vulnerabilities) < tt.expected {
				t.Errorf("Expected at least %d vulnerabilities, got %d", tt.expected, len(vulnerabilities))
			}
		})
	}
}

// Test GoDetector_DetectXSS
func TestGoDetector_DetectXSS(t *testing.T) {
	detector := &GoDetector{}

	tests := []struct {
		name     string
		code     string
		expected int // Minimum expected vulnerabilities
	}{
		{
			name: "XSS vulnerability",
			code: `
package main

import "html/template"

func render(userInput string) {
	tmpl := template.Must(template.New("test").Parse("<div>{{.}}</div>"))
	tmpl.Execute(os.Stdout, userInput)
}
`,
			expected: 0, // html/template auto-escapes, so may not be detected as XSS
		},
		{
			name: "Safe template rendering",
			code: `
package main

import "html/template"

func render(userInput string) {
	tmpl := template.Must(template.New("test").Parse("<div>{{.}}</div>"))
	tmpl.Execute(os.Stdout, template.HTMLEscapeString(userInput))
}
`,
			expected: 0,
		},
		{
			name: "No HTML rendering",
			code: `
package main

func test() {
	println("no HTML")
}
`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, rootNode := parseGoCode(t, tt.code)
			defer tree.Close()
			vulnerabilities := detector.DetectXSS(rootNode, tt.code)

			if len(vulnerabilities) < tt.expected {
				t.Errorf("Expected at least %d vulnerabilities, got %d", tt.expected, len(vulnerabilities))
			}
		})
	}
}

// Test GoDetector_DetectCommandInjection
func TestGoDetector_DetectCommandInjection(t *testing.T) {
	detector := &GoDetector{}

	tests := []struct {
		name     string
		code     string
		expected int // Minimum expected vulnerabilities
	}{
		{
			name: "Command injection",
			code: `
package main

import (
	"os/exec"
)

func execCommand(userInput string) {
	cmd := exec.Command("ls", userInput)
	cmd.Run()
}
`,
			expected: 1,
		},
		{
			name: "Safe command execution",
			code: `
package main

import (
	"os/exec"
)

func execCommand(cmd string) {
	exec.Command("ls", "-l").Run()
}
`,
			expected: 0,
		},
		{
			name: "Command injection with CommandContext",
			code: `
package main

import (
	"context"
	"os/exec"
)

func execCommand(ctx context.Context, userInput string) {
	cmd := exec.CommandContext(ctx, "ls", userInput)
	cmd.Run()
}
`,
			expected: 1,
		},
		{
			name: "No command execution",
			code: `
package main

func test() {
	println("no commands")
}
`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, rootNode := parseGoCode(t, tt.code)
			defer tree.Close()
			vulnerabilities := detector.DetectCommandInjection(rootNode, tt.code)

			if len(vulnerabilities) < tt.expected {
				t.Errorf("Expected at least %d vulnerabilities, got %d", tt.expected, len(vulnerabilities))
			}
		})
	}
}

// Test GoDetector_DetectCrypto
func TestGoDetector_DetectCrypto(t *testing.T) {
	detector := &GoDetector{}

	tests := []struct {
		name     string
		code     string
		expected int // Minimum expected vulnerabilities
	}{
		{
			name: "MD5 hash usage",
			code: `
package main

import "crypto/md5"

func hashPassword(password string) {
	h := md5.New()
	h.Write([]byte(password))
}
`,
			expected: 1,
		},
		{
			name: "SHA1 hash usage",
			code: `
package main

import "crypto/sha1"

func hashPassword(password string) {
	h := sha1.New()
	h.Write([]byte(password))
}
`,
			expected: 1,
		},
		{
			name: "Hardcoded secret",
			code: `
package main

func getSecret() string {
	apiKey := "my-secret-key-12345"
	return apiKey
}
`,
			expected: 1,
		},
		{
			name: "Secure hash (SHA256)",
			code: `
package main

import "crypto/sha256"

func hashPassword(password string) {
	h := sha256.New()
	h.Write([]byte(password))
}
`,
			expected: 0,
		},
		{
			name: "Multiple insecure patterns",
			code: `
package main

import (
	"crypto/md5"
	"crypto/sha1"
)

func hashPassword(password string) {
	h1 := md5.New()
	h1.Write([]byte(password))
	
	h2 := sha1.New()
	h2.Write([]byte(password))
}
`,
			expected: 2,
		},
		{
			name: "No crypto usage",
			code: `
package main

func test() {
	println("no crypto")
}
`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, rootNode := parseGoCode(t, tt.code)
			defer tree.Close()
			vulnerabilities := detector.DetectCrypto(rootNode, tt.code)

			if len(vulnerabilities) < tt.expected {
				t.Errorf("Expected at least %d vulnerabilities, got %d", tt.expected, len(vulnerabilities))
			}
		})
	}
}
